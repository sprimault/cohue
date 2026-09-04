// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Enemy est une créature de la horde telle qu'elle vit dans son bassin : sa
// position, l'index de son profil, et la résistance dont la chute est sa mort.

package game

// Enemy est une créature de la horde, telle qu'elle vit dans son bassin.
//
// Une struct nue, sans méthode ni pointeur : elle est copiée d'une place à
// l'autre à chaque suppression par échange, et tout ce qu'elle porte doit
// supporter ce déplacement. Ce qui la désigne d'une image à l'autre est un
// Handle, jamais un `*Enemy`.
type Enemy struct {
	// Profile est l'index de son profil dans `Profiles.Enemies`, et jamais une
	// copie de ses valeurs — même « pour éviter une indirection ». C'est ce qui
	// rend une modification de la table effective sans recharger le monde, et ce
	// qui empêche deux Badauds d'avoir des vitesses différentes.
	Profile int
	// X et Y sont sa position dans le monde, en tuiles.
	X, Y Fixed
	// Flash est ce qui reste de l'éclair d'impact, en ticks.
	//
	// **Un décompte et non la date du coup.** La date se comparerait au tick
	// courant, ce qui rendrait la valeur nulle indiscernable d'un coup reçu au
	// premier tick ; un décompte n'a pas de sentinelle et la horde est de toute
	// façon parcourue à chaque tick.
	//
	// Il vit ici plutôt que dans le rendu parce que c'est un état de la créature,
	// que la relance efface avec elle. Ce que le rendu en fait — la teinte, sa
	// vivacité — lui appartient.
	Flash Tick
	// Step est le déplacement que le tick précédent lui a appliqué, en tuiles.
	//
	// **Le pas appliqué et non l'intention**, parce que c'est lui qui prédit :
	// une créature qui glisse le long d'un mur ou qu'une voisine repousse
	// n'ira pas là où le champ de flux l'appelle. Il sert à la visée, qui tire
	// où la cible sera plutôt que là où elle est.
	Step Vec
	// ChargePhase est où elle en est de son cycle de charge, `ChargeNone` pour
	// une créature qui ne charge pas — ce que sa valeur zéro rend vrai sans
	// qu'on ait à l'initialiser, y compris après un échange de bassin.
	ChargePhase ChargePhase
	// ChargeTimer est ce qui reste de la phase en cours, en ticks. Un décompte
	// et non une date, pour la raison qui vaut déjà pour `Flash`.
	ChargeTimer Tick
	// ShotTimer est ce qui reste avant que la créature puisse tirer, en ticks.
	//
	// **Il ne se consomme pas hors de portée**, comme la cadence de l'arme du
	// joueur : une Buse qui sort d'un couloir désert tirerait sinon avec un
	// retard fonction du temps passé sans cible, ce que rien à l'écran
	// n'expliquerait.
	ShotTimer Tick
	// ChargeDir est la direction figée au départ de la course.
	//
	// **Figée, et c'est tout le comportement.** Une créature qui recalculerait
	// sa direction pendant la course serait un poursuivant rapide ; ce qui rend
	// la charge lisible est qu'elle ne corrige plus, donc qu'on peut s'en
	// écarter d'un pas de côté.
	ChargeDir Vec
	// Hits est ce qu'il lui reste à encaisser, dans l'unité où s'exprime la
	// résistance : des touches de l'arme de base à son premier niveau.
	//
	// **La mort est cet état, pas un événement.** Une résistance tombée à zéro
	// *est* la mort : pas de drapeau à côté, rien à synchroniser, et savoir si
	// une créature est une cible valide reste une lecture. Ce qui se déclenche
	// une fois est la transition — l'endroit qui applique les dégâts constate
	// qu'elle était positive et ne l'est plus.
	Hits int
}
