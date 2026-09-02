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
	return !s.monde.Alive() && presse(validation)
}

// peindreMort couvre l'écran et dit comment repartir.
//
// **Le monde reste visible dessous.** La conception veut que le joueur puisse se
// raconter sa mort : l'écran qui la masquerait retirerait ce qu'il y a à
// comprendre — où était la horde, par où il aurait pu passer.
//
// Deux lignes et rien d'autre. Ce qui l'a tué, ce qu'il allait obtenir et à
// combien il en était sont une extension, et supposent un score et des niveaux
// qui n'existent pas encore.
func (s *Screen) peindreMort(ecran *ebiten.Image) {
	if s.hud == nil {
		return
	}

	voile := s.hud.Color("bandeau_fond")
	s.hud.Rect(ecran, 0, 0, Width, Height, voile)

	titre, invite := "Mort", "Entrée pour relancer"
	hauteur := s.hud.Font.Height()
	y := Height/2 - hauteur

	s.hud.Font.Draw(ecran, titre, (Width-s.hud.Font.Advance(titre))/2, y,
		s.hud.Color("texte"))
	s.hud.Font.Draw(ecran, invite, (Width-s.hud.Font.Advance(invite))/2,
		y+2*hauteur, s.hud.Color("texte_attenue"))
}
