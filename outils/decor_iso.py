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
from primitives_iso import (HAUTEUR_PERSONNAGE, LARGEUR_TUILE, MATIERES,
                            PLAFOND_OBSTACLE_BAS, TRANSPARENT, aligner, appui,
                            bandeau, carrelage, categorie, contour, couvre,
                            creuser, decalque, elevation_reelle, empiler,
                            eventrer, fenetres, grain, grener, joint, nervures,
                            nuance,
                            position, poser, reduire, rivets, surface, tache,
                            texture, volume)

MARQUAGE = (232, 232, 224)
BANDE_JAUNE = (222, 186, 74)

# Ce qui se traverse, et ce que la traversée coûte. L'élévation ne suffit pas à
# le déduire : un trottoir et un quai dépassent du sol et se marchent, une flaque
# est plate et se traverse, alors qu'un muret de même hauteur qu'un trottoir
# arrête tout.
#
# Une seule table plutôt que deux : un ensemble et un dictionnaire de coûts
# finiraient par ne plus lister les mêmes formes, et le coût orphelin ne serait
# jamais lu.
#
# Le coût multiplie la longueur perçue d'une case. À 2, une flaque de deux cases
# vaut un détour de deux cases, et un ennemi la contournera plutôt que de la
# traverser : c'est fort, et c'est voulu comme point de départ. Un ralentissement
# qu'on ne voit pas ne se règle jamais.
FRANCHISSABLES = {
    "sol": 1, "sol_use": 1, "sol_carrele": 1, "sol_parking": 1,
    "sol_fissure": 2, "sol_sale": 2, "flaque": 2,
    "bouche_egout": 1, "fleche_sol": 1, "trottoir": 1, "quai": 1, "rail": 1,
    "porte_ouverte": 1, "chaussee": 1,
    "herbe": 1, "passage_pieton": 1, "moquette": 1, "bande_eveil": 1,
    # Le ballast est le premier ralentisseur d'un thème : il rend la voie
    # pénible à traverser, ce qui est ce qu'une voie doit être.
    "ballast": 2,
}

# Les revêtements qu'on peint sans les cerner : un liseré y doublerait le joint
# au raccord et dessinerait la grille du jeu sur toute la surface.
#
# Une table plutôt que le préfixe de nom qui en tenait lieu. « sol », « quai »,
# « trottoir », « rail » attrapaient exactement ces neuf-là, et n'attraperaient
# ni une herbe, ni des pavés, ni du ballast : le préfixe ne deviendrait pas
# faux, il cesserait d'attraper, ce qui ne se voit sur aucune planche.
#
# Ce qui se pose *sur* un revêtement est cerné, marquages au sol compris : le
# liseré est tout ce qui leur garantit un contraste sur un sol de thème qu'ils
# ne connaissent pas. Un revêtement est franchissable par nature, ce que le
# générateur vérifie contre la table ci-dessus.
#
# Une forme dessinée n'y figure pas, fût-elle un revêtement : on ne retouche
# rien de ce qui est tenu à la main, donc la question du contour ne se pose
# pas pour elle.
REVETEMENTS = {
    "sol", "sol_use", "sol_carrele", "sol_fissure", "sol_sale", "sol_parking",
    "trottoir", "quai", "rail",
    "herbe", "passage_pieton", "moquette", "ballast", "bande_eveil",
}


# --- Décor commun ----------------------------------------------------------

def sol():
    return joint(volume(peinture=texture("beton")))


def sol_use():
    return joint(volume(peinture=nuance(texture("beton"), facteur=0.82)))


def sol_carrele():
    # Le damier reste au code : il alterne deux carreaux par case, donc son pas
    # suit la grille et non la matière. La texture donne la céramique, l'écart
    # de valeur donne les carreaux.
    ceramique = texture("carrelage")
    sombre = nuance(ceramique, facteur=0.86)
    def damier(u, v):
        return sombre(u, v) if (int(u * 2) + int(v * 2)) % 2 else ceramique(u, v)
    return joint(volume(peinture=damier))


