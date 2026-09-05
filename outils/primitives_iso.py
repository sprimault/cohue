# Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
# SPDX-License-Identifier: Apache-2.0

"""Primitives isométriques : le noyau partagé par tous les générateurs.

Une seule fonction calcule la face supérieure d'un volume, en coordonnées de
tuile — c'est elle qui garantit l'escalier régulier du 2:1 quelle que soit
l'emprise, y compris fractionnaire. Tout le reste compose, détaille ou contrôle.

Ce module ne connaît aucune forme du jeu : ni décor, ni créature, ni objet. Il
est importé par `decor_iso.py`, `figurines.py` et `objets.py`, qui eux
connaissent le contenu.
"""

import random

from PIL import Image

LARGEUR_TUILE = 64
TRANSPARENT = (0, 0, 0, 0)

# Un personnage fait 64 pixels de haut ; ce qui dépasse 24 en masque un et
# devient un décor de bordure plutôt qu'un obstacle à contourner. Ces deux
# grandeurs vivent ici parce qu'elles décrivent la projection et sa lisibilité,
# pas une forme du jeu — décor et objets s'y réfèrent tous les deux.
HAUTEUR_PERSONNAGE = 64
PLAFOND_OBSTACLE_BAS = 24

# dessus, flanc gauche, flanc droit, arête éclairée
MATIERES = {
    "beton": ((166, 166, 170), (118, 118, 124), (94, 94, 100), (198, 198, 202)),
    "beton_sombre": ((132, 132, 138), (94, 94, 100), (74, 74, 80), (160, 160, 166)),
    "bitume": ((84, 84, 92), (60, 60, 68), (46, 46, 54), (108, 108, 116)),
    "metal": ((198, 204, 214), (138, 146, 160), (106, 114, 128), (228, 232, 240)),
    "acier_sombre": ((122, 130, 144), (86, 92, 104), (66, 72, 82), (152, 160, 172)),
    "carton": ((190, 148, 100), (146, 112, 72), (114, 86, 54), (214, 178, 134)),
    "bois": ((150, 108, 68), (112, 80, 50), (86, 60, 38), (178, 140, 100)),
    "peinture": ((90, 132, 170), (64, 96, 126), (48, 74, 98), (126, 168, 204)),
    "rouge": ((176, 66, 62), (132, 46, 44), (102, 34, 32), (206, 104, 98)),
    "vert": ((84, 148, 96), (60, 110, 70), (46, 86, 54), (120, 182, 132)),
    "jaune": ((214, 178, 68), (168, 138, 48), (132, 106, 36), (238, 210, 118)),
    "verre": ((172, 206, 216), (120, 150, 160), (92, 118, 128), (206, 232, 238)),
    "tissu": ((142, 62, 84), (106, 44, 62), (82, 32, 48), (176, 94, 118)),
}

MARQUAGE = (232, 232, 224)
BANDE_JAUNE = (222, 186, 74)


def surface(tx, ty, largeur_tuile=LARGEUR_TUILE):
    """Rend {(x, y): (u, v)} pour la face supérieure d'une emprise tx sur ty.

    Le test se fait en coordonnées de tuile, donc les bords tombent exactement
    sur des droites de pente 1/2 : c'est ce qui donne l'escalier du 2:1.
    tx et ty acceptent des valeurs fractionnaires, ce qui permet les cloisons.
    """
    demi = largeur_tuile / 2
    largeur = round((tx + ty) * demi)
    hauteur = round((tx + ty) * demi / 2)
    origine = ty * demi

    marge = 0.5 / largeur_tuile

    points = {}
    for y in range(hauteur):
        for x in range(largeur):
            px = x + 0.5 - origine
            py = y + 0.5
            u = px / largeur_tuile + py / demi
            v = -px / largeur_tuile + py / demi
            if -marge <= u <= tx + marge and -marge <= v <= ty + marge:
                points[(x, y)] = (min(max(u, 0.0), tx), min(max(v, 0.0), ty))
    return points, largeur, hauteur


