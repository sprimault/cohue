# Cohue

Action-roguelite urbain en vue isométrique 2D, sous pression de horde. Des salles enchaînées, du tir automatique, une build qui se compose en quinze minutes, et un éditeur qui permet à n'importe qui de fabriquer et de partager ses propres lieux.

Le titre dit la foule désordonnée et pressante, sans détour et sans jeu de mots
à expliquer. *Heure de pointe* a été écarté pour cette raison : il évoque
d'abord les embouteillages, et un titre doit orienter avant qu'on lise la
description. *Affluence*, *Marée* et *Terminus* étaient les autres candidats.

---

## 1. Le jeu en une page

Le joueur traverse une suite de lieux du quotidien qui se sont retournés : un parking souterrain, un supermarché, un quartier, un cinéma, une station de métro. La foule arrive et n'arrête plus d'arriver.

Il ne contrôle que son déplacement — le tir de base est automatique, seules les armes lourdes trouvées dans les caisses se déclenchent à la touche. Il monte de niveau en ramassant les gemmes, choisit ses armes et ses passifs, casse des caisses pour des ressources, trouve la porte de sortie en lisant la signalétique du décor, et passe au lieu suivant. Une run complète dure environ quinze minutes.

L'objectif de conception n'est pas la difficulté, c'est **la relance**. Tout ce qui suit est ordonné par ce que ça apporte à la boucle « je meurs, je recommence ».

L'enchaînement des lieux n'est pas figé : une campagne est un graphe de salles, composé dans l'éditeur. Supermarché puis quartier puis parking, ou tout autre ordre, y compris avec des embranchements.

---

## 2. La boucle de compulsion

### La décision de rejouer se prend en trois secondes

Point numéro un, et il est purement technique. À la mort, le joueur est dans l'état où il veut relancer. Un écran de résultats animé, un retour au menu, une sélection de personnage, une sélection de campagne : quatre occasions de se lever et de partir. Chaque friction coûte un pourcentage de relances.

Cible : **une touche, moins d'une seconde, même configuration**. L'écran de mort affiche le résumé et « Espace pour relancer » en évidence. Le reste est secondaire dans la hiérarchie visuelle. Ça paraît trivial, c'est la variable la plus lourde de tout le système.

Corollaire propre à la progression par salles : mourir au troisième lieu renvoie au premier, ce qui est bien plus punitif qu'une arène unique. Les trois premières minutes doivent donc être **rapides à repasser** — courbe basse, pas de temps mort, aucune cinématique. Sans ça, la relance devient une corvée et tout le bénéfice est perdu.

### Finir en plein élan

Une run doit s'arrêter alors que quelque chose est en cours : 200 XP avant une évolution, deux armes sur trois pour une fusion, une synergie composée mais jamais vue à l'œuvre. C'est la phrase intérieure — « la prochaine fois je prends la même chose mais je monte l'orbite d'abord » — qui déclenche la relance, jamais le score.

Concrètement : la mort typique d'un joueur moyen doit tomber vers 8-11 minutes, alors que les évolutions les plus intéressantes se débloquent vers 12-14. La majorité des runs meurent avec un plan inachevé.

### Le tempo des montées de niveau

C'est le métronome du plaisir. Front-load agressif : premier niveau à 12 secondes, puis toutes les 15-20 secondes pendant deux minutes, puis l'écart s'allonge. Règle dure : **jamais plus de 45 secondes sans un choix à faire**.

Le choix compte plus que la récompense. Trois cartes, dont deux tentantes. Si le joueur prend systématiquement la même, l'équilibrage est cassé — la bonne carte est celle qui fait hésiter.

### La montée de niveau ne casse pas le rythme

Le choix des trois cartes met le jeu en pause, mais **brièvement et sans cérémonie** : un fondu court, les cartes, le choix, un fondu, et on repart. Pas d'écran plein, pas d'animation d'entrée, pas de son long.

C'est ce qui compte en fin de run, quand les niveaux s'enchaînent toutes les vingt secondes : une transition d'une seconde répétée quinze fois hache la partie, alors que deux fois cent cinquante millisecondes se traversent sans qu'on les remarque.

La pause est réelle — la horde se fige — parce que choisir sous pression n'est pas un choix, c'est une loterie, et le document a posé que le choix compte plus que la récompense.

### Le pic de puissance

La courbe de sensation doit croiser celle de la difficulté.

- Minutes 0-3 : fragile, chaque ennemi compte.
- Minutes 4-9 : montée, on encaisse et on rend.
- Minutes 10-14 : le joueur traverse les hordes sans regarder, l'écran est blanc de dégâts.
- Dernière minute : tout est repris.

Cette phase de toute-puissance est indispensable. C'est le souvenir que le joueur emporte, et c'est ce qu'il essaie de retrouver en relançant.

### Le feedback par kill

Quinze minutes ne passent que si chaque seconde est satisfaisante.

- Temps d'arrêt de 2 à 3 frames sur les gros impacts.
- Chiffres de dégâts qui jaillissent, couleur distincte pour les critiques.
- Son de ramassage dont la hauteur monte à chaque gemme d'une même volée, et retombe après un silence.
- Tremblement d'écran très court, très faible, cumulatif quand ça part en masse.
- Cadavres au sol pendant quelques secondes, effacés progressivement, pour que la salle porte la trace du carnage.

Le moment de plaisir maximal du genre n'est pas le kill, c'est **l'aimant** : deux cents gemmes qui convergent d'un coup avec une montée sonore. Objet dédié, apparition régulière, déclenchement au choix du joueur.

### La lisibilité de l'échec

Le joueur doit toujours pouvoir se raconter pourquoi il est mort : « j'ai voulu traverser au lieu de contourner », « j'ai pris le passif défensif trop tard », « deux armes de zone et rien contre les élites ». S'il meurt sans comprendre, il attribue ça au jeu et il s'arrête. S'il comprend, il a une hypothèse à tester — et une hypothèse à tester est exactement ce qui déclenche la run suivante.

Conséquence directe sur le rendu : même à cent ennemis à l'écran, **le joueur et les projectiles ennemis restent visibles**. Contour lumineux sur le personnage, télégraphes en couleur réservée, ennemis désaturés. Un jeu illisible est un jeu abandonné à la troisième mort injuste.

### La méta-progression, à manier avec précaution

Le piège classique : des améliorations permanentes achetées avec l'or (+5 % de dégâts, +10 PV). Ça marche très fort pendant dix heures, puis ça détruit le jeu — la difficulté se résout par le grind, les runs deviennent identiques, et le joueur qui a tout acheté n'a plus de raison de revenir.

Ce qui tient sur la durée, ce sont les déblocages qui **ajoutent des options** plutôt que de la puissance : une arme qui entre dans le pool et redistribue toutes les builds, un personnage avec une règle différente (démarre sans arme ; ne peut pas se soigner ; monte plus vite mais prend double dégâts), un lieu de plus dans la liste. Chaque déblocage rouvre le jeu au lieu de l'aplatir.

Garder un petit socle d'améliorations permanentes pour les deux ou trois premières heures, plafonné bas. C'est la rampe d'accès, pas le moteur.

### La frontière

Tout ce qui précède crée l'envie de rejouer en rendant le jeu bon. Il existe une autre famille de mécaniques — jauges d'énergie, récompenses de connexion quotidienne, monnaie premium, timers — qui crée de l'envie en rendant l'absence désagréable. Elle marche aussi, mais elle produit des joueurs qui reviennent en s'en voulant, et c'est un jeu dont on parle mal. Vampire Survivors n'a rien de tout ça et s'est vendu à dix millions d'exemplaires.

---

## 3. Le déroulement d'une run

### Des segments, pas une arène

Quatre à cinq lieux de trois à quatre minutes, avec une build qui traverse les segments. Le total reste de l'ordre du quart d'heure.

### La sortie se gagne

Le joueur dispose d'un moyen d'échapper à la difficulté : courir vers la porte. Or l'XP est le moteur de tout — si on saute de lieu en lieu en dix secondes, on arrive au quatrième avec une arme au niveau 1 et le jeu se casse en deux.

La porte s'ouvre donc après un objectif : un temps de survie, un compteur de kills, trois points à réamorcer, une élite à abattre. Une fois l'objectif rempli, le joueur choisit **quand** partir — rester pour farmer ou partir avant que ça déborde est un des meilleurs choix du jeu, et le score du chapitre 6 lui donne une mesure.

### Le temps mort à la porte

Entre deux lieux, une pause courte : choix d'amélioration, gestion des consommables, et éventuellement choix de destination. C'est le seul endroit où l'on ouvre un écran ; jamais pendant l'action.

### Le choix de branche

Ce qui est inconnu n'est pas la position de la porte, c'est **ce qu'il y a derrière**. Quand une campagne bifurque, deux portes proposent chacune un aperçu court — « quartier, dense », « parking, sombre, réserve de caisses » — et le joueur tranche. C'est ce qui produit la phrase « cette fois je pars sur la branche parking », et donc la run suivante.

---

## 4. Le combat : un champ, pas mille pathfindings

La bonne structure, c'est le **flow field**. Un seul BFS depuis le joueur sur la grille de tuiles, recalculé toutes les 5-6 frames. Chaque cellule stocke deux choses : la distance au joueur et le vecteur vers la cellule voisine la plus proche. Un ennemi ne calcule rien, il lit la cellule sous ses pieds.

