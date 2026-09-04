// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le cas que la doctrine annonçait sans pouvoir le tenir : deux runs d'une même
// graine dont le flux cosmétique est consommé différemment rendent la même
// simulation.

package game

import (
	"testing"

	"github.com/sprimault/cohue"
)

// huitFigurants rend un peuplement fixe, sur une ligne dégagée du champ d'essai.
//
// **Les positions sont les mêmes des deux côtés de la comparaison**, et c'est ce
// qui isole la question : ce que le décalage du flux change alors n'est que le
// cap des figurants, jamais l'endroit où ils sont posés.
func huitFigurants() []AmbientPlacement {
	pose := make([]AmbientPlacement, 0, 8)
	for i := range 8 {
		pose = append(pose, AmbientPlacement{
			X: FromInt(4+i) + One/2,
			Y: FromInt(20) + One/2,
		})
	}
	return pose
}

// TestLeCosmetiqueNeDecideDeRien vérifie que ce qui n'a aucun effet n'en a
// aucun.
//
// **Il a attendu son premier consommateur.** `docs/go.md` l'annonce depuis
// l'étape 1 en notant qu'il passerait alors « sans rien séparer » : aucune
// entité ne tirait dans ce flux, si bien que le décaler ne changeait rien parce
// qu'il n'alimentait personne, et non parce que la séparation tenait. Une
// garantie annoncée que rien ne tenait est pire que son absence — c'est la même
// correction que celle du test d'empreinte.
//
// Les figurants d'un lieu en sont ce consommateur : leur position, leur cap et
// leurs changements de cap y puisent tous. Décaler le flux les déplace donc
// visiblement, et c'est précisément ce que la simulation ne doit pas voir.
//
// Le cas compare des empreintes plutôt que des positions choisies : ce qu'il
// garde n'est pas qu'une créature soit au même endroit, mais que **rien** de ce
// que la partie retient n'ait bougé.
func TestLeCosmetiqueNeDecideDeRien(t *testing.T) {
	const decalage = 37

	empreinte := func(tirages int) string {
		profils, err := LoadProfiles(cohue.Assets, manifestePersonnages)
		if err != nil {
			t.Fatalf("profils livrés : %v", err)
		}
		g := NewCostGrid(32, 32)
		w := NewWorld(profils, armesInertes(t), progressionLivree(t),
			vagueUnique(6, indexDuProfil(t, profils, "marcheur")), g,
			graineDeTest, capacitesDeTest)
		w.Place(FromInt(16)+One/2, FromInt(16)+One/2)

		// Le décalage se prend avant le peuplement, de sorte que les figurants
		// ne tombent pas aux mêmes endroits ni ne marchent dans le même sens.
		for range tirages {
			w.hasard.Cosmetic.IntN(1 << 20)
		}
		w.Populate(huitFigurants())

		for range 5 * TPS {
			w.Step(Vec{X: One})
		}
		return w.Fingerprint()
	}

	sans, avec := empreinte(0), empreinte(decalage)
	if sans != avec {
		t.Errorf("l'empreinte change avec le seul flux cosmétique décalé :\n%s\n%s",
			sans, avec)
	}
}

// TestLesFigurantsBougentAvecLeCosmetique est l'autre moitié, et sans elle la
// première ne prouverait rien.
//
// Un flux décalé qui ne déplacerait aucun figurant rendrait évidemment la même
// empreinte — et le cas passerait au vert en n'ayant rien séparé, ce que la
// doctrine reproche précisément à sa version d'avant. Il faut donc montrer que
// le décalage **change** quelque chose, avant de montrer qu'il ne change rien de
// ce qui compte.
func TestLesFigurantsBougentAvecLeCosmetique(t *testing.T) {
	positions := func(tirages int) []Vec {
		profils, err := LoadProfiles(cohue.Assets, manifestePersonnages)
		if err != nil {
			t.Fatalf("profils livrés : %v", err)
		}
		g := NewCostGrid(32, 32)
		w := NewWorld(profils, armesInertes(t), progressionLivree(t), sansVagues(), g,
			graineDeTest, capacitesDeTest)
		w.Place(FromInt(16)+One/2, FromInt(16)+One/2)

		for range tirages {
			w.hasard.Cosmetic.IntN(1 << 20)
		}
		w.Populate(huitFigurants())

		for range TPS {
			w.Step(Vec{})
		}
		var vus []Vec
		for i := range w.Ambients().Active() {
			a := w.Ambients().At(i)
			vus = append(vus, Vec{X: a.X, Y: a.Y})
		}
		return vus
	}

	sans, avec := positions(0), positions(37)
	if len(sans) == 0 {
		t.Fatal("aucun figurant posé : le cas ne peut rien séparer")
	}
	if len(sans) == len(avec) && sans[0] == avec[0] {
		t.Error("le décalage du flux cosmétique ne déplace aucun figurant : " +
			"l'autre cas passerait sans rien garder")
	}
}
