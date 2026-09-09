// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le manifeste des objets : ramassables, projectiles, effets et particules, avec
// ce que chaque famille exige. C'est le seul décodage qui en soit fait.

package game

import (
	"fmt"
	"io/fs"
	"maps"
	"slices"

	"github.com/sprimault/cohue/internal/manifest"
)

// FormatObjects est la version du manifeste d'objets que ce binaire lit.
const FormatObjects = 1

// Les familles que le manifeste distingue, et qui décident de ce qu'une entrée
// porte.
//
// **Exportées parce que deux paquets les lisent** : la simulation y prend les
// valeurs de jeu, le rendu y découpe les images. Deux tables de noms finiraient
// par ne plus lister les mêmes familles, et celle qui se tromperait laisserait
// passer une entrée au lieu de la refuser.
const (
	// FamilyWorld est un objet posé dans la scène : une taille, un ancrage, une
	// emprise.
	FamilyWorld = "monde"
	// FamilyWeapon est une arme lourde, qui porte **deux dessins de natures
	// différentes** : celui du sol suit la projection isométrique et se pose
	// comme un objet du monde, celui de l'icône se voit de face dans un
	// emplacement. En isométrie, une icône de vingt pixels ne se lirait plus.
	FamilyWeapon = "arme"
	// FamilyEffect est une animation brève et centrée sur ce qu'elle marque :
	// l'étincelle d'un impact, le souffle d'une déflagration.
	FamilyEffect = "effet"
	// FamilyParticle est un jeu de formes qu'on tire, et non une bande qu'on
	// déroule : trois éclats par matière, que le rendu envoie sur des
	// trajectoires qu'il calcule. Ce qui la sépare d'un effet est qu'elle n'a pas
	// d'ordre — la deuxième forme ne suit pas la première, elle en diffère.
	FamilyParticle = "particule"
	// FamilyCycle est une bande qui appartient à l'objet qui la cite, et qui
	// prend son ancrage plutôt que de se centrer : les deux cycles d'une caisse
	// qui cède. Rien ne la pose encore.
	FamilyCycle = "cycle"
)

// Objects est le manifeste que `outils/objets.py` écrit.
//
// **Il porte des tailles et des ancrages que ce paquet ne lit pas.** C'est le
// même arrangement que `Figure` pris par l'autre bout : là-bas la simulation
// tient le dessin d'un profil sans jamais l'ouvrir, ici elle tient celui d'un
// objet. Le décodage vit du côté qui décide, et le rendu reçoit le catalogue
// décodé pour n'en découper que les images — deux lectures du même fichier
// seraient deux vérités sur lui, avec leurs tolérances propres.
type Objects struct {
	manifest.Commentable
	// Format est la version du format de manifeste.
	Format int `json:"version_format"`
	// Items sont les objets, par nom.
	Items map[string]Object `json:"objets"`
}

