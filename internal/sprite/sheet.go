// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les feuilles de sprites : les bandes d'un profil, découpées en images et
// confrontées à ce que son manifeste annonce. Rien ici n'affiche.

package sprite

import (
	"fmt"
	"image"
	_ "image/png" // le format des bandes, décodé par image.Decode
	"io/fs"
	"maps"
	"path"
	"slices"

	"github.com/sprimault/cohue/internal/game"
	"github.com/sprimault/cohue/internal/manifest"
)

// Sheet porte les images d'un profil, prêtes à être posées.
//
// **Il ne lie pas Ebitengine, et c'est ce qui le rend vérifiable.**
// `internal/render` n'a aucun test par doctrine — importer la bibliothèque
// initialise GLFW, qui panique sans écran —, si bien qu'un découpage écrit
// là-bas serait invérifiable à jamais. Ce paquet rend des `image.Image` de la
// bibliothèque standard ; le rendu n'a plus qu'à les convertir.
type Sheet struct {
	figure game.Figure
	// images sont les bandes découpées, par cycle, direction et variante. La clé
	// est bâtie par `cle` plutôt que d'imbriquer trois tables : ce qu'on
	// demande à cette structure est une image et jamais une liste, et trois
	// niveaux de map coûteraient à écrire ce qu'ils n'apportent pas à lire.
	images map[string][]image.Image
}

// Load lit les bandes d'un profil et les découpe.
//
// **Tout est chargé au montage, jamais en jeu.** Le budget d'allocation interdit
// d'ouvrir un fichier dans la boucle, et une image manquante y serait de toute
// façon découverte au pire moment — la première fois qu'une créature meurt.
//
// Les manquements sont accumulés plutôt que rendus au premier : qui régénère les
// figurines veut la liste de ce qui cloche, pas un aller-retour par bande.
func Load(fsys fs.FS, racine string, figure game.Figure) (*Sheet, error) {
	var manques []string
	dire := func(format string, args ...any) {
		manques = append(manques, fmt.Sprintf(format, args...))
	}

	controler(figure, dire)
	if len(manques) > 0 {
		return nil, &manifest.Invalid{Path: path.Join(racine, figure.Key), Missing: manques}
	}

	feuille := &Sheet{figure: figure, images: map[string][]image.Image{}}
	for _, cycle := range slices.Sorted(maps.Keys(figure.Cycles)) {
		c := figure.Cycles[cycle]
		for _, direction := range figure.Directions {
			for variante := range figure.Variants {
				chemin := cheminDe(racine, figure, cycle, direction, variante)
				images, err := decouper(fsys, chemin, figure.Side, c.Frames)
				if err != nil {
					dire("%s", err)
					continue
				}
				feuille.images[cle(cycle, direction, variante)] = images
			}
		}
	}

	if len(manques) > 0 {
		return nil, &manifest.Invalid{Path: path.Join(racine, figure.Key), Missing: manques}
	}
	return feuille, nil
}

// Frame rend une image d'un cycle, ou dit qu'elle n'existe pas.
//
// **Le second retour n'est pas une précaution.** Tous les profils n'ont pas les
// mêmes cycles — le Molosse n'a ni repos ni attaque —, si bien que demander un
// cycle absent est un cas ordinaire du rendu et non une erreur de programmation.
// C'est à l'appelant de choisir son repli, parce que lui seul sait par quoi
// remplacer une pose qui n'existe pas.
func (s *Sheet) Frame(cycle, direction string, variante, image int) (image.Image, bool) {
	bande, ok := s.images[cle(cycle, direction, variante)]
	if !ok || image < 0 || image >= len(bande) {
		return nil, false
	}
	return bande[image], true
}

// Figure rend ce que le manifeste disait de ce profil.
func (s *Sheet) Figure() game.Figure { return s.figure }

