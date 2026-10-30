// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// L'écran de montée de niveau : les trois cartes en bas de l'image, la touche
// qui choisit, et la pause qu'ils tiennent le temps du choix.

package render

import (
	"fmt"

	"github.com/sprimault/lozengine/input"

	"github.com/sprimault/cohue/internal/game"
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
// Entrée du pavé numérique en fait partie : une `input.Key` désigne une place
// sur le clavier, et ce sont bien deux places différentes pour un même geste.
var validation = []input.Key{input.KeyEnter, input.KeyNumpadEnter, input.KeySpace}

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
func (s *Screen) choisir(in Input) {
	dernier := len(s.monde.Pending()) - 1
	if dernier < 0 {
		return
	}

	if in.JustPressed(input.KeyLeft) && s.carteChoisie > 0 {
		s.carteChoisie--
	}
	if in.JustPressed(input.KeyRight) && s.carteChoisie < dernier {
		s.carteChoisie++
	}

	if pressee(in, validation...) {
		s.monde.Choose(s.carteChoisie)
		s.carteChoisie = 0
		return
	}

	s.clicSurUneCarte(in)
}

// clicSurUneCarte prend la carte qu'on désigne du doigt, sans confirmation.
//
// **Le clic décide seul, là où la flèche désigne et Entrée valide**, et ce n'est
// pas une incohérence : la flèche a été séparée de la prise parce qu'elle sert
// aussi à courir, si bien qu'un panneau qui s'ouvre sous des doigts déjà posés
// prenait une carte avant qu'on l'ait lue. Un clic ne souffre pas de ce
// voisinage — on ne clique pas une carte par inadvertance —, et une carte mal
// choisie se rattrape au niveau suivant.
//
// **La désignation suit le clic** avant de repartir à zéro : si le choix ouvre
// autre chose plus tard, c'est la carte prise que la surbrillance doit avoir
// montrée, et non celle que les flèches avaient laissée.
func (s *Screen) clicSurUneCarte(in Input) {
	if s.hud == nil || !in.ButtonJustPressed(input.ButtonLeft) {
		return
	}

	_, zones := s.dispositionDesCartes(s.monde.Pending())
	for rang, z := range zones {
		if !s.souris.survole(z) {
			continue
		}
		s.carteChoisie = rang
		s.monde.Choose(rang)
		s.carteChoisie = 0
		return
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
func (s *Screen) poserCartes() {
	if s.hud == nil {
		return
	}
	cartes := s.monde.Pending()
	if len(cartes) == 0 {
		return
	}

	h := s.hud
	marge, ligne := h.Margin(), h.Font.Height()
	panneau, zones := s.dispositionDesCartes(cartes)

	// Le panneau se pose contre le bas de l'aire de jeu, pas contre le bas de la
	// surface : il forme un bloc avec elle, comme le bandeau au-dessus.
	h.Band(s.scene, panneau.x, panneau.y, panneau.largeur, panneau.hauteur)

	// **L'instruction est dans le titre, pas sous les cartes.** La règle qui veut
	// qu'un libellé porte sa touche vise ce qu'on lit *en jouant*, donc sous
	// pression et sans phrase disponible ; un panneau qui met le jeu en pause
	// n'est pas dans ce régime, et il a un titre où l'écrire une fois pour les
	// trois. Elle est utile ici parce que rien sur les cartes ne dit qu'on choisit
	// au clavier.
	titre := fmt.Sprintf("Niveau %d %c les flèches désignent, Entrée ou un clic valide",
		s.monde.Level(), '—')
	h.Font.Draw(s.scene, titre,
		panneau.x+(panneau.largeur-h.Font.Advance(titre))/2, panneau.y+marge,
		h.Color("texte"))

	for rang, z := range zones {
		c := cartes[rang]
		h.Frame(s.scene, z.x, z.y, z.largeur, z.hauteur)

		// Le survol d'abord, la désignation par-dessus : c'est elle qui décide,
		// et une carte à la fois désignée et survolée doit se lire comme
		// désignée.
		if s.souris.survole(z) {
			h.Outline(s.scene, z.x, z.y, z.largeur, z.hauteur, h.Color("survol"))
		}
		if rang == s.carteChoisie {
			// Le bord, et non un fond : le fond d'un cadre laisse deviner la
			// horde qui approche, et l'assombrir pour désigner une carte
			// retirerait au joueur ce qu'il doit lire pendant qu'il choisit.
			h.Outline(s.scene, z.x, z.y, z.largeur, z.hauteur, h.Color("cadre_choisi"))
		}

		texte, courante := z.x+marge+h.Border(), z.y+marge+h.Border()
		h.Font.Draw(s.scene, c.Name, texte, courante, h.Color("texte"))
		h.Font.Draw(s.scene, c.Effect, texte, courante+ligne+interligne,
			h.Color("texte_valeur"))
		h.Font.Draw(s.scene, c.Phrase, texte, courante+2*(ligne+interligne),
			h.Color("texte_attenue"))
	}
}

// dispositionDesCartes rend le rectangle du panneau et celui de chaque carte.
//
// **Une seule formule pour la pose et pour le clic**, comme pour les cases du
// bandeau : les recalculer ailleurs serait la seconde description que
// `docs/go.md` proscrit, et c'est le clic qui tomberait à côté le jour où la
// marge change. Le panneau en fait partie parce que le titre s'y centre : le
// séparer ferait deux fonctions qui recalculent la même ordonnée.
//
// **Toutes les cartes ont la largeur de la plus longue ligne de l'ensemble**, et
// non chacune la sienne : trois largeurs différentes se liraient comme trois
// importances différentes, alors que le choix est entre égaux. C'est aussi ce
// qui rend la rangée régulière, donc exprimable en un pas constant.
//
// **Elle alloue une tranche par ouverture de panneau**, ce que le budget admet :
// le choix met le jeu en pause, donc on n'est pas dans la boucle de mise à jour.
func (s *Screen) dispositionDesCartes(cartes []game.Card) (panneau zone, zones []zone) {
	h := s.hud
	marge, ligne := h.Margin(), h.Font.Height()
	hauteurCarte := 3*(ligne+interligne) + 2*(marge+h.Border())

	ax, ay, alargeur, ahauteur := aireDans(s.scene)
	hauteurPanneau := 2*marge + ligne + marge + hauteurCarte + marge
	panneau = zone{
		x: ax, y: ay + ahauteur - hauteurPanneau,
		largeur: alargeur, hauteur: hauteurPanneau,
	}

	largeur := 0
	for _, c := range cartes {
		for _, l := range []string{c.Name, c.Effect, c.Phrase} {
			largeur = max(largeur, h.Font.Advance(l))
		}
	}
	largeur += 2 * (marge + h.Border())

	x := ax + (alargeur-len(cartes)*largeur-(len(cartes)-1)*marge)/2
	y := panneau.y + 2*marge + ligne

	zones = make([]zone, len(cartes))
	for rang := range zones {
		zones[rang] = zone{
			x: x + rang*(largeur+marge), y: y,
			largeur: largeur, hauteur: hauteurCarte,
		}
	}
	return panneau, zones
}
