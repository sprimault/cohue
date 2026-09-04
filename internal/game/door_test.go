// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas de la sortie : ce qui l'ouvre, ce qui ne l'ouvre pas, ce que la
// franchir termine, et les refus qu'un lieu mal écrit reçoit au chargement.

package game

import "testing"

// La porte des cas ci-dessous, et le compte qu'elle demande. Une case du bord
// plutôt que le centre : le joueur monté au milieu n'est pas dessus au premier
// tick, ce qui laisse chaque cas dire quand il s'en approche.
const (
	porteU, porteV = 10, 10
	porteAbattus   = 3
)

// salleAvecPorte monte une salle ouverte dont une case est murée, et y pose une
// sortie qui demande trois créatures.
// La phase est vide de profils plutôt qu'absente : le spawner n'y achète rien,
// donc le compte des abattus ne bouge que par ce que chaque cas pose lui-même,
// et le durcissement d'une créature posée à la main trouve la table qu'il lit.
func salleAvecPorte(t *testing.T) (*World, *Profiles) {
	t.Helper()
	w, profils := salleOuverte(t, vagueUnique(0), 16)
	w.grille.Set(porteU, porteV, Blocked)
	w.SetExit(&Exit{
		X:     FromInt(porteU) + One/2,
		Y:     FromInt(porteV) + One/2,
		Kills: porteAbattus,
		U:     porteU,
		V:     porteV,
	})
	return w, profils
}

// TestLaPorteSOuvreAuCompte vérifie qu'elle ne s'ouvre ni avant ni jamais.
//
// Le seuil se franchit par le haut : le cas tue une créature de plus que
// l'objectif après l'avoir atteint, pour qu'un `==` mal placé se voie.
func TestLaPorteSOuvreAuCompte(t *testing.T) {
	w, _ := salleAvecPorte(t)

	for i := range porteAbattus + 1 {
		if ouverte := w.DoorOpen(); ouverte != (i >= porteAbattus) {
			t.Fatalf("après %d abattus, porte ouverte = %v", i, ouverte)
		}
		w.SpawnEnemy(0, FromInt(30), FromInt(30))
		w.Enemies().At(w.Enemies().Len() - 1).Hits = 0
		w.Step(Vec{})
	}
	if !w.DoorOpen() {
		t.Errorf("porte fermée après %d abattus, objectif %d", w.Kills(), porteAbattus)
	}
}

// TestUnFigurantNOuvrePasLaPorte garde la conséquence que le lot du Passant a
// posée sans pouvoir la vérifier.
//
// **Le chapitre 4 prévient qu'un lieu peuplé ouvrirait sa porte tout seul**, et
// que son auteur n'aurait aucun moyen de comprendre pourquoi. La règle était
// écrite depuis que les figurants existent, et rien ne pouvait la contredire
// faute d'objectif à compter : c'est l'objectif qui la rend exigible.
//
// Ce qui la tient est structurel — un figurant n'est pas dans le bassin des
// ennemis, et le compte se fait en sortant de ce bassin. Le cas garde donc une
// propriété de découpage, celle qui se perd le jour où quelqu'un range les deux
// ensemble pour mutualiser une boucle.
func TestUnFigurantNOuvrePasLaPorte(t *testing.T) {
	w, profils := salleAvecPorte(t)
	if len(profils.Ambient) == 0 {
		t.Fatal("aucun profil d'ambiance livré, ce cas ne sépare rien")
	}
	w.ambiants = NewPool[Ambient](8)
	for i := range 8 {
		w.SpawnAmbient(0, FromInt(20+i), FromInt(20))
	}

	for range 5 * TPS {
		w.Step(Vec{})
	}
	if w.Kills() != 0 {
		t.Errorf("%d abattus avec des figurants seuls et aucune créature", w.Kills())
	}
	if w.DoorOpen() {
		t.Error("la porte s'est ouverte sur des figurants : un lieu peuplé se termine tout seul")
	}
}

// TestFranchirTermineSansTuer vérifie que sortir est une fin, et pas la mort.
func TestFranchirTermineSansTuer(t *testing.T) {
	w, _ := salleAvecPorte(t)
	for range porteAbattus {
		w.SpawnEnemy(0, FromInt(30), FromInt(30))
		w.Enemies().At(w.Enemies().Len() - 1).Hits = 0
		w.Step(Vec{})
	}

	// Contre la porte, du côté franchissable : la case elle-même est murée.
	w.Place(FromInt(porteU)+One/2, FromInt(porteV+1)+One/2)
	w.Step(Vec{})

	if !w.Escaped() {
		t.Fatal("le joueur au contact d'une porte ouverte n'est pas sorti")
	}
	if !w.Over() {
		t.Error("la partie continue après la sortie")
	}
	if !w.Alive() {
		t.Error("sortir a tué le joueur : les deux fins se confondent")
	}
}

