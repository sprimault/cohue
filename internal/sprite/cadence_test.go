// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce que la dérivation d'une image garde : un cycle qui boucle avance et
// revient, deux entités voisines ne marchent pas au pas, et un cycle qui
// s'achève finit avec l'état qui le porte.

package sprite

import (
	"testing"

	"github.com/sprimault/cohue/internal/game"
)

// marche est un cycle qui boucle : cinq images, six ticks chacune.
var marche = game.Cycle{Frames: 5, Duration: 6, Loop: true}

// annonce est un cycle qui ne boucle pas : trois images, cinq ticks chacune,
// soit dix ticks pour aller de la première à la dernière.
var annonce = game.Cycle{Frames: 3, Duration: 5}

// TestUnCycleQuiBoucleAvanceEtRevient épingle la cadence et le retour à zéro.
//
// **Elle n'a aucun autre gardien.** Le rendu qui l'appelle n'a pas de test par
// doctrine, et une cadence fausse d'un facteur deux ne se voit pas : une horde
// qui marche trop vite ressemble à une horde qui marche.
func TestUnCycleQuiBoucleAvanceEtRevient(t *testing.T) {
	for _, c := range []struct {
		tick   game.Tick
		attend int
	}{
		{0, 0}, {5, 0}, // la première image tient ses six ticks
		{6, 1}, {11, 1},
		{24, 4}, {29, 4}, // la dernière
		{30, 0}, // et le cycle reprend
	} {
		if got := Loop(marche, c.tick, 0); got != c.attend {
			t.Errorf("tick %d : image %d, attendu %d", c.tick, got, c.attend)
		}
	}
}

// TestDeuxEntitesVoisinesNeMarchentPasAuPas garde le décalage par identifiant.
//
// **Le défaut qu'il évite n'a pas de symptôme partiel** : sans lui, une horde
// entière lève le pied à la même image, ce qui se voit d'un coup d'œil et ne
// ressemble à rien de vivant. Deux identifiants consécutifs suffisent à le
// montrer, et c'est le cas réel — l'anneau d'apparition les distribue à la
// suite.
func TestDeuxEntitesVoisinesNeMarchentPasAuPas(t *testing.T) {
	const tick = 12
	premier := Loop(marche, tick, 0)
	for id := 1; id < marche.Frames; id++ {
		if got := Loop(marche, tick, id); got == premier {
			t.Errorf("l'identifiant %d rend la même image que le premier, %d", id, got)
		}
	}

	// Le tour complet ramène à la même image : c'est ce qui dit que le décalage
	// est un décalage et non une dérive.
	if got := Loop(marche, tick, marche.Frames); got != premier {
		t.Errorf("après un tour complet, image %d, attendu %d", got, premier)
	}
}

// TestUnCycleQuiSacheveFinitAvecSonEtat épingle les trois régimes du décompte.
//
// **Ancrer sur la fin n'est pas un choix mais une contrainte** : au premier tick
// d'un état, tout ce qu'on lit est le décompte entier, et rien n'y distingue un
// état long qui commence d'un état court. Les trois régimes sont donc ceux du
// rapport entre la durée de l'état et celle du cycle, et chacun a son
// comportement voulu.
func TestUnCycleQuiSacheveFinitAvecSonEtat(t *testing.T) {
	// Un état plus long que le cycle : la première image tient jusqu'à ce que le
	// cycle ait la place de se dérouler, puis il se déroule.
	for _, c := range []struct {
		reste  game.Tick
		attend int
	}{
		{30, 0}, {11, 0}, {10, 0}, // dix ticks ou plus : rien n'a commencé
		{9, 1}, {5, 1},
		{4, 2}, {0, 2}, // et l'animation s'achève avec l'état
	} {
		if got := Once(annonce, c.reste); got != c.attend {
			t.Errorf("reste %d : image %d, attendu %d", c.reste, got, c.attend)
		}
	}

	// Un état plus court que le cycle : l'animation entre en cours de route et se
	// termine quand même. C'est ce que le décompte signifie — un état écourté est
	// un état interrompu, dont on montre la fin plutôt que le début.
	if got := Once(annonce, 7); got != 1 {
		t.Errorf("un état de sept ticks ouvre sur l'image %d, attendu la deuxième", got)
	}
}

// TestUnCycleDuneSeuleImageNeSeCadencePas garde le cas que le manifeste livré
// emploie le plus.
//
// Cinq des neuf profils ont un `degat` d'une image et un `repos` d'une image :
// une dérivation qui rendrait autre chose que zéro y demanderait une bande qui
// n'existe pas, et le rendu ne dessinerait plus rien.
func TestUnCycleDuneSeuleImageNeSeCadencePas(t *testing.T) {
	fixe := game.Cycle{Frames: 1, Duration: 12}
	for _, tick := range []game.Tick{0, 1, 12, 999} {
		if got := Loop(fixe, tick, 7); got != 0 {
			t.Errorf("boucle au tick %d : image %d, attendu 0", tick, got)
		}
		if got := Once(fixe, tick); got != 0 {
			t.Errorf("achèvement à %d : image %d, attendu 0", tick, got)
		}
	}
}

// TestUneCadenceNulleNeDivisePas garde la dérivation d'un cycle bâti à la main.
//
// Le chargement refuse une durée sous le pas, donc aucun manifeste ne produit ce
// cas ; un cycle construit dans un test, si. Une division par zéro y ferait
// tomber la suite pour une raison qui n'a rien à voir avec ce qu'elle éprouve.
func TestUneCadenceNulleNeDivisePas(t *testing.T) {
	muet := game.Cycle{Frames: 4}
	if got := Loop(muet, 10, 3); got != 0 {
		t.Errorf("boucle sans cadence : image %d, attendu 0", got)
	}
	if got := Once(muet, 10); got != 0 {
		t.Errorf("achèvement sans cadence : image %d, attendu 0", got)
	}
}
