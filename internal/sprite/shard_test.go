// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce que la course d'un éclat garde : elle part et retombe au sol, elle s'ouvre
// en gerbe, et huit rangs donnent huit directions.

package sprite

import (
	"testing"

	"github.com/sprimault/cohue/internal/game"
)

// vieDEssai est la durée d'une volée dans ces cas, choisie paire pour que la
// mi-vie tombe sur un tick.
const vieDEssai game.Tick = 24

// TestUnEclatPartDuSolEtYRetombe épingle la parabole.
//
// **Elle n'a aucun autre gardien**, le rendu qui l'appelle n'ayant pas de test.
// Une hauteur qui ne reviendrait pas à zéro laisserait les éclats en l'air à
// l'instant où ils disparaissent, ce qui se lit comme une coupure et non comme
// une chute.
func TestUnEclatPartDuSolEtYRetombe(t *testing.T) {
	_, depart := Shard(0, 0, vieDEssai)
	_, sommet := Shard(0, vieDEssai/2, vieDEssai)
	_, arrivee := Shard(0, vieDEssai, vieDEssai)

	if depart != 0 || arrivee != 0 {
		t.Errorf("hauteurs aux deux bouts : %d et %d, attendu zéro", depart, arrivee)
	}
	if sommet <= 0 {
		t.Fatalf("sommet à %d : l'éclat ne monte pas", sommet)
	}
	// Le sommet est au milieu : un quart avant doit être plus bas.
	if _, quart := Shard(0, vieDEssai/4, vieDEssai); quart >= sommet {
		t.Errorf("au quart de sa vie l'éclat est à %d, au sommet à %d", quart, sommet)
	}
}

// TestUnEclatSeloigneSansRevenir garde la course dans le plan.
//
// Linéaire et monotone : un éclat qui reviendrait vers son point d'émission se
// lirait comme aspiré, ce qui est l'inverse de ce qu'une gerbe montre.
func TestUnEclatSeloigneSansRevenir(t *testing.T) {
	var precedent int64
	for age := range vieDEssai + 1 {
		ecart, _ := Shard(0, age, vieDEssai)
		d := int64(ecart.X)*int64(ecart.X) + int64(ecart.Y)*int64(ecart.Y)
		if d < precedent {
			t.Fatalf("à l'âge %d l'éclat s'est rapproché : %d après %d", age, d, precedent)
		} else {
			precedent = d
		}
	}
	if precedent == 0 {
		t.Error("l'éclat n'a pas bougé de toute sa vie")
	}
}

// TestHuitRangsDonnentHuitDirections garde ce qui fait une gerbe.
//
// **Sans cela, huit éclats en font un.** Le pas de trois entre deux rangs est
// celui qui sépare déjà deux entités superposées ; employé ici, il ouvre la
// volée au lieu de l'envoyer en éventail serré. Deux rangs qui rendraient la
// même direction empileraient deux formes au même endroit, et la caisse aurait
// l'air d'en cracher six.
func TestHuitRangsDonnentHuitDirections(t *testing.T) {
	vues := map[game.Vec]bool{}
	for rang := range Shards {
		ecart, _ := Shard(rang, vieDEssai/2, vieDEssai)
		if vues[ecart] {
			t.Errorf("le rang %d reprend une direction déjà servie : %v", rang, ecart)
		}
		vues[ecart] = true
	}
	if len(vues) != Shards {
		t.Fatalf("%d directions pour %d rangs", len(vues), Shards)
	}
}

// TestUneVolieSansDureeNeBougePas garde le cas qu'un bassin vide produirait.
//
// Une division par zéro y ferait tomber la suite pour une raison sans rapport
// avec ce qu'elle éprouve, et un effet sans durée n'a de toute façon rien à
// montrer.
func TestUneVolieSansDureeNeBougePas(t *testing.T) {
	ecart, hauteur := Shard(3, 5, 0)
	if ecart != (game.Vec{}) || hauteur != 0 {
		t.Errorf("écart %v et hauteur %d, attendu l'immobilité", ecart, hauteur)
	}
}
