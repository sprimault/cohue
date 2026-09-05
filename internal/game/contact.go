// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le contact : ce que la horde retire au joueur tant qu'elle le touche, et le
// plafond par seconde sans lequel un encerclement tuerait instantanément.

package game

// subir accumule les dégâts du contact et en retire des points de vie.
//
// **Les dégâts sont continus, pas un choc à l'impact.** Une créature collée
// retire sa part chaque tick tant qu'elle touche ; se dégager arrête la perte
// sans qu'aucun coup n'ait été « porté ». C'est ce qui rend la fuite payante à
// tout instant plutôt qu'entre deux frappes.
//
// **Le plafond vaut pour la somme, jamais par créature.** Vingt par seconde
// quel que soit le nombre d'ennemis au contact : sans lui, un encerclement à
// douze tue en une demi-seconde et le joueur ne peut pas se raconter sa mort.
// C'est la règle du chapitre 5 de la conception, et la seule de ce fichier qui
// ne soit pas mécanique.
func (w *World) subir() {
	if !w.Alive() {
		return
	}

	somme, chocs := 0, 0
	for i := range w.ennemis.Len() {
		e := w.ennemis.At(i)
		if !w.auContact(e) {
			continue
		}
		profil := &w.profils.Enemies[e.Profile]

		// **Une charge qui touche s'arrête là**, et c'est ce qui évite un
		// drapeau « a déjà frappé » : les dégâts de charge sont un choc unique,
		// alors que tout ce qui l'entoure se compte par seconde. La fin de course
		// est la seule chose qui puisse les rendre uniques sans état de plus.
		if e.Charging() {
			chocs += profil.ChargeDamage
			w.finirLaCharge(e, profil)
			continue
		}
		somme += profil.ContactDamage
	}
	if plafond := w.profils.Player.DamageCap; somme > plafond {
		somme = plafond
	}

	// **Le choc se retire de la vie, il n'entre pas dans l'accumulateur.** Ce que
	// celui-ci compte est un débit — des points par seconde, en points-ticks —,
	// alors qu'une charge est un montant qui tombe d'un coup : l'y verser
	// diviserait dix-huit points par soixante, et le choc vaudrait un tiers de
	// point.
	//
	// **Et il ignore le plafond, sans le relever.** Ce que le plafond rend
	// lisible est l'encerclement, trente corps collés dont on ne distingue pas la
	// part ; une charge, elle, a été annoncée puis manquée. Les plafonner
	// ensemble ferait qu'une meute de trois Molosses infligerait ce qu'un seul
	// inflige, et le télégraphe n'annoncerait plus rien.
	if chocs > 0 {
		w.blesser(chocs)
	}
	if somme == 0 {
		return
	}

	// **L'accumulateur compte en points-ticks, et c'est ce qui évite une
	// troisième échelle.** Vingt points par seconde à soixante ticks font un
	// tiers de point par tick, qu'aucun entier ne représente : passer la vie en
	// virgule fixe la ferait entrer dans le type des tuiles, dont les bornes et
	// l'échelle sont celles d'une position. En accumulant `points × ticks` et en
	// retirant un point chaque fois qu'on atteint `TPS`, la perte est exacte sur
	// une seconde et le reste se garde d'un tick à l'autre.
	w.degatsSubis += somme

	// La condition d'arrêt porte la vie autant que l'accumulateur, et c'est elle
	// qui interdit une vie négative — sans garde-fou à côté, qui serait du code
	// mort et qu'un test finirait par déclarer nécessaire.
	//
	// La vie à zéro **est** la mort, comme la résistance d'une créature : pas de
	// drapeau à synchroniser. Ce qui se déclenche une fois est la transition, et
	// c'est cette boucle qui la produit — l'écran et la relance s'y brancheront,
	// jamais sur un tick où la vie est déjà nulle.
	for w.degatsSubis >= TPS && w.vie > 0 {
		w.degatsSubis -= TPS
		w.vie--
	}
}

// blesser retire un montant de vie d'un coup, sans passer par l'accumulateur.
//
// **C'est la voie des dégâts qui ne sont pas un débit** : le choc d'une charge,
// le projectile d'une Buse, la déflagration d'une Baudruche. Trois cas de même
// forme — ce qui diffère entre eux est le test de portée, jamais l'application.
// L'accumulateur compte en points-ticks, si bien qu'y verser dix-huit points
// d'un coup les diviserait par soixante et vaudrait un tiers de point.
//
// Le plancher est nécessaire ici, là où il serait du code mort pour le contact
// continu : celui-ci retire un point à la fois sous condition de vie positive,
// quand un montant peut la dépasser.
func (w *World) blesser(points int) {
	w.vie = max(w.vie-points, 0)
}

