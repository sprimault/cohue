// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// De quoi fabriquer un aspect d'interface : l'ancre qui pose un coin sur un
// pixel, l'aplat qui ne lit aucune planche, et les réductions vers les types du
// moteur.

package render

import (
	"math"

	"github.com/sprimault/lozengine/render"
	"github.com/sprimault/lozengine/scene"
)

// ancreDEcran réduit au type du moteur l'ancre qui pose un coin haut-gauche sur
// le pixel visé, et dit quand elle n'y tient pas.
//
// **Le refus plutôt que la saturation** : une ancre tronquée pose l'image
// ailleurs sans que rien ne le signale, ce qui est le pire des deux — le dessin
// reste visible et seulement décalé.
func ancreDEcran(x, y int) ([2]int16, bool) {
	if x < math.MinInt16 || x > math.MaxInt16 || y < math.MinInt16 || y > math.MaxInt16 {
		return [2]int16{}, false
	}
	return [2]int16{int16(x), int16(y)}, true
}

// pixelDEcran réduit une position d'image au type du moteur.
//
// La réduction est concentrée ici plutôt que justifiée à chaque pose : un pixel
// d'image reste de l'ordre du tampon de 960×540, à six ordres de grandeur de la
// borne du type, et un texte posé hors cadre ne change pas cet ordre.
func pixelDEcran(v int) int32 {
	// #nosec G115 -- un pixel d'image est de l'ordre du tampon, loin de la borne
	return int32(v)
}

// aplat rend l'aspect d'un rectangle uni, posé par son coin haut-gauche.
//
// **Il ne lit aucune planche**, et c'est ce que le drapeau du moteur apporte :
// la teinte se peint telle quelle, donc un cadre, une jauge ou un bandeau cesse
// d'être un pixel blanc qu'on étire. Les deux incidents du dépôt — l'éclair
// éteint par une multiplication, l'icône divisée par la teinte du bord laissée
// dans des options partagées — deviennent inexprimables : chaque quad porte sa
// propre teinte.
//
// Une taille nulle rend la valeur zéro, qui ne montre rien : c'est la façon dont
// la scène reconnaît déjà ce qu'elle n'a pas à dessiner, et une jauge vide ou un
// cadre sans épaisseur passent par là sans cas particulier.
//
// **Il quitte la mesure de recouvrement, parce qu'il cache sans que ce masquage
// soit une information de jeu.** Un aplat plein opaque — le bandeau, le voile de
// fin — compterait comme cachant ce qui passe dessous, donc il déclencherait le
// rappel de silhouette pour une chose que l'interface recouvre. Or le rappel se
// peint *sous* l'interface : il serait invisible et le relevé serait payé pour
// rien. Voir `reveler` pour l'autre usage de ce drapeau.
//
// **Il ne porte aucune ancre, et c'est le point à ne pas reperdre.** Une entité
// reçoit un point du sol, sur lequel le moteur centre le pied de son image ; un
// quad d'écran reçoit son **coin haut-gauche**. Un aplat d'interface ne se pose
// que par la seconde voie, donc il n'a rien à rattraper. L'ancre de
// `(largeur/2, hauteur)` qui était ici pour cela s'ajoutait au coin au lieu de
// le corriger : le bandeau partait d'une demi-largeur vers la droite et d'une
// hauteur vers le bas, c'est-à-dire hors de l'image, et seuls les textes — qui
// lisent une planche — restaient en place.
func aplat(largeur, hauteur int, teinte render.Color) scene.Aspect {
	if largeur <= 0 || hauteur <= 0 || largeur > math.MaxUint16 || hauteur > math.MaxUint16 {
		return scene.Aspect{}
	}
	return scene.Aspect{
		// #nosec G115 -- bornés juste au-dessus
		Width: uint16(largeur),
		// #nosec G115 -- bornés juste au-dessus
		Height:    uint16(hauteur),
		Tint:      teinte,
		Flags:     render.Flat,
		Layer:     coucheInterface,
		Uncounted: true,
	}
}
