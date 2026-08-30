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


def _entrees(chemin):
    """Rend le dictionnaire des entrées d'un manifeste, quelle que soit sa forme.

    Chaque manifeste porte désormais un en-tête — version de format, taille de
    tuile — et range ses entrées sous une clé. Un lecteur qui l'ignorerait
    prendrait `version_format` pour un asset.
    """
    contenu = json.loads(chemin.read_text(encoding="utf-8"))
    for cle in ("formes", "objets", "profils", "sons"):
        if cle in contenu:
            return contenu[cle]
    return contenu


def _cotes_de_grille(sortie):
    """Dossiers dont les images sont des cases de taille fixe.

    Une bande d'animation vit dans une grille : ses marges transparentes sont
    voulues, et son nombre de couleurs se compte image par image. Confondre les
    deux familles fait remonter des centaines de faux défauts.
    """
    cotes = {}
    for chemin in sortie.rglob("manifeste.json"):
        contenu = _entrees(chemin)
        for nom, info in contenu.items():
            if "cote" in info and "cycles" in info:
                cotes[chemin.parent / nom] = info["cote"]
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
        if chemin.name.startswith(("controle_", "apercu_", "palette")):
            continue
        image = Image.open(chemin).convert("RGBA")
        pixels = np.asarray(image)
        masque = pixels[:, :, 3] > 0
        controlees += 1
        nom = chemin.relative_to(sortie)

        cote = next((c for dossier, c in grilles.items()
                     if dossier in chemin.parents), None)

        # Icônes d'interface et bandes d'objets : leur case est fixe et leurs
        # marges sont voulues, comme pour les créatures.
        case_fixe = cote is not None or chemin.stem.endswith(
            ("_icone", "_scintille", "_appui", "_rupture")) or chemin.stem in (
            "etincelle", "souffle") or chemin.stem.startswith("eclats_")

        # Un anneau ou une gerbe de particules est creux par nature : chercher
        # une silhouette pleine n'a pas de sens sur un effet.
        effet = chemin.stem in ("etincelle", "souffle")

        if not set(np.unique(pixels[:, :, 3]).tolist()) <= {0, 255}:
            defauts.append((nom, "alpha non binaire : le moteur attend un masque net"))

        trous = int((ndimage.binary_fill_holes(masque) & ~masque).sum())
        if trous > TROUS_TOLERES and not effet:
            defauts.append((nom, f"{trous} pixels de trou dans la silhouette"))

        # Sur une bande, chaque image se compte séparément : concaténer trois
        # images de vingt couleurs n'en fait pas une de soixante.
        tranches = ([pixels[:, i * cote:(i + 1) * cote] for i in range(image.width // cote)]
                    if cote else [pixels])
        for tranche in tranches:
            visible = tranche[:, :, 3] > 0
            if not visible.any():
                defauts.append((nom, "image entièrement vide dans la bande"))
                continue
            couleurs = len({tuple(c) for c in tranche[visible][:, :3]})
            if couleurs > COULEURS_MAX:
                defauts.append((nom, f"{couleurs} couleurs, au-dessus de {COULEURS_MAX}"))
                break

        if cote:
            if image.height != cote or image.width % cote:
                defauts.append((nom, f"bande hors grille de {cote} px"))
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


# Les planches de contrôle et les aperçus sont régénérés mais pas versionnés :
# `.gitignore` les exclut parce qu'ils se refont à volonté. Les comparer ferait
# échouer tout clone frais, où ils n'existent pas encore.
NON_VERSIONNE = ("controle_", "apercu_")


def _versionne(chemin):
    """Dit si ce fichier a vocation à être dans le dépôt."""
    return not chemin.name.startswith(NON_VERSIONNE)


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
    attendus = {r for p in produit.rglob("*") if p.is_file() and _versionne(p)
                for r in [p.relative_to(produit)] if _engendre(r)}
    presents = {r for p in reference.rglob("*") if p.is_file() and _versionne(p)
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
