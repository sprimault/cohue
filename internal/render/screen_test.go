// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce que le rendu garde, maintenant qu'il peut être exercé : deux montages
// identiques peignent les mêmes octets, et la mise en page de l'interface se
// dérive de la surface au lieu de s'écrire.

package render

import (
	"testing"

	"github.com/sprimault/lozengine"

	"github.com/sprimault/cohue"
	"github.com/sprimault/cohue/internal/session"
	"github.com/sprimault/cohue/internal/sprite"
)

// graineDEssai est celle du lieu livré, la même que la première run du jeu.
//
// Une graine explicite et non celle du binaire : ce que ce fichier éprouve est
// le rendu, et il ne doit pas changer de sujet le jour où la run de départ
// change de graine.
const graineDEssai uint64 = 1

// pasDEssai est le nombre de pas joués avant de comparer les images.
//
// **Assez pour que le monde ne soit plus celui du montage, et pas plus.** La
// horde est apparue, les ambiants marchent, et le champ de flux a tourné —
// c'est là que deux exécutions divergeraient si quelque chose lisait une carte
// sans la trier. Au-delà, on paie des secondes sans rien éprouver de plus.
const pasDEssai = 120

// monter rend un écran et le moteur qui le peint, sur le lieu livré.
//
// **Le même montage que `cmd/cohue`, dans le même ordre**, et c'est ce qui rend
// le test utile : une divergence entre les deux se verrait au premier appel qui
// n'existe plus. L'atlas se construit avant les catalogues, qui y rangent leurs
// planches, et il se remet au moteur après eux.
func monter(t *testing.T) (*Screen, *lozengine.Engine) {
	t.Helper()

	partie, err := session.Open(cohue.Assets, cohue.StartingCampaign, graineDEssai)
	if err != nil {
		t.Fatalf("montage de la partie : %v", err)
	}

	atlas := &lozengine.Atlas{}
	hud, err := LoadHUD(cohue.Assets, cohue.InterfaceManifest, atlas)
	if err != nil {
		t.Fatalf("interface : %v", err)
	}
	tuiles, err := sprite.LoadTiles(cohue.Assets, cohue.DecorDir, partie.Decor, atlas)
	if err != nil {
		t.Fatalf("tuiles : %v", err)
	}
	sol, err := NewTerrain(partie.Tiles, tuiles)
	if err != nil {
		t.Fatalf("terrain : %v", err)
	}
	troupe, err := NewCast(cohue.Assets, cohue.CharacterDir, partie.Profiles, atlas)
	if err != nil {
		t.Fatalf("troupe : %v", err)
	}
	objets, err := NewStage(cohue.Assets, cohue.ObjectDir, cohue.ObjectManifest,
		partie.Objects, atlas)
	if err != nil {
		t.Fatalf("objets : %v", err)
	}

	retrait := hud.Inset()
	ecran, err := NewScreen(partie.World, partie.Grid, sol, troupe, objets, atlas, retrait)
	if err != nil {
		t.Fatalf("ecran : %v", err)
	}

	moteur, err := lozengine.NewEngine(ecran.Scene(), lozengine.Options{
		Width:      Width,
		Height:     Height,
		WorldInset: retrait,
		Background: Background,
		Atlas:      atlas,
	})
	if err != nil {
		t.Fatalf("moteur : %v", err)
	}
	return ecran.WithHUD(hud), moteur
}

// jouer avance l'écran et le moteur du nombre de pas donné, et rend l'image.
//
// **Aucune touche n'est tenue**, la valeur zéro d'`Input` étant ce qui le dit :
// une entrée simulée ferait dépendre le résultat de ce que le test presse, et
// c'est le rendu qu'on éprouve, pas la commande.
func jouer(ecran *Screen, moteur *lozengine.Engine, pas int) []lozengine.Color {
	for range pas {
		ecran.Overlaps(moteur.Overlaps())
		ecran.Step(Input{})
		ecran.Describe()
		moteur.Step(nil, lozengine.Input{})
	}
	return moteur.Frame().Pixels
}

// TestDeuxMontagesPeignentLesMemesOctets est ce qui remplace la doctrine qui
// interdisait de tester ce paquet.
//
// **La planche de relecture ne vaut que si elle est reproductible** : comparer
// une image d'avant et d'après un changement suppose que l'image d'avant ne
// dépendait que du code. C'est la seule chose qu'une machine sache dire du
// rendu — le reste se juge à l'œil, et c'est à cela que `make apercus` sert.
//
// **Deux montages complets et non deux passes sur le même**, parce que ce qu'on
// cherche est ce qui se glisse entre deux exécutions : le parcours d'une `map`,
// une adresse qui fuit dans un ordre, un `math/rand` global. Un second rendu de
// la même scène ne les verrait pas.
//
// **Même machine, et jamais une image de référence versionnée.** Ce paquet
// calcule en flottants, ce que sa godoc autorise ; l'identité de deux images
// prises sur deux systèmes n'est donc pas promise, et la figer en fichier de
// référence rendrait rouge un correctif juste. L'invariant est la reproductibilité
// d'une exécution à l'autre, pas d'une cible à l'autre.
func TestDeuxMontagesPeignentLesMemesOctets(t *testing.T) {
	premierEcran, premierMoteur := monter(t)
	premier := jouer(premierEcran, premierMoteur, pasDEssai)

	secondEcran, secondMoteur := monter(t)
	second := jouer(secondEcran, secondMoteur, pasDEssai)

	if len(premier) != len(second) {
		t.Fatalf("deux images de tailles différentes : %d et %d octets",
			len(premier), len(second))
	}
	for i := range premier {
		if premier[i] != second[i] {
			t.Fatalf("les deux images divergent au pixel %d sur %d : %v puis %v",
				i, len(premier), premier[i], second[i])
		}
	}
}
