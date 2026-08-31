// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package game

import (
	"testing"

	"github.com/sprimault/cohue"
)

// mondeDEssai monte une partie sur les profils livrés et une carte à obstacles.
//
// Les profils viennent du manifeste publié et non d'une table forgée : c'est ce
// qui fait de ce fichier une épreuve du montage entier, et pas seulement de la
// boucle. Un profil dont la vitesse cesserait d'être convertie casse ici.
func mondeDEssai(t *testing.T, largeur, hauteur int) (*World, *Profiles) {
	t.Helper()
	profils, err := LoadProfiles(cohue.Assets, "assets/personnages/manifeste.json")
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}

	g := NewCostGrid(largeur, hauteur)
	for u := range largeur {
		g.Set(u, 0, Blocked)
		g.Set(u, hauteur-1, Blocked)
	}
	for v := range hauteur {
		g.Set(0, v, Blocked)
		g.Set(largeur-1, v, Blocked)
	}
	// Trois piliers et une cloison percée : de quoi obliger au contournement
	// plutôt qu'à la ligne droite.
	for v := 8; v < hauteur-8; v++ {
		if v != hauteur/2 {
			g.Set(largeur/2, v, Blocked)
		}
	}
	for _, p := range [][2]int{{10, 10}, {12, 30}, {40, 18}} {
		g.Set(p[0], p[1], Blocked)
		g.Set(p[0]+1, p[1], Blocked)
	}

	armes, err := LoadWeapons(cohue.Assets, manifesteArmes)
	if err != nil {
		t.Fatalf("armes livrées : %v", err)
	}
	return NewWorld(profils, armes.Base, g, 300, 256), profils
}

// indexDuProfil rend la place d'un profil dans la table, ou arrête le test.
func indexDuProfil(t *testing.T, profils *Profiles, cle string) int {
	t.Helper()
	for i := range profils.Enemies {
		if profils.Enemies[i].Key == cle {
			return i
		}
	}
	t.Fatalf("« %s » absent de la table", cle)
	return 0
}

// peupler pose des créatures sur toutes les cases franchissables d'une carte,
// jusqu'à en avoir posé le compte demandé.
func peupler(t *testing.T, w *World, profil, combien int) {
	t.Helper()
	poses := 0
	for v := 1; v < 63 && poses < combien; v++ {
		for u := 1; u < 63 && poses < combien; u++ {
			if !w.grille.Passable(u, v) {
				continue
			}
			if _, ok := w.SpawnEnemy(profil, FromInt(u)+One/2, FromInt(v)+One/2); !ok {
				t.Fatalf("bassin plein après %d créatures", poses)
			}
			poses++
		}
	}
	if poses != combien {
		t.Fatalf("%d créatures posées, %d demandées", poses, combien)
	}
}

// distanceMoyenne rend l'éloignement moyen des créatures au joueur, en tuiles.
func distanceMoyenne(w *World) float64 {
	px, py := w.Player()
	var somme float64
	for i := range w.Enemies().Active() {
		e := w.Enemies().At(i)
		somme += (Vec{e.X - px, e.Y - py}).Len().Float()
	}
	return somme / float64(w.Enemies().Len())
}

// TestTroisCentsPoursuivantsConvergent est le critère de livraison de l'étape 1.
//
// Trois cents créatures, une cible qui se déplace, des obstacles à contourner.
// Ce n'est pas un test de valeurs mais d'une propriété : la horde se rapproche,
// et aucune de ses créatures n'a traversé un mur pour le faire.
func TestTroisCentsPoursuivantsConvergent(t *testing.T) {
	w, profils := mondeDEssai(t, 64, 64)
	marcheur := indexDuProfil(t, profils, "marcheur")

	w.Place(FromInt(32)+One/2, FromInt(32)+One/2)
	peupler(t, w, marcheur, 300)

	depart := distanceMoyenne(w)

	// Dix secondes, la cible dérivant vers le haut à gauche pour qu'aucune
	// créature ne l'atteigne par une simple ligne droite.
	const ticks = 10 * TPS
	for range ticks {
		w.Step(Vec{-One, -One})
	}

	arrivee := distanceMoyenne(w)
	if arrivee >= depart {
		t.Errorf("distance moyenne %0.2f au départ, %0.2f après %d ticks : la horde ne converge pas",
			depart, arrivee, ticks)
	}
	// La convergence doit être franche : une horde qui grignote un dixième de
	// tuile passerait le test précédent en ne poursuivant rien.
	if arrivee > depart/2 {
		t.Errorf("distance moyenne passée de %0.2f à %0.2f seulement", depart, arrivee)
	}

	for i := range w.Enemies().Active() {
		e := w.Enemies().At(i)
		if !w.passable(e.X, e.Y) {
			t.Fatalf("la créature %d est dans un mur, en (%0.2f, %0.2f)",
				i, e.X.Float(), e.Y.Float())
		}
	}
	t.Logf("distance moyenne : %0.2f -> %0.2f tuiles en %d ticks", depart, arrivee, ticks)
}

