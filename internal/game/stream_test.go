// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas des flux : même graine et même suite, indépendance des quatre,
// empreinte d'une suite connue, dérivation d'une run à la suivante, et bornes
// des tirages.

package game

import "testing"

// TestMemeGraineMemeSuite vérifie qu'une graine rejouée rend la même suite.
//
// C'est la promesse la plus basse du déterminisme, et celle dont tout le reste
// dépend : sans elle, ni le rejeu d'une mort injuste, ni la comparaison de deux
// équilibrages.
func TestMemeGraineMemeSuite(t *testing.T) {
	a, b := NewStreams(20260830), NewStreams(20260830)
	for i := 0; i < 100; i++ {
		if x, y := a.Waves.IntN(1000), b.Waves.IntN(1000); x != y {
			t.Fatalf("tirage %d : %d puis %d", i, x, y)
		}
	}
}

// TestLesFluxSontIndependants vérifie qu'épuiser un flux ne décale pas les autres.
//
// C'est ce que les quatre flux achètent, et l'unique raison de leur existence :
// une run simulée sans rendu ne tire pas les teintes, et doit pourtant produire
// exactement les mêmes vagues que la même graine jouée à l'écran.
func TestLesFluxSontIndependants(t *testing.T) {
	avec, sans := NewStreams(7), NewStreams(7)

	// La run jouée tire ses teintes ; la run simulée ne les tire pas.
	for i := 0; i < 10000; i++ {
		avec.Cosmetic.IntN(6)
	}

	for i := 0; i < 100; i++ {
		if x, y := avec.Waves.IntN(1000), sans.Waves.IntN(1000); x != y {
			t.Fatalf("vague %d : %d avec les teintes, %d sans", i, x, y)
		}
	}
}

// TestLesFluxNeSeSuiventPas vérifie que deux flux d'une même graine ne rendent
// pas la même suite.
//
// Un flux qui reprendrait la suite d'un autre corrélerait le contenu des caisses
// aux positions d'apparition, et personne ne verrait pourquoi le butin dépend de
// l'endroit où l'on se tient.
func TestLesFluxNeSeSuiventPas(t *testing.T) {
	s := NewStreams(99)
	var identiques int
	for i := 0; i < 100; i++ {
		if s.Waves.IntN(1<<30) == s.Positions.IntN(1<<30) {
			identiques++
		}
	}
	if identiques > 1 {
		t.Errorf("%d tirages sur 100 identiques entre deux flux", identiques)
	}
}

// TestEmpreinteDeLaSuite fige la suite d'une graine connue.
//
// C'est ce qui garde l'algorithme : si une version de Go changeait PCG, ou si
// quelqu'un remplaçait la source, ce test tomberait — alors que tous les autres
// continueraient de passer, puisqu'ils ne vérifient que la cohérence interne.
// Toutes les graines publiées deviendraient pourtant invalides.
//
// Il tourne sur les trois cibles natives de l'intégration continue : c'est là
// qu'il prouve ce que le déterminisme promet.
func TestEmpreinteDeLaSuite(t *testing.T) {
	// Relevées à l'écriture du test, sur une graine de 1. Elles ne se
	// régénèrent pas : un attendu réécrit sans être relu ne teste plus rien.
	const (
		vagues     = 282552
		positions  = 980240
		butin      = 747314
		cosmetique = 391763
	)
	s := NewStreams(1)
	cas := []struct {
		nom     string
		obtenu  int
		attendu int
	}{
		{"vagues", s.Waves.IntN(1 << 20), vagues},
		{"positions", s.Positions.IntN(1 << 20), positions},
		{"butin", s.Loot.IntN(1 << 20), butin},
		{"cosmetique", s.Cosmetic.IntN(1 << 20), cosmetique},
	}
	for _, c := range cas {
		if c.obtenu != c.attendu {
			t.Errorf("flux %s : premier tirage %d, attendu %d — la source a changé",
				c.nom, c.obtenu, c.attendu)
		}
	}
}

// TestLaGraineDeriveeEstStableEtNouvelle éprouve la dérivation d'une run à la
// suivante.
//
// Deux propriétés, et l'une sans l'autre ne vaut rien : la dérivation rend
// toujours la même chose de la même graine, sinon la suite des runs d'une session
// cesse d'être rejouable ; et elle ne rend jamais une graine déjà sortie, sinon
// les runs se répètent au bout de quelques morts et personne ne sait pourquoi.
//
// Une chaîne de cent plutôt qu'un appel unique : la répétition qu'on craint n'est
// pas `NextSeed(g) == g`, qui se verrait tout de suite, mais un cycle court, qui
// ne se voit qu'en déroulant.
//
// Ce test ne dit rien du branchement de la graine dans une partie —
// `TestLaSuiteDesRunsDescendDeLaGraineDeDepart`, dans `internal/session`, garde
// cette moitié-là : une dérivation juste que la relance n'utiliserait pas
// laisserait toutes les runs identiques sans que ce test-ci bronche.
func TestLaGraineDeriveeEstStableEtNouvelle(t *testing.T) {
	vues := map[uint64]bool{}
	graine := graineDeTest
	for i := range 100 {
		suivante := NextSeed(graine)
		if suivante == graine {
			t.Fatalf("relance %d : la graine ne change pas (%d)", i, graine)
		}
		if rejoue := NextSeed(graine); rejoue != suivante {
			t.Fatalf("relance %d : %d puis %d pour la même graine", i, suivante, rejoue)
		}
		if vues[suivante] {
			t.Fatalf("relance %d : la graine %d est déjà sortie", i, suivante)
		}
		vues[suivante] = true
		graine = suivante
	}
}

// TestFixedResteDansSaBorne vérifie qu'une longueur tirée ne dépasse pas son
// maximum, et qu'un maximum nul ou négatif rend zéro plutôt qu'une panique.
func TestFixedResteDansSaBorne(t *testing.T) {
	s := NewStreams(3)
	max := FromInt(16)
	for i := 0; i < 1000; i++ {
		if v := s.Positions.Fixed(max); v < 0 || v >= max {
			t.Fatalf("longueur tirée hors de [0, %v) : %v", max.Float(), v.Float())
		}
	}
	for _, degenere := range []Fixed{0, -1, FromInt(-4)} {
		if v := s.Positions.Fixed(degenere); v != 0 {
			t.Errorf("borne %v rend %v, attendu 0", degenere.Float(), v.Float())
		}
	}
}

// TestPickEnsembleVide vérifie qu'un ensemble vide rend zéro sans paniquer.
func TestPickEnsembleVide(t *testing.T) {
	s := NewStreams(5)
	for _, n := range []int{0, -1} {
		if v := s.Loot.Pick(n); v != 0 {
			t.Errorf("Pick(%d) rend %d, attendu 0", n, v)
		}
	}
}
