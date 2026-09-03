// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le choix de la montée de niveau : les trois cartes offertes, ce qui les
// compose, et ce qu'appliquer l'une d'elles change à l'arme.

package game

// Choices est le nombre de cartes offertes à chaque montée.
//
// **Trois, et ce n'est pas un réglage.** Le chapitre 2 en fait une règle du
// genre : le choix compte plus que la récompense, et trois cartes dont deux
// tentantes est ce qui fait hésiter. Deux ne se compare pas, quatre se lit trop
// lentement pour une pause qui doit durer une seconde. Un champ de manifeste
// inviterait à la bouger, et la bouger changerait la nature du moment plutôt que
// son équilibrage.
const Choices = 3

// Card est ce qu'une place de l'écran de montée propose.
//
// Elle porte son texte tout composé plutôt que de quoi le composer : ce que le
// rendu doit savoir d'une carte, c'est trois lignes à poser, et lui faire
// connaître les unités de chaque axe lui donnerait une seconde description du
// contenu de la table.
type Card struct {
	// Name est le nom de fiction.
	Name string
	// Effect est ce que la carte donne, en une ligne.
	Effect string
	// Phrase est ce qu'elle fait, en clair.
	Phrase string

	// axe est l'index dans `Passives.Axes`, ou -1 pour la soupape. Il ne sort
	// pas du paquet : l'appelant choisit une place, pas un effet.
	axe int
}

// Pending rend les cartes offertes, vide quand aucun choix n'est ouvert.
//
// La tranche est celle du monde et se réécrit à chaque choix : la parcourir
// après un appel à `Choose` n'a pas de sens, et rien n'en garde une copie.
func (w *World) Pending() []Card { return w.cartes }

// Choosing dit si un choix attend le joueur.
//
// **La pause est réelle, et c'est l'écran qui la tient**, comme il tient celle
// de la mort. La boucle ne se fige pas d'elle-même : ce qu'une pause suspend et
// ce qu'elle laisse courir est une décision d'affichage, et la simulation qui la
// prendrait la rendrait invérifiable.
func (w *World) Choosing() bool { return len(w.cartes) > 0 }

// Choose applique la carte de rang donné et ferme le choix.
//
// Un rang hors des cartes offertes ne fait rien : l'appelant est un clavier, et
// une touche pressée au moment où l'écran se ferme ne doit pas arrêter le jeu.
//
// Le choix suivant s'ouvre dans la foulée quand plusieurs montées se sont
// accumulées — une récolte abondante en donne deux d'un coup, et les présenter
// l'une après l'autre est la seule façon de ne pas en perdre une.
func (w *World) Choose(rang int) {
	if rang < 0 || rang >= len(w.cartes) {
		return
	}
	w.appliquer(w.cartes[rang])
	w.cartes = w.cartes[:0]

	if w.enAttente > 0 {
		w.enAttente--
		w.offrir()
	}
}

// appliquer porte l'effet d'une carte sur la partie.
//
// L'arme est une copie que le monde tient : la modifier ne touche pas la table
// du manifeste, si bien qu'une relance repart de l'arme neuve sans qu'on ait à
// défaire quoi que ce soit.
func (w *World) appliquer(c Card) {
	if c.axe < 0 {
		// La soupape ne dépasse jamais le maximum : un soin qui déborderait
		// donnerait une jauge pleine à un joueur qui n'a rien de plus, et la
		// carte cesserait d'être ignorable quand on est haut.
		w.vie = min(w.vie+w.passifs.Relief.Heal, w.profils.Player.Health)
		return
	}

	axe := &w.passifs.Axes[c.axe]
	w.paliers[c.axe]++
	switch axe.Axis {
	case AxisCadence:
		w.arme.Cooldown -= axe.CooldownStep
	case AxisRange:
		w.arme.Range += axe.RangeStep
	}
}

// offrir compose les cartes de la montée en cours.
//
// **Aucun tirage n'a lieu, et c'est une conséquence du nombre d'axes plutôt
// qu'une décision.** Deux axes plus la soupape font exactement trois places :
// il n'y a rien à choisir parmi les éligibles, donc rien à tirer au sort. Le
// jour où un troisième axe entre, le pool dépasse les trois places et cette
// fonction prend un flux — qui n'est pas déclaré aujourd'hui, parce qu'un flux
// que personne ne lit est un numéro réservé pour rien.
//
// La soupape complète, et elle se répète autant qu'il faut : c'est ce qui
// garantit qu'aucune place ne reste vide, y compris quand tous les axes sont
// épuisés.
func (w *World) offrir() {
	w.cartes = w.cartes[:0]

	for i := range w.passifs.Axes {
		if len(w.cartes) == Choices {
			break
		}
		if w.paliers[i] < w.passifs.Axes[i].Tiers {
			w.cartes = append(w.cartes, carte(&w.passifs.Axes[i], i, w.paliers[i]+1))
		}
	}
	for len(w.cartes) < Choices {
		w.cartes = append(w.cartes, soupape(w.passifs.Relief))
	}
}

// carte compose la carte d'un palier d'axe.
//
// **La ligne d'effet dit le palier atteint et la borne, pas la grandeur du
// gain.** Une grandeur demanderait une unité par axe que la table ne déclare
// pas, et « moins trente-trois millisecondes » ne dit rien à qui joue. Le rang
// sur la borne, lui, dit ce que le joueur ne peut pas déduire autrement : ce
// qu'il reste sur cet axe, alors que l'épuiser est un moment de jeu.
// La ligne est lue dans la table plutôt que composée ici : un tick qui ouvre un
// choix est un tick comme un autre, et le budget d'allocation ne connaît pas
// d'exception pour les ticks rares.
func carte(axe *Passive, index, palier int) Card {
	return Card{
		Name:   axe.Name,
		Effect: axe.Effects[palier-1],
		Phrase: axe.Phrase,
		axe:    index,
	}
}

// soupape compose la carte de secours.
func soupape(r Relief) Card {
	return Card{
		Name:   r.Name,
		Effect: r.Effect,
		Phrase: r.Phrase,
		axe:    -1,
	}
}
