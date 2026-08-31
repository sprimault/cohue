// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le point d'entrée : les chemins des ressources embarquées, leur chargement, le
// montage du monde, et le marqueur de l'étape qui reste à écrire. C'est le seul
// endroit du programme qui ait le droit de terminer le processus.

// Cohue est un action-roguelite urbain en vue isométrique, sous pression de
// horde.
//
// Le jeu n'existe pas encore : la feuille de route en donne les étapes, et
// `run` porte le marqueur de la première.
package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/sprimault/cohue"
	"github.com/sprimault/cohue/internal/game"
	"github.com/sprimault/cohue/internal/level"
	"github.com/sprimault/cohue/internal/render"
)

// Les trois chemins que le binaire connaît. Ils sont ici et nulle part
// ailleurs : le moteur ne reçoit qu'une grille de coûts et une table de
// profils, et c'est ce qui lui permet d'ignorer que les ressources sont
// embarquées plutôt que posées à côté.
const (
	manifesteDecor       = "assets/decors/manifeste.json"
	manifestePersonnages = "assets/personnages/manifeste.json"
	manifesteArmes       = "assets/armes/manifeste.json"
	lieuDepart           = "assets/lieux/place"

	// titreFenetre est ce que le gestionnaire de fenêtres affiche.
	titreFenetre = "Cohue"
)

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
// La fenêtre s'ouvre et ne montre qu'un fond uni : la simulation tourne en
// mémoire depuis l'étape 1, mais rien ne la dessine encore. Les marqueurs de
// l'étape 2 sont dans `render.Screen`, où ils attendent la projection.
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

	// Au lancement, et non à la première vague : un manifeste que le binaire
	// refuse doit le dire tout de suite, pas trois minutes après le démarrage
	// d'une partie.
	profils, err := game.LoadProfiles(cohue.Assets, manifestePersonnages)
	if err != nil {
		return err
	}
	slog.Info("profils chargés", "enemies", len(profils.Enemies))

	armes, err := game.LoadWeapons(cohue.Assets, manifesteArmes)
	if err != nil {
		return err
	}
	slog.Info("armes chargées", "base", armes.Base.Key)

	monde := game.NewWorld(profils, armes.Base, carte, capaciteHorde, capaciteTirs)
	slog.Info("monde monté", "capacite", monde.Enemies().Cap())

	ebiten.SetWindowTitle(titreFenetre)
	ebiten.SetWindowSize(render.Largeur, render.Hauteur)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	return ebiten.RunGame(&render.Screen{})
}
