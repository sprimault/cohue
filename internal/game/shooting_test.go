// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas du tir : l'arme prête qui reste prête, le plus proche visé quel que
// soit l'ordre du bassin, la créature qui en sort à l'instant de sa mort, et le
// projectile qui disparaît au bout de sa portée.

package game

import (
	"testing"

	"github.com/sprimault/cohue"
)

// champDeTir monte une salle vide, le joueur au centre, avec l'arme livrée.
func champDeTir(t *testing.T) (*World, *Profiles) {
	t.Helper()
	profils, err := LoadProfiles(cohue.Assets, manifestePersonnages)
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}
	armes, err := LoadWeapons(cohue.Assets, manifesteArmes)
	if err != nil {
		t.Fatalf("armes livrées : %v", err)
	}

	g := NewCostGrid(32, 32)
	for u := range 32 {
		g.Set(u, 0, Blocked)
		g.Set(u, 31, Blocked)
	}
	for v := range 32 {
		g.Set(0, v, Blocked)
		g.Set(31, v, Blocked)
	}

	w := NewWorld(profils, armes.Base, g, 16, 64)
	w.Place(FromInt(16)+One/2, FromInt(16)+One/2)
	return w, profils
}

// TestSansCibleLArmeNeConsommeRienEprouve le cas limite que la conception nomme.
//
// Si la cadence se consommait à vide, le joueur qui sort d'un couloir désert
// tirerait sa première salve avec un retard fonction du temps passé sans rien à
// viser. Rien à l'écran ne l'expliquerait, et le comportement de l'arme
// dépendrait du passé récent.
func TestSansCibleLArmeNeConsommeRien(t *testing.T) {
	w, profils := champDeTir(t)

	// Un compte volontairement premier avec la cadence : cent ticks tombaient
	// pile sur un multiple de vingt-cinq, si bien qu'une arme qui se consomme à
	// vide se retrouvait prête au moment de la mesure — et le test passait par
	// coïncidence.
	for range 101 {
		w.Step(Vec{})
	}
	if n := w.Shots().Len(); n != 0 {
		t.Fatalf("%d projectile(s) en vol sans aucune cible", n)
	}

	// Une créature apparaît à portée : le tir doit partir au tick suivant, sans
	// attendre une cadence.
	marcheur := indexDuProfil(t, profils, "marcheur")
	px, py := w.Player()
	if _, ok := w.SpawnEnemy(marcheur, px+FromInt(3), py); !ok {
		t.Fatal("créature refusée")
	}

	w.Step(Vec{})
	if n := w.Shots().Len(); n != 1 {
		t.Errorf("%d projectile(s) au premier tick avec cible, attendu 1 : "+
			"l'arme n'était pas prête", n)
	}
}

// TestLeTirViseLePlusProche fixe le ciblage.
//
// La visée est omnidirectionnelle et le joueur ne choisit pas : c'est ce qui
// donnera son rôle au Secouriste à l'étape 4, dont le seul moyen de se
// débarrasser est d'aller vers lui. Un ciblage qui prendrait n'importe quelle
// créature à portée le désactiverait par avance, et rien ici ne le dirait.
func TestLeTirViseLePlusProche(t *testing.T) {
	// Les deux cibles sont dans des directions perpendiculaires, et non alignées
	// : alignées, un projectile visant la lointaine traverse la proche et la
	// touche au passage, si bien que le test passe quel que soit le ciblage.
	//
	// Les deux ordres de pose, pour la même raison prise ailleurs : un ciblage
	// qui retiendrait la dernière créature rencontrée à portée désignerait la
	// bonne dans l'un des deux cas.
	for _, cas := range []struct {
		quoi         string
		procheDAbord bool
	}{
		{"la proche posée en premier", true},
		{"la proche posée en second", false},
	} {
		t.Run(cas.quoi, func(t *testing.T) {
			w, profils := champDeTir(t)
			px, py := w.Player()
			marcheur := indexDuProfil(t, profils, "marcheur")

			poser := func(dx, dy Fixed) Handle {
				h, ok := w.SpawnEnemy(marcheur, px+dx, py+dy)
				if !ok {
					t.Fatal("créature refusée")
				}
				return h
			}

			var pres, loin Handle
			if cas.procheDAbord {
				pres = poser(FromInt(2), 0)
				loin = poser(0, FromInt(5))
			} else {
				loin = poser(0, FromInt(5))
				pres = poser(FromInt(2), 0)
			}

			for range 10 * TPS {
				w.Step(Vec{})
				switch {
				case !w.Enemies().Alive(pres) && !w.Enemies().Alive(loin):
					t.Fatal("les deux sont mortes dans le même tick, le cas ne dit rien")
				case !w.Enemies().Alive(loin):
					t.Fatal("la créature à cinq tuiles est morte avant celle à deux")
				case !w.Enemies().Alive(pres):
					return
				}
			}
			t.Fatal("aucune des deux n'est morte")
		})
	}
}

