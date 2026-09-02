// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le montage d'une partie : les manifestes lus, le lieu cuit, le monde bâti et
// le joueur posé. Ce qu'il rend suffit à ouvrir une fenêtre ou à écrire une
// image, et rien d'autre n'a à savoir dans quel ordre tout cela se charge.

// Package session monte une partie à partir d'un système de fichiers.
//
// Il existe parce que deux programmes en ont besoin — le jeu et la planche de
// relecture — et qu'un montage recopié aurait fini par diverger : la planche
// aurait alors relu une partie que le jeu ne joue pas, ce qui lui retire tout
// intérêt. L'écran de choix des lieux en sera le troisième appelant, avec un
// lieu qui ne sera plus celui de départ.
//
// Il ne connaît ni fenêtre ni rendu, et ne peut donc pas en dépendre : ce qu'il
// rend est de la simulation et une taille de tuile.
package session

import (
	"io/fs"
	"log/slog"

	"github.com/sprimault/cohue"
	"github.com/sprimault/cohue/internal/game"
	"github.com/sprimault/cohue/internal/level"
)

// HordeCapacity plafonne le bassin des ennemis.
//
// Trois cents entités vivantes, ce que la conception donne comme total en
// comptant ce qui approche hors champ. Au-delà, ce n'est plus une horde mais un
// mur uni : les profils cessent d'être distinguables, et avec eux la lisibilité
// de l'échec. Le spawner rencontrera ce plafond, et c'est mieux que de laisser
// la horde croître jusqu'à ce que l'image s'effondre.
const HordeCapacity = 300

// ShotCapacity plafonne le bassin des projectiles.
//
// Large devant ce qu'une arme de base met en vol — quelques dizaines de ticks de
// portée pour une salve toutes les vingt-quatre —, parce que les passifs
// multiplient les projectiles bien plus vite que la cadence. Un bassin plein
// perd le tir plutôt que de le différer : une file d'attente rendrait la cadence
// élastique.
const ShotCapacity = 256

// GemCapacity plafonne le bassin des gemmes au sol.
//
// Deux fois et demie le pic que la conception nomme — deux cents gemmes qui
// convergent d'un coup, le moment de plaisir maximal du genre. Ce qui bornera
// vraiment le stock est leur effacement, qui vient avec l'aimant ; d'ici là
// elles s'accumulent, et ce plafond est le filet qui empêche d'allouer.
const GemCapacity = 512

// Les deux réglages du semis provisoire.
const (
	// pasDuSemis est l'écart entre deux créatures posées, en cases.
	pasDuSemis = 3
	// ecartAuJoueur tient la horde assez loin pour qu'on la voie converger.
	ecartAuJoueur = 8
	// porteeDuSemis la tient assez près pour qu'elle arrive.
	//
	// **Cette borne était tenue par le lieu, et par personne quand il a grandi.**
	// Sur trente-deux cases de côté, le semis couvrait la carte entière et le
	// bord faisait la limite ; sur quatre-vingt-dix-huit, le parcours remplit le
	// bassin avant d'avoir quitté la bande nord, et la horde naît toute d'un
	// côté. Une propriété qu'on lit dans un commentaire sans qu'aucune ligne ne
	// la tienne se perd au premier changement d'échelle.
	porteeDuSemis = 24
)

// Session est une partie montée, prête à tourner.
//
// La taille de tuile voyage avec le monde parce qu'elle vient du même
// chargement : elle est dans le manifeste de décor, que le rendu ne lit pas
// lui-même — il ne connaît que des profils et des cycles, et une taille reçue.
type Session struct {
	World *game.World
	Grid  *game.CostGrid
	Tile  [2]int

	// Seed est la graine de la run en cours, et le seul état de jeu qui traverse
	// une relance — sous une forme changée, puisque chaque run dérive la
	// suivante. L'écran de mort la montrera : sans elle affichée, un joueur qui
	// veut rejouer sa run n'a rien à noter.
	Seed uint64

	// Ce que la relance conserve, parce que rien ne l'a modifié : les tables du
	// manifeste et le lieu cuit. Les relire coûterait un décodage complet pour
	// rendre exactement les mêmes valeurs.
	profils     *game.Profiles
	armes       *game.Weapons
	progression *game.Progression
}

