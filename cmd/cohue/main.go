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
// La horde est semée au montage et n'arrive jamais par vagues : le spawner et sa
// courbe de pression sont le sujet de l'étape 4.
func run() error {
	partie, err := session.Open(cohue.Assets, cohue.StartingLevel)
	if err != nil {
		return err
	}

	ebiten.SetWindowTitle(titreFenetre)
	ebiten.SetWindowSize(render.Largeur, render.Hauteur)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	return ebiten.RunGame(render.NewScreen(partie.World, partie.Grid, partie.Tile))
}
