// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Level est le descripteur d'un lieu : son identifiant, le jeu de pièces dont il
// tire ses tuiles, et les pièces qu'il pose avec leurs positions.

// Package level lit les lieux et les cuit en une carte que la simulation
// consomme.
//
// Le moteur ne sait plus, après la cuisson, que le lieu était fait de pièces :
// il reçoit une grille de coûts. C'est ce qui permet aux lieux livrés d'emprunter
// exactement le chemin d'un niveau tiers — même code, une seule chose à
// déboguer, et un chargeur exercé à chaque partie plutôt qu'une fois de temps en
// temps.
package level

import (
	"github.com/sprimault/cohue/internal/game"
	"github.com/sprimault/cohue/internal/manifest"
)

// Level est un lieu : une liste de pièces posées, et rien de plus.
//
// Quelques centaines d'octets, parce qu'il ne porte que des identifiants et des
// positions — le destinataire possède déjà les tuiles.
type Level struct {
	manifest.Commentable
	// Format est la version du format de lieu, indépendante de celle d'une
	// pièce. Une pièce reste dans le binaire quand un lieu circule : le jour où
	// l'une gagne un champ, les lieux publiés ne deviennent pas suspects.
	Format int `json:"version_format"`
	// ID nomme le lieu.
	ID string `json:"identifiant"`
	// SetID est le jeu de pièces dont il tire ses pièces.
	SetID string `json:"jeu_pieces"`
	// SetFingerprint est l'empreinte de ce jeu au moment où le lieu a été
	// composé. Sans elle, une palette retouchée changerait le sens de toutes les
	// pièces en silence.
	SetFingerprint string `json:"empreinte_jeu_pieces,omitempty"`
	// Placements sont les pièces posées, avec leur case d'origine.
	Placements []Placement `json:"pieces"`
	// Waves est la courbe de pression du lieu, absente pour une salle sans horde.
	//
	// Le type vient d'`internal/game` : c'est une table de jeu, et son sens
	// appartient au paquet qui l'exécute. La décrire ici en aurait fait une
	// seconde description, à tenir d'accord avec la première.
	Waves game.WaveScenario `json:"vagues,omitempty"`
	// Ambient est le peuplement de figurants, absent le plus souvent.
	//
	// **Il vit à côté des vagues et non dedans**, parce qu'un figurant ne
	// s'achète pas : il n'a pas de coût de pression, donc rien à faire dans une
	// courbe qui dépense un budget. Un lieu dit combien il en veut, une fois pour
	// toutes ; c'est du décor qu'on pose, pas une horde qui arrive.
	//
	// Champ facultatif, donc `version_format` ne bouge pas : un lieu publié
	// avant lui reste lisible tel quel. Le type vient d'`internal/game` pour la
	// raison qui vaut déjà pour les vagues : c'est une table de jeu, et son sens
	// appartient au paquet qui l'exécute.
	Ambient game.AmbientSpec `json:"ambiance,omitempty"`
}

// Placement est une pièce posée dans un lieu.
type Placement struct {
	manifest.Commentable
	// RoomID nomme la pièce.
	RoomID string `json:"id"`
	// U et V sont la case d'origine de la pièce dans le lieu.
	U int `json:"u"`
	V int `json:"v"`
}
