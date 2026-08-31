// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La lecture du manifeste de décor, et le catalogue de coûts qu'il en tire.
// Aucun nom de forme n'est écrit dans le code : ajouter une flaque au générateur
// suffit à ce que le champ de flux la contourne.

package level

import (
	"fmt"
	"io/fs"
	"sort"

	"github.com/sprimault/cohue/internal/game"
	"github.com/sprimault/cohue/internal/manifest"
)

// FormatDecor est la version du manifeste de décor que ce binaire lit.
const FormatDecor = 1

// Decor est le manifeste que `outils/decor_iso.py` écrit.
//
// Le chargeur n'en tire qu'une chose, le catalogue de coûts ; le rendu y
// puisera taille, ancrage et élévation. Un seul type pour les deux usages,
// parce qu'un second type qui ne décrirait qu'une moitié du fichier laisserait
// la question de savoir lequel fait foi.
type Decor struct {
	manifest.Commentable
	// Format est la version du format de manifeste.
	Format int `json:"version_format"`
	// Tile est la taille d'une tuile en pixels, `[largeur, hauteur]`.
	Tile [2]int `json:"tuile"`
	// Shapes sont les formes du décor, par nom.
	Shapes map[string]Shape `json:"formes"`
}

// Shape est une forme du décor, telle que le générateur la déclare.
type Shape struct {
	manifest.Commentable
	// Theme est le lieu auquel la forme appartient.
	Theme string `json:"theme"`
	// Size est la taille de l'image en pixels.
	Size [2]int `json:"taille"`
	// Anchor est le point d'appui dans l'image, en pixels.
	Anchor [2]int `json:"ancrage"`
	// Elevation est la hauteur au-dessus du sol, en pixels.
	Elevation int `json:"elevation"`
	// Category est la hauteur telle que l'éditeur la lit : `sol`,
	// `obstacle_bas` ou `haut`.
	Category string `json:"categorie"`
	// Footprint est l'emprise au sol en tuiles.
	//
	// **Aucune pièce ne doit poser une forme de plus d'une tuile tant que rien
	// ne lit ce champ**, et c'est pourquoi le lieu livré n'en emploie aucune :
	// le chargeur l'ignore, si bien qu'une gondole de deux tuiles n'en
	// bloquerait qu'une, et que le lieu mentirait sans qu'aucun contrôle ne le
	// dise. Le champ est déclaré parce que le décodage refuse les clés
	// inconnues ; le retirer ferait échouer le chargement du manifeste livré.
	Footprint [2]float64 `json:"emprise"`
	// Blocking dit si la forme arrête ce qui s'y présente.
	Blocking bool `json:"bloquant"`
	// Cost est le prix de la traversée, en pas. Un pointeur, et non un entier
	// dont zéro vaudrait absence : c'est la présence même du champ qui doit
	// s'accorder avec `Blocking`, et un zéro implicite les rendrait
	// indiscernables.
	Cost *int `json:"cout_traversee,omitempty"`
	// Masking dit que la forme dépasse la hauteur d'un personnage, donc qu'elle
	// peut en cacher un.
	//
	// Le champ constate, il ne prescrit pas : ce que le rendu en fait — la
	// silhouette de ce qui est caché, redessinée par-dessus — est sa décision et
	// peut changer sans que le manifeste bouge. Il portait auparavant le nom de
	// la technique, ce qui figeait dans la donnée une solution qu'on n'avait pas
	// encore choisie.
	//
	// Comme `Footprint`, rien ne le lit encore. Le renommer n'a donc engagé
	// aucune garantie du compilateur : c'est `DisallowUnknownFields`, au
	// chargement du manifeste livré, qui a confronté les deux noms.
	Masking bool `json:"masquant"`
}

// LoadDecor lit le manifeste de décor et en dérive le catalogue de coûts.
//
// C'est ce qui fait du manifeste le contrat qu'il prétend être : aucun nom de
// forme n'est écrit dans le code, et ajouter une flaque au générateur suffit à
// ce que le champ de flux la contourne.
func LoadDecor(fsys fs.FS, chemin string) (*Decor, map[string]game.Cost, error) {
	decor, err := manifest.Decode[Decor](fsys, chemin)
	if err != nil {
		return nil, nil, err
	}
	if decor.Format != FormatDecor {
		return nil, nil, fmt.Errorf("%s: %w : %d, ce binaire lit la %d",
			chemin, manifest.ErrUnsupportedFormat, decor.Format, FormatDecor)
	}

	couts := make(map[string]game.Cost, len(decor.Shapes))
	var manques []string
	for _, nom := range noms(decor.Shapes) {
		forme := decor.Shapes[nom]
		cout, defaut := forme.cout()
		if defaut != "" {
			manques = append(manques, nom+" : "+defaut)
			continue
		}
		couts[nom] = cout
	}
	if len(manques) > 0 {
		return nil, nil, &manifest.Invalide{Chemin: chemin, Manques: manques}
	}
	return decor, couts, nil
}

// cout rend le prix de traversée de la forme, ou ce qui l'empêche de l'avoir.
//
// Le contrôle joue dans les deux sens, comme celui de `ressources.py` : un mur
// qui porterait un coût le porterait sans que rien ne le lise, et l'auteur
// croirait avoir réglé quelque chose. Les deux contrôles ne se doublent pas —
// le générateur vérifie ce qu'il écrit, le chargeur ce qu'il lit, et un lieu
// tiers n'aura jamais traversé le premier.
func (s Shape) cout() (game.Cost, string) {
	switch {
	case s.Blocking && s.Cost != nil:
		return 0, fmt.Sprintf("bloquant et pourtant un cout_traversee de %d", *s.Cost)
	case s.Blocking:
		return game.Blocked, ""
	case s.Cost == nil:
		return 0, "franchissable sans cout_traversee"
	case *s.Cost < int(game.Free) || *s.Cost >= int(game.Blocked):
		return 0, fmt.Sprintf("cout_traversee de %d, attendu entre %d et %d",
			*s.Cost, game.Free, game.Blocked-1)
	}
	return game.Cost(*s.Cost), ""
}

// noms rend les clés triées d'une table de formes.
//
// Le parcours d'une map n'a pas d'ordre stable : sans tri, deux chargements du
// même fichier invalide énuméreraient les manquements dans un ordre différent,
// et le message deviendrait impossible à comparer d'un essai à l'autre.
func noms(formes map[string]Shape) []string {
	tries := make([]string, 0, len(formes))
	for nom := range formes {
		tries = append(tries, nom)
	}
	sort.Strings(tries)
	return tries
}
