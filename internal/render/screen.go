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
	"github.com/sprimault/cohue/internal/sprite"
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
	fond = color.RGBA{R: 24, G: 24, B: 28, A: 255}

	// La porte, dans ses deux états. Elle teinte le décor de sa case au lieu de
	// le remplacer, parce qu'elle n'y change rien : fermée elle est un mur comme
	// un autre, ouverte elle l'est encore. Ce que ces deux teintes disent est ce
	// qu'aucune forme ne porte — le lieu est gagné, et la sortie est là.
	//
	// **Une seule teinte, éteinte puis vive**, plutôt que deux couleurs : c'est
	// le même objet dans deux états, et l'écart de valeur se voit en périphérie
	// là où un changement de teinte demande de regarder. Un cyan, parce que le
	// vert appartient à l'éclair de soin et qu'ouvrir une porte n'a rien à voir
	// avec lui — le voisin le plus proche est la gemme, à 28,5, et elle est au
	// sol quand la porte est sur un mur.
	//
	// **Deux écarts de valeur à tenir, et le premier jet n'en a tenu qu'un.**
	// L'écart entre les deux états était bon ; celui entre la porte fermée et le
	// mur qui l'entoure ne l'était pas, si bien qu'une partie jouée n'a pas
	// trouvé la porte en faisant les quatre coins de la salle. Une teinte ne se
	// choisit pas contre l'autre état du même objet, elle se choisit **aussi**
	// contre ce qui l'entoure — ici le mur de l'enceinte, sur lequel la porte
	// est posée et que ces deux teintes multiplient.
	porteFermee  = color.RGBA{R: 48, G: 132, B: 152, A: 255}
	porteOuverte = color.RGBA{R: 120, G: 232, B: 248, A: 255}

	// intact est l'élément neutre de la multiplication qu'Ebitengine applique :
	// une image posée sous cette échelle sort telle qu'elle est dessinée. C'est
	// de quoi partir quand on veut retirer de la lumière à un sprite sans avoir
	// à en connaître les couleurs.
	intact = color.RGBA{R: 255, G: 255, B: 255, A: 255}

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

	// **Le contour du joueur, et sa silhouette : la même teinte.** C'est la même
	// chose qui le détache — cernée quand on le voit, pleine quand une forme ou
	// une foule le recouvre —, et deux couleurs pour un seul rôle donneraient
	// deux réglages à tenir d'accord.
	//
	// Un blanc à peine tiède plutôt que le blanc pur : le tir du joueur est déjà
	// un blanc bleuté, et le personnage doit s'en distinguer autant que du reste.
	// La luminance reste au-delà de 240, ce que la mesure demande — le pixel le
	// plus clair d'un profil monte à 162, celui d'un figurant à 175.
	teinteContour = color.RGBA{R: 252, G: 246, B: 226, A: 255}

	// **Le projectile de la horde porte une teinte qu'aucune autre ne dispute.**
	// C'est entre projectiles que la distinction doit être maximale : « est-ce
	// que ça me fait mal ? » ne se pose que sur eux, si bien que les confondre
	// coûte plus cher que de confondre un projectile et une créature.
	//
	// Elle ne sert plus qu'à sa silhouette, le sprite portant sa propre couleur.
	// Ce n'en est donc pas une copie : un aplat n'a pas de masse sombre ni de
	// liseré clair, et ce qu'on lui demande est de se voir à travers ce qui le
	// cache.
	teinteTirHorde = color.RGBA{R: 186, G: 138, B: 232, A: 255}

	// La teinte de l'aimant, qui ne sert plus qu'à la case du bandeau : au sol,
	// c'est son sprite qu'on voit. Elle survit parce que l'objet n'a pas d'icône
	// d'interface, et `HUD.emplacement` dit ce qu'il faudrait pour l'éteindre.
	teinteAimant = color.RGBA{R: 198, G: 126, B: 78, A: 255}
)

