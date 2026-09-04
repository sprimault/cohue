// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La lecture du manifeste des personnages, et la table de profils que la
// simulation indexe. Les champs qu'un seul rôle ou un seul comportement porte y
// sont exigés là et refusés ailleurs, par une table unique qui dit les deux à la
// fois.

package game

import (
	"fmt"
	"io/fs"
	"maps"
	"slices"
	"strings"

	"github.com/sprimault/cohue/internal/manifest"
)

// FormatProfiles est la version du manifeste de personnages que ce binaire lit.
const FormatProfiles = 1

// Behaviour est la façon dont un profil casse le kiting du joueur.
//
// Un type nommé plutôt qu'une chaîne nue : c'est ce qui empêche de comparer un
// comportement à un rôle ou à un nom de gabarit, trois chaînes que le manifeste
// pose côte à côte. Les valeurs sont celles du fichier, donc en français, alors
// que les identifiants restent en anglais comme le reste de l'API.
type Behaviour string

// Les comportements que le manifeste sait nommer.
const (
	Chase  Behaviour = "poursuite"
	Charge Behaviour = "charge"
	Flank  Behaviour = "flanc"
	Ranged Behaviour = "tir"
	Burst  Behaviour = "explosion"
	Heal   Behaviour = "soin"
	Wander Behaviour = "va_et_vient"
)

// behaviours est la liste close des comportements admis.
//
// Elle décrit un fichier que ce dépôt produit lui-même, ce qui en fait une
// seconde description de ce que `outils/ressources.py` exige déjà. Ce qui la
// rend tenable est le manifeste livré : deux tables en désaccord sur un champ ne
// peuvent pas rester vertes toutes les deux, puisque l'une d'elles refuse alors
// le seul fichier qui existe. Leur unique angle mort est un comportement que
// plus aucun profil n'exerce — celui qu'on vient d'ajouter et qu'on croit avoir
// éprouvé —, et c'est `TestToutComportementEstExerce` qui le ferme.
var behaviours = []Behaviour{Chase, Charge, Flank, Ranged, Burst, Heal, Wander}

// Les natures que le champ `role` distingue. Le joueur a une vie et pas de
// points, un ennemi l'inverse, une entité d'ambiance ni l'un ni l'autre.
const (
	rolePlayer  = "joueur"
	roleEnemy   = "ennemi"
	roleAmbient = "ambiance"
)

// roles est la liste close des natures admises.
var roles = []string{rolePlayer, roleEnemy, roleAmbient}

// Player est le personnage que la partie déplace.
type Player struct {
	// Name est le nom de fiction, celui que l'interface montre.
	Name string
	// Speed est la vitesse de base, en tuiles par tick.
	Speed Fixed
	// Radius est le rayon du corps, en tuiles.
	Radius Fixed
	// Health est le nombre de points de vie au départ.
	Health int
	// DamageCap plafonne les dégâts subis par seconde, quel que soit le nombre
	// d'ennemis au contact. Sans lui, un encerclement tue instantanément et la
	// mort devient illisible.
	DamageCap int
	// LowHealth est la vie sous laquelle l'écran signale le danger, en points.
	//
	// Il vit avec la vie et non dans le manifeste d'interface, bien que rien
	// dans la simulation ne le lise : il s'y compare, et l'en séparer le
	// laisserait derrière le jour où la vie change de valeur.
	LowHealth int
}

