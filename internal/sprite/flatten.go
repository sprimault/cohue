// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// L'aplatissement d'un dessin en sa seule forme : ce qui reste d'un sprite quand
// on ne garde que sa découpe.

package sprite

import (
	"image"
	"image/color"
)

// Flatten rend la forme d'une image, tous ses pixels opaques en blanc.
//
// **Un aplat ne se tire pas d'un sprite coloré par une teinte.** L'échelle de
// couleur du rendu multiplie, et multiplier ne remplace pas : aucune valeur ne
// transforme un dessin en une seule couleur. Ce qu'on peut faire est le
// contraire — partir du blanc, qui est l'élément neutre, et le teindre à la
// pose. D'où cette copie, faite une fois au montage.
//
// **Le blanc plutôt que la teinte visée**, parce que la même forme sert deux
// fois et pas dans la même couleur : le contour qui détache un personnage en
// permanence, et la silhouette qui le révèle quand quelque chose le recouvre.
// Une copie par teinte en ferait deux.
//
// **L'alpha est recopié tel quel et non binarisé** : le contour d'un sprite est
// déjà franc, le générateur y veille, et arrondir ici déciderait à sa place. Un
// dessin qui aurait des bords fondus donnerait une silhouette aux bords fondus,
// ce qui est la réponse juste — la forme est celle du dessin, pas une idée de sa
// forme.
func Flatten(src image.Image) image.Image {
	boite := src.Bounds()
	forme := image.NewNRGBA(image.Rect(0, 0, boite.Dx(), boite.Dy()))
	// Le passage par le modèle plutôt que par `RGBA` et un décalage : celui-ci
	// rend des composantes sur seize bits qu'il faudrait rétrécir, et une
	// conversion qu'aucun outil ne peut borner devient une alerte à taire. Le
	// modèle rend l'octet directement, et il dépremultiplie au passage — ce qui
	// est ce qu'on veut, la forme étant l'alpha et non ce que la couleur en a
	// déjà consommé.
	for y := range boite.Dy() {
		for x := range boite.Dx() {
			c := color.NRGBAModel.Convert(src.At(boite.Min.X+x, boite.Min.Y+y)).(color.NRGBA)
			if c.A == 0 {
				continue
			}
			forme.SetNRGBA(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: c.A})
		}
	}
	return forme
}