// Screen est le jeu tel qu'Ebitengine le voit.
type Screen struct {
	monde  *game.World
	carte  *game.CostGrid
	sol    *Terrain
	troupe *Cast
	objets *Stage
	cam    *camera

	scene *scene

	// face est le losange blanc que le dessin teinte au blit.
	//
	// **C'est le dernier aplat, et il ne disparaîtra pas** : il marque l'emprise
	// d'une explosion, et le télégraphe s'y peint en cases pleines parce que le
	// rendu n'a rien à agrandir — redimensionner une image par une fraction
	// casserait le pixel entier. Tous les autres ont cédé la place à des sprites.
	face *ebiten.Image

	// emprises porte, par case et le temps d'une image, la teinte du télégraphe
	// qui s'y peint — l'alpha nul disant qu'aucune explosion ne la couvre.
	//
	// **Un relevé par case plutôt qu'une passe par explosion**, parce que le
	// marquage appartient à la case et doit suivre sa forme dans la passe qui la
	// dessine. Peint d'un bloc à la fin du sol, il passait sous le trottoir et le
	// quai, qui se trient et se peignent après lui.
	//
	// Longue de la carte, allouée au montage : la fenêtre visible s'efface au
	// début de chaque image, et rien hors d'elle ne se dessine.
	emprises []color.RGBA

	// aRevoir porte, le temps d'une image, ce qu'on redessinera si quelque chose
	// le recouvre. Réutilisée avec `[:0]` : la séquence la remplit soixante fois
	// par seconde, et une tranche qui croîtrait ferait passer le ramasse-miettes
	// là où il se voit.
	aRevoir []revele

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
// Le décor apporte sa taille de tuile, celle du manifeste et jamais une
// constante : le chargeur en exige le rapport de deux pour un, et c'est de lui
// que la projection la tient. La recevoir à part du terrain qui la porte aurait
// laissé deux appelants la prendre à deux endroits.
func NewScreen(monde *game.World, carte *game.CostGrid, sol *Terrain, troupe *Cast,
	objets *Stage) *Screen {
	tuile := sol.TileSize()
	cam := nouvelleCamera(tuile, carte)
	s := &Screen{
		monde:     monde,
		carte:     carte,
		sol:       sol,
		troupe:    troupe,
		objets:    objets,
		cam:       cam,
		scene:     nouvelleScene(carte, monde, sol, cam),
		face:      face(tuile),
		demiTuile: tuile[0] / 2,
		// Le joueur et, au pire, tous les projectiles de la horde en vol : c'est
		// exactement ce que la conception révèle, donc le majorant est le bassin
		// lui-même.
		aRevoir:  make([]revele, 0, 1+monde.EnemyShots().Cap()),
		emprises: make([]color.RGBA, carte.Width()*carte.Height()),
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
	s.releverEmprises()
	s.peindreSol(ecran)
	s.peindreEntites(ecran)
	s.peindreEffets(ecran)
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

// peindreSol pose le décor de chaque case visible qui ne dépasse pas du sol.
//
// **Ce qui dépasse n'est pas ici mais dans la séquence**, où il dispute sa
// profondeur au reste. Un sol, lui, est toujours dessous : l'ordre entre deux
// cases plates ne décide de rien, puisqu'elles pavent sans se recouvrir, et le
// balayage par rangées suffit.
//
// **Le tri par coût que ce sol montrait a disparu avec l'aplat**, et rien ne
// s'est perdu à le retirer : la grille descend désormais de la même carte de
// formes que ces images, si bien qu'un écart entre ce qu'on voit et ce que le
// champ de flux lit n'est plus exprimable.
func (s *Screen) peindreSol(ecran *ebiten.Image) {
	u0, v0, u1, v1 := s.cam.casesVisibles()
	for v := v0; v <= v1; v++ {
		for u := u0; u <= u1; u++ {
			if f, posee := s.sol.formeDe(u, v); posee && !f.triee {
				s.poserCase(ecran, u, v, f)
			}
		}
	}
}

// poserCase pose l'image d'une case, à l'ancrage que son manifeste lui donne, le
// sol du thème sous ce qui ne remplit pas son losange, et le télégraphe d'une
// explosion par-dessus.
//
// **Le sol vient juste avant la forme et non dans une passe à part.** Un pilier
// se trie, un rail ne se trie pas, et ce qui les comble doit suivre chacun dans
// sa passe : une passe de sol posée d'un bloc peindrait par-dessus les cases
// déjà triées de la même image.
//
// **La porte garde ses deux teintes, appliquées à ce que le décor y dessine.**
// Aucune forme du lieu ne dit qu'une case est la sortie — l'ouverture est un
// état de partie, et le décor n'en sait rien —, or une partie jouée a montré
// qu'on fait les quatre coins d'une salle sans trouver une porte qui ressemble
// au mur qui l'entoure. La teinte multiplie l'image au lieu de la remplacer :
// le mur reste un mur, et il vire au cyan. Le sol n'en prend rien — ce qu'elle
// désigne est la porte, pas la case qui la porte.
//
// **Le télégraphe d'une explosion suit la surface qu'il marque**, donc il
// s'intercale et ne se peint pas d'un bloc. Une forme qui remplit son losange
// *est* la surface, et le marquage se pose dessus, à son élévation ; une forme
// qui ne la remplit pas la traverse, et c'est le sol du thème, sous elle, qu'on
// marche. La passe unique qu'il remplaçait tombait sous le trottoir et le quai,
// qui ont une élévation et se trient donc.
func (s *Screen) poserCase(ecran *ebiten.Image, u, v int, f forme) trace {
	x, y := s.cam.ecran(game.FromInt(u), game.FromInt(v))
	if f.nue {
		if s.sol.sol != nil {
			s.poser(ecran, x, y, *s.sol.sol)
		}
		s.marquerEmprise(ecran, u, v, f)
	}

	s.op.ColorScale.Reset()
	if sortie := s.monde.Exit(); sortie != nil && sortie.U == u && sortie.V == v {
		if s.monde.DoorOpen() {
			s.op.ColorScale.ScaleWithColor(porteOuverte)
		} else {
			s.op.ColorScale.ScaleWithColor(porteFermee)
		}
	}
	s.op.GeoM.Reset()
	s.op.GeoM.Translate(float64(x+f.dx), float64(y+f.dy))
	ecran.DrawImage(f.image, &s.op)

	if !f.nue {
		s.marquerEmprise(ecran, u, v, f)
	}
	return trace{x: x + f.dx, y: y + f.dy, masque: f.masque, cache: f.cache}
}

// ecranAuSol projette une position du monde en la posant sur la surface qu'on y
// foule, et non sur le plan du sol.
//
// **Un trottoir a une marche, et ce qui marche dessus est six pixels plus haut.**
// La projection seule ignore l'élévation — elle ne connaît que deux axes —, si
// bien qu'un personnage sur une bordure était dessiné à l'intérieur : le décor
// perdait sa matière, il n'arrêtait rien et ne portait rien.
//
// **Le passage d'un niveau à l'autre saute, il ne s'interpole pas.** Une montée
// progressive ferait gravir une pente invisible là où le dessin montre une
// marche franche ; six pixels d'un coup sont ce qu'un pas sur un trottoir fait.
func (s *Screen) ecranAuSol(x, y game.Fixed) (int, int) {
	ex, ey := s.cam.ecran(x, y)
	return ex, ey - s.sol.hauteurSolEn(x, y)
}

// poser pose une forme sans la teinter, au coin que son ancrage lui donne.
func (s *Screen) poser(ecran *ebiten.Image, x, y int, f forme) {
	s.op.GeoM.Reset()
	s.op.GeoM.Translate(float64(x+f.dx), float64(y+f.dy))
	s.op.ColorScale.Reset()
	ecran.DrawImage(f.image, &s.op)
}

// releverEmprises note, case par case, ce qu'une explosion amorcée va couvrir.
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

	souffles := s.monde.Blasts()
	for i := range souffles.Active() {
		b := souffles.At(i)

		// L'imminence monte quand la mèche descend : le rapport rendu est ce
		// qu'il reste, et c'est son complément qui s'affiche.
		imminence := float32(1000-s.monde.FuseLeft(b)) / 1000
		vif := attenuer(teinteEmprise, emprisePlancher+(1-emprisePlancher)*imminence)

		bu0, bv0, bu1, bv1 := s.monde.BlastBounds(b)
		for v := max(bv0, v0); v <= min(bv1, v1); v++ {
			for u := max(bu0, u0); u <= min(bu1, u1); u++ {
				if !s.carte.InBounds(u, v) || !s.monde.BlastCovers(b, u, v) {
					continue
				}
				s.emprises[v*largeur+u] = vif
			}
		}
	}
}

// marquerEmprise pose sur une case le télégraphe qu'une explosion y a relevé.
//
// **Le marquage monte à la hauteur de ce qu'on marche.** Sur un trottoir ou un
// quai, la face supérieure est la surface, et le losange s'y superpose au pixel
// près — l'élévation que le générateur écrit *est* la distance entre le plan du
// sol et cette face. Posé à plat, il s'afficherait au pied de la marche.
//
// Sur un rail ou une porte ouverte, `hauteurSol` vaut zéro : ce qu'on y marche
// est le sol du thème, et l'appelant peint alors le marquage avant la forme,
// qui doit continuer de passer devant.
func (s *Screen) marquerEmprise(ecran *ebiten.Image, u, v int, f forme) {
	vif := s.emprises[v*s.sol.carte.Width()+u]
	if vif.A == 0 {
		return
	}
	x, y := s.cam.ecran(game.FromInt(u), game.FromInt(v))
	s.op.GeoM.Reset()
	s.op.GeoM.Translate(float64(x-s.demiTuile), float64(y-f.hauteurSol))
	s.op.ColorScale.Reset()
	s.op.ColorScale.ScaleWithColor(vif)
	ecran.DrawImage(s.face, &s.op)
}

// peindreEntites pose ce qui se tient sur le sol, du plus lointain au plus
// proche.
//
// La position est relue dans le bassin plutôt que portée par la séquence : le
// tri range des rangs, pas des coordonnées, et une copie faite au moment du tri
// aurait une image de retard le jour où quelque chose bougera entre les deux.
func (s *Screen) peindreEntites(ecran *ebiten.Image) {
	largeur := s.sol.carte.Width()
	s.aRevoir = s.aRevoir[:0]
	for _, e := range s.scene.ranger(s.monde) {
		var t trace
		switch e.sorte {
		case sorteDecor:
			u, v := e.place%largeur, e.place/largeur
			f, _ := s.sol.formeDe(u, v)
			t = s.poserCase(ecran, u, v, f)
		case sorteEnnemi:
			c := s.monde.Enemies().At(e.place)
			f := s.troupe.ennemis[c.Profile]
			t = s.peindreCreature(ecran, f, f.poseEnnemi(c, s.regardDe(c)),
				c.X, c.Y, e.identite, etatDe(c))
			if c.Flash > 0 {
				s.peindreEtincelle(ecran, t, c.Flash)
			}
		case sorteAmbiance:
			a := s.monde.Ambients().At(e.place)
			f := s.troupe.ambiants[a.Profile]
			t = s.peindreCreature(ecran, f, f.posePersonnage(a.Step, a.Step, 0),
				a.X, a.Y, e.identite, nil)
		case sorteTir:
			p := s.monde.Shots().At(e.place)
			t = s.peindreObjet(ecran, objetTir, p.X, p.Y, e.identite, nil)
		case sorteTirHorde:
			p := s.monde.EnemyShots().At(e.place)
			t = s.peindreObjet(ecran, objetTirHorde, p.X, p.Y, e.identite, nil)
		case sorteGemme:
			g := s.monde.Gems().At(e.place)
			// **Une gemme attirée reprend sa teinte pleine.** L'extinction dit
			// « ceci va disparaître » ; une gemme que l'aimant tient ne
			// disparaîtra pas, et la montrer éteinte serait montrer une
			// information fausse. Accessoirement, une ruée de gemmes anciennes
			// serait un feu d'artifice en gris — l'inverse de ce que la
			// conception appelle le moment de plaisir maximal du genre.
			//
			// **Sur un dessin, l'extinction multiplie**, et c'est ici le bon
			// sens : ce qu'elle dit est que la gemme s'efface, donc qu'elle perd
			// de sa lumière. C'est l'inverse d'un éclair d'état, qui ajoute.
			var voile *color.RGBA
			if !g.Pulled {
				eteinte := eteindre(intact, s.monde.GemAge(g), s.monde.GemLife())
				voile = &eteinte
			}
			t = s.peindreObjet(ecran, objetGemme, g.X, g.Y, e.identite, voile)
		case sorteAimant:
			a := s.monde.Magnets().At(e.place)
			t = s.peindreObjet(ecran, objetAimant, a.X, a.Y, e.identite, nil)
		case sorteCaisse:
			c := s.monde.Crates().At(e.place)
			t = s.peindreObjet(ecran, objetCaisse, c.X, c.Y, e.identite, nil)
		case sorteJoueur:
			x, y := s.monde.Player()
			f := s.troupe.joueur
			pas := s.monde.PlayerStep()
			t = s.peindreCreature(ecran, f,
				f.posePersonnage(pas, s.regardDuJoueur(pas), 0), x, y, 0, nil)
		}

		// **Le recouvrement se constate en dessinant, dans l'ordre où l'on
		// dessine.** Ce qui masque une chose est ce qui vient après elle, et la
		// séquence le dit déjà : il suffit de retenir la forme de ce qu'on
		// révélera, puis de confronter chaque dessin suivant. C'est ce que la
		// conception annonçait — « savoir quelle entité est masquée, c'est-à-dire
		// ce que le tri en profondeur calcule ».
		//
		// Le test précède l'inscription, sans quoi une chose se masquerait
		// elle-même.
		for i := range s.aRevoir {
			if !s.aRevoir[i].couvert && s.aRevoir[i].recouvertPar(t) {
				s.aRevoir[i].couvert = true
			}
		}
		if t.forme != nil {
			s.aRevoir = append(s.aRevoir, revele{
				forme: t.forme, masque: t.masque, x: t.x, y: t.y,
				teinte: teinteDeLaSorte(e.sorte),
			})
		}
	}
	s.reveler(ecran)
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
func (s *Screen) reveler(ecran *ebiten.Image) {
	for _, r := range s.aRevoir {
		if !r.couvert {
			continue
		}
		s.op.GeoM.Reset()
		s.op.GeoM.Translate(float64(r.x), float64(r.y))
		s.op.ColorScale.Reset()
		s.op.ColorScale.ScaleWithColor(r.teinte)
		ecran.DrawImage(r.forme, &s.op)
	}
}

// teinteDeLaSorte rend la couleur qu'une silhouette prend.
//
// Deux sortes, deux teintes réservées. Le joueur reprend celle de son contour :
// c'est la même chose qui le détache, cernée quand on le voit et pleine quand on
// ne le voit plus. Le projectile de la horde garde le violet que le chapitre 14
// lui réserve — « une couleur qui n'existe nulle part ailleurs dans le
// catalogue ».
func teinteDeLaSorte(quoi sorte) color.RGBA {
	if quoi == sorteTirHorde {
		return teinteTirHorde
	}
	return teinteContour
}

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

// peindreCreature pose l'image d'un personnage, son appui sur le point où le
// monde le situe.
//
// **L'image se dérive, elle ne se stocke pas.** Un cycle qui boucle se cadence
// sur le tick, décalé par l'identifiant de l'entité pour qu'une horde ne marche
// pas au pas ; un cycle qui s'achève se cadence sur le décompte de l'état qui le
// porte. Aucune des deux ne demande de compteur, donc aucune ne pose la question
// de savoir qui l'avance ni s'il entre dans l'empreinte d'une run.
//
// **L'orientation et la teinte viennent de la pose**, que l'appelant a bâtie :
// elles dépendent l'une de ce que le personnage regarde, l'autre de ce que son
// apparition a tiré, et aucune des deux n'est une propriété de son dessin.
// Le second retour est l'aplat de la pose posée, nul pour qui n'en a pas : seul
// le joueur en porte, et c'est ce qui décide de son contour comme de sa
// silhouette. L'appelant n'a donc pas à savoir lequel des personnages se révèle.
func (s *Screen) peindreCreature(ecran *ebiten.Image, f *figure, a anim,
	x, y game.Fixed, identite int, eclat *color.RGBA) trace {
	if a.nom == "" || a.direction == "" {
		return trace{}
	}

	i := sprite.Loop(a.cycle, s.monde.Tick(), identite)
	if !a.cycle.Loop {
		i = sprite.Once(a.cycle, a.reste)
	}
	p := pose{a.nom, a.direction, a.variante, i}
	img := f.image(p)
	if img == nil {
		return trace{}
	}

	ex, ey := s.ecranAuSol(x, y)
	coinX, coinY := ex-f.appui[0], ey-f.appui[1]

	// Le contour vient avant le sprite, qui le recouvre en son centre : ce qui
	// dépasse est le liseré, et il n'a pas d'autre épaisseur que ce dépassement.
	forme := f.forme(p)
	if forme != nil {
		s.contour(ecran, forme, coinX, coinY, teinteContour)
	}

	s.op.GeoM.Reset()
	s.op.GeoM.Translate(float64(coinX), float64(coinY))
	s.op.ColorScale.Reset()
	ecran.DrawImage(img, &s.op)

	if eclat != nil {
		s.eclairer(ecran, img, *eclat)
	}
	// **Une créature cache toujours, sans que le manifeste ait à le dire.** Elle
	// fait la taille d'un personnage par définition, et le chapitre 2 nomme
	// précisément ce cas : « ce qu'un mur ou une foule recouvre ».
	return trace{x: coinX, y: coinY, masque: f.masque(p), cache: true, forme: forme}
}

// intensiteEclair est la part de sa teinte qu'un éclair d'état ajoute au sprite.
//
// **Une valeur à juger à l'œil, et le premier chiffre est bas exprès.** Ce que
// l'éclair doit dire est « celle-ci vient d'être atteinte », pas « celle-ci est
// blanche » : au-delà, une mêlée dense devient une nappe claire où les
// silhouettes se perdent, ce qui coûte la lisibilité qu'on cherchait.
const intensiteEclair = 0.45

// eclairer rajoute la teinte d'un état par-dessus le sprite déjà posé.
//
// **Une passe additive et non une teinte multipliée**, et c'est ce que les
// sprites ont changé. Sur un aplat blanc, multiplier *était* colorer ; sur un
// dessin, multiplier par un rose presque blanc ne change à peu près rien, et
// l'éclair d'impact — le retour qui manque le plus quand on tire sans le voir —
// avait disparu sans qu'aucun test ne puisse le dire.
//
// La géométrie est celle de la passe précédente, laissée en place : ce qui
// s'ajoute doit se superposer au pixel près, et la recalculer ouvrirait la
// possibilité qu'elle diverge.
//
// Le mélange revient à sa valeur nulle en sortant, qui est l'alpha ordinaire :
// l'option est réutilisée d'un blit à l'autre, et un mélange laissé additif
// éclaircirait tout ce que l'image dessine ensuite.
func (s *Screen) eclairer(ecran, forme *ebiten.Image, teinte color.RGBA) {
	s.op.Blend = ebiten.BlendLighter
	s.op.ColorScale.Reset()
	s.op.ColorScale.ScaleWithColor(teinte)
	s.op.ColorScale.ScaleAlpha(intensiteEclair)
	ecran.DrawImage(forme, &s.op)
	s.op.Blend = ebiten.Blend{}
}

// etatDe rend la teinte que l'état d'une créature ajoute, ou rien.
//
// **Les éclairs d'état survivent aux sprites, la table par profil non.** Une
// couleur par profil disait qui était quoi, ce que le dessin dit désormais
// mieux ; un impact ou un soigneur qui s'allume disent ce qui vient de se
// passer, et aucune pose ne le dirait à cent créatures à l'écran. Les premiers
// s'éteignent donc avec la dérogation, les seconds restent.
func etatDe(e *game.Enemy) *color.RGBA {
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

// peindreObjet pose un objet du catalogue, son ancrage sur le point où le monde
// le situe.
//
// Le voile multiplie l'image quand il est donné : c'est ce que veut une
// extinction, qui retire de la lumière, et l'inverse de l'éclair d'un état, qui
// en ajoute.
// Le second retour est l'aplat de l'image posée, nul pour qui n'en a pas : seul
// le projectile de la horde en porte.
func (s *Screen) peindreObjet(ecran *ebiten.Image, nom string, x, y game.Fixed,
	identite int, voile *color.RGBA) trace {
	objet, img := s.objets.image(nom, s.monde.Tick(), identite)
	i := sprite.Loop(objet.cycle, s.monde.Tick(), identite)
	t := s.poserObjet(ecran, objet, img, x, y, voile)
	t.masque, t.forme, t.cache = objet.masque(i), objet.forme(i), objet.cache
	return t
}

// peindreEffets pose ce qui reste d'une chose qui n'existe plus.
//
// **Après la séquence et sans y entrer.** Un éclat vole au-dessus du sol et ne
// dispute sa profondeur à personne : le trier lui vaudrait de passer sous la
// créature qui piétine la caisse, ce qui est exactement l'inverse de ce qu'il
// doit montrer. Il en va de même de l'onde d'une déflagration, qui recouvre par
// définition ce qu'elle emporte.
//
// **Le bassin ne porte qu'un point et un décompte** ; la volée, ses directions
// et sa parabole se dérivent du rang de chaque éclat. C'est ce qui permet à
// huit éclats de ne rien coûter à la simulation, et de changer d'apparence sans
// qu'elle bouge.
func (s *Screen) peindreEffets(ecran *ebiten.Image) {
	effets := s.monde.Fxs()
	for i := range effets.Active() {
		e := effets.At(i)
		age := e.Total - e.Life
		switch e.Kind {
		case game.FxBlast:
			objet, img := s.objets.effet(objetSouffle, e.Life)
			s.poserObjet(ecran, objet, img, e.X, e.Y, nil)
		case game.FxCrate:
			s.peindreVolee(ecran, objetEclatsCaisse, e.X, e.Y, age, e.Total)
		}
	}
}

// peindreVolee ouvre une gerbe d'éclats depuis le point où quelque chose a cédé.
//
// La hauteur se retranche à l'ordonnée d'écran : monter est aller vers le haut,
// et c'est la seule endroit du rendu où l'élévation d'une chose se calcule au
// lieu de venir d'un manifeste — un éclat n'a pas d'ancrage, il a une
// trajectoire.
func (s *Screen) peindreVolee(ecran *ebiten.Image, nom string, x, y game.Fixed,
	age, total game.Tick) {
	objet, connu := s.objets.objets[nom]
	if !connu || len(objet.images) == 0 {
		return
	}

	for rang := range sprite.Shards {
		ecart, hauteur := sprite.Shard(rang, age, total)
		ex, ey := s.ecranAuSol(x+ecart.X, y+ecart.Y)
		s.op.GeoM.Reset()
		s.op.GeoM.Translate(float64(ex+objet.dx), float64(ey+objet.dy-hauteur))
		s.op.ColorScale.Reset()
		ecran.DrawImage(objet.images[rang%len(objet.images)], &s.op)
	}
}

// peindreEtincelle marque un tir qui vient de porter.
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
// la cible. Le milieu se prend sur la forme réellement dessinée plutôt que sur
// la bande, dont les trois quarts sont transparents.
//
// Une méthode pour un seul effet, et non une machinerie pour la famille : le
// souffle d'une déflagration est le second, il vient avec son bassin, et c'est
// alors qu'on saura ce que les deux partagent.
func (s *Screen) peindreEtincelle(ecran *ebiten.Image, sur trace, reste game.Tick) {
	_, img := s.objets.effet(objetEtincelle, reste)
	if img == nil || sur.masque == nil {
		return
	}

	corps := sur.masque.Bounds()
	taille := img.Bounds().Size()
	s.op.GeoM.Reset()
	s.op.GeoM.Translate(
		float64(sur.x+corps.Min.X+corps.Dx()/2-taille.X/2),
		float64(sur.y+corps.Min.Y+corps.Dy()/2-taille.Y/2))
	s.op.ColorScale.Reset()
	ecran.DrawImage(img, &s.op)
}

// poserObjet blitte une image d'objet au décalage que son manifeste lui donne.
func (s *Screen) poserObjet(ecran *ebiten.Image, objet prop, img *ebiten.Image,
	x, y game.Fixed, voile *color.RGBA) trace {
	if img == nil {
		return trace{}
	}

	ex, ey := s.ecranAuSol(x, y)
	coinX, coinY := ex+objet.dx, ey+objet.dy
	s.op.GeoM.Reset()
	s.op.GeoM.Translate(float64(coinX), float64(coinY))
	s.op.ColorScale.Reset()
	if voile != nil {
		s.op.ColorScale.ScaleWithColor(*voile)
	}
	ecran.DrawImage(img, &s.op)
	return trace{x: coinX, y: coinY}
}

// silhouette pose un aplat blanc dans la teinte donnée, son pied sur le point où
// le monde le situe.
//
// **Un aplat blanc, et la précision n'est pas une redondance.** La teinte
// multiplie l'image : sur du blanc, multiplier *est* colorer, et la fonction fait
// ce que son nom promet. Passez-lui un dessin et elle ne le teinte plus, elle
// l'assombrit — c'est exactement ce qui a fait disparaître l'éclair d'impact le
// jour où les créatures ont eu leurs bandes, sans qu'une ligne d'ici ne bouge.
// Ce qui s'ajoute à un dessin passe par `eclairer`, en mélange additif.
//
// L'appui est au milieu du bas : c'est le point qui touche le sol, et le seul
// qui puisse coïncider avec une position du monde.
func (s *Screen) silhouette(ecran, forme *ebiten.Image, x, y game.Fixed, teinte color.RGBA) {
	ex, ey := s.ecranAuSol(x, y)
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
func attenuer(teinte color.RGBA, part float32) color.RGBA {
	return color.RGBA{
		R: uint8(float32(teinte.R) * part),
		G: uint8(float32(teinte.G) * part),
		B: uint8(float32(teinte.B) * part),
		A: uint8(float32(teinte.A) * part),
	}
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
