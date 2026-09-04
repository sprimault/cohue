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
	// PickupRange est la distance à laquelle une gemme se ramasse, en tuiles.
	//
	// Elle vit avec la durée de vie et non sur le profil du joueur, parce que la
	// conception en fait un couple qu'on ne règle pas séparément : chacune punit
	// le non-ramassage, et le lecteur qui touche à l'une doit voir l'autre.
	PickupRange Fixed
	// GemLife est le temps qu'une gemme reste au sol, en ticks.
	//
	// **C'est la contre-force de l'aimant.** Sans elle, la valeur d'un
	// déclenchement croît strictement avec l'attente : attendre est toujours
	// rationnel, et le joueur meurt avec sa charge. Elle travaille au-delà de
	// l'aimant, d'ailleurs — ramasser oblige à revenir là où l'on vient de tuer,
	// donc là où la horde converge, et le kiting en cercle cesse d'être gratuit
	// sans qu'on ait rien interdit.
	GemLife Tick
	// MagnetPeriod est l'écart entre deux apparitions d'aimant, en ticks.
	MagnetPeriod Tick
	// MagnetMinRange est la distance au joueur sous laquelle une apparition est
	// refusée, en tuiles.
	//
	// **La contrainte est d'être loin, pas d'être hors champ** — l'inverse de
	// l'anneau qui pose les créatures. Une créature qui surgit devant est une
	// injustice ; un objet qu'on voit au loin est une invitation, et le trajet
	// pour l'atteindre est ce qui en fait une décision.
	MagnetMinRange Fixed
	// PullSpeed est la vitesse d'une gemme attirée, en tuiles par tick.
	PullSpeed Fixed
	// GemValue est ce qu'une gemme ramassée porte à l'expérience.
	//
	// Un scalaire et non une table, parce que la simulation ne distingue pas
	// deux sortes de gemmes : `Gem` ne porte pas de profil, et une table serait
	// un réglage que personne ne lirait — c'est exactement ce qui s'était
	// installé quand cette valeur vivait dans le manifeste d'objets sans que rien
	// n'aille l'y chercher. Le jour où une grosse gemme existera, c'est `Gem` qui
	// portera sa sorte, et ce champ deviendra une table indexée comme les profils.
	GemValue int
	// SpawnRadius est la distance au joueur à laquelle une créature apparaît, en
	// tuiles.
	//
	// **Une donnée et non une dérivée de la fenêtre**, et la contrainte est plus
	// forte qu'un cloisonnement de paquets : un rayon tiré de l'écran donnerait à
	// deux joueurs aux fenêtres différentes des apparitions différentes sur la
	// même graine, ce que le déterminisme de la run interdit. Le chiffre publié
	// est dérivé du tampon fixe et le manifeste dit lequel.
	SpawnRadius Fixed
	// CarryOver borne le budget de pression reporté d'un tick au suivant, en
	// ticks de budget.
	//
	// Une apparition abandonnée faute de place reporte son budget, sans quoi un
	// couloir étroit serait un abri où la pression tombe. Bornée, sans quoi le
	// temps passé à se terrer se libère d'un coup à la sortie — le mur d'ennemis
	// que la règle « jamais dans le champ de vision » existe pour interdire.
	CarryOver Tick
	// CrateGems est le nombre de gemmes qu'une caisse laisse en se cassant.
	CrateGems int
	// CrateRange est la distance à laquelle le joueur casse une caisse, en
	// tuiles.
	CrateRange Fixed
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
	} else if n.FloorMs != nil {
		// Un plancher nul n'est pas l'absence de plancher mais son contraire :
		// le compteur l'atteint à chaque tick, donc le joueur monte d'un niveau
		// et voit trois cartes soixante fois par seconde. Le champ absent est
		// déjà signalé par `exige`, et le dire deux fois ferait corriger deux
		// lignes pour une faute.
		dire("niveaux.plancher_ms : %d, un niveau à chaque tick n'est plus un choix", ms)
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

	table.PickupRange = FromFloat(exige("gemmes", "portee_ramassage_tuiles", g.TileRange, dire))
	if g.TileRange != nil && table.PickupRange < 1 {
		dire("gemmes.portee_ramassage_tuiles : %v, une portée que la virgule "+
			"fixe arrondit a zero ne ramasse rien", *g.TileRange)
	}

	if ms := exige("gemmes", "duree_vie_ms", g.LifeMs, dire); ms > 0 {
		ticks, err := TicksFromMs(ms)
		if err != nil {
			dire("gemmes.duree_vie_ms : %v", err)
		}
		table.GemLife = ticks
	} else if g.LifeMs != nil {
		// Une durée nulle efface la gemme dans le tick où elle tombe : le joueur
		// verrait la créature mourir sans rien laisser, ce qui ressemble à un
		// butin manquant et non à un effacement.
		dire("gemmes.duree_vie_ms : %d, une gemme qui ne dure pas n'existe pas", ms)
	}

	a := brut.Progression.Magnet
	if a.Object == "" {
		dire("aimant.objet : absent ou vide")
	}
	if ms := exige("aimant", "periode_ms", a.PeriodMs, dire); ms > 0 {
		ticks, err := TicksFromMs(ms)
		if err != nil {
			dire("aimant.periode_ms : %v", err)
		}
		table.MagnetPeriod = ticks
	} else if a.PeriodMs != nil {
		// Une période nulle poserait un aimant à chaque tick : le sol en serait
		// couvert, et l'objet cesserait d'être l'événement que la conception en
		// fait.
		dire("aimant.periode_ms : %d, un aimant à chaque tick n'est plus un événement", ms)
	}

	table.MagnetMinRange = FromFloat(exige("aimant", "distance_min_tuiles", a.MinTiles, dire))
	if a.MinTiles != nil && table.MagnetMinRange < One {
		// Sous une tuile, l'aimant tombe dans la portée de ramassage et se prend
		// sans qu'on ait bougé : il cesse d'être un trajet, donc une décision.
		dire("aimant.distance_min_tuiles : %v, un aimant qu'on ramasse sans "+
			"bouger n'est pas une decision", *a.MinTiles)
	}

	table.PullSpeed = parTick(exige("aimant", "vitesse_gemme_tuiles_s", a.GemSpeed, dire))
	if a.GemSpeed != nil && table.PullSpeed < 1 {
		dire("aimant.vitesse_gemme_tuiles_s : %v, une gemme attirée qui n'avance "+
			"pas ne rejoint jamais le joueur", *a.GemSpeed)
	}

	s := brut.Progression.Pressure
	table.SpawnRadius = FromFloat(exige("pression", "rayon_apparition_tuiles", s.Radius, dire))
	if s.Radius != nil && table.SpawnRadius < One {
		// Sous une tuile, la créature naît sur le joueur. La borne ne prétend pas
		// vérifier la règle vraie — hors du champ de vision —, que ce paquet ne
		// peut pas connaître : elle attrape le zéro et l'oubli de virgule.
		dire("pression.rayon_apparition_tuiles : %v, une créature apparaîtrait sur "+
			"le joueur", *s.Radius)
	}

	// Zéro passe sans rien dire, à l'inverse des autres durées de ce fichier : un
	// report nul rend la pression strictement instantanée, ce qui est un réglage
	// et non une saisie manquée. C'est aussi ce qu'on écrira pour retrouver le
	// comportement d'avant le report, le jour où l'on voudra le comparer.
	if ms := exige("pression", "report_ms", s.CarryMs, dire); ms > 0 {
		ticks, err := TicksFromMs(ms)
		if err != nil {
			dire("pression.report_ms : %v", err)
		}
		table.CarryOver = ticks
	}

	c := brut.Progression.Crates
	if c.Object == "" {
		dire("caisses.objet : absent ou vide")
	}
	table.CrateGems = exige("caisses", "gemmes", c.Gems, dire)
	if c.Gems != nil && table.CrateGems < 1 {
		// Une caisse vide est un objet qu'on casse pour rien : elle coûte le
		// détour qui l'a atteinte, et le joueur cesse d'en chercher après deux.
		dire("caisses.gemmes : %d, une caisse qui ne laisse rien ne vaut pas le "+
			"detour qu'elle coute", table.CrateGems)
	}

	table.CrateRange = FromFloat(exige("caisses", "portee_contact_tuiles", c.TileRange, dire))
	if c.TileRange != nil && table.CrateRange < 1 {
		dire("caisses.portee_contact_tuiles : %v, une portée que la virgule fixe "+
			"arrondit a zero ne casse rien", *c.TileRange)
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
// Des sections et non les champs à plat : les mettre à plat aurait mêlé des
// réglages sans rapport dans un même espace de noms, où `plancher_ms` aurait
// fini par désigner deux choses.
type rawSections struct {
	manifest.Commentable
	// Levels est le rythme des montées de niveau.
	Levels rawLevels `json:"niveaux"`
	// Gems porte ce que rapporte une gemme.
	Gems rawGems `json:"gemmes"`
	// Magnet porte le rythme de l'aimant et ce qu'il fait.
	Magnet rawMagnet `json:"aimant"`
	// Pressure porte ce que le spawner tient de la partie, par opposition à ce
	// que le lieu lui dit. Le partage est net : un auteur écrit le rythme de ses
	// vagues, il ne décide pas d'où sortent les créatures.
	Pressure rawPressure `json:"pression"`
	// Crates porte ce qu'une caisse laisse et à quelle distance elle se casse.
	//
	// **Le contenu est un réglage de jeu et non de lieu**, pour la raison qui
	// vaut déjà pour la valeur d'une gemme : ce que rapporte une chose doit
	// signifier la même partout, sans quoi un auteur règle sa difficulté en
	// gonflant ses caisses. Le lieu dit où elles sont, il ne dit pas ce qu'elles
	// donnent.
	Crates rawCrates `json:"caisses"`
}

// rawCrates déclare ce qu'une caisse laisse et comment elle se casse.
type rawCrates struct {
	manifest.Commentable

	// Object nomme la caisse dans le manifeste d'objets. Ce champ n'est pas lu
	// ici, pour la raison écrite sur `rawGems.Object` : c'est le contrôle des
	// ressources qui exige qu'il désigne un objet existant.
	Object string `json:"objet"`
	// Gems est le nombre de gemmes qu'elle laisse.
	Gems *int `json:"gemmes"`
	// TileRange est la distance à laquelle le joueur la casse, en tuiles.
	TileRange *float64 `json:"portee_contact_tuiles"`
}

// rawPressure déclare ce que le spawner tient de la partie.
type rawPressure struct {
	manifest.Commentable

	// Radius est la distance d'apparition, en tuiles.
	Radius *float64 `json:"rayon_apparition_tuiles"`
	// CarryMs borne le budget reporté d'un tick au suivant.
	CarryMs *int `json:"report_ms"`
}

// rawMagnet déclare l'apparition de l'aimant et la ruée qu'il déclenche.
type rawMagnet struct {
	manifest.Commentable

	// Object nomme l'aimant dans le manifeste d'objets. Ce champ n'est pas lu
	// ici, pour la raison écrite sur `rawGems.Object` : c'est le contrôle des
	// ressources qui exige qu'il désigne un objet existant.
	Object string `json:"objet"`
	// PeriodMs est l'écart entre deux apparitions.
	PeriodMs *int `json:"periode_ms"`
	// MinTiles est la distance au joueur sous laquelle l'apparition est refusée.
	MinTiles *float64 `json:"distance_min_tuiles"`
	// GemSpeed est la vitesse d'une gemme attirée, en tuiles par seconde.
	GemSpeed *float64 `json:"vitesse_gemme_tuiles_s"`
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
	// TileRange est la portée de ramassage, en tuiles.
	TileRange *float64 `json:"portee_ramassage_tuiles"`
	// LifeMs est le temps qu'une gemme reste au sol.
	LifeMs *int `json:"duree_vie_ms"`
}

// rawLevels porte les seuils, en pointeurs pour les valeurs dont zéro est une
// réponse plausible.
//
// Un incrément nul est un réglage légitime — un seuil constant —, si bien que
// rien ne distinguerait le champ absent du champ à zéro.
type rawLevels struct {
	manifest.Commentable

	// First est le nombre de gemmes qui mène du premier niveau au second.
	First *int `json:"seuil_premier"`
	// Increment est ce que chaque niveau ajoute au seuil du suivant. Zéro y est
	// un seuil constant, c'est-à-dire un réglage.
	Increment *int `json:"seuil_increment"`
	// FloorMs est le temps au bout duquel un niveau est donné sans rien
	// ramasser, converti en ticks au chargement. Zéro n'y est pas l'absence de
	// plancher mais son contraire — un niveau à chaque tick —, et le chargeur le
	// refuse.
	FloorMs *int `json:"plancher_ms"`
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
