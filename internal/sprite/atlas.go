// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Du fichier à l'index : une planche ouverte, confrontée à son manifeste,
// découpée en cases et rangée dans l'atlas que la scène désigne.

package sprite

import (
	"fmt"
	"io/fs"
	"math"

	"github.com/sprimault/lozengine/raster"
	"github.com/sprimault/lozengine/render"
	"github.com/sprimault/lozengine/scene"
)

// planche est ce qu'une planche rangée laisse à qui pose ses images : l'index
// qui la désigne dans l'atlas, et la découpe qui y prend une case.
//
// La planche elle-même voyage avec, parce que la forme d'une case se relève sur
// ses pixels — le masque, l'aplat, le bord. Elle ne coûte rien à garder : son
// tableau est celui que l'atlas partage, `raster.Atlas.Add` n'en recopiant que
// la description.
type planche struct {
	index  uint16
	cases  []raster.Rectangle
	source raster.Sheet
}

// ranger ouvre une planche, la confronte à la taille annoncée, la découpe en
// cases et la range dans l'atlas.
//
// **Une planche par fichier, jamais par image.** Le dépôt en compte six cent
// cinquante et un quand un atlas en porte mille vingt-quatre : découper en
// autant de planches qu'il y a d'images d'animation dépasserait la borne, et la
// refuser au chargement est tout ce que le moteur peut faire. La découpe est
// donc de la donnée dans la planche, pas une planche de plus.
//
// **La taille vient du manifeste, jamais du fichier.** C'est le contrat que le
// manifeste existe pour tenir : une forme redessinée plus haute sans que son
// manifeste bouge se poserait décalée de la différence, et rien ne le dirait.
//
// La case vaut la planche entière pour ce qui ne s'anime pas, ce qui donne une
// découpe d'une seule case et un chemin unique pour tout le catalogue.
func ranger(fsys fs.FS, atlas *raster.Atlas, chemin string, taille, cases [2]int) (planche, error) {
	f, err := fsys.Open(chemin)
	if err != nil {
		return planche{}, fmt.Errorf("%s : %w", chemin, err)
	}
	p, err := raster.LoadSheet(f)
	// Fermé ici et non en `defer` : ce qui suit ne lit plus le fichier, et une
	// lecture seule n'a rien à rendre à la fermeture.
	_ = f.Close()
	if err != nil {
		return planche{}, fmt.Errorf("%s : %w", chemin, err)
	}

	if int(p.Width) != taille[0] || int(p.Height) != taille[1] {
		return planche{}, fmt.Errorf("%s : %dx%d, le manifeste annonce %dx%d",
			chemin, p.Width, p.Height, taille[0], taille[1])
	}
	if cases[0] < 1 || cases[0] > taille[0] || cases[1] < 1 || cases[1] > taille[1] {
		return planche{}, fmt.Errorf("%s : cases de %dx%d dans une image de %dx%d",
			chemin, cases[0], cases[1], taille[0], taille[1])
	}

	// #nosec G115 -- bornés par la taille confrontée à la planche juste au-dessus
	largeur, hauteur := uint16(cases[0]), uint16(cases[1])
	if err := p.CutGrid(largeur, hauteur); err != nil {
		return planche{}, fmt.Errorf("%s : %w", chemin, err)
	}

	index, err := atlas.Add(p)
	if err != nil {
		return planche{}, fmt.Errorf("%s : %w", chemin, err)
	}
	return planche{index: index, cases: p.Rectangles, source: p}, nil
}

// poser range une planche et rend ses cases, posées autour du point que chacune
// vise.
//
// Le décalage est celui du dépôt — le coin haut-gauche relatif au point visé —
// et `AnchorFromOffset` le convertit vers le pied centré du moteur. Il porte sur
// une case et non sur la planche : c'est une image qu'on pose, pas une bande.
func poser(fsys fs.FS, atlas *raster.Atlas, chemin string, taille, cases, decalage [2]int) ([]Pose, error) {
	ancre, tient := ancre16(AnchorFromOffset(decalage, cases))
	if !tient {
		return nil, fmt.Errorf("%s : décalage %v dans une case de %v, l'ancre n'y tient pas",
			chemin, decalage, cases)
	}
	p, err := ranger(fsys, atlas, chemin, taille, cases)
	if err != nil {
		return nil, err
	}
	return p.poses(ancre), nil
}

// derive range dans l'atlas une planche tirée de celle-ci, à découpe identique.
//
// Les dérivés d'une bande — son aplat, son bord — en sont une par bande et non
// une par image, pour la raison qui tient déjà la découpe : une planche par
// image d'animation ferait des milliers d'entrées dans un atlas qui en porte
// mille vingt-quatre.
func (p planche) derive(atlas *raster.Atlas, tirer func(raster.Sheet) raster.Sheet) (planche, error) {
	tiree := tirer(p.source)
	index, err := atlas.Add(tiree)
	if err != nil {
		return planche{}, err
	}
	return planche{index: index, cases: tiree.Rectangles, source: tiree}, nil
}

// aspect rend de quoi dessiner une case de la planche, ancrée comme l'appelant
// le demande.
//
// **La teinte est blanche et non nulle.** Zéro est transparent sur une couleur,
// donc un aspect rendu sans teinte ne dessinerait rien et rien ne le dirait ;
// le blanc est l'élément neutre de la multiplication, c'est-à-dire l'absence de
// teinte plutôt que l'absence de couleur.
//
// Ce qui reste à l'appelant est ce qui ne regarde pas la découpe : la couche,
// l'étage, la révélation et la teinte d'un état — un éclair d'impact, une
// créature qui s'allume.
func (p planche) aspect(image int, ancre [2]int16) (scene.Aspect, bool) {
	if image < 0 || image >= len(p.cases) {
		return scene.Aspect{}, false
	}
	c := p.cases[image]
	return scene.Aspect{
		Sheet:   p.index,
		SourceX: c.X,
		SourceY: c.Y,
		Width:   c.Width,
		Height:  c.Height,
		Tint:    render.White(),
		AnchorX: ancre[0],
		AnchorY: ancre[1],
	}, true
}

// ancre16 réduit une ancre au type dont le moteur l'attend, et dit quand elle
// n'y tient pas.
//
// **Le refus plutôt que la saturation**, parce qu'une ancre tronquée pose
// l'image ailleurs sans que rien ne le signale : c'est le pire des deux, le
// dessin restant visible et seulement décalé. Une taille d'image et un ancrage
// tenant chacun dans seize bits non signés, leur combinaison n'y tient pas
// toujours — un manifeste qui déclarerait une planche de quarante mille pixels
// est un défaut de génération, et il se nomme ici plutôt qu'au dessin.
func ancre16(x, y int) ([2]int16, bool) {
	if x < math.MinInt16 || x > math.MaxInt16 || y < math.MinInt16 || y > math.MaxInt16 {
		return [2]int16{}, false
	}
	return [2]int16{int16(x), int16(y)}, true
}

// masque relève la forme d'une case de la planche.
func (p planche) masque(image int) *Mask {
	if image < 0 || image >= len(p.cases) {
		return nil
	}
	return NewMask(p.source, p.cases[image])
}
