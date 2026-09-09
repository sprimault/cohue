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
	// sorteDecor est une case du décor qui dépasse du sol. Elle vient en tête
	// pour ce que la comparaison en fait à égalité exacte : une créature posée
	// pile au centre de la case qu'elle foule se peint alors devant elle, ce qui
	// est le bon sens d'un trottoir sous des pieds.
	sorteDecor sorte = iota
	// sorteEpave est ce qu'une caisse a laissé. Elle vient juste après le décor,
	// donc elle passe sous tout ce qui partage sa profondeur : c'est l'exception
	// que la conception donne aux cadavres, et pour la même raison — une trace
	// aplatie au sol n'a jamais à cacher ce qui bouge.
	sorteEpave
	sorteEnnemi
	// sorteAmbiance est le décor mouvant. Il est trié avec le reste et non peint
	// à part : un figurant passe devant et derrière les créatures comme
	// n'importe quel corps posé au sol, et l'exclure du tri le mettrait toujours
	// au-dessus ou toujours en dessous — ce qui se verrait aussitôt.
	sorteAmbiance
	sorteTir
	sorteTirHorde
	sorteGemme
	sorteAimant
	// sorteCaisse est un objet posé au sol, trié avec le reste : elle coûte à
	// traverser plutôt qu'elle n'arrête, si bien qu'une créature lui passe dessus
	// et doit alors être peinte devant ou derrière selon sa profondeur, comme
	// n'importe quel corps.
	sorteCaisse
	// sorteArmeAuSol est une arme lourde tombée d'une caisse, triée comme le
	// reste : le joueur marche dessus pour la prendre, donc il passe devant ou
	// derrière selon sa profondeur.
	sorteArmeAuSol
	// sorteFioleAuSol est une fiole tombée d'une caisse, triée comme une arme et
	// pour la même raison : elle se ramasse en marchant dessus.
	sorteFioleAuSol
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
// Rien n'y est alloué après le montage, et c'est `source` qui le rend vrai : le
// dimensionnement était une seconde liste de bassins, plus courte que celle du
// relevé. Ce n'est pas l'invariant du budget, qui s'arrête à la simulation, mais
// un tri qui allouerait soixante fois par seconde ferait passer le
// ramasse-miettes exactement là où il se verrait — et, au-delà de la capacité du
// tampon, l'image ne se contentait pas d'allouer : elle paniquait.
//
// **Le décor y entre pour ce qui dépasse du sol, et pour cela seulement.** Une
// forme d'élévation nulle est toujours dessous : elle se peint en passe
// préalable, où l'ordre ne décide de rien. Un muret, lui, dispute sa profondeur
// à ce qui passe derrière, et le lui refuser ferait dessiner une créature
// par-dessus le mur qui la cache.
//
// **Seul le décor visible est relevé**, là où les bassins le sont en entier. La
// différence n'est pas un traitement de faveur : une carte de neuf mille cases
// en a quelques centaines à l'écran, quand trois cents créatures y sont presque
// toutes. Ce qui borne le relevé est donc la fenêtre et non la carte, et c'est
// ce qui rend le coût indépendant de la taille du lieu.
//
// **Les gemmes en sont, bien qu'elles soient au sol**, et la raison n'est pas
// statique : posées, elles pourraient être peintes avec le décor, puisqu'un
// personnage a son appui en bas de son image et ne descend jamais sous ses
// pieds. C'est l'aimant qui l'interdit — une gemme qui converge vers le joueur
// traverse la horde, si bien qu'elle a une profondeur à disputer. Les ranger avec
// le décor aujourd'hui obligerait à les en sortir dès ce moment-là.
//
// **L'anticipation était juste, sa raison ne l'était pas.** On avait écrit que la
// gemme volerait « à hauteur de torse » ; la ruée montre qu'elle rase le sol —
// huit pixels de haut contre quarante-huit pour une créature, l'appui de l'une et
// de l'autre sur leur position. Ce que la séquence décide n'est donc pas un
// croisement à mi-corps mais **le passage devant ou derrière des pieds** : une
// gemme un peu plus proche se peint sur le bas d'une silhouette, une gemme un peu
// plus loin disparaît dessous. La conclusion tient, le mécanisme est autre, et
// c'est ce genre d'écart qu'une anticipation laissée non vérifiée fait passer
// pour acquis.
type scene struct {
	// comptes porte, par seau, le nombre d'entités puis leur position de départ
	// dans la séquence. Sa taille vient de l'englobant du lieu et jamais d'une
	// constante : un lieu plus grand y déborderait, un plus petit y gaspillerait.
	comptes []int
	// recueil est ce qui a été relevé des bassins, tampon est la même séquence
	// rangée. Deux tranches parce que la distribution du tri ne peut pas se
	// faire en place, et préallouées parce qu'elle a lieu à chaque image.
	recueil, tampon []entite
	// sources sont les bassins que l'image dessine, montés une fois.
	sources []source
}

