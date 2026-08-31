// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas des profils : le manifeste livré monté sans rien injecter, les
// manquements listés en une fois, le zéro écrit qui n'est pas une absence, le
// rôle inconnu qui arrête là, et le comportement que plus aucun profil
// n'exercerait.

package game

import (
	"errors"
	"fmt"
	"slices"
	"testing"
	"testing/fstest"

	"github.com/sprimault/cohue"
	"github.com/sprimault/cohue/internal/manifest"
)

// manifestePersonnages est le manifeste livré, celui que le binaire embarque.
const manifestePersonnages = "assets/personnages/manifeste.json"

// TestManifesteLivreDonneLesProfils monte la table sur le manifeste publié, sans
// rien injecter.
//
// C'est le seul test qui exerce la chaîne entière — `go:embed`, le manifeste que
// `figurines.py` écrit, le décodage, la conversion —, donc le seul qui tombe
// quand le générateur et le chargeur cessent d'être d'accord. Un renommage de
// champ casse ici, pas au lancement du jeu.
func TestManifesteLivreDonneLesProfils(t *testing.T) {
	table, err := LoadProfiles(cohue.Assets, manifestePersonnages)
	if err != nil {
		t.Fatalf("chargement des profils : %v", err)
	}

	// 5 tuiles par seconde sur 60 ticks, en virgule fixe. La valeur est écrite
	// en clair et non recalculée : un test qui refait le calcul du code passe
	// même quand les deux sont faux.
	if table.Player.Speed != 5461 {
		t.Errorf("vitesse du joueur : %d, attendu 5461", table.Player.Speed)
	}
	if table.Player.Health != 100 || table.Player.DamageCap != 20 {
		t.Errorf("joueur : %d PV et %d de plafond, attendu 100 et 20",
			table.Player.Health, table.Player.DamageCap)
	}

	cles := make([]string, len(table.Enemies))
	for i, e := range table.Enemies {
		cles[i] = e.Key
	}
	if !slices.IsSorted(cles) {
		t.Errorf("profils rendus dans l'ordre %v, l'index doit être stable", cles)
	}
	if slices.Contains(cles, "civil") {
		t.Error("le Passant est entré dans les ennemis, il n'a ni dégâts ni points")
	}

	marcheur := profil(t, table, "marcheur")
	// 0,62 fois la vitesse du joueur, arrondi une seule fois.
	if marcheur.Speed != 3386 {
		t.Errorf("vitesse du Badaud : %d, attendu 3386", marcheur.Speed)
	}
	if marcheur.Behaviour != Chase {
		t.Errorf("comportement du Badaud : « %s », attendu « %s »", marcheur.Behaviour, Chase)
	}
	if marcheur.Name != "Badaud" {
		t.Errorf("nom du marcheur : « %s », attendu « Badaud »", marcheur.Name)
	}

	// Les champs qu'un seul comportement porte arrivent bien jusqu'à la table,
	// et restent à zéro ailleurs.
	if r := profil(t, table, "cracheur").Range; r != FromInt(6) {
		t.Errorf("portée de la Buse : %d, attendu %d", r, FromInt(6))
	}
	if d := marcheur.ChargeDamage; d != 0 {
		t.Errorf("le Badaud porte %d de dégâts de charge, il ne charge pas", d)
	}
}

// TestToutComportementEstExerce ferme le seul angle mort que laissent les deux
// tables, celle de `ressources.py` et celle d'ici.
//
// Elles ne peuvent diverger en silence que sur un comportement qu'aucun profil
// n'exerce : partout ailleurs, le manifeste livré satisfait l'une et pas
// l'autre, et l'un des deux contrôles rougit le jour même. Or ce comportement-là
// est justement celui qu'on vient d'ajouter et qu'on croit avoir éprouvé.
func TestToutComportementEstExerce(t *testing.T) {
	brut, err := manifest.Decode[rawManifest](cohue.Assets, manifestePersonnages)
	if err != nil {
		t.Fatalf("lecture du manifeste : %v", err)
	}

	exerces := make(map[Behaviour]int, len(behaviours))
	for _, p := range brut.Profiles {
		exerces[Behaviour(p.Behaviour)]++
	}
	couverts := 0
	for _, b := range behaviours {
		if exerces[b] == 0 {
			t.Errorf("« %s » n'est exercé par aucun profil livré : la table du "+
				"paquet a divergé de celle du générateur, ou le profil manque", b)
			continue
		}
		couverts++
	}
	t.Logf("comportements exercés : %d sur %d", couverts, len(behaviours))
}

