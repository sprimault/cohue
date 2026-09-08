// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce que le décodage du catalogue d'objets garde : le fichier livré se lit en
// entier, et une version de format que ce binaire ne connaît pas se refuse.

package game

import (
	"errors"
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
