// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce que le contact retire, ce que le plafond retient, et la mort comme état.

package game

import "testing"

// coller pose des créatures sur la position même du joueur.
//
// **Les tests qui mesurent une perte appellent `subir` et non `Step`.** Une
// créature posée sur le joueur ne l'y reste pas : la poussée de séparation
// l'écarte dès le tick suivant, si bien qu'un tour de boucle complet mesurerait
// la dispersion plutôt que la règle. Isoler la passe est ici légitime — ce qui
// est éprouvé est le rythme et le plafond, pas le chemin qui amène au contact,
// et `TestLaHordeFinitParBlesser` garde ce second point.
func coller(t *testing.T, w *World, profil, combien int) {
	t.Helper()
	px, py := w.Player()
	for range combien {
		if _, ok := w.SpawnEnemy(profil, px, py); !ok {
			t.Fatal("le bassin refuse une créature de plus")
		}
	}
}

// TestLeContactRetireDeLaVieEnContinu garde le rythme, pas seulement la perte.
//
// Un Badaud collé retire ce que son profil annonce par seconde : au bout d'une
// seconde exacte il en manque autant, ni un de plus ni un de moins. C'est ce que
// l'accumulateur en points-ticks rend exact — six points par seconde font un
// dixième de point par tick, qu'aucun entier ne représente, et une division
// faite à chaque tick perdrait le reste jusqu'à ne rien retirer du tout.
func TestLeContactRetireDeLaVieEnContinu(t *testing.T) {
	w, profils := champSansTir(t)
	badaud := indexDuProfil(t, profils, "marcheur")
	coller(t, w, badaud, 1)
	depart := w.Health()

	for range TPS {
		w.subir()
	}

	perdu := depart - w.Health()
	if attendu := profils.Enemies[badaud].ContactDamage; perdu != attendu {
		t.Errorf("%d points perdus en une seconde, attendu %d", perdu, attendu)
	}
}

// TestLePlafondTientQuelQueSoitLeNombreDEnnemis garde la règle du chapitre 5.
//
// Sans lui, un encerclement tue en une poignée de secondes et la mort devient
// illisible : le joueur ne distingue plus une erreur de placement d'un coup
// qu'il n'a pas vu venir.
//
// **Le plafond porte sur la somme et non sur chaque créature.** Un plafond par
// créature laisserait passer les dix, ce qui est le défaut que la règle existe
// pour empêcher — et un test qui n'en collerait qu'une ne verrait pas la
// différence entre les deux lectures.
func TestLePlafondTientQuelQueSoitLeNombreDEnnemis(t *testing.T) {
	const colles = 10

	w, profils := champSansTir(t)
	badaud := indexDuProfil(t, profils, "marcheur")
	coller(t, w, badaud, colles)

	if n := colles * profils.Enemies[badaud].ContactDamage; n <= profils.Player.DamageCap {
		t.Fatalf("le cas ne teste rien : %d de dégâts sous un plafond de %d",
			n, profils.Player.DamageCap)
	}
	depart := w.Health()

	for range TPS {
		w.subir()
	}

	if perdu := depart - w.Health(); perdu != profils.Player.DamageCap {
		t.Errorf("%d points perdus en une seconde, plafond à %d",
			perdu, profils.Player.DamageCap)
	}
}

// TestSansContactAucunePerte est le témoin des deux tests ci-dessus.
//
// Sans lui, une implémentation qui retirerait de la vie à chaque tick sans
// regarder les distances les ferait passer tous les deux : ils mesurent ce qu'on
// perd au contact, jamais qu'on ne perd rien sans lui.
func TestSansContactAucunePerte(t *testing.T) {
	w, profils := champSansTir(t)
	px, py := w.Player()
	if _, ok := w.SpawnEnemy(indexDuProfil(t, profils, "marcheur"),
		px+FromInt(6), py); !ok {
		t.Fatal("la créature n'a pas été posée")
	}
	depart := w.Health()

	for range TPS {
		w.subir()
	}

	if w.Health() != depart {
		t.Errorf("%d points perdus sans contact", depart-w.Health())
	}
}

// TestLaVieAZeroEstLaMort garde que l'état ne se double pas d'un drapeau.
//
// La vie ne descend pas sous zéro et la perte cesse : sans cela, un encerclement
// prolongé rendrait une vie négative, et l'écran de mort afficherait un nombre
// qu'aucune règle ne produit. C'est la forme qu'a déjà la résistance d'une
// créature, où la mort *est* l'état plutôt qu'un événement posé à côté.
func TestLaVieAZeroEstLaMort(t *testing.T) {
	w, profils := champSansTir(t)
	coller(t, w, indexDuProfil(t, profils, "marcheur"), 10)

	secondes := profils.Player.Health/profils.Player.DamageCap + 2
	for range secondes * TPS {
		w.subir()
	}

	if w.Alive() {
		t.Fatalf("encore vivant après %d secondes d'encerclement, %d points restants",
			secondes, w.Health())
	}
	if w.Health() != 0 {
		t.Errorf("vie de %d après la mort, attendu 0", w.Health())
	}
}

