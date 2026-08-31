// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le point d'entrée : le montage d'une partie et l'ouverture de la fenêtre.
// C'est le seul endroit du programme qui ait le droit de terminer le processus.

// Cohue est un action-roguelite urbain en vue isométrique, sous pression de
// horde.
//
// Le jeu se réduit pour l'instant à un lieu qu'on traverse : la feuille de route
// en donne les étapes, et il n'y a personne d'autre à l'écran.
package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/sprimault/cohue"
	"github.com/sprimault/cohue/internal/render"
	"github.com/sprimault/cohue/internal/session"
)

// titreFenetre est ce que le gestionnaire de fenêtres affiche.
const titreFenetre = "Cohue"

// capaciteHorde plafonne le bassin des ennemis.
//
// Trois cents entités vivantes, ce que la conception donne comme total en
// comptant ce qui approche hors champ. Au-delà, ce n'est plus une horde mais un
// mur uni : les profils cessent d'être distinguables, et avec eux la lisibilité
// de l'échec. Le spawner rencontrera ce plafond, et c'est mieux que de laisser
// la horde croître jusqu'à ce que l'image s'effondre.
const capaciteHorde = 300

// capaciteTirs plafonne le bassin des projectiles.
//
// Large devant ce qu'une arme de base met en vol — quelques dizaines de ticks de
// portée pour une salve toutes les vingt-quatre —, parce que les passifs
// multiplient les projectiles bien plus vite que la cadence. Un bassin plein
// perd le tir plutôt que de le différer : une file d'attente rendrait la cadence
// élastique.
const capaciteTirs = 256

// version est renseignée à la liaison par -ldflags, et vaut « dev » hors
// publication.
var version = "dev"

// main journalise la version, puis sort en échec si le montage du jeu échoue.
// C'est le seul endroit du programme qui a le droit de terminer le processus.
func main() {
	slog.Info("cohue", "version", version)

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run monte le jeu et le fait tourner jusqu'à ce que le joueur quitte.
//
// Le lieu s'affiche et se parcourt au clavier, mais il n'y a personne d'autre à
// l'écran : rien ne fait encore apparaître d'ennemi, ce qui est le sujet de
// l'étape 4.
func run() error {
	partie, err := session.Open(cohue.Assets, cohue.StartingLevel, capaciteHorde, capaciteTirs)
	if err != nil {
		return err
	}

	ebiten.SetWindowTitle(titreFenetre)
	ebiten.SetWindowSize(render.Largeur, render.Hauteur)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	return ebiten.RunGame(render.NewScreen(partie.World, partie.Grid, partie.Tile))
}
