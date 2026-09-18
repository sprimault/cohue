// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les objets d'une partie, prêts à peindre : leurs images converties une fois,
// et les noms de catalogue que le rendu emploie.

package render

import (
	"fmt"
	"io/fs"
	"maps"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/sprimault/cohue/internal/game"
	"github.com/sprimault/cohue/internal/sprite"
)

// Les objets que le rendu pose, par leur nom de catalogue.
//
// **Ce sont les seuls noms d'objet écrits dans du Go**, comme les cycles et les
// directions le sont ailleurs, et pour la même raison : la simulation tient des
// bassins — des gemmes, des projectiles —, et c'est le rendu qui sait lequel se
// dessine avec quoi. Elle n'a pas à savoir qu'une image existe.
const (
	objetGemme  = "gemme"
	objetAimant = "aimant"
	objetCaisse = "caisse"
	// La fiole, dont le nom canonique vit ailleurs : c'est le manifeste de
	// progression qui la désigne, et la simulation l'y lit pour savoir ce qu'elle
	// rend. Celui-ci ne sert qu'à charger son dessin, comme pour la caisse, et
	// tout ce qui se résout en partie — l'icône du bandeau, celle qui flotte
	// au-dessus d'une caisse — passe par la clé que le monde donne. Les deux
	// divergeraient franchement : `NewStage` refuserait un nom absent du
	// catalogue.
	objetFiole     = "fiole"
	objetTir       = "projectile_base"
	objetTirHorde  = "projectile_ennemi"
	objetEtincelle = "etincelle"
	objetSouffle   = "souffle"
	// Les flammes qu'une lourde pose. **Une bande par case et non une image par
	// flaque** : le rayon vient du manifeste des armes et se règle, quand un
	// dessin taillé pour lui se périmerait au premier réglage — et il ne suivrait
	// plus la zone que les impulsions appliquent.
	objetFlammes = "flammes"
	// Rien ne vole d'une Baudruche qui explose, l'onde suffit. Le jour où elle
	// laissera des éclats de chair, leur nom viendra de son profil, comme celui
	// d'une caisse vient de sa clé `destruction` — pas d'une constante ici.
)

// Stage porte les objets d'une partie, convertis une fois.
//
// Une table par nom et non une tranche par bassin : un bassin peut poser deux
// objets — une caisse et sa ruine —, et un objet peut n'appartenir à aucun,
// comme l'étincelle d'un impact.
type Stage struct {
	objets map[string]prop
	// icones sont les dessins de face des armes lourdes, que le bandeau pose
	// dans un emplacement. Séparées des `objets` parce qu'elles ne se posent ni
	// au même endroit ni dans le même repère : celles-là sont vues de face,
	// ceux-ci en isométrie.
	icones map[string]*ebiten.Image
	// bordsDIcone sont les mêmes icônes aplaties, dont le rendu tire le liseré
	// qui les détache d'un sol clair. Le bandeau n'en a pas besoin : ses cases
	// sont sombres, et c'est pour elles que les icônes ont été dessinées.
	bordsDIcone map[string]*ebiten.Image
	// cassables porte ce que le manifeste attache à chaque objet qui se casse :
	// ses cycles, sa ruine et ses éclats. Le monde nomme ce qui a cédé, et c'est
	// ici qu'on lit ce qu'il laisse.
	cassables map[string]game.Destruction
	// pivots nomme, pour ce qui se pose dans les deux sens, son dessin le long
	// de v.
	pivots map[string]string
	// caisse porte ce que le manifeste attache à la caisse, dont le cycle
	// d'appui que la caisse est seule à jouer.
	//
	// **Lus au catalogue plutôt qu'écrits ici**, à la différence des noms
	// ci-dessus, et la nuance tient à qui choisit. Le rendu décide qu'une gemme
	// se dessine avec `gemme` — rien ne le dit ailleurs. Ce qu'une caisse joue en
	// cédant, en revanche, est déjà déclaré sous sa clé `destruction` : l'écrire
	// une seconde fois en Go en ferait deux vérités, et le manifeste-contrat
	// existe pour que remplacer une bande soit un changement de fichier.
	caisse game.Destruction
}

