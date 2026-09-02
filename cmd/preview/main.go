// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La planche de relecture du rendu : les vues qu'elle écrit, le monde qu'elle
// pose pour chacune, et le détour par le cycle Ebitengine que le dessin impose.

// Preview écrit des vues du rendu dans `.tmp/apercus/`, à regarder.
//
// `internal/render` n'a pas de test et n'en aura pas : importer Ebitengine
// initialise GLFW, qui panique sans écran. Le rendu se juge donc à l'œil, et une
// planche est ce qui rend ce jugement possible ailleurs que dans une partie —
// on y compare une image d'avant et d'après un changement, ce qu'une observation
// au clavier ne permet pas.
//
// **Elle pilote le `render.Screen` du jeu, jamais une scène montée à côté.** Une
// planche qui dessinerait autrement que le jeu relirait la planche, et c'est le
// même piège qu'un test qui bâtit son entrée au lieu de passer par le chemin qui
// la produit.
//
// **Elle est déterministe, et doit le rester.** Deux exécutions écrivent des
// octets identiques, sans quoi comparer une planche d'avant et d'après un
// changement ne dirait rien. Rien n'y est tiré au sort : la horde vient du semis
// régulier du montage, et les vues qui jouent des pas les jouent avec une
// direction nulle. La graine, elle, est fixée d'avance, pour que la propriété
// tienne encore quand un tirage entrera dans la simulation.
//
// Elle exige un écran et ne tourne donc pas en intégration continue. Ce n'est
// pas un contrôle mais une planche : ce qui se vérifie mécaniquement vit du côté
// simulation, qui n'importe pas Ebitengine.
package main

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/sprimault/cohue"
	"github.com/sprimault/cohue/internal/game"
	"github.com/sprimault/cohue/internal/render"
	"github.com/sprimault/cohue/internal/session"
)

// sortie est le dossier des planches.
//
// Dans `.tmp/` et non dans `assets/`, dont le contrôle des ressources compare le
// contenu à ce que les générateurs produisent : une planche y deviendrait un
// écart à excepter, et elle pèserait dans le binaire de qui l'a régénérée.
const sortie = ".tmp/apercus"

// echelle agrandit le tampon avant l'écriture, en entier.
//
// **La planche montre ce que le joueur verra, et non le tampon nu.** Une image
// écrite à la taille du tampon se relit à la loupe, et une image agrandie d'un
// facteur qui n'est pas entier épaissit un pixel sur deux : un glyphe y paraît
// bancal alors que le jeu le rend net, et le jugement porte sur l'agrandissement
// plutôt que sur ce qu'on croit juger.
//
// Deux et non trois parce que c'est le facteur d'une fenêtre de 1080p, le cas
// courant. Ce que la fenêtre du jeu fera de son côté est un réglage d'affichage
// que l'étape 15 tranchera ; ici, on montre le tampon multiplié.
const echelle = 2

// graine est celle sur laquelle chaque vue monte sa partie.
//
// Fixée ici plutôt que reçue : la planche se compare d'une exécution à l'autre,
// et une graine qui varierait rendrait cette comparaison muette le jour où un
// tirage entre dans la simulation. C'est une exigence de la planche et non un
// réglage partagé avec le jeu, qui a la sienne pour une autre raison.
const graine uint64 = 1

// vue est une scène à écrire : où poser le joueur, combien de pas jouer avant de
// dessiner, ce qu'on pose par-dessus, et le nom du fichier qui en sort.
type vue struct {
	nom   string
	u, v  int
	ticks int
	// videLaHorde retire les créatures avant de jouer les pas.
	//
	// **Le seul moyen d'atteindre une montée de niveau aujourd'hui.** Avec le
	// semis provisoire, le joueur meurt vers six secondes en ayant ramassé au
	// plus six gemmes sur les dix du premier seuil : aucune position et aucune
	// direction de fuite n'ouvre un choix vivant. Ce n'est pas un défaut du choix
	// mais du semis, que la courbe de pression remplacera — et d'ici là, une vue
	// qui prétendrait obtenir le panneau en jouant montrerait un écran de mort.
	//
	// La salle vide reste la salle du jeu, dessinée par son écran : ce que la
	// planche met de côté est ce qui empêche d'y arriver, pas ce qu'elle donne à
	// juger.
	videLaHorde bool
	// jusquAuChoix arrête les pas dès qu'un choix s'ouvre, plutôt qu'au compte.
	// Le nombre de ticks devient alors un plafond : il tient la planche
	// déterministe même si l'équilibrage change ce qu'il faut de temps.
	jusquAuChoix bool
	// texte pose les échantillons nus, pour juger la police seule ; hud pose
	// l'interface, pour juger ce qui l'entoure. Les deux sont séparés parce
	// qu'un cadre sous un texte change ce qu'on lit de la police.
	texte bool
	hud   bool
}