// source est un bassin que l'image dessine : ce qu'il peut contenir, et de quoi
// le relever.
//
// **Les deux sur la même ligne, et c'est tout l'objet de ce type.** Le
// dimensionnement et le relevé étaient deux listes tenues à deux endroits : la
// première comptait cinq bassins quand la seconde en parcourait sept, si bien
// que le tampon pouvait être trop court — et le tri ne se contentait pas d'y
// allouer, il paniquait sur une tranche trop petite. L'écart s'est produit deux
// fois de suite, parce que rien dans la forme ne le signalait.
//
// Un bassin qui s'ajoute porte désormais sa capacité avec lui, et il n'y a plus
// de seconde liste à tenir d'accord.
type source struct {
	capacite int
	relever  func(s *scene)
}

// nouvelleScene dimensionne les seaux sur l'étendue d'un lieu et les bassins.
//
// La profondeur d'un point du lieu va de zéro, au sommet du losange, à la somme
// de ses deux côtés, à sa pointe basse : il faut donc un seau de plus que cette
// somme. La capacité des séquences couvre tous les bassins pleins et le joueur,
// c'est-à-dire le plus grand nombre d'entités qu'une image puisse porter — et
// elle se somme sur la liste même que le relevé parcourt.
func nouvelleScene(carte *game.CostGrid, monde *game.World, sol *Terrain, cam *camera) *scene {
	s := &scene{
		comptes: make([]int, carte.Width()+carte.Height()+1),
		sources: sources(monde, sol, cam),
	}

	// Le joueur ne vit dans aucun bassin, d'où celui qu'on ajoute.
	total := 1
	for _, src := range s.sources {
		total += src.capacite
	}
	s.recueil = make([]entite, 0, total)
	s.tampon = make([]entite, total)
	return s
}

