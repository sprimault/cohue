// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les passifs : les axes d'amélioration de l'arme unique, leurs paliers, et la
// carte de secours qui remplit une place quand aucun palier n'est disponible.

package game

import (
	"fmt"
	"maps"
	"slices"

	"github.com/sprimault/cohue/internal/manifest"
)

// Axis nomme un axe d'amélioration, et le moteur le connaît par ce nom.
//
// Comme un comportement de créature : la clé du manifeste décide quel champ
// l'entrée doit porter, ce qui rend le contrôle symétrique — un `pas_ms` sur
// l'axe de portée ne serait jamais lu et laisserait croire à un réglage.
type Axis string

// Les axes que le manifeste sait nommer.
//
// **Les six que la conception veut.** Ce qui reste à l'étape 6 après eux est ce
// qui les combine — les recettes de fusion —, et non un axe de plus.
const (
	AxisCadence     Axis = "cadence"
	AxisRange       Axis = "portee"
	AxisProjectiles Axis = "projectiles"
	AxisPierce      Axis = "perforant"
	AxisBounce      Axis = "ricochet"
	AxisSpread      Axis = "eventail"
)

// axes est la liste close des axes admis.
var axes = []Axis{
	AxisCadence, AxisRange, AxisProjectiles, AxisPierce, AxisBounce, AxisSpread,
}

// Passive est un axe d'amélioration, pris palier par palier.
//
// Un axe n'offre qu'une carte à la fois — son palier suivant —, et c'est ce qui
// permet à six axes de six paliers de remplir une trentaine de montées sans
// trente entrées de table.
//
// Les deux pas sont exclusifs : chaque axe porte le sien et refuse celui de
// l'autre, comme un profil de créature ne porte que les champs de son
// comportement. Celui qui ne le concerne pas vaut zéro et n'est jamais lu.
type Passive struct {
	// Axis est la clé du manifeste, et ce que le moteur reconnaît.
	Axis Axis
	// Name est le nom de fiction, celui que la carte affiche.
	Name string
	// Phrase est la ligne qui dit ce que l'axe fait, en clair.
	Phrase string
	// Tiers est le nombre de fois qu'on peut le prendre.
	//
	// Une borne, et c'est une règle de conception : l'épuiser oblige à basculer
	// sur un axe qu'on n'avait pas choisi, ce qui est un moment de jeu. Sans
	// elle, empiler un seul axe du début à la fin serait la stratégie unique.
	Tiers int

	// Effects porte la ligne d'effet de chaque palier, du premier au dernier.
	//
	// **Composée au chargement parce que la composer au tirage allouait.** Une
	// montée de niveau tombe dans un tick comme un autre, et l'invariant du
	// budget ne connaît pas d'exception pour les ticks rares. `Tiers` étant
	// borné et validé, la tranche est courte et se remplit une fois.
	Effects []string

	// CooldownStep est ce qu'un palier retire à la cadence, en ticks.
	CooldownStep Tick
	// RangeStep est ce qu'un palier ajoute à la portée, en tuiles.
	RangeStep Fixed
	// ProjectileStep est ce qu'un palier ajoute au nombre de projectiles.
	//
	// Il ne s'accompagne d'aucun élargissement : la largeur du front est celle
	// de l'arme et ne bouge pas d'un palier à l'autre — voir `Weapon.Front`.
	ProjectileStep int
	// PierceStep est ce qu'un palier ajoute aux créatures traversées, BounceStep
	// aux rebonds.
	//
	// Deux champs et deux axes, bien qu'ils portent la même unité : ce qu'ils
	// changent au tir n'est pas de même nature — l'un prolonge la course, l'autre
	// la redirige —, et `Projectile.Pierce` dit pourquoi l'ordre entre eux est
	// fixé.
	PierceStep int
	BounceStep int
	// SpreadStep est ce qu'un palier ajoute à la largeur de la salve à la portée.
	//
	// En tuiles comme `RangeStep`, et jamais en degrés : `Weapon.Spread` dit
	// pourquoi, et c'est une contrainte de déterminisme plutôt qu'un choix
	// d'unité.
	SpreadStep Fixed
}

