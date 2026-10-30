// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le tampon interne et ce qu'on y peint : le sol tel que la simulation le voit,
// ce qui s'y tient, et les touches qui déplacent le joueur. C'est ici que le
// repère de l'écran rencontre celui du monde ; l'ordre de dessin est à côté.

// Package render dessine ce que la simulation a calculé, et ne décide de rien.
//
// La frontière avec `internal/game` va dans un seul sens : ce paquet lit l'état
// du monde, et rien de ce qu'il produit n'y revient. C'est ce qui lui permet de
// calculer en flottants — interpolations, lissage de caméra — sans menacer le
// déterminisme de la run, et ce qui interdit à une notion d'écran de redescendre
// dans la simulation.
//
// **Ce paquet n'a pas de fichier de test, et la contrainte qui l'imposait est
// tombée.** Importer Ebitengine initialisait GLFW, qui panique sans `DISPLAY` :
// sur un runner sans écran, n'importe quel test du paquet tombait avant d'avoir
// commencé, y compris un test qui n'ouvrait aucune fenêtre. `lozengine` calcule
// ses images sans fenêtre, donc plus rien n'empêche d'en écrire.
//
// Ce qui reste vrai est que le rendu se juge à l'œil, et que ce qui doit être
// vérifié mécaniquement vit du côté simulation. La doctrine qui l'énonce est
// dans `docs/go.md` ; c'est elle qui dira ce qui la remplace, et ce n'est pas
// cette godoc qui le décide.
package render

import (
	"fmt"
	"math"

	"github.com/sprimault/lozengine/geometry"
	"github.com/sprimault/lozengine/input"
	"github.com/sprimault/lozengine/raster"
	"github.com/sprimault/lozengine/render"
	"github.com/sprimault/lozengine/scene"

	"github.com/sprimault/cohue/internal/game"
)

// Le plancher de la surface de dessin, en pixels.
//
// Une surface de 480×270 ne montrerait que sept tuiles de large, bien trop serré
// pour voir la horde arriver ; 960×540 en donne une quinzaine et se multiplie
// par deux pour du 1080p.
//
// **Un plancher et non une taille fixe**, et c'est ce que le moteur a changé :
// la surface suit la fenêtre, au facteur entier près, si bien qu'il ne reste
// aucune bande. Ce qui en est retiré est la bande du bandeau — `HUD.Inset` la
// déclare —, et l'aire de jeu prend tout le reste : agrandir la fenêtre montre
// donc un peu plus de lieu, ce qui est l'arbitrage jugé à l'écran.
//
// Ce qui ne bouge pas est le **grain** : le facteur d'agrandissement reste
// entier, donc une tuile couvre toujours un nombre entier de pixels de fenêtre,
// et c'est là qu'est la règle de pixel art — pas dans une aire de taille fixe.
//
// Rien ne se place plus contre ces deux nombres. Une pièce d'interface lit
// `Scene.Surface` pour son étendue et `Scene.WorldOrigin` pour s'accrocher au
// bord de l'aire ; les y remplacer par une constante poserait juste à
// l'ouverture et faux partout ailleurs, sans que rien ne le signale.
const (
	Width  = 960
	Height = 540
)

// aireDans rend le rectangle de l'aire de jeu dans la surface : son coin
// haut-gauche, puis sa taille.
//
// C'est la référence de tout ce que l'interface accroche au jeu — le bandeau
// au-dessus, le panneau des cartes en dessous, la vignette sur son bord. **Elle
// forme un bloc avec l'aire** plutôt que de se coller aux bords de la fenêtre :
// sur un grand écran, un bandeau au bord supérieur serait à trente centimètres
// de ce qu'on regarde, et la marge entre les deux se lirait comme un défaut.
//
// Le moteur rend l'origine et la taille séparément parce qu'un jeu qui ne borne
// pas son aire n'a besoin que de la seconde ; ici les deux comptent.
func aireDans(sc *scene.Scene) (x, y, largeur, hauteur int) {
	ox, oy := sc.WorldOrigin()
	_, _, w, h := sc.View()
	return int(ox), int(oy), int(w), int(h)
}

// Background est le fond du tampon, sous tout ce que la scène décrit.
//
// **Exporté parce que c'est le moteur qui l'efface désormais.** Le rendu
// peignait le tampon avant de dessiner ; la boucle le fait, et le fond est un
// réglage qu'elle reçoit au montage. Il reste déclaré ici — c'est une décision
// d'apparence, au même titre que les teintes d'en dessous — et c'est l'appelant
// qui le passe.
//
// Il se voit là où aucune case n'est posée : le hors-carte d'un lieu plus petit
// que l'écran, et la bordure pendant un cadrage.
var Background = render.Premultiply(24, 24, 28, 255)

// La vignette de danger : jusqu'où elle mord sur l'écran, et en combien de
// paliers d'opacité.
//
// Quarante-huit pixels font une tuile et demie de haut, assez pour se voir en
// périphérie sans entrer dans la zone où l'on suit son personnage. Ces deux
// nombres sont des pixels du tampon et n'appartiennent donc qu'au rendu — le
// manifeste porte la teinte, qui est un choix d'apparence, pas cette géométrie
// qui découle de la taille du tampon.
const (
	epaisseurVignette = 48
	paliersVignette   = 6
)

// poulsAnnonce est la demi-période du battement qui annonce une charge, en
// ticks.
//
// Six donne cinq alternances sur une anticipation d'une demi-seconde : assez
// lent pour se lire comme un signal et non comme un scintillement, assez rapide
// pour qu'on n'attende pas le second battement avant de réagir. Il vit ici
// parce qu'il est une cadence d'affichage — la durée qu'il découpe, elle,
// appartient au profil.
const poulsAnnonce = 6

// emprisePlancher est l'opacité de l'emprise d'une explosion au moment où elle
// s'amorce, en fraction de sa teinte pleine.
//
// Elle ne part pas de zéro : ce que la marque dit d'abord est « ne reste pas
// là », et une zone invisible à sa première image ne le dirait qu'une fois trop
// tard. Ce qui monte ensuite est le temps qui manque, pas l'existence du danger.
const emprisePlancher = 0.45

