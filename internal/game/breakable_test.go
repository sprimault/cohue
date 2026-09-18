// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas de l'obstacle fragile : la touche tenue et sa cadence, la portée que la
// géométrie impose, ce qu'il rend en cédant, et les refus qu'un semis mal écrit
// reçoit au chargement.

package game

import (
	"slices"
	"strings"
	"testing"

	"github.com/sprimault/cohue"
)

// obstaclesLivres rend la table des obstacles fragiles du catalogue livré.
func obstaclesLivres(t *testing.T) []BreakableKind {
	t.Helper()
	catalogue, err := LoadObjects(cohue.Assets, manifesteObjets)
	if err != nil {
		t.Fatalf("objets livrés : %v", err)
	}
	sortes, err := catalogue.Breakables()
	if err != nil {
		t.Fatalf("obstacles livrés : %v", err)
	}
	return sortes
}

// rangDeLObstacle rend la place d'une sorte dans la table, ou arrête le test.
func rangDeLObstacle(t *testing.T, sortes []BreakableKind, cle string) int {
	t.Helper()
	i := slices.IndexFunc(sortes, func(s BreakableKind) bool { return s.Key == cle })
	if i < 0 {
		t.Fatalf("« %s » n'est pas un obstacle du catalogue livré", cle)
	}
	return i
}

// salleAvecObstacle monte une salle coupée par un mur, dont les ouvertures sont
// fermées par les obstacles donnés, dans l'ordre du montage.
//
// Le mur court le long de u sur la rangée dix ; les obstacles se posent dans ses
// percées, aux colonnes données. Le semis passe par `CompileBreakables`, pour la
// raison qui y fait passer les caisses : une entrée de test doit être une entrée
// que le système accepterait.
//
// **Les armes livrées et non inertes**, parce que la frappe se cadence sur l'arme
// de base de la table : une arme vidée frapperait à chaque tick, et aucun cas de
// cadence n'aurait rien à dire.
func salleAvecObstacle(t *testing.T, cle string, colonnes ...int) *World {
	t.Helper()
	profils, err := LoadProfiles(cohue.Assets, manifestePersonnages)
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}
	armes, err := LoadWeapons(cohue.Assets, manifesteArmes)
	if err != nil {
		t.Fatalf("armes livrées : %v", err)
	}
	sortes := obstaclesLivres(t)

	g := NewCostGrid(24, 24)
	for u := range 24 {
		if !slices.Contains(colonnes, u) {
			g.Set(u, 10, Blocked)
		}
	}
	spec := make(BreakableSpec, 0, len(colonnes))
	for _, u := range colonnes {
		spec = append(spec, BreakablePlacementSpec{Object: cle, At: &[2]int{u, 10}})
	}
	poses, manques := CompileBreakables(spec, sortes, g, nil, nil)
	if len(manques) > 0 {
		t.Fatalf("semis d'essai refusé : %v", manques)
	}
	StampBreakables(g, poses)

	w := NewWorld(profils, armes, progressionLivree(t), caissesLivrees(t), fiolesLivrees(t),
		vagueUnique(0), g, graineDeTest, capacitesDeTest)
	// Sept dixièmes de tuile au sud du centre du premier : contre lui, dans la
	// case voisine.
	w.Place(FromInt(colonnes[0])+One/2, FromInt(11)+One/5)
	w.Erect(sortes, poses)
	return w
}

// tenir joue des ticks touche tenue jusqu'à ce que le semis perde un obstacle,
// et rend combien il en a fallu.
func tenir(t *testing.T, w *World, borne int) int {
	t.Helper()
	debout := w.Breakables().Len()
	for n := 1; n <= borne; n++ {
		w.Interact()
		w.Step(Vec{})
		if w.Breakables().Len() < debout {
			return n
		}
	}
	t.Fatalf("l'obstacle tient encore après %d ticks de touche tenue", borne)
	return 0
}

