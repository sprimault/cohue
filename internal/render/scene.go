// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// L'ordre de dessin : ce qui compose une image, la clé qui range chaque chose et
// le tri par compartiments qui les met en séquence. Les exceptions de la
// conception vivent dans la comparaison, et nulle part ailleurs.

package render

import "github.com/sprimault/cohue/internal/game"

// sorte dit dans quel bassin une place se résout.
//
// Le rendu mêle des entités venues de bassins différents dans une seule
// séquence, et une place ne veut rien dire sans savoir d'où elle vient.
type sorte uint8

// Les bassins d'où une place se résout.
//
// L'ordre des valeurs ne décide de rien : `avant` traite le joueur par sa sorte
// et non par son rang, et n'en compare deux que pour départager ce que deux
// bassins ont numéroté chacun de son côté.
const (
	sorteEnnemi sorte = iota
	sorteTir
	sorteGemme
	sorteJoueur
)

// entite est une chose à dessiner, réduite à ce qui décide de son rang.
//
// Elle porte ses clés et non sa position, que le dessin relit dans le bassin :
// la recopier ici en ferait une seconde description, qu'une séquence gardée d'une
// image à l'autre finirait par démentir.
type entite struct {
	// profondeur vaut `x + y`, et abscisse `x - y`.
	//
	// Ce sont les deux axes de l'écran à un facteur près : la projection
	// multiplie la première par la demi-hauteur et la seconde par la
	// demi-largeur, deux constantes positives qui ne changent aucun ordre.
	// Ranger sur ces sommes évite un flottant, et évite surtout de dépendre de
	// la taille de tuile pour comparer deux rangs.
	profondeur, abscisse game.Fixed
	// identite est l'identifiant stable de l'entité dans son bassin, le seul
	// critère qui ne bouge pas d'une image à l'autre.
	identite int
	sorte    sorte
	place    int
}

// scene met en séquence ce qu'une image doit dessiner.
//
// Un tri par compartiments et non un tri général : la profondeur est bornée par
// l'étendue du lieu, ce qui donne des seaux naturels d'une tuile d'épaisseur. À
// trois cents entités sur un lieu de trente-deux tuiles de côté, cela fait cinq
// entités par seau, qu'une insertion range sans y penser.
//
// Rien n'y est alloué après le montage. Ce n'est pas l'invariant du budget, qui
// s'arrête à la simulation, mais un tri qui allouerait soixante fois par seconde
// ferait passer le ramasse-miettes exactement là où il se verrait.
//
// **Elle ne range que les entités, et le décor est peint avant elle.** C'est
// juste tant que le sol est plat : une face de case n'a rien qui dépasse, donc
// rien à disputer. Le jour où les formes du décor auront leur volume, un muret
// devra entrer dans la même séquence — sans quoi une créature derrière lui se
// dessinera par-dessus.
//
// **Les gemmes en sont, bien qu'elles soient au sol**, et la raison n'est pas
// statique : posées, elles pourraient être peintes avec le décor, puisqu'un
// personnage a son appui en bas de son image et ne descend jamais sous ses
// pieds. C'est l'aimant qui l'interdit — une gemme qui converge vers le joueur
// se dessine à hauteur de torse et croise la horde, si bien qu'elle a une
// profondeur à disputer. Les ranger avec le décor aujourd'hui obligerait à les
// en sortir dès ce moment-là.
type scene struct {
	// comptes porte, par seau, le nombre d'entités puis leur position de départ
	// dans la séquence. Sa taille vient de l'englobant du lieu et jamais d'une
	// constante : un lieu plus grand y déborderait, un plus petit y gaspillerait.
	comptes []int
	// recueil est ce qui a été relevé des bassins, tampon est la même séquence
	// rangée. Deux tranches parce que la distribution du tri ne peut pas se
	// faire en place, et préallouées parce qu'elle a lieu à chaque image.
	recueil, tampon []entite
}