// Les teintes du rendu provisoire, qui tiendront jusqu'à ce que les créatures
// aient leurs sprites.
//
// Elles ne cherchent pas à ressembler à un lieu : ce sont des rôles de créature,
// choisis pour se distinguer et pour rien d'autre. La borne de couleurs vaut
// pour les images, qui sortent des générateurs, et non pour ces aplats qui
// disparaîtront avec elles.
//
// **Deux axes, et ils ne servent pas la même question.** La valeur sépare ce
// qu'on est de ce qu'on affronte — le joueur clair, la horde sombre —, et c'est
// elle qui tient à cent créatures à l'écran. La teinte sépare les rôles entre
// eux, ce qui ne se pose qu'une fois la première séparation acquise. Les
// confondre donnerait un profil plus visible que les autres, décision de
// conception que personne n'a prise.
var (
	// intact est l'élément neutre de la multiplication qu'une teinte applique :
	// une image posée sous cette échelle sort telle qu'elle est dessinée. C'est
	// de quoi partir quand on veut retirer de la lumière à un sprite sans avoir
	// à en connaître les couleurs.
	intact = render.White()

	// L'emprise d'une explosion se pose sur le sol et doit rester lisible sous
	// les corps qui la traversent : une teinte chaude que ni la horde ni le
	// joueur ne portent, et translucide **prémultipliée par son alpha** — écrite
	// à plat, elle rendrait un aplat trois fois trop dense.
	teinteEmprise = render.Color{R: 108, G: 54, B: 24, A: 132}
	// Le lit de braise sous les langues de feu, à la teinte de leur pied.
	//
	// **Il existe parce qu'un dessin par case ne joint pas ses voisines.** Les
	// langues sont posées au centre d'une case et débordent, mais leur silhouette
	// laisse passer le sol entre deux cases : la zone se lisait en tas alignés,
	// avec la grille visible dessous. Le losange, lui, est exactement la case, si
	// bien que deux voisines se touchent sans recouvrement.
	//
	// **Sombre, et c'est ce qui le sépare de l'aplat qu'il remplace.** Une nappe
	// orange se lit comme un tapis posé sur le sol ; des braises sous des flammes
	// se lisent comme ce qui brûle, et ce sont les langues qui portent la couleur.
	teinteBraise = render.Color{R: 96, G: 20, B: 14, A: 190}
	// **L'éclair d'une créature touchée est la même teinte, éclaircie**, et non
	// une couleur nouvelle : ce qu'il doit dire est « celle-ci vient d'être
	// atteinte », pas « ceci est autre chose ». Un blanc franc ferait clignoter
	// une foule dense en bandes qu'on ne relierait plus à des créatures.
	teinteImpact = render.Premultiply(236, 186, 178, 255)
	// **L'annonce d'une charge bat, elle ne se contente pas d'une couleur.** Une
	// teinte fixe se perd dans une foule qui porte déjà le rouge, et c'est
	// justement au milieu de la foule qu'il faut la repérer pour mettre un
	// obstacle entre soi et elle. Le battement se lit là où un aplat de plus ne
	// se lirait pas, et il dit l'imminence plutôt que l'état.
	//
	// La course, elle, n'est pas peinte : elle va vite et droit, ce qui est déjà
	// à l'image. Un signal ne redit pas ce qu'on voit.
	teinteAnnonce = render.Premultiply(232, 96, 72, 255)
	// **Le soigneur tranche, la soignée nuance.** Les deux disent le même
	// événement mais n'apprennent pas la même chose : qu'une créature récupère
	// explique pourquoi elle ne tombe pas, et cela peut rester dans la famille du
	// rouge ; savoir *lequel* d'une horde identique est le Secouriste change la
	// conduite du joueur, et cela doit sauter aux yeux là où tout est rouge.
	//
	// D'où un vert franc pour l'un et un rouge éclairci vers le vert pour
	// l'autre : les relier tient à la teinte commune, les séparer à la distance
	// qui les sépare du fond de la horde.
	teinteSoigneur = render.Premultiply(120, 226, 132, 255)
	teinteSoigne   = render.Premultiply(172, 178, 108, 255)

	// **Le contour du joueur, et sa silhouette : la même teinte.** C'est la même
	// chose qui le détache — cernée quand on le voit, pleine quand une forme ou
	// une foule le recouvre —, et deux couleurs pour un seul rôle donneraient
	// deux réglages à tenir d'accord.
	//
	// Un blanc à peine tiède plutôt que le blanc pur : le tir du joueur est déjà
	// un blanc bleuté, et le personnage doit s'en distinguer autant que du reste.
	// La luminance reste au-delà de 240, ce que la mesure demande — le pixel le
	// plus clair d'un profil monte à 162, celui d'un figurant à 175.
	teinteContour = render.Premultiply(252, 246, 226, 255)

	// **Le contour d'une icône posée dans le monde, et il va dans l'autre
	// sens.** Celui du joueur éclaircit parce qu'il cerne une silhouette sombre
	// dans une horde sombre ; une icône d'arme est faite de gris clairs, dessinée
	// pour être lue dans une case de bandeau à 28 de luminance. Posée sur un sol
	// à 162 elle n'a plus que trente d'écart, ce qui est le même défaut que celui
	// qui avait fait redessiner la grenade : une valeur juste, relevée contre un
	// fond qui n'est pas celui où elle s'affiche.
	//
	// Le presque-noir est celui que `outils/objets.py` donne aux objets du monde,
	// pour que ce qui cerne dans la scène cerne partout de la même façon.
	teinteContourSombre = render.Premultiply(24, 24, 28, 255)

	// **Le projectile de la horde porte une teinte qu'aucune autre ne dispute.**
	// C'est entre projectiles que la distinction doit être maximale : « est-ce
	// que ça me fait mal ? » ne se pose que sur eux, si bien que les confondre
	// coûte plus cher que de confondre un projectile et une créature.
	//
	// Elle ne sert plus qu'à sa silhouette, le sprite portant sa propre couleur.
	// Ce n'en est donc pas une copie : un aplat n'a pas de masse sombre ni de
	// liseré clair, et ce qu'on lui demande est de se voir à travers ce qui le
	// cache.
	teinteTirHorde = render.Premultiply(186, 138, 232, 255)

	// La teinte de l'aimant, qui ne sert plus qu'à la case du bandeau : au sol,
	// c'est son sprite qu'on voit. Elle survit parce que l'objet n'a pas d'icône
	// d'interface, et `HUD.emplacement` dit ce qu'il faudrait pour l'éteindre.
	teinteAimant = render.Premultiply(198, 126, 78, 255)
)

