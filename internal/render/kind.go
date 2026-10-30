// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// D'où vient une chose à poser : le bassin qu'une place désigne. C'est un
// aiguillage de lecture, et plus un critère de rang.

package render

// sorte dit dans quel bassin une place se résout.
//
// Le rendu relit des entités venues de bassins différents, et une place ne veut
// rien dire sans savoir d'où elle vient.
//
// **Ce n'est plus un critère de tri, et l'ordre des valeurs ne décide donc de
// rien.** La clé de rang appartient au moteur, qui range sur la profondeur, la
// couche et l'étage ; ce que la conception demande à égalité de profondeur — le
// joueur par-dessus, les traces au sol en dessous — se dit par l'étage d'un
// aspect, et c'est là que les justifications de rang vivent désormais. Les
// garder ici en ferait une seconde description de ce que la scène porte déjà.
type sorte uint8

// Les bassins d'où une place se résout.
const (
	// sorteDecor est une case du décor qui dépasse du sol.
	sorteDecor sorte = iota
	// sorteObstacle est un obstacle fragile debout. Il a la nature du décor tant
	// qu'il tient : sa case bloque, et ce qui la longe se départage avec lui
	// comme avec un mur.
	sorteObstacle
	// sorteEpave est ce qu'une caisse a laissé : une trace aplatie au sol, qui
	// n'a jamais à cacher ce qui bouge.
	sorteEpave
	sorteEnnemi
	// sorteAmbiance est le décor mouvant. Un figurant passe devant et derrière
	// les créatures comme n'importe quel corps posé au sol.
	sorteAmbiance
	sorteTir
	sorteTirHorde
	sorteGemme
	sorteAimant
	// sorteCaisse est un objet posé au sol : elle coûte à traverser plutôt
	// qu'elle n'arrête, si bien qu'une créature lui passe dessus.
	sorteCaisse
	// sorteArmeAuSol est une arme lourde tombée d'une caisse. Le joueur marche
	// dessus pour la prendre.
	sorteArmeAuSol
	// sorteFioleAuSol est une fiole tombée d'une caisse, qui se ramasse de la
	// même façon.
	sorteFioleAuSol
	// sorteTourelle est une tourelle posée. On marche autour.
	sorteTourelle
	// sorteEclair n'est pas un bassin de plus, c'est un second quad sur celui des
	// créatures : la teinte qu'un état ajoute par-dessus le sprite.
	//
	// **Elle a sa propre correspondance parce qu'elle a sa propre poignée.** Deux
	// quads de la même créature sont deux entités pour la scène, et l'éclair
	// apparaît et disparaît sans que la créature bouge — c'est le solde de son
	// attache qui le retire quand l'état s'achève.
	sorteEclair
	sorteJoueur
	// sorteLisere n'est pas un bassin non plus : ce sont les huit quads de
	// l'aplat qui cernent le joueur, posés sous son sprite.
	//
	// **Huit et non un, et ce n'est pas un choix de confort.** Le liseré déborde
	// du dessin, donc il demanderait une planche dilatée d'un pixel ; or une
	// silhouette grossie déborderait aussi de sa case et s'y ferait couper, ce
	// que la godoc de `sprite.Outline` constate pour le tracé intérieur. Huit
	// copies décalées par l'ancre débordent sans rien perdre, et elles coûtent
	// huit poses par image pour une seule entité.
	sorteLisere
)
