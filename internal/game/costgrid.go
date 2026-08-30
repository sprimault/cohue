// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package game

import "math"

// Cost est le prix de traversée d'une case, en pas.
//
// Pas un booléen. Une caisse ne bloque pas et ne se franchit pas librement :
// elle coûte cher, ce qui est le ralentissement que le chapitre 7 décrit, et
// c'est ce qui fait tenir ensemble ses trois règles — on ne ralentit pas ce qui
// est arrêté, et un joueur acculé ne se dégage pas à travers ce qui bloque.
//
// Ce que la conception veut par ailleurs devient gratuit : la flaque, le sol
// sale et le sol fissuré ralentiront ce qui les traverse sans qu'aucun mécanisme
// nouveau soit écrit, et un profil pourra ignorer un coût que le joueur paie.
type Cost uint16

const (
	// Free est le coût d'une case ordinaire.
	Free Cost = 1

	// Blocked est le coût d'un mur : infini, en pratique.
	Blocked Cost = math.MaxUint16
)

// CostGrid est la carte telle que la simulation la voit : des coûts, et rien
// d'autre.
//
// Le moteur ne sait plus, à ce stade, que le lieu était fait de pièces. C'est
// tout l'intérêt de la cuisson au chargement : l'assemblage est un sujet du
// chargeur, le parcours en est un autre, et aucun des deux n'a besoin de
// connaître le vocabulaire de l'autre.
type CostGrid struct {
	largeur, hauteur int
	couts            []Cost
}

// NewCostGrid rend une grille dont toutes les cases sont franchissables.
//
// Une seule allocation, faite au chargement : la boucle de mise à jour n'en fait
// aucune, et la grille ne change de taille qu'entre deux lieux.
func NewCostGrid(largeur, hauteur int) *CostGrid {
	couts := make([]Cost, largeur*hauteur)
	for i := range couts {
		couts[i] = Free
	}
	return &CostGrid{largeur: largeur, hauteur: hauteur, couts: couts}
}

// Width rend la largeur en tuiles.
func (g *CostGrid) Width() int { return g.largeur }

// Height rend la hauteur en tuiles.
func (g *CostGrid) Height() int { return g.hauteur }

// InBounds dit si une case appartient à la grille.
func (g *CostGrid) InBounds(x, y int) bool {
	return x >= 0 && y >= 0 && x < g.largeur && y < g.hauteur
}

// At rend le coût d'une case, et Blocked hors de la grille.
//
// Hors bornes plutôt qu'une panique ou un booléen de sortie : le parcours du
// champ de flux sonde les quatre voisins de chaque case, dont ceux du bord, et
// devoir tester l'appartenance avant chaque lecture alourdirait le seul endroit
// du code qui compte vraiment. Un bord de carte se comporte comme un mur, ce
// qu'il est.
func (g *CostGrid) At(x, y int) Cost {
	if !g.InBounds(x, y) {
		return Blocked
	}
	return g.couts[y*g.largeur+x]
}

// Set fixe le coût d'une case, et ignore ce qui sort de la grille.
func (g *CostGrid) Set(x, y int, c Cost) {
	if !g.InBounds(x, y) {
		return
	}
	g.couts[y*g.largeur+x] = c
}

// Passable dit si une case se franchit, à quelque prix que ce soit.
func (g *CostGrid) Passable(x, y int) bool {
	return g.At(x, y) != Blocked
}
