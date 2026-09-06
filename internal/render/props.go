// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les objets d'une partie, prêts à peindre : leurs images converties une fois,
// et les noms de catalogue que le rendu emploie.

package render

import (
	"fmt"
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/sprimault/cohue/internal/game"
	"github.com/sprimault/cohue/internal/sprite"
)

// Les objets que le rendu pose, par leur nom de catalogue.
//
// **Ce sont les seuls noms d'objet écrits dans du Go**, comme les cycles et les
// directions le sont ailleurs, et pour la même raison : la simulation tient des
// bassins — des gemmes, des projectiles —, et c'est le rendu qui sait lequel se
// dessine avec quoi. Elle n'a pas à savoir qu'une image existe.
const (
	objetGemme     = "gemme"
	objetAimant    = "aimant"
	objetCaisse    = "caisse"
	objetTir       = "projectile_base"
	objetTirHorde  = "projectile_ennemi"
	objetEtincelle = "etincelle"
)

// Stage porte les objets d'une partie, convertis une fois.
//
// Une table par nom et non une tranche par bassin : un bassin peut poser deux
// objets — une caisse et sa ruine —, et un objet peut n'appartenir à aucun,
// comme l'étincelle d'un impact.
type Stage struct {
	objets map[string]prop
}

// prop est ce qu'un objet donne à peindre.
type prop struct {
	images []*ebiten.Image
	dx, dy int
	cycle  game.Cycle
}

// NewStage résout le catalogue d'objets en images posables.
//
// Il ne retient que ce que le rendu pose : les armes au sol, les particules et
// les ruines attendent le mécanisme qui les fera exister, et les charger d'avance
// serait payer des textures pour ce que rien ne dessine.
func NewStage(fsys fs.FS, racine, chemin string) (*Stage, error) {
	_, catalogue, err := sprite.LoadObjects(fsys, racine, chemin)
	if err != nil {
		return nil, err
	}

	scene := &Stage{objets: map[string]prop{}}
	for _, nom := range []string{
		objetGemme, objetAimant, objetCaisse, objetTir, objetTirHorde, objetEtincelle,
	} {
		objet, connu := catalogue.Prop(nom)
		if !connu {
			return nil, fmt.Errorf("objets : « %s » n'est pas au catalogue", nom)
		}
		images := make([]*ebiten.Image, 0, len(objet.Images))
		for _, img := range objet.Images {
			images = append(images, ebiten.NewImageFromImage(img))
		}
		scene.objets[nom] = prop{
			images: images,
			dx:     objet.Offset[0],
			dy:     objet.Offset[1],
			cycle:  objet.Cycle,
		}
	}
	return scene, nil
}

// image rend l'image d'un objet au tick donné, décalée par une identité.
//
// **Le scintillement se dérive du tick**, comme les cycles d'un personnage, et
// le décalage par l'identifiant vaut ici ce qu'il valait là-bas : deux cents
// gemmes qui pulseraient ensemble feraient un stroboscope, quand décalées elles
// font un tapis qui bouge.
//
// Un objet d'une seule image ignore l'un et l'autre : sa cadence est nulle, et
// `Loop` rend alors la première.
func (s *Stage) image(nom string, tick game.Tick, identite int) (prop, *ebiten.Image) {
	return s.rendre(nom, func(objet prop) int {
		return sprite.Loop(objet.cycle, tick, identite)
	})
}

// effet rend l'image d'une animation brève, cadencée par le décompte de l'état
// qui la porte.
//
// C'est la même dérivation que celle d'un cycle d'attaque : l'animation s'achève
// quand l'état s'achève, et un état plus court qu'elle la fait entrer en cours
// de route. L'étincelle est dans ce dernier cas — trois images de deux ticks
// pour un éclair qui en dure quatre —, si bien qu'on en voit la fin. Ce qu'elle
// doit dire est que le tir a porté, pas comment l'impact se déroule.
func (s *Stage) effet(nom string, reste game.Tick) (prop, *ebiten.Image) {
	return s.rendre(nom, func(objet prop) int {
		return sprite.Once(objet.cycle, reste)
	})
}

// rendre résout un objet et l'image que son cadencement désigne.
func (s *Stage) rendre(nom string, quelle func(prop) int) (prop, *ebiten.Image) {
	objet, connu := s.objets[nom]
	if !connu || len(objet.images) == 0 {
		return prop{}, nil
	}
	i := quelle(objet)
	if i < 0 || i >= len(objet.images) {
		i = 0
	}
	return objet, objet.images[i]
}
