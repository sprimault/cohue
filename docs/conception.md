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

Cible : **une touche, moins d'une seconde, même configuration**. L'écran de mort affiche le résumé et « Entrée pour relancer » en évidence. Le reste est secondaire dans la hiérarchie visuelle. Ça paraît trivial, c'est la variable la plus lourde de tout le système.

Corollaire propre à la progression par salles : mourir au troisième lieu renvoie au premier, ce qui est bien plus punitif qu'une arène unique. Les trois premières minutes doivent donc être **rapides à repasser** — courbe basse, pas de temps mort, aucune cinématique. Sans ça, la relance devient une corvée et tout le bénéfice est perdu.

### Finir en plein élan

Une run doit s'arrêter alors que quelque chose est en cours : 200 XP avant une évolution, deux armes sur trois pour une fusion, une synergie composée mais jamais vue à l'œuvre. C'est la phrase intérieure — « la prochaine fois je prends la même chose mais je monte l'orbite d'abord » — qui déclenche la relance, jamais le score.

Concrètement : la mort typique d'un joueur moyen doit tomber vers 8-11 minutes, alors que les évolutions les plus intéressantes se débloquent vers 12-14. La majorité des runs meurent avec un plan inachevé — mais après avoir traversé la phase de toute-puissance, qui commence à la minute 7. Le joueur meurt en pleine gloire avec un plan en cours, pas avant d'avoir rien vu.

### Le tempo des montées de niveau

C'est le métronome du plaisir. Front-load agressif : premier niveau à 12 secondes, puis toutes les 15-20 secondes **pendant quatre minutes**, puis l'écart s'allonge. Règle dure : **jamais plus de 45 secondes sans un choix à faire**.

Quatre minutes et non deux, parce que c'est ce qui amène la toute-puissance à la minute 7 sans toucher au début. Le début est déjà à sa limite : en deçà de 15 secondes, les choix s'enchaînent trop vite pour être des choix. C'est donc l'allongement qui démarre plus tard, et cela vaut environ sept montées de niveau de plus au moment où la bascule doit se produire.

Le choix compte plus que la récompense. Trois cartes, dont deux tentantes. Si le joueur prend systématiquement la même, l'équilibrage est cassé — la bonne carte est celle qui fait hésiter.

**C'est le seuil qui monte, jamais la valeur d'une gemme.** Une gemme vaut la même chose du début à la fin, quel que soit ce qui l'a laissée tomber ; c'est le seuil du niveau suivant qui croît, et c'est lui qu'on règle pour espacer les choix. L'inverse mélangerait deux questions que ce chapitre traite séparément — ce que vaut un kill, et à quel rythme les choix arrivent —, et un Badaud vaudrait davantage en fin de run pour une raison qui ne le concerne pas. Un profil qui doit rapporter plus laisse tomber **plusieurs** gemmes : la quantité au sol dit alors ce qu'on va gagner, ce dont l'aimant a besoin pour qu'on estime sa récolte avant de déclencher.

**Le temps compte aussi**, et c'est ce qui rend la règle des quarante-cinq secondes vraie par construction. Un joueur qui a tenu quarante-cinq secondes a fait quelque chose — il a kité, esquivé, survécu — et le jeu le reconnaît par une montée. Ce n'est pas un rattrapage mais une seconde source de progression, trois fois plus lente que le tempo nominal, si bien que fuir sans ramasser reste le mauvais calcul.

Elle est nécessaire parce que les deux autres décisions tirent en sens contraire : à valeur fixe chaque niveau demande davantage de gemmes, et les gemmes s'effacent. Sans plancher, le tempo dépendrait du taux de ramassage alors qu'il est une règle dure — avec lui, une partie où l'on ne ramasse rien produit une montée toutes les quarante-cinq secondes, et c'est un test qui l'exige.

**Il ne remet aucun compteur à zéro.** Les gemmes déjà ramassées comptent pour le niveau suivant : un joueur puni par sa lenteur perdrait sinon ce qu'il avait collecté, et le plancher cesserait d'être purement additif.

### La montée de niveau ne casse pas le rythme

Le choix des trois cartes met le jeu en pause, mais **brièvement et sans cérémonie** : un fondu court, les cartes, le choix, un fondu, et on repart. Pas d'écran plein, pas d'animation d'entrée, pas de son long.

C'est ce qui compte en fin de run, quand les niveaux s'enchaînent toutes les vingt secondes : une transition d'une seconde répétée quinze fois hache la partie, alors que deux fois cent cinquante millisecondes se traversent sans qu'on les remarque.

La pause est réelle — la horde se fige — parce que choisir sous pression n'est pas un choix, c'est une loterie, et le document a posé que le choix compte plus que la récompense.

### Le pic de puissance

La courbe de sensation doit croiser celle de la difficulté.

- Minutes 0-2 : fragile, chaque ennemi compte.
- Minutes 3-6 : montée, on encaisse et on rend.
- Minutes 7-11 : le joueur traverse les hordes sans regarder, l'écran est blanc de dégâts.
- Dernière minute : tout est repris.

Cette phase de toute-puissance est indispensable. C'est le souvenir que le joueur emporte, et c'est ce qu'il essaie de retrouver en relançant.

**Elle tombe avant la mort typique, et c'est ce qui la rend possible.** Placée de la minute 10 à la 14, elle serait réservée à ceux qui survivent au-delà de la mort moyenne — c'est-à-dire absente de la partie du joueur qu'on cherche justement à faire relancer, et absente des premières heures, quand il décide s'il reste. Un jeu qui réserve son meilleur moment aux joueurs déjà convaincus perd les autres.

D'où un critère qui rend l'intention vérifiable, et qui est le vrai contenu du jalon éliminatoire : **si la bascule n'est pas ressentie avant la minute 9, la courbe est trop lente.**

### La conduite du jalon éliminatoire

Ce critère ne vaut que si l'on mesure ce qu'il prétend mesurer. Au moment de juger, tout réglage capable de fausser la lecture sans être ce que le critère interroge se sort du chemin : les gemmes ne s'effacent pas, la portée de ramassage est large. Sinon un échec est ambigu, et l'on conclut « la courbe est trop lente » là où le joueur perdait simplement ses gemmes.

**Neutraliser n'est pas adoucir**, et c'est la confusion qui rendrait la règle nuisible. La courbe de pression est exactement ce que le critère mesure : l'adoucir ferait passer le jalon trivialement. D'où une question qui classe un réglage au lieu de prescrire une direction — **ce réglage est-il ce que le critère chiffré mesure ?** Si oui, il se règle à l'intention et rien d'autre ; sinon, on le neutralise.

L'erreur à craindre n'est pas symétrique, et c'est ce qui justifie la précaution. Un échec ambigu se rejoue : on cherche, on trouve, on recommence. Un succès ambigu se clôt — c'est le seul jalon qui puisse arrêter le projet, donc un oui ferme la question définitivement, et l'on découvrirait trois étapes plus loin qu'on a validé une sensation fabriquée pour l'occasion.

Cette conduite ne vaut que pour ce jalon-là. Une règle générale d'évaluation n'aurait qu'un seul cas d'application, ce qui est le signe qu'elle n'est pas générale.

### Le feedback par kill

Quinze minutes ne passent que si chaque seconde est satisfaisante.

- Temps d'arrêt de 2 à 3 frames sur les gros impacts.
- Chiffres de dégâts qui jaillissent, couleur distincte pour les critiques.
- Son de ramassage dont la hauteur monte à chaque gemme d'une même volée, et retombe après un silence.
- Tremblement d'écran très court, très faible, cumulatif quand ça part en masse.
- Cadavres au sol pendant quelques secondes, effacés progressivement, pour que la salle porte la trace du carnage.

Le moment de plaisir maximal du genre n'est pas le kill, c'est **l'aimant** : deux cents gemmes qui convergent d'un coup avec une montée sonore.

### L'aimant

Objet dédié, apparition régulière, et **une charge que le joueur garde** : il se ramasse, il ne se déclenche pas au contact. C'est ce qui en fait une décision plutôt qu'un cadeau.

C'est aussi ce qui permet de garder la portée de ramassage courte. Sans aimant, elle devrait être généreuse — sinon chaque gemme laissée derrière est une perte sèche que le joueur voit s'accumuler sans recours —, et une portée généreuse retire au ramassage tout caractère de décision.

**Il a son emplacement propre, jamais partagé avec les fioles.** Mis en concurrence avec le soin, il ne serait jamais gardé : la vie est la seule ressource véritablement rare, le joueur prudent choisit toujours elle, et le pic n'aurait jamais lieu. Un objet dont la mécanique est détruite par la mise en concurrence ne partage pas ses emplacements.

**Sa contre-force est que les gemmes s'effacent.** Sans elle, la valeur du déclenchement croît strictement avec l'attente : attendre est toujours rationnel, et le joueur meurt avec sa charge. Un objet dont la valeur ne fait que monter n'est pas une décision, c'est un compte à rebours qu'on ne lance jamais.

L'effacement travaille au-delà de l'aimant. Ramasser oblige à revenir là où l'on vient de tuer, c'est-à-dire là où la horde converge : le trajet de collecte va contre le trajet de fuite, et le kiting en cercle dans un coin cesse d'être gratuit sans qu'on ait rien interdit.

**La portée de ramassage et la durée de vie d'une gemme forment un couple** et ne se règlent pas séparément, parce que chacune punit le non-ramassage. La tension vient de ce qu'on perd par choix, le vol de ce qu'on perd sans recours : la portée courte laisse le déplacement comme recours, l'effacement laisse l'aimant. Cumulés, ils n'en laissent qu'un, et il est consommable — c'est là que la collecte bascule de la tension au vol.

**Une gemme s'éteint progressivement, elle ne clignote pas**, et pas seulement pour épargner un signal clignotant à un écran déjà chargé où il concurrencerait les télégraphes. L'extinction donne une information continue : l'âge d'une gemme se lit, donc la récolte d'un déclenchement s'estime avant d'appuyer. C'est ce qui fait du déclenchement une lecture de la salle plutôt qu'un réflexe, et c'est ce qu'on casserait en revenant au clignotement pour une raison de lisibilité.

**La montée sonore du déclenchement reste à écrire, et les huit degrés du ramassage ordinaire ne la fournissent pas.** Ils sont conçus pour une volée de quelques gemmes, où la hauteur monte puis retombe après un silence ; deux cents gemmes les parcourent vingt-cinq fois, ce qui donne une scie. Ce qui est décidé est qu'il s'agit d'un son unique, attaché au déclenchement et non à chaque gemme aspirée — le reste se décide en écoutant.

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

Quatre lieux de trois à quatre minutes, avec une build qui traverse les segments. Le total reste de l'ordre du quart d'heure — cinq lieux de quatre minutes n'y tiendraient pas.

Quatre segments donnent trois portes, donc trois fois le choix « rester ou partir » : assez pour installer la tension, pas assez pour la répéter jusqu'à la lassitude. Et raccourcir les lieux plutôt qu'en retirer un aurait rogné ailleurs — trois minutes laissent à peine le temps d'une montée, d'un pic et d'un creux.

### La sortie se gagne

Le joueur dispose d'un moyen d'échapper à la difficulté : courir vers la porte. Or l'XP est le moteur de tout — si on saute de lieu en lieu en dix secondes, on arrive au quatrième avec une arme au niveau 1 et le jeu se casse en deux.

La porte s'ouvre donc après un objectif : un temps de survie, un compteur de kills, trois points à réamorcer, une élite à abattre. Une fois l'objectif rempli, le joueur choisit **quand** partir — rester pour farmer ou partir avant que ça déborde est un des meilleurs choix du jeu, et le score du chapitre 6 lui donne une mesure.

**Le compteur de kills est celui qui est écrit**, et les trois autres attendent leur usage. Il est le seul qui ne récompense pas l'attente : un temps de survie se remplit en tournant en rond dans un coin, c'est-à-dire par le comportement même que ce chapitre existe pour décourager. Il éprouve du même coup la règle du figurant — ce qui n'est pas hostile n'entre dans aucun compte —, jusque-là écrite sans qu'aucun compte existe pour la contredire.

**L'objectif est écrit par le lieu, pas par le binaire.** C'est son auteur qui compose sa longueur, comme il compose sa courbe : un seuil en dur ferait de tous les lieux la même durée. Il se refuse en dessous de un, faute de quoi le champ oublié donnerait une porte ouverte au premier tick.

**La porte se touche, elle ne se traverse pas.** Elle reste l'obstacle que le décor déclare, ouverte comme fermée, et le joueur qui l'atteint sort du lieu. Deux raisons vont dans le même sens : une horde qui sortirait par où le joueur sort n'aurait pas de sens, et rendre la case franchissable demanderait de modifier la carte cuite — or celle-ci est partagée par toutes les runs d'une session, si bien qu'une porte gagnée rouvrirait la suivante avant son premier tick. L'ouverture est un état de partie, elle vit là où vivent les états de partie.

### Le temps mort à la porte

Entre deux lieux, une pause courte : choix d'amélioration, gestion des consommables, et éventuellement choix de destination. C'est le seul endroit où l'on ouvre un écran ; jamais pendant l'action.

### Le choix de branche

Ce qui est inconnu n'est pas la position de la porte, c'est **ce qu'il y a derrière**. Quand une campagne bifurque, deux portes proposent chacune un aperçu court — « quartier, dense », « parking, sombre, réserve de caisses » — et le joueur tranche. C'est ce qui produit la phrase « cette fois je pars sur la branche parking », et donc la run suivante.

---

## 4. Le combat : un champ, pas mille pathfindings

La bonne structure, c'est le **flow field**. Un seul BFS depuis le joueur sur la grille de tuiles, recalculé toutes les 5-6 frames. Chaque cellule stocke deux choses : la distance au joueur et le vecteur vers la cellule voisine la plus proche. Un ennemi ne calcule rien, il lit la cellule sous ses pieds.

Coût : un BFS sur 128×128 cellules, quelques dizaines de microsecondes, une fois pour tout le monde.

Ordre de grandeur des effectifs, calé sur le tampon interne de 960×540. La demi-étendue visible y vaut 15,9 tuiles par axe du monde — pas « une quinzaine en largeur », qui serait un compte orthogonal et donnerait un champ deux fois trop petit. Sur cette surface : **60 à 100 ennemis à l'écran** en régime normal de fin de segment, **150 en pic** sur une vingtaine de secondes, **250 à 300 entités vivantes** au total en comptant ce qui approche hors champ. Au-delà, ce n'est plus une horde mais un mur uni : les profils cessent d'être distinguables, et avec eux la lisibilité de l'échec. Le levier de pression n'est de toute façon pas le nombre, c'est la vitesse relative et la fermeture des angles — vingt ennemis qui coupent une sortie font plus peur que deux cents qui suivent en file. Le contournement d'obstacles est gratuit — piliers, rayonnages, tourniquets se retrouvent dans le champ sans une ligne de code d'évitement. C'est ce qui rend le décor urbain jouable là où les survivors classiques se contentent d'un terrain vide.

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

**La dérive se prend sur l'attirance seule, jamais sur la somme des forces.** Le pseudo-code ci-dessus le fait sans le dire, et c'est ce que la prochaine réécriture perdrait. Sur la somme, un flanqueur serré par ses voisines tournerait autour d'elles plutôt qu'autour du joueur : son comportement dépendrait de la densité locale, et il cesserait d'être un flanqueur au moment précis où il devrait l'être, c'est-à-dire en groupe.

**Le vecteur nul a sa direction, et elle vient de l'entité.** La somme peut s'annuler — une créature pile sur le joueur, deux créatures exactement superposées dans le gradient de densité, ce qu'un anneau d'apparition finit toujours par produire. Normaliser n'a alors pas de réponse, et il en faut une : une direction fixe serait la pire, puisque deux entités superposées recevraient la même correction et le resteraient indéfiniment.