Coût : un BFS sur 128×128 cellules, quelques dizaines de microsecondes, une fois pour tout le monde.

Ordre de grandeur des effectifs, calé sur le tampon interne de 960×540, soit une quinzaine de tuiles visibles : **60 à 100 ennemis à l'écran** en régime normal de fin de segment, **150 en pic** sur une vingtaine de secondes, **250 à 300 entités vivantes** au total en comptant ce qui approche hors champ. Au-delà, ce n'est plus une horde mais un mur uni : les profils cessent d'être distinguables, et avec eux la lisibilité de l'échec. Le levier de pression n'est de toute façon pas le nombre, c'est la vitesse relative et la fermeture des angles — vingt ennemis qui coupent une sortie font plus peur que deux cents qui suivent en file. Le contournement d'obstacles est gratuit — piliers, rayonnages, tourniquets se retrouvent dans le champ sans une ligne de code d'évitement. C'est ce qui rend le décor urbain jouable là où les survivors classiques se contentent d'un terrain vide.

Sur les lieux étirés en longueur, le champ n'est calculé que sur une **fenêtre autour du joueur** : un ennemi à quarante tuiles n'a besoin d'aucune précision.

Le champ de distance sert aussi aux ennemis à distance : au lieu de descendre le gradient, ils se stabilisent sur une isodistance. Un seul champ, tous les comportements.

Reste la séparation, pour éviter que 200 monstres s'empilent sur un pixel. La méthode classique (spatial hash + voisinage) coûte cher. L'alternative, bien plus rapide : chaque ennemi incrémente sa cellule dans une grille de densité, et on soustrait le gradient de densité au vecteur du flow field. Deux passes O(n), pas de requête de voisinage.

```go
func (e *Enemy) desiredDirection(champ *FlowField, densite *DensityGrid) Vector {
    cx, cy := champ.Cell(e.X, e.Y)

    attirance := champ.Direction(cx, cy)
    repulsion := densite.Gradient(cx, cy)

    d := attirance.Sub(repulsion.Scale(e.profile.SeparationWeight))

    if e.profile.Tangential != 0 {
        d = d.Add(attirance.Perp().Scale(e.profile.Tangential))
    }

    return d.Normalize()
}
```

Le champ `Tangential` suffit à transformer un suiveur bête en flanqueur : il descend le gradient tout en dérivant sur le côté, ce qui referme progressivement le cercle autour du joueur.

### L'apparition

Les ennemis apparaissent **hors du champ de vision**, sur un anneau autour du joueur, à trois tuiles au-delà du bord de l'écran. Jamais dans le champ : voir une créature se matérialiser détruit la crédibilité de la horde et rend une mort incompréhensible.

Sur une carte fermée, aucune position de l'anneau n'est parfois valide — le joueur est dans un cul-de-sac, ou adossé à un mur de niveau. Dans ce cas **l'apparition est abandonnée**, pas déplacée : plutôt aucune créature qu'une créature surgie d'un mur.

Le budget de pression correspondant n'est pas perdu pour autant : il est reporté au tick suivant, où l'anneau aura peut-être une position libre. Sans ce report, un couloir étroit deviendrait un abri où la pression tombe à zéro, ce qui est exactement ce que la conception cherche à éviter.

### Rien ne traverse un mur

La poussée de séparation n'est qu'une force parmi d'autres : elle ne décide pas de la position finale. Le déplacement calculé est **projeté sur la grille de passabilité** avant d'être appliqué — s'il mène dans un obstacle, sa composante bloquée est annulée et l'entité glisse le long du mur au lieu de s'y enfoncer. Si les deux composantes sont bloquées, elle ne bouge pas.

Vingt Badauds qui poussent un Vigile contre une cloison ne le font donc pas passer au travers : ils s'entassent derrière lui. C'est le comportement voulu — un couloir bouché par un bloqueur doit rester bouché, c'est tout son intérêt.

Corollaire à assumer : **les entités se chevauchent** quand la place manque. Il n'y a pas de collision dure entre ennemis, seulement la répulsion douce du gradient de densité. Résoudre les chevauchements par un décalage géométrique produirait des tremblements en chaîne dans une foule dense, et pousserait mécaniquement les créatures du bord dans les murs — exactement ce qu'on vient d'interdire.

### Le recyclage de la traîne

Dès lors que le joueur progresse vers une sortie, les ennemis restés derrière ne servent plus à rien. Au-delà d'une distance seuil, l'entité est retirée du pool et réapparaît devant. Sans ça, la traîne grossit et le frame time avec.

### Aucun ennemi ne délibère

Décision, pas oubli. Il n'y a pas d'adversaire qui raisonne, pas d'exploration de positions, pas de niveaux de difficulté d'IA. Ce qui reste à écrire ressemble à des machines à états — trois ou quatre états par profil, en table — et non à de l'intelligence artificielle.

L'intelligence perçue vient d'ailleurs, de trois sources :

- **Le champ de flux**, qui donne à tous le contournement d'obstacles. Un ennemi qui évite un pilier et ressort de l'autre côté a l'air de savoir où il va ; il lit une case.
- **Les profils**, qui sont des jeux de paramètres. Le flanqueur n'a pas de plan : il descend le gradient avec une composante tangentielle, et c'est ce qui referme le cercle. Le sprinteur abandonne le champ pendant sa charge et ne corrige plus — comportement volontairement bête, mais lisible, et c'est la lisibilité qui compte.
- **La composition des vagues**, vrai levier de difficulté. Trois flanqueurs et deux bloqueurs dans un couloir produisent une situation que personne n'a scénarisée.

Une exception viendra plus tard si le jeu prend des **élites** : un ennemi unique qui alterne des schémas d'attaque demande une vraie machine à états et se traite à part, avec ses propres télégraphes.

### Ce qui fait « l'IA » perçue

Pas le pathfinding — la contrainte d'espace. Le joueur fait une seule chose : du kiting. Le travail de design consiste à casser ce kiting de manières différentes.

Le rôle est l'identifiant du moteur ; le nom est de la fiction, et vit dans la table des profils. Chaque profil a aussi sa silhouette propre — recolorer ne suffit pas, un joueur doit lire sa horde d'un coup d'œil.

| Rôle | Nom | Silhouette | Ce qu'il casse |
|---|---|---|---|
| marcheur | **Badaud** | humanoïde | la masse, pure pression |
| sprinteur | **Molosse** | quadrupède | la fuite en ligne droite |
| flanqueur | **Arpenteur** | six pattes hautes | le kiting circulaire |
| cracheur | **Buse** | bulbe posé au sol | le camping dans un coin |
| bloqueur | **Vigile** | colosse épaulé | les goulots et les couloirs |
| éclateur | **Baudruche** | corps gonflé, tête minuscule | le nettoyage à l'aveugle |

Résistance et points de chaque profil : chapitre 6.

