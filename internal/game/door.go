// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La porte de sortie : ce qui l'ouvre, et ce que la franchir termine. Elle est
// posée par le lieu comme les figurants, et l'objectif qui la déverrouille se
// compte en créatures abattues — donc jamais en figurants.

package game

import (
	"fmt"

	"github.com/sprimault/cohue/internal/manifest"
)

// ExitSpec est la sortie telle qu'un lieu l'écrit.
//
// **Un pointeur chez son porteur**, parce qu'un lieu peut n'en avoir aucune —
// une boutique, un passage. L'absence et une sortie mal écrite sont deux choses
// différentes, et seule la seconde se refuse.
type ExitSpec struct {
	manifest.Commentable
	// At est la case de la porte, en coordonnées de lieu.
	At *[2]int `json:"position"`
	// Kills est le nombre de créatures à abattre avant qu'elle s'ouvre.
	//
	// **Un pointeur, et c'est la règle de la valeur zéro qui l'impose.** Le champ
	// absent et « zéro créature » se liraient tous deux comme un entier nul,
	// c'est-à-dire comme une porte déjà ouverte au premier tick — exactement ce
	// que « la sortie se gagne » interdit, obtenu en oubliant une ligne.
	Kills *int `json:"abattus"`
}

// Exit est une sortie compilée : où elle est, et ce qu'elle demande.
type Exit struct {
	// X et Y sont le centre de sa case, où le joueur vient la toucher.
	X, Y Fixed
	// Kills est le nombre de créatures à abattre pour l'ouvrir.
	Kills int
	// U et V sont sa case, que le rendu peint à part.
	U, V int
}

// CompileExit résout une sortie écrite contre la carte cuite.
//
// Elle rend tout ce qui l'empêche de valoir plutôt que le premier écart, comme
// la compilation des vagues et celle du peuplement.
//
// **Une sortie se pose sur une case bloquée, et le refus de l'inverse n'est pas
// un caprice de format.** La porte fermée *est* le mur qui retient : posée sur
// du sol libre, elle serait franchissable avant d'être gagnée, et le lieu se
// terminerait en marchant dessus. Le fichier ne peut pas dire quelle forme de
// décor occupe la case — la cuisson n'en garde que le coût —, mais il peut dire
// qu'elle est infranchissable, ce qui est la propriété dont le mécanisme dépend.
func CompileExit(brut *ExitSpec, carte *CostGrid) (*Exit, []string) {
	if brut == nil {
		return nil, nil
	}

	var manques []string
	dire := func(format string, args ...any) {
		manques = append(manques, fmt.Sprintf(format, args...))
	}

	if brut.At == nil {
		dire("sortie.position : absente, une porte se place")
	}
	if brut.Kills == nil {
		dire("sortie.abattus : absent, une sortie se gagne")
	} else if *brut.Kills < 1 {
		dire("sortie.abattus : %d, attendu au moins 1 — a zero la porte est ouverte au premier tick",
			*brut.Kills)
	}
	if brut.At == nil {
		return nil, manques
	}

	u, v := brut.At[0], brut.At[1]
	switch {
	case !carte.InBounds(u, v):
		dire("sortie.position : (%d, %d) hors du lieu, qui fait %d sur %d",
			u, v, carte.Width(), carte.Height())
	case carte.Passable(u, v):
		dire("sortie.position : (%d, %d) est franchissable, une sortie se pose sur une porte fermee",
			u, v)
	}
	if len(manques) > 0 {
		return nil, manques
	}

	return &Exit{
		X:     FromInt(u) + One/2,
		Y:     FromInt(v) + One/2,
		Kills: *brut.Kills,
		U:     u,
		V:     v,
	}, nil
}

// SetExit pose la sortie du lieu, au montage et à chaque relance.
//
// Le pendant de `Populate` pour les figurants, et posé au montage pour la même
// raison : la carte cuite est partagée par toutes les runs d'une session, si
// bien que l'état d'ouverture ne peut pas y vivre. Une porte ouverte qui aurait
// modifié la grille rouvrirait la suivante avant le premier tick.
func (w *World) SetExit(sortie *Exit) { w.sortie = sortie }

// Exit rend la sortie du lieu, nulle quand il n'en a pas.
func (w *World) Exit() *Exit { return w.sortie }

// Kills rend le nombre de créatures abattues depuis le début de la run.
func (w *World) Kills() int { return w.abattus }

// DoorOpen dit si la porte est gagnée.
//
// Un lieu sans sortie n'en a jamais : rien à ouvrir, et la partie ne se termine
// que par la mort.
func (w *World) DoorOpen() bool {
	return w.sortie != nil && w.abattus >= w.sortie.Kills
}

// Escaped dit si le joueur a quitté le lieu par la porte.
func (w *World) Escaped() bool { return w.echappe }

// Over dit si la partie est terminée, d'une façon ou de l'autre.
//
// **Deux fins et non une, et c'est tout l'objet de ce lot.** Jusqu'ici la seule
// issue était la mort, si bien que `Alive` suffisait à dire que la partie
// continuait ; la sortie en ajoute une que la vie ne porte pas, et un appelant
// qui interrogerait encore la vie seule ferait tourner la horde sous l'écran de
// fin.
func (w *World) Over() bool { return !w.Alive() || w.echappe }

// franchir sort le joueur du lieu quand il touche une porte ouverte.
//
// **Toucher et non traverser.** La porte reste l'obstacle que le décor déclare,
// pour deux raisons qui vont dans le même sens : une horde qui sortirait par où
// le joueur sort n'aurait pas de sens, et rendre la case franchissable
// demanderait de modifier la grille cuite — donc de mettre de l'état de partie
// dans ce que les relances se partagent.
//
// Elle vient en fin de tick, après la moisson des morts : la créature qui
// complète l'objectif ouvre la porte dans le tick où elle tombe, et un joueur
// déjà contre elle sort sans attendre le suivant.
func (w *World) franchir() {
	if w.echappe || !w.Alive() || !w.DoorOpen() {
		return
	}

	// **Une case pleine en plus du rayon, et non une demi-case.** La porte est
	// murée : le joueur ne peut donc jamais atteindre son centre, ni même en
	// approcher à moins d'une demi-case plus son rayon. Une portée d'une
	// demi-case tomberait pile sur cette borne et la porte serait intouchable —
	// ouverte, signalée, et sans effet. Ce qu'il faut couvrir est la case
	// adjacente, pas le bord de la sienne.
	portee := w.profils.Player.Radius + One
	ecart := Vec{X: w.playerX - w.sortie.X, Y: w.playerY - w.sortie.Y}
	if ecart.carres() < int64(portee)*int64(portee) {
		w.echappe = true
	}
}
