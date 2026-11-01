// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La pose du décor dans la scène : les strates où il s'empile, le moment où il
// s'y range, ce qui s'y pose en entité pour cacher ce qui passe derrière, et la
// seule case dont l'état de partie change l'apparence.

package render

import (
	"fmt"

	"github.com/sprimault/lozengine/render"
	"github.com/sprimault/lozengine/scene"

	"github.com/sprimault/cohue/internal/game"
)

// Les strates où une cellule empile ce qu'elle porte.
//
// **Une strate dit l'empilement local, l'étage dit le niveau de décor.** La
// première ne passe devant que ce qu'elle recouvre, le second range toute une
// hauteur ; les confondre ferait d'un trottoir une couche d'interface.
const (
	// strateComblement porte le sol du thème sous ce qui ne remplit pas son
	// losange.
	strateComblement uint8 = iota
	// strateTerrain porte le revêtement que la grille déclare.
	strateTerrain
	// strateCouche est la première des couches du format de pièces, qui
	// s'empilent dans leur ordre.
	strateCouche
)

// La porte, dans ses deux états. Elle teinte le décor de sa case au lieu de le
// remplacer, parce qu'elle n'y change rien : fermée elle est un mur comme un
// autre, ouverte elle l'est encore. Ce que ces deux teintes disent est ce
// qu'aucune forme ne porte — le lieu est gagné, et la sortie est là.
//
// **Une seule teinte, éteinte puis vive**, plutôt que deux couleurs : c'est le
// même objet dans deux états, et l'écart de valeur se voit en périphérie là où
// un changement de teinte demande de regarder. Un cyan, parce que le vert
// appartient à l'éclair de soin et qu'ouvrir une porte n'a rien à voir avec
// lui — le voisin le plus proche est la gemme, à 28,5, et elle est au sol quand
// la porte est sur un mur.
//
// **Deux écarts de valeur à tenir, et le premier jet n'en a tenu qu'un.**
// L'écart entre les deux états était bon ; celui entre la porte fermée et le mur
// qui l'entoure ne l'était pas, si bien qu'une partie jouée n'a pas trouvé la
// porte en faisant les quatre coins de la salle. Une teinte ne se choisit pas
// contre l'autre état du même objet, elle se choisit **aussi** contre ce qui
// l'entoure — ici le mur de l'enceinte, sur lequel la porte est posée et que ces
// deux teintes multiplient.
var (
	porteFermee  = render.Premultiply(48, 132, 152, 255)
	porteOuverte = render.Premultiply(120, 232, 248, 255)
)

// strateNappe rend la strate où le lit d'une zone se pose, au-dessus de tout ce
// que le lieu empile.
//
// **Ce sont les deux seules strates relues par image**, parce qu'une zone naît
// et meurt ; les autres restent posées depuis le montage, et c'est ce qui permet
// au décor de ne plus se relever soixante fois par seconde.
//
// **Deux et non une, parce que le feu s'empile sur lui-même** : son lit de
// braise d'abord, ses langues par-dessus, qui montent de la braise. Les poser
// sur la même strate ferait que la seconde efface la première, une cellule ne
// portant qu'un aspect par strate.
func strateNappe(sol *Terrain) uint8 {
	// #nosec G115 -- le nombre de couches est borné par `MaxTiers` au montage
	return strateCouche + uint8(sol.Couches())
}

// strateFlammes rend la strate des langues de feu, juste au-dessus du lit.
func strateFlammes(sol *Terrain) uint8 {
	return strateNappe(sol) + 1
}

// poserLaNappe repose le marquage des zones sur les cases que l'image montre.
//
// **Elle efface autant qu'elle pose**, et c'est ce qu'une strate persistante
// impose : une case dont la zone a disparu garderait son aspect sans cela. La
// valeur zéro ne montrant rien, l'effacement est la pose de cette valeur — il n'y
// a donc pas deux chemins à tenir d'accord.
//
// **Bornée à ce que l'image montre**, parce que c'est la seule strate dont le
// coût suive le temps et non le lieu : un lieu de neuf mille cases en a un
// millier à l'écran, et le marquage ne se voit que là.
func (s *Screen) poserLaNappe() {
	u0, v0, u1, v1 := s.cam.casesVisibles()
	largeur, hauteur := s.sol.carte.Width(), s.sol.carte.Height()

	for v := max(v0, 0); v <= min(v1, hauteur-1); v++ {
		for u := max(u0, 0); u <= min(u1, largeur-1); u++ {
			lit, langues := s.nappeDe(u, v, largeur)
			s.scene.PlaceCell(u, v, strateNappe(s.sol), lit)
			s.scene.PlaceCell(u, v, strateFlammes(s.sol), langues)
		}
	}
}

