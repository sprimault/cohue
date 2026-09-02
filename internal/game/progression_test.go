// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas de la progression : le manifeste livré, le seuil qui monte, les deux
// sources d'une montée de niveau, et ce que le plancher de temps ne remet pas à
// zéro.

package game

import (
	"errors"
	"testing"
	"testing/fstest"

	"github.com/sprimault/cohue"
	"github.com/sprimault/cohue/internal/manifest"
)

// manifesteProgression est le manifeste livré, tenu à la main.
const manifesteProgression = "assets/progression/manifeste.json"

// progressionLivree rend la table des seuils publiée, ou arrête le test.
func progressionLivree(t *testing.T) *Progression {
	t.Helper()
	p, err := LoadProgression(cohue.Assets, manifesteProgression)
	if err != nil {
		t.Fatalf("progression livrée : %v", err)
	}
	return p
}

// champDeProgression monte une salle vide sur les seuils qu'on lui donne.
//
// **Les seuils sont forgés ici et non lus du fichier livré.** Le plancher publié
// vaut quarante-cinq secondes, soit deux mille sept cents ticks : un test qui
// les jouerait mesurerait la patience de la machine plutôt que la règle, et il
// changerait de sens au premier réglage d'équilibrage. Ce que le fichier contient
// a son propre test.
//
// L'arme est inerte : ces cas comptent des gemmes semées à la main, et un tir
// qui tuerait un cobaye en ajouterait sans qu'ils le sachent.
func champDeProgression(t *testing.T, seuils *Progression) (*World, *Profiles) {
	t.Helper()
	profils, err := LoadProfiles(cohue.Assets, manifestePersonnages)
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}
	w := NewWorld(profils, armesInertes(t), seuils, NewCostGrid(32, 32),
		graineDeTest, 16, 64, 32)
	w.Place(FromInt(16)+One/2, FromInt(16)+One/2)
	return w, profils
}

// semer pose des gemmes sous le joueur, par le chemin qui les produit.
//
// Par `lacher` et non par le bassin : c'est la mort d'une créature qui pose une
// gemme, et un test qui remplirait le bassin lui-même resterait vert le jour où
// la volée cesserait d'y écrire. La créature n'entre pas dans le bassin pour
// autant — vivante sous les pieds du joueur, elle le blesserait à chaque tick et
// fausserait des cas qui ne parlent pas de contact.
func semer(t *testing.T, w *World, profils *Profiles, combien int) {
	t.Helper()
	px, py := w.Player()
	e := Enemy{Profile: indexDuProfil(t, profils, "marcheur"), X: px, Y: py}
	for range combien {
		w.lacher(&e)
	}
}

// TestManifesteLivreDonneLesSeuils charge le fichier publié sans rien injecter.
//
// Comme celui des armes, aucun générateur ne le produit et aucun contrôle Python
// ne le relit : ce test est tout ce qui garde l'accord entre le fichier et le
// code qui le lit.
func TestManifesteLivreDonneLesSeuils(t *testing.T) {
	p := progressionLivree(t)

	if p.FirstThreshold != 10 {
		t.Errorf("premier seuil : %d gemme(s), attendu 10", p.FirstThreshold)
	}
	if p.Increment != 2 {
		t.Errorf("incrément : %d, attendu 2", p.Increment)
	}
	// 45 000 ms à 60 ticks par seconde. La valeur est écrite en clair : un test
	// qui refait la conversion du code passe même quand les deux sont faux.
	if p.Floor != 2700 {
		t.Errorf("plancher : %d ticks, attendu 2700", p.Floor)
	}
	if p.GemValue != 1 {
		t.Errorf("valeur d'une gemme : %d, attendu 1", p.GemValue)
	}
}

// TestUneGemmeVautCeQueLeManifesteDit garde la décision contre l'implémentation
// qu'elle remplace.
//
// **Un cas à une gemme valant un ne garderait rien.** Compter les gemmes et
// multiplier par leur valeur y donnent le même résultat, et c'est exactement
// ainsi que la confusion s'était installée : le manifeste d'objets déclarait
// `experience: 1` sur la gemme, personne n'allait la chercher, et aucun test
// n'aurait pu le dire. Une valeur autre que un est ce qui sépare les deux.
func TestUneGemmeVautCeQueLeManifesteDit(t *testing.T) {
	w, profils := champDeProgression(t, &Progression{
		FirstThreshold: 6, GemValue: 3, Floor: 1000,
	})

	semer(t, w, profils, 2)
	w.Step(Vec{})

	if w.Level() != 2 {
		t.Errorf("niveau : %d, attendu 2 : deux gemmes à trois valent le seuil de six",
			w.Level())
	}
}

