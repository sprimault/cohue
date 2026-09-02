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

	// ProgressionManifest porte le rythme des choix : les seuils de niveau et le
	// plancher de temps. Il ne vit pas auprès des armes parce qu'un seuil
	// n'appartient pas à l'arme équipée — il vaut quelle que soit la build, il
	// survivrait à un remplacement complet de la table d'armes, et il commande
	// le rythme des choix plutôt que leur contenu.
	ProgressionManifest = "assets/progression/manifeste.json"

	// StartingCampaign est la campagne sur laquelle le jeu s'ouvre.
	//
	// Une campagne et non un lieu : c'est elle que l'auteur compose, partage et
	// choisira dans le catalogue, et c'est son descripteur qui dit par quelle
	// salle on commence. En désigner une salle ici aurait fait du binaire le
	// seul à savoir laquelle est la première.
	StartingCampaign = "assets/campagnes/demonstration"
)
