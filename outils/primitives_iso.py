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
from pathlib import Path

from PIL import Image

LARGEUR_TUILE = 64
# Les textures carrées jointives, entrée versionnée des matières qu'un
# générateur de volumes ne sait pas produire. Elles vivent auprès des outils et
# non dans `assets/`, que `go:embed` embarque en entier : ce sont des sources,
# pas des ressources du jeu.
TEXTURES = Path(__file__).parent / "textures"
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


def couvre(img, emprise, largeur_tuile=LARGEUR_TUILE):
    """Dit si la forme peint entièrement le losange de sa case.

    C'est la question que le sol d'un thème existe pour résoudre, et elle se
    pose sur les pixels et non sur l'emprise. Les deux ont longtemps coïncidé
    parce qu'un volume peint tout son dessus : une emprise pleine valait un
    losange plein. Un marquage au sol les sépare — il occupe sa case, donc son
    emprise en vaut une, et n'en cache rien.

    Le losange se place là où le rendu pose l'image, c'est-à-dire en refaisant
    son calcul de coin : l'emprise se centre sur la case, si bien que le
    bas-centre d'une image ne retombe sur le sommet bas du losange que pour une
    emprise carrée. Le mesurer sans cela rendrait un comptoir de deux tuiles
    non couvrant, en cherchant son losange huit pixels trop bas.
    """
    points, _, _ = surface(1, 1, largeur_tuile)
    demi = largeur_tuile / 2
    ex, ey = emprise
    dx = round((ex - ey) * demi / 2)
    dy = round((2 + ex + ey) * (demi / 2) / 2)
    ax, ay = img.width // 2, img.height - 1

    px = img.load()
    for sx, sy in points:
        x, y = sx - largeur_tuile // 2 - dx + ax, sy - dy + 1 + ay
        if not (0 <= x < img.width and 0 <= y < img.height) or px[x, y][3] == 0:
            return False
    return True


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


def texture(nom, dossier=TEXTURES):
    """Rend une peinture qui lit une texture carrée jointive.

    C'est ce qui fait entrer une matière dans un volume sans rien lui retirer :
    la face supérieure prend le grain de la texture, les flancs et l'arête
    gardent la palette de leur matière. Un trottoir garde donc son relief, ce
    qu'une tuile plate dessinée ne saurait pas rendre.

    **La lecture est périodique**, en coordonnées de tuile : deux cases voisines
    lisent les bords opposés de la texture, comme deux carrés voisins d'un
    pavage carré, et le raccord d'une texture jointive survit à la projection.
    C'est aussi ce qui permet à une emprise de plusieurs tuiles de la répéter au
    lieu de l'étirer.

    Le fichier est une entrée versionnée, donc la forme reste reproductible à
    l'identique — ce qui la distingue d'une image tenue à la main, que rien ne
    régénère.
    """
    chemin = Path(dossier) / f"{nom}.png"
    with Image.open(chemin) as fichier:
        fichier.load()
        image = fichier.convert("RGB")
    px, largeur, hauteur = image.load(), image.width, image.height

    def peinture(u, v):
        return px[int(u * largeur) % largeur, int(v * hauteur) % hauteur]
    return peinture


def nuance(peinture, facteur=1.0, vers=None, force=0.0):
    """Décline une peinture sans changer son grain.

    Une même matière habille plusieurs revêtements — un béton neuf, un béton
    usé, un béton sali —, et ce qui les sépare est une valeur et une teinte, pas
    un dessin. Les décliner ici plutôt que de générer une texture par nuance
    garde un grain dont on sait qu'il se répète sans se voir : c'est la tache
    identifiable qui trahit un pavage, jamais le moucheté.
    """
    def teintee(u, v):
        r, v_, b = peinture(u, v)
        r, v_, b = int(r * facteur), int(v_ * facteur), int(b * facteur)
        if vers is not None:
            r, v_, b = (int(c + (t - c) * force) for c, t in zip((r, v_, b), vers))
        return min(255, max(0, r)), min(255, max(0, v_)), min(255, max(0, b))
    return teintee


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


