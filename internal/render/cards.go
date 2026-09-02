// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// L'écran de montée de niveau : les trois cartes en bas de l'image, la touche
// qui choisit, et la pause qu'ils tiennent le temps du choix.

package render

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// validation est la touche qui prend la carte désignée.
//
// **Désigner puis valider, et non prendre d'un coup.** La flèche prenait
// directement la carte de sa place, ce qui portait bien le geste et se jouait
// trop vite : à l'ouverture du panneau les doigts sont déjà sur les flèches pour
// courir, et le joueur voyait sa carte prise sans avoir eu le temps de lire
// laquelle. Séparer la désignation de la prise rend le choix délibéré, ce que la
// conception demande — choisir sous pression n'est pas un choix.
//
// Le conflit avec le déplacement tombe du même coup : une flèche ne fait plus
// que déplacer une surbrillance, et rien d'irréversible ne part d'une touche
// qu'on tenait pour une autre raison.
//
// **Espace et Entrée, les deux touches dont le sens est « je confirme ».** Elles
// valent l'une pour l'autre partout où le jeu attend un accord, y compris sur
// l'écran de mort : demander Espace ici et Entrée là ferait chercher laquelle,
// et chercher est la première des frictions que le chapitre 2 compte à la
// relance.
//
// Entrée du pavé numérique en fait partie : `ebiten.Key` désigne une place, et
// ce sont bien deux places différentes pour un même geste.
var validation = []ebiten.Key{ebiten.KeyEnter, ebiten.KeyNumpadEnter, ebiten.KeySpace}

// interligne est l'écart entre deux lignes d'une carte, en pixels.
//
// Deux pixels au-delà de la hauteur d'une ligne : les trois lignes d'une carte
// se lisent comme un bloc, et un interligne plus large les ferait lire comme
// trois textes qui se suivent.
const interligne = 2

// choisir déplace la désignation et prend la carte désignée.
//
// **Sur l'enfoncement et non sur le maintien**, pour la raison de l'écran de
// mort : la touche restant pressée d'une image à l'autre, un test d'état
// balaierait les trois cartes en trois images.
//
// **La désignation bute aux extrémités plutôt que de reboucler.** Trois cartes se
// parcourent d'un bout à l'autre sans qu'on ait à compter, et un rebouclage sur
// si peu d'éléments fait perdre le fil à qui appuie deux fois de suite.
//
// Elle revient à gauche après chaque prise : une récolte abondante ouvre deux
// panneaux d'affilée, et hériter de la désignation du premier ferait valider le
// second sur une carte qu'on n'a pas regardée.
func (s *Screen) choisir() {
	dernier := len(s.monde.Pending()) - 1
	if dernier < 0 {
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) && s.carteChoisie > 0 {
		s.carteChoisie--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) && s.carteChoisie < dernier {
		s.carteChoisie++
	}

	if presse(validation) {
		s.monde.Choose(s.carteChoisie)
		s.carteChoisie = 0
	}
}

// presse dit si l'une des touches vient d'être enfoncée.
func presse(touches []ebiten.Key) bool {
	for _, t := range touches {
		if inpututil.IsKeyJustPressed(t) {
			return true
		}
	}
	return false
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

	// Le titre, puis les cartes, chacun avec sa marge. Rien sous les cartes
	// depuis que l'instruction est passée dans le titre.
	panneau := 2*marge + ligne + marge + hauteurCarte + marge
	haut := Height - panneau
	h.Band(ecran, haut, panneau)

	// **L'instruction est dans le titre, pas sous les cartes.** La règle qui veut
	// qu'un libellé porte sa touche vise ce qu'on lit *en jouant*, donc sous
	// pression et sans phrase disponible ; un panneau qui met le jeu en pause
	// n'est pas dans ce régime, et il a un titre où l'écrire une fois pour les
	// trois. Elle est utile ici parce que rien sur les cartes ne dit qu'on choisit
	// au clavier.
	titre := fmt.Sprintf("Niveau %d %c les flèches désignent, Entrée valide",
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
		if rang == s.carteChoisie {
			// Le bord, et non un fond : le fond d'un cadre laisse deviner la
			// horde qui approche, et l'assombrir pour désigner une carte
			// retirerait au joueur ce qu'il doit lire pendant qu'il choisit.
			h.Outline(ecran, x, y, largeur, hauteurCarte, h.Color("cadre_choisi"))
		}

		texte, courante := x+marge+h.Border(), y+marge+h.Border()
		h.Font.Draw(ecran, c.Name, texte, courante, h.Color("texte"))
		h.Font.Draw(ecran, c.Effect, texte, courante+ligne+interligne,
			h.Color("texte_valeur"))
		h.Font.Draw(ecran, c.Phrase, texte, courante+2*(ligne+interligne),
			h.Color("texte_attenue"))

		x += largeur + marge
	}
}
