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
// **Ce paquet n'a pas de fichier de test, et n'en aura pas.** Importer
// Ebitengine initialise GLFW, qui panique sans `DISPLAY` : sur un runner sans
// écran, n'importe quel test du paquet tombe avant d'avoir commencé, y compris
// un test qui n'ouvrirait aucune fenêtre. La suite par défaut reste donc
// exécutable partout, et le rendu se juge à l'œil — c'est ce que la doctrine de
// test énonçait déjà, et l'absence de tests ici en est la conséquence et non un
// oubli. Ce qui doit être vérifié mécaniquement vit du côté simulation, qui
// n'importe pas Ebitengine.
package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/sprimault/cohue/internal/game"
)

// Les dimensions du tampon interne, fixes.
//
// Un tampon de 480×270 ne montrerait que sept tuiles de large, bien trop serré
// pour voir la horde arriver ; 960×540 en donne une quinzaine et se multiplie
// par deux pour du 1080p.
//
// C'est leur fixité qui est la règle de pixel art. Le facteur qui les agrandit
// vers la fenêtre est un réglage d'affichage, que l'étape 15 tranchera — voir
// `Layout`, qui dit ce qu'il en est aujourd'hui.
const (
	Width  = 960
	Height = 540
)

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

// Les teintes du rendu provisoire, qui tiendront jusqu'à ce que l'atlas entre.
//
// Elles ne cherchent pas à ressembler à un lieu : ce sont des états de la grille
// de coûts et des rôles de créature, choisis pour se distinguer et pour rien
// d'autre. La palette fermée du jeu vaut pour les images du décor, qui sortent
// des générateurs, et non pour ces aplats qui disparaîtront avec eux.
//
// **Deux axes, et ils ne servent pas la même question.** La valeur sépare ce
// qu'on est de ce qu'on affronte — le joueur clair, la horde sombre —, et c'est
// elle qui tient à cent créatures à l'écran. La teinte sépare les rôles entre
// eux, ce qui ne se pose qu'une fois la première séparation acquise. Les
// confondre donnerait un profil plus visible que les autres, décision de
// conception que personne n'a prise.
var (
	fond      = color.RGBA{R: 24, G: 24, B: 28, A: 255}
	solBloque = color.RGBA{R: 58, G: 58, B: 66, A: 255}
	solLibre  = color.RGBA{R: 96, G: 98, B: 104, A: 255}
	solLent   = color.RGBA{R: 74, G: 96, B: 120, A: 255}

	// La porte, dans ses deux états. Elle est peinte à part de la grille parce
	// qu'elle n'y change rien : fermée elle est un mur comme un autre, ouverte
	// elle l'est encore. Ce que ces deux teintes disent est ce qu'aucun coût ne
	// porte — le lieu est gagné, et la sortie est là.
	//
	// **Une seule teinte, éteinte puis vive**, plutôt que deux couleurs : c'est
	// le même objet dans deux états, et l'écart de valeur se voit en périphérie
	// là où un changement de teinte demande de regarder. Un cyan, parce que le
	// vert est deux fois pris — la gemme et le Secouriste — et qu'ouvrir une
	// porte n'a rien à voir avec l'un ni avec l'autre.
	//
	// **Deux écarts de valeur à tenir, et le premier jet n'en a tenu qu'un.**
	// L'écart entre les deux états était bon ; celui entre la porte fermée et le
	// mur qui l'entoure ne l'était pas, si bien qu'une partie jouée n'a pas
	// trouvé la porte en faisant les quatre coins de la salle. Une teinte ne se
	// choisit pas contre l'autre état du même objet, elle se choisit **aussi**
	// contre ce qui l'entoure — ici `solBloque`, qui occupe toute l'enceinte.
	porteFermee  = color.RGBA{R: 48, G: 132, B: 152, A: 255}
	porteOuverte = color.RGBA{R: 120, G: 232, B: 248, A: 255}

	// Le joueur en clair et les créatures en sombre : le chapitre de la
	// lisibilité veut que le personnage reste distinguable à cent ennemis à
	// l'écran, ce qui se joue d'abord sur la valeur et non sur la teinte.
	teinteJoueur = color.RGBA{R: 236, G: 214, B: 120, A: 255}
	teinteTir    = color.RGBA{R: 226, G: 232, B: 238, A: 255}
	// **Le projectile de la horde porte une teinte qu'aucune autre ne dispute**,
	// et c'est entre projectiles que la distinction doit être maximale : « est-ce
	// que ça me fait mal ? » ne se pose que sur eux, si bien que les confondre
	// coûte plus cher que de confondre un projectile et une créature.
	//
	// Le violet plutôt que le cyan pour cette raison : le cyan voisinerait le
	// blanc bleuté ci-dessus. **L'Arpenteur porte désormais un violet sourd**, et
	// les deux ne se disputent pas — un projectile est clair et vif, une créature
	// sombre, ce qui est la même séparation par la valeur que partout ailleurs.
	// L'arbitrage se rejugera à l'étape 5, entre deux teintes qui seront enfin
	// celles du jeu.
	teinteTirHorde = color.RGBA{R: 186, G: 138, B: 232, A: 255}
	// L'emprise d'une explosion se pose sur le sol et doit rester lisible sous
	// les corps qui la traversent : une teinte chaude que ni la horde ni le
	// joueur ne portent, et translucide **prémultipliée par son alpha** — écrite
	// à plat, elle rendrait un aplat trois fois trop dense.
	teinteEmprise = color.RGBA{R: 108, G: 54, B: 24, A: 132}
	// **L'éclair d'une créature touchée est la même teinte, éclaircie**, et non
	// une couleur nouvelle : ce qu'il doit dire est « celle-ci vient d'être
	// atteinte », pas « ceci est autre chose ». Un blanc franc ferait clignoter
	// une foule dense en bandes qu'on ne relierait plus à des créatures.
	teinteImpact = color.RGBA{R: 236, G: 186, B: 178, A: 255}
	// **L'annonce d'une charge bat, elle ne se contente pas d'une couleur.** Une
	// teinte fixe se perd dans une foule qui porte déjà le rouge, et c'est
	// justement au milieu de la foule qu'il faut la repérer pour mettre un
	// obstacle entre soi et elle. Le battement se lit là où un aplat de plus ne
	// se lirait pas, et il dit l'imminence plutôt que l'état.
	//
	// La course, elle, n'est pas peinte : elle va vite et droit, ce qui est déjà
	// à l'image. Un signal ne redit pas ce qu'on voit.
	teinteAnnonce = color.RGBA{R: 232, G: 96, B: 72, A: 255}
	// **Le soigneur tranche, la soignée nuance.** Les deux disent le même
	// événement mais n'apprennent pas la même chose : qu'une créature récupère
	// explique pourquoi elle ne tombe pas, et cela peut rester dans la famille du
	// rouge ; savoir *lequel* d'une horde identique est le Secouriste change la
	// conduite du joueur, et cela doit sauter aux yeux là où tout est rouge.
	//
	// D'où un vert franc pour l'un et un rouge éclairci vers le vert pour
	// l'autre : les relier tient à la teinte commune, les séparer à la distance
	// qui les sépare du fond de la horde.
	teinteSoigneur = color.RGBA{R: 120, G: 226, B: 132, A: 255}
	teinteSoigne   = color.RGBA{R: 172, G: 178, B: 108, A: 255}
	// Une gemme est minuscule et posée sur un sol gris : elle a besoin d'une
	// teinte saturée que rien d'autre ne porte, sans quoi un tas au sol
	// disparaît sous la horde au moment où l'on cherche à l'estimer.
	// **Le figurant se lit comme du décor, pas comme une menace.** Il ne blesse
	// pas, ne se vise pas et ne compte nulle part : lui donner la teinte de la
	// horde ferait tirer le joueur dessus par réflexe, et lui en donner une vive
	// le rendrait plus important que ce qui le tue. Un gris à peine teinté, plus
	// clair que le sol pour se détacher, plus terne que tout le reste.
	teinteAmbiance = color.RGBA{R: 132, G: 130, B: 138, A: 255}
	teinteGemme    = color.RGBA{R: 96, G: 214, B: 168, A: 255}
	// L'aimant doit se voir de loin, puisque tout son intérêt est qu'on décide
	// d'aller le chercher. Le cuivre d'une bobine, qui ne dispute sa teinte à
	// personne — le joueur tient le jaune, la horde le rouge, les gemmes le vert,
	// les projectiles le blanc.
	//
	// C'est la teinte du sprite de `assets/objets/`, recopiée ici le temps que le
	// rendu lise les images. Deux descriptions de la même couleur, donc, et elles
	// cesseront de l'être à l'étape 5 : c'est l'aplat qui disparaîtra.
	teinteAimant = color.RGBA{R: 198, G: 126, B: 78, A: 255}
	// **La caisse est du décor jusqu'à ce qu'on la touche**, et sa teinte le
	// dit : un bois sourd, moins saturé que l'aimant qu'elle voisine en teinte.
	// Ce qui doit rester distinct est l'objet qu'on va chercher — l'aimant — de
	// celui qu'on casse en passant.
	teinteCaisse = color.RGBA{R: 142, G: 108, B: 72, A: 255}

	// La horde, une teinte par profil.
	//
	// **Toutes sombres, et c'est la contrainte qui commande.** Le chapitre de la
	// lisibilité veut que le joueur reste distinguable à cent ennemis à l'écran,
	// ce qui se joue sur la valeur et non sur la teinte : la horde occupe donc la
	// moitié sombre, et le joueur seul la moitié claire. Ces sept-là se séparent
	// entre elles par la teinte, jamais par la valeur — l'inverse aurait rendu un
	// profil plus visible que les autres, ce qui est un choix de conception que
	// personne n'a fait.
	//
	// **Elles sont provisoires et le resteront jusqu'à l'atlas**, où chaque
	// créature aura son apparence. Ce qu'elles servent n'est pas l'apparence mais
	// une question de jeu : reconnaître un Vigile, un Secouriste ou une Baudruche
	// dans une foule décide de ce qu'on fait, et une horde d'une seule teinte
	// rend cette décision impossible à juger.
	//
	// **Rangées par clé de manifeste et non par index**, et le premier jet a
	// montré pourquoi : la table des profils est triée alphabétiquement, si bien
	// qu'un tableau indexé par position donnait le rouge du Quidam au Vigile et
	// l'olive du Vigile au Quidam. Rien ne pouvait le dire — les deux indices
	// sont valides, une couleur fausse ne casse aucun test, et il a fallu
	// demander « quelle couleur, le Vigile ? » pour que ça se voie.
	//
	// Une clé absente de cette table prend le rouge de la masse : un profil de
	// plus doit se peindre, pas faire tomber le rendu.
	teintesHorde = map[string]color.RGBA{
		"marcheur":  {R: 150, G: 78, B: 74, A: 255},  // le Quidam garde le rouge de la masse
		"flanqueur": {R: 122, G: 82, B: 150, A: 255}, // l'Arpenteur, violet sourd
		"sprinteur": {R: 168, G: 96, B: 52, A: 255},  // le Molosse, orange terreux
		"bloqueur":  {R: 84, G: 96, B: 140, A: 255},  // le Vigile, bleu d'uniforme
		"cracheur":  {R: 110, G: 124, B: 62, A: 255}, // la Buse, olive
		"eclateur":  {R: 162, G: 72, B: 116, A: 255}, // la Baudruche, magenta sourd
		"soigneur":  {R: 74, G: 128, B: 102, A: 255}, // le Secouriste, vert de son éclair
	}
	teinteHordeParDefaut = color.RGBA{R: 150, G: 78, B: 74, A: 255}
)

