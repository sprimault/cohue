// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// L'aperçu animé : ce qui transforme une suite d'images du tampon en GIF. La
// planche montre un instant, celui-ci montre ce qui ne se juge qu'en mouvement
// — une horde qui converge, une salve qui s'élargit, des gemmes qui partent.

package main

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"io/fs"
	"os"
	"slices"

	"github.com/sprimault/cohue/internal/session"
)

// imagesParSeconde est la cadence du GIF.
//
// Vingt plutôt que les soixante du jeu : l'œil ne distingue pas les deux sur du
// pixel art qui bouge, et le fichier est trois fois plus léger. Le délai d'une
// image s'exprime en centièmes de seconde, ce qui fait cinq — soixante ne se
// divise pas par vingt-cinq ou trente sans reste, et un délai fractionnaire
// n'existe pas dans le format.
const imagesParSeconde = 20

// delaiParImage est ce que le format inscrit entre deux images, en centièmes.
const delaiParImage = 100 / imagesParSeconde

// couleursGIF est ce qu'une palette de GIF peut porter.
const couleursGIF = 256

// pilotage joue une partie au pilote automatique, cartes comprises.
//
// **Les cartes se prennent, et c'est ce qui manquait à la planche.** Une montée
// de niveau met la horde en pause tant que personne ne tranche, et la première
// tombe à la douzième seconde : sans elles, une vue qui joue sept minutes
// s'arrête à la douzième et dessine une salle figée. Les places sont celles du
// test de déterminisme, pour que les deux relectures voient la même partie.
type pilotage struct {
	partie  *session.Session
	montees int
}

// pas joue un tick, en tranchant d'abord un choix s'il y en a un.
func (p *pilotage) pas() {
	if p.partie.World.Choosing() {
		p.partie.World.Choose(session.PilotChoice(p.montees))
		p.montees++
	}
	p.partie.World.Step(session.Pilot(p.partie.World.Tick()))
}

// avancer joue le nombre de ticks donné sans rien dessiner.
func (p *pilotage) avancer(ticks int) {
	for range ticks {
		p.pas()
	}
}

// ticksAvantLaMort rend le nombre de pas qui séparent le début de la partie du
// dernier instant où le joueur est vivant.
//
// **Un artefact s'arrête sur l'événement qu'il montre, pas sur un compte de
// pas** — la règle est dans `docs/go.md`, et ce lot l'a enfreinte avant de la
// relire : demander la septième minute rendait trois secondes d'écran de mort,
// la simulation continuant de tourner quand le joueur a succombé. Le pilote est
// médiocre par construction et meurt vers la quatrième minute, chiffre qu'un
// réglage d'équilibrage déplacera sans prévenir.
//
// La partie se rejoue depuis une session neuve plutôt que de se dérouler à
// l'envers, ce que rien ne permet. Deux passes sur la même graine voient la même
// chose, c'est ce que l'invariant du déterminisme garantit, et la première ne
// dessine rien — elle coûte quelques secondes.
func ticksAvantLaMort(assets fs.FS, campagne string, graine uint64, plafond int) (int, error) {
	partie, err := session.Open(assets, campagne, graine)
	if err != nil {
		return 0, err
	}
	pilote := &pilotage{partie: partie}
	for tick := range plafond {
		if !partie.World.Alive() {
			return tick, nil
		}
		pilote.pas()
	}
	return plafond, nil
}

// palettePartagee rend la palette commune à toutes les images.
//
// **Une palette exacte quand le rendu tient dans 256 teintes**, ce qui est le
// cas ordinaire : le décor est borné à vingt-six couleurs par image et les
// sprites le sont aussi. Le GIF est alors fidèle au pixel près, sans le
// délavage qu'une palette générique impose.
//
// Au-delà, les 256 teintes les plus employées sont retenues et le reste se
// rabat sur la plus proche. **Sans tramage** : un GIF tramé n'est plus du pixel
// art, et ce qu'on donne à relire est précisément la netteté.
//
// La palette est commune à toutes les images plutôt que calculée par image :
// une palette par image ferait clignoter les teintes d'une image à l'autre là
// où rien n'a bougé.
func palettePartagee(images [][]byte) color.Palette {
	compte := map[color.RGBA]int{}
	for _, pixels := range images {
		for i := 0; i+3 < len(pixels); i += 4 {
			compte[color.RGBA{pixels[i], pixels[i+1], pixels[i+2], pixels[i+3]}]++
		}
	}

	teintes := make([]color.RGBA, 0, len(compte))
	for teinte := range compte {
		teintes = append(teintes, teinte)
	}
	// Le tri décide quelles teintes survivent à l'écrêtage, donc il doit être
	// total : à fréquence égale, deux teintes départagées par le parcours d'une
	// map donneraient un GIF différent à chaque exécution.
	trierParUsage(teintes, compte)

	// La première place revient au transparent, qui n'est pas une teinte du
	// rendu mais ce qui dit « ce point n'a pas changé » : à l'intérieur du
	// rectangle réécrit, les points identiques à l'image précédente le prennent,
	// et de longues séries constantes remplacent le bruit des textures. C'est ce
	// qui fait passer la séquence de dix mégaoctets à un.
	if len(teintes) > couleursGIF-1 {
		teintes = teintes[:couleursGIF-1]
	}
	palette := make(color.Palette, 0, len(teintes)+1)
	palette = append(palette, color.RGBA{})
	for _, teinte := range teintes {
		palette = append(palette, teinte)
	}
	return palette
}

