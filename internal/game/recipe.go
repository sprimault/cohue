// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les recettes de fusion : ce qu'une recette exige d'axes montés, l'effet qu'elle
// donne au tir, et le vocabulaire fermé où la table le choisit.

package game

import (
	"fmt"
	"maps"
	"slices"

	"github.com/sprimault/cohue/internal/manifest"
)

// Effect nomme ce qu'une recette fait au tir, et le moteur le connaît par ce nom.
//
// **Un vocabulaire fermé, comme la liste des axes.** La table choisit un
// comportement, elle n'en invente pas : cinq recettes qui déclareraient chacune
// le leur feraient cinq mécanismes de tir, ce que l'étape 6 n'annonce pas. Le
// critère que la conception en tire décide aussi de leur nombre — une recette qui
// demanderait du code neuf dans `toucher` est trop ambitieuse pour la table où
// elle s'écrit, et il y en a deux parce que deux mécanismes s'y prêtaient.
type Effect string

// Les effets que le manifeste sait nommer.
const (
	// EffectRail retire sa borne à la perforation : le tir ne s'arrête plus sur
	// rien de vivant tant qu'il lui reste de la portée.
	EffectRail Effect = "rail"
	// EffectSpray fait converger la salve au point visé au lieu de l'en écarter.
	EffectSpray Effect = "gerbe"
)

// effets est la liste close des effets admis.
var effets = []Effect{EffectRail, EffectSpray}

// Ingredient est un axe et ce qu'une recette exige d'avoir pris dessus.
//
// **Les paliers pris, jamais l'état de l'arme.** « Trois projectiles » se lit
// « deux paliers de projectiles » : la table reste juste quand on règle ce dont
// l'arme part, et c'est la partie qui tient déjà ce compte.
type Ingredient struct {
	// Axis est l'axe exigé.
	Axis Axis
	// Tiers est le nombre de paliers exigés, quand la recette n'exige pas l'axe
	// au bout.
	Tiers int
	// Spent dit que la recette exige l'axe **épuisé**.
	//
	// **Un drapeau et non la borne recopiée**, qui serait une seconde
	// description : écrire six paliers ici laisserait la table mentir le jour où
	// la borne d'un axe bougerait, et rien ne le signalerait.
	//
	// Il sert à ce que la conception exige : un ingrédient dont la recette annule
	// l'utilité s'exige épuisé, faute de quoi les paliers restants ne rendraient
	// plus rien et le joueur dépenserait des choix pour zéro.
	Spent bool
}

// Recipe est une fusion : ce qu'elle exige, et ce qu'elle donne.
//
// **C'est une carte et non un compteur qui monte.** Elle entre dans le tirage
// quand ses ingrédients sont réunis, et se prend ou se laisse comme les autres —
// appliquer l'effet dès que les ingrédients sont là rendrait le joueur plus fort
// sans qu'il l'ait décidé, et rien ne lui dirait qu'une fusion a eu lieu.
type Recipe struct {
	// Key est la clé de la recette dans le manifeste.
	Key string
	// Name est le nom de fiction, celui que la carte affiche.
	Name string
	// Phrase est la ligne qui dit ce qu'elle fait, en clair.
	Phrase string
	// Effect est ce qu'elle donne au tir.
	Effect Effect
	// Ingredients sont les axes qu'elle exige, dans l'ordre du manifeste.
	Ingredients []Ingredient
}

// rawRecipe porte les champs d'une recette telle qu'elle s'écrit.
type rawRecipe struct {
	manifest.Commentable

	// Name et Phrase sont ce que la carte montre.
	Name   string `json:"nom"`
	Phrase string `json:"phrase"`
	// Effect est la clé de l'effet, prise dans le vocabulaire fermé.
	Effect string `json:"effet"`
	// Ingredients sont les axes exigés, par clé d'axe.
	Ingredients map[string]rawIngredient `json:"ingredients"`
}

// rawIngredient porte ce qu'une recette exige d'un axe.
//
// Les deux champs s'excluent, et le contrôle est symétrique : un ingrédient qui
// n'exige rien n'est pas un ingrédient, et un qui exigerait les deux ne dit pas
// lequel fait foi.
type rawIngredient struct {
	manifest.Commentable

	// Tiers est le nombre de paliers exigés.
	Tiers *int `json:"paliers,omitempty"`
	// Spent exige l'axe épuisé.
	Spent *bool `json:"epuise,omitempty"`
}

// recette convertit une recette brute, en signalant ce qui lui manque.
func (r rawRecipe) recette(cle string, table *Passives, dire func(string, ...any)) Recipe {
	nom := fmt.Sprintf("passifs.recettes.%s", cle)
	if r.Name == "" {
		dire("%s.nom : absent ou vide", nom)
	}
	if r.Phrase == "" {
		dire("%s.phrase : absent ou vide", nom)
	}
	if !slices.Contains(effets, Effect(r.Effect)) {
		dire("%s.effet : « %s » inconnu, attendu %s", nom, r.Effect, liste(effets))
	}

	recette := Recipe{
		Key:    cle,
		Name:   r.Name,
		Phrase: r.Phrase,
		Effect: Effect(r.Effect),
	}
	if len(r.Ingredients) == 0 {
		dire("%s.ingredients : absent ou vide, une recette sans ingredient est toujours offerte", nom)
		return recette
	}

	for _, axe := range slices.Sorted(maps.Keys(r.Ingredients)) {
		brut := r.Ingredients[axe]
		champ := fmt.Sprintf("%s.ingredients.%s", nom, axe)
		if !slices.Contains(axes, Axis(axe)) {
			dire("%s : axe inconnu, attendu %s", champ, liste(axes))
		}

		ingredient := Ingredient{Axis: Axis(axe)}
		switch {
		case brut.Tiers != nil && brut.Spent != nil:
			dire("%s : « paliers » et « epuise » ensemble, sans dire lequel fait foi", champ)
		case brut.Spent != nil:
			ingredient.Spent = *brut.Spent
			if !*brut.Spent {
				dire("%s.epuise : faux, ce qui n'exige rien — retirer la clé ou "+
					"écrire des paliers", champ)
			}
		case brut.Tiers != nil:
			ingredient.Tiers = *brut.Tiers
			if ingredient.Tiers < 1 {
				dire("%s.paliers : %d, un ingredient qui n'exige aucun palier est "+
					"toujours reuni", champ, ingredient.Tiers)
			}
			if borne := borneDe(table, Axis(axe)); borne > 0 && ingredient.Tiers > borne {
				dire("%s.paliers : %d sur un axe qui en porte %d, la recette ne "+
					"paraitrait jamais", champ, ingredient.Tiers, borne)
			}
		default:
			dire("%s : ni « paliers » ni « epuise », un ingredient doit exiger "+
				"quelque chose", champ)
		}
		recette.Ingredients = append(recette.Ingredients, ingredient)
	}
	return recette
}

// borneDe rend le nombre de paliers d'un axe, ou zéro s'il est inconnu.
//
// Zéro pour l'inconnu plutôt qu'une erreur : l'axe inconnu est déjà signalé, et
// le compter une seconde fois ferait corriger deux lignes pour une faute.
func borneDe(table *Passives, cle Axis) int {
	for i := range table.Axes {
		if table.Axes[i].Axis == cle {
			return table.Axes[i].Tiers
		}
	}
	return 0
}