// nappeDe rend ce qu'une case porte d'une zone : son lit, et ses langues quand
// elle brûle.
//
// Le télégraphe d'une explosion suit la surface qu'il marque, donc il monte de
// la hauteur du sol : une forme qui remplit son losange *est* cette surface, et
// le marquage se pose dessus ; une forme qui ne la remplit pas la traverse, et
// c'est le sol du thème, sous elle, qu'on marche.
func (s *Screen) nappeDe(u, v, largeur int) (lit, langues scene.Aspect) {
	n := s.emprises[v*largeur+u]
	if n.vif.A == 0 {
		return scene.Aspect{}, scene.Aspect{}
	}
	f, _ := s.sol.formeDe(u, v)

	if !n.feu {
		return s.losange(n.vif, f.hauteurSol), scene.Aspect{}
	}

	// Le lit d'abord, les langues par-dessus : elles montent de la braise.
	lit = s.losange(teinteBraise.Scale(n.vif.A), f.hauteurSol)
	catalogue, i, montre := s.objets.image(objetFlammes, s.monde.Tick(), u*7+v*3)
	if !montre {
		return lit, scene.Aspect{}
	}
	aspect, porte := catalogue.image(i)
	if !porte {
		return lit, scene.Aspect{}
	}
	langues = centrerSurLaCase(aspect, s.sol.TileSize())
	langues.Tint = n.vif
	langues.Floor = etageSol
	langues.AnchorY -= hauteurEnAncre(f.hauteurSol)
	return lit, langues
}

// losange rend l'aspect d'une case pleine, teintée et posée sur la surface
// qu'on y foule.
//
// C'est le seul dessin que le rendu calcule — `face` le compose au montage —
// parce qu'un aplat du moteur peint un rectangle quand une case est un losange.
func (s *Screen) losange(teinte render.Color, hauteurSol int) scene.Aspect {
	tuile := s.sol.TileSize()
	// Le losange a la taille de la tuile et se pose depuis le sommet nord de sa
	// case ; le centre que `PlaceCell` vise est une demi-hauteur plus bas, et la
	// surface foulée autant de pixels plus haut.
	ancre, tient := ancreDEcran(0, tuile[1]/2-hauteurSol)
	if !tient {
		return scene.Aspect{}
	}
	return scene.Aspect{
		Sheet: s.face,
		// #nosec G115 -- la taille de tuile vient du manifeste, confrontée à la
		// planche du losange au montage
		Width: uint16(tuile[0]),
		// #nosec G115 -- même borne
		Height:  uint16(tuile[1]),
		Tint:    teinte,
		AnchorX: ancre[0],
		AnchorY: ancre[1],
		Floor:   etageSol,
	}
}

// hauteurEnAncre réduit une hauteur de sol au type d'une ancre.
//
// Elle est bornée par l'élévation qu'un manifeste déclare, que le générateur
// tient sous le plafond d'un obstacle bas pour ce qui se marche.
func hauteurEnAncre(hauteurSol int) int16 {
	// #nosec G115 -- bornée par l'élévation d'une forme, de l'ordre de la dizaine
	return int16(hauteurSol)
}

