// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le Passant : un figurant qui va et vient sans rien décider du jeu. Il ne
// compte nulle part, ne se vise pas, ne pousse personne — et son errance tire
// dans le flux cosmétique, le seul dont rien ne dépend.

package game

import (
	"fmt"

	"github.com/sprimault/cohue/internal/manifest"
)

// AmbientSpec est le peuplement de figurants tel qu'un lieu l'écrit.
type AmbientSpec []AmbientGroup

// AmbientGroup pose un figurant à une case donnée.
//
// **Une position et non un nombre, et ce n'est pas une préférence de format.**
// Un peuplement tiré au sort abandonne en silence les positions qui tombent dans
// un mur : un lieu qui demande douze figurants en pose neuf sans que personne ne
// l'apprenne. Écrites, elles se refusent au chargement, où l'auteur les lit.
//
// Le reste suit : tout ce qu'un lieu contient se place — les pièces portent leur
// `u` et leur `v` —, et un tirage relancé à chaque mort ferait sauter les
// figurants d'un coin à l'autre entre deux essais, quand tout le reste de la
// salle ne bouge pas.
type AmbientGroup struct {
	manifest.Commentable
	// Profile est la clé du profil d'ambiance — `civil`.
	Profile string `json:"profil"`
	// At est la case où le figurant commence, en coordonnées de lieu.
	At *[2]int `json:"position"`
}

// AmbientPlacement est un figurant compilé : un index de profil et une position.
type AmbientPlacement struct {
	// Profile est l'index dans `Profiles.Ambient`.
	Profile int
	// X et Y sont sa position de départ, au centre de la case écrite.
	X, Y Fixed
}

// CompileAmbient résout un peuplement écrit contre la table des figurants et la
// carte cuite.
//
// Elle rend tout ce qui l'empêche de valoir plutôt que le premier écart, comme
// la compilation des vagues, et pour la même raison : qui met au point un lieu
// veut la liste.
//
// **Un profil d'ennemi y est refusé autant qu'un nom inconnu.** Poser un Badaud
// en figurant le sortirait du budget de pression et du plafond d'effectif tout
// en le laissant dans la salle — une horde gratuite, que rien dans la courbe
// n'expliquerait.
//
// **La carte entre ici pour que les positions se vérifient**, et c'est tout
// l'intérêt de les écrire : un figurant posé dans un mur est un défaut du
// fichier, et il se dit au chargement plutôt que de disparaître en silence.
func CompileAmbient(brut AmbientSpec, profils *Profiles, carte *CostGrid) ([]AmbientPlacement, []string) {
	var manques []string
	dire := func(format string, args ...any) {
		manques = append(manques, fmt.Sprintf(format, args...))
	}

	pose := make([]AmbientPlacement, 0, len(brut))
	for i, g := range brut {
		ou := fmt.Sprintf("ambiance[%d]", i)

		rang := -1
		for j := range profils.Ambient {
			if profils.Ambient[j].Key == g.Profile {
				rang = j
				break
			}
		}
		if rang < 0 {
			dire("%s.profil : « %s » n'est pas un profil d'ambiance, attendu %s",
				ou, g.Profile, listeDesFigurants(profils))
			continue
		}
		if g.At == nil {
			dire("%s.position : absente, un figurant se place", ou)
			continue
		}

		u, v := g.At[0], g.At[1]
		if !carte.InBounds(u, v) {
			dire("%s.position : (%d, %d) hors du lieu, qui fait %d sur %d",
				ou, u, v, carte.Width(), carte.Height())
			continue
		}
		if !carte.Passable(u, v) {
			dire("%s.position : (%d, %d) est dans un mur", ou, u, v)
			continue
		}

		// Au centre de la case et non sur son coin, comme le joueur : c'est là
		// que se tient une entité.
		pose = append(pose, AmbientPlacement{
			Profile: rang,
			X:       FromInt(u) + One/2,
			Y:       FromInt(v) + One/2,
		})
	}
	return pose, manques
}

// paliersErrance est le nombre de directions qu'un figurant sait prendre.
//
// Huit, comme les orientations du monde : un figurant qui marcherait dans une
// direction quelconque demanderait au rendu une image que le manifeste ne
// fournit pas, et le décalage se verrait sur une silhouette lente.
const paliersErrance = Headings

