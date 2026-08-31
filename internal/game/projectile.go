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
}
