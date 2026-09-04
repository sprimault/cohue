// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le pilote des runs sans joueur : ce qui déplace le personnage quand personne
// ne tient les touches. Il sert au test de déterminisme et aux planches de
// relecture, qui n'ont besoin ni l'un ni l'autre d'un bon joueur — seulement
// d'un joueur qui ne meure pas à la première minute.

package session

import "github.com/sprimault/cohue/internal/game"

// segmentPilote est le nombre de ticks pendant lesquels le pilote garde un cap.
//
// Quatre-vingt-dix ticks font une seconde et demie, soit sept tuiles et demie à
// cinq tuiles par seconde : l'octogone décrit fait une vingtaine de tuiles de
// large, assez pour que la horde la plus lente ne recoupe pas la corde et assez
// petit pour tenir dans un bloc du lieu livré.
const segmentPilote game.Tick = 90

// Pilot rend la direction que suit un joueur automatique au tick donné.
//
// **Il ne lit rien du monde, et c'est une contrainte et non une paresse.** Un
// pilote qui fuirait la créature la plus proche serait un meilleur joueur, mais
// son trajet dépendrait alors de la courbe de pression : l'empreinte de
// référence bougerait au premier réglage d'équilibrage, pour une raison sans
// rapport avec le déterminisme qu'elle garde. C'est le même argument qui fait
// prendre les instants à des ticks fixes plutôt qu'à la mort.
//
// **Un octogone plutôt qu'un cercle**, parce qu'un cercle demanderait un sinus :
// `math.Sin` est déterministe sur une machine, pas garanti identique d'une cible
// à l'autre, et ce test tourne sur trois. Les huit orientations du monde
// suffisent à décrire un tour, et elles sont exactes en virgule fixe.
//
// **Ce que le tour achète** : une direction fixe mène au mur, où la horde
// encercle un joueur qui ne peut plus reculer — la mort arrivait à 1:28 sur la
// planche et vers 2:00 dans la run de référence, avant même l'entrée du
// deuxième profil. En tournant, le joueur sème les poursuivants les plus lents
// et reste au milieu du lieu, ce qui suffit à voir arriver les paliers suivants.
//
// Il reste un joueur médiocre, et c'est voulu : ce qu'on veut mesurer n'est pas
// la meilleure run possible mais une run qui traverse la courbe.
func Pilot(tick game.Tick) game.Vec {
	return game.Heading(int(tick / segmentPilote))
}
