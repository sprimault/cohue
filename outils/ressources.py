# Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
# SPDX-License-Identifier: Apache-2.0

"""Régénère toutes les ressources du jeu, puis les contrôle.

    python outils/ressources.py                  -> tout, dans ./assets
    python outils/ressources.py --sortie /tmp/a  -> ailleurs
    python outils/ressources.py --controle       -> contrôle seul, sans écrire
    python outils/ressources.py --verifier       -> régénère et exige un dépôt propre

`--verifier` est ce que passe l'intégration continue : il régénère dans un
dossier temporaire et compare au contenu versionné. Un écart signale une
retouche manuelle d'un PNG ou une version de Pillow différente — dans les deux
cas un défaut, puisque les images doivent être reproductibles à l'identique.
"""

import argparse
import filecmp
import json
import subprocess
import sys
import tempfile
from pathlib import Path

import numpy as np
from PIL import Image
from scipy import ndimage

OUTILS = Path(__file__).parent

GENERATEURS = (
    ("décor", "decor_iso.py", "decors"),
    ("créatures", "figurines.py", "personnages"),
    ("objets", "objets.py", "objets"),
    ("sons", "sons.py", "sons"),
    ("interface", "interface.py", "interface"),
)

# Seuils du contrôle. Ils ne sont pas décoratifs : chacun a coûté une erreur.
COULEURS_MAX = 26          # le contour et le grain en fabriquent vite des dizaines
TROUS_TOLERES = 4          # quelques pixels isolés sur une forme éventrée
PENTE_ATTENDUE = 0.5       # projection 2:1
ECART_PENTE = 0.12


def generer(sortie, silencieux=False):
    for libelle, script, dossier in GENERATEURS:
        cible = sortie / dossier
        resultat = subprocess.run(
            [sys.executable, str(OUTILS / script), "--sortie", str(cible)],
            capture_output=True, text=True)
        if resultat.returncode != 0:
            print(resultat.stdout, resultat.stderr, sep="\n")
            raise SystemExit(f"échec du générateur {script}")
        if not silencieux:
            derniere = [l for l in resultat.stdout.splitlines() if l.strip()][-1]
            print(f"{libelle:10} {derniere.strip()}")


def _pentes(masque):
    """Pente des deux arêtes hautes, pour vérifier la projection 2:1."""
    hauteur, largeur = masque.shape
    sommets = []
    for x in range(largeur):
        colonne = np.nonzero(masque[:, x])[0]
        sommets.append(colonne[0] if len(colonne) else -1)
    sommets = np.array(sommets)
    milieu = largeur // 2
    mesures = {}
    for cote, (a, b) in {"gauche": (int(largeur * 0.15), milieu - 6),
                         "droite": (milieu + 6, int(largeur * 0.85))}.items():
        segment = sommets[a:b]
        abscisses = np.arange(len(segment))
        valides = segment >= 0
        if valides.sum() > 12:
            mesures[cote] = float(np.polyfit(abscisses[valides], segment[valides], 1)[0])
    return mesures


# Sous quelle clé chaque manifeste range ses entrées. Liste close : un manifeste
# nouveau s'y déclare, et c'est le geste qui le fait entrer dans les contrôles.
CLES_D_ENTREES = ("formes", "objets", "profils", "sons", "armes", "interface",
                  "progression")


def _entrees(chemin):
    """Rend le dictionnaire des entrées d'un manifeste, quelle que soit sa forme.

    Chaque manifeste porte un en-tête — version de format, taille de tuile — et
    range ses entrées sous une clé. Un lecteur qui l'ignorerait prendrait
    `version_format` pour un asset.

    Une clé inconnue arrête tout, plutôt que de rendre le document entier comme
    on le faisait : un contrôle privé de son entrée échoue, il ne passe pas. Le
    repli a fonctionné le jour où le manifeste des armes est arrivé — il a levé,
    donc on l'a vu ; un manifeste d'une autre forme aurait rendu quelque chose de
    plausible, et les contrôles auraient vérifié des clés d'en-tête en croyant
    lire des assets.

    Le `$comment` du bloc d'entrées est retiré. La convention l'autorise sur
    toute structure du format, et le décodeur Go l'accepte partout ; ici il
    serait rendu comme une entrée dont la valeur est une chaîne, et chaque
    contrôle qui parcourt les entrées casserait sur le premier `info.get`.
    """
    contenu = json.loads(chemin.read_text(encoding="utf-8"))
    for cle in CLES_D_ENTREES:
        if cle in contenu:
            return {nom: info for nom, info in contenu[cle].items()
                    if nom != "$comment"}
    raise SystemExit(f"{chemin} : aucune clé d'entrées connue, "
                     f"attendu {' ou '.join(CLES_D_ENTREES)}")


