// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas des armes : le manifeste livré monté sans rien injecter, les valeurs
// de tir qui viennent du tireur, l'absence d'arme de base, et la cadence sous le
// pas de simulation.

package game

import (
	"errors"
	"testing"
	"testing/fstest"

	"github.com/sprimault/cohue"
	"github.com/sprimault/cohue/internal/manifest"
)

// manifesteArmes est le manifeste livré, tenu à la main.
const manifesteArmes = "assets/armes/manifeste.json"

// TestManifesteLivreDonneLArmeDeBase charge le manifeste publié sans rien
// injecter.
//
// C'est le seul manifeste de `assets/` qu'aucun générateur ne produit, donc le
// seul qu'aucun contrôle Python ne relit : ce test est tout ce qui garde
// l'accord entre le fichier et le code qui le lit.
func TestManifesteLivreDonneLArmeDeBase(t *testing.T) {
	armes, err := LoadWeapons(cohue.Assets, manifesteArmes)
	if err != nil {
		t.Fatalf("chargement des armes : %v", err)
	}

	base := armes.Base
	if base.Key == "" {
		t.Fatal("aucune arme ne porte le rôle de base")
	}
	// 400 ms à 60 ticks par seconde. La valeur est écrite en clair : un test qui
	// refait la conversion du code passe même quand les deux sont faux.
	if base.Cooldown != 24 {
		t.Errorf("cadence : %d ticks, attendu 24", base.Cooldown)
	}
	if base.Range != FromInt(6) {
		t.Errorf("portée : %d, attendu %d", base.Range, FromInt(6))
	}
	// 12 tuiles par seconde, converties au pas de simulation comme toutes les
	// vitesses du jeu — c'est ce qui empêche trois fichiers d'avoir deux
	// conventions.
	if base.ProjectileSpeed != 13107 {
		t.Errorf("vitesse de projectile : %d, attendu 13107", base.ProjectileSpeed)
	}
	// L'unité de résistance est définie par cette arme à son premier niveau.
	if base.Hits != 1 {
		t.Errorf("dégâts : %d touche(s), attendu 1", base.Hits)
	}
}

// TestLesValeursDeTirViennentDuTireur est le test qui ferme la divergence.
//
// La portée de la Buse vivait dans deux fichiers avec deux valeurs, six sur son
// profil et sept sur son projectile, sans que rien ne le signale. Le projectile
// ne porte plus que son apparence ; ce test vérifie que le profil porte bien les
// trois valeurs de son tir, et le contrôle de `ressources.py` refuse qu'elles
// reviennent sur l'objet.
func TestLesValeursDeTirViennentDuTireur(t *testing.T) {
	profils, err := LoadProfiles(cohue.Assets, manifestePersonnages)
	if err != nil {
		t.Fatalf("chargement des profils : %v", err)
	}

	buse := profil(t, profils, "cracheur")
	if buse.Range != FromInt(6) {
		t.Errorf("portée de la Buse : %d, attendu %d", buse.Range, FromInt(6))
	}
	if buse.ShotDamage == 0 {
		t.Error("la Buse ne porte pas les dégâts de son tir")
	}
	if buse.ShotSpeed == 0 {
		t.Error("la Buse ne porte pas la vitesse de ses projectiles")
	}
}

// TestArmeSansRoleDeBase vérifie qu'un manifeste sans socle est refusé.
//
// Sans arme de base, le joueur n'a rien qui tire : le jeu se lancerait et le tir
// automatique ne partirait jamais, ce qui ne ressemble pas à un fichier invalide
// mais à un défaut du moteur.
func TestArmeSansRoleDeBase(t *testing.T) {
	fsys := fstest.MapFS{"a.json": &fstest.MapFile{Data: []byte(`{
		"version_format": 1,
		"armes": {
			"lourde": {"nom": "Fusil", "role": "lourde", "cadence_ms": 400,
			           "portee_tuiles": 6, "degats_touches": 3, "projectiles": 1,
			           "vitesse_projectile_tuiles_s": 12.0}
		}
	}`)}}

	_, err := LoadWeapons(fsys, "a.json")
	var invalide *manifest.Invalid
	if !errors.As(err, &invalide) {
		t.Fatalf("manifeste sans arme de base accepté : %v", err)
	}
}

// TestChampsDArmeManquantsListesEnUneFois vérifie que l'auteur reçoit la liste
// et non le premier manquement.
func TestChampsDArmeManquantsListesEnUneFois(t *testing.T) {
	fsys := fstest.MapFS{"a.json": &fstest.MapFile{Data: []byte(`{
		"version_format": 1,
		"armes": {"reglementaire": {"nom": "Réglementaire", "role": "base"}}
	}`)}}

	_, err := LoadWeapons(fsys, "a.json")
	var invalide *manifest.Invalid
	if !errors.As(err, &invalide) {
		t.Fatalf("arme sans aucune valeur acceptée : %v", err)
	}
	if len(invalide.Missing) != 5 {
		t.Errorf("%d manquement(s), attendu 5 :\n  %v", len(invalide.Missing), invalide.Missing)
	}
}

// TestCadenceSousLePasDeSimulation vérifie que la conversion commune refuse une
// durée qu'un tick ne peut pas porter.
//
// Une arme à cinq millisecondes ne tirerait pas deux cents fois par seconde,
// elle tirerait une fois par tick — le fichier mentirait sans que rien ne le
// dise.
func TestCadenceSousLePasDeSimulation(t *testing.T) {
	fsys := fstest.MapFS{"a.json": &fstest.MapFile{Data: []byte(`{
		"version_format": 1,
		"armes": {
			"reglementaire": {"nom": "Réglementaire", "role": "base", "cadence_ms": 5,
			                  "portee_tuiles": 6, "degats_touches": 1, "projectiles": 1,
			                  "vitesse_projectile_tuiles_s": 12.0}
		}
	}`)}}

	_, err := LoadWeapons(fsys, "a.json")
	if err == nil {
		t.Fatal("cadence de 5 ms acceptée")
	}
}

// TestFormatDArmesNonPrisEnCharge vérifie la sentinelle partagée.
func TestFormatDArmesNonPrisEnCharge(t *testing.T) {
	fsys := fstest.MapFS{"a.json": &fstest.MapFile{
		Data: []byte(`{"version_format": 99, "armes": {}}`)}}

	if _, err := LoadWeapons(fsys, "a.json"); !errors.Is(err, manifest.ErrUnsupportedFormat) {
		t.Errorf("format 99 accepté, ou refusé pour une autre raison : %v", err)
	}
}