// TestLaBoucleNalloueRien est l'autre moitié du critère de l'étape.
//
// La mesure porte sur des ticks entiers et non sur la seule mise à jour des
// entités : le champ de flux, la grille de densité et les seaux ont chacun leurs
// tableaux, et c'est leur réutilisation d'un tick à l'autre qui est en jeu. Une
// allocation par tick ne se verrait pas sur une passe isolée.
//
// Le champ ne se reconstruit qu'un tick sur `flowPeriod`, donc la mesure couvre
// forcément des ticks avec et sans reconstruction — c'est voulu : c'est
// justement la reconstruction qui est le candidat le plus probable.
//
// **Ce n'est pas un doublon de `TestLeChampNalloueRien`**, malgré l'apparence.
// Celui-là garde une propriété du champ pris seul ; celui-ci garde qu'aucun
// assemblage n'alloue — un tri, une tranche d'indices, un tampon partagé entre
// deux sous-systèmes qui vont bien chacun de leur côté. Le second peut tomber
// alors que le premier passe, et c'est le second qu'on serait tenté de
// supprimer, parce qu'il est le plus lent à exécuter.
func TestLaBoucleNalloueRien(t *testing.T) {
	w, profils := mondeDEssai(t, 64, 64)
	w.Place(FromInt(32)+One/2, FromInt(32)+One/2)
	peupler(t, w, indexDuProfil(t, profils, "marcheur"), 300)

	// Quelques ticks de préchauffage : les seaux atteignent leur capacité au
	// premier parcours, et c'est une allocation qu'on ne veut pas compter.
	for range 3 * flowPeriod {
		w.Step(Vec{One, 0})
	}

	moyenne := testing.AllocsPerRun(1000, func() {
		w.Step(Vec{One, 0})
	})
	if moyenne != 0 {
		t.Errorf("%v allocation(s) par tick à 300 entités, attendu aucune", moyenne)
	}
}

// TestRienNeTraverseUnMur éprouve la projection sur la passabilité.
//
// Une créature poussée droit dans une cloison perd sa composante bloquée et
// longe la paroi : c'est ce qui permet à la horde d'entasser un bloqueur contre
// un mur sans le faire passer au travers.
func TestRienNeTraverseUnMur(t *testing.T) {
	profils, err := LoadProfiles(cohue.Assets, "assets/personnages/manifeste.json")
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}
	// Un couloir d'une case de haut, la cible derrière le mur du fond.
	g := NewCostGrid(6, 3)
	for u := range 6 {
		g.Set(u, 0, Blocked)
		g.Set(u, 2, Blocked)
	}
	g.Set(5, 1, Blocked)

	// Arme inerte : ce test isole le déplacement, et un joueur qui abat la
	// créature dont on suit la trajectoire ne mesurerait plus rien.
	w := NewWorld(profils, Weapon{}, g, 4, 1)
	w.Place(FromInt(4)+One/2, FromInt(1)+One/2)
	if _, ok := w.SpawnEnemy(indexDuProfil(t, profils, "marcheur"), One/2+One, One/2+One); !ok {
		t.Fatal("créature refusée")
	}

	for range 5 * TPS {
		w.Step(Vec{})
		e := w.Enemies().At(0)
		if !w.passable(e.X, e.Y) {
			t.Fatalf("la créature est entrée dans un mur, en (%0.2f, %0.2f)", e.X.Float(), e.Y.Float())
		}
	}
}

// TestLeCoutDeLaCaseDiviseLaVitesse vérifie que le prix du terrain se paie au
// déplacement et pas seulement au calcul du chemin.
//
// Sans cela, le parcours pondéré serait une superstition : la horde
// contournerait la flaque et la traverserait à la même vitesse que le sol.
func TestLeCoutDeLaCaseDiviseLaVitesse(t *testing.T) {
	profils, err := LoadProfiles(cohue.Assets, "assets/personnages/manifeste.json")
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}

	// Deux couloirs identiques, l'un ordinaire et l'autre entièrement en flaque.
	parcours := func(cout Cost) Fixed {
		g := NewCostGrid(12, 3)
		for u := range 12 {
			g.Set(u, 0, Blocked)
			g.Set(u, 2, Blocked)
			g.Set(u, 1, cout)
		}
		w := NewWorld(profils, Weapon{}, g, 1, 1)
		w.Place(FromInt(1)+One/2, FromInt(1)+One/2)
		depart := FromInt(10) + One/2
		if _, ok := w.SpawnEnemy(indexDuProfil(t, profils, "marcheur"), depart, FromInt(1)+One/2); !ok {
			t.Fatal("créature refusée")
		}
		for range TPS {
			w.Step(Vec{})
		}
		return depart - w.Enemies().At(0).X
	}

	ordinaire := parcours(Free)
	flaque := parcours(3)
	if flaque >= ordinaire {
		t.Errorf("parcouru %0.2f tuile(s) dans la flaque contre %0.2f sur le sol",
			flaque.Float(), ordinaire.Float())
	}
	t.Logf("en une seconde : %0.2f tuiles sur le sol, %0.2f dans la flaque",
		ordinaire.Float(), flaque.Float())
}
