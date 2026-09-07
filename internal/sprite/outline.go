// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le bord d'une forme : ce qui reste d'une découpe quand on n'en garde que le
// pourtour, et laisse voir au travers.

package sprite

import (
	"image"
	"image/color"
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
func Outline(src image.Image) image.Image {
	boite := src.Bounds()
	largeur, hauteur := boite.Dx(), boite.Dy()
	bord := image.NewNRGBA(image.Rect(0, 0, largeur, hauteur))

	opaque := func(x, y int) bool {
		if x < 0 || y < 0 || x >= largeur || y >= hauteur {
			return false
		}
		return color.NRGBAModel.Convert(src.At(boite.Min.X+x, boite.Min.Y+y)).(color.NRGBA).A != 0
	}
	for y := range hauteur {
		for x := range largeur {
			if !opaque(x, y) {
				continue
			}
			for _, pas := range autour {
				if !opaque(x+pas[0], y+pas[1]) {
					bord.SetNRGBA(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
					break
				}
			}
		}
	}
	return bord
}
