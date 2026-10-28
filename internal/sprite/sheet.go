// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les feuilles de sprites : les bandes d'un profil, rangées dans l'atlas et
// confrontées à ce que son manifeste annonce. Rien ici n'affiche.

package sprite

import (
	"fmt"
	"io/fs"
	"maps"
	"path"
	"slices"

	"github.com/sprimault/lozengine/raster"
	"github.com/sprimault/lozengine/scene"

	"github.com/sprimault/cohue/internal/game"
	"github.com/sprimault/cohue/internal/manifest"
)

// Sheet porte les images d'un profil, prêtes à être posées.
//
// **Il ne lie aucune fenêtre, et c'est ce qui le garde vérifiable.** Ce qu'il
// touche du moteur est le rastériseur et la description d'une scène, dont aucun
// ne demande d'écran ni de contexte graphique : un découpage écrit dans le rendu
// serait resté invérifiable, et il ne l'est plus.
type Sheet struct {
	figure game.Figure
	// ancre pose l'image autour du point du monde, et elle est la même pour
	// toutes les bandes d'un profil : l'appui et le côté viennent du manifeste,
	// pas de la bande.
	ancre [2]int16
	// bandes sont les planches rangées, par cycle, direction et variante. La clé
	// est bâtie par `cle` plutôt que d'imbriquer trois tables : ce qu'on demande
	// à cette structure est une image et jamais une liste, et trois niveaux de
	// map coûteraient à écrire ce qu'ils n'apportent pas à lire.
	bandes map[string]planche
	// poses sont les cases de chaque bande, relevées une fois au chargement.
	// Elles dérivent des planches et ne peuvent donc pas en diverger.
	poses map[string][]Pose
	// formes et bords sont les dérivés d'une bande, absents tant que personne ne
	// les a demandés par `Silhouette`.
	formes map[string]planche
	bords  map[string]planche
}

// Load lit les bandes d'un profil, les range dans l'atlas et les découpe.
//
// **Tout est chargé au montage, jamais en jeu.** Le budget d'allocation interdit
// d'ouvrir un fichier dans la boucle, et une image manquante y serait de toute
// façon découverte au pire moment — la première fois qu'une créature meurt.
//
// **L'atlas est reçu et non créé** : il est unique et partagé par toute la
// scène, et chaque porteur ne garde que l'index que `raster.Atlas.Add` lui a
// rendu. Aucun index ne s'écrit donc en dur, et l'ordre de chargement est libre.
//
// Les manquements sont accumulés plutôt que rendus au premier : qui régénère les
// figurines veut la liste de ce qui cloche, pas un aller-retour par bande.
func Load(fsys fs.FS, racine string, figure game.Figure, atlas *raster.Atlas) (*Sheet, error) {
	var manques []string
	dire := func(format string, args ...any) {
		manques = append(manques, fmt.Sprintf(format, args...))
	}

	controler(figure, dire)
	cote := [2]int{figure.Side, figure.Side}
	ancre, tient := ancre16(AnchorFromFooting(figure.Anchor, cote))
	if !tient {
		dire("appui %v dans une case de %d : l'ancre n'y tient pas",
			figure.Anchor, figure.Side)
	}
	if len(manques) > 0 {
		return nil, &manifest.Invalid{Path: path.Join(racine, figure.Key), Missing: manques}
	}

	feuille := &Sheet{
		figure: figure,
		ancre:  ancre,
		bandes: map[string]planche{},
		poses:  map[string][]Pose{},
	}
	for _, cycle := range slices.Sorted(maps.Keys(figure.Cycles)) {
		c := figure.Cycles[cycle]
		for _, direction := range figure.Directions {
			for variante := range figure.Variants {
				chemin := cheminDe(racine, figure, cycle, direction, variante)
				bande, err := ranger(fsys, atlas, chemin,
					[2]int{c.Frames * figure.Side, figure.Side}, cote)
				if err != nil {
					dire("%s", err)
					continue
				}
				k := cle(cycle, direction, variante)
				feuille.bandes[k] = bande
				feuille.poses[k] = bande.poses(ancre)
			}
		}
	}

	if len(manques) > 0 {
		return nil, &manifest.Invalid{Path: path.Join(racine, figure.Key), Missing: manques}
	}
	return feuille, nil
}

// Silhouette range dans l'atlas l'aplat et le bord de chaque bande.
//
// **Deux planches par bande et non par image**, ce que la borne de l'atlas
// impose : les cases d'une bande gardent leur découpe, si bien qu'une pose y
// prend le même rectangle que dans son dessin.
//
// Demandé plutôt que fait d'office : seul ce qu'on révèle en a besoin, et
// dériver le catalogue entier triplerait ses planches pour des dessins que rien
// ne pose.
func (s *Sheet) Silhouette(atlas *raster.Atlas) error {
	s.formes = make(map[string]planche, len(s.bandes))
	s.bords = make(map[string]planche, len(s.bandes))
	for _, k := range slices.Sorted(maps.Keys(s.bandes)) {
		forme, err := s.bandes[k].derive(atlas, Flatten)
		if err != nil {
			return fmt.Errorf("aplat de %s : %w", k, err)
		}
		bord, err := s.bandes[k].derive(atlas, Outline)
		if err != nil {
			return fmt.Errorf("bord de %s : %w", k, err)
		}
		s.formes[k], s.bords[k] = forme, bord
	}
	return nil
}

// Pose rend une image d'un cycle, ou dit qu'elle n'existe pas.
//
// **Le second retour n'est pas une précaution.** Tous les profils n'ont pas les
// mêmes cycles — le Molosse n'a ni repos ni attaque —, si bien que demander un
// cycle absent est un cas ordinaire du rendu et non une erreur de programmation.
// C'est à l'appelant de choisir son repli, parce que lui seul sait par quoi
// remplacer une pose qui n'existe pas. La valeur zéro d'un aspect ne dessinant
// rien, la rendre seule le laisserait sans moyen de le savoir.
func (s *Sheet) Pose(cycle, direction string, variante, image int) (Pose, bool) {
	bande, ok := s.poses[cle(cycle, direction, variante)]
	if !ok || image < 0 || image >= len(bande) {
		return Pose{}, false
	}
	return bande[image], true
}

// Shape rend l'aplat d'une image, ou dit qu'il n'existe pas.
//
// Sans masque, à la différence d'une pose : un aplat a la forme du dessin dont
// il est tiré, et c'est celui-là qu'on a déjà mesuré.
func (s *Sheet) Shape(cycle, direction string, variante, image int) (scene.Aspect, bool) {
	return s.derive(s.formes, cycle, direction, variante, image)
}

// Edge rend le bord d'une image, ou dit qu'il n'existe pas.
func (s *Sheet) Edge(cycle, direction string, variante, image int) (scene.Aspect, bool) {
	return s.derive(s.bords, cycle, direction, variante, image)
}

// derive rend une case d'une table de dérivés, ou dit qu'elle n'y est pas.
func (s *Sheet) derive(table map[string]planche, cycle, direction string, variante, image int) (scene.Aspect, bool) {
	p, ok := table[cle(cycle, direction, variante)]
	if !ok {
		return scene.Aspect{}, false
	}
	return p.aspect(image, s.ancre)
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

// cle bâtit l'index d'une bande.
func cle(cycle, direction string, variante int) string {
	return fmt.Sprintf("%s/%s/%d", cycle, direction, variante)
}