// controler refuse une figure dont on ne saurait rien tirer.
//
// **Le contrôle vit ici et pas dans `internal/game`**, qui porte ces champs sans
// jamais les ouvrir : un paquet qui refuserait un manifeste sur une donnée qu'il
// ne lit pas jugerait à la place d'un autre, et les deux règles finiraient par
// diverger.
//
// Le point d'appui se contrôle contre la case et non dans l'absolu : c'est lui
// que le rendu pose sur la position du monde, et un appui hors de l'image
// décalerait la créature d'autant sans que rien ne dise pourquoi.
func controler(f game.Figure, dire func(string, ...any)) {
	if f.Side < 1 {
		dire("cote : %d, une image a un côté", f.Side)
	}
	if f.Variants < 1 {
		dire("variantes : %d, un profil a au moins une teinte", f.Variants)
	}
	if len(f.Directions) == 0 {
		dire("directions : aucune, un sprite s'oriente")
	}
	if len(f.Cycles) == 0 {
		dire("cycles : aucun, un profil se dessine")
	}
	if f.Side > 0 {
		for i, borne := range f.Anchor {
			if borne < 0 || borne >= f.Side {
				dire("appui[%d] : %d, hors d'une case de %d", i, borne, f.Side)
			}
		}
	}
	// Trié, parce que deux chargements du même manifeste doivent rendre la même
	// liste : le parcours d'une map ne le garantit pas.
	for _, nom := range slices.Sorted(maps.Keys(f.Cycles)) {
		if c := f.Cycles[nom]; c.Frames < 1 {
			dire("cycles.%s.images : %d, une bande porte au moins une image", nom, c.Frames)
		}
	}
}

// cheminDe bâtit le chemin d'une bande.
//
// Un profil à teinte unique n'a pas de sous-dossier : le chargeur n'a rien à
// savoir des variantes qui n'existent pas, et c'est la règle que suivent déjà
// les générateurs.
func cheminDe(racine string, f game.Figure, cycle, direction string, variante int) string {
	nom := fmt.Sprintf("%s_%s.png", cycle, direction)
	if f.Variants > 1 {
		return path.Join(racine, f.Key, fmt.Sprintf("v%d", variante), nom)
	}
	return path.Join(racine, f.Key, nom)
}

// decouper ouvre une bande et la tranche en images de côté fixe.
//
// **La taille est confrontée à ce que le manifeste annonce, jamais déduite du
// fichier.** Une bande de 320 pixels est indéchiffrable seule — cinq images de
// 64 ou quatre de 80 —, et c'est tout l'objet du manifeste-contrat : le moteur
// ne connaît que des profils et des cycles.
//
// Le découpage passe par `SubImage`, qui partage les pixels de la bande au lieu
// de les recopier : cinq cents bandes tiennent alors dans ce que pèsent les
// fichiers.
func decouper(fsys fs.FS, chemin string, cote, images int) ([]image.Image, error) {
	f, err := fsys.Open(chemin)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", chemin, err)
	}
	bande, _, err := image.Decode(f)
	// Fermé ici et non en `defer`, comme le fait déjà le chargeur d'icônes : ce
	// qui suit ne lit plus le fichier, et une lecture seule n'a rien à rendre à
	// la fermeture.
	_ = f.Close()
	if err != nil {
		return nil, fmt.Errorf("%s : %w", chemin, err)
	}

	taille := bande.Bounds().Size()
	if taille.X != images*cote || taille.Y != cote {
		return nil, fmt.Errorf("%s : %dx%d, attendu %dx%d pour %d image(s) de %d",
			chemin, taille.X, taille.Y, images*cote, cote, images, cote)
	}

	tranche, ok := bande.(interface {
		SubImage(r image.Rectangle) image.Image
	})
	if !ok {
		return nil, fmt.Errorf("%s : format sans sous-image, %T", chemin, bande)
	}

	origine := bande.Bounds().Min
	decoupe := make([]image.Image, images)
	for i := range images {
		coin := origine.Add(image.Pt(i*cote, 0))
		decoupe[i] = tranche.SubImage(image.Rectangle{Min: coin, Max: coin.Add(image.Pt(cote, cote))})
	}
	return decoupe, nil
}

// cle bâtit l'index d'une bande.
func cle(cycle, direction string, variante int) string {
	return fmt.Sprintf("%s/%s/%d", cycle, direction, variante)
}
