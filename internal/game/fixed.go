// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Package game porte la simulation : bassins d'entités, champ de flux, profils
// et boucle de mise à jour.
//
// Il n'importe ni Ebitengine, ni `time` : une run se joue en mémoire, sans
// fenêtre, sur un nombre de ticks fixé. C'est ce qui rend le rejeu testable et
// l'équilibrage comparable d'une version à l'autre.
package game

import "math"

// Fixed est une longueur du monde, en virgule fixe, dont l'unité est la tuile.
//
// Pas un flottant : la spécification Go autorise une implémentation à fusionner
// une multiplication et une addition en une seule opération arrondie une fois,
// et arm64 le fait là où amd64 ne le fait pas. Deux binaires publiés
// divergeraient alors sur la même graine, ce qui viderait le déterminisme de la
// run de sa substance — et avec lui le classement par graine et le partage d'un
// défi.
//
// Pas un alias non plus. `type Fixed = int32` laisserait `a * b` compiler sans
// rien dire, alors que le produit de deux longueurs demande une remise à
// l'échelle. C'est la seule vraie source d'erreur du procédé, et le seul
// garde-fou est que le compilateur refuse la multiplication nue.
//
// La frontière s'arrête à ce paquet : `internal/render` convertit en pixels et
// calcule ce qu'il veut en flottants, puisque rien de ce qu'il produit ne
// revient dans la simulation.
type Fixed int32

const (
	// fractionBits fixe l'échelle à 65536 unités par tuile. Puissance de deux,
	// donc la remise à l'échelle est un décalage et non une division, et la
	// plage d'un int32 couvre ±32768 tuiles — deux ordres de grandeur au-dessus
	// du plus grand lieu concevable.
	fractionBits = 16

	// One est la tuile, l'unité du monde.
	One Fixed = 1 << fractionBits

	// half sert à l'arrondi au plus proche. Le décalage arithmétique arrondit
	// vers l'infini négatif : sans cette correction, chaque opération biaiserait
	// d'une demi-unité toujours dans le même sens, ce qui fait de l'ordre de
	// quatre dixièmes de tuile de dérive sur une run de quinze minutes. Le rejeu
	// resterait déterministe, et faux.
	half = One / 2
)

// borner ramène un résultat intermédiaire dans la plage d'une longueur.
//
// Saturer plutôt que déborder. Une longueur qui sort de ±32768 tuiles est déjà
// un défaut — aucun lieu n'approche cette taille —, mais un débordement
// silencieux la ferait réapparaître de l'autre côté de la carte, avec le signe
// opposé : une entité qui se téléporte se diagnostique bien plus mal qu'une
// entité collée au bord. Le cas atteignable est la division par une longueur
// minuscule, qu'une normalisation de vecteur presque nul produit.
func borner(n int64) Fixed {
	switch {
	case n > math.MaxInt32:
		return math.MaxInt32
	case n < math.MinInt32:
		return math.MinInt32
	}
	return Fixed(n) // #nosec G115 -- borné aux deux lignes précédentes
}

// FromInt rend la longueur d'un nombre entier de tuiles.
func FromInt(tuiles int) Fixed {
	return borner(int64(tuiles) * int64(One))
}

// FromFloat convertit une valeur lue dans un manifeste.
//
// Le flottant s'arrête ici : il vient d'un fichier, il est converti une fois au
// chargement, et la simulation n'en voit jamais.
func FromFloat(tuiles float64) Fixed {
	if tuiles < 0 {
		return Fixed(tuiles*float64(One) - 0.5)
	}
	return Fixed(tuiles*float64(One) + 0.5)
}

// Mul rend le produit de deux longueurs, arrondi au plus proche et symétrique
// autour de zéro.
//
// Par `int64` : deux valeurs d'une tuile tiennent dans un int32, leur produit
// non — 65536 × 65536 déborde de vingt fois.
//
// Le signe est traité à part, et il faut l'écrire parce que le contraire paraît
// vrai : ajouter la demi-échelle avant un décalage arithmétique donne bien
// l'arrondi au plus proche, mais il arrondit les demis exacts vers l'infini
// positif. Une demi-unité monterait alors à un vers la droite et retomberait à
// zéro vers la gauche — une entité dont la vitesse tombe pile sur un demi
// n'avancerait pas pareil dans les deux sens, d'une unité par pas.
func (a Fixed) Mul(b Fixed) Fixed {
	p := int64(a) * int64(b)
	if p < 0 {
		return borner(-((-p + int64(half)) >> fractionBits))
	}
	return borner((p + int64(half)) >> fractionBits)
}

// Div rend le quotient de deux longueurs, arrondi au plus proche.
//
// La correction dépend du signe du résultat, contrairement à Mul : la division
// entière de Go tronque vers zéro là où le décalage arrondit vers l'infini
// négatif, et un demi-diviseur ajouté sans regarder les signes arrondirait dans
// le mauvais sens pour un quotient négatif.
func (a Fixed) Div(b Fixed) Fixed {
	n, d := int64(a)<<fractionBits, int64(b)
	if (n < 0) != (d < 0) {
		n -= d / 2
	} else {
		n += d / 2
	}
	return borner(n / d)
}

// Abs rend la valeur absolue.
func (a Fixed) Abs() Fixed {
	if a < 0 {
		return -a
	}
	return a
}

// Floor rend le nombre de tuiles entières, arrondi vers le bas.
//
// Vers le bas et non vers zéro : c'est l'index de la case qui contient le point,
// et une case ne change pas de numéro selon le côté de l'origine où elle se
// trouve.
func (a Fixed) Floor() int {
	return int(a >> fractionBits)
}

// Float rend la valeur en tuiles, pour le rendu et pour les messages.
//
// Jamais dans un calcul de simulation : ce serait rouvrir la porte que le type
// ferme.
func (a Fixed) Float() float64 {
	return float64(a) / float64(One)
}
