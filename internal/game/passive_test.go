// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas des passifs : la table livrée, le champ qu'un axe doit porter et celui
// qu'il doit refuser, et la cadence qu'un axe entier ne doit pas épuiser.

package game

import (
	"errors"
	"testing"
	"testing/fstest"

	"github.com/sprimault/cohue"
	"github.com/sprimault/cohue/internal/manifest"
)

// tableDEssai monte un manifeste d'armes complet autour de la section donnée.
//
// L'arme est celle du fichier livré, valeurs comprises : les contrôles croisés
// portent sur elle, et une arme forgée à côté ferait qu'un axe accepté ici
// serait refusé en jeu.
func tableDEssai(t *testing.T, passifs string) (*Weapons, error) {
	t.Helper()
	fsys := fstest.MapFS{"a.json": &fstest.MapFile{Data: []byte(`{
		"version_format": 1,
		"armes": {
			"reglementaire": {"nom": "Réglementaire", "role": "base", "cadence_ms": 400,
			                  "portee_tuiles": 6, "degats_touches": 1, "projectiles": 1,
			                  "vitesse_projectile_tuiles_s": 12.0}
		},
		` + passifs + `
	}`)}}
	return LoadWeapons(fsys, "a.json")
}

// TestManifesteLivreDonneLesPassifs charge la table publiée sans rien injecter.
//
// Comme pour les armes, aucun générateur ne la produit et aucun contrôle Python
// n'en relit les valeurs : ce test est tout ce qui garde l'accord entre le
// fichier et le code qui le lit.
func TestManifesteLivreDonneLesPassifs(t *testing.T) {
	armes, err := LoadWeapons(cohue.Assets, manifesteArmes)
	if err != nil {
		t.Fatalf("chargement des armes : %v", err)
	}
	table := armes.Passives

	if len(table.Axes) != 2 {
		t.Fatalf("%d axe(s), attendu 2", len(table.Axes))
	}
	// Triés par clé de manifeste : « cadence » avant « portee ».
	if table.Axes[0].Axis != AxisCadence || table.Axes[1].Axis != AxisRange {
		t.Errorf("axes dans l'ordre %s, %s", table.Axes[0].Axis, table.Axes[1].Axis)
	}

	// 33 ms à 60 ticks par seconde. La valeur est écrite en clair : un test qui
	// refait la conversion du code passe même quand les deux sont faux.
	if table.Axes[0].CooldownStep != 2 {
		t.Errorf("pas de cadence : %d ticks, attendu 2", table.Axes[0].CooldownStep)
	}
	if table.Axes[1].RangeStep != One/2 {
		t.Errorf("pas de portée : %d, attendu %d", table.Axes[1].RangeStep, One/2)
	}
	if table.Relief.Heal != 30 {
		t.Errorf("soin de la soupape : %d, attendu 30", table.Relief.Heal)
	}
}

// TestLAxeEntierNeVidePasLaCadence est le contrôle croisé que deux fichiers
// auraient rendu impossible.
//
// Six paliers de quatre ticks sur une cadence de vingt-quatre la ramènent à
// zéro : l'arme tirerait à chaque image. Le défaut ne se voit ni dans la table,
// où quatre ticks est une valeur ordinaire, ni dans l'arme, où vingt-quatre l'est
// aussi — seule leur rencontre le montre, et elle n'a lieu qu'au chargement.
func TestLAxeEntierNeVidePasLaCadence(t *testing.T) {
	_, err := tableDEssai(t, `"passifs": {
		"axes": {
			"cadence": {"nom": "Cadence", "phrase": "Plus souvent.",
			            "paliers": 6, "pas_ms": 67}
		},
		"soupape": {"nom": "Souffle", "phrase": "Des forces.", "soin": 30}
	}`)

	var invalide *manifest.Invalid
	if !errors.As(err, &invalide) {
		t.Fatalf("axe qui vide la cadence accepté : %v", err)
	}
}

