// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce qui se pose par-dessus le monde sans y être trié : l'éclair d'un état, la
// traînée d'un tir, l'annonce d'une caisse, et ce qui reste d'une chose qui
// n'existe plus.

package render

import (
	"math"
	"strconv"

	"github.com/sprimault/lozengine/render"
	"github.com/sprimault/lozengine/scene"

	"github.com/sprimault/cohue/internal/game"
	"github.com/sprimault/cohue/internal/sprite"
)

// intensiteEclair est la part de sa teinte qu'un éclair d'état ajoute au sprite,
// en fraction de 255.
//
// **Une valeur à juger à l'œil, et le premier chiffre est bas exprès.** Ce que
// l'éclair doit dire est « celle-ci vient d'être atteinte », pas « celle-ci est
// blanche » : au-delà, une mêlée dense devient une nappe claire où les
// silhouettes se perdent, ce qui coûte la lisibilité qu'on cherchait.
//
// **La fraction de 255 remplace le facteur à virgule**, et ce n'est pas une
// traduction : toute atténuation du moteur est entière, donc la même teinte rend
// les mêmes pixels d'une machine à l'autre.
const intensiteEclair = 115

// poserEclair superpose au sprite d'une créature la teinte de son état.
//
// **Une passe additive et non une teinte multipliée**, et c'est ce que les
// sprites ont changé. Sur un aplat blanc, multiplier *était* colorer ; sur un
// dessin, multiplier par un rose presque blanc ne change à peu près rien, et
// l'éclair d'impact — le retour qui manque le plus quand on tire sans le voir —
// avait disparu sans qu'aucun test ne puisse le dire.
//
// **Il était bloqué, et c'est `Entity.Tier` qui l'a rouvert.** Un quad additif
// posé au même point que le sprite doit se peindre après lui ; rien ne
// l'ordonnait entre deux entités, et les trois voies de contournement sont
// fermées — composer dans la planche bute sur la borne de l'atlas, décaler la
// hauteur déplace l'image, monter d'une couche passe aussi devant le décor. La
// strate le dit en un champ.
//
// **Et il reste trié avec la créature, là où les autres effets ne le sont pas.**
// Il appartient au sprite plutôt qu'à la scène : posé en couche haute, il
// s'afficherait sur le mur qui cache la créature qu'il désigne.
func (s *Screen) poserEclair(e scene.Entity, id int, teinte *render.Color) {
	if teinte == nil {
		return
	}
	e.Flags |= render.Additive
	e.Tint = teinte.Scale(intensiteEclair)
	e.Tier = rangDeLEclair
	s.attaches[sorteEclair].poser(s.scene, id, e, s.monde.Tick())
}

// hauteurAnnonce est ce dont l'icône d'une caisse flotte au-dessus du sol, en
// pixels.
//
// La caisse culmine à trente-deux pixels ; au-delà, l'icône décolle et cesse
// d'appartenir à l'objet qu'elle annonce.
const hauteurAnnonce = 38

