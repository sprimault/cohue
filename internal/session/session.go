// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le montage d'une partie : les manifestes lus, le lieu cuit, le monde bâti et
// le joueur posé. Ce qu'il rend suffit à ouvrir une fenêtre ou à écrire une
// image, et rien d'autre n'a à savoir dans quel ordre tout cela se charge.

// Package session monte une partie à partir d'un système de fichiers.
//
// Il existe parce que deux programmes en ont besoin — le jeu et la planche de
// relecture — et qu'un montage recopié aurait fini par diverger : la planche
// aurait alors relu une partie que le jeu ne joue pas, ce qui lui retire tout
// intérêt. L'écran de choix des lieux en sera le troisième appelant, avec un
// lieu qui ne sera plus celui de départ.
//
// Il ne connaît ni fenêtre ni rendu, et ne peut donc pas en dépendre : ce qu'il
// rend est de la simulation et une taille de tuile.
package session

import (
	"io/fs"
	"log/slog"

	"github.com/sprimault/cohue"
	"github.com/sprimault/cohue/internal/game"
	"github.com/sprimault/cohue/internal/level"
)

// Session est une partie montée, prête à tourner.
//
// La taille de tuile voyage avec le monde parce qu'elle vient du même
// chargement : elle est dans le manifeste de décor, que le rendu ne lit pas
// lui-même — il ne connaît que des profils et des cycles, et une taille reçue.
type Session struct {
	World *game.World
	Grid  *game.CostGrid
	Tile  [2]int
}

// Open monte une partie sur le lieu donné, avec les bassins demandés.
//
// L'ordre n'est pas libre : le catalogue de coûts vient du manifeste de décor et
// le chargeur de lieux en a besoin, si bien qu'un lieu ne peut pas se cuire avant
// que le décor soit lu.
//
// Les manifestes sont lus au montage et non à la première vague : un fichier que
// le binaire refuse doit le dire tout de suite, pas trois minutes après le début
// d'une partie.
func Open(fsys fs.FS, lieu string, ennemis, tirs int) (*Session, error) {
	decor, couts, err := level.LoadDecor(fsys, cohue.DecorManifest)
	if err != nil {
		return nil, err
	}
	grille, err := level.NewLoader(fsys, couts).Load(lieu)
	if err != nil {
		return nil, err
	}
	slog.Info("lieu chargé", "name", lieu, "largeur", grille.Width(), "hauteur", grille.Height())

	profils, err := game.LoadProfiles(fsys, cohue.CharacterManifest)
	if err != nil {
		return nil, err
	}
	slog.Info("profils chargés", "enemies", len(profils.Enemies))

	armes, err := game.LoadWeapons(fsys, cohue.WeaponManifest)
	if err != nil {
		return nil, err
	}
	slog.Info("armes chargées", "base", armes.Base.Key)

	monde := game.NewWorld(profils, armes.Base, grille, ennemis, tirs)
	placer(monde, grille)

	return &Session{World: monde, Grid: grille, Tile: decor.Tile}, nil
}

// placer pose le joueur au centre du lieu.
//
// Au centre, faute de mieux : le format ne porte pas encore d'ancrage de départ,
// et l'inventer maintenant demanderait de trancher un champ sans usage réel. La
// position de départ dépendra du lieu, de la campagne et de la porte par laquelle
// on entre — trois choses que l'étape 8 apporte.
//
// Au centre de la case et non sur son coin : c'est là que se tient une entité,
// et poser le joueur sur un coin montrerait un cas qu'aucune partie ne produit.
func placer(monde *game.World, grille *game.CostGrid) {
	monde.Place(
		game.FromInt(grille.Width()/2)+game.One/2,
		game.FromInt(grille.Height()/2)+game.One/2,
	)
}