// Object est une entrée du catalogue.
//
// Un seul type pour les familles, et non un par famille : le fichier est une
// table dont les valeurs doivent se décoder d'une seule passe, et une passe par
// famille demanderait de savoir laquelle avant d'avoir lu la clé qui le dit. Ce
// qu'une famille n'emploie pas reste à zéro.
type Object struct {
	manifest.Commentable
	// Family range l'entrée, et décide de ce qu'elle doit porter.
	Family string `json:"famille"`
	// Blocking dit si l'objet arrête ce qui s'y présente. Rien ne le lit encore :
	// les quatre destructibles attendent le lot qui les posera.
	Blocking bool `json:"bloquant"`
	// Cost est le prix de traversée de sa case, en pas.
	//
	// Un pointeur, et non un entier dont zéro vaudrait absence : c'est la
	// présence même du champ qui doit s'accorder avec `Blocking`, comme pour une
	// forme du décor. La plupart des entrées n'en portent aucun, n'étant sur
	// aucune grille — un projectile, un éclat, une icône.
	Cost *int `json:"cout_traversee,omitempty"`

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

	// Ce que porte une arme lourde, et que la table des armes lit à sa place.
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

// ModeContact est le mode de destruction de ce qui cède à l'appui, en le
// traversant : la caisse, et elle seule aujourd'hui.
//
// Le mode de l'obstacle fragile — on s'arrête contre lui et on presse la touche
// d'interaction — n'a pas sa constante : rien ne le lit, et une valeur écrite
// d'avance dans un paquet est une déclaration que personne n'exerce.
const ModeContact = "contact"

// Destruction est ce qu'un objet cassable déclare.
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

// LoadObjects lit le manifeste des objets et refuse une passabilité qui se
// contredit.
//
// **Elle ne juge que ce que ce paquet lit**, le reste appartenant à qui le lira :
// le rendu juge les dessins de son côté, et les charges d'une arme lourde se
// jugent dans la table des armes. C'est la leçon du contrôle de dessin, qui
// avait d'abord été posé dans la simulation.
//
// La passabilité est le premier champ à franchir cette porte, et le coût de
// traversée d'une caisse est ce qui l'y a fait entrer.
func LoadObjects(fsys fs.FS, chemin string) (*Objects, error) {
	catalogue, err := manifest.Decode[Objects](fsys, chemin)
	if err != nil {
		return nil, err
	}
	if catalogue.Format != FormatObjects {
		return nil, fmt.Errorf("%s: %w : %d, ce binaire lit la %d",
			chemin, manifest.ErrUnsupportedFormat, catalogue.Format, FormatObjects)
	}

	var manques []string
	for _, nom := range slices.Sorted(maps.Keys(catalogue.Items)) {
		if _, defaut := catalogue.Items[nom].cout(); defaut != "" {
			manques = append(manques, nom+" : "+defaut)
		}
	}
	if len(manques) > 0 {
		return nil, &manifest.Invalid{Path: chemin, Missing: manques}
	}
	return catalogue, nil
}

// cout rend le prix de traversée d'un objet, ou ce qui l'empêche de l'avoir.
//
// **Le contrôle a deux bouts et ils partent ensemble** : `outils/objets.py`
// refuse d'écrire ces couples, ce chargeur refuse de les lire. Ils ont le même
// déclencheur — quelqu'un qui lit la valeur —, et chacun sans l'autre est une
// moitié dont on ne peut plus voir à quoi elle sert.
//
// **La moitié qui manque au décor est celle qui n'a pas d'objet ici.** Là-bas
// tout ce qui se franchit doit déclarer son coût, parce que toute forme est une
// case ; ici la plupart des entrées ne sont sur aucune grille, et l'exiger
// d'elles leur inventerait une passabilité. Ce qui la remplace est le mode de
// destruction : ce qu'on casse **en le traversant** doit pouvoir se traverser et
// coûter, sans quoi le délai d'appui s'écoule pendant qu'on est déjà de l'autre
// côté.
// Sans coût déclaré elle rend `Free`, qui est ce que vaut la traversée de ce qui
// n'est sur aucune grille : c'est le cas de la plupart des entrées, et la seule
// qui lise ce résultat est une caisse, dont le refus ci-dessus garantit qu'elle
// en porte un.
func (o Object) cout() (Cost, string) {
	switch {
	case o.Blocking && o.Cost != nil:
		return 0, fmt.Sprintf("bloquant et pourtant un cout_traversee de %d", *o.Cost)
	case o.Cost == nil:
		if o.Destruction != nil && o.Destruction.Mode == ModeContact {
			return 0, "se casse en le traversant et se franchit pourtant sans cout_traversee"
		}
		return Free, ""
	case *o.Cost < int(Free) || *o.Cost >= int(Blocked):
		return 0, fmt.Sprintf("cout_traversee de %d, attendu entre %d et %d",
			*o.Cost, Free, Blocked-1)
	}
	return Cost(*o.Cost), "" // #nosec G115 -- borné par la branche précédente
}

// CrateRules est ce que la simulation tient d'une caisse : le temps d'appui
// avant qu'elle cède, et ce que sa case coûte tant qu'elle tient.
//
// **Résolues une fois au montage, jamais cherchées par nom dans un tick.** Le
// catalogue est une table par clé, et l'interroger à chaque image mettrait un
// nom d'asset dans la boucle de mise à jour — avec, en prime, une résolution qui
// peut échouer là où plus rien ne saurait quoi en dire.
type CrateRules struct {
	// Press est le temps d'appui avant rupture, en ticks.
	Press Tick
	// Cost est le prix de traversée de sa case, tant qu'elle tient.
	Cost Cost
}

// Crate résout la caisse du catalogue, désignée par son nom de manifeste.
//
// Le nom vient du manifeste de progression, où il était déjà écrit : la
// simulation ne porte donc aucun nom de catalogue, ce que le manifeste-contrat
// exige d'elle.
func (o *Objects) Crate(nom string) (CrateRules, error) {
	objet, connu := o.Items[nom]
	if !connu {
		return CrateRules{}, fmt.Errorf("objets : la caisse « %s » n'est pas au catalogue", nom)
	}
	if objet.Destruction == nil || objet.Destruction.Mode != ModeContact {
		return CrateRules{}, fmt.Errorf(
			"objets : la caisse « %s » ne se casse pas au contact", nom)
	}
	appui, err := TicksFromMs(objet.Destruction.DelayMs)
	if err != nil {
		return CrateRules{}, fmt.Errorf("objets : caisse « %s » : %w", nom, err)
	}
	// Le coût passe par le même contrôle que le chargement, plutôt que d'être
	// lu directement : les deux rendraient la même valeur aujourd'hui, et deux
	// lectures d'un même champ sont ce qui finit par diverger.
	cout, defaut := objet.cout()
	if defaut != "" {
		return CrateRules{}, fmt.Errorf("objets : caisse « %s » : %s", nom, defaut)
	}
	return CrateRules{Press: appui, Cost: cout}, nil
}
