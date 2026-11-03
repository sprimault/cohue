// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce que la résolution d'une forme de décor décide : par quelle voie elle se
// pose, et dans quelle colonne du relevé de recouvrement elle compte.

package render

import (
	"testing"

	"github.com/sprimault/lozengine/scene"
)

// TestToutCeQuiSeDresseEstReleve garde l'unité du critère de pose.
//
// **Deux critères s'étaient séparés, et le défaut restait entier du côté qui
// demeurait en cellule.** La voie de pose suivait l'obstacle bas quand l'étage
// suivait tout ce qui dispute sa profondeur, si bien que les murs gardaient la
// propriété du tri qu'on venait de corriger pour les murets : un obstacle
// fragile encastré dans une percée se peignait par-dessus le mur qui le borde.
//
// **Le compte attendu se refait par l'étage et non par le champ qui décide**,
// faute de quoi le test lirait deux fois la même décision : un relevé revenu à
// l'obstacle bas laisserait `forme.debout` juste et ne rendrait plus que les
// murets.
func TestToutCeQuiSeDresseEstReleve(t *testing.T) {
	ecran, _ := monter(t)
	sol := ecran.sol

	var attendu, masquantes int
	compter := func(f forme) {
		if f.aspect.Floor != etageDebout {
			return
		}
		attendu++
		if !f.bas {
			masquantes++
		}
	}
	for v := range sol.carte.Height() {
		for u := range sol.carte.Width() {
			if f, posee := sol.formeDe(u, v); posee {
				compter(f)
			}
			for couche := range sol.Couches() {
				if g, posee := sol.coucheDe(couche, u, v); posee {
					compter(g)
				}
			}
		}
	}

	if masquantes == 0 {
		t.Fatal("aucune case masquante dans le lieu livré : le cas ne pose pas la question")
	}
	if obtenu := len(ecran.debout); obtenu != attendu {
		t.Errorf("%d cases se dressent dans le lieu, %d relevées en entités", attendu, obtenu)
	}
	t.Logf("le lieu livré dresse %d cases, dont %d masquantes", attendu, masquantes)
}

// TestLeDecorLivreCompteCommeDecor garde ce qu'aucune image ne montre.
//
// **Le relevé de recouvrement classe sur la voie de pose quand l'aspect n'en
// dit rien**, et le rendu décrit en entités tout ce qui se dresse : sans
// `scene.MaskAsDecor`, un joueur qu'un mur cache cesse de vérifier
// `CoveredByDecor == Covered`, donc reçoit l'aplat du rappel au lieu du voile
// qui laisse voir la maçonnerie. Le défaut ne tombe ni à la compilation ni sur
// une vue de planche — il faut être caché pour le constater.
//
// Le comblement du thème en est, parce qu'il passe par la même résolution tout
// en restant posé en cellule : une colonne qui dépendrait de la voie se
// contredirait d'une forme à l'autre.
func TestLeDecorLivreCompteCommeDecor(t *testing.T) {
	ecran, _ := monter(t)

	if len(ecran.sol.formes) == 0 {
		t.Fatal("le lieu livré ne résout aucune forme : le cas ne pose rien")
	}
	for i, f := range ecran.sol.formes {
		if f.aspect.Masks != scene.MaskAsDecor {
			t.Errorf("forme %d : colonne %d du relevé, attendu %d",
				i, f.aspect.Masks, scene.MaskAsDecor)
		}
	}
	if comblement := ecran.sol.sol; comblement != nil &&
		comblement.aspect.Masks != scene.MaskAsDecor {
		t.Errorf("comblement du thème : colonne %d du relevé, attendu %d",
			comblement.aspect.Masks, scene.MaskAsDecor)
	}
}