- **Le Badaud** : masse lente, il ne fait qu'exister en nombre. Il existe en plusieurs teintes de vêtement — une foule d'un seul bleu se lit comme un bloc uni, alors que six variantes cassent la répétition sans coûter une silhouette de plus. La variante est tirée à l'apparition depuis la graine de la run, donc elle ne casse pas le déterminisme.
- **Le Molosse** : télégraphe une charge (une demi-seconde d'anticipation, un son), puis fonce en ligne droite et ne corrige plus. Sa charge inflige davantage qu'un contact ordinaire — sans cela, charger ne serait qu'un déplacement rapide. Il punit l'immobilité, mais s'esquive latéralement. Le fait qu'il abandonne le flow field pendant la charge est ce qui le rend lisible. **Il n'apparaît jamais seul** : une meute de trois qui charge en décalé impose d'arrêter de reculer en ligne droite, ce qu'un chien isolé n'obtient pas. La taille de groupe est un champ du profil, pas une exception du spawner.
- **L'Arpenteur** : `Tangential` élevé, il coupe la trajectoire de fuite. C'est lui qui donne l'impression que les monstres réfléchissent.
- **La Buse** : seul profil qui blesse à distance, elle se stabilise et tire. Elle punit le camping dans un coin, force à bouger vers le danger. Sans balancement de marche, elle reste identifiable même immobile au milieu d'une horde.
- **Le Vigile** : lent, encaissant, il bouche les goulots. Dans un couloir de supermarché, il transforme une route de fuite en piège.
- **La Baudruche** : explose en mourant. Sa silhouette disproportionnée dit « ne t'approche pas » avant même que le télégraphe ne s'allume. Elle punit le nettoyage à l'aveugle en mêlée.

Six profils suffisent pour tout le jeu. Ce sont des données, pas du code : une structure `EnemyProfile` avec nom, vitesse, résistance en touches, points, poids de séparation, tangentiel, portée, taille de groupe, comportement spécial. Le reste est du mixage de vagues.

### Le scénario de vagues : un budget, pas des compteurs

Point critique dès lors que des lieux sont créés par des tiers. Un scénario qui dit « à 7 min, 120 marcheurs » rendra le premier niveau amateur venu injouable ou vide. Un scénario exprime donc une **pression par seconde** ; le spawner achète des ennemis dans ce budget parmi les profils autorisés, et ça reste cohérent quel que soit le lieu.

```toml
[[phase]]
debut = "0:00"
pression = 8
profils = ["marcheur"]

[[phase]]
debut = "1:30"
pression = 25
profils = ["marcheur", "flanqueur", "sprinteur"]
pic = { a = "2:10", multiplicateur = 3, duree = 25 }
```

Chaque profil a un prix en points de pression. Le spawner remplit, respecte la passabilité, et lâche hors du champ de vision. Bonus : la difficulté globale devient un seul curseur multiplicateur, ce qui donne les modes de difficulté gratuitement.

---

## 5. Les dégâts subis

### Trois sources, une dominante

**Le contact** est le mode principal : le Badaud, le Molosse, l'Arpenteur et le Vigile n'ont pas d'autre moyen. Ils ne portent pas de coups, ils occupent l'espace — les dégâts ne sont donc pas des frappes isolées mais une **pression continue** quand on se laisse encercler.

**Le tir** n'appartient qu'à la Buse. C'est ce qui punit le camping dans un coin, et le seul cas où un projectile ennemi traverse l'écran : il porte une couleur qui n'existe nulle part ailleurs dans la palette.

**L'explosion** de la Baudruche, avec son télégraphe en anneaux croissants.

### Le contact fait mal en continu, avec un plafond

Un dégât par seconde tant que le contact dure, pas un coup unique suivi d'invulnérabilité. C'est ce qui rend l'encerclement mortel et le déplacement obligatoire — la lecture du jeu vient de là, pas d'un compteur de coups.

Mais **le total encaissé par seconde est plafonné**, quel que soit le nombre d'ennemis collés. Sans ce plafond, trente Badauds tuent instantanément, et la mort devient illisible : le joueur n'a rien vu venir et ne peut rien en apprendre. Avec lui, être encerclé est très dangereux mais laisse une fenêtre pour se dégager, ce qui est exactement le moment de jeu recherché.

Corollaire sur le retour : à cette cadence, un son de dégât par tick serait insupportable. Le son se déclenche à l'entrée en contact et se réarme après un silence, l'écran porte le reste — teinte rouge brève, et la jauge qui descend visiblement.

### La vie du joueur

Trois chiffres se tiennent et doivent être posés ensemble, sinon aucun n'a de sens : la vie totale, le plafond de dégâts par seconde, et ce que rend une fiole.

Point de départ proposé : **100 de vie, plafond à 20 par seconde, fiole à 30**. Ce n'est pas un équilibrage, c'est un rapport lisible — encerclé sans se dégager, le joueur tient cinq secondes ; une fiole rend un tiers de sa barre et lui rachète une seconde et demie d'encerclement.

La règle qui compte plus que les valeurs : **la vie ne se régénère pas seule**. Elle ne revient que par les fioles trouvées dans les caisses. Sans cela, attendre devient une stratégie et le joueur prudent finit par jouer un autre jeu que celui qu'on a écrit — or ici la seule ressource véritablement rare doit être la vie.

Une fiole ne dépasse jamais le maximum : le surplus est perdu, ce qui donne au joueur une raison de ne pas la boire tout de suite et fait de ses deux emplacements de consommables une petite décision de plus.

### La charge du Molosse

Elle inflige **plus qu'un contact ordinaire**. Sans cela, charger ne servirait à rien : ce ne serait qu'un déplacement rapide, et le télégraphe n'aurait rien à annoncer.

C'est ce qui donne son sens à la mécanique — une demi-seconde d'anticipation, une trajectoire droite qui ne corrige plus, et une esquive latérale qui annule tout. Le joueur qui reste immobile paie le prix fort ; celui qui se décale ne paie rien.

---

## 6. Résistance, points et score

### La résistance se compte en touches, pas en points de vie

Une valeur absolue de PV ne veut rien dire dans un jeu où l'arme grossit toute la run. La résistance s'exprime donc en **touches de l'arme de base à son premier niveau**, et c'est ce chiffre qui se lit et se règle.

| Profil | Touches | Points | Ce que le chiffre traduit |
|---|---|---|---|
| Badaud | 3 | 10 | il meurt vite, il revient en nombre |
| Molosse | 2 | 25 | fragile, mais il arrive à trois et vite |
| Arpenteur | 4 | 30 | il faut le suivre pendant qu'il tourne |
| Baudruche | 4 | 35 | l'abattre de près est une erreur |
| Buse | 5 | 40 | elle tire de loin, on va la chercher |
| Vigile | 12 | 60 | il bouche un couloir, on le contourne |

Ces valeurs sont un point de départ, pas un équilibrage : elles se règlent à partir du jalon 3, en jouant.

**La résistance monte au fil de la run**, par un multiplicateur adossé à la courbe de pression — sinon la fin de partie n'est qu'un tapis roulant, puisque l'arme a été multipliée par dix. C'est ce multiplicateur qu'on ajuste, jamais les touches de chaque profil, qui restent le rapport entre eux.

### Les points et la tension qu'ils créent

Chaque ennemi rapporte, et le score d'un lieu additionne les points récoltés et un **bonus de temps** : ce qui reste d'un temps de référence quand la porte est franchie.

C'est le seul intérêt réel du score ici. Il met deux envies en opposition directe : rester pour farmer rapporte des points d'ennemis, partir vite rapporte du bonus de temps. Le joueur qui optimise doit trancher à chaque salle, et ce choix — rester ou partir — est déjà, dans le document, le meilleur moment du système de porte. Le score ne fait que lui donner une mesure.

Deux garde-fous. Le score ne doit **jamais contredire la lisibilité** : il s'affiche à la fin d'un lieu et sur l'écran de mort, pas en gros pendant l'action, où l'attention appartient à la horde. Et il ne remplace pas la build comme moteur de rejouabilité — c'est un supplément pour qui veut se mesurer, pas la raison de relancer.

### Le classement, et pourquoi la graine existe

Un score n'est comparable que si les runs le sont. C'est à cela que sert la graine déterministe posée dans les invariants : deux joueurs qui affrontent la même graine sur le même lieu ont eu les mêmes vagues aux mêmes instants.

D'où deux classements possibles, sans serveur : la **graine du jour**, identique pour tout le monde, et le classement par lieu partagé, où celui qui diffuse un niveau diffuse aussi sa graine. Un fichier de niveau tenant dans un message, le défi se partage avec lui.

---

## 7. Les ressources : les caisses

Le joueur casse une caisse **en la traversant**. Aucune touche, aucun conflit avec l'auto-visée, et ça le garde en mouvement.

Trois règles rendent la mécanique juste :

**Un temps de contact.** Pas de destruction au frôlement : la caisse cède après environ un tiers de seconde d'appui, avec une déformation visible pendant le délai. On ne casse pas en passant, on casse en décidant d'y aller.

Ce délai a ses propres images : un cycle d'appui qui boucle tant que le joueur pousse — la caisse s'écrase en s'élargissant —, puis un cycle de rupture qui ne boucle pas et s'achève sur l'épave au sol. Sans ces images, le joueur ne sait pas qu'il casse quelque chose et croit à un blocage.

**Un ralentissement pendant le contact.** C'est le vrai coût de la ressource, et le seul qui compte : ramasser, c'est perdre du terrain. Sans ça, il n'y a plus de choix, on ramasse tout, tout le temps.

**Une distinction ferme entre caisse et obstacle.** Silhouette différente, teinte réservée, liseré lumineux. Si les caisses ressemblent aux piliers, le joueur ne sait jamais ce qui va céder et ce qui va le bloquer.

Le contenu est **visible avant la casse** — icône flottante ou liseré coloré. Sinon le joueur casse tout systématiquement et ce n'est plus un choix, c'est une corvée.

Contrainte technique : une caisse est **bloquante dans le flow field**, sinon les ennemis la traversent alors que le joueur non. Sa destruction déclenche un rafraîchissement local du champ, pas un BFS complet. Cas de jeu à préserver : un joueur acculé peut se dégager en cassant une caisse, à condition que le délai reste court.

### Les obstacles destructibles

Un mur ne cède **jamais** sous la pression d'une horde : la géométrie n'est pas négociable, sinon le Vigile perd son intérêt, la signalétique son objet, et l'éditeur ne peut plus rien valider — une carte qui change en cours de partie n'est plus vérifiable.

En revanche, l'auteur d'un niveau peut poser des **obstacles fragiles**, prévus dans la topologie et donc validables comme le reste :

| Obstacle | Touches | Ce qu'il ouvre |
|---|---|---|
| grille de ventilation | 3 | un raccourci, presque gratuitement |
| vitrine | 5 | une boutique, et on voit à travers avant de casser |
| cloison de placo | 8 | une réserve, un mur qui n'en était pas un |
| rideau de fer | 20 | une vraie décision : vingt touches sous la horde |

Ils se cassent **au tir**, contrairement à la caisse qui cède au contact — casser un mur en s'appuyant dessus n'aurait aucun sens, et cela permet de les distinguer d'un coup d'œil dans le jeu comme dans le code.

Deux conséquences reprises de la caisse : ils sont bloquants dans le champ de flux tant qu'ils tiennent, et leur destruction déclenche un rafraîchissement local du champ, pas un BFS complet. Chacun laisse une ruine basse, franchissable, qui garde la trace de ce qui a été ouvert.

L'intérêt de jeu est le prix : casser un rideau de fer coûte des secondes pendant lesquelles la horde arrive. Fuir par là est un choix, pas une porte de sortie gratuite.

### Les éclats

Un objet qui se détruit projette des **éclats de sa matière** : bois, verre, plâtre, métal, chair. Une explosion générique serait une erreur — c'est la matière qui dit au joueur ce qu'il vient d'ouvrir, et une vitrine qui se casse en poussière de plâtre ne se lit pas.

Ce ne sont pas des animations mais des **particules** : trois formes par matière, minuscules, que le moteur émet en nombre et déplace sur une parabole avec sa propre rotation et sa propre durée. Le principe est celui déjà retenu pour l'objet qui jaillit d'une caisse — la trajectoire appartient au moteur, le générateur ne fournit que les formes.

Deux effets font exception et sont bien des animations, parce qu'ils ont une géométrie propre. L'**étincelle** d'impact, trois images très courtes, qui ne dit pas ce qui a été touché mais que le tir a porté : c'est le retour qui manque le plus quand on tire sans le voir. Et le **souffle** de la Baudruche, cinq images d'anneaux qui s'élargissent et s'étalent dans le plan du sol — francs, jamais dégradés, un fondu lissé virerait à la tache brune une fois quantifié.

### Ce qui sort des caisses

Pas d'inventaire. Une arme ramassée s'ajoute directement à la build, comme un niveau gagné. Les soins sont des consommables à usage unique, **deux ou trois emplacements maximum**, une touche pour boire, aucun menu. Le joueur a une décision — boire maintenant ou garder — pas une gestion.

Toute gestion plus lourde se fait au temps mort de la porte, jamais pendant l'action.

---

## 8. Trouver la sortie : la signalétique

La porte n'est pas connue à l'avance, mais l'errance est mortelle dans un jeu de horde. La solution est **la direction sans la position**, portée par le décor lui-même : flèches de sortie de secours, panneaux « caisses », marquage au sol, fléchage de quai. C'est thématiquement gratuit — ces lieux sont déjà couverts de signalétique — et ça donne au joueur une raison d'apprendre à lire chaque décor.

Principe de fonctionnement : **le relais**. Un panneau donne la direction du prochain carrefour, jamais celle de la porte finale. On avance de repère en repère ; l'incertitude reste, l'errance non.

La boussole n'est qu'un filet de sécurité : elle apparaît discrètement en bord d'écran après quarante secondes sans progression vers la sortie. Le joueur compétent ne la voit jamais.

Côté données, la signalétique est une propriété des pièces : chaque pièce déclare ses emplacements de panneaux, et **l'orientation est calculée au chargement** depuis le chemin réel vers la sortie. L'auteur ne règle rien, il pose ses pièces et les panneaux se retournent tout seuls. C'est ce qui évite les niveaux communautaires avec des flèches qui mentent.

---

## 9. Armes et montées de niveau

### Une seule arme, qui évolue

Décision arrêtée. Le Survivant a un cycle d'attaque et une arme ; un lance-flammes ou une orbite demanderaient des animations que le générateur ne sait pas produire, et chaque arme supplémentaire coûterait un cycle par direction.

Le joueur garde donc **la même arme du début à la fin**, et ce sont des passifs qui la transforment : un projectile de plus, cadence, portée, perforant, ricochet, tir en éventail. La progression devient un chemin dans un arbre plutôt qu'une collection, et les **synergies se déclarent entre passifs** — « trois projectiles » plus « éventail » donne une gerbe, « perforant » plus « portée » donne un rail. Cinq ou six recettes suffisent, et elles motivent les runs suivantes plus que n'importe quelle progression méta.

L'évolution se fait **en nombre plutôt qu'en nature** : un projectile qui devient trois, pas un projectile qui change de comportement. Plus lisible, plus facile à équilibrer, et beaucoup moins de cas particuliers dans le code de collision.

**Le revers à assumer** : le contraste des patterns disparaît, alors que c'est un moteur de rejouabilité du genre. Il se récupère par deux voies qui ne coûtent aucune animation de personnage — les armes lourdes ramassées dans les caisses, qui sont un effet et non une pose, et les effets qui ne partent pas du personnage : zone au sol, orbite, onde de choc, dessinés au sol ou autour du joueur.

### La visée

Le tir de base est **automatique**, et il vise **le plus proche dans un cône devant**, autour de la direction de visée. Angle à caler entre 90 et 120 degrés : plus étroit, le joueur corrige son orientation au lieu de fuir ; plus large, on retombe sur « le plus proche » et le choix perd son sens.

**Cône vide, pas de tir.** Jamais de repli sur une cible derrière : c'est ce qui rend le comportement lisible — on comprend pourquoi ça ne part pas.

Le sprite s'oriente donc sur la **visée**, pas sur le déplacement. Les 8 directions étant fournies, reculer en tirant vers l'avant se lit immédiatement.

Contrainte : les projectiles ne ciblent **jamais** le mobilier. Si l'auto-visée choisit une caisse plutôt qu'un sprinteur qui charge, le joueur meurt sans comprendre.

### Armement de base infini, armes lourdes à charges

Une économie de munitions générale n'aurait pas de sens sur un tir automatique : le joueur subirait une jauge qu'il ne contrôle pas, avec une mort par panne sèche illisible, des caisses obligatoires — donc plus d'arbitrage, juste une corvée — et un pic de puissance de fin de run impossible.

L'armement de base est donc **infini**. C'est lui qui monte de niveau et porte la build.

Les armes lourdes sont **à charges** : lance-flammes, fusil à pompe, grenades, tourelle. Trouvées dans les caisses, utilisées un nombre limité de fois, jetées à vide. C'est ce qui donne un enjeu aux caisses et fait de chaque trouvaille un événement, sans jamais laisser le joueur sans défense.

**Elles se déclenchent à la touche.** Le joueur ne gère rien en continu — le socle tire tout seul — mais il décide du moment de sa grenade. C'est plus simple à coder qu'un déclenchement conditionnel qui devinerait quand la situation le mérite, plus lisible pour le joueur, et ça ajoute une décision au lieu d'une gestion.

Règles associées :

- **Affichage en pastilles**, pas en chiffres — trois pastilles qui s'éteignent se lisent en vision périphérique, un « 3/5 » demande de regarder. La dernière pulse. À l'épuisement, disparition de l'interface et son sec, pas de message.
- **Deux emplacements maximum**, une touche chacun. Une troisième arme ramassée propose l'échange sur place — aperçu au sol, on passe dessus pour prendre, on contourne pour laisser. Aucun menu.
- **Hors du système d'XP.** Les lourdes ne montent pas de niveau, ce qui garde la table d'évolutions lisible. En revanche les passifs de dégâts et de zone s'y appliquent, sinon elles deviennent inutiles en fin de run.
- **Une trouvaille toutes les 60 à 90 secondes environ.** Plus rare, le joueur oublie la mécanique et ne l'apprend jamais. Plus fréquent, elle remplace le socle infini.

Variante si les consommables gênent l'équilibrage : de la surchauffe plutôt que des charges. Même rythme, même retenue, pas de mort par assèchement.

### Ce qui reste à décider en jouant

Les valeurs — dégâts, cadences, portées, coût des passifs — ne se conçoivent pas sur le papier : elles tiennent au ressenti et seront jetées au premier essai. Elles se fixent à partir du jalon 3, une fois la boucle jouable, et vivent dans une table de données, pas dans le document.

Une option reste ouverte si le tir manuel manque au jalon 3 : garder l'automatique et ajouter un **tir d'appoint** sur touche, avec temps de recharge. Le joueur passif ne perd rien, le joueur actif gagne un peu.

---

## 10. Le rendu

### Les directions

Le pack de personnages retenu fournit 8 directions (voir chapitre 12), ce qui lève la contrainte initiale de 4 poses. Si un archétype devait être dessiné à la main, la solution de repli tient toujours : quatre orientations sur les diagonales écran (NE, SE, SO, NO), dont deux obtenues par miroir horizontal — on ne dessine alors que dos et face.

Dans tous les cas : orienter le sprite sur la direction de **visée**, pas de déplacement. Le joueur recule en tirant vers l'avant, et ça se lit immédiatement.

### La grille et les tailles

Tout découle des sprites de 64×64 du pack de personnages : la tuile de sol fait **64×32**, projection 2:1, origine au centre du losange.

Pour un objet couvrant plusieurs tuiles, `largeur = (tx + ty) × 32` et l'emprise au sol `hauteur = (tx + ty) × 16`.

| Élément | Taille image | Élévation au-dessus du sol |
|---|---|---|
| Tuile de sol | 64×32 | 0 |
| Obstacle bas (gondole, banc, caddie) | jusqu'à 96×62 | 24 max |
| Mur plein | 64×96 | 64 |
| Caisse cassable | 32×32 | 16 |
| Personnage et ennemis | 64×64 | appui en (32, 63) |
| Gemme d'XP | 10×8 | — |

La règle des **24 pixels** est celle qui compte : au-delà, l'objet masque un personnage de 64 et devient un piège visuel. Ce qui la dépasse — murs, piliers, véhicules, wagons, immeubles — n'est pas un obstacle à contourner en pleine action mais une limite de zone, et doit passer en semi-transparence quand le joueur est derrière.

**Résolution interne : 960×540**, agrandie en entier vers la fenêtre. Un tampon de 480×270 ne montrerait que 7 tuiles de large, bien trop serré pour voir la horde arriver ; 960×540 en donne une quinzaine et se multiplie par 2 pour du 1080p, donc pixels carrés garantis.

### Les trois hauteurs

Chaque tuile porte une hauteur parmi trois : sol, obstacle bas qu'on voit par-dessus, mur plein. C'est ce qui permet de juger la lisibilité d'une salle **avant** de la lancer, et c'est indispensable à l'éditeur en vue de dessus (voir plus bas).

### Le tri en profondeur

Le rendu iso a besoin d'un tri par `Y` écran : un tri par compartiments, pas un `sort.Slice` général à chaque frame.

**À égalité, la clé doit être totale et stable.** Deux entités sur la même case sont départagées par leur `X` écran, puis par leur identifiant — index et génération. Sans ce dernier critère, l'ordre dépend du parcours du bassin, qui change à chaque suppression par échange : deux sprites superposés se relaieraient au premier plan d'une image à l'autre, et le scintillement se voit immédiatement.

**Le joueur passe devant tout ce qui partage sa profondeur.** C'est une exception assumée à la règle de tri : perdre son personnage sous un empilement d'ennemis est la pire chose qui puisse arriver à la lisibilité, et cela survient précisément au moment où l'on est encerclé, c'est-à-dire quand il faut voir clair.

### La caméra

Elle suit le joueur, centrée, **en pixels entiers** — un déplacement en flottants ferait scintiller le pixel art, ce que la règle du chapitre du style interdit déjà.

Aux abords d'un bord de lieu, elle **se bloque sur les limites** plutôt que de découvrir du vide. Le joueur se décentre alors, ce qui est le comportement attendu : on voit le mur arriver, on comprend qu'on est acculé, et l'écran ne montre jamais un hors-champ noir qui n'existe pas.

Deux cas particuliers. Si un lieu est plus petit que le tampon de 960×540 — ce qui arrive avec une pièce unique — la caméra ne se bloque pas, elle **centre le lieu** une fois pour toutes. Et lors du passage d'un lieu au suivant, elle ne se déplace pas : le fondu de la porte couvre le saut, ce qui évite un travelling de plusieurs centaines de tuiles.

Un lissage est possible — la caméra rattrape le joueur en quelques images plutôt que de le coller — mais sa position finale est arrondie au pixel avant tout dessin. Sans cet arrondi, le lissage réintroduit exactement le scintillement qu'on cherche à éviter.

---

## 11. L'éditeur

### Vue de dessus pour éditer, iso pour jouer

L'édition se fait en 2D orthogonale, la partie se joue en iso. Trois raisons :

- une grille carrée se survole sans conversion de coordonnées, la souris tombe sur la bonne cellule au pixel près ;
- pas d'occlusion — en iso, un mur haut cache ce qu'on vient de poser derrière, et on passe son temps à déplacer la caméra ;
- la vue de dessus donne la **lecture topologique** : boucles, culs-de-sac et goulots se voient d'un coup d'œil, et c'est exactement l'information dont l'auteur a besoin.

Bénéfice secondaire : aucun renderer iso à écrire pour l'éditeur.

### Deux granularités

**Mode pièces** (par défaut). L'auteur glisse des pièces préfabriquées sur une grille : rayon de supermarché, zone de caisses, couloir large, salle de cinéma, palier d'escalator, réserve. Chaque pièce occupe un carré fixe de tuiles — 16×16 est une bonne base — et embarque son décor, sa passabilité, ses hauteurs et ses ancrages.

**Mode tuiles** (« ouvrir la pièce »), **différé**. Pour fabriquer des structures tordues, on entre dans une cellule et on peint tuile par tuile — avec les tuiles fournies par le jeu, jamais avec des images importées. Ce qui en sort est une pièce comme une autre, réutilisable, enregistrée dans le niveau.

Ce mode n'est pas au programme de la première version : le noyau de l'éditeur se réduit alors à la pose de pièces, ce qui allège sérieusement le chantier. Le champ reste prévu dans le format pour ne pas avoir à le rétro-ajouter.

Les deux modes partagent le même noyau : palette, pose, aimantation, annuler/refaire. Ce noyau est écrit **une fois**, paramétré par ce qu'on pose, sinon il y a deux éditeurs à maintenir et ils divergent.

### Les connecteurs

C'est le cœur du système. Chaque pièce déclare, sur ses quatre côtés, ce qu'elle offre : mur plein, ouverture centrée, ouverture large, passage double. Deux pièces voisines ne se posent que si leurs côtés en vis-à-vis sont compatibles ; sinon l'éditeur refuse ou insère une pièce de transition.

Conséquence majeure : **la connexité est garantie par construction**. Le piège classique du contenu communautaire — une carte coupée en deux qui fait tourner le flow field en rond — disparaît.

Pour une pièce peinte à la main, les connecteurs sont **déduits automatiquement** de la passabilité des bords. Personne ne les remplirait correctement à la main.

### La rotation

En iso, pivoter une pièce de 90° n'est pas gratuit : chaque tuile de mur doit exister dans les quatre orientations. Deux options honnêtes — concevoir le tileset avec les quatre variantes dès le départ, la rotation devenant un simple remappage d'index ; ou interdire la rotation et dessiner plus de pièces. La première coûte une journée de tileset, la seconde coûte des pièces à vie. **À trancher avant de dessiner la moindre tuile.**

### Le retour en direct

Chaque pièce connaissant son aire ouverte et ses ancrages, l'éditeur calcule pendant l'édition, sans lancer la partie :

- l'aire jouable totale, donc la **pression maximale supportable** avant embouteillage ;
- les culs-de-sac, surlignés — dans ce genre, un couloir sans issue est une mort injuste, pas une difficulté ;
- l'existence de **boucles** : le joueur doit pouvoir tourner en rond, c'est la mécanique de base du kiting. Un niveau en arbre ne se joue pas. Un parcours de graphe sur la grille de pièces donne le nombre de cycles indépendants ;
- l'atteignabilité de la sortie depuis l'entrée, et la présence d'au moins une boucle sur le trajet ;
- la couverture des points d'apparition par rapport à la surface.

Trois jauges en haut de l'écran, mises à jour à chaque pose. C'est ce qui fait la différence entre un outil de conception et un jouet.

### Le test

Deux modes, pas un bouton :

- **lancer la partie ici**, touche unique, retour à l'éditeur au même endroit — c'est celui qui fait itérer ;
- **caméra libre en iso**, sans ennemis, pour juger l'esthétique du décor sans mourir.

Le premier est indispensable, le second se code en dix minutes une fois le premier fait.

### L'édition de campagne

Un deuxième niveau d'édition, au-dessus des pièces : un graphe de lieux. Chaque nœud est un lieu (jeu de pièces, disposition, scénario de vagues, objectif de sortie, densité de caisses), chaque arête est une porte. Un graphe linéaire donne une campagne classique, un graphe qui bifurque donne des choix de branche.

Validation : tout nœud atteignable depuis l'entrée, au moins un chemin complet, pas de nœud orphelin.

### Le générateur, presque gratuit

Une fois les connecteurs en place, poser des pièces au hasard en respectant les compatibilités est le même algorithme que l'aimantation, sans la souris. On obtient un mode « lieu inédit à chaque run » sans écrire de générateur dédié, et un bouton « proposer une disposition » dans l'éditeur, que l'auteur retouche ensuite.

### Le vrai coût

Pas le code : le dessin des pièces. Compter une quinzaine de pièces par thème pour que les niveaux ne se ressemblent pas, conçues en pensant kiting — largeur de couloir au moins trois fois la boîte du joueur, piliers assez espacés pour tourner autour, toujours deux sorties.

---

## 12. Les formats de fichiers

Principe directeur : **le contenu utilisateur est de la donnée, jamais du code.** Pas de script, pas de greffon, pas d'expression arbitraire dans un paquet téléchargé. Un auteur ne compose que dans un vocabulaire fermé — profils d'ennemis existants, événements existants, pièces existantes. Les greffons, s'il y en a, restent réservés au contenu de première partie chargé depuis le binaire.

Trois champs à prévoir dès la première version, sinon ils ne pourront plus être ajoutés proprement : `version_format`, pour migrer sans casser l'existant ; une **graine déterministe** par run, qui donne les classements par lieu et le partage de run ; et `empreinte_jeu_pieces`, la somme de contrôle décrite plus bas.

### Une pièce

```toml
[piece]
identifiant = "rayon_long"
jeu = "supermarche"
version_format = 1
taille = [16, 16]
aire_ouverte = 0.62

cotes = { nord = "ouverture", est = "mur", sud = "ouverture", ouest = "mur" }

[[ancrages]]
type = "apparition"
position = [2, 14]

[[ancrages]]
type = "signaletique"
position = [8, 0]

[[ancrages]]
type = "caisse"
position = [11, 6]
```

La grille de tuiles de la pièce (indices d'atlas, passabilité, hauteurs) vit dans un fichier binaire ou JSON compact à côté. La passabilité et les hauteurs sont **dérivées des propriétés des tuiles au chargement**, jamais saisies à la main : l'auteur ne sait même pas que le flow field existe.

### Un jeu de pièces (thème)

```toml
[jeu_pieces]
identifiant = "supermarche"
nom = "Supermarché"
auteur = "stephane"
version = "1.2.0"
version_format = 1

[atlas]
fichier = "atlas.png"
taille_tuile = [64, 32]
variantes_rotation = true

[ambiance]
musique = "neons.ogg"
teinte = "#c8d4e0"
luminosite = 0.8
```

### Un lieu

Un lieu n'est pas une carte, c'est une **liste de pièces posées**. Quelques centaines d'octets.

```json
{
  "version_format": 1,
  "identifiant": "supermarche_nuit",
  "jeu_pieces": "supermarche@1.2",
  "empreinte_jeu_pieces": "a41f7c92",
  "grille": [4, 3],
  "pieces": [
    { "id": "entree_caisses", "x": 0, "y": 0, "rotation": 0 },
    { "id": "rayon_long",     "x": 1, "y": 0, "rotation": 1 },
    { "id": "carrefour",      "x": 2, "y": 0, "rotation": 0 }
  ],
  "pieces_personnalisees": [],
  "scenario": "standard_4min",
  "objectif_sortie": { "type": "kills", "valeur": 250 },
  "densite_caisses": 0.4
}
```

`pieces_personnalisees` embarque les pièces peintes à la main, quand il y en a — et il reste vide tant que le mode tuiles n'est pas implémenté. C'est le seul cas où le fichier grossit ; même alors, une pièce de 16×16 tuiles compressée pèse quelques centaines d'octets.

`empreinte_jeu_pieces` est une somme de contrôle du jeu de pièces, pas seulement son numéro de version. Sans elle, une pièce retouchée sans changement de numéro produit des niveaux qui **se chargent en silence avec une géométrie différente** de celle qu'a construite leur auteur. Quelques octets de plus, et le message devient explicite au lieu d'être trompeur.

Un niveau qui ne référence que des pièces officielles ne contient donc que des identifiants et des positions : le destinataire possède déjà les tuiles, les objets et les images. Mesuré sur un supermarché de douze pièces, ça fait 1189 octets lisibles, 902 compacts, **548 caractères une fois compressé et encodé en base64** — copiable dans un message.

### Une campagne

```json
{
  "version_format": 1,
  "nom": "Descente",
  "auteur": "stephane",
  "entree": "supermarche_nuit",
  "noeuds": [
    { "id": "supermarche_nuit", "sorties": ["quartier_est", "quartier_ouest"] },
    { "id": "quartier_est",     "sorties": ["parking_niveau3"] },
    { "id": "quartier_ouest",   "sorties": ["parking_niveau3"] },
    { "id": "parking_niveau3",  "sorties": [] }
  ]
}
```

### La cuisson au chargement

À l'ouverture d'un lieu : assemblage des pièces en une seule tilemap, dérivation de la grille de passabilité et des hauteurs, collecte des ancrages, orientation de la signalétique depuis le chemin réel vers la sortie.

À partir de là, **le moteur ne sait plus que le lieu était modulaire** : le flow field tourne sur une grille ordinaire. Les lieux officiels sont faits de pièces exactement comme les lieux communautaires — même chemin de code, une seule chose à déboguer, et l'éditeur est utilisé en permanence par son auteur, ce qui garantit qu'il sera bon.

### La validation au chargement

Un lieu invalide est rejeté avec un message clair, il ne fait pas planter la run :

- identifiants de pièces tous connus, sinon refus propre ;
- empreinte du jeu de pièces conforme, sinon avertissement explicite plutôt qu'un chargement silencieux ;
- zone jouable connexe (garantie par les connecteurs, revérifiée) ;
- sortie atteignable depuis l'entrée, au moins une boucle sur le trajet ;
- au moins un point d'apparition atteignable depuis toute position du joueur ;
- aire ouverte suffisante pour la pression maximale du scénario ;
- version de format supportée.

Les contrôles de graphe utilisent le BFS déjà écrit pour le flow field. C'est le même code.

---

## 13. Distribution

Étape 1, la plus robuste : un dossier `lieux/` scanné au démarrage, une archive par lieu avec une extension propre au jeu, l'identifiant faisant office de clé. Glisser-déposer et ça marche. C'est déjà tout ce dont une communauté a besoin pour échanger sur un salon Discord.

Mieux : un lieu qui ne référence que des pièces officielles tient dans quelques centaines d'octets. Encodé en base64, il se colle dans un message ou tient dans un QR code. Pas de serveur, pas de somme de contrôle, pas d'atelier — on copie une chaîne.

Étape 2, si le besoin apparaît : un catalogue, Steam Workshop ou un simple index JSON servi par le site avec téléchargement et somme de contrôle. Le format de paquet ne change pas, seul le transport change.

---

## 14. Les ressources graphiques et sonores

Le format suppose qu'existent un atlas de tuiles, des feuilles de sprites et de la musique. Rien de tout ça n'est produit par le moteur, l'éditeur ou le format : c'est le poste le plus coûteux du projet, celui qui ne bénéficie d'aucun raccourci technique.

Ordre de grandeur : une quinzaine de pièces par thème, cinq thèmes, plus six archétypes d'ennemis et un joueur.

### Le style : pixel art

Décision arrêtée. Elle est esthétique — le registre des jeux d'arcade et des isométriques des années 90 — mais elle a trois conséquences pratiques qui pèsent lourd :

- **Les décors manquants deviennent faisables soi-même.** Un rayonnage de supermarché en 64×32, ce n'est pas de l'illustration, c'est de la géométrie et trois teintes. Rayons, caddies, tourniquets, distributeurs : des soirées de travail plutôt qu'une commande.
- **La taille des sprites reste raisonnable** à 800 entités affichées, ce que du 3D précalculé haute résolution ne permettrait pas.
- **Le mélange de sources devient possible**, à condition de tenir la palette (voir ci-dessous).

Corollaire : la voie du rendu 3D précalculé (modèles Mixamo ou low-poly rendus dans Blender) est écartée. Elle reste notée ici comme option de repli si le bestiaire devait grossir au point que le dessin à la main ne suive plus.

### Les trois règles du pixel art

À poser dès le premier asset, pénibles à corriger ensuite.

**Une palette fermée.** Trente-deux couleurs, pas plus, extraites du pack de personnages retenu et fixées dans un fichier. Tout asset entrant est recoloré dessus. C'est le seul moyen de rendre cohérents des paquets venant d'auteurs différents et ses propres tuiles.

**Aucun filtrage à l'affichage.** Échantillonnage au plus proche voisin, et caméra déplacée en pixels entiers, jamais en flottants. Le scrolling sous-pixel fait scintiller le pixel art : c'est le défaut qui trahit immédiatement un jeu bâclé.

**Une résolution interne fixe.** Rendu dans le tampon de 960×540 arrêté au chapitre 10, agrandi en entier vers la fenêtre. Pixels carrés à toutes les tailles d'écran, et surtout décision de game design déguisée en détail technique : ce tampon détermine à quelle distance le joueur voit arriver la horde. C'est sa fixité qui est la règle de pixel art ; sa valeur, elle, est déjà tranchée.

### La lisibilité en masse

Point de vigilance sans précédent à copier : les jeux rétro n'ont jamais affiché autant de sprites simultanés. Contours foncés sur les ennemis, teinte réservée au seul personnage joueur, projectiles ennemis dans une couleur qui n'existe nulle part ailleurs dans la palette. Une palette fermée rend cette discipline tenable ; des sprites rendus l'auraient rendue impossible.

### Les tuiles

**Générées**, par `outils/decor_iso.py` : une soixantaine de formes réparties en six thèmes, à partir d'une seule fonction qui calcule la face supérieure en coordonnées de tuile. Le compte exact est celui du manifeste, qui est le seul endroit où il ne peut pas dériver.

Un générateur donne une cohérence gratuite là où assembler des dessins d'auteurs différents mélange les palettes et les conventions d'angle. Et une forme se corrige dans le script, pas dans le PNG.

Règle qui vaudra encore si le jeu passe un jour à du pixel art dessiné : une seule taille de tuile, une seule projection, une seule palette.

### Les personnages

**Générés, comme le décor**, par `outils/figurines.py` : une créature est un
empilement de volumes isométriques, et rien d'autre. Six gabarits donnent des
silhouettes distinctes — bipède, quadrupède, rampant, bulbe, colosse, gonflé —
parce que recolorer ne suffit pas : un joueur doit lire sa horde d'un coup
d'œil. Le colosse est celui du Vigile, dont le chapitre 4 dit qu'il se
reconnaît à ses épaules ; sur un bipède ordinaire il ne bouchait pas un couloir
à l'œil, seulement dans la grille.

Les huit orientations viennent de la place des membres et du regard, pas d'une
rotation des volumes : à cette taille un torse pivoté ne se lit pas, alors qu'un
bras avancé se lit tout de suite. Un corps composé de blocs enfilés sur le
vecteur du regard s'oriente donc sans qu'aucun volume ne tourne — c'est ce qui a
réglé le chien, dont la tête sortait du corps dès qu'il changeait de direction.

Un profil peut déclarer plusieurs **teintes de vêtement**. La variante est tirée
à l'apparition depuis la graine de la run, jamais depuis l'horloge, sinon deux
rejeux de la même graine divergent.

**Aucune ressource tierce.** Un générateur est relisible en pull request, un PNG
ne l'est pas : c'est ce qui rend les créatures **contribuables**, et ce qui évite
toute question de licence sur ce qui entre dans le jeu.

### Les outils de fabrication

Tous déterministes — relancés, ils produisent des fichiers identiques, donc versionnables.

Cinq modules, découpés par nature, plus `ressources.py` qui les enchaîne et
contrôle ce qu'ils ont produit. **`primitives_iso.py`** est le noyau : tout part d'une fonction `volume(tx, ty, elevation, matiere)` qui calcule la face supérieure **en coordonnées de tuile**, ce qui garantit l'escalier régulier du 2:1 quelle que soit l'emprise, y compris fractionnaire — `ty=0.15` donne une cloison mince, une barrière ou un écran de cinéma. Il ne connaît aucune forme du jeu.

**`decor_iso.py`** génère les lieux : une soixantaine de formes réparties en commun, supermarché, parking, quartier, cinéma, station.

Fonctions de détail : `grain` (moucheture), `nervures` (tôle, lattes), `bandeau` (réglette d'étiquettes), `rivets`, `fenetres` (bande vitrée, restreignable à une portion pour un pare-brise), `creuser` (porte ou vitrine renfoncée), `eventrer` (arêtes mangées), `roues` (accrochées au bord inférieur de la silhouette, avec débord), et `contour`, le plus efficace de tous.

Composition : `poser` place un objet **en coordonnées de tuile**, pas en pixels — trois caisses alignées le sont dans le monde, pas seulement à l'image ; `aligner` compose des volumes partageant la même origine, ce qui donne les angles de mur. Un véhicule se déclare en une ligne via `vehicule(longueur, largeur, hauteur, caisse, cabine)` : une seule caisse, la cabine étant une zone teintée et non un bloc rapporté — deux volumes empilés laissent un joint visible et cassent la lecture.

`MATIERES` est le point d'entrée unique de la palette : trois teintes par matière, dessus et deux flancs.

**`figurines.py`** génère les créatures : six gabarits, neuf profils, variantes de teinte, une bande horizontale par cycle et par direction.

**`objets.py`** génère ce qui se ramasse ou se tire : caisse et ses cycles d'appui et de rupture, gemme, fiole, projectiles, armes lourdes en version posée au sol et en icône d'interface. Une caisse se casse et se ramasse — elle appartient au jeu, pas au lieu, même si elle est posée sur la grille comme un mur.

**`sons.py`** génère les bruitages par synthèse, sur le même principe de graine et de manifeste — le procédé est décrit plus bas, à la section du son.

### Les manifestes

Chaque lot produit un manifeste JSON, et c'est lui qui fait contrat entre les images et le moteur.

Chaque manifeste porte un **en-tête** : `version_format`, et pour le décor la taille de tuile. Sans lui, aucune migration n'est possible le jour où un champ change de sens.

Côté **décor** : taille, ancrage, élévation, catégorie, thème, et trois champs qui commandent le moteur — `bloquant`, dont le chargeur tire la grille de passabilité, `emprise` en tuiles, sans laquelle une gondole de deux tuiles n'en bloquerait qu'une, et `transparence_si_derriere` pour ce qui dépasse 24 pixels. Aucun des deux ne se devine : un trottoir et un quai dépassent du sol et se marchent, une flaque est plate et se traverse, alors qu'un muret de même hauteur qu'un trottoir arrête tout.

Côté **personnages** : le rendu — cycles, cadences, bouclage, directions, point d'appui, gabarit, variantes — **et les valeurs de jeu**, dans le même fichier. Vitesse rapportée à celle du joueur, résistance en touches, points, dégâts de contact par seconde, rayon de collision, et ce qui est propre à un profil : tangentiel du flanqueur, portée de la Buse, dégâts de charge du Molosse, rayon d'explosion de la Baudruche.

Les mettre ailleurs aurait dupliqué la liste des profils à deux endroits. Un nouveau profil reste une ligne de table.

Côté **sons** : durée, gain, bouclage, et une **catégorie de mixage** — le joueur doit pouvoir baisser les effets sans toucher à la musique, et l'interface doit rester audible quand tout le reste est baissé.

Côté **objets** : emprise, ce qui bloque, ce qui détruit, ce qui est projeté, ce qui est entendu, et les valeurs de jeu — expérience d'une gemme, soin d'une fiole, dégâts et portée d'un projectile, charges d'une arme lourde. Un bloc `destruction` porte le mode — `contact` pour la caisse, où le délai est la mécanique elle-même, `tir` pour les obstacles —, le nombre de touches, le nom de la ruine, la matière des éclats, les cycles d'appui et de rupture, et les clés de sons. Le moteur ne code donc rien en dur : un futur obstacle se déclare dans une table.

Le contrôle vérifie la cohérence de ces renvois — une ruine qui n'existe pas, des éclats sans particules générées, un destructible au tir sans nombre de touches, une ruine qui bloque encore, un son introuvable. Ce sont des défauts qui ne se manifestent qu'au moment de la destruction, c'est-à-dire le plus tard possible et souvent chez un joueur.

Sans ces fichiers, une bande de 320 pixels est indéchiffrable — 5 images de 64 ou 4 de 80 ? Avec eux, le code de rendu ne connaît que des profils et des cycles, jamais des noms de fichiers ni des nombres codés en dur. Remplacer plus tard le chien du pack par un sprinteur dessiné à la main avec 6 images se fait en changeant une ligne du manifeste.

### Le son

**Les bruitages sont générés**, par `outils/sons.py` : une enveloppe appliquée à un oscillateur, plus un peu de bruit. C'est le procédé de sfxr, et il couvre exactement le registre d'un survivor — tirs, impacts, ramassages, explosions. Seule la bibliothèque standard est employée, et chaque son a sa graine, donc les fichiers sont reproductibles au bit près comme les images.

**Le tir de base est le son le plus contraint du jeu.** En tir automatique il part plusieurs fois par seconde pendant quinze minutes : au même niveau que les autres, il recouvrirait la musique et saturerait l'oreille. Il est donc très court, aigu et mat — un son bref et haut se superpose à une nappe sans occuper sa place — et son gain est le plus bas du catalogue, environ 20 % de l'échelle contre 70 % pour une explosion.

C'est une règle générale et pas un réglage : **les sons rares ont le droit d'être forts, les sons répétés doivent rester sous la nappe**. Le catalogue porte donc un gain par son, qui fixe le rapport entre eux ; le volume absolu et les réglages par catégorie restent au moteur.

Un détail vaut d'être noté parce qu'il porte le moment de plaisir maximal du genre : le ramassage de gemme existe en **huit degrés d'une gamme**. Le moteur joue le degré suivant à chaque gemme d'une même volée et repart du premier après un silence. Une hauteur unique répétée deux cents fois deviendrait vite pénible.

**La musique reste tierce.** Elle demande un vrai métier, et une nappe d'ambiance par lieu suffit à ce jeu. Une piste sous licence entre dans `CREDITS.md` avec son auteur et sa licence, dans le commit qui l'introduit — CC0 ne demande rien, CC-BY impose une ligne de crédit, CC-BY-NC est exclu pour un jeu vendu.

### Les placeholders

Jusqu'au jalon 3, des capsules colorées avec ombre au sol, une couleur par archétype, générées par code. On apprend davantage sur la boucle avec des formes lisibles qu'avec de jolis sprites obtenus trois semaines plus tard.

### Les crédits

Un fichier `CREDITS.md` à la racine du dépôt, tenu **dès la première ressource tierce** et jamais à la fin — retrouver la provenance d'une tuile six mois plus tard est impossible. Décor et personnages étant générés, il ne porte pour l'instant que la mention de la référence ayant servi à la mise au point.

Une ligne par paquet : nom du paquet, auteur, URL, licence, date de récupération. La CC0 n'exige pas l'attribution mais elle est appréciée ; la CC-BY l'exige. Conserver aussi le fichier de licence original dans le dossier de l'asset.

Cette section reste écrite parce que les sons, eux, viendront de sources tierces.

### Le contenu utilisateur

Décision arrêtée : un paquet communautaire **ne peut embarquer ni image ni son**. Un auteur compose exclusivement avec le vocabulaire fourni par le binaire — pièces, tuiles, objets, profils d'ennemis, événements.

Trois problèmes disparaissent d'un coup : la vérification des droits sur ce qui transite par le jeu, la modération de ce qui est distribué, et le poids des paquets. C'est aussi ce qui rend la distribution triviale, puisqu'un niveau se réduit à une chaîne de caractères.

Le mode tuiles différé ne remet pas ça en cause quand il arrivera : peindre une pièce se fera avec les tuiles du jeu, sans import possible.

---

## 15. Le langage et le budget technique

### Go avec Ebitengine

Retenu, principalement parce que c'est déjà la pile de [[jeu-fugitif]] : même chaîne de compilation, mêmes runners pour macOS et Linux, WebAssembly déjà prévu. Deux jeux qui partagent leur outillage font un chantier de moins.

Sur le fond, le travail à écrire — BFS sur grille, pools, boucle sur quelques centaines d'entités, blitting, tri par compartiments — tombe bien pour Go, et ne demande pas de moteur.

Deux manques à connaître d'avance. **Ebitengine ne fournit aucune interface** : pour le jeu ce n'est pas grave (l'écran de montée de niveau, c'est trois cartes dessinées), pour l'éditeur c'est un vrai sujet — palette, jauges, dialogues. `ebitenui` fait le travail, à évaluer avant de s'engager. Et **le rendu isométrique est entièrement à écrire** : tri en profondeur, conversion écran/tuile, caméra en pixels entiers. Une journée, écrite une fois pour les deux jeux.

L'alternative sérieuse serait Godot, qui donnerait tilemap isométrique et interface gratuitement — mais l'éditeur voulu ici est intégré au jeu, avec test en direct, donc l'avantage fond et il faudrait apprendre un moteur entier au lieu d'un langage.

### La structure des entités

À trancher au premier jalon, pas après.

```go
type Enemy struct {
    X, Y       float32
    VX, VY     float32
    HP         int16
    Profile    uint8   // index dans []EnemyProfile, jamais un pointeur
    Cycle      uint8
    Frame      uint8
    Generation uint16
}

type Pool struct {
    enemies []Enemy // capacité fixe, jamais réallouée
    active  int
}
```

Un `[]Enemy` de structures pleines plutôt qu'un `[]*Enemy`. À ce volume, l'argument du cache est secondaire — 300 structures tiennent en L2 et le temps d'image est dominé par les appels de dessin. Le vrai motif est **la pression sur le ramasse-miettes** : 300 objets alloués à chaque vague produisent les micro-saccades qui se voient.

Trois principes, plus importants que le choix structures/pointeurs :

- **Préallouer et ne jamais réallouer.** `make([]Enemy, 0, 512)` au démarrage ; suppression par échange avec le dernier actif puis décrément. Pas d'`append`, pas de trou.
- **Ne pas faire sortir de pointeur du bassin.** Après un échange, un `*Enemy` conservé ailleurs désigne une autre entité. Pour référencer un ennemi qui vit plusieurs images — une cible verrouillée —, utiliser index + génération : au recyclage la génération est incrémentée, et le détenteur d'une référence périmée le voit.
- **Séparer le chaud du froid.** Vitesse, PV max, poids de séparation, tangentiel sont partagés par tous les ennemis d'un type : ils vivent dans un `[]EnemyProfile` et l'entité n'en garde que l'index. C'est ce qui garde la structure petite, seul levier qui compte réellement au parcours.

En itérant, prendre l'adresse plutôt que la copie :

```go
for i := range p.enemies[:p.active] {
    e := &p.enemies[i]
    ...
}
```

Deux pièges. L'échange à la suppression **casse l'ordre**, donc le tri en profondeur travaille sur une slice d'indices réutilisée avec `indices = indices[:0]`, jamais sur le bassin lui-même. Et le passage en tableaux séparés par champ n'est à envisager que si le profileur le réclame : pénible à écrire, sans gain mesurable à ce volume.

Le même modèle sert pour les projectiles, les gemmes, les caisses et les particules.

### La persistance

Ce qui survit à une run est peu de chose : les déblocages, les meilleurs scores par lieu et par graine, les réglages, et le compteur de parties. Rien de tout cela n'a besoin d'un moteur de base de données, mais deux raisons peuvent en justifier un — l'historique des scores par graine, qui grossit, et l'habitude déjà prise sur [[jeu-fugitif]].

**Un point à vérifier avant de s'engager, et il est bloquant** : la cible `js/wasm` est compilée par l'intégration continue, précisément pour qu'aucune dépendance n'introduise de cgo sans qu'on le voie. Un pilote SQLite qui ne se construit pas pour `js/wasm` casserait ce contrôle, et l'on découvrirait le problème sur le premier commit qui l'ajoute. À tester dans un bac à sable avant de l'inscrire dans `go.mod`.

Le repli, s'il ne passe pas : un fichier unique en JSON, écrit par le point de sérialisation commun. Quelques kilo-octets, aucune dépendance, lisible à la main quand un joueur signale un défaut — ce qui n'est pas un mince avantage.

Dans les deux cas, la sauvegarde est **écrite à la fin d'une run et à la sortie**, jamais en cours de partie : une écriture disque au milieu d'une vague est exactement le genre de saccade qu'on passe le reste du document à éviter. Quitter en pleine partie perd la run en cours, et c'est assumé — une run dure quinze minutes, et la reprendre à froid n'aurait pas de sens dans un jeu qui tient sur l'élan.

### Les autres postes

Le rafraîchissement local du flow field à la destruction d'une caisse, le tri en profondeur par compartiments, et le recyclage de la traîne d'ennemis sur les lieux étirés.

---

## 16. Les jalons

Fait :

- Le format de pièces et la géométrie 2:1, validés au pixel : la tuile carrèle sans couture ni joint double.
- Le générateur de décor, une soixantaine de formes avec manifeste, extensible en quelques lignes par forme.
- Le générateur de personnages : six gabarits, neuf profils, huit orientations, et les valeurs de jeu dans le même manifeste que le rendu.
- Les objets et les bruitages, générés et contrôlés au même titre : caisse et ses cycles, gemme, fiole, projectiles, armes lourdes, éclats par matière.
- Le test de projection : personnages et tuiles s'alignent.

Reste, dans cet ordre :

1. **Le noyau nerveux.** Un rectangle qui bouge, 300 rectangles qui le poursuivent via le flow field, un tir automatique. Si ce n'est pas déjà satisfaisant, aucun sprite ne le sauvera. C'est ici que se fige la structure des entités (chapitre 13) : la reprendre plus tard touche tout le code de jeu.
2. **La boucle mort → relance**, avec une seule arme et un seul type d'ennemi. Se surprendre à enchaîner cinq runs sur ce prototype est la condition pour continuer.
3. **La salle complète** : vagues, caisses, objectif, porte, build. Jalon décisif — si cette salle donne envie d'être refaite, tout le reste n'est que du contenu.
4. Les six profils d'ennemis, désormais disponibles en images, et la courbe de pression.
5. Le système d'armes, de niveaux et de synergies.
6. L'enchaînement de salles et la signalétique.
7. L'assemblage des pièces au chargement, les lieux officiels bâtis en pièces.
8. L'éditeur : pose de pièces, connecteurs, jauges, test en direct.
9. Le mode tuiles, l'édition de campagne, le partage.

Note de prudence : survivor, roguelite, exploration, ressources et éditeur avec campagne, empilés, font un projet de plusieurs années à une personne. L'ordre ci-dessus est conçu pour que la question « est-ce que c'est amusant ? » soit tranchée à l'étape 3, en quelques semaines, et non après trois ans.

## 17. Ce qui reste à trancher

- **La rotation des pièces** : quatre variantes de mur dans le tileset, ou aucune rotation et plus de pièces à dessiner. À décider avant de dessiner la moindre tuile.
- **La taille de la maille des pièces** : 16×16 tuiles est la base proposée, elle conditionne tout le travail d'édition.
- **La palette définitive** : 32 couleurs extraites du pack de personnages, à reporter dans `MATIERES` pour que décor et sprites deviennent cohérents d'un coup.
- **La semi-transparence** des objets hauts quand le joueur passe derrière : le manifeste porte déjà `transparence_si_derriere` sur les vingt-sept formes concernées, reste à décider comment le rendu l'applique — opacité fixe, découpe, ou seulement autour du personnage.
- **L'angle du cône de visée**, entre 90 et 120 degrés : à régler en jouant, c'est lui qui décide si le kiting est agréable.
- **Le plafond de dégâts par seconde** et le rapport entre contact ordinaire et charge du Molosse : deux chiffres qui décident si l'encerclement est tendu ou injuste.
- **La vitesse du joueur rapportée à celle des profils** : à 60 % de sa vitesse un Badaud ne rattrape jamais, à 90 % la fuite ne suffit plus. Tout le kiting tient dans ce rapport, et il se règle en jouant.
- **La portée de ramassage des gemmes** : c'est elle qui rend l'aimant nécessaire ou décoratif.
- **Le pilote de persistance** : SQLite si et seulement si il se compile pour `js/wasm`, sinon un fichier JSON.
- **La courbe de résistance** au fil de la run, et le **temps de référence** de chaque lieu : c'est lui qui fixe le poids réel du bonus de temps face aux points d'ennemis.
- **La bibliothèque d'interface pour l'éditeur** : `ebitenui` ou tout dessiner à la main. À évaluer avant le chantier de l'éditeur, pas pendant.

Tranché en cours de route : aucun asset importable dans le contenu utilisateur, et mode tuiles différé après la première version de l'éditeur.

Tranché par le générateur : la catégorie n'est pas saisie, elle est **dérivée de l'élévation** — `PLAFOND_OBSTACLE_BAS` vaut 24, au-delà la forme est `haut` et porte `transparence_si_derriere`. Un obstacle bas trop haut est donc impossible à produire, plutôt que refusé après coup. Le distributeur de billets, l'abribus et les véhicules sont `haut` au même titre que les murs : il n'y a rien à baisser.
