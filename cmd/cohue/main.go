// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le point d'entrée : le montage d'une partie et l'ouverture de la fenêtre.

// Cohue est un action-roguelite urbain en vue isométrique, sous pression de
// horde.
//
// Le jeu se réduit pour l'instant à un lieu qu'on traverse sous la pression
// d'une horde semée au montage : la feuille de route en donne les étapes.
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

// graineDeDepart est celle de la première run d'une session ; les suivantes en
// descendent.
//
// **Fixe à titre provisoire**, et c'est ce qui la sépare de celle de la planche
// de relecture, fixe par nature : ici elle l'est faute d'écran qui en choisisse
// une, et le premier à le faire la remplacera. Sans cette note, une graine en
// dur se lit comme une décision et personne ne la rouvre.
//
// Ce qui n'est pas provisoire, c'est qu'elle se reçoive : le montage ne devine
// pas de quelle run il s'agit. Deux lancements jouent donc aujourd'hui la même
// suite, ce qui ne se voit pas tant qu'aucun tirage n'entre dans la simulation.
const graineDeDepart uint64 = 1

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

// boucle enchaîne les parties : elle tient celle qui se joue et la remplace
// quand le joueur relance.
//
// **Elle n'est ici que parce qu'aucun des deux autres n'a le droit de la
// porter.** Le remontage appartient à `internal/session`, qui sait ce qu'une
// relance conserve et le prouve par un test ; la lecture de la touche appartient
// à `internal/render`, qui connaît déjà le clavier. Ce qui reste — appeler l'un
// quand l'autre le demande — est le seul geste que rien ne peut éprouver, et
// c'est pour cela qu'il est réduit à trois lignes.
type boucle struct {
	partie *session.Session
	ecran  *render.Screen
	hud    *render.HUD
}

// Update avance la partie, ou en monte une neuve si le joueur relance.
func (b *boucle) Update() error {
	if b.ecran.WantsRestart() {
		b.partie.Restart()
		b.monter()
		return nil
	}
	return b.ecran.Update()
}

// Draw peint la partie en cours.
func (b *boucle) Draw(ecran *ebiten.Image) { b.ecran.Draw(ecran) }

// Layout délègue à l'écran, qui fixe le tampon interne.
func (b *boucle) Layout(largeur, hauteur int) (int, int) { return b.ecran.Layout(largeur, hauteur) }

// monter accroche un écran neuf sur la partie courante.
//
// L'écran se remonte parce qu'il cadre sur le joueur à sa construction : le
// réutiliser laisserait la caméra là où la partie précédente s'est terminée, et
// la relance montrerait un premier instant décadré.
func (b *boucle) monter() {
	b.ecran = render.NewScreen(b.partie.World, b.partie.Grid, b.partie.Tile).WithHUD(b.hud)
}

// run monte le jeu et le fait tourner jusqu'à ce que le joueur quitte.
//
// La horde est semée au montage et n'arrive jamais par vagues : le spawner et sa
// courbe de pression sont le sujet de l'étape 4.
func run() error {
	partie, err := session.Open(cohue.Assets, cohue.StartingLevel, graineDeDepart)
	if err != nil {
		return err
	}
	hud, err := render.LoadHUD(cohue.Assets, cohue.InterfaceManifest)
	if err != nil {
		return err
	}

	jeu := &boucle{partie: partie, hud: hud}
	jeu.monter()

	ebiten.SetWindowTitle(titreFenetre)
	ebiten.SetWindowSize(render.Width, render.Height)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	return ebiten.RunGame(jeu)
}