// EnemyProfile est ce qu'une sorte d'ennemi partage avec toutes ses instances.
//
// Une ligne de table, jamais une branche de code : ajouter une sorte d'ennemi
// est une entrée de plus dans le manifeste. Si l'une d'elles demandait du code,
// c'est qu'il manquerait un champ ici.
type EnemyProfile struct {
	// Key est la clé du profil dans le manifeste — `marcheur`, `sprinteur`.
	// C'est l'identifiant du moteur ; le nom est de la fiction.
	Key string
	// Name est le nom de fiction — Badaud, Molosse.
	Name string
	// Behaviour est la manière dont il approche le joueur.
	Behaviour Behaviour
	// Speed est sa vitesse, en tuiles par tick.
	Speed Fixed
	// Radius est le rayon de son corps, en tuiles.
	Radius Fixed
	// Hits est sa résistance, en touches de l'arme de base au premier niveau.
	// Jamais en points absolus : l'arme grossit toute la run, un chiffre absolu
	// ne voudrait rien dire. La conversion vivra là où l'on applique des dégâts.
	Hits int
	// Points est ce que sa mort rapporte au score du lieu.
	Points int
	// PressureCost est ce que le spawner paie pour en poser un.
	PressureCost int
	// SeparationWeight pondère la poussée qu'il exerce sur ses voisins.
	SeparationWeight Fixed
	// MaxAlive plafonne le nombre de vivants à la fois, zéro valant « aucun
	// plafond ». Il compte les vivants et non les apparus : un quota par run
	// ferait disparaître le profil après le premier.
	MaxAlive int
	// Group est le nombre de créatures que le spawner pose d'un coup, et il vaut
	// au moins un.
	//
	// Le Molosse n'apparaît jamais seul : trois qui chargent en décalé sont ce
	// qui oblige à cesser de reculer en ligne droite, quand un chien isolé se
	// contourne. Une meute est donc indivisible, et son prix est `PressureCost`
	// multiplié par elle — le coût du manifeste est unitaire.
	Group int
	// Solid dit si son corps arrête le joueur, ce qu'un seul profil fait.
	//
	// **Exigé de tous les ennemis, et non réservé à un comportement**, parce que
	// le Vigile poursuit comme le Badaud : rien dans son comportement ne
	// distingue celui qui bouche de celui qui suit. Un champ facultatif ferait
	// d'un Vigile dont on l'aurait oublié un colosse qui ne bloque rien — lent,
	// encaissant, et sans le rôle qui justifie ces deux traits.
	//
	// Il n'arrête que le joueur. Les créatures se traversent entre elles, et ce
	// qui empêche vingt Badauds de pousser un Vigile à travers une cloison est la
	// projection sur la passabilité, pas leurs corps.
	Solid bool
	// ContactDamage est ce qu'il inflige par seconde au contact.
	ContactDamage int
	// Gems est le nombre de gemmes que sa mort laisse au sol.
	//
	// Un nombre et non une valeur : une gemme rapporte la même chose du début à
	// la fin de la run, et c'est le seuil du niveau suivant qui monte. Ce qui
	// distingue une créature qui rapporte est donc la quantité qu'elle laisse,
	// ce que le joueur lit au sol avant de déclencher son aimant.
	Gems int

	// Ce qui suit n'a de sens que pour un comportement, et vaut zéro ailleurs —
	// le manifeste refusant de porter le champ sur un autre.

	// ChargeDamage est le choc d'une charge aboutie. Un coup unique, hors du
	// plafond de dégâts par seconde : ce que le plafond couvre est le contact
	// continu, qu'on ne voit pas venir dans une foule, là où une charge a été
	// annoncée puis manquée.
	ChargeDamage int
	// ChargeRange est la distance à laquelle la charge se déclenche, en tuiles,
	// et **son zéro est ce qui désactive le mécanisme** — un profil qui ne
	// charge pas ne porte aucun des quatre champs, le manifeste les réservant au
	// comportement.
	//
	// C'est elle qui décide, et non `Telegraph` : zéro milliseconde
	// d'anticipation est un réglage légitime autant que ce qu'un oubli produit,
	// quand une portée nulle n'a aucun autre sens.
	//
	// **Aucune ligne de vue n'est vérifiée.** Il charge à travers un pilier et
	// s'y arrête : c'est la mécanique, pas un défaut. La vérifier retirerait au
	// décor le seul usage défensif que la conception lui donne.
	ChargeRange Fixed
	// Telegraph est l'anticipation avant la course, en ticks. La créature y est
	// immobile — c'est ce qui laisse le temps de se décaler, et une anticipation
	// qui avance encore n'annonce rien.
	Telegraph Tick
	// ChargeDuration est la durée de la course, en ticks, quand rien ne
	// l'interrompt.
	ChargeDuration Tick
	// Recovery est le temps mort qui suit toute fin de course, en ticks.
	//
	// **Toute fin, et pas seulement l'échec au mur.** Sans lui, une charge
	// aboutie enchaîne sur la suivante et la créature n'a aucun moment
	// vulnérable ; l'esquive latérale que la conception lui oppose suppose
	// qu'esquiver laisse quelque chose.
	Recovery Tick
	// Tangential est la part du déplacement portée sur le côté plutôt que vers
	// la cible, ce qui produit le contournement.
	Tangential Fixed
	// Range est la distance à laquelle il ouvre le feu, en tuiles.
	Range Fixed
	// ShotDamage est ce qu'un de ses projectiles retire au joueur.
	ShotDamage int
	// ShotSpeed est la vitesse de ses projectiles, en tuiles par tick. Elle vit
	// ici et non sur l'objet qui vole : c'est elle qui décide si un tir de Buse
	// s'esquive, donc la seule vraie question d'équilibrage du profil.
	ShotSpeed Fixed
	// ShotCooldown est le temps entre deux tirs, en ticks.
	//
	// Sans lui, une créature à portée tirerait soixante fois par seconde : la
	// portée dit d'où elle atteint, elle ne dit rien de la fréquence, et les deux
	// se règlent séparément.
	ShotCooldown Tick
	// BurstDamage est ce que son explosion inflige au centre.
	BurstDamage int
	// BurstRadius est la portée de cette explosion, en tuiles.
	BurstRadius Fixed
	// HealRange est la distance à laquelle il rend des touches à ses voisines, en
	// tuiles, et **son zéro ferme le mécanisme** comme celui de `ChargeRange`
	// ferme la charge.
	HealRange Fixed
	// HealCooldown est le temps entre deux soins, en ticks.
	HealCooldown Tick
	// HealHits est ce qu'un soin rend, dans l'unité de la résistance.
	//
	// **Il ne se soigne pas lui-même**, et c'est ce qui garde la mécanique
	// entière : trois touches font de lui une cible qui tombe vite une fois
	// atteinte, et c'est cette récompense qui paie le trajet dans la horde. Un
	// soigneur qui se régénère la retirerait au moment de l'obtenir.
	//
	// **Une seule voisine par soin, la plus blessée à portée.** Soigner tout le
	// monde ferait de lui une invulnérabilité collective, quand la conception
	// écrit qu'il annule le travail — pas qu'il rend la horde invincible. La plus
	// blessée est en outre celle que le joueur est en train d'abattre, donc celle
	// dont la guérison se lit.
	HealHits int
	// Fuse est le temps entre la mort de la créature et sa détonation, en ticks.
	//
	// **C'est la durée du danger, donc une donnée de jeu**, et c'est ce qui la
	// range ici plutôt que dans la cadence du cycle d'animation qui l'annonce.
	// Le jour où le rendu lira les images de `assets/objets/`, l'animation devra
	// s'étirer sur cette durée : un télégraphe qui s'éteint avant la détonation
	// ment, et un mécanisme qui existe pour avertir ne peut pas se le permettre.
	Fuse Tick
}