// TestLaHordeFinitParBlesser est l'autre moitié : le contact arrive vraiment.
//
// Les quatre tests ci-dessus appellent `subir` sur des créatures posées à la
// main, donc aucun ne dirait rien si la boucle ne l'appelait jamais ou si le
// champ de flux n'amenait personne. Celui-ci part d'une horde à distance, joue
// des ticks entiers, et exige que la vie ait baissé — c'est le chemin complet,
// des intentions au contact.
func TestLaHordeFinitParBlesser(t *testing.T) {
	w, profils := champSansTir(t)
	badaud := indexDuProfil(t, profils, "marcheur")
	px, py := w.Player()

	for i := range 8 {
		if _, ok := w.SpawnEnemy(badaud, px+FromInt(2+i%3), py+FromInt(i%4)); !ok {
			t.Fatal("la créature n'a pas été posée")
		}
	}
	depart := w.Health()

	for range 5 * TPS {
		w.Step(Vec{})
	}

	if w.Health() >= depart {
		t.Errorf("aucune perte en cinq secondes : la horde n'atteint pas le joueur")
	}
}

// TestLeVoileDureAutantQueLeContact garde ce que le voile suit.
//
// **Allumé tant que la horde coûte de la vie**, quelle que soit la durée du
// contact : c'est la première moitié, et elle tomberait sur un voile qui
// s'éteindrait sous une horde toujours collée, c'est-à-dire précisément quand le
// joueur est en train de mourir.
//
// **Éteint après le contact, et pas au premier tick.** La seconde moitié tient
// les deux bouts : encore allumé au tick qui suit le dernier contact, parti une
// fois la durée écoulée. Sans le premier, une créature qui entre et sort de
// portée ferait battre l'écran entier.
//
// **Ce qu'il ne garde pas, et ne peut pas garder** : la façon dont le voile est
// entretenu pendant le contact. `World.Hurt` rendant un booléen, le reposer à
// chaque tick ou le rallumer quand il retombe à zéro produisent le même allumé —
// éprouvé contre l'autre implémentation, ce cas passe sur les deux. La godoc de
// `eclairContact` dit ce qui les séparerait.
//
// La durée se lit dans `eclairContact` plutôt que de s'écrire ici : ce que le
// test garde est que le voile suit le contact, jamais le chiffre qui le règle.
func TestLeVoileDureAutantQueLeContact(t *testing.T) {
	w, profils := champSansTir(t)
	if w.Hurt() {
		t.Fatal("le voile est allumé avant tout contact")
	}

	coller(t, w, indexDuProfil(t, profils, "marcheur"), 1)

	// Plus long que la durée du voile : c'est ce qui distingue « allumé pendant
	// le contact » de « allumé un instant au début ».
	for range 2*eclairContact + 1 {
		w.subir()
	}
	if !w.Hurt() {
		t.Error("le voile s'est éteint sous une horde toujours collée")
	}

	for w.ennemis.Len() > 0 {
		w.ennemis.RemoveAt(0)
	}

	w.subir()
	if !w.Hurt() {
		t.Error("le voile s'éteint au premier tick sans contact, il clignoterait")
	}
	for range eclairContact {
		w.subir()
	}
	if w.Hurt() {
		t.Errorf("le voile dure plus de %d ticks après la fin du contact", eclairContact)
	}
}

// TestLeContactNalloueRien est le jumeau, sur le contact, de
// `TestLaBoucleNalloueRien`.
//
// Celui-là mesure des ticks entiers sur une partie montée, où presque aucune
// créature ne touche le joueur ; celui-ci pose le cas que la boucle ne produit
// pas d'elle-même — un bassin entier collé —, où le parcours et la somme
// tournent à chaque tick.
func TestLeContactNalloueRien(t *testing.T) {
	w, profils := champSansTir(t)
	coller(t, w, indexDuProfil(t, profils, "marcheur"), w.Enemies().Cap())

	moyenne := testing.AllocsPerRun(200, func() {
		w.subir()
	})
	if moyenne != 0 {
		t.Errorf("%v allocation(s) par tick de contact, attendu aucune", moyenne)
	}
}
