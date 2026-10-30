// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les icônes de la fenêtre : ce que le manifeste en déclare, et leur lecture en
// planches. Rien n'est dessiné ici — le gestionnaire de fenêtres choisit la
// taille qui convient à l'endroit où il l'affiche.

package ui

import (
	"fmt"
	"io/fs"
	"path"

	"github.com/sprimault/lozengine/raster"

	"github.com/sprimault/cohue/internal/manifest"
)

// LoadIcons lit les icônes de fenêtre que le manifeste d'interface déclare.
//
// **Le manifeste ne dit que des noms de fichiers.** La taille de chacun se lit
// dans son image, et la déclarer serait une seconde description du dessin, qui
// se démentirait le jour où le générateur change de grille sans que le manifeste
// suive.
//
// **Une planche par taille, et aucune n'est redimensionnée.** C'est ce que la
// fenêtre attend : le moteur ne sait pas passer de vingt-quatre à seize pixels
// par une échelle entière, et une réduction douce brouillerait un dessin fait
// pixel par pixel. Le système retient la plus proche et réduit s'il y tient.
//
// Une liste vide fait échouer plutôt que de rendre une tranche nulle : une
// fenêtre sans icône est une icône par défaut, c'est-à-dire un défaut qu'on ne
// remarque pas — et le seul endroit capable de le dire est celui-ci.
func LoadIcons(fsys fs.FS, chemin string) ([]raster.Sheet, error) {
	brut, err := manifest.Decode[rawInterface](fsys, chemin)
	if err != nil {
		return nil, err
	}
	if brut.Format != FormatFont {
		return nil, fmt.Errorf("%s: %w : %d, ce binaire lit la %d",
			chemin, manifest.ErrUnsupportedFormat, brut.Format, FormatFont)
	}

	fichiers := brut.Interface.Icon.Files
	if len(fichiers) == 0 {
		return nil, &manifest.Invalid{Path: chemin,
			Missing: []string{"icone : aucun fichier déclaré"}}
	}

	dossier := path.Dir(chemin)
	icones := make([]raster.Sheet, 0, len(fichiers))
	for _, nom := range fichiers {
		f, err := fsys.Open(path.Join(dossier, nom))
		if err != nil {
			return nil, fmt.Errorf("icone %s: %w", nom, err)
		}
		icone, err := raster.LoadSheet(f)
		// Fermé ici et non en `defer` : ce qui suit ne lit plus le fichier, et
		// une lecture seule n'a rien à rendre à la fermeture.
		_ = f.Close()
		if err != nil {
			return nil, fmt.Errorf("icone %s: %w", nom, err)
		}
		icones = append(icones, icone)
	}
	return icones, nil
}

// rawIcon porte ce que le manifeste dit des icônes.
type rawIcon struct {
	manifest.Commentable

	Files []string `json:"fichiers"`
}
