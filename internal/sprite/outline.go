// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le bord d'une forme : ce qui reste d'une découpe quand on n'en garde que le
// pourtour, et laisse voir au travers.

package sprite

import (
	"slices"

	"github.com/sprimault/lozengine/raster"
	"github.com/sprimault/lozengine/render"
)

// Les huit voisins d'un pixel, ceux dont l'absence fait un bord.
//
// **Huit et non quatre**, pour la raison qui vaut au liseré du rendu de l'être
// aussi : un sprite isométrique n'a pas d'arête droite, ses bords descendent en
// escalier, et un voisinage en croix y déclare bord un pixel sur deux dans les
// diagonales.
var autour = [8][2]int{
	{-1, -1}, {0, -1}, {1, -1},
	{-1, 0}, {1, 0},
	{-1, 1}, {0, 1}, {1, 1},
}

// Outline rend le bord intérieur d'une forme, en blanc opaque.
//
// **C'est la silhouette de qui passe derrière un mur.** L'aplat plein le montre
// aussi, mais il efface le mur là où il se pose : le personnage s'y lit posé
// *sur* la maçonnerie plutôt que derrière elle, et sur l'enceinte d'un lieu — un
// mur avec le vide au-delà — cela se lit comme une sortie de carte. Le bord seul
// laisse voir ce qui recouvre à l'intérieur du tracé, ce qui rend la profondeur
// sans rien retirer au décor.
//
// **Intérieur et non extérieur, à la différence du liseré permanent.** Celui-ci
// est posé huit fois autour du sprite qui le recouvre ensuite, donc il déborde ;
// ici l'image est rendue seule et à sa taille, et un bord extérieur y perdrait
// ce qui dépasse dès qu'un dessin touche le cadre de sa case. Le tracé coïncide
// ainsi exactement avec le personnage.
//
// **Le seuil est celui de `Flatten`, et il ne peut pas en différer** : un pixel
// appartient à la forme dès que son alpha n'est pas nul, faute de quoi le bord
// rendu ici ne cernerait pas la forme que le masque a mesurée.
//
// **Les bandes voisines d'une planche ne se voient pas l'une l'autre.** Le bord
// se relève par case de la découpe et non sur la planche entière : deux images
// d'un cycle se touchent, et un pixel du bord droit de la première y perdrait
// son bord parce que la seconde commence là. Une planche sans découpe vaut pour
// elle-même, comme partout ailleurs dans le moteur.
func Outline(p raster.Sheet) raster.Sheet {
	bord := raster.NewSheet(p.Width, p.Height)
	bord.Rectangles = slices.Clone(p.Rectangles)

	for _, r := range decoupeOuEntier(p) {
		cerner(p, bord, r)
	}
	return bord
}

// cerner relève le bord des pixels opaques d'une case, sans regarder au-delà.
func cerner(src, dst raster.Sheet, r raster.Rectangle) {
	largeur, hauteur := int(r.Width), int(r.Height)
	x0, y0 := int(r.X), int(r.Y)
	pas := int(src.Width)

	opaque := func(x, y int) bool {
		if x < 0 || y < 0 || x >= largeur || y >= hauteur {
			return false
		}
		return src.Pixels[(y0+y)*pas+x0+x].A != 0
	}
	for y := range hauteur {
		for x := range largeur {
			if !opaque(x, y) {
				continue
			}
			for _, voisin := range autour {
				if !opaque(x+voisin[0], y+voisin[1]) {
					dst.Pixels[(y0+y)*pas+x0+x] = render.White()
					break
				}
			}
		}
	}
}

// decoupeOuEntier rend les cases d'une planche, ou la planche entière quand
// elle n'est pas découpée.
//
// Une planche vaut pour elle-même faute de découpe, et c'est la convention du
// moteur plutôt qu'un repli : une tuile unique n'a rien à trancher.
func decoupeOuEntier(p raster.Sheet) []raster.Rectangle {
	if len(p.Rectangles) > 0 {
		return p.Rectangles
	}
	return []raster.Rectangle{{Width: p.Width, Height: p.Height}}
}
