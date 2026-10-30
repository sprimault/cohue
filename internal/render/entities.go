// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La pose des entités dans la scène : ce qu'une créature et un objet y
// décrivent, et le liseré qui détache le joueur de la foule.

package render

import (
	"github.com/sprimault/lozengine/render"
	"github.com/sprimault/lozengine/scene"

	"github.com/sprimault/cohue/internal/game"
	"github.com/sprimault/cohue/internal/sprite"
)

// poserEntites décrit dans la scène tout ce que les bassins portent, puis repose
// le rappel de ce qui est caché.
//
// **Le rappel vient après la pose et lit le relevé de l'image précédente.** Un
// recouvrement ne se connaît qu'une fois l'image triée, donc le plus récent est
// celui d'avant ; à soixante images par seconde l'écart ne se voit pas, et
// aucune décision de jeu n'en dépend.
//
// **Le solde vient en dernier et porte sur tous les bassins.** Une entité
// disparue ne se signale pas : elle cesse simplement d'être marquée, et c'est en
// parcourant les attaches qu'on la retire de la scène.
func (s *Screen) poserEntites() {
	s.aRevoir = s.aRevoir[:0]

	poserBassin(s, sorteEnnemi, s.monde.Enemies(),
		// **Les deux retours de combat se posent ici et non dans `creature`.** Ils
		// n'appartiennent pas au dessin d'un personnage — un figurant n'en a
		// jamais — mais à l'état d'une créature, que seule cette closure connaît.
		func(c *game.Enemy, id int) (scene.Entity, trace, bool) {
			f := s.troupe.ennemis[c.Profile]
			e, t, montre := s.creature(f, f.poseEnnemi(c, s.regardDe(c)), c.X, c.Y, id)
			if montre {
				s.poserEclair(e, id, etatDe(c))
				if c.Flash > 0 {
					s.poserEtincelle(t, c.Flash)
				}
			}
			return e, t, montre
		})
	poserBassin(s, sorteAmbiance, s.monde.Ambients(),
		func(a *game.Ambient, id int) (scene.Entity, trace, bool) {
			f := s.troupe.ambiants[a.Profile]
			return s.creature(f, f.posePersonnage(a.Step, a.Step, a.Variant), a.X, a.Y, id)
		})
	poserBassin(s, sorteTir, s.monde.Shots(),
		func(p *game.Projectile, id int) (scene.Entity, trace, bool) {
			s.poserTrainee(p.X, p.Y, p.Step, id)
			return s.objet(objetTir, p.X, p.Y, id, nil)
		})
	poserBassin(s, sorteTirHorde, s.monde.EnemyShots(),
		func(p *game.Projectile, id int) (scene.Entity, trace, bool) {
			return s.objet(objetTirHorde, p.X, p.Y, id, nil)
		})
	poserBassin(s, sorteGemme, s.monde.Gems(),
		func(g *game.Gem, id int) (scene.Entity, trace, bool) {
			return s.objet(objetGemme, g.X, g.Y, id, s.voileDeLaGemme(g))
		})
	poserBassin(s, sorteAimant, s.monde.Magnets(),
		func(a *game.Magnet, id int) (scene.Entity, trace, bool) {
			return s.objet(objetAimant, a.X, a.Y, id, nil)
		})
	poserBassin(s, sorteEpave, s.monde.Wrecks(), s.epave)
	poserBassin(s, sorteObstacle, s.monde.Breakables(), s.obstacle)
	poserBassin(s, sorteCaisse, s.monde.Crates(),
		// **L'annonce reste pendant l'appui**, et ce n'est pas indifférent : elle
		// disparaîtrait au moment précis où le joueur décide de rester ou de
		// repartir, c'est-à-dire là où elle sert le plus.
		func(c *game.Crate, id int) (scene.Entity, trace, bool) {
			e, t, montre := s.caisse(c, id)
			if montre {
				s.poserAnnonce(c, t)
			}
			return e, t, montre
		})
	poserBassin(s, sorteArmeAuSol, s.monde.Drops(),
		func(d *game.Drop, id int) (scene.Entity, trace, bool) {
			return s.objet(s.monde.DropWeapon(d).Key, d.X, d.Y, id, nil)
		})
	poserBassin(s, sorteFioleAuSol, s.monde.Vials(),
		func(f *game.Vial, id int) (scene.Entity, trace, bool) {
			return s.objet(objetFiole, f.X, f.Y, id, nil)
		})
	poserBassin(s, sorteTourelle, s.monde.Turrets(),
		// Le même dessin que l'arme au sol : ce qu'on pose est l'objet qu'on
		// tenait, et le distinguer demanderait un second dessin pour dire ce que
		// la position dit déjà — une tourelle posée est là où le joueur n'est plus.
		func(t *game.Turret, id int) (scene.Entity, trace, bool) {
			return s.objet(s.monde.TurretWeapon(t).Key, t.X, t.Y, id, nil)
		})
	s.poserJoueur()

	s.reveler()
	for _, a := range s.attaches {
		a.solder(s.scene, s.monde.Tick())
	}
}

