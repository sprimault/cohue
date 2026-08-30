// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package manifest

// Commentable donne à toute structure du format le droit de porter un
// commentaire.
//
// Le champ n'est jamais lu. Il existe pour que le refus des clés inconnues
// accepte `$comment` là où un auteur en met un — et il en mettra là où sa pièce
// pose question, dans un ancrage ou au milieu d'une liste, pas seulement en tête
// de fichier. L'embarquer plutôt que le répéter fait garantir la règle par le
// compilateur : une structure qui l'oublierait ferait échouer le premier fichier
// commenté, pas le centième.
type Commentable struct {
	Comment string `json:"$comment,omitempty"`
}
