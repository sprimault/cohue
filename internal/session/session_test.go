// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le montage de la partie publiée, sans rien injecter : la chaîne entière depuis
// l'embed jusqu'au joueur posé.

package session

import (
	"testing"

	"github.com/sprimault/cohue"
	"github.com/sprimault/cohue/internal/game"
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

// TestLaRelanceNeConserveRienDeLaPartie fixe la règle du remontage.
//
// **Ce qui compte n'est pas que la partie se remonte, mais ce qu'elle emporte.**
// Une relance qui garderait la vie entamée, la horde en place ou le tick courant
// serait une reprise et non une nouvelle run, et le joueur ne le verrait
// qu'après coup — au moment où il meurt deux fois plus vite qu'il ne devrait.
//
// **L'absence de survivant est énumérée plutôt que supposée**, et c'est ce qui
// fait de ce test une garde : le premier élément de méta-progression devra le
// modifier pour entrer, donc le décider. Ce qui traverse aujourd'hui est ce que
// la partie n'a pas touché — les tables du manifeste et le lieu cuit —, et rien
// de cela n'est un état de jeu.
func TestLaRelanceNeConserveRienDeLaPartie(t *testing.T) {
	partie, err := Open(cohue.Assets, cohue.StartingLevel)
	if err != nil {
		t.Fatalf("montage de la partie livrée : %v", err)
	}

	monde, grille, tuile := partie.World, partie.Grid, partie.Tile
	vie := monde.Health()
	semis := monde.Enemies().Len()

	// Jouer assez pour que tout ait bougé : le tick avance, la horde converge,
	// et l'arme abat de quoi changer le compte des vivants.
	for range 3 * game.TPS {
		monde.Step(game.Vec{})
	}
	if monde.Tick() == 0 || monde.Enemies().Len() == semis {
		t.Fatal("la partie n'a pas assez avancé : la relance n'aurait rien à effacer")
	}

	partie.Restart()

	if partie.World == monde {
		t.Fatal("la relance rend le même monde : l'état précédent y survit en entier")
	}
	if got := partie.World.Health(); got != vie {
		t.Errorf("vie de %d après la relance, attendu %d", got, vie)
	}
	if got := partie.World.Tick(); got != 0 {
		t.Errorf("tick à %d après la relance, attendu 0", got)
	}
	if got := partie.World.Enemies().Len(); got != semis {
		t.Errorf("%d créatures après la relance, attendu le semis de %d", got, semis)
	}
	if got := partie.World.Shots().Len(); got != 0 {
		t.Errorf("%d projectiles encore en vol après la relance", got)
	}

	// Le lieu et sa taille de tuile traversent, parce que la partie ne les a pas
	// touchés : les recuire rendrait les mêmes octets pour le prix d'un
	// décodage complet.
	if partie.Grid != grille || partie.Tile != tuile {
		t.Error("la relance recharge le lieu, qu'aucune partie ne modifie")
	}
}
