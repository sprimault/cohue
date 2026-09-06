// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Set est un jeu de pièces : la palette qui donne un sens aux caractères d'une
// grille de pièce.

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
	// Ground est la forme peinte sous une case dont la forme ne remplit pas son
	// losange : un pilier occupe un quart de la sienne, une cloison un cinquième.
	//
	// **Le sol appartient au lieu et non à la forme.** Le même banc se pose sur
	// du carrelage dans un supermarché et sur du bitume dans un parking : le
	// dessiner sur son sol demanderait deux bancs, et le thème cesserait de
	// décider de son propre décor. Il tient donc ici, à côté de la palette, qui
	// est déjà le vocabulaire partagé du thème.
	//
	// **Un sol par thème et non par pièce**, parce que le thème est l'unité qui
	// porte la cohérence visuelle et qu'aucune pièce livrée ne mêle deux
	// revêtements sous une forme non couvrante. Ce qui le rendrait insuffisant
	// est nommé plutôt que prévu : une pièce dont un pilier serait sur moquette
	// quand ses voisins sont sur carrelage. Ce jour-là, c'est une surcharge par
	// pièce qu'il faudra, pas une refonte.
	//
	// Facultatif, donc `version_format` ne bouge pas : un thème dont toutes les
	// formes couvrent leur case n'en a pas l'usage, et le chargement l'exige
	// exactement quand une palette en nomme une qui ne couvre pas.
	//
	// **Il nomme une forme et non un caractère de palette.** Un caractère
	// désigne ce que les pièces écrivent ; le sol, lui, n'est écrit nulle part,
	// et le faire passer par la palette obligerait un thème à réserver un
	// caractère pour une tuile que personne ne pose.
	Ground string `json:"sol,omitempty"`
}
