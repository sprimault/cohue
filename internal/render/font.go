// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La pose du texte : la planche de glyphes rangée dans l'atlas, découpée une
// fois, et les deux façons de poser une chaîne — nue sur un fond connu,
// contourée sur le monde.

package render

import (
	"fmt"
	"io/fs"
	"path"

	"github.com/sprimault/lozengine/raster"
	"github.com/sprimault/lozengine/render"
	"github.com/sprimault/lozengine/scene"
)

// Font pose du texte à partir de la planche de glyphes.
//
// Les mesures viennent de `Metrics`, qui les décode et les vérifie sans atlas ;
// ce type n'ajoute que la planche et le quad. Aucune métrique n'est écrite ici :
// le manifeste fait contrat, et changer de fonte ne touche pas ce fichier.
//
// **La mise en page reste à cohue, et ce n'est pas un choix.** Le moteur sait
// lire une planche de glyphes à chasse fixe ou uniformément resserrée ; la fonte
// du jeu a sept avances distinctes, de quatre à onze pixels pour une cellule de
// onze. Mesurer une ligne demande donc l'avance par glyphe, que `Metrics` porte
// et que `Advance` rend.
type Font struct {
	mesures *Metrics
	// planche est l'index de la planche de glyphes dans l'atlas, et cases sa
	// découpe — une colonne par glyphe, dans l'ordre de la chaîne du manifeste.
	planche uint16
	cases   []raster.Rectangle
}

// LoadFont lit le manifeste d'interface et range la planche qu'il désigne.
//
// La planche est cherchée à côté du manifeste : c'est le manifeste qui la nomme,
// et le chemin du dossier ne se recompose pas ailleurs.
//
// **La découpe vient du moteur et non d'une boucle de sous-images.** `CutGrid`
// range ses cases de gauche à droite, ce qui est l'ordre de la chaîne de glyphes
// du manifeste : `Metrics.Place` indexe donc directement la découpe, et un
// glyphe ne coûte plus une image à part.
func LoadFont(fsys fs.FS, chemin string, atlas *raster.Atlas) (*Font, error) {
	mesures, err := LoadMetrics(fsys, chemin)
	if err != nil {
		return nil, err
	}

	nom := path.Join(path.Dir(chemin), mesures.Sheet)
	f, err := fsys.Open(nom)
	if err != nil {
		return nil, fmt.Errorf("lecture de %s: %w", nom, err)
	}
	planche, err := raster.LoadSheet(f)
	_ = f.Close()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", nom, err)
	}

	largeur, hauteur := mesures.Cell[0], mesures.Cell[1]
	if int(planche.Height) != hauteur {
		return nil, fmt.Errorf("%s: planche haute de %d px pour une cellule de %d",
			nom, planche.Height, hauteur)
	}
	// La largeur se confronte par la découpe, qui refuse une cellule qui ne
	// divise pas : une colonne tombée en silence serait une planche de glyphes où
	// personne ne verrait le décalage avant de lire un texte accentué.
	// #nosec G115 -- la cellule est confrontée à la hauteur de la planche, bornée
	// par le type du moteur
	if err := planche.CutGrid(uint16(largeur), planche.Height); err != nil {
		return nil, fmt.Errorf("%s: %w", nom, err)
	}

	index, err := atlas.Add(planche)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", nom, err)
	}
	// **Un glyphe n'a pas d'ancre, et lui en donner une le décalait.** L'ancre
	// convertit une pose de monde — où le moteur centre le pied de l'image sur le
	// point visé — en coin haut-gauche ; `Scene.PlaceScreen` vise déjà ce coin et
	// ajoute l'ancre par-dessus. Celle que `AnchorFromOffset` rendait pour une
	// cellule de 11×9 valait donc cinq pixels vers la droite et neuf vers le bas,
	// et tout le texte du bandeau tombait une ligne plus bas que sa jauge.
	return &Font{mesures: mesures, planche: index, cases: planche.Rectangles}, nil
}

// Advance rend la largeur d'une chaîne, en pixels de tampon.
//
// C'est par elle qu'on pose un texte contre un bord droit : un texte placé à
// distance fixe déborde dès qu'il s'allonge, et un minuteur déborde exactement
// quand la partie dure.
func (f *Font) Advance(s string) int { return f.mesures.Advance(s) }

// Height rend la hauteur d'une ligne, en pixels de tampon.
func (f *Font) Height() int { return f.mesures.Height() }