// poserAnnonce pose au-dessus d'une caisse l'icône de ce qu'elle porte.
//
// **Le contenu est visible avant la casse, sans quoi on casse tout
// systématiquement et ce n'est plus un choix.** C'est la règle du chapitre 7, et
// la seule information qui la remplisse est ce que la caisse a **de plus** que
// les autres : toutes laissent des gemmes, donc les annoncer ne dirait rien et
// couvrirait l'écran de marques qui ne départagent aucune.
//
// **L'icône plutôt que le liseré coloré, que la conception donnait pour
// alternative**, et deux raisons vont dans le même sens. Elle dit **laquelle**,
// ce qui décide dès que le catalogue portera ses quatre lourdes : un liseré ne
// sait dire que « il y a quelque chose », et il faudrait alors une teinte par
// arme. Or la palette n'a plus de place — le violet du projectile ennemi a coûté
// une exploration entière pour trouver son seul creux —, si bien que quatre
// teintes d'armes rouvriraient quatre fois ce problème, en portant en prime dans
// le manifeste une décision d'apparence que le chapitre 14 lui refuse.
//
// **Le dessin de face au milieu de l'isométrie est le vrai coût, et il est plus
// étroit qu'il n'y paraît** : le bandeau pose déjà ces mêmes icônes dans ses
// emplacements. Ce que le joueur lit ici est donc l'objet qu'il connaît, au même
// endroit du langage visuel — « tu l'auras dans ta case ». C'est une continuité
// entre le monde et le bandeau, pas un registre mêlé.
//
// **La fiole s'y annonce comme une arme**, et c'est ce que l'argument du dessus
// prédisait : elle a sa case au bandeau, donc son icône y est déjà connue. La
// clé vient du monde, si bien que rien ici ne distingue les deux sortes — une
// icône se pose, quelle que soit la chose qu'elle nomme.
func (s *Screen) poserAnnonce(c *game.Crate, t trace) {
	cle, porte := s.monde.CrateContent(c)
	if !porte {
		return
	}
	icone := s.objets.Icon(cle)
	if icone.Width == 0 {
		return
	}

	x, y := t.ex, t.ey-hauteurAnnonce

	// **Le liseré vient avant l'icône, qui le recouvre en son centre**, comme
	// pour le personnage : ce qui dépasse est le contour, et il n'a pas d'autre
	// épaisseur que ce dépassement. Sans lui, une icône de gris clairs posée sur
	// un sol clair ne se détache pas — c'est la règle que le chapitre 14 pose
	// pour les chiffres de dégâts, et une icône flottante est dans le même cas.
	if bord := s.objets.bordDIcone(cle); bord.Width != 0 {
		bord.Tint = teinteContourSombre
		for _, d := range contourAutour {
			b := bord
			b.AnchorX += d[0]
			b.AnchorY += d[1]
			b.Layer = coucheEffets
			s.poserAuPoint(x, y, b)
		}
	}
	icone.Layer = coucheEffets
	icone.Tint = render.White()
	s.poserAuPoint(x, y, icone)
}

// poserTrainee relie un tir à la place qu'il occupait au tick précédent.
//
// **Un point de six pixels qui saute de sept ne se lit pas comme un tir.** Relevé
// le 6 septembre 2026 : les projectiles du joueur sont à l'image sur 23,5 % d'une
// partie de dix minutes, ils portent et ils tuent, et on ne les voit pas partir.
// Le contraste n'était pas en cause — 84 de luminance contre le sol, 85 contre un
// Quidam, c'est-à-dire les deux à la fois. Ce qui manquait est la trajectoire,
// que rien ne montrait entre deux positions.
//
// C'est le sprite lui-même qui comble l'écart, plutôt qu'un trait tracé au code :
// aucune forme nouvelle à dessiner, et le pixel entier tient puisque aucune image
// n'est étirée ni pivotée.
//
// **Le nombre de copies se dérive du pas et de la largeur du sprite**, et n'est
// pas un chiffre choisi : une arme plus rapide écarte davantage ses positions, et
// une valeur figée y laisserait des trous le jour où
// `vitesse_projectile_tuiles_s` monte. Un pas plus court que le sprite ne laisse
// rien à combler, et la fonction ne pose alors personne.
//
// Au premier tick de vie d'un projectile, la place précédente est celle du
// joueur : la traînée part alors du canon, ce qui est l'autre moitié de ce que la
// mesure demandait.
func (s *Screen) poserTrainee(x, y game.Fixed, pas game.Vec, identite int) {
	catalogue, i, montre := s.objets.image(objetTir, s.monde.Tick(), identite)
	if !montre {
		return
	}
	aspect, porte := catalogue.image(i)
	if !porte {
		return
	}

	ex, ey := s.pixelDe(x, y)
	ax, ay := s.pixelDe(x-pas.X, y-pas.Y)
	dx, dy := float64(ex-ax), float64(ey-ay)

	largeur := float64(aspect.Width)
	saut := math.Hypot(dx, dy)
	if saut <= largeur {
		return
	}

	aspect.Layer = coucheEffets
	copies := int(math.Ceil(saut / largeur))
	for k := 1; k < copies; k++ {
		part := float64(k) / float64(copies)
		s.poserAuPoint(
			int32(math.Round(float64(ax)+dx*part)),
			int32(math.Round(float64(ay)+dy*part)),
			aspect)
	}
}