// poserJoueur pose le personnage, qui n'est dans aucun bassin, et le liseré qui
// le cerne.
//
// **Le sprite monte d'un rang pour que son liseré tienne dessous.** Une strate
// ne descend pas sous zéro, donc c'est le dessin qui se range au-dessus de ce
// qu'il doit recouvrir, et non l'inverse. Entre deux entités, le rang ne
// départage que ce que la profondeur et l'abscisse laissent à égalité — soit
// deux choses exactement superposées —, si bien que ce décalage ne change rien
// à la place du joueur dans la horde.
func (s *Screen) poserJoueur() {
	x, y := s.monde.Player()
	f := s.troupe.joueur
	pas := s.monde.PlayerStep()
	e, t, montre := s.creature(f, f.posePersonnage(pas, s.regardDuJoueur(pas), 0), x, y, 0)
	if !montre {
		return
	}

	e.Measure = true
	e.Tier = rangDuJoueur
	poignee := s.attaches[sorteJoueur].poser(s.scene, 0, e, s.monde.Tick())
	s.retenir(poignee, t, sorteJoueur)
	s.poserLisere(e, t)
}

// poserLisere cerne le joueur d'un trait, sous son sprite.
//
// **Le chapitre 2 le veut permanent**, et il l'est : c'est ce qui détache le
// personnage d'un décor dont il partage les gris, là où le rappel de silhouette
// ne répond qu'à ce qui le cache. Les deux emploient le même aplat et ne se
// confondent pas — celui-ci est sous le sprite et tout le temps, l'autre
// par-dessus et seulement caché.
//
// **Il était inexprimable jusqu'à la 0.5.4.** Un aplat n'est jamais la planche
// du dessin qu'il double, et la clé de tri rangeait la planche avant le rang de
// strate : les huit quads passaient par-dessus le sprite quoi qu'on demande. Le
// moteur ayant retiré la planche de la clé, le rang décide, et le liseré tient
// enfin sous le personnage.
//
// **Aucun des huit n'entre dans la mesure de recouvrement.** Ils cernent le
// joueur sans rien cacher de ce qui le cacherait, et les compter ferait du
// personnage sa propre occultation — c'est le défaut que le rappel vient de
// refermer, et il se rouvrirait ici.
func (s *Screen) poserLisere(e scene.Entity, t trace) {
	if t.forme.Width == 0 {
		return
	}

	for i, d := range contourAutour {
		q := e
		q.Aspect = t.forme
		q.Tint = teinteContour
		q.AnchorX += d[0]
		q.AnchorY += d[1]
		q.Uncounted = true
		q.Measure = false
		q.Tier = rangDuLisere
		s.attaches[sorteLisere].poser(s.scene, i, q, s.monde.Tick())
	}
}

// retenir inscrit une chose posée parmi celles dont on repose le rappel.
//
// Une chose dont le profil ne porte pas d'aplat n'est pas inscrite : il n'y
// aurait rien à reposer, et la marque de mesure qu'elle a reçue ne coûte qu'un
// relevé que personne ne lit — ce qui n'arrive pas, le joueur et le projectile
// de la horde étant les deux seuls dessins dérivés du catalogue.
func (s *Screen) retenir(poignee scene.Handle, t trace, quoi sorte) {
	if poignee == 0 || t.forme.Width == 0 {
		return
	}
	s.aRevoir = append(s.aRevoir, revele{
		poignee: poignee,
		trace:   t,
		teinte:  teinteDuRappel(quoi),
	})
}