// Relief est la carte qui remplit une place quand aucun palier ne le peut.
//
// **Répétable dans un même tirage**, et c'est ce qui garantit qu'un écran de
// montée ne se vide jamais — le pire défaut possible à cet endroit, puisqu'il
// tombe sur le moment de récompense. Sans cette propriété, le cas extrême où
// tout est épuisé rendrait un écran vide au moment précis où le joueur a le
// mieux joué.
//
// Un soin et non un gain d'expérience : le second se rembourserait en accélérant
// la montée suivante, donc il serait toujours au moins aussi bon que ce qu'il
// remplace, et il aplatirait la table entière sans qu'on voie pourquoi.
type Relief struct {
	// Name est le nom de fiction.
	Name string
	// Phrase est la ligne qui dit ce qu'elle fait.
	Phrase string
	// Heal est ce qu'elle rend au joueur, en points de vie.
	Heal int
	// Effect est sa ligne d'effet, composée au chargement comme celles des axes.
	Effect string
}

// Passives est la table des améliorations.
type Passives struct {
	// Axes sont les axes, triés par clé de manifeste.
	//
	// Triés pour la même raison que les profils d'ennemis : ce qui est retenu
	// d'une carte est une place dans cette tranche, et l'ordre de parcours d'une
	// map la ferait changer d'un lancement à l'autre.
	Axes []Passive
	// Relief est la carte de secours.
	Relief Relief
	// Recipes sont les fusions, triées par clé de manifeste.
	//
	// Triées pour la même raison que les axes : une carte retient une place dans
	// cette tranche, et l'ordre de parcours d'une map la ferait changer d'un
	// lancement à l'autre.
	Recipes []Recipe
}

// Axes rend la table des axes, dans l'ordre où les cartes les offrent.
//
// En lecture : c'est la table du manifeste, que le monde ne recopie pas. Elle
// est indexée comme `TiersTaken`, et les deux se lisent ensemble — celle-ci
// porte les noms et les bornes, l'autre ce qui est acquis.
func (w *World) Axes() []Passive { return w.passifs.Axes }

// TiersTaken rend les paliers pris sur chaque axe, indexé comme `Axes`.
//
// **C'est l'état de la partie que le joueur ne peut pas restituer de mémoire.**
// Ce qu'il a choisi décide de ce que son arme est devenue, et l'écart entre deux
// parties se lit là avant de se lire ailleurs ; les deux tranches séparées plutôt
// qu'une structure composée évitent une allocation à chaque lecture.
func (w *World) TiersTaken() []int { return w.paliers }

// rawPassives est la section telle qu'elle s'écrit.
type rawPassives struct {
	manifest.Commentable
	// Axes sont les axes, par clé.
	Axes map[string]rawAxis `json:"axes"`
	// Relief est la carte de secours.
	Relief rawRelief `json:"soupape"`
	// Recipes sont les fusions, par clé.
	Recipes map[string]rawRecipe `json:"recettes"`
}

// rawAxis porte les champs d'un axe, en pointeurs pour les pas dont zéro est
// une réponse plausible — un axe qui n'améliore rien.
type rawAxis struct {
	manifest.Commentable

	// Name et Phrase sont ce que la carte montre : son titre, et la ligne qui dit
	// ce qu'elle fait.
	Name   string `json:"nom"`
	Phrase string `json:"phrase"`
	// Tiers est le nombre de fois qu'on peut prendre l'axe. Épuisé, il sort du
	// tirage, et c'est ce qui fait de sa borne un moment de jeu.
	Tiers *int `json:"paliers"`

	// Les pas, dont chaque axe porte le sien et refuse ceux des autres.
	// CadenceMs se retranche de l'écart entre deux salves, Tiles s'ajoute à la
	// portée, Shots au nombre de projectiles, Pierce à ce qu'un tir traverse et
	// Bounces à ce vers quoi il repart ; d'où les unités qui les séparent, des
	// millisecondes, des tuiles et des comptes.
	//
	// Les trois derniers comptent la même chose sans être interchangeables : la
	// clé décide lequel l'entrée doit porter, et un `pas_rebonds` posé sur le
	// perforant est refusé plutôt que lu.
	CadenceMs *int     `json:"pas_ms,omitempty"`
	Tiles     *float64 `json:"pas_tuiles,omitempty"`
	Shots     *int     `json:"pas_projectiles,omitempty"`
	Pierce    *int     `json:"pas_perforations,omitempty"`
	Bounces   *int     `json:"pas_rebonds,omitempty"`
	Spread    *float64 `json:"pas_eventail,omitempty"`
}

