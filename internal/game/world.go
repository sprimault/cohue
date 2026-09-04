// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La partie en cours — bassins, champ de flux, densité, profils, arme, flux
// aléatoires — et l'ordre d'un tick. Rien n'y est alloué après le montage : la
// réutilisation des tableaux d'un tick à l'autre est ce qui tient le budget
// d'allocation.

package game

// flowPeriod est le nombre de ticks entre deux calculs du champ de flux.
//
// Le champ ne dépend que du joueur et des obstacles, et six ticks font un
// dixième de seconde : à cinq tuiles par seconde, le joueur y parcourt une
// demi-tuile. Le recalculer à chaque image paierait plein tarif pour une
// information qui n'a pas bougé d'une case. C'est un
// compromis de coût, pas un réglage de jeu — le baisser ne rend pas la horde
// plus maligne, il la rend plus chère.
const flowPeriod Tick = 6

// separationScale convertit la pente de densité en poussée.
//
// La grille rend une pente en entités par cellule ; l'attirance du champ vaut
// une tuile. Sans ce facteur, deux voisines suffiraient à retourner une créature
// contre le joueur. Le poids de séparation du profil multiplie ce vecteur et
// reste un rapport — c'est ici qu'est l'échelle, et c'est un des chiffres qu'on
// cherchera en équilibrant, avec la vitesse relative et le plafond de dégâts. Le
// jalon de l'étape 3 ne les a pas tranchés : il jugeait une sensation, contre un
// seul profil.
const separationScale = One / 8

// eclairImpact est la durée de l'éclair qu'une créature touchée porte, en ticks.
//
// **Une créature encaisse trois touches et le joueur ne voyait que la troisième.**
// Rien ne distinguait un tir qui rate d'un tir qui entame, si bien qu'on
// déduisait les ratés au lieu de les lire — c'est le premier reproche qu'une
// partie jouée a fait au tir.
//
// Quatre ticks, un quinzième de seconde : assez pour se voir sur une cadence de
// vingt-quatre, trop court pour qu'une foule dense clignote en permanence. Le
// chiffre vit ici parce que le décompte y vit ; ce que le rendu en fait lui
// appartient.
const eclairImpact Tick = 4

// eclairSoin est la durée des deux éclairs de soin, en ticks.
//
// Plus long que celui de l'impact, et pour une raison de lecture opposée : un
// impact se répète à chaque salve, donc il doit s'éteindre vite ou la foule
// clignote ; un soin arrive une fois par seconde et demie, et sa rareté est ce
// qui doit se remarquer. Un quart de seconde laisse le temps de trouver d'où il
// vient au milieu d'une horde.
const eclairSoin Tick = 15

// rabattement est la distance sous laquelle une créature vise le joueur plutôt
// que la cellule suivante du champ de flux, en tuiles.
//
// Une tuile et demie : de quoi couvrir la cellule de la cible et le bord des
// huit voisines, là où la direction tabulée cesse de rapprocher. Plus large, on
// court-circuiterait le contournement d'obstacles qui est la raison d'être du
// champ ; plus étroit, on retrouverait la horde arrêtée au bord de la case.
const rabattement = One * 3 / 2

