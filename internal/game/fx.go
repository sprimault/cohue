// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les effets brefs : ce qui reste à l'écran d'une chose qui n'existe plus, et
// que la simulation n'interroge jamais.

package game

// FxKind dit ce qu'un effet marque.
//
// **Ce que l'événement était, jamais ce qu'il faut dessiner.** Une caisse qui
// cède crache du bois, une déflagration une onde : ces deux phrases appartiennent
// au rendu, qui a le manifeste où la matière d'un objet est déclarée. La
// simulation, elle, sait qu'une caisse a cédé — c'est un fait de jeu, et il se
// nomme comme tel.
type FxKind uint8

// Ce qui produit un effet aujourd'hui. Les destructibles et la mort d'une
// créature s'y ajouteront avec leur mécanisme.
const (
	// FxCrate est une caisse qui vient de céder.
	FxCrate FxKind = iota
	// FxBlast est une déflagration qui vient de partir.
	FxBlast
)

// Les durées de vie d'un effet, en ticks.
//
// **Elles vivent ici comme les deux durées d'éclair**, et pour la même raison :
// c'est la simulation qui tient le décompte, et ce que le rendu en fait lui
// appartient. Ce ne sont pas des réglages de jeu — rien ne dépend d'elles.
//
// Celle du souffle porte une condition plutôt qu'un motif : **elle doit couvrir
// l'animation**, que le manifeste des objets déclare en cinq images de soixante
// millisecondes. Le jour où cette bande s'allonge, ce nombre suit — sinon
// l'onde se coupe en plein vol.
const (
	dureeEclats  Tick = 24
	dureeSouffle Tick = 18
)

// Fx est un effet bref, posé là où quelque chose a eu lieu.
//
// **Une entrée par événement et non par éclat.** Une caisse en crache une
// dizaine, tous partis du même point au même instant : leurs directions, leur
// parabole et leur rotation se dérivent du rang de chacun et de l'âge commun,
// exactement comme l'image d'un cycle se dérive du tick. Les stocker
// individuellement paierait dix fois ce qu'un seul point suffit à dire.
//
// **Rien ici n'entre dans l'empreinte d'une run.** Le bassin est entièrement
// cosmétique, comme celui des cadavres que la conception décrit : il ne pousse
// personne, ne se vise pas, ne compte nulle part, et une run simulée sans rendu
// peut ne pas l'alimenter sans que rien ne diverge.
type Fx struct {
	// X et Y sont le point où l'événement a eu lieu. L'effet n'y bouge plus :
	// ce qui l'a produit n'existe plus, donc il n'a rien à suivre.
	X, Y Fixed
	// Kind dit ce qui l'a produit.
	Kind FxKind
	// Life est ce qui lui reste à vivre, et Total ce qu'il a reçu.
	//
	// **Les deux, parce que le rendu a besoin de l'avancement et non du reste.**
	// Une parabole se lit sur le chemin parcouru, une bande d'images aussi ; le
	// seul décompte obligerait chaque lecteur à retrouver le total par la sorte,
	// c'est-à-dire à redire ici ce que la constante dit déjà.
	Life, Total Tick
}

// Fxs rend le bassin des effets brefs, que le rendu parcourt.
func (w *World) Fxs() *Pool[Fx] { return w.effets }

// emettre pose un effet là où quelque chose vient d'avoir lieu.
//
// Le bassin plein perd l'effet plutôt que d'en chasser un : deux caisses cassées
// dans la même seconde valent mieux qu'une caisse dont les éclats sautent au
// milieu de leur vol. Et perdre un effet ne coûte rien qu'un peu de décor —
// c'est ce que veut dire cosmétique.
func (w *World) emettre(x, y Fixed, quoi FxKind) {
	vie := dureeEclats
	if quoi == FxBlast {
		vie = dureeSouffle
	}
	w.effets.Spawn(Fx{X: x, Y: y, Kind: quoi, Life: vie, Total: vie})
}

// vieillirEffets fait vivre les effets et retire ceux qui ont fini.
//
// **Au début du tick, avant tout ce qui émet.** Un effet posé pendant le tick
// doit se montrer entier à l'image qui suit : décompté à la fin du même tick, il
// naîtrait déjà vieux d'un pas et perdrait sa première image — celle où une
// volée est encore groupée et où une onde est la plus serrée.
//
// Sa place est libre par ailleurs, et c'est la meilleure preuve qu'il ne décide
// de rien : rien de la partie ne le lit, comme l'errance des figurants.
func (w *World) vieillirEffets() {
	for i := 0; i < w.effets.Len(); {
		e := w.effets.At(i)
		e.Life--
		if e.Life <= 0 {
			w.effets.RemoveAt(i)
			continue
		}
		i++
	}
}
