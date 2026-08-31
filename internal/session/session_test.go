// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le montage de la partie publiée, sans rien injecter : la chaîne entière depuis
// l'embed jusqu'au joueur posé.

package session

import (
	"testing"

	"github.com/sprimault/cohue"
)

// TestPartieLivreeSeMonte monte le jeu publié par le chemin qu'empruntent les
// deux binaires, et n'injecte rien.
//
// Il va plus loin que `TestLieuLivre`, dans `internal/level`, qui s'arrête à la
// grille de coûts : ni profils, ni armes, ni joueur n'y sont montés, si bien
// qu'un manifeste de personnages devenu illisible le laisserait au vert. Celui-ci
// tombe. À l'inverse, il ne dit rien de ce que la cuisson a mis dans chaque case,
// que l'autre relève une par une — supprimer l'un des deux laisse donc une moitié
// de la chaîne sans épreuve.
//
// Ce qu'il garde et que rien d'autre ne garde : **le joueur est posé sur une
// case franchissable**. `placer` le met au centre du lieu sans rien vérifier, et
// `World.Place` ne rattrape rien par principe — un point de départ dans un mur
// est un défaut du niveau. Le jour où le lieu livré changera de dessin, c'est ce
// test qui le dira plutôt qu'une partie où l'on ne peut pas bouger.
func TestPartieLivreeSeMonte(t *testing.T) {
	partie, err := Open(cohue.Assets, cohue.StartingLevel)
	if err != nil {
		t.Fatalf("montage de la partie livrée : %v", err)
	}

	// La horde est semée au montage, et le lieu livré est assez ouvert pour en
	// porter. Zéro créature signifierait que le semis ne trouve aucune case, ce
	// qu'un changement de pas ou d'écart au joueur produirait en silence.
	if n := partie.World.Enemies().Len(); n == 0 {
		t.Error("aucune créature semée sur le lieu livré")
	}

	if partie.Tile != [2]int{64, 32} {
		t.Errorf("tuile %v, attendu [64 32] — la taille du manifeste ne voyage pas",
			partie.Tile)
	}

	x, y := partie.World.Player()
	u, v := x.Floor(), y.Floor()
	if !partie.Grid.Passable(u, v) {
		t.Errorf("le joueur est posé en (%d, %d), qui ne se franchit pas", u, v)
	}
}
