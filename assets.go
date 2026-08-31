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