def maconnerie(tx=1, ty=1, elevation=64, largeur_tuile=LARGEUR_TUILE,
               matiere="pierre"):
    """Un volume de pierre : appareil de face sur les flancs, projeté au-dessus.

    **Les deux lectures ne se font pas dans le même repère**, et c'est ce qui
    les sépare : un flanc est un plan vertical, que la projection ne parcourt
    pas, quand le dessus est un losange en coordonnées de tuile. Le même
    appareil, lu des deux façons, donne des pierres de face et des pierres vues
    de dessus.

    **La répétition d'une case à l'autre est ici le motif et non un défaut.**
    Une tache trahit un pavage ; un appareil de maçonnerie est fait pour se
    répéter, et c'est ce qui rend la pierre traitable là où le béton ne l'était
    pas — une tuile de mur étant la même image à chaque case.
    """
    appareil = texture(matiere)
    return grener(volume(tx=tx, ty=ty, elevation=elevation,
                         largeur_tuile=largeur_tuile, peinture=appareil),
                  appareil, force=1.3)


def mur():
    return contour(maconnerie())


def muret():
    # Du moellon plutôt que l'appareil des murs, pour la raison qui met du
    # parpaing sur un pilier : vingt-deux pixels de haut ne montreraient qu'une
    # assise et demie du grand appareil, et le muret passerait pour un mur
    # tronqué. De petites pierres serrées lui donnent sa propre trame.
    return contour(maconnerie(elevation=22, matiere="caillou"))


def pilier():
    # **Du parpaing et non la pierre des murs**, et ce n'est pas qu'une
    # variation : un pilier n'occupe qu'un quart de case, si bien que le même
    # appareil n'y montrerait que deux ou trois pierres et le ferait lire comme
    # un bout de mur. Des blocs petits et réguliers le distinguent par leur
    # trame plutôt que par leur valeur — un moellon sombre en aurait fait une
    # tache au milieu du sol, là où c'est la horde qui doit accrocher l'œil.
    #
    # Les rivets sont partis avec le béton coulé qu'ils disaient.
    return contour(maconnerie(largeur_tuile=32, matiere="parpaing"))


def cloison():
    return maconnerie(ty=0.18, elevation=48)


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
    enrobe = texture("bitume")
    def marquage(u, v):
        return MARQUAGE if v < 0.06 or v > 0.94 else enrobe(u, v)
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

def herbe():
    # **Calmée, et c'est une exigence de lisibilité et non de goût.** Brute, la
    # texture donne un vert qui accroche l'œil avant la horde ; tirée vers le
    # gris et assombrie, elle garde ses brins et reste au second plan, où un sol
    # doit être.
    return joint(volume(peinture=nuance(texture("herbe"), facteur=0.82,
                                        vers=(96, 104, 84), force=0.30)))


def passage_pieton():
    # Les bandes suivent `v`, donc un axe du monde, et quatre demi-bandes par
    # case se raccordent d'une case à la suivante : posées en file, elles
    # dessinent un passage continu.
    enrobe = texture("bitume")
    def zebres(u, v):
        return MARQUAGE if int(v * 4) % 2 else enrobe(u, v)
    return joint(volume(matiere="bitume", peinture=zebres))


def trottoir():
    # Le grain vient d'une texture, le relief et les flancs du volume : c'est la
    # répartition qui vaut pour tous les revêtements, la matière d'un côté et la
    # géométrie de l'autre. Le joint de case reste au code — une rainure est un
    # tracé, et un tracé doit rester net au pixel.
    return joint(volume(elevation=6, matiere="beton", peinture=texture("beton_clair")))


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

def moquette():
    # Le cinéma n'avait aucun sol à lui et empruntait le carrelage du
    # supermarché. Son motif se répète d'une case à l'autre, et c'est ce qu'on
    # attend d'une moquette — comme d'un appareil de pierre, et à la différence
    # d'une tache.
    return joint(volume(peinture=texture("moquette")))


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