// Screen est le jeu tel que `lozengine` le voit : une scène qu'il décrit, et non
// une image qu'il peint.
type Screen struct {
	monde  *game.World
	carte  *game.CostGrid
	sol    *Terrain
	troupe *Cast
	objets *Stage
	cam    *camera

	// scene porte ce que l'image contient, et c'est le moteur qui la trie. Le
	// champ garde le nom du paquet sans l'ombrer : les deux vivent dans des
	// espaces distincts, et l'appeler autrement ferait chercher la scène sous un
	// autre mot.
	scene *scene.Scene

	// face est la planche du losange blanc que la nappe teinte.
	//
	// **C'est le dernier dessin que le rendu calcule, et il ne disparaîtra pas** :
	// il marque l'emprise d'une explosion, et le télégraphe s'y peint en cases
	// pleines parce qu'il n'y a rien à agrandir — redimensionner une image par une
	// fraction casserait le pixel entier. Un aplat du moteur ne le remplacerait
	// pas davantage : `render.Flat` peint un rectangle, quand une case est un
	// losange.
	face uint16

	// emprises porte, par case et le temps d'une image, la nappe qui s'y peint —
	// l'alpha nul disant qu'aucune zone ne la couvre.
	//
	// **Un relevé par case plutôt qu'une passe par explosion**, parce que le
	// marquage appartient à la case et doit suivre sa forme dans la passe qui la
	// dessine. Peint d'un bloc à la fin du sol, il passait sous le trottoir et le
	// quai, qui se trient et se peignent après lui.
	//
	// Longue de la carte, allouée au montage : la fenêtre visible s'efface au
	// début de chaque image, et rien hors d'elle ne se dessine.
	emprises []nappe

	// aRevoir porte, le temps d'une image, ce qu'on redessinera si quelque chose
	// le recouvre. Réutilisée avec `[:0]` : la passe la remplit soixante fois
	// par seconde, et une tranche qui croîtrait ferait passer le ramasse-miettes
	// là où il se voit.
	aRevoir []revele

	// attaches relient chaque bassin à la scène, indexées par sorte — qui n'est
	// plus un critère de rang mais exactement cet aiguillage.
	attaches [sorteLisere + 1]*attache

	// demiTuile est l'abscisse du sommet dans l'image d'une face, ce que le
	// manifeste appellera son ancrage quand les images viendront de lui.
	demiTuile int

	// carteChoisie est la place que la désignation occupe dans le panneau de
	// choix. Elle revient à gauche après chaque prise, pour la raison écrite sur
	// `choisir`.
	carteChoisie int

	// finRepere est le tick jusqu'auquel l'accusé d'un repère reste affiché.
	//
	// Zéro dit qu'il n'y en a pas, et rien de valide ne le produit : une pose au
	// premier tick le fixe déjà deux secondes plus loin. C'est ce qui permet de
	// ne pas tenir un second champ pour distinguer l'absence du début de partie.
	finRepere game.Tick

	// hud pose le bandeau de la partie et le texte de l'écran de fin. Il peut
	// être nul : la planche de relecture monte des écrans sans lui, et une partie
	// sans interface se dessine quand même — ce qu'elle perd est tout le texte du
	// jeu.
	hud *HUD

	// recouvrements est le relevé que le moteur a rendu pour l'image précédente,
	// recopié parce que la tranche qu'il passe ne vaut que pendant l'appel.
	// Réutilisée avec `[:0]`, comme tout ce que ce paquet remplit par image.
	recouvrements render.OverlapList
}

// WithHUD attache l'interface à un écran.
//
// Séparée du constructeur parce que tous les appelants n'en ont pas : la planche
// de relecture dessine des scènes sans interface, et l'ajouter au montage
// l'aurait obligée à charger un manifeste dont elle ne se sert pas.
func (s *Screen) WithHUD(h *HUD) *Screen {
	s.hud = h
	return s
}

// projection rend la matrice de cohue : `u` vers la droite, `v` vers la gauche,
// le sommet nord du losange en (0, 0).
//
// **Écrite à la main plutôt que prise à `Dimetric`, et le sens est la seule
// raison.** Le constructeur du cas courant fait croître la première coordonnée
// vers la **gauche** — sa godoc donne l'avance en pixels par cellule et jamais
// son signe, donc cela se mesure. Cohue pose l'est à droite depuis la première
// ligne : les noms de direction des bandes dessinées, les touches, et les lieux
// déjà écrits en dépendent tous. Prendre l'autre sens retournerait le monde en
// miroir, ce qui se voit sur chaque personnage.
//
// L'étage vaut un pixel, ce qui permet à `Entity.Z` de porter directement ceux
// que le manifeste donne — six pour un trottoir, dix pour un quai. Sans ce
// réglage, l'étage monterait d'une hauteur de cellule et la moindre bordure
// décollerait le personnage du sol.
func projection(tuile [2]int) geometry.Matrix {
	demi, moitie := float64(tuile[0])/2, float64(tuile[1])/2
	return geometry.Raw(demi, -demi, moitie, moitie).WithFloorHeight(1)
}

// NewScreen monte le rendu sur une partie et le lieu qu'elle joue.
//
// Le décor apporte sa taille de tuile, celle du manifeste et jamais une
// constante : le chargeur en exige le rapport de deux pour un, et c'est de lui
// que la projection la tient. La recevoir à part du terrain qui la porte aurait
// laissé deux appelants la prendre à deux endroits.
// **La scène se monte ici parce que c'est l'écran qui la décrit.** Sa capacité
// d'entités reste au défaut du moteur jusqu'à ce que les entités y entrent : la
// fixer maintenant serait un chiffre qu'aucune pose n'exerce, et le majorant est
// la somme des bassins, qu'on relève en les parcourant.
//
// **Le retrait est celui que les options du moteur reçoivent**, et c'est la
// seule chose à ne pas laisser diverger entre les deux : il sert ici à cadrer la
// première image, et là à dimensionner toutes les suivantes. L'arithmétique est
// celle du moteur et n'est pas recopiée — `scene.Inset.Available` est publique
// pour cela.
func NewScreen(monde *game.World, carte *game.CostGrid, sol *Terrain, troupe *Cast,
	objets *Stage, atlas *raster.Atlas, retrait scene.Inset) (*Screen, error) {
	tuile := sol.TileSize()
	sc, err := scene.New(carte.Width(), carte.Height(), projection(tuile), scene.Options{})
	if err != nil {
		return nil, fmt.Errorf("scene du lieu : %w", err)
	}

	// **La vue prend la taille de l'aire de jeu ici, et pas plus tard.** Le
	// moteur ne pose son propre cadrage que sur une scène dont la vue n'a pas de
	// taille — il recentrerait alors sur le milieu du lieu, loin du joueur —, et
	// il ne redimensionne la vue ensuite que par la boucle de `Run`, qu'un hôte
	// sans fenêtre n'emprunte pas. Une vue laissée à zéro donnerait donc un
	// cadrage faux au jeu et une aire fausse à la planche.
	sc.ResizeView(retrait.Available(Width, Height))
	losange, err := face(tuile, atlas)
	if err != nil {
		return nil, err
	}

	cam := nouvelleCamera(sc, carte)
	s := &Screen{
		monde:     monde,
		carte:     carte,
		sol:       sol,
		troupe:    troupe,
		objets:    objets,
		cam:       cam,
		scene:     sc,
		face:      losange,
		demiTuile: tuile[0] / 2,
		// Le joueur et, au pire, tous les projectiles de la horde en vol : c'est
		// exactement ce que la conception révèle, donc le majorant est le bassin
		// lui-même.
		aRevoir:       make([]revele, 0, 1+monde.EnemyShots().Cap()),
		recouvrements: make(render.OverlapList, 0, 1+monde.EnemyShots().Cap()),
		emprises:      make([]nappe, carte.Width()*carte.Height()),
	}
	s.attacher(monde)
	s.cam.suivre(monde.Player())
	if err := s.LeDecor(); err != nil {
		return nil, err
	}
	return s, nil
}

