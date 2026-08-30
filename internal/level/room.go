// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package level

import "github.com/sprimault/cohue/internal/manifest"

// Room est une pièce : sa grille de tuiles, ses côtés et ses ancrages.
type Room struct {
	manifest.Commentable
	// Format est la version du format de pièce.
	Format int `json:"version_format"`
	// ID nomme la pièce dans son jeu.
	ID string `json:"identifiant"`
	// Set est le jeu de pièces dont elle tire sa palette et son atlas.
	Set string `json:"jeu"`
	// Size est la taille en tuiles, `[u, v]`.
	Size [2]int `json:"taille"`
	// OpenArea est la part de cases praticables, indicative pour l'éditeur.
	OpenArea float64 `json:"aire_ouverte,omitempty"`
	// Sides dit ce que chaque côté offre à ses voisines. Les côtés se nomment
	// par la grille et jamais par l'écran : `nord` est `v = 0`.
	Sides map[string]string `json:"cotes,omitempty"`
	// Rows porte la grille, une chaîne par `v` croissant, un caractère par `u`
	// croissant. La première dimension n'est pas une ligne d'écran : `u` descend
	// vers le sud-est et `v` vers le sud-ouest, et la case (0, 0) est le sommet
	// du losange.
	Rows []string `json:"grille"`
	// Anchors sont les emplacements que le lieu déclare — apparition,
	// signalétique, caisse.
	Anchors []Anchor `json:"ancrages,omitempty"`
}

// Anchor est un emplacement déclaré dans une pièce.
type Anchor struct {
	manifest.Commentable
	// Kind dit à quoi sert l'emplacement.
	Kind string `json:"type"`
	// At est la case, en coordonnées de tuile `[u, v]`.
	At [2]int `json:"position"`
}