// voileDeLaGemme rend l'extinction d'une gemme, nulle quand l'aimant la tient.
//
// **Une gemme attirée reprend sa teinte pleine.** L'extinction dit « ceci va
// disparaître » ; une gemme que l'aimant tient ne disparaîtra pas, et la montrer
// éteinte serait montrer une information fausse. Accessoirement, une ruée de
// gemmes anciennes serait un feu d'artifice en gris — l'inverse de ce que la
// conception appelle le moment de plaisir maximal du genre.
//
// **Sur un dessin, l'extinction multiplie**, et c'est ici le bon sens : ce
// qu'elle dit est que la gemme s'efface, donc qu'elle perd de sa lumière. C'est
// l'inverse d'un éclair d'état, qui ajoute.
func (s *Screen) voileDeLaGemme(g *game.Gem) *render.Color {
	if g.Pulled {
		return nil
	}
	eteinte := eteindre(render.White(), s.monde.GemAge(g), s.monde.GemLife())
	return &eteinte
}

// creature décrit ce qu'une créature montre, et ce qu'elle laisse pour son
// rappel.
//
// **Elle décrit, elle ne pose pas.** Ce qui pose est `attache`, qui sait si
// l'entité est déjà dans la scène ; séparer les deux évite que chaque sorte
// redise la correspondance entre un identifiant de bassin et une poignée.
//
// Le troisième retour est faux quand le profil ne porte pas la pose demandée :
// le Molosse n'a ni repos ni attaque, et `Sheet.Frame` l'annonce par son second
// retour. L'entité n'est alors pas marquée, donc elle quitte la scène au solde.
func (s *Screen) creature(f *figure, a anim, x, y game.Fixed,
	identite int) (scene.Entity, trace, bool) {
	if a.nom == "" || a.direction == "" {
		return scene.Entity{}, trace{}, false
	}

	i := sprite.Loop(a.cycle, s.monde.Tick(), identite)
	if !a.cycle.Loop {
		i = sprite.Once(a.cycle, a.reste)
	}
	p := pose{a.nom, a.direction, a.variante, i}
	aspect, porte := f.image(p)
	if !porte {
		return scene.Entity{}, trace{}, false
	}
	aspect.Layer = coucheMonde

	e, t := s.decrire(aspect, x, y)
	t.forme, _ = f.forme(p)
	t.bord, _ = f.bord(p)
	return e, t, true
}

// objet décrit ce qu'un objet du catalogue montre, cadencé par le tick.
func (s *Screen) objet(nom string, x, y game.Fixed, identite int,
	voile *render.Color) (scene.Entity, trace, bool) {
	catalogue, i, montre := s.objets.image(nom, s.monde.Tick(), identite)
	return s.poseDuCatalogue(catalogue, i, montre, x, y, voile)
}

// effet décrit une animation brève, cadencée par le décompte de l'état qui la
// porte.
//
// C'est la même dérivation qu'un cycle d'attaque : l'animation s'achève quand
// l'état s'achève, et un état plus court qu'elle la fait entrer en cours de
// route.
func (s *Screen) effet(nom string, x, y game.Fixed,
	reste game.Tick) (scene.Entity, trace, bool) {
	catalogue, i, montre := s.objets.effet(nom, reste)
	return s.poseDuCatalogue(catalogue, i, montre, x, y, nil)
}

// poseDuCatalogue rend l'entité et la trace d'une image du catalogue.
//
// Les deux cadences — le tick et le décompte — ne diffèrent que par le rang
// qu'elles désignent, et tout le reste est commun : la pose, la teinte, ce que
// l'objet laisse pour son rappel. Les séparer entièrement aurait dupliqué ce
// reste.
func (s *Screen) poseDuCatalogue(catalogue prop, i int, montre bool, x, y game.Fixed,
	voile *render.Color) (scene.Entity, trace, bool) {
	if !montre {
		return scene.Entity{}, trace{}, false
	}
	aspect, porte := catalogue.image(i)
	if !porte {
		return scene.Entity{}, trace{}, false
	}

	aspect.Layer = coucheMonde
	if voile != nil {
		aspect.Tint = *voile
	}

	e, t := s.decrire(aspect, x, y)
	t.forme, _ = catalogue.forme(i)
	return e, t, true
}

