// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les créatures, prêtes à peindre : leurs bandes converties une fois, le cycle
// que leur état appelle, et l'image de ce cycle. La table qui relie un état de
// jeu à un nom de cycle vit ici et pas ailleurs.

package render

import (
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/sprimault/cohue/internal/game"
	"github.com/sprimault/cohue/internal/sprite"
)

// Les cycles que le manifeste des personnages déclare.
//
// **Ce sont les seuls noms de cycle écrits dans du Go, et ils le sont ici.** La
// simulation n'en connaît aucun : elle tient des états — un décompte d'éclair,
// une phase de charge, un pas appliqué — et c'est le rendu qui décide lequel se
// dessine comment. Les mettre côté jeu l'aurait obligé à savoir qu'un dessin
// existe.
const (
	cycleRepos   = "repos"
	cycleMarche  = "marche"
	cycleDegat   = "degat"
	cycleAttaque = "attaque"
)

// Cast porte les images des personnages, converties une fois.
//
// Une tranche par rôle plutôt qu'une table par clé : ce que le rendu tient d'une
// entité est l'index de son profil, et c'est l'index qui doit résoudre.
type Cast struct {
	joueur   *figure
	ennemis  []*figure
	ambiants []*figure
}

// figure est un profil dessiné : ses images, et ce que son manifeste dit de la
// façon de les poser.
type figure struct {
	appui [2]int
	// directions sont les orientations dessinées, dans l'ordre du manifeste.
	directions []string
	cycles     map[string]game.Cycle
	images     map[pose]*ebiten.Image
}

// pose désigne une image dans une figure.
type pose struct {
	cycle, direction string
	variante, image  int
}

// NewCast lit les bandes de tous les profils et les convertit.
//
// **Tout au montage, jamais en jeu.** Le budget d'allocation interdit d'ouvrir
// un fichier dans la boucle, et une texture créée à la première apparition d'un
// profil ferait sauter l'image précise où la sixième minute entre.
func NewCast(fsys fs.FS, racine string, profils *game.Profiles) (*Cast, error) {
	troupe := &Cast{
		ennemis:  make([]*figure, len(profils.Enemies)),
		ambiants: make([]*figure, len(profils.Ambient)),
	}

	joueur, err := charger(fsys, racine, profils.Player.Figure)
	if err != nil {
		return nil, err
	}
	troupe.joueur = joueur

	for i, p := range profils.Enemies {
		f, err := charger(fsys, racine, p.Figure)
		if err != nil {
			return nil, err
		}
		troupe.ennemis[i] = f
	}
	for i, p := range profils.Ambient {
		f, err := charger(fsys, racine, p.Figure)
		if err != nil {
			return nil, err
		}
		troupe.ambiants[i] = f
	}
	return troupe, nil
}

// charger découpe les bandes d'un profil et en fait des textures.
//
// La conversion est exhaustive et non paresseuse : une image créée au premier
// usage se paierait pendant une partie, et c'est précisément ce que le montage
// existe pour éviter.
func charger(fsys fs.FS, racine string, f game.Figure) (*figure, error) {
	feuille, err := sprite.Load(fsys, racine, f)
	if err != nil {
		return nil, err
	}

	dessin := &figure{
		appui:      f.Anchor,
		directions: f.Directions,
		cycles:     f.Cycles,
		images:     map[pose]*ebiten.Image{},
	}
	for nom, c := range f.Cycles {
		for _, direction := range f.Directions {
			for variante := range f.Variants {
				for image := range c.Frames {
					img, ok := feuille.Frame(nom, direction, variante, image)
					if !ok {
						continue
					}
					dessin.images[pose{nom, direction, variante, image}] =
						ebiten.NewImageFromImage(img)
				}
			}
		}
	}
	return dessin, nil
}

// image rend la texture d'une pose, ou nil quand la figure ne la porte pas.
func (f *figure) image(p pose) *ebiten.Image { return f.images[p] }

// anim est le cycle qu'un état demande, et ce qui le cadence.
//
// Les trois ensemble parce qu'ils se décident ensemble : le cycle vient de
// l'état, sa cadence de son manifeste, et le décompte est celui du même état.
// Les rendre séparément ferait redemander plus loin « lequel des trois
// décomptes était-ce ».
type anim struct {
	nom   string
	cycle game.Cycle
	// reste est le décompte qui cadence un cycle non bouclé, nul pour un cycle
	// qui boucle — auquel cas c'est le tick qui le cadence.
	reste game.Tick
}

// cycleTenu rend le premier des cycles proposés que la figure porte.
//
// **Tous les profils n'ont pas les mêmes**, et c'est le manifeste qui décide :
// le Molosse n'a ni repos ni attaque, il court et il meurt. Demander un cycle
// absent est donc un cas ordinaire, ce que `Sheet.Frame` annonçait déjà en
// rendant un second retour — et le repli appartient à l'appelant, seul à savoir
// par quoi remplacer une pose qui n'existe pas.
func (f *figure) cycleTenu(choix ...string) (string, game.Cycle) {
	for _, nom := range choix {
		if c, porte := f.cycles[nom]; porte {
			return nom, c
		}
	}
	return "", game.Cycle{}
}

// cycleEnnemi dit quel cycle une créature demande.
//
// **L'ordre suit celui des teintes d'état, et pour les mêmes raisons.** Le
// soigneur passe avant l'impact parce que repérer d'où vient le soin est ce qui
// change la conduite du joueur ; l'impact passe avant le télégraphe parce qu'un
// coup qui vient de porter est plus récent qu'une annonce en cours.
//
// **Un état sans cycle retombe sur le déplacement**, il ne fige pas la créature.
// Le Molosse n'a pas d'`attaque`, si bien que son télégraphe se lit à sa teinte
// qui bat et non à une pose : c'est le repli que `Sheet.Frame` annonçait en
// rendant un second retour, et le rendu est le seul à savoir par quoi remplacer
// une pose absente.
func (f *figure) cycleEnnemi(e *game.Enemy) anim {
	switch {
	case e.Healing > 0:
		if nom, c := f.cycleTenu(cycleAttaque); nom != "" {
			return anim{nom, c, e.Healing}
		}
	case e.Flash > 0:
		if nom, c := f.cycleTenu(cycleDegat); nom != "" {
			return anim{nom, c, e.Flash}
		}
	case e.Telegraphing():
		if nom, c := f.cycleTenu(cycleAttaque); nom != "" {
			return anim{nom, c, e.ChargeTimer}
		}
	}
	return f.deplacement(e.Step)
}

// deplacement rend le cycle d'un personnage selon qu'il avance ou non.
//
// Les deux cycles se replient l'un sur l'autre : un profil qui n'a que `marche`
// s'anime à l'arrêt plutôt que de disparaître, et c'est moins faux qu'une
// créature absente.
func (f *figure) deplacement(pas game.Vec) anim {
	if pas != (game.Vec{}) {
		nom, c := f.cycleTenu(cycleMarche, cycleRepos)
		return anim{nom: nom, cycle: c}
	}
	nom, c := f.cycleTenu(cycleRepos, cycleMarche)
	return anim{nom: nom, cycle: c}
}