// World est une partie en cours : tout ce qu'un tick lit et modifie.
//
// La struct porte les tableaux plutôt que de les rendre à l'appelant, parce que
// leur réutilisation d'un tick à l'autre est ce qui tient le budget
// d'allocation. Rien n'y est alloué après la construction.
type World struct {
	profils *Profiles
	// arme est la copie que la partie transforme. Les passifs la modifient, si
	// bien qu'une relance repart de la table sans qu'on ait à défaire quoi que
	// ce soit.
	arme        Weapon
	passifs     *Passives
	progression *Progression
	grille      *CostGrid
	flux        *FlowField
	densite     *DensityGrid
	ennemis     *Pool[Enemy]
	tirs        *Pool[Projectile]
	tirsEnnemis *Pool[Projectile]
	souffles    *Pool[Blast]
	// ambiants sont les figurants du lieu. Un bassin à part parce qu'ils ne sont
	// ni comptés, ni visés, ni hostiles — ce qui les range hors de tout ce que le
	// bassin des ennemis suppose.
	ambiants *Pool[Ambient]
	gemmes   *Pool[Gem]
	// aimants tient au plus un objet, la règle du lot étant qu'un seul soit au
	// sol à la fois. Un bassin quand même : le rendu parcourt les bassins, et une
	// entité rangée autrement y serait un cas particulier.
	aimants *Pool[Magnet]
	// hasard porte les quatre flux de la partie. Ils vivent ici plutôt que dans
	// le montage parce que c'est le tick qui les consomme, et qu'un hasard tenu à
	// côté de la partie serait un état de simulation hors de la simulation.
	hasard *Streams
	// scenario est la courbe de pression du lieu. Elle vient de son fichier et
	// non des tables de la partie : c'est le rythme que son auteur compose.
	scenario *Scenario

	playerX, playerY Fixed
	// vie est ce qu'il reste au joueur, en points. À zéro, il est mort — la
	// valeur est l'état, comme la résistance d'une créature.
	vie int
	// degatsSubis est le reste de l'accumulateur de contact, en points-ticks.
	// Il porte ce qui n'a pas encore fait un point entier ; sa raison d'être est
	// dans `subir`.
	degatsSubis int
	// cooldown est ce qui reste à attendre avant le prochain tir. Il descend à
	// zéro et y demeure : sans cible, l'arme ne tire pas et ne consomme rien.
	cooldown Tick
	tick     Tick

	// niveau est celui du joueur, le premier valant un.
	niveau int
	// experience est ce qui a été ramassé vers le niveau suivant, en gemmes.
	// Ce que le seuil dépasse est reporté et non perdu.
	experience int
	// depuisChoix compte les ticks écoulés depuis la dernière montée, quelle
	// qu'en soit la source. C'est lui que le plancher de temps surveille, et son
	// nom dit qu'il ne mesure pas l'âge de la run.
	depuisChoix Tick
	// cartes sont les places offertes, vides quand aucun choix n'est ouvert.
	// La tranche est réutilisée d'un choix à l'autre.
	cartes []Card
	// paliers compte ce qui a été pris sur chaque axe, indexé comme
	// `Passives.Axes`. Un compteur par axe plutôt qu'une liste de cartes prises :
	// c'est le rang qui décide de ce que la carte suivante offre.
	paliers []int
	// enAttente est le nombre de choix dus au joueur, en plus de celui qui est
	// ouvert. Une récolte abondante en donne deux d'un coup, et les présenter
	// l'un après l'autre est la seule façon de n'en perdre aucun.
	enAttente int
	// charge dit que le joueur tient un aimant. Un booléen et non un compteur :
	// la conception parle d'une charge, et pouvoir en empiler retirerait la
	// décision de dépenser.
	charge bool
	// dernierAimant est le tick de la dernière apparition, ou de la dernière
	// tentative abandonnée.
	dernierAimant Tick
	// budget est la pression accumulée et non encore dépensée. En virgule fixe :
	// une phase à huit par seconde accorde moins d'un point par tick, et un
	// compteur entier ne saurait pas l'exprimer autrement qu'en tronquant à zéro.
	budget Fixed
	// vivants compte la horde par profil, refait à chaque tick d'achat. Il sert
	// au plafond de simultanéité, qui compte les vivants et non les apparus.
	vivants []int
	// achetables est la tranche de travail du spawner, réutilisée d'un achat à
	// l'autre.
	achetables []int
	// sortie est la porte du lieu, nulle quand il n'en a pas. Elle vient de son
	// fichier, comme le scénario et les figurants.
	sortie *Exit
	// abattus compte les créatures tombées depuis le début de la run. C'est ce
	// que l'objectif de porte mesure, et les figurants n'y entrent pas — ils ne
	// sont pas dans ce bassin.
	abattus int
	// echappe dit que le joueur est sorti par la porte. Un drapeau, là où la
	// mort n'en a pas besoin : la vie à zéro *est* la mort, quand rien dans la
	// position du joueur ne distingue « contre la porte » de « sorti ».
	echappe bool
	// convoite est le profil pour lequel le spawner épargne, **décalé de un** :
	// zéro veut dire « aucun ».
	//
	// Le décalage n'est pas une coquetterie. L'index brut ferait du zéro à la
	// fois « aucun convoité » et « le premier profil de la table », c'est-à-dire
	// qu'un champ oublié désignerait silencieusement le Badaud. Décalé, le zéro
	// est l'état que rien de valide ne produit — le même geste que les
	// générations d'un `Handle`, qui partent à un.
	convoite int
}