// HitsAt rend la résistance de ce profil sous un durcissement donné.
//
// **L'arrondi est au plus proche et il n'a qu'un domicile.** Trois touches
// multipliées par 1,4 en font 4,2, et un profil doit tomber sur un entier : que
// deux profils de bases différentes arrondissent différemment sous le même
// multiplicateur est correct, mais qu'une seconde règle d'arrondi existe
// ailleurs changerait toutes les résistances du jeu le jour où l'une des deux
// bouge.
//
// **Le durcissement ne descend jamais sous un**, ce que le chargement refuse
// plutôt que ce qu'un plancher rattraperait : un profil à une touche sous un
// demi rendrait zéro, c'est-à-dire une créature morte à l'apparition. Ce champ
// dit une courbe qui durcit ; en faire aussi une courbe qui adoucit lui donnerait
// deux sens.
func (p *EnemyProfile) HitsAt(durcissement Fixed) int {
	return (FromInt(p.Hits).Mul(durcissement) + One/2).Floor()
}

// PackCost est ce que le spawner dépense pour une apparition de ce profil.
//
// Le coût du manifeste est unitaire et la meute est indivisible, si bien que le
// prix d'un achat n'est jamais celui d'une créature. Les deux endroits qui
// jugent un prix passent par ici — le test d'achat, et le plancher sous lequel
// la borne de report ne descend pas : le second oublié, une phase qui ne
// convoque que le Molosse plafonnerait son budget au prix d'un chien et
// n'achèterait jamais la meute.
func (p *EnemyProfile) PackCost() Fixed {
	return FromInt(p.PressureCost * p.Group)
}

// AmbientProfile est ce que partagent les figurants d'un lieu.
//
// **Une table à part, et tout en découle.** Ce qui n'est pas hostile n'entre
// dans aucun compte : ni le budget de pression, ni le plafond d'effectif, ni un
// objectif de porte fondé sur les kills. Le ranger parmi les ennemis avec un
// drapeau aurait obligé chaque compte à connaître l'exception, quand une table
// distincte les en dispense tous.
//
// **Et il n'est pas une cible.** La visée prend le plus proche sans que le
// joueur puisse choisir : un figurant dans le bassin des ennemis détournerait le
// tir sur des civils, et la mécanique du Secouriste — qui repose entièrement sur
// cette visée — tomberait avec. La séparation le garantit par construction, là
// où un test devrait le surveiller.
type AmbientProfile struct {
	// Key est la clé du profil dans le manifeste — `civil`.
	Key string
	// Name est le nom de fiction — Passant.
	Name string
	// Speed est sa vitesse, en tuiles par tick.
	Speed Fixed
	// Radius est le rayon de son corps, en tuiles.
	Radius Fixed
}

