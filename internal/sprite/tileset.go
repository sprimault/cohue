// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les formes du décor : leurs images, confrontées au manifeste, et le coin où
// les poser. Comme les feuilles de personnages, rien ici n'affiche.

package sprite

import (
	"fmt"
	"image"
	"io/fs"
	"maps"
	"math"
	"path"
	"slices"

	"github.com/sprimault/cohue/internal/level"
	"github.com/sprimault/cohue/internal/manifest"
)

// Tile est une forme du décor, prête à être posée.
type Tile struct {
	// Image est le dessin de la forme.
	Image image.Image
	// Offset est le coin haut-gauche où poser l'image, en pixels, relatif au
	// sommet nord de la case qui la porte.
	//
	// Il est calculé ici plutôt qu'au dessin parce que c'est de l'arithmétique
	// de manifeste — un ancrage, une emprise, une taille de tuile — et que le
	// rendu n'a aucun test pour la garder. Ce qui se pose sur l'écran vaut mieux
	// d'être fixé là où un test peut l'épingler.
	Offset [2]int
	// Elevation est la hauteur au-dessus du sol, en pixels. Zéro dit que la
	// forme est un sol, donc qu'elle ne peut rien recouvrir.
	Elevation int
	// Covers dit que la forme remplit le losange de sa case, donc qu'il n'y a
	// rien à peindre dessous.
	Covers bool
}

// Tileset porte les formes du décor, par nom.
//
// **Tout le catalogue est lu, pas seulement ce qu'un lieu emploie.** Une forme
// dont l'image manque ou dément son manifeste est un défaut du dépôt et non du
// lieu qui l'aurait citée : la découvrir au premier lieu qui la pose la ferait
// remonter chez un joueur plutôt qu'au montage. Soixante et une images sont par
// ailleurs déjà dans le binaire, décodées ou non.
type Tileset struct {
	// taille est la taille de tuile du manifeste, celle contre laquelle les
	// coins ont été calculés. Elle voyage avec eux plutôt que d'être redemandée
	// à l'appelant : deux tailles pour un même jeu de formes poseraient tout de
	// travers, et rien ne le dirait.
	taille [2]int
	tuiles map[string]Tile
}

// LoadTiles lit les images du décor et calcule où chacune se pose.
//
// La racine est le dossier des thèmes : une forme vit dans celui que son
// manifeste lui donne, et son nom de fichier est le sien. C'est le manifeste qui
// dit où chercher, jamais une table de chemins écrite à côté.
//
// Les manquements sont accumulés, comme pour les feuilles de personnages : qui
// régénère le décor veut la liste de ce qui cloche.
func LoadTiles(fsys fs.FS, racine string, decor *level.Decor) (*Tileset, error) {
	jeu := &Tileset{taille: decor.Tile, tuiles: make(map[string]Tile, len(decor.Shapes))}

	var manques []string
	for _, nom := range slices.Sorted(maps.Keys(decor.Shapes)) {
		forme := decor.Shapes[nom]
		chemin := path.Join(racine, forme.Theme, nom+".png")
		img, err := lire(fsys, chemin, forme.Size)
		if err != nil {
			manques = append(manques, err.Error())
			continue
		}
		if defaut := controlerForme(forme); defaut != "" {
			manques = append(manques, nom+" : "+defaut)
			continue
		}
		jeu.tuiles[nom] = Tile{
			Image:     img,
			Offset:    coin(forme, decor.Tile),
			Elevation: forme.Elevation,
			Covers:    forme.Covers(),
		}
	}
	if len(manques) > 0 {
		return nil, &manifest.Invalid{Path: racine, Missing: manques}
	}
	return jeu, nil
}

// TileSize rend la taille de tuile contre laquelle les coins ont été calculés.
func (t *Tileset) TileSize() [2]int { return t.taille }

