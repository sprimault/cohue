// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce que cohue ajoute aux entrées du moteur : une commande peut se donner par
// plusieurs touches, et c'est la seule chose qui manque à ce qu'il rend.

package render

import "github.com/sprimault/lozengine/input"

// Input est l'état des entrées pour l'image en cours, tel que le moteur le rend.
//
// **Le type du moteur et non le nôtre.** `internal/render` en portait un, écrit
// pour recevoir les entrées en valeur plutôt que de sonder les fonctions
// globales d'Ebitengine ; `lozengine` rend exactement cela — un état et non une
// file, une valeur qu'on lit sans pouvoir la modifier, et une valeur zéro qui ne
// tient aucune touche. Le garder en aurait fait la seconde description que
// `docs/go.md` proscrit, avec une conversion à écrire à chaque image.
//
// L'alias plutôt qu'un emprunt direct dans chaque signature : ce que les
// méthodes de dessin reçoivent est « l'entrée de cette image », et le mot vaut
// mieux que le chemin du paquet répété quarante fois.
type Input = input.Input

// tenue dit si l'une des touches est tenue à cette image.
//
// **Plusieurs touches pour une commande, et c'est tout ce que cohue ajoute.** Le
// moteur répond sur une touche, ce qui est la bonne granularité pour lui ; ici
// une commande a plusieurs places — les flèches ou ZQSD, Entrée ou Espace ou
// l'Entrée du pavé —, et c'est une décision de jeu que la conception porte : le
// joueur qui appuie sur le chiffre qu'il a sous les doigts n'a aucun moyen de
// savoir que le jeu écoutait l'autre.
func tenue(in Input, touches ...input.Key) bool {
	for _, t := range touches {
		if in.Held(t) {
			return true
		}
	}
	return false
}

// pressee dit si l'une des touches vient d'être enfoncée, et une seule fois par
// appui.
//
// **Les deux régimes ne sont pas interchangeables.** L'état maintenu sert au
// déplacement et au forçage d'un obstacle, qui « se tient, ne se martèle pas » ;
// le front montant sert à ce qui se consomme — une charge, une fiole, la
// validation d'une carte, la relance. Au maintien, une charge partirait à l'image
// où le doigt se pose et le joueur ne saurait jamais s'il l'a dépensée exprès.
func pressee(in Input, touches ...input.Key) bool {
	for _, t := range touches {
		if in.JustPressed(t) {
			return true
		}
	}
	return false
}