// auContact dit si une créature touche le joueur.
//
// La comparaison se fait sur les carrés, pour n'extraire aucune racine dans la
// boucle.
func (w *World) auContact(e *Enemy) bool {
	portee := w.porteeContact(w.profils.Enemies[e.Profile].Radius)
	ecart := Vec{X: e.X - w.playerX, Y: e.Y - w.playerY}
	return ecart.carres() < int64(portee)*int64(portee)
}

// porteeContact rend la distance sous laquelle une créature de ce rayon touche
// le joueur.
//
// Les rayons viennent des profils et jamais d'une distance choisie : ce sont eux
// que le manifeste porte, et c'est la même mesure qui décide du contact et de la
// séparation.
func (w *World) porteeContact(rayon Fixed) Fixed {
	return w.profils.Player.Radius + rayon
}

// porteeBlocage rend la distance sous laquelle un corps solide et le joueur se
// refusent l'un l'autre ; `surLeJoueur` et `dansUnCorps` la lisent, chacune d'un
// bord de l'exclusion réciproque.
//
// **Elle se dérive de la portée de contact au lieu de se poser sur la même somme
// de rayons, et c'est ce qui fait tenir les deux règles ensemble.** Écrite comme
// elle, elle valait exactement la borne que le contact exige de franchir : le
// Vigile ne pouvait approcher qu'à la distance où le contact cesse, si bien
// qu'il publiait dix dégâts par seconde qu'il n'infligeait jamais.
//
// **Le retrait vaut un pas, et il se dérive plutôt qu'il ne se choisit.** La
// projection annule un axe entier au lieu de coller le mobile à sa borne, si
// bien qu'un arrêt tombe n'importe où dans le pas qui l'a précédé : un retrait
// plus court rendrait le contact dépendant de l'alignement des pas, c'est-à-dire
// vrai une fois sur deux. C'est le plus grand des deux pas, l'un ou l'autre
// pouvant être celui qui approche — et le coût du terrain ne fait que diviser
// une vitesse, donc aucun pas réel ne dépasse celui du profil.
func (w *World) porteeBlocage(profil *EnemyProfile) Fixed {
	return w.porteeContact(profil.Radius) - max(w.profils.Player.Speed, profil.Speed)
}

// Health rend les points de vie restants.
func (w *World) Health() int { return w.vie }

// MaxHealth rend la vie du profil, celle d'où la partie est partie.
//
// Elle sort du monde plutôt que de la table parce qu'une jauge a besoin des deux
// termes : l'afficheur qui irait chercher le maximum dans les profils tiendrait
// une seconde référence sur ce que la partie modifie, et une régénération de
// vie au-delà du maximum ne se verrait que d'un côté.
func (w *World) MaxHealth() int { return w.profils.Player.Health }

// Alive dit si la partie continue.
//
// Elle se lit plutôt qu'elle ne se retient : la vie à zéro est la mort, donc
// aucun état parallèle ne peut en diverger.
func (w *World) Alive() bool { return w.vie > 0 }

// InDanger dit si la vie est descendue sous le seuil d'alerte du profil.
//
// **Un état lu, jamais un événement retenu.** C'est ce qui distingue ce signal
// de celui qu'on avait d'abord écrit — un voile posé au contact —, et ce qui le
// rend incapable de battre : la vie ne remonte que par une fiole, si bien que le
// seuil se franchit rarement et dans un sens. Un retour attaché au contact
// clignotait à chaque créature qui frôle, sur toute la surface de l'écran.
//
// **Et il porte ce que la scène ne montre pas.** Qu'une créature touche se voit
// — elle est collée au personnage, au centre du regard ; qu'il ne reste que
// quelques points de vie ne se lit qu'en haut à gauche, là où l'on ne regarde
// pas en kitant. Un signal redit rarement ce qui est déjà à l'image.
//
// **La partie finie, le joueur n'est plus en danger : c'est fait.** L'écran de
// fin prend le relais, et laisser l'alerte allumée sous lui la ferait durer
// jusqu'à la relance.
//
// **Finie et non morte**, depuis qu'il y a deux issues. La vie seule éteignait
// l'alerte à la mort et la laissait allumée sur une sortie, si bien que le même
// écran se présentait rouge ou non selon la façon dont on y était arrivé —
// alors que ce qu'il montre est justement commun aux deux.
func (w *World) InDanger() bool {
	return !w.Over() && w.vie <= w.profils.Player.LowHealth
}
