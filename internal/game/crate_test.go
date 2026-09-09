// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas de la caisse : le délai d'appui et ce qu'il refuse, le coût qu'elle
// met sur sa case et lui rend, ce qu'elle laisse, et les refus qu'un semis mal
// écrit reçoit au chargement.

package game

import (
	"testing"

	"github.com/sprimault/cohue"
)

// salleAvecCaisse monte une salle et y pose une caisse, dans l'ordre du montage.
//
// **Cet ordre est ce que le cas éprouve autant que le reste** : la grille reçoit
// le coût de la caisse avant que le monde bâtisse son champ de flux, dont le
// nombre de seaux se dérive du plus grand coût qu'il y trouve. Une salle montée
// dans l'autre sens passerait ici et casserait la file de Dial en jeu.
//
// Le semis passe par `CompileCrates` plutôt que d'être bâti à la main : c'est
// elle qui relève le sol que la caisse recouvre, et une entrée de test doit être
// une entrée que le système accepterait.
//
// La caisse est loin du joueur, pour que chaque cas décide lui-même quand il
// l'atteint.
func salleAvecCaisse(t *testing.T) (*World, Fixed, Fixed) {
	t.Helper()
	profils, err := LoadProfiles(cohue.Assets, manifestePersonnages)
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}
	regles := caissesLivrees(t)

	g := NewCostGrid(48, 48)
	pose, manques := CompileCrates(CrateSpec{{At: &[2]int{10, 10}}}, g)
	if len(manques) > 0 {
		t.Fatalf("semis d'essai refusé : %v", manques)
	}
	StampCrates(g, pose, regles)

	w := NewWorld(profils, armesInertes(t), progressionLivree(t), regles, fiolesLivrees(t),
		vagueUnique(0), g, graineDeTest, capacitesDeTest)
	w.Place(FromInt(24)+One/2, FromInt(24)+One/2)
	w.Stock(pose)
	return w, pose[0].X, pose[0].Y
}

// casserLaCaisse pousse le joueur sur la caisse jusqu'à ce qu'elle cède.
//
// **Elle boucle plutôt qu'elle ne compte.** Écrire ici le nombre de ticks en
// ferait une seconde description du délai que le manifeste déclare, et la borne
// n'est là que pour arrêter un cas qui ne casserait rien plutôt que de tourner
// sans fin.
func casserLaCaisse(t *testing.T, w *World, x, y Fixed) {
	t.Helper()
	w.Place(x, y)
	borne := 2 * int(w.CratePress())
	for range borne {
		w.Step(Vec{})
		if w.Crates().Len() == 0 {
			return
		}
	}
	t.Fatalf("la caisse tient encore après %d ticks d'appui", borne)
}

