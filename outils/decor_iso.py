# Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
# SPDX-License-Identifier: Apache-2.0

"""Génère le décor isométrique du jeu, thème par thème.

    python decor_iso.py                   -> tout, dans ./assets
    python decor_iso.py --theme parking   -> un seul thème
    python decor_iso.py voiture gondole   -> des formes précises
    python decor_iso.py --liste

Les primitives — volume, composition, détails, contrôle — sont dans
`primitives_iso.py`. Ce module ne porte que des formes : ce qui existe dans les
lieux du jeu, et rien d'autre.

L'ombrage est fait en bandes franches, pas en dégradé continu : deux ou trois
teintes par flanc, la plus sombre en bas. Un dégradé lissé produirait des
centaines de couleurs et trahirait le pixel art.
"""

import argparse
import json
import math
from pathlib import Path

from PIL import Image

from manifestes import ecrire_manifeste
from primitives_iso import (LARGEUR_TUILE, MATIERES, TRANSPARENT, aligner,
                            bandeau, carrelage, contour, creuser, empiler,
                            eventrer, fenetres, grain, joint, nervures,
                            position, poser, reduire, rivets, surface, tache,
                            volume)

HAUTEUR_PERSONNAGE = 64
PLAFOND_OBSTACLE_BAS = 24

MARQUAGE = (232, 232, 224)
BANDE_JAUNE = (222, 186, 74)

# Ce qui se traverse. L'élévation ne suffit pas à le déduire : un trottoir et un
# quai dépassent du sol et se marchent, une flaque est plate et se traverse,
# alors qu'un muret de même hauteur qu'un trottoir arrête tout.
FRANCHISSABLES = {
    "sol", "sol_use", "sol_carrele", "sol_fissure", "sol_sale", "sol_parking",
    "flaque", "bouche_egout", "fleche_sol", "trottoir", "quai", "rail",
    "porte_ouverte",
}


# --- Décor commun ----------------------------------------------------------

def sol():
    return joint(grain(volume(), densite=0.14, graine=1))


def sol_use():
    return joint(grain(volume(matiere="beton_sombre"), densite=0.22, force=0.22, graine=2))


def sol_carrele():
    def damier(u, v):
        return MARQUAGE if (int(u * 2) + int(v * 2)) % 2 else None
    return joint(volume(peinture=damier))


def mur():
    return contour(grain(volume(elevation=64), densite=0.10, graine=3,
                         faces=("dessus", "gauche", "droite")))


def muret():
    return contour(grain(volume(elevation=22), densite=0.12, graine=4,
                         faces=("dessus", "gauche", "droite")))


def pilier():
    return contour(rivets(volume(elevation=64, matiere="beton_sombre",
                                 largeur_tuile=32), pas=6))


def cloison():
    return volume(tx=1, ty=0.18, elevation=48, matiere="beton")


def panneau():
    poteau = volume(elevation=26, matiere="acier_sombre", largeur_tuile=8)
    plaque = volume(tx=0.5, ty=0.12, elevation=10, matiere="vert")
    return empiler((poteau, 12, 0), (plaque, 0, 24))


def porte_sortie():
    encadrement = volume(tx=1, ty=0.2, elevation=44, matiere="acier_sombre")
    lampe = volume(tx=0.4, ty=0.16, elevation=4, matiere="vert")
    return empiler((encadrement, 0, 0), (lampe, 8, 44))


# --- Supermarché -----------------------------------------------------------

def gondole():
    socle = bandeau(nervures(volume(tx=2, ty=0.7, elevation=14, matiere="metal"), pas=7))
    marchandises = [
        (contour(volume(elevation=7, matiere=m, largeur_tuile=20)), 0.35 + i * 0.65, 0.35)
        for i, m in enumerate(("carton", "rouge", "vert"))
    ]
    return poser(socle, *marchandises)


def gondole_courte():
    socle = bandeau(nervures(volume(tx=1, ty=0.7, elevation=14, matiere="metal"), pas=7))
    caisse_ = contour(volume(elevation=7, matiere="carton", largeur_tuile=20))
    return poser(socle, (caisse_, 0.5, 0.35))


