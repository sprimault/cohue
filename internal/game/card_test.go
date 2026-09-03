// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas du choix : les trois places toujours remplies, ce qu'une carte change,
// l'axe épuisé qui sort du menu, et les montées qui s'accumulent sans se perdre.

package game

import (
	"testing"

	"github.com/sprimault/cohue"
)

// champDeCartes monte une salle vide sur les passifs livrés.
//
// L'arme est celle du fichier, non neutralisée : la table des passifs se contrôle
// contre elle au chargement, et une arme nulle rendrait le monde d'essai
// incapable de montrer ce qu'un palier de cadence retire. Elle ne tire pas pour
// autant — aucune créature n'entre dans le bassin de ces cas.
func champDeCartes(t *testing.T, seuils *Progression) (*World, *Profiles) {
	t.Helper()
	profils, err := LoadProfiles(cohue.Assets, manifestePersonnages)
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}
	armes, err := LoadWeapons(cohue.Assets, manifesteArmes)
	if err != nil {
		t.Fatalf("armes livrées : %v", err)
	}

	w := NewWorld(profils, armes, seuils, sansVagues(), NewCostGrid(32, 32), graineDeTest, 16, 64, 32)
	w.Place(FromInt(16)+One/2, FromInt(16)+One/2)
	return w, profils
}

// monteeSimple rend un réglage où une gemme suffit à monter d'un niveau.
//
// Le plancher est hors d'atteinte : ces cas comptent des choix, et une montée
// donnée par le temps au milieu d'une boucle ferait dériver le compte sans que
// le message le dise.
func monteeSimple() *Progression {
	return collecte(&Progression{FirstThreshold: 1, GemValue: 1, Floor: 100000})
}

// TestUneMonteeOuvreTroisPlaces garde la règle du choix ternaire.
//
// Trois, et jamais moins : un écran de montée qui n'offrirait que ce qui reste
// tomberait à deux places dès qu'un axe s'épuise, c'est-à-dire au moment où le
// joueur a le plus joué. C'est la soupape qui remplit, et c'est sa raison d'être.
func TestUneMonteeOuvreTroisPlaces(t *testing.T) {
	w, profils := champDeCartes(t, monteeSimple())

	if w.Choosing() {
		t.Fatal("un choix est ouvert avant toute montée")
	}
	semer(t, w, profils, 1)
	w.Step(Vec{})

	if len(w.Pending()) != Choices {
		t.Fatalf("%d place(s) offerte(s), attendu %d", len(w.Pending()), Choices)
	}
}

// TestLeMenuEstLesDeuxAxesPuisLaSoupape écrit ce que le jalon offre réellement.
//
// **Le menu ne varie pas, et c'est une conséquence arithmétique.** Deux axes
// plus la soupape font exactement trois places : il n'y a rien à choisir parmi
// les éligibles, donc aucun tirage n'a lieu. Ce test est ce qui rendra le
// changement visible le jour où un troisième axe entrera — il tombera, et c'est
// à ce moment qu'un flux aléatoire prendra son numéro.
func TestLeMenuEstLesDeuxAxesPuisLaSoupape(t *testing.T) {
	w, profils := champDeCartes(t, monteeSimple())
	semer(t, w, profils, 1)
	w.Step(Vec{})

	cartes := w.Pending()
	if cartes[0].Name != "Cadence" || cartes[1].Name != "Portée" {
		t.Errorf("les deux premières places : %q et %q", cartes[0].Name, cartes[1].Name)
	}
	if cartes[2].Name != w.passifs.Relief.Name {
		t.Errorf("la troisième place : %q, attendu la soupape", cartes[2].Name)
	}
	// Le rang sur la borne, et non la grandeur du gain : c'est ce que le joueur
	// ne peut pas déduire autrement, l'épuisement d'un axe étant un moment de jeu.
	if cartes[0].Effect != "Palier 1 sur 6" {
		t.Errorf("effet de la première carte : %q", cartes[0].Effect)
	}
}

// TestChoisirAppliqueLePalier vérifie que la carte agit sur l'arme de la partie.
func TestChoisirAppliqueLePalier(t *testing.T) {
	w, profils := champDeCartes(t, monteeSimple())
	avant := w.arme.Cooldown

	semer(t, w, profils, 1)
	w.Step(Vec{})
	w.Choose(0)

	// Deux ticks, ce que le manifeste déclare comme pas de cadence.
	if w.arme.Cooldown != avant-2 {
		t.Errorf("cadence : %d ticks, attendu %d", w.arme.Cooldown, avant-2)
	}
	if w.Choosing() {
		t.Error("le choix reste ouvert après avoir été pris")
	}
}