// Restart rejoue le même lieu, sans rien redemander.
//
// **La règle de ce remontage n'est pas qu'il ait lieu, mais ce qu'il conserve et
// ce qu'il remet à zéro.** Il ne conserve rien de la partie précédente : ni vie,
// ni horde, ni tick, ni compteur. Ce qui survit est ce que la partie n'a pas
// touché — les tables et la carte —, et cette liste est vide de tout état de jeu
// par construction plutôt que par vigilance.
//
// C'est pour cela que le remontage vit ici et non dans le rendu : ce qu'il
// conserve se mesure, et un remontage piloté par l'écran serait juste et
// invérifiable.
//
// **La graine fait exception, et c'est la seule** : elle ne traverse pas, elle
// engendre. La run suivante en reçoit une neuve, dérivée de celle qui s'achève,
// de sorte que deux morts ne rejouent pas la même chose et que toute la suite
// des runs d'une session descende de sa graine de départ.
func (s *Session) Restart() {
	s.Seed = game.NextSeed(s.Seed)
	s.monter()
}

// monter bâtit la run de la graine courante : le monde, le joueur, la horde.
//
// Séparé de Restart parce que la première run d'une session ne dérive rien —
// elle se joue sur la graine reçue. Les confondre aurait fait qu'ouvrir une
// session sur une graine en jouerait une autre, ce qui se serait vu au moment
// d'écrire un lieu de défi.
func (s *Session) monter() {
	s.World = game.NewWorld(s.profils, s.armes, s.progression, s.Grid, s.Seed,
		HordeCapacity, ShotCapacity, GemCapacity)
	placer(s.World, s.Grid)
	peupler(s.World, s.Grid, s.profils)
}

// Open monte une partie sur la campagne donnée, à son lieu de départ.
//
// **Une campagne et non un lieu**, parce que c'est elle que l'auteur compose et
// partage, et parce que le lieu de départ est une propriété de la campagne : le
// binaire n'a pas à savoir laquelle de ses salles vient en premier. Quand
// l'étape 8 apportera les portes, c'est ce même descripteur qui dira où mène
// chacune, et le montage n'aura pas à changer de forme.
//
// L'ordre n'est pas libre : le catalogue de coûts vient du manifeste de décor et
// le chargeur de lieux en a besoin, si bien qu'un lieu ne peut pas se cuire avant
// que le décor soit lu.
//
// Les manifestes sont lus au montage et non à la première vague : un fichier que
// le binaire refuse doit le dire tout de suite, pas trois minutes après le début
// d'une partie.
//
// Les capacités des bassins ne sont pas des paramètres : ce sont des plafonds de
// la partie, pas des réglages de l'appelant. Les laisser choisir aurait fait
// qu'une planche montre une horde plus petite que le jeu, donc qu'elle relise
// autre chose que ce qui se joue.
//
// La graine, elle, est un paramètre, et elle n'a pas de valeur par défaut : le
// montage ne la tire ni de l'horloge, que l'invariant du déterminisme proscrit,
// ni d'une constante qui ferait de deux appelants aux intentions différentes
// deux copies de la même valeur. La planche de relecture en exige une fixe par
// nature ; le jeu n'en a une fixe que faute d'écran pour la choisir.
func Open(fsys fs.FS, campagne string, graine uint64) (*Session, error) {
	decor, couts, err := level.LoadDecor(fsys, cohue.DecorManifest)
	if err != nil {
		return nil, err
	}

	graphe, err := level.LoadCampaign(fsys, campagne)
	if err != nil {
		return nil, err
	}
	lieu := graphe.StartPath(campagne)

	grille, err := level.NewLoader(fsys, couts).Load(lieu)
	if err != nil {
		return nil, err
	}
	slog.Info("lieu chargé", "campaign", graphe.ID, "name", lieu,
		"width", grille.Width(), "height", grille.Height())

	profils, err := game.LoadProfiles(fsys, cohue.CharacterManifest)
	if err != nil {
		return nil, err
	}
	slog.Info("profils chargés", "enemies", len(profils.Enemies))

	armes, err := game.LoadWeapons(fsys, cohue.WeaponManifest)
	if err != nil {
		return nil, err
	}
	slog.Info("armes chargées", "base", armes.Base.Key)

	progression, err := game.LoadProgression(fsys, cohue.ProgressionManifest)
	if err != nil {
		return nil, err
	}

	partie := &Session{
		Grid:        grille,
		Tile:        decor.Tile,
		Seed:        graine,
		profils:     profils,
		armes:       armes,
		progression: progression,
	}
	partie.monter()
	slog.Info("partie montée", "seed", graine)
	return partie, nil
}