// TestUnPasEtrangerALAxeEstRefuse garde le contrôle dans les deux sens.
//
// Un `pas_tuiles` resté sur la cadence après un copier-coller ne serait jamais
// lu et laisserait croire à un réglage : c'est le champ de trop, aussi grave que
// le champ absent, et le même défaut que la portée oubliée sur un Quidam.
func TestUnPasEtrangerALAxeEstRefuse(t *testing.T) {
	_, err := tableDEssai(t, `"passifs": {
		"axes": {
			"cadence": {"nom": "Cadence", "phrase": "Plus souvent.",
			            "paliers": 6, "pas_ms": 33, "pas_tuiles": 0.5}
		},
		"soupape": {"nom": "Souffle", "phrase": "Des forces.", "soin": 30}
	}`)

	var invalide *manifest.Invalid
	if !errors.As(err, &invalide) {
		t.Fatalf("pas de portée sur l'axe de cadence accepté : %v", err)
	}
}

// TestUnAxeInconnuEstRefuse vérifie que la liste des axes est close.
//
// Le moteur reconnaît un axe par sa clé, comme un profil reconnaît un
// comportement : une clé qu'il ne connaît pas se chargerait sans effet, et la
// carte s'afficherait sans rien améliorer.
//
// **L'axe inconnu ne porte aucun pas, et c'est ce qui rend le cas
// discriminant.** Avec un `pas_ms`, il tombait sous la règle du champ de trop —
// réservé à la cadence — et le test restait vert quand la liste close
// disparaissait : deux règles déclenchées par un seul fichier ne se distinguent
// pas. Sans pas, seule la clé inconnue peut le refuser.
func TestUnAxeInconnuEstRefuse(t *testing.T) {
	_, err := tableDEssai(t, `"passifs": {
		"axes": {
			"perforant": {"nom": "Perforant", "phrase": "Les tirs traversent.",
			              "paliers": 6}
		},
		"soupape": {"nom": "Souffle", "phrase": "Des forces.", "soin": 30}
	}`)

	var invalide *manifest.Invalid
	if !errors.As(err, &invalide) {
		t.Fatalf("axe « perforant » accepté : %v", err)
	}
}

// TestLaSoupapeManquanteEstRefusee garde ce qui empêche l'écran vide.
//
// Elle est la seule carte répétable, donc la seule qui puisse remplir trois
// places à elle seule. Sans elle, la montée qui suit l'épuisement des axes
// n'aurait rien à proposer — au moment précis où le joueur a le mieux joué.
func TestLaSoupapeManquanteEstRefusee(t *testing.T) {
	_, err := tableDEssai(t, `"passifs": {
		"axes": {
			"cadence": {"nom": "Cadence", "phrase": "Plus souvent.",
			            "paliers": 6, "pas_ms": 33}
		}
	}`)

	var invalide *manifest.Invalid
	if !errors.As(err, &invalide) {
		t.Fatalf("table sans soupape acceptée : %v", err)
	}
}

// TestChampsDePassifsManquantsListesEnUneFois vérifie que l'auteur reçoit la
// liste et non le premier manquement.
func TestChampsDePassifsManquantsListesEnUneFois(t *testing.T) {
	_, err := tableDEssai(t, `"passifs": {
		"axes": {"cadence": {}},
		"soupape": {}
	}`)

	var invalide *manifest.Invalid
	if !errors.As(err, &invalide) {
		t.Fatalf("section vide acceptée : %v", err)
	}
	// Nom, phrase, paliers et pas pour l'axe ; nom, phrase et soin pour la
	// soupape. Une absence compte pour une ligne : les bornes ne se prononcent
	// que sur un champ présent.
	if len(invalide.Missing) != 7 {
		t.Errorf("%d manquement(s), attendu 7 :\n  %v", len(invalide.Missing), invalide.Missing)
	}
}
