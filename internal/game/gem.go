// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La gemme : ce qu'une créature laisse en mourant, la façon dont une volée
// s'étale au sol, et le ramassage à la portée du joueur.

package game

// rayonVolee est l'écart entre une gemme et le point de la mort, en tuiles.
//
// Un quart de tuile, ce qui n'est pas un réglage mais une conséquence : huit
// gemmes de dix pixels de large réparties tous les quarante-cinq degrés se
// chevauchent en deçà d'environ treize pixels de rayon, soit un cinquième de
// tuile. Un quart les tient sans les serrer.
//
// **Le critère est le chevauchement, et rien d'autre.** Que la figure ainsi
// dessinée soit lisible n'a jamais été jugé : aucune donnée livrée ne produit de
// volée, et un calcul qui écarte des disques ne dit rien d'un anneau bien fait.
// Qui trouvera la figure mauvaise ne doit donc pas commencer par ce chiffre —
// il est le seul de ceux qui décident de l'image à porter un nom et une
// justification, ce qui le rend lisible, pas coupable. Le nombre de gemmes, le
// rang zéro qui reste au centre et le pas de la table décident autant, en ligne
// dans `lacher`.
const rayonVolee = One / 4

// Gem est une gemme au sol.
//
// Elle ne porte pas ce qu'elle vaut : une gemme rapporte la même chose du début
// à la fin de la run, et c'est le seuil du niveau suivant qui monte. Loger la
// valeur ici en ferait un champ recopié autant de fois qu'il y a de gemmes, pour
// une information qui n'en a qu'une.
type Gem struct {
	X, Y Fixed
	// Born est le tick où elle est tombée.
	//
	// La naissance et non le reste à vivre : un compte à rebours demanderait une
	// écriture par gemme et par tick, là où une date se pose une fois. Et c'est
	// l'âge que le rendu veut — l'extinction progressive est une fraction de la
	// durée de vie, pas un seuil.
	Born Tick
	// Pulled dit que l'aimant l'a saisie.
	//
	// **Une gemme attirée cesse de vieillir.** Sans cela, déclencher l'aimant sur
	// un tas ancien en perdrait la moitié en vol : il échouerait précisément sur
	// les gemmes pour lesquelles on l'a dépensé. Il est le recours contre
	// l'effacement, il ne peut pas en être la victime.
	Pulled bool
}

// Gems rend le bassin des gemmes au sol.
func (w *World) Gems() *Pool[Gem] { return w.gemmes }

// lacher pose au sol ce que la mort d'une créature laisse.
//
// **Les gemmes tombent à l'endroit de la mort**, et non dispersées au hasard :
// le tas dit au joueur où il a tué, donc où revenir, et c'est ce lien qu'un
// tirage brouillerait. Il dit aussi ce qu'il y a à récolter, ce dont l'aimant a
// besoin pour qu'on estime sa prise avant de le déclencher.
//
// **Une volée s'étale par la table des huit orientations**, celle qui sépare
// déjà deux entités superposées dans le gradient de densité. C'est le même
// problème — plusieurs choses au même point exact —, donc la même réponse, et
// elle ne consomme aucun tirage. Le pas de trois est celui de `Vec.Direction`,
// pour la même raison : il écarte deux rangs consécutifs de cent trente-cinq
// degrés au lieu de quarante-cinq.
//
// **Ce pas a été choisi pour séparer deux poses voisines, pas pour dessiner un
// anneau.** Ce sont deux exigences distinctes qui se trouvent partager un seul
// paramètre, et rien ne le signale tant qu'elles s'accordent. Le jour où elles
// divergent, on ne pourra pas ajuster l'une sans déplacer l'autre : il faudra
// alors les séparer, et nommer l'arbitrage plutôt que déplacer le chiffre.
//
// **Le rayon croît d'un tour de table au suivant.** La neuvième gemme retombe
// exactement sur la première, la table n'ayant que huit entrées : sans cela, une
// volée de plus de huit poserait deux gemmes au même point, ce que l'étalement
// existe précisément pour éviter.
//
// Le bassin plein arrête la volée plutôt que d'allouer. Une gemme perdue est une
// perte invisible pour le joueur, et c'est pourquoi le plafond est large — ce
// qui borne vraiment le stock est l'effacement, qui retire une gemme au bout de
// sa durée de vie.
func (w *World) lacher(e *Enemy) {
	for rang := range w.profils.Enemies[e.Profile].Gems {
		x, y := e.X, e.Y
		// La première reste au point de la mort ; les suivantes s'écartent. Une
		// créature qui n'en laisse qu'une la pose donc exactement où elle est
		// tombée, ce qui est le cas de toutes aujourd'hui.
		if rang > 0 {
			ecart := Heading(rang * 3).Scale(rayonVolee * Fixed(1+rang/Headings))
			x, y = x+ecart.X, y+ecart.Y
		}
		if _, pose := w.gemmes.Spawn(Gem{X: x, Y: y, Born: w.tick}); !pose {
			return
		}
	}
}

