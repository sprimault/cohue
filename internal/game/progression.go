// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le rythme des choix : la lecture du manifeste des seuils, l'expérience qui
// s'accumule, et les deux façons de monter d'un niveau — la gemme et le temps.

package game

import (
	"fmt"
	"io/fs"

	"github.com/sprimault/cohue/internal/manifest"
)

// FormatProgression est la version du manifeste de progression que ce binaire
// lit.
const FormatProgression = 1

// Progression est ce qui commande le rythme des montées de niveau.
//
// Elle ne vit pas auprès des armes, et le critère n'est pas qu'un manifeste de
// plus soit plus propre : un seuil appartient à la partie. Il vaut quelle que
// soit l'arme équipée, il continuerait d'exister si la table d'armes changeait
// entièrement, et il commande le rythme des choix plutôt que leur contenu.
type Progression struct {
	// FirstThreshold est le nombre de gemmes qui mène du premier niveau au
	// second.
	FirstThreshold int
	// Increment est ce que chaque niveau ajoute au seuil du suivant.
	//
	// C'est le seuil qui monte, jamais la valeur d'une gemme : une gemme rapporte
	// la même chose du début à la fin de la run, et une créature qui doit
	// rapporter davantage en laisse plusieurs. L'inverse mélangerait ce que vaut
	// un kill et le rythme auquel les choix arrivent.
	Increment int
	// Floor est le temps au bout duquel un niveau est donné sans rien ramasser,
	// en ticks.
	Floor Tick
	// GemValue est ce qu'une gemme ramassée porte à l'expérience.
	//
	// Un scalaire et non une table, parce que la simulation ne distingue pas
	// deux sortes de gemmes : `Gem` ne porte pas de profil, et une table serait
	// un réglage que personne ne lirait — c'est exactement ce qui s'était
	// installé quand cette valeur vivait dans le manifeste d'objets sans que rien
	// n'aille l'y chercher. Le jour où une grosse gemme existera, c'est `Gem` qui
	// portera sa sorte, et ce champ deviendra une table indexée comme les profils.
	GemValue int
}

// Threshold rend ce que coûte le passage d'un niveau au suivant, en gemmes.
//
// La forme est affine et non tabulée. Ce qui se règle est le seuil ; ce qui se
// juge est l'écart entre deux choix, c'est-à-dire ce seuil divisé par une
// cadence de récolte qui croît elle aussi. Aucune des deux courbes n'est
// mesurable avant le spawner, et une table de trente valeurs donnerait
// l'illusion d'une précision qu'on n'a pas — avec, en prime, une règle à
// inventer pour ce qui se passe après sa dernière ligne.
func (p *Progression) Threshold(niveau int) int {
	return p.FirstThreshold + p.Increment*(niveau-1)
}

// LoadProgression lit le manifeste des seuils.
//
// Tenu à la main comme celui des armes, et pour la même raison : ces chiffres
// s'ajustent en rejouant, et les loger dans un fichier généré ferait passer
// chaque essai d'équilibrage par un script Python.
func LoadProgression(fsys fs.FS, chemin string) (*Progression, error) {
	brut, err := manifest.Decode[rawProgression](fsys, chemin)
	if err != nil {
		return nil, err
	}
	if brut.Format != FormatProgression {
		return nil, fmt.Errorf("%s: %w : %d, ce binaire lit la %d",
			chemin, manifest.ErrUnsupportedFormat, brut.Format, FormatProgression)
	}

	var manques []string
	dire := func(format string, args ...any) {
		manques = append(manques, fmt.Sprintf(format, args...))
	}

	n := brut.Progression.Levels
	table := &Progression{
		FirstThreshold: exige("niveaux", "seuil_premier", n.First, dire),
		Increment:      exige("niveaux", "seuil_increment", n.Increment, dire),
	}

	// Un seuil nul ou négatif ferait boucler la montée sans fin : chaque tour
	// retirerait zéro gemme et donnerait un niveau. Le refus est ici plutôt
	// qu'un plafond de tours dans la boucle, parce que le fichier est faux et
	// qu'un plafond le laisserait tourner en le masquant.
	//
	// Les bornes ne se prononcent que sur un champ présent : une absence est déjà
	// signalée, et la compter deux fois donnerait à l'auteur une liste où le
	// nombre de lignes ne correspond plus au nombre de choses à corriger.
	if n.First != nil && table.FirstThreshold < 1 {
		dire("niveaux.seuil_premier : %d, il en faut au moins une", table.FirstThreshold)
	}
	if n.Increment != nil && table.Increment < 0 {
		dire("niveaux.seuil_increment : %d, un seuil ne redescend pas", table.Increment)
	}

	if ms := exige("niveaux", "plancher_ms", n.FloorMs, dire); ms > 0 {
		ticks, err := TicksFromMs(ms)
		if err != nil {
			dire("niveaux.plancher_ms : %v", err)
		}
		table.Floor = ticks
	}

	g := brut.Progression.Gems
	if g.Object == "" {
		dire("gemmes.objet : absent ou vide")
	}
	table.GemValue = exige("gemmes", "experience", g.Experience, dire)
	if g.Experience != nil && table.GemValue < 1 {
		dire("gemmes.experience : %d, une gemme qui ne rapporte rien est une "+
			"gemme que le joueur ramasse pour rien", table.GemValue)
	}

	if len(manques) > 0 {
		return nil, &manifest.Invalid{Path: chemin, Missing: manques}
	}
	return table, nil
}

