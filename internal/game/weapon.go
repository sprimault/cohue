// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La lecture du manifeste des armes, l'un des deux de `assets/` tenus à la main
// avec celui de la progression. Le tireur y porte les valeurs de son tir :
// cadence, portée, dégâts, nombre de projectiles et vitesse.

package game

import (
	"fmt"
	"io/fs"
	"maps"
	"slices"

	"github.com/sprimault/cohue/internal/manifest"
)

// FormatWeapons est la version du manifeste d'armes que ce binaire lit.
const FormatWeapons = 1

// roleBase désigne l'armement infini, celui qui monte de niveau et porte la
// build. Exactement une arme le porte, comme exactement un profil porte le rôle
// de joueur.
const roleBase = "base"

// Weapon est une arme, telle que le manifeste tenu à la main la décrit.
//
// C'est le tireur qui porte les valeurs de son tir : cadence, portée, dégâts,
// nombre de projectiles et vitesse. Le projectile, lui, n'est qu'un objet qui
// vole, et son manifeste ne décrit que son apparence.
type Weapon struct {
	// Key est la clé de l'arme dans le manifeste.
	Key string
	// Name est le nom de fiction.
	Name string
	// Cooldown est le délai entre deux tirs, en ticks.
	Cooldown Tick
	// Range est la portée, en tuiles.
	Range Fixed
	// Hits est ce qu'un projectile retire à une créature, dans l'unité où
	// s'exprime leur résistance. L'arme de base au premier niveau en inflige
	// une : c'est elle qui définit l'unité, et le chiffre est donc tautologique
	// ici — il cessera de l'être à la première arme qui frappe plus fort.
	Hits int
	// Projectiles est le nombre de projectiles par tir.
	//
	// **Ils partent en front parallèle** : décalés perpendiculairement à la
	// visée, même direction et même vitesse. Les trois autres lectures que ce
	// champ a longtemps laissées ouvertes sont écartées, et deux d'entre elles
	// parce qu'elles feraient doublon avec un axe voisin — la superposition est
	// ce que `perforant` apporte, l'étalement en angle est `eventail`, et la
	// recette « trois projectiles plus éventail » n'aurait alors rien à
	// combiner. La troisième, une rafale étalée dans le temps, dirait ce que
	// `cadence` dit déjà et ne montrerait jamais qu'un projectile à la fois,
	// quand ce qu'on attend d'un nombre est d'abord qu'il se voie.
	//
	// **À cible unique, trois projectiles ne valent pas trois fois un**, et cela
	// se lirait comme un défaut sans cette phrase. Seul celui du centre suit la
	// ligne d'interception ; les autres passent à côté d'une cible ponctuelle et
	// vont chercher derrière. C'est ce qui sépare l'axe d'un multiplicateur de
	// dégâts : il paie contre la masse, et ne rend rien contre une Buse isolée.
	//
	// **Cela cesse d'être vrai en montant, et ce n'est pas une contradiction.**
	// `Front` gardant une largeur constante, la salve se densifie à chaque
	// palier : à sept projectiles ils sont assez serrés pour qu'une même
	// créature en prenne deux ou trois, si bien que l'axe redevient en partie le
	// multiplicateur qu'il n'était pas à trois. C'est une progression où l'axe
	// change de nature plutôt qu'un effet de bord — le spécialiste qui a dépensé
	// six choix sur lui gagne contre la masse d'abord, et contre la cible unique
	// ensuite.
	Projectiles int
	// Front est la largeur totale du front, en tuiles, quel que soit le nombre.
	//
	// **C'est la largeur qui se règle, et l'écartement entre voisins qui s'en
	// dérive.** L'inverse — un écartement fixe entre deux projectiles — élargit
	// le front d'un palier à l'autre, et les extrêmes d'une salve de sept
	// passeraient à plusieurs tuiles de la visée : le dernier palier tirerait de
	// plus en plus large au lieu de rapporter. Dérivée de la largeur, la salve se
	// densifie.
	//
	// Un front nul n'est pas une absence légitime : il superposerait les
	// projectiles, c'est-à-dire la salve confondue que ce champ remplace.
	Front Fixed
	// Pierce est le nombre de créatures qu'un tir traverse au-delà de la
	// première, et Bounces le nombre de fois qu'il repart vers une autre cible.
	//
	// Zéro pour l'arme de base, et c'est une valeur et non une absence : un tir
	// qui s'arrête sur ce qu'il touche est le comportement ordinaire. Les deux
	// champs restent exigés du fichier pour cette raison même — omis, ils
	// vaudraient zéro sans qu'on sache si c'était voulu.
	Pierce  int
	Bounces int
	// Spread est ce que la salve gagne en largeur à la portée de l'arme.
	//
	// **Une largeur et non un angle, et c'est le déterminisme qui l'impose.**
	// L'IEEE-754 ne garantit pas le dernier bit de `sin` et `cos` d'une
	// architecture à l'autre, et cette simulation tourne sur trois cibles dont
	// deux arm64 : un éventail exprimé en degrés y aurait fait diverger deux
	// binaires publiés sur la même graine. Exprimée en tuiles, l'ouverture donne
	// un point à viser, et la direction sort de `Direction`, qui normalise par
	// `sqrt` — la seule opération dont l'arrondi correct est garanti. **Ne pas
	// « simplifier » ce champ en angle.**
	//
	// **Elle s'ajoute à `Front` au lieu de la remplacer.** Un premier palier qui
	// ramènerait les départs au canon retirerait la couverture rapprochée que le
	// front donne : ce serait le seul palier de la table à faire perdre quelque
	// chose. À ouverture nulle la salve reste parallèle, au bit près.
	//
	// **Elle ne fait rien sur une salve d'un seul tir**, la répartition étant
	// centrée : c'est la synergie « projectiles plus éventail » prise par l'autre
	// bout. La carte reste offerte à qui n'a pas encore de front — les cartes ne
	// s'auto-censurent pas —, et c'est son libellé qui doit le dire.
	Spread Fixed
	// ProjectileSpeed est la vitesse d'un projectile, en tuiles par tick.
	ProjectileSpeed Fixed
}

