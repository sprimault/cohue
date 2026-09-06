// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le catalogue des objets : ce que le manifeste déclare, et les images qu'on en
// tire. Ramassables, projectiles, effets et particules y voisinent, chacun avec
// ce que sa famille exige.

package sprite

import (
	"fmt"
	"image"
	"io/fs"
	"maps"
	"path"
	"slices"

	"github.com/sprimault/cohue/internal/game"
	"github.com/sprimault/cohue/internal/manifest"
)

// FormatObjects est la version du manifeste d'objets que ce binaire lit.
const FormatObjects = 1

// Les familles que le manifeste distingue, et qui décident de ce qu'une entrée
// porte.
const (
	// familleMonde est un objet posé dans la scène : une taille, un ancrage, une
	// emprise. C'est la seule famille que ce paquet pose aujourd'hui.
	familleMonde = "monde"
	// familleEffet est une animation brève et centrée sur ce qu'elle marque :
	// l'étincelle d'un impact, le souffle d'une déflagration.
	familleEffet = "effet"
)

// Objects est le manifeste que `outils/objets.py` écrit.
//
// **Il porte des valeurs de jeu que ce paquet ne lit pas** — le soin d'une
// fiole, les charges d'une arme lourde, les touches d'un destructible. Elles
// sont déclarées parce que le décodage refuse toute clé inconnue, et c'est le
// même arrangement que celui de `game.Figure` pris par l'autre bout : là-bas la
// simulation porte le dessin sans le lire, ici le dessin porte le jeu sans le
// lire.
//
// **Le jour où l'étape 6 ou 7 les consommera, ce décodeur déménage** plutôt
// qu'il ne se dédouble : deux lectures du même fichier seraient deux vérités sur
// lui, avec leurs tolérances propres.
type Objects struct {
	manifest.Commentable
	// Format est la version du format de manifeste.
	Format int `json:"version_format"`
	// Items sont les objets, par nom.
	Items map[string]Item `json:"objets"`
}

// Item est une entrée du catalogue.
//
// Un seul type pour les quatre familles, et non un par famille : le fichier est
// une table dont les valeurs doivent se décoder d'une seule passe, et une passe
// par famille demanderait de savoir laquelle avant d'avoir lu la clé qui le dit.
// Ce qu'une famille n'emploie pas reste à zéro, et `controlerObjet` refuse une
// entrée dont la famille et les champs ne s'accordent pas.
type Item struct {
	manifest.Commentable
	// Family range l'entrée, et décide de ce qu'elle doit porter.
	Family string `json:"famille"`
	// Blocking dit si l'objet arrête ce qui s'y présente. Rien ne le lit encore :
	// une caisse ne bloque pas le champ de flux avant l'étape 7.
	Blocking bool `json:"bloquant"`

	// Ce que porte un objet du monde.
	Size      [2]int     `json:"taille,omitempty"`
	Anchor    [2]int     `json:"ancrage,omitempty"`
	Footprint [2]float64 `json:"emprise,omitempty"`
	Elevation int        `json:"elevation,omitempty"`
	Category  string     `json:"categorie,omitempty"`
	Masking   bool       `json:"masquant,omitempty"`
	// Twinkle est la boucle qui dit qu'un objet se ramasse, absente sinon.
	Twinkle *Twinkle `json:"scintillement,omitempty"`

	// Ce que porte un effet ou une particule : une bande d'images de côté fixe,
	// ou un jeu de formes tirées au hasard.
	Frames   int    `json:"images,omitempty"`
	Shapes   int    `json:"formes,omitempty"`
	Cell     [2]int `json:"cote,omitempty"`
	Duration int    `json:"duree_ms,omitempty"`
	Loop     bool   `json:"boucle,omitempty"`

	// Ce que porte une arme lourde, et que rien ne lit avant l'étape 6.
	IconSize   [2]int `json:"taille_icone,omitempty"`
	GroundSize [2]int `json:"taille_sol,omitempty"`
	GroundAt   [2]int `json:"ancrage_sol,omitempty"`
	Charges    int    `json:"charges,omitempty"`

	// Les valeurs de jeu qui restent, et que rien ne lit encore.
	Heal        int          `json:"soin,omitempty"`
	Slots       int          `json:"emplacements,omitempty"`
	Sound       string       `json:"son,omitempty"`
	SoundFamily string       `json:"famille_sons,omitempty"`
	Destruction *Destruction `json:"destruction,omitempty"`
}