def elevation_reelle(img):
    """Hauteur d'une forme au-dessus du sol, en pixels.

    L'image contient la face supérieure et les flancs ; c'est la part sous le
    dessus qui dit de combien la forme dépasse, et donc si elle masque un
    personnage.
    """
    return max(0, img.height - img.info["hauteur_dessus"])


def categorie(bloquant, elevation):
    """Range une forme parmi les trois hauteurs de la vue de dessus.

    **La passabilité décide d'abord**, l'élévation ne départageant que ce qui
    bloque. Une porte ouverte dépasse de 48 pixels et reste un passage ; un quai
    ou un trottoir dépassent du sol et se marchent. Les classer sur leur seule
    hauteur ferait afficher des obstacles là où l'on passe, et la lecture
    topologique que l'éditeur promet ne vaudrait plus rien.

    La catégorie ne sert donc qu'à la topologie. Le rendu, lui, a `elevation` et
    `masquant` : une propriété qui servirait aux deux finirait par mal servir
    les deux.
    """
    if not bloquant:
        return "sol"
    return "obstacle_bas" if elevation <= PLAFOND_OBSTACLE_BAS else "haut"


def volume(tx=1, ty=1, elevation=0, matiere="beton", largeur_tuile=LARGEUR_TUILE,
           bandes=3, arete=True, peinture=None):
    """Volume isométrique : face supérieure, deux flancs, arête haute éclairée.

    peinture(u, v) peut rendre une couleur pour marquer la face supérieure.
    """
    dessus, gauche, droite, clair = MATIERES[matiere]
    points, largeur, hauteur = surface(tx, ty, largeur_tuile)
    img = Image.new("RGBA", (largeur, hauteur + elevation), TRANSPARENT)
    px = img.load()

    masque_dessus = set()
    masque_gauche = set()
    masque_droite = set()

    bas = {}
    haut = {}
    for (x, y), (u, v) in points.items():
        teinte = dessus
        if peinture is not None:
            choix = peinture(u, v)
            if choix is not None:
                teinte = choix
        px[x, y] = teinte + (255,)
        masque_dessus.add((x, y))
        if y > bas.get(x, -1):
            bas[x] = y
        if y < haut.get(x, 10**6):
            haut[x] = y

    if bas:
        creux = max(bas, key=lambda c: bas[c])
        for x, y in bas.items():
            base = gauche if x <= creux else droite
            for k in range(1, elevation + 1):
                niveau = (k - 1) * bandes // max(1, elevation)
                facteur = 1.0 - 0.18 * niveau
                px[x, y + k] = tuple(int(c * facteur) for c in base) + (255,)
                (masque_gauche if x <= creux else masque_droite).add((x, y + k))

    if arete and elevation:
        for x, y in haut.items():
            px[x, y] = clair + (255,)

    img.info["hauteur_dessus"] = hauteur
    # Emprise réelle, en tuiles de 64 : c'est elle que le chargeur marque dans la
    # grille de passabilité. Un volume bâti sur une tuile plus petite occupe
    # moins qu'une tuile entière, et un wagon en occupe trois.
    echelle = largeur_tuile / LARGEUR_TUILE
    img.info["emprise"] = (round(tx * echelle, 3), round(ty * echelle, 3))
    img.info["tx"] = tx
    img.info["ty"] = ty
    img.info["largeur_tuile"] = largeur_tuile
    img.info["dessus"] = masque_dessus
    img.info["gauche"] = masque_gauche
    img.info["droite"] = masque_droite
    return img


def empiler(*couches):
    largeur = max(c.width + dx for c, dx, _ in couches)
    hauteur = max(c.height + dy for c, _, dy in couches)
    img = Image.new("RGBA", (largeur, hauteur), TRANSPARENT)
    for couche, dx, dy in couches:
        img.alpha_composite(couche, (dx, hauteur - couche.height - dy))
    img.info["hauteur_dessus"] = couches[0][0].info["hauteur_dessus"]
    img.info["emprise"] = couches[0][0].info.get("emprise", (1.0, 1.0))
    for face in ("dessus", "gauche", "droite"):
        img.info[face] = set()
    return img


