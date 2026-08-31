// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas du vecteur : la direction toujours unitaire, la plus petite diagonale
// qui reste une diagonale, le vecteur nul qui tire son orientation de l'index de
// son entité, et le budget d'allocation.

package game

import "testing"

// TestDirectionEstUnitaire vérifie que la normalisation rend bien une longueur
// d'une tuile, quelle que soit celle du vecteur d'entrée.
//
// Les cas sont choisis pour couvrir les trois régimes : un axe, une diagonale
// exacte, un triplet pythagoricien, et un vecteur d'une seule unité par axe —
// ce dernier étant celui qu'une norme arrondie avant division ramènerait à une
// longueur et demie.
func TestDirectionEstUnitaire(t *testing.T) {
	cas := []struct {
		quoi string
		v    Vec
	}{
		{"un axe", Vec{FromInt(3), 0}},
		{"une diagonale", Vec{FromInt(2), FromInt(2)}},
		{"un triplet 3-4-5", Vec{FromInt(3), FromInt(4)}},
		{"une unité par axe", Vec{1, 1}},
		{"un axe négatif", Vec{0, FromInt(-7)}},
	}
	for _, c := range cas {
		if l := c.v.Direction(0).Len(); l != One {
			t.Errorf("%s : la direction mesure %d, attendu %d", c.quoi, l, One)
		}
	}
}

// TestLaPlusPetiteDiagonaleResteUneDiagonale éprouve la somme exacte des carrés.
//
// Une norme calculée puis arrondie à l'unité vaudrait 1 pour ce vecteur, et la
// division rendrait deux composantes d'une tuile chacune — une direction une
// fois et demie trop longue, qui ferait avancer deux fois plus vite en
// diagonale. Le défaut ne se verrait que sur les vecteurs minuscules, c'est-à-
// dire au moment où deux entités se séparent.
func TestLaPlusPetiteDiagonaleResteUneDiagonale(t *testing.T) {
	if d := (Vec{1, 1}).Direction(0); d != (Vec{diag, diag}) {
		t.Errorf("la direction vaut %v, attendu {%d %d}", d, diag, diag)
	}
}

// TestVecteurNulPrendLaDirectionDeSonIndex vérifie ce que la conception demande
// au cas nul : une réponse par entité, et des réponses qui s'écartent vite.
//
// Deux entités exactement superposées reçoivent des corrections opposées à plus
// d'un quart de tour, donc elles se séparent au tick suivant. Une direction fixe
// leur donnerait la même correction et les laisserait superposées indéfiniment.
func TestVecteurNulPrendLaDirectionDeSonIndex(t *testing.T) {
	vues := make(map[Vec]int, 8)
	for index := range 8 {
		d := (Vec{}).Direction(index)
		if l := d.Len(); l != One {
			t.Errorf("index %d : direction de longueur %d, attendu %d", index, l, One)
		}
		if precedent, deja := vues[d]; deja {
			t.Errorf("les index %d et %d reçoivent la même direction %v", precedent, index, d)
		}
		vues[d] = index
	}

	// Le produit scalaire de deux index consécutifs est négatif : leur écart
	// dépasse le quart de tour. Un pas de un dans la table n'en donnerait que
	// 45 degrés, et deux créatures apparues côte à côte se sépareraient
	// lentement — or leurs index sont justement voisins.
	for index := range 8 {
		a := (Vec{}).Direction(index)
		b := (Vec{}).Direction(index + 1)
		if scalaire := int64(a.X)*int64(b.X) + int64(a.Y)*int64(b.Y); scalaire >= 0 {
			t.Errorf("les index %d et %d s'écartent d'un quart de tour ou moins", index, index+1)
		}
	}
}

// TestDirectionResteDansLaTable vérifie que le cas nul répond aussi pour un
// index quelconque, sans sortir de la table.
func TestDirectionResteDansLaTable(t *testing.T) {
	for _, index := range []int{0, 1, 7, 8, 300, 2999} {
		if l := (Vec{}).Direction(index).Len(); l != One {
			t.Errorf("index %d : direction de longueur %d", index, l)
		}
	}
}

// TestPerpTourneSansAllonger vérifie le quart de tour dont dérive la dérive
// latérale d'un flanqueur.
func TestPerpTourneSansAllonger(t *testing.T) {
	v := Vec{FromInt(3), FromInt(4)}
	p := v.Perp()

	if p != (Vec{FromInt(-4), FromInt(3)}) {
		t.Errorf("la perpendiculaire vaut %v, attendu {%d %d}", p, FromInt(-4), FromInt(3))
	}
	if p.Len() != v.Len() {
		t.Errorf("la perpendiculaire mesure %d, l'original %d", p.Len(), v.Len())
	}
	// Perpendiculaires : leur produit scalaire est nul.
	if scalaire := int64(v.X)*int64(p.X) + int64(v.Y)*int64(p.Y); scalaire != 0 {
		t.Errorf("produit scalaire %d, attendu 0", scalaire)
	}
}

// TestVecteurNalloueRien garde le budget de la boucle : le calcul de direction
// est fait par chaque entité et à chaque image.
func TestVecteurNalloueRien(t *testing.T) {
	v := Vec{FromInt(3), FromInt(4)}
	moyenne := testing.AllocsPerRun(1000, func() {
		_ = v.Direction(0).Perp().Scale(One / 2).Add(v).Sub(v)
	})
	if moyenne != 0 {
		t.Errorf("%v allocation(s) par calcul, attendu aucune", moyenne)
	}
}