// Capacities dit combien de places chaque bassin préalloue.
//
// **Une struct plutôt que des entiers de rang.** Quatre `int` consécutifs à
// l'appel sont une inversion silencieuse : les valeurs sont toutes plausibles,
// le compilateur les accepte dans n'importe quel ordre, et un bassin de gemmes
// réduit à seize places ne se manifesterait que par des gemmes qui n'apparaissent
// pas, longtemps après. C'est le geste que le projet fait partout où le
// compilateur ne peut pas aider — `Fixed` plutôt qu'un entier nu, `Spawn` par
// valeur —, appliqué à un appel plutôt qu'à une valeur.
//
// **Elle ne se valide pas, et le motif tient à d'où viennent ses valeurs.**
// Elles sont des constantes de `internal/session` et ne sortent d'aucun fichier :
// aucune donnée tierce ne peut les rendre nulles, si bien qu'un zéro y serait
// une faute de programmation et non une entrée invalide. Le premier test qui
// fait apparaître une créature la trouverait — un bassin sans place ne pose
// rien.
//
// C'est ce qui sépare ce cas des durées nulles que le chargement refuse : celles-
// là viennent d'un manifeste, donc d'une main qu'on ne contrôle pas. Ce que cette
// struct ferme est l'inversion, pas l'oubli.
type Capacities struct {
	// Enemies est la horde vivante à un instant.
	Enemies int
	// Shots est le nombre de projectiles du joueur en vol.
	Shots int
	// EnemyShots est celui des projectiles de la horde, qui a son bassin parce
	// que ce qu'ils retirent n'a pas la même unité.
	EnemyShots int
	// Blasts est le nombre d'explosions amorcées à la fois.
	Blasts int
	// Ambients est le nombre de figurants qu'un lieu peut porter.
	Ambients int
	// Gems est le nombre de gemmes au sol.
	Gems int
}

// NewWorld monte une partie sur une carte et les tables du manifeste.
//
// La table d'armes entre entière plutôt que sa seule arme de base : les passifs
// y vivent, et le monde en a besoin dès la première montée de niveau. Les passer
// à côté aurait fait deux paramètres pour ce qui est un seul fichier.
//
// Les capacités ne changent plus après le montage. Les plafonds eux-mêmes et ce
// qui les justifie vivent dans `internal/session`, qui monte les parties : ce
// sont des valeurs de jeu, pas un paramètre que chaque appelant choisirait.
//
// La graine en est un, en revanche, et elle vient du montage : lui seul sait de
// quelle run il s'agit dans la suite d'une session. Une partie qui tirerait la
// sienne ne se rejouerait plus.
func NewWorld(profils *Profiles, armes *Weapons, progression *Progression, scenario *Scenario,
	grille *CostGrid, graine uint64, capacites Capacities) *World {
	return &World{
		profils:     profils,
		arme:        armes.Base,
		passifs:     armes.Passives,
		progression: progression,
		scenario:    scenario,
		vie:         profils.Player.Health,
		niveau:      1,
		grille:      grille,
		flux:        NewFlowField(grille),
		densite:     NewDensityGrid(grille.Width(), grille.Height()),
		ennemis:     NewPool[Enemy](capacites.Enemies),
		tirs:        NewPool[Projectile](capacites.Shots),
		tirsEnnemis: NewPool[Projectile](capacites.EnemyShots),
		souffles:    NewPool[Blast](capacites.Blasts),
		ambiants:    NewPool[Ambient](capacites.Ambients),
		gemmes:      NewPool[Gem](capacites.Gems),
		aimants:     NewPool[Magnet](1),
		hasard:      NewStreams(graine),
		cartes:      make([]Card, 0, Choices),
		paliers:     make([]int, len(armes.Passives.Axes)),
		vivants:     make([]int, len(profils.Enemies)),
		achetables:  make([]int, 0, len(profils.Enemies)),
	}
}

