// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// L'écran de montée de niveau : les trois cartes en bas de l'image, la touche
// qui choisit, et la pause qu'ils tiennent le temps du choix.

package render

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/sprimault/cohue/internal/game"
)

// choix sont les touches qui prennent une carte, dans l'ordre de l'écran.
//
// Les chiffres du rang principal et non le pavé numérique : c'est la rangée que
// tout clavier porte, et celle où la main gauche est déjà posée pendant qu'on
// joue de l'autre. La touche s'écrit sous chaque carte, parce que ce que le
// joueur cherche est ce qu'il doit presser.
var choix = [game.Choices]ebiten.Key{ebiten.Key1, ebiten.Key2, ebiten.Key3}

// interligne est l'écart entre deux lignes d'une carte, en pixels.
//
// Deux pixels au-delà de la hauteur d'une ligne : les trois lignes d'une carte
// se lisent comme un bloc, et un interligne plus large les ferait lire comme
// trois textes qui se suivent.
const interligne = 2

// choisir prend la carte que le joueur demande, s'il en demande une.
//
// **Sur l'enfoncement et non sur le maintien**, pour la raison de l'écran de
// mort : la touche restant pressée d'une image à l'autre, un test d'état
// prendrait les trois cartes en trois images.
func (s *Screen) choisir() {
	for rang, touche := range choix {
		if inpututil.IsKeyJustPressed(touche) {
			s.monde.Choose(rang)
			return
		}
	}
}

// peindreCartes pose le panneau de choix en bas de l'image.
//
// **En bas et sur toute la largeur, pas au milieu.** La horde figée reste
// visible au-dessus : la conception veut que la pause soit brève et sans
// cérémonie, et un cadre flottant qui masquerait le combat ferait du choix un
// écran plutôt qu'un moment. Le titre tient sur le bandeau du panneau plutôt que
// sur le décor, où il disparaîtrait dès que le sol s'éclaircit.
//
// Toutes les cartes ont la largeur de la plus longue ligne de l'ensemble, et non
// chacune la sienne : trois largeurs différentes se liraient comme trois
// importances différentes, alors que le choix est entre égaux.
func (s *Screen) peindreCartes(ecran *ebiten.Image) {
	if s.hud == nil {
		return
	}
	cartes := s.monde.Pending()
	if len(cartes) == 0 {
		return
	}

	h := s.hud
	marge, ligne := h.Margin(), h.Font.Height()
	hauteurCarte := 3*(ligne+interligne) + 2*(marge+h.Border())

	// La touche s'écrit sous la carte, donc le panneau la compte.
	panneau := 2*marge + ligne + marge + hauteurCarte + ligne + marge
	haut := Height - panneau
	h.Band(ecran, haut, panneau)

	titre := fmt.Sprintf("Niveau %d %c choisissez une amélioration",
		s.monde.Level(), '—')
	h.Font.Draw(ecran, titre, (Width-h.Font.Advance(titre))/2, haut+marge,
		h.Color("texte"))

	largeur := 0
	for _, c := range cartes {
		for _, l := range []string{c.Name, c.Effect, c.Phrase} {
			largeur = max(largeur, h.Font.Advance(l))
		}
	}
	largeur += 2 * (marge + h.Border())

	x := (Width - len(cartes)*largeur - (len(cartes)-1)*marge) / 2
	y := haut + 2*marge + ligne
	for rang, c := range cartes {
		h.Frame(ecran, x, y, largeur, hauteurCarte)

		texte, courante := x+marge+h.Border(), y+marge+h.Border()
		h.Font.Draw(ecran, c.Name, texte, courante, h.Color("texte"))
		h.Font.Draw(ecran, c.Effect, texte, courante+ligne+interligne,
			h.Color("texte_valeur"))
		h.Font.Draw(ecran, c.Phrase, texte, courante+2*(ligne+interligne),
			h.Color("texte_attenue"))

		touche := fmt.Sprintf("%d", rang+1)
		h.Font.Draw(ecran, touche, x+(largeur-h.Font.Advance(touche))/2,
			y+hauteurCarte+interligne, h.Color("texte_attenue"))
		x += largeur + marge
	}
}
