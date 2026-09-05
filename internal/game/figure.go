// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le versant dessin d'un profil : la case d'une image, le point d'appui, les
// orientations et les cycles. La simulation n'en lit rien.

package game

// Figure est ce qu'un profil dit de son dessin.
//
// **Elle vit ici parce que le manifeste des personnages décrit les deux à la
// fois**, le rendu et les valeurs de jeu, et qu'un second décodeur du même
// fichier en serait une seconde description. `internal/game` la porte donc sans
// jamais la lire — elle ne contient que des entiers et des chaînes, et rien de
// ce paquet ne dépend d'elle.
//
// Ce qui la remplit est déclaré depuis l'étape 4, avec un commentaire annonçant
// que le rendu le lirait à l'étape 5 : le décodage refusant toute clé inconnue,
// ces champs étaient déjà obligatoires sans avoir d'usage.
type Figure struct {
	// Key est la clé du manifeste, qui nomme aussi le dossier des bandes —
	// `assets/personnages/marcheur/`.
	Key string
	// Side est le côté de la case carrée d'une image, en pixels.
	Side int
	// Anchor est le point d'appui dans cette case, en pixels : c'est lui que le
	// rendu pose sur la position du monde, jamais le coin de l'image.
	Anchor [2]int
	// Variants est le nombre de teintes de vêtement. À un, les bandes sont dans
	// le dossier du profil ; au-delà, dans des sous-dossiers `v0`, `v1`…
	//
	// Le chargeur n'a rien à savoir des variantes qui n'existent pas, ce qui est
	// aussi pourquoi un profil à teinte unique n'a pas de sous-dossier.
	Variants int
	// Directions sont les orientations dessinées, dans l'ordre du manifeste.
	// Elles nomment les fichiers — `marche_SO.png`.
	Directions []string
	// Cycles sont les animations, par nom.
	//
	// **Tous les profils n'ont pas les mêmes**, et le rendu ne peut pas supposer
	// qu'un cycle existe : le Molosse n'a ni repos ni attaque, il charge et il
	// meurt. C'est le manifeste qui dit ce que chacun porte.
	Cycles map[string]Cycle
}

// Cycle est une animation déclarée par un profil.
type Cycle struct {
	// Frames est le nombre d'images de la bande, qui en fixe la largeur :
	// `Frames × Side`.
	Frames int
	// Duration est la durée d'une image, convertie une fois au chargement.
	//
	// Le manifeste l'écrit en millisecondes et la simulation ne connaît que le
	// tick : convertir à l'usage rouvrirait la question à chaque appel, et
	// 100 ms vaudraient six pas ou sept selon qui écrit le code.
	Duration Tick
	// Loop dit si le cycle reprend à la fin. Une mort ne boucle pas.
	Loop bool
}

// figure rend le versant dessin d'un profil brut.
func (p rawProfile) figure(cle string) Figure {
	cycles := make(map[string]Cycle, len(p.Cycles))
	for nom, c := range p.Cycles {
		duree, _ := TicksFromMs(c.DurationMs)
		cycles[nom] = Cycle{Frames: c.Frames, Duration: duree, Loop: c.Loop}
	}
	return Figure{
		Key:        cle,
		Side:       p.Side,
		Anchor:     p.Anchor,
		Variants:   p.Variants,
		Directions: p.Directions,
		Cycles:     cycles,
	}
}
