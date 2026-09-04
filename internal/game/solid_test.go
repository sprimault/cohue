// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas du corps solide : le Vigile qui arrête le joueur, la horde qu'il ne
// gêne pas, le corps qui cesse de bloquer en mourant, et le joueur recouvert qui
// peut toujours sortir.

package game

import (
	"testing"

	"github.com/sprimault/cohue"
)

// areneSolide monte une salle vide et rend le profil du Vigile.
func areneSolide(t *testing.T) (*World, *EnemyProfile) {
	t.Helper()
	profils, err := LoadProfiles(cohue.Assets, manifestePersonnages)
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}

	g := NewCostGrid(32, 32)
	w := NewWorld(profils, armesInertes(t), progressionLivree(t), sansVagues(), g,
		graineDeTest, capacitesDeTest)
	w.Place(FromInt(16)+One/2, FromInt(16)+One/2)

	p := &profils.Enemies[indexDuProfil(t, profils, "bloqueur")]
	if !p.Solid {
		t.Fatal("le Vigile ne bloque pas : ces cas n'ont plus de sujet")
	}
	return w, p
}

// TestLeVigileArreteLeJoueur vérifie la seule exception que la conception
// accorde à la traversée de la horde.
//
// Un corps qui bouche un couloir, et pas seulement une résistance qui met du
// temps à tomber : sans lui, le Vigile serait un ennemi lent de plus, et son
// rôle — les goulots — n'existerait pas.
//
// Le joueur pousse vers lui pendant une seconde, ce qui suffirait largement à le
// traverser : cinq tuiles par seconde contre un tiers de tuile d'écart.
func TestLeVigileArreteLeJoueur(t *testing.T) {
	w, _ := areneSolide(t)
	if _, ok := w.SpawnEnemy(indexDuProfil(t, w.profils, "bloqueur"),
		w.playerX+One, w.playerY); !ok {
		t.Fatal("bassin plein")
	}
	vigile := w.Enemies().At(0)
	depart := w.playerX

	for range TPS {
		w.Step(Vec{X: One})
	}
	if w.playerX >= vigile.X {
		t.Errorf("le joueur est en %v et le Vigile en %v : il l'a traversé",
			w.playerX, vigile.X)
	}
	if w.playerX <= depart {
		t.Error("le joueur n'a pas avancé du tout : le corps bloque de trop loin")
	}
}

// TestUnProfilSansCorpsNArretePas garde ce qui rend l'exception une exception.
//
// Toute la horde partage la passe de déplacement du joueur : si le champ ne
// fermait pas le mécanisme, une foule dense figerait le joueur sur place et la
// mort deviendrait illisible — ce que la conception refuse explicitement.
func TestUnProfilSansCorpsNArretePas(t *testing.T) {
	w, _ := areneSolide(t)
	marcheur := indexDuProfil(t, w.profils, "marcheur")
	if w.profils.Enemies[marcheur].Solid {
		t.Fatal("le Quidam bloque, ce cas suppose le contraire")
	}
	if _, ok := w.SpawnEnemy(marcheur, w.playerX+One, w.playerY); !ok {
		t.Fatal("bassin plein")
	}
	quidam := w.Enemies().At(0)

	for range TPS {
		w.Step(Vec{X: One})
	}
	if w.playerX <= quidam.X {
		t.Errorf("le joueur est resté en deçà du Quidam (%v contre %v) : un corps "+
			"qui ne bloque pas l'a pourtant arrêté", w.playerX, quidam.X)
	}
}

// TestUnCorpsCesseDeBloquerEnMourant garde ce qui empêche le blocage de devenir
// un piège.
//
// La conception y tient l'exception entière : un joueur coincé tire
// nécessairement sur ce qui le coince, puisque la visée prend le plus proche, et
// douze touches finissent par tomber. Si le corps restait solide après la mort,
// le temps du nettoyage suffirait à rendre la situation inexplicable.
func TestUnCorpsCesseDeBloquerEnMourant(t *testing.T) {
	w, _ := areneSolide(t)
	if _, ok := w.SpawnEnemy(indexDuProfil(t, w.profils, "bloqueur"),
		w.playerX+One, w.playerY); !ok {
		t.Fatal("bassin plein")
	}
	vigile := w.Enemies().At(0)

	if !w.dansUnCorps(vigile.X, vigile.Y) {
		t.Fatal("le centre du Vigile n'est pas dans son propre corps")
	}
	vigile.Hits = 0
	if w.dansUnCorps(vigile.X, vigile.Y) {
		t.Error("le corps bloque encore alors que la résistance est tombée")
	}
}

// TestUnJoueurRecouvertPeutSortir garde la règle de résolution : on empêche
// d'entrer dans un corps, jamais d'y être.
//
// **L'état qu'il force n'est plus atteignable en jouant, et c'est assumé.** La
// réciprocité de la solidité l'a fermé : ni le joueur ni un corps solide ne peut
// entrer dans l'autre, donc rien ne les superpose. Ce que ce cas garde n'est
// donc pas une situation de jeu mais la **totalité de la projection** — sa
// signature ne promet aucune position hors des corps, et une entrée sans réponse
// définie figerait le joueur jusqu'à la fin de la partie.
//
// C'est pour cela qu'il pose la superposition lui-même plutôt que d'attendre
// qu'une partie la produise : la propriété gardée est celle de la fonction, pas
// celle du monde monté.
func TestUnJoueurRecouvertPeutSortir(t *testing.T) {
	w, _ := areneSolide(t)
	if _, ok := w.SpawnEnemy(indexDuProfil(t, w.profils, "bloqueur"),
		w.playerX, w.playerY); !ok {
		t.Fatal("bassin plein")
	}
	if !w.dansUnCorps(w.playerX, w.playerY) {
		t.Fatal("le joueur n'est pas dans le corps : le cas ne pose plus sa question")
	}

	depart := w.playerX
	for range TPS / 2 {
		w.Step(Vec{X: One})
	}
	if w.playerX <= depart {
		t.Errorf("le joueur est resté en %v : recouvert, il ne peut plus bouger",
			w.playerX)
	}
}