// trierParUsage range les teintes de la plus employée à la moins employée, et
// départage celles qui s'égalent par leurs composantes.
func trierParUsage(teintes []color.RGBA, compte map[color.RGBA]int) {
	rang := func(t color.RGBA) uint32 {
		return uint32(t.R)<<24 | uint32(t.G)<<16 | uint32(t.B)<<8 | uint32(t.A)
	}
	slices.SortFunc(teintes, func(a, b color.RGBA) int {
		if compte[a] != compte[b] {
			return compte[b] - compte[a]
		}
		switch {
		case rang(a) < rang(b):
			return -1
		case rang(a) > rang(b):
			return 1
		}
		return 0
	})
}

// difference rend le rectangle des points qui changent d'une image à l'autre.
//
// **C'est ce qui décide du poids du fichier**, et de loin : la caméra se
// déplaçant en pixels entiers, une grande part de l'écran est identique d'une
// image à la suivante. Écrire l'écran entier à chaque fois coûtait seize
// mégaoctets là où le rectangle utile en coûte un — le format prévoit
// exactement cela, et ne pas s'en servir revient à lui faire compresser du
// bruit qu'il a déjà vu.
//
// Un rectangle vide dit que rien n'a bougé ; l'appelant écrit alors un point,
// parce qu'une image de surface nulle n'est pas représentable.
func difference(avant, apres []byte, taille image.Point) image.Rectangle {
	minX, minY, maxX, maxY := taille.X, taille.Y, -1, -1
	for y := range taille.Y {
		for x := range taille.X {
			i := 4 * (y*taille.X + x)
			if avant[i] == apres[i] && avant[i+1] == apres[i+1] &&
				avant[i+2] == apres[i+2] && avant[i+3] == apres[i+3] {
				continue
			}
			minX, maxX = min(minX, x), max(maxX, x)
			minY, maxY = min(minY, y), max(maxY, y)
		}
	}
	if maxX < 0 {
		return image.Rectangle{}
	}
	return image.Rect(minX, minY, maxX+1, maxY+1)
}

// ecrireGIF assemble les images lues du tampon et les écrit.
//
// Les pixels arrivent tels que le tampon les rend, quatre octets par point dans
// l'ordre RGBA, et l'index de palette se cherche d'abord dans une table : la
// recherche linéaire de `color.Palette.Index` coûterait deux cent cinquante-six
// comparaisons par point, soit des milliards pour une séquence entière.
func ecrireGIF(chemin string, taille image.Point, images [][]byte) error {
	palette := palettePartagee(images)
	index := make(map[color.RGBA]uint8, len(palette))
	for i, teinte := range palette {
		index[teinte.(color.RGBA)] = uint8(i)
	}

	anime := &gif.GIF{
		Image:    make([]*image.Paletted, 0, len(images)),
		Delay:    make([]int, 0, len(images)),
		Disposal: make([]byte, 0, len(images)),
		// La taille logique se déclare, puisque les images suivantes ne
		// couvrent qu'une part de l'écran : sans elle, le lecteur prendrait la
		// première image pour la toile entière.
		Config: image.Config{ColorModel: palette, Width: taille.X, Height: taille.Y},
	}

	place := func(teinte color.RGBA) uint8 {
		if p, connue := index[teinte]; connue {
			return p
		}
		// Le masque dit ce que le type ne dit pas : une palette de GIF compte au
		// plus deux cent cinquante-six places, donc `Index` tient dans un octet.
		// Sans lui, la conversion est une promesse que rien ne tient.
		return uint8(palette.Index(teinte) & 0xff)
	}

	for n, pixels := range images {
		rect := image.Rect(0, 0, taille.X, taille.Y)
		if n > 0 {
			if rect = difference(images[n-1], pixels, taille); rect.Empty() {
				rect = image.Rect(0, 0, 1, 1)
			}
		}
		cadre := image.NewPaletted(rect, palette)
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			for x := rect.Min.X; x < rect.Max.X; x++ {
				i := 4 * (y*taille.X + x)
				if n > 0 && images[n-1][i] == pixels[i] &&
					images[n-1][i+1] == pixels[i+1] &&
					images[n-1][i+2] == pixels[i+2] &&
					images[n-1][i+3] == pixels[i+3] {
					continue // l'index zéro est déjà là, et il est transparent
				}
				cadre.SetColorIndex(x, y, place(color.RGBA{
					pixels[i], pixels[i+1], pixels[i+2], pixels[i+3]}))
			}
		}
		anime.Image = append(anime.Image, cadre)
		anime.Delay = append(anime.Delay, delaiParImage)
		// Chaque image se pose sur la précédente et n'en efface rien : c'est ce
		// qui donne son sens au rectangle partiel.
		anime.Disposal = append(anime.Disposal, gif.DisposalNone)
	}

	// Le chemin n'a pas de part variable, comme celui des vues : `sortie` est une
	// constante et le nom vient de la table.
	f, err := os.OpenFile(chemin, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600) // #nosec G304
	if err != nil {
		return err
	}
	err = gif.EncodeAll(f, anime)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return fmt.Errorf("%s: %w", chemin, err)
	}
	fmt.Println(chemin)
	return nil
}
