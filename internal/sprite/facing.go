// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Vers quoi un personnage regarde : le passage d'un vecteur du monde au nom de
// bande que le manifeste emploie. C'est de la géométrie, et elle se garde ici.

package sprite

import "github.com/sprimault/cohue/internal/game"

// Les huit orientations, telles que le manifeste des personnages les nomme.
//
// **Ce sont des noms d'écran et non des axes du monde**, et c'est le générateur
// qui l'a décidé : il compose une figurine à partir d'un angle en degrés d'écran,
// si bien que « SE » désigne un personnage vu de trois quarts vers le bas à
// droite. Le monde, lui, a deux axes obliques — son `+x` descend vers le
// sud-est et son `+y` vers le sud-ouest. La correspondance entre les deux est
// une rotation d'un huitième de tour, et c'est tout ce que ce fichier contient.
//
// **Ils sont écrits ici et nulle part ailleurs**, comme les noms de cycles le
// sont dans le rendu. Le manifeste dit quelles bandes existent, ce fichier dit
// laquelle regarde où ; l'ordre du manifeste ne fait donc pas contrat, et un
// profil qui déclarerait ses directions autrement se dessinerait quand même.
const (
	Sud       = "S"
	SudOuest  = "SO"
	Ouest     = "O"
	NordOuest = "NO"
	Nord      = "N"
	NordEst   = "NE"
	Est       = "E"
	SudEst    = "SE"
)

// Les deux termes du rapport qui sépare une direction franche d'une diagonale.
//
// Les huit secteurs sont d'égale largeur, donc leurs frontières tombent à
// 22,5 degrés des axes, où le rapport des deux composantes vaut la tangente de
// cet angle — √2 − 1, soit 0,41421. Le rapport entier l'approche à un
// cent-millième, ce qui déplace une frontière de bien moins qu'un pixel : rien
// ne justifierait d'ouvrir la porte au flottant pour cela, et la comparaison
// tient dans un `int64` avec de la marge.
const (
	tangenteNum = 41421
	tangenteDen = 100000
)

// Facing rend le nom de bande vers lequel un vecteur du monde regarde, ou une
// chaîne vide pour un vecteur nul.
//
// **Le vecteur nul n'a pas de direction, et il ne s'en invente pas une.** Rendre
// le sud par défaut ferait pivoter un personnage à l'arrêt vers l'écran sans
// qu'on sache si c'est un choix ou l'absence de réponse ; c'est à l'appelant de
// décider ce que regarde ce qui ne va nulle part, parce que lui seul sait s'il
// reste une cible.
//
// Le passage au repère de l'écran se fait ici plutôt que chez l'appelant : les
// noms sont ceux de l'écran, donc la conversion appartient à ce qui les rend.
func Facing(v game.Vec) string {
	// L'abscisse d'écran croît vers la droite, l'ordonnée vers le bas : c'est la
	// projection isométrique, réduite à ce qui décide d'un signe.
	a, b := int64(v.X)-int64(v.Y), int64(v.X)+int64(v.Y)
	if a == 0 && b == 0 {
		return ""
	}

	horizontal, vertical := abs(a), abs(b)
	switch {
	case vertical*tangenteDen < horizontal*tangenteNum:
		if a > 0 {
			return Est
		}
		return Ouest
	case horizontal*tangenteDen < vertical*tangenteNum:
		if b > 0 {
			return Sud
		}
		return Nord
	case a > 0 && b > 0:
		return SudEst
	case a > 0:
		return NordEst
	case b > 0:
		return SudOuest
	}
	return NordOuest
}

// abs rend la valeur absolue d'un entier.
func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
