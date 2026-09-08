// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les images du catalogue d'objets : ce qu'on découpe de ce que le manifeste
// déclare. Ramassables, projectiles, effets et particules y voisinent, chacun
// avec ce que sa famille exige.

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

// Prop est un objet prêt à poser : ses images et le coin où elles vont.
type Prop struct {
	// Images sont les images du cycle, une seule pour un objet immobile.
	Images []image.Image
	// Offset est le coin haut-gauche où poser, relatif au point du monde.
	Offset [2]int
	// Cycle est la cadence de la bande, nulle pour une image unique.
	Cycle game.Cycle
	// Masking dit que l'objet dépasse la hauteur d'un personnage, donc qu'il
	// peut en cacher un. Une vitrine et un rideau de fer le déclarent, une
	// caisse de seize pixels non.
	Masking bool
	// Icon est le dessin de face d'une arme lourde, nul pour tout le reste.
	//
	// **Séparé de `Images` parce qu'il ne se pose pas au même endroit** : celles-ci
	// vont dans la scène, en isométrie et avec un ancrage au sol, quand celui-ci
	// va dans un emplacement du bandeau, vu de face. Les confondre ferait poser
	// une icône dans le monde le jour où quelqu'un parcourrait `Images`.
	Icon image.Image
}

// Props porte les objets du catalogue, par nom.
type Props struct {
	objets map[string]Prop
	// armes sont les noms de famille « arme », triés.
	//
	// **Une liste rendue plutôt que des noms écrits dans le rendu.** Les autres
	// objets sont nommés en Go parce que c'est le rendu qui sait lequel va avec
	// quel bassin ; une arme lourde, elle, est désignée par la table des armes, et
	// le rendu pose celle que le monde lui donne. Les écrire en dur obligerait à
	// y revenir à chaque arme ajoutée, pour une correspondance que personne ne
	// choisit.
	armes []string
}

// Weapons rend les noms des armes lourdes du catalogue, triés.
func (p *Props) Weapons() []string { return p.armes }

// LoadProps découpe les images que le catalogue d'objets déclare.
//
// **Le manifeste est décodé ailleurs** — `game.LoadObjects` —, et ce paquet en
// reçoit le résultat : il n'y a qu'une lecture du fichier, donc une seule vérité
// sur lui.
//
// `source` est le manifeste à citer, et non un fichier à ouvrir. Un manquement
// porte sur l'écart entre une déclaration et l'image qui devrait la tenir : le
// nommer par le dossier des images ferait accuser l'image, alors que c'est le
// couple qui est en cause et que seule la déclaration dit ce qui était promis.
//
// **Tout le catalogue, comme pour le décor.** Une image absente ou qui dément
// son manifeste est un défaut du dépôt et non de la partie qui l'aurait
// affichée : la découvrir au premier tir remonterait chez un joueur.
//
// Les manquements sont accumulés : qui régénère les objets veut la liste.
func LoadProps(fsys fs.FS, racine, source string, catalogue *game.Objects) (*Props, error) {
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
		return nil, &manifest.Invalid{Path: source, Missing: manques}
	}
	return props, nil
}

// Prop rend un objet du catalogue, ou dit qu'il n'existe pas.
func (p *Props) Prop(nom string) (Prop, bool) {
	objet, connu := p.objets[nom]
	return objet, connu
}