// Profiles est la table que la simulation indexe.
//
// Une entité ne garde que l'index de son profil dans `Enemies`, jamais une copie
// de ses valeurs : c'est ce qui rend une modification de table effective sans
// recharger le monde, et ce qui empêche deux Badauds d'avoir des valeurs
// différentes.
type Profiles struct {
	// Player est le profil du personnage joué.
	Player Player
	// Ambient sont les figurants, triés par clé comme les ennemis et pour la même
	// raison : leur index est ce qu'une entité conserve.
	Ambient []AmbientProfile
	// Enemies sont les profils d'ennemis, triés par clé de manifeste.
	//
	// Triés, parce que l'index dans cette tranche est ce que l'entité conserve :
	// rendus dans l'ordre de parcours d'une map, ils changeraient de numéro d'un
	// lancement à l'autre et le rejeu d'une graine ne vaudrait plus rien.
	Enemies []EnemyProfile
}

// LoadProfiles lit le manifeste des personnages et en tire la table de jeu.
//
// Les entités d'ambiance n'entrent pas dans `Enemies` : le Passant n'a ni dégâts
// ni points, il n'est pas un ennemi et n'aura pas sa place dans leur bassin. Le
// manifeste les porte quand même, parce qu'il décrit aussi ce que le rendu
// dessine, et elles sont validées comme les autres.
func LoadProfiles(fsys fs.FS, chemin string) (*Profiles, error) {
	brut, err := manifest.Decode[rawManifest](fsys, chemin)
	if err != nil {
		return nil, err
	}
	if brut.Format != FormatProfiles {
		return nil, fmt.Errorf("%s: %w : %d, ce binaire lit la %d",
			chemin, manifest.ErrUnsupportedFormat, brut.Format, FormatProfiles)
	}

	var manques []string
	dire := func(format string, args ...any) {
		manques = append(manques, fmt.Sprintf(format, args...))
	}

	cles := slices.Sorted(maps.Keys(brut.Profiles))
	for _, cle := range cles {
		controler(cle, brut.Profiles[cle], dire)
	}

	// Le joueur d'abord : sa vitesse de base est le référentiel de toutes les
	// autres, et rien ne garantit qu'il vienne en tête de l'ordre alphabétique.
	table := &Profiles{}
	var base float64
	joueurs := 0
	for _, cle := range cles {
		p := brut.Profiles[cle]
		if p.Role != rolePlayer {
			continue
		}
		joueurs++
		base = ou0(p.TilesPerSec)
		table.Player = p.joueur()
	}
	if joueurs != 1 {
		dire("profils : %d de rôle « %s », il en faut exactement un", joueurs, rolePlayer)
	}

	for _, cle := range cles {
		switch p := brut.Profiles[cle]; p.Role {
		case roleEnemy:
			table.Enemies = append(table.Enemies, p.ennemi(cle, base))
		case roleAmbient:
			table.Ambient = append(table.Ambient, p.figurant(cle, base))
		}
	}

	if len(manques) > 0 {
		return nil, &manifest.Invalid{Path: chemin, Missing: manques}
	}
	return table, nil
}

// rawManifest est le fichier tel que `outils/figurines.py` l'écrit.
type rawManifest struct {
	manifest.Commentable
	// Format est la version du format de manifeste.
	Format int `json:"version_format"`
	// Profiles sont les personnages, par clé.
	Profiles map[string]rawProfile `json:"profils"`
}