// nouvelleScene dimensionne les seaux sur l'étendue d'un lieu et les bassins.
//
// La profondeur d'un point du lieu va de zéro, au sommet du losange, à la somme
// de ses deux côtés, à sa pointe basse : il faut donc un seau de plus que cette
// somme. La capacité des séquences couvre les trois bassins pleins et le joueur,
// c'est-à-dire le plus grand nombre d'entités qu'une image puisse porter.
func nouvelleScene(carte *game.CostGrid, ennemis, tirs, gemmes int) *scene {
	total := ennemis + tirs + gemmes + 1
	return &scene{
		comptes: make([]int, carte.Width()+carte.Height()+1),
		recueil: make([]entite, 0, total),
		tampon:  make([]entite, total),
	}
}

// ranger rend ce qu'il faut dessiner, du plus lointain au plus proche.
//
// La tranche rendue vaut pour l'image en cours : le prochain appel la réécrit.
func (s *scene) ranger(monde *game.World) []entite {
	s.recueillir(monde)
	return s.compartimenter()
}

// recueillir relève dans les bassins ce que l'image doit porter.
func (s *scene) recueillir(monde *game.World) {
	s.recueil = s.recueil[:0]

	ennemis := monde.Enemies()
	for i := range ennemis.Active() {
		e := ennemis.At(i)
		s.ajouter(e.X, e.Y, ennemis.IDAt(i), sorteEnnemi, i)
	}

	tirs := monde.Shots()
	for i := range tirs.Active() {
		p := tirs.At(i)
		s.ajouter(p.X, p.Y, tirs.IDAt(i), sorteTir, i)
	}

	gemmes := monde.Gems()
	for i := range gemmes.Active() {
		g := gemmes.At(i)
		s.ajouter(g.X, g.Y, gemmes.IDAt(i), sorteGemme, i)
	}

	// Le joueur ne vit dans aucun bassin : il est seul, donc son identité ne
	// départage rien.
	x, y := monde.Player()
	s.ajouter(x, y, 0, sorteJoueur, 0)
}

// ajouter place une entité dans le recueil, sans encore la ranger.
func (s *scene) ajouter(x, y game.Fixed, identite int, quoi sorte, place int) {
	s.recueil = append(s.recueil, entite{
		profondeur: x + y,
		abscisse:   x - y,
		identite:   identite,
		sorte:      quoi,
		place:      place,
	})
}

// compartimenter range le recueil et rend la séquence triée.
//
// Comptage, sommes cumulées, distribution : le tri par compartiments dans sa
// forme qui n'alloue pas. Les entités d'un même seau se retrouvent contiguës
// dans un ordre quelconque, et l'insertion les range ensuite — elle est le bon
// choix sur cinq éléments, où elle bat tout ce qui divise.
func (s *scene) compartimenter() []entite {
	for i := range s.comptes {
		s.comptes[i] = 0
	}
	for _, e := range s.recueil {
		s.comptes[s.seau(e)]++
	}

	depart := 0
	for i, n := range s.comptes {
		s.comptes[i] = depart
		depart += n
	}

	sequence := s.tampon[:len(s.recueil)]
	for _, e := range s.recueil {
		seau := s.seau(e)
		sequence[s.comptes[seau]] = e
		s.comptes[seau]++
	}

	insertion(sequence)
	return sequence
}

// seau rend l'indice du compartiment d'une entité, borné à ceux qui existent.
//
// Une position hors du lieu n'est pas un défaut à faire remonter : le spawner
// posera des créatures au-delà du bord visible, et une image ne doit pas
// s'interrompre parce que l'une d'elles est trop loin. Elle se range au bord, ce
// qui déplace son rang parmi des entités qu'on ne voit pas.
func (s *scene) seau(e entite) int {
	i := e.profondeur.Floor()
	switch {
	case i < 0:
		return 0
	case i >= len(s.comptes):
		return len(s.comptes) - 1
	}
	return i
}