// vues énumère ce que la planche donne à relire.
//
// Le centre montre la projection sans rien qui la borde ; les quatre suivantes
// posent le joueur près d'un coin du losange, là où la caméra bute et où le
// joueur se décentre. Ce sont les seuls endroits où le cadrage décide de quelque
// chose, et ils couvrent les deux axes dans les deux sens.
//
// **La mêlée est la seule qui juge l'exception du joueur**, et c'est pourquoi
// elle joue des pas : la horde semée converge, et il faut qu'elle soit arrivée
// pour qu'on voie si le personnage reste devant ce qui l'entoure. Une horde qui
// approche encore ne le dirait pas — elle est derrière lui ou devant lui, jamais
// autour. Quatre secondes suffisent à ce que les plus proches le rejoignent sans
// que l'arme de base en ait abattu assez pour dégager la place.
var vues = []vue{
	{nom: "centre", u: 16, v: 16},
	{nom: "nord", u: 2, v: 2},
	{nom: "ouest", u: 2, v: 29},
	{nom: "est", u: 29, v: 2},
	{nom: "sud", u: 29, v: 29},
	{nom: "melee", u: 16, v: 16, ticks: 4 * game.TPS},
	{nom: "texte", u: 16, v: 16, texte: true},
	{nom: "interface", u: 16, v: 16, hud: true},

	// **Le choix ne se pose pas non plus, il s'obtient** — comme la mort, et pour
	// la même raison : une maquette de cartes divergerait du panneau que le jeu
	// peint dès le premier déplacement de colonne. La horde est retirée parce
	// qu'elle rend la montée inatteignable, et c'est alors le plancher de temps
	// qui la donne. Le plafond de ticks est large : c'est le choix qui arrête les
	// pas, pas le compte, si bien que la vue tient encore quand le plancher change.
	{nom: "cartes", u: 16, v: 16, ticks: 90 * game.TPS,
		videLaHorde: true, jusquAuChoix: true},

	// La mort ne se pose pas, elle s'obtient : le joueur reste immobile au
	// milieu du lieu et la horde finit par l'avoir. Vingt secondes couvrent
	// largement la convergence puis les cinq secondes que le plafond de dégâts
	// impose — c'est la seule vue dont la scène est jouée plutôt que montée.
	{nom: "mort", u: 16, v: 16, ticks: 20 * game.TPS},
}

// echantillons sont les chaînes que la vue de texte affiche.
//
// **Ce sont des échantillons de mesure, pas les libellés du jeu.** Ni les cartes
// ni l'écran de mort n'existent encore ; ces chaînes sont ici parce qu'il faut
// du texte réel pour juger une police — des mots français avec leurs accents,
// des chiffres, un pourcentage, une durée. Le jour où l'étape 3 écrira les vrais
// libellés, ils viendront de là et cette table disparaîtra : deux listes de
// libellés divergeraient, et c'est celle-ci qu'on oublierait.
//
// Les espaces insécables y sont écrites par leur code : posées en littéral,
// elles se confondent avec des espaces ordinaires dans le source. Elles sont ce
// que le français impose devant un pourcentage et entre les milliers, et la vue
// est le premier usage réel du glyphe que la table déclare.
var echantillons = []string{
	"Niveau 5 " + string(rune(0x2014)) + " choisissez une amélioration",
	"Rafale",
	"+1 projectile",
	"Trois projectiles au lieu de deux.",
	"Cadence +15" + string(rune(0x00A0)) + "%",
	"Portée +2 tuiles",
	"L'arme tire plus souvent.",
	"Espace pour relancer",
	"ÀÉÈÊÇÎÔÙŸŒÆ «»",
}

// planche écrit toutes les vues au premier pas, puis demande l'arrêt.
//
// Elle est un jeu Ebitengine sans l'être : le dessin et la lecture de pixels ne
// sont valides que dans le cycle de la bibliothèque, si bien qu'un programme qui
// écrirait ses images depuis `main` n'obtiendrait rien. D'où ce détour, dont
// `Draw` reste vide — la fenêtre qui s'ouvre une fraction de seconde n'a rien à
// montrer.
//
// **Elle n'appelle jamais `render.Screen.Update`, et c'est délibéré.** Cette
// méthode lit les touches : une direction pressée pendant l'écriture déplacerait
// le joueur, et les images cesseraient d'être comparables d'une exécution à
// l'autre — le déterminisme reposerait sur la précaution de ne pas toucher au
// clavier plutôt que sur une propriété. Ne pas l'appeler n'est pas une garde
// qu'on pourrait contourner, c'est un chemin qui n'existe pas.
//
// Ce qu'elle en perd est nul : monter un écran cadre déjà sur le joueur, donc un
// écran monté après la scène est cadré juste. Ce qui doit avancer d'un pas passe
// par `World.Step`, où la direction est écrite et non lue.
//
// **Chaque vue monte sa propre partie**, et n'hérite donc pas de ce que les
// précédentes ont joué. Une vue qui avance de quatre secondes laisserait sinon
// une horde déplacée aux suivantes, et l'ordre de la table déciderait de ce
// qu'on voit — ce qui est exactement le genre de dépendance cachée qu'une
// planche ne doit pas avoir. Le montage coûte quelques millisecondes.
type planche struct {
	tampon *ebiten.Image
	// agrandi porte le tampon multiplié par `echelle`, et c'est lui qu'on écrit.
	agrandi *ebiten.Image
	hud     *render.HUD
	ecrit   bool
}

