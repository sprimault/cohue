// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas de la campagne : celle qui est livrée, le dossier renommé sans son
// identifiant, et l'absence de lieu de départ.

package level

import (
	"errors"
	"testing"
	"testing/fstest"

	"github.com/sprimault/cohue"
	"github.com/sprimault/cohue/internal/manifest"
)

// TestCampagneLivreeDonneSonLieuDeDepart charge celle du binaire sans rien
// injecter.
//
// Aucun générateur ne l'écrit et aucun contrôle Python ne la relit : ce test est
// tout ce qui garde l'accord entre le dossier livré et le code qui l'ouvre.
func TestCampagneLivreeDonneSonLieuDeDepart(t *testing.T) {
	campagne, err := LoadCampaign(cohue.Assets, cohue.StartingCampaign)
	if err != nil {
		t.Fatalf("campagne livrée : %v", err)
	}

	if campagne.ID != "demonstration" {
		t.Errorf("identifiant : %q", campagne.ID)
	}
	if attendu := cohue.StartingCampaign + "/place"; campagne.StartPath(cohue.StartingCampaign) != attendu {
		t.Errorf("lieu de départ : %q, attendu %q",
			campagne.StartPath(cohue.StartingCampaign), attendu)
	}
}

// TestCampagneDupliqueeSansSonIdentifiant garde le piège du copier-coller.
//
// Quelqu'un duplique une campagne pour en faire une variante, renomme le dossier
// et oublie l'identifiant : sans ce refus, la copie se charge en se croyant
// l'original, et le catalogue affiche deux fois le même nom.
func TestCampagneDupliqueeSansSonIdentifiant(t *testing.T) {
	fsys := fstest.MapFS{"copie/campagne.json": &fstest.MapFile{Data: []byte(`{
		"version_format": 1, "identifiant": "originale", "lieu_depart": "place"
	}`)}}

	_, err := LoadCampaign(fsys, "copie")
	var invalide *manifest.Invalid
	if !errors.As(err, &invalide) {
		t.Fatalf("campagne dont l'identifiant dément son dossier acceptée : %v", err)
	}
}

// TestCampagneSansLieuDeDepart refuse celle par où l'on ne peut pas commencer.
//
// C'est la seule chose que le descripteur porte aujourd'hui : sans elle, il ne
// resterait qu'un dossier dont personne ne sait quelle salle ouvre la run.
func TestCampagneSansLieuDeDepart(t *testing.T) {
	fsys := fstest.MapFS{"vide/campagne.json": &fstest.MapFile{Data: []byte(`{
		"version_format": 1, "identifiant": "vide"
	}`)}}

	_, err := LoadCampaign(fsys, "vide")
	var invalide *manifest.Invalid
	if !errors.As(err, &invalide) {
		t.Fatalf("campagne sans lieu de départ acceptée : %v", err)
	}
}

// TestFormatDeCampagneNonPrisEnCharge vérifie la sentinelle partagée.
func TestFormatDeCampagneNonPrisEnCharge(t *testing.T) {
	fsys := fstest.MapFS{"x/campagne.json": &fstest.MapFile{Data: []byte(`{
		"version_format": 99, "identifiant": "x", "lieu_depart": "place"
	}`)}}

	if _, err := LoadCampaign(fsys, "x"); !errors.Is(err, manifest.ErrUnsupportedFormat) {
		t.Errorf("format 99 accepté, ou refusé pour une autre raison : %v", err)
	}
}
