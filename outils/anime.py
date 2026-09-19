#!/usr/bin/env python3
# Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
# SPDX-License-Identifier: Apache-2.0

"""Convertit l'aperçu animé en WebP, douze fois plus léger que le GIF.

    python outils/anime.py .tmp/apercus/partie.gif .tmp/apercus/partie.webp

**Sans perte, et ce n'est pas un scrupule.** Le mode avec perte est ici le plus
lourd — treize mégaoctets à qualité 90 contre un et demi sans perte : il est
fait pour des dégradés photographiques, quand le pixel art n'est que des arêtes
franches et des aplats, que le sans-perte compresse très bien et que l'autre
brouille avant de mal les coder.

**Le GIF reste la sortie du générateur** parce qu'il se lit partout et qu'aucune
bibliothèque tierce ne l'écrit — `image/gif` est dans la bibliothèque standard
de Go. Cette conversion vient après, pour ce qui se publie : le format ne
change pas ce que l'image montre, seulement ce qu'elle pèse.

Ce qui est écrit n'entre pas dans le dépôt : `.tmp/` est le seul endroit où l'on
produit, et la vidéo d'un jeu n'est pas une ressource qu'il embarque.
"""
import argparse
from pathlib import Path

from PIL import Image, ImageSequence


def convertir(source, sortie):
    """Écrit le WebP animé et rend son poids, en octets."""
    with Image.open(source) as anime:
        images = [image.convert("RGB") for image in ImageSequence.Iterator(anime)]
        # La cadence se relit dans le fichier plutôt que de se redéclarer : deux
        # descriptions du même rythme finiraient par diverger, et c'est le
        # générateur qui décide.
        duree = anime.info.get("duration", 50)

    images[0].save(sortie, save_all=True, append_images=images[1:],
                   duration=duree, loop=0, lossless=True, quality=100, method=4)
    return Path(sortie).stat().st_size


def main():
    a = argparse.ArgumentParser(description=__doc__)
    a.add_argument("source")
    a.add_argument("sortie")
    o = a.parse_args()

    avant = Path(o.source).stat().st_size
    apres = convertir(o.source, o.sortie)
    print(f"{o.sortie} {apres / 1e6:.2f} Mo, {avant / apres:.0f} fois plus léger")


if __name__ == "__main__":
    main()
