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

	"github.com/sprimault/cohue"
	"github.com/sprimault/cohue/internal/level"
)

// Les deux chemins que le binaire connaît. Ils sont ici et nulle part ailleurs :
// le moteur ne reçoit qu'une grille de coûts, et c'est ce qui lui permet
// d'ignorer que les ressources sont embarquées plutôt que posées à côté.
const (
	manifesteDecor = "assets/decors/manifeste.json"
	lieuDepart     = "assets/lieux/place.json"
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
// rendu — quand il y aura quelque chose à monter. Le chargement, lui, se fait
// déjà : c'est ce qui met les ressources dans le binaire, un `go:embed` que rien
// n'importe restant un fichier sans effet.
func run() error {
	_, couts, err := level.LoadDecor(cohue.Assets, manifesteDecor)
	if err != nil {
		return err
	}
	carte, err := level.NewLoader(cohue.Assets, couts).Load(lieuDepart)
	if err != nil {
		return err
	}
	slog.Info("lieu chargé", "largeur", carte.Width(), "hauteur", carte.Height())

	return errors.New("à implémenter : étape 1")
}
