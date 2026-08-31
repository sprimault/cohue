// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La conversion entre les tuiles du monde et les pixels du lieu, dans les deux
// sens. Elle ignore la caméra et le tampon : ce qu'elle rend est un pixel du
// lieu, pas un pixel de l'écran.

package render

import "github.com/sprimault/cohue/internal/game"

// projection convertit entre le repère du monde et celui de l'image.
//
// Elle ne s'atteint que par la caméra, qui compose ses résultats avec son
// décalage et les arrondit. Deux conversions offertes côte à côte laisseraient
// choisir celle qui ne cadre ni n'arrondit, et le pixel art demande l'autre.
//
// Les demi-dimensions plutôt que la taille de tuile, parce que ce sont elles qui
// apparaissent dans les formules : une tuile de 64×32 avance de 32 pixels par
// tuile en abscisse et de 16 en ordonnée.
type projection struct {
	demiLargeur, demiHauteur float64
}

// nouvelleProjection monte la conversion sur une taille de tuile, telle que le
// manifeste de décor la porte.
func nouvelleProjection(tuile [2]int) projection {
	return projection{
		demiLargeur: float64(tuile[0]) / 2,
		demiHauteur: float64(tuile[1]) / 2,
	}
}

// pixel rend le point de l'image où tombe une position du monde.
//
// L'axe des X du monde descend vers le sud-est de l'écran et celui des Y vers le
// sud-ouest : le monde (0, 0) est donc le sommet du losange de la première case,
// et non son coin supérieur gauche. C'est la convention des pièces, où le nord
// est v = 0.
func (p projection) pixel(x, y game.Fixed) (float64, float64) {
	tx, ty := x.Float(), y.Float()
	return (tx - ty) * p.demiLargeur, (tx + ty) * p.demiHauteur
}

// tuile rend la position du monde qui tombe sur un point de l'image.
//
// C'est le seul calcul du rendu qui remonte vers le monde, et il ne sert qu'au
// cadrage : la zone que le tampon montre est un losange dans le monde, dont il
// faut l'englobant en cases pour savoir quoi dessiner. Sans appelant, elle
// n'aurait pas sa place ici — un paquet sans tests ne garde rien de ce que rien
// n'exerce.
func (p projection) tuile(ex, ey float64) (game.Fixed, game.Fixed) {
	demiDiff := ex / (2 * p.demiLargeur)
	demiSomme := ey / (2 * p.demiHauteur)
	return game.FromFloat(demiSomme + demiDiff), game.FromFloat(demiSomme - demiDiff)
}
