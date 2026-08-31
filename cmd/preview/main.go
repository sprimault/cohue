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
// changement ne dirait rien. Rien n'y est tiré au sort : la horde vient du semis
// régulier du montage, et les vues qui jouent des pas les jouent avec une
// direction nulle. Le jour où un tirage entrera dans la simulation, sa graine
// devra être fixée pour que cette propriété tienne.
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

// vue est une scène à écrire : où poser le joueur, combien de pas jouer avant de
// dessiner, et le nom du fichier qui en sort.
type vue struct {
	nom   string
	u, v  int
	ticks int
}

// vues énumère ce que la planche donne à relire.
//
// Le centre montre la projection sans rien qui la borde ; les quatre suivantes
// posent le joueur près d'un coin du losange, là où la caméra bute et où le
// joueur se décentre. Ce sont les seuls endroits où le cadrage décide de quelque
// chose, et ils couvrent les deux axes dans les deux sens.
//
// **La mêlée est la seule qui juge l'exception du joueur**, et c'est pourquoi
// elle joue des pas : la horde semée converge, et il faut qu'elle soit arrivée
// pour qu'on voie si le personnage reste devant ce qui l'entoure. Une horde qui
// approche encore ne le dirait pas — elle est derrière lui ou devant lui, jamais
// autour. Quatre secondes suffisent à ce que les plus proches le rejoignent sans
// que l'arme de base en ait abattu assez pour dégager la place.
var vues = []vue{
	{nom: "centre", u: 16, v: 16},
	{nom: "nord", u: 2, v: 2},
	{nom: "ouest", u: 2, v: 29},
	{nom: "est", u: 29, v: 2},
	{nom: "sud", u: 29, v: 29},
	{nom: "melee", u: 16, v: 16, ticks: 4 * game.TPS},
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
//
// **Chaque vue monte sa propre partie**, et n'hérite donc pas de ce que les
// précédentes ont joué. Une vue qui avance de quatre secondes laisserait sinon
// une horde déplacée aux suivantes, et l'ordre de la table déciderait de ce
// qu'on voit — ce qui est exactement le genre de dépendance cachée qu'une
// planche ne doit pas avoir. Le montage coûte quelques millisecondes.
type planche struct {
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

// vue monte une partie, y pose la scène, la dessine et écrit le fichier.
//
// Le joueur est posé au centre de sa case et non sur son coin, faute de quoi la
// planche montrerait un cas qu'aucune partie ne produit. Les pas se jouent
// ensuite, avec une direction nulle : la horde avance, le joueur ne bouge pas, et
// rien de ce qui se passe ne dépend de ce qu'on presse. L'écran vient en dernier,
// puisque c'est son montage qui cadre.
func (p *planche) vue(v vue) error {
	partie, err := session.Open(cohue.Assets, cohue.StartingLevel)
	if err != nil {
		return err
	}
	partie.World.Place(game.FromInt(v.u)+game.One/2, game.FromInt(v.v)+game.One/2)
	for range v.ticks {
		partie.World.Step(game.Vec{})
	}
	render.NewScreen(partie.World, partie.Grid, partie.Tile).Draw(p.tampon)

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

// run ouvre la fenêtre que le dessin exige, puis lance la planche.
//
// La partie, elle, se monte une fois par vue : voir la godoc de `planche`.
func run() error {
	if err := os.MkdirAll(sortie, 0o750); err != nil {
		return err
	}

	ebiten.SetWindowTitle("Cohue — planche")
	ebiten.SetWindowSize(render.Largeur, render.Hauteur)
	return ebiten.RunGame(&planche{
		tampon: ebiten.NewImage(render.Largeur, render.Hauteur),
	})
}