La direction se dérive donc de **l'index de l'entité dans son bassin**. Deux entités vivantes ont toujours des index distincts, donc deux directions différentes, donc elles se séparent au tick suivant. C'est déterministe, identique sur toutes les machines, et cela ne consomme aucun tirage aléatoire.

Elle se prend dans une **table de huit directions**, à un pas premier avec leur nombre : `(index × 3) mod 8` porte l'écart entre deux index consécutifs à 135 degrés, là où un pas de 1 n'en donnerait que 45 — deux créatures apparues côte à côte ont des index voisins, et se sépareraient lentement.

Une table, et non un angle calculé, pour une raison qui dépasse la vitesse : **`sin` et `cos` ne sont pas correctement arrondis par l'IEEE-754**, contrairement à `sqrt`, qui l'est et que la simulation admet pour cette raison. Leur dernier bit peut différer d'une implémentation à l'autre, donc d'une architecture à l'autre : faire entrer de la trigonométrie ici rouvrirait la porte que la virgule fixe a fermée, et sur le cas le plus difficile à diagnostiquer — une divergence qui n'apparaît que lorsque deux entités se superposent.

Les huit valeurs sont donc tabulées : `0`, `±65536` pour les axes, `±46341` pour les diagonales — ce dernier étant `65536 × √2 ⁄ 2` arrondi. **Une diagonale n'est donc pas exactement unitaire** : sa norme vaut 65536,07 pour une tuile de 65536, soit un millionième de trop. Autant dire exacte, et le chiffre est là pour qui réemploiera la table ailleurs — il répond d'avance à la question plutôt que de la laisser chercher pourquoi ses diagonales seraient lentes.

L'index n'est pas stable dans le temps — la suppression par échange le change dès qu'une entité meurt devant — et ce n'est pas nécessaire : ce qu'il faut est l'unicité **à un instant donné**, pas la permanence. Le cas dure un tick, le temps que les deux se séparent. Ajouter un identifiant d'apparition à l'entité pour le stabiliser reviendrait à payer un champ pour une propriété dont le mécanisme n'a pas besoin.

Deux garanties de niveaux différents se répondent ici, et il vaut mieux ne pas les confondre : la virgule fixe garantit qu'aucune opération ne produit de valeur aberrante, la normalisation garantit qu'aucune direction n'est arbitraire. C'est la seconde qui répond à la question du jeu — un résultat saturé donnerait une direction juste et une longueur absurde. Le cas se traite donc **avant** de diviser, pas après.

### L'apparition

Les ennemis apparaissent **hors du champ de vision**, sur un anneau autour du joueur, à trois tuiles au-delà du bord de l'écran. Jamais dans le champ : voir une créature se matérialiser détruit la crédibilité de la horde et rend une mort incompréhensible.

Sur une carte fermée, aucune position de l'anneau n'est parfois valide — le joueur est dans un cul-de-sac, ou adossé à un mur de niveau. Dans ce cas **l'apparition est abandonnée**, pas déplacée : plutôt aucune créature qu'une créature surgie d'un mur.

Le budget de pression correspondant n'est pas perdu pour autant : il est reporté au tick suivant, où l'anneau aura peut-être une position libre. Sans ce report, un couloir étroit deviendrait un abri où la pression tombe à zéro, ce qui est exactement ce que la conception cherche à éviter.

**Mais le report est borné à quelques secondes.** Sans borne, vingt secondes passées dans un cul-de-sac s'accumulent et se libèrent d'un coup à la sortie — c'est-à-dire le mur d'ennemis que la règle « jamais dans le champ de vision » cherche à interdire. Ce qui déborde la borne est perdu, et c'est voulu : se terrer doit coûter du temps, pas produire une punition différée qu'on ne relie plus à sa cause.

### Rien ne traverse un mur

La poussée de séparation n'est qu'une force parmi d'autres : elle ne décide pas de la position finale. Le déplacement calculé est **projeté sur la grille de passabilité** avant d'être appliqué — s'il mène dans un obstacle, sa composante bloquée est annulée et l'entité glisse le long du mur au lieu de s'y enfoncer. Si les deux composantes sont bloquées, elle ne bouge pas.

Vingt Badauds qui poussent un Vigile contre une cloison ne le font donc pas passer au travers : ils s'entassent derrière lui. C'est le comportement voulu — un couloir bouché par un bloqueur doit rester bouché, c'est tout son intérêt.

Corollaire à assumer : **les entités se chevauchent** quand la place manque. Il n'y a pas de collision dure entre ennemis, seulement la répulsion douce du gradient de densité. Résoudre les chevauchements par un décalage géométrique produirait des tremblements en chaîne dans une foule dense, et pousserait mécaniquement les créatures du bord dans les murs — exactement ce qu'on vient d'interdire.

**Le joueur traverse la horde, sauf le Vigile.** C'est la seule exception, et elle lui donne ce que le tableau des rôles lui promet : un corps qui bouche un couloir, et pas seulement une résistance qui met du temps à tomber. S'il ne traversait rien, une foule dense le figerait et la mort deviendrait illisible ; s'il traversait tout, l'encerclement ne serait qu'un mur de dégâts et le bloqueur n'aurait plus de rôle.

**La solidité joue dans les deux sens** : le Vigile ne traverse pas le joueur plus que le joueur ne le traverse. Ce n'est pas une symétrie décorative, et l'ordre du raisonnement compte parce que la conclusion inverse paraît plus naturelle.

Un obstacle immobile n'appelle qu'une règle : on ne peut y entrer que par accident, donc en sortir doit toujours être permis. Appliquée telle quelle à un poursuivant, elle se retourne — c'est lui qui crée le recouvrement en avançant, et il annule alors sa propre solidité. **Un corps qui se rend traversable en avançant n'est plus un corps**, et le blocage cesserait à l'instant précis de la rencontre, c'est-à-dire au seul moment où il sert.

La réciprocité règle les deux d'un coup : le Vigile ne pouvant pas entrer dans le joueur, le recouvrement n'a pas lieu et la question de la sortie ne se pose plus. Qui voudra corriger le cas du joueur emmuré en rétablissant la règle asymétrique retirera donc le blocage lui-même, sans le voir.

Le blocage ne peut pas devenir un piège, parce que le corps solide ne l'est que vivant : un joueur pris entre un Vigile et un mur tire nécessairement dessus — c'est le plus proche, et la visée est omnidirectionnelle. Douze touches, c'est long et c'est fini.

**L'exception tient, et le sursis n'est levé qu'à moitié.** Ce qui est éprouvé : le blocage ne piège pas et ne se lit pas comme une injustice — le joueur s'arrête, comprend qu'il doit contourner, et repart. Ce qui ne l'est pas : le cas décrit juste au-dessus, celui d'un joueur **pris entre un Vigile et un mur**, sans direction de repli. C'est lui qui déciderait, et il ne s'est pas présenté.

**L'arrêt se comprend par le mouvement empêché avant de se comprendre par la vue**, ce qui est une nuance et non un détail : le joueur sait qu'il est bloqué parce qu'il n'avance plus, pas parce qu'il a reconnu la créature devant lui. Tant qu'une issue existe, cela suffit — c'est acculé que la distinction compterait, puisqu'il faudrait alors savoir qu'on tire sur ce qui retient plutôt que chercher un passage qui n'existe pas.

### Le recyclage de la traîne

Dès lors que le joueur progresse vers une sortie, les ennemis restés derrière ne servent plus à rien. Au-delà d'une distance seuil, l'entité est retirée du pool et réapparaît devant. Sans ça, la traîne grossit et le frame time avec.

**Réapparaît par l'anneau**, et non « devant » : c'est la même règle que pour une apparition neuve — hors du champ de vision, sur une position passable, abandonnée si aucune ne convient. Un raccourci qui poserait l'entité directement devant le joueur produirait exactement ce que le paragraphe précédent interdit, une créature surgie de nulle part, avec en prime l'impression que le jeu triche pour rattraper celui qui court.

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
| soigneur | **Secouriste** | humanoïde clair | le nettoyage tout court |

Résistance et points de chaque profil : chapitre 6.

**Le Secouriste est le seul qui crée une priorité de cible.** Les six autres cassent le kiting par la position ; aucun ne rend un ennemi plus urgent qu'un autre. Lui annule le travail tant qu'il vit, et comme le tir vise le plus proche sans que le joueur puisse choisir, le seul moyen de l'abattre est **d'aller vers lui** — donc d'abandonner sa position de kiting pour entrer dans la horde.

**Il soigne une créature à la fois, la plus blessée à portée, et jamais lui-même.** Une par impulsion parce qu'il annule le travail sans rendre la horde invincible ; la plus blessée parce que c'est celle que le joueur est en train d'abattre, donc celle dont la guérison se lit ; jamais lui-même parce que ses trois touches sont la récompense qui paie le trajet jusqu'à lui, et qu'un soigneur qui se régénère la retirerait au moment de l'obtenir. Un mort n'est pas une cible : une résistance nulle *est* la mort, et le soin ne rouvre pas cet état.

**Deux retours le montrent, et le plus utile n'est pas celui qu'on attend.** Voir une créature récupérer explique pourquoi le travail est perdu ; voir *lequel* soigne dit où aller, et c'est la seule information qui change la conduite du joueur — la visée prenant le plus proche, il faut d'abord savoir qui chercher dans une horde qui se ressemble.

Cette mécanique n'existe que parce que la visée est omnidirectionnelle et automatique. Avec le cône avant qu'on a écarté, il aurait suffi de s'orienter pour le prioriser, et il n'aurait rien coûté. Le lien mérite d'être noté ici, parce qu'aucun fichier ne le montre : réintroduire un jour un moyen de viser — un mode, une arme dirigée, une option d'accessibilité — désactiverait le Secouriste sans que personne ne touche au Secouriste.

**Le Passant n'est pas dans cette table**, parce qu'il n'est pas un ennemi. Son profil porte le rôle `ambiance`, et tout en découle d'une seule ligne : **ce qui n'est pas hostile n'entre dans aucun compte** — ni le budget de pression, ni le plafond d'effectif, ni un objectif de porte fondé sur les kills. Sans cette dernière conséquence, un lieu peuplé de Passants ouvrirait sa porte tout seul, et son auteur n'aurait aucun moyen de comprendre pourquoi.

**La règle vaut par son absence d'exception, et deux s'y ajoutent.** Il n'est pas une cible : la visée prend le plus proche sans que le joueur choisisse, donc un figurant dans le bassin des ennemis détournerait chaque salve et emporterait avec lui la mécanique du Secouriste, qui repose entièrement sur cette visée. Et **il n'entre pas dans la grille de densité** : s'il poussait la horde, on apprendrait à se placer derrière une foule de civils pour la dévier — une tactique réelle, obtenue sans qu'aucune décision de conception ne l'ait voulue, et qu'il faudrait ensuite équilibrer ou retirer.

**Son errance tire donc dans le flux cosmétique**, le seul dont rien ne dépend. C'est ce qui rend enfin vérifiable ce que ce document promettait sans pouvoir le montrer : deux runs d'une même graine dont le cosmétique est consommé différemment rendent la même simulation. Tant qu'aucune entité n'y puisait, la garantie était annoncée et ne séparait rien.

**Il se pose par le lieu, pas par le spawner** : il n'a pas de coût de pression, donc rien à acheter dans une courbe qui dépense un budget. **Et il se place, il ne se sème pas** — chaque figurant porte sa case, comme une pièce porte la sienne. Un peuplement tiré au sort abandonnerait en silence les positions tombées dans un mur, si bien qu'un lieu demandant douze figurants en poserait neuf sans que son auteur l'apprenne ; écrites, elles se refusent au chargement. Elles ne sauteraient pas non plus d'un coin à l'autre entre deux relances, quand tout le reste de la salle ne bouge pas.

Il n'est pas davantage une cible. Le tir étant automatique, le rendre ciblable ferait tuer des civils au joueur **sans qu'il l'ait décidé** — le jeu orienterait la salve à sa place, et la charge du geste lui reviendrait sans le choix. Le moment où cela se produirait est le pire : salle nettoyée, quand plus rien ne presse et que ça se voit. La règle existe déjà pour le mobilier, et une seule phrase couvre les deux cas — la visée ne choisit que ce qui menace.

Il traverse la scène en va-et-vient — il avance tout droit et repart quand il bute, ce qui n'exige aucune trajectoire posée dans la pièce — et son seul effet est d'occuper l'espace où l'on voulait passer. Lui donner des dégâts de contact en referait un Badaud affaibli, c'est-à-dire un doublon ; le ranger parmi les ennemis obligerait chaque boucle écrite ensuite à demander « celui-là attaque-t-il ? ».