// rawProfile porte toutes les clés du fichier, et les pointeurs qui font la
// différence entre un champ absent et un zéro écrit.
//
// C'est la seule structure du paquet où cette distinction transparaît :
// `EnemyProfile` n'a aucun pointeur, la déréférence a lieu après le contrôle, et
// rien en aval ne teste `nil`. Sans elle, un `portee_tuiles` oublié sur la Buse
// donnerait une portée nulle, donc un profil qui ne tire jamais — un défaut qui
// passe en silence parce que la valeur d'absence est une valeur valide.
type rawProfile struct {
	manifest.Commentable

	Role       string   `json:"role"`
	Name       string   `json:"nom"`
	Behaviour  string   `json:"comportement,omitempty"`
	TileRadius *float64 `json:"rayon_tuiles"`

	TilesPerSec *float64 `json:"vitesse_tuiles_s,omitempty"`
	Health      *int     `json:"vie,omitempty"`
	DamageCap   *int     `json:"plafond_degats_s,omitempty"`
	LowHealth   *int     `json:"seuil_alerte_vie,omitempty"`

	RelSpeed     *float64 `json:"vitesse_relative,omitempty"`
	Hits         *int     `json:"touches,omitempty"`
	Points       *int     `json:"points,omitempty"`
	PressureCost *int     `json:"cout_pression,omitempty"`
	Separation   *float64 `json:"poids_separation,omitempty"`
	MaxAlive     *int     `json:"max_simultane,omitempty"`
	Group        *int     `json:"groupe,omitempty"`
	Solid        *bool    `json:"corps_bloquant,omitempty"`
	Contact      *int     `json:"degats_contact_s,omitempty"`
	Gems         *int     `json:"gemmes,omitempty"`

	ChargeDamage *int     `json:"degats_charge,omitempty"`
	ChargeRange  *float64 `json:"portee_charge_tuiles,omitempty"`
	TelegraphMs  *int     `json:"telegraphe_ms,omitempty"`
	ChargeMs     *int     `json:"duree_charge_ms,omitempty"`
	RecoveryMs   *int     `json:"recuperation_ms,omitempty"`
	Tangential   *float64 `json:"tangentiel,omitempty"`
	Range        *float64 `json:"portee_tuiles,omitempty"`
	ShotDamage   *int     `json:"degats_tir,omitempty"`
	ShotSpeed    *float64 `json:"vitesse_projectile_tuiles_s,omitempty"`
	ShotEveryMs  *int     `json:"cadence_tir_ms,omitempty"`
	BurstDamage  *int     `json:"degats_explosion,omitempty"`
	BurstRadius  *float64 `json:"rayon_explosion_tuiles,omitempty"`
	FuseMs       *int     `json:"amorce_ms,omitempty"`
	HealRange    *float64 `json:"portee_soin_tuiles,omitempty"`
	HealEveryMs  *int     `json:"cadence_soin_ms,omitempty"`
	HealHits     *int     `json:"soin_touches,omitempty"`

	// Ce qui suit décrit la figurine et son identité, et la simulation n'en lit
	// rien. Ces champs sont déclarés parce que le décodage refuse toute clé
	// inconnue : les retirer en constatant qu'ils ne servent à rien ferait
	// échouer le chargement du seul manifeste qui existe. Le rendu les lira à
	// l'étape 5 ; contrôler leur présence appartient au générateur qui les
	// écrit.
	Variants   int                 `json:"variantes"`
	Template   string              `json:"gabarit"`
	Origin     string              `json:"origine"`
	Side       int                 `json:"cote"`
	Anchor     [2]int              `json:"appui"`
	Directions []string            `json:"directions"`
	Cycles     map[string]rawCycle `json:"cycles"`
}

// rawCycle est une animation déclarée par le manifeste.
type rawCycle struct {
	manifest.Commentable
	// Frames est le nombre d'images de la bande.
	Frames int `json:"images"`
	// DurationMs est la durée d'une image.
	DurationMs int `json:"duree_ms"`
	// Loop dit si le cycle reprend à la fin.
	Loop bool `json:"boucle"`
}

// champsConditionnels dit, pour chaque champ que le manifeste ne porte pas
// partout, qui a le droit de le porter — et donc qui doit le porter.
//
// Une entrée dit les deux choses à la fois, et la seconde est celle qui attrape
// le copier-coller d'un profil vers un autre : un `tangentiel` resté sur un
// Badaud ne serait jamais lu, et laisserait croire à un réglage. Le contrôle est
// donc symétrique sans qu'il ait fallu l'écrire deux fois.
var champsConditionnels = []struct {
	// nom est la clé telle qu'elle s'écrit dans le fichier.
	nom string
	// qui désigne ce qui autorise le champ, pour le message.
	qui string
	// pour dit si ce profil doit porter le champ.
	pour func(rawProfile) bool
	// present dit s'il le porte.
	present func(rawProfile) bool
}{
	{"comportement", "un ennemi ou une ambiance", nonJoueur, func(p rawProfile) bool { return p.Behaviour != "" }},

	{"vitesse_tuiles_s", "un joueur", estRole(rolePlayer), func(p rawProfile) bool { return p.TilesPerSec != nil }},
	{"vie", "un joueur", estRole(rolePlayer), func(p rawProfile) bool { return p.Health != nil }},
	{"plafond_degats_s", "un joueur", estRole(rolePlayer), func(p rawProfile) bool { return p.DamageCap != nil }},
	{"seuil_alerte_vie", "un joueur", estRole(rolePlayer), func(p rawProfile) bool { return p.LowHealth != nil }},

	{"vitesse_relative", "un ennemi ou une ambiance", nonJoueur, func(p rawProfile) bool { return p.RelSpeed != nil }},
	{"touches", "un ennemi", estRole(roleEnemy), func(p rawProfile) bool { return p.Hits != nil }},
	{"points", "un ennemi", estRole(roleEnemy), func(p rawProfile) bool { return p.Points != nil }},
	{"cout_pression", "un ennemi", estRole(roleEnemy), func(p rawProfile) bool { return p.PressureCost != nil }},
	{"poids_separation", "un ennemi", estRole(roleEnemy), func(p rawProfile) bool { return p.Separation != nil }},
	{"max_simultane", "un ennemi", estRole(roleEnemy), func(p rawProfile) bool { return p.MaxAlive != nil }},
	{"groupe", "un ennemi", estRole(roleEnemy), func(p rawProfile) bool { return p.Group != nil }},
	{"corps_bloquant", "un ennemi", estRole(roleEnemy), func(p rawProfile) bool { return p.Solid != nil }},
	{"degats_contact_s", "un ennemi", estRole(roleEnemy), func(p rawProfile) bool { return p.Contact != nil }},
	{"gemmes", "un ennemi", estRole(roleEnemy), func(p rawProfile) bool { return p.Gems != nil }},

	{"degats_charge", "« charge »", estComportement(Charge), func(p rawProfile) bool { return p.ChargeDamage != nil }},
	{"portee_charge_tuiles", "« charge »", estComportement(Charge), func(p rawProfile) bool { return p.ChargeRange != nil }},
	{"telegraphe_ms", "« charge »", estComportement(Charge), func(p rawProfile) bool { return p.TelegraphMs != nil }},
	{"duree_charge_ms", "« charge »", estComportement(Charge), func(p rawProfile) bool { return p.ChargeMs != nil }},
	{"recuperation_ms", "« charge »", estComportement(Charge), func(p rawProfile) bool { return p.RecoveryMs != nil }},
	{"tangentiel", "« flanc »", estComportement(Flank), func(p rawProfile) bool { return p.Tangential != nil }},
	{"portee_tuiles", "« tir »", estComportement(Ranged), func(p rawProfile) bool { return p.Range != nil }},
	{"degats_tir", "« tir »", estComportement(Ranged), func(p rawProfile) bool { return p.ShotDamage != nil }},
	{"vitesse_projectile_tuiles_s", "« tir »", estComportement(Ranged), func(p rawProfile) bool { return p.ShotSpeed != nil }},
	{"cadence_tir_ms", "« tir »", estComportement(Ranged), func(p rawProfile) bool { return p.ShotEveryMs != nil }},
	{"degats_explosion", "« explosion »", estComportement(Burst), func(p rawProfile) bool { return p.BurstDamage != nil }},
	{"rayon_explosion_tuiles", "« explosion »", estComportement(Burst), func(p rawProfile) bool { return p.BurstRadius != nil }},
	{"amorce_ms", "« explosion »", estComportement(Burst), func(p rawProfile) bool { return p.FuseMs != nil }},

	{"portee_soin_tuiles", "« soin »", estComportement(Heal), func(p rawProfile) bool { return p.HealRange != nil }},
	{"cadence_soin_ms", "« soin »", estComportement(Heal), func(p rawProfile) bool { return p.HealEveryMs != nil }},
	{"soin_touches", "« soin »", estComportement(Heal), func(p rawProfile) bool { return p.HealHits != nil }},
}