// Twinkle est la boucle de scintillement d'un ramassable.
type Twinkle struct {
	manifest.Commentable
	// Frames est le nombre d'images de la bande.
	Frames int `json:"images"`
	// Amplitude est le bombement vertical, en pixels.
	//
	// **C'est la seule chose que la bande ne dit pas.** Sa cellule fait la
	// largeur de l'objet sur sa hauteur plus ce nombre, et son point d'appui
	// descend d'autant. Sans lui, une bande de quarante sur dix pour un objet de
	// dix sur huit ne se découpe qu'en devinant — exactement ce que le
	// manifeste-contrat existe pour éviter.
	Amplitude int `json:"amplitude"`
	// Duration est la durée d'une image, en millisecondes.
	Duration int `json:"duree_ms"`
	// Loop dit si la boucle reprend. Un scintillement boucle toujours ; le champ
	// existe parce que le générateur l'écrit comme pour les autres cycles.
	Loop bool `json:"boucle"`
}

// Destruction est ce qu'un objet cassable déclare, et que rien ne lit avant
// l'étape 7.
type Destruction struct {
	manifest.Commentable
	Mode       string `json:"mode"`
	DelayMs    int    `json:"delai_ms,omitempty"`
	Hits       int    `json:"touches,omitempty"`
	Ruin       string `json:"ruine"`
	Shards     string `json:"eclats"`
	PressCycle string `json:"cycle_appui,omitempty"`
	BreakCycle string `json:"cycle_rupture,omitempty"`
	PressSound string `json:"son_appui,omitempty"`
	BreakSound string `json:"son_rupture,omitempty"`
}

// Prop est un objet prêt à poser : ses images et le coin où elles vont.
type Prop struct {
	// Images sont les images du cycle, une seule pour un objet immobile.
	Images []image.Image
	// Offset est le coin haut-gauche où poser, relatif au point du monde.
	Offset [2]int
	// Cycle est la cadence de la bande, nulle pour une image unique.
	Cycle game.Cycle
}

// Props porte les objets du catalogue, par nom.
type Props struct {
	objets map[string]Prop
}

// LoadObjects lit le manifeste des objets et découpe ce qu'il déclare.
//
// **Tout le catalogue, comme pour le décor.** Une image absente ou qui dément
// son manifeste est un défaut du dépôt et non de la partie qui l'aurait
// affichée : la découvrir au premier tir remonterait chez un joueur.
//
// Les manquements sont accumulés : qui régénère les objets veut la liste.
func LoadObjects(fsys fs.FS, racine, chemin string) (*Objects, *Props, error) {
	catalogue, err := manifest.Decode[Objects](fsys, chemin)
	if err != nil {
		return nil, nil, err
	}
	if catalogue.Format != FormatObjects {
		return nil, nil, fmt.Errorf("%s: %w : %d, ce binaire lit la %d",
			chemin, manifest.ErrUnsupportedFormat, catalogue.Format, FormatObjects)
	}

	props := &Props{objets: make(map[string]Prop, len(catalogue.Items))}
	var manques []string
	for _, nom := range slices.Sorted(maps.Keys(catalogue.Items)) {
		objet := catalogue.Items[nom]
		if defaut := controlerObjet(nom, objet); defaut != "" {
			manques = append(manques, defaut)
			continue
		}
		if err := props.charger(fsys, racine, nom, objet); err != nil {
			manques = append(manques, err.Error())
		}
	}
	if len(manques) > 0 {
		return nil, nil, &manifest.Invalid{Path: chemin, Missing: manques}
	}
	return catalogue, props, nil
}

// Prop rend un objet du catalogue, ou dit qu'il n'existe pas.
func (p *Props) Prop(nom string) (Prop, bool) {
	objet, connu := p.objets[nom]
	return objet, connu
}

// charger découpe ce qu'une entrée déclare, et rien de plus.
//
// **Les familles que le rendu ne pose pas encore ne se chargent pas.** Trois
// restent dehors, et chacune attend un mécanisme plutôt qu'une décision :
// « arme » attend les armes lourdes de l'étape 6, « particule » attend le bassin
// qui émettra des éclats, et « cycle » — les deux bandes d'une caisse qui cède —
// attend que l'étape 7 donne à la caisse son délai de contact. Les découper
// d'avance serait charger ce que rien n'exerce, dans un paquet dont le seul
// lecteur n'a pas de test.
//
// Un cycle se distingue d'un effet par son ancrage : il appartient à l'objet qui
// le cite et prend le sien, quand un effet se centre sur le point qu'il marque.
// C'est ce qui les sépare le jour où les deux se posent.
func (p *Props) charger(fsys fs.FS, racine, nom string, objet Item) error {
	switch objet.Family {
	case familleMonde:
		img, err := lire(fsys, path.Join(racine, nom+".png"), objet.Size)
		if err != nil {
			return err
		}
		p.objets[nom] = Prop{
			Images: []image.Image{img},
			Offset: [2]int{-objet.Anchor[0], -objet.Anchor[1]},
		}
		if objet.Twinkle != nil {
			return p.chargerScintillement(fsys, racine, nom, objet)
		}
	case familleEffet:
		images, err := decouperLarge(fsys, path.Join(racine, nom+".png"),
			objet.Cell, objet.Frames)
		if err != nil {
			return err
		}
		duree, _ := game.TicksFromMs(objet.Duration)
		p.objets[nom] = Prop{
			// Un effet se centre sur ce qu'il marque : il n'a pas de point
			// d'appui parce qu'il ne se pose pas au sol, il recouvre.
			Images: images,
			Offset: [2]int{-objet.Cell[0] / 2, -objet.Cell[1] / 2},
			Cycle:  game.Cycle{Frames: objet.Frames, Duration: duree, Loop: objet.Loop},
		}
	}
	return nil
}

