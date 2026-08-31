// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package level

import (
	"errors"
	"fmt"
	"io/fs"
	"path"

	"github.com/sprimault/cohue/internal/game"
	"github.com/sprimault/cohue/internal/manifest"
)

// Les refus que l'appelant peut vouloir distinguer.
var (
	// ErrUnknownRoom signale un lieu qui cite une pièce absente du jeu.
	ErrUnknownRoom = errors.New("piece inconnue du jeu")
	// ErrEmptyLevel signale un lieu sans aucune pièce posée.
	ErrEmptyLevel = errors.New("lieu sans piece")
)

// Formats lus par ce binaire. Chacun a sa vie : un lieu circule entre joueurs,
// une pièce et un jeu restent dans le binaire.
const (
	FormatLevel = 1
	FormatRoom  = 1
	FormatSet   = 1
)

// LevelFile est le nom que porte le descripteur dans un dossier de lieu.
//
// Fixe, et non dérivé de l'identifiant : c'est ce qui permet de reconnaître un
// dossier de lieu sans l'ouvrir, et de renommer un lieu en renommant son
// dossier.
const LevelFile = "lieu.json"

// Loader lit un lieu et ses pièces dans un système de fichiers.
//
// Une interface de lecture plutôt qu'un chemin : les lieux livrés viennent d'un
// `embed.FS`, ceux d'un joueur d'un dossier, et les cas de test d'une carte en
// mémoire. Un seul chemin de code pour les trois, ce que la conception exige —
// une voie rapide pour le contenu livré ne serait exercée qu'à moitié.
type Loader struct {
	fsys fs.FS
	// couts dit ce que coûte la traversée d'une forme de décor. Il vient du
	// manifeste : le chargeur ne connaît aucun nom de tuile en dur.
	couts map[string]game.Cost
}

// NewLoader monte un chargeur sur un système de fichiers et un catalogue de coûts.
func NewLoader(fsys fs.FS, couts map[string]game.Cost) *Loader {
	return &Loader{fsys: fsys, couts: couts}
}

// Load lit le lieu que porte un dossier, ses pièces et son jeu, puis les cuit
// en grille de coûts.
//
// Un dossier, et non un fichier : c'est ce qui donne aux pièces un espace de
// noms local. À plat, deux auteurs qui nomment chacun leur pièce « carrefour »
// s'écraseraient, et un lieu ne pourrait pas circuler sans emporter le
// vocabulaire de tous les autres.
//
// L'ordre importe : décoder, valider, cuire. Le décodage s'arrête au premier
// écart — `encoding/json` ne sait pas faire autrement —, la validation liste
// tout ce qui manque en une fois, parce que c'est là que l'aller-retour coûte à
// qui met au point un niveau.
func (l *Loader) Load(dossier string) (*game.CostGrid, error) {
	nom := path.Base(dossier)
	if nom == "." || nom == "/" {
		return nil, fmt.Errorf("%q : un lieu se charge par son dossier, qui porte son nom", dossier)
	}

	chemin := path.Join(dossier, LevelFile)
	lieu, err := manifest.Decode[Level](l.fsys, chemin)
	if err != nil {
		return nil, err
	}
	if lieu.Format != FormatLevel {
		return nil, fmt.Errorf("%s: %w : %d, ce binaire lit la %d",
			chemin, manifest.ErrUnsupportedFormat, lieu.Format, FormatLevel)
	}
	if len(lieu.Placements) == 0 {
		return nil, fmt.Errorf("%s: %w", chemin, ErrEmptyLevel)
	}

	jeu, err := manifest.Decode[Set](l.fsys, path.Join(dossier, lieu.SetID+".json"))
	if err != nil {
		return nil, err
	}
	if jeu.Format != FormatSet {
		return nil, fmt.Errorf("%w : jeu de pièces en %d", manifest.ErrUnsupportedFormat, jeu.Format)
	}

	pieces := make([]*Room, 0, len(lieu.Placements))
	for _, pose := range lieu.Placements {
		piece, err := manifest.Decode[Room](l.fsys, path.Join(dossier, pose.RoomID+".json"))
		if err != nil {
			return nil, fmt.Errorf("%w : %s", ErrUnknownRoom, pose.RoomID)
		}
		pieces = append(pieces, piece)
	}

	if manques := valider(nom, lieu, jeu, pieces); len(manques) > 0 {
		return nil, &manifest.Invalide{Chemin: chemin, Manques: manques}
	}
	return cuire(lieu, jeu, pieces, l.couts), nil
}

// cuire assemble les pièces posées en une seule grille de coûts.
//
// Après quoi le moteur ne sait plus que le lieu était modulaire : le parcours du
// champ de flux tourne sur une grille ordinaire.
func cuire(lieu *Level, jeu *Set, pieces []*Room, couts map[string]game.Cost) *game.CostGrid {
	var largeur, hauteur int
	for i, pose := range lieu.Placements {
		largeur = max(largeur, pose.U+pieces[i].Size[0])
		hauteur = max(hauteur, pose.V+pieces[i].Size[1])
	}

	grille := game.NewCostGrid(largeur, hauteur)
	for i, pose := range lieu.Placements {
		for v, ligne := range pieces[i].Rows {
			for u, jeton := range ligne {
				forme := jeu.Palette[string(jeton)]
				cout, connu := couts[forme]
				if !connu {
					cout = game.Blocked
				}
				grille.Set(pose.U+u, pose.V+v, cout)
			}
		}
	}
	return grille
}