// emplacement1 est la touche qui dépense la charge du premier emplacement, celui
// de l'aimant.
//
// Les deux places du chiffre, comme les deux places d'Entrée : le pavé numérique
// tombe sous la main droite quand la gauche tient le déplacement, et un joueur
// qui appuie sur le 1 qu'il a sous les doigts n'a aucun moyen de savoir que le
// jeu écoutait l'autre.
var emplacement1 = []input.Key{input.Key1, input.KeyNumpad1}

// emplacement2 est la touche qui déclenche l'arme lourde tenue.
//
// **L'aimant garde son emplacement propre**, ce que la conception exige de lui :
// mis en concurrence avec ce qui se déclenche, il ne serait jamais gardé. Les
// deux places du chiffre pour la raison écrite au-dessus.
var emplacement2 = []input.Key{input.Key2, input.KeyNumpad2}

// emplacement3 est la touche du second emplacement d'arme lourde.
//
// Deux emplacements et non trois : la conception en fait une règle, le joueur
// ayant une décision — laquelle garder — et non une gestion.
var emplacement3 = []input.Key{input.Key3, input.KeyNumpad3}

// emplacement4 est la touche qui boit une fiole.
//
// **Une case et non deux, bien que le joueur en tienne deux.** Deux cases qui ne
// pourraient contenir que la même chose seraient un compteur de deux avec une
// interface en plus ; ce que la conception limite à deux ou trois est le nombre
// de sortes de consommables, et la seconde case arrivera avec la seconde sorte.
var emplacement4 = []input.Key{input.Key4, input.KeyNumpad4}

// interaction est la touche qui force un obstacle fragile.
//
// **Sous la main droite, que le déplacement laisse libre.** E a été essayé et
// écarté par une partie : en ZQSD l'index repose sur D, et tenir E l'en retire —
// on frappe, mais on ne décroche plus sans relâcher d'abord, au moment précis où
// la horde arrive. Maj droite se tient sans rien lâcher, et le 0 du pavé
// numérique est la seconde place, à côté des emplacements que la même main y
// tient déjà.
var interaction = []input.Key{input.KeyRightShift, input.KeyNumpad0}

// Step avance la simulation d'un pas, puis recadre.
//
// **Un pas par appel, et rien qui lise l'horloge.** La boucle du moteur appelle
// à pas fixe et **abandonne** le retard au-delà de deux pas au lieu de le
// rattraper — c'est ce que sa godoc énonce, et c'est ce qu'il faut : la
// simulation ne veut qu'« avance d'un pas », jamais « voici le temps écoulé ».
//
// Le corollaire vaut d'être su : le compteur de ticks est un nombre d'images et
// non une horloge, donc le minuteur du bandeau dérive de l'heure murale sur une
// machine chargée. Le déterminisme n'en souffre pas, puisque rien de ce qui
// décide ne lit autre chose que le compteur.
func (s *Screen) Step(in Input) {
	// **La mort fige la scène**, et c'est ici que la décision se prend puisque
	// `World.Step` la laisse ouverte. Le chapitre 2 veut que le joueur puisse se
	// raconter sa mort ; une horde qui continue d'avancer sous le voile efface
	// en deux secondes la configuration qui l'a tué, c'est-à-dire ce qu'il y
	// avait à comprendre.
	if s.monde.Over() {
		return
	}

	// **La pause du choix est réelle, et elle se tient ici pour la même raison.**
	// Choisir sous pression n'est pas un choix mais une loterie, et le chapitre 2
	// a posé que le choix compte plus que la récompense. Elle vient après la
	// mort : mourir dans le tick qui ouvre un choix laisse l'écran de fin, et non
	// trois cartes suspendues au-dessus d'un cadavre.
	if s.monde.Choosing() {
		s.choisir(in)
		return
	}

	if pressee(in, emplacement1...) {
		// Sur l'enfoncement, comme la relance et le choix d'une carte : au
		// maintien, la charge partirait à l'image où le doigt se pose et le
		// joueur ne saurait jamais s'il l'a dépensée exprès.
		s.monde.Attract()
	}

	// **La même touche prend, échange ou déclenche**, et l'ordre n'est pas libre :
	// une arme sous les pieds l'emporte, sinon la touche déclenche ce que
	// l'emplacement tient. Sans cette priorité, presser la touche en passant sur
	// une trouvaille dépenserait une charge au lieu de la ramasser, et le joueur
	// verrait son arme partir en croyant en prendre une.
	for place, touches := range [][]input.Key{emplacement2, emplacement3} {
		if !pressee(in, touches...) {
			continue
		}
		if !s.monde.TakeDrop(place) {
			s.monde.Trigger(place)
		}
	}

	if pressee(in, emplacement4...) {
		// Sur l'enfoncement comme l'aimant, et pour la même raison : au maintien,
		// la fiole partirait à l'image où le doigt se pose. Rien à prendre au sol
		// ici — une fiole se ramasse en marchant dessus, la touche ne fait que
		// boire.
		s.monde.Drink()
	}

	if pressee(in, repere...) {
		s.poserRepere()
	}

	// **Au maintien, à la différence des emplacements** : on force un obstacle
	// en tenant la touche, une frappe par cadence. Marteler douze fois un rideau
	// de fer serait une corvée des doigts, pas une décision.
	if tenue(in, interaction...) {
		s.monde.Interact()
	}

	s.monde.Step(voulu(in))
	s.cam.suivre(s.monde.Player())
}

