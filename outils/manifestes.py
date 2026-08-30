# Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
# SPDX-License-Identifier: Apache-2.0

"""Écriture des manifestes, commune aux quatre générateurs.

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


def ecrire_manifeste(chemin, generateur, contenu):
    """Écrit un manifeste, mention de licence en première clé.

    `newline` explicite et saut de ligne final, parce que le manifeste est le
    seul fichier texte que ces scripts produisent. Sans eux, Python écrit du
    CRLF sous Windows quand `.gitattributes` restitue du LF au clone, et la
    vérification de `ressources.py --verifier`, qui compare les octets, déclare
    les quatre manifestes différents sans qu'un caractère du contenu ait bougé.
    """
    entete = (f"{COPYRIGHT} — {LICENCE} — généré par outils/{generateur},"
              " ne pas modifier à la main")
    texte = json.dumps({"$comment": entete, **contenu}, indent=2, ensure_ascii=False)
    Path(chemin).write_text(texte + "\n", encoding="utf-8", newline="\n")
