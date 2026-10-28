// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce que le clavier dit à une image, reçu en valeur plutôt que sondé.
//
// Le relevé appartient à la boucle, qui seule sait quand une image commence ;
// ce paquet ne fait que lire ce qu'elle lui passe.

package render

import "github.com/hajimehoshi/ebiten/v2"

// Input est l'état du clavier pour l'image en cours : les touches tenues, et
// celles qui viennent d'être enfoncées.
//
// **Une valeur reçue et non un sondage.** Les fonctions d'état d'Ebitengine sont
// globales, et c'est ce qui avait placé la lecture du clavier dans ce paquet :
// il n'y avait pas de valeur à passer, donc pas de raison de la passer. Le prix
// en était qu'une méthode de dessin pouvait lire une touche sans que sa signature
// le dise, et qu'aucun appelant ne pouvait décider ce qu'elle lirait.
//
// **Les deux régimes ne sont pas interchangeables.** L'état maintenu sert au
// déplacement et au forçage d'un obstacle, qui « se tient, ne se martèle pas » ;
// le front montant sert à ce qui se consomme — une charge, une fiole, la
// validation d'une carte, la relance. Au maintien, une charge partirait à l'image
// où le doigt se pose et le joueur ne saurait jamais s'il l'a dépensée exprès.
//
// Sa valeur zéro ne tient aucune touche, ce qui est l'entrée d'une image où rien
// n'est arrivé — et ce qu'un appelant passe quand il veut avancer la partie sans
// rien commander.
type Input struct {
	tenues, pressees []ebiten.Key
}

// NewInput rend l'entrée d'une image depuis les touches relevées.
//
// **Les deux tranches appartiennent à l'appelant et ne sont pas copiées.** Une
// entrée ne vaut que pour l'image en cours, donc la boucle réemploie ses tranches
// d'une image à la suivante sans allouer — ce que le budget d'allocation exige
// d'un chemin parcouru soixante fois par seconde. Garder l'entrée au-delà de son
// image la ferait changer sous celui qui la lit.
func NewInput(tenues, pressees []ebiten.Key) Input {
	return Input{tenues: tenues, pressees: pressees}
}

// Held dit si l'une des touches est tenue à cette image.
func (in Input) Held(touches ...ebiten.Key) bool {
	return contient(in.tenues, touches)
}

// JustPressed dit si l'une des touches vient d'être enfoncée, et une seule fois
// par appui.
func (in Input) JustPressed(touches ...ebiten.Key) bool {
	return contient(in.pressees, touches)
}

// contient dit si l'un des candidats est dans le relevé.
//
// Deux boucles plutôt qu'un ensemble : le relevé porte les quelques touches
// qu'une main tient, les candidats sont au plus quatre, et bâtir une table par
// image coûterait l'allocation que ce chemin n'a pas le droit de faire.
func contient(releve, candidats []ebiten.Key) bool {
	for _, t := range releve {
		for _, c := range candidats {
			if t == c {
				return true
			}
		}
	}
	return false
}
