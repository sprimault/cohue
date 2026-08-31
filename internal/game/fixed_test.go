// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas de la virgule fixe : l'arrondi au plus proche et sa symétrie, la
// dérive qu'une troncature produirait sur une run, l'absence de débordement au
// produit, et la saturation d'une valeur hors plage.

package game

import (
	"math"
	"testing"
)

// TestMulRoundsToNearest vérifie que le produit arrondit au plus proche et non
// vers l'infini négatif.
//
// C'est la propriété dont dépend l'absence de dérive : le décalage arithmétique
// arrondit toujours vers le bas, et une demi-unité perdue à chaque opération
// finit par déplacer une entité.
func TestMulRoundsToNearest(t *testing.T) {
	cas := []struct {
		nom     string
		a, b    Fixed
		attendu Fixed
	}{
		{"un demi par un demi", One / 2, One / 2, One / 4},
		{"l unite est neutre", 12345, One, 12345},
		{"une demi-unite monte", 1, One / 2, 1},
		{"un quart d unite descend", 1, One / 4, 0},
		{"le signe ne change pas l ecart", -1, One / 2, -1},
	}
	for _, c := range cas {
		if obtenu := c.a.Mul(c.b); obtenu != c.attendu {
			t.Errorf("%s : %d.Mul(%d) = %d, attendu %d", c.nom, c.a, c.b, obtenu, c.attendu)
		}
	}
}

// TestMulErreurBorneeADemiUnite vérifie que le produit ne s'écarte jamais de
// plus d'une demi-unité du résultat exact.
//
// C'est la propriété qui justifie l'arrondi, et la seule qu'il puisse tenir : la
// quantification, elle, ne s'élimine pas — une longueur qui n'est pas un
// multiple de l'échelle ne se représente pas. Ce que l'arrondi divise par deux,
// c'est le pire cas d'une opération, donc la dérive d'une run entière.
func TestMulErreurBorneeADemiUnite(t *testing.T) {
	valeurs := []Fixed{1, 3, 677, 12345, One / 3, One / 2, One, One * 7, One * 100}
	for _, a := range valeurs {
		for _, b := range valeurs {
			for _, signe := range []Fixed{1, -1} {
				exact := float64(a*signe) * float64(b) / float64(One)
				ecart := math.Abs(float64((a * signe).Mul(b)) - exact)
				if ecart > 0.5 {
					t.Errorf("%d × %d : écart de %v unité", a*signe, b, ecart)
				}
			}
		}
	}
}

// TestMulDeriveMoinsQueLaTroncature compare l'arrondi à ce qu'un décalage nu
// aurait donné, sur une run de quinze minutes.
//
// L'écart se voit en jeu : une entité dont la vitesse se quantifie mal accumule
// le même biais à chaque pas, toujours dans le même sens. Ce test dit de combien
// l'arrondi le réduit, et il échouerait si quelqu'un retirait la correction en
// croyant simplifier.
func TestMulDeriveMoinsQueLaTroncature(t *testing.T) {
	const pas = TPS * 60 * 15 // une run de quinze minutes

	// Ce que ferait un décalage sans correction : l'arrondi vers l'infini
	// négatif, tel que Go l'applique.
	tronque := func(a, b Fixed) Fixed {
		return Fixed((int64(a) * int64(b)) >> fractionBits)
	}

	for _, tuilesParSeconde := range []float64{0.45, 0.62, 0.82, 1.0, 1.35} {
		vitesse := FromFloat(tuilesParSeconde)
		dt := One / TPS

		var arrondi, tronquee int64
		for i := 0; i < pas; i++ {
			arrondi += int64(vitesse.Mul(dt))
			tronquee += int64(tronque(vitesse, dt))
		}
		exact := int64(vitesse) * int64(dt) * pas / int64(One)

		ecartArrondi := abs64(arrondi - exact)
		ecartTronque := abs64(tronquee - exact)
		if ecartArrondi > ecartTronque {
			t.Errorf("à %v tuile/s, l'arrondi dérive de %d unités contre %d pour la troncature",
				tuilesParSeconde, ecartArrondi, ecartTronque)
		}
	}
}

// abs64 rend la valeur absolue d'un entier, que la bibliothèque standard
// n'offre que sur les flottants.
func abs64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

// TestMulNeDeborde vérifie que le produit passe par int64.
//
// Deux longueurs de cent tuiles tiennent dans un int32 ; leur produit non, et
// une multiplication écrite en int32 rendrait un résultat de signe arbitraire.
func TestMulNeDeborde(t *testing.T) {
	cent := FromInt(100)
	if obtenu := cent.Mul(cent); obtenu != FromInt(10000) {
		t.Errorf("100 × 100 = %v tuiles, attendu 10000", obtenu.Float())
	}
}

