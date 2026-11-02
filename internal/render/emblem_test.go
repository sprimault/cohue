// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce que le bandeau et l'annonce d'une caisse ont le droit de montrer : tout
// objet qu'ils nomment doit avoir un dessin, propre ou emprunté à la scène.

package render

import "testing"

// TestToutObjetMontrableAUnDessin exige qu'`Icon` rende quelque chose pour
// chaque objet que le rendu pose hors de la scène.
//
// **Le défaut qu'il garde n'avait rien de théorique.** La fiole portait une
// icône dessinée à part ; le jour où sa case du bandeau est passée au dessin de
// la scène, ce fichier a été retiré faute de lecteur — et l'annonce flottée
// au-dessus d'une caisse, qui lisait la même icône, a cessé de montrer quoi que
// ce soit. Rien ne tombait : l'annonce se tait proprement quand l'aspect est
// nul, ce qui est juste pour un objet sans dessin et masque l'erreur pour les
// autres.
//
// **Il porte sur le catalogue livré**, parce que ce qui est éprouvé est la
// chaîne entière — ce que le générateur écrit, et ce que le chargeur en fait.
func TestToutObjetMontrableAUnDessin(t *testing.T) {
	ecran, _ := monter(t)

	// Les deux objets qui ont une case au bandeau, et les quatre armes lourdes
	// dont une caisse annonce la trouvaille.
	for _, nom := range []string{objetAimant, objetFiole,
		"fusil", "lance_flammes", "grenade", "tourelle"} {
		if ecran.objets.Icon(nom).Width == 0 {
			t.Errorf("%s : aucun dessin à montrer, ni icône ni sprite de scène", nom)
		}
	}
}