// rawRelief porte les champs de la carte de secours.
type rawRelief struct {
	manifest.Commentable

	// Name et Phrase sont ce que la carte montre, comme pour un axe.
	Name   string `json:"nom"`
	Phrase string `json:"phrase"`
	// Heal est la vie qu'elle rend, en points. C'est ce qui la rend
	// situationnelle — précieuse quand on est bas, ignorable quand on est
	// haut —, là où un gain d'expérience se rembourserait et serait toujours bon.
	Heal *int `json:"soin"`
}

// passifs convertit la section, en signalant ce qui lui manque.
//
// Le contrôle croisé avec l'arme de base est la raison d'être de cette
// signature : un axe de cadence dont tous les paliers pris tomberaient à zéro
// tick produirait une arme qui tire à chaque image, ce qu'aucun test de la table
// seule ne verrait — les deux valeurs vivent dans le même fichier et se lisent
// ensemble.
func (p rawPassives) passifs(base Weapon, dire func(string, ...any)) *Passives {
	table := &Passives{}

	for _, cle := range slices.Sorted(maps.Keys(p.Axes)) {
		table.Axes = append(table.Axes, p.Axes[cle].axe(Axis(cle), base, dire))
	}
	table.Relief = p.Relief.soupape(dire)

	// Après les axes, parce qu'une recette se contrôle contre eux : ses
	// ingrédients doivent nommer des axes existants, et n'en exiger que ce que
	// leur borne permet.
	for _, cle := range slices.Sorted(maps.Keys(p.Recipes)) {
		table.Recipes = append(table.Recipes, p.Recipes[cle].recette(cle, table, dire))
	}
	return table
}

