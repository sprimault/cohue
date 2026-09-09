# Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
# SPDX-License-Identifier: Apache-2.0

"""Fabrique les objets du jeu : projectiles, ramassables, icônes d'armes.

    python outils/objets.py --sortie assets/objets
    python outils/objets.py --apercu

Même principe que le décor et les personnages : des volumes isométriques
composés, aucune source tierce.

Deux familles, à ne pas confondre. Les **objets de monde** sont posés sur la
grille et suivent la projection 2:1 ; ils ont une ombre et un point d'appui. Les
**icônes d'interface** sont vues de face, sans ombre ni perspective, et vivent
dans un repère d'écran. La même arme a donc deux images, et c'est voulu : une
icône dessinée en isométrique est illisible dans une case de 20 pixels.

Le vol d'un objet qui sort d'une caisse n'est pas une animation : c'est une
trajectoire calculée par le moteur, qui déplace ce sprite le long d'une parabole
jusqu'à la jauge. Le générateur ne fournit que la boucle de scintillement.
"""

import argparse
import json
import math
from pathlib import Path

from PIL import Image

import primitives_iso as prim
from manifestes import ecrire_manifeste

TRANSPARENT = (0, 0, 0, 0)

TEINTES = {
    "vie": (206, 74, 82),
    "verre": (176, 208, 216),
    "acier": (188, 194, 204),
    "acier_sombre": (108, 114, 128),
    "feu": (232, 148, 60),
    # Le crachat de la Buse, seule teinte du catalogue à porter une obligation :
    # la conception veut que le projectile ennemi n'existe nulle part ailleurs,
    # parce que « est-ce que ça me fait mal ? » est la question qu'on se pose
    # sous pression.
    #
    # **Elle est ce violet-ci et pas un violet plus sage, et la raison n'est pas
    # esthétique.** Le catalogue est fait de gris, de bruns, de verts et de bleus
    # sourds : la seule région libre est du côté du magenta. La ramener vers le
    # violet la rapprocherait du bleu du décor et du tir perforant, c'est-à-dire
    # exactement des deux choses dont elle doit se détacher — le grief qui la
    # ferait déplacer est donc le symptôme de ce qu'elle évite.
    #
    # Mesuré contre les deux tirs du joueur, en écart perceptif : 126 de l'or du
    # tir de base, 69 du bleu du perforant, et 38 de son plus proche voisin dans
    # les six cents images. Sa saturation vaut 0,57 contre 0,67 pour l'or : elle
    # ne sera pas plus vive que ce qui est déjà à l'écran.
    "venin": (178, 96, 224),
    "poudre": (96, 96, 104),
    "energie": (108, 178, 224),
    "or": (222, 186, 74),
    "platre": (206, 202, 194),
    "bois": (168, 126, 78),
    # Le cuivre, celui d'une bobine, qui va au décor urbain. Le rouge du fer à
    # cheval classique se serait perdu au milieu de cent créatures rouges,
    # c'est-à-dire au moment précis où l'on cherche l'aimant.
    #
    # **Il n'est pas réservé, et deux voisins le disent** : 13,4 du corps de
    # l'Éclateur, 15,6 de la caisse. C'est assumé — la godoc d'`aimant()` dit
    # pourquoi : la silhouette porte la lecture, et le fer à cheval est la seule
    # forme du catalogue qu'on nomme sans légende.
    "aimant": (198, 126, 78),
    "aimant_pole": (238, 202, 156),
    # **Un turquoise, et le choix se tient à trois écarts en même temps** — c'est
    # ce qui l'a fait préférer au vert, qui paraissait pourtant le mieux
    # argumenté : le vert (72, 214, 132) tombe à 10,6 de `teinteSoigneur`, soit
    # moins que ce qui séparait le Passant du Secouriste avant qu'on le corrige.
    # Il aurait déplacé la collision au lieu de la fermer.
    #
    # Mesuré : 29,0 du plus proche des quatre-vingt-quinze teintes de tête du
    # catalogue, 28,5 de la plus proche des teintes que le rendu ajoute, et 45,6
    # du sol livré le plus proche pour 11,5 de luminance. Aucun des trois n'est le
    # meilleur pris seul, et c'est le seul candidat où aucun n'est mauvais.
    "gemme": (64, 206, 178),
}


def _matiere(nom, teinte, contraste=1.0):
    """Déclare une matière et ses flancs, comme le fait `figurines.py`.

    `contraste` atténue l'ombrage. Il existe parce qu'un volume tire son relief
    de trois niveaux de teinte, et qu'un objet de six pixels de large n'a pas la
    place de les porter : l'ombrage qui donne du volume à une caisse y mange la
    couleur, et ce qu'on voit à l'écran n'est plus la teinte choisie mais sa
    moitié sombre.
    """
    def melange(cible, force):
        return tuple(round(c + (t - c) * force) for c, t in zip(teinte, cible))

    prim.MATIERES[nom] = (teinte,
                           melange((0, 0, 0), 0.28 * contraste),
                           melange((0, 0, 0), 0.44 * contraste),
                           melange((255, 255, 255), 0.32))
    return nom


def _bloc(largeur, hauteur, teinte, arete=True, contraste=1.0):
    # Le contraste entre dans le nom de la matière : deux blocs de la même teinte
    # à deux ombrages sont deux matières, et partager le nom ferait rendre au
    # second ce que le premier a déclaré.
    nom = f"_o_{teinte}" if contraste == 1.0 else f"_o_{teinte}_{contraste}"
    return prim.volume(elevation=hauteur,
                        matiere=_matiere(nom, TEINTES[teinte], contraste),
                        largeur_tuile=largeur, bandes=2, arete=arete)