// TestUneGemmeSansValeurRefusee garde le fichier qui ferait ramasser pour rien.
//
// Un objet à zéro se distingue mal d'un objet oublié, et le joueur qui traverse
// un tas sans que sa jauge bouge n'a aucun moyen de savoir lequel des deux il
// regarde.
func TestUneGemmeSansValeurRefusee(t *testing.T) {
	fsys := fstest.MapFS{"p.json": &fstest.MapFile{Data: []byte(`{
		"version_format": 1,
		"progression": {
			"niveaux": {"seuil_premier": 10, "seuil_increment": 2, "plancher_ms": 45000},
			"gemmes": {"objet": "gemme", "experience": 0}
		}
	}`)}}

	_, err := LoadProgression(fsys, "p.json")
	var invalide *manifest.Invalid
	if !errors.As(err, &invalide) {
		t.Fatalf("gemme sans valeur acceptée : %v", err)
	}
}

// TestLeSeuilMonteAvecLeNiveau garde la forme affine de la courbe.
//
// Le premier niveau coûte le seuil de départ et non un seuil déjà incrémenté :
// c'est le décalage d'un rang qui se trompe le plus facilement, et il ne se
// verrait à l'écran qu'en comparant deux runs.
func TestLeSeuilMonteAvecLeNiveau(t *testing.T) {
	p := &Progression{FirstThreshold: 10, Increment: 2}

	for _, cas := range []struct{ niveau, seuil int }{{1, 10}, {2, 12}, {5, 18}} {
		if got := p.Threshold(cas.niveau); got != cas.seuil {
			t.Errorf("seuil du niveau %d : %d, attendu %d", cas.niveau, got, cas.seuil)
		}
	}
}

// TestLesGemmesRamasseesMontentLeNiveau éprouve la première source de montée.
func TestLesGemmesRamasseesMontentLeNiveau(t *testing.T) {
	w, profils := champDeProgression(t, &Progression{
		FirstThreshold: 3, Increment: 4, GemValue: 1, Floor: 1000,
	})

	semer(t, w, profils, 5)
	w.Step(Vec{})

	if w.Level() != 2 {
		t.Fatalf("niveau : %d, attendu 2 après cinq gemmes pour un seuil de trois", w.Level())
	}
	// Deux gemmes de trop, reportées et non perdues.
	if w.Experience() != 2 {
		t.Errorf("expérience : %d, attendu 2", w.Experience())
	}
	if w.Threshold() != 7 {
		t.Errorf("seuil suivant : %d, attendu 7", w.Threshold())
	}
}

// TestUneRecolteAbondanteDonnePlusieursNiveaux garde la boucle qui distribue.
//
// Sept gemmes pour un seuil constant de trois en valent deux, pas un. L'aimant
// rendra le cas ordinaire — une récolte entière ramassée d'un coup —, et une
// montée par tick étalerait sur sept secondes ce que le joueur vient de gagner.
func TestUneRecolteAbondanteDonnePlusieursNiveaux(t *testing.T) {
	w, profils := champDeProgression(t, &Progression{
		FirstThreshold: 3, GemValue: 1, Floor: 1000,
	})

	semer(t, w, profils, 7)
	w.Step(Vec{})

	if w.Level() != 3 {
		t.Errorf("niveau : %d, attendu 3 après sept gemmes pour un seuil de trois", w.Level())
	}
	if w.Experience() != 1 {
		t.Errorf("expérience : %d, attendu 1", w.Experience())
	}
}

// TestLePlancherDonneUnNiveauSansRienRamasser éprouve la seconde source.
//
// La conception en fait une règle dure : jamais plus de quarante-cinq secondes
// sans un choix à faire. Sans elle, le tempo dépendrait du taux de ramassage
// alors que les deux autres décisions du chapitre tirent en sens contraire — à
// valeur de gemme fixe chaque niveau en demande davantage, et les gemmes
// s'effacent.
//
// Sept ticks, qui ne sont un multiple ni de la période du champ de flux ni de la
// cadence de l'arme : un plancher en phase avec un autre mécanisme se
// déclencherait au même tick que lui, et le test ne dirait plus lequel des deux
// a donné le niveau.
func TestLePlancherDonneUnNiveauSansRienRamasser(t *testing.T) {
	w, _ := champDeProgression(t, &Progression{
		FirstThreshold: 1000, GemValue: 1, Floor: 7,
	})

	for range 6 {
		w.Step(Vec{})
	}
	if w.Level() != 1 {
		t.Fatalf("niveau : %d au sixième tick, attendu 1 pour un plancher de sept", w.Level())
	}

	w.Step(Vec{})
	if w.Level() != 2 {
		t.Errorf("niveau : %d au septième tick, attendu 2", w.Level())
	}
}

