// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le dessin du texte : la planche de glyphes chargée en image, découpée une fois
// pour toutes, et les deux façons de poser une chaîne — nue sur un fond connu,
// contourée sur le monde.

package render

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/png" // le décodeur que la planche de glyphes exige
	"io/fs"
	"path"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/sprimault/cohue/internal/ui"
)

// Font dessine du texte à partir de la planche de glyphes.
//
// Les mesures viennent de `internal/ui`, qui les décode et les vérifie sans
// écran ; ce type n'ajoute que l'image et le blit. Aucune métrique n'est écrite
// ici : le manifeste fait contrat, et changer de fonte ne touche pas ce fichier.
type Font struct {
	mesures *ui.Font
	// glyphes est la planche découpée une fois au chargement. Un `SubImage` par
	// caractère et par image ferait une allocation par lettre affichée, soixante
	// fois par seconde.
	glyphes []*ebiten.Image

	op ebiten.DrawImageOptions
}

// LoadFont lit le manifeste d'interface et charge la planche qu'il désigne.
//
// La planche est cherchée à côté du manifeste : c'est le manifeste qui la nomme,
// et le chemin du dossier ne se recompose pas ailleurs.
func LoadFont(fsys fs.FS, chemin string) (*Font, error) {
	mesures, err := ui.LoadFont(fsys, chemin)
	if err != nil {
		return nil, err
	}

	planche := path.Join(path.Dir(chemin), mesures.Sheet)
	brut, err := fs.ReadFile(fsys, planche)
	if err != nil {
		return nil, fmt.Errorf("lecture de %s: %w", planche, err)
	}

	source, _, err := image.Decode(bytes.NewReader(brut))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", planche, err)
	}
	atlas := ebiten.NewImageFromImage(source)

	largeur, hauteur := mesures.Cell[0], mesures.Cell[1]
	if atlas.Bounds().Dy() != hauteur {
		return nil, fmt.Errorf("%s: planche haute de %d px pour une cellule de %d",
			planche, atlas.Bounds().Dy(), hauteur)
	}

	colonnes := atlas.Bounds().Dx() / largeur
	glyphes := make([]*ebiten.Image, colonnes)
	for i := range glyphes {
		rect := image.Rect(i*largeur, 0, (i+1)*largeur, hauteur)
		glyphes[i] = atlas.SubImage(rect).(*ebiten.Image)
	}

	return &Font{mesures: mesures, glyphes: glyphes}, nil
}

// Advance rend la largeur d'une chaîne, en pixels de tampon.
//
// C'est par elle qu'on pose un texte contre un bord droit : un texte placé à
// distance fixe déborde dès qu'il s'allonge, et un minuteur déborde exactement
// quand la partie dure.
func (f *Font) Advance(s string) int { return f.mesures.Advance(s) }

// Height rend la hauteur d'une ligne, en pixels de tampon.
func (f *Font) Height() int { return f.mesures.Height() }

// Draw pose une chaîne, coin haut-gauche de la première cellule en (x, y).
//
// La teinte multiplie le blanc de la planche, qui ne porte que des glyphes
// opaques : le texte prend donc exactement la couleur demandée.
func (f *Font) Draw(dst *ebiten.Image, texte string, x, y int, teinte color.RGBA) {
	for _, r := range texte {
		place, connu := f.mesures.Place(r)
		if !connu {
			// Le trou que `ui.Font.Advance` compte : rien n'est dessiné, mais la
			// cellule est franchie, si bien qu'un caractère manquant se voit au
			// lieu de se fondre dans un texte plausible.
			x += f.mesures.Cell[0]
			continue
		}
		f.op = ebiten.DrawImageOptions{}
		f.op.GeoM.Translate(float64(x), float64(y))
		f.op.ColorScale.ScaleWithColor(teinte)
		dst.DrawImage(f.glyphes[place], &f.op)
		x += f.mesures.Advance(string(r))
	}
}

// DrawOutlined pose une chaîne avec un contour d'un pixel.
//
// C'est ce que la conception exige de tout texte posé **sur le monde** plutôt
// que sur un cadre : un chiffre de dégâts jaillit au-dessus de n'importe quoi —
// carrelage clair, flaque, créature sombre — et sa seule couleur ne suffit pas à
// le détacher. Le contour est en croix et non sur les huit voisins : à cette
// taille, les diagonales épaissiraient le glyphe au point de le fermer.
func (f *Font) DrawOutlined(dst *ebiten.Image, texte string, x, y int, teinte, contour color.RGBA) {
	for _, d := range [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
		f.Draw(dst, texte, x+d[0], y+d[1], contour)
	}
	f.Draw(dst, texte, x, y, teinte)
}
