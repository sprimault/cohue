// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas de la caisse : ce qui la casse, ce qui ne la casse pas, ce qu'elle
// laisse, et les refus qu'un semis mal écrit reçoit au chargement.

package game

import "testing"

// salleAvecCaisse monte une salle ouverte et pose une caisse loin du joueur.
//
// Loin, pour que chaque cas décide lui-même quand le joueur l'atteint : posée
// sous lui, elle serait cassée au premier tick et aucun cas ne pourrait
// distinguer « au contact » de « n'importe où ».
func salleAvecCaisse(t *testing.T) (*World, Fixed, Fixed) {
	t.Helper()
	w, _ := salleOuverte(t, vagueUnique(0), 16)
	x, y := FromInt(10)+One/2, FromInt(10)+One/2
	w.Stock([]CratePlacement{{X: x, Y: y}})
	return w, x, y
}

// TestUneCaisseSeCasseAuContact vérifie qu'elle laisse ce que la progression
// annonce, et qu'elle ne le laisse qu'une fois.
func TestUneCaisseSeCasseAuContact(t *testing.T) {
	w, x, y := salleAvecCaisse(t)
	attendu := w.progression.CrateGems
	if attendu < 1 {
		t.Fatalf("la progression livrée annonce %d gemme(s) par caisse", attendu)
	}

	w.Place(x, y)
	w.Step(Vec{})

	if n := w.Crates().Len(); n != 0 {
		t.Errorf("%d caisse(s) encore debout après le contact", n)
	}
	if n := w.Gems().Len(); n != attendu {
		t.Errorf("%d gemme(s) laissées, attendu %d", n, attendu)
	}

	// Une seconde passe au même endroit : la caisse a quitté le bassin, donc
	// rien de plus ne tombe. Sans la suppression, le contact continu en poserait
	// une volée par tick.
	avant := w.Gems().Len()
	w.Step(Vec{})
	if n := w.Gems().Len(); n > avant {
		t.Errorf("%d gemmes après un second tick au contact, contre %d : la "+
			"caisse cassée laisse encore du butin", n, avant)
	}
}

// TestUneCaisseHorsDePorteeTient garde ce qui fait d'elle un détour.
//
// Sans cette borne, casser ne coûterait rien : les caisses tomberaient au fur et
// à mesure que la partie avance, et le choix d'aller les chercher — qui est ce
// que la sonde existe pour faire sentir — n'existerait pas.
func TestUneCaisseHorsDePorteeTient(t *testing.T) {
	w, x, y := salleAvecCaisse(t)

	// Trois tuiles, largement au-delà de la portée de contact et bien en deçà de
	// ce qu'un cas de portée nulle laisserait passer.
	w.Place(x+FromInt(3), y)
	for range TPS {
		w.Step(Vec{})
	}

	if n := w.Crates().Len(); n != 1 {
		t.Errorf("%d caisse(s) debout à trois tuiles, attendu une", n)
	}
	if n := w.Gems().Len(); n != 0 {
		t.Errorf("%d gemme(s) tombées sans que personne ne touche la caisse", n)
	}
}

// TestUneCaisseNEstPasUneCible garde la décision qui la sort du bassin des
// ennemis.
//
// **La visée automatique prend la plus proche sans que le joueur choisisse.**
// Une caisse rangée parmi les cibles détournerait donc chaque salve vers le
// décor, et emporterait avec elle la mécanique du Secouriste, qui repose
// entièrement sur cette visée. C'est la même règle que pour le figurant.
//
// Le cas met une caisse seule dans la salle : l'arme ne doit rien avoir à viser.
func TestUneCaisseNEstPasUneCible(t *testing.T) {
	w, x, y := salleAvecCaisse(t)
	w.Place(x+FromInt(2), y)

	for range 2 * TPS {
		w.Step(Vec{})
	}
	if n := w.Shots().Len(); n != 0 {
		t.Errorf("%d projectile(s) partis vers une caisse : l'arme la vise", n)
	}
	if n := w.Crates().Len(); n != 1 {
		t.Errorf("%d caisse(s) debout après deux secondes de tir, attendu une", n)
	}
}

// TestUnSemisMalEcritSeRefuse vérifie que le chargement nomme la caisse fautive.
//
// Le rang est dans le message, comme pour un figurant : un lieu qui en pose
// trente n'apprendrait rien d'un refus qui dit « une position est dans un mur ».
func TestUnSemisMalEcritSeRefuse(t *testing.T) {
	carte := NewCostGrid(16, 16)
	carte.Set(4, 4, Blocked)

	cas := []struct {
		nom    string
		brut   CrateSpec
		attend string
	}{
		{"position absente", CrateSpec{{}}, "position"},
		{"hors du lieu", CrateSpec{{At: &[2]int{99, 2}}}, "hors du lieu"},
		{"dans un mur", CrateSpec{{At: &[2]int{4, 4}}}, "dans un mur"},
	}
	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			pose, manques := CompileCrates(c.brut, carte)
			if len(pose) > 0 {
				t.Error("une caisse a été posée là où le fichier est faux")
			}
			if !contient(manques, c.attend) {
				t.Errorf("le refus ne nomme pas %q : %v", c.attend, manques)
			}
			if !contient(manques, "caisses[0]") {
				t.Errorf("le refus ne dit pas laquelle : %v", manques)
			}
		})
	}
}

// TestUnSemisEcritSeCompile vérifie le cas valide, et où il pose les caisses.
func TestUnSemisEcritSeCompile(t *testing.T) {
	carte := NewCostGrid(16, 16)
	pose, manques := CompileCrates(CrateSpec{{At: &[2]int{4, 4}}, {At: &[2]int{9, 2}}}, carte)
	if len(manques) > 0 {
		t.Fatalf("un semis bien écrit est refusé : %v", manques)
	}
	if len(pose) != 2 {
		t.Fatalf("%d caisse(s) posées, attendu deux", len(pose))
	}
	// Au centre de la case et non sur son coin, comme un figurant et comme la
	// porte : c'est de là que se mesure la distance au joueur.
	if pose[0].X != FromInt(4)+One/2 || pose[0].Y != FromInt(4)+One/2 {
		t.Errorf("caisse en (%v, %v), attendue au centre de sa case", pose[0].X, pose[0].Y)
	}
}

// TestUnLieuSansCaisseNeSeRefusePas vérifie que l'absence n'est pas une faute.
func TestUnLieuSansCaisseNeSeRefusePas(t *testing.T) {
	pose, manques := CompileCrates(nil, NewCostGrid(16, 16))
	if len(pose) > 0 || len(manques) > 0 {
		t.Errorf("un lieu sans caisse est refusé : %v, %v", pose, manques)
	}
}