def ballast():
    # **Réduite en moyenne et non au plus proche voisin**, à l'inverse des
    # autres : un gravier dont les grains sont plus fins que le pas
    # d'échantillonnage donne du bruit qui grouille, là où la moyenne rend la
    # masse qu'on voit à trois mètres. Le coût de deux est ce qui rend la voie
    # pénible à traverser, sa fonction même.
    return joint(volume(peinture=texture("ballast")))


def quai():
    # **La bande d'éveil n'est plus ici, et c'est un défaut corrigé.** Chaque
    # case en portait une : un quai de trois cases de large était rayé de jaune
    # sur toute sa profondeur, alors que la bande borde un quai et ne le pave
    # pas. Elle a sa forme, que l'auteur pose sur la seule rangée de bord.
    return joint(volume(elevation=10, matiere="beton",
                        peinture=texture("beton_clair")))


def bande_eveil():
    # Les plots alternés sont ce qui distingue une bande podotactile d'un simple
    # marquage : ils se comptent en pixels de tuile, comme le damier du
    # carrelage, et restent nets là où une texture les aurait dilués.
    beton = texture("beton_clair")
    sombre = tuple(int(c * 0.78) for c in BANDE_JAUNE)
    def bande(u, v):
        if v <= 0.62:
            return beton(u, v)
        return sombre if (int(u * 10) + int(v * 10)) % 2 else BANDE_JAUNE
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
    return aligner(maconnerie(ty=0.2), maconnerie(tx=0.2))


def mur_te():
    return aligner(maconnerie(ty=0.2), maconnerie(tx=0.2, ty=0.6))


def mur_ouverture():
    return creuser(maconnerie(ty=0.2), depuis_haut=26, hauteur=38)


def porte_fermee():
    return creuser(volume(tx=1, ty=0.2, elevation=48), depuis_haut=16, hauteur=32,
                   couleur=(74, 92, 108))


def porte_ouverte():
    return creuser(volume(tx=1, ty=0.2, elevation=48), depuis_haut=16, hauteur=32,
                   couleur=(26, 26, 32))


# --- Variantes de sol et marquages -----------------------------------------

def sol_fissure():
    # La matière vient de la texture, les fissures restent tracées : ce sont
    # elles qui disent que la case coûte à traverser, et un motif de texture ne
    # se placerait pas là où on le veut.
    #
    # **Elles sont franches, et c'est ce qui a manqué au premier essai.** La
    # texture ayant donné son grain au sol ordinaire, un fissuré qui partageait
    # sa valeur et ses marques pâles ne se distinguait plus de lui : la zone
    # coûteuse devenait invisible, alors qu'un ralentissement qu'on ne voit pas
    # ne se règle jamais. Le léger assombrissement accompagne les marques ; seul,
    # il doublerait le sol usé.
    return joint(tache(volume(peinture=nuance(texture("beton"), facteur=0.94)),
                       (46, 46, 52), densite=0.28, graine=12))


def sol_sale():
    return joint(volume(peinture=nuance(texture("beton"), facteur=0.86,
                                        vers=(96, 82, 58), force=0.35)))


def flaque():
    # **L'eau couvre sa case plutôt que de s'y inscrire en ovale**, et c'est ce
    # qui la fait lire comme une flaque. Un disque par case donnait autant de
    # ronds séparés qu'un lieu en posait — six cases d'affilée faisaient six
    # pastilles, jamais une étendue. Couvrante, elle s'étale sur ce que le lieu
    # lui donne, et les reflets en stries disent la surface.
    #
    # **Elle laisse voir le sol au travers**, en plus sombre et plus froid : la
    # matière est celle du sol, assombrie et tirée vers le bleu, et le grain qui
    # transparaît fait le reste — un aplat uni se lisait comme un trou.
    #
    # Le bord est rongé, et la flaque n'est pas cernée : un liseré en referait
    # un objet posé là, alors qu'elle altère le sol.
    fond = nuance(texture("beton"), facteur=0.52, vers=(42, 66, 84), force=0.55)
    reflet = nuance(texture("beton"), facteur=0.86, vers=(96, 126, 146), force=0.62)

    def eau(u, v):
        return reflet(u, v) if 0.60 < u + v < 0.78 or 1.30 < u + v < 1.42 else fond(u, v)
    return eventrer(decalque(eau, cerne=False), densite=0.14, graine=17)


