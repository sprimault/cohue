// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// L'ouverture d'un dossier de lieu et l'assemblage de sa grille de coûts, plus les
// refus que l'appelant peut vouloir distinguer. Le nom du dossier et
// l'identifiant doivent s'accorder, ce qui attrape la copie qu'on a renommée
// sans toucher au champ.

package level

import (
	"errors"
	"fmt"
	"io/fs"
	"path"

	"github.com/sprimault/cohue/internal/game"
	"github.com/sprimault/cohue/internal/manifest"
)

// Les refus que l'appelant peut vouloir distinguer.
var (
	// ErrUnknownRoom signale un lieu qui cite une pièce absente du jeu.
	ErrUnknownRoom = errors.New("piece inconnue du jeu")
	// ErrEmptyLevel signale un lieu sans aucune pièce posée.
	ErrEmptyLevel = errors.New("lieu sans piece")
)

// Formats lus par ce binaire. Chacun a sa vie : un lieu circule entre joueurs,
// une pièce et un jeu restent dans le binaire.
const (
	FormatLevel = 1
	FormatRoom  = 1
	FormatSet   = 1
)

// Les noms fixes d'un dossier de lieu, et le sous-dossier des pièces.
//
// **Ce qui existe une fois par dossier porte un nom fixe ; ce qui existe en
// plusieurs exemplaires garde son identifiant et se range dans un sous-dossier
// qui dit sa nature.** Sans cette règle, le jeu de pièces et une pièce sont deux
// noms libres au même niveau : `quartier.json` posé à côté de `carrefour.json`
// ne dit pas lequel est une palette et lequel est un plan, et « quartier » se lit
// même comme un endroit qu'on construirait.
//
// L'identité ne se perd pas pour autant, elle vit dans le champ `identifiant` —
// comme un lieu nommé « place » n'a jamais eu de fichier `place.json`.
const (
	LevelFile = "lieu.json"
	SetFile   = "jeu.json"
	RoomsDir  = "pieces"
)

// Loader lit un lieu et ses pièces dans un système de fichiers.
//
// Une interface de lecture plutôt qu'un chemin : les lieux livrés viennent d'un
// `embed.FS`, ceux d'un joueur d'un dossier, et les cas de test d'une carte en
// mémoire. Un seul chemin de code pour les trois, ce que la conception exige —
// une voie rapide pour le contenu livré ne serait exercée qu'à moitié.
type Loader struct {
	fsys fs.FS
	// decor est le manifeste des formes. Le chargeur ne connaît aucun nom de
	// tuile en dur : il y lit ce que coûte une traversée, et si une forme
	// remplit sa case.
	//
	// **Le manifeste entier et non le seul catalogue de coûts**, depuis qu'un
	// thème doit déclarer un sol quand sa palette emploie une forme qui laisse
	// du losange nu. Les deux dérivent du même fichier ; en recevoir un et
	// dériver l'autre aurait fait entrer deux descriptions par la porte du
	// paramètre.
	decor *Decor
	// couts est dérivé de `decor`, une fois, parce que l'assemblage le consulte
	// par case.
	couts map[string]Footing
	// profils sert à résoudre les profils qu'un scénario de vagues autorise.
	//
	// Il entre ici pour la même raison que le catalogue de coûts : un lieu cite
	// des noms, et refuser ceux qui n'existent pas fait partie de sa validation.
	// Les résoudre plus tard aurait rendu à l'auteur deux listes de manquements
	// au lieu d'une.
	profils *game.Profiles
	// report est la borne du budget de pression reporté d'un tick au suivant.
	//
	// Elle vient de la progression et non du lieu, et elle entre pourtant dans sa
	// validation : une phase qui autorise un profil coûtant plus que ce plafond
	// ne le fera jamais apparaître. Le refus appartient au fichier — c'est le
	// scénario qui est mal formé, quelle que soit la partie qui le monte —, donc
	// il se fait ici plutôt qu'au montage.
	report game.Tick
	// obstacles sont les sortes d'obstacles fragiles du catalogue des objets,
	// qu'un lieu cite par leur nom. Elles entrent pour la raison des profils :
	// refuser un nom inconnu fait partie de la validation du fichier.
	obstacles []game.BreakableKind
}