// estRole rend le prédicat qui reconnaît une nature.
func estRole(role string) func(rawProfile) bool {
	return func(p rawProfile) bool { return p.Role == role }
}

// estComportement rend le prédicat qui reconnaît un comportement.
func estComportement(b Behaviour) func(rawProfile) bool {
	return func(p rawProfile) bool { return Behaviour(p.Behaviour) == b }
}

// nonJoueur reconnaît ce qui se déplace par rapport au joueur, ennemi comme
// ambiance : les deux tirent leur vitesse d'un rapport, et non d'une valeur.
func nonJoueur(p rawProfile) bool { return p.Role != rolePlayer }

// controler énumère tout ce qui empêche un profil d'être lu.
//
// Le rôle décide d'abord, et un rôle inconnu arrête là : sans lui, aucun
// prédicat de la table ne reconnaîtrait le profil et l'on déverserait seize
// manquements dont pas un ne nommerait la cause.
func controler(cle string, p rawProfile, dire func(string, ...any)) {
	if !slices.Contains(roles, p.Role) {
		dire("%s.role : « %s », attendu %s", cle, p.Role, liste(roles))
		return
	}
	if p.Name == "" {
		// Une chaîne vide et un champ absent sont le même défaut, et n'appellent
		// pas deux messages.
		dire("%s.nom : absent ou vide", cle)
	}
	if p.TileRadius == nil {
		dire("%s.rayon_tuiles : absent", cle)
	}
	if p.Behaviour != "" && !slices.Contains(behaviours, Behaviour(p.Behaviour)) {
		dire("%s.comportement : « %s », attendu %s", cle, p.Behaviour, liste(behaviours))
	}

	for _, champ := range champsConditionnels {
		switch {
		case champ.pour(p) && !champ.present(p):
			dire("%s.%s : absent, alors que %s l'exige", cle, champ.nom, champ.qui)
		case !champ.pour(p) && champ.present(p):
			dire("%s.%s : réservé à %s", cle, champ.nom, champ.qui)
		}
	}

	// La taille de meute est le seul champ dont le zéro serait accepté par la
	// table ci-dessus tout en n'ayant pas de sens : un profil qui n'apparaît
	// jamais est une entrée que rien ne signale. Le message ne double pas celui
	// de l'absence, qui vient d'être écrit le cas échéant.
	if p.Group != nil && *p.Group < 1 {
		dire("%s.groupe : %d, un profil apparaît au moins par un", cle, *p.Group)
	}

	// Les durées de la charge passent par le refus que `TicksFromMs` oppose déjà
	// partout ailleurs — zéro comme toute durée sous le pas de simulation. Elles
	// se convertissent ici sans qu'on garde le résultat : c'est ce qui laisse
	// `ennemi` convertir sans avoir à juger, comme `ou0` déréférence sans
	// contrôler.
	for _, d := range []struct {
		nom string
		ms  *int
	}{
		{"telegraphe_ms", p.TelegraphMs},
		{"duree_charge_ms", p.ChargeMs},
		{"recuperation_ms", p.RecoveryMs},
		{"cadence_tir_ms", p.ShotEveryMs},
		{"amorce_ms", p.FuseMs},
		{"cadence_soin_ms", p.HealEveryMs},
	} {
		if d.ms == nil {
			continue
		}
		if _, err := TicksFromMs(*d.ms); err != nil {
			dire("%s.%s : %v", cle, d.nom, err)
		}
	}

	// **Un coût nul n'est pas gratuit, il est impossible.** Le spawner tire le
	// profil qu'il va acheter avec un poids inversement proportionnel à son
	// prix : un prix de zéro y diviserait par zéro. Et le sens de jeu suit le
	// sens arithmétique — une créature qu'aucun budget ne limite remplirait le
	// bassin au premier tick.
	if p.PressureCost != nil && *p.PressureCost < 1 {
		dire("%s.cout_pression : %d, une créature que la pression n'achète pas "+
			"apparaîtrait sans limite", cle, *p.PressureCost)
	}

	// Une portée nulle laisserait un profil déclaré chargeur qui ne charge
	// jamais : le comportement serait au manifeste, les quatre champs présents,
	// et rien n'arriverait en jeu.
	if p.ChargeRange != nil && FromFloat(*p.ChargeRange) <= 0 {
		dire("%s.portee_charge_tuiles : %v, une portée que la virgule fixe "+
			"arrondit à zéro ne déclenche jamais la charge", cle, *p.ChargeRange)
	}

	// Une meute plus grande que le plafond de simultanéité ne tiendrait jamais
	// entière, et le spawner l'écarterait à chaque tick sans que rien ne le
	// dise : le profil serait au manifeste, autorisé par une phase, et
	// n'apparaîtrait pas une fois de la partie.
	if p.Group != nil && p.MaxAlive != nil && *p.MaxAlive > 0 && *p.Group > *p.MaxAlive {
		dire("%s.groupe : %d au-dessus de max_simultane %d, la meute ne tiendrait jamais",
			cle, *p.Group, *p.MaxAlive)
	}
}