def _cotes_de_grille(sortie):
    """Ce dont les images sont des cases de taille fixe, et ce que la case admet.

    Une bande d'animation vit dans une grille : ses marges transparentes sont
    voulues, et son nombre de couleurs se compte image par image. Confondre les
    deux familles fait remonter des centaines de faux défauts.

    Rend `chemin -> (largeur, hauteur, vides_admises)`. La clé est un dossier
    pour une bande — toutes les images d'un profil partagent sa case — et un
    fichier pour une planche de glyphes, qui est une image unique.

    **Le manifeste déclare `cote` pour un carré et `cellule` pour un rectangle,
    et ce n'est pas la même notion.** Un sprite isométrique vit dans une case
    carrée par construction ; une police a une cellule dont la hauteur porte la
    ligne d'accent et le jambage, et qui n'a aucune raison d'égaler sa largeur.
    C'est ici que les deux se ramènent à un couple, et nulle part dans les
    manifestes : le format des quatre autres n'a pas à changer pour un besoin
    qui n'est pas le leur.
    """
    cotes = {}
    for chemin in sortie.rglob("manifeste.json"):
        contenu = _entrees(chemin)
        for nom, info in contenu.items():
            if "cote" in info and "cycles" in info:
                cotes[chemin.parent / nom] = (info["cote"], info["cote"], False)
            # Une planche de glyphes porte des cases vides — l'espace, l'espace
            # insécable — qui sont des glyphes légitimes sans dessin, là où une
            # image vide dans une bande d'animation est un défaut.
            elif "cellule" in info and "glyphes" in info:
                largeur, hauteur = info["cellule"]
                cotes[chemin.parent / info["fichier"]] = (largeur, hauteur, True)
    return cotes


def controler_sons(sortie):
    """Niveau, saturation, coupure brutale : les trois défauts audibles.

    Une coupure en pleine amplitude produit un clic, et un clic répété cent
    fois par minute est ce qui fait couper le son d'un jeu.
    """
    import struct
    import wave

    defauts = []
    for chemin in sorted(sortie.rglob("*.wav")):
        with wave.open(str(chemin)) as fichier:
            images = fichier.getnframes()
            brut = fichier.readframes(images)
            taux = fichier.getframerate()
        echantillons = struct.unpack(f"<{images}h", brut)
        nom = chemin.relative_to(sortie)
        crete = max(abs(v) for v in echantillons)
        fin = max(abs(v) for v in echantillons[-int(0.005 * taux):])

        # Seuil bas : plusieurs sons sont volontairement discrets — le tir de
        # base part plusieurs fois par seconde et doit passer sous la musique.
        if crete < 1200:
            defauts.append((nom, f"inaudible, crête {crete}"))
        if crete >= 32767:
            defauts.append((nom, "saturé"))
        if fin > crete * 0.25:
            defauts.append((nom, "coupure brutale en fin : produit un clic"))
    return defauts