// epave décrit ce qu'un objet cédé a laissé : sa rupture tant qu'elle se
// déroule, sa ruine au sol ensuite. Un objet sans cycle de rupture — les
// obstacles fragiles — passe directement à sa ruine, ses éclats disant la
// rupture à sa place.
//
// **L'avancement se prend sur l'âge et non sur un décompte**, à l'inverse de
// toutes les autres animations de ce paquet. La règle générale ancre sur la fin
// parce que le rendu n'a que le reste d'un état ; ici il n'y a pas d'état qui
// finisse — l'épave demeure —, et la durée de la bande est connue puisque le
// cycle est chargé. Le reste s'en déduit, et `Once` retrouve son régime
// habituel.
func (s *Screen) epave(e *game.Wreck, id int) (scene.Entity, trace, bool) {
	d := s.objets.cassable(e.Object)
	if age, duree := s.monde.WreckAge(e), s.objets.duree(d.BreakCycle); age < duree {
		return s.effet(d.BreakCycle, e.X, e.Y, duree-age)
	}

	// **Une épave gît, elle ne se dresse pas**, donc elle descend d'un étage et
	// passe sous tout ce qui marche dessus. Sans cela elle se départagerait par
	// la profondeur avec une créature debout sur elle, et le tirage d'une
	// égalité la ferait passer devant une image sur deux.
	ent, t, montre := s.objet(s.objets.sens(d.Ruin, e.Across), e.X, e.Y, id, nil)
	ent.Floor = etageTraine
	return ent, t, montre
}

// obstacle décrit un obstacle fragile tant qu'il tient.
func (s *Screen) obstacle(o *game.Breakable, id int) (scene.Entity, trace, bool) {
	nom := s.objets.sens(s.monde.BreakableKey(o.Kind), o.Across)
	return s.objet(nom, o.X, o.Y, id, nil)
}

// caisse décrit une caisse, écrasée sous l'appui ou intacte.
//
// **La déformation dit qu'on est en train de casser**, et c'est tout ce qui
// sépare un délai d'un blocage : sans elle, un joueur qui pousse contre une
// caisse ralentie croit avoir buté sur un mur.
//
// **Elle se dérive du décompte d'appui et non du tick**, ce qui est le régime de
// toute animation qui s'achève : l'écrasement se complète à l'instant où la
// caisse cède. Cadencée sur le tick, sa phase n'aurait aucun rapport avec la
// poussée — on toucherait une caisse déjà écrasée, qui se redresserait ensuite.
func (s *Screen) caisse(c *game.Crate, id int) (scene.Entity, trace, bool) {
	if c.Press < s.monde.CratePress() {
		return s.effet(s.objets.caisse.PressCycle, c.X, c.Y, c.Press)
	}
	return s.objet(objetCaisse, c.X, c.Y, id, nil)
}

// decrire rend l'entité d'un aspect posé à une position du monde, et le pixel où
// elle tombe.
//
// **La hauteur du sol passe par `Entity.Z`**, et c'est ce que le portage a
// déplacé : elle était retranchée de l'ordonnée au moment du dessin, ce qui
// mêlait une hauteur à une profondeur. Le moteur la range dans son propre champ
// de la clé de tri, exactement là où un corps surélevé doit se départager — un
// personnage sur un trottoir passe devant ce qui est au pied de la bordure, et
// non derrière.
//
// **Le passage d'un niveau à l'autre saute, il ne s'interpole pas.** Une montée
// progressive ferait gravir une pente invisible là où le dessin montre une
// marche franche ; six pixels d'un coup sont ce qu'un pas sur un trottoir fait.
//
// **Le pixel se projette à la main, et il n'y a pas d'autre voie.** Le rappel se
// pose en couche haute pour passer devant ce qui le cache, donc par un quad
// d'écran ; la scène n'a pas à dire où elle a posé une entité, puisque c'est
// elle qui le décide.
func (s *Screen) decrire(aspect scene.Aspect, x, y game.Fixed) (scene.Entity, trace) {
	ex, ey := s.pixelDe(x, y)

	// **Une entité se dresse sur le sol, donc elle prend l'étage de ce qui s'y
	// dresse.** L'étage passant avant la profondeur dans la clé de tri, la
	// laisser à zéro la rangeait sous tout le décor occultant quelle que soit sa
	// place : un pilier devant elle la recouvrait. Voir `etageDebout`, qui porte
	// le motif.
	aspect.Floor = etageDebout

	return scene.Entity{
		X: x.Float(), Y: y.Float(),
		Z:      float64(s.sol.hauteurSolEn(x, y)),
		Aspect: aspect,
	}, trace{ex: ex, ey: ey}
}

