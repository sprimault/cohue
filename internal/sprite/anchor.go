// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La conversion des deux conventions d'ancrage du dépôt vers celle d'un moteur
// qui centre le pied d'une image sur le point qu'elle vise.
//
// Deux fonctions parce que le dépôt pose de deux façons opposées : le décor et
// les objets ajoutent un décalage au point, les personnages retranchent un point
// d'appui. Les confondre inverse le signe.

package sprite

// AnchorFromOffset rend l'ancre d'une image que l'on pose en **ajoutant** un
// décalage au pixel du point visé : le décor et les objets.
//
// L'ancre est relative au **pied centré** — coin haut-gauche en `(-largeur/2,
// -hauteur)` du point visé —, qui est la convention d'un moteur de rendu
// isométrique et non celle du dépôt. La division est entière, comme celle du
// moteur : une largeur paire laisse une colonne de plus à gauche du point, une
// largeur impaire centre exactement, et le catalogue a les deux — 64 pour un mur,
// 61 pour un banc.
//
// **Elle ne convertit que la pose de l'image autour du point, jamais le point
// lui-même**, et c'est la moitié qu'on croira faite. Le dépôt vise le sommet nord
// du losange d'une case, un moteur qui compte en cellules vise son centre : seize
// pixels d'écart en ordonnée à la tuile 64×32, que personne n'absorbe ici.
// Confondre les deux conversions donne une scène décalée d'un quart de tuile en
// diagonale, et le décalage se cherche dans la projection, où il n'est pas.
//
// Il n'y a donc pas d'ancre nulle ni d'écart constant à attendre : le résultat
// dépend de la forme, et vaut 32 pour un sol comme pour un mur, 24 pour un
// pilier, 64 pour un immeuble.
//
// Le résultat reste en `int` : le paquet ne lie aucun moteur, et la réduction
// vers le type d'un champ de rendu appartient à qui compose l'image.
func AnchorFromOffset(decalage, taille [2]int) (x, y int) {
	return decalage[0] + taille[0]/2, decalage[1] + taille[1]
}

// AnchorFromFooting rend l'ancre d'une image que l'on pose en **retranchant** un
// point d'appui du pixel du point visé : les personnages.
//
// C'est AnchorFromOffset de l'appui opposé, et les deux existent plutôt qu'une
// seule parce que le signe est précisément ce qu'on se trompe à reporter : un
// appui de (32, 63) et un décalage de (32, 63) décrivent des images posées à cent
// vingt-six pixels l'une de l'autre, ce qu'aucune des deux valeurs ne trahit.
//
// Le dépôt pose bien de ces deux façons : `Tile.Offset` et `Prop.Offset`
// s'ajoutent au point — `x + f.dx` dans le dessin d'une case —, quand
// `game.Figure.Anchor` s'en retranche.
func AnchorFromFooting(appui, taille [2]int) (x, y int) {
	return AnchorFromOffset([2]int{-appui[0], -appui[1]}, taille)
}
