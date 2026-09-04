// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas de la charge : l'annonce immobile, la direction qui ne se corrige
// plus, les trois fins de course et la récupération qui les suit toutes, le choc
// unique hors plafond, et la horde ordinaire qui ne passe jamais par le cycle.

package game

import (
	"testing"

	"github.com/sprimault/cohue"
)

// arene monte une salle vide de 32 cases, joueur au centre, arme inerte.
//
// L'arme ne tire pas : ces cas comptent des ticks de charge, et un Molosse
// abattu en pleine course ne dirait rien de ce qu'ils gardent. La grille est
// rendue pour que le cas qui veut un pilier le pose lui-même.
func arene(t *testing.T) (*World, *EnemyProfile, *CostGrid) {
	t.Helper()
	profils, err := LoadProfiles(cohue.Assets, manifestePersonnages)
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}

	g := NewCostGrid(32, 32)
	w := NewWorld(profils, armesInertes(t), progressionLivree(t), sansVagues(), g,
		graineDeTest, capacitesDeTest)
	w.Place(FromInt(16)+One/2, FromInt(16)+One/2)

	molosse := &profils.Enemies[indexDuProfil(t, profils, "sprinteur")]
	if molosse.ChargeRange == 0 || molosse.Telegraph == 0 || molosse.ChargeDuration == 0 {
		t.Fatalf("le Molosse ne charge pas : portée %v, télégraphe %d, course %d",
			molosse.ChargeRange, molosse.Telegraph, molosse.ChargeDuration)
	}
	return w, molosse, g
}

// poserMolosse place un Molosse à l'est du joueur, à la distance donnée.
//
// À l'est et sur la même ligne : la charge part alors plein ouest, si bien qu'un
// écart en Y à l'arrivée ne peut venir que d'une correction en cours de course.
func poserMolosse(t *testing.T, w *World, tuiles Fixed) *Enemy {
	t.Helper()
	if _, ok := w.SpawnEnemy(indexDuProfil(t, w.profils, "sprinteur"),
		w.playerX+tuiles, w.playerY); !ok {
		t.Fatal("bassin plein")
	}
	return w.Enemies().At(0)
}

// TestLaChargeSAnnonceImmobile vérifie que l'anticipation est un temps mort et
// non une approche.
//
// **C'est elle qui rend l'esquive possible.** Une créature qui avancerait
// pendant son annonce n'annoncerait rien : le joueur verrait un Molosse
// s'approcher, ce qu'il fait déjà le reste du temps. Le cas relève donc la
// position, pas seulement la phase.
func TestLaChargeSAnnonceImmobile(t *testing.T) {
	w, molosse, _ := arene(t)
	e := poserMolosse(t, w, FromInt(4))
	depart := Vec{X: e.X, Y: e.Y}

	for range molosse.Telegraph {
		w.Step(Vec{})
		if !e.Telegraphing() {
			t.Fatalf("phase %d pendant l'anticipation, attendu l'annonce", e.ChargePhase)
		}
		if e.X != depart.X || e.Y != depart.Y {
			t.Fatalf("la créature a bougé de %v,%v pendant son annonce",
				e.X-depart.X, e.Y-depart.Y)
		}
	}

	w.Step(Vec{})
	if !e.Charging() {
		t.Errorf("phase %d après l'anticipation, attendu la course", e.ChargePhase)
	}
}

// TestLaChargeNeCorrigePlus vérifie que la direction est figée au départ.
//
// **C'est tout le comportement**, et ce qui le sépare d'un poursuivant rapide :
// le joueur se décale d'un pas de côté et la charge passe à côté. Le cas déplace
// donc le joueur perpendiculairement pendant la course, et exige que le Molosse
// n'ait pas dévié d'un pouce en Y.
func TestLaChargeNeCorrigePlus(t *testing.T) {
	w, molosse, _ := arene(t)
	e := poserMolosse(t, w, FromInt(4))
	ligne := e.Y

	for range molosse.Telegraph + 1 {
		w.Step(Vec{})
	}
	for range molosse.ChargeDuration - 1 {
		w.Step(Vec{Y: -One})
		if !e.Charging() {
			t.Fatalf("la course s'est arrêtée avant terme, phase %d", e.ChargePhase)
		}
	}

	if e.Y != ligne {
		t.Errorf("la créature a dévié de %v en Y : elle a corrigé sa course", e.Y-ligne)
	}
	if e.X >= w.playerX+FromInt(4) {
		t.Error("la créature n'a pas avancé pendant sa course")
	}
}

