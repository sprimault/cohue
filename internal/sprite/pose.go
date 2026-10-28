// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce qu'une image découpée laisse à qui la pose : de quoi remplir un quad, et
// la forme qui dit ce qu'elle recouvre vraiment.

package sprite

import "github.com/sprimault/lozengine/scene"

// Pose est une image prête à poser, et la forme de ce qu'elle peint.
//
// **Les deux voyagent ensemble parce qu'elles décrivent la même case.** Un
// aspect dit où lire dans l'atlas et comment s'ancrer, un masque dit lesquels de
// ces pixels sont opaques ; les tenir dans deux tables indexées par le même
// numéro d'image serait deux descriptions d'une découpe, et c'est celle qu'on
// oublierait de remplir qui déciderait du recouvrement.
type Pose struct {
	// Aspect est de quoi dessiner la case. Ce qui n'y est pas rempli ne regarde
	// pas la découpe : la couche, l'étage, la révélation, et la teinte d'un état
	// comme l'éclair d'un impact.
	Aspect scene.Aspect
	// Mask est la forme des pixels opaques, nulle pour une case vide. C'est ce
	// qui décide si quelque chose cache vraiment ce qu'il chevauche, une boîte
	// de bande étant quatre fois plus large qu'un personnage.
	Mask *Mask
}

// poses rend les cases d'une planche, ancrées comme l'appelant le demande.
//
// Relevées une fois au chargement : ce qui les compose lit des pixels, et un
// relevé par image serait du travail dans la boucle.
func (p planche) poses(ancre [2]int16) []Pose {
	liste := make([]Pose, len(p.cases))
	for i := range p.cases {
		aspect, _ := p.aspect(i, ancre)
		liste[i] = Pose{Aspect: aspect, Mask: p.masque(i)}
	}
	return liste
}
