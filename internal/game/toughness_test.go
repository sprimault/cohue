// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas du durcissement : l'arrondi au plus proche, la résistance figée à
// l'apparition, le refus d'une courbe qui adoucirait, et l'absence qui vaut un.

package game

import (
	"testing"

	"github.com/sprimault/cohue"
)

// TestLaResistanceSArrondit garde la règle d'arrondi, qui n'a qu'un domicile.
//
// **Aucune donnée livrée ne l'exerce**, et c'est pourquoi ce cas la force : le
// lieu de démonstration ne déclare aucun durcissement, si bien que le champ
// vaudrait un partout et que la multiplication rendrait toujours la base. Un
// mécanisme que rien n'exerce est un mécanisme qu'on croit écrit.
//
// Les valeurs sont choisies pour tomber des deux côtés du demi : 4,2 arrondit
// vers le bas, 4,5 vers le haut. Un cas qui ne verrait qu'un seul côté passerait
// sur une troncature.
func TestLaResistanceSArrondit(t *testing.T) {
	profils, err := LoadProfiles(cohue.Assets, manifestePersonnages)
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}
	badaud := &profils.Enemies[indexDuProfil(t, profils, "marcheur")]
	if badaud.Hits != 3 {
		t.Fatalf("le Badaud encaisse %d touches, ces chiffres en attendent trois",
			badaud.Hits)
	}

	for _, c := range []struct {
		durcissement float64
		veut         int
	}{
		{1.0, 3},
		{1.3, 4}, // 3,9 — vers le haut
		{1.4, 4}, // 4,2 — vers le bas
		{1.5, 5}, // 4,5 — le demi va vers le haut
		{2.0, 6}, // le doublement, qu'un multiplicateur entier imposerait
		{3.34, 10},
	} {
		if got := badaud.HitsAt(FromFloat(c.durcissement)); got != c.veut {
			t.Errorf("durcissement %v : %d touches, attendu %d",
				c.durcissement, got, c.veut)
		}
	}
}

// TestLaResistanceEstFigeeALApparition garde ce qui fait tenir l'unité.
//
// **« Trois touches » est une unité, pas un nombre**, et elle ne le reste que si
// une créature demande le même nombre de coups du premier au dernier. Une
// résistance qui suivrait la courbe en cours de vie ferait qu'un Badaud entamé
// avant un palier en réclamerait davantage après, sans que rien à l'écran ne
// l'explique.
//
// Le cas fait apparaître sous un palier tendre, puis passe au palier dur : la
// créature née avant doit garder sa résistance, et celle qui naît après prendre
// la nouvelle.
func TestLaResistanceEstFigeeALApparition(t *testing.T) {
	w, profils := salleOuverte(t, nil, 16)
	marcheur := indexDuProfil(t, profils, "marcheur")
	base := profils.Enemies[marcheur].Hits

	// **Une pression lente, et c'est nécessaire.** À forte pression le bassin se
	// remplit avant le second palier : toutes les créatures naissent sous le
	// premier, et le cas relève alors une résistance qu'il croit neuve. Trois par
	// seconde pour un Badaud à trois en fait un par seconde, ce qui laisse de la
	// place à celle qui naîtra durcie.
	w.scenario = &Scenario{Phases: []Phase{
		{Start: 0, Pressure: parTick(3), Profiles: []int{marcheur}, Toughness: One},
		{Start: 2 * TPS, Pressure: parTick(3), Profiles: []int{marcheur},
			Toughness: FromFloat(2.0)},
	}}

	// L'attente porte sur l'apparition et non sur un compte de ticks : le budget
	// met une seconde à payer un Badaud, et un relevé pris trop tôt trouverait la
	// salle vide.
	for w.tick < 2*TPS && w.Enemies().Len() == 0 {
		w.Step(Vec{})
	}
	if w.Enemies().Len() == 0 {
		t.Fatal("aucune créature au premier palier")
	}
	ancienne := w.Enemies().At(0)
	if ancienne.Hits != base {
		t.Fatalf("résistance %d au palier tendre, attendu %d", ancienne.Hits, base)
	}

	// **Le second palier se laisse franchement passer avant de compter.** Le tick
	// s'incrémente après les apparitions, si bien qu'une créature relevée au tick
	// où le palier commence est née sous le précédent : `avant` se prend donc une
	// fois la frontière derrière soi, et tout ce qui naît ensuite est durci.
	for w.tick <= 2*TPS+1 {
		w.Step(Vec{})
	}
	avant := w.Enemies().Len()
	for w.tick < 8*TPS && w.Enemies().Len() == avant {
		w.Step(Vec{})
	}
	if ancienne.Hits != base {
		t.Errorf("la créature née au premier palier est passée à %d touches : la "+
			"résistance suit la courbe au lieu d'être figée", ancienne.Hits)
	}
	if w.Enemies().Len() <= avant {
		t.Fatal("aucune créature neuve au second palier")
	}
	if neuve := w.Enemies().At(w.Enemies().Len() - 1); neuve.Hits != 2*base {
		t.Errorf("la créature née au second palier a %d touches, attendu %d",
			neuve.Hits, 2*base)
	}
}

// TestUneCourbeQuiAdoucitEstRefusee garde ce que le champ signifie.
//
// Un profil d'une touche sous un demi rendrait zéro, c'est-à-dire une créature
// morte à l'instant où elle apparaît — et le refus au chargement est plus simple
// qu'un plancher, qui laisserait le fichier dire une chose et le jeu en faire une
// autre.
func TestUneCourbeQuiAdoucitEstRefusee(t *testing.T) {
	mou := 0.5
	_, manques := CompileScenario(WaveScenario{Phases: []WavePhase{
		func() WavePhase {
			p := phaseEcrite("0:00", 8, "marcheur")
			p.Toughness = &mou
			return p
		}(),
	}}, profilsLivres(t), reportDeTest)

	if !contient(manques, "durcit et n'adoucit pas") {
		t.Errorf("une courbe qui adoucit passe : %v", manques)
	}
}

// TestUnDurcissementAbsentVautUn vérifie que le champ facultatif l'est vraiment.
//
// C'est ce qui permet au lieu livré de ne rien déclarer et de garder exactement
// la résistance de sa table — donc à ce lot de ne rien changer à ce qui existe.
func TestUnDurcissementAbsentVautUn(t *testing.T) {
	scenario, manques := CompileScenario(WaveScenario{Phases: []WavePhase{
		phaseEcrite("0:00", 8, "marcheur"),
	}}, profilsLivres(t), reportDeTest)
	if len(manques) > 0 {
		t.Fatalf("phase refusée : %v", manques)
	}
	if got := scenario.Phases[0].Toughness; got != One {
		t.Errorf("durcissement %v pour un champ absent, attendu %v", got, One)
	}
}
