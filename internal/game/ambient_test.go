// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas du figurant : il n'est pas une cible, il ne pousse personne, il change
// de cap, et un peuplement mal écrit se refuse au chargement.

package game

import (
	"testing"

	"github.com/sprimault/cohue"
)

// arenePeuplee monte une salle avec l'arme du joueur active et des figurants
// posés autour de lui.
//
// **L'arme tire, contrairement à la plupart des arènes du paquet** : ce que le
// premier cas garde est qu'elle ne vise pas les figurants, et une arme inerte ne
// pourrait pas le dire.
func arenePeuplee(t *testing.T, figurants int) *World {
	t.Helper()
	profils, err := LoadProfiles(cohue.Assets, manifestePersonnages)
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}
	armes, err := LoadWeapons(cohue.Assets, manifesteArmes)
	if err != nil {
		t.Fatalf("armes livrées : %v", err)
	}
	if len(profils.Ambient) == 0 {
		t.Fatal("aucun profil d'ambiance au manifeste : ces cas n'ont plus de sujet")
	}

	g := NewCostGrid(32, 32)
	w := NewWorld(profils, armes, progressionLivree(t), sansVagues(), g,
		graineDeTest, capacitesDeTest)
	w.Place(FromInt(16)+One/2, FromInt(16)+One/2)
	for range figurants {
		w.SpawnAmbient(0, w.playerX+One, w.playerY)
	}
	return w
}

// TestLArmeNeViseJamaisUnFigurant garde ce que la séparation des bassins donne
// par construction.
//
// **C'est le cas qui justifie la table à part.** La visée prend le plus proche
// sans que le joueur puisse choisir : un figurant collé au personnage
// détournerait chaque salve, et la mécanique du Secouriste — qui repose
// entièrement sur cette visée — tomberait avec. Le cas pose donc les figurants
// **plus près** que la créature, c'est-à-dire à l'endroit exact où une visée
// naïve les prendrait.
func TestLArmeNeViseJamaisUnFigurant(t *testing.T) {
	w := arenePeuplee(t, 4)
	if _, ok := w.SpawnEnemy(indexDuProfil(t, w.profils, "marcheur"),
		w.playerX+FromInt(3), w.playerY); !ok {
		t.Fatal("bassin plein")
	}
	cible := w.Enemies().At(0)
	plein := cible.Hits

	for range 3 * TPS {
		w.Step(Vec{})
	}
	if cible.Hits >= plein {
		t.Errorf("la créature a encore %d touches sur %d : l'arme a tiré ailleurs, "+
			"donc sur des figurants", cible.Hits, plein)
	}
	if n := w.Ambients().Len(); n != 4 {
		t.Errorf("%d figurants sur 4 : quelque chose en a retiré", n)
	}
}

// TestUnFigurantNePoussePersonne garde la décision qui aurait produit une
// mécanique que personne n'a voulue.
//
// S'ils entraient dans la grille de densité, on apprendrait à se placer derrière
// une foule de civils pour dévier la horde : une tactique réelle, obtenue sans
// qu'aucune décision de conception ne l'ait demandée, qu'il faudrait ensuite
// équilibrer ou retirer.
//
// Le cas compare deux mondes identiques dont l'un est peuplé, et exige la même
// trajectoire à la position près.
func TestUnFigurantNePoussePersonne(t *testing.T) {
	trajet := func(figurants int) Vec {
		w := arenePeuplee(t, figurants)
		if _, ok := w.SpawnEnemy(indexDuProfil(t, w.profils, "marcheur"),
			w.playerX+FromInt(4), w.playerY); !ok {
			t.Fatal("bassin plein")
		}
		for range 2 * TPS {
			w.Step(Vec{})
		}
		e := w.Enemies().At(0)
		return Vec{X: e.X, Y: e.Y}
	}

	if seul, entoure := trajet(0), trajet(8); seul != entoure {
		t.Errorf("la créature finit en %v seule et en %v au milieu des figurants : "+
			"ils la poussent", seul, entoure)
	}
}

// TestUnFigurantChangeDeCap vérifie qu'il va et vient au lieu de filer droit.
//
// Un figurant qui garderait son cap quitterait l'écran en quelques secondes, et
// l'ambiance se viderait là où le joueur regarde — c'est le seul endroit où elle
// sert.
func TestUnFigurantChangeDeCap(t *testing.T) {
	w := arenePeuplee(t, 1)
	a := w.Ambients().At(0)
	depart := a.Heading

	change := false
	for range 4 * int(erranceMin+erranceEcart) {
		w.Step(Vec{})
		if a.Heading != depart {
			change = true
			break
		}
	}
	if !change {
		t.Error("le figurant n'a jamais changé de cap : il traverse le lieu en " +
			"ligne droite et sort du champ")
	}
}

// TestUnPeuplementMalEcritEstRefuse vérifie que le chargement dit ce qui cloche.
//
// **Les deux derniers cas sont ceux qui justifient d'écrire les positions.** Un
// peuplement tiré au sort ne pourrait ni refuser un mur ni refuser un hors-bord :
// il déplacerait ou abandonnerait en silence, et un lieu qui demande douze
// figurants en poserait neuf sans que son auteur l'apprenne.
//
// Les deux premiers gardent la séparation des tables : un Badaud posé en
// figurant serait une horde gratuite, hors du budget de pression et du plafond
// d'effectif, que rien dans la courbe n'expliquerait.
func TestUnPeuplementMalEcritEstRefuse(t *testing.T) {
	profils, err := LoadProfiles(cohue.Assets, manifestePersonnages)
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}

	carte := NewCostGrid(8, 8)
	carte.Set(2, 2, Blocked)
	libre := [2]int{1, 1}
	mur := [2]int{2, 2}
	dehors := [2]int{99, 1}

	for _, c := range []struct {
		nom     string
		groupe  AmbientGroup
		attendu string
	}{
		{"un profil d'ennemi", AmbientGroup{Profile: "marcheur", At: &libre},
			"n'est pas un profil d'ambiance"},
		{"un nom inconnu", AmbientGroup{Profile: "badaud", At: &libre},
			"n'est pas un profil d'ambiance"},
		{"une position absente", AmbientGroup{Profile: "civil"},
			"un figurant se place"},
		{"une position dans un mur", AmbientGroup{Profile: "civil", At: &mur},
			"est dans un mur"},
		{"une position hors du lieu", AmbientGroup{Profile: "civil", At: &dehors},
			"hors du lieu"},
	} {
		t.Run(c.nom, func(t *testing.T) {
			_, manques := CompileAmbient(AmbientSpec{c.groupe}, profils, carte)
			if !contient(manques, c.attendu) {
				t.Errorf("manquements %v, aucun ne dit « %s »", manques, c.attendu)
			}
		})
	}
}