// rawProgression est le fichier tel qu'il s'écrit.
type rawProgression struct {
	manifest.Commentable
	// Format est la version du format de manifeste.
	Format int `json:"version_format"`
	// Progression range les sections, dont il n'existe qu'une aujourd'hui.
	Progression rawSections `json:"progression"`
}

// rawSections groupe ce qui commande le déroulé d'une run.
//
// Une section et non les champs à plat : la courbe de pression du spawner vient
// s'y ranger, et elle appartient à la partie au même titre que les seuils. Les
// mettre à plat aurait mêlé deux réglages sans rapport dans un même espace de
// noms, où `plancher_ms` aurait fini par désigner deux choses.
type rawSections struct {
	manifest.Commentable
	// Levels est le rythme des montées de niveau.
	Levels rawLevels `json:"niveaux"`
	// Gems porte ce que rapporte une gemme.
	Gems rawGems `json:"gemmes"`
}

// rawGems déclare la valeur d'une gemme et l'objet qu'elle décrit.
type rawGems struct {
	manifest.Commentable

	// Object nomme la gemme dans le manifeste d'objets. Ce champ n'est pas lu
	// ici : c'est `outils/ressources.py` qui exige qu'il désigne un objet
	// existant. Sans ce lien, renommer la gemme laisserait ce réglage orphelin,
	// et c'est exactement ainsi qu'une valeur déclarée sans lecteur avait pu
	// s'installer dans le manifeste d'objets.
	Object string `json:"objet"`
	// Experience est ce que son ramassage porte au compteur.
	Experience *int `json:"experience"`
}

// rawLevels porte les seuils, en pointeurs pour les valeurs dont zéro est une
// réponse plausible.
//
// Un incrément nul est un réglage légitime — un seuil constant —, si bien que
// rien ne distinguerait le champ absent du champ à zéro.
type rawLevels struct {
	manifest.Commentable

	First     *int `json:"seuil_premier"`
	Increment *int `json:"seuil_increment"`
	FloorMs   *int `json:"plancher_ms"`
}

// Level rend le niveau atteint. La partie commence au premier.
func (w *World) Level() int { return w.niveau }

// Experience rend les gemmes acquises vers le niveau suivant.
func (w *World) Experience() int { return w.experience }

// Threshold rend ce que le niveau suivant coûte, en gemmes.
func (w *World) Threshold() int { return w.progression.Threshold(w.niveau) }

// progresser encaisse la récolte du tick et distribue les niveaux qu'elle ouvre.
//
// **Ce qu'une gemme vaut vient du manifeste et n'est pas son unité.** Compter
// les gemmes donnerait aujourd'hui le même résultat, la valeur publiée étant
// justement un — et c'est ce qui rend la confusion facile : les deux
// implémentations sont indistinguables tant qu'on ne change pas ce chiffre.
//
// **Le plancher de temps se compte en ticks et jamais en millisecondes réelles.**
// La garantie de la conception — jamais plus de quarante-cinq secondes sans un
// choix — porte sur le temps de la partie, pas sur celui du joueur : un menu
// ouvert ne doit pas la faire avancer. Compter les ticks le donne sans rien
// écrire, puisque ce compteur ne monte que dans un tick simulé.
//
// **Le compteur repart à chaque montée, quelle qu'en soit la source.** C'est ce
// que « jamais plus de quarante-cinq secondes sans un choix » veut dire : une
// garantie sur l'intervalle entre deux choix. Sur un calendrier absolu, on
// obtiendrait un métronome capable de donner un niveau forcé trois secondes
// après un niveau gagné.
//
// **Il ne remet pas l'expérience à zéro pour autant**, et c'est un autre
// compteur : les gemmes déjà ramassées comptent pour le niveau suivant. Un
// joueur puni par sa lenteur perdrait sinon ce qu'il avait récolté, et le
// plancher cesserait d'être purement additif.
//
// L'ordre des deux sources n'est pas indifférent : la récolte d'abord, le
// plancher ensuite. Un niveau gagné à ce tick vient de remettre le compteur à
// zéro, si bien que les deux ne peuvent pas se déclencher ensemble.
func (w *World) progresser(recoltees int) {
	w.experience += recoltees * w.progression.GemValue
	w.depuisChoix++

	for w.experience >= w.progression.Threshold(w.niveau) {
		w.experience -= w.progression.Threshold(w.niveau)
		w.monter()
	}
	if w.depuisChoix >= w.progression.Floor {
		w.monter()
	}
}

// monter passe au niveau suivant et ouvre le choix qui l'accompagne.
//
// Chemin unique : la montée par expérience et la montée forcée l'empruntent
// toutes deux jusqu'au choix présenté. Deux chemins parallèles resteraient
// d'accord tant que personne n'y touche, et cesseraient de l'être au premier
// réglage — c'est ici que les trois cartes se branchent, et elles ne connaissent
// pas la source de la montée.
//
// **Le niveau monte avant le choix, pas après.** C'est ce qui fait terminer la
// boucle de distribution : le seuil du niveau suivant est plus haut, donc
// l'expérience finit par ne plus l'atteindre. Un niveau qui n'avancerait qu'au
// choix laisserait cette boucle tourner sur un seuil qui ne bouge pas.
//
// Un choix déjà ouvert n'est pas remplacé, il fait la queue. Une récolte
// abondante donne deux montées dans le même tick, et écraser la première en
// retirerait une au joueur sans que rien ne le dise.
func (w *World) monter() {
	w.niveau++
	w.depuisChoix = 0

	if w.Choosing() {
		w.enAttente++
		return
	}
	w.offrir()
}
