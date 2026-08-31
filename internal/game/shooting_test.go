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

// champSansTir monte la même salle, mais avec une arme inerte : les tests qui
// éprouvent la mort ne doivent pas voir le joueur abattre leurs cobayes.
func champSansTir(t *testing.T) (*World, *Profiles) {
	t.Helper()
	profils, err := LoadProfiles(cohue.Assets, manifestePersonnages)
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}
	g := NewCostGrid(32, 32)
	w := NewWorld(profils, Weapon{}, g, 16, 64)
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

// TestUnMortCesseDEtreUneCibleSansQuitterLeBassin garde la règle que la
// conception fixe : l'entité morte reste en place jusqu'à la fin du tick, pour
// que les index tiennent, mais un projectile traité plus tard l'ignore.
//
// Le test force une résistance à zéro, ce que le système ne produit qu'au milieu
// d'un tick — c'est le seul moyen d'observer de l'extérieur un état qui, en
// marche normale, ne survit pas à la passe de nettoyage.
//
// Le test précédent vérifiait qu'aucun mort ne reste après un `Step` : c'est vrai
// des deux implémentations, donc il ne gardait rien. Celui-ci tombe si la garde
// disparaît.
func TestUnMortCesseDEtreUneCibleSansQuitterLeBassin(t *testing.T) {
	w, profils := champSansTir(t)
	px, py := w.Player()

	if _, ok := w.SpawnEnemy(indexDuProfil(t, profils, "marcheur"), px+FromInt(2), py); !ok {
		t.Fatal("créature refusée")
	}
	w.Enemies().At(0).Hits = 0

	// Le projectile est posé un pas en arrière de la créature et avance d'un pas :
	// il arrive donc sur elle au moment où `toucher` s'exécute. La poser à un
	// rayon d'elle ne suffirait pas — la créature se déplace plus tôt dans le
	// tick, et le contact se manquerait pour une raison géométrique, sans rapport
	// avec ce que le test annonce garder.
	e := w.Enemies().At(0)
	if _, ok := w.Shots().Spawn(Projectile{
		X: e.X - One/8, Y: e.Y,
		Step:      Vec{One / 8, 0},
		Remaining: FromInt(4),
		Hits:      1,
	}); !ok {
		t.Fatal("projectile refusé")
	}

	w.Step(Vec{})

	if n := w.Shots().Len(); n != 1 {
		t.Errorf("%d projectile(s) en vol, attendu 1 : le tir a été absorbé par une "+
			"créature qui n'était plus une cible", n)
	}
	if n := w.Enemies().Len(); n != 0 {
		t.Errorf("%d créature(s) restante(s) : le nettoyage de fin de tick n'a pas eu lieu", n)
	}
}

// TestLeNettoyageNeLaisseAucunMort vérifie que la passe de fin de tick réexamine
// la place qu'elle libère.
//
// C'est le seul endroit du paquet où la place libérée doit être réexaminée : une
// passe de mise à jour ferait avancer deux fois l'entité remontée, alors que
// celle-ci ne fait que filtrer. La sauter laisserait un mort jusqu'au tick
// suivant, et deux morts adjacents suffisent à le montrer.
func TestLeNettoyageNeLaisseAucunMort(t *testing.T) {
	w, profils := champSansTir(t)
	px, py := w.Player()
	marcheur := indexDuProfil(t, profils, "marcheur")

	for i := range 3 {
		if _, ok := w.SpawnEnemy(marcheur, px+FromInt(2+i), py); !ok {
			t.Fatal("créature refusée")
		}
	}
	// Les deux dernières places : la première retirée fait remonter la seconde
	// à l'endroit qu'on vient de vider.
	w.Enemies().At(1).Hits = 0
	w.Enemies().At(2).Hits = 0

	w.Step(Vec{})

	if n := w.Enemies().Len(); n != 1 {
		t.Errorf("%d créature(s) vivante(s), attendu 1", n)
	}
	for i := range w.Enemies().Active() {
		if h := w.Enemies().At(i).Hits; h <= 0 {
			t.Errorf("une créature à %d touche(s) a survécu au nettoyage", h)
		}
	}
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
