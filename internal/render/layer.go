// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les couches du rendu : ce qui passe devant toute l'image, et dans quel ordre.
// Trois suffisent, et c'est leur raison d'être qui les sépare.

package render

// Les couches, de la plus basse à la plus haute.
//
// **Une couche passe devant toute l'image, un étage ne range qu'une hauteur de
// décor.** Le monde entier tient donc dans la couche zéro, où la profondeur et
// l'étage décident ; ce qui suit est ce que la conception veut voir quoi qu'il
// arrive.
//
// **La silhouette est au-dessus du monde et sous l'interface**, et les deux
// bornes comptent. Au-dessus, parce que perdre son personnage sous un
// empilement est ce qui peut arriver de pire à la lisibilité, et que cela
// survient précisément quand on est encerclé. En dessous, parce qu'un bandeau
// qu'une silhouette traverserait cesserait de se lire — et c'est le seul endroit
// du jeu où une information critique vit hors du regard.
//
// **Les effets sont au-dessus de la silhouette, et c'est l'ordre d'avant le
// portage.** Un éclat, un chiffre de dégâts, une gerbe d'étincelles disent ce
// qui vient de se passer ; ils sont brefs et petits, là où le rappel est large
// et continu. Les mettre dessous les ferait disparaître sous l'aplat d'un joueur
// encerclé, au moment précis où ils comptent.
//
// **Quatre et non cinq** : rien ne distingue aujourd'hui le bandeau d'une carte
// de choix, qui ne se recouvrent jamais. Le jour où deux panneaux se
// chevaucheront, c'est l'ordre de pose qui les départagera — `PlaceScreen` le
// garantit entre pièces d'une même couche —, et une couche de plus ne se
// justifiera que si l'un doit passer devant l'autre sans être posé après.
const (
	// coucheMonde porte le décor et les entités. La profondeur y décide, et c'est
	// la seule couche où la caméra déplace ce qu'on voit.
	coucheMonde uint8 = iota
	// coucheSilhouette porte le rappel de ce qui est caché : le joueur et les
	// projectiles de la horde, jamais une créature.
	coucheSilhouette
	// coucheEffets porte ce qui ne se trie avec rien : la traînée d'un tir, les
	// éclats d'une casse, un chiffre de dégâts, l'annonce d'une caisse.
	coucheEffets
	// coucheInterface porte ce qui ne bouge pas avec la caméra — bandeau, cartes,
	// vignette, écran de fin.
	coucheInterface
)

// Les rangs de strate, qui ne départagent que ce qui se superpose exactement.
//
// **Une strate n'est pas une couche** : elle ne range qu'à profondeur et
// abscisse égales, donc deux entités qui ne se recouvrent pas ne se comparent
// jamais par elle. Les rangs d'un groupe sont de ce fait indépendants de ceux
// d'un autre, et c'est pourquoi le sprite d'une créature reste à zéro quand
// celui du joueur monte à un.
//
// **Ils ne valaient rien avant la 0.5.4**, la clé de tri plaçant l'index de
// planche avant le rang : deux entités de planches différentes étaient rangées
// par l'ordre de chargement de leurs images, et le rang n'avait pas voix. C'est
// ce qui tenait le liseré et l'éclair d'état hors du jeu.
const (
	// rangDuLisere porte les huit quads qui cernent le joueur, sous son dessin.
	rangDuLisere uint8 = iota
	// rangDuJoueur porte le dessin du joueur, au-dessus de son liseré.
	rangDuJoueur
)

// rangDeLEclair porte la teinte additive d'un état, au-dessus du sprite de la
// créature qu'elle désigne — lequel reste au rang zéro, n'ayant rien dessous.
const rangDeLEclair uint8 = 1