// TestDivArrondiSymetrique vérifie que la division arrondit de la même façon des
// deux côtés de zéro.
//
// La division entière de Go tronque vers zéro : sans correction par le signe, un
// quotient négatif s'arrondirait dans l'autre sens que son opposé, et une entité
// qui va vers la gauche n'avancerait pas comme celle qui va vers la droite.
func TestDivArrondiSymetrique(t *testing.T) {
	for _, n := range []Fixed{1, 3, 7, 12345, One, One * 100} {
		positif := n.Div(FromInt(3))
		negatif := (-n).Div(FromInt(3))
		if positif != -negatif {
			t.Errorf("%d/3 = %d mais -%d/3 = %d : l arrondi n est pas symétrique",
				n, positif, n, negatif)
		}
	}
}

// TestConversionsAllerRetour vérifie qu'une valeur de manifeste survit à sa
// conversion à une unité près.
//
// À une unité près et non à l'identique : l'échelle est de 65536 par tuile, et
// une valeur qui n'en est pas un multiple ne se représente pas. Une emprise de
// 0,18 tuile devient 11796 unités, soit 0,17999. Exiger l'égalité stricte
// reviendrait à tester la représentabilité du décimal, pas la conversion.
func TestConversionsAllerRetour(t *testing.T) {
	for _, tuiles := range []float64{0, 0.156, 0.18, 0.5, 1, 1.5, 16, 128.25} {
		obtenu := FromFloat(tuiles).Float()
		if math.Abs(obtenu-tuiles) > 1/float64(One) {
			t.Errorf("%v tuile devient %v, au-delà d'une unité d'écart", tuiles, obtenu)
		}
	}
}

// TestFromFloatSatureHorsPlage vérifie qu'une valeur de manifeste trop grande
// sature au lieu de changer de signe.
//
// Le cas est atteignable : rien n'empêche d'écrire quarante mille dans un
// champ de tuiles, et la conversion directe rendait alors le plus petit int32 —
// une distance positive devenue la plus grande distance négative. Pire, la
// spécification Go laisse ce résultat à l'implémentation : deux binaires publiés
// divergeraient sur la même donnée.
func TestFromFloatSatureHorsPlage(t *testing.T) {
	cas := map[float64]Fixed{
		40000:  math.MaxInt32,
		-40000: math.MinInt32,
		1e9:    math.MaxInt32,
		-1e9:   math.MinInt32,
	}
	for tuiles, attendu := range cas {
		if obtenu := FromFloat(tuiles); obtenu != attendu {
			t.Errorf("FromFloat(%v) = %d, attendu la saturation à %d",
				tuiles, obtenu, attendu)
		}
	}
}

// TestDivSatureAuLieuDeDeborder vérifie qu'une division par une longueur
// minuscule sature au lieu de changer de signe.
//
// C'est le seul débordement atteignable en jeu : normaliser un vecteur presque
// nul divise par presque rien. Sans la borne, le résultat repasserait par le
// négatif et l'entité se téléporterait à l'autre bout de la carte — un symptôme
// qu'on ne relie jamais à sa cause.
func TestDivSatureAuLieuDeDeborder(t *testing.T) {
	cas := []struct {
		nom     string
		a, b    Fixed
		attendu Fixed
	}{
		{"positif", FromInt(100), 1, math.MaxInt32},
		{"négatif", FromInt(-100), 1, math.MinInt32},
		{"diviseur négatif", FromInt(100), -1, math.MinInt32},
	}
	for _, c := range cas {
		if obtenu := c.a.Div(c.b); obtenu != c.attendu {
			t.Errorf("%s : %d/%d = %d, attendu la saturation à %d",
				c.nom, c.a, c.b, obtenu, c.attendu)
		}
	}
}

// TestFloorVersLeBas vérifie que l'index de case ne dépend pas du côté de
// l'origine.
func TestFloorVersLeBas(t *testing.T) {
	cas := map[Fixed]int{
		0: 0, One / 2: 0, One: 1, One*3 + One/4: 3,
		-One / 2: -1, -One: -1, -One - 1: -2,
	}
	for valeur, attendu := range cas {
		if obtenu := valeur.Floor(); obtenu != attendu {
			t.Errorf("Floor(%v) = %d, attendu %d", valeur.Float(), obtenu, attendu)
		}
	}
}
