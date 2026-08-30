// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

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
		"bloquant": ` + bloquant + `, "transparence_si_derriere": false` + cout + `}`
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
	invalide, ok := err.(*manifest.Invalide)
	if !ok {
		t.Fatalf("erreur %T, attendu *manifest.Invalide : %v", err, err)
	}
	if len(invalide.Manques) != 2 {
		t.Fatalf("%d manquement(s), attendu 2 :\n%v", len(invalide.Manques), invalide.Manques)
	}
	dit := strings.Join(invalide.Manques, "\n")
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

// TestDecorRefuseUneAutreVersion vérifie qu'un format inconnu ne se lit pas à moitié.
func TestDecorRefuseUneAutreVersion(t *testing.T) {
	fsys := fstest.MapFS{"decor.json": &fstest.MapFile{
		Data: []byte(`{"version_format": 99, "tuile": [64, 32], "formes": {}}`)}}
	if _, _, err := LoadDecor(fsys, "decor.json"); err == nil {
		t.Fatal("version 99 acceptée")
	}
}