def decalque(peinture, tx=1, ty=1, largeur_tuile=LARGEUR_TUILE, cerne=True):
    """Marque posée à plat sur une case, sans matière dessous.

    Le sol appartient au lieu et non à la marque : une bouche d'égout se pose
    sur du bitume comme sur du carrelage, et un losange plein lui ferait
    emporter le sien. `peinture` ne peint donc que ce qu'elle nomme, et le reste
    de la case reste transparent — c'est le thème qui le comble.

    **L'image garde le cadre de la case et ne se recadre pas.** Le rendu pose
    son bas-centre sur le sommet bas du losange : une marque rognée à son motif
    descendrait d'autant qu'on lui aurait retiré de creux.

    **`cerne` sépare deux natures qu'on confondrait.** Ce qu'on pose sur le sol
    — une grille de fonte, un marquage peint — se cerne, faute de quoi rien ne
    le détache d'un revêtement qu'il ne connaît pas. Ce qui *altère* le sol —
    une flaque, une tache — ne se cerne pas : un liseré en ferait un objet posé
    là, et c'est exactement ce qui fait lire une flaque comme une pastille.
    """
    points, largeur, hauteur = surface(tx, ty, largeur_tuile)
    img = Image.new("RGBA", (largeur, hauteur), TRANSPARENT)
    px = img.load()
    for (x, y), (u, v) in points.items():
        teinte = peinture(u, v)
        if teinte is not None:
            px[x, y] = teinte + (255,)

    img.info["hauteur_dessus"] = hauteur
    echelle = largeur_tuile / LARGEUR_TUILE
    img.info["emprise"] = (round(tx * echelle, 3), round(ty * echelle, 3))
    img.info["tx"] = tx
    img.info["ty"] = ty
    img.info["largeur_tuile"] = largeur_tuile
    img.info["decalque"] = True
    img.info["cerner"] = cerne
    for face in ("dessus", "gauche", "droite"):
        img.info[face] = set()
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
                "largeur_tuile", "emprise", "decalque", "cerner"):
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


def grener(img, peinture, faces=("gauche", "droite"), force=0.55):
    """Porte une matière sur les flancs d'un volume, sans toucher à leur ombre.

    **Un mur se voit par ses flancs, pas par son dessus** — soixante-quatre
    pixels d'élévation contre une bande de face supérieure —, et `volume` ne
    sait peindre que celle-ci. La texture ne peut donc pas y entrer comme elle
    entre dans un sol.

    Ce qui est appliqué est **l'écart de la texture à sa moyenne**, en facteur :
    l'ombrage par bandes du volume survit intact, et seul le grain s'ajoute. Le
    remplacement pur aurait aplati les deux flancs et le dégradé avec eux.

    La lecture se fait en pixels d'écran et non en coordonnées de tuile : un
    flanc est un plan vertical, que la projection ne parcourt pas. Le raccord
    entre deux cases voisines tient tant que la texture mesure la largeur de
    tuile ou un de ses diviseurs.
    """
    px = img.load()
    lire = peinture
    # Moyenne de la texture, prise une fois : c'est elle qui fait du grain un
    # écart plutôt qu'une teinte, donc ce qui laisse l'ombrage intact.
    echantillons = [lire(x / LARGEUR_TUILE, y / LARGEUR_TUILE)
                    for x in range(LARGEUR_TUILE) for y in range(LARGEUR_TUILE)]
    moyenne = sum(sum(c) / 3 for c in echantillons) / len(echantillons) or 1

    for face in faces:
        for x, y in sorted(img.info.get(face, ())):
            r, v, b, a = px[x, y]
            grise = sum(lire((x % LARGEUR_TUILE) / LARGEUR_TUILE,
                             (y % LARGEUR_TUILE) / LARGEUR_TUILE)) / 3
            facteur = 1.0 + (grise / moyenne - 1.0) * force
            px[x, y] = (min(255, max(0, int(r * facteur))),
                        min(255, max(0, int(v * facteur))),
                        min(255, max(0, int(b * facteur))), a)
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


def lames(img, pas=3, force=0.22):
    """Lignes sombres parallèles au bord haut des flancs : lames, bardage, store.

    Le rang se compte depuis le haut de chaque colonne, si bien que la ligne
    suit la pente de la face comme le fait son arête : c'est ce qui la lit
    horizontale en isométrie, là où une rangée de pixels à y fixe la couperait
    en biais.
    """
    px = img.load()
    for face in ("gauche", "droite"):
        colonnes = {}
        for x, y in img.info.get(face, ()):
            colonnes.setdefault(x, []).append(y)
        for x, ys in colonnes.items():
            ys.sort()
            for rang, y in enumerate(ys):
                if rang % pas == pas - 1:
                    r, v, b, a = px[x, y]
                    px[x, y] = _melange((r, v, b), (0, 0, 0), force) + (a,)
    return img


def bandeau(img, depuis_haut=3, epaisseur=2, couleur=(226, 226, 214), depuis_bas=None):
    """Réglette d'étiquettes sur le bord des flancs, juste sous la tablette.

    `depuis_bas` la pose au pied des flancs plutôt que sous leur arête haute :
    une plinthe, la barre d'un store.
    """
    px = img.load()
    for face in ("gauche", "droite"):
        colonnes = {}
        for x, y in img.info.get(face, ()):
            colonnes.setdefault(x, []).append(y)
        for x, ys in colonnes.items():
            ys.sort()
            if depuis_bas is None:
                prises = ys[depuis_haut:depuis_haut + epaisseur]
            else:
                prises = ys[max(0, len(ys) - depuis_bas - epaisseur):len(ys) - depuis_bas]
            for y in prises:
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


