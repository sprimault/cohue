// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Cohue est un action-roguelite urbain en vue isométrique, sous pression de
// horde.
//
// Le jeu n'existe pas encore : la feuille de route en donne les étapes, et
// `run` porte le marqueur de la première.
package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
)

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
// Elle rendra la struct de dépendances construite ici — bassins, champ de flux,
// rendu — quand il y aura quelque chose à monter.
func run() error {
	return errors.New("à implémenter : étape 1")
}
