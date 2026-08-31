// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

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
	// ContactDamage est ce qu'il inflige par seconde au contact.
	ContactDamage int

	// Ce qui suit n'a de sens que pour un comportement, et vaut zéro ailleurs —
	// le manifeste refusant de porter le champ sur un autre.

	// ChargeDamage est le choc d'une charge aboutie.
	ChargeDamage int
	// Tangential est la part du déplacement portée sur le côté plutôt que vers
	// la cible, ce qui produit le contournement.
	Tangential Fixed
	// Range est la distance à laquelle il ouvre le feu, en tuiles.
	Range Fixed
	// BurstDamage est ce que son explosion inflige au centre.
	BurstDamage int
	// BurstRadius est la portée de cette explosion, en tuiles.
	BurstRadius Fixed
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
		if p := brut.Profiles[cle]; p.Role == roleEnemy {
			table.Enemies = append(table.Enemies, p.ennemi(cle, base))
		}
	}

	if len(manques) > 0 {
		return nil, &manifest.Invalide{Chemin: chemin, Manques: manques}
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

	RelSpeed     *float64 `json:"vitesse_relative,omitempty"`
	Hits         *int     `json:"touches,omitempty"`
	Points       *int     `json:"points,omitempty"`
	PressureCost *int     `json:"cout_pression,omitempty"`
	Separation   *float64 `json:"poids_separation,omitempty"`
	MaxAlive     *int     `json:"max_simultane,omitempty"`
	Contact      *int     `json:"degats_contact_s,omitempty"`

	ChargeDamage *int     `json:"degats_charge,omitempty"`
	Tangential   *float64 `json:"tangentiel,omitempty"`
	Range        *float64 `json:"portee_tuiles,omitempty"`
	BurstDamage  *int     `json:"degats_explosion,omitempty"`
	BurstRadius  *float64 `json:"rayon_explosion_tuiles,omitempty"`

	// Ce qui suit décrit la figurine et son identité, et la simulation n'en lit
	// rien. Ces champs sont déclarés parce que le décodage refuse toute clé
	// inconnue : les retirer en constatant qu'ils ne servent à rien ferait
	// échouer le chargement du seul manifeste qui existe. Le rendu les lira à
	// l'étape 5 ; contrôler leur présence appartient au générateur qui les
	// écrit.
	Group      int                 `json:"groupe"`
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

	{"vitesse_relative", "un ennemi ou une ambiance", nonJoueur, func(p rawProfile) bool { return p.RelSpeed != nil }},
	{"touches", "un ennemi", estRole(roleEnemy), func(p rawProfile) bool { return p.Hits != nil }},
	{"points", "un ennemi", estRole(roleEnemy), func(p rawProfile) bool { return p.Points != nil }},
	{"cout_pression", "un ennemi", estRole(roleEnemy), func(p rawProfile) bool { return p.PressureCost != nil }},
	{"poids_separation", "un ennemi", estRole(roleEnemy), func(p rawProfile) bool { return p.Separation != nil }},
	{"max_simultane", "un ennemi", estRole(roleEnemy), func(p rawProfile) bool { return p.MaxAlive != nil }},
	{"degats_contact_s", "un ennemi", estRole(roleEnemy), func(p rawProfile) bool { return p.Contact != nil }},

	{"degats_charge", "« charge »", estComportement(Charge), func(p rawProfile) bool { return p.ChargeDamage != nil }},
	{"tangentiel", "« flanc »", estComportement(Flank), func(p rawProfile) bool { return p.Tangential != nil }},
	{"portee_tuiles", "« tir »", estComportement(Ranged), func(p rawProfile) bool { return p.Range != nil }},
	{"degats_explosion", "« explosion »", estComportement(Burst), func(p rawProfile) bool { return p.BurstDamage != nil }},
	{"rayon_explosion_tuiles", "« explosion »", estComportement(Burst), func(p rawProfile) bool { return p.BurstRadius != nil }},
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
		ContactDamage:    ou0(p.Contact),
		ChargeDamage:     ou0(p.ChargeDamage),
		Tangential:       FromFloat(ou0(p.Tangential)),
		Range:            FromFloat(ou0(p.Range)),
		BurstDamage:      ou0(p.BurstDamage),
		BurstRadius:      FromFloat(ou0(p.BurstRadius)),
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

// parTick convertit une vitesse en tuiles par seconde vers le pas de simulation.
func parTick(tuilesParSeconde float64) Fixed {
	return FromFloat(tuilesParSeconde / TPS)
}

// liste énumère des valeurs admises pour un message de refus, ce qui vaut mieux
// qu'un adjectif : l'auteur y lit du premier coup ce qu'il pouvait écrire.
func liste[T ~string](valeurs []T) string {
	libelles := make([]string, len(valeurs))
	for i, v := range valeurs {
		libelles[i] = "« " + string(v) + " »"
	}
	return strings.Join(libelles, ", ")
}