// Ambient est un figurant : une position, une direction, et le temps qui reste
// avant qu'il en change.
//
// **Il ne porte ni résistance ni dégâts, et ce n'est pas une simplification.**
// Une entité d'ambiance qu'on pourrait tuer entrerait dans un objectif de porte
// fondé sur les kills ; une qui blesserait entrerait dans le plafond de dégâts.
// Ce qui n'est pas hostile n'entre dans aucun compte, et la façon la plus sûre
// de le tenir est qu'il n'ait rien à compter.
type Ambient struct {
	// Profile est l'index de son profil dans `Profiles.Ambient`.
	Profile int
	// X et Y sont sa position dans le monde, en tuiles.
	X, Y Fixed
	// Heading est le rang de sa direction parmi les huit du monde.
	Heading int
	// Until est ce qui reste avant qu'il choisisse une autre direction.
	Until Tick
	// Step est le déplacement que le tick précédent lui a appliqué. Le rendu s'en
	// sert comme de celui d'une créature ; rien d'autre ne le lit.
	Step Vec
}

// errer déplace les figurants et leur fait changer de cap de temps en temps.
//
// **Tous ses tirages passent par le flux cosmétique**, et c'est ce qui les rend
// sans conséquence : ce flux ne décide de rien par construction, si bien qu'un
// figurant de plus ou de moins ne déplace pas d'un pixel ce que la simulation
// retient. C'est aussi ce qui donne enfin un consommateur à ce flux, et rend
// exigible le test qui sépare le cosmétique du reste — jusqu'ici annoncé sans
// pouvoir rien séparer.
//
// **Ils ne sont pas comptés dans la densité.** Un figurant qui pousserait la
// horde deviendrait exploitable : on apprendrait à se placer derrière une foule
// de civils pour la dévier, mécanique tactique réelle que personne n'a voulue et
// qu'il faudrait ensuite équilibrer ou retirer.
func (w *World) errer() {
	for i := range w.ambiants.Active() {
		a := w.ambiants.At(i)
		profil := &w.profils.Ambient[a.Profile]

		if a.Until > 0 {
			a.Until--
		} else {
			a.Heading = w.hasard.Cosmetic.Pick(paliersErrance)
			a.Until = w.palierDErrance()
		}

		avantX, avantY := a.X, a.Y
		pas := Heading(a.Heading).Scale(w.vitesse(profil.Speed, a.X, a.Y))
		a.X, a.Y = w.glisser(a.X, a.Y, pas)
		a.Step = Vec{X: a.X - avantX, Y: a.Y - avantY}

		// Un mur ne l'arrête pas longtemps : la direction se retire au tick
		// suivant plutôt qu'à la fin du palier, sans quoi un figurant coincé dans
		// un angle y frotterait plusieurs secondes.
		if a.Step == (Vec{}) {
			a.Until = 0
		}
	}
}

// Les bornes d'un palier d'errance, en ticks.
//
// Une seconde et demie au moins, quatre au plus : en deçà, un figurant tremble
// sur place au lieu de traverser ; au-delà, il quitte l'écran en ligne droite et
// l'ambiance se vide là où le joueur regarde.
const (
	erranceMin   Tick = 90
	erranceEcart Tick = 150
)

// Populate pose le peuplement d'un lieu, au montage et à chaque relance.
//
// Les positions viennent du fichier et sont déjà vérifiées : cette passe ne
// décide de rien, elle recopie. Ce qui reste tiré est le cap de départ, dans le
// flux cosmétique — deux relances placent donc les mêmes figurants aux mêmes
// endroits, marchant vers autre chose.
func (w *World) Populate(figurants []AmbientPlacement) {
	for _, f := range figurants {
		w.SpawnAmbient(f.Profile, f.X, f.Y)
	}
}

// SpawnAmbient pose un figurant, sa première direction tirée au cosmétique.
//
// Le palier de départ est tiré comme les suivants : sans cela, tous les
// figurants d'un lieu changeraient de cap au même tick, et une foule censée
// paraître quelconque marcherait au pas.
func (w *World) SpawnAmbient(profil int, x, y Fixed) (Handle, bool) {
	return w.ambiants.Spawn(Ambient{
		Profile: profil,
		X:       x,
		Y:       y,
		Heading: w.hasard.Cosmetic.Pick(paliersErrance),
		Until:   w.palierDErrance(),
	})
}

// palierDErrance tire la durée d'un cap, en ticks.
//
// Elle existe pour que la conversion vers `Tick` n'ait qu'un domicile : `Pick`
// rend un `int` borné par son argument, ce qu'un analyseur ne peut pas déduire,
// et la garde s'écrirait autrement à chacun des deux appels.
func (w *World) palierDErrance() Tick {
	pas := w.hasard.Cosmetic.Pick(int(erranceEcart))
	return erranceMin + Tick(pas) // #nosec G115 -- borné par erranceEcart
}

// Ambients rend le bassin des figurants.
func (w *World) Ambients() *Pool[Ambient] { return w.ambiants }

// AmbientRadius rend le rayon d'un figurant, que le rendu pose au sol.
func (w *World) AmbientRadius(a *Ambient) Fixed { return w.profils.Ambient[a.Profile].Radius }
