// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package game

import "testing"

// TestLeGradientPointeVersLaFoule fixe le sens, qui est la seule chose qu'on
// puisse inverser sans que rien ne compile de travers.
//
// L'appelant soustrait ce vecteur : orienté à l'envers, la horde s'agglutinerait
// au lieu de se desserrer, et le symptôme — des créatures qui s'empilent —
// ressemble exactement à une répulsion trop faible.
func TestLeGradientPointeVersLaFoule(t *testing.T) {
	d := NewDensityGrid(5, 5)
	d.Add(3, 2)
	d.Add(3, 2)
	d.Add(3, 2)

	g := d.Gradient(2, 2)
	if g.X <= 0 {
		t.Errorf("gradient %v : il ne pointe pas vers les trois entités de droite", g)
	}
	if g.Y != 0 {
		t.Errorf("gradient %v : rien ne le pousse verticalement", g)
	}
}

// TestUneEntiteDEcartVautUneTuile fixe l'échelle, que la boucle multipliera.
//
// Elle est écrite en clair plutôt que dérivée d'un calcul : un test qui refait
// l'arithmétique du code passe même quand les deux sont faux.
func TestUneEntiteDEcartVautUneTuile(t *testing.T) {
	d := NewDensityGrid(5, 5)
	d.Add(3, 2)
	d.Add(3, 2)

	if g := d.Gradient(2, 2); g.X != FromInt(2) {
		t.Errorf("deux entités d'écart donnent %d, attendu %d", g.X, FromInt(2))
	}
}

// TestUneFouleUniformeNeRepousseRien vérifie une propriété qui ressemble à un
// défaut et n'en est pas un.
//
// Dans une masse de densité égale, aucune direction n'est moins encombrée
// qu'une autre : c'est la poussée des bords qui desserre la foule, de proche en
// proche. Quelqu'un qui « corrigerait » ce cas ajouterait une force sans
// direction juste.
func TestUneFouleUniformeNeRepousseRien(t *testing.T) {
	d := NewDensityGrid(5, 5)
	for v := range 5 {
		for u := range 5 {
			d.Add(u, v)
			d.Add(u, v)
		}
	}

	if g := d.Gradient(2, 2); g != (Vec{}) {
		t.Errorf("gradient %v au milieu d'une foule uniforme, attendu nul", g)
	}
}

// TestLeBordPousseVersLInterieur vérifie que l'extérieur de la carte compte pour
// vide.
//
// Une créature collée au mur de la carte n'a de voisins que d'un côté ; le
// gradient la renvoie donc vers l'intérieur, ce qui est le comportement voulu.
func TestLeBordPousseVersLInterieur(t *testing.T) {
	d := NewDensityGrid(4, 4)
	d.Add(1, 0)

	if g := d.Gradient(0, 0); g.X <= 0 {
		t.Errorf("gradient %v au bord gauche, attendu tourné vers l'intérieur", g)
	}
}

// TestHorsGrilleNeCompteNiNePlante éprouve les deux entrées que la boucle peut
// produire sans le vouloir.
//
// Une entité poussée hors carte par la séparation, ou apparue au mauvais
// endroit, ne doit ni faire paniquer le comptage ni compter ailleurs.
func TestHorsGrilleNeCompteNiNePlante(t *testing.T) {
	d := NewDensityGrid(3, 3)
	d.Add(-1, 0)
	d.Add(0, 3)
	d.Add(99, 99)

	for v := range 3 {
		for u := range 3 {
			if n := d.At(u, v); n != 0 {
				t.Errorf("(%d, %d) compte %d entité(s), aucune n'était dans la grille", u, v, n)
			}
		}
	}
	if n := d.At(-1, 0); n != 0 {
		t.Errorf("une cellule hors grille compte %d, attendu 0", n)
	}
}

// TestClearRemetLaGrilleAZero vérifie ce que le début de chaque tick suppose.
//
// Sans cet effacement, les comptes s'accumuleraient d'une image à l'autre et la
// répulsion croîtrait sans fin — une horde qui explose au bout de quelques
// secondes, sans qu'aucune entité ait bougé anormalement.
func TestClearRemetLaGrilleAZero(t *testing.T) {
	d := NewDensityGrid(4, 4)
	d.Add(1, 1)
	d.Add(2, 2)
	d.Clear()

	for v := range 4 {
		for u := range 4 {
			if n := d.At(u, v); n != 0 {
				t.Fatalf("(%d, %d) compte encore %d après l'effacement", u, v, n)
			}
		}
	}
}

// TestLaDensiteNalloueRien garde le budget : la grille est effacée et remplie à
// chaque tick, pour toute la horde.
func TestLaDensiteNalloueRien(t *testing.T) {
	d := NewDensityGrid(64, 64)

	moyenne := testing.AllocsPerRun(1000, func() {
		d.Clear()
		for i := range 300 {
			d.Add(i%64, i/64)
		}
		_ = d.Gradient(10, 3)
	})
	if moyenne != 0 {
		t.Errorf("%v allocation(s) par tick, attendu aucune", moyenne)
	}
}