// NewLoader monte un chargeur sur un système de fichiers et les catalogues
// qu'un lieu cite : les formes du décor, les profils de créatures et les
// obstacles fragiles.
func NewLoader(fsys fs.FS, decor *Decor, profils *game.Profiles, report game.Tick,
	obstacles []game.BreakableKind) *Loader {
	return &Loader{
		fsys:      fsys,
		decor:     decor,
		couts:     decor.Costs(),
		profils:   profils,
		report:    report,
		obstacles: obstacles,
	}
}

// Load lit le lieu que porte un dossier, ses pièces et son jeu, puis les
// assemble en grille de coûts.
//
// Un dossier, et non un fichier : c'est ce qui donne aux pièces un espace de
// noms local. À plat, deux auteurs qui nomment chacun leur pièce « carrefour »
// s'écraseraient, et un lieu ne pourrait pas circuler sans emporter le
// vocabulaire de tous les autres.
//
// L'ordre importe : décoder, valider, assembler. Le décodage s'arrête au premier
// écart — `encoding/json` ne sait pas faire autrement —, la validation liste
// tout ce qui manque en une fois, parce que c'est là que l'aller-retour coûte à
// qui met au point un niveau.
func (l *Loader) Load(dossier string) (*Loaded, error) {
	nom := path.Base(dossier)
	if nom == "." || nom == "/" {
		return nil, fmt.Errorf("%q : un lieu se charge par son dossier, qui porte son nom", dossier)
	}

	chemin := path.Join(dossier, LevelFile)
	lieu, err := manifest.Decode[Level](l.fsys, chemin)
	if err != nil {
		return nil, err
	}
	if lieu.Format != FormatLevel {
		return nil, fmt.Errorf("%s: %w : %d, ce binaire lit la %d",
			chemin, manifest.ErrUnsupportedFormat, lieu.Format, FormatLevel)
	}
	if len(lieu.Placements) == 0 {
		return nil, fmt.Errorf("%s: %w", chemin, ErrEmptyLevel)
	}

	jeu, err := manifest.Decode[Set](l.fsys, path.Join(dossier, SetFile))
	if err != nil {
		return nil, err
	}
	if jeu.Format != FormatSet {
		return nil, fmt.Errorf("%w : jeu de pieces en %d", manifest.ErrUnsupportedFormat, jeu.Format)
	}

	pieces := make([]*Room, 0, len(lieu.Placements))
	for _, pose := range lieu.Placements {
		piece, err := manifest.Decode[Room](l.fsys, path.Join(dossier, RoomsDir, pose.RoomID+".json"))
		if err != nil {
			return nil, fmt.Errorf("%w : %s", ErrUnknownRoom, pose.RoomID)
		}
		pieces = append(pieces, piece)
	}

	// La géométrie et la courbe de pression se valident ensemble et se refusent
	// ensemble : ce sont deux moitiés du même fichier, et rendre la seconde après
	// avoir corrigé la première ferait un aller-retour de plus.
	// **L'assemblage précède la dernière validation**, et l'ordre annoncé plus
	// haut s'entend donc « décoder, valider, assembler, valider ce qui a besoin
	// de la carte ». Il ne peut pas échouer — la grille est dimensionnée sur les
	// placements qu'elle recopie —, si bien que la faire tôt ne coûte rien et
	// donne aux positions de figurants la seule chose qui permette de les
	// refuser : une carte où lire la passabilité.
	tuiles := assembler(lieu, pieces, jeu)
	grille := tuiles.couts(l.couts)

	scenario, ecarts := game.CompileScenario(lieu.Waves, l.profils, l.report)
	ambiance, ecartsAmbiance := game.CompileAmbient(lieu.Ambient, l.profils, grille)
	sortie, ecartsSortie := game.CompileExit(lieu.Exit, grille)
	caisses, ecartsCaisses := game.CompileCrates(lieu.Crates, grille)
	obstacles, ecartsObstacles := game.CompileBreakables(lieu.Breakables, l.obstacles,
		grille, caisses, ambiance)
	manques := append(valider(nom, lieu, jeu, pieces, l.decor), ecarts...)
	manques = append(manques, ecartsAmbiance...)
	manques = append(manques, ecartsSortie...)
	manques = append(manques, ecartsCaisses...)
	manques = append(manques, ecartsObstacles...)
	if len(manques) > 0 {
		return nil, &manifest.Invalid{Path: chemin, Missing: manques}
	}
	return &Loaded{
		Grid: grille, Tiles: tuiles, Scenario: scenario, Ambient: ambiance,
		Exit: sortie, Crates: caisses, Breakables: obstacles,
	}, nil
}

