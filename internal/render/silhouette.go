// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le rappel de ce qui est caché : ce qu'on suit, et ce que le relevé du moteur
// en dit.

package render

import (
	"github.com/sprimault/lozengine/render"
	"github.com/sprimault/lozengine/scene"
)

// seuilDeRappel est la part du dessin qui doit être couverte pour qu'on
// rappelle, en centièmes.
//
// **Une part et non un nombre de pixels**, parce que le relevé compte désormais
// ce que la chose peint réellement : une fraction dit « la moitié du corps »
// quelle que soit la planche, là où un nombre absolu serait vrai pour une taille
// de sprite et faux pour la suivante.
//
// **La valeur se calibre en jouant et la conception ne la fixe pas.** La moitié
// du dessin est le moment où l'on cesse de reconnaître son personnage, et c'est
// l'intention que portait déjà la version en pixels.
//
// **Celle-là valait 2 048 et était inatteignable**, ce qui n'était pas un défaut
// de calibrage mais d'échelle : elle comptait des pixels de boîte — des
// rectangles cumulés, transparents compris — quand une case de personnage de
// cohue n'en peint que 954 au plus. Un seuil réglé sur l'une ne vaut rien pour
// l'autre, et c'est le moteur qui porte les deux mesures côte à côte pour cette
// raison.
const seuilDeRappel = 50

// voileDuRappel est ce qui reste de l'aplat quand c'est le décor seul qui cache,
// en fraction de 255.
//
// **Il tient entre deux lectures fausses, et les deux ont été vues à l'image.**
// À 255 l'aplat efface le mur, et l'enceinte d'un lieu se lit comme une sortie
// de carte. À zéro il ne reste que le contour, et le mur qu'on voit à
// l'intérieur du tracé referme la forme en trou : le personnage se lit *dans* la
// maçonnerie. Entre les deux, la silhouette est une surface et la maçonnerie
// reste lisible au travers.
//
// **Une valeur à juger à l'œil**, comme `intensiteEclair` : ce qui la fixe est
// qu'on voie l'appareil du mur sans hésiter sur le contour du personnage.
const voileDuRappel = 110

// contourAutour sont les huit décalages d'un pixel qui font le liseré permanent.
//
// **Huit et non quatre** : un sprite isométrique descend en escalier, et un
// voisinage en croix laisse les diagonales à nu un pixel sur deux.
var contourAutour = [8][2]int16{
	{-1, -1}, {0, -1}, {1, -1},
	{-1, 0}, {1, 0},
	{-1, 1}, {0, 1}, {1, 1},
}

// trace est ce qu'une chose posée laisse de quoi reposer son rappel : les deux
// formes qu'il peut prendre, et le pixel où elle s'est posée.
//
// **Elle ne porte plus rien de ce qui mesurait.** Le masque, la profondeur et le
// drapeau de ce qui cache existaient pour confronter les choses entre elles ; le
// relevé du moteur le fait, et il ne reste que ce dont le rappel a besoin pour
// se dessiner.
type trace struct {
	forme, bord scene.Aspect
	ex, ey      int32
}

// revele est une chose suivie dont on redessine le rappel quand elle est cachée.
//
// **Ce qui mesurait vit désormais dans le moteur.** Cohue tenait un masque par
// pixel, relevait le décor visible et confrontait les profondeurs pour savoir ce
// qui recouvrait quoi — trois cents lignes qui refaisaient ce que le tri calcule
// déjà pour décider si une chose marquée se redessine. `Entity.Measure` le
// demande, `Overlapped` le rend, et il ne reste ici que ce que le moteur ne peut
// pas savoir : sous quelle forme on rappelle.
//
// La poignée est le lien : le relevé la porte, et c'est elle que la pose a
// rendue.
type revele struct {
	poignee scene.Handle
	trace
	teinte render.Color
}

// Overlaps reçoit le relevé du pas précédent.
//
// **La tranche ne vaut que pendant l'appel**, et le moteur le dit : ce qui est
// gardé est une copie, dans une tranche réutilisée avec `[:0]` pour qu'une image
// n'alloue pas.
//
// **Le décalage d'une image est sans conséquence ici et il faut le savoir.** Un
// recouvrement ne se connaît qu'après un tri, donc le relevé le plus récent est
// celui de l'image d'avant ; à soixante images par seconde, une silhouette qui
// apparaît un seizième de seconde plus tard ne se distingue pas. Ce serait une
// autre affaire pour une décision de jeu, et c'est pourquoi aucune n'en dépend.
func (s *Screen) Overlaps(list render.OverlapList) {
	s.recouvrements = append(s.recouvrements[:0], list...)
}