// TestUnObstacleCedeAuBoutDeSesTouches garde le prix, et ce que la rupture rend.
//
// **Le prix se compte en touches à la cadence du socle**, la première portant
// dès l'appui : un obstacle de cinq touches cède au bout de quatre cadences et un
// tick. Le compte se dérive du catalogue et de la table d'armes plutôt que de
// s'écrire ici, faute de quoi le cas resterait vert le jour où la vitrine en
// encaisserait six.
//
// Ce qu'il rend est gardé dans le même cas, parce que c'est la même rupture : la
// case au sol, une ruine qui dit ce qui a cédé, une volée qui le dit aussi.
func TestUnObstacleCedeAuBoutDeSesTouches(t *testing.T) {
	w := salleAvecObstacle(t, "vitrine", 12)
	o := *w.Breakables().At(0)
	cle := w.BreakableKey(o.Kind)
	attendu := (o.Hits-1)*int(w.armes.Base.Cooldown) + 1

	if n := tenir(t, w, 4*attendu); n != attendu {
		t.Errorf("cédé au tick %d, attendu %d : %d touches à %d ticks d'écart",
			n, attendu, o.Hits, w.armes.Base.Cooldown)
	}

	if got := w.grille.At(12, 10); got != o.Floor {
		t.Errorf("la case porte %d après la rupture, attendu le sol de %d", got, o.Floor)
	}
	if w.Wrecks().Len() != 1 || w.Wrecks().At(0).Object != cle {
		t.Errorf("aucune ruine de « %s » après la rupture", cle)
	}
	if w.Fxs().Len() == 0 || w.Fxs().At(0).Kind != FxBreak || w.Fxs().At(0).Object != cle {
		t.Errorf("aucune volée de « %s » après la rupture", cle)
	}
}

// TestMartelerNeFrappePasPlusVite garde la touche tenue contre le martèlement.
//
// **C'est la décision du geste qu'il protège.** Un décompte remis à zéro au
// relâchement ferait frapper chaque appui : un joueur qui martèle forcerait un
// rideau de fer en vingt-quatre ticks au lieu de quatre secondes et demie, et
// tenir la touche deviendrait la mauvaise façon de jouer. Il faut donc un décompte qui descend
// que la touche soit tenue ou non.
func TestMartelerNeFrappePasPlusVite(t *testing.T) {
	tenu := tenir(t, salleAvecObstacle(t, "vitrine", 12), 1000)

	w := salleAvecObstacle(t, "vitrine", 12)
	for n := 1; n <= 1000; n++ {
		if n%2 == 1 {
			w.Interact()
		}
		w.Step(Vec{})
		if w.Breakables().Len() == 0 {
			if n < tenu {
				t.Errorf("cédé en %d ticks à coups répétés, contre %d touche tenue", n, tenu)
			}
			return
		}
	}
	t.Fatal("l'obstacle tient encore après mille ticks de martèlement")
}

// TestLaPorteeDepasseLaBorneDeLObstacle garde la portée contre la géométrie.
//
// **Le cas se joue au contact réel, pas à une distance choisie.** Le joueur
// marche droit sur l'obstacle jusqu'à ce que sa case l'arrête ; c'est la plus
// petite distance qu'il puisse atteindre, et une portée posée sur elle serait
// fausse en permanence sans jamais le dire. Au centre de la case voisine, il ne
// frappe pas : on force en venant contre.
func TestLaPorteeDepasseLaBorneDeLObstacle(t *testing.T) {
	w := salleAvecObstacle(t, "rideau_fer", 12)
	plein := w.Breakables().At(0).Hits

	w.Place(FromInt(12)+One/2, FromInt(11)+One/2)
	w.Interact()
	w.Step(Vec{})
	if got := w.Breakables().At(0).Hits; got != plein {
		t.Fatalf("frappé depuis le centre de la case voisine : %d touches sur %d", got, plein)
	}

	for range 2 * TPS {
		w.Step(Vec{Y: -One})
	}
	if _, y := w.Player(); y.Floor() != 11 {
		t.Fatalf("le joueur a quitté sa case en poussant contre l'obstacle : rangée %d", y.Floor())
	}
	w.frappe = 0
	w.Interact()
	w.Step(Vec{Y: -One})
	if got := w.Breakables().At(0).Hits; got != plein-1 {
		t.Errorf("%d touches sur %d au contact : la portée n'atteint pas la borne", got, plein)
	}
}

// TestUneFrappeNePorteQueSurUnObstacle garde qu'on paie chaque obstacle entier.
//
// Deux obstacles côte à côte sont tous deux à portée d'un joueur posé entre eux.
// Frappés ensemble, tenir contre deux diviserait le prix de chacun, et une
// rangée de vitrines se forcerait au prix d'une seule.
func TestUneFrappeNePorteQueSurUnObstacle(t *testing.T) {
	w := salleAvecObstacle(t, "vitrine", 11, 12)
	w.Place(FromInt(12), FromInt(11)+One/5)
	avant := w.Breakables().At(0).Hits + w.Breakables().At(1).Hits

	w.Interact()
	w.Step(Vec{})

	if apres := w.Breakables().At(0).Hits + w.Breakables().At(1).Hits; apres != avant-1 {
		t.Errorf("%d touches retirées d'une frappe, attendu une", avant-apres)
	}
}

