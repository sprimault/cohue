// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas du choix : les trois places toujours remplies, ce qu'une carte change,
// l'axe épuisé qui sort du menu, et les montées qui s'accumulent sans se perdre.

package game

import (
	"fmt"
	"slices"
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

	w := NewWorld(profils, armes, seuils, sansVagues(), NewCostGrid(32, 32), graineDeTest, capacitesDeTest)
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

// TestLeMenuEstLesTroisAxesSansSoupape écrit ce que la table offre désormais.
//
// **La soupape a quitté le menu ordinaire, et c'est le vrai gain du troisième
// axe.** Trois axes remplissent exactement les trois places, si bien qu'elle ne
// reparaît qu'à l'épuisement de l'un d'eux — ce que la conception attend d'elle,
// et ce que deux axes lui interdisaient : elle occupait alors une place sur trois
// du début à la fin d'une run.
//
// **Aucun tirage n'a lieu pour autant**, et la version précédente de ce test
// annonçait le contraire pour ce jour-ci. Le pool ne dépasse les trois places
// qu'au quatrième axe, et c'est alors qu'un flux aléatoire prendra son numéro. Ce
// test tombera une seconde fois ce jour-là, ce qui est son second emploi.
func TestLeMenuEstLesTroisAxesSansSoupape(t *testing.T) {
	w, profils := champDeCartes(t, monteeSimple())
	semer(t, w, profils, 1)
	w.Step(Vec{})

	cartes := w.Pending()
	if cartes[0].Name != "Cadence" || cartes[1].Name != "Portée" ||
		cartes[2].Name != "Projectiles" {
		t.Errorf("les trois places : %q, %q et %q",
			cartes[0].Name, cartes[1].Name, cartes[2].Name)
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

// tableLarge bâtit une table de passifs plus large que le nombre de places.
//
// **Elle est bâtie en Go et non décodée d'un manifeste, et c'est une nécessité
// plutôt qu'une entorse** : la liste des axes admis est close dans
// `passive.go`, si bien qu'aucun fichier ne peut en déclarer un quatrième tant
// que le code n'en connaît pas un quatrième. Ce qu'on isole ici est la
// sélection, jamais le décodage — celui-ci reste gardé par les tests qui
// chargent le manifeste livré.
//
// Les axes reprennent les clés existantes : ce que le tirage manipule est une
// place dans la tranche, et leurs effets ne sont pas ce qu'on mesure.
func tableLarge(t *testing.T) *Weapons {
	t.Helper()
	armes, err := LoadWeapons(cohue.Assets, manifesteArmes)
	if err != nil {
		t.Fatalf("armes livrées : %v", err)
	}

	large := *armes.Passives
	large.Axes = make([]Passive, 0, 5)
	for _, nom := range []string{"A", "B", "C", "D", "E"} {
		large.Axes = append(large.Axes, Passive{
			Axis: AxisCadence, Name: nom, Phrase: "Essai.", Tiers: 6,
			Effects: []string{"1", "2", "3", "4", "5", "6"},
		})
	}
	copie := *armes
	copie.Passives = &large
	return &copie
}

// offresDe joue une montée sur la table donnée et rend les noms offerts.
func offresDe(t *testing.T, armes *Weapons, graine uint64) []string {
	t.Helper()
	profils, err := LoadProfiles(cohue.Assets, manifestePersonnages)
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}
	w := NewWorld(profils, armes, monteeSimple(), sansVagues(), NewCostGrid(32, 32),
		graine, capacitesDeTest)
	w.Place(FromInt(16)+One/2, FromInt(16)+One/2)

	semer(t, w, profils, 1)
	w.Step(Vec{})

	noms := make([]string, 0, Choices)
	for _, c := range w.Pending() {
		noms = append(noms, c.Name)
	}
	return noms
}

// TestLeTirageNeSeConsommePasSansChoix garde ce que la table livrée ne fait pas.
//
// **Trois axes pour trois places ne laissent rien à choisir**, et le flux ne doit
// alors pas être touché : un tirage inconditionnel le décalerait sans qu'aucune
// décision en dépende, et l'attendu d'empreinte bougerait pour une raison qui
// n'est pas une règle de jeu.
//
// C'est aussi ce qui donne son statut au lot qui a introduit `Cards` : le
// mécanisme est en place, et aucune donnée livrée ne le consomme encore.
func TestLeTirageNeSeConsommePasSansChoix(t *testing.T) {
	w, profils := champDeCartes(t, monteeSimple())
	semer(t, w, profils, 1)
	w.Step(Vec{})

	if len(w.Pending()) != Choices {
		t.Fatalf("%d carte(s) offertes", len(w.Pending()))
	}
	// Un flux neuf de la même graine : si l'ouverture avait tiré, celui de la
	// partie aurait pris de l'avance et les deux rendraient des valeurs
	// différentes.
	neuf := NewStreams(graineDeTest).Cards
	if attendu, obtenu := neuf.IntN(1<<30), w.hasard.Cards.IntN(1<<30); attendu != obtenu {
		t.Errorf("le flux des cartes a été consommé : %d, attendu %d", obtenu, attendu)
	}
}

// TestPlusDAxesQueDePlacesFaitTirer garde le mécanisme que le quatrième axe
// réveillera.
//
// Deux propriétés, et la seconde est celle qui compte : trois places sur cinq
// axes, et deux graines qui n'offrent pas la même chose. Sans elle, un tirage
// qui rendrait toujours les trois premiers passerait la première.
func TestPlusDAxesQueDePlacesFaitTirer(t *testing.T) {
	armes := tableLarge(t)

	offres := offresDe(t, armes, graineDeTest)
	if len(offres) != Choices {
		t.Fatalf("%d carte(s) offertes, attendu %d", len(offres), Choices)
	}
	vus := map[string]bool{}
	for _, nom := range offres {
		if vus[nom] {
			t.Errorf("« %s » offert deux fois : %v", nom, offres)
		}
		vus[nom] = true
	}

	// Les graines sont choisies distinctes ; deux tirages de trois parmi cinq
	// peuvent coïncider, donc on en compare plusieurs plutôt qu'une paire.
	distinctes := map[string]bool{}
	for _, graine := range []uint64{1, 2, 3, 4, 5} {
		distinctes[fmt.Sprint(offresDe(t, armes, graine))] = true
	}
	if len(distinctes) < 2 {
		t.Errorf("cinq graines offrent toutes la même chose : %v", distinctes)
	}
}

// TestChoisirAppliqueLePalierDeProjectiles ferme le chemin de la carte à l'arme.
//
// Les tests de géométrie de `shooting_test.go` posent le nombre sur l'arme pour
// isoler la forme du front ; celui-ci garde l'autre moitié — qu'une carte prise
// l'accroisse réellement. Sans lui, l'axe pourrait n'être qu'une entrée de table
// que rien ne branche, et les deux autres tests passeraient quand même.
func TestChoisirAppliqueLePalierDeProjectiles(t *testing.T) {
	w, profils := champDeCartes(t, monteeSimple())
	avant := w.arme.Projectiles

	semer(t, w, profils, 1)
	w.Step(Vec{})

	rang := slices.IndexFunc(w.Pending(), func(c Card) bool { return c.Name == "Projectiles" })
	if rang < 0 {
		t.Fatal("la carte des projectiles n'est pas offerte")
	}
	w.Choose(rang)

	if w.arme.Projectiles != avant+1 {
		t.Errorf("projectiles : %d, attendu %d", w.arme.Projectiles, avant+1)
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
// Épuiser un axe oblige à basculer sur ceux qu'on n'avait pas choisis : c'est ce
// que la conception attend de la borne, et ça ne se voit qu'en prenant six fois
// la même carte.
//
// **C'est aussi le seul chemin qui fasse reparaître la soupape**, depuis que
// trois axes remplissent les trois places : un axe épuisé libère une place, et
// c'est elle qu'elle vient prendre.
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
	if cartes[0].Name != "Portée" || cartes[1].Name != "Projectiles" {
		t.Errorf("les deux axes restants : %q et %q", cartes[0].Name, cartes[1].Name)
	}
	if soupape := w.passifs.Relief.Name; cartes[2].Name != soupape {
		t.Errorf("place libérée : %q, attendu la soupape", cartes[2].Name)
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
