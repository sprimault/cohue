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
// Un Quidam collé retire ce que son profil annonce par seconde : au bout d'une
// seconde exacte il en manque autant, ni un de plus ni un de moins. C'est ce que
// l'accumulateur en points-ticks rend exact — six points par seconde font un
// dixième de point par tick, qu'aucun entier ne représente, et une division
// faite à chaque tick perdrait le reste jusqu'à ne rien retirer du tout.
func TestLeContactRetireDeLaVieEnContinu(t *testing.T) {
	w, profils := champSansTir(t)
	quidam := indexDuProfil(t, profils, "marcheur")
	coller(t, w, quidam, 1)
	depart := w.Health()

	for range TPS {
		w.subir()
	}

	perdu := depart - w.Health()
	if attendu := profils.Enemies[quidam].ContactDamage; perdu != attendu {
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
	quidam := indexDuProfil(t, profils, "marcheur")
	coller(t, w, quidam, colles)

	if n := colles * profils.Enemies[quidam].ContactDamage; n <= profils.Player.DamageCap {
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

// TestSansContactAucunePerte est le témoin de
// `TestLeContactRetireDeLaVieEnContinu` et de
// `TestLePlafondTientQuelQueSoitLeNombreDEnnemis`.
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
// `TestLeContactRetireDeLaVieEnContinu`,
// `TestLePlafondTientQuelQueSoitLeNombreDEnnemis`, `TestSansContactAucunePerte`
// et `TestLaVieAZeroEstLaMort` appellent `subir` sur des créatures posées à la
// main, donc aucun ne dirait rien si la boucle ne l'appelait jamais ou si le
// champ de flux n'amenait personne. Celui-ci part d'une horde à distance, joue
// des ticks entiers, et exige que la vie ait baissé — c'est le chemin complet,
// des intentions au contact.
func TestLaHordeFinitParBlesser(t *testing.T) {
	w, profils := champSansTir(t)
	quidam := indexDuProfil(t, profils, "marcheur")
	px, py := w.Player()

	for i := range 8 {
		if _, ok := w.SpawnEnemy(quidam, px+FromInt(2+i%3), py+FromInt(i%4)); !ok {
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

// TestLeVigileBlesseAuContact est le cas qu'un profil solide rend possible et
// que `TestLaHordeFinitParBlesser` ne peut pas voir : celui-là joue un Quidam,
// que rien n'arrête avant le joueur.
//
// **Ce qui est gardé est que la distance où le corps arrête soit une distance où
// le contact a lieu.** Les deux portées se posaient sur la même somme de rayons,
// si bien que la seule position que le Vigile pouvait atteindre était celle où le
// contact cesse : il publiait dix dégâts par seconde qu'il n'infligeait jamais,
// et rien autour ne le disait — il bloquait, il se contournait, il mourait sous
// les tirs.
//
// **Il passe donc par `Step`**, à l'inverse des cas qui appellent `subir` sur des
// créatures posées à la main : collé de force, le Vigile blessait déjà. C'était
// l'approche qui n'aboutissait pas.
//
// La perte n'est pas celle de deux secondes pleines — le joueur met quelques
// ticks à le rejoindre —, et le rythme exact est gardé par
// `TestLeContactRetireDeLaVieEnContinu`.
func TestLeVigileBlesseAuContact(t *testing.T) {
	w, profil := areneSolide(t)
	if _, ok := w.SpawnEnemy(indexDuProfil(t, w.profils, "bloqueur"),
		w.playerX+One, w.playerY); !ok {
		t.Fatal("bassin plein")
	}
	depart := w.Health()

	for range 2 * TPS {
		w.Step(Vec{X: One})
	}

	if perdu := depart - w.Health(); perdu < profil.ContactDamage {
		t.Errorf("%d points perdus en deux secondes contre le Vigile, attendu au "+
			"moins %d", perdu, profil.ContactDamage)
	}
}

// TestLAlerteSuitLaVieEtNonLeContact garde ce que l'alerte signale.
//
// **Elle ne s'allume pas sous les coups**, et c'est ce que la première moitié
// sépare : une horde collée qui entame la vie sans la faire descendre sous le
// seuil ne l'allume pas. C'est ce qui distingue cette version de celle qu'on
// avait d'abord écrite, où le voile suivait le contact et battait à chaque
// créature qui frôle.
//
// **Elle s'allume au franchissement du seuil et y reste**, quel que soit le
// nombre de ticks : c'est un état lu, pas un compteur qui s'épuise. La seconde
// moitié le tient en laissant tourner le contact bien après.
//
// Le seuil se lit dans le profil livré plutôt que de s'écrire ici, et la
// première assertion le confronte à la vie de départ : sans elle, un manifeste
// où les deux seraient égaux rendrait le cas vide, l'alerte étant allumée dès le
// premier tick.
func TestLAlerteSuitLaVieEtNonLeContact(t *testing.T) {
	w, profils := champSansTir(t)
	seuil := profils.Player.LowHealth
	if seuil <= 0 || seuil >= profils.Player.Health {
		t.Fatalf("seuil d'alerte à %d pour %d de vie : le cas ne teste rien",
			seuil, profils.Player.Health)
	}
	if w.InDanger() {
		t.Fatal("l'alerte est allumée à pleine vie")
	}

	coller(t, w, indexDuProfil(t, profils, "marcheur"), 1)

	// Assez pour entamer la vie, pas pour atteindre le seuil : c'est là que
	// l'alerte doit rester éteinte alors que le joueur encaisse.
	for w.Health() > seuil+1 && w.Health() > 0 {
		if w.InDanger() {
			t.Fatalf("alerte allumée à %d points, seuil %d", w.Health(), seuil)
		}
		w.subir()
	}

	for w.Health() > seuil {
		w.subir()
	}
	if !w.InDanger() {
		t.Fatalf("alerte éteinte à %d points, seuil %d", w.Health(), seuil)
	}

	// Elle ne s'épuise pas : la vie ne remonte pas toute seule, donc rien ne doit
	// la rendre au calme tant que le joueur vit.
	for range 3 * TPS {
		w.subir()
		if w.Alive() && !w.InDanger() {
			t.Fatalf("alerte retombée à %d points, seuil %d", w.Health(), seuil)
		}
	}
}

// TestLAlerteSEteintALaMort garde ce que l'écran de fin reprend.
//
// Mort, le joueur n'est plus en danger : c'est fait. Sans ce cas, une alerte
// laissée allumée sous le voile de mort durerait jusqu'à la relance, et le seul
// endroit qui l'aurait dit est un écran qu'aucun test ne peut regarder.
//
// `TestLAlerteSEteintALaSortie` garde l'autre fin, et il faut les deux : la vie
// éteint l'alerte à la mort par elle-même, jamais sur une sortie.
func TestLAlerteSEteintALaMort(t *testing.T) {
	w, profils := champSansTir(t)
	coller(t, w, indexDuProfil(t, profils, "marcheur"), 10)

	for range 60 * TPS {
		w.subir()
		if !w.Alive() {
			break
		}
	}
	if w.Alive() {
		t.Fatal("le joueur survit : le cas n'atteint pas la mort")
	}
	if w.InDanger() {
		t.Error("l'alerte reste allumée après la mort")
	}
}

// TestLAlerteSEteintALaSortie est le jumeau du précédent, sur l'autre fin.
//
// **Les deux issues partagent leur écran**, ce que le rendu affirme. Une alerte
// qui s'éteignait à la mort et restait allumée sur une sortie les séparait
// pourtant : le joueur qui s'en tire sous le seuil voyait un bord rouge sur
// l'écran qui lui annonce qu'il a gagné le lieu.
//
// **La vie ne pouvait pas le dire**, et c'est pourquoi ce cas manquait : elle
// tombe à zéro dans un cas et pas dans l'autre, si bien qu'un prédicat adossé à
// elle traite une fin sur deux. C'est `Over` qui porte les deux.
func TestLAlerteSEteintALaSortie(t *testing.T) {
	w, _ := salleAvecPorte(t)

	// Sous le seuil avant de sortir : sinon l'alerte serait éteinte pour la
	// raison ordinaire, et le cas passerait sur le code fautif.
	w.vie = w.profils.Player.LowHealth
	if !w.InDanger() {
		t.Fatalf("l'alerte n'est pas allumée à %d points pour un seuil de %d",
			w.Health(), w.profils.Player.LowHealth)
	}

	for range porteAbattus {
		w.SpawnEnemy(0, FromInt(30), FromInt(30))
		w.Enemies().At(w.Enemies().Len() - 1).Hits = 0
		w.Step(Vec{})
	}
	w.Place(FromInt(porteU)+One/2, FromInt(porteV+1)+One/2)
	w.Step(Vec{})

	if !w.Escaped() {
		t.Fatal("le joueur n'est pas sorti : le cas ne pose plus sa question")
	}
	if w.InDanger() {
		t.Error("l'alerte reste allumée après la sortie : les deux fins ne " +
			"montrent pas le même écran")
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