// TestSansLaToucheRienNeCede garde que se tenir contre ne suffit pas.
//
// C'est ce qui sépare l'obstacle de la caisse : l'une cède à l'appui, l'autre à
// la touche. Un obstacle qui céderait au contact ferait de chaque couloir étroit
// un passage qu'on ouvre en le longeant.
func TestSansLaToucheRienNeCede(t *testing.T) {
	w := salleAvecObstacle(t, "grille_ventilation", 12)
	plein := w.Breakables().At(0).Hits

	for range 5 * TPS {
		w.Step(Vec{Y: -One})
	}
	if got := w.Breakables().At(0).Hits; got != plein {
		t.Errorf("%d touches sur %d sans que la touche soit tenue", got, plein)
	}
}

// TestLesTouchesPorteesNeSeRegagnentPas garde qu'un obstacle entamé le reste.
//
// À l'inverse de l'appui d'une caisse, que s'écarter remet à neuf. Revenir finir
// un rideau entamé est une décision que le jeu doit payer moins cher que de le
// recommencer ; l'effacer rendrait chaque retraite sous la horde vaine.
func TestLesTouchesPorteesNeSeRegagnentPas(t *testing.T) {
	w := salleAvecObstacle(t, "rideau_fer", 12)
	plein := w.Breakables().At(0).Hits

	w.Interact()
	w.Step(Vec{})
	w.Place(FromInt(12)+One/2, FromInt(20)+One/2)
	for range 5 * TPS {
		w.Step(Vec{})
	}
	if got := w.Breakables().At(0).Hits; got != plein-1 {
		t.Errorf("%d touches sur %d après s'être écarté, attendu %d", got, plein, plein-1)
	}
}

// TestFrapperUnObstacleNalloueRien garde le budget sur le chemin de la frappe.
//
// **Une frappe par exécution, et c'est ce qui rend la mesure honnête.**
// `AllocsPerRun` rend une moyenne arrondie : une frappe toutes les vingt-quatre
// exécutions y disparaîtrait. Les armes inertes y pourvoient — leur socle vidé a
// une cadence nulle, si bien que chaque tick frappe —, et l'obstacle encaisse
// assez pour ne jamais céder pendant la mesure.
func TestFrapperUnObstacleNalloueRien(t *testing.T) {
	profils, err := LoadProfiles(cohue.Assets, manifestePersonnages)
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}
	w := NewWorld(profils, armesInertes(t), progressionLivree(t), caissesLivrees(t),
		fiolesLivrees(t), vagueUnique(0), NewCostGrid(16, 16), graineDeTest, capacitesDeTest)
	w.Place(FromInt(8)+One/2, FromInt(8)+One/5)
	w.Erect([]BreakableKind{{Key: "vitrine", Hits: 1 << 20}},
		[]BreakablePlacement{{X: FromInt(8) + One/2, Y: FromInt(7) + One/2}})

	moyenne := testing.AllocsPerRun(1000, func() {
		w.Interact()
		w.Step(Vec{})
	})
	if moyenne != 0 {
		t.Errorf("%v allocation(s) par frappe, attendu aucune", moyenne)
	}
	if w.Breakables().At(0).Hits == 1<<20 {
		t.Error("aucune frappe n'a porté : la mesure n'a rien mesuré")
	}
}

// TestUnSemisDObstaclesSeCompile garde ce que la compilation relève.
//
// Le sol recouvert se lit sur la carte avant qu'aucun blocage n'y soit écrit :
// c'est lui que l'obstacle rend en cédant, et le relever plus tard ferait
// mémoriser le blocage comme étant le sol.
func TestUnSemisDObstaclesSeCompile(t *testing.T) {
	sortes := obstaclesLivres(t)
	g := NewCostGrid(8, 8)
	g.Set(3, 3, 2)

	poses, manques := CompileBreakables(BreakableSpec{
		{Object: "vitrine", At: &[2]int{3, 3}},
		{Object: "rideau_fer", At: &[2]int{5, 5}, Axis: AxisV},
	}, sortes, g, nil, nil)
	if len(manques) > 0 {
		t.Fatalf("semis valide refusé : %v", manques)
	}
	if poses[0].Floor != 2 || poses[0].Across {
		t.Errorf("vitrine : sol %d, travers %t ; attendu le sol de 2, le long de u",
			poses[0].Floor, poses[0].Across)
	}
	if sortes[poses[1].Kind].Key != "rideau_fer" || !poses[1].Across {
		t.Errorf("rideau : sorte « %s », travers %t ; attendu le rideau, le long de v",
			sortes[poses[1].Kind].Key, poses[1].Across)
	}
}

