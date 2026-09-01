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
// direction nulle. Le jour où un tirage entrera dans la simulation, sa graine
// devra être fixée pour que cette propriété tienne.
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

// vue est une scène à écrire : où poser le joueur, combien de pas jouer avant de
// dessiner, ce qu'on pose par-dessus, et le nom du fichier qui en sort.
type vue struct {
	nom   string
	u, v  int
	ticks int
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

	// La mort ne se pose pas, elle s'obtient : le joueur reste immobile au
	// milieu du lieu et la horde finit par l'avoir. Vingt secondes couvrent
	// largement la convergence puis les cinq secondes que le plafond de dégâts
	// impose — c'est la seule vue dont la scène est jouée plutôt que montée.
	{nom: "mort", u: 16, v: 16, ticks: 20 * game.TPS},
}

// carte est une amélioration proposée, telle qu'une carte l'affiche.
//
// Trois lignes de nature différente — un nom, un effet chiffré, une phrase —
// parce que c'est ce qui décide de la largeur d'une carte et de la lisibilité
// d'une hiérarchie à cette taille.
type carte struct {
	nom, effet, phrase string
}

// cartes sont les échantillons du choix de niveau, provisoires comme les autres.
var cartes = []carte{
	{"Rafale", "+1 projectile", "Trois projectiles au lieu de deux."},
	{"Cadence", "+15" + string(rune(0x00A0)) + "%", "L'arme tire plus souvent."},
	{"Portée", "+2 tuiles", "Les tirs vont plus loin."},
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
	partie, err := session.Open(cohue.Assets, cohue.StartingLevel)
	if err != nil {
		return err
	}
	partie.World.Place(game.FromInt(v.u)+game.One/2, game.FromInt(v.v)+game.One/2)
	for range v.ticks {
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

	// Les jauges et leur état, en haut à gauche.
	p.poserJauges(12, 12)

	// Les emplacements, sous les jauges. Le contenu vaut deux lignes faute
	// d'icône : elles viendront à l'étape 5, et c'est leur taille réelle qui
	// fixera le côté de la case — d'où un côté calculé plutôt que réglé.
	x := 12
	for _, touche := range []string{"1", "2"} {
		x += h.Slot(p.tampon, x, 62, hauteur*2, touche) + marge
	}

	// Le minuteur et le score, alignés sur le bord droit par mesure.
	for i, s := range []string{"07:41", "1" + string(rune(0x00A0)) + "340 pts"} {
		h.Font.Draw(p.tampon, s, render.Width-12-h.Font.Advance(s),
			12+i*(hauteur+2), h.Color("texte"))
	}

	// Le choix occupe le bas de l'écran jusqu'au bord, et non un cadre flottant
	// au milieu. Deux raisons, dont la seconde est la vraie : une bande basse
	// laisse toute la hauteur utile au combat qui continue derrière, et le titre
	// y est lisible sans bandeau propre — posé à même le décor, il disparaissait
	// dès que le sol s'éclaircissait.
	titre := "Niveau 5 " + string(rune(0x2014)) + " choisissez une amélioration"
	panneau := 2*marge + hauteur + marge + p.hauteurCarte() + marge
	haut := render.Height - panneau
	h.Band(p.tampon, haut, panneau)
	h.Font.Draw(p.tampon, titre, (render.Width-h.Font.Advance(titre))/2, haut+marge,
		h.Color("texte"))

	p.poserCartes(haut + 2*marge + hauteur)
}

// hauteurCarte rend la hauteur d'une carte, qui vaut ses trois lignes.
//
// Elle est calculée et non réglée : le nombre de lignes d'une carte est ce qui
// la fixe, et une hauteur déclarée deviendrait fausse à la quatrième ligne sans
// que rien ne le dise.
func (p *planche) hauteurCarte() int {
	h := p.hud
	return 3*(h.Font.Height()+2) + 2*(h.Margin()+h.Border())
}

// poserJauges pose la vie et l'expérience, avec ce qu'elles valent.
func (p *planche) poserJauges(x, y int) {
	h := p.hud
	const largeur = 148

	h.Gauge(p.tampon, x, y, largeur, 0.62, h.Color("jauge_vie"))
	h.Font.Draw(p.tampon, "62 / 100", x+largeur+8, y-1, h.Color("texte"))

	y += h.Font.Height()
	h.Gauge(p.tampon, x, y, largeur, 0.35, h.Color("jauge_experience"))
	h.Font.Draw(p.tampon, "Niveau 4", x+largeur+8, y-1, h.Color("texte_attenue"))
}

// poserCartes pose les trois choix, chacun large de sa plus longue ligne.
func (p *planche) poserCartes(y int) {
	h := p.hud
	marge, hauteur := h.Margin(), h.Font.Height()
	interligne := hauteur + 2

	largeur := 0
	for _, c := range cartes {
		for _, ligne := range []string{c.nom, c.effet, c.phrase} {
			largeur = max(largeur, h.Font.Advance(ligne))
		}
	}
	largeur += 2 * (marge + h.Border())

	x := (render.Width - 3*largeur - 2*marge) / 2
	for _, c := range cartes {
		h.Frame(p.tampon, x, y, largeur, p.hauteurCarte())
		texte := x + marge + h.Border()
		ligne := y + marge + h.Border()
		h.Font.Draw(p.tampon, c.nom, texte, ligne, h.Color("texte"))
		h.Font.Draw(p.tampon, c.effet, texte, ligne+interligne, h.Color("texte_valeur"))
		h.Font.Draw(p.tampon, c.phrase, texte, ligne+2*interligne, h.Color("texte_attenue"))
		x += largeur + marge
	}
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