// poserEffets pose ce qui reste d'une chose qui n'existe plus.
//
// **Après la séquence et sans y entrer.** Un éclat vole au-dessus du sol et ne
// dispute sa profondeur à personne : le trier lui vaudrait de passer sous la
// créature qui piétine la caisse, ce qui est exactement l'inverse de ce qu'il
// doit montrer. Il en va de même de l'onde d'une déflagration, qui recouvre par
// définition ce qu'elle emporte.
//
// **C'est pourquoi ils sont des quads d'écran et non des entités.** Ce qui ne se
// trie pas n'a pas de profondeur à porter, et une entité se suivrait par une
// poignée qu'il faudrait tenir d'une image à l'autre pour une chose qui ne vit
// qu'une image. La couche s'efface à chaque début de passe, ce qui est
// exactement leur durée de vie.
//
// **Le bassin ne porte qu'un point et un décompte** ; la volée, ses directions
// et sa parabole se dérivent du rang de chaque éclat. C'est ce qui permet à
// huit éclats de ne rien coûter à la simulation, et de changer d'apparence sans
// qu'elle bouge.
func (s *Screen) poserEffets() {
	effets := s.monde.Fxs()
	for i := range effets.Active() {
		e := effets.At(i)
		age := e.Total - e.Life
		switch e.Kind {
		case game.FxBlast:
			s.poserAuSol(objetSouffle, e.Life, e.X, e.Y)
		case game.FxBreak:
			s.poserVolee(s.objets.cassable(e.Object).Shards, e.X, e.Y, age, e.Total)
		case game.FxDamage:
			s.poserChiffre(e, age)
		}
	}
}

// poserAuSol pose l'image d'un effet au pixel où son point du monde tombe.
func (s *Screen) poserAuSol(nom string, reste game.Tick, x, y game.Fixed) {
	catalogue, i, montre := s.objets.effet(nom, reste)
	if !montre {
		return
	}
	aspect, porte := catalogue.image(i)
	if !porte {
		return
	}

	aspect.Layer = coucheEffets
	ex, ey := s.pixelDe(x, y)
	s.poserAuPoint(ex, ey, aspect)
}

// montantMarquant est le montant à partir duquel un chiffre passe en pleine
// teinte.
//
// **Trois, parce que c'est ce qu'un Quidam encaisse en entier.** En dessous, le
// coup est ordinaire et le chiffre reste discret ; au-dessus, il vient d'une
// arme lourde ou d'une build montée, et c'est ce que le joueur doit remarquer. Le
// seuil suit donc une unité du jeu plutôt qu'un nombre choisi — le jour où la
// résistance de base changera, c'est là qu'il faudra revenir.
const montantMarquant = 3

// hauteurChiffre est ce dont un chiffre s'élève sur sa vie, et departChiffre ce
// dont il part au-dessus du sol.
//
// Le départ vaut la moitié d'un sprite de créature : posé à ses pieds, le chiffre
// serait recouvert par elle ; posé au-dessus de sa tête, il flotterait sans dire
// ce qu'il compte.
const (
	hauteurChiffre = 10
	departChiffre  = 32
)