- **Le Badaud** : masse lente, il ne fait qu'exister en nombre. Il existe en plusieurs teintes de vêtement — une foule d'un seul bleu se lit comme un bloc uni, alors que six variantes cassent la répétition sans coûter une silhouette de plus. La variante est tirée à l'apparition depuis la graine de la run, donc elle ne casse pas le déterminisme.
- **Le Molosse** : télégraphe une charge (une demi-seconde d'anticipation, un son), puis fonce en ligne droite et ne corrige plus. Sa charge inflige davantage qu'un contact ordinaire — sans cela, charger ne serait qu'un déplacement rapide. Il punit l'immobilité, mais s'esquive latéralement. Le fait qu'il abandonne le flow field pendant la charge est ce qui le rend lisible. **Il n'apparaît jamais seul** : une meute de trois qui charge en décalé impose d'arrêter de reculer en ligne droite, ce qu'un chien isolé n'obtient pas. La taille de groupe est un champ du profil, pas une exception du spawner.

  Trois décisions complètent la charge, et elles se tiennent ensemble parce que chacune sans les deux autres retire à la mécanique ce qui la rend jouable.

  **La charge s'arrête sur ce qu'elle heurte, et le décor devient une réponse.** Ne corrigeant plus, elle ne contourne rien : un pilier interposé la reçoit, et la créature a dépensé son annonce pour rien. Rien ne vérifie la ligne de vue au déclenchement — elle part *malgré* l'obstacle, faute de quoi un Molosse embusqué attendrait sagement que la voie soit libre et le décor perdrait son seul usage défensif.

  **Toute fin de course ouvre un temps mort**, qu'elle ait touché, manqué ou heurté. Sans lui, une charge aboutie enchaîne sur la suivante et la créature n'a aucun moment vulnérable, ce qui contredit l'esquive latérale qu'on lui oppose. C'est ce temps mort qui paie le pas de côté.

  **Le choc est unique parce qu'il clôt la course.** Il vaut un montant et non un débit — le contact ordinaire se compte par seconde, lui tombe d'un coup — et il ne passe pas sous le plafond de dégâts : celui-ci existe pour rendre lisible un encerclement dont on ne distingue pas les parts, quand une charge a été annoncée puis manquée. Les plafonner ensemble ferait qu'une meute de trois infligerait ce qu'un seul inflige, et l'annonce n'annoncerait plus rien.
- **L'Arpenteur** : `Tangential` élevé, il coupe la trajectoire de fuite. C'est lui qui donne l'impression que les monstres réfléchissent.
- **La Buse** : seul profil qui blesse à distance, elle se stabilise et tire. Elle punit le camping dans un coin, force à bouger vers le danger. Sans balancement de marche, elle reste identifiable même immobile au milieu d'une horde.

  **Sa portée est aussi la distance où elle s'arrête**, et les deux ne se règlent pas séparément : approcher davantage la mènerait au contact, où un profil qui blesse de loin ne vaut plus rien. Elle vise où le joueur est et non où il sera — c'est ce qui rend le tir esquivable, donc ce qui punit l'immobilité sans punir le déplacement. Et rien ne vérifie que la voie est libre : le projectile part et meurt sur le pilier, comme la charge du Molosse s'y arrête. Le décor protège par le fait, jamais par une condition.
- **Le Vigile** : lent, encaissant, il bouche les goulots. Dans un couloir de supermarché, il transforme une route de fuite en piège.
- **La Baudruche** : explose en mourant. Sa silhouette disproportionnée dit « ne t'approche pas » avant même que le télégraphe ne s'allume. Elle punit le nettoyage à l'aveugle en mêlée.

  **La déflagration n'emporte que le joueur, jamais la horde autour.** Une explosion qui nettoie ses voisines serait plus imitative du réel et retournerait la mécanique : tuer sans regarder ce qu'on tue deviendrait la bonne façon d'éclaircir une foule, c'est-à-dire exactement le geste que ce profil existe pour décourager.

  **Ce qui sépare sa mort de sa détonation est une amorce**, et c'est elle qui rend l'esquive possible. Le télégraphe couvre ce délai et doit le couvrir entier : une annonce qui s'éteint avant l'explosion ment sur le seul point qui compte. La durée vit donc dans le profil, avec le rayon et les dégâts, et l'animation qui l'annonce s'étire dessus plutôt que l'inverse.

  **L'emprise se montre entière dès la première image**, et c'est l'intensité qui dit le temps restant. Une zone qui s'élargirait laisserait croire qu'on est hors de portée jusqu'à l'instant où l'on ne l'est plus.

Sept profils suffisent pour tout le jeu — les six décrits ci-dessus et le Secouriste. Ce sont des données, pas du code : une structure `EnemyProfile` avec nom, vitesse, résistance en touches, points, **coût de pression**, poids de séparation, tangentiel, portée, taille de groupe, comportement spécial. Le reste est du mixage de vagues.

Le **coût de pression** est ce que le spawner dépense pour acheter la créature ; les **points** sont ce que le joueur gagne en la tuant. Deux monnaies sans rapport, que le mot « points » a d'abord désignées toutes les deux à cinquante lignes d'écart. Le mot reste au joueur, qui le voit à l'écran.

Ces valeurs vivent dans `assets/personnages/manifeste.json`, avec le rendu : les mettre ailleurs dupliquerait la liste des profils à deux endroits.

### Le scénario de vagues : un budget, pas des compteurs

Point critique dès lors que des lieux sont créés par des tiers. Un scénario qui dit « à 7 min, 120 marcheurs » rendra le premier niveau amateur venu injouable ou vide. Un scénario exprime donc une **pression par seconde** ; le spawner achète des ennemis dans ce budget parmi les profils autorisés, et ça reste cohérent quel que soit le lieu.

```json
{
  "phases": [
    { "debut": "0:00", "pression": 8, "profils": ["marcheur"] },
    {
      "debut": "1:30",
      "pression": 25,
      "profils": ["marcheur", "flanqueur", "sprinteur"],
      "pic": { "a": "2:10", "multiplicateur": 3, "duree_s": 25 }
    }
  ]
}
```

Chaque profil a son coût de pression. Le spawner remplit, respecte la passabilité, et lâche hors du champ de vision. Bonus : la difficulté globale devient un seul curseur multiplicateur, ce qui donne les modes de difficulté gratuitement.

**Le spawner épargne au lieu d'acheter ce qu'il peut.** Il tire un profil parmi ceux que la phase autorise, puis met de côté jusqu'à pouvoir le payer. Sans cela, un budget qui se dépense au fil de l'eau se vide au prix le moins cher de la phase et n'atteint jamais les prix élevés : une phase qui ouvre sept profils n'en fait apparaître qu'un, et les six autres sont écrits sans jamais arriver. Le défaut est silencieux — chaque achat est légitime, c'est leur suite qui ne l'est pas.

**Le tirage est pondéré par l'inverse du prix**, ce qui donne à chaque profil la même part de budget et fait donc suivre au nombre l'inverse du prix : quatre Badauds pour un Vigile qui coûte quatre fois plus. C'est de là que vient une horde faite de masse avec des exceptions, sans qu'aucun réglage nouveau l'exprime — le coût de pression, qui disait déjà la rareté, la dit maintenant en nombres. Un tirage uniforme donnerait autant de Vigiles que de Badauds, donc quatre fois plus de budget aux gros : une horde chère et clairsemée, qui n'est pas celle que ce document décrit.

**Le profil convoité s'abandonne dès qu'il cesse d'être achetable** — son plafond de simultanéité atteint, ou le bassin qui se remplit. Sinon la horde entière attendrait derrière un profil que rien ne débloquera avant une mort. Ce qui a été mis de côté reste acquis : le budget ne se perd que par les plafonds, comme avant.

**Le scénario vit dans le `lieu.json`, sous la clé `vagues`**, parce que c'est le rythme que son auteur compose et partage avec sa salle. Une phase vaut jusqu'à ce que la suivante la remplace, ce qui évite à la dernière une fin à écrire ; la première date le début de la partie, faute de quoi les premières secondes n'auraient aucun palier. Un lieu peut n'en avoir aucune — une boutique, un passage —, et cela ne se refuse pas : ce qui garde le lieu livré d'y tomber par mégarde est un contrôle de conformité et non une règle de format.

**Le multiplicateur d'une pointe est consommé au chargement.** Le fichier écrit un multiplicateur, qui se lit bien ; ce que la phase porte ensuite est le budget qu'il donne. La multiplication faite une fois plutôt qu'à chaque image retire du même coup une valeur d'origine extérieure du chemin de l'arithmétique en virgule fixe.

**Un budget est un débit, les effectifs sont un stock, et il faut les deux.** Le premier dit combien de créatures arrivent par seconde ; le second, combien sont vivantes à un instant. Le stock vaut le débit multiplié par la durée de vie — laquelle dépend d'une arme qui aura été multipliée par dix au cours de la run. Un scénario qui ne piloterait que le débit laisserait donc les effectifs du début de ce chapitre à l'état de vœu : ils seraient espérés, pas tenus.

Le spawner porte donc aussi un **plafond d'effectif**, et cesse d'acheter quand il est atteint, quel que soit son budget. Le budget non dépensé pour cette raison est perdu et non reporté — sinon on retrouve le mur d'ennemis différé.

**Une phase ne peut autoriser que ce qu'elle sait payer.** Le budget s'accumule jusqu'à sa borne de report et pas au-delà : un profil dont l'apparition coûte davantage est écrit dans le fichier et n'arrive jamais. Pire depuis que le spawner épargne : convoité, il arrête toute la horde au lieu d'être seulement absent, puisque le budget qu'il attend est celui que la borne lui refuse. Le refus se fait donc au chargement, et il nomme les trois nombres dont la relation est fausse : le prix, le plafond, et la pression qui le produit. C'est la seule façon pour l'auteur de savoir lequel changer.

**Un profil peut aussi porter son propre plafond**, en nombre de vivants. Un coût règle une fréquence moyenne, il ne sait pas exprimer une simultanéité : le Secouriste ne vaut rien seul et double la difficulté au milieu de vingt Badauds, si bien que sa rareté ne peut pas se régler par son prix — trop bas il déséquilibre, trop haut il n'apparaît jamais et sa mécanique ne s'apprend pas. Le plafond compte les vivants et non les apparus, faute de quoi il deviendrait un quota par run et le profil disparaîtrait après le premier. Plafond atteint, le spawner achète autre chose ; s'il ne peut rien acheter, le budget est perdu.

**Un scénario ne peut que restreindre les propriétés d'un profil, jamais les étendre.** Il resserre un plafond, il ne le desserre pas ; et le jour où quelqu'un voudra donner plus de résistance ou une autre vitesse à un Badaud dans son lieu, la réponse est déjà écrite et c'est non. C'est ce qui garantit qu'une créature signifie la même chose partout — les touches annoncées, le coût de pression, ce que le joueur a appris de la première salle. La règle se vérifie au chargement : un plafond de scénario supérieur à celui du profil est refusé, avec le nom du profil et les deux valeurs.

---

## 5. Les dégâts subis

### Trois sources, une dominante

**Le contact** est le mode principal : le Badaud, le Molosse, l'Arpenteur et le Vigile n'ont pas d'autre moyen. Ils ne portent pas de coups, ils occupent l'espace — les dégâts ne sont donc pas des frappes isolées mais une **pression continue** quand on se laisse encercler.

**Le tir** n'appartient qu'à la Buse. C'est ce qui punit le camping dans un coin, et le seul cas où un projectile ennemi traverse l'écran : il porte une couleur qui n'existe nulle part ailleurs dans la palette.

**L'explosion** de la Baudruche, annoncée par son emprise marquée au sol pendant que l'amorce brûle.

### Le contact fait mal en continu, avec un plafond

Un dégât par seconde tant que le contact dure, pas un coup unique suivi d'invulnérabilité. C'est ce qui rend l'encerclement mortel et le déplacement obligatoire — la lecture du jeu vient de là, pas d'un compteur de coups.

Mais **le total encaissé par seconde est plafonné**, quel que soit le nombre d'ennemis collés. Sans ce plafond, trente Badauds tuent instantanément, et la mort devient illisible : le joueur n'a rien vu venir et ne peut rien en apprendre. Avec lui, être encerclé est très dangereux mais laisse une fenêtre pour se dégager, ce qui est exactement le moment de jeu recherché.

Corollaire sur le retour : à cette cadence, un son de dégât par tick serait insupportable. Le son se déclenche à l'entrée en contact et se réarme après un silence ; l'écran, lui, porte la jauge qui descend visiblement.

**L'écran ne rougit pas au contact, il rougit quand la vie est basse.** Ce document a d'abord demandé une teinte rouge brève à chaque fois qu'une créature touche, et c'est ce qui a été écrit puis retiré. Deux raisons, dont la seconde est la vraie :

- **le contact est intermittent**, et un retour qui le suit bat à chaque créature qui frôle. Ce défaut-là se corrige en allongeant la rémanence ;
- **le contact est déjà à l'image.** Une créature qui touche est collée au personnage, au centre du regard : le signal redisait ce que la scène montrait. La vie basse est au contraire la seule information critique qui vive hors du regard, en haut à gauche, là où l'on ne va pas en kitant.

Ce que l'écran signale est donc un **état** et non un événement : sous un seuil exprimé en points de vie, une vignette rouge cerne le bord. Elle ne peut pas battre, la vie ne remontant que par une fiole, et elle ne coûte rien à la lisibilité — c'est le second enseignement, obtenu en essayant : **un aplat plein écran teinte le sol vers la couleur de la horde**, si bien que les créatures s'en détachent moins au moment précis où il faut voir pour s'échapper. Un bord laisse le centre intact et s'adresse à la vision périphérique, qui est ce à quoi un signal permanent doit parler.

Le seuil vaut ce que rend une fiole. En dessous, en boire une ne gaspille rien : l'alerte annonce alors la décision qu'elle doit déclencher, au lieu d'être un chiffre choisi pour lui-même.

### La vie du joueur

Trois chiffres se tiennent et doivent être posés ensemble, sinon aucun n'a de sens : la vie totale, le plafond de dégâts par seconde, et ce que rend une fiole.

Point de départ proposé : **100 de vie, plafond à 20 par seconde, fiole à 30**. Ce n'est pas un équilibrage, c'est un rapport lisible — encerclé sans se dégager, le joueur tient cinq secondes ; une fiole rend un tiers de sa barre et lui rachète une seconde et demie d'encerclement.

La règle qui compte plus que les valeurs : **la vie ne se régénère pas seule**. Elle ne revient que par les fioles trouvées dans les caisses. Sans cela, attendre devient une stratégie et le joueur prudent finit par jouer un autre jeu que celui qu'on a écrit — or ici la seule ressource véritablement rare doit être la vie.

Une fiole ne dépasse jamais le maximum : le surplus est perdu, ce qui donne au joueur une raison de ne pas la boire tout de suite et fait de ses deux emplacements de consommables une petite décision de plus.

**Le plafond ne couvre que le contact continu.** La charge du Molosse et l'explosion de la Baudruche s'y ajoutent sans en relever. C'est le contact qui rend la mort illisible en masse — trente corps collés dont on ne distingue pas la contribution —, alors qu'une charge télégraphée ou un anneau qui s'élargit sont deux choses qu'on a vues venir et qu'on n'a pas esquivées. Les plafonner ensemble ferait qu'une meute de trois Molosses infligerait ce qu'un seul inflige, et le télégraphe n'annoncerait plus rien.

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
| Secouriste | 3 | 15 | tant qu'il vit, le reste ne meurt pas |
| Vigile | 12 | 60 | il bouche un couloir, on le contourne |

Ces valeurs sont un point de départ, pas un équilibrage : elles se règlent à partir du jalon 3, en jouant.

**La résistance monte au fil de la run**, par un multiplicateur adossé à la courbe de pression — sinon la fin de partie n'est qu'un tapis roulant, puisque l'arme a été multipliée par dix. C'est ce multiplicateur qu'on ajuste, jamais les touches de chaque profil, qui restent le rapport entre eux.

**Il vit sur la phase, à côté du budget qu'elle dépense**, et il est fractionnaire : la première valeur entière au-dessus de un est deux, si bien qu'un multiplicateur entier ferait commencer toute progression par un Badaud à six touches. Un virgule trois en donne quatre, qui est le pas naturel. Il ne descend pas sous un — une courbe durcit, et un profil d'une touche sous un demi rendrait une créature morte à l'apparition.

**Il s'applique à l'apparition et jamais après, et c'est ce qui garde l'unité.** Une créature qui durcirait pendant qu'on la frappe demanderait plus de coups qu'au coup précédent : « trois touches » cesserait d'être une unité pour redevenir un nombre, et le joueur ne pourrait plus compter. Le nombre et la dureté restent par ailleurs deux réglages distincts — monter la pression pour avoir plus d'ennemis ne doit pas les rendre plus durs par la même écriture.

### Les points et la tension qu'ils créent

Chaque ennemi rapporte, et le score d'un lieu additionne les points récoltés et un **bonus de temps** : ce qui reste d'un temps de référence quand la porte est franchie.

C'est le seul intérêt réel du score ici. Il met deux envies en opposition directe : rester pour farmer rapporte des points d'ennemis, partir vite rapporte du bonus de temps. Le joueur qui optimise doit trancher à chaque salle, et ce choix — rester ou partir — est déjà, dans le document, le meilleur moment du système de porte. Le score ne fait que lui donner une mesure.

Deux garde-fous. Le score ne doit **jamais contredire la lisibilité** : il s'affiche à la fin d'un lieu et sur l'écran de mort, pas en gros pendant l'action, où l'attention appartient à la horde. Et il ne remplace pas la build comme moteur de rejouabilité — c'est un supplément pour qui veut se mesurer, pas la raison de relancer.

### Le classement, et pourquoi la graine existe

Ce que la graine garantit est précis, et ce n'est pas ce qu'on croit d'abord : **la même suite de décisions produit la même run**. Deux joueurs qui affrontent la même graine sur le même lieu ne rencontrent pas les mêmes vagues aux mêmes instants — le report du budget d'apparition et le recyclage de la traîne dépendent du trajet, donc de la façon de jouer. Ils partagent un terrain et une distribution, pas une séquence.

C'est un speedrun sur la même carte, pas un puzzle identique. Et c'est mieux ainsi : celui qui esquive mieux subit moins de report, donc moins de pression — le classement récompense le jeu plutôt que la chance d'un tirage.

D'où deux classements possibles, sans serveur : la **graine du jour**, identique pour tout le monde, et le classement par lieu partagé, où celui qui diffuse un niveau diffuse aussi sa graine. Un fichier de niveau tenant dans un message, le défi se partage avec lui.

**Un score se vérifie par rejeu, il ne devient pas inviolable pour autant.** Le journal de run — un `Input` par tick, sous-produit gratuit du chapitre 15 — se joint au score : qui veut vérifier rejoue, et retrouve le même total ou ne le retrouve pas. Quelques kilo-octets une fois les répétitions encodées, donc joignable à un message comme le lieu lui-même.

Ce que cela établit est la cohérence entre un score et un journal, ce qui suffit à éliminer le fichier modifié — le cas réel. Cela n'établit pas qu'un humain a joué : un journal produit par un programme passerait la vérification sans faute. La distinction décide de ce qu'on affiche — un journal joignable au score, jamais la mention « vérifié », qui promettrait une garantie qu'aucun classement sans serveur ne peut donner. Tant qu'il n'y a rien à gagner, la fraude n'a d'ailleurs pas d'enjeu.

---

## 7. Les ressources : les caisses

Le joueur casse une caisse **en la traversant**. Aucune touche, aucun conflit avec l'auto-visée, et ça le garde en mouvement.

**Ce qui en est écrit aujourd'hui est une sonde, et le reste de ce chapitre attend son étape.** Une caisse est posée par le lieu, se casse au contact du joueur et laisse des gemmes ; elle n'a ni délai de contact, ni ralentissement, ni coût dans le champ de flux, et son contenu n'est pas visible avant la casse. Ce qui est déjà tenu est la seule règle qui coûterait cher à rétablir plus tard : **elle n'est pas une cible.** Rangée parmi les ennemis, elle détournerait la visée automatique — qui prend la plus proche sans que le joueur choisisse — et emporterait avec elle la mécanique du Secouriste ; c'est la même règle que pour le figurant, et pour la même raison.

**Ce qu'une caisse laisse est un réglage de partie, pas de lieu.** Un auteur écrit où sont ses caisses, il n'écrit pas ce qu'elles donnent : sans cela il règle la difficulté de sa salle par son butin, et une caisse cesse de signifier la même chose d'un lieu à l'autre. C'est la règle qui vaut déjà pour la valeur d'une gemme.

Trois règles rendent la mécanique juste :

**Un temps de contact.** Pas de destruction au frôlement : la caisse cède après environ un tiers de seconde d'appui, avec une déformation visible pendant le délai. On ne casse pas en passant, on casse en décidant d'y aller.

Ce délai a ses propres images : un cycle d'appui qui boucle tant que le joueur pousse — la caisse s'écrase en s'élargissant —, puis un cycle de rupture qui ne boucle pas et s'achève sur l'épave au sol. Sans ces images, le joueur ne sait pas qu'il casse quelque chose et croit à un blocage.

**Un ralentissement pendant le contact.** C'est le vrai coût de la ressource, et le seul qui compte : ramasser, c'est perdre du terrain. Sans ça, il n'y a plus de choix, on ramasse tout, tout le temps.

**Une distinction ferme entre caisse et obstacle.** Silhouette différente, teinte réservée, liseré lumineux. Si les caisses ressemblent aux piliers, le joueur ne sait jamais ce qui va céder et ce qui va le bloquer.

Le contenu est **visible avant la casse** — icône flottante ou liseré coloré. Sinon le joueur casse tout systématiquement et ce n'est plus un choix, c'est une corvée.

Contrainte technique : **la passabilité n'est pas un booléen, c'est un coût par case.** Une caisse ne bloque pas et ne se franchit pas librement : elle coûte cher à traverser, ce qui est exactement le ralentissement décrit plus haut. Un mur, lui, a un coût infini.

C'est ce qui fait tenir ensemble les trois règles de la caisse, qu'un booléen rendait contradictoires — on ne ralentit pas ce qui est arrêté, et un joueur acculé ne se dégage pas à travers ce qui bloque. Et ce que la conception veut par ailleurs devient gratuit : la flaque, le sol sale et le sol fissuré ralentissent ce qui les traverse sans qu'aucun mécanisme nouveau soit écrit, et un profil pourra ignorer un coût que le joueur paie.

Le champ de flux devient donc un parcours pondéré. **Un tri par seaux, pas un tas** : les coûts sont trois ou quatre valeurs entières, ce qui ramène le calcul au même temps linéaire qu'un parcours en largeur ordinaire — un Dijkstra général se paierait pour une variété de coûts qui n'existe pas ici. La destruction d'une caisse déclenche un rafraîchissement local, pas un recalcul complet.

**Le coût se paie au déplacement, et il divise la vitesse.** Sans quoi le parcours pondéré serait une superstition : il contournerait au prix de deux cases ce qui ne coûte rien à traverser, et l'écart entre le chemin choisi et le chemin payé ne se verrait nulle part.

**Une seule grille, lue par toutes les entités**, joueur compris — la caisse veut déjà le ralentir. Deux grilles se paieraient à chaque destruction, où il faudrait garder d'accord deux rafraîchissements locaux ; et le jour où un profil devra ignorer un coût, la Buse qui glisse ou le Molosse qui charge, c'est un champ de profil, pas une seconde grille.

La case de référence est **celle du point d'appui, avant le pas**. Le point d'appui parce qu'une entité est presque toujours à cheval sur deux cases et que c'est déjà lui qui la situe partout ailleurs, tri en profondeur compris ; avant le pas parce que diviser par le coût de la case d'arrivée est circulaire — le pas plein entre dans la flaque, la division le raccourcit, il n'y entre plus, et l'entité tremble à la frontière. L'écart avec ce que le parcours a compté est d'un tick à l'entrée et d'un tick à la sortie, qui se compensent sur une traversée.

Corollaire à connaître : **le coût s'échantillonne au tick, il ne s'intègre pas.** Une entité assez rapide pour franchir une case entière en un pas ne paierait jamais une case isolée, son point d'appui n'y séjournant à aucun tick. Sans portée aujourd'hui — il faudrait soixante tuiles par seconde là où le joueur en fait cinq — mais c'est ce qui cède en premier le jour où un coût élevé servira à ralentir quelque chose de mince.

### Les obstacles destructibles

Un mur ne cède **jamais** sous la pression d'une horde : la géométrie n'est pas négociable, sinon le Vigile perd son intérêt, la signalétique son objet, et l'éditeur ne peut plus rien valider — une carte qui change en cours de partie n'est plus vérifiable.

En revanche, l'auteur d'un niveau peut poser des **obstacles fragiles**, prévus dans la topologie et donc validables comme le reste :

| Obstacle | Touches | Ce qu'il ouvre |
|---|---|---|
| grille de ventilation | 3 | un raccourci, presque gratuitement |
| vitrine | 5 | une boutique, et on voit à travers avant de casser |
| cloison de placo | 8 | une réserve, un mur qui n'en était pas un |
| rideau de fer | 20 | une vraie décision : vingt touches sous la horde |

Ils se cassent **sur une touche d'interaction**, en se tenant contre eux, et jamais sous le tir de base — qui ne cible que des ennemis et ne saurait pas distinguer un rideau de fer d'une créature.

Ce qui les sépare de la caisse n'est donc pas la nature du dégât mais le geste : la caisse cède à l'appui, en la traversant, et n'interrompt pas la course ; un destructible demande de s'arrêter et d'appuyer. C'est ce qui donne son prix aux vingt touches du rideau de fer — vingt touches immobile sous la horde, contre un tiers de seconde de ralentissement pour une caisse. Les armes lourdes, elles, emportent ce qui se trouve dans leur zone, destructibles compris : c'est un usage de plus pour une charge, et une raison d'en garder une.

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

### Les passifs montent par paliers

Un axe se prend plusieurs fois, et c'est ce qui remplit une trentaine de montées avec six axes plutôt que trente entrées de table.

Trois règles, qui visent toutes la même chose — la stratégie unique consistant à empiler un seul axe du début à la fin :

- **Chaque palier coûte plus que le précédent.** Sans quoi le troisième projectile vaut le premier, alors qu'il apporte proportionnellement moins.
- **Chaque axe a une borne**, six paliers pour commencer. L'épuiser devient un moment de jeu : il faut basculer sur un axe qu'on n'avait pas choisi.
- **L'offre dépasse la demande.** Six axes de six paliers font trente-six choix pour une trentaine de montées : le joueur laisse l'équivalent d'un axe derrière lui. L'inverse ne produirait pas un choix difficile mais un écran de montée de niveau vide, ce qui est le pire défaut possible à cet endroit — il tombe sur le moment de récompense.

Reste la run exceptionnelle qui dépasse trente-six montées. Une carte inépuisable la couvre, mais elle **n'entre dans le tirage que lorsqu'il ne reste plus assez de paliers pour remplir les trois cartes** — sinon elle occupe une place trente fois pour un cas qui arrive une.

Sa nature n'est pas indifférente : un soin fait l'affaire, précieux quand on est bas et ignorable quand on est haut, donc situationnel. Un gain d'expérience serait un mauvais choix — il accélère la montée suivante, donc il se rembourse, donc il est toujours au moins aussi bon que ce qu'il remplace, et il aplatirait la table entière sans qu'on voie pourquoi.

Et rien n'interdit le cas extrême, celui où tout est épuisé : la soupape doit alors **remplir les trois places à elle seule**. Elle est donc répétable dans un même tirage, ou il en existe plusieurs de natures différentes — faute de quoi on retombe sur l'écran vide qu'elle était censée éviter, au moment précis où le joueur a le mieux joué.

**Ce que l'étape 3 livre est une table réduite, et c'est une décision plutôt qu'un provisoire mal fini.** Deux axes — cadence et portée —, six paliers chacun, et la soupape répétable. Le perforant, le ricochet et l'éventail demandent du travail dans le tir qui appartient à l'étape 6 ; le nombre de projectiles est écarté parce que le mécanisme écrit ne produit aujourd'hui ni un nombre ni un éventail mais une salve confondue qui perfore, et trancher laquelle des trois lectures est la bonne ne se fait pas sous la pression d'avoir un axe de plus.

Deux conséquences à assumer, puisqu'elles portent sur ce que le jalon peut juger. **La soupape cesse d'y être une soupape** : douze paliers pour une trentaine de montées, elle occupe le tiers final de chaque run et devient un mécanisme. C'est acceptable pour éprouver le choix, ce ne le sera plus à l'étape 6. Et **l'offre ne varie pas** : deux axes plus la soupape font exactement trois cartes, donc aucun tirage n'a lieu. Le jalon exerce l'arbitrage entre deux axes, il n'exerce pas « la bonne carte est celle qui fait hésiter » — c'est une limite du jugement, connue et datée, pas un défaut du lot.

**Le revers à assumer** : le contraste des patterns disparaît, alors que c'est un moteur de rejouabilité du genre. Il se récupère par deux voies qui ne coûtent aucune animation de personnage — les armes lourdes ramassées dans les caisses, qui sont un effet et non une pose, et les effets qui ne partent pas du personnage : zone au sol, orbite, onde de choc, dessinés au sol ou autour du joueur.

### La visée

Le tir de base est **automatique** et vise **le plus proche, dans toutes les directions**. Il n'y a pas de cône, et rien ne retient le tir : s'il existe une cible à portée, ça part.

Un cône avant a d'abord été retenu, avec la règle « cône vide, pas de tir ». Les deux sont abandonnés, et pas au vu d'une mesure : le conflit est logique et aucune partie ne l'aurait tranché autrement. Le chapitre 1 pose que le joueur ne contrôle que son déplacement, le chapitre 4 que tout son jeu est du kiting — et kiter, c'est avoir la horde derrière soi. Un cône avant ferait donc de la fuite un moment sans dégâts, et le seul moyen de tirer serait de cesser de fuir. Aucun angle ne répare ça.

Ce que le cône apportait vraiment était visuel, et se garde sans lui : **le sprite s'oriente sur la cible**, pas sur le déplacement. Les 8 directions étant fournies, reculer en tirant vers l'avant se lit immédiatement. Sans cible à portée, il s'oriente sur le déplacement — un personnage figé dans une direction morte se lirait comme un défaut.

Contrainte : les projectiles ne ciblent **jamais** le mobilier. Si l'auto-visée choisit une caisse plutôt qu'un sprinteur qui charge, le joueur meurt sans comprendre. C'est aussi pourquoi les obstacles destructibles ne se cassent pas au tir de base — voir le chapitre 7.

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

Cette table est **`assets/armes/manifeste.json`, tenu à la main** — l'un des deux de `assets/` qui ne sortent pas d'un générateur, avec celui de la progression. C'est délibéré : c'est le fichier qu'on rouvrira le plus souvent pendant l'équilibrage, et le loger dans un manifeste généré ferait passer chaque réglage de cadence par un script Python, donc par une régénération de six cents images pour changer un chiffre. La boucle courte compte davantage ici que l'uniformité.

Il porte dès l'étape 1 ce que le tir automatique réclame : cadence, portée, dégâts, nombre de projectiles. Sans lui, ces valeurs deviennent des constantes Go et l'invariant des données tombe à la première ligne écrite.

Une option reste ouverte si le tir manuel manque au jalon 3 : garder l'automatique et ajouter un **tir d'appoint** sur touche, avec temps de recharge. Le joueur passif ne perd rien, le joueur actif gagne un peu.

---

## 10. Le rendu

### Les directions

`figurines.py` produit 8 directions, ce qui lève la contrainte initiale de 4 poses. Si un archétype devait un jour être dessiné à la main, la solution de repli tient toujours : quatre orientations sur les diagonales écran (NE, SE, SO, NO), dont deux obtenues par miroir horizontal — on ne dessine alors que dos et face.

Dans tous les cas : orienter le sprite sur la direction de **visée**, pas de déplacement. Le joueur recule en tirant vers l'avant, et ça se lit immédiatement.

### La grille et les tailles

Tout découle des sprites de personnages, en 64×64 : la tuile de sol fait **64×32**, projection 2:1.

**Le point du monde (u, v) est le sommet nord du losange de la case**, et non son centre — le centre tombe en (u+0,5 ; v+0,5), seize pixels plus bas. C'est la convention que le chapitre 12 pose déjà pour les pièces, où « la case (0, 0) est le sommet du losange ». L'ancrage d'une image, lui, n'est pas ce point : il est déclaré forme par forme dans le manifeste, parce qu'un mur de 64×96 et une flaque plate ne se posent pas de la même façon.

Pour un objet couvrant plusieurs tuiles, `largeur = (tx + ty) × 32` et l'emprise au sol `hauteur = (tx + ty) × 16`.

Ce sont là des tailles d'image, donc des pixels, et elles n'appartiennent qu'au rendu. La simulation, elle, ne connaît que la tuile — voir « Les repères » au chapitre 15.

| Élément | Taille image | Élévation au-dessus du sol |
|---|---|---|
| Tuile de sol | 64×32 | 0 |
| Obstacle bas (gondole, banc, caddie) | jusqu'à 96×62 | 24 max |
| Mur plein | 64×96 | 64 |
| Caisse cassable | 32×32 | 16 |
| Personnage et ennemis | 64×64 | appui en (32, 63) |
| Gemme d'XP | 10×8 | — |

La règle des **24 pixels** est celle qui compte : au-delà, l'objet masque un personnage de 64 et devient un piège visuel. Ce qui la dépasse — murs, piliers, véhicules, wagons, immeubles — n'est pas un obstacle à contourner en pleine action mais une limite de zone, et le manifeste le déclare forme par forme. Ce que le rendu en fait est plus bas, à « La silhouette plutôt que la transparence ».

**Résolution interne : 960×540**, agrandie en entier vers la fenêtre. Un tampon de 480×270 ne montrerait que 7 tuiles de large, bien trop serré pour voir la horde arriver ; 960×540 en donne une quinzaine et se multiplie par 2 pour du 1080p, donc pixels carrés garantis.

Deux choses distinctes, et elles n'arrivent pas ensemble : **la fixité du tampon est acquise dès l'étape 2**, l'agrandissement en entier attend l'étape 15. Le facteur dépend du facteur d'échelle du système, qu'il faut lire séparément, et il oblige à choisir ce qu'on fait du reste de la fenêtre — bandes noires, taille contrainte, ou redimensionnement par pas. C'est un réglage d'affichage, et il vit avec le plein écran et la résolution.

### Les trois hauteurs

Chaque tuile porte une hauteur parmi trois : sol, obstacle bas qu'on voit par-dessus, mur plein. C'est ce qui permet de juger la lisibilité d'une salle **avant** de la lancer, et c'est indispensable à l'éditeur en vue de dessus (voir plus bas).

**Les trois se dérivent, et la passabilité décide d'abord** : ce qui se franchit est du sol, ce qui bloque est un obstacle bas jusqu'à 24 pixels et un mur au-delà. L'élévation ne départage donc que ce qui arrête.

L'ordre compte, parce que la vue de dessus sert à voir **où l'on passe**. Une porte ouverte dépasse de 48 pixels et reste un passage — c'est même la seule chose que l'auteur ait besoin d'y lire ; un quai, un trottoir, un rail dépassent du sol et se marchent. Les classer sur leur seule hauteur afficherait des obstacles là où il n'y en a pas, et la lecture topologique ne vaudrait plus rien.

Le rendu, lui, ne lit pas cette catégorie : il a l'élévation et le drapeau qui dit qu'une forme masque un personnage. Une propriété qui servirait aux deux finirait par mal servir les deux.

### Le tri en profondeur

Le rendu iso a besoin d'un tri par `Y` écran : un tri par compartiments, pas un `sort.Slice` général à chaque frame.

**À égalité, la clé doit être totale et stable.** Le tri range par seau de profondeur ; deux entités d'un même seau sont départagées par leur profondeur exacte, puis par leur abscisse écran, puis par leur sorte, et enfin par leur identifiant. Sans ces derniers critères, l'ordre dépend du parcours du bassin, qui change à chaque suppression par échange : deux sprites superposés se relaieraient au premier plan d'une image à l'autre, et le scintillement se voit immédiatement.

**La sorte est un critère et non une décoration** : chaque bassin numérote ses entités pour lui seul, si bien qu'un ennemi et un projectile peuvent porter le même identifiant sans avoir rien de commun.

**La génération n'entre pas dans la clé**, et le bassin refuse même de la rendre ici : il donne l'identifiant seul, jamais une référence complète, parce qu'une référence sortie pour un tri finirait conservée quelque part et l'invariant tomberait par la porte qu'on aurait ouverte. L'identifiant suffit à ce qu'on lui demande — ne jamais bouger, là où la place change parce qu'une *autre* entité est morte.

**Le joueur passe devant tout ce qui partage sa profondeur.** C'est une exception assumée à la règle de tri : perdre son personnage sous un empilement d'ennemis est la pire chose qui puisse arriver à la lisibilité, et cela survient précisément au moment où l'on est encerclé, c'est-à-dire quand il faut voir clair.

**Les cadavres passent sous tout ce qui est vivant**, à profondeur égale. C'est l'exception symétrique, et pour la même raison : en fin de run, un tapis de cadavres finirait par masquer la horde qui arrive.

### La silhouette plutôt que la transparence

Une forme qui dépasse 24 pixels cache un personnage. Trois façons de le traiter, et deux d'entre elles paient la lisibilité du lieu pour celle du personnage :

- **l'opacité fixe** sur la forme entière — triviale à écrire, mais les élévations montent jusqu'à 120 pixels, et un immeuble à demi effacé retire plus d'information qu'il n'en donne ;
- **la découpe** autour du personnage, en dégradé — le plus beau rendu, au prix d'un shader et d'un second passage ;
- **la silhouette** — le décor reste opaque, et c'est le personnage qu'on redessine par-dessus, en aplat, quand une forme le masque.

**C'est la silhouette**, et l'argument tient en une phrase : la transparence retire de l'information au décor, la silhouette en ajoute au personnage. Le chapitre 2 veut les deux — voir la horde arriver *et* comprendre où sont les murs —, et c'est la seule des trois qui n'enlève rien.

Elle traite en outre un cas qu'aucune propriété du décor ne pourra jamais décrire : un joueur derrière trois Vigiles empilés est invisible sans qu'aucune forme ne soit en cause. Une transparence ne sait effacer que ce qu'un manifeste a déclaré ; la silhouette ne dépend que de ce qui est dessiné devant.

Ce qu'elle suppose est déjà là : savoir quelle entité est masquée, c'est-à-dire ce que le tri en profondeur calcule. Aucun mécanisme de plus, seulement un second blit en teinte plate.

**Elle ne révèle que le joueur et les projectiles ennemis**, comme le chapitre 2 l'exige, et jamais un ennemi : voir la horde à travers un bus retirerait au décor le seul pouvoir qu'il a sur le combat. Ce qui se cache derrière un camion doit rester une inconnue.

Le champ du manifeste s'appelle donc `masquant`, par symétrie avec `bloquant` : deux adjectifs, deux constats, et aucun des deux ne dit ce que le moteur doit en faire. Le nommer d'après la technique aurait figé dans la donnée une décision de rendu.

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

```json
{
  "$comment": "Rayon central. Mur plein au nord, ouverture pleine au sud.",
  "identifiant": "rayon_long",
  "jeu": "supermarche",
  "version_format": 1,
  "taille": [8, 4],
  "aire_ouverte": 0.62,
  "cotes": { "nord": "ouverture", "est": "mur", "sud": "ouverture", "ouest": "mur" },
  "grille": [
    "########",
    "#......#",
    "#..O...#",
    "........"
  ],
  "ancrages": [
    { "type": "apparition", "position": [2, 3] },
    { "type": "caisse", "position": [4, 2] }
  ]
}
```

**La grille est dans le même fichier, et sous forme de lignes.** Un fichier séparé serait une seconde description du même objet : une pièce dont la grille annonce seize cases et dont le descripteur en déclare douze ment, sans qu'on sache lequel des deux. Et des lignes plutôt qu'un tableau de nombres parce qu'on **voit la pièce** en lisant le fichier — l'exemple ci-dessus a son mur au nord et son ouverture au sud, cela se lit sans rien décoder. Pendant les étapes où les pièces s'écrivent à la main, cela vaut mieux qu'un mur de virgules ; en revue, un diff montre les lignes changées.

Deux règles rendent la forme exploitable :

- **Une chaîne par `v` croissant, un caractère par `u` croissant** — l'ordre direct des axes du losange, celui que les emprises du manifeste emploient déjà. Un format cohérent avec le reste s'oublie moins qu'un format justifié.

  Attention à ce que `grille[v][u]` ne suggère pas : la première dimension n'est **pas une ligne d'écran**. En projection 2:1, `u` croissant descend vers le sud-est et `v` croissant vers le sud-ouest ; la case (0, 0) est le sommet du losange, et la première chaîne est l'arête qui en part vers la droite. Les côtés se nomment donc par la grille et jamais par l'écran — `nord` est `v = 0`, `sud` est `v` maximal, `ouest` est `u = 0`, `est` est `u` maximal — ce que `cotes` employait déjà sans le dire. L'exemple est volontairement asymétrique : son mur est au nord au sens de la grille, et il apparaît à l'écran comme l'arête haute-droite du losange.
- **Toutes les chaînes ont la même longueur, et elle vaut la taille annoncée.** C'est le désaccord de dimensions qu'un fichier séparé aurait rendu indétectable : il ne disparaît pas en fusionnant les deux, il se déplace à l'intérieur du fichier, où il se vérifie d'un coup au chargement.

La correspondance entre un caractère et une tuile n'est pas dans la pièce mais dans son **jeu de pièces** : c'est là que vivent déjà l'atlas et la taille de tuile, et une palette est de même nature. La déclarer par pièce la dupliquerait, avec le même caractère désignant deux choses selon le fichier.

C'est aussi le premier usage concret d'`empreinte_jeu_pieces` : un caractère réattribué dans un thème change le sens de **toutes** ses pièces à la fois, et en silence. L'empreinte est ce qui transforme ce désastre en message explicite.

La passabilité et les hauteurs sont **dérivées des propriétés des tuiles au chargement**, jamais saisies à la main : l'auteur ne sait même pas que le flow field existe.

**Tout est en JSON, y compris ce que le TOML rendrait plus agréable à écrire.** Un lieu partagé est du JSON compact compressé — c'est ce qui donne les 548 caractères mesurés plus bas —, et un second format sur le même objet imposerait une conversion pour partager, donc deux représentations qui finiraient par diverger. Le format retenu est celui que le jeu lit toute sa vie, pas celui qui arrange les neuf étapes pendant lesquelles les pièces s'écrivent à la main. Corollaire pratique : rien à lire hors de la bibliothèque standard, et une seule forme d'en-tête de licence.

Ce qu'on y perd est le commentaire, et `$comment` le rend : il est **autorisé partout, pas seulement en première clé**, où il porte la mention de licence. Un fichier partagé se lit en refusant les clés inconnues — sans quoi `rotaton` au lieu de `rotation` se charge en silence avec une valeur par défaut, et l'auteur ne comprend pas pourquoi sa pièce est de travers. `$comment` est la seule clé exemptée de ce refus.

**Les `version_format` du lieu et de la pièce sont indépendants.** Un lieu circule entre joueurs, une pièce reste dans le binaire ; le jour où une pièce gagne un champ, les lieux publiés ne deviennent pas suspects pour autant. Ce sont deux contrats, avec deux durées de vie et deux migrations.

### Un jeu de pièces (thème)

```json
{
  "identifiant": "supermarche",
  "nom": "Supermarché",
  "auteur": "stephane",
  "version": "1.2.0",
  "version_format": 1,
  "atlas": {
    "fichier": "atlas.png",
    "taille_tuile": [64, 32],
    "variantes_rotation": true
  },
  "palette": {
    ".": "sol",
    "#": "mur",
    "O": "pilier"
  },
  "ambiance": {
    "musique": "neons.ogg",
    "teinte": "#c8d4e0",
    "luminosite": 0.8
  }
}
```

### Une campagne

**Une campagne est un dossier de lieux, et c'est l'unité que l'on compose, partage et choisit.** Le graphe de salles du chapitre 2 vit là : par où l'on commence, où mènent les portes, comment les branches se rejoignent.

```
assets/campagnes/monjeu/
    campagne.json          identifiant, lieu de départ, portes et branches
    supermarche_nuit/
        lieu.json
        jeu.json
        pieces/
            carrefour.json
    parking/
        …
```

**Le dossier existe pour cloisonner l'espace de noms**, exactement comme celui d'un lieu le fait pour ses pièces. À plat, deux auteurs nommeront tous les deux une salle `parking`, et celui qui reçoit les deux en perd une. C'est ce qui écarte l'autre forme possible — des lieux à plat et un fichier de graphe qui les cite par identifiant : plus économe, mais elle remet le vocabulaire global que le dossier existe pour supprimer.

Le jeu de pièces reste dans le lieu et ne remonte pas à la campagne, bien que quatre salles d'un même thème le recopient. C'est ce qui garde un lieu autonome et extractible seul, et c'est le prix déjà accepté un cran plus bas pour les pièces.

Deux conséquences à assumer. **L'unité de partage est la campagne**, pas le lieu — une campagne d'un seul lieu tient toujours dans un message, et cela donne une seule chose à partager au lieu de deux régimes. Et **le catalogue énumère des campagnes** : c'est le nom que l'auteur donne à ce qu'il a fait.

Ce que le descripteur porte aujourd'hui, c'est l'identifiant, le nom lisible et le lieu de départ :

```json
{
  "version_format": 1,
  "identifiant": "demonstration",
  "nom": "Démonstration",
  "lieu_depart": "place"
}
```

**Le graphe arrive à l'étape 8, avec les portes.** Il prendra la forme d'une liste de nœuds portant chacun ses sorties, ce qui donne une campagne linéaire quand chaque nœud n'en a qu'une et un choix de branche quand il en a deux. L'écrire maintenant ferait un champ que personne ne lit, et la validation qui l'accompagne — tout nœud atteignable depuis l'entrée, au moins un chemin complet, pas de nœud orphelin — n'aurait rien à valider.

**Le nom du dossier et le champ `identifiant` s'accordent**, comme pour un lieu et pour la même raison : celui qui duplique une campagne pour en faire une variante renomme le dossier, oublie l'identifiant, et charge une copie qui se croit l'original.

### Un lieu

Un lieu n'est pas une carte, c'est une **liste de pièces posées**. Quelques centaines d'octets.

**Un lieu est un dossier**, qui porte son `lieu.json`, son jeu de pièces et ses pièces. Trois raisons, dont la première suffirait : à plat, l'espace de noms des pièces est global, et deux auteurs qui nomment chacun la leur `carrefour` s'écrasent. Un lieu qu'on télécharge se suffit alors à lui-même — sinon il ne se télécharge pas, il s'installe. Et `empreinte_jeu_pieces` ne veut dire quelque chose que si deux exemplaires du même jeu peuvent différer, ce qui suppose précisément que chacun emporte le sien.

**Ce qui existe une fois par dossier porte un nom fixe ; ce qui existe en plusieurs exemplaires garde son identifiant et se range dans un sous-dossier qui dit sa nature.** Un lieu porte donc `lieu.json` et `jeu.json`, et ses pièces vivent dans `pieces/`. Sans cette règle, le jeu de pièces et une pièce sont deux noms libres au même niveau — `quartier.json` posé à côté de `carrefour.json` ne dit pas lequel est une palette et lequel est un plan, et « quartier » se lit même comme un endroit qu'on construirait.

L'identité ne se perd pas pour autant : elle vit dans le champ `identifiant`, comme un lieu nommé « place » n'a jamais eu de fichier `place.json`. Elle cesse en revanche d'être vérifiée par le chemin, ce qui oblige à la contrôler explicitement — un `jeu.json` déposé dans le mauvais lieu se chargerait sinon en silence avec sa palette, donc en changeant le sens de tous les caractères des pièces.

Un dossier de lieu se reconnaît alors sans être ouvert, et renommer un lieu se fait en renommant son dossier. Le nom du dossier et le champ `identifiant` décrivant dès lors la même chose, **le chargeur exige qu'ils s'accordent** — sans quoi celui qui duplique un lieu pour en faire une variante renomme le dossier, oublie l'identifiant, et charge une copie qui se croit l'original.

```json
{
  "version_format": 1,
  "identifiant": "supermarche_nuit",
  "jeu_pieces": "supermarche@1.2",
  "empreinte_jeu_pieces": "a41f7c92",
  "pieces": [
    { "id": "entree_caisses", "u":  0, "v": 0 },
    { "id": "rayon_long",     "u": 24, "v": 0 },
    { "id": "carrefour",      "u": 56, "v": 0 }
  ],
  "pieces_personnalisees": [],
  "scenario": "standard_4min",
  "objectif_sortie": { "type": "kills", "valeur": 250 },
  "densite_caisses": 0.4
}
```

**Les axes sont `u` et `v`**, ceux que ce document pose plus haut et que le lieu livré emploie ; un lieu écrit avec `x` et `y` est refusé, le décodage n'admettant aucune clé inconnue. Et tout ce que montre cet exemple n'existe pas encore : la rotation reste à trancher, le scénario et l'objectif de sortie sont les étapes 4 et 8. Ce qu'un lieu porte aujourd'hui se lit dans `assets/campagnes/demonstration/place/lieu.json`.

**`u` et `v` sont la case d'origine de la pièce, pas son rang dans une trame.** Le lieu livré pose des blocs de trente-deux cases et une enceinte qui n'en fait qu'une d'épaisseur : des pièces de tailles différentes se composent dans un même lieu, ce qu'un rang ne saurait pas exprimer. Cette version du document portait un champ `grille` et des positions de rang, qui n'ont jamais été lus — c'est la pose de l'enceinte qui a rendu la contradiction visible.

**Un lieu couvre son étendue exactement une fois, et le chargeur refuse les deux écarts.** Une case qu'aucune pièce ne pose garde le coût d'une grille neuve, celui d'un sol ordinaire : le trou se traverse, ne se dessine pas, et n'apparaît qu'au moment où une créature y flotte. Et deux pièces qui se recouvrent se départagent par l'ordre des poses, la dernière écrivant par-dessus la première — un ordre que rien n'annonce et dont aucun auteur d'éditeur n'aura idée. Les refuser rend cet ordre sans effet, ce qui vaut mieux que de l'écrire quelque part.

`pieces_personnalisees` embarque les pièces peintes à la main, quand il y en a — et il reste vide tant que le mode tuiles n'est pas implémenté. C'est le seul cas où le fichier grossit ; même alors, une pièce de 16×16 tuiles compressée pèse quelques centaines d'octets.

`empreinte_jeu_pieces` est une somme de contrôle du jeu de pièces, pas seulement son numéro de version. Sans elle, une pièce retouchée sans changement de numéro produit des niveaux qui **se chargent en silence avec une géométrie différente** de celle qu'a construite leur auteur. Quelques octets de plus, et le message devient explicite au lieu d'être trompeur.

Un niveau qui ne référence que des pièces officielles ne contient donc que des identifiants et des positions : le destinataire possède déjà les tuiles, les objets et les images. Mesuré sur un supermarché de douze pièces, ça fait 1189 octets lisibles, 902 compacts, **548 caractères une fois compressé et encodé en base64** — copiable dans un message.

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

Étape 1, la plus robuste : un dossier `campagnes/` scanné au démarrage, une archive par campagne avec une extension propre au jeu, l'identifiant faisant office de clé. Glisser-déposer et ça marche. C'est déjà tout ce dont une communauté a besoin pour échanger sur un salon Discord.

Mieux : un lieu qui ne référence que des pièces officielles tient dans quelques centaines d'octets. Encodé en base64, il se colle dans un message ou tient dans un QR code. Pas de serveur, pas de somme de contrôle, pas d'atelier — on copie une chaîne.

Étape 2, si le besoin apparaît : un catalogue, Steam Workshop ou un simple index JSON servi par le site avec téléchargement et somme de contrôle. Le format de paquet ne change pas, seul le transport change.

---

## 14. Les ressources graphiques et sonores

Le format suppose qu'existent un atlas de tuiles, des feuilles de sprites et de la musique. Rien de tout ça n'est produit par le moteur, l'éditeur ou le format : c'est le poste le plus coûteux du projet, celui qui ne bénéficie d'aucun raccourci technique.

Ordre de grandeur : une quinzaine de pièces par thème, cinq thèmes, plus sept archétypes d'ennemis et un joueur.

### Le style : pixel art

Décision arrêtée. Elle est esthétique — le registre des jeux d'arcade et des isométriques des années 90 — mais elle a trois conséquences pratiques qui pèsent lourd :

- **Les décors manquants deviennent faisables soi-même.** Un rayonnage de supermarché en 64×32, ce n'est pas de l'illustration, c'est de la géométrie et trois teintes. Rayons, caddies, tourniquets, distributeurs : des soirées de travail plutôt qu'une commande.
- **La taille des sprites reste raisonnable** aux effectifs du chapitre 4 — 150 entités à l'écran en pic —, ce que du 3D précalculé haute résolution ne permettrait pas.
- **Le mélange de sources devient possible**, à condition de tenir la palette (voir ci-dessous).

Corollaire : la voie du rendu 3D précalculé (modèles Mixamo ou low-poly rendus dans Blender) est écartée. Elle reste notée ici comme option de repli si le bestiaire devait grossir au point que le dessin à la main ne suive plus.

### Les trois règles du pixel art

À poser dès le premier asset, pénibles à corriger ensuite.

**Une palette fermée.** Trente-deux couleurs, pas plus, fixées dans `MATIERES` et partagées par tous les générateurs. Tout asset entrant est recoloré dessus. C'est le seul moyen de rendre cohérents des paquets venant d'auteurs différents et ses propres tuiles.

**Aucun filtrage à l'affichage.** Échantillonnage au plus proche voisin, et caméra déplacée en pixels entiers, jamais en flottants. Le scrolling sous-pixel fait scintiller le pixel art : c'est le défaut qui trahit immédiatement un jeu bâclé.

**Une résolution interne fixe.** Rendu dans le tampon de 960×540 arrêté au chapitre 10, agrandi en entier vers la fenêtre — l'agrandissement étant réglé à l'étape 15, avec le reste de l'affichage. Pixels carrés à toutes les tailles d'écran une fois qu'il l'est, et surtout décision de game design déguisée en détail technique : ce tampon détermine à quelle distance le joueur voit arriver la horde. C'est sa fixité qui est la règle de pixel art ; sa valeur, elle, est déjà tranchée.

### La lisibilité en masse

Point de vigilance sans précédent à copier : les jeux rétro n'ont jamais affiché autant de sprites simultanés. Contours foncés sur les ennemis, teinte réservée au seul personnage joueur, projectiles ennemis dans une couleur qui n'existe nulle part ailleurs dans la palette. Une palette fermée rend cette discipline tenable ; des sprites rendus l'auraient rendue impossible.

### La lisibilité du texte

Les libellés s'écrivent en casse mixte, jamais en capitales. Ce n'est pas une concession à la fonte mais le meilleur choix en soi : les mots en capitales ont tous la même silhouette rectangulaire, et c'est la silhouette qui porte la lecture rapide. Une carte d'amélioration se lit en une seconde, et « Cadence +15 % » s'y lit plus vite que « CADENCE +15 % ».

L'ordre de ces deux raisons compte. Les capitales accentuées sont aussi ce qui déborde le plus souvent d'une cellule bitmap, mais adosser la règle à cette limite la ferait tomber au premier changement de fonte : on remettrait des capitales au motif que la nouvelle les supporte, en perdant ce qu'on avait gagné en lecture.

Un libellé d'emplacement porte la **touche**, jamais le nom. L'icône dit déjà de quoi il s'agit, et ce que le joueur cherche sous la case en jouant est ce qu'il doit presser. Un « 1 » tient là où « Aimant » déborde de sa case, mais la largeur n'est pas la raison : le nom serait au mauvais endroit même s'il tenait.

**Cette règle vise ce qu'on lit en jouant, donc sous pression, et rien d'autre.** Un panneau qui met le jeu en pause n'est pas dans ce régime : il a un titre, où l'instruction s'écrit une fois pour tous ses choix au lieu d'être répétée sous chacun. C'est ce qui permet à l'écran de montée de niveau de dire « les flèches désignent, Entrée valide » plutôt que d'étiqueter trois cartes. Sans cette précision, quelqu'un appliquera la règle au panneau et retirera l'instruction, qui est la seule chose y indiquant qu'on choisit au clavier.

**Les chiffres appartiennent aux emplacements et les gardent toute la partie.** Ce qui a écarté les chiffres du choix des cartes n'est pas l'encombrement mais l'asymétrie du coût : une carte mal choisie se rattrape à la montée suivante, un aimant déclenché à vide est perdu jusqu'à la prochaine apparition.

**On désigne puis on valide, et la première version prenait d'un coup.** La flèche prenait la carte de sa place : le geste portait le sens sans qu'aucune étiquette l'explique, et ce document en concluait que les flèches étaient libres puisque le déplacement est suspendu. Elles le sont dans le code, pas sous les mains. À l'ouverture du panneau les doigts sont déjà dessus pour courir, et l'enfoncement suivant prenait une carte que personne n'avait lue — le panneau ne durait que quelques images, si bien qu'on ne le voyait jamais. **Une partie jouée l'a montré en une minute ; aucune relecture ne l'avait vu**, et aucune vue de la planche ne le pouvait, puisqu'elle ne presse aucune touche.

Une flèche ne fait donc que déplacer une désignation, qu'un bord de couleur porte, et la prise demande **Espace ou Entrée** — les deux touches dont le sens est « je confirme », qui valent l'une pour l'autre partout où le jeu attend un accord, relance comprise. Rien d'irréversible ne part plus d'une touche tenue pour autre chose, et le choix redevient délibéré, ce que la pause existe pour permettre.

Les chiffres de dégâts jaillissent au-dessus de n'importe quoi — décor clair, flaque, carrelage. Ils portent donc un contour ou une ombre portée d'un pixel, sans quoi leur couleur, distincte pour les critiques, ne suffit pas à les détacher du fond. C'est la discipline des contours foncés du paragraphe précédent, appliquée à ce qui n'est pas un sprite.

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

Des modules découpés par nature, plus `ressources.py` qui les enchaîne et
contrôle ce qu'ils ont produit. **`primitives_iso.py`** est le noyau : tout part d'une fonction `volume(tx, ty, elevation, matiere)` qui calcule la face supérieure **en coordonnées de tuile**, ce qui garantit l'escalier régulier du 2:1 quelle que soit l'emprise, y compris fractionnaire — `ty=0.15` donne une cloison mince, une barrière ou un écran de cinéma. Il ne connaît aucune forme du jeu.

**`decor_iso.py`** génère les lieux : une soixantaine de formes réparties en commun, supermarché, parking, quartier, cinéma, station.

Fonctions de détail : `grain` (moucheture), `nervures` (tôle, lattes), `bandeau` (réglette d'étiquettes), `rivets`, `fenetres` (bande vitrée, restreignable à une portion pour un pare-brise), `creuser` (porte ou vitrine renfoncée), `eventrer` (arêtes mangées), `roues` (accrochées au bord inférieur de la silhouette, avec débord), et `contour`, le plus efficace de tous.

Composition : `poser` place un objet **en coordonnées de tuile**, pas en pixels — trois caisses alignées le sont dans le monde, pas seulement à l'image ; `aligner` compose des volumes partageant la même origine, ce qui donne les angles de mur. Un véhicule se déclare en une ligne via `vehicule(longueur, largeur, hauteur, caisse, cabine)` : une seule caisse, la cabine étant une zone teintée et non un bloc rapporté — deux volumes empilés laissent un joint visible et cassent la lecture.

`MATIERES` est le point d'entrée unique de la palette : trois teintes par matière, dessus et deux flancs.

**`figurines.py`** génère les créatures : six gabarits, neuf profils, variantes de teinte, une bande horizontale par cycle et par direction.

**`objets.py`** génère ce qui se ramasse ou se tire : caisse et ses cycles d'appui et de rupture, gemme, fiole, projectiles, armes lourdes en version posée au sol et en icône d'interface. Une caisse se casse et se ramasse — elle appartient au jeu, pas au lieu, même si elle est posée sur la grille comme un mur.

**`sons.py`** génère les bruitages par synthèse, sur le même principe de graine et de manifeste — le procédé est décrit plus bas, à la section du son.

**`interface.py`** ne dessine pas, il rastérise : il cuit la police tierce en planche de glyphes et déclare dans son manifeste la cellule, la ligne de base, la chaîne des glyphes et leur avance. C'est le seul générateur dont la source est un fichier reçu plutôt qu'une fonction — `assets/polices/` porte ce qu'on a téléchargé, `assets/interface/` ce qu'on en fabrique —, et c'est ce qui met le texte dans le même régime que le reste : une image qu'on régénère et qu'on compare, au lieu d'un rendu qui dépendrait de la version d'une bibliothèque.

### Les manifestes

Chaque lot produit un manifeste JSON, et c'est lui qui fait contrat entre les images et le moteur.

**Un manifeste est généré quand son contenu est dérivé, tenu à la main quand il est décidé.** Un générateur gagne sa place en calculant ce qui n'existe pas dans son entrée — des volumes composés, une enveloppe synthétisée, une fonte cuite en planche. Il n'en a aucune quand il transcrirait les mêmes chiffres depuis un script : les valeurs déménageraient du JSON vers le Python, on éditerait toujours un fichier, et l'on paierait une commande de plus.

Ce n'est donc pas une exception que d'écrire à la main la table d'armes et celle de la progression, c'est le même critère appliqué à des contenus d'une autre nature. Et il dit d'avance ce qui les ferait changer de camp : le jour où l'outil d'équilibrage mesure une cadence de récolte, les seuils de niveau cessent d'être décidés pour devenir un calcul. La table d'armes, elle, ne changera pas — ses valeurs tiennent au ressenti, et un outil mesure ce qu'une cadence produit sans décider ce qui est agréable.

Chaque manifeste porte un **en-tête** : `version_format`, et pour le décor la taille de tuile. Sans lui, aucune migration n'est possible le jour où un champ change de sens.

**Mais il n'y sert pas à ce qu'il sert dans un niveau**, et la nuance mérite d'être écrite avant que quelqu'un l'aligne dans un sens ou dans l'autre. Un niveau circule entre joueurs, donc son numéro dit à un binaire quoi faire d'un fichier qu'il n'a pas produit. Un manifeste, lui, est embarqué par `go:embed` : il voyage avec son lecteur et ne peut pas en être désynchronisé. Ce qu'il accorde n'est pas deux machines mais **deux chaînes d'outils** — des scripts Python qui écrivent, du Go qui lit — et c'est pour cela qu'il en porte un quand même. D'où deux réflexes à ne pas avoir : incrémenter par symétrie avec les niveaux quand un champ disparaît, ou retirer le champ en constatant qu'aucune migration ne l'attend.

Côté **décor** : taille, ancrage, élévation, catégorie, thème, et quatre champs qui commandent le moteur — `bloquant` et `cout_traversee`, dont le chargeur tire la grille de coûts, `emprise` en tuiles, sans laquelle une gondole de deux tuiles n'en bloquerait qu'une, et le drapeau qui signale ce qui dépasse 24 pixels, donc masque un personnage. Rien de tout cela ne se devine : un trottoir et un quai dépassent du sol et se marchent, une flaque est plate et se traverse, alors qu'un muret de même hauteur qu'un trottoir arrête tout.

`cout_traversee` est exigé sur ce qui se franchit et refusé sur ce qui bloque. Deux champs plutôt qu'un entier où une valeur réservée vaudrait l'infini : une sentinelle laisse `bloquant` et un coût fini coexister dans le même fichier, et quelqu'un finit par l'écrire. Contrôlé dans les deux sens, l'état absurde n'est pas exprimable — et un coût orphelin sur un mur, jamais lu, ne fait croire à aucun réglage.

Côté **personnages** : le rendu — cycles, cadences, bouclage, directions, point d'appui, gabarit, variantes — **et les valeurs de jeu**, dans le même fichier.

Un **rôle** décide d'abord, et il en a trois : le joueur porte une vie, un plafond de dégâts et la seule vitesse absolue du jeu ; un ennemi porte résistance, points, coût de pression, poids de séparation, plafond de simultanéité, dégâts de contact et nombre de gemmes ; une entité d'ambiance ne porte ni l'un ni l'autre.

Le nombre de gemmes est un **nombre et non une valeur**, ce que le chapitre 2 impose : une gemme rapporte la même chose du début à la fin, et c'est le seuil du niveau suivant qui monte. Un profil qui doit rapporter davantage en laisse donc plusieurs, et c'est la quantité au sol qui dit au joueur ce qu'il va gagner. Un booléen n'en aurait couvert que deux, le troisième cas se lisant dans son absence — le genre de défaut par défaut qu'on ne voit pas.

Un **comportement** ajoute ensuite ce qui n'a de sens que pour lui : tangentiel du flanqueur, portée de la Buse, dégâts de charge du Molosse, rayon d'explosion de la Baudruche. Déclarés avec le comportement et non dans une liste de champs facultatifs, ils se contrôlent dans les deux sens — une portée sur un Badaud ne serait jamais lue et laisserait croire qu'il tire.

Deux unités enfin, et jamais deux fois le même nom pour deux unités : `vitesse_tuiles_s` sur le seul joueur, `vitesse_relative` sur les autres. Et `rayon_tuiles`, pas un rayon en pixels : la simulation ne connaît que la tuile, et une distance mesurée à l'écran décrirait une ellipse dans le monde.

Côté **objets et armes**, la ligne de partage tient en une phrase : **le tireur porte les valeurs de son tir, le projectile ne porte que son apparence.** Cadence, portée, dégâts, nombre de projectiles et vitesse appartiennent à l'arme quand c'est le joueur qui tire, au profil quand c'est une créature — et le projectile garde sa taille, son ancrage, son emprise, ce qui se règle en dessinant.

Ce n'est pas une préférence de rangement. Les projectiles ont d'abord porté leurs dégâts et leur portée, si bien que la Buse avait une portée de six sur son profil et de sept sur son projectile, dans deux fichiers, sans que rien ne le signale — la seconde description avait divergé avant même d'être lue. Et la vitesse d'un projectile est le chiffre qui décide si un tir de Buse s'esquive : la laisser dans un manifeste généré signifiait que la seule vraie question d'équilibrage du Cracheur se réglait en régénérant six cents images.

Le critère qui range vaut au-delà de ce cas : **une valeur vit à côté de ce qu'elle alimente.** Ce n'est pas « rester dans un manifeste généré » qui décide — la génération n'est pas ce qui gêne —, mais qui édite le fichier et pourquoi : passer un soin de trente à vingt-cinq ne change rien à ce qui est dessiné, et redessiner la fiole ne change rien à ce qui est soigné. Deux publics, deux cadences.

C'est ce qui a fait descendre **l'expérience d'une gemme** dans le manifeste de progression : elle alimente les seuils de niveau, et c'est là qu'ils se règlent. Elle était restée sur l'objet, déclarée et jamais lue — le moteur comptait les gemmes en les supposant à un, ce qui donnait le même résultat et rendait les deux implémentations indistinguables. Un champ que personne ne lit ne se voit pas, et c'est ce qui rend ce rangement moins une préférence qu'un contrôle.

`outils/ressources.py` porte les deux moitiés du geste : il **refuse** les champs déménagés, pour qu'on ne les remette pas sur un objet par symétrie avec le manifeste des personnages, qui porte bien ses valeurs de jeu ; et il exige que tout renvoi `objet` d'un autre manifeste désigne un objet du catalogue, faute de quoi le lien par nom que le déménagement crée casserait en silence au premier renommage.

Côté **armes** : `assets/armes/manifeste.json`, tenu à la main et non généré — cadence, portée, dégâts, nombre de projectiles, puis la table des passifs et les recettes de fusion. C'est l'une des deux exceptions de `assets/`, et le chapitre 9 dit pourquoi.

Les mettre ailleurs aurait dupliqué la liste des profils à deux endroits. Un nouveau profil reste une ligne de table.

Côté **progression** : `assets/progression/manifeste.json`, tenu à la main lui aussi — le seuil du premier niveau, ce que chaque niveau ajoute au seuil du suivant, le plancher de quarante-cinq secondes, le rythme de l'aimant avec la distance minimale de son apparition et la vitesse de sa ruée, et tout ce qui concerne la gemme : ce qu'elle rapporte, la portée à laquelle elle se ramasse, le temps qu'elle reste au sol. Ces deux dernières forment le couple que ce document interdit de régler séparément, et c'est pour cela que la portée a quitté le profil du joueur — celui qui touche à l'une doit voir l'autre. Ces chiffres se règlent en rejouant, et c'est le même critère qui les sort d'un fichier généré.

Il ne vit pas auprès des armes, et la raison n'est pas qu'un manifeste de plus soit plus propre : **un seuil appartient à la partie**. Il vaut quelle que soit l'arme équipée, il continuerait d'exister si la table d'armes changeait entièrement, et il commande le rythme des choix plutôt que leur contenu. D'où des sections nommées plutôt que des champs à plat.

Une section `pression` y range **ce que le spawner tient de la partie, par opposition à ce que le lieu lui dit** : le rayon de l'anneau d'apparition et la borne du budget reporté. Le partage suit le même critère que le reste — qui édite le fichier, et pourquoi. Un auteur de lieu compose le rythme de ses vagues ; il ne décide ni de la distance à laquelle une créature se matérialise, ni de ce que devient un budget non dépensé. **Et le rayon ne peut pas se dériver de la fenêtre** : deux joueurs aux fenêtres différentes n'auraient alors pas les mêmes apparitions sur la même graine, ce que l'invariant du déterminisme interdit. Il est donc une donnée, dérivée du tampon fixe, et le manifeste dit lequel pour qu'on sache quoi rouvrir le jour où ce tampon change.

Le plafond d'effectif, lui, n'y est pas : c'est la capacité du bassin des ennemis, et deux chiffres pour une même limite finiraient par se contredire en silence — le plus bas masquerait l'autre, et un bassin qui refuse est muet là où le spawner qui s'arrête perd son budget sciemment.

Côté **sons** : durée, gain, bouclage, et une **catégorie de mixage** — le joueur doit pouvoir baisser les effets sans toucher à la musique, et l'interface doit rester audible quand tout le reste est baissé.

Côté **objets** : emprise, élévation, catégorie et masquage — les trois mêmes que le décor, et pour la même raison, un rideau de fer culmine à 46 pixels et masque un personnage —, ce qui bloque, ce qui détruit, ce qui est projeté, ce qui est entendu, et les valeurs de jeu qui restent — soin d'une fiole, charges d'une arme lourde. Un bloc `destruction` porte le mode — `contact` pour la caisse, où le délai est la mécanique elle-même, `interaction` pour les obstacles fragiles —, le nombre de touches, le nom de la ruine, la matière des éclats, les cycles d'appui et de rupture, et les clés de sons. Le moteur ne code donc rien en dur : un futur obstacle se déclare dans une table.

Un renvoi de son dit **s'il nomme un fichier ou une famille**. `son` désigne l'un, `famille_sons` une suite de degrés — `gemme_0` à `gemme_7` — que le moteur parcourt en avançant d'un cran à chaque déclenchement rapproché, et qu'il reprend au premier après un silence. Deux clés plutôt qu'une seule à interpréter : sans la distinction, le contrôle ne peut que comparer des préfixes, et accepte alors « gem » et « g » aussi bien que « gemme ».

Le contrôle vérifie la cohérence de ces renvois — une ruine qui n'existe pas, des éclats sans particules générées, un destructible sans nombre de touches, une ruine qui bloque encore, **une ruine qui n'est pas plus basse que son original**, un son introuvable au nom exact, **une famille de sons dont la suite de degrés a un trou**. Ce sont des défauts qui ne se manifestent qu'au moment de la destruction, c'est-à-dire le plus tard possible et souvent chez un joueur.

Deux de ces contrôles ont été écrits après coup, sur des défauts que la liste promettait de couvrir et ne couvrait pas : une vitrine dont la ruine culminait aussi haut que la devanture intacte, et un renvoi de son que la comparaison par préfixe validait par accident. Un contrôle qui passe par accident est pire qu'absent — on se croit couvert.

Sans ces fichiers, une bande de 320 pixels est indéchiffrable — 5 images de 64 ou 4 de 80 ? Avec eux, le code de rendu ne connaît que des profils et des cycles, jamais des noms de fichiers ni des nombres codés en dur. Remplacer plus tard le chien du pack par un sprinteur dessiné à la main avec 6 images se fait en changeant une ligne du manifeste.

### Le son

**Les bruitages sont générés**, par `outils/sons.py` : une enveloppe appliquée à un oscillateur, plus un peu de bruit. C'est le procédé de sfxr, et il couvre exactement le registre d'un survivor — tirs, impacts, ramassages, explosions. Seule la bibliothèque standard est employée, et chaque son a sa graine, donc les fichiers sont reproductibles au bit près — davantage que les images, dont seul le dessin l'est, la compression PNG dépendant du système.

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

### Le temps

**Soixante pas par seconde, fixes, jamais réglables.** Ebitengine appelle la mise à jour à cadence constante et rattrape un retard en l'appelant plusieurs fois d'affilée ; la simulation n'a donc jamais à connaître le temps écoulé. Le tick est l'unité de temps unique d'`internal/game`, qui **n'accepte aucun delta** : une fonction de mise à jour qui prendrait une durée en paramètre est un défaut, pas une souplesse.

Toute durée écrite dans un manifeste — les cadences d'animation, le tiers de seconde d'appui sur une caisse, le télégraphe du Molosse — est en millisecondes, et se convertit en ticks **une seule fois au chargement**, par arrondi au plus proche. Convertir à l'usage rouvrirait la question à chaque appel : à 60 Hz, 330 ms valent 19,8 pas, et la caisse céderait au 19e ou au 20e selon qui écrit le code.

**Une exception, et une seule : la frise d'un scénario de vagues.** Ses instants s'écrivent `m:ss` et la durée d'une pointe en secondes, parce que ce n'est pas une cadence de mécanisme mais un déroulé que son auteur relit comme une minuterie — `"2:10"` se place sur une courbe de quinze minutes, `130000` non. Elle passe par la même conversion que le reste et n'arrive dans la simulation qu'en ticks ; **ce qui change est la notation, pas la règle.** Et elle se lit strictement, au point de refuser ce qui se laisserait interpréter : `0:60` n'est pas une minute, `1:5` n'est pas cinq secondes, et deviner sur le chiffre qui commande tout le rythme d'un lieu serait le pire choix possible.

Cette exception est écrite ici parce qu'une première entorse devient le précédent qui autorise les suivantes. Elle vaut pour ce que lit un auteur de lieu, jamais pour ce que produit un générateur.

**Une durée sous le pas est refusée, jamais relevée à un tick.** La relever produirait un fichier qui ment : quelqu'un écrirait 8 ms en croyant obtenir du 125 Hz et obtiendrait du 60, sans que rien ne le dise. Le refus est tenable ici parce que ces fichiers ne sont saisis par personne — ils sortent des générateurs, donc une telle valeur est un défaut dans un script, pas une intention d'auteur. Le contrôle est aux deux bouts : le générateur refuse de l'écrire, le chargeur refuse de la lire. Le premier la montre à sa source, le second protège du fichier retouché à la main.

Un seul compteur de ticks porte la partie. Il n'avance ni pendant la pause de la montée de niveau, ni pendant le gel d'impact des gros coups — sans quoi le scénario de vagues, l'objectif de survie et le bonus de temps du chapitre 6 mesureraient trois choses différentes. Le temps réel n'entre jamais dans la simulation, ce qui se vérifie simplement : `internal/game` n'importe pas `time`.

### Les repères

**La tuile est le repère du monde, et le seul.** Le pixel n'existe que dans `internal/render` ; l'élévation est une troisième grandeur qui ne participe à aucun calcul de simulation. Une distance de jeu se mesure dans le plan du sol et jamais à l'écran, sinon un rayon exprimé en pixels décrit une ellipse et deux Badauds se touchent à une distance différente selon qu'ils sont alignés est-ouest ou nord-sud.

**Les positions sont en virgule fixe entière, une tuile valant 65536.** Pas des flottants : la spécification Go autorise une implémentation à fusionner une multiplication et une addition en une seule opération arrondie une fois, et arm64 le fait là où amd64 ne le fait pas. Deux binaires publiés divergeraient sur la même graine, ce qui viderait l'invariant du déterminisme de sa substance — et avec lui le classement par graine, la graine du jour et le partage d'un défi, que le chapitre 6 promet tous les trois.

Trois règles font tenir le procédé :

- **Un type fermé, jamais un alias.** `type Fixed int32` avec ses opérations en méthodes, pas `type Fixed = int32` sur lequel `a * b` compilerait sans rien dire. La remise à l'échelle après produit est la seule vraie source d'erreur du procédé ; le seul garde-fou est que le compilateur refuse la multiplication nue.
- **Les produits passent par `int64`** avant remise à l'échelle. Deux valeurs d'une tuile tiennent dans un `int32`, leur produit non.
- **L'arrondi est au plus proche et symétrique autour de zéro.** L'échelle étant une puissance de deux, la remise à l'échelle est un décalage arithmétique, qui arrondit vers l'infini négatif : sans correction, chaque opération biaise d'une demi-unité toujours dans le même sens, ce qui fait de l'ordre de quatre dixièmes de tuile de dérive sur une run de quinze minutes — le rejeu resterait déterministe, et faux. Ajouter la demi-échelle avant le décalage suffit à l'arrondi, mais pas à la symétrie : les demis exacts monteraient vers l'infini positif, et une vitesse tombant pile sur un demi ferait avancer différemment vers la gauche et vers la droite. Le signe se traite donc à part.
- **Un résultat hors plage sature, il ne déborde pas.** Aucun lieu n'approche les ±32768 tuiles, mais la division par une longueur minuscule est atteignable — c'est ce que produit la normalisation d'un vecteur presque nul. Un débordement silencieux ferait réapparaître l'entité de l'autre côté de la carte avec le signe opposé ; la saturation la colle au bord, ce qui se diagnostique. C'est un comportement de simulation, donc observable et identique partout, pas un détail d'implémentation.

**La frontière s'arrête à `internal/game`.** Le rendu convertit en pixels et calcule ce qu'il veut en flottants — interpolations, lissage de caméra, paraboles d'éclats — puisque rien de ce qu'il produit ne revient dans la simulation. Sans cette ligne, la verbosité de la virgule fixe contaminerait tout le projet pour un déterminisme dont le rendu n'a que faire.

Corollaire sur la normalisation : le champ de flux stocke des vecteurs **déjà normalisés**, calculés une fois par rafraîchissement et non par entité et par image. `math.Sqrt` reste utilisable ailleurs — c'est l'une des rares opérations dont l'IEEE-754 exige l'arrondi correct, donc elle est portable —, à condition d'arrondir son résultat au plus proche plutôt que de le tronquer : une troncature raccourcit toujours, et les diagonales deviendraient plus lentes que les axes.

**L'invariant se vérifie, il ne se surveille pas.** Un test joue une graine sur un nombre de ticks fixé et compare l'empreinte de l'état — positions, vies, générations et état de chaque flux, parcourus dans l'ordre des index du bassin — à un attendu versionné. Il tourne sur les trois cibles natives de l'intégration continue, dont deux arm64. L'empreinte porte sur l'état et non sur un résumé : un compte d'ennemis vivants passerait au vert alors que deux trajectoires ont divergé puis se sont recroisées, ce qui est précisément le cas qu'on cherche. Sa mise à jour passe par `-maj-attendus`, jamais automatiquement.

### L'aléatoire

Une graine par partie, et **quatre flux nommés** qui en dérivent : `vagues`, `positions`, `butin`, `cosmetique`. Nommés par leur usage et jamais par leur rang — un flux qu'on désigne par son numéro finit par changer de sens.

Un flux unique suffirait à rejouer une partie jouée, et casserait le test qui compte : une run simulée **sans rendu** ne tire pas les teintes de vêtement, et chaque tirage manquant décale tous les suivants. Les vagues d'une run simulée cesseraient de correspondre à celles de la même graine jouée à l'écran — c'est-à-dire que l'outil d'équilibrage mesurerait autre chose que le jeu.

Le tirage cosmétique a donc lieu **dans la simulation**, à l'apparition, comme les autres. Le sortir vers le rendu paraîtrait plus propre et serait pire : une partie rejouée depuis son journal ne retrouverait pas les mêmes vêtements, et un rejeu visuellement infidèle est plus trompeur qu'un rejeu qui diverge, puisqu'il a l'air juste.

Ce qui est interdit n'est pas le flux, c'est **qu'une décision lise ce qu'il alimente**. Et cela se vérifie plutôt que de se surveiller : deux runs sur la même graine, teintes forcées différentes, doivent rendre la même empreinte de simulation.

L'algorithme est **PCG**, celui de `math/rand/v2`. Il est spécifié, stable d'une version de Go à l'autre, et le changer invaliderait toutes les graines publiées — donc il se choisit maintenant. Le générateur global reste proscrit, l'invariant 4 le dit déjà.

### L'ordre de mise à jour

Il n'y a pas de bon ordre, il y a un ordre écrit une fois. Dans un tick :

1. les entrées ;
2. les apparitions ;
3. le champ de flux, si c'est son tick ;
4. la grille de densité ;
5. les intentions de déplacement ;
6. la projection sur la passabilité ;
7. les contacts et les dégâts ;
8. l'aimant : son apparition, sa ruée, sa prise ;
9. le ramassage, et ce qu'il fait monter ;
10. le tir — du joueur puis de la horde —, et le vol des projectiles avec ce qu'ils touchent ;
11. les suppressions.

Les apparitions avant la densité, et c'est ce qui commande leur place : le champ de flux ne dépend que du joueur et des obstacles, une créature apparue après son calcul n'y perd rien. La densité, elle, dépend des ennemis — deux créatures apparues au même endroit se superposeraient exactement le temps d'une image, et personne ne retrouverait jamais l'origine de ce scintillement.

**Le contact se constate après le déplacement et non avant**, sinon une créature qui vient de se coller ne blesserait qu'au tick suivant, et le joueur verrait la horde le traverser sans effet pendant une image.

**La ruée de l'aimant avance avant le ramassage.** Sans quoi une gemme arrivée sur le joueur attendrait le tick suivant pour être prise, et la convergence de deux cents gemmes — le moment de plaisir maximal du chapitre 2 — se terminerait par un temps mort d'une image, exactement là où l'on veut un coup.

**Le ramassage est rangé avec les contacts, dont il est un** : ce que le joueur touche en se déplaçant. Il vient après les dégâts parce qu'une gemme ramassée dans le tick où l'on meurt ne change rien, alors que l'inverse ferait dépendre la mort de ce qu'on a récolté.

**Les étapes 5 et 6 peuvent tenir en une seule passe, à une condition qui vaut mieux que la conclusion** : aucune intention ne lit l'état d'une autre entité. Elles sont énumérées séparément parce que ce sont deux décisions, et l'équivalence tient tant qu'une intention ne lit que le champ et la densité, tous deux figés avant la passe. Le jour où un ennemi devra éviter celui qui le précède, elle tombe, et il faut les deux passes — donc une tranche d'intentions à préallouer.

La suppression par échange remonte la dernière entité active à l'index libéré. Cet index **n'est pas réexaminé par la passe de mise à jour en cours** : l'entité remontée y attend le tick suivant. Sans cette règle, elle serait mise à jour deux fois ou zéro selon le sens du parcours, et le déterminisme dépendrait d'un détail d'écriture.

La passe qui *retire* les morts, elle, réexamine bien la place libérée, et elle le doit : elle ne fait que filtrer, sans rien avancer, et la sauter y laisserait un cadavre un tick de plus. C'est la même distinction dans les deux sens — ce qui avance ne repasse pas, ce qui trie repasse.

### La mort est un état, pas un événement

Une résistance tombée à zéro **est** la mort : pas de drapeau à côté, rien à synchroniser, et savoir si une entité est une cible valide reste une lecture.

Ce qui se déclenche une fois est la **transition** — l'endroit qui applique les dégâts constate que la résistance était positive et ne l'est plus. Trois conséquences en partent, au même point : le butin, les points, et l'émission du cadavre. Les rattacher à l'état plutôt qu'à la transition les ferait repartir à chaque passage, et une Baudruche exploserait autant de fois qu'un projectile la traverse.

L'entité morte reste en place jusqu'à la fin du tick, pour que les index tiennent, mais **cesse d'être une cible** : un projectile traité plus tard dans la même passe l'ignore et va chercher derrière. C'est ce qui empêche deux projectiles de tuer le même ennemi sans rien devoir à l'ordre des étapes.

**Un cadavre n'est pas un ennemi mort, c'est une autre nature de chose** — une position, un cycle, une durée, et rien d'autre. Il ne pense pas, ne bloque pas, ne compte dans aucune densité et n'est jamais visé. Il a donc son bassin, comme les particules et pour la même raison : un type qui signifierait deux choses selon un drapeau obligerait chaque boucle écrite ensuite à demander « est-il mort ? », et l'oubli se paierait dans une boucle qui n'existe pas encore.

Trois règles le tiennent :

- **Il partage la clé de tri en profondeur des autres entités, et passe sous tout ce qui est vivant** à profondeur égale. Sans cela, un tapis de cadavres finit par masquer la horde en fin de run — ce que le chapitre 2 interdit.
- **Son bassin est borné comme les autres.** Plein, le plus ancien cède sa place plutôt que d'allouer. À vingt morts par seconde en pic, c'est le cas ordinaire et non une erreur : un cadavre disparaît un peu tôt au milieu d'une mêlée, ce qui ne se voit pas, et le budget d'allocation ne bouge pas d'un pic à l'autre.
- **Il porte la teinte de variante de l'ennemi dont il vient**, sinon un Badaud bleu laisse un cadavre vert et cela se remarque immédiatement.

Le bassin de cadavres est donc entièrement cosmétique : il n'entre pas dans l'empreinte d'état, et ne consommant aucun tirage, une run simulée sans rendu peut ne pas l'alimenter.

### La structure des entités

Tranchée au premier jalon, et voici ce qu'elle est devenue :

```go
type Enemy struct {
    Profile int   // index dans []EnemyProfile, jamais un pointeur
    X, Y    Fixed // en tuiles, virgule fixe — voir « Les repères »
    Hits    int   // la mort est cet état, pas un événement
}

type Pool[T any] struct {
    entities []T // capacité fixe, jamais réallouée
    active   int
}
```

**Les positions sont en `Fixed` et non en flottants**, ce que « Les repères » exige quelques paragraphes plus haut : un bloc de code qui montrerait des `float32` inviterait à casser le déterminisme sans que rien ne le signale.

Ce que la struct ne porte pas est aussi décidé. Pas de vitesse stockée : une créature lit le champ de flux sous ses pieds à chaque pas. Pas de génération : elle appartient au bassin, pour que rien n'incite à la lire hors du `Handle` — voir plus bas. Les champs d'animation, `Cycle` et `Frame`, viendront avec les sprites.

Le bassin est **générique**, et c'est ce qui met le mécanisme en facteur plutôt que de le recopier pour chaque sorte d'entité. Autant de copies seraient autant d'endroits où tenir la règle, et une copie qui la manquerait ne ferait échouer aucun test : elle ferait qu'une référence périmée désigne une entité vivante.

Un `[]Enemy` de structures pleines plutôt qu'un `[]*Enemy`. À ce volume, l'argument du cache est secondaire — 300 structures tiennent en L2 et le temps d'image est dominé par les appels de dessin. Le vrai motif est **la pression sur le ramasse-miettes** : 300 objets alloués à chaque vague produisent les micro-saccades qui se voient.

Trois principes, plus importants que le choix structures/pointeurs :

- **Préallouer et ne jamais réallouer.** `make([]Enemy, 0, 512)` au démarrage ; suppression par échange avec le dernier actif puis décrément. Pas d'`append`, pas de trou.
- **Ne pas faire sortir de pointeur du bassin.** Après un échange, un `*Enemy` conservé ailleurs désigne une autre entité. Pour référencer un ennemi qui vit plusieurs images — une cible verrouillée —, utiliser **un identifiant stable et sa génération, tous deux tenus par le bassin** : au recyclage la génération est incrémentée, et le détenteur d'une référence périmée le voit. L'identifiant n'est pas la place : l'échange ramène la dernière entité dans le trou, si bien qu'une référence indexée par la place se briserait parce qu'une *autre* entité est morte.
- **Séparer le chaud du froid.** Vitesse, PV max, poids de séparation, tangentiel sont partagés par tous les ennemis d'un type : ils vivent dans un `[]EnemyProfile` et l'entité n'en garde que l'index. C'est ce qui garde la structure petite, seul levier qui compte réellement au parcours.

En itérant, prendre l'adresse plutôt que la copie :

```go
for i := range p.enemies[:p.active] {
    e := &p.enemies[i]
    ...
}
```

Deux pièges. L'échange à la suppression **casse l'ordre**, donc le tri en profondeur travaille sur une slice d'indices réutilisée avec `indices = indices[:0]`, jamais sur le bassin lui-même. Et le passage en tableaux séparés par champ n'est à envisager que si le profileur le réclame : pénible à écrire, sans gain mesurable à ce volume.

Le même modèle sert pour toutes les autres sortes d'entités.

### Les écrans hors de l'action

Le seul écran décrit jusqu'ici est celui de la mort. Il en faut trois autres — pause, réglages, sortie —, et leur **structure** se décide maintenant parce qu'elle touche la boucle ; leur contenu attend d'avoir quelque chose à régler.

- **Ce qu'une pause fige.** Le compteur de ticks s'arrête, comme pour la montée de niveau : c'est la même règle, et il n'y en a qu'une. Rien ne se sauvegarde à ce moment.
- **Quand on écrit sur le disque.** À la fin d'une run et à la sortie, jamais pendant. Une écriture au milieu d'une vague est la saccade que tout le reste du document cherche à éviter.
- **Ce qui persiste.** Déblocages, meilleurs scores par lieu et par graine, réglages, compteur de parties. Rien d'autre : quitter en pleine partie perd la run en cours, et c'est assumé.

Le contenu de l'écran de réglages dépend de ce qu'il y aura à régler, et la moitié n'existe pas encore — les catégories de mixage, oui ; la sensibilité de quoi, la difficulté de quoi, pas encore. Il a donc son étape dans la feuille de route plutôt qu'une place ici : une tâche datée n'est pas une dette, une tâche absente en est une.

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

Reste, dans l'ordre qu'établit [`../ROADMAP.md`](../ROADMAP.md) — **et cette numérotation-là est la seule**. En tenir une seconde ici avait produit deux « jalons décisifs » qui ne désignaient pas la même chose.

Deux jalons, et ils ne répondent pas à la même question.

**Le jalon éliminatoire, à l'étape 3** — la boucle mort → relance, une arme, un profil, une courbe de pression. Il tranche : le déplacement et le tir sont-ils agréables ? Si la réponse est non, rien de ce qui suit ne les rendra bons et on ne continue pas. C'est le seul jalon qui peut tuer le projet, et il arrive tôt pour cette raison. Son critère, écrit au chapitre 2 : si la bascule de puissance n'est pas ressentie avant la minute 9, la courbe est trop lente.

**Franchi le 3 septembre 2026** : bascule à 6:13, attribuée au dernier palier de cadence — l'axe était épuisé à l'instant où elle a été relevée. Le trajet est passé du contournement à la traversée, le déplacement s'est lu sans blocage incompris, et la visée automatique n'a pas été subie.

La même partie n'a pas tué en seize minutes, ce qui ne dit rien de la courbe : elle s'est jouée contre un seul profil, comme cette étape l'annonce, et le plus faible des sept. L'équilibrage de fin de run se juge avec la horde complète, à l'étape 4.

**Le jalon décisif, à l'étape 8** — l'enchaînement de salles, le score, le choix de la porte. Il tranche : a-t-on envie de relancer ? Un oui à l'étape 3 ne le prouvait pas ; seul un enchaînement complet le dit. Celui-là ne peut que valider.

Cinq étapes séparent les deux, ce qui est long sans retour. D'où la sonde de l'étape 4 : une porte et une caisse, assez pour sentir la tension « rester ou partir » bien avant que tout soit écrit.

Note de prudence : survivor, roguelite, exploration, ressources et éditeur avec campagne, empilés, font un projet de plusieurs années à une personne. L'ordre est conçu pour que la question « est-ce agréable ? » soit tranchée en quelques semaines, et « est-ce qu'on y revient ? » en quelques mois — pas après trois ans.

## 17. Ce qui reste à trancher

- **La rotation des pièces** : quatre variantes de mur dans le tileset, ou aucune rotation et plus de pièces à dessiner. À décider avant de dessiner la moindre tuile.
- **La taille de la maille des pièces** : 16×16 tuiles est la base proposée, elle conditionne tout le travail d'édition.
- **La palette définitive** : le plafond de couleurs de `MATIERES`, et les teintes réservées — celle du personnage joueur et celle des projectiles ennemis, qui ne doivent apparaître nulle part ailleurs.
- **La portée du tir de base**, qui remplace l'angle du cône comme réglage décisif du kiting : trop courte, il faut faire face pour toucher ; trop longue, la horde meurt avant d'être une menace.
- **Le plafond de dégâts par seconde** et le rapport entre contact ordinaire et charge du Molosse : deux chiffres qui décident si l'encerclement est tendu ou injuste.
- **La vitesse du joueur rapportée à celle des profils** : à 60 % de sa vitesse un Badaud ne rattrape jamais, à 90 % la fuite ne suffit plus. Tout le kiting tient dans ce rapport, et il se règle en jouant. Le point de départ est cinq tuiles par seconde pour le joueur — un peu plus de trois secondes pour traverser l'écran — ce qui donne au Molosse en charge un gain d'une tuile trois quarts par seconde, franc plutôt que rapide.
- **La portée de ramassage des gemmes et la durée de vie d'une gemme** : deux valeurs, un seul réglage — le chapitre 2 dit pourquoi elles ne se décident pas séparément.
- **Le pilote de persistance** : SQLite si et seulement si il se compile pour `js/wasm`, sinon un fichier JSON.
- **La courbe de résistance** au fil de la run, et le **temps de référence** de chaque lieu : c'est lui qui fixe le poids réel du bonus de temps face aux points d'ennemis.
- **La bibliothèque d'interface pour l'éditeur** : `ebitenui` ou tout dessiner à la main. À évaluer avant le chantier de l'éditeur, pas pendant.

Tranché en cours de route : aucun asset importable dans le contenu utilisateur, mode tuiles différé après la première version de l'éditeur, et l'aimant gardé — avec son emplacement propre et l'effacement des gemmes pour contre-force.

Tranché avant d'écrire la première ligne d'`internal/game`, parce que ce sont les décisions que reprendre plus tard toucherait tout le code de jeu : le pas de simulation et la conversion des durées, la tuile en virgule fixe comme repère unique, les quatre flux aléatoires et leur algorithme, la visée omnidirectionnelle sans cône, la passabilité par coût plutôt que par booléen, le Vigile comme seul corps que le joueur ne traverse pas, le domicile de la table d'armes, et l'ordre de mise à jour dans une image.

Tranché par le générateur : la catégorie n'est pas saisie, elle est **dérivée de l'élévation** — `PLAFOND_OBSTACLE_BAS` vaut 24, au-delà la forme est `haut` et se déclare masquante. Un obstacle bas trop haut est donc impossible à produire, plutôt que refusé après coup. Le distributeur de billets, l'abribus et les véhicules sont `haut` au même titre que les murs : il n'y a rien à baisser.
