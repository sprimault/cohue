// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas de la tourelle : ce qu'une charge pose, ce qu'elle tire sans le
// joueur, et le tir qu'elle garde tant que personne n'est à portée.

package game

import "testing"

// TestUneTourelleSePoseSousLeJoueur garde ce qui la sépare des deux autres
// effets.
//
// **Sans cible, et c'est tout l'objet du cas.** La grenade et la gerbe ne
// partent pas dans le vide ; une tourelle se pose pour tenir un passage avant
// que la horde n'arrive, et exiger une cible interdirait l'anticipation qu'elle
// récompense.
func TestUneTourelleSePoseSousLeJoueur(t *testing.T) {
	w, _ := champDeTir(t)
	tomber(t, w, "tourelle")
	w.ramasserUneArme()
	px, py := w.Player()
	_, avant := w.HeldHeavy(0)

	w.Trigger(0)

	if n := w.Turrets().Len(); n != 1 {
		t.Fatalf("%d tourelle(s) posée(s) dans une salle vide, attendu 1", n)
	}
	if _, apres := w.HeldHeavy(0); apres != avant-1 {
		t.Errorf("%d charge(s) après la pose, attendu %d", apres, avant-1)
	}
	if tr := w.Turrets().At(0); tr.X != px || tr.Y != py {
		t.Errorf("posée en (%d, %d), attendu sous le joueur en (%d, %d)",
			tr.X, tr.Y, px, py)
	}
}

// TestUneTourelleTireSansLeJoueur ferme le chemin de la pose au projectile.
//
// La créature est posée à portée de la tourelle et le monde avance d'un tick :
// ce qui part n'a été demandé par aucune touche.
func TestUneTourelleTireSansLeJoueur(t *testing.T) {
	w, profils := champDeTir(t)
	w.arme = Weapon{} // le socle se tairait sinon, et ses tirs se compteraient ici
	tomber(t, w, "tourelle")
	w.ramasserUneArme()
	w.Trigger(0)
	px, py := w.Player()
	if _, ok := w.SpawnEnemy(indexDuProfil(t, profils, "marcheur"), px+FromInt(3), py); !ok {
		t.Fatal("créature refusée")
	}
	avant := w.Turrets().At(0).Shots

	w.Step(Vec{})

	if n := w.Shots().Len(); n == 0 {
		t.Error("la tourelle n'a rien tiré alors qu'une créature est à portée")
	}
	if reste := w.Turrets().At(0).Shots; reste != avant-1 {
		t.Errorf("%d tir(s) restant(s), attendu %d", reste, avant-1)
	}
}

// TestUneTourelleGardeSesTirsSansCible garde ce qui rend une charge toujours
// rendue.
//
// **Une tourelle posée d'avance ne s'use pas en attendant.** C'est la même règle
// que la cadence du socle, qui ne se consomme pas à vide, et c'est ce qui a fait
// préférer un compte de tirs à une durée : celle-ci se serait écoulée dans un
// couloir vide, et la charge aurait été perdue pour une raison que le joueur ne
// contrôle pas.
//
// Il tourne assez de ticks pour couvrir plusieurs cadences, sans quoi le cas
// passerait sur une tourelle qui n'a simplement pas encore eu son tour.
func TestUneTourelleGardeSesTirsSansCible(t *testing.T) {
	w, _ := champDeTir(t)
	w.arme = Weapon{}
	tomber(t, w, "tourelle")
	w.ramasserUneArme()
	w.Trigger(0)
	arme, _ := w.HeldHeavy(0)
	avant := w.Turrets().At(0).Shots

	for range 3 * int(arme.Cooldown) {
		w.Step(Vec{})
	}

	if reste := w.Turrets().At(0).Shots; reste != avant {
		t.Errorf("%d tir(s) restant(s) après une attente à vide, attendu %d", reste, avant)
	}
}

// TestUneTourelleEpuiseeDisparait garde la fin de sa vie.
//
// Sans ce cas, une tourelle qui ne se retirerait jamais tiendrait sa place dans
// le bassin — et le plafond ferait refuser des poses pour des tourelles qui ne
// peuvent plus rien.
func TestUneTourelleEpuiseeDisparait(t *testing.T) {
	w, profils := champDeTir(t)
	w.arme = Weapon{}
	tomber(t, w, "tourelle")
	w.ramasserUneArme()
	w.Trigger(0)
	arme, _ := w.HeldHeavy(0)
	px, py := w.Player()

	// Une créature remplacée à chaque fois qu'elle tombe : ce qu'on éprouve est
	// l'épuisement, pas la survie d'un cobaye.
	for range arme.Shots * (int(arme.Cooldown) + 1) {
		if w.Enemies().Len() == 0 {
			w.SpawnEnemy(indexDuProfil(t, profils, "marcheur"), px+FromInt(3), py)
		}
		w.Step(Vec{})
		if w.Turrets().Len() == 0 {
			return
		}
	}
	t.Errorf("la tourelle tient encore avec %d tir(s) après ses %d",
		w.Turrets().At(0).Shots, arme.Shots)
}