// Enemies rend le bassin des ennemis, pour le parcourir ou le peupler.
func (w *World) Enemies() *Pool[Enemy] { return w.ennemis }

// Shots rend le bassin des projectiles du joueur en vol.
func (w *World) Shots() *Pool[Projectile] { return w.tirs }

// EnemyShots rend le bassin des projectiles tirés par la horde.
//
// **Deux bassins pour un même type, et ce n'est pas une commodité.** Ce que
// `Projectile.Hits` retire s'exprime dans l'unité de ce qu'il touche : des
// touches pour une créature, des points de vie pour le joueur. Un bassin unique
// donnerait donc au champ deux sens qu'aucun type ne sépare — même déclaration,
// valeur également plausible, et seule la lecture du code dirait laquelle
// s'applique. C'est le bassin qui porte l'unité, comme le cadavre est une nature
// et non un ennemi à drapeau.
func (w *World) EnemyShots() *Pool[Projectile] { return w.tirsEnnemis }

// Blasts rend le bassin des explosions amorcées.
func (w *World) Blasts() *Pool[Blast] { return w.souffles }

// Streams rend les flux aléatoires de la partie.
//
// Ils sortent du monde parce que la relance se juge sur eux : deux runs d'une
// même session doivent tirer autrement, et deux sessions ouvertes sur la même
// graine doivent tirer pareil. Sans cet accès, ces deux propriétés
// n'existeraient que dans un commentaire.
func (w *World) Streams() *Streams { return w.hasard }

// Player rend la position du joueur.
func (w *World) Player() (Fixed, Fixed) { return w.playerX, w.playerY }

// Place pose le joueur, au montage du lieu.
//
// Sans projection ni contrôle de passabilité : c'est au chargeur de savoir où
// commence un lieu, et un point de départ dans un mur est un défaut du niveau,
// pas un cas que la simulation rattrape.
func (w *World) Place(x, y Fixed) {
	w.playerX, w.playerY = x, y
	w.flux.Rebuild(x, y)
}

// SpawnEnemy pose une créature d'un profil donné.
//
// L'index est celui de la table, jamais une copie de ses valeurs. Le second
// résultat est faux quand le bassin est plein.
func (w *World) SpawnEnemy(profil int, x, y Fixed) (Handle, bool) {
	return w.ennemis.Spawn(Enemy{
		Profile: profil,
		X:       x,
		Y:       y,
		Hits:    w.profils.Enemies[profil].HitsAt(w.durcissement()),
	})
}

// Step avance la simulation d'un tick, le joueur suivant la direction voulue.
//
// L'ordre est celui du chapitre 15 de la conception, qui l'énumère et porte les
// arbitrages : les entrées, les apparitions, le champ de flux si c'est son tick,
// la densité, les intentions et leur projection, les dégâts de contact, l'aimant
// et sa ruée, puis le ramassage et ce qu'il fait monter, le tir puis le vol des
// projectiles avec ce qu'ils touchent, et les suppressions en dernier. La place
// des apparitions se justifie dans `apparaitre`, qui la tient.
//
// **La simulation continue de tourner après la mort**, et `subir` cesse
// seulement d'appliquer des dégâts. Figer le monde ici serait une décision
// d'écran — ce que la mort suspend, ce qu'elle laisse courir — et elle
// appartient à qui l'affichera, pas à la boucle.
//
// Les intentions et la projection tiennent ici en une seule passe alors que la
// conception les énumère séparément, ce qu'elle autorise sous la condition
// qu'elle pose : aucune intention ne lit l'état d'une autre entité.
// `deplacerEnnemis` ne lit que le champ et la densité, tous deux figés avant la
// passe, si bien que rien de ce qu'une créature déplacée modifie n'est lu par les
// suivantes.
func (w *World) Step(voulu Vec) {
	w.deplacerJoueur(voulu)
	w.apparaitre()

	if w.tick%flowPeriod == 0 {
		w.flux.Rebuild(w.playerX, w.playerY)
	}
	w.compterDensite()
	w.deplacerEnnemis()
	w.subir()
	// Les mèches brûlent avec les dégâts de contact, dont l'explosion est une
	// source au même titre. Elle vient après le déplacement du joueur, ce qui lui
	// laisse le dernier tick pour sortir de l'emprise.
	w.detoner()
	w.poserAimant()
	w.attirer()
	w.prendreAimant()
	w.progresser(w.ramasser())
	w.tirer()
	w.deplacerTirs()
	w.tirerLaHorde()
	w.deplacerTirsEnnemis()
	w.soigner()
	// L'ambiance en dernier, et hors de tout ce qui précède : elle ne lit rien de
	// la partie et rien de la partie ne la lit. Sa place dans le tick est donc
	// indifférente, ce qui est la meilleure preuve qu'elle ne décide de rien.
	w.errer()
	w.retirerLesMorts()
	w.franchir()

	w.tick++
}