// duree rend la longueur d'un cycle du catalogue, en ticks, et zéro pour ce qui
// n'en a pas.
func (s *Stage) duree(nom string) game.Tick {
	objet, connu := s.objets[nom]
	if !connu {
		return 0
	}
	// #nosec G115 -- le nombre d'images est celui d'une bande découpée, donc
	// borné par la largeur de l'image que le chargement a lue
	return objet.cycle.Duration * game.Tick(objet.cycle.Frames)
}

// cassable rend ce que le manifeste attache à un objet qui se casse.
func (s *Stage) cassable(nom string) game.Destruction { return s.cassables[nom] }

// sens rend le dessin d'un objet dans le sens où il est posé : le sien le long
// de u, celui que `pivote` nomme le long de v.
func (s *Stage) sens(nom string, travers bool) string {
	if pivot, connu := s.pivots[nom]; travers && connu {
		return pivot
	}
	return nom
}

// Icon rend l'icône d'une arme lourde, nulle si le catalogue n'en a pas.
//
// **Le bandeau ne charge aucune image et n'en cherche aucune** : il pose celle
// qu'on lui donne, ce qui garde la frontière — il ne connaît pas le catalogue,
// donc il ne peut pas savoir quelle icône va où.
func (s *Stage) Icon(nom string) *ebiten.Image { return s.icones[nom] }

// bordDIcone rend le liseré d'une icône, nul si le catalogue n'en a pas.
func (s *Stage) bordDIcone(nom string) *ebiten.Image { return s.bordsDIcone[nom] }

// prop est ce qu'un objet donne à peindre.
type prop struct {
	images []*ebiten.Image
	// masques sont les formes des mêmes images, un bit par pixel, et tous les
	// objets en ont : une gemme ne se révèle pas, mais elle recouvre.
	masques []*sprite.Mask
	// cache dit que l'objet dépasse la hauteur d'un personnage. Une vitrine et un
	// rideau de fer le déclarent au manifeste, une caisse de seize pixels non.
	cache  bool
	dx, dy int
	cycle  game.Cycle
	// formes sont les mêmes images aplaties en blanc, et seul le projectile de
	// la horde en a : c'est le second des deux que la conception révèle quand
	// quelque chose les cache, avec le joueur.
	formes []*ebiten.Image
}

// NewStage résout le catalogue d'objets en images posables.
//
// Il ne retient que ce que le rendu pose : ce que le monde peut nommer, et ce
// que le manifeste attache à ce qui se casse. Charger le reste serait payer des
// textures pour ce que rien ne dessine.
//
// `source` est le manifeste à citer dans un manquement, jamais un fichier à
// ouvrir : le catalogue arrive décodé.
func NewStage(fsys fs.FS, racine, source string, objets *game.Objects) (*Stage, error) {
	catalogue, err := sprite.LoadProps(fsys, racine, source, objets)
	if err != nil {
		return nil, err
	}

	caisse, connue := objets.Items[objetCaisse]
	if !connue || caisse.Destruction == nil {
		return nil, fmt.Errorf("objets : « %s » ne declare pas comment elle se casse", objetCaisse)
	}

	scene := &Stage{
		objets:      map[string]prop{},
		icones:      map[string]*ebiten.Image{},
		bordsDIcone: map[string]*ebiten.Image{},
		caisse:      *caisse.Destruction,
		cassables:   map[string]game.Destruction{},
		pivots:      map[string]string{},
	}

	// **Les armes viennent du catalogue et non d'une liste écrite ici**, à la
	// différence de tout le reste : le rendu sait qu'une gemme se dessine avec
	// `gemme`, mais une arme lourde est désignée par la table des armes, et il
	// pose celle que le monde lui donne. Les nommer en dur demanderait d'y revenir
	// à chaque arme ajoutée.
	noms := append([]string{
		objetGemme, objetAimant, objetCaisse, objetFiole, objetTir, objetTirHorde,
		objetEtincelle, objetSouffle, objetFlammes,
	}, catalogue.Weapons()...)

	// Ce qui se casse vient du catalogue pour la même raison : la clé
	// `destruction` de chaque objet nomme déjà ses cycles, sa ruine et ses
	// éclats, et `pivote` son dessin le long de v. Le monde désigne ce qui a
	// cédé par son nom, et c'est ici qu'on lit ce que ce nom laisse.
	for _, nom := range slices.Sorted(maps.Keys(objets.Items)) {
		d := objets.Items[nom].Destruction
		if d == nil {
			continue
		}
		scene.cassables[nom] = *d
		for _, cle := range []string{nom, d.Ruin, d.Shards, d.PressCycle, d.BreakCycle} {
			if cle == "" {
				continue
			}
			noms = append(noms, cle)
			if pivot := objets.Items[cle].Pivot; pivot != "" {
				scene.pivots[cle] = pivot
				noms = append(noms, pivot)
			}
		}
	}

	for _, nom := range noms {
		if _, chargee := scene.objets[nom]; chargee {
			continue
		}
		objet, connu := catalogue.Prop(nom)
		if !connu {
			return nil, fmt.Errorf("objets : « %s » n'est pas au catalogue", nom)
		}
		images := make([]*ebiten.Image, 0, len(objet.Images))
		masques := make([]*sprite.Mask, 0, len(objet.Images))
		var formes []*ebiten.Image
		for _, img := range objet.Images {
			images = append(images, ebiten.NewImageFromImage(img))
			masques = append(masques, sprite.NewMask(img))
			if nom == objetTirHorde {
				formes = append(formes, aplatir(img))
			}
		}
		scene.objets[nom] = prop{
			images:  images,
			masques: masques,
			cache:   objet.Masking,
			dx:      objet.Offset[0],
			dy:      objet.Offset[1],
			cycle:   objet.Cycle,
			formes:  formes,
		}
		if objet.Icon != nil {
			scene.icones[nom] = ebiten.NewImageFromImage(objet.Icon)
			scene.bordsDIcone[nom] = cerner(objet.Icon)
		}
	}
	return scene, nil
}