// TestLesPaliersSIndexentCommeLesAxes garde ce qu'un repère lit en jouant.
//
// `Axes` et `TiersTaken` sont deux tranches séparées, donc rien dans le type ne
// les tient ensemble : un croisement d'indices nommerait la cadence en montrant
// le compte de la portée, et le repère attribuerait une bascule au mauvais axe
// sans que rien d'autre ne le dise.
//
// Le cas prend le second axe et non le premier : sur le premier, deux indices
// croisés donneraient le même résultat.
func TestLesPaliersSIndexentCommeLesAxes(t *testing.T) {
	w, profils := champDeCartes(t, monteeSimple())
	semer(t, w, profils, 1)
	w.Step(Vec{})

	pris := w.Pending()[1].Name
	w.Choose(1)

	axes := w.Axes()
	for i := range axes {
		attendu := 0
		if axes[i].Name == pris {
			attendu = 1
		}
		if w.TiersTaken()[i] != attendu {
			t.Errorf("axe %q : %d palier(s) pris, attendu %d",
				axes[i].Name, w.TiersTaken()[i], attendu)
		}
	}
}

// TestUnAxeEpuiseSortDuMenu éprouve la borne, et ce que la soupape fait alors.
//
// Épuiser un axe oblige à basculer sur celui qu'on n'avait pas choisi : c'est ce
// que la conception attend de la borne, et ça ne se voit qu'en prenant six fois
// la même carte.
func TestUnAxeEpuiseSortDuMenu(t *testing.T) {
	w, profils := champDeCartes(t, monteeSimple())

	for range 6 {
		semer(t, w, profils, 1)
		w.Step(Vec{})
		w.Choose(0)
	}

	semer(t, w, profils, 1)
	w.Step(Vec{})

	cartes := w.Pending()
	if len(cartes) != Choices {
		t.Fatalf("%d place(s) après épuisement, attendu %d", len(cartes), Choices)
	}
	if cartes[0].Name != "Portée" {
		t.Errorf("première place : %q, attendu l'axe restant", cartes[0].Name)
	}
	soupape := w.passifs.Relief.Name
	if cartes[1].Name != soupape || cartes[2].Name != soupape {
		t.Errorf("places de remplissage : %q et %q, attendu la soupape deux fois",
			cartes[1].Name, cartes[2].Name)
	}
}

// TestDeuxMonteesDansLeMemeTickNEnPerdentAucune garde la file d'attente.
//
// Une récolte abondante donne deux niveaux dans le même tick — l'aimant en fera
// le cas ordinaire. Écraser le premier choix par le second en retirerait un au
// joueur au moment où il vient d'en gagner deux, et rien à l'écran ne le dirait.
func TestDeuxMonteesDansLeMemeTickNEnPerdentAucune(t *testing.T) {
	w, profils := champDeCartes(t, monteeSimple())

	semer(t, w, profils, 2)
	w.Step(Vec{})
	if w.Level() != 3 {
		t.Fatalf("niveau : %d, attendu 3 après deux gemmes pour un seuil d'une", w.Level())
	}

	w.Choose(0)
	if !w.Choosing() {
		t.Fatal("le second choix ne s'est pas ouvert")
	}
	w.Choose(0)
	if w.Choosing() {
		t.Error("un troisième choix s'est ouvert")
	}
}

// TestLaSoupapeNeDepassePasLeMaximum garde ce qui la rend ignorable en pleine vie.
//
// Un soin qui déborderait donnerait une jauge pleine à un joueur qui n'a rien de
// plus : la carte cesserait d'être situationnelle, donc d'être un choix.
func TestLaSoupapeNeDepassePasLeMaximum(t *testing.T) {
	w, profils := champDeCartes(t, monteeSimple())
	semer(t, w, profils, 1)
	w.Step(Vec{})

	w.Choose(2)
	if w.Health() != w.MaxHealth() {
		t.Errorf("vie : %d, attendu le maximum de %d", w.Health(), w.MaxHealth())
	}
}

// TestUnRangHorsDesPlacesNeFaitRien garde le clavier contre lui-même.
//
// L'appelant est une touche, et une touche pressée au moment où l'écran se ferme
// ne doit pas arrêter le jeu.
func TestUnRangHorsDesPlacesNeFaitRien(t *testing.T) {
	w, profils := champDeCartes(t, monteeSimple())
	semer(t, w, profils, 1)
	w.Step(Vec{})

	w.Choose(Choices)
	if !w.Choosing() {
		t.Error("un rang hors des places a fermé le choix")
	}
	w.Choose(-1)
	if !w.Choosing() {
		t.Error("un rang négatif a fermé le choix")
	}
}
