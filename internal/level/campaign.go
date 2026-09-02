// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La campagne : le dossier qui groupe des lieux et dit par où l'on commence.

package level

import (
	"errors"
	"fmt"
	"io/fs"
	"path"

	"github.com/sprimault/cohue/internal/manifest"
)

// FormatCampaign est la version du format de campagne que ce binaire lit.
const FormatCampaign = 1

// CampaignFile est le nom que porte le descripteur dans un dossier de campagne.
//
// Fixe, pour la raison qui vaut déjà pour `LevelFile` un cran plus bas : un
// dossier se reconnaît sans être ouvert, et renommer une campagne se fait en
// renommant son dossier.
const CampaignFile = "campagne.json"

// ErrNoStart refuse une campagne qui ne dit pas par où commencer.
var ErrNoStart = errors.New("campagne sans lieu de depart")

// Campaign est un dossier de lieux, et l'unité que l'on partage.
//
// **Le dossier existe pour cloisonner l'espace de noms**, exactement comme celui
// d'un lieu le fait pour ses pièces. À plat, deux auteurs nommeront tous les
// deux une salle « parking », et celui qui reçoit les deux en perd une. C'est ce
// qui écarte l'autre forme possible — des lieux à plat et un fichier de graphe
// qui les cite par identifiant : plus économe, mais elle remet le vocabulaire
// global que le dossier existe pour supprimer.
//
// Elle ne porte pas encore de graphe. La conception en fait un graphe de salles
// composé dans l'éditeur, avec des portes et des embranchements ; les portes
// arrivent à l'étape 8, et déclarer leur description avant qu'on la lise en
// ferait un champ que personne ne consulte.
type Campaign struct {
	manifest.Commentable
	// Format est la version du format de campagne, indépendante de celle d'un
	// lieu : une campagne peut gagner son graphe sans que les lieux publiés
	// deviennent suspects.
	Format int `json:"version_format"`
	// ID nomme la campagne, et doit valoir le nom de son dossier.
	ID string `json:"identifiant"`
	// Name est le nom lisible, celui que le catalogue affichera.
	Name string `json:"nom,omitempty"`
	// Start est le dossier du lieu par lequel une run commence.
	Start string `json:"lieu_depart"`
}

// LoadCampaign lit le descripteur que porte un dossier de campagne.
//
// Il ne charge aucun lieu : une campagne dit où commencer, et c'est l'appelant
// qui décide s'il monte cette run maintenant. L'écran de catalogue les
// énumérera toutes sans en cuire une seule.
func LoadCampaign(fsys fs.FS, dossier string) (*Campaign, error) {
	nom := path.Base(dossier)
	if nom == "." || nom == "/" {
		return nil, fmt.Errorf("%q : une campagne se charge par son dossier, qui porte son nom", dossier)
	}

	chemin := path.Join(dossier, CampaignFile)
	campagne, err := manifest.Decode[Campaign](fsys, chemin)
	if err != nil {
		return nil, err
	}
	if campagne.Format != FormatCampaign {
		return nil, fmt.Errorf("%s: %w : %d, ce binaire lit la %d",
			chemin, manifest.ErrUnsupportedFormat, campagne.Format, FormatCampaign)
	}

	var manques []string
	if campagne.ID == "" {
		manques = append(manques, "identifiant : la campagne n'en a pas")
	} else if campagne.ID != nom {
		// Le même piège que pour un lieu : quelqu'un duplique une campagne pour
		// en faire une variante, renomme le dossier et oublie l'identifiant.
		manques = append(manques, fmt.Sprintf(
			"identifiant : « %s », alors que le dossier se nomme « %s »", campagne.ID, nom))
	}
	if campagne.Start == "" {
		manques = append(manques, "lieu_depart : "+ErrNoStart.Error())
	}
	if len(manques) > 0 {
		return nil, &manifest.Invalid{Path: chemin, Missing: manques}
	}
	return campagne, nil
}

// StartPath rend le chemin du lieu de départ, prêt pour le chargeur de lieux.
func (c *Campaign) StartPath(dossier string) string {
	return path.Join(dossier, c.Start)
}