// reveler repose en couche haute le rappel de ce que quelque chose recouvre.
//
// **La forme du rappel dépend de ce qui cache, et c'est `CoveredByDecor` qui le
// dit.** Un aplat quand une créature s'interpose : il faut suivre le personnage
// dans la mêlée, et une foule ne porte aucun indice de profondeur à préserver.
// Un aplat voilé **et** son contour quand le décor seul recouvre, le voile
// laissant voir la maçonnerie et le contour épinglant le bord.
//
// **Les deux formes pures ont été essayées et échouent chacune d'un côté**,
// c'est une partie jouée qui l'a montré et une planche qui le reproduit — voir
// la vue `mur-sud-est`. L'aplat plein efface le mur, et l'enceinte d'un lieu se
// lit comme une sortie de carte. Le contour seul laisse voir le mur à
// l'intérieur du tracé, et l'œil referme la forme en trou : le personnage se lit
// *dans* la maçonnerie, ce qui est précisément la lecture que ce contour
// existait pour éviter. `voileDuRappel` dit ce qui les sépare.
//
// **L'ordre de pose décide, et il est le bon sens** : le voile d'abord, le
// contour par-dessus. Les deux partagent leur couche, donc ce qui est posé après
// recouvre ce qui précède.
//
// **Le seuil est celui de cohue, et le moteur le veut ainsi** : il rend la
// mesure et laisse l'hôte choisir son seuil comme la forme de son rappel. Voir
// `seuilDeRappel` pour ce qui le fixe.
func (s *Screen) reveler() {
	for _, r := range s.aRevoir {
		o, releve := s.recouvrementDe(r.poignee)
		if !releve || o.Painted == 0 || o.Covered*100 < o.Painted*seuilDeRappel {
			continue
		}

		if o.CoveredByDecor == o.Covered && r.bord.Width != 0 {
			s.poserRappel(r, r.forme, voiler(r.teinte, voileDuRappel))
			s.poserRappel(r, r.bord, r.teinte)
			continue
		}
		s.poserRappel(r, r.forme, r.teinte)
	}
}

// voiler rend une teinte de planche à l'opacité demandée.
//
// **Ce n'est pas `Color.Scale`, et l'écart est un défaut du moteur.** Sa godoc
// en fait « la seule façon d'atténuer » une couleur, et elle met les quatre
// composantes à l'échelle, alpha compris — ce qui est juste pour une couleur
// prémultipliée, et ce que fait le chemin de l'aplat uni. Mais le rastériseur
// module un quad qui lit une planche par `product(product(src, teinte),
// teinte.A)` : il applique l'alpha de la teinte une **seconde** fois aux
// composantes de couleur, déjà prémultipliées. Le voile assombrit alors au lieu
// d'éclaircir — relevé sur la planche `mur-sud-est`, 96 de luminance dans le
// corps contre 116 pour le mur autour, là où un crème à 43 % doit en rendre 171.
//
// Sur ce chemin, les composantes de couleur sont donc un filtre et l'alpha
// l'opacité. La teinte de base étant opaque, sa forme prémultipliée est sa forme
// droite, et poser l'alpha suffit. **Le jour où le moteur accorde les deux
// chemins, cette fonction disparaît au profit de `Scale`.**
func voiler(c render.Color, opacite uint8) render.Color {
	c.A = opacite
	return c
}

// poserRappel pose un quad de rappel au-dessus de ce qui cachait.
//
// **Il quitte la mesure, et sans cela il ne s'éteint plus.** Posé en couche
// haute, il recouvre ce qui cachait : le relevé suivant le compterait comme
// cachant la chose qu'il rappelle, dépasserait le seuil qui vient de le
// déclencher, et s'entretiendrait seul. C'est le défaut qui laissait le joueur
// en aplat blanc dès qu'une caisse l'avait frôlé une fois. `Aspect.Uncounted`
// l'exclut à la source, et c'est la voie que le moteur prévoit pour cela :
// retrancher l'aire du quad de la mesure marcherait aujourd'hui, mais sa godoc
// l'interdit explicitement — la loi additive sur laquelle cette soustraction
// repose n'est pas garantie d'une version à l'autre.
func (s *Screen) poserRappel(r revele, a scene.Aspect, teinte render.Color) {
	a.Tint = teinte
	a.Layer = coucheSilhouette
	a.Uncounted = true
	s.poserAuPoint(r.ex, r.ey, a)
}

// recouvrementDe cherche ce que le relevé dit d'une poignée.
//
// Un parcours et non une table : le relevé ne porte que ce qu'on a marqué — le
// joueur et les projectiles de la horde —, donc quelques dizaines d'entrées, et
// bâtir une table par image coûterait l'allocation que ce chemin n'a pas le
// droit de faire.
func (s *Screen) recouvrementDe(p scene.Handle) (render.Overlap, bool) {
	for _, o := range s.recouvrements {
		if o.Identity == uint32(p) {
			return o, true
		}
	}
	return render.Overlap{}, false
}
