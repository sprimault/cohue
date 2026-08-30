# Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
# SPDX-License-Identifier: Apache-2.0

"""Fabrique les personnages du jeu : des volumes composés, façon figurine.

    python outils/figurines.py --sortie assets/personnages
    python outils/figurines.py --apercu           # planches de contrôle

Ce sont les personnages définitifs du jeu, versionnés au même titre que le
décor. Ils se modifient dans ce fichier, ce qui les rend relisibles en pull
request là où un PNG ne l'est pas.

Le corps est bâti avec les primitives du décor : une créature est un empilement
de volumes isométriques, et rien d'autre. Six gabarits — bipède, quadrupède,
rampant, bulbe, colosse, gonflé — donnent des silhouettes distinctes ; recolorer
ne suffit pas, un joueur doit lire sa horde d'un coup d'œil.

Les huit orientations viennent de la place des membres et du regard, pas d'une
rotation des volumes : à cette taille un torse pivoté ne se lit pas, alors qu'un
bras avancé se lit tout de suite.
"""

import argparse
import json
import math
from pathlib import Path

from PIL import Image

import primitives_iso as prim
from manifestes import ecrire_manifeste

COTE = 64
APPUI = (COTE // 2, COTE - 1)
DIRECTIONS = ["S", "SO", "O", "NO", "N", "NE", "E", "SE"]


# Valeurs de jeu, par profil. Elles vivent ici plutôt que dans le code : un
# nouveau profil est une ligne de table, jamais une branche.
#
# `touches` est la résistance en touches de l'arme de base au premier niveau —
# une valeur absolue de PV ne voudrait rien dire dans un jeu où l'arme est
# multipliée par dix au cours d'une run.
#
# `vitesse_relative` est un rapport à celle du joueur, la seule vitesse absolue
# du jeu. Deux champs plutôt qu'un `vitesse` qui changerait d'unité selon le
# profil : c'est le chiffre le plus structurant du jeu — au-delà de 0,9 la fuite
# ne suffit plus, en dessous de 0,6 on ne peut plus être encerclé en terrain
# ouvert — et il ne doit pas se lire de travers.
#
# `rayon_tuiles` et non un rayon en pixels : la simulation ne connaît que la
# tuile, et une distance mesurée à l'écran décrit une ellipse dans le monde. La
# conversion se fait ici, une fois, plutôt que dans le chargeur.
#
# `role` distingue trois natures, et non deux : le joueur n'a ni touches ni
# points mais une vie et un plafond de dégâts, un ennemi a l'inverse, et le
# Passant n'a ni l'un ni l'autre. Un booléen `hostile` n'en portait que deux, le
# troisième cas se lisant dans son absence — le genre de défaut par défaut qu'on
# ne voit pas.
#
# `cout_pression` est ce que le spawner dépense pour l'acheter, `points` ce que
# le joueur gagne à le tuer : deux monnaies sans rapport. Le coût se calque sur
# la résistance, c'est-à-dire sur ce que la créature coûte au joueur, et jamais
# sur ce qu'elle lui rapporte. Il est unitaire, le spawner le multipliant par la
# taille du groupe.
#
# `max_simultane` borne les vivants d'un profil, pas les apparus. Un coût élevé
# règle une fréquence moyenne, pas une simultanéité : le Secouriste ne vaut rien
# seul et double la difficulté au milieu de vingt Badauds, donc sa rareté ne peut
# pas se régler par son prix. Un scénario peut resserrer ce plafond, jamais le
# desserrer.
JEU = {
    "joueur":    {"role": "joueur", "vitesse_tuiles_s": 5.0, "rayon_tuiles": 0.125,
                  "vie": 100, "plafond_degats_s": 20},
    "marcheur":  {"role": "ennemi", "comportement": "poursuite",
                  "vitesse_relative": 0.62, "rayon_tuiles": 0.125, "touches": 3,
                  "points": 10, "cout_pression": 3, "poids_separation": 1.0,
                  "max_simultane": 0, "degats_contact_s": 6},
    "sprinteur": {"role": "ennemi", "comportement": "charge",
                  "vitesse_relative": 1.35, "rayon_tuiles": 0.109, "touches": 2,
                  "points": 25, "cout_pression": 4, "poids_separation": 1.0,
                  "max_simultane": 0, "degats_contact_s": 8, "degats_charge": 18},
    "flanqueur": {"role": "ennemi", "comportement": "flanc",
                  "vitesse_relative": 0.82, "rayon_tuiles": 0.109, "touches": 4,
                  "points": 30, "cout_pression": 5, "poids_separation": 1.0,
                  "max_simultane": 0, "degats_contact_s": 7, "tangentiel": 0.55},
    "cracheur":  {"role": "ennemi", "comportement": "tir",
                  "vitesse_relative": 0.55, "rayon_tuiles": 0.125, "touches": 5,
                  "points": 40, "cout_pression": 6, "poids_separation": 1.3,
                  "max_simultane": 0, "degats_contact_s": 4, "portee_tuiles": 6},
    # Poids faible : dans le mécanisme du chapitre 4, ce poids dit combien une
    # créature s'écarte de ses voisines, et non combien elle résiste à être
    # poussée — personne n'est poussé, chacun s'applique la force à soi-même.
    # Un Vigile qui ignore la foule tient sa position, et c'est ce qui bouche.
    "bloqueur":  {"role": "ennemi", "comportement": "poursuite",
                  "vitesse_relative": 0.45, "rayon_tuiles": 0.188, "touches": 12,
                  "points": 60, "cout_pression": 12, "poids_separation": 0.4,
                  "max_simultane": 0, "degats_contact_s": 10},
    "eclateur":  {"role": "ennemi", "comportement": "explosion",
                  "vitesse_relative": 0.70, "rayon_tuiles": 0.141, "touches": 4,
                  "points": 35, "cout_pression": 5, "poids_separation": 1.0,
                  "max_simultane": 0, "degats_contact_s": 5, "degats_explosion": 35,
                  "rayon_explosion_tuiles": 1.5},
    # Un seul à la fois : sa menace est multiplicative, pas additive.
    "soigneur":  {"role": "ennemi", "comportement": "soin",
                  "vitesse_relative": 0.70, "rayon_tuiles": 0.109, "touches": 3,
                  "points": 15, "cout_pression": 6, "poids_separation": 1.0,
                  "max_simultane": 1, "degats_contact_s": 4},
    # Le Passant n'est pas un monstre : il ne blesse pas, ne rapporte rien, ne
    # se cible pas et ne se paie pas dans le budget de pression. Il traverse la
    # scène et occupe l'espace, ce qui suffit à le rendre gênant. Lui laisser
    # des dégâts de contact en ferait un Badaud affaibli, donc un doublon.
    #
    # `va_et_vient` dit ce qu'on voit, comme `poursuite` pour les autres ; le
    # moyen est un rebond — il avance tout droit jusqu'à buter, puis repart —
    # ce qui n'exige aucune trajectoire posée dans la pièce.
    #
    # Le rôle porte tout le reste : ce qui n'est pas hostile n'entre dans aucun
    # compte — ni le budget de pression, ni le plafond d'effectif, ni un objectif
    # de porte fondé sur les kills. Sans cette dernière conséquence, un lieu
    # peuplé de Passants ouvrirait sa porte tout seul.
    "civil":     {"role": "ambiance", "comportement": "va_et_vient",
                  "vitesse_relative": 0.75, "rayon_tuiles": 0.109},
}

CADENCES = {"repos": 200, "marche": 100, "attaque": 80, "degat": 120, "mort": 120}
BOUCLENT = {"repos", "marche"}

# Corps, teinte de vêtement, cycles. Le nombre d'images par cycle est ce que le
# manifeste annonce au moteur : le changer ici change l'animation en jeu.
PROFILS = {
    "joueur":    {"nom": "Survivant", "habit": (72, 108, 152),  "peau": (214, 170, 132), "carrure": 1.0,
                  "cycles": {"repos": 1, "attaque": 1, "marche": 4, "mort": 2}},
    "marcheur":  {"nom": "Badaud", "habit": (120, 148, 116), "peau": (128, 168, 116), "carrure": 1.0,
                  # Une foule d'un seul bleu se lit comme un bloc : les variantes
                  # cassent la répétition sans coûter une silhouette de plus.
                  "variantes": [(120, 148, 116), (92, 116, 140), (146, 122, 96),
                                (132, 108, 120), (104, 132, 128), (150, 140, 108)],
                  "cycles": {"repos": 1, "marche": 5, "attaque": 3, "degat": 1, "mort": 3}},
    "sprinteur": {"nom": "Molosse", "groupe": 3, "habit": (112, 88, 72),   "peau": (150, 118, 88),  "carrure": 1.0,
                  "gabarit": "quadrupede",
                  "cycles": {"marche": 8, "mort": 2}},
    "flanqueur": {"nom": "Arpenteur", "habit": (108, 76, 132),  "peau": (168, 132, 196), "carrure": 0.9,
                  "gabarit": "rampant",
                  "cycles": {"repos": 1, "attaque": 2, "marche": 5, "degat": 1, "mort": 3}},
    "cracheur":  {"nom": "Buse", "habit": (168, 140, 72),  "peau": (120, 100, 52),  "carrure": 0.95,
                  "gabarit": "bulbe",
                  "cycles": {"repos": 1, "attaque": 1, "marche": 4, "mort": 2}},
    "bloqueur":  {"nom": "Vigile", "habit": (96, 96, 104),   "peau": (140, 172, 128), "carrure": 1.35,
                  "gabarit": "colosse",
                  "cycles": {"repos": 1, "attaque": 2, "marche": 5, "degat": 1, "mort": 3}},
    "eclateur":  {"nom": "Baudruche", "habit": (176, 92, 52),   "peau": (222, 152, 96),  "carrure": 1.0,
                  "gabarit": "gonfle",
                  "cycles": {"repos": 1, "marche": 5, "attaque": 3, "degat": 1, "mort": 3}},
    # Le Secouriste est le seul profil qui crée une priorité de cible : tant
    # qu'il vit, il annule le nettoyage. Les six autres cassent le kiting par la
    # position, aucun ne rend un ennemi plus urgent qu'un autre.
    "soigneur":  {"nom": "Secouriste", "habit": (206, 206, 200), "peau": (214, 170, 132), "carrure": 0.95,
                  "cycles": {"repos": 1, "attaque": 1, "marche": 4, "mort": 2}},
    # Pas de cycle d'attaque : il n'en a pas, et lui en donner un serait la
    # première marche vers un profil de combat.
    "civil":     {"nom": "Passant", "habit": (176, 160, 142), "peau": (214, 170, 132), "carrure": 0.95,
                  "cycles": {"repos": 1, "marche": 4, "mort": 2}},
}

CHEVEUX = (58, 44, 38)


def _matiere(nom, teinte):
    """Enregistre une matière à la volée : dessus, deux flancs, arête claire."""
    def melange(cible, force):
        return tuple(round(c + (t - c) * force) for c, t in zip(teinte, cible))

    prim.MATIERES[nom] = (teinte, melange((0, 0, 0), 0.28), melange((0, 0, 0), 0.44),
                           melange((255, 255, 255), 0.30))
    return nom


def _bloc(largeur, hauteur, matiere, arrondi=True):
    """Une boîte isométrique, coins chanfreinés.

    Le chanfrein s'applique au bloc lui-même et pas seulement à la silhouette
    finale : sans lui, les angles internes — jonction bras-torse, sommet du
    crâne — restent taillés au carré alors que ce sont eux qu'on regarde.
    """
    bloc = prim.volume(elevation=hauteur, matiere=matiere, largeur_tuile=largeur,
                        bandes=2)
    if arrondi and largeur >= 10:
        _adoucir(bloc)
    return bloc


def _coller(fond, piece, cx, cy):
    """Colle une pièce en visant le milieu de sa base."""
    fond.alpha_composite(piece, (round(cx - piece.width / 2),
                                 round(cy - piece.height)))



def _adoucir(img, passes=1):
    """Chanfreine les coins convexes de la silhouette.

    Un empilement de boîtes n'a que des angles droits : épaules, crânes et
    caisses en gardent une allure taillée. Retirer le pixel des coins convexes
    les arrondit sans rien redessiner, et à 64 pixels une seule passe suffit.
    """
    for _ in range(passes):
        px = img.load()
        l, h = img.size
        opaque = [[px[x, y][3] > 0 for x in range(l)] for y in range(h)]

        def plein(x, y):
            return 0 <= x < l and 0 <= y < h and opaque[y][x]

        coins = []
        for y in range(h):
            for x in range(l):
                if not opaque[y][x]:
                    continue
                for dx, dy in ((-1, -1), (1, -1), (-1, 1), (1, 1)):
                    if not plein(x + dx, y) and not plein(x, y + dy):
                        coins.append((x, y))
                        break
        for x, y in coins:
            px[x, y] = (0, 0, 0, 0)
    return img


def _avant(lateral, profondeur, distance):
    """Vecteur écran vers l'avant du regard, en pixels.

    Tout ce qui doit se placer « devant » ou « derrière » passe par lui : c'est
    ce qui fait qu'un corps composé de blocs s'oriente dans les huit directions
    sans qu'aucun volume ne pivote.
    """
    return round(lateral * distance), round(-profondeur * distance * 0.45)


def _quadrupede(cycle, image, total, img, contexte):
    """Chien : deux blocs de corps enfilés sur l'axe du regard, quatre pattes.

    Un volume unique allongé ne marche pas : son axe long est celui de la tuile,
    pas celui du regard, et la tête se retrouvait à côté du corps dès que le
    chien tournait. Deux blocs posés le long du vecteur avant tiennent, eux,
    dans les huit orientations.
    """
    habit, peau, museau = contexte["matieres"]
    lateral, profondeur, de_dos = contexte["regard"]
    avancement, cx, sol = contexte["pose"]

    foulee = math.sin(avancement * 2 * math.pi)
    bond = round(2 * max(0.0, foulee))
    ecrase = round(11 * avancement) if cycle == "mort" else 0

    ombre = prim.volume(tx=0.5, ty=0.34, elevation=0, matiere="bitume", arete=False)
    ombre.putalpha(ombre.getchannel("A").point(lambda v: 80 if v else 0))
    _coller(img, ombre, cx, sol + ombre.height // 2)

    h_pattes = max(2, 8 - ecrase)
    for rang, (cote, avant) in enumerate(((-1, 1), (1, -1), (-1, -1), (1, 1))):
        phase = foulee if (cote * avant) > 0 else -foulee
        dx, dy = _avant(lateral, profondeur, avant * 7)
        patte = _bloc(5, max(2, h_pattes + round(2 * phase)), habit)
        _coller(img, patte, cx + dx + cote * 3, sol + dy)

    # Arrière puis avant : le tri en profondeur se fait à la main, l'avant du
    # chien étant toujours plus près de la caméra quand il descend l'écran.
    for cote, largeur, hauteur in ((-1, 15, 9), (1, 17, 10)):
        dx, dy = _avant(lateral, profondeur, cote * 5)
        bloc = _bloc(largeur, max(3, hauteur - ecrase), habit)
        _coller(img, bloc, cx + dx, sol - h_pattes + 5 + dy - bond)

    dxq, dyq = _avant(lateral, profondeur, -12)
    _coller(img, _bloc(4, 5, habit), cx + dxq, sol - h_pattes + 1 + dyq - bond)

    dxt, dyt = _avant(lateral, profondeur, 11)
    base_tete = sol - h_pattes - 2 + dyt - bond
    tete = _bloc(11, 8, peau)
    _coller(img, tete, cx + dxt, base_tete)
    sommet = base_tete - tete.height

    dxm, dym = _avant(lateral, profondeur, 16)
    _coller(img, _bloc(6, 4, museau), cx + dxm, base_tete - 1 + (dym - dyt))

    if not de_dos:
        oeil = Image.new("RGBA", (2, 2), (222, 74, 62, 255))
        ecarts = (round(lateral * 2),) if abs(lateral) > 0.8 else (-3, 2)
        for ecart in ecarts:
            img.alpha_composite(oeil, (cx + dxt + ecart, sommet + 6))

    return img


def _rampant(cycle, image, total, img, contexte):
    """Flanqueur : corps bas sur six pattes écartées, sans tête distincte.

    Une silhouette d'insecte se reconnaît de loin à ses pattes qui dépassent du
    corps — c'est le seul détail qui survit à cette taille, donc le seul qui
    mérite d'être dessiné.
    """
    habit, peau, _ = contexte["matieres"]
    lateral, profondeur, de_dos = contexte["regard"]
    avancement, cx, sol = contexte["pose"]

    battement = math.sin(avancement * 2 * math.pi)
    ecrase = round(9 * avancement) if cycle == "mort" else 0

    ombre = prim.volume(tx=0.46, ty=0.34, elevation=0, matiere="bitume", arete=False)
    ombre.putalpha(ombre.getchannel("A").point(lambda v: 80 if v else 0))
    _coller(img, ombre, cx, sol + ombre.height // 2)

    for cote in (-1, 1):
        for rang, distance in enumerate((-7, 0, 7)):
            phase = battement if (rang % 2 == 0) == (cote > 0) else -battement
            dx, dy = _avant(lateral, profondeur, distance)
            patte = _bloc(3, max(2, 11 + round(3 * phase) - ecrase), habit)
            _coller(img, patte, cx + dx + cote * 10, sol + dy)

    corps = prim.volume(tx=0.30, ty=0.24, elevation=max(3, 6 - ecrase),
                         matiere=habit, bandes=2)
    _coller(img, corps, cx, sol - 9)
    haut = sol - 9 - corps.height

    dxt, dyt = _avant(lateral, profondeur, 9)
    _coller(img, _bloc(8, 4, peau), cx + dxt, sol - 10 + dyt)

    if not de_dos:
        oeil = Image.new("RGBA", (2, 2), (236, 214, 96, 255))
        for ecart in (-3, 0, 3):
            img.alpha_composite(oeil, (cx + dxt + ecart, haut + 6))

    return img


def _bulbe(cycle, image, total, img, contexte):
    """Cracheur : masse ronde posée au sol, sans jambes, avec un canon à l'avant.

    Il ne se déplace pas comme les autres — il glisse — donc il n'a ni pattes ni
    balancement, seulement une respiration verticale. C'est ce qui le rend
    lisible même immobile au milieu d'une horde.
    """
    habit, peau, _ = contexte["matieres"]
    lateral, profondeur, de_dos = contexte["regard"]
    avancement, cx, sol = contexte["pose"]

    souffle = round(2 * math.sin(avancement * 2 * math.pi))
    ecrase = round(12 * avancement) if cycle == "mort" else 0

    ombre = prim.volume(tx=0.42, ty=0.38, elevation=0, matiere="bitume", arete=False)
    ombre.putalpha(ombre.getchannel("A").point(lambda v: 80 if v else 0))
    _coller(img, ombre, cx, sol + ombre.height // 2)

    socle = prim.volume(tx=0.40, ty=0.36, elevation=max(2, 5 - ecrase),
                         matiere=habit, bandes=2)
    _coller(img, socle, cx, sol)

    corps = prim.volume(tx=0.28, ty=0.24, elevation=max(3, 15 - ecrase - souffle),
                         matiere=habit, bandes=3)
    _coller(img, corps, cx, sol - 4)
    haut = sol - 4 - corps.height

    dxc, dyc = _avant(lateral, profondeur, 11)
    _coller(img, _bloc(6, 4, peau), cx + dxc, haut + 12 + dyc)

    if not de_dos:
        oeil = Image.new("RGBA", (3, 3), (240, 172, 84, 255))
        img.alpha_composite(oeil, (cx + round(lateral * 3) - 1, haut + 5))

    return img


def _gonfle(cycle, image, total, img, contexte):
    """Éclateur : sphère instable sur deux pattes courtes, prête à céder.

    Sa silhouette doit dire « ne t'approche pas » avant même que le télégraphe
    de mort ne se déclenche : d'où le corps disproportionné et la tête minuscule.
    """
    habit, peau, _ = contexte["matieres"]
    lateral, profondeur, de_dos = contexte["regard"]
    avancement, cx, sol = contexte["pose"]

    palpite = round(1.5 * math.sin(avancement * 2 * math.pi))
    ecrase = round(14 * avancement) if cycle == "mort" else 0

    ombre = prim.volume(tx=0.44, ty=0.36, elevation=0, matiere="bitume", arete=False)
    ombre.putalpha(ombre.getchannel("A").point(lambda v: 80 if v else 0))
    _coller(img, ombre, cx, sol + ombre.height // 2)

    for cote in (-1, 1):
        _coller(img, _bloc(6, max(2, 8 - ecrase), habit), cx + cote * 7, sol)

    corps = prim.volume(tx=0.42 + palpite * 0.01, ty=0.36,
                         elevation=max(4, 12 - ecrase), matiere=habit, bandes=3)
    _coller(img, corps, cx, sol - 7)
    haut = sol - 7 - corps.height

    tete = _bloc(8, 5, peau)
    dxt, dyt = _avant(lateral, profondeur, 3)
    _coller(img, tete, cx + dxt, haut + 8 + dyt)

    if not de_dos:
        oeil = Image.new("RGBA", (2, 2), (26, 26, 30, 255))
        for ecart in ((round(lateral * 2),) if abs(lateral) > 0.8 else (-2, 1)):
            img.alpha_composite(oeil, (cx + dxt + ecart, haut + 5))

    return img


def _colosse_epaules(img, contexte, h_torse, sol, h_jambes):
    """Épaulières du bloqueur : une masse au-dessus du torse, pas un bipède plus gros."""
    habit = contexte["matieres"][0]
    epauliere = prim.volume(tx=0.54, ty=0.40, elevation=6, matiere=habit, bandes=2)
    _coller(img, epauliere, contexte["pose"][1], sol - h_jambes - h_torse + 12)
    return img


GABARITS = {"quadrupede": _quadrupede, "rampant": _rampant,
            "bulbe": _bulbe, "gonfle": _gonfle}


def _elan(image, total):
    """Amplitude du coup, image par image : appel, frappe, maintien.

    Une courbe en sinus sur la durée du cycle vaut zéro à ses deux extrémités :
    sur deux images, l'attaque ne bougeait donc pas du tout. Un cycle non bouclé
    se décrit par ses poses clés, pas par une fonction continue.
    """
    if total <= 1:
        return 6
    if total == 2:
        return (-3, 7)[image]
    return (-3, 7) [image] if image < 2 else 5


def figurine(profil, direction, cycle, image, total, variante=0):
    """Une pose : jambes, torse, tête, deux bras, orientés par la direction."""
    reglages = PROFILS[profil]
    teintes = reglages.get("variantes", [reglages["habit"]])
    teinte_habit = teintes[variante % len(teintes)]
    habit = _matiere(f"_habit_{profil}_{variante % len(teintes)}", teinte_habit)
    peau = _matiere(f"_peau_{profil}", reglages["peau"])
    cheveux = _matiere("_cheveux", CHEVEUX)
    carrure = reglages["carrure"]

    avancement = image / max(1, total - 1) if total > 1 else 0.0
    angle = math.radians(DIRECTIONS.index(direction) * 45 + 90)
    lateral = math.cos(angle)                # +1 vers la droite de l'écran
    profondeur = -math.sin(angle)            # +1 vers le fond, donc de dos
    de_dos = profondeur > 0.3

    balance = math.sin(avancement * 2 * math.pi)
    pas = 3 * balance if cycle == "marche" else 0.0
    tassement = 0
    allonge = 0

    if cycle == "attaque":
        allonge = _elan(image, total)
    elif cycle == "degat":
        tassement = 3
    elif cycle == "mort":
        tassement = round(22 * avancement)

    img = Image.new("RGBA", (COTE, COTE), (0, 0, 0, 0))
    cx = APPUI[0]
    sol = APPUI[1] - 1

    contexte = {"matieres": (habit, peau, cheveux),
                "regard": (lateral, profondeur, de_dos),
                "pose": (avancement, cx, sol)}

    gabarit = reglages.get("gabarit", "bipede")
    if gabarit in GABARITS:
        img = GABARITS[gabarit](cycle, image, total, img, contexte)
        return prim.reduire(prim.contour(_adoucir(img), force=0.40), couleurs=20)

    ombre = prim.volume(tx=0.30 * carrure, ty=0.30, elevation=0, matiere="bitume",
                         arete=False)
    ombre.putalpha(ombre.getchannel("A").point(lambda v: 80 if v else 0))
    _coller(img, ombre, cx, sol + ombre.height // 2)

    # Proportions, mesurées depuis le sol. Le tassement écrase le personnage sur
    # place plutôt que de le faire descendre : c'est ce qui rend une mort lisible
    # à cette taille.
    h_jambes = max(3, 13 - tassement)
    h_torse = max(4, 16 - tassement)
    l_torse = round(17 * carrure)

    # Jambes : un écart franc, sinon elles se lisent comme un socle unique.
    for cote in (-1, 1):
        avance = round(cote * pas)
        jambe = _bloc(round(7 * carrure), h_jambes, habit)
        _coller(img, jambe, cx + cote * round(5 * carrure) + avance, sol)

    epaules = sol - h_jambes - h_torse

    def poser_torse():
        _coller(img, _bloc(l_torse, h_torse, habit),
                cx + round(allonge * 0.5 * lateral) + allonge // 3,
                sol - h_jambes + 2 - max(0, allonge) // 3)

    def poser_bras():
        for cote in (-1, 1):
            devant = lateral * cote > 0
            avance = round(lateral * cote * 3)
            oscille = round(-cote * pas * 0.8)
            # Le bras qui frappe part en avant et monte ; l'autre recule à
            # peine. Les décalages restent petits : au-delà de quatre pixels le
            # bras se détache du torse et flotte à côté du corps.
            frappe = max(-2, min(4, round(allonge * (0.45 if devant else -0.15))))
            bras = _bloc(round(5 * carrure), max(3, h_torse - 6), habit)
            _coller(img, bras,
                    cx + cote * round(l_torse * 0.34) + avance + frappe,
                    sol - h_jambes + 1 + oscille - max(0, allonge) // 3)

    # De dos, les bras passent derrière le torse : c'est le seul indice
    # d'orientation qui survive quand le visage n'est plus visible.
    if de_dos:
        poser_bras()
        poser_torse()
    else:
        poser_torse()
        poser_bras()

    if gabarit == "colosse":
        _colosse_epaules(img, contexte, h_torse, sol, h_jambes)

    # Tête posée sur les épaules, décalée vers l'avant du regard. Le repère
    # vertical se prend sur l'image collée : un volume est plus haut que son
    # élévation, du losange de sa face supérieure.
    penche = round(lateral * 2)
    largeur_tete = round(11 * carrure)
    tete = _bloc(largeur_tete, 9, peau)
    base_tete = epaules + 1
    _coller(img, tete, cx + penche + round(allonge * 0.4 * lateral),
            base_tete - max(0, allonge) // 4)

    sommet = base_tete - tete.height
    visage = sommet + largeur_tete // 2 - 1      # juste sous le losange du crâne

    calotte = _bloc(largeur_tete + 2, 2, cheveux)
    _coller(img, calotte, cx + penche + round(allonge * 0.4 * lateral), visage + 1)

    if de_dos:
        # Nuque : sans elle, une figurine de dos ressemble à une figurine de
        # face à qui il manque les yeux.
        nuque = _bloc(largeur_tete, 4, cheveux)
        _coller(img, nuque, cx + penche + allonge // 4, base_tete - 1)
    else:
        oeil = Image.new("RGBA", (2, 2), (26, 26, 30, 255))
        base = cx + penche + round(lateral * 3)
        # De profil, un seul œil est visible ; deux donneraient un visage plat.
        ecarts = (round(lateral * 2),) if abs(lateral) > 0.8 else (-3, 2)
        for ecart in ecarts:
            img.alpha_composite(oeil, (base + ecart, visage + 2))

    if cycle == "mort" and avancement > 0:
        # Une figurine qui rapetisse sur place ne se lit pas comme une mort :
        # on l'écrase vers le bas et on l'étale, comme une chute.
        etale = 1.0 + 0.5 * avancement
        ecrase = 1.0 - 0.45 * avancement
        couche = img.resize((round(COTE * etale), max(1, round(COTE * ecrase))),
                            Image.NEAREST)
        img = Image.new("RGBA", (COTE, COTE), (0, 0, 0, 0))
        img.alpha_composite(couche, ((COTE - couche.width) // 2,
                                     APPUI[1] - couche.height + 1))

    return prim.reduire(prim.contour(_adoucir(img), force=0.40), couleurs=20)


def sens_bras(cote):
    """Les deux bras oscillent en opposition, comme les jambes."""
    return cote


def bande(profil, direction, cycle, images, variante=0):
    planche = Image.new("RGBA", (COTE * images, COTE), (0, 0, 0, 0))
    for i in range(images):
        planche.alpha_composite(
            figurine(profil, direction, cycle, i, images, variante), (i * COTE, 0))
    return planche


def apercu(profil, echelle=4):
    """Les huit directions au repos, puis un cycle de marche."""
    lignes = [
        [figurine(profil, d, "repos", 0, 1) for d in DIRECTIONS],
        [figurine(profil, "SE", "marche", i, 5) for i in range(5)],
        [figurine(profil, "SE", "attaque", i, 3) for i in range(3)],
        [figurine(profil, "SE", "mort", i, 3) for i in range(3)],
    ]
    largeur = max(len(l) for l in lignes) * COTE
    planche = Image.new("RGBA", (largeur, len(lignes) * COTE), (26, 26, 32, 255))
    for j, ligne in enumerate(lignes):
        for i, im in enumerate(ligne):
            planche.alpha_composite(im, (i * COTE, j * COTE))
    return planche.resize((planche.width * echelle, planche.height * echelle),
                          Image.NEAREST)


def main():
    a = argparse.ArgumentParser(description=__doc__)
    a.add_argument("--sortie", type=Path, default=Path("assets/personnages"))
    a.add_argument("--apercu", action="store_true")
    # Voir `decor_iso.py` : ce qui sert à relire ne va pas dans ce qui est livré.
    a.add_argument("--controles", default=Path(".tmp/controle"), type=Path)
    o = a.parse_args()

    if o.apercu:
        o.controles.mkdir(parents=True, exist_ok=True)
        for profil in PROFILS:
            apercu(profil).save(o.controles / f"apercu_{profil}.png")
            print("aperçu:", profil)
        return

    manifeste = {}
    for profil, reglages in PROFILS.items():
        variantes = reglages.get("variantes", [reglages["habit"]])
        for indice in range(len(variantes)):
            # Une seule teinte : pas de sous-dossier, le chargeur n'a rien à
            # savoir des variantes qui n'existent pas.
            dossier = o.sortie / profil
            if len(variantes) > 1:
                dossier = dossier / f"v{indice}"
            dossier.mkdir(parents=True, exist_ok=True)
            for cycle, images in reglages["cycles"].items():
                for direction in DIRECTIONS:
                    bande(profil, direction, cycle, images, indice).save(
                        dossier / f"{cycle}_{direction}.png")
        manifeste[profil] = {
            "nom": reglages["nom"],
            "groupe": reglages.get("groupe", 1),
            "variantes": len(variantes),
            "gabarit": reglages.get("gabarit", "bipede"),
            **JEU.get(profil, {}),
            "origine": "figurine générée",
            "cote": COTE,
            "appui": list(APPUI),
            "directions": DIRECTIONS,
            "cycles": {c: {"images": n, "duree_ms": CADENCES.get(c, 120),
                           "boucle": c in BOUCLENT}
                       for c, n in reglages["cycles"].items()},
        }
        total = sum(reglages["cycles"].values()) * 8 * len(variantes)
        suffixe = f"  {len(variantes)} variantes" if len(variantes) > 1 else ""
        print(f"{profil:11} {total:4} images{suffixe}")

    ecrire_manifeste(o.sortie / "manifeste.json", "figurines.py",
                     {"version_format": 1, "profils": manifeste})
    print(f"\n{len(manifeste)} profils générés dans {o.sortie}")


if __name__ == "__main__":
    main()