def frigo():
    caisson = rivets(nervures(volume(tx=1, ty=1, elevation=20,
                                     matiere="acier_sombre"), pas=5))
    vitre = volume(tx=0.9, ty=0.9, elevation=6, matiere="verre", largeur_tuile=58)
    return empiler((caisson, 0, 0), (vitre, 3, 22))


def comptoir():
    plan = bandeau(volume(tx=2, ty=1, elevation=20, matiere="peinture"),
                   depuis_haut=2, epaisseur=1)
    terminal = volume(elevation=8, matiere="acier_sombre", largeur_tuile=20)
    return empiler((plan, 0, 0), (terminal, 56, 22))


def caddie():
    chassis = volume(elevation=6, matiere="acier_sombre", largeur_tuile=32)
    panier = volume(elevation=10, matiere="metal", largeur_tuile=28)
    return empiler((chassis, 0, 0), (panier, 2, 8))


# --- Parking ---------------------------------------------------------------

def sol_parking():
    def marquage(u, v):
        return MARQUAGE if v < 0.06 or v > 0.94 else None
    return joint(volume(matiere="bitume", peinture=marquage))


def voiture():
    """Trois étages : train roulant sombre, carrosserie, habitacle vitré."""
    chassis = volume(tx=1.7, ty=0.78, elevation=5, matiere="bitume")
    caisse_ = grain(volume(tx=1.7, ty=0.78, elevation=8, matiere="rouge"),
                    densite=0.05, force=0.10, graine=6)
    habitacle = fenetres(volume(tx=0.85, ty=0.66, elevation=7, matiere="verre"),
                         pas=7, largeur=4, depuis_haut=2, hauteur=4,
                         vitre=(58, 78, 92), allumee=(120, 150, 164), graine=7)

    corps = poser(chassis, (caisse_, 0.85, 0.39))
    return roues(poser(corps, (habitacle, 0.75, 0.39)))


def borne():
    return volume(elevation=18, matiere="jaune", largeur_tuile=16)


def plot():
    return volume(elevation=12, matiere="rouge", largeur_tuile=12)


def barriere():
    return volume(tx=1, ty=0.12, elevation=14, matiere="jaune")


# --- Quartier --------------------------------------------------------------

def trottoir():
    return joint(volume(elevation=6, matiere="beton"))


def banc():
    assise = nervures(volume(tx=1.4, ty=0.5, elevation=8, matiere="bois"), pas=5)
    dossier = volume(tx=1.4, ty=0.12, elevation=10, matiere="bois")
    return empiler((assise, 0, 0), (dossier, 0, 8))


def poubelle():
    return contour(nervures(volume(elevation=18, matiere="vert",
                                   largeur_tuile=26), pas=4))


def lampadaire():
    mat = volume(elevation=52, matiere="acier_sombre", largeur_tuile=8)
    tete = volume(tx=0.6, ty=0.3, elevation=4, matiere="jaune", largeur_tuile=20)
    return empiler((mat, 10, 0), (tete, 0, 50))


def jardiniere():
    bac = volume(elevation=14, matiere="beton_sombre", largeur_tuile=40)
    feuillage = volume(elevation=10, matiere="vert", largeur_tuile=34)
    return empiler((bac, 0, 0), (feuillage, 3, 15))


# --- Cinéma ----------------------------------------------------------------

def rangee_fauteuils():
    assise = nervures(volume(tx=2, ty=0.8, elevation=10, matiere="tissu"), pas=11)
    dossier = volume(tx=2, ty=0.2, elevation=12, matiere="tissu")
    return empiler((assise, 0, 0), (dossier, 0, 10))


def ecran():
    return volume(tx=2.4, ty=0.15, elevation=46, matiere="verre")


def poteau_cordon():
    socle = volume(elevation=4, matiere="acier_sombre", largeur_tuile=14)
    mat = volume(elevation=22, matiere="jaune", largeur_tuile=6)
    return empiler((socle, 0, 0), (mat, 4, 4))


