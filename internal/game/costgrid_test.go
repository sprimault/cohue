// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas de la grille de coûts : une grille neuve est franchissable,
// l'extérieur est un mur, une écriture hors bornes ne déborde pas, et un coût
// intermédiaire ralentit sans arrêter.

package game

import "testing"

// TestGrilleNeuveEstFranchissable vérifie qu'une grille naît sans obstacle.
func TestGrilleNeuveEstFranchissable(t *testing.T) {
	g := NewCostGrid(4, 3)
	if g.Width() != 4 || g.Height() != 3 {
		t.Fatalf("grille %dx%d, attendu 4x3", g.Width(), g.Height())
	}
	for y := 0; y < g.Height(); y++ {
		for x := 0; x < g.Width(); x++ {
			if g.At(x, y) != Free {
				t.Fatalf("case (%d,%d) coûte %d à la création", x, y, g.At(x, y))
			}
		}
	}
}

// TestHorsGrilleEstUnMur vérifie que lire au-delà du bord rend Blocked.
//
// C'est ce qui dispense le parcours du champ de flux de tester l'appartenance
// avant chaque lecture de voisin, dans la boucle la plus chaude du jeu. Sans
// cette règle, un oubli de test lirait la case d'en face à la ligne suivante —
// la grille étant un tableau à une dimension, le bord droit toucherait le bord
// gauche et les ennemis traverseraient la carte.
func TestHorsGrilleEstUnMur(t *testing.T) {
	g := NewCostGrid(4, 3)
	dehors := [][2]int{{-1, 0}, {0, -1}, {4, 0}, {0, 3}, {4, 3}, {-1, -1}}
	for _, c := range dehors {
		if g.At(c[0], c[1]) != Blocked {
			t.Errorf("(%d,%d) hors grille rend %d, attendu Blocked", c[0], c[1], g.At(c[0], c[1]))
		}
		if g.Passable(c[0], c[1]) {
			t.Errorf("(%d,%d) hors grille se dit franchissable", c[0], c[1])
		}
	}
}

// TestSetHorsGrilleNeDeborde vérifie qu'une écriture hors bornes est ignorée
// plutôt que d'écraser une autre case.
//
// L'index se calcule en y×largeur+x : sans la garde, poser un mur en (-1, 2)
// écrirait sur la case (3, 1) d'une grille de quatre de large, et l'obstacle
// apparaîtrait ailleurs qu'on ne l'a mis.
func TestSetHorsGrilleNeDeborde(t *testing.T) {
	g := NewCostGrid(4, 3)
	g.Set(-1, 2, Blocked)
	g.Set(4, 0, Blocked)
	for y := 0; y < g.Height(); y++ {
		for x := 0; x < g.Width(); x++ {
			if g.At(x, y) != Free {
				t.Fatalf("case (%d,%d) modifiée par une écriture hors grille", x, y)
			}
		}
	}
}

// TestCoutIntermediaire vérifie qu'une case peut coûter cher sans bloquer.
//
// C'est la raison d'être du type : une caisse se traverse en y perdant du
// terrain, elle n'arrête pas.
func TestCoutIntermediaire(t *testing.T) {
	g := NewCostGrid(2, 2)
	g.Set(1, 1, 20)
	if !g.Passable(1, 1) {
		t.Error("une case à 20 se dit infranchissable")
	}
	if g.At(1, 1) != 20 {
		t.Errorf("coût relu à %d au lieu de 20", g.At(1, 1))
	}
}
