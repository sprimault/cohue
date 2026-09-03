# Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
# SPDX-License-Identifier: Apache-2.0

"""Rastérise la police en planche de glyphes et dessine l'icône de fenêtre.

    python outils/interface.py                     -> dans ./assets/interface
    python outils/interface.py --sortie /tmp/a     -> ailleurs

La police est une ressource tierce : `assets/polices/` porte ce qu'on a reçu,
`assets/interface/` ce qu'on en fabrique. Le jeu ne lit jamais le `.ttf` — le
rastériser à l'exécution mettrait l'apparence du texte sous la version d'une
bibliothèque, alors qu'une planche générée se compare comme les autres images.

La rastérisation est en 1 bit à la taille native : une fonte pixel y rend sa
grille d'origine, sans anticrénelage ni hinting à faire varier. C'est ce qui la
rend reproductible d'un système à l'autre.
"""

import argparse
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont

from manifestes import ecrire_manifeste

RACINE = Path(__file__).resolve().parent.parent
POLICE = RACINE / "assets" / "polices" / "PixelOperator8.ttf"

# La taille native de la fonte. Les tailles d'affichage admises en sont les
# multiples entiers : à 16 chaque pixel devient un bloc de 2×2 et la grille
# tient, à 12 elle se casse et le rendu redevient une interpolation.
TAILLE = 8

# Un pixel libre au-dessus de l'ascent. Les capitales accentuées y débordent —
# la fonte les dessine, sa hauteur nominale ne les loge pas — et une cellule
# calée sur ascent + descent les tronquerait en silence.
MARGE_ACCENT = 1

# La chaîne fait la déclaration : l'ordre des glyphes dans la planche est celui
# de ces caractères, et le manifeste la recopie telle quelle. Une plage — « ASCII
# 32 à 126 plus les accents » — serait une seconde description de ce que la
# planche contient, fausse au premier glyphe ajouté.
#
# Elle porte ce qu'un usage réclame et rien de plus : le tiret cadratin parce que
# l'écran de choix écrit « Niveau 5 — choisissez une amélioration », l'espace
# insécable parce qu'elle sépare les milliers et manquerait au premier score à
# quatre chiffres. L'espace fine insécable n'y est pas : la fonte ne la dessine
# pas, et la déclarer poserait un rectangle à sa place.
#
# Les caractères invisibles s'écrivent en échappement. Une espace insécable posée
# telle quelle se confond avec une espace ordinaire dans le source, et un éditeur
# la normalise sans que personne ne le voie.
GLYPHES = (
    " !\"#$%&'()*+,-./0123456789:;<=>?@"
    "ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`"
    "abcdefghijklmnopqrstuvwxyz{|}~"
    "àâäçéèêëîïôöùûüÿœæ"
    "ÀÂÄÇÉÈÊËÎÏÔÖÙÛÜŸŒÆ"
    "«»°— "
)

# Un caractère de la zone à usage privé, qu'aucune fonte ne mappe. Il sert de
# témoin : tous les caractères absents rendent le même dessin, `.notdef`.
TEMOIN = ""

