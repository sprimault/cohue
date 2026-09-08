// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le manifeste des objets : ramassables, projectiles, effets et particules, avec
// ce que chaque famille exige. C'est le seul décodage qui en soit fait.

package game

import (
	"fmt"
	"io/fs"

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

// LoadObjects lit le manifeste des objets et rend le catalogue tel qu'il est
// déclaré.
//
// **Elle ne juge encore rien de ce qu'elle décode**, et ce n'est pas un oubli :
// le contrôle d'une valeur appartient à ce qui la lit, sans quoi deux règles
// finissent par diverger sur le même champ. Le rendu juge donc les dessins de
// son côté, et les valeurs de jeu se jugeront ici quand un mécanisme les
// consommera. C'est la leçon du contrôle de dessin, qui avait d'abord été posé
// dans la simulation.
func LoadObjects(fsys fs.FS, chemin string) (*Objects, error) {
	catalogue, err := manifest.Decode[Objects](fsys, chemin)
	if err != nil {
		return nil, err
	}
	if catalogue.Format != FormatObjects {
		return nil, fmt.Errorf("%s: %w : %d, ce binaire lit la %d",
			chemin, manifest.ErrUnsupportedFormat, catalogue.Format, FormatObjects)
	}
	return catalogue, nil
}
