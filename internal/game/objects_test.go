// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce que le décodage du catalogue d'objets garde : le fichier livré se lit en
// entier, une version de format inconnue se refuse, et une passabilité qui se
// contredit aussi.

package game

import (
	"errors"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/sprimault/cohue"
	"github.com/sprimault/cohue/internal/manifest"
)

// manifesteObjets est le catalogue des objets, tel que le binaire l'embarque.
const manifesteObjets = "assets/objets/manifeste.json"

// TestLeCatalogueDobjetsLivreSeDecode monte tout `assets/objets` sans rien
// injecter.
//
// **Le décodage refuse toute clé inconnue**, si bien qu'un champ ajouté au
// générateur sans l'être ici fait tomber ce test plutôt que de disparaître en
// silence. C'est ce qui donne son prix à un cas qui, autrement, ne ferait que
// relire un fichier.
//
// Il compte ce qu'il a visité, et il épingle les valeurs de jeu qui n'ont pas
// encore de lecteur : ce sont elles que ce paquet a été chargé de tenir, et une
// boucle qui ne trouverait aucune entrée passerait au vert en ne vérifiant rien.
func TestLeCatalogueDobjetsLivreSeDecode(t *testing.T) {
	catalogue, err := LoadObjects(cohue.Assets, manifesteObjets)
	if err != nil {
		t.Fatalf("catalogue livré : %v", err)
	}

	familles := map[string]int{}
	for _, objet := range catalogue.Items {
		familles[objet.Family]++
	}
	for _, famille := range []string{FamilyWorld, FamilyWeapon, FamilyEffect,
		FamilyParticle, FamilyCycle} {
		if familles[famille] == 0 {
			t.Errorf("aucune entrée de la famille « %s »", famille)
		}
	}

	// La caisse porte tout ce que l'étape 7 lui demandera, et le vérifier ici
	// dit que le décodage traverse la structure imbriquée — un `destruction`
	// laissé nul se lirait comme une caisse incassable.
	caisse, connue := catalogue.Items["caisse"]
	if !connue {
		t.Fatal("« caisse » n'est pas au catalogue")
	}
	if caisse.Destruction == nil {
		t.Fatal("la caisse ne déclare pas comment elle se casse")
	}
	if mode := caisse.Destruction.Mode; mode != "contact" {
		t.Errorf("mode de destruction « %s », attendu « contact »", mode)
	}
	if delai := caisse.Destruction.DelayMs; delai != 330 {
		t.Errorf("délai de %d ms, attendu 330 — le tiers de seconde du chapitre 7", delai)
	}

	// La fiole porte son soin et ses emplacements, que le chapitre 5 chiffre
	// ensemble avec la vie et le plafond de dégâts.
	fiole, connue := catalogue.Items["fiole"]
	if !connue {
		t.Fatal("« fiole » n'est pas au catalogue")
	}
	if fiole.Heal != 30 || fiole.Slots != 2 {
		t.Errorf("fiole : %d de soin sur %d emplacement(s), attendu 30 sur 2",
			fiole.Heal, fiole.Slots)
	}

	if len(catalogue.Items) < 25 {
		t.Fatalf("%d entrées au catalogue, trop peu pour dire quoi que ce soit",
			len(catalogue.Items))
	}
	t.Logf("%d entrées, %d familles", len(catalogue.Items), len(familles))
}

// TestUnFormatDobjetsInconnuEstRefuse garde la borne que tous les manifestes
// portent.
//
// Un fichier d'une version que ce binaire ne lit pas se refuse au chargement et
// non à l'usage : décoder ce qu'on ne comprend pas rendrait des valeurs
// plausibles et fausses.
func TestUnFormatDobjetsInconnuEstRefuse(t *testing.T) {
	fsys := fstest.MapFS{"objets.json": &fstest.MapFile{Data: []byte(`{
		"version_format": 99,
		"objets": {}
	}`)}}

	_, err := LoadObjects(fsys, "objets.json")
	if !errors.Is(err, manifest.ErrUnsupportedFormat) {
		t.Fatalf("erreur %v, attendu %v", err, manifest.ErrUnsupportedFormat)
	}
}