// Update écrit les vues, puis rend la fin de partie.
func (p *planche) Update() error {
	if p.ecrit {
		return ebiten.Termination
	}
	for _, v := range vues {
		if err := p.vue(v); err != nil {
			return err
		}
	}
	p.ecrit = true
	return nil
}

// Draw ne dessine rien : la fenêtre n'est ouverte que pour le contexte
// graphique.
func (p *planche) Draw(*ebiten.Image) {}

// Layout donne au tampon la taille de celui du jeu.
func (p *planche) Layout(_, _ int) (int, int) { return render.Width, render.Height }

// vue monte une partie, y pose la scène, la dessine et écrit le fichier.
//
// Le joueur est posé au centre de sa case et non sur son coin, faute de quoi la
// planche montrerait un cas qu'aucune partie ne produit. Les pas se jouent
// ensuite, avec une direction nulle : la horde avance, le joueur ne bouge pas, et
// rien de ce qui se passe ne dépend de ce qu'on presse. L'écran vient en dernier,
// puisque c'est son montage qui cadre.
func (p *planche) vue(v vue) error {
	partie, err := session.Open(cohue.Assets, cohue.StartingLevel, graine)
	if err != nil {
		return err
	}
	partie.World.Place(game.FromInt(v.u)+game.One/2, game.FromInt(v.v)+game.One/2)
	if v.videLaHorde {
		horde := partie.World.Enemies()
		for horde.Len() > 0 {
			horde.RemoveAt(0)
		}
	}

	// **La mort arrête les pas dès qu'une vue en dépend.** `World.Step` continue
	// de tourner après elle — c'est l'écran qui fige, et la planche l'appelle
	// directement —, si bien qu'un cadavre continue de tirer et de ramasser. La
	// vue du choix montrait ainsi un niveau gagné trente secondes après la mort.
	for range v.ticks {
		if v.jusquAuChoix && (partie.World.Choosing() || !partie.World.Alive()) {
			break
		}
		partie.World.Step(game.Vec{})
	}
	render.NewScreen(partie.World, partie.Grid, partie.Tile).WithHUD(p.hud).Draw(p.tampon)
	if v.texte {
		p.poser()
	}
	if v.hud {
		p.poserInterface()
	}

	// L'agrandissement se fait au plus proche voisin, qui est le filtre par
	// défaut : un lissage rendrait la planche inutilisable pour juger du pixel
	// art, qui est précisément ce qu'elle donne à relire.
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(echelle, echelle)
	p.agrandi.DrawImage(p.tampon, &op)

	chemin := filepath.Join(sortie, v.nom+".png")

	// Le chemin n'a aucune part variable : `sortie` est une constante et le nom
	// vient de la table ci-dessus. `os.Root` ne fermerait donc rien qui soit
	// ouvert, et l'ajouter pour taire l'avertissement mettrait un mécanisme là
	// où il n'y a pas de question.
	f, err := os.OpenFile(chemin, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600) // #nosec G304
	if err != nil {
		return err
	}

	// La fermeture est signalée, mais elle ne masque pas l'écriture : un encodage
	// qui échoue dit ce qui ne va pas, là où la fermeture d'un fichier déjà
	// fautif ne dirait que le symptôme.
	err = png.Encode(f, p.agrandi)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return fmt.Errorf("%s: %w", chemin, err)
	}
	fmt.Println(chemin)
	return nil
}