// sources énumère ce qu'une image dessine, chacun avec sa capacité.
//
// Les bassins sont pris une fois : `World` les tient pour toute la partie, et
// une relance monte un écran neuf.
//
// Le décor ouvre la liste et n'est pas un bassin : sa capacité est celle de la
// fenêtre, bornée par la carte quand celle-ci est plus petite. Un lieu de neuf
// mille cases n'en montre jamais plus d'un millier ; un lieu d'une pièce n'en a
// pas mille à montrer.
func sources(monde *game.World, sol *Terrain, cam *camera) []source {
	ennemis := monde.Enemies()
	ambiants := monde.Ambients()
	tirs := monde.Shots()
	tirsHorde := monde.EnemyShots()
	gemmes := monde.Gems()
	aimants := monde.Magnets()
	caisses := monde.Crates()
	epaves := monde.Wrecks()
	armesAuSol := monde.Drops()
	fiolesAuSol := monde.Vials()

	largeur := sol.carte.Width()
	fenetre := min(cam.casesMax(), largeur*sol.carte.Height())

	return []source{
		{fenetre, func(s *scene) {
			u0, v0, u1, v1 := cam.casesVisibles()
			for v := v0; v <= v1; v++ {
				for u := u0; u <= u1; u++ {
					f, posee := sol.formeDe(u, v)
					if !posee || !f.triee {
						continue
					}
					// Le centre de la case, comme une créature se tient au
					// centre de la sienne : c'est ce qui met les deux sur le
					// même point de comparaison. Le sommet bas de l'emprise,
					// qui est pourtant le point où l'image se pose, mettrait un
					// mur et la créature qui le longe à égalité.
					s.ajouter(game.FromInt(u)+game.One/2, game.FromInt(v)+game.One/2,
						v*largeur+u, sorteDecor, v*largeur+u)
				}
			}
		}},
		{ennemis.Cap(), func(s *scene) {
			for i := range ennemis.Active() {
				e := ennemis.At(i)
				s.ajouter(e.X, e.Y, ennemis.IDAt(i), sorteEnnemi, i)
			}
		}},
		{ambiants.Cap(), func(s *scene) {
			for i := range ambiants.Active() {
				a := ambiants.At(i)
				s.ajouter(a.X, a.Y, ambiants.IDAt(i), sorteAmbiance, i)
			}
		}},
		{tirs.Cap(), func(s *scene) {
			for i := range tirs.Active() {
				p := tirs.At(i)
				s.ajouter(p.X, p.Y, tirs.IDAt(i), sorteTir, i)
			}
		}},
		{tirsHorde.Cap(), func(s *scene) {
			for i := range tirsHorde.Active() {
				p := tirsHorde.At(i)
				s.ajouter(p.X, p.Y, tirsHorde.IDAt(i), sorteTirHorde, i)
			}
		}},
		{gemmes.Cap(), func(s *scene) {
			for i := range gemmes.Active() {
				g := gemmes.At(i)
				s.ajouter(g.X, g.Y, gemmes.IDAt(i), sorteGemme, i)
			}
		}},
		{aimants.Cap(), func(s *scene) {
			for i := range aimants.Active() {
				a := aimants.At(i)
				s.ajouter(a.X, a.Y, aimants.IDAt(i), sorteAimant, i)
			}
		}},
		{epaves.Cap(), func(s *scene) {
			for i := range epaves.Active() {
				e := epaves.At(i)
				s.ajouter(e.X, e.Y, epaves.IDAt(i), sorteEpave, i)
			}
		}},
		{caisses.Cap(), func(s *scene) {
			for i := range caisses.Active() {
				c := caisses.At(i)
				s.ajouter(c.X, c.Y, caisses.IDAt(i), sorteCaisse, i)
			}
		}},
		{armesAuSol.Cap(), func(s *scene) {
			for i := range armesAuSol.Active() {
				d := armesAuSol.At(i)
				s.ajouter(d.X, d.Y, armesAuSol.IDAt(i), sorteArmeAuSol, i)
			}
		}},
		{fiolesAuSol.Cap(), func(s *scene) {
			for i := range fiolesAuSol.Active() {
				f := fiolesAuSol.At(i)
				s.ajouter(f.X, f.Y, fiolesAuSol.IDAt(i), sorteFioleAuSol, i)
			}
		}},
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
	for _, src := range s.sources {
		src.relever(s)
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
// **Le décor a fermé la moitié d'une dette.** La profondeur exacte et l'abscisse
// s'atteignaient jusqu'ici si peu qu'on ne pouvait rien dire d'elles : deux
// entités ne partagent une profondeur en virgule fixe qu'en étant posées
// exactement au même point du monde, ce que seul l'anneau d'apparition pourrait
// produire. Les cases, elles, tombent sur des profondeurs entières, et celles
// d'une même diagonale se départagent par l'abscisse à chaque image — sans
// conséquence visible, deux cases voisines ne se recouvrant pas, mais le critère
// est exercé au lieu d'être supposé.
//
// **La sorte et l'identifiant restent hors d'atteinte.** Il y faudrait deux
// entités au même point, ou une créature dont la position tombe exactement au
// centre d'une case — le premier des deux départage alors une case et ce qui la
// foule, dans le sens que dit `sorteDecor`. Le seau et l'exception du joueur,
// eux, sont éprouvés : les inverser change la planche.
//
// **Les gemmes n'y changent rien**, contrairement à ce qu'on pourrait croire
// d'un tas : deux créatures meurent à des positions distinctes, donc leurs
// gemmes le sont aussi, et une volée est écartée exprès pour ne pas se
// superposer.
//
// **Le joueur passe devant ce qui partage sa profondeur**, exception que la
// conception assume : perdre son personnage sous un empilement est ce qui peut
// arriver de pire à la lisibilité, et cela survient précisément quand on est
// encerclé, c'est-à-dire quand il faut voir clair.
//
// **Elle vaut aussi contre le décor depuis qu'il entre dans la séquence**, et
// c'est assumé plutôt que subi : un mur du même seau ne mord sur le joueur que
// de quelques pixels, et le montrer par-dessus est ce que la silhouette fera de
// toute façon — elle redessine le personnage sur ce qui le cache. Ce que
// l'exception anticipe ici, elle le rendra alors exact.
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