# Ce que l'interface règle, et que le rendu ne code pas.
#
# **Aucune image ne sort d'ici.** Une jauge, un cadre, une case sont des
# rectangles unis dont la taille dépend de leur contenu : une image fixe devrait
# être étirée, ce qui casserait le pixel entier, et un découpage en neuf morceaux
# n'a d'intérêt que si le cadre porte un motif — biseau, rivets — qu'aucune
# décision n'a demandé. Ce qui reste à sortir du code est donc l'apparence seule.
#
# **Des faits d'apparence, jamais des dimensions calculées.** L'épaisseur d'un
# bord et une marge sont des réglages ; la largeur d'une carte dépend de son
# texte et le côté d'une case de la taille de son icône, si bien que les déclarer
# ici en ferait une seconde description de ce que le contenu impose.
#
# Elles vivent dans ce script plutôt que dans un fichier tenu à la main parce que
# le régler ne coûte qu'une régénération d'une seconde — la police est le seul
# dessin de ce générateur. Le manifeste des armes, lui, se tient à la main parce
# que le sien coûterait six cents images.
REGLAGES = {
    "bord_px": 1,
    "marge_px": 4,
    "hauteur_jauge_px": 6,
    # Les teintes sont en RVBA. Le fond d'un cadre et celui d'un bandeau sont
    # translucides : posés sur le décor, ils doivent laisser deviner ce qu'ils
    # couvrent, sans quoi une carte de choix masquerait la horde qui approche.
    "teintes": {
        "cadre_fond": [26, 28, 34, 205],
        "cadre_bord": [92, 96, 106, 255],
        # Le bord de la carte sur laquelle porte le choix, dans l'ambre qui
        # designe deja la valeur qui change : le meme sens, applique au cadre.
        "cadre_choisi": [236, 196, 96, 255],
        "bandeau_fond": [16, 17, 21, 200],
        "jauge_fond": [40, 42, 48, 255],
        "jauge_vie": [176, 62, 58, 255],
        # Le voile que le contact pose sur l'écran, dans le rouge de la jauge de
        # vie : c'est la même chose qu'il annonce, et une seconde teinte
        # obligerait à relier les deux de tête. Son alpha est le plus faible du
        # catalogue — il couvre toute la surface, là où une jauge occupe cent
        # cinquante pixels, et la conception veut voir la horde pendant qu'on
        # encaisse.
        #
        # **Une teinte translucide s'écrit prémultipliée par son alpha**, comme
        # les deux fonds ci-dessus : le rendu la passe telle quelle à Ebitengine,
        # qui l'attend sous cette forme. Écrit à plat, ce rouge donnait un aplat
        # trois fois trop dense — 176 par-dessus un décor à 86, alors qu'un quart
        # d'opacité en demande 44. Aucun contrôle ne l'aurait dit : seule la
        # planche de mêlée le montre.
        "voile_degat": [44, 16, 15, 64],
        "jauge_experience": [86, 132, 186, 255],
        "texte": [232, 234, 238, 255],
        "texte_attenue": [150, 154, 162, 255],
        "texte_valeur": [236, 196, 96, 255],
        "texte_contour": [0, 0, 0, 255],
    },
}


# L'icône de fenêtre se dessine sur une grille de cases, jamais en pixels.
#
# Chaque case devient un carré de côté entier, si bien que les trois tailles
# portent le même dessin sans qu'aucune soit l'agrandissement d'une autre : une
# mise à l'échelle interpolerait, et le pixel entier tombe avec elle.
GRILLE_ICONE = 16

# Les tailles parmi lesquelles le gestionnaire de fenêtres choisit selon
# l'endroit où il l'affiche. Toutes multiples de la grille, faute de quoi une
# case ne tomberait pas sur un compte entier de pixels.
TAILLES_ICONE = (16, 32, 48)

# **La palette de l'icône est la sienne, et ne descend pas du rendu.**
#
# Les aplats de `internal/render` sont provisoires et disparaîtront avec les
# sprites ; les recopier ici ferait deux descriptions d'une même couleur, dont
# l'une survivrait à l'autre sans que rien ne les confronte. Une icône est une
# composition autonome — ce qu'elle emprunte au jeu est son sujet, pas ses
# teintes.
#
# Le liseré n'est pas décoratif : sans lui, un carré presque noir se fond dans
# une barre des tâches sombre, et l'icône disparaît sur la moitié des postes.
PALETTE_ICONE = {
    "fond": (20, 22, 30, 255),
    "bord": (58, 62, 74, 255),
    "joueur": (240, 208, 110, 255),
    "horde": (168, 70, 66, 255),
}

# Les assaillants, par le coin de leur bloc de deux cases sur deux.
#
# Six et non quatre : posés aux seuls coins, ils se lisent comme un cadre. Ce
# qu'il faut donner à voir est un encerclement, donc une répartition sans
# échappée.
HORDE_ICONE = ((2, 2), (12, 2), (2, 12), (12, 12), (7, 1), (7, 13))

# Le joueur, ligne par ligne : ordonnée, abscisse de départ, largeur.
#
# Un losange à pente d'une case, et non la projection 2:1 du jeu : à seize
# pixels, celle-ci rendrait une barre de trois de haut qu'on ne lit plus comme
# une figure. L'icône dit le sujet, elle ne reproduit pas la caméra.
JOUEUR_ICONE = ((5, 7, 2), (6, 6, 4), (7, 5, 6), (8, 4, 8),
                (9, 5, 6), (10, 6, 4), (11, 7, 2))