// Les teintes de la vue de texte, provisoires comme ses chaînes.
//
// Elles vivent ici et non dans le rendu pour la même raison que les
// échantillons : le jeu n'a pas encore d'interface, donc pas de teintes à en
// tirer. Quand il en aura, c'est de lui qu'elles viendront.
// poser écrit les échantillons sur la scène déjà dessinée.
//
// Trois situations, et ce sont elles qui font la vue : une colonne de texte nu
// sur le décor, deux chiffres contourés posés là où le fond est le plus clair,
// et deux lignes alignées sur le bord droit par mesure plutôt qu'à distance
// fixe. La troisième est ce qui manquait à la planche qui a fait écrire la règle
// — un minuteur placé au jugé déborde dès qu'il s'allonge.
func (p *planche) poser() {
	h := p.hud
	y := 12
	for _, s := range echantillons {
		h.Font.Draw(p.tampon, s, 12, y, h.Color("texte"))
		y += h.Font.Height() + 4
	}

	// Sur le monde et non sur un cadre : c'est le cas que la conception vise
	// quand elle exige un contour.
	//
	// **Le fond le plus hostile n'est pas à l'image, et ne peut pas y être
	// aujourd'hui.** Le contour existe pour un chiffre posé sur du décor clair,
	// or le rendu provisoire ne peint que trois gris et un bleu : le cas se
	// jugera quand les sprites entreront, à l'étape 5. Ce que la vue montre est
	// que le contour ne nuit pas sur fond moyen, pas qu'il suffit sur fond clair.
	contour := h.Color("texte_contour")
	h.Font.DrawOutlined(p.tampon, "247", render.Width/2-40, render.Height/2-60,
		h.Color("texte_valeur"), contour)
	h.Font.DrawOutlined(p.tampon, "12", render.Width/2+30, render.Height/2-20,
		h.Color("texte"), contour)

	for i, s := range []string{"07:41", "1" + string(rune(0x00A0)) + "340 pts"} {
		x := render.Width - 12 - h.Font.Advance(s)
		h.Font.Draw(p.tampon, s, x, 12+i*(h.Font.Height()+2), h.Color("texte"))
	}
}

// poserInterface compose l'écran de jeu tel que l'étape 3 le demandera.
//
// **Chaque élément se dimensionne sur son contenu**, jamais sur une constante :
// la carte prend la largeur de sa plus longue ligne, la case le côté de ce
// qu'elle contient, le minuteur sa place mesurée depuis le bord droit. C'est ce
// que la planche doit donner à juger — si une dimension y était écrite, on
// jugerait le chiffre plutôt que la règle qui le produit.
func (p *planche) poserInterface() {
	h := p.hud
	marge, hauteur := h.Margin(), h.Font.Height()

	// Le bandeau n'est pas remaquetté ici : `Screen.Draw` vient de le poser, avec
	// les valeurs de la partie montée. Une maquette qui le doublerait
	// superposerait deux jeux de chiffres — c'est ce qui est arrivé, et l'image
	// l'a montré tout de suite : les glyphes se chevauchaient au point de rendre
	// « 62 / 100 » illisible. Les vues qui le jugent sur des valeurs parlantes
	// sont `melee` et `mort`, qui jouent assez de ticks pour cela.

	// Les emplacements, sous les jauges. Le contenu vaut deux lignes faute
	// d'icône : elles viendront à l'étape 5, et c'est leur taille réelle qui
	// fixera le côté de la case — d'où un côté calculé plutôt que réglé.
	x := 12
	for _, touche := range []string{"1", "2"} {
		x += h.Slot(p.tampon, x, 62, hauteur*2, touche) + marge
	}

	// Le score, sous le minuteur du bandeau. Il n'est pas encore une lecture du
	// jeu — rien ne le compte —, et la planche le porte pour que sa place soit
	// jugée avec le reste.
	score := "1" + string(rune(0x00A0)) + "340 pts"
	h.Font.Draw(p.tampon, score, render.Width-12-h.Font.Advance(score),
		12+hauteur+2, h.Color("texte"))

	// Le panneau de choix n'est pas remaquetté : la vue `cartes` montre celui que
	// le jeu peint, sur une partie où le plancher de temps l'a ouvert. Une
	// maquette qui doublerait la mise en page livrée resterait juste jusqu'au
	// premier déplacement de colonne, puis montrerait un écran que personne ne
	// joue — c'est ce qui était arrivé au bandeau.
}

// main écrit la planche, ou sort en échec.
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run ouvre la fenêtre que le dessin exige, puis lance la planche.
//
// La partie, elle, se monte une fois par vue : voir la godoc de `planche`.
func run() error {
	if err := os.MkdirAll(sortie, 0o750); err != nil {
		return err
	}

	hud, err := render.LoadHUD(cohue.Assets, cohue.InterfaceManifest)
	if err != nil {
		return err
	}

	ebiten.SetWindowTitle("Cohue — planche")
	ebiten.SetWindowSize(render.Width, render.Height)
	return ebiten.RunGame(&planche{
		tampon:  ebiten.NewImage(render.Width, render.Height),
		agrandi: ebiten.NewImage(render.Width*echelle, render.Height*echelle),
		hud:     hud,
	})
}
