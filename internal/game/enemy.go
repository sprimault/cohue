// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package game

// Enemy est une créature de la horde, telle qu'elle vit dans son bassin.
//
// Une struct nue, sans méthode ni pointeur : elle est copiée d'une place à
// l'autre à chaque suppression par échange, et tout ce qu'elle porte doit
// supporter ce déplacement. Ce qui la désigne d'une image à l'autre est un
// Handle, jamais un `*Enemy`.
type Enemy struct {
	// Profile est l'index de son profil dans `Profiles.Enemies`, et jamais une
	// copie de ses valeurs — même « pour éviter une indirection ». C'est ce qui
	// rend une modification de la table effective sans recharger le monde, et ce
	// qui empêche deux Badauds d'avoir des vitesses différentes.
	Profile int
	// X et Y sont sa position dans le monde, en tuiles.
	X, Y Fixed
}