// axe convertit un axe brut et contrôle ce que sa clé impose.
func (a rawAxis) axe(cle Axis, base Weapon, dire func(string, ...any)) Passive {
	if !slices.Contains(axes, cle) {
		dire("passifs.axes.%s : axe inconnu, attendu %s", cle, liste(axes))
	}
	if a.Name == "" {
		dire("passifs.axes.%s.nom : absent ou vide", cle)
	}
	if a.Phrase == "" {
		dire("passifs.axes.%s.phrase : absent ou vide", cle)
	}

	nom := fmt.Sprintf("passifs.axes.%s", cle)
	axe := Passive{
		Axis:   cle,
		Name:   a.Name,
		Phrase: a.Phrase,
		Tiers:  exige(nom, "paliers", a.Tiers, dire),
	}
	if a.Tiers != nil && axe.Tiers < 1 {
		dire("%s.paliers : %d, un axe qu'on ne peut pas prendre n'est pas un axe", nom, axe.Tiers)
	}
	for palier := 1; palier <= axe.Tiers; palier++ {
		axe.Effects = append(axe.Effects, fmt.Sprintf("Palier %d sur %d", palier, axe.Tiers))
	}

	// Chaque axe porte son pas et refuse celui de l'autre. Le contrôle est
	// symétrique sans qu'il ait fallu l'écrire deux fois : un `pas_tuiles` resté
	// sur la cadence après un copier-coller ne serait jamais lu.
	for _, c := range []struct {
		champ   string
		pour    Axis
		present bool
	}{
		{"pas_ms", AxisCadence, a.CadenceMs != nil},
		{"pas_tuiles", AxisRange, a.Tiles != nil},
		{"pas_projectiles", AxisProjectiles, a.Shots != nil},
		{"pas_perforations", AxisPierce, a.Pierce != nil},
		{"pas_rebonds", AxisBounce, a.Bounces != nil},
		{"pas_eventail", AxisSpread, a.Spread != nil},
	} {
		switch {
		case cle == c.pour && !c.present:
			dire("%s.%s : absent, et « %s » l'exige", nom, c.champ, c.pour)
		case cle != c.pour && c.present:
			dire("%s.%s : présent, réservé à « %s »", nom, c.champ, c.pour)
		}
	}

	switch cle {
	case AxisCadence:
		if a.CadenceMs == nil {
			return axe
		}
		ticks, err := TicksFromMs(*a.CadenceMs)
		if err != nil {
			dire("%s.pas_ms : %v", nom, err)
			return axe
		}
		axe.CooldownStep = ticks

		// **L'axe entier ne doit pas épuiser la cadence.** À zéro tick, l'arme
		// tirerait à chaque image et la cadence cesserait d'être un réglage ;
		// négative, elle deviendrait un compteur qui ne redescend jamais. Le
		// contrôle est ici parce que c'est le seul endroit qui voie les deux
		// valeurs — la table seule ne peut pas savoir de quelle arme elle part.
		//
		// Il se tait quand l'arme n'a pas de cadence : cette absence est déjà
		// signalée, et la compter une seconde fois ferait corriger deux lignes
		// pour une faute — celle-ci, en prime, désignerait la table alors que le
		// défaut est dans l'arme.
		if base.Cooldown < 1 {
			return axe
		}

		// **Le contrôle divise, il ne multiplie pas.** `paliers` vient du fichier
		// et n'a pas de borne propre : le produit déborderait le compteur de
		// ticks avant d'être comparé, et la comparaison rendrait n'importe quoi
		// sur la valeur qui aurait justement dû être refusée. La division ne peut
		// pas déborder, et elle dit la règle directement — le nombre de paliers
		// qu'une cadence supporte.
		if maxi := int(base.Cooldown-1) / int(ticks); axe.Tiers > maxi {
			dire("%s : %d paliers de %d ticks sur une cadence de %d, %d au plus",
				nom, axe.Tiers, ticks, base.Cooldown, maxi)
		}
	case AxisRange:
		if a.Tiles != nil {
			axe.RangeStep = FromFloat(*a.Tiles)
			if axe.RangeStep < 1 {
				dire("%s.pas_tuiles : %v, un pas que la virgule fixe arrondit à "+
					"zero n ameliore rien", nom, *a.Tiles)
			}
		}
	case AxisProjectiles:
		if a.Shots != nil {
			axe.ProjectileStep = *a.Shots
			if axe.ProjectileStep < 1 {
				dire("%s.pas_projectiles : %d, un palier qui n'ajoute aucun "+
					"projectile ne change rien à la salve", nom, axe.ProjectileStep)
			}
		}
	case AxisPierce:
		if a.Pierce != nil {
			axe.PierceStep = *a.Pierce
			if axe.PierceStep < 1 {
				dire("%s.pas_perforations : %d, un palier qui ne traverse "+
					"personne de plus ne change rien au tir", nom, axe.PierceStep)
			}
		}
	case AxisBounce:
		if a.Bounces != nil {
			axe.BounceStep = *a.Bounces
			if axe.BounceStep < 1 {
				dire("%s.pas_rebonds : %d, un palier qui n'ajoute aucun rebond "+
					"ne change rien au tir", nom, axe.BounceStep)
			}
		}
	case AxisSpread:
		if a.Spread != nil {
			axe.SpreadStep = FromFloat(*a.Spread)
			if axe.SpreadStep < 1 {
				dire("%s.pas_eventail : %v, un pas que la virgule fixe arrondit "+
					"à zero n ecarte rien", nom, *a.Spread)
			}
		}
	}
	return axe
}

// soupape convertit la carte de secours.
func (r rawRelief) soupape(dire func(string, ...any)) Relief {
	if r.Name == "" {
		dire("passifs.soupape.nom : absent ou vide")
	}
	if r.Phrase == "" {
		dire("passifs.soupape.phrase : absent ou vide")
	}

	secours := Relief{
		Name:   r.Name,
		Phrase: r.Phrase,
		Heal:   exige("passifs.soupape", "soin", r.Heal, dire),
	}
	if r.Heal != nil && secours.Heal < 1 {
		dire("passifs.soupape.soin : %d, une carte qui ne fait rien laisse le "+
			"joueur devant un choix vide", secours.Heal)
	}
	secours.Effect = fmt.Sprintf("+%d vie", secours.Heal)
	return secours
}