// image rend l'image d'un objet au tick donné, décalée par une identité.
//
// **Le scintillement se dérive du tick**, comme les cycles d'un personnage, et
// le décalage par l'identifiant vaut ici ce qu'il valait là-bas : deux cents
// gemmes qui pulseraient ensemble feraient un stroboscope, quand décalées elles
// font un tapis qui bouge.
//
// Un objet d'une seule image ignore l'un et l'autre : sa cadence est nulle, et
// `Loop` rend alors la première.
func (s *Stage) image(nom string, tick game.Tick, identite int) (prop, *ebiten.Image) {
	return s.rendre(nom, func(objet prop) int {
		return sprite.Loop(objet.cycle, tick, identite)
	})
}

// effet rend l'image d'une animation brève, cadencée par le décompte de l'état
// qui la porte.
//
// C'est la même dérivation que celle d'un cycle d'attaque : l'animation s'achève
// quand l'état s'achève, et un état plus court qu'elle la fait entrer en cours
// de route. L'étincelle est dans ce dernier cas — trois images de deux ticks
// pour un éclair qui en dure quatre —, si bien qu'on en voit la fin. Ce qu'elle
// doit dire est que le tir a porté, pas comment l'impact se déroule.
func (s *Stage) effet(nom string, reste game.Tick) (prop, *ebiten.Image) {
	return s.rendre(nom, func(objet prop) int {
		return sprite.Once(objet.cycle, reste)
	})
}

// forme rend l'aplat d'une image de l'objet, ou nil quand il n'en a pas.
func (p prop) forme(i int) *ebiten.Image {
	if i < 0 || i >= len(p.formes) {
		return nil
	}
	return p.formes[i]
}

// masque rend la forme d'une image de l'objet, ou nil quand le rang n'en a pas.
func (p prop) masque(i int) *sprite.Mask {
	if i < 0 || i >= len(p.masques) {
		return nil
	}
	return p.masques[i]
}

// rendre résout un objet et l'image que son cadencement désigne.
func (s *Stage) rendre(nom string, quelle func(prop) int) (prop, *ebiten.Image) {
	objet, connu := s.objets[nom]
	if !connu || len(objet.images) == 0 {
		return prop{}, nil
	}
	i := quelle(objet)
	if i < 0 || i >= len(objet.images) {
		i = 0
	}
	return objet, objet.images[i]
}