// Tile rend une forme du décor, ou dit qu'elle n'existe pas.
//
// Le second retour n'est pas une précaution : un lieu tiers cite les formes
// qu'il veut, et ce qui l'assemble a besoin de savoir laquelle manque pour le
// dire à son auteur.
func (t *Tileset) Tile(nom string) (Tile, bool) {
	tuile, connue := t.tuiles[nom]
	return tuile, connue
}

// controlerForme refuse une forme dont on ne saurait rien poser.
//
// L'ancrage se contrôle contre l'image et non dans l'absolu : c'est lui qu'on
// pose sur la case, et un ancrage hors du dessin décalerait la forme d'autant
// sans que rien ne dise pourquoi. L'emprise, elle, ne peut pas être nulle — le
// sommet bas d'un losange sans surface n'est nulle part.
func controlerForme(f level.Shape) string {
	switch {
	case f.Anchor[0] < 0 || f.Anchor[0] >= f.Size[0],
		f.Anchor[1] < 0 || f.Anchor[1] >= f.Size[1]:
		return fmt.Sprintf("ancrage %v, hors d'une image de %v", f.Anchor, f.Size)
	case f.Footprint[0] <= 0 || f.Footprint[1] <= 0:
		return fmt.Sprintf("emprise %v, une forme occupe du sol", f.Footprint)
	case f.Elevation < 0:
		return fmt.Sprintf("elevation %d, une forme ne s'enfonce pas", f.Elevation)
	}
	return ""
}

// coin rend le coin haut-gauche où poser une forme, relatif au sommet nord de
// sa case.
//
// **L'emprise se centre sur la case**, ce qui est la convention de `poser` dans
// les primitives isométriques : une position y désigne le centre de l'emprise
// d'un objet, et non son coin. Un pilier d'une demi-tuile se pose donc au milieu
// de son losange plutôt que dans son quart sud, et une forme d'une tuile pleine
// retombe exactement sur le sommet bas de la case — ce que faisait déjà le sol
// peint en aplat.
//
// L'ancrage du manifeste est le pixel bas-centre de l'image, c'est-à-dire ce
// sommet bas à une rangée près : la dernière rangée d'une image de n pixels est
// la n-1, d'où le pixel retranché en ordonnée.
//
// Ce n'est exact que pour une emprise carrée, la seule que le décor livré
// emploie : le bas-centre d'une image cesse d'être le sommet bas du losange dès
// que les deux côtés diffèrent. Le manifeste écrivant l'ancrage de la même
// façon pour toutes, l'écart appartient au générateur et se corrigera là-bas, le
// jour où une pièce posera une forme allongée.
func coin(f level.Shape, tuile [2]int) [2]int {
	demiLargeur, demiHauteur := float64(tuile[0])/2, float64(tuile[1])/2
	ex, ey := f.Footprint[0], f.Footprint[1]

	dx := (ex - ey) * demiLargeur / 2
	dy := (2 + ex + ey) * demiHauteur / 2
	return [2]int{
		int(math.Round(dx)) - f.Anchor[0],
		int(math.Round(dy)) - 1 - f.Anchor[1],
	}
}

// lire ouvre l'image d'une forme et la confronte à la taille annoncée.
//
// **La taille vient du manifeste, jamais du fichier.** C'est le contrat que le
// manifeste existe pour tenir : une forme redessinée plus haute sans que son
// manifeste bouge se poserait décalée de la différence, et rien ne le dirait.
func lire(fsys fs.FS, chemin string, taille [2]int) (image.Image, error) {
	f, err := fsys.Open(chemin)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", chemin, err)
	}
	img, _, err := image.Decode(f)
	_ = f.Close()
	if err != nil {
		return nil, fmt.Errorf("%s : %w", chemin, err)
	}

	if vue := img.Bounds().Size(); vue.X != taille[0] || vue.Y != taille[1] {
		return nil, fmt.Errorf("%s : %dx%d, le manifeste annonce %dx%d",
			chemin, vue.X, vue.Y, taille[0], taille[1])
	}
	return img, nil
}