// pixelDe rend le pixel de l'aire de jeu où un point du monde tombe, hauteur du
// sol comprise.
//
// **Elle sert ce que la scène ne place pas**, et c'est son seul usage hors de la
// description d'une entité : un chiffre de dégâts, l'annonce d'une caisse, le
// rappel d'un personnage caché se posent au-dessus d'une position mais en couche
// haute, donc par une pièce — qui se compte en pixels et jamais en cellules.
//
// **La vue se retranche parce qu'une pièce ne la connaît pas.** La projection
// rend une position dans la carte entière ; la pièce se pose dans l'aire, et le
// moteur n'y applique aucun déplacement de caméra. C'est le même calcul que la
// scène fait pour ses propres quads, et c'est exactement le repère que
// `Scene.PlaceArea` et `Font.DrawOutlinedArea` attendent.
func (s *Screen) pixelDe(x, y game.Fixed) (int32, int32) {
	z := float64(s.sol.hauteurSolEn(x, y))
	ex, ey, _, _ := s.scene.Projection().Project(x.Float(), y.Float(), z)
	vx, vy, _, _ := s.scene.View()
	return ex - vx, ey - vy
}

// poserAuPoint pose un aspect en couche haute, son pied centré sur le pixel
// donné, comme le moteur poserait une entité sur un point du monde.
//
// **Les deux voies de pose n'ont pas la même origine, et rien ne le dit.** Une
// entité reçoit un point du sol, et le moteur y centre le pied de son image ;
// un quad d'écran reçoit le **coin haut-gauche**. Passer le pixel projeté tel
// quel décale donc d'une demi-largeur vers la droite et d'une hauteur entière
// vers le bas tout ce qui se pose au-dessus d'une position — la silhouette d'un
// personnage caché, un chiffre de dégâts, une gerbe d'étincelles.
//
// **L'ancre ne se refait pas ici, et c'est le piège de cette conversion.** La
// pose l'ajoute d'elle-même au coin qu'on lui donne, sa godoc le dit en toutes
// lettres : « les ajouter au calcul les compterait deux fois, et l'image
// partirait du double de l'ancre ». Ce qui reste à faire est la seule chose
// qu'elle ne fait pas — passer du pied au coin, soit une demi-largeur et une
// hauteur.
//
// **La pose vise l'aire et non la surface, et c'est ce qui ferme le reste.**
// `pixelDe` rend un pixel relatif à l'**aire de monde** ; `PlaceArea` attend
// exactement ce repère et découpe la pièce sur l'aire. Par `PlaceScreen` il
// fallait y ajouter l'origine de l'aire — oubli qui posait le chiffre de dégâts
// loin de sa créature —, puis refuser ce qui n'y touchait pas, et un quad à
// cheval sur le bord débordait quand même dans la marge de sa propre taille.
// Le découpage exact demandait de refaire celui du moteur, miroirs compris ;
// c'est lui qui le fait.
func (s *Screen) poserAuPoint(ex, ey int32, a scene.Aspect) {
	s.scene.PlaceArea(ex-int32(a.Width)/2, ey-int32(a.Height), a)
}

// teinteDuRappel rend la teinte sous laquelle une sorte se rappelle.
func teinteDuRappel(quoi sorte) render.Color {
	if quoi == sorteTirHorde {
		return teinteTirHorde
	}
	return teinteContour
}
