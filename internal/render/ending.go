// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// L'écran de mort : ce qu'il montre, et la touche qui relance. Il ne remonte
// rien lui-même — c'est la session qui sait ce qu'une relance conserve.

package render

import "github.com/hajimehoshi/ebiten/v2"

// La relance emprunte les touches de validation, et c'est le même geste : le
// chapitre 2 compte quatre frictions à la relance, et la première est de devoir
// chercher quoi presser. Deux touches larges et connues plutôt qu'« une touche
// quelconque », les mêmes qu'ailleurs.

// WantsRestart dit si le joueur demande à rejouer.
//
// **Sur l'enfoncement et non sur le maintien.** La touche restant pressée d'une
// image à l'autre, un test d'état relancerait soixante fois par seconde tant que
// le doigt n'est pas levé — et la partie repartirait à chaque image, donc jamais
// vraiment.
//
// Elle ne rend vrai qu'une fois mort : pendant la partie, ces touches ne font
// rien hors du panneau de choix, qui met la mort hors de portée puisque `Update`
// traite l'une avant l'autre.
func (s *Screen) WantsRestart() bool {
	return s.monde.Over() && presse(validation)
}

// peindreFin couvre l'écran et dit comment repartir.
//
// **Le monde reste visible dessous.** La conception veut que le joueur puisse se
// raconter sa mort : l'écran qui la masquerait retirerait ce qu'il y a à
// comprendre — où était la horde, par où il aurait pu passer.
//
// Deux lignes et rien d'autre, parce que le reste est déjà là : `peindreBandeau`
// pose le bandeau sous le voile, et le niveau atteint comme le temps tenu se
// lisent au travers. Ce qui manque est le score, qui suppose l'enchaînement de
// salles — c'est lui qui donne un total, et le bonus de temps qui s'y oppose.
//
// **Les deux fins partagent cet écran, et seul le titre les sépare.** Sortir
// n'ouvre pas encore sur un lieu suivant — l'étape 8 apporte l'enchaînement, le
// temps mort et le choix de branche —, si bien qu'un écran propre à la sortie
// n'aurait rien de plus à montrer que celui-ci. Ce qui compte pour l'instant est
// que la fin se voie et se distingue.
func (s *Screen) peindreFin(ecran *ebiten.Image) {
	if s.hud == nil {
		return
	}

	voile := s.hud.Color("bandeau_fond")
	s.hud.Rect(ecran, 0, 0, Width, Height, voile)

	titre, invite := "Mort", "Entrée pour relancer"
	if s.monde.Escaped() {
		titre = "Sorti"
	}
	hauteur := s.hud.Font.Height()
	y := Height/2 - hauteur

	s.hud.Font.Draw(ecran, titre, (Width-s.hud.Font.Advance(titre))/2, y,
		s.hud.Color("texte"))

	// **L'invite en pleine teinte, comme le titre.** Atténuer range un texte au
	// second plan, ce qu'on fait d'une explication et non de la seule chose que
	// cet écran demande. Le chapitre 2 en fait la variable la plus lourde du
	// système : la décision de rejouer se prend en trois secondes, et « Entrée
	// pour relancer » doit être en évidence plutôt que rangé sous le résumé.
	// C'est le même défaut que le bandeau portait sur son niveau.
	s.hud.Font.Draw(ecran, invite, (Width-s.hud.Font.Advance(invite))/2,
		y+2*hauteur, s.hud.Color("texte"))
}