// TestUnSemisDObstaclesMalEcritSeRefuse garde chacun des refus, et qu'ils se
// disent tous.
//
// Chaque cas porte un seul défaut, pour qu'un refus qui en masquerait un autre
// se voie ; le dernier les réunit, pour que la liste soit rendue entière plutôt
// qu'au premier.
func TestUnSemisDObstaclesMalEcritSeRefuse(t *testing.T) {
	sortes := obstaclesLivres(t)
	g := NewCostGrid(8, 8)
	g.Set(1, 1, Blocked)
	caisses := []CratePlacement{{X: FromInt(2) + One/2, Y: FromInt(2) + One/2}}
	ambiance := []AmbientPlacement{{X: FromInt(3) + One/2, Y: FromInt(3) + One/2}}
	a := func(u, v int) *[2]int { return &[2]int{u, v} }

	cas := []struct {
		nom    string
		spec   BreakableSpec
		attend string
	}{
		{"nom inconnu", BreakableSpec{{Object: "caisse", At: a(5, 5)}}, "connus : ["},
		{"sens inconnu", BreakableSpec{{Object: "vitrine", At: a(5, 5), Axis: "x"}}, "attendu u ou v"},
		{"sans position", BreakableSpec{{Object: "vitrine"}}, "absente"},
		{"hors du lieu", BreakableSpec{{Object: "vitrine", At: a(9, 5)}}, "hors du lieu"},
		{"dans un mur", BreakableSpec{{Object: "vitrine", At: a(1, 1)}}, "dans un mur"},
		{"sur une caisse", BreakableSpec{{Object: "vitrine", At: a(2, 2)}}, "caisses[0]"},
		{"sur un figurant", BreakableSpec{{Object: "vitrine", At: a(3, 3)}}, "ambiance[0]"},
		{"deux sur une case", BreakableSpec{
			{Object: "vitrine", At: a(5, 5)}, {Object: "rideau_fer", At: a(5, 5)},
		}, "destructibles[0]"},
	}
	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			_, manques := CompileBreakables(c.spec, sortes, g, caisses, ambiance)
			if len(manques) != 1 || !strings.Contains(manques[0], c.attend) {
				t.Errorf("manquements %q, attendu un seul qui dise %q", manques, c.attend)
			}
		})
	}

	var tous BreakableSpec
	for _, c := range cas {
		tous = append(tous, c.spec[len(c.spec)-1])
	}
	if _, manques := CompileBreakables(tous, sortes, g, caisses, ambiance); len(manques) < len(cas)-1 {
		t.Errorf("%d manquement(s) rendus pour %d défauts : la liste s'arrête au premier",
			len(manques), len(cas))
	}
}

// TestLeCatalogueRangeSesObstacles garde l'ordre de la table et le refus de zéro.
//
// **L'ordre parce que le rang est ce qu'une entité retient** : une table rangée
// dans l'ordre d'une `map` changerait d'un lancement à l'autre, et avec elle
// l'empreinte d'une run. Le refus parce qu'une clé `touches` oubliée se décode à
// zéro, et un obstacle qui cède sans qu'on frappe n'est pas un choix d'auteur.
func TestLeCatalogueRangeSesObstacles(t *testing.T) {
	sortes := obstaclesLivres(t)
	if len(sortes) != 4 {
		t.Fatalf("%d obstacle(s) au catalogue livré, la conception en nomme quatre", len(sortes))
	}
	if !slices.IsSortedFunc(sortes, func(a, b BreakableKind) int { return strings.Compare(a.Key, b.Key) }) {
		t.Errorf("table non triée : %v", sortes)
	}

	catalogue := &Objects{Items: map[string]Object{
		"vitrine": {Destruction: &Destruction{Mode: ModeInteraction}},
	}}
	if _, err := catalogue.Breakables(); err == nil {
		t.Error("un obstacle à zéro touche s'est chargé")
	}
}