// catalogueForge rend un manifeste d'un seul objet, écrit à la main.
//
// Bâti plutôt que livré, parce que ce qu'on éprouve ici est un refus : le
// catalogue livré ne porte aucune des paires que le chargeur doit rejeter, et
// c'est heureux.
func catalogueForge(entree string) fstest.MapFS {
	return fstest.MapFS{"objets.json": &fstest.MapFile{Data: []byte(`{
		"version_format": 1,
		"objets": {` + entree + `}
	}`)}}
}

// TestUnePassabiliteQuiSeContreditEstRefusee garde le bout que ce paquet tient
// d'un contrôle à deux bouts.
//
// **L'autre est dans `outils/objets.py`, et ils partent ensemble.** Le
// générateur refuse d'écrire ces couples, ce chargeur refuse de les lire : ils
// ont le même déclencheur — quelqu'un qui lit la valeur —, et chacun sans
// l'autre est une moitié dont on ne peut plus voir à quoi elle sert.
func TestUnePassabiliteQuiSeContreditEstRefusee(t *testing.T) {
	cas := []struct {
		nom    string
		entree string
		attend string
	}{
		{
			"bloquant avec un coût",
			`"muret": {"famille": "monde", "bloquant": true, "cout_traversee": 2}`,
			"bloquant et pourtant",
		},
		{
			"cassé au contact sans coût",
			`"caisse": {"famille": "monde", "bloquant": false,
			 "destruction": {"mode": "contact", "delai_ms": 330,
			  "ruine": "epave", "eclats": "bois"}}`,
			"sans cout_traversee",
		},
		{
			"coût nul",
			`"flaque": {"famille": "monde", "bloquant": false, "cout_traversee": 0}`,
			"attendu entre",
		},
	}
	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			_, err := LoadObjects(catalogueForge(c.entree), "objets.json")
			if err == nil {
				t.Fatal("le catalogue est accepté")
			}
			if !strings.Contains(err.Error(), c.attend) {
				t.Errorf("le refus ne dit pas %q : %v", c.attend, err)
			}
		})
	}
}

// TestUnObjetSansCoutNiBlocagePasse garde ce que le refus ci-dessus ne doit pas
// emporter.
//
// **La moitié qui vaut pour le décor ne vaut pas ici**, et c'est ce cas qui le
// dit : là-bas tout ce qui se franchit doit déclarer son coût, parce que toute
// forme est une case. Un projectile, un éclat, une icône ne sont sur aucune
// grille, et leur en réclamer un leur inventerait une passabilité.
func TestUnObjetSansCoutNiBlocagePasse(t *testing.T) {
	entree := `"projectile_base": {"famille": "monde", "bloquant": false}`
	if _, err := LoadObjects(catalogueForge(entree), "objets.json"); err != nil {
		t.Errorf("un objet qui n'est sur aucune grille est refusé : %v", err)
	}
}

// TestLaCaisseDuCatalogueSeResout garde le renvoi qui relie les deux manifestes.
//
// Le nom vient de la progression, les valeurs du catalogue : c'est ce qui évite
// à la simulation de porter le premier nom d'asset de son histoire. Le lien se
// casserait en silence à un renommage, d'où un refus qui nomme ce qu'il n'a pas
// trouvé.
func TestLaCaisseDuCatalogueSeResout(t *testing.T) {
	catalogue, err := LoadObjects(cohue.Assets, manifesteObjets)
	if err != nil {
		t.Fatalf("catalogue livré : %v", err)
	}

	regles, err := catalogue.Crate(progressionLivree(t).CrateObject)
	if err != nil {
		t.Fatalf("caisse du catalogue livré : %v", err)
	}
	// Le tiers de seconde du chapitre 7, converti une fois au chargement.
	if attendu, _ := TicksFromMs(330); regles.Press != attendu {
		t.Errorf("appui de %d ticks, attendu %d", regles.Press, attendu)
	}
	if regles.Cost <= Free {
		t.Errorf("coût de traversée de %d : une caisse qui ne ralentit pas ne "+
			"coute rien a ramasser", regles.Cost)
	}

	if _, err := catalogue.Crate("caiise"); err == nil {
		t.Error("un nom de caisse introuvable se résout quand même")
	}
}