// poserChiffre pose ce qu'un coup vient de retirer, au-dessus de ce qu'il a
// touché.
//
// **Ce qui distingue deux chiffres est leur montant, jamais leur nature.** Une
// horde dense en produit des dizaines par seconde, et le chapitre 14 veut que
// l'écran chargé appartienne au joueur et aux menaces : les petits restent
// atténués, les gros passent en pleine teinte. C'est la variation que la teinte
// des critiques aurait apportée, obtenue sans mécanisme — aucun axe ni aucune
// arme ne produit de coup critique, et la conception date cette teinte du jour où
// l'un d'eux le fera.
//
// **Il monte et s'efface**, sur la même parabole que les éclats : ce qui jaillit
// se lit comme un événement, ce qui reste en place se lit comme un état.
func (s *Screen) poserChiffre(e *game.Fx, age game.Tick) {
	texte := strconv.Itoa(e.Amount)
	ex, ey := s.pixelDe(e.X, e.Y)

	// La montée est linéaire et non parabolique : un chiffre n'est pas un objet
	// qui retombe, c'est une valeur qui s'en va.
	monte := hauteurChiffre * int(age) / int(e.Total)
	x := int(ex) - s.hud.Font.Advance(texte)/2
	y := int(ey) - departChiffre - monte

	teinte := s.hud.Color("texte_attenue")
	if e.Amount >= montantMarquant {
		teinte = s.hud.Color("texte")
	}

	// Le contour d'un pixel, que le chapitre 14 exige : un chiffre jaillit
	// au-dessus de n'importe quoi — décor clair, flaque, carrelage —, et sa seule
	// couleur ne suffit pas à le détacher du fond.
	//
	// **Dans l'aire et non sur la surface**, parce que le pixel vient de
	// `pixelDe` : posé sur la surface, le chiffre se décalait de la marge et une
	// créature touchée juste hors du champ écrivait le sien à côté du jeu.
	s.hud.Font.DrawOutlinedArea(s.scene, texte, x, y, teinte, s.hud.Color("texte_contour"))
}

// poserVolee ouvre une gerbe d'éclats depuis le point où quelque chose a cédé.
//
// La hauteur se retranche à l'ordonnée d'écran : monter est aller vers le haut,
// et c'est le seul endroit du rendu où l'élévation d'une chose se calcule au
// lieu de venir d'un manifeste — un éclat n'a pas d'ancrage, il a une
// trajectoire.
func (s *Screen) poserVolee(nom string, x, y game.Fixed, age, total game.Tick) {
	catalogue, connu := s.objets.objets[nom]
	if !connu || len(catalogue.images) == 0 {
		return
	}

	for rang := range sprite.Shards {
		aspect := catalogue.images[rang%len(catalogue.images)]
		aspect.Layer = coucheEffets

		ecart, hauteur := sprite.Shard(rang, age, total)
		ex, ey := s.pixelDe(x+ecart.X, y+ecart.Y)
		s.poserAuPoint(ex, ey-pixelDEcran(hauteur), aspect)
	}
}

// milieuDuCorps est la hauteur, au-dessus du sol, où se centre ce qui marque une
// créature touchée.
//
// **Elle remplace le milieu mesuré sur la forme, et c'est une perte assumée.**
// Le relevé du moteur porte sur des boîtes et ne rend plus de masque, si bien
// qu'il n'y a plus de forme dessinée à mesurer ; la moitié d'un corps de
// quarante-cinq pixels est ce que la mesure donnait en moyenne, et l'écart ne se
// voit pas sur une gerbe de six pixels qui vit quatre images.
const milieuDuCorps = 22

// poserEtincelle marque un tir qui vient de porter.
//
// **Elle dit que le coup est parti et qu'il a touché**, là où l'éclair de la
// créature dit laquelle a été touchée : deux retours pour un même événement,
// mais qui n'apprennent pas la même chose — c'est le partage qui vaut déjà pour
// le soigneur et la soignée. Sur une créature collée au joueur, l'éclair passe
// sous le personnage et l'étincelle non, ce qui la rend d'autant plus utile là
// où la conception dit que le retour manque le plus.
//
// **Elle se pose au milieu du dessin de la créature, et non à ses pieds.** Une
// position du monde est un point au sol : centrée dessus, la gerbe s'affichait
// sous une créature de quarante-cinq pixels, là où l'œil ne suit ni le tir ni
// la cible.
//
// Une méthode pour un seul effet, et non une machinerie pour la famille : le
// souffle d'une déflagration est le second, il vient avec son bassin, et c'est
// alors qu'on saura ce que les deux partagent.
func (s *Screen) poserEtincelle(sur trace, reste game.Tick) {
	catalogue, i, montre := s.objets.effet(objetEtincelle, reste)
	if !montre {
		return
	}
	aspect, porte := catalogue.image(i)
	if !porte {
		return
	}

	aspect.Layer = coucheEffets
	s.poserAuPoint(sur.ex, sur.ey-milieuDuCorps, aspect)
}
