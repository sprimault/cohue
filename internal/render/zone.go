// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les rectangles de l'interface : ce qu'on pose, et ce qu'un clic peut
// atteindre. Une seule description par élément, parce que deux finiraient par
// diverger.

package render

// zone est un rectangle de l'interface, en pixels de la surface de dessin.
//
// **Le repère est celui de la surface et non celui de l'aire de jeu**, et c'est
// ce qui rend le clic direct : l'interface se pose par `Scene.PlaceScreen`, qui
// vise la surface, et `Input.Pointer` rend sa position dans le même repère — le
// moteur défait l'agrandissement entier et les bandes avant de rendre l'entrée.
// Il n'y a donc aucune conversion entre ce qu'on dessine et ce qu'on désigne.
type zone struct {
	x, y, largeur, hauteur int
}

// contient dit si un point de la surface tombe dans le rectangle.
//
// Bord gauche et bord haut compris, bords opposés exclus : c'est la convention
// d'un rectangle en pixels, et elle évite que deux cases voisines se disputent
// la colonne qui les sépare.
func (z zone) contient(x, y int) bool {
	return x >= z.x && x < z.x+z.largeur &&
		y >= z.y && y < z.y+z.hauteur
}

// vide dit que la zone ne désigne rien.
//
// C'est la valeur zéro qui le porte, et elle sert aux cases que la partie ne
// montre pas : un emplacement d'arme vide n'a pas de rectangle, donc rien à y
// cliquer. Le distinguer par un booléen à côté ferait deux choses à tenir
// d'accord là où une largeur nulle suffit.
func (z zone) vide() bool { return z.largeur == 0 || z.hauteur == 0 }

// pointeur est ce que la souris désigne à cette image.
//
// **Il se retient d'un pas à l'autre parce que la pose n'a pas les entrées.**
// `Screen.Step` les reçoit, `Screen.Describe` dessine, et le survol a besoin
// des deux : la position se note au pas, et la pose la relit. C'est le seul
// état que la souris ajoute à l'écran.
type pointeur struct {
	x, y   int
	dedans bool
}

// pointeurDe lit la position de la souris dans les entrées d'une image.
//
// La conversion de `int16` à `int` est sans danger et sans garde : le moteur
// rend un pixel de la surface, bornée par son propre `MaxBufferSide`.
func pointeurDe(in Input) pointeur {
	x, y, dedans := in.Pointer()
	return pointeur{x: int(x), y: int(y), dedans: dedans}
}

// survole dit si le pointeur tombe dans la zone, et faux s'il est hors fenêtre.
//
// Le second résultat d'`Input.Pointer` décide : sorti de la fenêtre, le moteur
// garde la dernière position plutôt que de la remettre à zéro — qui désignerait
// un coin, donc la première case du bandeau —, et c'est ce drapeau qui dit de
// ne plus s'y fier.
//
// **Une zone vide n'a pas de garde à part**, et c'est ce qui a fait retirer
// celui qui était écrit ici : `contient` la rejette déjà, ses bornes opposées
// étant exclues. Le garde en plus ne changeait rien et masquait le seul test qui
// éprouve cette borne — un test qu'aucune mutation ne fait tomber ne prouve
// rien.
func (p pointeur) survole(z zone) bool {
	return p.dedans && z.contient(p.x, p.y)
}
