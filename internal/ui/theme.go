// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// L'apparence de l'interface telle que son manifeste la règle : épaisseurs,
// marges et teintes. Aucune dimension calculée — ce qui dépend d'un contenu se
// mesure au rendu.

package ui

import (
	"fmt"
	"image/color"
	"io/fs"
	"maps"
	"slices"

	"github.com/sprimault/cohue/internal/manifest"
)

// Teintes que le manifeste doit porter, et rien d'autre.
//
// La table est close, et c'est elle qui décide : une teinte absente fait échouer
// le chargement, une teinte inconnue aussi. Sans le second refus, une clé mal
// orthographiée s'ajouterait au manifeste sans que rien ne la lise, et la
// couleur qu'on croyait avoir réglée resterait celle d'avant.
var teintesRequises = []string{
	"cadre_fond", "cadre_bord", "bandeau_fond",
	"jauge_fond", "jauge_vie", "jauge_experience",
	"texte", "texte_attenue", "texte_valeur", "texte_contour",
}

// Theme porte ce que l'interface règle.
//
// Il ne porte aucune largeur ni hauteur d'élément : une carte tient à son texte
// et une case à son icône, si bien qu'une dimension déclarée ici serait une
// seconde description de ce que le contenu impose.
type Theme struct {
	// Border est l'épaisseur d'un bord de cadre, en pixels de tampon.
	Border int
	// Margin est la marge entre un bord et ce qu'il entoure.
	Margin int
	// GaugeHeight est l'épaisseur d'une jauge.
	GaugeHeight int

	teintes map[string]color.RGBA
}

// LoadTheme lit les réglages d'apparence du manifeste d'interface.
func LoadTheme(fsys fs.FS, chemin string) (*Theme, error) {
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

	r := brut.Interface.Settings
	if r.Border <= 0 {
		dire("réglages : bord de %d px", r.Border)
	}
	if r.Margin < 0 {
		dire("réglages : marge de %d px", r.Margin)
	}
	if r.GaugeHeight <= 0 {
		dire("réglages : jauge haute de %d px", r.GaugeHeight)
	}

	teintes := make(map[string]color.RGBA, len(r.Colors))
	for _, nom := range teintesRequises {
		brute, presente := r.Colors[nom]
		if !presente {
			dire("réglages : teinte « %s » absente", nom)
			continue
		}
		teintes[nom] = color.RGBA{R: brute[0], G: brute[1], B: brute[2], A: brute[3]}
	}
	for _, nom := range slices.Sorted(maps.Keys(r.Colors)) {
		if !slices.Contains(teintesRequises, nom) {
			dire("réglages : teinte « %s » inconnue, que rien ne lira", nom)
		}
	}

	if len(manques) > 0 {
		return nil, &manifest.Invalid{Path: chemin, Missing: manques}
	}
	return &Theme{
		Border:      r.Border,
		Margin:      r.Margin,
		GaugeHeight: r.GaugeHeight,
		teintes:     teintes,
	}, nil
}

// Color rend une teinte du thème.
//
// Elle panique sur un nom inconnu, et c'est délibéré : les noms sont une table
// close vérifiée au chargement, donc un nom absent ici est une faute de frappe
// dans du code, pas une donnée douteuse. Rendre une couleur par défaut la ferait
// passer pour un réglage, et l'on chercherait dans le manifeste ce qui est dans
// le code.
func (t *Theme) Color(nom string) color.RGBA {
	teinte, connue := t.teintes[nom]
	if !connue {
		panic("teinte d'interface inconnue : " + nom)
	}
	return teinte
}

// rawSettings porte les réglages tels qu'ils s'écrivent.
type rawSettings struct {
	manifest.Commentable

	Border      int                 `json:"bord_px"`
	Margin      int                 `json:"marge_px"`
	GaugeHeight int                 `json:"hauteur_jauge_px"`
	Colors      map[string][4]uint8 `json:"teintes"`
}
