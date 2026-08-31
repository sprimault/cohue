// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La planche de relecture du rendu : les vues qu'elle écrit, le monde qu'elle
// pose pour chacune, et le détour par le cycle Ebitengine que le dessin impose.

// Preview écrit des vues du rendu dans `.tmp/apercus/`, à regarder.
//
// `internal/render` n'a pas de test et n'en aura pas : importer Ebitengine
// initialise GLFW, qui panique sans écran. Le rendu se juge donc à l'œil, et une
// planche est ce qui rend ce jugement possible ailleurs que dans une partie —
// on y compare une image d'avant et d'après un changement, ce qu'une observation
// au clavier ne permet pas.
//
// **Elle pilote le `render.Screen` du jeu, jamais une scène montée à côté.** Une
// planche qui dessinerait autrement que le jeu relirait la planche, et c'est le
// même piège qu'un test qui bâtit son entrée au lieu de passer par le chemin qui
// la produit.
//
// **Elle est déterministe, et doit le rester.** Deux exécutions écrivent des
// octets identiques, sans quoi comparer une planche d'avant et d'après un
// changement ne dirait rien. C'est gratuit tant que rien n'est tiré au sort ; le
// jour où des créatures seront posées, leurs positions viendront d'une graine
// fixée ici et non d'un tirage libre.
//
// Elle exige un écran et ne tourne donc pas en intégration continue. Ce n'est
// pas un contrôle mais une planche : ce qui se vérifie mécaniquement vit du côté
// simulation, qui n'importe pas Ebitengine.
package main

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/sprimault/cohue"
	"github.com/sprimault/cohue/internal/game"
	"github.com/sprimault/cohue/internal/render"
	"github.com/sprimault/cohue/internal/session"
)

// sortie est le dossier des planches.
//
// Dans `.tmp/` et non dans `assets/`, dont le contrôle des ressources compare le
// contenu à ce que les générateurs produisent : une planche y deviendrait un
// écart à excepter, et elle pèserait dans le binaire de qui l'a régénérée.
const sortie = ".tmp/apercus"

// Les capacités des bassins, qui n'ont ici aucune importance : la planche ne
// peuple rien. Basses plutôt que recopiées de celles du jeu, qui auraient donné
// deux valeurs à tenir d'accord pour un montage qui ne les exerce pas.
const (
	capacite = 1
	tirs     = 1
)

// vue est une position du joueur et le nom du fichier qui en sort.
type vue struct {
	nom  string
	u, v int
}

// vues énumère ce que la planche donne à relire.
//
// Le centre montre la projection sans rien qui la borde ; les quatre autres
// posent le joueur près d'un coin du losange, là où la caméra bute et où le
// joueur se décentre. Ce sont les seuls endroits où le cadrage décide de quelque
// chose, et ils couvrent les deux axes dans les deux sens.
var vues = []vue{
	{"centre", 16, 16},
	{"nord", 2, 2},
	{"ouest", 2, 29},
	{"est", 29, 2},
	{"sud", 29, 29},
}

// planche écrit toutes les vues au premier pas, puis demande l'arrêt.
//
// Elle est un jeu Ebitengine sans l'être : le dessin et la lecture de pixels ne
// sont valides que dans le cycle de la bibliothèque, si bien qu'un programme qui
// écrirait ses images depuis `main` n'obtiendrait rien. D'où ce détour, dont
// `Draw` reste vide — la fenêtre qui s'ouvre une fraction de seconde n'a rien à
// montrer.
//
// **Elle n'appelle jamais `render.Screen.Update`, et c'est délibéré.** Cette
// méthode lit les touches : une direction pressée pendant l'écriture déplacerait
// le joueur, et les images cesseraient d'être comparables d'une exécution à
// l'autre — le déterminisme reposerait sur la précaution de ne pas toucher au
// clavier plutôt que sur une propriété. Ne pas l'appeler n'est pas une garde
// qu'on pourrait contourner, c'est un chemin qui n'existe pas.
//
// Ce qu'elle en perd est nul : monter un écran cadre déjà sur le joueur, donc un
// écran monté après la scène est cadré juste. Ce qui doit avancer d'un pas passe
// par `World.Step`, où la direction est écrite et non lue.
type planche struct {
	monde  *game.World
	grille *game.CostGrid
	tuile  [2]int
	tampon *ebiten.Image
	ecrit  bool
}

// Update écrit les vues, puis rend la fin de partie.
func (p *planche) Update() error {
	if p.ecrit {
		return ebiten.Termination
	}
	for _, v := range vues {
		if err := p.vue(v); err != nil {
			return err
		}
	}
	p.ecrit = true
	return nil
}

// Draw ne dessine rien : la fenêtre n'est ouverte que pour le contexte
// graphique.
func (p *planche) Draw(*ebiten.Image) {}

// Layout donne au tampon la taille de celui du jeu.
func (p *planche) Layout(_, _ int) (int, int) { return render.Largeur, render.Hauteur }

// vue pose la scène, la dessine et écrit le fichier.
//
// Le joueur est posé au centre de sa case et non sur son coin, faute de quoi la
// planche montrerait un cas qu'aucune partie ne produit. L'écran vient après
// lui, et non l'inverse : c'est son montage qui cadre.
func (p *planche) vue(v vue) error {
	p.monde.Place(game.FromInt(v.u)+game.One/2, game.FromInt(v.v)+game.One/2)
	render.NewScreen(p.monde, p.grille, p.tuile).Draw(p.tampon)

	chemin := filepath.Join(sortie, v.nom+".png")

	// Le chemin n'a aucune part variable : `sortie` est une constante et le nom
	// vient de la table ci-dessus. `os.Root` ne fermerait donc rien qui soit
	// ouvert, et l'ajouter pour taire l'avertissement mettrait un mécanisme là
	// où il n'y a pas de question.
	f, err := os.OpenFile(chemin, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600) // #nosec G304
	if err != nil {
		return err
	}

	// La fermeture est signalée, mais elle ne masque pas l'écriture : un encodage
	// qui échoue dit ce qui ne va pas, là où la fermeture d'un fichier déjà
	// fautif ne dirait que le symptôme.
	err = png.Encode(f, p.tampon)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return fmt.Errorf("%s: %w", chemin, err)
	}
	fmt.Println(chemin)
	return nil
}

// main écrit la planche, ou sort en échec.
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run monte une partie par le même chemin que le binaire, puis lance la planche.
func run() error {
	if err := os.MkdirAll(sortie, 0o750); err != nil {
		return err
	}

	partie, err := session.Open(cohue.Assets, cohue.StartingLevel, capacite, tirs)
	if err != nil {
		return err
	}

	ebiten.SetWindowTitle("Cohue — planche")
	ebiten.SetWindowSize(render.Largeur, render.Hauteur)
	return ebiten.RunGame(&planche{
		monde:  partie.World,
		grille: partie.Grid,
		tuile:  partie.Tile,
		tampon: ebiten.NewImage(render.Largeur, render.Hauteur),
	})
}
