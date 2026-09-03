// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La lecture du manifeste des armes, le seul de `assets/` tenu à la main. Le
// tireur y porte les valeurs de son tir : cadence, portée, dégâts, nombre de
// projectiles et vitesse.

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
	// **Ce chemin n'a jamais été parcouru : le manifeste livré en déclare un.**
	// `tirer` engendre les copies au même point, dans la même direction et au
	// même pas — elles sont rigoureusement confondues à l'écran, et `toucher`
	// écartant ce qui n'a plus de résistance, la seconde va chercher derrière la
	// première. Ce que le mécanisme produit aujourd'hui est donc une salve qui
	// perfore en profondeur, dessinée comme un seul point.
	//
	// Ce n'est aucune des choses que la conception nomme, et trois lectures
	// restent ouvertes : un étalement dans le temps, qui ferait du nombre une
	// rafale ; un étalement dans l'espace, qui est l'éventail, un axe distinct ;
	// ou la superposition telle quelle, assumée comme de la perforation. Trancher
	// est une décision de conception, et la prendre pour avoir un axe de plus
	// serait la prendre pour la mauvaise raison — c'est pourquoi l'axe du nombre
	// n'est pas ouvert au jalon 3.
	Projectiles int
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

	Name        string   `json:"nom"`
	Role        string   `json:"role"`
	CadenceMs   *int     `json:"cadence_ms"`
	TileRange   *float64 `json:"portee_tuiles"`
	Hits        *int     `json:"degats_touches"`
	Projectiles *int     `json:"projectiles"`
	Speed       *float64 `json:"vitesse_projectile_tuiles_s"`
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
		ProjectileSpeed: parTick(exige(cle, "vitesse_projectile_tuiles_s", a.Speed, dire)),
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