def bouche_egout():
    # La fonte cerne la grille : sans elle, les barreaux posés à même le sol du
    # thème flottent au lieu d'être encastrés.
    def grille(u, v):
        ecart = max(abs(u - 0.5), abs(v - 0.5))
        if ecart > 0.24:
            return None
        if ecart > 0.19:
            return (64, 64, 70)
        return (44, 44, 50) if int(v * 22) % 2 else (92, 92, 98)
    return decalque(grille)


def fleche_sol():
    # Hampe fine et pointe courte : dessinée en coordonnées de tuile, une flèche
    # est écrasée de moitié par la projection, et celle d'avant s'y perdait.
    def marque(u, v):
        if 0.16 < u < 0.60 and abs(v - 0.5) < 0.07:
            return MARQUAGE
        if 0.60 <= u <= 0.86 and abs(v - 0.5) <= 0.22 * (0.86 - u) / 0.26:
            return MARQUAGE
        return None
    return decalque(marque)


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
        "trottoir": trottoir, "herbe": herbe, "passage_pieton": passage_pieton,
        "banc": banc, "poubelle": poubelle,
        "lampadaire": lampadaire, "jardiniere": jardiniere,
        "boutique": boutique, "abribus": abribus, "conteneur": conteneur,
        "feu_tricolore": feu_tricolore,
    },
    "cinema": {
        "moquette": moquette,
        "rangee_fauteuils": rangee_fauteuils, "ecran": ecran,
        "poteau_cordon": poteau_cordon, "comptoir_confiserie": comptoir_confiserie,
    },
    "station": {
        "quai": quai, "bande_eveil": bande_eveil, "ballast": ballast,
        "tourniquet": tourniquet, "distributeur": distributeur,
        "rail": rail, "wagon": wagon, "wagon_tete": wagon_tete,
        "panneau_horaires": panneau_horaires,
    },
}

# Les formes dont l'image est tenue à la main, et le thème auquel elles
# appartiennent. Ce script en écrit toujours l'entrée de manifeste — taille,
# ancrage, élévation, catégorie, couverture se mesurent sur le dessin livré,
# comme elles se mesuraient sur le volume qu'il remplace —, mais plus l'image,
# que `make decors` écraserait.
#
# **L'emprise est la seule chose qu'un dessin ne porte pas.** Tout le reste se
# lit dans ses pixels ; elle, non : la largeur d'un losange dit la somme de ses
# deux côtés, jamais lequel est lequel.
#
# Une texture ne se régénère pas à l'identique — c'est ce qui range ces formes
# à part, comme les personnages, et leur vaut le `LICENSE.txt` du dossier.
DESSINES = {
    "chaussee": {"theme": "quartier", "emprise": (1.0, 1.0)},
    # **Les véhicules dessinés sont couchés le long de `v`**, quand ceux que
    # `vehicule` compose le sont le long de `u`. Ce n'est pas un choix de
    # cadrage : le modèle ne rend que cette diagonale, et le format prévoit les
    # deux sens depuis que les obstacles fragiles se posent dans l'un ou dans
    # l'autre. Le catalogue y gagne ce qu'il n'avait pas — une rue où les
    # véhicules ne sont pas tous garés dans le même sens.
    "bus": {"theme": "quartier", "emprise": (0.8, 2.6)},
    "minibus": {"theme": "quartier", "emprise": (0.8, 2.6)},
    "autocar": {"theme": "quartier", "emprise": (0.8, 2.6)},
    "bus_imperiale": {"theme": "quartier", "emprise": (0.8, 2.6)},
    "fourgon": {"theme": "quartier", "emprise": (0.8, 2.0)},
    "utilitaire": {"theme": "quartier", "emprise": (0.8, 2.0)},
    "pickup": {"theme": "quartier", "emprise": (0.8, 2.0)},
    "camion_livraison": {"theme": "quartier", "emprise": (0.8, 2.0)},
    "camion_ordures": {"theme": "quartier", "emprise": (0.8, 2.0)},
    "citadine": {"theme": "quartier", "emprise": (0.8, 2.0)},
    "scooter": {"theme": "quartier", "emprise": (0.3, 0.8)},
    # Trois hauteurs qui se distinguent d'un coup d'œil : une boutique basse,
    # un immeuble de rapport, une façade de brique. Leur emprise est carrée,
    # donc leur sens ne se pose pas.
    "immeuble_petit": {"theme": "quartier", "emprise": (2.0, 2.0)},
    "immeuble_haut": {"theme": "quartier", "emprise": (2.0, 2.0)},
    "immeuble_brique": {"theme": "quartier", "emprise": (2.0, 2.0)},
    "immeuble_pierre": {"theme": "quartier", "emprise": (2.0, 2.0)},
    "immeuble_bureau": {"theme": "quartier", "emprise": (2.0, 2.0)},
    "immeuble_tour": {"theme": "quartier", "emprise": (2.0, 2.0)},
}