def comptoir_confiserie():
    plan = volume(tx=2, ty=1, elevation=20, matiere="rouge")
    vitrine = volume(tx=1.8, ty=0.8, elevation=10, matiere="verre", largeur_tuile=58)
    return empiler((plan, 0, 0), (vitrine, 6, 22))


# --- Station ---------------------------------------------------------------

def quai():
    def bande(u, v):
        return BANDE_JAUNE if v > 0.86 else None
    return joint(volume(elevation=10, matiere="beton", peinture=bande))


def tourniquet():
    socle = volume(elevation=12, matiere="acier_sombre", largeur_tuile=34)
    bras = volume(tx=0.5, ty=0.1, elevation=6, matiere="metal")
    return empiler((socle, 0, 0), (bras, 6, 13))


def distributeur():
    caisson = rivets(volume(elevation=30, matiere="peinture", largeur_tuile=30), pas=5)
    ecran_ = volume(tx=0.6, ty=0.12, elevation=8, matiere="verre", largeur_tuile=24)
    return empiler((caisson, 0, 0), (ecran_, 2, 24))


def rail():
    return volume(tx=2, ty=0.1, elevation=4, matiere="acier_sombre")



# --- Murs et ouvertures ----------------------------------------------------

def mur_angle():
    return aligner(volume(tx=1, ty=0.2, elevation=64),
                   volume(tx=0.2, ty=1, elevation=64))


def mur_te():
    return aligner(volume(tx=1, ty=0.2, elevation=64),
                   volume(tx=0.2, ty=0.6, elevation=64))


def mur_ouverture():
    return creuser(volume(tx=1, ty=0.2, elevation=64), depuis_haut=26, hauteur=38)


def porte_fermee():
    return creuser(volume(tx=1, ty=0.2, elevation=48), depuis_haut=16, hauteur=32,
                   couleur=(74, 92, 108))


def porte_ouverte():
    return creuser(volume(tx=1, ty=0.2, elevation=48), depuis_haut=16, hauteur=32,
                   couleur=(26, 26, 32))


# --- Variantes de sol et marquages -----------------------------------------

def sol_fissure():
    return joint(tache(grain(volume(), densite=0.16, graine=11),
                       (60, 60, 66), densite=0.10, graine=12))


def sol_sale():
    return joint(tache(grain(volume(matiere="beton_sombre"), densite=0.20, graine=13),
                       (72, 64, 48), densite=0.22, graine=14))


def flaque():
    def forme(u, v):
        d = ((u - 0.5) ** 2 + (v - 0.5) ** 2) ** 0.5
        return (58, 74, 84) if d < 0.30 else None
    return volume(peinture=forme)


def bouche_egout():
    def grille(u, v):
        if max(abs(u - 0.5), abs(v - 0.5)) > 0.22:
            return None
        return (52, 52, 58) if int(v * 22) % 2 else (86, 86, 92)
    return joint(volume(matiere="beton_sombre", peinture=grille))


def fleche_sol():
    def marque(u, v):
        if abs(v - 0.5) < 0.10 and 0.2 < u < 0.8:
            return MARQUAGE
        if abs(u - 0.62) < 0.18 and abs(v - 0.5) < (0.62 + 0.18 - u):
            return MARQUAGE
        return None
    return joint(volume(matiere="beton_sombre", peinture=marque))


def immeuble_petit():
    corps = volume(tx=2, ty=2, elevation=72, matiere="beton")
    return fenetres(corps, pas=12, largeur=5, depuis_haut=8, hauteur=8, graine=21)


def immeuble_haut():
    corps = volume(tx=2, ty=2, elevation=120, matiere="beton_sombre")
    for etage in range(0, 96, 18):
        fenetres(corps, pas=12, largeur=5, depuis_haut=8, hauteur=8,
                 etage=etage, graine=22 + etage)
    return corps


def boutique():
    corps = volume(tx=2, ty=1.4, elevation=44, matiere="peinture")
    return creuser(fenetres(corps, pas=14, largeur=6, depuis_haut=6, hauteur=6,
                            graine=23), depuis_haut=16, hauteur=24,
                   largeur_relative=0.5, couleur=(70, 96, 112))