// ramasser retire les gemmes que le joueur atteint et celles qui se sont
// éteintes, et rend le nombre des premières.
//
// **Deux causes, une seule suppression**, comme pour un projectile qui disparaît
// qu'il ait touché ou épuisé sa portée. Les écrire en deux passes finirait par en
// laisser une oublier de libérer sa place, et le nom de cette fonction dit son
// produit — la récolte — plutôt que la liste de ce qu'elle retire.
//
// **L'ordre entre les deux n'est pas indifférent : le ramassage passe en
// premier.** Une gemme atteinte au tick même où elle expire est ramassée, et
// c'est le sens qu'il faut : le joueur l'avait sous les pieds, la lui retirer
// pour une milliseconde ferait un vol que rien à l'écran n'expliquerait.
//
// Une gemme que l'aimant a saisie n'expire plus : la raison est sur `Gem.Pulled`.
//
// Le compte est rendu plutôt que porté à l'expérience ici : ce que vaut une
// gemme est une question de progression, et une passe de ramassage qui la
// trancherait aurait donné un second domicile au rythme des choix.
//
// La distance se mesure dans le plan du sol et non à l'écran : un rayon exprimé
// en pixels décrirait une ellipse dans le monde, et une gemme se ramasserait de
// plus loin vers l'est que vers le nord.
//
// La place libérée est réexaminée, comme dans le retrait des morts et pour la
// même raison : cette passe ne fait que filtrer, sans rien avancer, et la sauter
// laisserait au sol une gemme que le joueur a déjà traversée.
func (w *World) ramasser() int {
	portee := int64(w.progression.PickupRange)
	recoltees := 0
	for i := 0; i < w.gemmes.Len(); {
		g := w.gemmes.At(i)
		switch {
		case (Vec{X: g.X - w.playerX, Y: g.Y - w.playerY}).carres() <= portee*portee:
			w.gemmes.RemoveAt(i)
			recoltees++
		case !g.Pulled && w.tick-g.Born >= w.progression.GemLife:
			w.gemmes.RemoveAt(i)
		default:
			i++
		}
	}
	return recoltees
}

// GemAge rend l'âge d'une gemme, en ticks.
//
// Elle sort du paquet pour le rendu, qui en tire l'extinction progressive : une
// gemme s'éteint et ne clignote pas, parce que l'âge doit se lire en continu —
// c'est ce qui permet d'estimer sa récolte avant de déclencher l'aimant, et ce
// qui fait du déclenchement une lecture de la salle plutôt qu'un réflexe.
func (w *World) GemAge(g *Gem) Tick { return w.tick - g.Born }

// GemLife rend le temps qu'une gemme reste au sol, en ticks.
func (w *World) GemLife() Tick { return w.progression.GemLife }
