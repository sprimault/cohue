// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le décor d'un lieu, prêt à peindre : ses images converties une fois, le coin
// où chacune se pose, et ce qui départage le sol de ce qui le surplombe.

package render

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/sprimault/cohue/internal/level"
	"github.com/sprimault/cohue/internal/sprite"
)

// Terrain est le décor cuit d'un lieu, résolu en images.
//
// Il existe pour que `Screen` n'ait ni à connaître le catalogue de formes, ni à
// résoudre un nom par case : la carte cuite cite des index, et la résolution se
// fait une fois au montage.
type Terrain struct {
	taille [2]int
	carte  *level.Tilemap
	// formes est alignée sur `Tilemap.Shapes()` : l'index d'une case y désigne
	// directement ce qu'il faut dessiner.
	formes []forme
	// sol est le comblement du thème, nul quand aucune forme n'en a besoin.
	sol *forme
}

// forme est ce qu'une case donne à peindre.
type forme struct {
	image  *ebiten.Image
	dx, dy int
	// nue dit que la forme ne remplit pas le losange de sa case, donc qu'il faut
	// peindre le sol du thème avant elle.
	nue bool
	// elevee dit que la forme dépasse du sol, donc qu'elle entre dans le tri en
	// profondeur au lieu d'être peinte en passe préalable.
	//
	// **Le critère se dérive de l'élévation plutôt que de se déclarer.** Un
	// champ « triable » dans le manifeste serait une seconde description de ce
	// que l'élévation dit déjà, et l'aplatir à zéro pour épargner un tri ferait
	// dire au générateur qu'un trottoir n'a pas de relief. Quatre formes de
	// catégorie « sol » ont une élévation non nulle et entrent donc dans le
	// tri : c'est exact, et seulement coûteux.
	elevee bool
	// hauteurSol est la hauteur, en pixels d'écran, de la surface qu'on marche
	// sur cette case — ce sur quoi un marquage au sol se pose. Elle vient du
	// manifeste par `sprite`, qui porte l'arithmétique et le test.
	hauteurSol int
	// masque est la forme de l'image, ce qui décide si elle recouvre un
	// personnage. Un mur remplit sa boîte, un pilier n'en occupe qu'une bande
	// étroite.
	masque *sprite.Mask
	// cache dit que la forme dépasse la hauteur d'un personnage, donc qu'elle en
	// dissimule un au lieu de simplement empiéter sur sa case.
	//
	// **Le manifeste le déclare et le rendu en décide**, ce que sa godoc annonce
	// depuis l'étape 5 sans que rien ne le lise : le champ constate qu'une forme
	// peut cacher, la silhouette est la réponse qu'on y apporte. Vingt-sept
	// formes le portent, aucune sous vingt-cinq pixels d'élévation.
	cache bool
}

// NewTerrain résout la carte cuite d'un lieu en images posables.
//
// La conversion vers Ebitengine a lieu ici et une seule fois : `sprite` rend des
// images de la bibliothèque standard pour rester vérifiable sans écran, et les
// convertir à chaque case coûterait une texture par image dessinée.
func NewTerrain(carte *level.Tilemap, tuiles *sprite.Tileset) (*Terrain, error) {
	sol := &Terrain{
		taille: tuiles.TileSize(),
		carte:  carte,
		formes: make([]forme, 0, len(carte.Shapes())),
	}
	for _, nom := range carte.Shapes() {
		f, err := resoudre(tuiles, nom)
		if err != nil {
			return nil, err
		}
		sol.formes = append(sol.formes, f)
	}

	// Le sol n'est résolu que si le thème en déclare un. Le chargement l'exige
	// dès qu'une forme en a besoin, donc son absence dit qu'il n'y a rien à
	// combler et non qu'il faudra s'en passer.
	if nom := carte.Ground(); nom != "" {
		f, err := resoudre(tuiles, nom)
		if err != nil {
			return nil, err
		}
		sol.sol = &f
	}
	return sol, nil
}

// resoudre convertit une forme du catalogue en ce que le dessin pose.
func resoudre(tuiles *sprite.Tileset, nom string) (forme, error) {
	tuile, connue := tuiles.Tile(nom)
	if !connue {
		return forme{}, fmt.Errorf("decor : la forme %q du lieu n'est pas au catalogue", nom)
	}
	return forme{
		image:      ebiten.NewImageFromImage(tuile.Image),
		dx:         tuile.Offset[0],
		dy:         tuile.Offset[1],
		nue:        !tuile.Covers,
		elevee:     tuile.Elevation != 0,
		hauteurSol: tuile.GroundHeight,
		masque:     sprite.NewMask(tuile.Image),
		cache:      tuile.Masking,
	}, nil
}

// TileSize rend la taille de tuile du décor, celle dont la projection dépend.
func (t *Terrain) TileSize() [2]int { return t.taille }

// formeDe rend ce qu'une case donne à peindre, et dit qu'il n'y a rien hors de
// la carte.
//
// Rien, et non une forme par défaut : une case qu'aucune pièce ne pose se
// traverse et ne se dessine pas, ce que le contrôle de couverture refuse par
// ailleurs. Lui inventer un sol ici ferait disparaître le seul symptôme du
// trou.
func (t *Terrain) formeDe(u, v int) (forme, bool) {
	i := t.carte.At(u, v)
	if i < 0 {
		return forme{}, false
	}
	return t.formes[i], true
}
