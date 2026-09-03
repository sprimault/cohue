// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// L'empreinte d'un état de partie : ce que deux exécutions d'une même graine
// doivent rendre identique, écrit en texte comparable ligne à ligne.

package game

import (
	"fmt"
	"strings"
)

// Fingerprint rend l'état de la partie sous une forme comparable.
//
// **Elle énumère ce qu'elle inclut, jamais ce qu'elle écarte.** Une liste de
// champs à ignorer se périmerait au premier champ ajouté, et l'empreinte
// changerait pour une raison sans rapport avec ce qu'elle garde ; une liste de ce
// qui compte laisse un champ nouveau dehors par défaut, ce qui l'affaiblit au
// lieu de la casser. Des deux erreurs, c'est celle qui se rattrape.
//
// **De l'état, jamais un résumé.** Un compte d'entités vivantes ou un score
// passerait au vert alors que deux trajectoires ont divergé puis se sont
// recroisées, ce qui est précisément le cas recherché.
//
// **L'ordre est celui des index du bassin**, qui n'est stable qu'au sein d'un
// tick : la suppression par échange déplace la dernière entité dans le trou. Une
// empreinte prise en fin de run n'en souffre pas ; une empreinte par tick
// devrait trier par identifiant, et sa lecture ne serait plus la même chose.
//
// Les positions s'écrivent en unités de virgule fixe et non en tuiles : ce qui
// est comparé est ce que la simulation porte, et une conversion en flottant
// rendrait deux états distincts égaux à l'affichage.
func (w *World) Fingerprint() string {
	var b strings.Builder

	fmt.Fprintf(&b, "tick %d\n", w.tick)
	fmt.Fprintf(&b, "joueur x=%d y=%d vie=%d niveau=%d xp=%d charge=%t\n",
		w.playerX, w.playerY, w.vie, w.niveau, w.experience, w.charge)

	for i := range w.ennemis.Active() {
		e := w.ennemis.At(i)
		id := w.ennemis.IDAt(i)
		fmt.Fprintf(&b, "ennemi %d id=%d gen=%d profil=%d x=%d y=%d hits=%d pas=%d,%d flash=%d\n",
			i, id, w.ennemis.gens[id], e.Profile, e.X, e.Y, e.Hits, e.Step.X, e.Step.Y, e.Flash)
	}
	for i := range w.tirs.Active() {
		p := w.tirs.At(i)
		id := w.tirs.IDAt(i)
		fmt.Fprintf(&b, "tir %d id=%d gen=%d x=%d y=%d reste=%d hits=%d pas=%d,%d\n",
			i, id, w.tirs.gens[id], p.X, p.Y, p.Remaining, p.Hits, p.Step.X, p.Step.Y)
	}
	for i := range w.gemmes.Active() {
		g := w.gemmes.At(i)
		id := w.gemmes.IDAt(i)
		fmt.Fprintf(&b, "gemme %d id=%d gen=%d x=%d y=%d ne=%d attiree=%t\n",
			i, id, w.gemmes.gens[id], g.X, g.Y, g.Born, g.Pulled)
	}
	for i := range w.aimants.Active() {
		a := w.aimants.At(i)
		id := w.aimants.IDAt(i)
		fmt.Fprintf(&b, "aimant %d id=%d gen=%d x=%d y=%d\n",
			i, id, w.aimants.gens[id], a.X, a.Y)
	}

	// **Un témoin par flux, et non son état interne.** Ce qu'il faut garder est
	// que deux exécutions ont consommé le hasard identiquement ; un générateur
	// dont le prochain tirage coïncide a consommé le même nombre de valeurs, ce
	// qui décide la question sans exposer d'état.
	//
	// Il ne prouve donc pas que les deux états sont égaux, il prouve qu'ils
	// rendent le même tirage suivant — c'est suffisant ici, et ça ne l'est pas
	// pour qui lirait « empreinte de l'état des flux ».
	//
	// Le tirage se consomme, ce qui est sans conséquence : une empreinte se
	// prend d'une partie qu'on ne joue plus.
	fmt.Fprintf(&b, "flux vagues=%d positions=%d butin=%d cosmetique=%d\n",
		w.hasard.Waves.IntN(1<<30), w.hasard.Positions.IntN(1<<30),
		w.hasard.Loot.IntN(1<<30), w.hasard.Cosmetic.IntN(1<<30))

	return b.String()
}
