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

	// sorte dit d'où la carte vient, et index où la trouver dans sa table.
	//
	// **Une sorte et un index plutôt que deux index dont l'un vaut moins un.**
	// Depuis que les recettes entrent dans le menu, il y a trois provenances : un
	// axe, une fusion, la soupape. Deux champs à moins un rendraient l'état « ni
	// l'un ni l'autre » exprimable, et c'est ce qu'un champ oublié produirait.
	//
	// Ni l'une ni l'autre ne sort du paquet : l'appelant choisit une place, pas un
	// effet.
	sorte sorteDeCarte
	index int
}

// sorteDeCarte dit de quelle table une carte vient.
type sorteDeCarte uint8

const (
	// carteSoupape est la valeur zéro, et c'est délibéré : une `Card` bâtie sans
	// rien désigne la carte qui n'a pas d'index, jamais un axe.
	carteSoupape sorteDeCarte = iota
	carteAxe
	carteFusion
)

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
	switch c.sorte {
	case carteSoupape:
		// La soupape ne dépasse jamais le maximum : un soin qui déborderait
		// donnerait une jauge pleine à un joueur qui n'a rien de plus, et la
		// carte cesserait d'être ignorable quand on est haut.
		w.vie = min(w.vie+w.passifs.Relief.Heal, w.profils.Player.Health)
		return
	case carteFusion:
		w.fusions[c.index] = true
		switch w.passifs.Recipes[c.index].Effect {
		case EffectRail:
			w.arme.Rail = true
		case EffectSpray:
			w.arme.Spray = true
		}
		return
	}

	axe := &w.passifs.Axes[c.index]
	w.paliers[c.index]++
	switch axe.Axis {
	case AxisCadence:
		w.arme.Cooldown -= axe.CooldownStep
	case AxisRange:
		w.arme.Range += axe.RangeStep
	case AxisProjectiles:
		w.arme.Projectiles += axe.ProjectileStep
	case AxisPierce:
		w.arme.Pierce += axe.PierceStep
	case AxisBounce:
		w.arme.Bounces += axe.BounceStep
	case AxisSpread:
		w.arme.Spread += axe.SpreadStep
	}
}

// offrir compose les cartes de la montée en cours.
//
// **Le tirage ne se consomme que s'il départage.** Tant que les axes éligibles
// tiennent dans les trois places, ils sont tous offerts et rien n'est tiré : la
// table livrée en porte trois, donc aucune partie ne consomme `Cards`
// aujourd'hui. C'est au quatrième axe que le choix commence — la version
// précédente de cette godoc annonçait le troisième, et se trompait d'un cran :
// trois éligibles pour trois places ne laissent rien à choisir.
//
// Un tirage inconditionnel serait pire qu'inutile : il décalerait le flux sans
// qu'aucune décision en dépende, et deux graines identiques n'offriraient plus la
// même chose pour une raison qui n'est pas une règle de jeu.
//
// **Le mélange est partiel et se fait en place**, sur une tranche dont la
// capacité vient du montage : trois échanges suffisent à tirer trois éléments
// sans biais, et rien n'est alloué dans un tick qui ouvre un choix.
//
// La soupape complète, et elle se répète autant qu'il faut : c'est ce qui
// garantit qu'aucune place ne reste vide, y compris quand tous les axes sont
// épuisés.
func (w *World) offrir() {
	w.cartes = w.cartes[:0]
	w.candidats = w.candidats[:0]

	for i := range w.passifs.Axes {
		if w.paliers[i] < w.passifs.Axes[i].Tiers {
			w.candidats = append(w.candidats, carte(&w.passifs.Axes[i], i, w.paliers[i]+1))
		}
	}
	// **Les fusions concourent avec les axes plutôt que de passer devant.** Une
	// recette qui prendrait sa place d'office cesserait d'être un choix, et le
	// joueur la subirait au lieu de la préférer à ce qu'elle écarte.
	for i := range w.passifs.Recipes {
		if !w.fusions[i] && w.reunie(&w.passifs.Recipes[i]) {
			w.candidats = append(w.candidats, fusion(&w.passifs.Recipes[i], i))
		}
	}

	if len(w.candidats) > Choices {
		for k := range Choices {
			j := k + w.hasard.Cards.IntN(len(w.candidats)-k)
			w.candidats[k], w.candidats[j] = w.candidats[j], w.candidats[k]
		}
		w.candidats = w.candidats[:Choices]
	}

	w.cartes = append(w.cartes, w.candidats...)
	for len(w.cartes) < Choices {
		w.cartes = append(w.cartes, soupape(w.passifs.Relief))
	}
}

// reunie dit si les ingrédients d'une recette sont tous réunis.
//
// **Elle reste vraie tant qu'ils le sont**, ce qui garde la carte offerte au
// tirage suivant : celui qui préfère autre chose au moment où elle paraît ne perd
// pas la recette. C'est `w.fusions` qui la retire, une fois prise.
//
// L'axe d'un ingrédient se cherche par sa clé à chaque ouverture plutôt que
// d'être résolu au chargement. Six axes et deux ingrédients font une douzaine de
// comparaisons dans un tick qui arrive toutes les vingt secondes, et rien n'y est
// alloué — un index résolu d'avance serait une seconde description du rang d'un
// axe, qui change quand la table en gagne un.
func (w *World) reunie(r *Recipe) bool {
	for _, ingredient := range r.Ingredients {
		rang := -1
		for i := range w.passifs.Axes {
			if w.passifs.Axes[i].Axis == ingredient.Axis {
				rang = i
				break
			}
		}
		if rang < 0 {
			// Un axe inconnu est refusé au chargement : ici, il ne peut venir que
			// d'une table bâtie à la main dans un test, et la recette n'est alors
			// jamais offerte plutôt que d'ouvrir un index hors bornes.
			return false
		}

		exige := ingredient.Tiers
		if ingredient.Spent {
			exige = w.passifs.Axes[rang].Tiers
		}
		if w.paliers[rang] < exige {
			return false
		}
	}
	return true
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
		sorte:  carteAxe,
		index:  index,
	}
}

// fusion compose la carte d'une recette.
//
// **Sa ligne d'effet nomme la fusion et non un palier**, parce qu'elle n'en a
// qu'un : ce que le joueur doit lire est qu'une carte de cette sorte n'existait
// pas au tirage précédent.
func fusion(r *Recipe, index int) Card {
	return Card{
		Name:   r.Name,
		Effect: "Fusion",
		Phrase: r.Phrase,
		sorte:  carteFusion,
		index:  index,
	}
}

// soupape compose la carte de secours.
func soupape(r Relief) Card {
	return Card{
		Name:   r.Name,
		Effect: r.Effect,
		Phrase: r.Phrase,
		sorte:  carteSoupape,
	}
}