def _poser(fond, piece, cx, cy):
    fond.alpha_composite(piece, (round(cx - piece.width / 2), round(cy - piece.height)))


def _ombre(fond, cx, cy, largeur=0.16):
    ombre = prim.volume(tx=largeur, ty=largeur, elevation=0, matiere="bitume", arete=False)
    ombre.putalpha(ombre.getchannel("A").point(lambda v: 70 if v else 0))
    _poser(fond, ombre, cx, cy + ombre.height // 2)


# --- Projectiles -----------------------------------------------------------
# Petits, sans ombre : ils volent. Leur lisibilité tient à la couleur, pas à la
# forme — à cette taille une silhouette ne se distingue plus.

def projectile(teinte="or", cote=8, hauteur=4, contraste=0.45, liseré=None):
    """Un projectile, dont l'ombrage est atténué faute de place pour le porter.

    Six pixels de large ne tiennent pas trois niveaux de teinte : à l'ombrage
    plein, la moitié sombre l'emporte et l'objet se lit plus foncé que la couleur
    qu'on lui a donnée. Ce qui se juge ici n'est pas le relief mais le contraste
    avec le sol qu'il traverse, puisque tout ce qu'un projectile doit faire est
    d'être suivi des yeux.
    """
    img = Image.new("RGBA", (cote, cote), TRANSPARENT)
    corps = _bloc(cote - 2, hauteur, teinte, contraste=contraste)
    _poser(img, corps, cote / 2, cote - 1)
    if liseré:
        img.info["contour"] = liseré
    return img


def projectile_base():
    return projectile("or", 8, 3)


def projectile_perforant():
    return projectile("energie", 10, 3)


def projectile_ennemi():
    """Le tir de la Buse : masse sombre, pourtour clair.

    **C'est le seul objet du catalogue dont le contour est inversé**, et la
    raison est qu'il est le seul qu'on doive voir *arriver*. Les tirs du joueur,
    on ne les esquive pas : leur donner la même saillance remplirait l'écran
    d'objets qui crient tous, et le liseré cesserait de dire « danger » faute de
    distinguer quoi que ce soit.

    **Ne pas éclaircir le violet en croyant améliorer.** L'éclaircissement était
    la bonne réponse tant que le contour restait foncé — il détachait le tir de
    la horde. Le liseré inversé le rend nuisible : il rapproche la masse du sol,
    qui est à 162 de luminance quand les créatures sont à 83, et fait retomber le
    pire cas de 81 à 69. Les deux réglages tirent en sens contraire, et c'est
    celui-ci qui gagne parce qu'il ne dépend d'aucun fond.
    """
    return projectile("venin", 8, 4, liseré=((248, 240, 255), 0.55))


# --- Ramassables -----------------------------------------------------------
# Posés au sol, donc ombre et point d'appui. Le scintillement est une boucle de
# quatre images : c'est ce qui les distingue du décor, qui ne bouge jamais.

SCINTILLEMENT_IMAGES = 4
SCINTILLEMENT_AMPLITUDE = 2


def _scintillement(base, images=SCINTILLEMENT_IMAGES,
                   amplitude=SCINTILLEMENT_AMPLITUDE):
    planche = Image.new("RGBA", (base.width * images, base.height + amplitude), TRANSPARENT)
    for i in range(images):
        hauteur = round(amplitude * (1 - math.cos(i / images * 2 * math.pi)) / 2)
        planche.alpha_composite(base, (i * base.width, amplitude - hauteur))
    return planche


def fiole():
    """Flacon de soin : verre, liquide rouge, bouchon."""
    img = Image.new("RGBA", (20, 22), TRANSPARENT)
    _ombre(img, 10, 21, 0.14)
    _poser(img, _bloc(12, 7, "verre"), 10, 20)
    _poser(img, _bloc(9, 5, "vie"), 10, 18)
    _poser(img, _bloc(5, 3, "acier_sombre"), 10, 12)
    return img


def _arme_au_sol(icone, socle="acier_sombre"):
    img = Image.new("RGBA", (24, 26), TRANSPARENT)
    _ombre(img, 12, 25, 0.18)
    _poser(img, _bloc(14, 4, socle), 12, 24)
    img.alpha_composite(icone, ((24 - icone.width) // 2, 24 - 4 - icone.height))
    return img


# --- Icônes d'armes --------------------------------------------------------
# Vues de face, sans perspective : dans une case de 20 pixels, l'isométrie ne
# se lit plus. Ce sont des pictogrammes, pas des objets.

def _icone(cote=20):
    return Image.new("RGBA", (cote, cote), TRANSPARENT)


def _rect(img, x0, y0, x1, y1, teinte, bord=True):
    px = img.load()
    couleur = TEINTES[teinte] + (255,)
    sombre = tuple(round(c * 0.55) for c in TEINTES[teinte]) + (255,)
    for y in range(y0, y1 + 1):
        for x in range(x0, x1 + 1):
            if not (0 <= x < img.width and 0 <= y < img.height):
                continue
            bordure = bord and (x in (x0, x1) or y in (y0, y1))
            px[x, y] = sombre if bordure else couleur


def icone_fusil():
    img = _icone()
    _rect(img, 2, 9, 15, 12, "acier")
    _rect(img, 13, 7, 17, 11, "acier_sombre")
    _rect(img, 4, 12, 8, 16, "poudre")
    return img


def icone_lance_flammes():
    img = _icone()
    _rect(img, 3, 8, 12, 12, "acier_sombre")
    _rect(img, 12, 9, 17, 11, "feu")
    _rect(img, 4, 12, 7, 16, "poudre")
    return img


def icone_grenade():
    """La seule icone d'arme qu'une partie jouee n'a pas reconnue.

    **Elle etait batie sur « poudre », la teinte la plus sombre du catalogue.**
    Le bord d'un rectangle vaut 0,55 de sa teinte, ce qui mettait le contour de
    l'icone a 53 de luminance contre 28 pour le fond de sa case : vingt-cinq
    d'ecart, quand le fusil en a soixante-dix-huit par « acier » et le
    lance-flammes soixante et un par « feu ». Elle se noyait dans son propre
    cadre, et ce projet retient quatre-vingts comme l'ordre de grandeur qui
    detache.

    **Et sa forme ne disait rien** : trois rectangles empiles dont un corps carre
    de neuf sur neuf, qui se lisait comme une boite. Ce qui nomme une grenade est
    la cuillere le long du corps et l'anneau de goupille — deux signes que les
    autres icones n'ont pas, donc deux signes qui la distinguent d'elles autant
    que d'un objet quelconque.

    L'anneau prend l'or plutot que l'acier : c'est le seul detail de trois
    pixels de l'icone, et il lui faut la valeur la plus haute pour survivre a
    cette taille.
    """
    img = _icone()
    # Le corps d'un seul bloc, plus haut que large : c'est ce qui le separe
    # d'une caisse. Deux blocs empiles y faisaient une marche, et l'icone se
    # lisait comme une pile.
    _rect(img, 5, 7, 12, 18, "acier")
    # Une nervure courte, sans bord et sans toucher les cotes : elle marque la
    # fonte a fragmentation. Deux qui traversaient decoupaient le corps en
    # bandes, et rendaient l'objet plus large qu'il n'est haut a l'oeil.
    _rect(img, 7, 12, 10, 12, "acier_sombre", bord=False)
    # Le col, puis la cuillere qui descend le long du corps et l'anneau.
    _rect(img, 7, 4, 10, 7, "acier_sombre")
    _rect(img, 12, 3, 14, 12, "acier_sombre")
    _rect(img, 15, 3, 17, 6, "or")
    return img


def icone_tourelle():
    img = _icone()
    _rect(img, 4, 13, 16, 17, "acier_sombre")
    _rect(img, 7, 7, 13, 13, "acier")
    _rect(img, 12, 9, 18, 11, "acier_sombre")
    return img


def icone_fiole():
    img = _icone()
    _rect(img, 6, 7, 13, 17, "verre")
    _rect(img, 7, 11, 12, 16, "vie")
    _rect(img, 8, 4, 11, 7, "acier_sombre")
    return img


ARMES = {"fusil": icone_fusil, "lance_flammes": icone_lance_flammes,
         "grenade": icone_grenade, "tourelle": icone_tourelle}


# --- Ramassables et caisse -------------------------------------------------
# Déplacés depuis le générateur de décor : ce sont des éléments de jeu, pas des
# éléments de lieu. Une caisse se casse, une gemme se ramasse — rien de commun
# avec un mur.

def caisse():
    return prim.contour(prim.nervures(prim.grain(prim.volume(elevation=16, matiere="carton",
                                         largeur_tuile=32), graine=5), pas=9))


def palette():
    return prim.volume(elevation=6, matiere="bois")


def gemme():
    """Le ramassable de base, et le seul qu'on voie par centaines.

    **Sa teinte est à elle et n'est empruntée à rien.** Elle a longtemps été
    celle de la matière « peinture » des primitives, partagée avec le décor : la
    gemme était alors identique **au byte près** à la teinte de tête de six
    pièces livrées — comptoir, wagon, boutique, distributeur, scooter. Un tas de
    gemmes sur un comptoir ne se voyait pas, et le chapitre 2 lui demande
    l'inverse : « la quantité au sol dit ce qu'on va gagner ».

    Plus petite que l'aimant de moitié, parce qu'on en compte un tas plutôt
    qu'on ne repère une pièce.
    """
    return prim.volume(elevation=3, matiere=_matiere("_o_gemme", TEINTES["gemme"]),
                       largeur_tuile=10, arete=False)


def aimant():
    """Fer à cheval : deux branches, une culasse, et les pôles en clair.

    **La silhouette porte la lecture, pas la couleur.** Le fer à cheval est la
    seule forme qu'on identifie comme un aimant sans légende, et c'est ce qu'il
    faut à un objet qu'on doit reconnaître de l'autre bout de la salle pour
    décider d'aller le chercher. Une bobine ou un cylindre seraient plus justes
    dans un décor urbain et ne diraient rien à cette taille.

    Les pôles en clair ne sont pas un ornement : ce sont eux qui distinguent le
    fer à cheval d'un simple U, donc l'aimant d'une pièce de mobilier.

    Plus grand que la gemme du double, comme à l'écran — il ne s'agit pas
    d'estimer un tas mais de repérer un objet unique.
    """
    img = Image.new("RGBA", (22, 20), TRANSPARENT)
    _ombre(img, 11, 19, 0.18)

    # **Trapu, et c'est ce qui le distingue d'une table.** Des branches longues
    # sur une culasse mince lisent comme un meuble ; le fer à cheval est plus
    # large que haut, et sa culasse pèse autant que ses branches.
    _poser(img, _bloc(6, 9, "aimant"), 5, 18)
    _poser(img, _bloc(6, 9, "aimant"), 17, 18)
    _poser(img, _bloc(18, 5, "aimant"), 11, 13)

    # **L'ouverture entre les branches est ce qui fait le fer à cheval.** Trop
    # serrée, la culasse la recouvre et l'objet devient un bloc à deux pieds ;
    # c'est le vide qu'on reconnaît, pas la matière.
    _poser(img, _bloc(6, 3, "aimant_pole"), 5, 18)
    _poser(img, _bloc(6, 3, "aimant_pole"), 17, 18)
    return img


def caisse_cassee():
    base = prim.volume(elevation=9, matiere="carton", largeur_tuile=32)
    return prim.eventrer(prim.grain(base, densite=0.18, graine=15,
                          faces=("dessus", "gauche", "droite")), graine=16)


# --- Quartier : bâtiments --------------------------------------------------



# --- Obstacles destructibles -----------------------------------------------
# Ils appartiennent aux objets et non au décor : un mur qui cède est une
# mécanique de jeu, pas un élément de lieu. L'auteur du niveau les pose comme
# des obstacles ordinaires, la topologie reste donc validable — c'est ce qui
# distingue une ouverture prévue d'un mur qui s'effondre sous la pression.

def _mince(teinte, elevation, epaisseur=0.18, longueur=1.0):
    return prim.volume(tx=longueur, ty=epaisseur, elevation=elevation,
                       matiere=_matiere(f"_o_{teinte}", TEINTES[teinte]), bandes=2)


def cloison_fragile():
    """Placo : haute mais légère, deux montants apparents."""
    corps = _mince("platre", 40)
    prim.nervures(corps, pas=9, force=0.10)
    return corps


def cloison_fragile_cassee():
    """Après rupture : un moignon bas, franchissable, qui garde la trace."""
    corps = _mince("platre", 10)
    return prim.eventrer(corps, densite=0.26, graine=61)


def vitrine():
    """Verre : le seul obstacle destructible qu'on voit à travers.

    Par `empiler` et non par un canevas monté à la main : la composition
    manuelle perd `info`, et l'emprise retombait sur la valeur par défaut d'un
    quart de tuile — celle d'une fiole, pour une cloison mince que le champ de
    flux doit contourner sur toute sa longueur.
    """
    cadre = _mince("acier_sombre", 34)
    verre = prim.volume(tx=0.86, ty=0.10, elevation=26,
                        matiere=_matiere("_o_verre", TEINTES["verre"]), bandes=2)
    verre.putalpha(verre.getchannel("A").point(lambda v: 150 if v else 0))
    return prim.empiler((cadre, 0, 0),
                        (verre, (cadre.width - verre.width) // 2,
                         cadre.height - verre.height - 4))


def vitrine_cassee():
    """Après rupture : le châssis bas, qui garde la trace de la devanture."""
    cadre = _mince("acier_sombre", 12)
    return prim.eventrer(cadre, densite=0.30, graine=62)


def grille_ventilation():
    """Bouche d'aération : basse, elle se casse vite et ouvre un raccourci."""
    corps = _mince("acier", 22, epaisseur=0.16, longueur=0.7)
    prim.nervures(corps, pas=3, force=0.22)
    return corps


def grille_ventilation_cassee():
    corps = _mince("acier", 8, epaisseur=0.16, longueur=0.7)
    return prim.eventrer(corps, densite=0.32, graine=63)


def rideau_fer():
    """Rideau de boutique : le plus résistant des trois, et le plus voyant."""
    corps = _mince("acier_sombre", 46)
    prim.nervures(corps, pas=4, force=0.16)
    return corps


def rideau_fer_casse():
    corps = _mince("acier_sombre", 14)
    return prim.eventrer(corps, densite=0.24, graine=64)



# --- Éclats et effets ------------------------------------------------------
# Un éclat n'est pas une animation : c'est une particule que le moteur émet en
# nombre et déplace sur une parabole, avec sa propre rotation et sa propre
# durée. Le générateur ne fournit que les formes — trois par matière, pour que
# deux éclats voisins ne soient pas identiques.
#
# Une explosion générique serait une erreur : le verre, le plâtre et la tôle ne
# se cassent pas de la même façon, et c'est ce qui dit au joueur ce qu'il vient
# d'ouvrir.

ECLATS = {
    "bois": ("bois", (5, 4, 6)),
    "verre": ("verre", (4, 3, 5)),
    "platre": ("platre", (4, 3, 5)),
    "metal": ("acier", (4, 5, 3)),
    "chair": ("vie", (3, 4, 3)),
}


def eclats(matiere):
    """Bande de trois particules d'une même matière, dans une case commune.

    Elles restent minuscules : à l'échelle du jeu un éclat n'est qu'un point de
    matière projeté, et le grossir en fait un objet posé au sol.
    """
    teinte, tailles = ECLATS[matiere]
    cote = 8
    planche = Image.new("RGBA", (cote * len(tailles), cote), TRANSPARENT)
    couleur = TEINTES[teinte]
    sombre = tuple(round(c * 0.6) for c in couleur)
    for i, taille in enumerate(tailles):
        origine = i * cote + (cote - taille) // 2
        haut = (cote - taille) // 2
        for y in range(taille):
            for x in range(taille):
                if x + y >= taille - 1 and x - y <= taille - 2:
                    teinte_pixel = couleur if y < taille - 1 else sombre
                    planche.putpixel((origine + x, haut + y), teinte_pixel + (255,))
    return planche


def etincelle():
    """Impact d'un projectile : trois images, très courtes, sans matière.

    Elle ne dit pas ce qui a été touché, seulement que le tir a porté — c'est
    ce retour-là qui manque le plus quand on tire sans le voir.

    **La gerbe part de deux et non de un**, parce qu'un rayon d'un pixel ne
    survit pas à l'aplatissement isométrique : la moitié verticale vaut alors un
    demi-pixel, l'arrondi la ramène à zéro, et les huit directions retombent sur
    trois pixels alignés. La première image montrait un tiret de trois pixels
    là où elle doit montrer une gerbe.

    Pas d'alpha dégressif : `prim.reduire` le binarise à 128, si bien que la
    troisième image, écrite à 115, sortait **entièrement vide** — trois images
    déclarées, deux dessinées. Ce qui s'éteint est la teinte.
    """
    images, cote = 3, 10
    planche = Image.new("RGBA", (cote * images, cote), TRANSPARENT)
    # Du plus vif au plus éteint : la gerbe s'élargit en refroidissant.
    teintes = ((255, 248, 214), (250, 226, 150), (214, 178, 108))
    for i in range(images):
        rayon = 2 + i
        for angle in range(0, 360, 45):
            dx = round(rayon * math.cos(math.radians(angle)))
            dy = round(rayon * math.sin(math.radians(angle)) * 0.5)
            x = i * cote + cote // 2 + dx
            y = cote // 2 + dy
            if 0 <= x - i * cote < cote and 0 <= y < cote:
                planche.putpixel((x, y), teintes[i] + (255,))
        planche.putpixel((i * cote + cote // 2, cote // 2), (255, 248, 214, 255))
    return planche


def souffle():
    """Explosion de la Baudruche : anneaux francs qui s'élargissent.

    Non bouclée, cinq images. Un dégradé lissé virerait à la tache brune une
    fois quantifié — même écueil que pour le télégraphe de sa mort.

    **L'épaisseur croît avec le rayon**, et c'est une contrainte de tracé et non
    un choix : un anneau d'épaisseur constante se troue quand son périmètre
    s'allonge, parce que le nombre de pixels qui le composent ne suit pas. Le
    dernier — le seul que l'œil lit comme la portée — était continu à 73,9 % sur
    sept cent vingt directions, et l'aplatissement isométrique le rompait en haut
    et en bas, c'est-à-dire là où l'on cherche la limite.

    Pas d'alpha dégressif : `prim.reduire` le binarise à 128, si bien qu'un
    dégradé écrit ici ressortirait plein. Ce qui s'éteint est la teinte.
    """
    images, cote = 5, 48
    planche = Image.new("RGBA", (cote * images, cote), TRANSPARENT)
    centre = cote // 2
    for i in range(images):
        rayon = 4 + i * 4
        epaisseur = 3 + i
        teinte = ((255, 226, 150), (250, 190, 96), (232, 148, 60),
                  (188, 104, 48), (128, 72, 44))[i]
        for y in range(cote):
            for x in range(cote):
                dx = x - centre
                dy = (y - centre) * 2          # l'onde s'étale dans le plan du sol
                distance = math.hypot(dx, dy)
                if rayon - epaisseur <= distance <= rayon:
                    planche.putpixel((i * cote + x, y), teinte + (255,))
    return planche


# --- Caisse cassable -------------------------------------------------------
# Trois états ne suffisent pas : le délai de contact doit se voir, sinon le
# joueur ne sait pas qu'il est en train de casser quelque chose et croit à un
# blocage. D'où un cycle d'appui qui boucle tant qu'il pousse, puis une rupture
# qui ne boucle pas.

def _comprimer(img, tassement, elargissement):
    """Écrase la caisse verticalement en l'élargissant : elle encaisse.

    Le canevas suit l'élargissement au lieu d'une marge fixe. Il en portait
    quatre, ce qui rognait toute image plus large — la dernière du cycle de
    rupture le dépassait déjà d'un pixel, et rien ne le disait.
    """
    largeur = img.width + elargissement
    hauteur = max(1, img.height - tassement)
    petite = img.resize((largeur, hauteur), Image.NEAREST)
    fond = Image.new("RGBA", (largeur, img.height), TRANSPARENT)
    fond.alpha_composite(petite, (0, img.height - hauteur))
    return fond


def caisse_appui(images=3):
    """Jouée pendant que le joueur pousse : la caisse s'écrase en s'élargissant.

    **Elle ne boucle pas, et le manifeste le déclare.** Sa durée est celle du
    délai d'appui, si bien qu'elle se joue une fois et s'achève à l'instant où la
    caisse cède : le moteur en tire l'image du décompte qui reste, comme de toute
    animation qui finit. Bouclée, sa phase viendrait du tick et n'aurait aucun
    rapport avec la poussée — on toucherait une caisse déjà écrasée, qui se
    redresserait ensuite.

    L'écrasement est franc parce qu'il porte à lui seul l'information : sans lui,
    un joueur ralenti sur une caisse croit avoir buté sur un mur.
    """
    base = caisse()
    cadres = []
    for i in range(images):
        avancement = i / max(1, images - 1)
        # Un cinquième de la hauteur au dernier tiers de seconde. Trois pixels
        # avaient été essayés d'abord : la planche montrait une caisse intacte, et
        # une déformation qu'on ne voit pas ne dit rien à qui pousse dessus.
        cadres.append(_comprimer(base, round(6 * avancement), round(6 * avancement)))
    largeur = max(c.width for c in cadres)
    hauteur = max(c.height for c in cadres)
    planche = Image.new("RGBA", (largeur * images, hauteur), TRANSPARENT)
    for i, c in enumerate(cadres):
        planche.alpha_composite(c, (i * largeur + (largeur - c.width) // 2,
                                    hauteur - c.height))
    return planche, (largeur, hauteur)


def caisse_rupture(images=3):
    """Éclatement, non bouclé : la dernière image reste au sol."""
    base = caisse()
    cadres = []
    for i in range(images):
        avancement = (i + 1) / images
        morceau = _comprimer(base, round(6 * avancement), round(5 * avancement))
        prim.eventrer(morceau, densite=0.18 + 0.30 * avancement, graine=90 + i)
        cadres.append(morceau)
    largeur = max(c.width for c in cadres)
    hauteur = max(c.height for c in cadres)
    planche = Image.new("RGBA", (largeur * images, hauteur), TRANSPARENT)
    for i, c in enumerate(cadres):
        planche.alpha_composite(c, (i * largeur + (largeur - c.width) // 2,
                                    hauteur - c.height))
    return planche, (largeur, hauteur)


# Tout ce que le moteur doit savoir d'un objet sans rien coder en dur : ce qui
# le bloque, ce qui le détruit, ce qu'il projette et ce qu'il fait entendre.
#
# `mode` distingue les deux façons de casser du jeu, et c'est le geste qui les
# sépare, pas la nature du dégât : la caisse cède à l'appui, en la traversant,
# sans interrompre la course ; un obstacle fragile demande de s'arrêter contre
# lui et d'appuyer sur la touche d'interaction. Le tir de base ne les casse
# jamais — il ne cible que des ennemis et ne saurait pas distinguer un rideau de
# fer d'une créature.
#
# `touches` est en touches de l'arme de base au premier niveau, même unité que
# la résistance des créatures.
DESTRUCTION = {
    "caisse": {"mode": "contact", "delai_ms": 330, "ruine": "caisse_cassee",
               "eclats": "bois", "cycle_appui": "caisse_appui",
               "cycle_rupture": "caisse_rupture",
               "son_appui": "caisse_appui", "son_rupture": "caisse_rupture"},
    "cloison_fragile": {"mode": "interaction", "touches": 8, "ruine": "cloison_fragile_cassee",
                        "eclats": "platre", "son_rupture": "caisse_rupture"},
    "vitrine": {"mode": "interaction", "touches": 5, "ruine": "vitrine_cassee",
                "eclats": "verre", "son_rupture": "caisse_rupture"},
    "grille_ventilation": {"mode": "interaction", "touches": 3,
                           "ruine": "grille_ventilation_cassee",
                           "eclats": "metal", "son_rupture": "caisse_rupture"},
    "rideau_fer": {"mode": "interaction", "touches": 20, "ruine": "rideau_fer_casse",
                   "eclats": "metal", "son_rupture": "caisse_rupture"},
}

# Ce qui arrête un déplacement. Les ruines ne bloquent plus : c'est tout
# l'intérêt d'avoir cassé quelque chose.
BLOQUANTS = {"palette", "cloison_fragile", "vitrine",
             "grille_ventilation", "rideau_fer"}

# Ce qui se franchit en payant, et le prix en pas.
#
# Le coût multiplie la longueur perçue d'une case et divise la vitesse de qui la
# traverse. C'est le vrai prix de la ressource : ramasser, c'est perdre du
# terrain. Trois quand une flaque en vaut deux — casser doit coûter plus que
# patauger —, et le chiffre se juge en jouant.
#
# **Deux tables ici, une seule dans le décor, et la différence n'est pas un
# relâchement.** Là-bas toute forme est une case, si bien que ce qui n'est pas
# franchissable bloque et qu'une table dit les deux. Ici la plupart des entrées
# ne sont sur aucune grille — une gemme, un éclat, une icône —, et les ranger
# d'un côté ou de l'autre leur inventerait une passabilité qu'elles n'ont pas.
# Ce que l'unique table du décor fermait, un coût orphelin sur ce qui bloque, se
# ferme ici par le refus de l'intersection.
COUTS = {"caisse": 3}


def _passabilite(nom):
    """Rend le couple `bloquant` / `cout_traversee`, et refuse ses deux
    contradictions.

    Le fait porte le booléen, la valeur ne se déclare que quand il est faux :
    c'est la forme du décor, et le chargeur Go refuse les mêmes couples.

    **La seconde n'a pas d'équivalent dans le décor, et sa place est prise par
    le mode de destruction.** Là-bas tout ce qui se franchit doit déclarer son
    coût ; l'exiger de tout objet franchissable reviendrait ici à en réclamer un
    au projectile et à l'icône. Ce qu'on peut exiger, en revanche, est qu'une
    chose qu'on casse **en la traversant** puisse se traverser et coûte : sans
    cela la caisse cède au premier contact, et le ralentissement, qui est la
    mécanique elle-même, ne se paie jamais.
    """
    bloque, cout = nom in BLOQUANTS, COUTS.get(nom)
    if bloque and cout is not None:
        raise ValueError(f"{nom} bloque et déclare pourtant un coût de traversée :"
                         " on ne ralentit pas ce qui est arrêté")
    if DESTRUCTION.get(nom, {}).get("mode") == "contact" and cout is None:
        raise ValueError(f"{nom} se casse en le traversant sans coûter à traverser :"
                         " le délai s'écoulerait pendant qu'on est déjà de l'autre côté")
    return bloque, cout

# Son joué au ramassage ou à l'usage, quand il y en a un.
SONS = {"fiole": "soin", "aimant": "aimant"}

# Un renvoi vers une famille et non vers un fichier : le catalogue porte
# `gemme_0` à `gemme_7`, et le moteur avance d'un degré à chaque ramassage
# rapproché. Deux clés distinctes plutôt qu'une seule à interpréter — sans
# quoi le contrôle ne peut que comparer des préfixes, et accepte alors « gem »
# et « g » aussi bien que « gemme ».
FAMILLES_SONS = {"gemme": "gemme"}

# Ce que vaut un objet en jeu. Ces nombres vivent ici et non dans le code : la
# règle du dépôt est que les données ne sont pas du code, et un manifeste est
# déjà l'endroit où le moteur va les chercher.
#
# Les projectiles n'y figurent plus. Un projectile est un objet qui vole : il
# porte sa taille, son ancrage et son emprise, et rien de ce qui se règle en
# jouant. Dégâts, portée et vitesse appartiennent à celui qui tire — l'arme pour
# le joueur, le profil pour une créature —, parce que ce sont les chiffres qu'on
# rouvrira le plus, et que les régler ne doit pas coûter une régénération de six
# cents images. La règle est au chapitre 4 de la conception ; `ressources.py`
# refuse ces champs pour qu'on ne les y remette pas par symétrie.
# L'expérience d'une gemme n'y figure plus non plus, et pour la même raison
# poussée d'un cran : une valeur vit à côté de ce qu'elle alimente. Celle-ci
# alimente les seuils de niveau, donc elle est descendue dans
# `assets/progression/manifeste.json`, où le seuil se règle. Le test n'est pas
# que le fichier soit généré mais qui l'édite et pourquoi — changer le dessin
# d'une gemme ne change rien à ce qu'elle rapporte, et l'inverse est vrai aussi.
VALEURS = {
    "fiole": {"soin": 30, "emplacements": 2},
}

# Les charges d'une arme lourde ne sont plus ici : elles ont déménagé dans
# assets/armes/manifeste.json, tenu à la main, par le critère qui avait déjà fait
# descendre l'expérience d'une gemme dans celui de la progression — une valeur vit
# à côté de ce qu'elle alimente, et régler une charge ne touche pas au dessin.
# ressources.py refuse qu'elles reviennent sur un objet.

CATALOGUE = {
    "caisse": caisse,
    "caisse_cassee": caisse_cassee,
    "palette": palette,
    "gemme": gemme,
    "cloison_fragile": cloison_fragile,
    "cloison_fragile_cassee": cloison_fragile_cassee,
    "vitrine": vitrine,
    "vitrine_cassee": vitrine_cassee,
    "grille_ventilation": grille_ventilation,
    "grille_ventilation_cassee": grille_ventilation_cassee,
    "rideau_fer": rideau_fer,
    "rideau_fer_casse": rideau_fer_casse,
    "projectile_base": projectile_base,
    "projectile_perforant": projectile_perforant,
    "projectile_ennemi": projectile_ennemi,
    "fiole": fiole,
    "gemme": gemme,
    "aimant": aimant,
}


def main():
    a = argparse.ArgumentParser(description=__doc__)
    a.add_argument("--sortie", type=Path, default=Path("assets/objets"))
    a.add_argument("--apercu", action="store_true")
    o = a.parse_args()

    o.sortie.mkdir(parents=True, exist_ok=True)
    manifeste = {}

    for nom, fabrique in CATALOGUE.items():
        brut = fabrique()
        # Une emprise absente n'est pas une petite emprise. Elle se perd dès
        # qu'une forme recompose ses volumes dans une image neuve, et un défaut
        # silencieux poserait alors un carré d'un quart de tuile là où la
        # vitrine est une cloison mince — le champ de flux contournerait autre
        # chose que ce qui est dessiné. Ce qui entre dans la grille doit donc la
        # déclarer, qu'il l'arrête ou qu'il la fasse payer.
        bloquant, cout = _passabilite(nom)
        emprise = brut.info.get("emprise")
        if emprise is None:
            if bloquant or cout is not None:
                raise ValueError(f"{nom} entre dans la grille sans déclarer son emprise :"
                                 " la composition l'a perdue en chemin")
            emprise = (0.25, 0.25)

        # Le pourtour va vers un presque-noir sauf déclaration contraire : seul
        # ce qu'on doit voir arriver porte un liseré clair, et `contour` dit
        # pourquoi cette dérogation ne s'étend pas.
        cible, force = brut.info.get("contour", ((24, 24, 28), 0.40))
        img = prim.reduire(prim.contour(brut, force=force, cible=cible), couleurs=12)
        # Recadrage : le moteur pose un objet par son point d'appui, pas par le
        # coin d'une image. Des marges transparentes ne feraient que décaler la
        # position réelle sans que rien ne le signale.
        img = img.crop(img.getbbox())
        img.info.update(brut.info)
        img.save(o.sortie / f"{nom}.png")

        # Les trois champs qui commandent le rendu et la topologie, dérivés
        # comme pour le décor et non saisis. Sans eux, le rideau de fer fait
        # 65 pixels de haut et rien ne dit au moteur qu'il masque un
        # personnage : il devrait en coder la liste, contre l'invariant du
        # manifeste-contrat.
        haut = prim.elevation_reelle(img) if "hauteur_dessus" in img.info else 0
        manifeste[nom] = {"taille": list(img.size),
                          "ancrage": [img.width // 2, img.height - 1],
                          "emprise": list(emprise),
                          "elevation": haut,
                          "categorie": prim.categorie(bloquant, haut),
                          "masquant": haut > prim.PLAFOND_OBSTACLE_BAS,
                          "famille": "monde",
                          "bloquant": bloquant,
                          **({} if cout is None else {"cout_traversee": cout})}
        if nom in VALEURS:
            manifeste[nom].update(VALEURS[nom])
        if nom in DESTRUCTION:
            manifeste[nom]["destruction"] = DESTRUCTION[nom]
        if nom in SONS:
            manifeste[nom]["son"] = SONS[nom]
        if nom in FAMILLES_SONS:
            manifeste[nom]["famille_sons"] = FAMILLES_SONS[nom]
        # L'aimant scintille comme les deux autres ramassables : c'est ce qui
        # dit qu'un objet se prend plutôt qu'il ne décore.
        if nom in ("fiole", "gemme", "aimant"):
            bande = _scintillement(img)
            bande.save(o.sortie / f"{nom}_scintille.png")
            # L'amplitude est déclarée parce qu'elle est la seule chose que la
            # bande ne dit pas. Sa cellule fait la largeur de l'objet sur sa
            # hauteur plus le bombement, et son point d'appui descend d'autant :
            # sans ce nombre, une bande de 40 sur 10 pour un objet de 10 sur 8
            # ne se découpe qu'en devinant, ce que le manifeste-contrat existe
            # pour éviter.
            manifeste[nom]["scintillement"] = {"images": SCINTILLEMENT_IMAGES,
                                               "amplitude": SCINTILLEMENT_AMPLITUDE,
                                               "duree_ms": 140,
                                               "boucle": True}
        print(f"{nom:22} {img.size}")

    for matiere in ECLATS:
        planche = prim.reduire(eclats(matiere), couleurs=8)
        planche.save(o.sortie / f"eclats_{matiere}.png")
        cote = planche.width // 3
        # `cote` est une paire comme partout ailleurs, et non l'entier d'un
        # carré : un même nom qui porte deux formes selon la famille ne se lit
        # pas d'un seul type, et c'est son lecteur qui l'a montré. Le mot sur la
        # trajectoire est un commentaire, donc il en prend la clé — déclaré en
        # donnée, il obligeait à un champ que personne ne lirait jamais.
        manifeste[f"eclats_{matiere}"] = {"$comment": "trajectoire calculée par le moteur",
                                          "formes": 3, "cote": [cote, cote],
                                          "famille": "particule", "bloquant": False}
        print(f"{'eclats_' + matiere:22} 3 formes de {cote} px")

    for nom, fabrique, duree, boucle in (("etincelle", etincelle, 40, False),
                                         ("souffle", souffle, 60, False),
                                         ("caisse_appui", caisse_appui, 110, False),
                                         ("caisse_rupture", caisse_rupture, 90, False)):
        rendu = fabrique()
        if isinstance(rendu, tuple):
            planche, (largeur, hauteur) = rendu
        else:
            planche, largeur, hauteur = rendu, rendu.height, rendu.height
        planche = prim.reduire(planche, couleurs=14)
        planche.save(o.sortie / f"{nom}.png")
        # **« cycle » et non « monde » pour les deux bandes de la caisse.** Une
        # entrée du monde porte une taille, un ancrage et une emprise ; celles-ci
        # n'en ont aucun, et les ranger là donnait une famille dont les entrées
        # n'avaient pas la même forme — ce que leur lecteur a refusé. Un cycle
        # appartient à l'objet qui le cite, dont il prend l'ancrage ; un effet se
        # centre sur ce qu'il marque et n'appartient à personne.
        manifeste[nom] = {"images": planche.width // largeur, "cote": [largeur, hauteur],
                          "duree_ms": duree, "boucle": boucle,
                          "famille": "effet" if nom in ("etincelle", "souffle") else "cycle",
                          "bloquant": False}
        print(f"{nom:22} {planche.width // largeur} images de {largeur}x{hauteur}")

    (o.sortie / "armes").mkdir(exist_ok=True)
    for nom, fabrique in ARMES.items():
        icone = fabrique()
        icone.save(o.sortie / "armes" / f"{nom}_icone.png")
        au_sol = prim.reduire(prim.contour(_arme_au_sol(icone), force=0.40), couleurs=14)
        au_sol = au_sol.crop(au_sol.getbbox())
        au_sol.save(o.sortie / "armes" / f"{nom}_sol.png")
        manifeste[nom] = {"taille_icone": list(icone.size),
                          "taille_sol": list(au_sol.size),
                          "ancrage_sol": [au_sol.width // 2, au_sol.height - 1],
                          "famille": "arme", "bloquant": False,
                          "son": "ramassage_arme"}
        print(f"{nom:22} icône {icone.size}  sol {au_sol.size}")

    icone_fiole().save(o.sortie / "armes" / "fiole_icone.png")

    ecrire_manifeste(o.sortie / "manifeste.json", "objets.py",
                     {"version_format": 1, "objets": manifeste})
    print(f"\n{len(manifeste)} objets dans {o.sortie}")


if __name__ == "__main__":
    main()
