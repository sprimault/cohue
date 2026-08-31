// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package manifest

import "errors"

// ErrUnsupportedFormat signale une version de format que ce binaire ne lit pas.
//
// Elle vit ici, et non chez chacun de ses lecteurs, parce que c'est le seul
// défaut de chargement qu'un joueur puisse rencontrer sans avoir rien fait de
// mal : il ouvre un fichier plus ancien ou plus récent que son binaire. Sa
// réponse utile n'est donc pas « corrige ton fichier » mais « celui-ci vient
// d'une autre version du jeu », et cette distinction-là ne se fait que sur une
// sentinelle. Deux sentinelles la rendraient impossible à faire proprement, ce
// qui vaut de la déménager avant que quiconque en ait l'usage.
//
// Elle voyage seule, sans fonction qui vérifierait un numéro : ses lecteurs
// traitent ce défaut de deux façons volontairement différentes, immédiate au
// décodage et accumulée à la validation, et un contrôle commun effacerait cette
// distinction sans que personne s'en aperçoive.
var ErrUnsupportedFormat = errors.New("version de format non prise en charge")
