// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas de la progression : le manifeste livré, le seuil qui monte, les deux
// sources d'une montée de niveau, et ce que le plancher de temps ne remet pas à
// zéro.

package game

import (
	"errors"
	"strings"
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

// collecte complète des seuils forgés par ce que la collecte exige.
//
// **Sans cela, ces cas passent pour une mauvaise raison.** `semer` pose les
// gemmes exactement sur le joueur : une portée nulle les ramasse quand même, la
// distance valant zéro, et une durée de vie nulle ne se voit pas puisque le
// ramassage est évalué avant l'expiration. Les deux réglages seraient
// indiscernables de leur absence, et c'est ce que la mutation a montré.
//
// **La période d'aimant est mise hors d'atteinte pour la raison inverse** : à
// zéro, un aimant apparaît dès le premier tick de tous ces cas, consomme le flux
// des positions et se ramasse peut-être — un mécanisme entier qui tourne dans
// des cas qui ne parlent pas de lui.
//
// Ces cas comptent des niveaux, pas des distances, des âges ni des aimants, et
// ce qui les concerne a ses propres tests.
func collecte(p *Progression) *Progression {
	p.PickupRange = One
	p.GemLife = 100000
	p.MagnetPeriod = 100000
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
	w := NewWorld(profils, armesInertes(t), seuils, sansVagues(), NewCostGrid(32, 32),
		graineDeTest, capacitesDeTest)
	w.Place(FromInt(16)+One/2, FromInt(16)+One/2)
	return w, profils
}

// TestLeTickQuiOuvreUnChoixNalloueRien garde l'invariant là où l'autre ne va pas.
//
// **`TestLaBoucleNalloueRien` ne peut pas voir ce tick-ci**, et pour deux
// raisons qui se cumulent. Son monde monte sur les seuils livrés, dont le
// plancher vaut deux mille sept cents ticks, et il en joue mille dix-huit : la
// montée par le temps n'y arrive jamais, donc `offrir` n'y est jamais appelé.
//
// **Et l'allonger n'aurait pas suffi**, ce qui est le point à retenir :
// `AllocsPerRun` arrondit sa moyenne à l'entier. Un plancher franchi trente
// fois en mille ticks donne trois allocations par ouverture, soit un dixième
// par tick, soit zéro après arrondi. Un test d'allocation ne voit qu'un coût
// que **chacune** de ses exécutions paie.
//
// D'où la forme : un plancher d'un tick, et la carte prise avant le pas. Chaque
// exécution mesurée ouvre alors exactement un choix. Le réglage est extrême et
// c'est ce qui isole le tick — ce que le test garde n'est pas le plancher mais
// ce que son franchissement coûte.
func TestLeTickQuiOuvreUnChoixNalloueRien(t *testing.T) {
	w, _ := champDeProgression(t, collecte(&Progression{
		FirstThreshold: 100000, GemValue: 1, Floor: 1,
	}))

	moyenne := testing.AllocsPerRun(1000, func() {
		w.Choose(0)
		w.Step(Vec{})
	})
	if moyenne != 0 {
		t.Errorf("%v allocation(s) par tick ouvrant un choix, attendu aucune", moyenne)
	}
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
	w, profils := champDeProgression(t, collecte(&Progression{
		FirstThreshold: 6, GemValue: 3, Floor: 1000,
	}))

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

// TestUnPlancherNulRefuse garde la frontière entre l'absence et le zéro.
//
// **Un plancher à zéro n'est pas « pas de plancher », c'est son contraire** :
// `progresser` compare le compteur au plancher à chaque tick, donc un zéro donne
// un niveau et un choix soixante fois par seconde. Le champ absent est déjà
// signalé par `exige` ; c'est le champ écrit à zéro qui passait, parce que la
// conversion n'était appelée que pour une durée strictement positive.
//
// Ses deux voisins de fichier, `duree_vie_ms` et `periode_ms`, ont ce refus
// depuis toujours et disent chacun sa conséquence. Celui-ci l'avait perdu.
func TestUnPlancherNulRefuse(t *testing.T) {
	fsys := fstest.MapFS{"p.json": &fstest.MapFile{Data: []byte(`{
		"version_format": 1,
		"progression": {
			"niveaux": {"seuil_premier": 10, "seuil_increment": 2, "plancher_ms": 0},
			"gemmes": {"objet": "gemme", "experience": 1, "portee_ramassage_tuiles": 1.0,
			           "duree_vie_ms": 6000},
			"aimant": {"objet": "aimant", "periode_ms": 30000, "distance_min_tuiles": 6.0,
			           "vitesse_gemme_tuiles_s": 12.0},
			"pression": {"rayon_apparition_tuiles": 19.0, "report_ms": 3000}
		}
	}`)}}

	_, err := LoadProgression(fsys, "p.json")
	var invalide *manifest.Invalid
	if !errors.As(err, &invalide) {
		t.Fatalf("plancher nul accepté : %v", err)
	}
	if !strings.Contains(err.Error(), "plancher_ms") {
		t.Errorf("le refus ne nomme pas la clé fautive : %v", err)
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
	w, profils := champDeProgression(t, collecte(&Progression{
		FirstThreshold: 3, Increment: 4, GemValue: 1, Floor: 1000,
	}))

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
	w, profils := champDeProgression(t, collecte(&Progression{
		FirstThreshold: 3, GemValue: 1, Floor: 1000,
	}))

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
	w, _ := champDeProgression(t, collecte(&Progression{
		FirstThreshold: 1000, GemValue: 1, Floor: 7,
	}))

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
	w, profils := champDeProgression(t, collecte(&Progression{
		FirstThreshold: 5, GemValue: 1, Floor: 7,
	}))

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
	w, profils := champDeProgression(t, collecte(&Progression{
		FirstThreshold: 3, GemValue: 1, Floor: 7,
	}))

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
	// Les trois seuils, les quatre champs de la gemme, les quatre de l'aimant et
	// les deux de la pression. Une absence compte pour une ligne : les bornes ne
	// se prononcent que sur un champ présent, faute de quoi le nombre de lignes
	// cesserait d'être le nombre de choses à corriger.
	if len(invalide.Missing) != 13 {
		t.Errorf("%d manquement(s), attendu 13 :\n  %v", len(invalide.Missing), invalide.Missing)
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