// Tick rend le numéro du tick courant.
func (w *World) Tick() Tick { return w.tick }

// deplacerJoueur applique la direction voulue, à la vitesse de son profil.
func (w *World) deplacerJoueur(voulu Vec) {
	if voulu == (Vec{}) {
		return
	}
	pas := voulu.Direction(0).Scale(w.vitesse(w.profils.Player.Speed, w.playerX, w.playerY))
	w.playerX, w.playerY = w.glisserJoueur(w.playerX, w.playerY, pas)
}

// compterDensite refait le comptage des ennemis, cellule par cellule.
func (w *World) compterDensite() {
	w.densite.Clear()
	for i := range w.ennemis.Active() {
		e := w.ennemis.At(i)
		u, v := w.flux.Cell(e.X, e.Y)
		w.densite.Add(u, v)
	}
}

// deplacerEnnemis calcule l'intention de chaque créature et la projette.
//
// L'index de l'entité sert deux fois : il désigne son profil dans la table, et
// il départage les directions quand la somme des forces s'annule.
func (w *World) deplacerEnnemis() {
	for i := range w.ennemis.Active() {
		e := w.ennemis.At(i)
		profil := &w.profils.Enemies[e.Profile]

		// L'éclair s'éteint dans la passe qui parcourt déjà la horde : lui en
		// donner une à lui seul ferait un tour de bassin par tick pour un
		// décompte.
		if e.Flash > 0 {
			e.Flash--
		}

		// La charge décide avant le champ de flux, et quand elle décide elle
		// décide seule : une créature en course ne lit ni le champ ni la densité,
		// c'est ce qui la fait aller droit. Les autres passent ici sans effet,
		// leur portée de charge étant nulle.
		if impose, charge := w.charger(e, profil); charge {
			w.avancer(e, impose)
			w.arreterAuMur(e, profil, impose)
			continue
		}

		// **Une créature qui tire se stabilise dès qu'elle est à portée.** C'est
		// ce qui la rend identifiable immobile au milieu d'une horde qui
		// converge, et ce qui l'empêche de finir au contact — venir au corps à
		// corps annulerait le seul rôle d'un profil qui blesse de loin.
		if _, portee := w.viseeDe(e, profil); portee {
			w.avancer(e, Vec{})
			continue
		}

		u, v := w.flux.Cell(e.X, e.Y)
		attirance := w.flux.Direction(u, v)

		// **Près du joueur, viser sa position et non la cellule suivante.** Le
		// champ mène de case en case, et sa direction est nulle dans celle de la
		// cible : une créature qui y entre n'a plus rien qui l'attire, tandis que
		// la densité — forte là où tout le monde converge — continue de la
		// pousser dehors. La horde encerclait donc à un demi-tuile sans jamais
		// toucher, et le contact n'arrivait que par accident.
		//
		// **Le rabattement se déclenche sur une distance et non sur la case.**
		// Sur la case, une créature restée dans la voisine garderait la direction
		// tabulée jusqu'au bord, ce qui est exactement la position mesurée.
		//
		// Le seuil vaut une tuile et demie : au-delà, le champ fait mieux
		// puisqu'il contourne les obstacles ; en deçà, il ne dit plus rien
		// d'utile, et le glissement empêche de toute façon de traverser un mur.
		ecart := Vec{X: w.playerX - e.X, Y: w.playerY - e.Y}
		if ecart.carres() < int64(rabattement)*int64(rabattement) {
			attirance = ecart.Direction(i)
		}
		repulsion := w.densite.Gradient(u, v).Scale(profil.SeparationWeight).Scale(separationScale)

		voulu := attirance.Sub(repulsion)
		if profil.Tangential != 0 {
			// La dérive latérale d'un flanqueur se prend sur l'attirance et non
			// sur la somme : sur la somme, une créature serrée par ses voisines
			// tournerait autour d'elles au lieu de tourner autour du joueur.
			voulu = voulu.Add(attirance.Perp().Scale(profil.Tangential))
		}

		w.avancer(e, voulu.Direction(i).Scale(w.vitesse(profil.Speed, e.X, e.Y)))
	}
}