// joueur tire le profil du personnage joué.
//
// La conversion suppose le contrôle passé, mais elle s'exécute même quand il a
// signalé quelque chose : les manquements se listent en une fois, donc rien ne
// s'interrompt en route. Ce qu'elle produit alors n'est jamais rendu.
func (p rawProfile) joueur() Player {
	return Player{
		Name:      p.Name,
		Speed:     parTick(ou0(p.TilesPerSec)),
		Radius:    FromFloat(ou0(p.TileRadius)),
		Health:    ou0(p.Health),
		DamageCap: ou0(p.DamageCap),
		LowHealth: ou0(p.LowHealth),
	}
}

// ennemi tire un profil d'ennemi, sa vitesse rapportée à celle du joueur.
//
// Le produit se fait ici, une fois, et contre la vitesse **de base** du joueur :
// contre la vitesse courante, un bonus de vitesse accélérerait la horde d'autant
// et n'aurait plus aucun effet. Il se fait aussi en flottants, avant le passage
// en virgule fixe, pour n'arrondir qu'une fois — l'arrondi de la vitesse du
// joueur suivi de celui du produit ferait dériver la horde d'une fraction de
// tuile par minute, sans que rien ne le montre.
func (p rawProfile) ennemi(cle string, base float64) EnemyProfile {
	return EnemyProfile{
		Key:              cle,
		Name:             p.Name,
		Behaviour:        Behaviour(p.Behaviour),
		Speed:            parTick(ou0(p.RelSpeed) * base),
		Radius:           FromFloat(ou0(p.TileRadius)),
		Hits:             ou0(p.Hits),
		Points:           ou0(p.Points),
		PressureCost:     ou0(p.PressureCost),
		SeparationWeight: FromFloat(ou0(p.Separation)),
		MaxAlive:         ou0(p.MaxAlive),
		Group:            ou0(p.Group),
		Solid:            ou0(p.Solid),
		ContactDamage:    ou0(p.Contact),
		Gems:             ou0(p.Gems),
		ChargeDamage:     ou0(p.ChargeDamage),
		ChargeRange:      FromFloat(ou0(p.ChargeRange)),
		Telegraph:        ticks(p.TelegraphMs),
		ChargeDuration:   ticks(p.ChargeMs),
		Recovery:         ticks(p.RecoveryMs),
		Tangential:       FromFloat(ou0(p.Tangential)),
		Range:            FromFloat(ou0(p.Range)),
		ShotDamage:       ou0(p.ShotDamage),
		ShotSpeed:        parTick(ou0(p.ShotSpeed)),
		ShotCooldown:     ticks(p.ShotEveryMs),
		BurstDamage:      ou0(p.BurstDamage),
		BurstRadius:      FromFloat(ou0(p.BurstRadius)),
		Fuse:             ticks(p.FuseMs),
		HealRange:        FromFloat(ou0(p.HealRange)),
		HealCooldown:     ticks(p.HealEveryMs),
		HealHits:         ou0(p.HealHits),
	}
}

