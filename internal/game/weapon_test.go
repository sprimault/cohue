// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas des armes : le manifeste livré monté sans rien injecter, les valeurs
// de tir qui viennent du tireur, l'absence d'arme de base, et la cadence sous le
// pas de simulation.

package game

import (
	"errors"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/sprimault/cohue"
	"github.com/sprimault/cohue/internal/manifest"
)

// manifesteArmes est le manifeste livré, tenu à la main.
const manifesteArmes = "assets/armes/manifeste.json"

// passifsValides est une section de passifs sans défaut, à coller dans les
// fixtures dont le sujet est l'arme.
//
// **Sans elle, ces tests deviendraient verts pour la mauvaise raison.** Le
// manifeste porte deux sections, et une fixture qui n'en déclare qu'une rend un
// `manifest.Invalid` quoi qu'il arrive : un test qui n'assure que le type de
// l'erreur passerait alors même si le contrôle qu'il garde disparaissait. Le cas
// ne relève pas de la théorie — il est apparu à l'écriture de cette section, et
// la mutation l'a confirmé.
const passifsValides = `"passifs": {
		"axes": {
			"cadence": {"nom": "Cadence", "phrase": "Plus souvent.",
			            "paliers": 6, "pas_ms": 33}
		},
		"soupape": {"nom": "Souffle", "phrase": "Des forces.", "soin": 30}
	}`

// TestManifesteLivreDonneLArmeDeBase charge le manifeste publié sans rien
// injecter.
//
// Aucun générateur ne le produit, donc aucun contrôle Python n'en relit les
// valeurs : ce test est tout ce qui garde l'accord entre le fichier et le code
// qui le lit. Le manifeste de progression est dans le même cas.
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
//
// **Ce que ce cas éprouve est le rôle déclaré par l'arme, pas le compte.** La
// seule arme du fichier porte « lourde », ce que le contrôle par arme refuse
// déjà : la règle du « exactement une base » ne peut pas se distinguer ici, et
// la mutation le confirme. C'est `TestDeuxArmesDeBaseRefusees` qui la garde.
func TestArmeSansRoleDeBase(t *testing.T) {
	fsys := fstest.MapFS{"a.json": &fstest.MapFile{Data: []byte(`{
		"version_format": 1,
		"armes": {
			"lourde": {"nom": "Fusil", "role": "lourde", "cadence_ms": 400,
			           "portee_tuiles": 6, "degats_touches": 3, "projectiles": 1,
			           "vitesse_projectile_tuiles_s": 12.0}
		},
		` + passifsValides + `
	}`)}}

	_, err := LoadWeapons(fsys, "a.json")
	var invalide *manifest.Invalid
	if !errors.As(err, &invalide) {
		t.Fatalf("manifeste sans arme de base accepté : %v", err)
	}
}

// TestUneCadenceNulleRefusee garde ce que la godoc de la conversion promet.
//
// **Le commentaire qui la précède dit que la conversion commune refuse une durée
// sous le pas de simulation.** Elle le fait — quand on l'appelle : la conversion
// n'était tentée que pour une durée strictement positive, si bien qu'un zéro
// écrit dans le fichier passait sans un mot et donnait une arme qui tire à chaque
// image. C'est la même forme que `plancher_ms`, dans l'autre manifeste tenu à la
// main, et le même correctif.
func TestUneCadenceNulleRefusee(t *testing.T) {
	fsys := fstest.MapFS{"a.json": &fstest.MapFile{Data: []byte(`{
		"version_format": 1,
		"armes": {
			"reglementaire": {"nom": "Réglementaire", "role": "base", "cadence_ms": 0,
			                  "portee_tuiles": 6, "degats_touches": 1, "projectiles": 1,
			                  "vitesse_projectile_tuiles_s": 12.0}
		},
		` + passifsValides + `
	}`)}}

	_, err := LoadWeapons(fsys, "a.json")
	var invalide *manifest.Invalid
	if !errors.As(err, &invalide) {
		t.Fatalf("cadence nulle acceptée : %v", err)
	}
	if !strings.Contains(err.Error(), "cadence_ms") {
		t.Errorf("le refus ne nomme pas la clé fautive : %v", err)
	}
}

// TestDeuxArmesDeBaseRefusees garde la règle du compte, seule ici à pouvoir
// parler.
//
// Deux armes irréprochables, toutes deux de rôle « base » : aucun contrôle par
// arme n'a rien à dire, et seule la règle du compte refuse le fichier. C'est le
// cas qui manquait — le socle est celui que le joueur porte, et deux socles ne
// désignent rien.
func TestDeuxArmesDeBaseRefusees(t *testing.T) {
	fsys := fstest.MapFS{"a.json": &fstest.MapFile{Data: []byte(`{
		"version_format": 1,
		"armes": {
			"reglementaire": {"nom": "Réglementaire", "role": "base", "cadence_ms": 400,
			                  "portee_tuiles": 6, "degats_touches": 1, "projectiles": 1,
			                  "vitesse_projectile_tuiles_s": 12.0},
			"seconde": {"nom": "Seconde", "role": "base", "cadence_ms": 400,
			            "portee_tuiles": 6, "degats_touches": 1, "projectiles": 1,
			            "vitesse_projectile_tuiles_s": 12.0}
		},
		` + passifsValides + `
	}`)}}

	_, err := LoadWeapons(fsys, "a.json")
	var invalide *manifest.Invalid
	if !errors.As(err, &invalide) {
		t.Fatalf("deux armes de base acceptées : %v", err)
	}
}

// TestChampsDArmeManquantsListesEnUneFois vérifie que l'auteur reçoit la liste
// et non le premier manquement.
func TestChampsDArmeManquantsListesEnUneFois(t *testing.T) {
	fsys := fstest.MapFS{"a.json": &fstest.MapFile{Data: []byte(`{
		"version_format": 1,
		"armes": {"reglementaire": {"nom": "Réglementaire", "role": "base"}},
		` + passifsValides + `
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
		},
		` + passifsValides + `
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