// Rebind rattache l'écran à la run qui remplace la précédente.
//
// **Relier plutôt que remonter, et ce n'est pas une économie de lignes.** La
// boucle du moteur reçoit une scène et la garde : un écran neuf en construirait
// une seconde, que personne ne présenterait. Ce qui l'accompagne est que la
// relance n'alloue rien — ni scène, ni attaches, ni relevé d'emprises —, ce que
// le budget d'allocation demande et que la seconde de relance de la conception
// demandait déjà.
//
// **Le lieu ne change pas, et c'est la condition.** Une relance rejoue la même
// salle sur une graine dérivée : la carte garde ses dimensions, donc les bornes
// de la caméra et le relevé par case restent justes. Le décor se repose quand
// même, parce que la porte y porte un état de partie et qu'elle vient de se
// refermer.
func (s *Screen) Rebind(monde *game.World, carte *game.CostGrid) error {
	for _, a := range s.attaches {
		a.vider(s.scene)
	}
	s.monde, s.carte = monde, carte
	s.carteChoisie = 0
	s.finRepere = 0
	s.aRevoir = s.aRevoir[:0]
	s.recouvrements = s.recouvrements[:0]
	clear(s.emprises)

	if err := s.LeDecor(); err != nil {
		return err
	}
	s.cam.suivre(monde.Player())
	return nil
}

// Follow recadre la caméra sur le joueur.
//
// **`Step` le fait déjà, et cette méthode est pour ce qui n'y passe pas.** La
// planche de relecture avance le monde elle-même — elle n'a ni touches à lire
// ni pause à tenir — puis dessine ; sans ce recadrage, sa caméra reste où le
// montage l'a posée et le joueur sort du champ au bout de quelques pas.
// L'aperçu animé l'a montré en soixante images où le décor ne bougeait pas.
func (s *Screen) Follow() { s.cam.suivre(s.monde.Player()) }

// Scene rend la scène que l'écran décrit, à qui la présente.
//
// **Le moteur la reçoit de l'extérieur** — `NewEngine` et `Run` la prennent tous
// deux en premier argument —, donc un descripteur qui la garderait pour lui
// n'aurait aucun appelant. Ce qu'elle expose est une description et jamais une
// image : ce paquet pose, le moteur trie et peint.
func (s *Screen) Scene() *scene.Scene { return s.scene }

// Describe décrit l'image : ce que les zones marquent au sol, les entités, leurs
// effets, puis l'interface par-dessus.
func (s *Screen) Describe() {
	// **Les quads d'écran se vident, le monde ne se vide pas.** L'interface est
	// recomposée à chaque image — une jauge suit la vie, un compteur change —
	// quand le décor reste posé depuis le montage et les entités sont suivies par
	// leur poignée. Le moteur ne vide jamais de lui-même, et c'est ce qui laisse
	// ce choix au jeu.
	s.scene.ClearScreen()

	s.releverEmprises()
	s.poserLaNappe()
	s.poserEntites()
	s.poserEffets()
	s.poserDanger()
	s.poserBandeau()
	if s.monde.Over() {
		s.poserFin()
		return
	}
	s.poserCartes()
}

// peindreDanger cerne l'écran de rouge quand la vie passe sous son seuil.
//
// **Une vignette de bord et non un aplat plein**, et c'est ce que l'aplat a
// coûté : teinté en entier, le sol se rapprochait de la teinte de la horde, si
// bien que les créatures s'en détachaient moins au moment précis où il faut voir
// pour s'échapper. La vignette laisse le centre intact et se lit malgré tout, la
// vision périphérique étant ce à quoi un bord s'adresse.
//
// **Après les entités et sous le bandeau.** Par-dessus le monde, parce que c'est
// lui qu'on regarde ; sous le bandeau, parce que teinter la jauge de vie en
// rouge la rendrait illisible au moment où elle décide de tout.
//
// Le dégradé vient de l'empilement et non d'une image : chaque bande couvre la
// précédente en s'éloignant moins du bord, si bien que l'opacité croît par
// paliers vers l'arête sans qu'aucune teinte de plus soit déclarée.
//
// **Six paliers et non trois**, ce qui n'est pas un réglage esthétique : à trois,
// la marche fait seize pixels et la vignette se lit comme un cadre d'interface,
// c'est-à-dire comme un élément posé sur le jeu plutôt que comme un état du
// joueur. Huit pixels la ramènent à ce qu'elle doit être, une teinte qui monte.
func (s *Screen) poserDanger() {
	if s.hud == nil || !s.monde.InDanger() {
		return
	}
	// **Elle cerne l'aire de jeu et non la surface**, parce qu'elle s'adresse à
	// la vision périphérique de qui regarde le jeu : au bord d'une fenêtre en
	// 4K, elle serait loin de là où l'œil est posé, et elle cesserait de dire ce
	// qu'elle existe pour dire.
	x, y, largeur, hauteur := aireDans(s.scene)
	teinte := s.hud.Color("vignette_danger")
	for palier := range paliersVignette {
		e := epaisseurVignette * (paliersVignette - palier) / paliersVignette
		s.hud.Rect(s.scene, x, y, largeur, e, teinte)
		s.hud.Rect(s.scene, x, y+hauteur-e, largeur, e, teinte)
		s.hud.Rect(s.scene, x, y+e, e, hauteur-2*e, teinte)
		s.hud.Rect(s.scene, x+largeur-e, y+e, e, hauteur-2*e, teinte)
	}
}

// LeDecor pose le décor du lieu dans la scène, une fois pour toutes.
//
// **Il n'y a plus de passe de sol**, et c'est ce que la scène a retiré : le
// décor se range au montage et le moteur ne relit que la fenêtre visible, là où
// deux boucles balayaient un millier de cases à chaque image. Ce qui dépasse du
// sol n'est plus peint à part non plus — son étage le range sous les vivants
// sans qu'aucune passe l'y mette.
func (s *Screen) LeDecor() error {
	return poserLeDecor(s.scene, s.sol, s.monde, s.carte)
}