// avancer applique un pas à une créature et retient celui qui a eu lieu.
//
// Le pas retenu est celui qui a eu lieu, projection sur la passabilité comprise :
// c'est la visée qui le lit, et prédire un déplacement que le mur a annulé ferait
// tirer dans le mur. C'est aussi ce que l'arrêt de la charge compare à son pas
// voulu, d'où un seul endroit qui l'écrive — deux copies finiraient par ne plus
// dire la même chose du même tick.
//
// **Une créature en traverse une autre, y compris un Molosse lancé sur un
// Vigile.** Ce qu'une créature solide refuse est le joueur, jamais ses
// congénères : une charge ne s'interrompt donc que sur le décor, et un bloqueur
// qui servirait de bouclier à la horde contre les charges n'aurait de sens pour
// personne.
func (w *World) avancer(e *Enemy, pas Vec) {
	avantX, avantY := e.X, e.Y
	if profil := &w.profils.Enemies[e.Profile]; profil.Solid {
		e.X, e.Y = w.glisserSolide(e.X, e.Y, pas, profil.Radius)
	} else {
		e.X, e.Y = w.glisser(e.X, e.Y, pas)
	}
	e.Step = Vec{X: e.X - avantX, Y: e.Y - avantY}
}

// vitesse rend la vitesse effective sur la case du point d'appui.
//
// Le coût de la case divise la vitesse, sans quoi le parcours pondéré serait une
// superstition : le champ contournerait au prix de deux cases ce qui ne coûte
// rien à traverser, et l'écart entre le chemin choisi et le chemin payé ne se
// verrait nulle part.
//
// La case est celle d'avant le pas. Diviser par celle d'arrivée serait
// circulaire — le pas plein entre dans la flaque, la division le raccourcit, il
// n'y entre plus, et l'entité tremble à la frontière.
func (w *World) vitesse(base Fixed, x, y Fixed) Fixed {
	cout := w.grille.At(x.Floor(), y.Floor())
	if cout <= Free {
		return base
	}
	return base.Div(FromInt(int(cout)))
}

// glisser applique un pas en annulant ce qui entrerait dans un obstacle.
//
// Les deux axes se traitent séparément : un déplacement qui mène dans un mur
// perd sa composante bloquée et l'entité longe la paroi au lieu de s'y enfoncer.
// Les deux bloqués, elle ne bouge pas.
//
// **Elle annule une composante, elle n'en raccourcit jamais aucune, et rien
// n'arrondit** : sans obstacle, le pas rendu est exactement celui qu'on lui
// passe. `arreterAuMur` en dépend — elle reconnaît un mur à l'inégalité stricte
// entre le pas voulu et le pas obtenu, ce qu'un raccourcissement partiel, ou un
// arrondi de la position, rendrait faux en silence.
//
// La poussée de séparation ne décide donc pas de la position finale, et c'est ce
// qui permet à vingt Badauds de pousser un Vigile contre une cloison sans le
// faire passer au travers : ils s'entassent derrière lui, ce qui est tout
// l'intérêt d'un couloir bouché.
func (w *World) glisser(x, y Fixed, pas Vec) (Fixed, Fixed) {
	return w.projeter(x, y, pas, exclusionAucune, 0)
}

// glisserJoueur applique un pas au joueur, que les corps solides arrêtent en
// plus du décor.
//
// **Deux fonctions nommées plutôt qu'un booléen à l'appel.** `glisser(x, y, pas,
// true)` ne dirait pas *true quoi*, et c'est le défaut qu'on vient de fermer sur
// les capacités de bassin ; un prédicat passé en argument le dirait, mais un
// method value alloue et cette ligne est dans la boucle de mise à jour.
func (w *World) glisserJoueur(x, y Fixed, pas Vec) (Fixed, Fixed) {
	return w.projeter(x, y, pas, exclusionCorps, 0)
}