// chargerScintillement remplace l'image fixe d'un ramassable par sa boucle.
//
// La cellule et son ancrage se dérivent des deux nombres déclarés : la bande
// fait la largeur de l'objet sur sa hauteur plus le bombement, et l'objet y
// flotte, donc son appui descend de ce bombement. Rien n'est deviné du fichier.
func (p *Props) chargerScintillement(fsys fs.FS, racine, nom string, objet Item) error {
	s := objet.Twinkle
	cellule := [2]int{objet.Size[0], objet.Size[1] + s.Amplitude}
	images, err := decouperLarge(fsys, path.Join(racine, nom+"_scintille.png"),
		cellule, s.Frames)
	if err != nil {
		return err
	}
	duree, _ := game.TicksFromMs(s.Duration)
	p.objets[nom] = Prop{
		Images: images,
		Offset: [2]int{-objet.Anchor[0], -(objet.Anchor[1] + s.Amplitude)},
		Cycle:  game.Cycle{Frames: s.Frames, Duration: duree, Loop: s.Loop},
	}
	return nil
}

// controler refuse une entrée dont la famille et les champs se contredisent.
//
// **Le contrôle porte sur ce que ce paquet pose**, et laisse le reste passer :
// juger les charges d'une arme lourde serait juger à la place de l'étape qui les
// lira, et les deux règles finiraient par diverger. C'est la leçon du contrôle
// de dessin, qui avait d'abord été posé dans la simulation.
func controlerObjet(nom string, objet Item) string {
	switch objet.Family {
	case familleMonde:
		if objet.Size[0] < 1 || objet.Size[1] < 1 {
			return fmt.Sprintf("%s : taille %v, une image a deux côtés", nom, objet.Size)
		}
		for i, borne := range objet.Anchor {
			if borne < 0 || borne >= objet.Size[i] {
				return fmt.Sprintf("%s : ancrage %v, hors d'une image de %v",
					nom, objet.Anchor, objet.Size)
			}
		}
		if s := objet.Twinkle; s != nil && (s.Frames < 1 || s.Amplitude < 0) {
			return fmt.Sprintf("%s : scintillement de %d image(s) et %d d'amplitude",
				nom, s.Frames, s.Amplitude)
		}
	case familleEffet:
		if objet.Frames < 1 || objet.Cell[0] < 1 || objet.Cell[1] < 1 {
			return fmt.Sprintf("%s : %d image(s) de %v, un effet a une bande",
				nom, objet.Frames, objet.Cell)
		}
	}
	return ""
}

// decouperLarge tranche une bande en images d'une cellule rectangulaire.
//
// `decouper` ne sait trancher que du carré, ce qu'une feuille de personnage est
// toujours ; une bande d'objet ne l'est presque jamais — dix sur huit pour une
// gemme, quarante-huit sur quarante-huit pour un souffle. Les deux confrontent
// la taille au manifeste et jamais au fichier, ce qui est le point commun qui
// compte.
func decouperLarge(fsys fs.FS, chemin string, cellule [2]int, images int) ([]image.Image, error) {
	bande, err := lire(fsys, chemin, [2]int{cellule[0] * images, cellule[1]})
	if err != nil {
		return nil, err
	}

	tranche, ok := bande.(interface {
		SubImage(r image.Rectangle) image.Image
	})
	if !ok {
		return nil, fmt.Errorf("%s : format sans sous-image, %T", chemin, bande)
	}

	origine := bande.Bounds().Min
	decoupe := make([]image.Image, images)
	for i := range images {
		coin := origine.Add(image.Pt(i*cellule[0], 0))
		decoupe[i] = tranche.SubImage(image.Rectangle{Min: coin, Max: coin.Add(image.Pt(cellule[0], cellule[1]))})
	}
	return decoupe, nil
}