// TestLePlancherNeRemetPasLExperienceAZero garde la clause additive.
//
// Les gemmes déjà ramassées comptent pour le niveau suivant. Sans cela, un
// joueur puni par sa lenteur perdrait ce qu'il avait récolté au moment précis où
// le jeu prétend le récompenser, et le plancher cesserait d'être une seconde
// source de progression pour devenir une remise à zéro.
func TestLePlancherNeRemetPasLExperienceAZero(t *testing.T) {
	w, profils := champDeProgression(t, &Progression{
		FirstThreshold: 5, GemValue: 1, Floor: 7,
	})

	semer(t, w, profils, 2)
	for range 7 {
		w.Step(Vec{})
	}

	if w.Level() != 2 {
		t.Fatalf("niveau : %d, attendu 2 : le plancher devait donner la montée", w.Level())
	}
	if w.Experience() != 2 {
		t.Errorf("expérience : %d, attendu 2 gemmes conservées", w.Experience())
	}
}

// TestUneMonteeParGemmesRepousseLePlancher sépare l'intervalle du calendrier.
//
// « Jamais plus de quarante-cinq secondes sans un choix » est une garantie sur
// l'écart entre deux choix, donc un compteur qui repart à chaque montée, quelle
// qu'en soit la source. Sur un calendrier absolu, le plancher donnerait un
// niveau forcé quelques secondes après un niveau gagné — deux choix coup sur
// coup, puis un long silence.
func TestUneMonteeParGemmesRepousseLePlancher(t *testing.T) {
	w, profils := champDeProgression(t, &Progression{
		FirstThreshold: 3, GemValue: 1, Floor: 7,
	})

	w.Step(Vec{})
	w.Step(Vec{})
	semer(t, w, profils, 3)
	w.Step(Vec{})
	if w.Level() != 2 {
		t.Fatalf("niveau : %d au troisième tick, attendu 2 par les gemmes", w.Level())
	}

	// Le septième tick de la partie est passé ; celui du compteur ne l'est pas.
	for range 6 {
		w.Step(Vec{})
	}
	if w.Level() != 2 {
		t.Fatalf("niveau : %d au neuvième tick, attendu 2 : le plancher compte "+
			"depuis la montée, pas depuis le début", w.Level())
	}

	w.Step(Vec{})
	if w.Level() != 3 {
		t.Errorf("niveau : %d au dixième tick, attendu 3", w.Level())
	}
}

// TestSeuilPremierNulRefuse garde le fichier qui ferait boucler la montée.
//
// À seuil nul, chaque tour de la distribution retire zéro gemme et donne un
// niveau : la partie se figerait dans une boucle infinie au premier tick. Le
// refus est au chargement plutôt qu'un plafond de tours dans la boucle, qui
// aurait laissé tourner un fichier faux en le masquant.
func TestSeuilPremierNulRefuse(t *testing.T) {
	fsys := fstest.MapFS{"p.json": &fstest.MapFile{Data: []byte(`{
		"version_format": 1,
		"progression": {
			"niveaux": {"seuil_premier": 0, "seuil_increment": 2, "plancher_ms": 45000}
		}
	}`)}}

	_, err := LoadProgression(fsys, "p.json")
	var invalide *manifest.Invalid
	if !errors.As(err, &invalide) {
		t.Fatalf("seuil premier nul accepté : %v", err)
	}
}

// TestChampsDeProgressionManquantsListesEnUneFois vérifie que l'auteur reçoit la
// liste et non le premier manquement.
func TestChampsDeProgressionManquantsListesEnUneFois(t *testing.T) {
	fsys := fstest.MapFS{"p.json": &fstest.MapFile{Data: []byte(`{
		"version_format": 1,
		"progression": {"niveaux": {}}
	}`)}}

	_, err := LoadProgression(fsys, "p.json")
	var invalide *manifest.Invalid
	if !errors.As(err, &invalide) {
		t.Fatalf("section de niveaux vide acceptée : %v", err)
	}
	// Les trois seuils et les deux champs de la gemme. Une absence compte pour
	// une ligne : les bornes ne se prononcent que sur un champ présent, faute de
	// quoi le nombre de lignes cesserait d'être le nombre de choses à corriger.
	if len(invalide.Missing) != 5 {
		t.Errorf("%d manquement(s), attendu 5 :\n  %v", len(invalide.Missing), invalide.Missing)
	}
}

// TestFormatDeProgressionNonPrisEnCharge vérifie la sentinelle partagée.
func TestFormatDeProgressionNonPrisEnCharge(t *testing.T) {
	fsys := fstest.MapFS{"p.json": &fstest.MapFile{Data: []byte(`{
		"version_format": 99,
		"progression": {"niveaux": {}}
	}`)}}

	_, err := LoadProgression(fsys, "p.json")
	if !errors.Is(err, manifest.ErrUnsupportedFormat) {
		t.Fatalf("format 99 accepté : %v", err)
	}
}
