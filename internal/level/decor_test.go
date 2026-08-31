// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas du décor : le coût contradictoire d'une forme bloquante, le coût hors
// bornes, la taille de tuile hors du deux pour un, et la version de format que
// ce binaire ne lit pas.

package level

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/sprimault/cohue/internal/manifest"
)

// manifesteDecor bâtit un manifeste de décor autour de formes données.
//
// Les formes arrivent en JSON brut et non en `Shape` : c'est le décodage qui est
// éprouvé, et une structure Go ne saurait pas exprimer l'absence de
// `cout_traversee` autrement que par le pointeur qu'on cherche justement à
// vérifier.
func manifesteDecor(formes string) fstest.MapFS {
	return fstest.MapFS{"decor.json": &fstest.MapFile{Data: []byte(`{
		"version_format": 1,
		"tuile": [64, 32],
		"formes": {` + formes + `}
	}`)}}
}

// forme rend le JSON d'une forme, avec ou sans coût de traversée.
func forme(bloquant, cout string) string {
	if cout != "" {
		cout = `, "cout_traversee": ` + cout
	}
	return `{"theme": "commun", "taille": [64, 32], "ancrage": [32, 31],
		"elevation": 0, "categorie": "sol", "emprise": [1.0, 1.0],
		"bloquant": ` + bloquant + `, "masquant": false` + cout + `}`
}

// TestDecorRefuseLesCoutsContradictoires vérifie le contrôle dans les deux sens.
//
// Les deux manquements sont attendus ensemble : un auteur qui corrigerait le
// premier pour découvrir le second au chargement suivant paierait deux
// aller-retours là où le format n'en demande qu'un.
func TestDecorRefuseLesCoutsContradictoires(t *testing.T) {
	fsys := manifesteDecor(`
		"mur": ` + forme("true", "3") + `,
		"sol": ` + forme("false", "") + `,
		"flaque": ` + forme("false", "2"))

	_, _, err := LoadDecor(fsys, "decor.json")
	if err == nil {
		t.Fatal("manifeste contradictoire accepté")
	}
	invalide, ok := err.(*manifest.Invalid)
	if !ok {
		t.Fatalf("erreur %T, attendu *manifest.Invalid : %v", err, err)
	}
	if len(invalide.Missing) != 2 {
		t.Fatalf("%d manquement(s), attendu 2 :\n%v", len(invalide.Missing), invalide.Missing)
	}
	dit := strings.Join(invalide.Missing, "\n")
	for _, attendu := range []string{"mur", "sol"} {
		if !strings.Contains(dit, attendu) {
			t.Errorf("« %s » absent du refus :\n%s", attendu, dit)
		}
	}
	if strings.Contains(dit, "flaque") {
		t.Errorf("la flaque est correcte et pourtant citée :\n%s", dit)
	}
}

// TestDecorRefuseUnCoutHorsBornes vérifie qu'un coût ne peut pas valoir un mur.
//
// Zéro laisserait une case franchie sans rien payer, et le maximum vaudrait
// `Blocked` sans le dire : le lieu serait muré par une valeur que son auteur
// croit être un ralentissement.
func TestDecorRefuseUnCoutHorsBornes(t *testing.T) {
	for _, cout := range []string{"0", "-1", "65535"} {
		_, _, err := LoadDecor(manifesteDecor(`"sol": `+forme("false", cout)), "decor.json")
		if err == nil {
			t.Errorf("cout_traversee de %s accepté", cout)
		}
	}
}

// TestDecorRefuseUneTuileHorsDuDeuxPourUn couvre les quatre façons de perdre la
// projection : la clé absente, la taille nulle, la négative et le carré.
//
// Le premier cas est celui qui motive le contrôle. Refuser les clés inconnues
// attrape la faute de frappe qui *ajoute* une clé, jamais celle qui en supprime
// une : sans lui, un manifeste sans `tuile` se charge sur un couple nul, et
// c'est le rendu qui divise par zéro, deux paquets plus loin.
func TestDecorRefuseUneTuileHorsDuDeuxPourUn(t *testing.T) {
	for _, entete := range []string{
		`"version_format": 1`,
		`"version_format": 1, "tuile": [0, 0]`,
		`"version_format": 1, "tuile": [-64, -32]`,
		`"version_format": 1, "tuile": [64, 64]`,
	} {
		fsys := fstest.MapFS{"decor.json": &fstest.MapFile{
			Data: []byte(`{` + entete + `, "formes": {}}`)}}
		if _, _, err := LoadDecor(fsys, "decor.json"); err == nil {
			t.Errorf("en-tête accepté : {%s}", entete)
		}
	}
}

// TestDecorRefuseUneAutreVersion vérifie qu'un format inconnu ne se lit pas à moitié.
func TestDecorRefuseUneAutreVersion(t *testing.T) {
	fsys := fstest.MapFS{"decor.json": &fstest.MapFile{
		Data: []byte(`{"version_format": 99, "tuile": [64, 32], "formes": {}}`)}}
	if _, _, err := LoadDecor(fsys, "decor.json"); err == nil {
		t.Fatal("version 99 acceptée")
	}
}