// Weapons est la table des armes, et des passifs qui les transforment.
//
// Les passifs voyagent avec elles parce qu'ils n'améliorent rien d'autre : une
// valeur vit à côté de ce qu'elle alimente. C'est aussi ce qui permet de
// contrôler qu'un axe de cadence n'épuise pas celle de l'arme de base, ce que
// deux fichiers auraient rendu invérifiable au chargement.
type Weapons struct {
	// Base est l'armement infini du joueur.
	Base Weapon
	// All sont toutes les armes, triées par clé de manifeste.
	All []Weapon
	// Passives sont les axes d'amélioration et la carte de secours.
	Passives *Passives
}

// LoadWeapons lit le manifeste des armes.
//
// L'un des deux de `assets/` qui ne sortent d'aucun générateur, avec celui de la
// progression, et c'est délibéré : ce sont les chiffres qu'on rouvrira le plus
// pendant l'équilibrage, et les loger dans un fichier généré ferait passer chaque
// réglage de cadence par un script Python, donc par une régénération de six cents
// images.
func LoadWeapons(fsys fs.FS, chemin string) (*Weapons, error) {
	brut, err := manifest.Decode[rawWeapons](fsys, chemin)
	if err != nil {
		return nil, err
	}
	if brut.Format != FormatWeapons {
		return nil, fmt.Errorf("%s: %w : %d, ce binaire lit la %d",
			chemin, manifest.ErrUnsupportedFormat, brut.Format, FormatWeapons)
	}

	var manques []string
	dire := func(format string, args ...any) {
		manques = append(manques, fmt.Sprintf(format, args...))
	}

	table := &Weapons{}
	bases := 0
	for _, cle := range slices.Sorted(maps.Keys(brut.Weapons)) {
		a := brut.Weapons[cle]
		arme := a.arme(cle, dire)
		table.All = append(table.All, arme)
		if a.Role == roleBase {
			bases++
			table.Base = arme
		}
	}
	if bases != 1 {
		dire("armes : %d de rôle « %s », il en faut exactement une", bases, roleBase)
	}

	// Après les armes, parce que le contrôle d'un axe de cadence se fait contre
	// celle de l'arme de base. Sans arme de base, il se ferait contre une valeur
	// nulle et signalerait un second défaut qui n'est que la conséquence du
	// premier — l'auteur corrigerait deux lignes pour une faute.
	table.Passives = brut.Passives.passifs(table.Base, dire)

	if len(manques) > 0 {
		return nil, &manifest.Invalid{Path: chemin, Missing: manques}
	}
	return table, nil
}

// rawWeapons est le fichier tel qu'il s'écrit.
type rawWeapons struct {
	manifest.Commentable
	// Format est la version du format de manifeste.
	Format int `json:"version_format"`
	// Weapons sont les armes, par clé.
	Weapons map[string]rawWeapon `json:"armes"`
	// Passives est la table des améliorations.
	Passives rawPassives `json:"passifs"`
}

