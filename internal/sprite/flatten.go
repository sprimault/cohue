// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// L'aplatissement d'une planche en sa seule forme : ce qui reste d'un dessin
// quand on ne garde que sa découpe.

package sprite

import (
	"slices"

	"github.com/sprimault/lozengine/raster"
	"github.com/sprimault/lozengine/render"
)

// Flatten rend la forme d'une planche, tous ses pixels opaques en blanc.
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
//
// **La découpe suit, parce qu'une forme se découpe comme son dessin.** Les deux
// planches entrent séparément dans l'atlas et un aspect y prend le même
// rectangle ; redécouper la dérivée à côté serait une seconde description de la
// même grille.
func Flatten(p raster.Sheet) raster.Sheet {
	forme := raster.NewSheet(p.Width, p.Height)
	forme.Rectangles = slices.Clone(p.Rectangles)
	// L'alpha d'une couleur prémultipliée est l'alpha du dessin : la
	// prémultiplication ne touche que les trois autres composantes, si bien que
	// la forme se lit sans dépremultiplier quoi que ce soit.
	for i, c := range p.Pixels {
		if c.A == 0 {
			continue
		}
		forme.Pixels[i] = render.Premultiply(255, 255, 255, c.A)
	}
	return forme
}