// LaPorte repose la case de sortie, dont l'état de partie change l'apparence.
//
// **C'est la seule case qui se repose**, et elle existe parce qu'aucune forme du
// lieu ne dit qu'une case est la sortie : l'ouverture est un état de partie, et
// le décor n'en sait rien. Une partie jouée a montré qu'on fait les quatre coins
// d'une salle sans trouver une porte qui ressemble au mur qui l'entoure.
func (s *Screen) LaPorte() {
	sortie := s.monde.Exit()
	if sortie == nil {
		return
	}
	poserLaCase(s.scene, s.sol, s.monde, sortie.U, sortie.V)
}

// nappe est ce qu'une case porte d'une zone, le temps d'une image.
//
// **Une sorte plutôt qu'une teinte convenue.** Les deux natures se peignent par
// case et au même moment, mais l'une est un aplat teinté et l'autre un dessin :
// les distinguer par la couleur reviendrait à lire une intention dans une
// valeur, et le jour où les deux se ressembleraient, la nappe de feu se
// peindrait en losange sans que rien ne le dise.
type nappe struct {
	// vif est la teinte de l'aplat, ou l'atténuation du dessin. L'alpha nul dit
	// qu'aucune zone ne couvre la case.
	vif render.Color
	// feu dit que la case porte des flammes plutôt que le télégraphe d'une
	// explosion.
	feu bool
}

// releverEmprises note, case par case, ce qu'une zone couvre — l'emprise qu'une
// explosion annonce, et les flammes qui brûlent.
//
// **Les deux partagent ce tampon parce qu'elles partagent une contrainte**, et
// non par économie : l'une et l'autre appartiennent à la case et doivent suivre
// sa forme dans la passe qui la dessine. Ce qui les sépare est la teinte et le
// sens de l'intensité, et c'est assez pour qu'on ne les confonde pas.
//
// **Le télégraphe se peint en cases pleines, et cette forme n'est pas un
// pis-aller.** Le rendu ne lit aucune image : il n'a rien à agrandir, et
// redimensionner un aplat par une fraction casserait le pixel entier qui est
// toute la règle du pixel art. Des cases se peignent à l'échelle exacte du
// décor, avec la face qui sert déjà au sol.
//
// **L'emprise est montrée entière dès la première image**, et c'est ce qui la
// sépare d'un cercle qui grandirait : une zone qui s'élargit ferait croire qu'on
// est hors de portée jusqu'à l'instant où l'on ne l'est plus, c'est-à-dire
// exactement le mensonge qu'un avertissement ne peut pas se permettre. Ce qui
// croît est l'intensité, qui dit le temps restant sans rien dire de faux sur
// l'espace.
//
// **Deux explosions qui se recouvrent laissent la dernière relevée**, comme la
// passe qu'elle remplace laissait le dernier blit. Les intensités ne s'ajoutent
// pas : ce que le marquage dit est « ici », pas « combien ».
//
// Seule la fenêtre visible est effacée puis remplie. Ce qui est hors d'elle ne
// se dessine pas, et la caméra ne peut pas montrer une case sans que l'image où
// elle entre l'ait effacée d'abord.
func (s *Screen) releverEmprises() {
	largeur, hauteur := s.sol.carte.Width(), s.sol.carte.Height()

	// **La fenêtre visible déborde de la carte, et c'est voulu** : elle est
	// l'englobant d'un losange, plus une case de marge. Ses bornes se ramènent
	// donc à la carte avant de servir d'indices — les autres consommateurs le
	// font par `formeDe`, qui refuse une case hors carte ; ici l'indice est
	// calculé, et un `u0` négatif faisait un début de tranche négatif.
	u0, v0, u1, v1 := s.cam.casesVisibles()
	u0, u1 = max(u0, 0), min(u1, largeur-1)
	v0, v1 = max(v0, 0), min(v1, hauteur-1)
	for v := v0; v <= v1; v++ {
		clear(s.emprises[v*largeur+u0 : v*largeur+u1+1])
	}

	// **Les flammes se relèvent avant les mèches, et l'ordre porte la
	// priorité** : deux nappes qui se recouvrent laissent la dernière relevée, et
	// ce qui doit l'emporter est l'avertissement. Une flaque dit ce que le joueur
	// a posé, un télégraphe dit ce qu'il va prendre — masquer le second par le
	// premier retirerait au joueur la seule information qu'il ne peut pas
	// deviner.
	feux := s.monde.Fires()
	for i := range feux.Active() {
		f := feux.At(i)

		// **Le dessin porte ses couleurs, et l'atténuation part d'un blanc.** La
		// teinte multiplie l'image : un voile coloré sur des flammes rouges les
		// virerait au brun, ce qui est l'accident qui avait éteint l'éclair
		// d'impact le jour où les créatures ont eu leurs bandes. Ce qui s'éteint
		// ici est l'opacité, et elle faiblit là où une mèche s'intensifie.
		reste := float64(s.monde.FireLeft(f)) / 1000
		vif := intact.Scale(fraction(emprisePlancher + (1-emprisePlancher)*reste))

		fu0, fv0, fu1, fv1 := s.monde.FireBounds(f)
		for v := max(fv0, v0); v <= min(fv1, v1); v++ {
			for u := max(fu0, u0); u <= min(fu1, u1); u++ {
				if !s.carte.InBounds(u, v) || !s.monde.FireCovers(f, u, v) {
					continue
				}
				s.emprises[v*largeur+u] = nappe{vif: vif, feu: true}
			}
		}
	}

	souffles := s.monde.Blasts()
	for i := range souffles.Active() {
		b := souffles.At(i)

		// L'imminence monte quand la mèche descend : le rapport rendu est ce
		// qu'il reste, et c'est son complément qui s'affiche.
		imminence := float64(1000-s.monde.FuseLeft(b)) / 1000
		vif := teinteEmprise.Scale(fraction(emprisePlancher + (1-emprisePlancher)*imminence))

		bu0, bv0, bu1, bv1 := s.monde.BlastBounds(b)
		for v := max(bv0, v0); v <= min(bv1, v1); v++ {
			for u := max(bu0, u0); u <= min(bu1, u1); u++ {
				if !s.carte.InBounds(u, v) || !s.monde.BlastCovers(b, u, v) {
					continue
				}
				s.emprises[v*largeur+u] = nappe{vif: vif}
			}
		}
	}
}