// figurant tire le profil d'une entité d'ambiance.
//
// Elle ne retient que ce qu'il faut pour le déplacer et le dessiner : un
// figurant n'a ni résistance, ni dégâts, ni points, et le manifeste lui refuse
// ces champs. Sa vitesse se rapporte à celle du joueur comme celle d'un ennemi —
// c'est la seule chose qu'ils partagent, et `nonJoueur` le dit déjà.
func (p rawProfile) figurant(cle string, base float64) AmbientProfile {
	return AmbientProfile{
		Key:    cle,
		Name:   p.Name,
		Speed:  parTick(ou0(p.RelSpeed) * base),
		Radius: FromFloat(ou0(p.TileRadius)),
	}
}

// ou0 déréférence un champ facultatif, l'absence valant zéro.
//
// Sans contrôle : c'est `controler` qui décide si l'absence est un défaut, et
// dupliquer ce jugement ici produirait deux messages pour un seul manquement.
func ou0[T any](v *T) T {
	if v == nil {
		var zero T
		return zero
	}
	return *v
}

// ticks convertit une durée facultative de manifeste, l'absence valant zéro.
//
// Sans contrôle, pour la même raison qu'`ou0` : `controler` a déjà refusé les
// durées qu'un tick ne porte pas, et rejuger ici donnerait deux messages pour un
// seul manquement.
func ticks(ms *int) Tick {
	if ms == nil {
		return 0
	}
	t, _ := TicksFromMs(*ms)
	return t
}

// parTick convertit un débit par seconde vers le pas de simulation.
//
// Les vitesses en tuiles par seconde d'abord, et le budget de pression d'un
// scénario ensuite : un débit se ramène au tick de la même façon, et l'écrire
// deux fois aurait donné deux arrondis à tenir d'accord.
func parTick(parSeconde float64) Fixed {
	return FromFloat(parSeconde / TPS)
}

// liste énumère des valeurs admises pour un message de refus, ce qui vaut mieux
// qu'un adjectif : l'auteur y lit du premier coup ce qu'il pouvait écrire.
// listeDesEnnemis énumère les profils achetables, chaque clé suivie de son nom
// de fiction.
//
// **Les deux, parce que l'auteur d'un lieu arrive avec le mauvais des deux.** Le
// fichier attend la clé du moteur — `flanqueur` —, quand tout ce qu'un humain
// lit ailleurs porte le nom — Arpenteur : la règle du jeu, la table des rôles,
// l'écran. Un refus qui ne listerait que les clés laisserait donc chercher la
// correspondance dans un manifeste, et un refus qui ne listerait que les noms
// serait inutilisable. Les apparier ne crée aucune seconde description : le
// manifeste porte déjà les deux champs, et c'est lui qu'on relit.
//
// L'ordre est celui de la table, triée par clé au chargement, donc stable d'un
// lancement à l'autre.
func listeDesEnnemis(profils *Profiles) string {
	libelles := make([]string, len(profils.Enemies))
	for i, e := range profils.Enemies {
		libelles[i] = "« " + e.Key + " » (" + e.Name + ")"
	}
	return strings.Join(libelles, ", ")
}

// listeDesFigurants énumère les profils d'ambiance, clé et nom de fiction.
//
// Même appariement que pour les ennemis, et pour la même raison : le fichier
// attend la clé quand tout ce qu'un humain lit porte le nom.
func listeDesFigurants(profils *Profiles) string {
	libelles := make([]string, len(profils.Ambient))
	for i, a := range profils.Ambient {
		libelles[i] = "« " + a.Key + " » (" + a.Name + ")"
	}
	return strings.Join(libelles, ", ")
}

func liste[T ~string](valeurs []T) string {
	libelles := make([]string, len(valeurs))
	for i, v := range valeurs {
		libelles[i] = "« " + string(v) + " »"
	}
	return strings.Join(libelles, ", ")
}
