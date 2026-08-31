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
    "poudre": (96, 96, 104),
    "energie": (108, 178, 224),
    "or": (222, 186, 74),
    "platre": (206, 202, 194),
    "bois": (168, 126, 78),
}


def _matiere(nom, teinte):
    def melange(cible, force):
        return tuple(round(c + (t - c) * force) for c, t in zip(teinte, cible))

    prim.MATIERES[nom] = (teinte, melange((0, 0, 0), 0.28), melange((0, 0, 0), 0.44),
                           melange((255, 255, 255), 0.32))
    return nom


def _bloc(largeur, hauteur, teinte, arete=True):
    return prim.volume(elevation=hauteur, matiere=_matiere(f"_o_{teinte}", TEINTES[teinte]),
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

def projectile(teinte="or", cote=8, hauteur=4):
    img = Image.new("RGBA", (cote, cote), TRANSPARENT)
    corps = _bloc(cote - 2, hauteur, teinte)
    _poser(img, corps, cote / 2, cote - 1)
    return img


def projectile_base():
    return projectile("or", 8, 3)


def projectile_perforant():
    return projectile("energie", 10, 3)


def projectile_ennemi():
    return projectile("feu", 8, 4)


# --- Ramassables -----------------------------------------------------------
# Posés au sol, donc ombre et point d'appui. Le scintillement est une boucle de
# quatre images : c'est ce qui les distingue du décor, qui ne bouge jamais.

def _scintillement(base, images=4, amplitude=2):
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
    img = _icone()
    _rect(img, 6, 8, 14, 16, "poudre")
    _rect(img, 8, 5, 12, 8, "acier_sombre")
    _rect(img, 12, 4, 15, 6, "acier")
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
    return prim.volume(elevation=3, matiere="peinture", largeur_tuile=10, arete=False)


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
    """
    images, cote = 3, 10
    planche = Image.new("RGBA", (cote * images, cote), TRANSPARENT)
    for i in range(images):
        rayon = 1 + i
        alpha = 255 - i * 70
        for angle in range(0, 360, 45):
            dx = round(rayon * math.cos(math.radians(angle)))
            dy = round(rayon * math.sin(math.radians(angle)) * 0.5)
            x = i * cote + cote // 2 + dx
            y = cote // 2 + dy
            if 0 <= x - i * cote < cote and 0 <= y < cote:
                planche.putpixel((x, y), (250, 226, 150, alpha))
        planche.putpixel((i * cote + cote // 2, cote // 2), (255, 248, 214, alpha))
    return planche


def souffle():
    """Explosion de la Baudruche : anneaux francs qui s'élargissent.

    Non bouclée, cinq images. Un dégradé lissé virerait à la tache brune une
    fois quantifié — même écueil que pour le télégraphe de sa mort.
    """
    images, cote = 5, 48
    planche = Image.new("RGBA", (cote * images, cote), TRANSPARENT)
    centre = cote // 2
    for i in range(images):
        rayon = 4 + i * 4
        epaisseur = max(1, 3 - i // 2)
        teinte = ((255, 226, 150), (250, 190, 96), (232, 148, 60),
                  (188, 104, 48), (128, 72, 44))[i]
        for y in range(cote):
            for x in range(cote):
                dx = x - centre
                dy = (y - centre) * 2          # l'onde s'étale dans le plan du sol
                distance = math.hypot(dx, dy)
                if rayon - epaisseur <= distance <= rayon:
                    planche.putpixel((i * cote + x, y), teinte + (255 - i * 30,))
    return planche


# --- Caisse cassable -------------------------------------------------------
# Trois états ne suffisent pas : le délai de contact doit se voir, sinon le
# joueur ne sait pas qu'il est en train de casser quelque chose et croit à un
# blocage. D'où un cycle d'appui qui boucle tant qu'il pousse, puis une rupture
# qui ne boucle pas.

def _comprimer(img, tassement, elargissement):
    """Écrase la caisse verticalement en l'élargissant : elle encaisse."""
    largeur = img.width + elargissement
    hauteur = max(1, img.height - tassement)
    petite = img.resize((largeur, hauteur), Image.NEAREST)
    fond = Image.new("RGBA", (img.width + 4, img.height), TRANSPARENT)
    fond.alpha_composite(petite, ((fond.width - largeur) // 2, img.height - hauteur))
    return fond


def caisse_appui(images=3):
    """Boucle jouée pendant que le joueur pousse : la caisse tremble et cède."""
    base = caisse()
    cadres = []
    for i in range(images):
        avancement = i / max(1, images - 1)
        cadres.append(_comprimer(base, round(3 * avancement), round(3 * avancement)))
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
BLOQUANTS = {"caisse", "palette", "cloison_fragile", "vitrine",
             "grille_ventilation", "rideau_fer"}

# Son joué au ramassage ou à l'usage, quand il y en a un.
SONS = {"fiole": "soin"}

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
VALEURS = {
    "gemme": {"experience": 1},
    "fiole": {"soin": 30, "emplacements": 2},
}

CHARGES = {"fusil": 5, "lance_flammes": 3, "grenade": 3, "tourelle": 2}

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
        # chose que ce qui est dessiné. Ce qui bloque doit donc la déclarer.
        emprise = brut.info.get("emprise")
        if emprise is None:
            if nom in BLOQUANTS:
                raise ValueError(f"{nom} bloque sans déclarer son emprise :"
                                 " la composition l'a perdue en chemin")
            emprise = (0.25, 0.25)

        img = prim.reduire(prim.contour(brut, force=0.40), couleurs=12)
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
                          "categorie": prim.categorie(nom in BLOQUANTS, haut),
                          "masquant": haut > prim.PLAFOND_OBSTACLE_BAS,
                          "famille": "monde",
                          "bloquant": nom in BLOQUANTS}
        if nom in VALEURS:
            manifeste[nom].update(VALEURS[nom])
        if nom in DESTRUCTION:
            manifeste[nom]["destruction"] = DESTRUCTION[nom]
        if nom in SONS:
            manifeste[nom]["son"] = SONS[nom]
        if nom in FAMILLES_SONS:
            manifeste[nom]["famille_sons"] = FAMILLES_SONS[nom]
        if nom in ("fiole", "gemme"):
            bande = _scintillement(img)
            bande.save(o.sortie / f"{nom}_scintille.png")
            manifeste[nom]["scintillement"] = {"images": 4, "duree_ms": 140,
                                               "boucle": True}
        print(f"{nom:22} {img.size}")

    for matiere in ECLATS:
        planche = prim.reduire(eclats(matiere), couleurs=8)
        planche.save(o.sortie / f"eclats_{matiere}.png")
        cote = planche.width // 3
        manifeste[f"eclats_{matiere}"] = {"formes": 3, "cote": cote,
                                          "famille": "particule", "bloquant": False,
                                          "note": "trajectoire calculée par le moteur"}
        print(f"{'eclats_' + matiere:22} 3 formes de {cote} px")

    for nom, fabrique, duree, boucle in (("etincelle", etincelle, 40, False),
                                         ("souffle", souffle, 60, False),
                                         ("caisse_appui", caisse_appui, 110, True),
                                         ("caisse_rupture", caisse_rupture, 90, False)):
        rendu = fabrique()
        if isinstance(rendu, tuple):
            planche, (largeur, hauteur) = rendu
        else:
            planche, largeur, hauteur = rendu, rendu.height, rendu.height
        planche = prim.reduire(planche, couleurs=14)
        planche.save(o.sortie / f"{nom}.png")
        manifeste[nom] = {"images": planche.width // largeur, "cote": [largeur, hauteur],
                          "duree_ms": duree, "boucle": boucle,
                          "famille": "effet" if nom in ("etincelle", "souffle") else "monde",
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
                          "charges": CHARGES[nom],
                          "son": "ramassage_arme"}
        print(f"{nom:22} icône {icone.size}  sol {au_sol.size}")

    icone_fiole().save(o.sortie / "armes" / "fiole_icone.png")

    ecrire_manifeste(o.sortie / "manifeste.json", "objets.py",
                     {"version_format": 1, "objets": manifeste})
    print(f"\n{len(manifeste)} objets dans {o.sortie}")


if __name__ == "__main__":
    main()
