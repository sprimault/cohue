// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce que le bassin des effets garde : une caisse qui cède en pose un, il vit ce
// qu'on lui donne, et rien de ce qu'il contient ne décide de quoi que ce soit.

package game

import "testing"

// TestUneCaisseCasseeLaisseUnEffet vérifie qu'un événement se retient.
//
// **C'est la seule chose que la simulation doive au rendu ici** : le point et
// l'instant. Une caisse qui quitterait son bassin sans rien laisser ne pourrait
// plus être dessinée en train de céder, puisque plus rien n'en garderait la
// trace — c'est ce qui distingue un effet bref d'une animation d'entité.
func TestUneCaisseCasseeLaisseUnEffet(t *testing.T) {
	w, x, y := salleAvecCaisse(t)
	casserLaCaisse(t, w, x, y)

	effets := w.Fxs()
	if effets.Len() != 1 {
		t.Fatalf("%d effet(s) après la casse, attendu un", effets.Len())
	}
	e := effets.At(0)
	if e.Kind != FxCrate {
		t.Errorf("sorte %d, attendu celle d'une caisse", e.Kind)
	}
	if e.X != x || e.Y != y {
		t.Errorf("effet posé en (%d, %d), la caisse était en (%d, %d)", e.X, e.Y, x, y)
	}
	if e.Life != e.Total || e.Total <= 0 {
		t.Errorf("vie %d sur %d : un effet naît entier", e.Life, e.Total)
	}
}

// TestUnEffetVitCeQuOnLuiDonnePuisSEnVa garde les deux bouts de sa durée.
//
// **Le premier tick compte, et c'est ce qui a placé le vieillissement en fin de
// tick.** Décompté aussitôt émis, un effet perdrait l'image qui le rend lisible ;
// gardé un tick de trop, il resterait à l'écran après avoir fini de s'y
// montrer. Les deux se vérifient au tick près, et aucun ne se voit à l'œil.
func TestUnEffetVitCeQuOnLuiDonnePuisSEnVa(t *testing.T) {
	w, x, y := salleAvecCaisse(t)
	casserLaCaisse(t, w, x, y)

	total := w.Fxs().At(0).Total
	if total <= 1 {
		t.Fatalf("un effet qui dure %d tick ne peut rien montrer", total)
	}
	// Il reste entier au tick de son émission : le vieillissement vient après
	// tout ce qui émet.
	if vie := w.Fxs().At(0).Life; vie != total {
		t.Errorf("vie %d au tick de l'émission, attendu %d", vie, total)
	}

	for range total - 1 {
		w.Step(Vec{})
	}
	if n := w.Fxs().Len(); n != 1 {
		t.Fatalf("%d effet(s) au dernier tick de sa vie, attendu un", n)
	}
	w.Step(Vec{})
	if n := w.Fxs().Len(); n != 0 {
		t.Errorf("%d effet(s) une fois la durée écoulée", n)
	}
}

// TestUnBassinDeffetsPleinPerdLeSuivant garde le refus plutôt que l'éviction.
//
// Deux caisses cassées dans la même seconde valent mieux qu'une caisse dont les
// éclats sautent en plein vol : un effet chassé se verrait, un effet jamais posé
// non. Et le perdre ne coûte rien qu'un peu de décor, ce qui est le sens du mot
// cosmétique.
func TestUnBassinDeffetsPleinPerdLeSuivant(t *testing.T) {
	w, x, y := salleAvecCaisse(t)
	w.effets = NewPool[Fx](1)

	w.emettre(x, y, FxCrate)
	w.emettre(x, y, FxBlast)

	if n := w.Fxs().Len(); n != 1 {
		t.Fatalf("%d effet(s) dans un bassin d'un seul", n)
	}
	if sorte := w.Fxs().At(0).Kind; sorte != FxCrate {
		t.Errorf("sorte %d retenue : le second a chassé le premier", sorte)
	}
}