// TestUneCaisseCedeApresLAppui garde les deux bouts du délai de contact.
//
// **Le premier compte autant que le second.** Sans lui, on ne casse pas en
// décidant d'y aller mais en frôlant, ce que la conception refuse ; sans le
// second, le délai serait une attente dont rien ne sort. Un cas qui ne
// vérifierait que la casse passerait sur un délai nul.
func TestUneCaisseCedeApresLAppui(t *testing.T) {
	w, x, y := salleAvecCaisse(t)
	attendu := w.progression.CrateGems
	if attendu < 1 {
		t.Fatalf("la progression livrée annonce %d gemme(s) par caisse", attendu)
	}
	appui := int(w.CratePress())
	if appui < 2 {
		t.Fatalf("un appui de %d tick ne se distingue pas d'une casse au contact", appui)
	}

	w.Place(x, y)
	for range appui - 1 {
		w.Step(Vec{})
	}
	if n := w.Crates().Len(); n != 1 {
		t.Fatalf("la caisse a cédé avant la fin de l'appui, %d debout", n)
	}
	if n := w.Gems().Len(); n != 0 {
		t.Errorf("%d gemme(s) tombées avant que la caisse cède", n)
	}

	w.Step(Vec{})
	if n := w.Crates().Len(); n != 0 {
		t.Errorf("%d caisse(s) encore debout après %d ticks d'appui", n, appui)
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

// TestUneCaisseNeSeCassePasEnPassant garde la remise à zéro du décompte.
//
// **Deux passages qui totalisent plus que le délai ne cassent rien.** C'est ce
// que « on ne casse pas en passant, on casse en décidant d'y aller » veut dire :
// un décompte qui se cumulerait ferait céder une caisse qu'on longe trois fois
// sans jamais s'y arrêter, et le délai cesserait d'être une décision pour
// devenir une usure.
func TestUneCaisseNeSeCassePasEnPassant(t *testing.T) {
	w, x, y := salleAvecCaisse(t)
	appui := int(w.CratePress())

	for range 3 {
		w.Place(x, y)
		for range appui - 1 {
			w.Step(Vec{})
		}
		// Quatre tuiles, franchement au-delà de la portée de contact : le
		// décompte doit repartir entier.
		w.Place(x+FromInt(4), y)
		w.Step(Vec{})
	}

	if n := w.Crates().Len(); n != 1 {
		t.Errorf("%d caisse(s) debout après trois passages, attendu une : le "+
			"decompte se cumule d'un passage a l'autre", n)
	}
	if n := w.Gems().Len(); n != 0 {
		t.Errorf("%d gemme(s) tombées sans qu'un appui soit allé au bout", n)
	}
}

// TestUneCaisseCouteAtraverserPuisRendSaCase garde le ralentissement et sa fin.
//
// **Le coût est le vrai prix de la ressource** : ramasser, c'est perdre du
// terrain. Et il doit disparaître avec elle — une case qui garderait son coût
// ralentirait la horde sur un point qu'aucun pixel ne montre plus, et le champ
// de flux la contournerait, c'est-à-dire ferait un détour autour de rien.
func TestUneCaisseCouteAtraverserPuisRendSaCase(t *testing.T) {
	w, x, y := salleAvecCaisse(t)
	u, v := x.Floor(), y.Floor()

	if cout := w.grille.At(u, v); cout <= Free {
		t.Fatalf("la case d'une caisse coûte %d : elle se traverse comme le sol", cout)
	}
	// La vitesse s'en trouve divisée, et c'est ce que le coût existe pour faire :
	// sans cela le parcours pondéré contournerait au prix de deux cases ce qui ne
	// coûte rien à traverser.
	base := w.profils.Player.Speed
	if freinee := w.vitesse(base, x, y); freinee >= base {
		t.Errorf("vitesse de %v sur la caisse contre %v sur le sol", freinee, base)
	}

	casserLaCaisse(t, w, x, y)

	if got := w.grille.At(u, v); got != Free {
		t.Errorf("la case coûte encore %d après la casse, attendu %d", got, Free)
	}
}

// TestUneCaisseCasseeLaisseUneEpave garde la trace qu'elle doit laisser.
//
// **Une épave et non un effet bref.** Ce qu'elle dit est qu'une caisse a été
// ouverte ici, et cette information reste vraie jusqu'à la fin de la run : la
// ranger parmi les effets l'aurait fait disparaître au bout d'une demi-seconde,
// et le lieu n'aurait plus gardé la trace de ce qu'on y a pris.
func TestUneCaisseCasseeLaisseUneEpave(t *testing.T) {
	w, x, y := salleAvecCaisse(t)
	if n := w.Wrecks().Len(); n != 0 {
		t.Fatalf("%d épave(s) avant qu'une caisse cède", n)
	}

	casserLaCaisse(t, w, x, y)

	if n := w.Wrecks().Len(); n != 1 {
		t.Fatalf("%d épave(s) après la casse, attendu une", n)
	}
	e := w.Wrecks().At(0)
	if e.X != x || e.Y != y {
		t.Errorf("épave en (%d, %d), la caisse était en (%d, %d)", e.X, e.Y, x, y)
	}
	// Un tick au sortir du tick où elle naît, comme l'âge d'une gemme : le
	// compteur de la partie avance en fin de tick, et l'âge se lit après.
	if age := w.WreckAge(e); age != 1 {
		t.Errorf("épave d'âge %d au sortir du tick où elle naît, attendu 1", age)
	}

	// Elle ne s'efface pas : une seconde d'attente ne doit rien lui retirer.
	for range TPS {
		w.Step(Vec{})
	}
	if n := w.Wrecks().Len(); n != 1 {
		t.Errorf("%d épave(s) une seconde plus tard : elle s'efface", n)
	}
	if age := w.WreckAge(w.Wrecks().At(0)); age != TPS+1 {
		t.Errorf("âge de %d après une seconde, attendu %d", age, TPS+1)
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

// TestDeuxCaissesSurUneCaseSeRefusent garde ce que le coût a rendu nécessaire.
//
// Elles étaient une redondance sans conséquence tant qu'une caisse n'écrivait
// rien : maintenant la première cassée rend la case au sol sous la seconde, qui
// cesse de ralentir ce qu'elle devrait ralentir. Le refus nomme les deux rangs,
// sans quoi l'auteur d'un semis de trente ne saurait pas laquelle déplacer.
func TestDeuxCaissesSurUneCaseSeRefusent(t *testing.T) {
	brut := CrateSpec{{At: &[2]int{2, 2}}, {At: &[2]int{9, 9}}, {At: &[2]int{2, 2}}}
	_, manques := CompileCrates(brut, NewCostGrid(16, 16))

	if !contient(manques, "caisses[2]") || !contient(manques, "caisses[0]") {
		t.Errorf("le refus ne nomme pas les deux rangs en cause : %v", manques)
	}
	if len(manques) != 1 {
		t.Errorf("%d manquement(s) pour un seul doublon : %v", len(manques), manques)
	}
}

// TestUnSemisEcritSeCompile vérifie le cas valide, où il pose les caisses, et le
// sol qu'il relève sous chacune.
func TestUnSemisEcritSeCompile(t *testing.T) {
	carte := NewCostGrid(16, 16)
	// Une case lente sous la seconde caisse : c'est ce qu'elle devra rendre en
	// cédant, et le relever au montage serait trop tard — la case porte alors
	// déjà le coût de la caisse.
	const boue Cost = 2
	carte.Set(9, 2, boue)

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
	if pose[0].Floor != Free {
		t.Errorf("sol de %d sous la première caisse, attendu %d", pose[0].Floor, Free)
	}
	if pose[1].Floor != boue {
		t.Errorf("sol de %d sous la seconde caisse, attendu %d", pose[1].Floor, boue)
	}
}

// TestUnLieuSansCaisseNeSeRefusePas vérifie que l'absence n'est pas une faute.
func TestUnLieuSansCaisseNeSeRefusePas(t *testing.T) {
	pose, manques := CompileCrates(nil, NewCostGrid(16, 16))
	if len(pose) > 0 || len(manques) > 0 {
		t.Errorf("un lieu sans caisse est refusé : %v, %v", pose, manques)
	}
}
