// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les images du catalogue d'objets : ce qu'on découpe de ce que le manifeste
// déclare. Ramassables, projectiles, effets et particules y voisinent, chacun
// avec ce que sa famille exige.

package sprite

import (
	"fmt"
	"io/fs"
	"maps"
	"path"
	"slices"

	"github.com/sprimault/lozengine/raster"
	"github.com/sprimault/lozengine/scene"

	"github.com/sprimault/cohue/internal/game"
	"github.com/sprimault/cohue/internal/manifest"
)

// Prop est un objet prêt à poser : ses images et l'ancre qui les place.
type Prop struct {
	// Images sont les images du cycle, une seule pour un objet immobile.
	Images []scene.Aspect
	// Cycle est la cadence de la bande, nulle pour une image unique.
	Cycle game.Cycle
	// Masking dit que l'objet dépasse la hauteur d'un personnage, donc qu'il
	// peut en cacher un. Une vitrine et un rideau de fer le déclarent, une
	// caisse de seize pixels non.
	Masking bool
	// Icon est le dessin de face d'une arme lourde, et sa valeur zéro dit qu'il
	// n'y en a pas — une largeur nulle ne couvre aucun pixel, ce qui est la
	// façon dont le moteur reconnaît déjà ce qui ne montre rien.
	//
	// **Séparé de `Poses` parce qu'il ne se pose pas au même endroit** :
	// celles-ci vont dans la scène, en isométrie et avec un ancrage au sol,
	// quand celui-ci va dans un emplacement du bandeau, vu de face. Les
	// confondre ferait poser une icône dans le monde le jour où quelqu'un
	// parcourrait `Poses`.
	//
	// Son ancre pose son coin haut-gauche sur le point visé, parce qu'un
	// emplacement d'interface se mesure par son coin et non par un pied.
	Icon scene.Aspect
	// Shapes et IconShape sont les mêmes dessins aplatis en blanc, absents tant
	// que `Props.Silhouette` ne les a pas demandés.
	//
	// **Deux usages seulement, et c'est pourquoi ils se demandent** : le
	// projectile de la horde, que la conception révèle quand quelque chose le
	// cache, et le liseré qui détache une icône d'un sol clair. Dériver le
	// catalogue entier doublerait ses planches pour des dessins que rien ne pose.
	Shapes    []scene.Aspect
	IconShape scene.Aspect
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

// LoadProps range dans l'atlas les images que le catalogue d'objets déclare.
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
func LoadProps(fsys fs.FS, racine, source string, catalogue *game.Objects, atlas *raster.Atlas) (*Props, error) {
	props := &Props{objets: make(map[string]Prop, len(catalogue.Items))}
	var manques []string
	for _, nom := range slices.Sorted(maps.Keys(catalogue.Items)) {
		objet := catalogue.Items[nom]
		if defaut := controlerObjet(nom, objet); defaut != "" {
			manques = append(manques, defaut)
			continue
		}
		if err := props.charger(fsys, racine, nom, objet, atlas); err != nil {
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

// Silhouette range dans l'atlas l'aplat des images des objets nommés.
//
// **Par nom et non pour tout le catalogue**, parce qu'un seul objet en a besoin
// — le projectile de la horde, que la conception révèle quand quelque chose le
// cache — et que les planches d'un atlas sont comptées. Un nom inconnu est une
// erreur de programmation en amont et se dit comme telle : le rendu choisit ce
// qu'il révèle, donc il sait ce qu'il demande.
//
// **Séparée de `IconSilhouette` parce que les deux dessins ne servent pas au
// même geste**, et qu'un objet qui a les deux n'en révèle qu'un : une arme pose
// son icône au-dessus d'une caisse et son dessin au sol dans la scène, où rien
// ne la révèle. Les dériver ensemble aurait rangé une planche que personne ne
// pose, par objet et pour toujours.
//
// L'aplat d'une bande garde sa découpe, si bien qu'une image y prend le même
// rectangle que dans son dessin.
func (p *Props) Silhouette(atlas *raster.Atlas, noms ...string) error {
	for _, nom := range noms {
		objet, connu := p.objets[nom]
		if !connu {
			return fmt.Errorf("aplat de %q : objet absent du catalogue", nom)
		}
		if len(objet.Images) == 0 {
			return fmt.Errorf("aplat de %q : aucune image à aplatir", nom)
		}

		plate, err := aplatirLaPlanche(atlas, objet.Images[0])
		if err != nil {
			return fmt.Errorf("aplat de %q : %w", nom, err)
		}
		objet.Shapes = make([]scene.Aspect, len(objet.Images))
		for i, a := range objet.Images {
			a.Sheet = plate
			objet.Shapes[i] = a
		}
		p.objets[nom] = objet
	}
	return nil
}

// IconSilhouette range dans l'atlas l'aplat des icônes des objets nommés.
//
// C'est le liseré qui détache une icône d'un sol clair, là où le bandeau n'en a
// pas besoin : ses cases sont sombres, et c'est pour elles que les icônes ont
// été dessinées. Un objet sans icône est refusé pour la même raison qu'un nom
// inconnu — l'appelant sait ce qu'il pose.
func (p *Props) IconSilhouette(atlas *raster.Atlas, noms ...string) error {
	for _, nom := range noms {
		objet, connu := p.objets[nom]
		if !connu {
			return fmt.Errorf("aplat d'icône de %q : objet absent du catalogue", nom)
		}
		if objet.Icon.Width == 0 {
			return fmt.Errorf("aplat d'icône de %q : cet objet n'en a pas", nom)
		}

		plate, err := aplatirLaPlanche(atlas, objet.Icon)
		if err != nil {
			return fmt.Errorf("aplat d'icône de %q : %w", nom, err)
		}
		objet.IconShape = objet.Icon
		objet.IconShape.Sheet = plate
		p.objets[nom] = objet
	}
	return nil
}

// aplatirLaPlanche range l'aplat de la planche qu'un aspect désigne et rend son
// index.
//
// L'aspect sert à désigner la planche et non à décrire ce qu'on aplatit : c'est
// la planche entière qui se dérive, avec sa découpe, pour que toutes les images
// d'une bande partagent un seul aplat.
func aplatirLaPlanche(atlas *raster.Atlas, a scene.Aspect) (uint16, error) {
	return atlas.Add(Flatten(atlas.Sheets[a.Sheet]))
}

// charger découpe ce qu'une entrée déclare, et rien de plus.
//
// **Les cinq familles se posent maintenant**, la dernière étant « cycle » — les
// deux bandes d'une caisse qui cède —, qui attendait le délai de contact pour
// avoir quelque chose à cadencer. Les découper d'avance aurait été charger ce que
// rien n'exerce, dans un paquet dont le seul lecteur n'a pas de test.
//
// Un cycle se distingue d'un effet par son ancrage : il appartient à l'objet qui
// le cite et se pose au sol comme lui, quand un effet se centre sur le point
// qu'il marque.
func (p *Props) charger(fsys fs.FS, racine, nom string, objet game.Object, atlas *raster.Atlas) error {
	switch objet.Family {
	case game.FamilyWorld:
		poses, err := poser(fsys, atlas, path.Join(racine, nom+".png"),
			objet.Size, objet.Size, decalageDe(objet.Anchor))
		if err != nil {
			return err
		}
		pose := Prop{Images: poses, Masking: objet.Masking}
		// **Un objet du monde peut avoir une icône, et la fiole est le premier.**
		// Son fichier est à côté de son dessin de monde quand celui d'une arme vit
		// dans `armes/` : ce dossier porte ce qui est une arme, et le suffixe
		// nomme l'icône dans les deux cas.
		if objet.IconSize != [2]int{} {
			icone, err := chargerIcone(fsys, atlas, path.Join(racine, nom+"_icone.png"), objet.IconSize)
			if err != nil {
				return err
			}
			pose.Icon = icone
		}
		p.objets[nom] = pose
		if objet.Twinkle != nil {
			return p.chargerScintillement(fsys, racine, nom, objet, atlas)
		}
	case game.FamilyWeapon:
		// Les deux dessins d'une arme vivent dans `armes/`, où le générateur les
		// range : le nom du fichier porte le suffixe, pas le catalogue.
		poses, err := poser(fsys, atlas, path.Join(racine, "armes", nom+"_sol.png"),
			objet.GroundSize, objet.GroundSize, decalageDe(objet.GroundAt))
		if err != nil {
			return err
		}
		icone, err := chargerIcone(fsys, atlas, path.Join(racine, "armes", nom+"_icone.png"), objet.IconSize)
		if err != nil {
			return err
		}
		p.objets[nom] = Prop{Images: poses, Icon: icone}
		// Les entrées sont parcourues dans l'ordre trié du manifeste, si bien que
		// cette liste l'est aussi sans qu'on ait à la trier.
		p.armes = append(p.armes, nom)
	case game.FamilyEffect:
		poses, err := poser(fsys, atlas, path.Join(racine, nom+".png"),
			// Un effet se centre sur ce qu'il marque : il n'a pas de point
			// d'appui parce qu'il ne se pose pas au sol, il recouvre.
			bandeDe(objet.Cell, objet.Frames), objet.Cell, centre(objet.Cell))
		if err != nil {
			return err
		}
		duree, _ := game.TicksFromMs(objet.Duration)
		p.objets[nom] = Prop{
			Images: poses,
			Cycle:  game.Cycle{Frames: objet.Frames, Duration: duree, Loop: objet.Loop},
		}
	case game.FamilyCycle:
		poses, err := poser(fsys, atlas, path.Join(racine, nom+".png"),
			bandeDe(objet.Cell, objet.Frames), objet.Cell,
			// Le bas-centre de la cellule, qui est l'ancrage que le générateur
			// donne à tout objet du monde : un cycle est dessiné autour du même
			// point au sol que l'objet qui le cite. Le déduire de sa cellule
			// évite de lui faire porter un renvoi vers cet objet, que rien
			// d'autre ne lirait.
			[2]int{-objet.Cell[0] / 2, -(objet.Cell[1] - 1)})
		if err != nil {
			return err
		}
		duree, _ := game.TicksFromMs(objet.Duration)
		p.objets[nom] = Prop{
			Images: poses,
			Cycle:  game.Cycle{Frames: objet.Frames, Duration: duree, Loop: objet.Loop},
		}
	case game.FamilyParticle:
		poses, err := poser(fsys, atlas, path.Join(racine, nom+".png"),
			// Un éclat se centre sur sa position comme un effet — il vole, il ne
			// repose sur rien.
			bandeDe(objet.Cell, objet.Shapes), objet.Cell, centre(objet.Cell))
		if err != nil {
			return err
		}
		// Pas de cadence : ces formes ne se suivent pas, on en choisit une.
		p.objets[nom] = Prop{Images: poses}
	}
	return nil
}

// chargerScintillement remplace l'image fixe d'un ramassable par sa boucle.
//
// La cellule et son ancrage se dérivent des deux nombres déclarés : la bande
// fait la largeur de l'objet sur sa hauteur plus le bombement, et l'objet y
// flotte, donc son appui descend de ce bombement. Rien n'est deviné du fichier.
func (p *Props) chargerScintillement(fsys fs.FS, racine, nom string, objet game.Object, atlas *raster.Atlas) error {
	s := objet.Twinkle
	cellule := [2]int{objet.Size[0], objet.Size[1] + s.Amplitude}
	poses, err := poser(fsys, atlas, path.Join(racine, nom+"_scintille.png"),
		bandeDe(cellule, s.Frames), cellule,
		decalageDe([2]int{objet.Anchor[0], objet.Anchor[1] + s.Amplitude}))
	if err != nil {
		return err
	}
	duree, _ := game.TicksFromMs(s.Duration)
	// **La pose déjà écrite est reprise, elle n'est pas rebâtie.** Ce que la
	// bande remplace est l'image et son ancrage ; ce qui ne dépend pas d'elle —
	// l'icône d'une fiole — s'y perdrait en silence, et un bandeau sans dessin ne
	// se distingue pas d'un bandeau qu'on n'a pas encore rempli.
	pose := p.objets[nom]
	pose.Images = poses
	pose.Cycle = game.Cycle{Frames: s.Frames, Duration: duree, Loop: s.Loop}
	pose.Masking = objet.Masking
	p.objets[nom] = pose
	return nil
}

// chargerIcone range une icône d'interface et rend de quoi la dessiner.
//
// Son coin suffit à la poser : un emplacement de bandeau se mesure par là, et
// non par un pied.
func chargerIcone(fsys fs.FS, atlas *raster.Atlas, chemin string, taille [2]int) (scene.Aspect, error) {
	images, err := poser(fsys, atlas, chemin, taille, taille, [2]int{})
	if err != nil {
		return scene.Aspect{}, err
	}
	return images[0], nil
}

// bandeDe rend la taille d'une planche qui porte des cases côte à côte.
func bandeDe(cellule [2]int, cases int) [2]int {
	return [2]int{cellule[0] * cases, cellule[1]}
}

// centre rend le décalage d'une image posée centrée sur le point qu'elle marque.
func centre(cellule [2]int) [2]int {
	return [2]int{-cellule[0] / 2, -cellule[1] / 2}
}

// decalageDe rend le décalage d'une image posée par son point d'appui.
//
// Les deux conventions du dépôt se croisent ici, et `AnchorFromFooting` dit ce
// que leur confusion coûte : un appui s'ajoute en négatif, il ne se passe pas
// tel quel.
func decalageDe(appui [2]int) [2]int {
	return [2]int{-appui[0], -appui[1]}
}

// controlerObjet refuse une entrée dont la famille et les champs se contredisent.
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
	case game.FamilyCycle:
		if objet.Frames < 1 || objet.Cell[0] < 1 || objet.Cell[1] < 1 {
			return fmt.Sprintf("%s : %d image(s) de %v, un cycle a une bande",
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