def abribus():
    return creuser(volume(tx=1.6, ty=0.7, elevation=30, matiere="verre"),
                   depuis_haut=4, hauteur=22, largeur_relative=0.7,
                   couleur=(140, 172, 182))


def conteneur():
    return nervures(volume(tx=1, ty=0.7, elevation=22, matiere="vert"), pas=5)


def feu_tricolore():
    mat = volume(elevation=44, matiere="acier_sombre", largeur_tuile=8)
    tete = volume(elevation=10, matiere="rouge", largeur_tuile=10)
    return poser(mat, (tete, 0.5, 0.5))


# --- Station : matériel roulant --------------------------------------------

def wagon():
    chassis = volume(tx=3.05, ty=0.94, elevation=5, matiere="bitume")
    corps = volume(tx=3, ty=0.9, elevation=26, matiere="peinture")
    fenetres(corps, pas=11, largeur=6, depuis_haut=5, hauteur=8, graine=31)
    bandeau(corps, depuis_haut=15, epaisseur=1, couleur=(214, 214, 206))
    return roues(poser(chassis, (corps, 3.05 / 2, 0.94 / 2)),
                 essieux=(0.14, 0.26, 0.74, 0.86), largeur=5, hauteur=3)


def wagon_tete():
    chassis = volume(tx=2.45, ty=0.94, elevation=5, matiere="bitume")
    corps = volume(tx=2.4, ty=0.9, elevation=26, matiere="peinture")
    fenetres(corps, pas=11, largeur=6, depuis_haut=5, hauteur=8, graine=32)
    creuser(corps, depuis_haut=5, hauteur=10, largeur_relative=0.22,
            couleur=(36, 44, 52))
    return roues(poser(chassis, (corps, 2.45 / 2, 0.94 / 2)),
                 essieux=(0.16, 0.28, 0.72, 0.84), largeur=5, hauteur=3)


def panneau_horaires():
    mat = volume(elevation=30, matiere="acier_sombre", largeur_tuile=8)
    plaque = volume(tx=0.7, ty=0.1, elevation=14, matiere="verre")
    return poser(mat, (plaque, 0.5, 0.5))


# --- Parking et supermarché ------------------------------------------------

def camionnette():
    return vehicule(2.0, 0.82, 20, "metal", "acier_sombre", part_cabine=0.32,
                    graine=34)


def tete_gondole():
    socle = bandeau(nervures(volume(tx=1, ty=1, elevation=14, matiere="metal"), pas=7))
    caisses = [(contour(volume(elevation=7, matiere=m, largeur_tuile=20)), u, v)
               for m, u, v in (("rouge", 0.3, 0.3), ("carton", 0.7, 0.3),
                               ("vert", 0.5, 0.7))]
    return poser(socle, *caisses)


def portique_antivol():
    return aligner(volume(tx=0.14, ty=0.14, elevation=40, matiere="peinture"),
                   volume(tx=0.14, ty=1, elevation=40, matiere="peinture"))



def _train_roulant(tx, ty=0.8, hauteur=5):
    return volume(tx=tx, ty=ty, elevation=hauteur, matiere="bitume")