// glisserSolide applique un pas à une créature dont le corps bloque, que le
// joueur arrête en retour.
//
// C'est l'autre moitié de l'exception, et elle est ce qui la rend vraie : sans
// elle, le Vigile entre dans le joueur et le blocage cesse au moment précis où
// il devrait jouer.
func (w *World) glisserSolide(x, y Fixed, pas Vec, rayon Fixed) (Fixed, Fixed) {
	return w.projeter(x, y, pas, exclusionJoueur, rayon)
}

// exclusion dit quels corps, en plus du décor, arrêtent le mobile qu'on projette.
//
// **La solidité du Vigile joue dans les deux sens, et c'est ce qui la rend
// effective.** Le premier jet ne l'arrêtait que du côté du joueur : le Vigile
// poursuivant, il finissait toujours par le recouvrir, et la règle qui rend ses
// mouvements libres une fois dedans annulait le blocage à chaque rencontre. Un
// corps qui bouche un couloir doit donc refuser d'entrer dans le joueur autant
// que le joueur refuse d'entrer en lui — sans quoi il n'a jamais bouché quoi que
// ce soit.
//
// Les créatures continuent de se traverser entre elles : ce qui est exclu ici
// est le couple joueur / corps solide, jamais deux corps de la horde.
type exclusion uint8

const (
	// exclusionAucune est le cas de la horde ordinaire, que seul le décor arrête.
	exclusionAucune exclusion = iota
	// exclusionCorps est celui du joueur, qu'un corps solide arrête en plus.
	exclusionCorps
	// exclusionJoueur est celui d'une créature solide, que le joueur arrête.
	exclusionJoueur
)

// projeter est le corps commun des trois, `exclut` disant ce qui bloque en plus
// du décor.
//
// `rayon` est celui du mobile projeté, et il ne sert qu'à l'exclusion par le
// joueur : les deux autres cas mesurent depuis le joueur, dont le rayon est déjà
// connu ici. Le porter en paramètre plutôt que de le retrouver évite à cette
// passe de traverser la table des profils à chaque axe.
//
// **On n'empêche pas d'être dans un recouvrement, on empêche d'y entrer.**
// Depuis l'intérieur, toutes les directions redeviennent libres ; venant de
// dehors, chaque axe qui entrerait est annulé comme il le serait par un mur.
//
// **Ce n'est pas la parade à un cas du jeu : la réciprocité l'a rendu
// inatteignable.** Le recouvrement n'arrive que si un corps solide et le joueur
// se trouvent superposés, ce qu'aucun chemin ne produit plus une fois qu'aucun
// des deux ne peut entrer dans l'autre.
//
// **Ce qui justifie de le garder est le mode d'échec, pas la totalité par
// principe.** Sans lui, une position que rien ne devrait produire ne plante pas
// et ne signale rien : elle fige le joueur jusqu'à la fin de la partie, sans
// qu'un pixel dise pourquoi. C'est le silence, et le projet le ferme partout où
// il peut. Si l'absence de garde produisait une panique, il faudrait le retirer —
// un état impossible qui panique est une information.
//
// **Il n'est éprouvé que par un test qui forge son entrée**, jamais par une
// partie. Il doit donc rester de cette taille : le jour où on voudra l'étendre —
// deux corps, une sortie préférentielle —, c'est qu'il faut le retirer plutôt
// que le faire grandir. On ne développe pas du code que rien n'exerce.
//
// Le garde a d'ailleurs précédé la réciprocité, et l'ordre vaut d'être connu :
// il avait été écrit pour libérer un joueur recouvert par un Vigile qui avance,
// et c'est en le voyant annuler le blocage à chaque rencontre qu'on a compris
// qu'un corps qui se rend traversable en avançant n'est plus un corps.
func (w *World) projeter(x, y Fixed, pas Vec, exclut exclusion, rayon Fixed) (Fixed, Fixed) {
	dedans := w.recouvert(x, y, exclut, rayon)
	nx, ny := x+pas.X, y+pas.Y
	if !w.libre(nx, y, exclut, rayon, dedans) {
		nx = x
	}
	if !w.libre(nx, ny, exclut, rayon, dedans) {
		ny = y
	}
	return nx, ny
}

