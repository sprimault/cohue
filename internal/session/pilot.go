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

// PilotChoice rend la place que le pilote prend à sa n-ième montée.
//
// **Le pilote ne prenait aucune carte, et c'était l'angle mort du test qui vend
// le projet** : les montées s'ouvraient, s'accumulaient, et l'arme restait à son
// premier palier du premier au dernier tick. L'empreinte gardait donc le
// déterminisme d'une partie que personne ne joue — sans axes, sans synergies,
// sans la bascule de puissance que la conception met au cœur de la boucle.
//
// **La rotation porte sur les places et non sur les axes**, et c'est ce qui la
// rend sans exception. Cinq axes pour trois places : un axe visé ne serait pas
// toujours offert, il faudrait un repli, et le repli deviendrait la politique
// réelle sans qu'aucune ligne ne le dise. Une place est toujours là. Le tirage
// remplissant les places, la rotation couvre les axes sans avoir à les nommer.
//
// **Elle ne consomme aucun tirage**, à la différence d'un choix au hasard : le
// flux `Cards` reste alimenté par la seule offre, donc le témoin de l'empreinte
// continue de garder ce qu'il gardait.
//
// **Ce que la sonde garde est le déterminisme du mécanisme, jamais la justesse
// de l'équilibrage.** Aucune politique arbitraire ne joue comme un humain :
// celle-ci répartit ses prises pour visiter la table, là où un joueur suivrait
// une intention. Un chiffre tiré d'une run pilotée — le temps de survie, le
// niveau atteint — décrit donc cette politique et pas le jeu, et les confondre
// ferait croire qu'une run mesurée dit quelque chose de la courbe.
func PilotChoice(montees int) int {
	return montees % game.Choices
}