// rawWeapon porte les champs d'une arme, en pointeurs pour les valeurs dont
// zéro est une réponse plausible.
//
// Une cadence, une portée ou des dégâts nuls sont des absences déguisées : une
// arme qui tire à portée zéro ne tire jamais, et rien à l'écran ne dirait que le
// champ manque au fichier.
type rawWeapon struct {
	manifest.Commentable

	// Name est le nom lisible de l'arme.
	Name string `json:"nom"`
	// Role dit ce qu'elle est dans la partie. Seul l'armement de base existe, et
	// c'est le champ qui refusera une arme lourde tombée dans cette table.
	Role string `json:"role"`
	// CadenceMs est l'écart entre deux salves, converti en ticks au chargement.
	CadenceMs *int `json:"cadence_ms"`
	// TileRange est la distance au-delà de laquelle une cible n'est plus visée.
	TileRange *float64 `json:"portee_tuiles"`
	// Hits est ce qu'une touche retire à une créature, dans l'unité où le
	// manifeste des personnages compte leur résistance.
	Hits *int `json:"degats_touches"`
	// Projectiles est le nombre de projectiles d'une salve.
	Projectiles *int `json:"projectiles"`
	// FrontTuiles est la largeur sur laquelle une salve se répartit.
	FrontTuiles *float64 `json:"front_tuiles"`
	// Pierce et Bounces sont ce qu'un tir traverse et ce vers quoi il repart.
	Pierce  *int `json:"perforations"`
	Bounces *int `json:"rebonds"`
	// EventailTuiles est ce que la salve gagne en largeur à la portée.
	EventailTuiles *float64 `json:"eventail_tuiles"`
	// Speed est la vitesse d'un projectile, en tuiles par seconde.
	Speed *float64 `json:"vitesse_projectile_tuiles_s"`
}

// arme convertit une arme brute, en signalant ce qui lui manque.
func (a rawWeapon) arme(cle string, dire func(string, ...any)) Weapon {
	if a.Name == "" {
		dire("%s.nom : absent ou vide", cle)
	}
	if a.Role != roleBase {
		dire("%s.role : « %s », attendu « %s »", cle, a.Role, roleBase)
	}

	w := Weapon{
		Key:             cle,
		Name:            a.Name,
		Range:           FromFloat(exige(cle, "portee_tuiles", a.TileRange, dire)),
		Hits:            exige(cle, "degats_touches", a.Hits, dire),
		Projectiles:     exige(cle, "projectiles", a.Projectiles, dire),
		Front:           FromFloat(exige(cle, "front_tuiles", a.FrontTuiles, dire)),
		Pierce:          exige(cle, "perforations", a.Pierce, dire),
		Bounces:         exige(cle, "rebonds", a.Bounces, dire),
		Spread:          FromFloat(exige(cle, "eventail_tuiles", a.EventailTuiles, dire)),
		ProjectileSpeed: parTick(exige(cle, "vitesse_projectile_tuiles_s", a.Speed, dire)),
	}

	// Un front que la virgule fixe ramène à zéro superpose les projectiles, ce
	// qui est la salve confondue d'avant l'axe : le nombre monterait sans que
	// rien ne se voie. Le refus se tait quand le champ manque, son absence étant
	// déjà signalée.
	if a.FrontTuiles != nil && w.Front < 1 {
		dire("%s.front_tuiles : %v, un front que la virgule fixe arrondit à zéro "+
			"superpose les projectiles", cle, *a.FrontTuiles)
	}

	// La cadence passe par la conversion commune, qui refuse une durée sous le
	// pas de simulation : une arme à cinq millisecondes ne tirerait pas deux
	// cents fois par seconde, elle tirerait une fois par tick sans que le
	// fichier le dise.
	if ms := exige(cle, "cadence_ms", a.CadenceMs, dire); ms > 0 {
		ticks, err := TicksFromMs(ms)
		if err != nil {
			dire("%s.cadence_ms : %v", cle, err)
		}
		w.Cooldown = ticks
	} else if a.CadenceMs != nil {
		// Zéro échappait à la conversion, donc au refus qu'elle porte : l'arme
		// tirait à chaque image, ce qui ne ressemble pas à un fichier invalide
		// mais à un moteur cassé.
		dire("%s.cadence_ms : %d, une arme qui tire à chaque image n'a plus de cadence", cle, ms)
	}
	return w
}

// exige déréférence un champ obligatoire, ou signale son absence.
func exige[T any](cle, champ string, v *T, dire func(string, ...any)) T {
	if v == nil {
		dire("%s.%s : absent", cle, champ)
		var zero T
		return zero
	}
	return *v
}