// Screen est le jeu tel qu'Ebitengine le voit.
type Screen struct {
	monde *game.World
	carte *game.CostGrid
	cam   *camera

	scene *scene

	// Les cinq formes blanches que le dessin teinte au blit : la face d'une
	// case, la silhouette d'un personnage, le point d'un projectile, celui d'une
	// gemme et celui d'un aimant.
	face     *ebiten.Image
	figurine *ebiten.Image
	eclat    *ebiten.Image
	gemme    *ebiten.Image
	aimant   *ebiten.Image
	caisse   *ebiten.Image
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

	// hud pose le bandeau de la partie et le texte de l'écran de mort. Il peut
	// être nul : la planche de relecture monte des écrans sans lui, et une partie
	// sans interface se dessine quand même — ce qu'elle perd est tout le texte du
	// jeu.
	hud *HUD

	// op est réutilisée d'un blit à l'autre, et remise à zéro à chaque fois :
	// une case visible en produit un millier par image.
	op ebiten.DrawImageOptions
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

// NewScreen monte le rendu sur une partie et le lieu qu'elle joue.
//
// La taille de tuile est celle du manifeste de décor et jamais une constante :
// le manifeste la porte, le chargeur en exige le rapport de deux pour un, et
// c'est de lui que la projection la tient.
func NewScreen(monde *game.World, carte *game.CostGrid, tuile [2]int) *Screen {
	s := &Screen{
		monde: monde,
		carte: carte,
		cam:   nouvelleCamera(tuile, carte),
		scene: nouvelleScene(carte, monde.Enemies().Cap(), monde.Shots().Cap(),
			monde.Gems().Cap(), monde.Magnets().Cap(), monde.Crates().Cap()),
		face:     face(tuile),
		figurine: aplat(tuile[0]/4, tuile[0]*3/4),
		eclat:    aplat(tuile[1]/8, tuile[1]/8),
		// Deux fois l'éclat : assez pour qu'un tas se compte d'un coup d'œil,
		// assez peu pour qu'une gemme ne masque pas ce qui la piétine.
		gemme: aplat(tuile[1]/4, tuile[1]/4),
		// Deux fois la gemme : il ne s'agit pas d'estimer un tas mais de
		// repérer un objet unique à l'autre bout de la salle, et c'est la
		// taille qui porte ça avant la teinte.
		aimant: aplat(tuile[1]/2, tuile[1]/2),
		// Plus large que haute, à l'inverse d'une figurine : ce qui doit se lire
		// est un objet posé au sol qu'on va casser, pas une créature qu'on
		// affronte. La confusion coûterait un détour ou une salve, et elle est
		// facile à faire tant que tout est un aplat.
		caisse:    aplat(tuile[0]/3, tuile[1]/2),
		demiTuile: tuile[0] / 2,
	}
	s.cam.suivre(monde.Player())
	return s
}

// emplacement1 est la touche qui dépense la charge du premier emplacement, celui
// de l'aimant.
//
// Les deux places du chiffre, comme les deux places d'Entrée : le pavé numérique
// tombe sous la main droite quand la gauche tient le déplacement, et un joueur
// qui appuie sur le 1 qu'il a sous les doigts n'a aucun moyen de savoir que le
// jeu écoutait l'autre.
var emplacement1 = []ebiten.Key{ebiten.Key1, ebiten.KeyNumpad1}

// Update avance la simulation d'un pas, puis recadre.
//
// Un pas par appel et rien qui lise l'horloge : Ebitengine appelle cette méthode
// à cadence fixe et rattrape un retard en l'appelant plusieurs fois d'affilée,
// ce qui est exactement ce que la simulation attend d'un appelant.
func (s *Screen) Update() error {
	// **La mort fige la scène**, et c'est ici que la décision se prend puisque
	// `World.Step` la laisse ouverte. Le chapitre 2 veut que le joueur puisse se
	// raconter sa mort ; une horde qui continue d'avancer sous le voile efface
	// en deux secondes la configuration qui l'a tué, c'est-à-dire ce qu'il y
	// avait à comprendre.
	if s.monde.Over() {
		return nil
	}

	// **La pause du choix est réelle, et elle se tient ici pour la même raison.**
	// Choisir sous pression n'est pas un choix mais une loterie, et le chapitre 2
	// a posé que le choix compte plus que la récompense. Elle vient après la
	// mort : mourir dans le tick qui ouvre un choix laisse l'écran de fin, et non
	// trois cartes suspendues au-dessus d'un cadavre.
	if s.monde.Choosing() {
		s.choisir()
		return nil
	}

	if presse(emplacement1) {
		// Sur l'enfoncement, comme la relance et le choix d'une carte : au
		// maintien, la charge partirait à l'image où le doigt se pose et le
		// joueur ne saurait jamais s'il l'a dépensée exprès.
		s.monde.Attract()
	}

	if presse(repere) {
		s.poserRepere()
	}

	s.monde.Step(voulu())
	s.cam.suivre(s.monde.Player())
	return nil
}

// Draw peint le tampon interne : le sol, ce qui s'y tient en profondeur, puis
// l'interface par-dessus.
func (s *Screen) Draw(ecran *ebiten.Image) {
	ecran.Fill(fond)
	s.peindreSol(ecran)
	s.peindreEntites(ecran)
	s.peindreDanger(ecran)
	s.peindreBandeau(ecran)
	if s.monde.Over() {
		s.peindreFin(ecran)
		return
	}
	s.peindreCartes(ecran)
}

// Layout fixe la taille du tampon interne, quelle que soit celle de la fenêtre.
//
// Ebitengine agrandit ensuite vers la fenêtre, et le facteur n'est pas
// nécessairement entier : une fenêtre large de 1400 pixels montre le tampon
// agrandi de 1,45 fois, donc des pixels de tailles inégales. Ce qui reste à
// trancher n'appartient ni à la projection ni à la caméra, qui travaillent
// toutes deux dans le tampon : c'est le facteur d'échelle du système qui décide,
// il se lit par `LayoutF`, et aucun des deux ne le connaît.
func (s *Screen) Layout(largeurFenetre, hauteurFenetre int) (int, int) {
	return Width, Height
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
func (s *Screen) peindreDanger(ecran *ebiten.Image) {
	if s.hud == nil || !s.monde.InDanger() {
		return
	}
	teinte := s.hud.Color("vignette_danger")
	for palier := range paliersVignette {
		e := epaisseurVignette * (paliersVignette - palier) / paliersVignette
		s.hud.Rect(ecran, 0, 0, Width, e, teinte)
		s.hud.Rect(ecran, 0, Height-e, Width, e, teinte)
		s.hud.Rect(ecran, 0, e, e, Height-2*e, teinte)
		s.hud.Rect(ecran, Width-e, e, e, Height-2*e, teinte)
	}
}

// peindreSol pose la face de chaque case visible, teintée par son coût.
//
// Ce que ce sol montre n'est pas le décor mais la grille de coûts, c'est-à-dire
// ce que le champ de flux lit : franchissable, coûteux, ou mur. Tant qu'aucun
// atlas n'est chargé, c'est l'information la plus utile qu'une case puisse
// porter — un lieu se relit alors comme la simulation le voit, et un écart entre
// les deux se verrait ici avant de se deviner ailleurs.
func (s *Screen) peindreSol(ecran *ebiten.Image) {
	u0, v0, u1, v1 := s.cam.casesVisibles()
	for v := v0; v <= v1; v++ {
		for u := u0; u <= u1; u++ {
			if !s.carte.InBounds(u, v) {
				continue
			}
			x, y := s.cam.ecran(game.FromInt(u), game.FromInt(v))
			s.op.GeoM.Reset()
			s.op.GeoM.Translate(float64(x-s.demiTuile), float64(y))
			s.op.ColorScale.Reset()
			s.op.ColorScale.ScaleWithColor(s.teinteCase(u, v))
			ecran.DrawImage(s.face, &s.op)
		}
	}
	s.peindreEmprises(ecran)
}

// teinteDuProfil rend la couleur d'une créature, le rouge de la masse pour une
// clé que la table ne connaît pas.
func teinteDuProfil(cle string) color.RGBA {
	if teinte, connue := teintesHorde[cle]; connue {
		return teinte
	}
	return teinteHordeParDefaut
}

// teinteCase rend la couleur d'une case : celle de la porte si c'en est une,
// celle de son coût sinon.
//
// **La porte prime sur le coût**, sans quoi elle se peindrait en mur et le
// joueur n'aurait aucun moyen de la trouver — c'est la seule case du lieu dont
// la grille ne dit pas ce qu'il faut en savoir.
func (s *Screen) teinteCase(u, v int) color.RGBA {
	sortie := s.monde.Exit()
	if sortie != nil && sortie.U == u && sortie.V == v {
		if s.monde.DoorOpen() {
			return porteOuverte
		}
		return porteFermee
	}
	return teinte(s.carte.At(u, v))
}

// peindreEmprises marque au sol ce qu'une explosion amorcée va couvrir.
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
func (s *Screen) peindreEmprises(ecran *ebiten.Image) {
	souffles := s.monde.Blasts()
	for i := range souffles.Active() {
		b := souffles.At(i)

		// L'imminence monte quand la mèche descend : le rapport rendu est ce
		// qu'il reste, et c'est son complément qui s'affiche.
		imminence := float32(1000-s.monde.FuseLeft(b)) / 1000
		vif := teinteEmprise
		vif.A = uint8(float32(teinteEmprise.A) * (emprisePlancher + (1-emprisePlancher)*imminence))

		u0, v0, u1, v1 := s.monde.BlastBounds(b)
		for v := v0; v <= v1; v++ {
			for u := u0; u <= u1; u++ {
				if !s.carte.InBounds(u, v) || !s.monde.BlastCovers(b, u, v) {
					continue
				}
				x, y := s.cam.ecran(game.FromInt(u), game.FromInt(v))
				s.op.GeoM.Reset()
				s.op.GeoM.Translate(float64(x-s.demiTuile), float64(y))
				s.op.ColorScale.Reset()
				s.op.ColorScale.ScaleWithColor(vif)
				ecran.DrawImage(s.face, &s.op)
			}
		}
	}
}

// peindreEntites pose ce qui se tient sur le sol, du plus lointain au plus
// proche.
//
// La position est relue dans le bassin plutôt que portée par la séquence : le
// tri range des rangs, pas des coordonnées, et une copie faite au moment du tri
// aurait une image de retard le jour où quelque chose bougera entre les deux.
func (s *Screen) peindreEntites(ecran *ebiten.Image) {
	for _, e := range s.scene.ranger(s.monde) {
		switch e.sorte {
		case sorteEnnemi:
			c := s.monde.Enemies().At(e.place)
			teinte := teinteDuProfil(s.monde.EnemyKey(c.Profile))
			// **Le soigneur passe avant l'impact, la soignée après.** Ce qui
			// change la conduite du joueur est de repérer d'où vient le soin :
			// la visée prend le plus proche, donc abattre un Secouriste demande
			// d'abord de savoir lequel c'est. Une créature soignée qu'on est en
			// train de toucher, elle, porte les deux informations — et celle qui
			// compte alors est le coup qui vient de partir.
			switch {
			case c.Healing > 0:
				teinte = teinteSoigneur
			case c.Flash > 0:
				teinte = teinteImpact
			case c.Healed > 0:
				teinte = teinteSoigne
			case c.Telegraphing() && (c.ChargeTimer/poulsAnnonce)%2 == 0:
				teinte = teinteAnnonce
			}
			s.silhouette(ecran, s.figurine, c.X, c.Y, teinte)
		case sorteAmbiance:
			a := s.monde.Ambients().At(e.place)
			s.silhouette(ecran, s.figurine, a.X, a.Y, teinteAmbiance)
		case sorteTir:
			p := s.monde.Shots().At(e.place)
			s.silhouette(ecran, s.eclat, p.X, p.Y, teinteTir)
		case sorteTirHorde:
			p := s.monde.EnemyShots().At(e.place)
			s.silhouette(ecran, s.eclat, p.X, p.Y, teinteTirHorde)
		case sorteGemme:
			g := s.monde.Gems().At(e.place)
			// **Une gemme attirée reprend sa teinte pleine.** L'extinction dit
			// « ceci va disparaître » ; une gemme que l'aimant tient ne
			// disparaîtra pas, et la montrer éteinte serait montrer une
			// information fausse. Accessoirement, une ruée de gemmes anciennes
			// serait un feu d'artifice en gris — l'inverse de ce que la
			// conception appelle le moment de plaisir maximal du genre.
			teinte := teinteGemme
			if !g.Pulled {
				teinte = eteindre(teinte, s.monde.GemAge(g), s.monde.GemLife())
			}
			s.silhouette(ecran, s.gemme, g.X, g.Y, teinte)
		case sorteAimant:
			a := s.monde.Magnets().At(e.place)
			s.silhouette(ecran, s.aimant, a.X, a.Y, teinteAimant)
		case sorteCaisse:
			c := s.monde.Crates().At(e.place)
			s.silhouette(ecran, s.caisse, c.X, c.Y, teinteCaisse)
		case sorteJoueur:
			x, y := s.monde.Player()
			s.silhouette(ecran, s.figurine, x, y, teinteJoueur)
		}
	}
}

// silhouette pose une forme teintée, son pied sur le point où le monde la situe.
//
// L'appui est au milieu du bas, ce que sera l'ancrage d'un sprite de personnage
// quand le manifeste en fournira : c'est le point qui touche le sol, et le seul
// qui puisse coïncider avec une position du monde.
func (s *Screen) silhouette(ecran, forme *ebiten.Image, x, y game.Fixed, teinte color.RGBA) {
	ex, ey := s.cam.ecran(x, y)
	taille := forme.Bounds()
	s.op.GeoM.Reset()
	s.op.GeoM.Translate(float64(ex-taille.Dx()/2), float64(ey-taille.Dy()))
	s.op.ColorScale.Reset()
	s.op.ColorScale.ScaleWithColor(teinte)
	ecran.DrawImage(forme, &s.op)
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
func eteindre(teinte color.RGBA, age, vie game.Tick) color.RGBA {
	if vie <= 0 {
		return teinte
	}
	part := 1 - (1-braiseGemme)*min(float64(age)/float64(vie), 1)
	return color.RGBA{
		R: uint8(float64(teinte.R) * part),
		G: uint8(float64(teinte.G) * part),
		B: uint8(float64(teinte.B) * part),
		A: teinte.A,
	}
}

// teinte dit de quelle couleur une case se peint, selon ce qu'elle coûte.
func teinte(cout game.Cost) color.RGBA {
	switch {
	case cout == game.Blocked:
		return solBloque
	case cout > game.Free:
		return solLent
	}
	return solLibre
}

// voulu lit les touches et rend la direction demandée, dans le repère du monde.
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
// `ebiten.Key` désigne une touche par sa **place** sur un clavier américain :
// `KeyW` est la touche marquée Z sur un clavier français, et le carré ZQSD tombe
// donc juste sans qu'on ait à le nommer.
func voulu() game.Vec {
	var v game.Vec
	if enfonce(ebiten.KeyW, ebiten.KeyArrowUp) {
		v = v.Add(game.Vec{X: -game.One, Y: -game.One})
	}
	if enfonce(ebiten.KeyS, ebiten.KeyArrowDown) {
		v = v.Add(game.Vec{X: game.One, Y: game.One})
	}
	if enfonce(ebiten.KeyA, ebiten.KeyArrowLeft) {
		v = v.Add(game.Vec{X: -game.One, Y: game.One})
	}
	if enfonce(ebiten.KeyD, ebiten.KeyArrowRight) {
		v = v.Add(game.Vec{X: game.One, Y: -game.One})
	}
	return v
}

// enfonce dit si l'une des touches est pressée.
func enfonce(touches ...ebiten.Key) bool {
	for _, t := range touches {
		if ebiten.IsKeyPressed(t) {
			return true
		}
	}
	return false
}

// face peint la face supérieure d'une case, en blanc, pour être teintée au blit.
//
// La forme reprend le test de `outils/primitives_iso.py` — bords sur des droites
// de pente 1/2, avec la même marge d'un demi-pixel — pour que les faces se
// jointent comme le feront les images du décor. Ce n'est pas une seconde
// description qu'il faudrait tenir d'accord avec lui : les deux ne s'affichent
// jamais ensemble, et celle-ci s'en va quand l'atlas entre.
func face(tuile [2]int) *ebiten.Image {
	largeur, hauteur := tuile[0], tuile[1]
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

	img := ebiten.NewImage(largeur, hauteur)
	img.WritePixels(pixels)
	return img
}

// aplat rend un rectangle blanc plein, à teinter au blit.
//
// Un quart de tuile de large et trois quarts de haut pour la silhouette du
// joueur : un personnage tient dans une image de la largeur d'une tuile et s'y
// dresse presque entier, si bien que ces proportions donnent l'échelle sans
// prétendre au sprite. Elles se dérivent de la tuile pour ne pas mentir si elle
// change.
func aplat(largeur, hauteur int) *ebiten.Image {
	img := ebiten.NewImage(largeur, hauteur)
	img.Fill(color.White)
	return img
}