// reveler redessine en aplat ce que la scène a recouvert.
//
// **La silhouette plutôt que la transparence, et l'argument tient en une
// phrase** : la transparence retire de l'information au décor, la silhouette en
// ajoute au personnage. Le chapitre 2 veut les deux — voir la horde arriver et
// comprendre où sont les murs —, et c'est la seule des trois voies qui n'enlève
// rien.
//
// **Elle ne révèle que le joueur et les projectiles de la horde**, jamais une
// créature : voir la horde à travers un bus retirerait au décor le seul pouvoir
// qu'il a sur le combat. Ce qui se cache derrière un camion doit rester une
// inconnue.
//
// **Et elle ne dépend pas de ce que le décor déclare.** Un joueur derrière trois
// Vigiles empilés est invisible sans qu'aucune forme ne soit en cause : une
// transparence ne sait effacer que ce qu'un manifeste a nommé, celle-ci ne
// regarde que ce qui est dessiné devant.
//
// **Le tracé derrière une chose du lieu, l'aplat derrière une créature.** Une
// partie jouée a montré ce que l'aplat coûte contre un mur : il l'efface là où
// il se pose, si bien que le personnage se lit *sur* la maçonnerie — et sur
// l'enceinte, qui n'a que le vide au-delà, comme sorti de la carte. Le tracé
// laisse le mur au travers et rend la profondeur. Contre une foule, l'aplat
// reste : ce qu'il faut alors est de retrouver le joueur dans une masse qui
// partage sa gamme, et un liseré s'y perdrait — c'est la mesure qui avait fait
// écarter la valeur au profit de la forme.

// teinteDeLaSorte rend la couleur qu'une silhouette prend.
//
// Deux sortes, deux teintes réservées. Le joueur reprend celle de son contour :
// c'est la même chose qui le détache, cernée quand on le voit et pleine quand on
// ne le voit plus. Le projectile de la horde garde le violet que le chapitre 14
// lui réserve — « une couleur qui n'existe nulle part ailleurs dans le
// catalogue ».

// regardDuJoueur rend ce que le personnage regarde : sa cible s'il en a une, son
// déplacement sinon.
//
// **La cible passe avant le pas, et c'est toute la décision.** Le chapitre 9
// pose que le sprite s'oriente sur la visée et non sur le déplacement : reculer
// en tirant vers l'avant est ce que le joueur fait tout le temps, et c'est ce
// qui doit se lire. Un cône de visée avait été écarté ; ce qu'il apportait de
// visuel se garde ainsi, sans lui.
//
// Sans cible, le déplacement — parce qu'un personnage figé dans une direction
// morte se lirait comme un défaut. Et sans l'un ni l'autre, `regarder` retombe
// sur la première bande du manifeste : le joueur à l'arrêt devant une salle vide
// fait face à l'écran, ce qui est la pose de repos.
func (s *Screen) regardDuJoueur(pas game.Vec) game.Vec {
	if vers, vise := s.monde.PlayerAim(); vise {
		return vers
	}
	return pas
}

// regardDe rend ce qu'une créature regarde.
//
// **Trois cas, et le dernier est celui qui compte.** Une charge regarde où elle
// va et non où le joueur est parti, ce qui est tout son comportement — elle ne
// corrige plus. Une créature qui avance regarde son pas. Et une créature à
// l'arrêt regarde le joueur : c'est le cas de la Buse, qui s'immobilise à sa
// portée pour tirer, et le sien seul rendrait un profil tourné au hasard.
func (s *Screen) regardDe(e *game.Enemy) game.Vec {
	if e.Charging() {
		return e.ChargeDir
	}
	if e.Step != (game.Vec{}) {
		return e.Step
	}
	x, y := s.monde.Player()
	return game.Vec{X: x - e.X, Y: y - e.Y}
}

// etatDe rend la teinte que l'état d'une créature ajoute, ou rien.
//
// **Les éclairs d'état survivent aux sprites, la table par profil non.** Une
// couleur par profil disait qui était quoi, ce que le dessin dit désormais
// mieux ; un impact ou un soigneur qui s'allume disent ce qui vient de se
// passer, et aucune pose ne le dirait à cent créatures à l'écran. Les premiers
// s'éteignent donc avec la dérogation, les seconds restent.
func etatDe(e *game.Enemy) *render.Color {
	switch {
	case e.Healing > 0:
		return &teinteSoigneur
	case e.Flash > 0:
		return &teinteImpact
	case e.Healed > 0:
		return &teinteSoigne
	case e.Telegraphing() && (e.ChargeTimer/poulsAnnonce)%2 == 0:
		return &teinteAnnonce
	}
	return nil
}

// braiseGemme est ce qu'il reste d'une gemme au dernier tick de sa vie.
//
// **Elle ne descend pas à zéro**, et ce n'est pas une prudence d'affichage :
// une gemme invisible mais ramassable est un objet qui ment. Le joueur qui ne la
// voit plus a toutes les raisons de la croire partie, et la ramasser au passage
// sans comprendre pourquoi son compteur bouge est un défaut de lisibilité plus
// insidieux qu'une absence d'affichage — il ne manque rien, quelque chose de
// faux est montré. Un quart de la teinte reste lisible sur le sol tout en disant
// que la gemme s'en va.
const braiseGemme = 0.25

// eteindre affaiblit une teinte à mesure que la gemme vieillit.
//
// **Une extinction continue et non un clignotement.** Le clignotement dirait
// « bientôt » sans dire « dans combien de temps », et il entrerait en
// concurrence avec les télégraphes d'attaque sur un écran déjà chargé.
// L'extinction, elle, donne l'âge en continu : c'est ce qui permet d'estimer une
// récolte avant de déclencher l'aimant, donc de faire du déclenchement une
// lecture de la salle plutôt qu'un réflexe.
//
// Linéaire sur toute la vie, et non sur sa fin seule : une gemme qui ne
// changerait qu'au dernier moment ne se distinguerait pas d'une gemme neuve
// pendant l'essentiel de son existence, et l'information n'arriverait qu'une
// fois inutile.
//
// **La fin de l'extinction et la disparition coïncident tant que l'échelle vient
// de la durée de vie elle-même**, ce qu'assure la signature : il n'existe pas de
// durée d'extinction distincte qui pourrait s'en écarter. En introduire une
// rouvrirait l'écart — une vie rallongée sans que l'extinction le soit rendrait
// des gemmes éteintes bien avant de partir, et la braise ne protégerait plus de
// rien.
// **Elle survit à `render.Color.Scale`, qui ne sait pas faire cela.** Celle du
// moteur met les quatre composantes à l'échelle, alpha compris, et sa godoc le
// justifie : amputer les trois couleurs seules donnerait « quelque chose qui
// n'est plus une couleur prémultipliée du tout ». C'est pourtant exactement ce
// qu'il faut ici — un assombrissement d'opaque, et non un estompement. Les vingt
// lignes qui l'expliquent restent donc, et c'est le seul endroit du rendu où une
// teinte se calcule à la main.
func eteindre(teinte render.Color, age, vie game.Tick) render.Color {
	if vie <= 0 {
		return teinte
	}
	part := 1 - (1-braiseGemme)*min(float64(age)/float64(vie), 1)
	return render.Color{
		R: uint8(float64(teinte.R) * part),
		G: uint8(float64(teinte.G) * part),
		B: uint8(float64(teinte.B) * part),
		// L'alpha ne bouge pas, et c'est ce qui la sépare d'`attenuer` : la gemme
		// est opaque, donc l'assombrir revient bien à l'éteindre. Mettre son
		// alpha à l'échelle la rendrait transparente au lieu de sombre, et une
		// gemme qui s'efface sur un sol clair n'a plus de silhouette.
		A: teinte.A,
	}
}

