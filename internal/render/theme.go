// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// L'apparence de l'interface telle que son manifeste la règle : épaisseurs,
// marges et teintes. Aucune dimension calculée — ce qui dépend d'un contenu se
// mesure au rendu.

package render

import (
	"fmt"
	"io/fs"
	"maps"
	"slices"
	"strings"

	"github.com/sprimault/lozengine/render"

	"github.com/sprimault/cohue/internal/manifest"
)

// Teintes que le manifeste doit porter, et rien d'autre.
//
// La table est close, et c'est elle qui décide : une teinte absente fait échouer
// le chargement, une teinte inconnue aussi. Sans le second refus, une clé mal
// orthographiée s'ajouterait au manifeste sans que rien ne la lise, et la
// couleur qu'on croyait avoir réglée resterait celle d'avant.
var teintesRequises = []string{
	"cadre_fond", "cadre_bord", "cadre_choisi", "bandeau_fond",
	"jauge_fond", "jauge_vie", "jauge_experience", "vignette_danger",
	"texte", "texte_attenue", "texte_valeur", "texte_contour",
	"survol",
}

// Theme porte ce que l'interface règle.
//
// Il ne porte aucune largeur ni hauteur d'élément : une carte tient à son texte
// et une case à son icône, si bien qu'une dimension déclarée ici serait une
// seconde description de ce que le contenu impose.
type Theme struct {
	// Border est l'épaisseur d'un bord de cadre, en pixels de tampon.
	Border int
	// Margin est la marge entre un bord et ce qu'il entoure.
	Margin int
	// GaugeHeight est l'épaisseur d'une jauge.
	GaugeHeight int

	teintes map[string]render.Color
}

// LoadTheme lit les réglages d'apparence du manifeste d'interface.
func LoadTheme(fsys fs.FS, chemin string) (*Theme, error) {
	brut, err := manifest.Decode[rawInterface](fsys, chemin)
	if err != nil {
		return nil, err
	}
	if brut.Format != FormatInterface {
		return nil, fmt.Errorf("%s: %w : %d, ce binaire lit la %d",
			chemin, manifest.ErrUnsupportedFormat, brut.Format, FormatInterface)
	}

	var manques []string
	dire := func(format string, args ...any) {
		manques = append(manques, fmt.Sprintf(format, args...))
	}

	r := brut.Interface.Settings
	if r.Border <= 0 {
		dire("reglages.bord_px : %d, un cadre sans bord n'en est pas un", r.Border)
	}
	if r.Margin < 0 {
		dire("reglages.marge_px : %d, une marge ne se retranche pas", r.Margin)
	}
	if r.GaugeHeight <= 0 {
		dire("reglages.hauteur_jauge_px : %d, une jauge sans épaisseur ne se voit pas",
			r.GaugeHeight)
	}

	teintes := make(map[string]render.Color, len(r.Colors))
	for _, nom := range teintesRequises {
		brute, presente := r.Colors[nom]
		if !presente {
			dire("reglages.teintes.%s : absente", nom)
			continue
		}
		// **Une teinte s'écrit prémultipliée par son alpha**, et le manifeste la
		// porte déjà ainsi : elle se recopie donc champ par champ, sans passer
		// par `render.Premultiply`, qui l'appliquerait une seconde fois.
		//
		// **Le refus se délègue à `render.Color.Valid`**, qui porte la même règle
		// et le même argument — une composante au-dessus de l'alpha ne décrit
		// aucune couleur, elle donne un aplat bien plus dense que l'opacité
		// annoncée, et rien d'autre ne le signale. Un voile écrit à plat est passé
		// ainsi, et seule une planche l'a montré. La garder ici en serait une
		// seconde description, que rien ne tiendrait d'accord.
		teinte := render.Color{R: brute[0], G: brute[1], B: brute[2], A: brute[3]}
		if !teinte.Valid() {
			dire("reglages.teintes.%s : %s, une teinte s'écrit prémultipliée",
				nom, debordement(brute))
			continue
		}
		teintes[nom] = teinte
	}
	for _, nom := range slices.Sorted(maps.Keys(r.Colors)) {
		if !slices.Contains(teintesRequises, nom) {
			dire("reglages.teintes.%s : inconnue, que rien ne lira", nom)
		}
	}

	if len(manques) > 0 {
		return nil, &manifest.Invalid{Path: chemin, Missing: manques}
	}
	return &Theme{
		Border:      r.Border,
		Margin:      r.Margin,
		GaugeHeight: r.GaugeHeight,
		teintes:     teintes,
	}, nil
}

// Color rend une teinte du thème.
//
// Elle panique sur un nom inconnu, et c'est délibéré : les noms sont une table
// close vérifiée au chargement, donc un nom absent ici est une faute de frappe
// dans du code, pas une donnée douteuse. Rendre une couleur par défaut la ferait
// passer pour un réglage, et l'on chercherait dans le manifeste ce qui est dans
// le code.
func (t *Theme) Color(nom string) render.Color {
	teinte, connue := t.teintes[nom]
	if !connue {
		panic("teinte d'interface inconnue : " + nom)
	}
	return teinte
}

// debordement nomme ce qui déborde dans un prémultiplié incohérent.
//
// `render.Color.Valid` décide, cette fonction explique : un manquement doit dire
// où il est, et « teinte invalide » enverrait l'auteur du manifeste comparer
// quatre nombres à la main. Les deux ne font pas doublon pour autant — l'une
// porte la règle, l'autre la phrase.
func debordement(brute [4]uint8) string {
	var fautifs []string
	for i, canal := range []string{"rouge", "vert", "bleu"} {
		if brute[i] > brute[3] {
			fautifs = append(fautifs, fmt.Sprintf("%s à %d", canal, brute[i]))
		}
	}
	return fmt.Sprintf("%s au-dessus de l'alpha %d", strings.Join(fautifs, ", "), brute[3])
}

// rawSettings porte les réglages tels qu'ils s'écrivent.
type rawSettings struct {
	manifest.Commentable

	Border      int                 `json:"bord_px"`
	Margin      int                 `json:"marge_px"`
	GaugeHeight int                 `json:"hauteur_jauge_px"`
	Colors      map[string][4]uint8 `json:"teintes"`
}
