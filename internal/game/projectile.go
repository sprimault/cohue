// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Projectile est un tir en vol : sa position, son pas, ce qui lui reste de
// portée et ce qu'il retire à ce qu'il touche. Second occupant d'un bassin, il
// n'a demandé aucune contrainte que l'ennemi n'ait demandée.

package game

// Projectile est un tir en vol.
//
// Il porte tout ce dont son déplacement a besoin et rien de plus : ni renvoi
// vers l'arme qui l'a tiré, ni vers sa cible. Une arme peut monter de niveau
// pendant qu'un de ses projectiles vole, et une cible peut mourir — un tir en
// vol ne doit dépendre ni de l'une ni de l'autre.
//
// C'est le second type à vivre dans un `Pool`, et il n'a rien demandé que
// `Enemy` n'ait demandé : pas de propriétaire, pas de durée de vie tenue par le
// bassin, donc aucune contrainte sur le paramètre de type. Ce qu'un projectile
// veut de plus vit dans le projectile.
type Projectile struct {
	// X et Y sont sa position dans le monde, en tuiles.
	X, Y Fixed
	// Step est ce dont il avance à chaque tick, direction et vitesse ensemble.
	Step Vec
	// Remaining est la distance qu'il peut encore parcourir avant d'avoir
	// épuisé la portée de son arme.
	Remaining Fixed
	// Hits est ce qu'il retire à ce qu'il touche, dans l'unité où s'exprime la
	// résistance des créatures.
	Hits int
	// Pierce est le nombre de créatures qu'il peut encore traverser.
	//
	// Il décide de la course : tant qu'il en reste, le projectile poursuit tout
	// droit après avoir frappé. C'est ce qui le sépare de `Bounces`, qui décide
	// de la fin de la course — et pourquoi la perforation passe d'abord quand un
	// projectile a les deux. L'ordre inverse ferait rebondir avant d'avoir
	// traversé, et la perforation ne servirait jamais à qui a pris les deux axes.
	Pierce int
	// Bounces est le nombre de fois qu'il peut encore repartir vers une autre
	// cible.
	//
	// Le rebond ne recharge pas la portée : `Remaining` continue de descendre,
	// si bien qu'un projectile ne peut rebondir que vers ce qu'il pouvait déjà
	// atteindre. Sans cela il deviendrait perpétuel dans une foule, chaque
	// créature en amenant une autre.
	Bounces int
	// LastHit désigne la dernière créature touchée, pour ne pas la retoucher.
	//
	// **Un projectile qui meurt à l'impact n'a jamais eu ce problème** : il
	// disparaissait avant de repasser. Celui qui perfore survit, et une créature
	// à trois touches survit elle aussi au premier coup — sans cette référence,
	// elle serait reprise au tick suivant tant qu'elle reste dans le segment.
	//
	// Une seule et non toutes : une tranche par projectile serait une allocation
	// dans la boucle, ce que l'invariant du budget refuse. Le cas qu'elle laisse
	// ouvert est celui d'une créature retrouvée deux touches plus loin, ce qui
	// demanderait qu'elle contourne le projectile pendant qu'il traverse sa
	// voisine — un mouvement que la horde ne fait pas, puisqu'elle converge vers
	// le joueur.
	//
	// C'est un `Handle` et non un identifiant nu : la place change à chaque
	// suppression par échange, et un identifiant privé de sa génération finirait
	// par désigner l'entité qui a recyclé la sienne.
	LastHit Handle
}
