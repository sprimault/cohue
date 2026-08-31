// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La lecture d'un fichier JSON partagé, toute clé inconnue refusée. Le message
// du décodeur ressort tel quel jusqu'à l'auteur du fichier, avec le chemin de la
// clé fautive : le remplacer par « fichier invalide » détruirait la seule
// information utile de la ligne.

// Package manifest lit les fichiers JSON que le jeu partage, et rien de plus.
//
// Y entre ce qu'au moins deux paquets lisent, et rien d'autre. Le nom
// invite à y ranger « ce qui a rapport aux manifestes » ; c'est le critère qui
// décide, pas le nom. Un type de manifeste qui n'a qu'un lecteur reste chez lui
// et ne déménage que le jour où un second apparaît — la symétrie n'est pas un
// motif, le cycle d'import en est un.
//
// Ce qui l'a fait naître : `internal/level` importe `internal/game`, donc `game`
// ne peut pas emprunter le décodeur du chargeur de lieux. Et dupliquer
// `Commentable` aurait retiré au compilateur la garantie qui fait tout son
// intérêt, une structure pouvant alors l'oublier d'un côté sans que l'autre s'en
// aperçoive.
//
// Il porte la lecture, jamais l'interprétation : aucune conversion en virgule
// fixe, aucune règle de jeu, aucun catalogue dérivé. Ce qu'un manifeste veut
// dire appartient à qui le consomme, et une décision de jeu qui se glisserait
// ici aurait trouvé un second domicile sans que personne l'ait voulu.
package manifest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
)

// Decode lit un fichier JSON en refusant toute clé inconnue.
//
// Le refus attrape la faute de frappe qui *ajoute* une clé — `rotaton` au lieu
// de `rotation` se chargerait sinon en silence sur la valeur par défaut, et
// l'auteur ne comprendrait pas pourquoi sa pièce est de travers. Il n'attrape
// jamais celle qui en supprime une : c'est la validation, derrière, qui exige
// ce qui doit être là.
//
// Le message du décodeur ressort tel quel, avec le chemin de la clé fautive :
// le remplacer par « fichier invalide » détruirait la seule information utile.
func Decode[T any](fsys fs.FS, chemin string) (*T, error) {
	brut, err := fs.ReadFile(fsys, chemin)
	if err != nil {
		return nil, fmt.Errorf("lecture de %s: %w", chemin, err)
	}
	d := json.NewDecoder(bytes.NewReader(brut))
	d.DisallowUnknownFields()

	var valeur T
	if err := d.Decode(&valeur); err != nil {
		return nil, fmt.Errorf("%s: %w", chemin, err)
	}
	return &valeur, nil
}