// Draw pose une chaîne, coin haut-gauche de la première cellule en (x, y).
//
// La teinte multiplie le blanc de la planche, qui ne porte que des glyphes
// opaques : le texte prend donc exactement la couleur demandée.
//
// **Un quad par glyphe, et l'ordre de pose les départage.** Ils partagent leur
// couche et leur planche, et c'est ce que le moteur garantit entre pièces
// d'écran : ce qui est posé après recouvre ce qui précède.
//
// Le pixel est celui de la surface, marge comprise : c'est le repère d'un
// bandeau, d'un panneau, d'un écran de fin. Un texte accroché à une chose du
// monde passe par `Font.DrawOutlinedArea`.
func (f *Font) Draw(sc *scene.Scene, texte string, x, y int, teinte render.Color) {
	f.poser(sc, texte, x, y, teinte, false)
}

// DrawOutlined pose une chaîne avec un contour d'un pixel.
//
// C'est ce que la conception exige de tout texte posé **sur le monde** plutôt
// que sur un cadre : un chiffre de dégâts jaillit au-dessus de n'importe quoi —
// carrelage clair, flaque, créature sombre — et sa seule couleur ne suffit pas à
// le détacher. Le contour est en croix et non sur les huit voisins : à cette
// taille, les diagonales épaissiraient le glyphe au point de le fermer.
//
// Cinq quads par glyphe, et le texte vient en dernier : l'ordre de pose décide,
// donc le contour doit précéder ce qu'il entoure.
func (f *Font) DrawOutlined(sc *scene.Scene, texte string, x, y int, teinte, contour render.Color) {
	f.contourer(sc, texte, x, y, teinte, contour, false)
}

// DrawOutlinedArea pose une chaîne contourée dans l'aire de jeu, et non sur la
// surface.
//
// **C'est la voie d'un texte accroché à une chose du monde** — le chiffre de
// dégâts est le seul du jeu : le pixel donné est celui que `pixelDe` rend, la
// ligne se découpe sur l'aire, et une créature touchée juste hors du champ
// n'écrit donc pas son chiffre dans la marge, à côté du jeu.
//
// **Il n'y a pas de variante sans contour, et il n'y en aura pas sans besoin.**
// Un texte posé sur le monde en demande un, c'est la règle du chapitre 14 ; une
// seconde porte qui s'en passe inviterait à l'enfreindre sans rien apporter.
//
// **Elle pose dans la couche des effets, là que les deux autres posent dans
// celle de l'interface**, et le repère décide donc de la couche. Ce n'est pas un
// raccourci : un texte accroché à une chose du monde *est* un effet — il dit ce
// qui vient de se passer —, et le laisser dans l'interface le ferait passer
// devant le panneau des cartes pendant la pause où le joueur choisit.
func (f *Font) DrawOutlinedArea(sc *scene.Scene, texte string, x, y int, teinte, contour render.Color) {
	f.contourer(sc, texte, x, y, teinte, contour, true)
}

// contourer pose le contour en croix puis le texte, dans l'un ou l'autre repère.
func (f *Font) contourer(sc *scene.Scene, texte string, x, y int,
	teinte, contour render.Color, dansLAire bool) {
	for _, d := range [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
		f.poser(sc, texte, x+d[0], y+d[1], contour, dansLAire)
	}
	f.poser(sc, texte, x, y, teinte, dansLAire)
}

// poser avance la ligne glyphe par glyphe, dans l'un ou l'autre repère.
//
// Un drapeau et non la méthode de pose en paramètre : une valeur de méthode liée
// est une fermeture, et une ligne se pose à chaque image — qu'elle tienne sur la
// pile dépendrait de l'analyse d'échappement, ce qui n'est pas une hypothèse à
// laisser dans le budget d'allocation.
func (f *Font) poser(sc *scene.Scene, texte string, x, y int,
	teinte render.Color, dansLAire bool) {
	for _, r := range texte {
		place, connu := f.mesures.Place(r)
		if !connu || place >= len(f.cases) {
			// Le trou que `Metrics.Advance` compte : rien n'est posé, mais la
			// cellule est franchie, si bien qu'un caractère manquant se voit au
			// lieu de se fondre dans un texte plausible.
			x += f.mesures.Cell[0]
			continue
		}
		c := f.cases[place]
		a := scene.Aspect{
			Sheet:   f.planche,
			SourceX: c.X,
			SourceY: c.Y,
			Width:   c.Width,
			Height:  c.Height,
			Tint:    teinte,
			Layer:   coucheInterface,
		}
		if dansLAire {
			a.Layer = coucheEffets
			sc.PlaceArea(pixelDEcran(x), pixelDEcran(y), a)
		} else {
			sc.PlaceScreen(pixelDEcran(x), pixelDEcran(y), a)
		}
		x += f.mesures.Advance(string(r))
	}
}
