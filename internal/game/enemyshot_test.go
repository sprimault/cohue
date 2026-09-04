// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas du tir de la horde : la portée qui arrête l'approche, la cadence qui
// ne se consomme pas à vide, le projectile qui blesse hors plafond et meurt sur
// un pilier, et la horde ordinaire qui ne tire jamais.

package game

import (
	"testing"

	"github.com/sprimault/cohue"
)

// buse monte l'arène et rend le profil du Cracheur, ses quatre champs vérifiés.
//
// Sans eux les cas qui suivent perdraient leur sujet : une portée nulle les
// ferait tous passer sur une créature qui ne tire pas.
func buse(t *testing.T) (*World, *EnemyProfile, *CostGrid) {
	t.Helper()
	profils, err := LoadProfiles(cohue.Assets, manifestePersonnages)
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}

	g := NewCostGrid(32, 32)
	w := NewWorld(profils, armesInertes(t), progressionLivree(t), sansVagues(), g,
		graineDeTest, 16, 64, 16, 32)
	w.Place(FromInt(16)+One/2, FromInt(16)+One/2)

	p := &profils.Enemies[indexDuProfil(t, profils, "cracheur")]
	if p.Range == 0 || p.ShotCooldown == 0 || p.ShotDamage == 0 || p.ShotSpeed == 0 {
		t.Fatalf("la Buse ne tire pas : portée %v, cadence %d, dégâts %d, vitesse %v",
			p.Range, p.ShotCooldown, p.ShotDamage, p.ShotSpeed)
	}
	return w, p, g
}

// poserBuse place un Cracheur à l'est du joueur, à la distance donnée.
func poserBuse(t *testing.T, w *World, tuiles Fixed) *Enemy {
	t.Helper()
	if _, ok := w.SpawnEnemy(indexDuProfil(t, w.profils, "cracheur"),
		w.playerX+tuiles, w.playerY); !ok {
		t.Fatal("bassin plein")
	}
	return w.Enemies().At(0)
}

// TestLaBuseSeStabilisePourTirer vérifie qu'elle cesse d'avancer dès qu'elle est
// à portée.
//
// **C'est ce qui la sépare du reste de la horde.** Une créature dont tout le
// rôle est de blesser de loin et qui continue de converger finit au contact, où
// elle ne vaut plus rien ; c'est aussi son immobilité qui la rend repérable au
// milieu d'une foule qui, elle, avance.
func TestLaBuseSeStabilisePourTirer(t *testing.T) {
	w, profil, _ := buse(t)
	e := poserBuse(t, w, FromInt(9))

	for range 8 * TPS {
		w.Step(Vec{})
	}

	ecart := Vec{X: e.X - w.playerX, Y: e.Y - w.playerY}.Len()
	if ecart > profil.Range {
		t.Fatalf("la Buse est à %v du joueur : elle n'a pas approché jusqu'à sa "+
			"portée de %v", ecart, profil.Range)
	}
	// Une tuile de marge sous la portée : elle s'arrête au tick où elle entre,
	// pas au point exact, et exiger l'égalité serait exiger un arrondi.
	if ecart < profil.Range-One {
		t.Errorf("la Buse est à %v pour une portée de %v : elle a continué "+
			"d'avancer au lieu de se stabiliser", ecart, profil.Range)
	}
}

// TestLaBuseNeConsommePasSaCadenceAVide reprend sur la horde le cas limite que
// la conception nomme pour l'arme du joueur.
//
// Une cadence qui tourne hors de portée ferait de la première salve une fonction
// du temps passé sans cible : la Buse qui voit le joueur revenir tirerait aussi
// tôt qu'elle a attendu, ce que rien à l'écran n'expliquerait.
//
// **Le cas ne vaut qu'après un premier tir**, et son premier jet l'ignorait :
// parti d'une cadence à zéro, il ne laissait rien à consommer et passait sur le
// code fautif. Il fait donc tirer la Buse, l'éloigne plus longtemps que sa
// cadence, puis la ramène — si le compteur avait tourné à vide, elle tirerait
// aussitôt.
//
// La durée d'éloignement est première avec la cadence, pour que le compteur ne
// retombe pas prêt par coïncidence.
func TestLaBuseNeConsommePasSaCadenceAVide(t *testing.T) {
	w, profil, _ := buse(t)
	e := poserBuse(t, w, FromInt(3))

	w.Step(Vec{})
	if w.EnemyShots().Len() != 1 {
		t.Fatalf("%d projectile(s) au premier tick : la Buse n'a pas ouvert le feu",
			w.EnemyShots().Len())
	}

	// Hors de portée, reposée à chaque tick : sans cela le champ de flux la
	// ramènerait vers le joueur et elle rentrerait dans sa portée en route.
	loin := w.playerX + FromInt(20)
	for range profil.ShotCooldown + 7 {
		e.X, e.Y = loin, w.playerY
		w.Step(Vec{})
	}
	if n := w.EnemyShots().Len(); n != 0 {
		t.Fatalf("%d projectile(s) en vol après l'éloignement, attendu aucun", n)
	}

	e.X, e.Y = w.playerX+FromInt(3), w.playerY
	w.Step(Vec{})
	if n := w.EnemyShots().Len(); n != 0 {
		t.Errorf("%d projectile(s) au retour à portée : la cadence a tourné à vide "+
			"pendant l'absence de cible", n)
	}
}