// poserLeDecor range le décor d'un lieu dans la scène, une fois pour toutes.
//
// **Une fois et non par image, et c'est tout le gain du portage.** La scène
// garde ce qu'on y pose et ne relit que la fenêtre visible : les deux boucles
// qui balayaient un millier de cases à chaque image disparaissent, et ce qui se
// repose ensuite est ce qui change — la porte quand son état bascule.
//
// Le refus porte sur le nombre de strates et non sur le nombre de cases : un
// lieu profond passerait la borne du moteur sans qu'aucune case soit en cause,
// et le message doit nommer ce qu'il faut changer.
func poserLeDecor(sc *scene.Scene, sol *Terrain, monde *game.World, carte *game.CostGrid) error {
	if besoin := int(strateCouche) + sol.Couches() + 1; besoin > scene.MaxTiers {
		return fmt.Errorf("decor : %d strates pour %d couches, le moteur en porte %d",
			besoin, sol.Couches(), scene.MaxTiers)
	}

	for v := range carte.Height() {
		for u := range carte.Width() {
			poserLaCase(sc, sol, monde, u, v)
		}
	}
	return nil
}

// poserLaCase range dans la scène ce qu'une case du décor empile.
//
// **Le comblement précède la forme sur la même cellule**, et c'est la strate qui
// l'ordonne au lieu de l'ordre des appels : une forme qui ne remplit pas son
// losange laisserait voir le fond sans lui.
//
// **Un obstacle bas ne prend pas de cellule et laisse le comblement seul**,
// parce que `poserLesObstaclesBas` le pose en entité. Sa case serait sinon
// peinte deux fois, et ses pixels translucides s'assombriraient de leur propre
// double.
//
// **Une couche ne comble pas, et c'est tout l'objet des couches.** Ce qu'il y a
// dessous est déjà posé : y remettre le sol du thème effacerait la chaussée
// avant d'y poser le véhicule, ce qui est exactement le défaut que le calque
// existe pour fermer. Un obstacle bas d'une couche laisse donc sa strate vide et
// rien d'autre — ce qui est sous lui est déjà peint.
func poserLaCase(sc *scene.Scene, sol *Terrain, monde *game.World, u, v int) {
	f, posee := sol.formeDe(u, v)
	if !posee {
		return
	}

	if (f.nue || f.bas) && sol.sol != nil {
		sc.PlaceCell(u, v, strateComblement, sol.sol.aspect)
	}
	if !f.bas {
		sc.PlaceCell(u, v, strateTerrain, teinterLaPorte(f.aspect, monde, u, v))
	}

	for couche := range sol.Couches() {
		g, posee := sol.coucheDe(couche, u, v)
		if !posee || g.bas {
			continue
		}
		// #nosec G115 -- le nombre de couches est borné par `MaxTiers` au montage
		sc.PlaceCell(u, v, strateCouche+uint8(couche), g.aspect)
	}
}

// obstacleBas est un obstacle du lieu que le rendu pose en entité : sa case, son
// dessin, et la poignée sous laquelle la scène le connaît quand l'image le
// montre.
//
// **La liste des obstacles plutôt qu'une table par case**, et c'est ce qui rend
// la passe bon marché : le lieu livré en porte deux cent vingt-cinq pour les
// neuf mille six cents cases de sa carte, donc le balayage suit ce qui existe au
// lieu de suivre la carte. Elle porte aussi la poignée, si bien qu'aucune
// seconde tranche n'a à rester d'accord avec elle.
type obstacleBas struct {
	u, v   int
	aspect scene.Aspect
	// poignee est zéro quand l'obstacle n'est pas dans la scène, les générations
	// du moteur partant à un.
	poignee scene.Handle
}

// entite rend ce qu'une pose dans la scène reçoit : le centre de la case, et
// l'image telle quelle.
//
// **Le centre, parce que c'est là que le moteur projette une cellule**, et c'est
// ce qui garantit que la bascule en entité ne déplace pas le dessin d'un pixel.
// Ce qui change est la profondeur : une cellule recule du sien d'une cellule de
// somme, une entité porte le point qu'on lui donne.
//
// **La hauteur reste nulle et l'étage vient de l'aspect.** Une cellule n'a pas
// d'autre champ où porter le sien, donc le moteur la projette à la hauteur de son
// étage ; une entité somme les deux, et ajouter ici l'un ou l'autre lèverait le
// décor au-dessus du sol qu'il touche.
func (o obstacleBas) entite(a scene.Aspect) scene.Entity {
	return scene.Entity{
		X:      float64(o.u) + 0.5,
		Y:      float64(o.v) + 0.5,
		Aspect: a,
	}
}