// insertion range une séquence déjà compartimentée.
//
// Elle n'y déplace rien au-delà d'un seau, puisque aucune entité ne précède
// celles du seau d'avant : son coût est celui d'un seau et non celui de la
// séquence, ce qui autorise à l'appeler une fois sur le tout plutôt qu'une fois
// par compartiment.
func insertion(s []entite) {
	for i := 1; i < len(s); i++ {
		e := s[i]
		j := i - 1
		for j >= 0 && avant(e, s[j]) {
			s[j+1] = s[j]
			j--
		}
		s[j+1] = e
	}
}

// avant dit si une entité se dessine avant une autre, donc derrière elle.
//
// La clé est **totale et stable**, ce que la conception exige et ce que ses
// derniers critères achètent. La profondeur range ; l'abscisse départage deux
// entités d'une même bande ; la sorte sépare ce que deux bassins numérotent
// chacun de son côté ; l'identifiant tranche à l'intérieur d'un bassin. Sans ces
// deux derniers, l'ordre retomberait sur celui des bassins, que l'échange à la
// suppression change dès qu'une entité meurt ailleurs — deux sprites superposés
// se relaieraient au premier plan d'une image à l'autre, et le scintillement se
// voit tout de suite.
//
// La sorte n'est donc pas décorative : `Pool.IDAt` numérote par bassin, si bien
// qu'un ennemi et un projectile peuvent porter le même identifiant sans avoir
// rien de commun.
//
// **Quatre de ces critères ne sont pas encore éprouvés, et il faut le savoir.**
// La profondeur exacte, l'abscisse, la sorte et l'identifiant ne s'atteignent que si deux
// entités ont exactement la même profondeur en virgule fixe, ce que rien ne
// produit aujourd'hui : un seau fait une tuile, soit seize pixels d'ordonnée,
// alors que deux entités d'un même seau sont le plus souvent très écartées en
// abscisse — celles qui se chevauchent à l'écran appartiennent à des seaux
// différents, et c'est le premier critère qui les range. Ce qui les atteindra est
// l'anneau d'apparition de l'étape 4, qui superpose des créatures par
// construction. Le seau et l'exception du joueur, eux, sont éprouvés : les
// inverser change la planche.
//
// **Les gemmes ne les atteignent pas non plus**, contrairement à ce qu'on
// pourrait croire d'un tas : deux créatures meurent à des positions distinctes,
// donc leurs gemmes le sont aussi, et une volée est écartée exprès pour ne pas
// se superposer. La dette reste donc entière.
//
// **Le joueur passe devant ce qui partage sa profondeur**, exception que la
// conception assume : perdre son personnage sous un empilement est ce qui peut
// arriver de pire à la lisibilité, et cela survient précisément quand on est
// encerclé, c'est-à-dire quand il faut voir clair.
//
// Le grain de l'exception est le **seau**, une bande d'une tuile d'épaisseur, et
// non l'égalité exacte des profondeurs. Deux positions en virgule fixe ne sont
// jamais exactement à la même profondeur, si bien qu'une exception posée sur
// cette égalité ne se déclencherait pour ainsi dire jamais et ne protégerait de
// rien — alors que le chevauchement, lui, se produit sur toute la bande.
func avant(a, b entite) bool {
	if sa, sb := a.profondeur.Floor(), b.profondeur.Floor(); sa != sb {
		return sa < sb
	}
	if (a.sorte == sorteJoueur) != (b.sorte == sorteJoueur) {
		return b.sorte == sorteJoueur
	}
	if a.profondeur != b.profondeur {
		return a.profondeur < b.profondeur
	}
	if a.abscisse != b.abscisse {
		return a.abscisse < b.abscisse
	}
	if a.sorte != b.sorte {
		return a.sorte < b.sorte
	}
	return a.identite < b.identite
}