// TestManquementsListesEnUneFois éprouve les deux moitiés du contrôle
// conditionnel sur le même fichier : le champ que le comportement exige et qui
// manque, et celui qu'il ne reconnaît pas et qui est là.
//
// Les deux ensemble, parce que la promesse n'est pas seulement de refuser mais
// de tout dire d'un coup : qui met au point un manifeste veut la liste, pas un
// aller-retour par manquement.
func TestManquementsListesEnUneFois(t *testing.T) {
	fsys := fstest.MapFS{"m.json": &fstest.MapFile{Data: []byte(`{
		"version_format": 1,
		"profils": {
			"joueur": {"role": "joueur", "nom": "Survivant", "rayon_tuiles": 0.125,
			           "vitesse_tuiles_s": 5.0, "vie": 100, "plafond_degats_s": 20},
			"cracheur": {"role": "ennemi", "nom": "Buse", "rayon_tuiles": 0.125,
			             "comportement": "tir", "vitesse_relative": 0.55,
			             "touches": 5, "points": 40, "cout_pression": 6,
			             "poids_separation": 1.3, "max_simultane": 0,
			             "degats_contact_s": 4, "degats_tir": 6,
			             "vitesse_projectile_tuiles_s": 7.0},
			"marcheur": {"role": "ennemi", "nom": "Badaud", "rayon_tuiles": 0.125,
			             "comportement": "poursuite", "vitesse_relative": 0.62,
			             "touches": 3, "points": 10, "cout_pression": 3,
			             "poids_separation": 1.0, "max_simultane": 0,
			             "degats_contact_s": 6, "tangentiel": 0.55}
		}
	}`)}}

	_, err := LoadProfiles(fsys, "m.json")
	var invalide *manifest.Invalide
	if !errors.As(err, &invalide) {
		t.Fatalf("manifeste accepté, ou refusé autrement : %v", err)
	}

	attendus := []string{
		"cracheur.portee_tuiles : absent, alors que « tir » l'exige",
		"marcheur.tangentiel : réservé à « flanc »",
	}
	if !slices.Equal(invalide.Manques, attendus) {
		t.Errorf("manquements :\n  %v\nattendu :\n  %v", invalide.Manques, attendus)
	}
}

// TestZeroEcritNestPasUneAbsence éprouve le champ qui a imposé les pointeurs.
//
// `max_simultane` vaut zéro pour « aucun plafond », si bien qu'un entier nu
// rendrait l'oubli du champ indiscernable du réglage le plus courant. Le profil
// se chargerait sans plafond, et le Secouriste apparaîtrait en nombre.
func TestZeroEcritNestPasUneAbsence(t *testing.T) {
	const modele = `{
		"version_format": 1,
		"profils": {
			"joueur": {"role": "joueur", "nom": "Survivant", "rayon_tuiles": 0.125,
			           "vitesse_tuiles_s": 5.0, "vie": 100, "plafond_degats_s": 20},
			"soigneur": {"role": "ennemi", "nom": "Secouriste", "rayon_tuiles": 0.109,
			             "comportement": "soin", "vitesse_relative": 0.7,
			             "touches": 3, "points": 15, "cout_pression": 6,
			             "poids_separation": 1.0, "degats_contact_s": 4%s}
		}
	}`

	ecrit := fstest.MapFS{"m.json": &fstest.MapFile{
		Data: []byte(fmt.Sprintf(modele, `, "max_simultane": 0`))}}
	table, err := LoadProfiles(ecrit, "m.json")
	if err != nil {
		t.Fatalf("plafond écrit à zéro refusé : %v", err)
	}
	if table.Enemies[0].MaxAlive != 0 {
		t.Errorf("plafond : %d, attendu 0", table.Enemies[0].MaxAlive)
	}

	absent := fstest.MapFS{"m.json": &fstest.MapFile{Data: []byte(fmt.Sprintf(modele, ""))}}
	if _, err := LoadProfiles(absent, "m.json"); err == nil {
		t.Error("plafond absent accepté, il ne se distingue plus d'un zéro écrit")
	}
}

// TestRoleInconnuArreteLa vérifie qu'un rôle que le paquet ne connaît pas rend
// un seul manquement.
//
// Sans cet arrêt, aucun prédicat de la table ne reconnaîtrait le profil et l'on
// déverserait seize refus dont pas un ne nommerait la cause.
func TestRoleInconnuArreteLa(t *testing.T) {
	fsys := fstest.MapFS{"m.json": &fstest.MapFile{Data: []byte(`{
		"version_format": 1,
		"profils": {"chose": {"role": "monstre", "nom": "Chose", "rayon_tuiles": 0.1}}
	}`)}}

	_, err := LoadProfiles(fsys, "m.json")
	var invalide *manifest.Invalide
	if !errors.As(err, &invalide) {
		t.Fatalf("rôle inconnu accepté, ou refusé autrement : %v", err)
	}
	// Le manquement du rôle, et celui du joueur qu'aucun profil ne fournit.
	if len(invalide.Manques) != 2 {
		t.Errorf("%d manquements, attendu 2 :\n  %v", len(invalide.Manques), invalide.Manques)
	}
}

// TestFormatNonPrisEnCharge vérifie qu'un manifeste d'une autre version se
// reconnaît à sa sentinelle, et non à son message.
func TestFormatNonPrisEnCharge(t *testing.T) {
	fsys := fstest.MapFS{"m.json": &fstest.MapFile{
		Data: []byte(`{"version_format": 99, "profils": {}}`)}}

	_, err := LoadProfiles(fsys, "m.json")
	if !errors.Is(err, manifest.ErrUnsupportedFormat) {
		t.Errorf("format 99 accepté, ou refusé pour une autre raison : %v", err)
	}
}

// profil rend le profil de clé donnée, ou arrête le test.
func profil(t *testing.T, table *Profiles, cle string) EnemyProfile {
	t.Helper()
	for _, e := range table.Enemies {
		if e.Key == cle {
			return e
		}
	}
	t.Fatalf("« %s » absent de la table", cle)
	return EnemyProfile{}
}