// TestUnePorteFermeeNeLaissePasSortir garde ce qui fait qu'une sortie se gagne.
func TestUnePorteFermeeNeLaissePasSortir(t *testing.T) {
	w, _ := salleAvecPorte(t)
	w.Place(FromInt(porteU)+One/2, FromInt(porteV+1)+One/2)

	for range TPS {
		w.Step(Vec{})
	}
	if w.Escaped() {
		t.Errorf("sorti par une porte fermée, %d abattus sur %d", w.Kills(), porteAbattus)
	}
	if w.Over() {
		t.Error("la partie s'est terminée sans que rien la termine")
	}
}

// TestUnLieuSansPorteNeSeTerminePas vérifie qu'une salle sans sortie reste
// jouable jusqu'à la mort.
//
// Un lieu peut n'en avoir aucune — une boutique, un passage —, et ce n'est pas
// un défaut de fichier : c'est le cas où `DoorOpen` doit répondre non plutôt que
// de lire une porte absente.
func TestUnLieuSansPorteNeSeTerminePas(t *testing.T) {
	w, _ := salleOuverte(t, vagueUnique(0), 16)
	for range TPS {
		w.Step(Vec{})
	}
	if w.DoorOpen() || w.Escaped() || w.Over() {
		t.Errorf("lieu sans sortie : ouverte=%v sorti=%v fini=%v",
			w.DoorOpen(), w.Escaped(), w.Over())
	}
}

// TestUneSortieMalEcriteSeRefuse vérifie que le chargement nomme ce qui cloche.
//
// **Le cas de la case franchissable est le seul qui ne saute pas aux yeux**, et
// c'est le plus grave : la porte s'y traverserait avant d'être gagnée, si bien
// que le lieu se terminerait en marchant dessus. Les autres se voient en jouant
// dix secondes ; celui-là se voit une fois, par accident.
func TestUneSortieMalEcriteSeRefuse(t *testing.T) {
	carte := NewCostGrid(16, 16)
	carte.Set(4, 4, Blocked)
	trois := 3
	zero := 0

	cas := []struct {
		nom    string
		brut   ExitSpec
		attend string
	}{
		{"position absente", ExitSpec{Kills: &trois}, "position"},
		{"objectif absent", ExitSpec{At: &[2]int{4, 4}}, "abattus"},
		{"objectif nul", ExitSpec{At: &[2]int{4, 4}, Kills: &zero}, "abattus"},
		{"hors du lieu", ExitSpec{At: &[2]int{99, 4}, Kills: &trois}, "hors du lieu"},
		{"sur du sol libre", ExitSpec{At: &[2]int{5, 5}, Kills: &trois}, "franchissable"},
	}
	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			sortie, manques := CompileExit(&c.brut, carte)
			if sortie != nil {
				t.Fatal("une sortie a été compilée là où le fichier est faux")
			}
			if len(manques) == 0 {
				t.Fatal("aucun manquement rendu")
			}
			if !contient(manques, c.attend) {
				t.Errorf("le refus ne nomme pas %q : %v", c.attend, manques)
			}
		})
	}
}

// TestUneSortieEcriteSeCompile vérifie que le cas valide passe, et où il pose la
// porte.
//
// Au centre de la case et non sur son coin, comme un figurant : c'est de là que
// se mesure la distance au joueur, et un décalage d'une demi-case ferait sortir
// par le mur voisin.
func TestUneSortieEcriteSeCompile(t *testing.T) {
	carte := NewCostGrid(16, 16)
	carte.Set(4, 4, Blocked)
	trois := 3

	sortie, manques := CompileExit(&ExitSpec{At: &[2]int{4, 4}, Kills: &trois}, carte)
	if len(manques) > 0 {
		t.Fatalf("une sortie bien écrite est refusée : %v", manques)
	}
	if sortie.X != FromInt(4)+One/2 || sortie.Y != FromInt(4)+One/2 {
		t.Errorf("porte en (%v, %v), attendue au centre de sa case", sortie.X, sortie.Y)
	}
	if sortie.Kills != trois {
		t.Errorf("objectif %d, attendu %d", sortie.Kills, trois)
	}
}

// TestUnLieuSansSortieNeSeRefusePas vérifie que l'absence n'est pas une faute.
func TestUnLieuSansSortieNeSeRefusePas(t *testing.T) {
	sortie, manques := CompileExit(nil, NewCostGrid(16, 16))
	if sortie != nil || len(manques) > 0 {
		t.Errorf("un lieu sans sortie est refusé : %v, %v", sortie, manques)
	}
}