// releverLesObstaclesBas énumère les obstacles bas du lieu, terrain et couches
// confondus.
//
// Une fois au montage : ce qu'une case porte ne change pas en cours de partie, et
// seule la fenêtre visible le fait.
func releverLesObstaclesBas(sol *Terrain) []obstacleBas {
	var bas []obstacleBas
	for v := range sol.carte.Height() {
		for u := range sol.carte.Width() {
			if f, posee := sol.formeDe(u, v); posee && f.bas {
				bas = append(bas, obstacleBas{u: u, v: v, aspect: f.aspect})
			}
			for couche := range sol.Couches() {
				if g, posee := sol.coucheDe(couche, u, v); posee && g.bas {
					bas = append(bas, obstacleBas{u: u, v: v, aspect: g.aspect})
				}
			}
		}
	}
	return bas
}

// poserLesObstaclesBas pose en entités les obstacles bas que l'image montre, et
// retire ceux qu'elle ne montre plus.
//
// **Une entité pour que la case voisine ne passe pas devant.** Une cellule se
// trie à la profondeur de son coin le plus lointain, ce qui la met derrière tout
// ce que la case du nord contient : une créature qui s'abrite derrière un muret
// se peignait par-dessus la maçonnerie, et se lisait comme debout dessus. Posée
// au centre de sa case, l'entité porte la profondeur de ce centre, qui départage
// les quatre voisines dans le bon sens — le dessin ne bouge pas d'un pixel, le
// moteur projetant une cellule depuis ce même centre.
//
// **Bornée à la fenêtre, parce que c'est ce qui borne la scène.** Le décor en
// cellules ne coûte rien à poser une fois pour toutes, là qu'une entité occupe
// une place de la capacité tant qu'elle est là : la suivre à la fenêtre fait que
// le nombre d'entités de décor suit l'écran et jamais la taille du lieu, et
// qu'un lieu tiers chargé de murets ne prend pas la place du joueur.
//
// **Les obstacles bas comptent dans la mesure de recouvrement comme des
// entités, et c'est le prix de cette voie.** `render.Overlap` sépare ce que le
// décor cache de ce que les entités cachent par la voie de pose, qu'un hôte ne
// peut pas déclarer : un joueur qu'un mur et un muret cachent ensemble reçoit
// donc l'aplat du rappel plutôt que son voile. Les marquer `Uncounted` fermerait
// ce cas et en ouvrirait un pire — le muret ne compterait plus du tout, et le
// seuil de rappel ne serait plus atteint pour un joueur réellement caché.
func (s *Screen) poserLesObstaclesBas() {
	u0, v0, u1, v1 := s.cam.casesVisibles()

	for i := range s.bas {
		o := &s.bas[i]
		if o.u < u0 || o.u > u1 || o.v < v0 || o.v > v1 {
			if o.poignee != 0 {
				s.scene.Remove(o.poignee)
				o.poignee = 0
			}
			continue
		}

		// La sortie peut tomber sur une case d'obstacle bas — rien ne l'interdit à
		// un lieu, et la teinte est ce qui la signale au joueur.
		aspect := teinterLaPorte(o.aspect, s.monde, o.u, o.v)
		if o.poignee != 0 {
			s.scene.SetAspect(o.poignee, aspect)
			continue
		}
		if p, pose := s.scene.Place(o.entite(aspect)); pose {
			o.poignee = p
		}
	}
}

// teinterLaPorte donne à la case de sortie la teinte de son état, et laisse tout
// le reste intact.
//
// Aucune forme du lieu ne dit qu'une case est la sortie — l'ouverture est un état
// de partie, et le décor n'en sait rien. La teinte multiplie donc l'image au lieu
// de la remplacer : le mur reste un mur, et il vire au cyan.
func teinterLaPorte(a scene.Aspect, monde *game.World, u, v int) scene.Aspect {
	sortie := monde.Exit()
	if sortie == nil || sortie.U != u || sortie.V != v {
		return a
	}
	if monde.DoorOpen() {
		a.Tint = porteOuverte
		return a
	}
	a.Tint = porteFermee
	return a
}