// libre dit si une position accueille le mobile qu'on y projette.
//
// `dedans` porte l'état du **départ** et non celui de l'arrivée : c'est lui qui
// rend la règle asymétrique, et le passer en paramètre plutôt que de le
// recalculer ici évite un parcours de horde par axe.
func (w *World) libre(x, y Fixed, exclut exclusion, rayon Fixed, dedans bool) bool {
	if !w.passable(x, y) {
		return false
	}
	return dedans || !w.recouvert(x, y, exclut, rayon)
}

// recouvert dit si une position tombe dans ce que l'exclusion refuse.
//
// Le cas sans exclusion sort avant tout parcours : la horde ordinaire traverse
// la horde, et c'est elle qui appelle cette passe trois cents fois par tick.
func (w *World) recouvert(x, y Fixed, exclut exclusion, rayon Fixed) bool {
	switch exclut {
	case exclusionCorps:
		return w.dansUnCorps(x, y)
	case exclusionJoueur:
		return w.surLeJoueur(x, y, rayon)
	default:
		return false
	}
}

// surLeJoueur dit si un corps posé là recouvrirait le joueur.
//
// Mort, il ne s'oppose plus à rien : une créature figée contre un cadavre serait
// arrêtée par ce que la partie ne montre plus.
func (w *World) surLeJoueur(x, y, rayon Fixed) bool {
	if !w.Alive() {
		return false
	}
	portee := w.profils.Player.Radius + rayon
	return (Vec{X: w.playerX - x, Y: w.playerY - y}).carres() < int64(portee)*int64(portee)
}

// dansUnCorps dit si un point tombe dans une créature qui bloque.
//
// **Le corps n'est solide que vivant**, ce que la conception pose pour que le
// blocage ne puisse pas devenir un piège : un joueur coincé entre un Vigile et
// un mur tire nécessairement dessus, puisque la visée prend le plus proche, et
// douze touches finissent par tomber. Une créature dont la résistance est
// tombée cesse donc de bloquer dans le tick même, sans attendre le nettoyage.
func (w *World) dansUnCorps(x, y Fixed) bool {
	for i := range w.ennemis.Active() {
		e := w.ennemis.At(i)
		profil := &w.profils.Enemies[e.Profile]
		if !profil.Solid || e.Hits <= 0 {
			continue
		}
		portee := w.profils.Player.Radius + profil.Radius
		if (Vec{X: e.X - x, Y: e.Y - y}).carres() < int64(portee)*int64(portee) {
			return true
		}
	}
	return false
}

// retirerLesMorts ferme le tick en vidant le bassin de ce qui n'a plus de
// résistance.
//
// Une créature morte a cessé d'être une cible dès l'instant où sa résistance est
// tombée, mais elle est restée en place : c'est ce qui permet aux boucles du tick
// de garder leurs index, et c'est pourquoi les suppressions viennent en dernier.
//
// La place libérée **est** réexaminée ici, à l'inverse de ce que fait une passe
// de mise à jour : celle-ci ferait avancer deux fois l'entité remontée, alors que
// ce nettoyage ne fait que filtrer — la sauter y laisserait un mort jusqu'au tick
// suivant.
// **C'est aussi le seul endroit qui compte les morts**, et il le peut parce que
// sortir de ce bassin veut dire mourir : rien d'autre n'y retire une créature.
// Le recyclage de la traîne, à l'étape 8, sera la première exception — il
// retirera des vivantes, et devra donc entrer par une autre porte que celle-ci
// sous peine d'ouvrir la sortie en éloignant la horde.
func (w *World) retirerLesMorts() {
	for i := 0; i < w.ennemis.Len(); {
		if w.ennemis.At(i).Hits <= 0 {
			w.ennemis.RemoveAt(i)
			w.abattus++
			continue
		}
		i++
	}
}

// passable dit si une position du monde tombe sur une case franchissable.
func (w *World) passable(x, y Fixed) bool {
	return w.grille.Passable(x.Floor(), y.Floor())
}