// TestLePilierArreteLaChargeSansEmpecherSonDepart éprouve les deux moitiés de la
// mécanique du décor, qui ne valent que prises ensemble.
//
// Rien ne vérifie la ligne de vue au déclenchement : la charge part **malgré**
// l'obstacle, et c'est ce qui la rend punissable. Vérifier la vue ferait qu'un
// Molosse derrière un pilier attendrait sagement, et le décor perdrait le seul
// usage défensif que la conception lui donne.
//
// Puis elle s'arrête au mur — le pas voulu et le pas obtenu cessent d'être
// égaux — et la récupération commence, sans que le joueur ait été touché.
func TestLePilierArreteLaChargeSansEmpecherSonDepart(t *testing.T) {
	w, molosse, g := arene(t)
	e := poserMolosse(t, w, FromInt(4))

	// Le pilier est posé entre les deux, sur la ligne de la charge.
	g.Set(18, 16, Blocked)

	for range molosse.Telegraph + 1 {
		w.Step(Vec{})
	}
	if !e.Charging() {
		t.Fatalf("phase %d : le pilier a empêché la charge de partir", e.ChargePhase)
	}

	// **Le compte de ticks est ce qui discrimine.** Une course qui va à son
	// terme finit elle aussi en récupération : sans mesurer *quand*, le cas
	// passerait à l'identique sur un arrêt au mur désarmé — ce qu'une mutation a
	// montré. Deux tuiles et demie séparent la créature du pilier, soit un peu
	// plus de vingt ticks, là où la course en dure quarante-deux.
	ticks := Tick(0)
	for range molosse.ChargeDuration {
		w.Step(Vec{})
		ticks++
		if e.ChargePhase == ChargeRecover {
			break
		}
	}

	if e.ChargePhase != ChargeRecover {
		t.Fatalf("phase %d en fin de course : le mur ne l'a pas arrêtée", e.ChargePhase)
	}
	if ticks >= molosse.ChargeDuration {
		t.Errorf("la course a tenu ses %d ticks : elle s'est éteinte d'elle-même "+
			"et non contre le pilier", ticks)
	}
	if e.X < FromInt(19) {
		t.Errorf("la créature est à %v, elle a traversé le pilier", e.X)
	}
	if w.Health() != w.MaxHealth() {
		t.Errorf("vie %d : la charge a touché malgré le pilier", w.Health())
	}
}

// TestLeChocDeChargeEstUniqueEtHorsPlafond garde les deux propriétés que la
// conception donne au choc, qu'un seul relevé confondrait.
//
// **Unique**, parce que la fin de course est ce qui le rend tel : sans elle, une
// créature collée infligerait dix-huit points par tick. **Hors plafond**, parce
// que le plafond couvre le contact continu — trente corps dont on ne distingue
// pas la part — et non un coup annoncé puis manqué.
//
// **C'est une meute qui l'éprouve, et pas un chien seul.** Dix-huit points sont
// sous le plafond de vingt : un choc unique passerait à l'identique plafonné ou
// non, et le cas ne séparerait rien. Trois qui percutent ensemble font
// cinquante-quatre, ce qui est exactement l'exemple par lequel la conception
// justifie la règle.
func TestLeChocDeChargeEstUniqueEtHorsPlafond(t *testing.T) {
	w, molosse, _ := arene(t)
	const meute = 3
	if meute*molosse.ChargeDamage <= w.profils.Player.DamageCap {
		t.Fatalf("%d chocs de %d sous le plafond de %d : le cas ne sépare plus rien",
			meute, molosse.ChargeDamage, w.profils.Player.DamageCap)
	}
	// **Hors contact, à deux tuiles.** Collés, ils infligeraient leur contact
	// ordinaire pendant les trente ticks immobiles de l'annonce, et le relevé
	// mêlerait deux mécanismes : c'est ce que le premier jet de ce cas mesurait,
	// dix points de trop.
	for range meute {
		poserMolosse(t, w, FromInt(2))
	}

	for range molosse.Telegraph + molosse.ChargeDuration {
		w.Step(Vec{})
		if w.Health() < w.MaxHealth() {
			break
		}
	}
	if perdu := w.MaxHealth() - w.Health(); perdu != meute*molosse.ChargeDamage {
		t.Errorf("%d points perdus au choc, attendu %d : le plafond a mordu dessus",
			perdu, meute*molosse.ChargeDamage)
	}
	for i := range w.Enemies().Active() {
		if e := w.Enemies().At(i); e.ChargePhase != ChargeRecover {
			t.Errorf("créature %d en phase %d après le choc : sa course y a survécu",
				i, e.ChargePhase)
		}
	}

	// **Le tick suivant, et lui seul.** Ce qu'il faut séparer est un choc par
	// course d'un choc par tick — collées, elles en infligeraient un à chaque
	// image et le joueur tomberait en deux. Laisser courir deux secondes ne le
	// dirait pas : la récupération finie, un second cycle est légitime, et il
	// aurait achevé un joueur déjà à quarante-six.
	avant := w.Health()
	w.Step(Vec{})
	if perdu := avant - w.Health(); perdu > 1 {
		t.Errorf("%d points au tick suivant le choc : il se répète au lieu de "+
			"clore la course", perdu)
	}
}

