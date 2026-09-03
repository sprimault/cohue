// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La police telle que son manifeste la décrit : la cellule, la ligne de base, la
// place de chaque glyphe dans la planche et son avance. Rien qui dessine — le
// dessin est dans `internal/render`, qui lit ces mesures.

// Package ui porte ce que l'interface lit, sans rien afficher.
//
// Il existe pour que le décodage d'un manifeste d'interface reste vérifiable :
// `internal/render` importe Ebitengine, donc n'a pas de test et n'en aura pas,
// alors que refuser un manifeste incohérent est exactement ce qui doit l'être.
// La frontière est celle du pixel : ce qui mesure vit ici, ce qui peint vit là.
package ui

import (
	"fmt"
	"io/fs"

	"github.com/sprimault/cohue/internal/manifest"
)

// FormatFont est la version de format que ce binaire sait lire.
const FormatFont = 1

// Font est la planche de glyphes, mesurée.
//
// Elle ne porte aucune image : la planche est un fichier que le rendu charge, et
// ce type n'en connaît que le nom et la géométrie. C'est ce qui lui permet de se
// tester sans écran.
type Font struct {
	// Sheet est le nom de la planche, relatif au dossier du manifeste.
	Sheet string
	// Cell est la cellule d'un glyphe, largeur puis hauteur. Elle n'est pas
	// carrée : la hauteur loge la ligne d'accent au-dessus de la capitale et le
	// jambage en dessous.
	Cell [2]int
	// Baseline est la ligne de base, comptée depuis le haut de la cellule.
	Baseline int
	// Native est la taille à laquelle la fonte a été rastérisée. Les tailles
	// d'affichage admises en sont les multiples entiers : à deux fois, chaque
	// pixel devient un bloc de deux sur deux et la grille tient ; à une fois et
	// demie elle se casse et le rendu redevient une interpolation.
	Native int

	// place donne la colonne d'un glyphe dans la planche.
	place map[rune]int
	// advance donne l'avance d'un glyphe, indexée par sa place.
	advance []int
}

// LoadFont lit le manifeste d'interface et rend la police qu'il décrit.
//
// Les manques sont listés en une fois : qui met au point un manifeste veut la
// liste, pas un aller-retour par défaut.
func LoadFont(fsys fs.FS, chemin string) (*Font, error) {
	brut, err := manifest.Decode[rawInterface](fsys, chemin)
	if err != nil {
		return nil, err
	}
	if brut.Format != FormatFont {
		return nil, fmt.Errorf("%s: %w : %d, ce binaire lit la %d",
			chemin, manifest.ErrUnsupportedFormat, brut.Format, FormatFont)
	}

	var manques []string
	dire := func(format string, args ...any) {
		manques = append(manques, fmt.Sprintf(format, args...))
	}

	p := brut.Interface.Font
	glyphes := []rune(p.Glyphs)

	if p.Sheet == "" {
		dire("police : planche non nommée")
	}
	if p.Cell[0] <= 0 || p.Cell[1] <= 0 {
		dire("police : cellule de %dx%d", p.Cell[0], p.Cell[1])
	}
	if len(glyphes) == 0 {
		dire("police : table des glyphes vide")
	}
	if len(p.Advances) != len(glyphes) {
		dire("police : %d avances pour %d glyphes", len(p.Advances), len(glyphes))
	}
	// La ligne de base hors de la cellule poserait tout le texte au-dessus ou
	// au-dessous de là où il doit être, ce qui ne se verrait qu'à l'écran.
	if p.Baseline <= 0 || p.Baseline >= p.Cell[1] {
		dire("police : ligne de base %d hors d'une cellule haute de %d",
			p.Baseline, p.Cell[1])
	}

	place := make(map[rune]int, len(glyphes))
	for i, r := range glyphes {
		// Un doublon rend le second exemplaire inatteignable : le texte qui
		// l'emploie dessine le premier, et la planche porte une colonne que rien
		// ne lit.
		if _, vu := place[r]; vu {
			dire("police : le glyphe U+%04X est déclaré deux fois", r)
			continue
		}
		place[r] = i
	}

	if len(manques) > 0 {
		return nil, &manifest.Invalid{Path: chemin, Missing: manques}
	}
	return &Font{
		Sheet:    p.Sheet,
		Cell:     p.Cell,
		Baseline: p.Baseline,
		Native:   p.Native,
		place:    place,
		advance:  p.Advances,
	}, nil
}

// Place rend la colonne d'un glyphe dans la planche.
//
// Le second retour dit si la table le porte. Un caractère absent n'est pas une
// erreur de chargement — le manifeste ne peut pas connaître ce que le jeu
// écrira — mais il ne se dessine pas pour autant.
func (f *Font) Place(r rune) (int, bool) {
	i, ok := f.place[r]
	return i, ok
}

// Advance rend la largeur d'une chaîne, en pixels de tampon.
//
// C'est elle qui permet de poser un texte contre un bord droit ou de le centrer.
// Le minuteur de la première planche débordait pour avoir été placé à distance
// fixe du bord : un texte se mesure avant d'être posé.
//
// **Un caractère absent de la table occupe une cellule pleine.** Il pourrait
// n'occuper rien, mais un texte auquel il manque une lettre se lit comme un
// texte, alors qu'un trou se voit. Le défaut doit se remarquer là où il se
// produit, pas se dissimuler dans une chaîne plausible.
func (f *Font) Advance(s string) int {
	total := 0
	for _, r := range s {
		if i, ok := f.place[r]; ok {
			total += f.advance[i]
			continue
		}
		total += f.Cell[0]
	}
	return total
}

// Height rend la hauteur d'une ligne, en pixels de tampon.
func (f *Font) Height() int { return f.Cell[1] }

// rawInterface est le manifeste tel qu'il s'écrit.
type rawInterface struct {
	manifest.Commentable
	// Format est la version du format de manifeste.
	Format int `json:"version_format"`
	// Interface porte les entrées, dont la police est la première.
	Interface rawEntries `json:"interface"`
}

// rawEntries regroupe ce que le manifeste d'interface déclare.
type rawEntries struct {
	manifest.Commentable
	Font     rawFont     `json:"police"`
	Icon     rawIcon     `json:"icone"`
	Settings rawSettings `json:"reglages"`
}

// rawFont porte les mesures de la planche de glyphes.
//
// La chaîne des glyphes est la déclaration de leur ordre, et les avances lui
// sont parallèles : un décalage se lit sur les longueurs, là où un dictionnaire
// l'aurait laissé sans trace.
type rawFont struct {
	manifest.Commentable

	Sheet    string `json:"fichier"`
	Source   string `json:"source"`
	Cell     [2]int `json:"cellule"`
	Baseline int    `json:"ligne_de_base"`
	Native   int    `json:"taille_native"`
	Glyphs   string `json:"glyphes"`
	Advances []int  `json:"avances"`
}
