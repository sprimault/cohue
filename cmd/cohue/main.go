// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le point d'entrée : le montage d'une partie et l'ouverture de la fenêtre.

// Cohue est un action-roguelite urbain en vue isométrique, sous pression de
// horde.
//
// Le jeu se réduit pour l'instant à un lieu qu'on traverse sous une horde qui
// s'achète dans un budget de pression : la feuille de route en donne les étapes.
package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/sprimault/lozengine"

	"github.com/sprimault/cohue"
	"github.com/sprimault/cohue/internal/render"
	"github.com/sprimault/cohue/internal/session"
	"github.com/sprimault/cohue/internal/sprite"
	"github.com/sprimault/cohue/internal/ui"
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
// pas de quelle run il s'agit. Deux lancements jouent donc la même suite, et cela
// se voit désormais — mêmes vagues achetées aux mêmes instants, mêmes places
// d'aimant.
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
	// echec porte ce qu'une relance a refusé. La boucle du moteur ne sait pas
	// recevoir d'erreur en cours de route — un jeu qui veut partir le dit par
	// `Quitting` —, et la taire laisserait une partie tourner sur un décor que
	// personne n'a reposé.
	echec error
}

// Step avance la partie, ou relie l'écran à une partie neuve si le joueur
// relance, puis décrit l'image.
//
// **Les entrées viennent du moteur, et la boucle n'interroge plus rien.** Elles
// étaient relevées ici parce que les fonctions d'état de la bibliothèque
// précédente étaient globales, et qu'une méthode de dessin pouvait lire une
// touche sans que sa signature le dise. Le pas les reçoit désormais en valeur.
//
// **Avancer et décrire sont le même appel**, parce que c'est ce que le moteur
// attend : il ne dessine pas ce qu'on lui demande, il trie et peint ce que la
// scène porte quand le pas rend la main.
func (b *boucle) Step(in lozengine.Input) {
	if b.echec != nil {
		return
	}
	if b.ecran.WantsRestart(in) {
		b.partie.Restart()
		b.echec = b.ecran.Rebind(b.partie.World, b.partie.Grid)
		if b.echec != nil {
			return
		}
	} else {
		b.ecran.Step(in)
	}
	b.ecran.Describe()
}

// Overlaps relaie à l'écran le relevé de ce qui recouvre les choses qu'il suit.
//
// La boucle du moteur la cherche sur le jeu, et le jeu est cet objet-ci : sans
// ce relais, le rappel de ce qui est caché ne recevrait jamais sa mesure.
func (b *boucle) Overlaps(list lozengine.OverlapList) { b.ecran.Overlaps(list) }

// Quitting arrête la boucle quand une relance a échoué.
//
// C'est la seule sortie de secours d'un jeu qui passe par la boucle du moteur,
// et l'erreur est rendue par `run` une fois la main reprise : sortir du
// processus ici sauterait tout ce que la boucle fait en se terminant.
func (b *boucle) Quitting() bool { return b.echec != nil }

// run monte le jeu et le fait tourner jusqu'à ce que le joueur quitte.
func run() error {
	partie, err := session.Open(cohue.Assets, cohue.StartingCampaign, graineDeDepart)
	if err != nil {
		return err
	}

	// L'atlas se remplit à mesure que les catalogues se chargent, chacun y
	// rangeant ses planches et retenant leurs index. Il se construit donc avant
	// eux, et se remet au moteur après eux.
	atlas := &lozengine.Atlas{}
	hud, err := render.LoadHUD(cohue.Assets, cohue.InterfaceManifest, atlas)
	if err != nil {
		return err
	}
	tuiles, err := sprite.LoadTiles(cohue.Assets, cohue.DecorDir, partie.Decor, atlas)
	if err != nil {
		return err
	}
	sol, err := render.NewTerrain(partie.Tiles, tuiles)
	if err != nil {
		return err
	}
	troupe, err := render.NewCast(cohue.Assets, cohue.CharacterDir, partie.Profiles, atlas)
	if err != nil {
		return err
	}
	objets, err := render.NewStage(cohue.Assets, cohue.ObjectDir, cohue.ObjectManifest,
		partie.Objects, atlas)
	if err != nil {
		return err
	}
	// Le retrait que l'interface réserve se lit avant l'écran, qui en tire le
	// cadrage de sa première image, et il part aux options du moteur, qui en tire
	// toutes les suivantes.
	retrait := hud.Inset()
	ecran, err := render.NewScreen(partie.World, partie.Grid, sol, troupe, objets, atlas,
		retrait)
	if err != nil {
		return err
	}

	// Les icônes ne passent pas par l'atlas : la fenêtre les porte, aucun quad
	// n'y puise.
	icones, err := ui.LoadIcons(cohue.Assets, cohue.InterfaceManifest)
	if err != nil {
		return err
	}

	jeu := &boucle{partie: partie, ecran: ecran.WithHUD(hud)}

	// **Le monde remplit la fenêtre moins la bande du bandeau, et c'est un
	// arbitrage jugé à l'écran.** `Width` et `Height` sont le plancher de la
	// surface : la fenêtre s'ouvre à cette taille, puis la surface grandit au
	// facteur entier près. Le retrait en sort la hauteur du bandeau, toujours la
	// même, et l'aire de monde prend tout le reste — il ne reste donc aucune
	// bande, et agrandir montre un peu plus de lieu.
	//
	// **Le plafond a été essayé et écarté, et c'est l'écran qui a tranché.** Il
	// borne l'aire en pixels, donc la marge est ce que la fenêtre a de plus :
	// 44 % de la surface sur un écran de 1440p, le lieu collé dans un cadre noir.
	// Le reste d'une division entière va au monde ou à la marge, et il n'y a pas
	// de troisième endroit — c'est de l'arithmétique, pas un réglage. Un retrait
	// fait l'inverse, et c'est celui-là qu'on veut : la marge est une donnée de
	// mise en page, l'aire est ce qui reste.
	//
	// L'alignement n'a donc rien à dire : une aire qui prend tout ce que le
	// retrait laisse part du retrait, quel que soit `WorldAlignY`.
	//
	// `Bars` n'est pas posé : son défaut noir est voulu par le moteur, qui tient
	// à ce que la frontière de l'image reste visible, et il ne reste de toute
	// façon au plus qu'un résidu de `facteur − 1` pixels.
	err = lozengine.Run(ecran.Scene(), jeu, lozengine.Options{
		Title:      titreFenetre,
		Width:      render.Width,
		Height:     render.Height,
		WorldInset: retrait,
		Background: render.Background,
		Atlas:      atlas,
		Icons:      icones,
	})
	if err != nil {
		return err
	}
	return jeu.echec
}