// Loaded est ce qu'un lieu devient une fois assemblé, jamais le fichier qu'il
// était.
//
// La distinction porte : `Level` est la structure décodée, avec ses pièces
// posées et ses noms de profils ; ceci est le produit de l'assemblage, où la
// géométrie est devenue une grille et les noms des index. Rien ici ne se
// réécrit dans un fichier.
//
// **N'y entre que ce que l'assemblage produit**, jamais ce qu'un appelant
// trouverait commode d'avoir sous la main : une struct de retour est une
// invitation permanente à devenir un fourre-tout, et ce qui l'en garde est ce
// critère plutôt que la vigilance.
type Loaded struct {
	// Grid est la carte assemblée, où le champ de flux tourne. Elle descend de
	// `Tiles`, et c'est ce qui interdit au dessin et au coût de se contredire.
	Grid *game.CostGrid
	// Tiles est la forme de décor de chaque case, ce que le rendu pose.
	Tiles *Tilemap
	// Scenario est la courbe de pression compilée.
	Scenario *game.Scenario
	// Ambient est le peuplement de figurants, résolu en index de profils.
	Ambient []game.AmbientPlacement
	// Exit est la porte de sortie, nulle quand le lieu n'en a pas.
	Exit *game.Exit
	// Crates sont les caisses posées, vides quand le lieu n'en a pas.
	Crates []game.CratePlacement
	// Breakables sont les obstacles fragiles posés, vides quand le lieu n'en a
	// pas.
	Breakables []game.BreakablePlacement
}

// assembler réunit les pièces posées en une seule carte de formes.
//
// Après quoi le moteur ne sait plus que le lieu était modulaire : le parcours du
// champ de flux tourne sur une grille ordinaire, dérivée de cette carte.
func assembler(lieu *Level, pieces []*Room, jeu *Set) *Tilemap {
	var largeur, hauteur int
	for i, pose := range lieu.Placements {
		largeur = max(largeur, pose.U+pieces[i].Size[0])
		hauteur = max(hauteur, pose.V+pieces[i].Size[1])
	}

	tuiles := newTilemap(largeur, hauteur)
	tuiles.sol = jeu.Ground
	for i, pose := range lieu.Placements {
		for v, ligne := range pieces[i].Rows {
			// Les runes et non les octets, comme le fait le contrôle de la
			// grille : un caractère de palette hors de l'ASCII décalerait
			// autrement toute la fin de sa ligne, sur une pièce que la
			// validation vient d'accepter.
			for u, jeton := range []rune(ligne) {
				tuiles.set(pose.U+u, pose.V+v, jeu.Palette[string(jeton)])
			}
		}
		// Les couches viennent après le terrain de la même pièce, et l'espace
		// y laisse la case telle qu'elle est : c'est ce qui permet d'en poser
		// une pour trois véhicules sans redire le revêtement partout ailleurs.
		for couche, lignes := range pieces[i].Layers {
			for v, ligne := range lignes {
				for u, jeton := range []rune(ligne) {
					if jeton == ' ' {
						continue
					}
					tuiles.poser(pose.U+u, pose.V+v, couche, jeu.Palette[string(jeton)])
				}
			}
		}
	}
	return tuiles
}