// placer pose le joueur au centre du lieu.
//
// Au centre, faute de mieux : le format ne porte pas encore d'ancrage de départ,
// et l'inventer maintenant demanderait de trancher un champ sans usage réel. La
// position de départ dépendra du lieu, de la campagne et de la porte par laquelle
// on entre — trois choses que l'étape 8 apporte.
//
// Au centre de la case et non sur son coin : c'est là que se tient une entité,
// et poser le joueur sur un coin montrerait un cas qu'aucune partie ne produit.
func placer(monde *game.World, grille *game.CostGrid) {
	monde.Place(
		game.FromInt(grille.Width()/2)+game.One/2,
		game.FromInt(grille.Height()/2)+game.One/2,
	)
}

// peupler sème la horde de départ, à intervalle régulier et à distance du joueur.
//
// **Provisoire, et ce n'est pas une approximation du spawner.** L'étape 4
// achètera des créatures dans un budget de pression et les fera apparaître hors
// du champ, sur un anneau autour du joueur. Ce qui est ici n'a qu'un objet : que
// l'étape 1 se voie. Trois cents poursuivants qui convergent en contournant les
// obstacles est ce qu'elle a livré, et un rendu qui ne le montrerait pas ne
// livrerait pas ce que l'étape 2 promet.
//
// **Le semis est régulier et non tiré au sort.** La raison n'est pas le confort
// de comparer deux planches — cela en découle, mais ne le motive pas : `World` ne
// porte pas de graine, et lui en donner une depuis un montage ferait entrer une
// décision de simulation par la porte de service. L'aléatoire d'une partie
// appartient aux flux nommés, que le spawner branchera.
//
// L'écart au joueur se mesure en cases et non en tuiles du monde : ce qui est
// semé l'est sur la grille, et un rayon exact n'apporterait rien à un motif dont
// le pas vaut trois cases.
//
// **Ce qu'il coûte, mesuré :** cent vingt et une créatures posées d'un coup dans
// la couronne convergent en cinq secondes, et le joueur tombe à la sixième avec
// une gemme des dix que le premier niveau demande. Aucune position de départ ni
// direction de fuite n'ouvre une montée de niveau vivant. Ce n'est pas un défaut
// de la progression mais du semis : la courbe de pression achète les créatures
// dans un budget qui commence bas, et c'est elle qui rendra le jalon mesurable.
func peupler(monde *game.World, grille *game.CostGrid, profils *game.Profiles) {
	px, py := monde.Player()
	pu, pv := px.Floor(), py.Floor()

	profil := 0
	for v := 0; v < grille.Height(); v += pasDuSemis {
		for u := 0; u < grille.Width(); u += pasDuSemis {
			if d := ecart(u, v, pu, pv); d < ecartAuJoueur || d > porteeDuSemis {
				continue
			}
			if !grille.Passable(u, v) {
				continue
			}
			_, pose := monde.SpawnEnemy(
				profil%len(profils.Enemies),
				game.FromInt(u)+game.One/2,
				game.FromInt(v)+game.One/2,
			)
			if !pose {
				return
			}
			profil++
		}
	}
}

// ecart rend la distance entre deux cases, comptée en pas de grille.
func ecart(u, v, pu, pv int) int {
	return abs(u-pu) + abs(v-pv)
}

// abs rend la valeur absolue d'un entier.
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