// TestLaRecuperationSuitUneChargeAboutie vérifie que le temps mort ne récompense
// pas que l'esquive.
//
// Une charge qui finit sa course sans rien heurter enchaînerait sinon sur la
// suivante, et la créature n'aurait aucun moment vulnérable — ce que la
// conception lui refuse en lui opposant l'esquive latérale. Le cas laisse donc
// la course aller à son terme, loin du joueur et de tout mur.
//
// Le joueur fuit, ce qui décale le déclenchement d'un tick ou deux — la distance
// se rouvre pendant qu'il monte : la course est donc attendue par son issue et
// non par un compte de ticks, qui supposerait un départ au premier tick.
func TestLaRecuperationSuitUneChargeAboutie(t *testing.T) {
	w, molosse, _ := arene(t)
	e := poserMolosse(t, w, FromInt(4))

	for range 4 * (molosse.Telegraph + molosse.ChargeDuration) {
		w.Step(Vec{Y: One})
		if e.ChargePhase == ChargeRecover {
			break
		}
	}
	if e.ChargePhase != ChargeRecover {
		t.Fatalf("phase %d en fin de course aboutie, attendu la récupération",
			e.ChargePhase)
	}
	if w.Health() != w.MaxHealth() {
		t.Fatalf("vie %d : la course s'est terminée sur le joueur, pas d'elle-même",
			w.Health())
	}

	fige := Vec{X: e.X, Y: e.Y}
	for range molosse.Recovery - 1 {
		w.Step(Vec{})
	}
	if e.X != fige.X || e.Y != fige.Y {
		t.Errorf("la créature a bougé de %v,%v pendant sa récupération",
			e.X-fige.X, e.Y-fige.Y)
	}
	if e.ChargePhase != ChargeRecover {
		t.Errorf("phase %d avant la fin du temps mort", e.ChargePhase)
	}
}

// TestUnProfilSansPorteeNePasseJamaisParLeCycle garde ce qui permet à la charge
// de vivre dans la boucle commune sans y peser.
//
// Le Badaud partage la passe de tous les autres : si son cycle s'ouvrait, il
// s'immobiliserait un tiers du temps sans que rien dans son profil ne le
// demande. C'est la portée nulle qui ferme le mécanisme, et non un test sur le
// comportement.
func TestUnProfilSansPorteeNePasseJamaisParLeCycle(t *testing.T) {
	w, _, _ := arene(t)
	marcheur := indexDuProfil(t, w.profils, "marcheur")
	if w.profils.Enemies[marcheur].ChargeRange != 0 {
		t.Fatal("le Badaud porte une portée de charge, ce cas suppose le contraire")
	}
	if _, ok := w.SpawnEnemy(marcheur, w.playerX+FromInt(3), w.playerY); !ok {
		t.Fatal("bassin plein")
	}

	e := w.Enemies().At(0)
	for range 3 * TPS {
		w.Step(Vec{})
		if e.ChargePhase != ChargeNone {
			t.Fatalf("phase %d sur un profil sans portée de charge", e.ChargePhase)
		}
	}
}