// TestLeTirDeLaBuseBlesseHorsPlafond garde la seconde des trois sources de
// dégâts que la conception distingue.
//
// Le plafond couvre le contact continu, qu'on ne voit pas venir dans une foule ;
// un projectile se voit arriver. Le cas mesure la vie perdue au tick de
// l'impact, la Buse étant seule et hors de portée de contact — ce qui s'y perd
// ne peut donc venir que du tir.
func TestLeTirDeLaBuseBlesseHorsPlafond(t *testing.T) {
	w, profil, _ := buse(t)
	poserBuse(t, w, FromInt(3))

	for range 4 * TPS {
		w.Step(Vec{})
		if w.Health() < w.MaxHealth() {
			break
		}
	}
	if perdu := w.MaxHealth() - w.Health(); perdu != profil.ShotDamage {
		t.Errorf("%d points perdus, attendu %d", perdu, profil.ShotDamage)
	}
	if n := w.EnemyShots().Len(); n != 0 {
		t.Errorf("%d projectile(s) survivent à leur impact", n)
	}
}

// TestLaBuseTireASaCadence vérifie qu'elle tire à répétition et non une fois.
//
// **Une mutation l'a réclamé** : cesser de décrémenter le compteur laisse une
// Buse qui tire son premier projectile puis plus jamais, et aucun des autres cas
// ne s'en apercevait — tous se contentent d'un tir.
//
// Le compte se lit sur la vie plutôt que sur le bassin, dont les projectiles
// sortent dès qu'ils touchent. La durée est choisie première avec la cadence, et
// laisse au dernier tir le temps d'arriver : sans cette marge, le cas mesurerait
// la vitesse du projectile autant que la cadence.
func TestLaBuseTireASaCadence(t *testing.T) {
	w, profil, _ := buse(t)
	poserBuse(t, w, FromInt(3))

	const duree = 330
	for range duree {
		w.Step(Vec{})
	}

	// Le premier part au tick où elle entre à portée, les suivants toutes les
	// `ShotCooldown` : quatre en trois cent trente ticks pour une cadence de
	// quatre-vingt-seize.
	veut := int(1 + (duree-1)/profil.ShotCooldown)
	if perdu := w.MaxHealth() - w.Health(); perdu != veut*profil.ShotDamage {
		t.Errorf("%d points perdus en %d ticks, attendu %d pour %d tirs",
			perdu, duree, veut*profil.ShotDamage, veut)
	}
}

// TestLePilierArreteLeTirDeLaBuse vérifie que le décor protège par le fait.
//
// Rien ne vérifie la voie au déclenchement, comme pour la charge : la Buse tire,
// et c'est le projectile qui meurt sur l'obstacle par le chemin qu'emprunte déjà
// un tir du joueur. Le cas exige donc les deux — qu'elle tire, et que le joueur
// ne perde rien.
func TestLePilierArreteLeTirDeLaBuse(t *testing.T) {
	w, _, g := buse(t)
	poserBuse(t, w, FromInt(3))
	g.Set(18, 16, Blocked)

	tire := false
	for range 4 * TPS {
		w.Step(Vec{})
		if w.EnemyShots().Len() > 0 {
			tire = true
		}
	}
	if !tire {
		t.Error("la Buse n'a pas tiré : le pilier l'en a empêchée au lieu de " +
			"stopper ses projectiles")
	}
	if w.Health() != w.MaxHealth() {
		t.Errorf("vie %d sur %d : un tir a traversé le pilier",
			w.Health(), w.MaxHealth())
	}
}

// TestUnProfilSansPorteeNeTirePas garde ce qui permet au tir de vivre dans la
// passe commune sans y peser.
//
// Le Badaud partage la boucle de la horde : si sa portée nulle ne le sortait
// pas, il tirerait des projectiles sans dégâts ni vitesse, et le bassin se
// remplirait de tirs immobiles.
func TestUnProfilSansPorteeNeTirePas(t *testing.T) {
	w, _, _ := buse(t)
	marcheur := indexDuProfil(t, w.profils, "marcheur")
	if w.profils.Enemies[marcheur].Range != 0 {
		t.Fatal("le Badaud porte une portée, ce cas suppose le contraire")
	}
	if _, ok := w.SpawnEnemy(marcheur, w.playerX+FromInt(3), w.playerY); !ok {
		t.Fatal("bassin plein")
	}

	for range 4 * TPS {
		w.Step(Vec{})
		if n := w.EnemyShots().Len(); n != 0 {
			t.Fatalf("%d projectile(s) tirés par un profil sans portée", n)
		}
	}
}