CATALOGUE = {nom: fn for formes in THEMES.values() for nom, fn in formes.items()}
THEME_DE = {nom: theme for theme, formes in THEMES.items() for nom in formes}

# Un nom ne peut pas relever des deux tables : la seconde gagnerait en silence,
# et l'on croirait régénérer une forme que plus rien ne dessine.
if _deux := sorted(set(DESSINES) & set(THEME_DE)):
    raise SystemExit(f"à la fois dessinée et générée : {', '.join(_deux)}")
THEME_DE |= {nom: quoi["theme"] for nom, quoi in DESSINES.items()}


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


def dessin(dossier, nom):
    """Charge l'image tenue à la main d'une forme de `DESSINES`.

    Elle est lue là où elle est versionnée et jamais dans le dossier de sortie :
    la vérification régénère à côté, dans un répertoire vide, et y chercher le
    dessin ferait échouer le contrôle au lieu de comparer les manifestes.

    Son absence est une erreur et non un passage en silence — sans le fichier,
    il n'y a ni taille, ni élévation, ni couverture à mesurer, et une entrée de
    manifeste écrite sur des valeurs par défaut serait pire que pas d'entrée.

    La hauteur du dessus se dérive de l'emprise, seule chose que le dessin ne
    porte pas : c'est la hauteur qu'aurait le losange, et c'est d'elle que se
    déduit l'élévation.

    **Le point d'appui s'en dérive aussi, et il le faut dès qu'une forme
    dessinée n'est pas carrée.** Sans lui, l'entrée de manifeste retombe sur le
    bas-centre de l'image, qui n'est le sommet bas du losange que lorsque les
    deux côtés de l'emprise sont égaux — le défaut que le générateur a corrigé
    pour les volumes, et qui reviendrait par cette porte. Il ne s'est pas vu
    tant que la seule forme dessinée était un sol d'une tuile.
    """
    ex, ey = DESSINES[nom]["emprise"]
    chemin = Path(dossier) / DESSINES[nom]["theme"] / f"{nom}.png"
    if not chemin.exists():
        raise SystemExit(f"{chemin} : dessin absent, {nom} est tenue à la main")
    img = Image.open(chemin).convert("RGBA")
    img.load()
    img.info["hauteur_dessus"] = round((ex + ey) * LARGEUR_TUILE / 4)
    img.info["emprise"] = (ex, ey)
    img.info["appui"] = appui(ex, LARGEUR_TUILE, img.height)
    return img


