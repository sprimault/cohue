// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas des flammes : ce qu'une charge pose, ce qu'une traversée coûte à la
// horde, ce que le joueur n'y prend pas, et l'extinction au bout de la durée.

package game

import "testing"

// TestDesFlammesSePosentSousLeJoueur garde ce qui les sépare de la grenade.
//
// **Sans cible, comme la tourelle et pour la même raison** : ce qu'une flaque
// interdit est un passage, et on la pose avant que la horde n'arrive. Exiger une
// cible la mettrait là où la horde est déjà, c'est-à-dire trop tard.
func TestDesFlammesSePosentSousLeJoueur(t *testing.T) {
	w, _ := champDeTir(t)
	tomber(t, w, "lance_flammes")
	w.ramasserUneArme()
	px, py := w.Player()
	_, avant := w.HeldHeavy(0)

	w.Trigger(0)

	if n := w.Fires().Len(); n != 1 {
		t.Fatalf("%d flaque(s) posée(s) dans une salle vide, attendu 1", n)
	}
	if _, apres := w.HeldHeavy(0); apres != avant-1 {
		t.Errorf("%d charge(s) après la pose, attendu %d", apres, avant-1)
	}
	if f := w.Fires().At(0); f.X != px || f.Y != py {
		t.Errorf("posée en (%d, %d), attendu sous le joueur en (%d, %d)", f.X, f.Y, px, py)
	}
}

// TestUneTraverseeLenteCouteDavantageQuUneRapide garde ce qui décide du chiffre
// de cet effet.
//
// **Ce qui punit les gros profils n'est pas une règle mais un rapport** : celui
// de la cadence des impulsions à la vitesse de traversée. Deux créatures
// posées sur la même flaque, l'une laissée deux fois plus longtemps que
// l'autre, doivent avoir payé des montants distincts — sans quoi l'effet
// retirerait un forfait à l'entrée, ce qui est l'implémentation qu'on a écartée
// et que rien d'autre ne sépare de celle-ci.
//
// La durée courte est choisie hors d'un multiple de la cadence, faute de quoi le
// cas passerait par la coïncidence qui met le compteur à zéro au moment du
// relevé.
func TestUneTraverseeLenteCouteDavantageQuUneRapide(t *testing.T) {
	brule := func(ticks int) (paye int, cadence Tick) {
		w, profils := champDeTir(t)
		w.arme = Weapon{} // le socle abattrait le cobaye, et ses touches compteraient ici
		tomber(t, w, "lance_flammes")
		w.ramasserUneArme()
		arme, _ := w.HeldHeavy(0)
		w.Trigger(0)
		px, py := w.Player()

		// Sur le centre de la flaque : ce qu'on mesure est le temps passé dedans,
		// pas une géométrie.
		// Le Vigile, et pas un Quidam : douze touches lui permettent de tenir tout
		// le relevé long, quand trois le feraient mourir à la quatrième impulsion —
		// le cas mesurerait alors la survie du cobaye au lieu du coût du temps.
		if _, ok := w.SpawnEnemy(indexDuProfil(t, profils, "bloqueur"), px, py); !ok {
			t.Fatal("créature refusée")
		}
		avant := w.Enemies().At(0).Hits
		for range ticks {
			w.Step(Vec{})
			if w.Enemies().Len() == 0 {
				t.Fatal("la créature est morte avant la fin du relevé")
			}
		}
		return avant - w.Enemies().At(0).Hits, arme.Cooldown
	}

	_, cadence := brule(0)
	bref, long := int(cadence)+1, 3*int(cadence)+1
	court, _ := brule(bref)
	tenu, _ := brule(long)
	if tenu <= court {
		t.Errorf("%d touche(s) en %d ticks contre %d en %d : une traversée lente doit coûter plus",
			tenu, long, court, bref)
	}
}

// TestLesFlammesNAtteignentPasLeJoueur garde la règle du chapitre 9 sur l'effet
// qui l'exposerait le plus.
//
// **Il se tient dans sa propre flaque par construction**, puisqu'elle se pose
// sous lui : c'est le seul effet du jeu dont la zone contienne son auteur au
// moment où elle s'applique, et celui où l'oubli coûterait le plus cher.
func TestLesFlammesNAtteignentPasLeJoueur(t *testing.T) {
	w, _ := champDeTir(t)
	tomber(t, w, "lance_flammes")
	w.ramasserUneArme()
	arme, _ := w.HeldHeavy(0)
	w.Trigger(0)
	avant := w.Health()

	for range 2*int(arme.Cooldown) + 1 {
		w.Step(Vec{})
	}

	if apres := w.Health(); apres != avant {
		t.Errorf("le joueur a perdu %d point(s) de vie dans ses propres flammes", avant-apres)
	}
}

// TestUneFlaqueSEteintAuBoutDeSaDuree garde la fin de sa vie.
//
// Sans ce cas, une flaque qui ne se retirerait jamais brûlerait indéfiniment et
// tiendrait sa place : le passage interdit deviendrait un mur, et le plafond du
// bassin ferait refuser des poses pour des flaques éteintes.
func TestUneFlaqueSEteintAuBoutDeSaDuree(t *testing.T) {
	w, _ := champDeTir(t)
	tomber(t, w, "lance_flammes")
	w.ramasserUneArme()
	arme, _ := w.HeldHeavy(0)
	w.Trigger(0)

	for range int(arme.Duration) - 1 {
		w.Step(Vec{})
		if w.Fires().Len() == 0 {
			t.Fatalf("éteinte au tick %d, avant ses %d", w.Tick(), arme.Duration)
		}
	}
	w.Step(Vec{})

	if n := w.Fires().Len(); n != 0 {
		t.Errorf("%d flaque(s) au-delà de la durée déclarée", n)
	}
}

// TestUneFlaqueRefuseeNeCoutePasSaCharge garde ce que toutes les branches
// promettent.
//
// Une charge dépensée pour un effet qui n'est pas parti retirerait au joueur une
// des trois choses qu'il possède, pour une raison qu'aucun écran ne peut lui
// montrer.
func TestUneFlaqueRefuseeNeCoutePasSaCharge(t *testing.T) {
	w, _ := champDeTir(t)
	tomber(t, w, "lance_flammes")
	w.ramasserUneArme()
	arme, _ := w.HeldHeavy(0)

	// Le bassin est rempli à la main : atteindre son plafond par des poses
	// coûterait plus de charges que le stock n'en donne.
	for w.Fires().Len() < w.Fires().Cap() {
		if _, ok := w.flammes.Spawn(Fire{Weapon: rangDeLArme(t, w, "lance_flammes"),
			Life: arme.Duration}); !ok {
			t.Fatal("le bassin refuse avant d'être plein")
		}
	}
	_, avant := w.HeldHeavy(0)

	w.Trigger(0)

	if _, apres := w.HeldHeavy(0); apres != avant {
		t.Errorf("%d charge(s) après un refus, attendu %d", apres, avant)
	}
}