// attenuer met une teinte prémultipliée à l'échelle, ses quatre canaux ensemble.
//
// **Les quatre, et c'est la seule façon d'atténuer une couleur prémultipliée.**
// Toucher au seul alpha laisse les composantes à leur valeur pleine, si bien que
// le rapport de chacune à l'alpha monte : la couleur s'affirme au lieu de
// s'effacer. L'emprise d'une explosion faisait ainsi l'inverse de ce qu'elle
// promet — la plus marquée quand la mèche s'allume, la plus pâle à l'instant du
// souffle —, et l'état intermédiaire n'était même plus une couleur prémultipliée
// valide, sa composante rouge dépassant son alpha.
//
// `eteindre` est sa voisine et ne s'y ramène pas : elle assombrit une teinte
// opaque, où l'alpha n'a rien à dire.
// fraction convertit une part de un en la fraction de 255 que le moteur attend.
//
// **`render.Color.Scale` remplace `attenuer`**, et le gain n'est pas d'avoir
// retiré huit lignes : sa fraction entière rend les mêmes pixels d'une machine à
// l'autre, là où un facteur flottant appliqué composante par composante
// dépendait de l'arrondi du poste. C'est le même argument que la virgule fixe de
// la simulation, appliqué à une couleur.
func fraction(part float64) uint8 {
	return uint8(min(max(part, 0), 1) * 255)
}

// voulu rend la direction que les touches tenues demandent, dans le repère du
// monde.
//
// Les touches sont en repère d'écran et le monde ne connaît que ses deux axes
// obliques : « haut » vaut donc (-1, -1), et « gauche » (-1, +1). C'est la
// conversion dont ce paquet a la charge, prise par son autre bout, et c'est
// pourquoi elle vit ici plutôt que dans la boucle de jeu.
//
// La somme des touches enfoncées donne les huit directions sans table de
// combinaisons — haut et gauche font (-2, 0), que la simulation ramène à
// l'unité. Elle ne normalise pas non plus : c'est `Vec.Direction` qui le fait,
// avec l'arrondi que le déterminisme exige, et le refaire ici en flottants
// donnerait deux réponses pour une question.
//
// `input.Key` désigne une touche par sa **place** sur un clavier américain :
// `KeyW` est la touche marquée Z sur un clavier français, et le carré ZQSD tombe
// donc juste sans qu'on ait à le nommer.
func voulu(in Input) game.Vec {
	var v game.Vec
	if tenue(in, input.KeyW, input.KeyUp) {
		v = v.Add(game.Vec{X: -game.One, Y: -game.One})
	}
	if tenue(in, input.KeyS, input.KeyDown) {
		v = v.Add(game.Vec{X: game.One, Y: game.One})
	}
	if tenue(in, input.KeyA, input.KeyLeft) {
		v = v.Add(game.Vec{X: -game.One, Y: game.One})
	}
	if tenue(in, input.KeyD, input.KeyRight) {
		v = v.Add(game.Vec{X: game.One, Y: -game.One})
	}
	return v
}

// face peint la face supérieure d'une case, en blanc, pour être teintée au blit.
//
// La forme reprend le test de `outils/primitives_iso.py` — bords sur des droites
// de pente 1/2, avec la même marge d'un demi-pixel — pour que les faces se
// jointent comme le feront les images du décor. Ce n'est pas une seconde
// description qu'il faudrait tenir d'accord avec lui : les deux ne s'affichent
// jamais ensemble, et celle-ci s'en va quand l'atlas entre.
// **Elle entre dans l'atlas par `NewSheetFromStraight`**, qui est la porte du
// moteur pour qui tient déjà ses pixels : la prémultiplication y a lieu une fois,
// comme au chargement d'un PNG, au lieu d'être refaite ou oubliée ici.
//
// C'est le dernier dessin que cohue calcule plutôt que de le charger, et il le
// reste : un losange plein dont la taille suit la tuile n'est pas une ressource
// mais une conséquence du manifeste de décor.
func face(tuile [2]int, atlas *raster.Atlas) (uint16, error) {
	largeur, hauteur := tuile[0], tuile[1]
	if largeur < 1 || hauteur < 1 || largeur > math.MaxUint16 || hauteur > math.MaxUint16 {
		return 0, fmt.Errorf("losange de %dx%d, hors de ce qu'une planche mesure",
			largeur, hauteur)
	}
	demi := float64(largeur) / 2
	marge := 0.5 / float64(largeur)

	pixels := make([]byte, 4*largeur*hauteur)
	for y := range hauteur {
		for x := range largeur {
			px, py := float64(x)+0.5-demi, float64(y)+0.5
			u := px/float64(largeur) + py/demi
			v := -px/float64(largeur) + py/demi
			if u < -marge || u > 1+marge || v < -marge || v > 1+marge {
				continue
			}
			i := 4 * (y*largeur + x)
			pixels[i], pixels[i+1], pixels[i+2], pixels[i+3] = 255, 255, 255, 255
		}
	}

	// #nosec G115 -- bornés à l'entrée
	p, err := raster.NewSheetFromStraight(uint16(largeur), uint16(hauteur), pixels)
	if err != nil {
		return 0, fmt.Errorf("losange : %w", err)
	}
	return atlas.Add(p)
}
