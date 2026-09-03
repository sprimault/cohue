# Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
# SPDX-License-Identifier: Apache-2.0

"""Écriture des manifestes, commune aux générateurs.

**Ce module ne s'exécute pas** : comme `primitives_iso.py`, il est importé — par
`decor_iso.py`, `figurines.py`, `objets.py` et `sons.py`, les quatre qui
produisent un manifeste. Le lancer directement ne fait rien, et c'est normal.

Un manifeste fait contrat entre les images et le moteur : c'est lui, et jamais
le code, qui dit les tailles, les cycles et les valeurs de jeu. Les quatre
générateurs l'écrivent de la même façon, et le passage par une fonction unique
tient les deux détails qu'on oublie — la mention de licence, que JSON ne peut
porter qu'en `$comment`, et les fins de ligne.

Bibliothèque standard uniquement : `sons.py` n'a aucune dépendance, et ce module
ne doit pas être ce qui lui en donne une.
"""

import json
from pathlib import Path

COPYRIGHT = "Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>"
LICENCE = "SPDX-License-Identifier: Apache-2.0"

# Un pas de simulation, en millisecondes. La conception fixe soixante pas par
# seconde et convertit les durées une fois au chargement : une valeur sous le
# pas s'arrondirait à zéro, et le cycle ne changerait jamais d'image.
TICK_MS = 1000 / 60


def _durees_courtes(contenu, chemin=""):
    """Rend les clés en `_ms` dont la valeur tomberait sous un pas."""
    fautives = []
    if isinstance(contenu, dict):
        for cle, valeur in contenu.items():
            ou = f"{chemin}.{cle}" if chemin else cle
            if cle.endswith("_ms") and isinstance(valeur, (int, float)):
                if valeur < TICK_MS:
                    fautives.append((ou, valeur))
            else:
                fautives += _durees_courtes(valeur, ou)
    elif isinstance(contenu, list):
        for i, valeur in enumerate(contenu):
            fautives += _durees_courtes(valeur, f"{chemin}[{i}]")
    return fautives


def ecrire_manifeste(chemin, generateur, contenu):
    """Écrit un manifeste, mention de licence en première clé.

    `newline` explicite et saut de ligne final, parce que le manifeste est le
    seul fichier texte que ces scripts produisent. Sans eux, Python écrit du
    CRLF sous Windows quand `.gitattributes` restitue du LF au clone, et la
    vérification de `ressources.py --verifier`, qui compare les octets, déclare
    les quatre manifestes différents sans qu'un caractère du contenu ait bougé.

    Une durée sous le pas de simulation fait échouer l'écriture. Ces fichiers ne
    sont saisis par personne : une telle valeur est un défaut dans le script qui
    la produit, et le chargeur du jeu la refusera de toute façon. La signaler
    ici la montre à sa source, avec le chemin de la clé et la valeur fautive,
    plutôt qu'au lancement d'une partie.
    """
    if fautives := _durees_courtes(contenu):
        details = ", ".join(f"{ou} = {v} ms" for ou, v in fautives)
        raise ValueError(
            f"{generateur} : duree sous le pas de {TICK_MS:.2f} ms — {details}")

    entete = (f"{COPYRIGHT} — {LICENCE} — généré par outils/{generateur},"
              " ne pas modifier à la main")
    texte = json.dumps({"$comment": entete, **contenu}, indent=2, ensure_ascii=False)
    Path(chemin).write_text(texte + "\n", encoding="utf-8", newline="\n")
