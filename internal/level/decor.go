// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La lecture du manifeste de décor, et le catalogue de coûts qu'il en tire.
// Aucun nom de forme n'est écrit dans le code : ajouter une flaque au générateur
// suffit à ce que le champ de flux la contourne.

package level

import (
	"fmt"
	"io/fs"
	"math"
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
	//
	// Le rapport est de deux pour un, et le chargeur l'exige : les formes sont
	// dessinées en 2:1 par le générateur, si bien qu'un manifeste annonçant
	// autre chose décrirait des images que le dossier ne contient pas. C'est
	// d'ici que le rendu tient sa projection, et de nulle part ailleurs.
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
	// **Le rendu et le chargeur la lisent, et c'est ce qui les accorde** : elle
	// dit où poser l'image, et quelles cases la forme ferme. Les deux passent
	// par `Block`, si bien qu'aucun pixel dessiné ne tombe sur une case qu'on
	// peut traverser. Ce qu'elle couvre est autre chose, et se mesure :
	// `Covering`.
	Footprint [2]float64 `json:"emprise"`
	// Covering dit que la forme peint tout le losange de sa case, donc qu'il n'y
	// a rien à combler dessous.
	//
	// **C'est la question que le sol d'un thème existe pour résoudre.** Elle se
	// déduisait de l'emprise, ce qui était vrai tant qu'un volume peignait tout
	// son dessus : une emprise pleine valait un losange plein. Un marquage au
	// sol les sépare — il occupe sa case, donc son emprise en vaut une, et n'en
	// cache rien. Le générateur la mesure désormais sur les pixels.
	//
	// Le champ absent vaut `false`, et c'est le sens sûr : croire qu'une forme
	// couvre laisse un trou à l'écran, croire l'inverse coûte un blit que rien
	// ne voit.
	Covering bool `json:"couvrant"`
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

// LoadDecor lit le manifeste de décor et refuse une forme dont le rôle se
// contredit.
//
// C'est ce qui fait du manifeste le contrat qu'il prétend être : aucun nom de
// forme n'est écrit dans le code, et ajouter une flaque au générateur suffit à
// ce que le champ de flux la contourne.
//
// La taille de tuile entre dans les mêmes manquements que les formes. La clé
// absente vaut un couple nul, qu'aucun consommateur ne saurait distinguer d'un
// réglage, et un refus qui s'arrêterait à elle cacherait ce que le fichier a
// d'autre à corriger.
func LoadDecor(fsys fs.FS, chemin string) (*Decor, error) {
	decor, err := manifest.Decode[Decor](fsys, chemin)
	if err != nil {
		return nil, err
	}
	if decor.Format != FormatDecor {
		return nil, fmt.Errorf("%s: %w : %d, ce binaire lit la %d",
			chemin, manifest.ErrUnsupportedFormat, decor.Format, FormatDecor)
	}

	var manques []string
	if largeur, hauteur := decor.Tile[0], decor.Tile[1]; largeur <= 0 || hauteur*2 != largeur {
		manques = append(manques, fmt.Sprintf(
			"tuile : [%d, %d], il en faut une largeur positive et une hauteur de moitie",
			largeur, hauteur))
	}

	for _, nom := range noms(decor.Shapes) {
		if _, defaut := decor.Shapes[nom].cout(); defaut != "" {
			manques = append(manques, nom+" : "+defaut)
		}
	}
	if len(manques) > 0 {
		return nil, &manifest.Invalid{Path: chemin, Missing: manques}
	}
	return decor, nil
}

// Footing est ce qu'une forme fait à la grille : ce que sa traversée coûte, et
// sur combien de cases ce coût s'applique.
//
// **Les deux voyagent ensemble parce qu'ils ne veulent rien dire l'un sans
// l'autre.** Un coût sans son bloc s'appliquerait à la seule case d'ancrage, ce
// qui laisserait une gondole de deux tuiles n'en fermer qu'une ; un bloc sans
// son coût ne dirait pas ce qu'il faut y écrire. Deux tables parallèles auraient
// fini par ne plus parler des mêmes formes.
type Footing struct {
	// Cost est le prix de traversée, `game.Blocked` pour ce qui arrête.
	Cost game.Cost
	// Block est le nombre de cases fermées sur chaque axe, au moins une.
	Block [2]int
}

// Costs dérive le catalogue d'assises que le chargeur de lieux consulte.
//
// **Il se dérive au lieu d'être rendu par la lecture**, parce que le manifeste
// est désormais lu pour deux choses — les assises et les images — et qu'un couple
// rendu à tous ferait porter à chaque appelant ce dont il n'a pas l'usage. Une
// forme dont le rôle se contredit vaut un mur ici ; c'est `LoadDecor` qui la
// refuse, et personne n'atteint ce cas sur un manifeste qu'il a lu.
func (d *Decor) Costs() map[string]Footing {
	assises := make(map[string]Footing, len(d.Shapes))
	for nom, forme := range d.Shapes {
		cout, defaut := forme.cout()
		if defaut != "" {
			cout = game.Blocked
		}
		assises[nom] = Footing{Cost: cout, Block: forme.Block()}
	}
	return assises
}

// Block rend le nombre de cases qu'une forme ferme sur chaque axe.
//
// **Le plafond de l'emprise, et jamais un seuil de recouvrement.** L'emprise se
// centre sur sa case : une forme de deux tuiles recouvre donc exactement la
// moitié de chacune de ses voisines, et un seuil « plus de la moitié » lui
// donnerait une case ou neuf selon qu'on écrive `>` ou `>=`. Huit formes du
// catalogue ont une emprise de deux exactement, le cas n'a rien d'exotique. Le
// plafond ne compare aucune fraction, et à l'entier ses deux branches rendent le
// même bloc — la borne est tranchée par la géométrie au lieu de l'être par une
// écriture.
//
// **Le rendu et la passabilité l'appellent tous les deux**, et c'est ce qui les
// tient d'accord : le dessin se centre sur le bloc que ce compte donne, si bien
// qu'aucun pixel ne tombe sur une case franchissable. Recopier la règle d'un
// côté ou de l'autre la ferait diverger en silence, la scène restant plausible.
//
// Une emprise nulle ou négative est refusée à la lecture des images ; le
// plancher à une case est ici pour que ce qui échapperait à ce refus ferme au
// moins sa propre case.
func (s Shape) Block() [2]int {
	return [2]int{
		max(int(math.Ceil(s.Footprint[0])), 1),
		max(int(math.Ceil(s.Footprint[1])), 1),
	}
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