// charger découpe ce qu'une entrée déclare, et rien de plus.
//
// **Les familles que le rendu ne pose pas encore ne se chargent pas.** Deux
// restent dehors, et chacune attend un mécanisme plutôt qu'une décision :
// « arme » attend les armes lourdes de l'étape 6, et « cycle » — les deux bandes
// d'une caisse qui cède — attend que l'étape 7 donne à la caisse son délai de
// contact. Les découper d'avance serait charger ce que rien n'exerce, dans un
// paquet dont le seul lecteur n'a pas de test.
//
// Un cycle se distingue d'un effet par son ancrage : il appartient à l'objet qui
// le cite et prend le sien, quand un effet se centre sur le point qu'il marque.
// C'est ce qui les sépare le jour où les deux se posent.
func (p *Props) charger(fsys fs.FS, racine, nom string, objet game.Object) error {
	switch objet.Family {
	case game.FamilyWorld:
		img, err := lire(fsys, path.Join(racine, nom+".png"), objet.Size)
		if err != nil {
			return err
		}
		p.objets[nom] = Prop{
			Images:  []image.Image{img},
			Offset:  [2]int{-objet.Anchor[0], -objet.Anchor[1]},
			Masking: objet.Masking,
		}
		if objet.Twinkle != nil {
			return p.chargerScintillement(fsys, racine, nom, objet)
		}
	case game.FamilyWeapon:
		// Les deux dessins d'une arme vivent dans `armes/`, où le générateur les
		// range : le nom du fichier porte le suffixe, pas le catalogue.
		sol, err := lire(fsys, path.Join(racine, "armes", nom+"_sol.png"), objet.GroundSize)
		if err != nil {
			return err
		}
		icone, err := lire(fsys, path.Join(racine, "armes", nom+"_icone.png"), objet.IconSize)
		if err != nil {
			return err
		}
		p.objets[nom] = Prop{
			Images: []image.Image{sol},
			Offset: [2]int{-objet.GroundAt[0], -objet.GroundAt[1]},
			Icon:   icone,
		}
		// Les entrées sont parcourues dans l'ordre trié du manifeste, si bien que
		// cette liste l'est aussi sans qu'on ait à la trier.
		p.armes = append(p.armes, nom)
	case game.FamilyEffect:
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
	case game.FamilyParticle:
		formes, err := decouperLarge(fsys, path.Join(racine, nom+".png"),
			objet.Cell, objet.Shapes)
		if err != nil {
			return err
		}
		p.objets[nom] = Prop{
			// Pas de cadence : ces formes ne se suivent pas, on en choisit une.
			// Un éclat se centre sur sa position comme un effet — il vole, il ne
			// repose sur rien.
			Images: formes,
			Offset: [2]int{-objet.Cell[0] / 2, -objet.Cell[1] / 2},
		}
	}
	return nil
}

// chargerScintillement remplace l'image fixe d'un ramassable par sa boucle.
//
// La cellule et son ancrage se dérivent des deux nombres déclarés : la bande
// fait la largeur de l'objet sur sa hauteur plus le bombement, et l'objet y
// flotte, donc son appui descend de ce bombement. Rien n'est deviné du fichier.
func (p *Props) chargerScintillement(fsys fs.FS, racine, nom string, objet game.Object) error {
	s := objet.Twinkle
	cellule := [2]int{objet.Size[0], objet.Size[1] + s.Amplitude}
	images, err := decouperLarge(fsys, path.Join(racine, nom+"_scintille.png"),
		cellule, s.Frames)
	if err != nil {
		return err
	}
	duree, _ := game.TicksFromMs(s.Duration)
	p.objets[nom] = Prop{
		Images:  images,
		Offset:  [2]int{-objet.Anchor[0], -(objet.Anchor[1] + s.Amplitude)},
		Cycle:   game.Cycle{Frames: s.Frames, Duration: duree, Loop: s.Loop},
		Masking: objet.Masking,
	}
	return nil
}

// controler refuse une entrée dont la famille et les champs se contredisent.
//
// **Le contrôle porte sur ce que ce paquet pose**, et laisse le reste passer :
// juger les charges d'une arme lourde serait juger à la place de l'étape qui les
// lira, et les deux règles finiraient par diverger. C'est la leçon du contrôle
// de dessin, qui avait d'abord été posé dans la simulation.
func controlerObjet(nom string, objet game.Object) string {
	switch objet.Family {
	case game.FamilyWorld:
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
	case game.FamilyWeapon:
		if objet.GroundSize[0] < 1 || objet.IconSize[0] < 1 {
			return fmt.Sprintf("%s : sol %v et icône %v, une arme porte les deux",
				nom, objet.GroundSize, objet.IconSize)
		}
		for i, borne := range objet.GroundAt {
			if borne < 0 || borne >= objet.GroundSize[i] {
				return fmt.Sprintf("%s : ancrage %v, hors d'un dessin au sol de %v",
					nom, objet.GroundAt, objet.GroundSize)
			}
		}
	case game.FamilyEffect:
		if objet.Frames < 1 || objet.Cell[0] < 1 || objet.Cell[1] < 1 {
			return fmt.Sprintf("%s : %d image(s) de %v, un effet a une bande",
				nom, objet.Frames, objet.Cell)
		}
	case game.FamilyParticle:
		if objet.Shapes < 1 || objet.Cell[0] < 1 || objet.Cell[1] < 1 {
			return fmt.Sprintf("%s : %d forme(s) de %v, une particule en a au moins une",
				nom, objet.Shapes, objet.Cell)
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