// TestLaCadenceEspaceLesTirs vérifie que l'arme ne tire pas à chaque tick une
// fois qu'elle a une cible.
func TestLaCadenceEspaceLesTirs(t *testing.T) {
	w, profils := champDeTir(t)
	px, py := w.Player()
	// Un Vigile : douze touches, il survivra à la mesure.
	if _, ok := w.SpawnEnemy(indexDuProfil(t, profils, "bloqueur"), px+FromInt(3), py); !ok {
		t.Fatal("créature refusée")
	}

	tirs := 0
	precedent := 0
	for range 3 * 24 {
		w.Step(Vec{})
		if n := w.Shots().Len(); n > precedent {
			tirs++
		}
		precedent = w.Shots().Len()
	}
	if tirs != 3 {
		t.Errorf("%d tir(s) en trois cadences, attendu 3", tirs)
	}
}

// TestLeTirTueEtLaCreatureQuitteLeBassin éprouve la mort comme état.
//
// La résistance du Badaud vaut trois touches de l'arme de base : trois tirs, et
// il n'est plus dans le bassin. Aucun drapeau, aucune liste de morts en attente.
func TestLeTirTueEtLaCreatureQuitteLeBassin(t *testing.T) {
	w, profils := champDeTir(t)
	px, py := w.Player()
	marcheur := indexDuProfil(t, profils, "marcheur")
	if _, ok := w.SpawnEnemy(marcheur, px+FromInt(2), py); !ok {
		t.Fatal("créature refusée")
	}
	depart := profils.Enemies[marcheur].Hits

	for range 10 * TPS {
		w.Step(Vec{})
		if w.Enemies().Len() == 0 {
			break
		}
	}
	if n := w.Enemies().Len(); n != 0 {
		t.Fatalf("%d créature(s) vivante(s) après dix secondes de tir sur %d touches",
			n, depart)
	}
}

// TestLaSuppressionEstImmediate dit lequel des deux chemins le code emprunte.
//
// Le résultat visible serait le même avec une suppression différée à la fin du
// tick, mais pas le chemin : là, une garde sur la résistance empêcherait un
// second projectile de toucher un mort resté dans le bassin. Ici la créature en
// sort à l'instant où sa résistance tombe, si bien que le ciblage du même tick
// ne peut pas la voir — et ce test tomberait si quelqu'un différait la
// suppression en croyant coller à l'ordre de mise à jour.
func TestLaSuppressionEstImmediate(t *testing.T) {
	w, profils := champDeTir(t)
	px, py := w.Player()
	// Un Secouriste, trois touches, posé au contact du joueur pour que le
	// projectile l'atteigne dès le tick de son tir.
	if _, ok := w.SpawnEnemy(indexDuProfil(t, profils, "soigneur"), px+One/4, py); !ok {
		t.Fatal("créature refusée")
	}

	vivantes := w.Enemies().Len()
	for range 10 * TPS {
		w.Step(Vec{})
		// À chaque tick, le bassin ne contient que des créatures vivantes : une
		// résistance nulle ou négative signifierait une mort en attente.
		for i := range w.Enemies().Active() {
			if h := w.Enemies().At(i).Hits; h <= 0 {
				t.Fatalf("une créature à %d touche(s) est encore dans le bassin", h)
			}
		}
		if w.Enemies().Len() < vivantes {
			return
		}
	}
	t.Fatal("la créature n'est jamais morte")
}

// TestLeProjectileMeurtAuBoutDeSaPortee vérifie la seconde cause de suppression.
//
// Même chemin que celle du projectile qui touche : deux causes, une seule
// suppression, sinon l'une des deux oublie un jour de libérer sa place — et le
// bassin se remplit de tirs qui n'existent plus.
func TestLeProjectileMeurtAuBoutDeSaPortee(t *testing.T) {
	w, profils := champDeTir(t)
	px, py := w.Player()
	// Une cible juste à portée, qu'on retire aussitôt le tir parti : le
	// projectile poursuit dans le vide.
	h, ok := w.SpawnEnemy(indexDuProfil(t, profils, "marcheur"), px+FromInt(5), py)
	if !ok {
		t.Fatal("créature refusée")
	}

	w.Step(Vec{})
	if w.Shots().Len() != 1 {
		t.Fatalf("%d projectile(s), attendu 1", w.Shots().Len())
	}
	w.Enemies().Remove(h)

	// La portée vaut six tuiles et le projectile en parcourt un cinquième par
	// tick : trente ticks suffisent largement.
	for range 60 {
		w.Step(Vec{})
	}
	if n := w.Shots().Len(); n != 0 {
		t.Errorf("%d projectile(s) encore en vol après leur portée", n)
	}
}

// TestLeTirNalloueRien garde le budget sur la boucle complète, tir compris.
func TestLeTirNalloueRien(t *testing.T) {
	w, profils := champDeTir(t)
	px, py := w.Player()
	// Des Vigiles, qui encaissent : la mesure ne doit pas vider le bassin.
	bloqueur := indexDuProfil(t, profils, "bloqueur")
	for i := range 16 {
		if _, ok := w.SpawnEnemy(bloqueur, px+FromInt(2+i%3), py+FromInt(i%4)); !ok {
			t.Fatal("créature refusée")
		}
	}
	for range 3 * flowPeriod {
		w.Step(Vec{})
	}

	moyenne := testing.AllocsPerRun(500, func() {
		w.Step(Vec{})
	})
	if moyenne != 0 {
		t.Errorf("%v allocation(s) par tick avec tir, attendu aucune", moyenne)
	}
}