def controler(sortie, pentes=False):
    """Rend la liste des défauts. Vide, tout va bien."""
    defauts = []
    controlees = 0
    grilles = _cotes_de_grille(sortie)

    for chemin in sorted(sortie.rglob("*.png")):
        image = Image.open(chemin).convert("RGBA")
        pixels = np.asarray(image)
        masque = pixels[:, :, 3] > 0
        controlees += 1
        nom = chemin.relative_to(sortie)

        case = grilles.get(chemin) or next(
            (c for dossier, c in grilles.items() if dossier in chemin.parents), None)
        cote = case[0] if case else None

        # Icônes d'interface et bandes d'objets : leur case est fixe et leurs
        # marges sont voulues, comme pour les créatures.
        case_fixe = cote is not None or chemin.stem.endswith(
            ("_icone", "_scintille", "_appui", "_rupture")) or chemin.stem in (
            "etincelle", "souffle") or chemin.stem.startswith("eclats_")

        # Un anneau ou une gerbe de particules est creux par nature : chercher
        # une silhouette pleine n'a pas de sens sur un effet. Une planche de
        # glyphes non plus : le contre-intérieur d'un « o », d'un « e », d'un
        # « à » en fabrique des centaines, et ce sont eux qui font la lettre.
        creux = chemin.stem in ("etincelle", "souffle") or (case and case[2])

        if not set(np.unique(pixels[:, :, 3]).tolist()) <= {0, 255}:
            defauts.append((nom, "alpha non binaire : le moteur attend un masque net"))

        trous = int((ndimage.binary_fill_holes(masque) & ~masque).sum())
        if trous > TROUS_TOLERES and not creux:
            defauts.append((nom, f"{trous} pixels de trou dans la silhouette"))

        # Sur une bande, chaque image se compte séparément : concaténer trois
        # images de vingt couleurs n'en fait pas une de soixante.
        tranches = ([pixels[:, i * cote:(i + 1) * cote] for i in range(image.width // cote)]
                    if cote else [pixels])
        for tranche in tranches:
            visible = tranche[:, :, 3] > 0
            if not visible.any():
                if not (case and case[2]):
                    defauts.append((nom, "image entièrement vide dans la bande"))
                continue
            couleurs = len({tuple(c) for c in tranche[visible][:, :3]})
            if couleurs > COULEURS_MAX:
                defauts.append((nom, f"{couleurs} couleurs, au-dessus de {COULEURS_MAX}"))
                break

        if case:
            largeur, hauteur, _ = case
            if image.height != hauteur or image.width % largeur:
                defauts.append((nom, f"bande hors grille de {largeur}x{hauteur} px"))
        elif not case_fixe and image.getbbox() != (0, 0, image.width, image.height):
            defauts.append((nom, "non recadré : des colonnes ou lignes vides au bord"))

        # Mesure indicative : elle ne vaut que pour un volume nu. Sur une forme
        # composée — objet posé dessus, cabine, roues — le sommet de la
        # silhouette n'est plus l'arête du volume, et la mesure ment.
        if pentes and not case_fixe and image.width >= 48:
            for arete, pente in _pentes(masque).items():
                if abs(abs(pente) - PENTE_ATTENDUE) > ECART_PENTE:
                    defauts.append((nom, f"pente {arete} = {pente:.3f}, attendu ±0,5"))

    return controlees, defauts


# Ce qu'un profil doit porter, et rien d'autre. Deux descriptions du même
# ensemble — le générateur qui l'écrit, cette table qui l'exige — et c'est
# voulu : la question « qu'est-ce qui garantit qu'elles restent d'accord »
# trouve ici une bonne réponse, un contrôle joué à chaque `make controle`. La
# duplication est le mécanisme, pas son prix : une table qui se mettrait à jour
# depuis le générateur ne contrôlerait rien, et ajouter un champ doit être un
# geste délibéré en deux endroits.
#
# Le rôle décide d'abord, parce que les trois natures n'ont pas les mêmes
# valeurs : le joueur a une vie et pas de points, un ennemi l'inverse, une
# entité d'ambiance ni l'un ni l'autre.
CHAMPS_COMMUNS = {"role", "vitesse_relative", "rayon_tuiles", "groupe",
                  "gabarit", "nom", "origine", "cote", "appui", "directions",
                  "cycles", "variantes"}

CHAMPS_PAR_ROLE = {
    "joueur": {"vitesse_tuiles_s", "vie", "plafond_degats_s"},
    "ennemi": {"comportement", "touches", "points", "cout_pression",
               "poids_separation", "max_simultane", "degats_contact_s",
               "gemmes"},
    "ambiance": {"comportement"},
}

# Ce qu'un comportement ajoute, et qui n'a de sens que pour lui. Un
# `portee_tuiles` sur un Quidam ne serait jamais lu et laisserait croire qu'il
# tire : le contrôle attrape donc le champ de trop autant que le champ absent.
CHAMPS_PAR_COMPORTEMENT = {
    "poursuite":   set(),
    "va_et_vient": set(),
    "charge":      {"degats_charge"},
    "flanc":       {"tangentiel"},
    "tir":         {"portee_tuiles", "degats_tir", "vitesse_projectile_tuiles_s"},
    "explosion":   {"degats_explosion", "rayon_explosion_tuiles"},
    "soin":        set(),
}


# Ce qu'un manifeste ne porte plus, et chez qui c'est parti.
#
# Une liste de champs bannis plutôt que la liste blanche qu'on préfère ailleurs :
# ici on ne se protège pas d'un fichier tiers mais d'une régression connue, et le
# message doit apprendre où ranger la valeur plutôt que de dire qu'elle est de
# trop. Les deux fichiers portaient la portée de la Buse, avec deux valeurs
# différentes, et personne ne l'a vu.
#
# Elle sert aux profils autant qu'aux objets. Un champ déménagé ne se distingue
# pas d'une faute de frappe pour qui lit « champ inconnu », et les deux se
# corrigent dans des fichiers différents.
CHAMPS_DEMENAGES = {
    "degats": "chez le tireur : l'arme pour le joueur, le profil pour une créature",
    "portee_tuiles": "chez le tireur, qui la porte déjà",
    "vitesse_px_s": "chez le tireur, en tuiles par seconde et non en pixels",
    "traverse": "chez l'arme, où ce sera un passif",
    "experience": "chez la progression, à côté des seuils qu'elle alimente",
    "portee_ramassage_tuiles": "chez la progression, avec la durée de vie d'une "
                               "gemme dont elle forme un couple",
}


def profils(sortie):
    """Vérifie que chaque profil porte les champs de son rôle, et rien de plus.

    Le joueur est le seul à porter une vitesse absolue ; les autres n'ont qu'un
    rapport. Un profil qui aurait gardé l'ancien nom de champ se chargerait à
    vitesse nulle — une créature immobile au milieu de la horde, qu'on mettrait
    longtemps à relier à un renommage.
    """
    chemin = sortie / "personnages" / "manifeste.json"
    if not chemin.exists():
        return []

    defauts = []
    propres = set().union(*CHAMPS_PAR_COMPORTEMENT.values())
    for nom, info in _entrees(chemin).items():
        role = info.get("role")
        if role not in CHAMPS_PAR_ROLE:
            defauts.append((nom, f"role « {role} » inconnu, attendu "
                                 f"{' ou '.join(sorted(CHAMPS_PAR_ROLE))}"))
            continue

        attendus = CHAMPS_COMMUNS | CHAMPS_PAR_ROLE[role]
        if role == "joueur":
            attendus -= {"vitesse_relative", "comportement"}
        else:
            comportement = info.get("comportement")
            if comportement not in CHAMPS_PAR_COMPORTEMENT:
                defauts.append((nom, f"comportement « {comportement} » inconnu"))
                continue
            attendus |= CHAMPS_PAR_COMPORTEMENT[comportement]

        for manquant in sorted(attendus - set(info)):
            defauts.append((nom, f"champ « {manquant} » absent"))
        for surnumeraire in sorted(set(info) - attendus):
            # Un champ qui a déménagé se signale par sa destination, pas par son
            # inutilité : « inconnu » ferait chercher une faute de frappe là où
            # il faut chercher un autre fichier. C'est la même table que pour les
            # objets, et elle sert ici pour la même raison.
            if surnumeraire in CHAMPS_DEMENAGES:
                defauts.append((nom, f"champ « {surnumeraire} » : il a déménagé "
                                     f"{CHAMPS_DEMENAGES[surnumeraire]}"))
                continue
            quoi = "propre à un autre comportement" if surnumeraire in propres else "inconnu"
            defauts.append((nom, f"champ « {surnumeraire} » {quoi}"))
    return defauts


def objets(sortie):
    """Vérifie qu'aucun objet ne reprend une valeur qui appartient au tireur.

    Un projectile est un objet qui vole : sa taille et son ancrage le décrivent,
    ses dégâts non. La règle est au chapitre 4 de la conception, et ce contrôle
    est ce qui empêche de la reperdre — sans lui, quelqu'un remet `degats` sur un
    projectile par symétrie avec le manifeste des personnages, qui porte bien ses
    valeurs de jeu.
    """
    chemin = sortie / "objets" / "manifeste.json"
    if not chemin.exists():
        return []

    defauts = []
    for nom, info in _entrees(chemin).items():
        for champ in sorted(set(info) & set(CHAMPS_DEMENAGES)):
            defauts.append((nom, f"champ « {champ} » : il a déménagé "
                                 f"{CHAMPS_DEMENAGES[champ]}"))
    return defauts


def renvois_d_objets(sortie):
    """Exige que tout renvoi `objet` désigne un objet du catalogue.

    Une valeur de jeu sortie du manifeste d'objets y laisse un lien par nom entre
    deux fichiers, et un renommage le casse en silence : le moteur lit toujours
    son chiffre, l'objet ne s'appelle plus pareil, et personne ne l'apprend. Ce
    contrôle est ce qui rend le lien visible — c'est le prix du déménagement, et
    il se paie une fois.

    Le champ n'est lu que par ici. C'est voulu et c'est dit dans les deux
    manifestes : le moteur n'a pas besoin du nom, il a besoin du chiffre, et un
    champ que personne ne lit du tout est exactement ce qui avait laissé
    `experience` dormir sur la gemme sans que sa valeur serve.
    """
    catalogue = sortie / "objets" / "manifeste.json"
    if not catalogue.exists():
        return []
    connus = set(_entrees(catalogue))

    defauts = []
    for chemin in sorted(sortie.rglob("manifeste.json")):
        if chemin == catalogue:
            continue
        for nom, info in _entrees(chemin).items():
            vise = info.get("objet")
            if vise and vise not in connus:
                defauts.append((f"{chemin.parent.name}/{nom}",
                                f"objet « {vise} » introuvable au catalogue"))
    return defauts


def formes(sortie):
    """Vérifie que le coût de traversée est déclaré là, et seulement là, où il est lu.

    Exigé sur ce qui se franchit, refusé sur ce qui bloque. Un champ facultatif
    dans un seul sens laisserait un `cout_traversee` orphelin sur un mur, jamais
    lu, qui ferait croire à un réglage — et `bloquant` avec un coût fini est un
    état que le format ne doit pas savoir exprimer.
    """
    chemin = sortie / "decors" / "manifeste.json"
    if not chemin.exists():
        return []

    defauts = []
    for nom, info in _entrees(chemin).items():
        cout = info.get("cout_traversee")
        if info.get("bloquant"):
            if cout is not None:
                defauts.append((nom, "bloquant et pourtant un cout_traversee"))
        elif cout is None:
            defauts.append((nom, "franchissable sans cout_traversee"))
        # `isinstance(True, int)` est vrai : sans l'exclusion, un `true` recopié
        # de la ligne `bloquant` juste au-dessus passerait pour un coût de 1.
        elif isinstance(cout, bool) or not isinstance(cout, int) or cout < 1:
            defauts.append((nom, f"cout_traversee « {cout} » : entier supérieur à zéro attendu"))
    return defauts


def manifestes(sortie):
    """Vérifie que chaque manifeste décrit bien ce qui est sur le disque."""
    defauts = []
    for chemin in sorted(sortie.rglob("manifeste.json")):
        entete = json.loads(chemin.read_text(encoding="utf-8"))
        if "version_format" not in entete:
            defauts.append((chemin.parent.name,
                            "manifeste sans version_format : impossible de migrer"))
        # Le manifeste part dans l'archive publiée comme le reste des ressources,
        # et c'est le seul fichier généré où la mention de licence peut se perdre
        # sans que rien ne le montre : il n'a pas de commentaire à relire.
        if "$comment" not in entete:
            defauts.append((chemin.parent.name,
                            "manifeste sans $comment : mention de licence absente"))
        dossier = chemin.parent
        contenu = _entrees(chemin)
        for nom, info in contenu.items():
            if "cycles" in info:
                for cycle, reglages in info["cycles"].items():
                    for direction in info["directions"]:
                        bandes = list(dossier.rglob(f"{nom}/**/{cycle}_{direction}.png"))
                        if not bandes:
                            defauts.append((f"{nom}/{cycle}_{direction}", "bande absente"))
                            continue
                        largeur = Image.open(bandes[0]).width
                        attendu = reglages["images"] * info["cote"]
                        if largeur != attendu:
                            defauts.append((f"{nom}/{cycle}_{direction}",
                                            f"{largeur} px pour {reglages['images']} images"))
            elif "taille" in info and not list(dossier.rglob(f"{nom}.png")):
                defauts.append((nom, "fichier annoncé au manifeste mais absent"))

            # Un renvoi qui ne pointe nulle part casse le jeu au moment précis
            # où l'objet est détruit, c'est-à-dire le plus tard possible.
            destruction = info.get("destruction", {})
            for cle in ("ruine", "cycle_appui", "cycle_rupture"):
                cible = destruction.get(cle)
                if cible and cible not in contenu:
                    defauts.append((nom, f"{cle} renvoie à « {cible} », absent du manifeste"))
            matiere = destruction.get("eclats")
            if matiere and f"eclats_{matiere}" not in contenu:
                defauts.append((nom, f"éclats « {matiere} » sans particules générées"))
            if destruction.get("mode") == "interaction" and "touches" not in destruction:
                defauts.append((nom, "destructible sans nombre de touches"))
            ruine = contenu.get(destruction.get("ruine") or "", {})
            if ruine.get("bloquant"):
                defauts.append((nom, "sa ruine bloque encore : casser n'ouvre rien"))
            # Une ruine plus haute ou aussi haute que son original signale une
            # forme cassée qu'on a oublié de raccourcir : elle masquerait encore
            # un personnage, et rien à l'écran ne dirait que l'obstacle est
            # tombé. Vrai de toutes les ruines livrées, sauf celle qui a fait
            # écrire ce contrôle.
            if ruine and ruine.get("elevation", 0) >= info.get("elevation", 0):
                defauts.append((nom, f"sa ruine culmine à {ruine['elevation']} px"
                                     f" contre {info.get('elevation')} : elle n'est pas tombée"))
    return defauts


def police(sortie):
    """Confronte la planche de glyphes à ce que son manifeste en déclare.

    C'est le contrôle des bandes appliqué à une grille rectangulaire : la
    largeur de la planche vaut le nombre de glyphes multiplié par celle de la
    cellule. Il attrape le glyphe ajouté à la table sans régénération, qui
    décalerait tout le texte d'une cellule à partir de lui.

    Les avances sont une liste parallèle aux glyphes, donc un décalage se lit
    sur les longueurs — c'est ce qui a décidé de la liste contre un
    dictionnaire, où un décalage n'aurait eu aucune trace.
    """
    chemin = sortie / "interface" / "manifeste.json"
    if not chemin.exists():
        return []

    defauts = []
    for nom, info in _entrees(chemin).items():
        if "glyphes" not in info:
            continue
        glyphes, avances = info["glyphes"], info["avances"]
        largeur, hauteur = info["cellule"]
        planche = chemin.parent / info["fichier"]
        if not planche.exists():
            defauts.append((nom, f"planche annoncée mais absente : {info['fichier']}"))
            continue
        image = Image.open(planche)
        if image.width != largeur * len(glyphes):
            defauts.append((nom, f"{image.width} px pour {len(glyphes)} glyphes"
                                 f" de {largeur} px"))
        if len(avances) != len(glyphes):
            defauts.append((nom, f"{len(avances)} avances pour {len(glyphes)} glyphes"))
        if not 0 < info["ligne_de_base"] < hauteur:
            defauts.append((nom, f"ligne de base {info['ligne_de_base']} hors"
                                 f" d'une cellule de {hauteur} px"))
        # Un doublon décale la lecture sans rien casser à la génération : le
        # second exemplaire est simplement inatteignable, et le texte qui
        # l'emploie rend le premier.
        if len(set(glyphes)) != len(glyphes):
            defauts.append((nom, "la table des glyphes porte un doublon"))
    return defauts


def icones(sortie):
    """Vérifie que les icônes déclarées existent et sont carrées.

    Le manifeste ne dit que des noms de fichiers : la taille de chacun se lit
    dans son image, et c'est ce qui évite une seconde description du dessin. En
    échange, rien d'autre que ce contrôle ne rattrape une taille retirée de la
    liste du générateur sans que le fichier disparaisse — le manifeste
    n'annoncerait plus une image qui reste au dépôt, et le jeu s'ouvrirait avec
    une icône de moins sans un mot.

    Carrées parce qu'un gestionnaire de fenêtres qui reçoit un rectangle
    l'étire : l'icône ne serait pas refusée, elle serait déformée.
    """
    chemin = sortie / "interface" / "manifeste.json"
    if not chemin.exists():
        return []

    entree = _entrees(chemin).get("icone")
    if entree is None:
        return [("interface/manifeste.json", "aucune icone declaree")]

    defauts = []
    for fichier in entree["fichiers"]:
        image = chemin.parent / fichier
        if not image.exists():
            defauts.append((f"interface/{fichier}", "icone annoncee mais absente"))
            continue
        with Image.open(image) as ouverte:
            if ouverte.width != ouverte.height:
                defauts.append((f"interface/{fichier}",
                                f"{ouverte.width}x{ouverte.height}, une icone est carree"))
    return defauts


def renvois_de_sons(sortie):
    """Vérifie que chaque son nommé par un objet existe réellement.

    Un renvoi mort ne se voit qu'au moment où l'objet est détruit — c'est-à-dire
    le plus tard possible, et souvent chez un joueur.

    La correspondance est **exacte**. Une comparaison par préfixe passerait sur
    « gem » et sur « g » aussi bien que sur « gemme » : elle ne couvrirait pas,
    elle en donnerait l'air. Ce qui désigne une suite de degrés le déclare avec
    sa propre clé, et se vérifie autrement.
    """
    catalogue = sortie / "sons" / "manifeste.json"
    if not catalogue.exists():
        return []
    sons = set(_entrees(catalogue))
    defauts = []
    for chemin in sorted(sortie.rglob("manifeste.json")):
        if chemin == catalogue:
            continue
        for nom, info in _entrees(chemin).items():
            attendus = [info.get("son")] + [info.get("destruction", {}).get(cle)
                                            for cle in ("son_appui", "son_rupture")]
            for cle in filter(None, attendus):
                if cle not in sons:
                    defauts.append((nom, f"son « {cle} » introuvable"))

            famille = info.get("famille_sons")
            if famille:
                defauts += _famille_de_sons(nom, famille, sons)
    return defauts


def _famille_de_sons(nom, famille, sons):
    """Exige qu'une famille soit une suite de degrés complète, à partir de zéro.

    Le moteur avance d'un degré à chaque déclenchement rapproché et repart du
    premier après un silence : un trou dans la suite se traduirait par un
    silence au milieu d'une volée, ce qui s'entend et ne se diagnostique pas.
    """
    degres = sorted(int(s.rsplit("_", 1)[1]) for s in sons
                    if s.startswith(f"{famille}_") and s.rsplit("_", 1)[1].isdigit())
    if len(degres) < 2:
        return [(nom, f"famille « {famille} » : {len(degres)} degré(s), il en faut au moins deux")]
    if degres != list(range(len(degres))):
        return [(nom, f"famille « {famille} » : degrés {degres}, suite incomplète")]
    return []


def _engendre(chemin):
    """Dit si ce chemin relève d'un dossier qu'un générateur écrit.

    Le périmètre vient de la table des générateurs et non d'une liste de noms à
    tenir : tout `assets/` n'est pas généré — la table d'armes, que l'on règle à
    la main pendant l'équilibrage, ne l'est pas — et un dossier tenu à la main ne
    doit être ni comparé, ni signalé comme n'étant plus produit.

    Le dossier et non le fichier : à l'intérieur d'un dossier généré, un fichier
    que plus rien ne produit reste une anomalie. Une forme retirée d'un script
    laisserait sinon son image versionnée, elle partirait dans le binaire par
    `go:embed`, et personne ne la verrait.
    """
    return bool(chemin.parts) and chemin.parts[0] in {d for _, _, d in GENERATEURS}


def _meme_dessin(reference, produit):
    """Dit si deux images portent les mêmes pixels, quels que soient leurs octets.

    Un PNG est un conteneur compressé, et la comparaison stricte n'est pas
    portable : Pillow n'embarque pas la même bibliothèque de compression selon
    la plateforme — la wheel Windows charge zlib-ng, la wheel Linux zlib —, si
    bien que toutes les images du dépôt diffèrent d'un système à l'autre à
    dessin rigoureusement identique. C'est le dessin qui est versionné ici, pas
    la façon de le compresser.

    Ce que le contrôle attrape n'en souffre pas : une retouche manuelle, un
    script modifié sans régénération et un rendu de Pillow qui changerait
    déplacent tous des pixels. Sons et manifestes gardent la comparaison
    stricte, qu'ils passent sur les deux systèmes.
    """
    with Image.open(reference) as a, Image.open(produit) as b:
        return a.size == b.size and a.mode == b.mode and a.tobytes() == b.tobytes()


def _ecart(reference, produit):
    """Rend la description de l'écart entre deux fichiers, ou None s'il n'y en a pas."""
    if filecmp.cmp(reference, produit, shallow=False):
        return None
    if reference.suffix != ".png":
        return "diffère de la régénération"
    with Image.open(reference) as a, Image.open(produit) as b:
        if a.size != b.size or a.mode != b.mode:
            return f"taille ou mode différents : {a.size}{a.mode} contre {b.size}{b.mode}"
    return None if _meme_dessin(reference, produit) else "pixels différents"


def comparer(reference, produit):
    """Liste les fichiers qui diffèrent entre le dépôt et une régénération."""
    ecarts = []
    attendus = {r for p in produit.rglob("*") if p.is_file()
                for r in [p.relative_to(produit)] if _engendre(r)}
    presents = {r for p in reference.rglob("*") if p.is_file()
                for r in [p.relative_to(reference)] if _engendre(r)}

    for manquant in sorted(attendus - presents):
        ecarts.append((manquant, "absent du dépôt"))
    for surnumeraire in sorted(presents - attendus):
        ecarts.append((surnumeraire, "présent dans le dépôt mais plus généré"))
    for commun in sorted(attendus & presents):
        quoi = _ecart(reference / commun, produit / commun)
        if quoi:
            ecarts.append((commun, quoi))
    return ecarts


def main():
    analyseur = argparse.ArgumentParser(description=__doc__)
    analyseur.add_argument("--sortie", type=Path, default=Path("assets"))
    analyseur.add_argument("--controle", action="store_true",
                           help="contrôler l'existant sans rien régénérer")
    analyseur.add_argument("--verifier", action="store_true",
                           help="régénérer à côté et exiger l'identique")
    analyseur.add_argument("--pentes", action="store_true",
                           help="mesurer la projection des formes simples (indicatif)")
    options = analyseur.parse_args()

    if options.verifier:
        with tempfile.TemporaryDirectory() as temporaire:
            temoin = Path(temporaire) / "assets"
            generer(temoin, silencieux=True)
            ecarts = comparer(options.sortie, temoin)
        if ecarts:
            print(f"{len(ecarts)} écart(s) entre le dépôt et la régénération :")
            for nom, quoi in ecarts[:20]:
                print(f"  {nom} — {quoi}")
            raise SystemExit(1)
        print("ressources reproductibles à l'identique")
        return

    if not options.controle:
        generer(options.sortie)

    controlees, defauts = controler(options.sortie, pentes=options.pentes)
    defauts += manifestes(options.sortie)
    defauts += profils(options.sortie)
    defauts += objets(options.sortie)
    defauts += renvois_d_objets(options.sortie)
    defauts += formes(options.sortie)
    defauts += police(options.sortie)
    defauts += icones(options.sortie)
    defauts += controler_sons(options.sortie)
    defauts += renvois_de_sons(options.sortie)
    sons = len(list(options.sortie.rglob("*.wav")))

    print(f"\n{controlees} images et {sons} sons contrôlés")
    if defauts:
        for nom, quoi in defauts:
            print(f"  {nom} — {quoi}")
        raise SystemExit(f"{len(defauts)} défaut(s)")
    print("aucun défaut")


if __name__ == "__main__":
    main()