def roues(img, essieux=(0.22, 0.78), largeur=7, hauteur=4,
          pneu=(26, 26, 30), jante=(150, 150, 158)):
    """Ajoute des roues sous la silhouette, une fois le véhicule composé.

    Les poser sur le châssis ne sert à rien : la caisse le recouvre. On les
    accroche donc au bord inférieur de la silhouette, où elles débordent.
    """
    haut = Image.new("RGBA", (img.width, img.height + hauteur), TRANSPARENT)
    haut.alpha_composite(img, (0, 0))
    haut.info.update(img.info)
    px = haut.load()

    bas = {}
    for x in range(img.width):
        colonne = [y for y in range(img.height) if img.getpixel((x, y))[3] > 0]
        if colonne:
            bas[x] = max(colonne)
    if not bas:
        return haut

    pointe = max(bas, key=lambda c: bas[c])
    aretes = ([x for x in sorted(bas) if x <= pointe],
              [x for x in sorted(bas) if x >= pointe])

    for arete in aretes:
        if len(arete) < largeur * 2:
            continue
        for essieu in essieux:
            centre = arete[min(len(arete) - 1, round((len(arete) - 1) * essieu))]
            for x in range(centre - largeur // 2, centre + largeur // 2 + 1):
                if x not in bas:
                    continue
                retrait = 1 if abs(x - centre) >= largeur // 2 else 0
                for k in range(retrait, hauteur):
                    y = bas[x] + 1 + k - retrait
                    if y < haut.height:
                        px[x, y] = pneu + (255,)
                if abs(x - centre) <= 1:
                    px[x, bas[x] + hauteur // 2] = jante + (255,)

    return haut


def vehicule(tx, ty, hauteur, caisse, cabine, part_cabine=0.30, graine=0,
             liseré=None, cannelures=None, essieux=(0.20, 0.80)):
    """Véhicule d'une seule caisse, à la manière du wagon.

    Empiler une cabine et une remorque comme deux volumes distincts laisse un
    joint visible et casse la lecture : la cabine est ici une zone de la même
    caisse, différenciée par la teinte du toit et son pare-brise.
    """
    def toit(u, v):
        return MATIERES[cabine][0] if u < tx * part_cabine else None

    chassis = volume(tx=tx + 0.1, ty=ty + 0.06, elevation=5, matiere="bitume")
    corps = volume(tx=tx, ty=ty, elevation=hauteur, matiere=caisse, peinture=toit)

    if cannelures:
        nervures(corps, pas=cannelures)

    limite = round((tx * part_cabine + ty) * 32)
    fenetres(corps, pas=8, largeur=5, depuis_haut=3, hauteur=6,
             vitre=(58, 78, 92), allumee=(120, 150, 164), graine=graine,
             bornes=(0, limite))

    if liseré is not None:
        bandeau(corps, depuis_haut=hauteur // 2, epaisseur=1, couleur=liseré)

    return roues(poser(chassis, (corps, (tx + 0.1) / 2, (ty + 0.06) / 2)),
                 essieux=essieux)


def bus():
    """Même construction que le wagon : châssis, caisse, bandeau, fenêtres."""
    chassis = _train_roulant(2.6)
    caisse_ = volume(tx=2.6, ty=0.8, elevation=24, matiere="vert")
    fenetres(caisse_, pas=10, largeur=6, depuis_haut=4, hauteur=8, graine=41)
    bandeau(caisse_, depuis_haut=14, epaisseur=1, couleur=(228, 228, 218))
    return poser(chassis, (caisse_, 1.3, 0.4))


def camion():
    return vehicule(2.8, 0.85, 26, "metal", "rouge", part_cabine=0.26,
                    graine=42, cannelures=6)


def camion_benne():
    return vehicule(2.3, 0.85, 20, "jaune", "acier_sombre", part_cabine=0.30,
                    graine=43, cannelures=5)


def taxi():
    chassis = _train_roulant(1.7, ty=0.78)
    caisse_ = volume(tx=1.7, ty=0.78, elevation=8, matiere="jaune")
    habitacle = fenetres(volume(tx=0.85, ty=0.66, elevation=7, matiere="verre"),
                         pas=7, largeur=4, depuis_haut=2, hauteur=4,
                         vitre=(58, 78, 92), allumee=(120, 150, 164), graine=44)
    enseigne = volume(tx=0.3, ty=0.2, elevation=3, matiere="rouge")
    corps = poser(chassis, (caisse_, 0.85, 0.39))
    corps = poser(corps, (habitacle, 0.75, 0.39))
    return roues(poser(corps, (enseigne, 0.7, 0.39)))


def voiture_epave():
    chassis = _train_roulant(1.7, ty=0.78)
    caisse_ = grain(volume(tx=1.7, ty=0.78, elevation=7, matiere="beton_sombre"),
                    densite=0.08, force=0.20, graine=45,
                    faces=("gauche", "droite"))
    habitacle = eventrer(volume(tx=0.8, ty=0.62, elevation=5, matiere="acier_sombre"),
                         densite=0.30, graine=46)
    corps = poser(chassis, (caisse_, 0.85, 0.39))
    return eventrer(poser(corps, (habitacle, 0.75, 0.39)), densite=0.10, graine=47)


def ambulance():
    base = vehicule(2.1, 0.85, 22, "metal", "metal", part_cabine=0.32,
                    graine=47, liseré=(196, 72, 68))
    gyrophare = volume(tx=0.35, ty=0.3, elevation=3, matiere="rouge")
    return poser(base, (gyrophare, 0.55, 0.30))


def scooter():
    chassis = _train_roulant(0.8, ty=0.3, hauteur=4)
    selle = volume(tx=0.7, ty=0.3, elevation=8, matiere="peinture")
    guidon = volume(tx=0.14, ty=0.3, elevation=12, matiere="acier_sombre")
    corps = poser(chassis, (selle, 0.4, 0.15))
    return roues(poser(corps, (guidon, 0.12, 0.15)),
                 essieux=(0.15, 0.85), largeur=5, hauteur=3)


THEMES = {
    "commun": {
        "sol": sol, "sol_use": sol_use, "sol_carrele": sol_carrele,
        "mur": mur, "muret": muret, "pilier": pilier, "cloison": cloison,
        "panneau": panneau, "porte_sortie": porte_sortie,
        "mur_angle": mur_angle, "mur_te": mur_te, "mur_ouverture": mur_ouverture,
        "porte_fermee": porte_fermee, "porte_ouverte": porte_ouverte,
        "sol_fissure": sol_fissure, "sol_sale": sol_sale, "flaque": flaque,
        "bouche_egout": bouche_egout, "fleche_sol": fleche_sol,
    },
    "supermarche": {
        "gondole": gondole, "gondole_courte": gondole_courte, "frigo": frigo,
        "comptoir": comptoir, "caddie": caddie,
        "tete_gondole": tete_gondole, "portique_antivol": portique_antivol,
    },
    "parking": {
        "sol_parking": sol_parking, "voiture": voiture, "borne": borne,
        "plot": plot, "barriere": barriere, "camionnette": camionnette,
        "camion": camion, "camion_benne": camion_benne, "taxi": taxi,
        "voiture_epave": voiture_epave, "ambulance": ambulance,
    },
    "quartier": {
        "trottoir": trottoir, "banc": banc, "poubelle": poubelle,
        "lampadaire": lampadaire, "jardiniere": jardiniere,
        "bus": bus, "scooter": scooter,
        "immeuble_petit": immeuble_petit, "immeuble_haut": immeuble_haut,
        "boutique": boutique, "abribus": abribus, "conteneur": conteneur,
        "feu_tricolore": feu_tricolore,
    },
    "cinema": {
        "rangee_fauteuils": rangee_fauteuils, "ecran": ecran,
        "poteau_cordon": poteau_cordon, "comptoir_confiserie": comptoir_confiserie,
    },
    "station": {
        "quai": quai, "tourniquet": tourniquet, "distributeur": distributeur,
        "rail": rail, "wagon": wagon, "wagon_tete": wagon_tete,
        "panneau_horaires": panneau_horaires,
    },
}

CATALOGUE = {nom: fn for formes in THEMES.values() for nom, fn in formes.items()}
THEME_DE = {nom: theme for theme, formes in THEMES.items() for nom in formes}


def elevation_reelle(img):
    return max(0, img.height - img.info["hauteur_dessus"])


def planche(images, echelle=3):
    """Aligne les formes sur un sol commun, avec la silhouette du joueur."""
    silhouette = Image.new("RGBA", (24, HAUTEUR_PERSONNAGE), (226, 70, 70, 110))
    elements = list(images) + [silhouette]
    dalle = sol()
    pas = max(max(e.width for e in elements), dalle.width) + 10
    largeur = pas * len(elements) + 10
    hauteur = max(e.height for e in elements) + 40

    fond = Image.new("RGBA", (largeur, hauteur), (26, 26, 32, 255))
    x = 10
    for e in elements:
        fond.alpha_composite(dalle, (x + (pas - 10 - dalle.width) // 2, hauteur - 16 - dalle.height))
        fond.alpha_composite(e, (x + (pas - 10 - e.width) // 2, hauteur - 16 - e.height))
        x += pas
    return fond.resize((fond.width * echelle, fond.height * echelle), Image.NEAREST)


def main():
    analyseur = argparse.ArgumentParser(description=__doc__)
    analyseur.add_argument("formes", nargs="*")
    analyseur.add_argument("--theme", choices=sorted(THEMES))
    analyseur.add_argument("--sortie", default="assets", type=Path)
    analyseur.add_argument("--liste", action="store_true")
    options = analyseur.parse_args()

    if options.liste:
        for theme, formes in THEMES.items():
            print(f"{theme:14} {' '.join(sorted(formes))}")
        return

    if options.formes:
        noms = options.formes
    elif options.theme:
        noms = sorted(THEMES[options.theme])
    else:
        noms = sorted(CATALOGUE)

    inconnus = [n for n in noms if n not in CATALOGUE]
    if inconnus:
        analyseur.error(f"forme inconnue : {', '.join(inconnus)}")

    options.sortie.mkdir(parents=True, exist_ok=True)
    manifeste = {}
    produites = {}

    for nom in noms:
        img = CATALOGUE[nom]()
        if not nom.startswith(("sol", "quai", "trottoir", "rail")):
            img = contour(img)
        emprise = list(img.info.get("emprise", (1.0, 1.0)))
        img = reduire(img)
        # Les roues agrandissent le canevas d'une marge qu'elles n'occupent pas
        # toujours : sans recadrage, le manifeste annonce une taille fausse et
        # l'objet se pose décalé.
        boite = img.getbbox()
        if boite and boite != (0, 0, img.width, img.height):
            info = dict(img.info)
            img = img.crop(boite)
            img.info.update(info)
        dossier = options.sortie / THEME_DE[nom]
        dossier.mkdir(exist_ok=True)
        img.save(dossier / f"{nom}.png")
        produites[nom] = img

        haut = elevation_reelle(img)
        manifeste[nom] = {
            "theme": THEME_DE[nom],
            "taille": list(img.size),
            "ancrage": [img.width // 2, img.height - 1],
            "elevation": haut,
            "categorie": "obstacle_bas" if haut <= PLAFOND_OBSTACLE_BAS else "haut",
            # Emprise au sol, en tuiles : le chargeur marque toutes les cases
            # couvertes, pas seulement celle de l'ancrage. Sans elle, une gondole
            # de deux tuiles n'en bloquerait qu'une.
            "emprise": emprise,
            # Le chargeur en tire la grille de passabilité : c'est la seule
            # source, et elle est déclarée plutôt que devinée.
            "bloquant": nom not in FRANCHISSABLES,
            # Au-delà de la limite, l'objet masque un personnage : le rendu doit
            # le passer en semi-transparence quand le joueur est derrière.
            "transparence_si_derriere": haut > PLAFOND_OBSTACLE_BAS,
        }
        alerte = "" if haut <= PLAFOND_OBSTACLE_BAS else "   masque le joueur"
        print(f"{THEME_DE[nom]:12} {nom:22} {img.width:3}x{img.height:<3} "
              f"élévation {haut:2}{alerte}")

    ecrire_manifeste(options.sortie / "manifeste.json", "decor_iso.py",
                     {"version_format": 1,
                      "tuile": [LARGEUR_TUILE, LARGEUR_TUILE // 2],
                      "formes": manifeste})

    for theme, formes in THEMES.items():
        lot = [produites[n] for n in sorted(formes) if n in produites]
        if lot:
            planche(lot).save(options.sortie / f"controle_{theme}.png")

    if "sol" in produites:
        carrelage(produites["sol"]).save(options.sortie / "controle_carrelage.png")


if __name__ == "__main__":
    main()
