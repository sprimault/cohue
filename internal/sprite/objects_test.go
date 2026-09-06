// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce que le chargeur d'objets garde : le catalogue livré se lit en entier, la
// cellule d'un scintillement se dérive de son amplitude, et une bande qui dément
// ce couple se refuse.

package sprite

import (
	"fmt"
	"testing"
	"testing/fstest"

	"github.com/sprimault/cohue"
)

// objetsLivres est le manifeste des objets, tel que le binaire l'embarque.
const objetsLivres = "assets/objets/manifeste.json"

// racineObjets est le dossier de leurs images, toutes à plat.
const racineObjets = "assets/objets"

// manifesteObjets bâtit un manifeste d'un seul objet, en JSON brut.
//
// Le JSON plutôt qu'une structure Go : ce qui est éprouvé est le décodage, et
// une structure ne saurait pas dire qu'une clé est absente du fichier.
func manifesteObjets(entree string) fstest.MapFS {
	return fstest.MapFS{"objets.json": &fstest.MapFile{Data: []byte(`{
		"version_format": 1,
		"objets": {` + entree + `}
	}`)}}
}

// ramassable rend le JSON d'un objet de dix sur huit qui scintille.
func ramassable(amplitude int) string {
	return fmt.Sprintf(`"gemme": {
		"taille": [10, 8], "ancrage": [5, 7], "emprise": [0.156, 0.156],
		"elevation": 3, "categorie": "sol", "masquant": false,
		"famille": "monde", "bloquant": false,
		"scintillement": {"images": 4, "amplitude": %d, "duree_ms": 140, "boucle": true}
	}`, amplitude)
}

// TestLeCatalogueDobjetsLivreSeCharge monte tout `assets/objets` sans rien
// injecter.
//
// C'est la conformité que la doctrine exige, et elle porte ici sur un fichier
// que rien ne lisait : le manifeste des objets était écrit depuis l'étape 4 et
// n'avait jamais été décodé, si bien qu'aucun de ses champs n'était confronté à
// quoi que ce soit. Trois écarts sont sortis du premier lecteur — une amplitude
// qui manquait, un `cote` de deux formes selon la famille, une bande rangée dans
// « monde » sans en avoir la forme.
//
// Il compte ce qu'il a visité : une boucle qui ne trouverait aucun objet
// passerait au vert en ne vérifiant rien.
func TestLeCatalogueDobjetsLivreSeCharge(t *testing.T) {
	catalogue, props, err := LoadObjects(cohue.Assets, racineObjets, objetsLivres)
	if err != nil {
		t.Fatalf("catalogue livré : %v", err)
	}

	poses, scintillants := 0, 0
	for nom, objet := range catalogue.Items {
		if objet.Family != familleMonde && objet.Family != familleEffet {
			continue
		}
		p, connu := props.Prop(nom)
		if !connu {
			t.Errorf("« %s » est de la famille « %s » et n'a pas été chargé", nom, objet.Family)
			continue
		}
		poses++
		if len(p.Images) > 1 {
			scintillants++
		}
	}

	if len(catalogue.Items) < 25 || poses < 15 || scintillants < 4 {
		t.Fatalf("%d entrées, %d posables, %d animées : trop peu pour dire quoi que ce soit",
			len(catalogue.Items), poses, scintillants)
	}
	t.Logf("%d entrées au catalogue, %d posables, %d animées",
		len(catalogue.Items), poses, scintillants)
}

// TestLaCelluleDunScintillementSeDeriveDeLamplitude épingle la seule
// arithmétique de ce chargeur.
//
// **L'amplitude est le seul nombre que la bande ne dit pas**, et tout en
// descend : la cellule fait la hauteur de l'objet plus le bombement, et l'appui
// descend d'autant puisque l'objet y flotte. Une erreur d'un pixel décale le
// ramassable de la même quantité, ce qui se voit à peine et ne se mesure nulle
// part ailleurs.
func TestLaCelluleDunScintillementSeDeriveDeLamplitude(t *testing.T) {
	const amplitude = 2
	fsys := manifesteObjets(ramassable(amplitude))
	fsys["gemme.png"] = &fstest.MapFile{Data: bande(t, 10, 8)}
	fsys["gemme_scintille.png"] = &fstest.MapFile{Data: bande(t, 4*10, 8+amplitude)}

	_, props, err := LoadObjects(fsys, ".", "objets.json")
	if err != nil {
		t.Fatalf("chargement : %v", err)
	}
	gemme, connue := props.Prop("gemme")
	if !connue {
		t.Fatal("la gemme n'a pas été chargée")
	}

	if n := len(gemme.Images); n != 4 {
		t.Fatalf("%d image(s) découpée(s), attendu 4", n)
	}
	if taille := gemme.Images[0].Bounds().Size(); taille.X != 10 || taille.Y != 8+amplitude {
		t.Errorf("cellule %v, attendu 10x%d", taille, 8+amplitude)
	}
	// L'appui de l'objet est en (5, 7) ; celui de la cellule descend du
	// bombement, donc le coin remonte d'autant.
	if attendu := [2]int{-5, -(7 + amplitude)}; gemme.Offset != attendu {
		t.Errorf("coin %v, attendu %v", gemme.Offset, attendu)
	}
}

// TestUneBandeDeScintillementQuiMentEstRefusee garde le manifeste-contrat sur le
// couple taille-amplitude.
//
// Une bande dont la hauteur ne s'accorde pas avec ce que le manifeste annonce se
// découperait en images tronquées, et le ramassable se poserait de travers sans
// qu'aucune erreur ne le dise.
func TestUneBandeDeScintillementQuiMentEstRefusee(t *testing.T) {
	fsys := manifesteObjets(ramassable(2))
	fsys["gemme.png"] = &fstest.MapFile{Data: bande(t, 10, 8)}
	// Trois pixels de bombement là où le manifeste en annonce deux.
	fsys["gemme_scintille.png"] = &fstest.MapFile{Data: bande(t, 4*10, 8+3)}

	if _, _, err := LoadObjects(fsys, ".", "objets.json"); err == nil {
		t.Fatal("bande d'un pixel trop haute acceptée")
	}
}

// TestUnAncrageDobjetHorsDeSonImageEstRefuse garde le point qu'on pose sur le
// monde, comme pour une forme du décor.
func TestUnAncrageDobjetHorsDeSonImageEstRefuse(t *testing.T) {
	fsys := manifesteObjets(`"gemme": {
		"taille": [10, 8], "ancrage": [5, 8], "emprise": [0.156, 0.156],
		"elevation": 3, "categorie": "sol", "masquant": false,
		"famille": "monde", "bloquant": false
	}`)
	fsys["gemme.png"] = &fstest.MapFile{Data: bande(t, 10, 8)}

	if _, _, err := LoadObjects(fsys, ".", "objets.json"); err == nil {
		t.Fatal("ancrage d'une rangée hors de l'image accepté")
	}
}
