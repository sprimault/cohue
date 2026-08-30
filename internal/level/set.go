// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package level

import "github.com/sprimault/cohue/internal/manifest"

// Set est un jeu de pièces : l'atlas, la palette et l'ambiance d'un thème.
type Set struct {
	manifest.Commentable
	// Format est la version du format de jeu de pièces.
	Format int `json:"version_format"`
	// ID nomme le jeu, et c'est lui que les pièces citent.
	ID string `json:"identifiant"`
	// Name est le nom lisible du thème.
	Name string `json:"nom,omitempty"`
	// Palette associe un caractère de grille à une forme du décor. Elle vit ici
	// et non dans chaque pièce : la dupliquer ferait qu'un même caractère
	// désignerait deux choses selon le fichier, et c'est ce qui donne son
	// premier usage concret à l'empreinte du jeu de pièces — un caractère
	// réattribué change le sens de toutes les pièces d'un thème, en silence.
	Palette map[string]string `json:"palette"`
}