def sur_les_sols(marquages, sols, echelle=3):
    """Croise les marques au sol avec les revêtements qui les portent.

    C'est la seule vue où leur défaut se voit : une marque qui emporte son
    propre sol paraît juste sur le revêtement dont elle vient, et fausse partout
    ailleurs. Une planche par thème ne la montre jamais, chacune ne posant ses
    formes que sur le béton de commun.

    Les images s'alignent par le haut, où se trouve la face supérieure : c'est
    elle qu'on marque, qu'elle soit au niveau du sol ou en haut d'un trottoir.
    """
    pas = LARGEUR_TUILE + 8
    hauteur = max(s.height for s in sols) + 8
    planche = Image.new("RGBA", (pas * len(marquages) + 8, hauteur * len(sols) + 8),
                        (26, 26, 32, 255))
    for j, sol_ in enumerate(sols):
        for i, marque in enumerate(marquages):
            x, y = 8 + i * pas, 8 + j * hauteur
            planche.alpha_composite(sol_, (x, y))
            planche.alpha_composite(marque, (x, y))
    return planche.resize((planche.width * echelle, planche.height * echelle),
                          Image.NEAREST)


def main():
    analyseur = argparse.ArgumentParser(description=__doc__)
    analyseur.add_argument("formes", nargs="*")
    analyseur.add_argument("--theme", choices=sorted(THEMES))
    analyseur.add_argument("--sortie", default="assets", type=Path)
    # Hors de `--sortie` : ce qui est livré et ce qui sert à relire ne se mêlent
    # pas. `assets/` part dans le binaire par `go:embed`, qui ne sait pas
    # exclure, et une planche oubliée là s'y retrouverait.
    analyseur.add_argument("--controles", default=Path(".tmp/controle"), type=Path)
    # Là où les dessins tenus à la main sont versionnés, et non `--sortie` : la
    # vérification régénère dans un répertoire vide, où ils n'existent pas.
    analyseur.add_argument("--dessins", default=Path("assets/decors"), type=Path)
    analyseur.add_argument("--liste", action="store_true")
    options = analyseur.parse_args()

    if options.liste:
        for theme, formes in THEMES.items():
            dessinees = [n for n, q in DESSINES.items() if q["theme"] == theme]
            marque = "".join(f" {n}*" for n in sorted(dessinees))
            print(f"{theme:14} {' '.join(sorted(formes))}{marque}")
        print("\n* tenue à la main : le manifeste seul est écrit")
        return

    if options.formes:
        noms = options.formes
    elif options.theme:
        noms = sorted(set(THEMES[options.theme])
                      | {n for n, q in DESSINES.items() if q["theme"] == options.theme})
    else:
        noms = sorted(set(CATALOGUE) | set(DESSINES))

    inconnus = [n for n in noms if n not in CATALOGUE and n not in DESSINES]
    if inconnus:
        analyseur.error(f"forme inconnue : {', '.join(inconnus)}")

    # Les deux tables portent deux questions sur les mêmes objets — ce qui se
    # traverse, ce qui ne se cerne pas —, et rien ne les tiendrait d'accord :
    # un revêtement qu'on ne pourrait pas fouler serait une contradiction, et
    # elle ne se verrait qu'à l'œil, sur une planche que personne n'ouvre.
    hors = sorted(REVETEMENTS - set(FRANCHISSABLES))
    if hors:
        analyseur.error(f"revêtement que rien ne traverse : {', '.join(hors)}")

    options.sortie.mkdir(parents=True, exist_ok=True)
    manifeste = {}
    produites = {}

    for nom in noms:
        tenue = nom in DESSINES
        img = dessin(options.dessins, nom) if tenue else CATALOGUE[nom]()
        emprise = list(img.info["emprise"] if tenue
                       else img.info.get("emprise", (1.0, 1.0)))
        # Mesurée avant le contour, qui ne touche à aucun alpha, et avant la
        # réduction, qui reseuille le même masque.
        couvrant = couvre(img, emprise)
        if not tenue:
            if nom not in REVETEMENTS and img.info.get("cerner", True):
                img = contour(img)
            img = reduire(img)
            # Les roues agrandissent le canevas d'une marge qu'elles n'occupent
            # pas toujours : sans recadrage, le manifeste annonce une taille
            # fausse et l'objet se pose décalé. Un décalque en est dispensé —
            # son cadre est celui de la case, et le rogner descendrait la marque
            # d'autant.
            boite = None if img.info.get("decalque") else img.getbbox()
            if boite and boite != (0, 0, img.width, img.height):
                info = dict(img.info)
                img = img.crop(boite)
                img.info.update(info)
                if "appui" in info:
                    ax, ay = info["appui"]
                    img.info["appui"] = (ax - boite[0], ay - boite[1])
            dossier = options.sortie / THEME_DE[nom]
            dossier.mkdir(exist_ok=True)
            img.save(dossier / f"{nom}.png")
        produites[nom] = img

        haut = elevation_reelle(img)
        manifeste[nom] = {
            "theme": THEME_DE[nom],
            "taille": list(img.size),
            # Le sommet bas du losange de l'emprise, et non le bas-centre de
            # l'image : les deux ne coïncident que pour une emprise carrée, et
            # l'écart vaut un quart de tuile en diagonale sur les allongées. Une
            # forme dessinée n'a pas de volume dont le dériver, mais elle occupe
            # sa case entière, donc son losange est celui de l'image.
            "ancrage": list(img.info.get("appui", (img.width // 2, img.height - 1))),
            "elevation": haut,
            # Trois hauteurs, dérivées de la passabilité d'abord : c'est la vue
            # de dessus de l'éditeur qui les lit, et elle veut savoir où l'on
            # passe. Un quai se marche, une porte ouverte aussi, quelle que soit
            # leur hauteur.
            "categorie": categorie(nom not in FRANCHISSABLES, haut),
            # Emprise au sol, en tuiles : le chargeur marque toutes les cases
            # couvertes, pas seulement celle de l'ancrage. Sans elle, une gondole
            # de deux tuiles n'en bloquerait qu'une.
            "emprise": emprise,
            # Si la forme peint tout le losange de sa case, donc s'il faut
            # peindre le sol du thème avant elle. Le rendu le déduisait de
            # l'emprise, ce qui était vrai tant qu'un volume peignait tout son
            # dessus : une marque au sol occupe sa case sans rien en cacher, et
            # sépare les deux. Le sens du doute est sûr — croire qu'une forme
            # couvre laisse un trou à l'écran, croire l'inverse coûte un blit.
            "couvrant": couvrant,
            # Le chargeur en tire la grille de passabilité : c'est la seule
            # source, et elle est déclarée plutôt que devinée.
            "bloquant": nom not in FRANCHISSABLES,
            **({} if nom not in FRANCHISSABLES
               else {"cout_traversee": FRANCHISSABLES[nom]}),
            # Au-delà de la limite, la forme masque un personnage de 64. Le
            # champ dit ce fait et non le remède : le rendu redessine la
            # silhouette de ce qui est caché plutôt que d'effacer le décor, et
            # c'est sa décision, pas celle du manifeste.
            "masquant": haut > PLAFOND_OBSTACLE_BAS,
        }
        alerte = "" if haut <= PLAFOND_OBSTACLE_BAS else "   masque le joueur"
        if tenue:
            alerte = "   dessinée, manifeste seul"
        print(f"{THEME_DE[nom]:12} {nom:22} {img.width:3}x{img.height:<3} "
              f"élévation {haut:2}{alerte}")

    ecrire_manifeste(options.sortie / "manifeste.json", "decor_iso.py",
                     {"version_format": 1,
                      "tuile": [LARGEUR_TUILE, LARGEUR_TUILE // 2],
                      "formes": manifeste})

    options.controles.mkdir(parents=True, exist_ok=True)
    for theme, formes in THEMES.items():
        lot = [produites[n] for n in sorted(formes) if n in produites]
        if lot:
            planche(lot).save(options.controles / f"controle_{theme}.png")

    if "sol" in produites:
        carrelage(produites["sol"]).save(options.controles / "controle_carrelage.png")

    marquages = [produites[n] for n in noms
                 if n in produites and produites[n].info.get("decalque")]
    revetements = [produites[n] for n in noms
                   if n in produites and manifeste[n]["couvrant"] and n in FRANCHISSABLES]
    if marquages and revetements:
        sur_les_sols(marquages, revetements).save(
            options.controles / "controle_marquages.png")


if __name__ == "__main__":
    main()
