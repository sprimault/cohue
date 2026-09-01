// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// L'unique `go:embed` du dépôt : tout ce que le binaire distribue — manifestes,
// images, sons et lieux livrés — pour que l'exécutable se suffise à lui-même.

// Package cohue n'existe que pour embarquer les ressources dans le binaire.
//
// `go:embed` ne remonte pas au-dessus du répertoire de son paquet, et les
// générateurs écrivent à la racine : le seul paquet qui puisse embarquer
// `assets/` est donc celui de la racine. Poser un fichier Go dans `assets/`
// serait l'autre solution, mais le contrôle des ressources compare ce dossier à
// ce que les générateurs produisent, et il y deviendrait un écart à excepter.
package cohue

import "embed"

// Assets porte tout ce que le binaire distribue : manifestes, images, sons et
// lieux livrés.
//
// Le motif n'exclut rien, et c'est ce qui exige que `assets/` ne contienne que
// du livré. Les planches de relecture vont dans `.tmp/` pour cette raison :
// embarquées, elles pèseraient dans le binaire de qui les a régénérées et pas
// dans celui de l'intégration continue, qui part d'un clone frais.
//
//go:embed assets
var Assets embed.FS

// Ce que le binaire charge au lancement, par leur chemin dans `Assets`.
//
// Ils vivent auprès de l'embed et non dans le programme qui les lit, parce
// qu'ils en décrivent le contenu : un lieu renommé casse ici, où le renommage se
// fait, plutôt que dans un `cmd/` qu'on ne pense pas à rouvrir. Deux programmes
// les lisent, et deux copies d'un chemin ne restent d'accord que par vigilance.
const (
	DecorManifest     = "assets/decors/manifeste.json"
	CharacterManifest = "assets/personnages/manifeste.json"
	WeaponManifest    = "assets/armes/manifeste.json"
	InterfaceManifest = "assets/interface/manifeste.json"
	StartingLevel     = "assets/lieux/place"
)