def icone(taille):
    """Rend l'icône de fenêtre à une taille, en carrés de côté entier."""
    cote = taille // GRILLE_ICONE
    image = Image.new("RGBA", (taille, taille), PALETTE_ICONE["fond"])
    dessin = ImageDraw.Draw(image)

    def bloc(x, y, largeur, hauteur, teinte):
        dessin.rectangle((x * cote, y * cote,
                          (x + largeur) * cote - 1, (y + hauteur) * cote - 1),
                         fill=teinte)

    dernier = GRILLE_ICONE - 1
    for i in range(GRILLE_ICONE):
        bloc(i, 0, 1, 1, PALETTE_ICONE["bord"])
        bloc(i, dernier, 1, 1, PALETTE_ICONE["bord"])
        bloc(0, i, 1, 1, PALETTE_ICONE["bord"])
        bloc(dernier, i, 1, 1, PALETTE_ICONE["bord"])

    for x, y in HORDE_ICONE:
        bloc(x, y, 2, 2, PALETTE_ICONE["horde"])
    for y, x, largeur in JOUEUR_ICONE:
        bloc(x, y, largeur, 1, PALETTE_ICONE["joueur"])
    return image


def absents(police):
    """Rend les glyphes que la fonte ne dessine pas.

    Un caractère absent ne lève rien : FreeType rend `.notdef`, un rectangle
    plein, avec une avance plausible. Sans ce contrôle, l'ajouter à la table
    graverait la boîte dans la planche et le manifeste la déclarerait comme un
    glyphe ordinaire — le défaut ne se verrait qu'à l'écran, le jour où le texte
    qui l'emploie s'affiche.
    """
    boite = bytes(police.getmask(TEMOIN, mode="1"))
    return [c for c in GLYPHES if bytes(police.getmask(c, mode="1")) == boite]


def planche(police):
    """Rend la planche et ses métriques : cellule, ligne de base, avances.

    Les avances sont une liste parallèle à `GLYPHES` plutôt qu'un dictionnaire :
    l'ordre y est celui de la planche, et un décalage se voit sur la longueur. Un
    dictionnaire aurait porté la même information en la détachant de l'ordre
    qu'elle décrit.
    """
    ascent, descent = police.getmetrics()
    largeur = max(int(police.getlength(c)) for c in GLYPHES)
    hauteur = MARGE_ACCENT + ascent + descent

    masque = Image.new("1", (largeur * len(GLYPHES), hauteur), 0)
    dessin = ImageDraw.Draw(masque, "1")
    for i, c in enumerate(GLYPHES):
        dessin.text((i * largeur, MARGE_ACCENT), c, font=police, fill=1)

    # Blanc opaque sur transparent : le moteur teinte au dessin, et le contrôle
    # des ressources exige un alpha binaire.
    image = Image.new("RGBA", masque.size, (255, 255, 255, 0))
    image.putalpha(masque.convert("L").point(lambda v: 255 if v else 0))

    avances = [int(police.getlength(c)) for c in GLYPHES]
    return image, (largeur, hauteur), MARGE_ACCENT + ascent, avances


def main():
    a = argparse.ArgumentParser(description=__doc__)
    a.add_argument("--sortie", type=Path, default=Path("assets/interface"))
    o = a.parse_args()

    # Sans la fonte il n'y a pas de planche à produire, et une planche vide
    # passerait les contrôles en ne portant aucun glyphe. Un contrôle privé de
    # son entrée échoue, il ne passe pas.
    if not POLICE.exists():
        raise SystemExit(f"police introuvable : {POLICE}")

    police = ImageFont.truetype(str(POLICE), TAILLE)
    if manquants := absents(police):
        details = ", ".join(f"U+{ord(c):04X}" for c in manquants)
        raise SystemExit(f"{POLICE.name} ne dessine pas : {details}")

    o.sortie.mkdir(parents=True, exist_ok=True)
    image, cellule, ligne_de_base, avances = planche(police)
    image.save(o.sortie / "police.png")

    icones = [f"icone_{taille}.png" for taille in TAILLES_ICONE]
    for taille, fichier in zip(TAILLES_ICONE, icones):
        icone(taille).save(o.sortie / fichier)

    ecrire_manifeste(o.sortie / "manifeste.json", "interface.py", {
        "version_format": 1,
        "interface": {
            "police": {
                "fichier": "police.png",
                "source": POLICE.name,
                "cellule": list(cellule),
                "ligne_de_base": ligne_de_base,
                "taille_native": TAILLE,
                "glyphes": GLYPHES,
                "avances": avances,
            },
            # Les fichiers seuls : la taille de chacun se lit dans son image, et
            # la déclarer ici en ferait une seconde description que rien ne
            # confronterait au dessin.
            "icone": {"fichiers": icones},
            "reglages": REGLAGES,
        },
    })
    print(f"police    {len(GLYPHES)} glyphes, cellule {cellule[0]}x{cellule[1]},"
          f" ligne de base {ligne_de_base}")
    print(f"icone     {len(icones)} tailles, "
          + "x".join(str(t) for t in TAILLES_ICONE))


if __name__ == "__main__":
    main()