def joint(img, largeur_tuile=LARGEUR_TUILE):
    """Liseré sur les deux arêtes hautes seulement.

    Chaque tuile n'apporte que la moitié du joint : deux tuiles voisines
    donnent une ligne d'un pixel, pas deux.
    """
    px = img.load()
    largeur, _ = img.size
    for y in range(largeur_tuile // 4):
        ligne = [x for x in range(largeur) if px[x, y][3] > 0]
        if not ligne:
            continue
        for x in (ligne[0], ligne[0] + 1, ligne[-1] - 1, ligne[-1]):
            r, v, b, _ = px[x, y]
            px[x, y] = (int(r * 0.72), int(v * 0.72), int(b * 0.72), 255)
    return img





def reduire(img, couleurs=24):
    """Ramène une forme à un nombre borné de couleurs, alpha binaire conservé.

    Le grain et le contour mélangent des teintes et en fabriquent des dizaines ;
    sans cette passe, un seul objet peut dépasser la palette du projet.
    """
    alpha = img.getchannel("A").point(lambda v: 255 if v > 128 else 0)
    plat = img.convert("RGB").quantize(colors=couleurs, dither=Image.NONE).convert("RGB")
    plat.putalpha(alpha)
    for cle in ("hauteur_dessus", "dessus", "gauche", "droite", "tx", "ty",
                "largeur_tuile", "emprise"):
        if cle in img.info:
            plat.info[cle] = img.info[cle]
    return plat


def position(base, u, v):
    """Point de la face supérieure de base, en pixels, aux coordonnées (u, v)."""
    lt = base.info["largeur_tuile"]
    x = (u - v) * lt / 2 + base.info["ty"] * lt / 2
    y = (u + v) * lt / 4
    return round(x), round(y)


def poser(base, *objets):
    """Pose des objets sur la face supérieure, aux coordonnées de tuile données.

    Les objets suivent les axes du losange, pas ceux de l'écran : trois caisses
    alignées le sont dans le monde, pas seulement à l'image.
    """
    marge = max((o.height for o, _, _ in objets), default=0)
    canevas = Image.new("RGBA", (base.width, base.height + marge), TRANSPARENT)
    canevas.alpha_composite(base, (0, marge))

    for objet, u, v in sorted(objets, key=lambda t: t[1] + t[2]):
        # (u, v) désigne le centre de l'emprise de l'objet, pas son coin bas :
        # sans ce recentrage, tout ce qu'on pose part vers le fond à gauche.
        lt_o = objet.info.get("largeur_tuile", LARGEUR_TUILE)
        # l'objet peut être bâti sur une tuile plus petite que celle de la base :
        # son emprise doit être ramenée à l'échelle de la base avant tout calcul.
        echelle = lt_o / base.info["largeur_tuile"]
        tx_o = objet.info.get("tx", 0) * echelle
        ty_o = objet.info.get("ty", 0) * echelle

        # (u, v) désigne le centre de l'emprise de l'objet. Son point d'appui
        # dans sa propre image est le sommet bas du losange, en tx * lt / 2 —
        # pas au milieu de l'image dès que l'emprise n'est pas carrée.
        x, y = position(base, u + tx_o / 2, v + ty_o / 2)
        appui = round(tx_o * base.info["largeur_tuile"] / 2) if tx_o else objet.width // 2
        canevas.alpha_composite(objet, (x - appui, marge + y - objet.height + 1))

    boite = canevas.getbbox()
    resultat = canevas.crop(boite)
    resultat.info["hauteur_dessus"] = base.info["hauteur_dessus"]
    for cle in ("tx", "ty", "largeur_tuile", "emprise"):
        if cle in base.info:
            resultat.info[cle] = base.info[cle]
    for face in ("dessus", "gauche", "droite"):
        resultat.info[face] = set()
    return resultat


# --- Détails ---------------------------------------------------------------
# Ces fonctions posent ce qu'on ajouterait à la main : elles s'appliquent à un
# volume avant empilement, en s'appuyant sur les masques de faces.

def _melange(couleur, cible, force):
    return tuple(int(c + (t - c) * force) for c, t in zip(couleur, cible))


def grain(img, densite=0.10, force=0.16, graine=0, faces=("dessus",)):
    """Moucheture déterministe : casse l'aplat sans ajouter de couleur vive."""
    alea = random.Random(graine)
    px = img.load()
    for face in faces:
        for x, y in sorted(img.info.get(face, ())):
            if alea.random() >= densite:
                continue
            r, v, b, a = px[x, y]
            cible = (0, 0, 0) if alea.random() < 0.5 else (255, 255, 255)
            px[x, y] = _melange((r, v, b), cible, force) + (a,)
    return img


def nervures(img, pas=6, force=0.14):
    """Lignes verticales claires sur les flancs : tôle, montants, cannelures."""
    px = img.load()
    for face, sens in (("gauche", 255), ("droite", 0)):
        colonnes = {x for x, _ in img.info.get(face, ())}
        for x in sorted(colonnes):
            if x % pas:
                continue
            for xx, yy in sorted(img.info[face]):
                if xx == x:
                    r, v, b, a = px[xx, yy]
                    px[xx, yy] = _melange((r, v, b), (sens,) * 3, force) + (a,)
    return img


def bandeau(img, depuis_haut=3, epaisseur=2, couleur=(226, 226, 214)):
    """Réglette d'étiquettes sur le bord des flancs, juste sous la tablette."""
    px = img.load()
    for face in ("gauche", "droite"):
        colonnes = {}
        for x, y in img.info.get(face, ()):
            colonnes.setdefault(x, []).append(y)
        for x, ys in colonnes.items():
            ys.sort()
            for y in ys[depuis_haut:depuis_haut + epaisseur]:
                px[x, y] = couleur + (255,)
    return img


def rivets(img, pas=8, couleur=(70, 70, 76)):
    """Points sombres réguliers en haut des flancs."""
    px = img.load()
    for face in ("gauche", "droite"):
        colonnes = {}
        for x, y in img.info.get(face, ()):
            colonnes.setdefault(x, []).append(y)
        for x, ys in sorted(colonnes.items()):
            if x % pas:
                continue
            y = min(ys) + 1
            if (x, y) in img.info[face]:
                px[x, y] = couleur + (255,)
    return img


def contour(img, force=0.45, cible=(24, 24, 28)):
    """Tire le pourtour de la silhouette vers `cible`, un presque-noir par défaut.

    C'est le détail qui détache le plus un objet du fond en pixel art. À ne pas
    appliquer aux tuiles de sol : il doublerait le joint au raccord.

    **La cible se change pour ce qu'on doit voir arriver, et pour rien d'autre.**
    Un pourtour clair sur une masse sombre détache l'objet de tous les fonds à la
    fois — aucun ne peut être proche des deux —, là où un contour foncé ne le
    détache que des fonds clairs. Mesuré sur le tir de la Buse contre les cinq
    familles de décor : le pire cas passe de 40 à 81 de luminance d'écart.

    Ce n'est pas une option esthétique offerte à toutes les formes. Un signal qui
    vaut pour tout ne vaut plus rien : si les tirs du joueur le prenaient aussi,
    l'écran se remplirait d'objets qui crient et le liseré cesserait de dire
    « danger ». C'est la règle qui gouverne déjà la silhouette du rendu, qui ne
    révèle que le joueur et ses menaces.
    """
    px = img.load()
    l, h = img.size
    bord = []
    for y in range(h):
        for x in range(l):
            if px[x, y][3] == 0:
                continue
            for dx, dy in ((1, 0), (-1, 0), (0, 1), (0, -1)):
                vx, vy = x + dx, y + dy
                if not (0 <= vx < l and 0 <= vy < h) or px[vx, vy][3] == 0:
                    bord.append((x, y))
                    break
    for x, y in bord:
        r, v, b, a = px[x, y]
        px[x, y] = _melange((r, v, b), cible, force) + (a,)
    return img



def aligner(*volumes):
    """Compose plusieurs volumes qui partagent la même origine de tuile.

    Chaque volume place son point (0, 0) à une abscisse différente selon sa
    profondeur ; on les recale avant de composer, sinon un angle de mur se
    retrouve décalé d'une demi-tuile.
    """
    origines = [v.info["ty"] * v.info["largeur_tuile"] / 2 for v in volumes]
    ref = max(origines)
    decalages = [round(ref - o) for o in origines]
    largeur = max(d + v.width for d, v in zip(decalages, volumes))
    hauteur = max(v.height for v in volumes)

    img = Image.new("RGBA", (largeur, hauteur), TRANSPARENT)
    for decalage, v in sorted(zip(decalages, volumes), key=lambda c: -c[1].info["ty"]):
        img.alpha_composite(v, (decalage, 0))
    img.info["hauteur_dessus"] = min(v.info["hauteur_dessus"] for v in volumes)
    for face in ("dessus", "gauche", "droite"):
        img.info[face] = set()
    return img


def fenetres(img, pas=10, largeur=5, depuis_haut=6, hauteur=7, etage=0,
             vitre=(96, 132, 152), allumee=(228, 206, 132), graine=0, bornes=None):
    """Bande de fenêtres sur les flancs, une sur quatre éclairée."""
    alea = random.Random(graine)
    px = img.load()
    for face in ("gauche", "droite"):
        colonnes = {}
        for x, y in img.info.get(face, ()):
            colonnes.setdefault(x, []).append(y)
        for x, ys in sorted(colonnes.items()):
            if bornes and not bornes[0] <= x <= bornes[1]:
                continue
            if (x // largeur) % 2 or x % pas >= largeur:
                continue
            ys.sort()
            depart = depuis_haut + etage
            teinte = allumee if alea.random() < 0.25 else vitre
            for y in ys[depart:depart + hauteur]:
                px[x, y] = teinte + (255,)
    return img


def creuser(img, depuis_haut=4, hauteur=18, largeur_relative=0.34,
            couleur=(38, 38, 44)):
    """Renfoncement sombre au centre d'un flanc : porte, ouverture, vitrine."""
    px = img.load()
    for face in ("gauche", "droite"):
        colonnes = sorted({x for x, _ in img.info.get(face, ())})
        if not colonnes:
            continue
        debut = colonnes[0] + int(len(colonnes) * (1 - largeur_relative) / 2)
        fin = debut + int(len(colonnes) * largeur_relative)
        for x in colonnes:
            if not debut <= x <= fin:
                continue
            ys = sorted(y for xx, y in img.info[face] if xx == x)
            for y in ys[depuis_haut:depuis_haut + hauteur]:
                px[x, y] = couleur + (255,)
    return img


def tache(img, couleur, densite=0.30, graine=0):
    """Salissure plate sur la face supérieure, sans relief."""
    alea = random.Random(graine)
    px = img.load()
    for x, y in sorted(img.info.get("dessus", ())):
        if alea.random() < densite:
            r, v, b, a = px[x, y]
            px[x, y] = _melange((r, v, b), couleur, 0.45) + (a,)
    return img


def eventrer(img, densite=0.16, graine=0):
    """Entames aléatoires sur les arêtes : caisse cassée, mobilier abîmé."""
    alea = random.Random(graine)
    px = img.load()
    l, h = img.size
    for y in range(h):
        for x in range(l):
            if px[x, y][3] == 0:
                continue
            bord = any(not (0 <= x + dx < l and 0 <= y + dy < h) or px[x + dx, y + dy][3] == 0
                       for dx, dy in ((1, 0), (-1, 0), (0, 1), (0, -1)))
            if bord and alea.random() < densite:
                px[x, y] = (0, 0, 0, 0)
    return img



def carrelage(tuile, colonnes=8, rangees=8, echelle=3):
    l, h = tuile.size
    plan = Image.new("RGBA", (l * colonnes, h * rangees), TRANSPARENT)
    for ty in range(rangees * 2):
        for tx in range(colonnes * 2):
            plan.alpha_composite(tuile, ((tx - ty) * (l // 2) + l * colonnes // 2,
                                         (tx + ty) * (h // 2) - h * 2))
    return plan.resize((plan.width * echelle, plan.height * echelle), Image.NEAREST)


