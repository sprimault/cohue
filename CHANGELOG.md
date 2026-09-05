# Journal des versions

Le format suit [Keep a Changelog](https://keepachangelog.com/fr/1.1.0/).

SemVer avec la clause du zéro : **en `0.x`, rien n'est imposé**. Le mineur
marque une étape de la feuille de route, pas une rupture d'API — tout le reste
s'accumule en correctif, correctifs, fonctionnalités et ruptures confondus.

Trois numéros à ne pas confondre :

| Numéro | Où | Ce qu'il suit |
|---|---|---|
| version du dépôt | tag git | le binaire |
| `version_format` | chaque niveau et chaque pièce | le format de fichier |
| `empreinte_jeu_pieces` | chaque niveau | l'état réel du jeu de pièces |

Les deux derniers ne suivent pas SemVer. `version_format` est un entier :
ajouter un champ optionnel ne l'incrémente pas, tout le reste l'incrémente, et
un incrément oblige à écrire la migration des niveaux existants.
`empreinte_jeu_pieces` n'est pas une version mais une somme de contrôle. Une
version peut sortir sans qu'ils bougent ; ils ne bougent jamais sans version.

Un titre de section s'écrit `## [version] — date — titre`. La publication en
tire le nom et les notes de la version : ce qui est relu ici est ce qui sera lu
sur la page des versions, et il n'y a rien à recopier ensuite. Le titre est
facultatif ; sans lui, la version se nomme par son tag.

**Une section absente arrête la publication.** Une version sans notes ne dit ni
ce qui change, ni ce qu'un auteur de niveau doit reprendre.

La section en cours s'écrit `## [Non publié]`, sans date ni titre, et chaque lot
y ajoute ce qu'il change : écrite au fil de l'eau, elle est relue en même temps
que le code qu'elle décrit, alors qu'une section rédigée le jour du tag résume
de mémoire un mois de travail. **Elle prend son numéro et sa date au moment du
tag**, faute de quoi la publication cherche une section qui n'existe pas et
s'arrête.

Chaque section est **bilingue, français d'abord, séparé par `***`** — les notes
de version sont ce que lit un auteur de niveau étranger avant de savoir s'il
doit reprendre son travail. Ce préambule reste en français : il n'est jamais
publié, et explique les conventions du dépôt à qui y contribue.

## [Non publié]

## [0.4.0] — 2026-09-05 — Les profils d'ennemis

### Ajouté

- **Chaque créature a sa teinte.** Le Quidam reste rouge et les six autres se
  distinguent enfin — l'Arpenteur violet, le Molosse orange, le Vigile bleu, la
  Buse olive, la Baudruche magenta, le Secouriste vert. Toutes restent sombres :
  c'est la valeur qui garde le personnage visible dans une foule, la teinte ne
  sert qu'à séparer les rôles. La porte fermée, elle, tranche maintenant sur le
  mur qui l'entoure.
- **Les caisses se cassent et laissent des gemmes.** Un lieu déclare les siennes
  dans un champ `caisses` facultatif, chacune à sa case ; le joueur les casse en
  arrivant dessus, jamais son arme, qui ne les vise pas. Ce qu'elles laissent est
  un réglage de partie et non de lieu, comme la valeur d'une gemme. La place de
  démonstration en pose huit autour du carrefour.
- **Un lieu peut se gagner et se quitter.** Il déclare sa sortie dans un champ
  `sortie` facultatif : la case de sa porte, et le nombre de créatures à abattre
  pour l'ouvrir. La porte se touche une fois gagnée, elle ne se traverse jamais,
  et les figurants n'entrent pas dans le compte. La place de démonstration
  demande cent créatures.
- **Le lieu de démonstration convoque de nouveau les sept créatures.** Une entre
  par palier, du plus spatial au plus subtil : l'Arpenteur à une minute, le
  Molosse et le Vigile à quatre, la Buse à six, la Baudruche à huit, le
  Secouriste à dix. La résistance monte à partir de la sixième minute, sans quoi
  la fin de partie n'est qu'un tapis roulant.
- **Un profil qu'une phase ne peut pas payer est refusé au chargement.** Le
  budget d'une phase s'accumule jusqu'à un plafond ; un profil qui coûte
  davantage était écrit dans le fichier et n'apparaissait jamais, sans un mot. Le
  refus nomme le prix, le plafond et la pression qui le produit, pour qu'on
  puisse choisir lequel des trois changer.
- **Une phase peut durcir ce qu'elle fait apparaître.** Un champ `resistance`
  facultatif multiplie les touches des créatures nées sous elle — 1,3 fait passer
  un Quidam de trois à quatre. La résistance est figée à l'apparition : une
  créature demande le même nombre de coups du premier au dernier.
- **Les Passants peuplent un lieu.** Un lieu déclare son ambiance dans un champ
  `ambiance` facultatif, chaque figurant à la case où il commence ; ils vont et
  viennent sans être visés, sans gêner la horde et sans compter nulle part. Une
  position dans un mur ou hors du lieu est refusée au chargement. La place de
  démonstration en pose douze autour du carrefour central.
- **Le Secouriste soigne la horde.** Il rend des touches à la créature la plus
  entamée autour de lui, une à la fois, et jamais à lui-même. Deux éclairs le
  disent : l'un sur la soignée, l'autre sur lui — c'est ce dernier qui indique
  qui aller chercher, puisque l'arme vise toujours le plus proche.
- **Le corps du Vigile arrête le joueur.** C'est la seule créature qu'on ne
  traverse pas : elle bouche un goulot au lieu d'être un ennemi lent de plus.
  Elle ne traverse pas le joueur non plus, et son corps cesse de bloquer dès
  qu'elle tombe.
- **La Baudruche explose en mourant.** Son emprise se marque au sol entière dès
  le premier instant — ce qui est marqué est ce qui sera touché —, et c'est son
  intensité qui dit le temps restant pour en sortir ; passé ce délai, le souffle
  retire ses points hors du plafond. Il n'atteint que le joueur : une
  déflagration qui emporterait la horde récompenserait le nettoyage à l'aveugle
  qu'elle punit.
- **La Buse tire.** Elle s'arrête dès que le joueur entre dans sa portée, par un
  chemin et non à vol d'oiseau : un mur entre les deux la fait contourner au lieu
  de se figer. Elle envoie alors un projectile par intervalles, visant où il est
  plutôt que où il ira. Le tir s'esquive, un mur l'arrête, et ses dégâts
  s'ajoutent hors du plafond.
- **Le Molosse charge.** Il s'immobilise le temps d'annoncer, fonce en ligne
  droite sans plus corriger, et son choc s'ajoute hors du plafond de dégâts. Un
  mur ou un pilier interposé l'arrête : il charge sans regarder si la voie est
  libre, et toute fin de course lui coûte un temps mort.
- **Le Molosse arrive par meutes de trois.** La meute apparaît d'un seul côté et
  d'un seul coup : elle se paie trois fois le prix d'un Molosse, et jamais
  un chien n'arrive seul parce qu'il ne restait de la place que pour lui.

### Modifié

- **Le Badaud s'appelle désormais le Quidam.** Les deux noms de la foule — celui
  de la horde et celui des figurants — désignaient à peu près la même personne
  en français, et on les confondait. Le Passant garde le sien, qui dit ce qu'il
  est ; la masse hostile prend celui qui dit ce qu'elle a de terrible, à savoir
  n'importe qui. La clé `marcheur` ne change pas : les lieux déjà écrits restent
  valides.

### Corrigé

- **Une phase fait apparaître tous les profils qu'elle autorise.** Le budget se
  vidait au prix le moins cher et n'atteignait jamais les prix élevés : une phase
  qui en ouvrait sept n'en montrait qu'un. Le spawner met désormais de côté pour
  un profil tiré au sort, d'autant plus souvent qu'il est bon marché — la masse
  arrive vite, les exceptions se paient.
- **Un profil de vague inconnu dit lequel écrire.** Le refus énumère désormais
  les clés attendues avec le nom de chaque créature — « flanqueur » (Arpenteur)
  —, là où il se contentait de déclarer le nom inconnu. Un lieu s'écrit avec la
  clé, quand tout ce qui se lit ailleurs porte le nom.

***

### Added

- **Each creature has its own hue.** The Quidam stays red and the other six are
  finally told apart — the Arpenteur purple, the Molosse orange, the Vigile blue,
  the Buse olive, the Baudruche magenta, the Secouriste green. All stay dark:
  value is what keeps the character visible in a crowd, hue only separates roles.
  The closed door now stands out against the wall around it.
- **Crates break and leave gems.** A place declares its own in an optional
  `caisses` field, each at its tile; the player breaks them by walking into
  them, never their weapon, which does not aim at them. What they leave is a
  game setting rather than a place setting, like a gem's value. The
  demonstration place puts eight around the crossroads.
- **A place can now be won and left.** It declares its exit in an optional
  `sortie` field: the door's tile, and how many creatures must be killed to open
  it. Once earned the door is touched, never walked through, and bystanders do
  not count towards it. The demonstration place asks for a hundred creatures.
- **The demonstration place summons all seven creatures again.** One enters per
  step, from the most spatial to the most subtle: the Arpenteur at one minute,
  the Molosse and the Vigile at four, the Buse at six, the Baudruche at eight,
  the Secouriste at ten. Toughness rises from the sixth minute on, or the endgame
  is just a conveyor belt.
- **A profile a phase cannot afford is rejected at load time.** A phase's budget
  accumulates up to a ceiling; a profile costing more was written in the file and
  never showed up, without a word. The rejection names the price, the ceiling and
  the pressure that produces it, so you can choose which of the three to change.
- **A phase can toughen what it spawns.** An optional `resistance` field
  multiplies the hits of creatures born under it — 1.3 takes a Quidam from three
  to four. Toughness is fixed at spawn: a creature takes the same number of hits
  from first to last.
- **Passants populate a place.** A place declares its ambience in an optional
  `ambiance` field, each extra at the cell where it starts; they wander without
  being targeted, without hindering the horde and without counting anywhere. A
  position inside a wall or outside the place is rejected at load time. The
  demonstration place puts twelve of them around the central crossroads.
- **The Secouriste heals the horde.** It restores hits to the most damaged
  creature around it, one at a time, and never to itself. Two flashes say so: one
  on the healed creature, one on it — the latter is what tells you who to go
  after, since the weapon always aims at the nearest.
- **The Vigile's body stops the player.** It is the only creature you cannot walk
  through: it plugs a chokepoint instead of being one more slow enemy. It does
  not walk through the player either, and its body stops blocking the moment it
  falls.
- **The Baudruche explodes on death.** Its footprint is marked on the ground in
  full from the first instant — what is marked is what will be hit — and its
  intensity tells how long is left to step out; past that delay the blast takes
  its points outside the cap. It only reaches the player: a blast that
  swept the horde would reward the blind clearing it exists to punish.
- **The Buse shoots.** It stops as soon as the player is within range along a
  path rather than as the crow flies: a wall between them makes it go around
  instead of freezing. It then sends a projectile at intervals, aiming where they
  are rather than where they are headed. The shot can be dodged, a wall stops it,
  and its damage lands outside the cap.
- **The Molosse charges.** It stands still while it winds up, runs in a straight
  line without correcting, and its impact lands outside the damage cap. A wall or
  a pillar in the way stops it: it charges without checking whether the path is
  clear, and any end of a run costs it a pause.
- **The Molosse arrives in packs of three.** The pack shows up from one side and
  all at once: it costs three times the price of a single Molosse, and no hound
  ever arrives alone because there was only room for one.

### Changed

- **The Badaud is now the Quidam.** The two names for the crowd — the horde's
  and the bystanders' — meant roughly the same person in French, and they were
  being confused. The Passant keeps its own, which says what it is; the hostile
  mass takes the one that says what is dreadful about it, namely anyone at all.
  The `marcheur` key does not change: places already written stay valid.

### Fixed

- **A phase now spawns every profile it allows.** The budget drained at the
  cheapest price and never reached the high ones: a phase opening seven profiles
  showed only one. The spawner now saves up for a randomly drawn profile, drawn
  the more often the cheaper it is — the mass arrives fast, the exceptions are
  paid for.
- **An unknown wave profile now says which one to write.** The rejection lists
  the expected keys along with each creature's name — "flanqueur" (Arpenteur) —
  where it merely declared the name unknown. A place is written with the key,
  while everything read elsewhere carries the name.

## [0.3.2] — 2026-09-03 — L'alerte de vie basse

Une version de correction. Ce qui se voit en jouant tient en une chose : sous
trente points de vie, les bords de l'écran rougissent et le restent jusqu'au
soin.

### Ajouté

- **Les bords de l'écran rougissent quand la vie est basse.** Sous trente points
  — ce que rend une fiole —, une vignette cerne l'écran et y reste jusqu'au
  soin. C'est la seule alerte visible sans quitter son personnage des yeux, là où
  la jauge de vie demande de regarder en haut à gauche.

### Corrigé

- **Un manquement du manifeste d'interface nomme la clé fautive.** Les refus
  désignaient leur groupe et non la clé — « planche non nommée », un mot que le
  fichier ne porte nulle part —, si bien qu'on savait qu'une chose manquait sans
  savoir laquelle corriger.

***

A maintenance release. Only one thing shows while playing: below thirty points
of health, the screen edges redden and stay that way until you heal.

### Added

- **The screen edges redden when health runs low.** Below thirty points — what a
  flask restores — a vignette rings the screen and stays until you heal. It is
  the only warning you can see without taking your eyes off your character,
  where the health gauge asks you to look top left.

### Fixed

- **A shortcoming in the interface manifest now names the offending key.**
  Rejections pointed at their group rather than the key — "unnamed sheet", a
  word the file carries nowhere — so you knew something was missing without
  knowing what to fix.

## [0.3.1] — 2026-09-03 — La boucle mort → relance

**On y meurt, et on relance.** La horde qui convergeait sans effet blesse
désormais au contact : collé sans se dégager, le joueur tombe en cinq secondes.
L'écran de mort laisse voir ce qui a eu raison de lui, et une touche remonte une
partie neuve sur le même lieu.

Les créatures abattues laissent des gemmes, qui font monter de niveau — et une
montée met le jeu en pause pour offrir **trois cartes**, dont on prend une.

### Ajouté

- **La vie du joueur et les dégâts au contact.** Une créature collée blesse en
  continu plutôt qu'en un coup, et le total encaissé par seconde est plafonné
  quel que soit le nombre d'ennemis au contact : l'encerclement reste très
  dangereux, mais il laisse une fenêtre pour se dégager plutôt que de tuer d'un
  bloc.
- **La mort fige la scène, une touche relance.** Le décor et la horde restent à
  l'écran, assombris, et Espace remonte la partie sans rien redemander — même
  lieu, aucune sélection à refaire.
- **Les gemmes.** Une créature laisse, à l'endroit exact où elle meurt, le
  nombre de gemmes que son profil déclare ; le joueur les ramasse en passant à
  courte portée.
- **L'expérience et la montée de niveau.** Les gemmes ramassées font monter d'un
  niveau, à un seuil qui croît d'un niveau au suivant : c'est le seuil qui monte,
  une gemme vaut la même chose du début à la fin de la partie. Un niveau est de
  plus donné toutes les quarante-cinq secondes sans rien ramasser, et il ne
  retire rien de ce qui est déjà acquis.
- **Le bandeau de la partie.** La vie, l'expérience et le temps écoulé
  s'affichent en haut de l'écran, du premier tick à la mort.
- **`assets/progression/manifeste.json`.** Les seuils de niveau, le plancher de
  temps et ce qu'une gemme rapporte se règlent dans ce fichier, tenu à la main
  comme la table d'armes.

- **L'aimant entre au catalogue** : un fer à cheval de cuivre dans
  `assets/objets/`, et sa montée sonore dans `assets/sons/`. Le son est une
  glissade d'une octave qui finit là où commence la gamme du ramassage ordinaire,
  pour qu'on reconnaisse la même chose en grand. Rien ne joue encore de son.
- **L'aimant.** Un objet apparaît toutes les trente secondes, toujours à
  bonne distance : on le ramasse et on le garde, il ne se déclenche pas au
  contact. La touche 1, celle du pavé numérique comme celle de la rangée du
  haut, dépense la charge, et **toutes les gemmes au sol convergent d'un
  coup** — elles cessent alors de s'effacer, l'aimant étant le recours contre
  l'effacement et non sa victime. On n'en tient qu'un, et aucun nouveau
  n'apparaît tant que la charge n'est pas dépensée.
- **La fenêtre a son icône**, dessinée par `outils/interface.py` en trois tailles
  parmi lesquelles le système choisit : une figure claire cernée de six autres,
  ce que le jeu raconte en seize pixels. Sa palette est la sienne et ne suit pas
  celle du rendu, dont les aplats sont provisoires.
- **Un repère se pose en jouant.** Espace horodate l'instant dans le journal de
  la partie, avec le niveau atteint et les paliers pris sur chaque axe. Le
  bandeau le confirme deux secondes sous le minuteur : marquer ce qu'on vient de
  ressentir ne demande pas de quitter la horde des yeux, et ce qui l'a produit
  est écrit plutôt que retenu.
- **Les gemmes s'effacent.** Une gemme laissée au sol s'éteint progressivement
  et disparaît au bout de six secondes. L'extinction est continue et non
  clignotante : l'âge d'une gemme se lit, donc ce qu'un tas vaut encore
  s'estime. C'est ce qui obligera à revenir là où l'on vient de tuer, là où la
  horde converge.
- **La courbe de pression.** La horde n'est plus posée au montage : une partie
  commence sur une salle vide, et les créatures s'achètent dans un budget de
  pression par seconde. Elles apparaissent hors du champ de vision, à dix-neuf
  tuiles du joueur, et jamais dans un mur — plutôt aucune créature qu'une
  créature surgie de nulle part. **La première montée de niveau s'atteint
  désormais en jouant.**
- **Un lieu porte son scénario de vagues**, sous `vagues` dans `lieu.json` : des
  phases datées sur une frise `m:ss`, une pression par seconde, les profils
  qu'elles autorisent, et une pointe facultative qui multiplie le budget pendant
  quelques secondes. Une phase vaut jusqu'à ce que la suivante la remplace. Un
  lieu sans scénario est un lieu sans horde, ce qui est admis : toutes les salles
  ne sont pas des arènes.
- **`assets/progression/manifeste.json` gagne une section `pression`** : le rayon
  d'apparition et la borne du budget reporté. Ce que le lieu décide est le
  rythme ; ce que la partie décide est d'où sortent les créatures et ce que
  devient un budget qu'on n'a pas pu dépenser.
- **Une créature touchée s'éclaircit un instant.** Elle encaisse plusieurs
  touches, et rien ne distinguait jusqu'ici un tir qui rate d'un tir qui entame :
  on déduisait ses ratés au lieu de les lire.
- **Les trois cartes de la montée de niveau.** La horde se fige et trois cartes
  s'affichent en bas de l'écran : les flèches gauche et droite déplacent un
  liseré ambre, Entrée ou Espace prend la carte désignée. Deux axes améliorent
  l'arme — cadence et portée, six paliers chacun — et une troisième carte rend
  de la vie quand rien d'autre n'est disponible. La table se règle dans
  `assets/armes/manifeste.json`.
- **Les chiffres appartiennent aux emplacements**, qui les gardent toute la
  partie : une carte mal choisie se rattrape au niveau suivant, un aimant
  déclenché à vide est perdu.

### Modifié

- **La portée de ramassage a quitté le profil du joueur** pour
  `assets/progression/`, où elle rejoint la durée de vie d'une gemme. Les deux
  punissent le non-ramassage et ne se règlent pas séparément ; les laisser dans
  deux fichiers, dont l'un est généré, aurait fait payer une régénération de six
  cents images pour ajuster l'un des deux termes.
- **Les lieux vivent désormais dans une campagne**, `assets/campagnes/<nom>/`,
  et c'est elle qu'on compose et qu'on partagera : un dossier de salles dont le
  descripteur `campagne.json` dit par laquelle on commence. Le dossier existe
  pour cloisonner les noms, comme celui d'un lieu le fait pour ses pièces — sans
  lui, deux auteurs appellent tous les deux une salle « parking ».
- **Les fichiers d'un lieu portent des noms fixes.** Le jeu de pièces s'appelle
  `jeu.json` et les pièces vivent dans `pieces/`, là où deux noms libres au même
  niveau ne disaient pas lequel était une palette et lequel un plan. Un lieu
  existant se reprend en renommant son jeu de pièces et en déplaçant ses pièces.
- **Ce qu'une gemme rapporte a quitté `assets/objets/manifeste.json`** pour le
  manifeste de progression : une valeur vit à côté de ce qu'elle alimente, et
  celle-ci alimente les seuils. Le manifeste de progression nomme l'objet qu'elle
  décrit, et le contrôle des ressources exige qu'il existe.
- **Le lieu de démonstration fait neuf fois la surface** — quatre-vingt-dix-huit
  cases de côté au lieu de trente-deux. L'ancienne carte tenait tout entière dans
  un écran et se traversait en six secondes et demie. Neuf blocs et une enceinte
  la composent, et c'est le premier lieu du dépôt à poser plus d'une pièce.
- **Un lieu couvre son étendue exactement une fois**, et le chargeur refuse les
  deux écarts en nommant la première case fautive : une case qu'aucune pièce ne
  pose se traverse sans se dessiner, et deux pièces qui se recouvrent se
  départagent par l'ordre des poses, que rien n'annonce.
- **`u` et `v` sont la case d'origine d'une pièce**, jamais son rang dans une
  trame : des pièces de tailles différentes se composent dans un même lieu, ce
  que l'enceinte du lieu livré fait. La conception montrait un champ `grille` et
  des positions de rang, que rien n'a jamais lus.
- **Le tir vise où la cible sera.** Le projectile met une demi-seconde à
  parcourir sa portée, pendant laquelle un Badaud avance de douze fois son rayon :
  tout ce qui traversait était manqué, et seul ce qui venait droit sur le joueur
  était touché de façon fiable. Les ratés demeurent — une créature qu'on repousse
  ou qui longe un mur change de cap en vol —, mais ils viennent de la situation
  et non d'un retard de visée.
- **La courbe de pression du lieu livré est réglée.** Elle s'ouvre à deux points
  par seconde et monte d'environ un tiers par palier, ce qui suit à peu près la
  façon dont l'arme grossit. Les pointes doublent le débit au lieu de le tripler.
- **Le lieu livré ne convoque plus qu'une créature**, le Badaud. Sept sortes
  qu'aucune teinte ne distingue rendaient la difficulté illisible : on ne pouvait
  pas savoir si une minute plus dure venait de la courbe ou d'une créature
  arrivée sans se signaler. Les six autres reviennent avec ce qui les rendra
  reconnaissables ; la courbe, elle, ne bouge pas d'un chiffre.

### Corrigé

- **Une durée nulle est refusée au chargement.** Un `plancher_ms` à zéro donnait
  un niveau et trois cartes à chaque tick, une `cadence_ms` à zéro une arme qui
  tire à chaque image : les deux se chargeaient sans un mot, dans les deux seuls
  manifestes tenus à la main. Le refus nomme la clé et dit ce que la valeur
  produirait.
- **Une phase de faible pression n'achetait rien.** La borne du budget reporté
  pouvait tomber sous le prix de la créature la moins chère de la phase : le
  budget montait, butait sur la borne, et rien n'apparaissait jamais — sans refus
  au chargement ni message, une salle simplement vide. La borne limite désormais
  l'accumulation sans jamais l'empêcher d'atteindre un achat.
- **Un tir pouvait traverser un obstacle.** Sa passabilité n'était mesurée qu'au
  point d'arrivée : un pas qui rasait l'angle où quatre cases se rencontrent
  entrait dans l'une et en ressortait sans y être vu.
- **Le bandeau se lisait mal.** Il n'avait pas de fond et se posait à même le
  décor, où le texte disparaissait sur un sol clair ; le niveau était de surcroît
  écrit dans la teinte atténuée, celle des phrases d'explication.

### À savoir

- **La version 0.3.0 a été construite le 3 septembre 2026 et n'a jamais été
  publiée.** Son tag reste dans le dépôt ; aucune archive n'en est sortie.
- **Le lieu livré ne convoque qu'une créature**, le Badaud. Les six autres
  profils existent et attendent ce qui les rendra reconnaissables à l'écran :
  une horde de sept sortes qu'aucune couleur ne distingue rend la difficulté
  illisible.
- **La courbe de pression n'est pas jugée.** Une séance de jeu a tranché ce
  qu'elle devait trancher — le déplacement et le tir —, pas l'équilibrage de fin
  de run : jouée contre une seule sorte de créature, elle ne dit rien de ce que
  vaudra la horde complète. Une partie peut donc ne pas se terminer.

***

**You can die here, and start over.** The horde that used to converge without
effect now hurts on contact: caught and unable to break free, the player falls
in five seconds. The death screen leaves in view whatever got the better of
them, and one key brings up a fresh run on the same place.

Killed creatures drop gems, and those gems raise the level — and a level up
pauses the game to offer **three cards**, one of which you take.

### Added

- **Player health and contact damage.** A creature pressed against the player
  hurts continuously rather than in a single blow, and the total taken per second
  is capped whatever the number of enemies in contact: being surrounded stays
  very dangerous, but it leaves a window to break free instead of killing
  outright.
- **Death freezes the scene, one key restarts.** The place and the horde stay on
  screen, darkened, and Space brings the run back with nothing to re-enter — same
  place, no selection to redo.
- **Gems.** A creature drops, exactly where it dies, as many gems as its profile
  declares; the player picks them up by passing within a short range.
- **Experience and levelling up.** Gems picked up raise the level, at a threshold
  that grows from one level to the next: it is the threshold that rises, a gem is
  worth the same from the start of a run to its end. A level is further granted
  every forty-five seconds without picking anything up, and it takes nothing away
  from what is already banked.
- **The run panel.** Health, experience and elapsed time are shown at the top of
  the screen, from the first tick to death.
- **`assets/progression/manifeste.json`.** Level thresholds, the time floor and
  what a gem is worth are tuned in this file, hand-held like the weapon table.

- **The magnet joins the catalogue**: a copper horseshoe in `assets/objets/`, and
  its rising sound in `assets/sons/`. The sound sweeps up one octave and ends
  where the ordinary pickup scale begins, so it is recognisably the same thing
  writ large. Nothing plays sound yet.
- **The magnet.** An object appears every thirty seconds, always at a
  distance: you pick it up and keep it, it does not trigger on contact. Key 1,
  on the numeric keypad as well as on the top row, spends the charge and
  **every gem on the ground converges at once** — those gems then stop fading,
  the magnet being the recourse against erasure rather than its victim. You hold
  only one, and no new one appears until the charge is spent.
- **The window has an icon**, drawn by `outils/interface.py` at three sizes for
  the system to choose from: a bright figure ringed by six others, which is what
  the game is about in sixteen pixels. Its palette is its own and does not follow
  the renderer's, whose flat colours are temporary.
- **A mark can be dropped while playing.** Space timestamps the moment in the
  run's log, along with the level reached and the tiers taken on each axis. The
  banner confirms it for two seconds under the timer: marking what you have just
  felt does not mean looking away from the horde, and what produced it is
  written down rather than remembered.
- **Gems fade away.** A gem left on the ground dims progressively and vanishes
  after six seconds. The fade is continuous rather than blinking: a gem's age
  can be read, so what a pile is still worth can be estimated. This is what will
  force you back to where you just killed, where the horde converges.
- **The pressure curve.** The horde is no longer placed at setup: a run starts in
  an empty room, and creatures are bought from a pressure budget per second. They
  appear outside the field of view, nineteen tiles from the player, and never
  inside a wall — better no creature at all than one out of nowhere. **The first
  level up can now be reached by playing.**
- **A place carries its wave scenario**, under `vagues` in `lieu.json`: phases
  dated on an `m:ss` timeline, a pressure per second, the profiles they allow,
  and an optional peak that multiplies the budget for a few seconds. A phase
  holds until the next one replaces it. A place without a scenario is a place
  without a horde, and that is allowed: not every room is an arena.
- **`assets/progression/manifeste.json` gains a `pression` section**: the spawn
  radius and the bound on carried-over budget. What the place decides is the
  rhythm; what the run decides is where creatures come from and what becomes of a
  budget that could not be spent.
- **A creature that is hit brightens for a moment.** It takes several hits, and
  until now nothing told a shot that missed from one that landed: misses were
  inferred rather than read.
- **The three level-up cards.** The horde freezes and three cards appear at the
  bottom of the screen: the left and right arrows move an amber outline, Enter
  or Space takes the card it points at. Two axes improve the weapon — fire rate
  and range, six tiers each — and a third card restores health when nothing else
  is available. The table is tuned in `assets/armes/manifeste.json`.
- **Numbers belong to the slots**, which keep them for the whole run: a badly
  chosen card is made up for at the next level, a magnet fired with nothing held
  is lost.

### Changed

- **Pickup range has left the player profile** for `assets/progression/`, where
  it joins a gem's lifetime. Both punish not collecting and are not tuned
  separately; leaving them in two files, one of them generated, would have made
  adjusting either term cost a regeneration of six hundred images.
- **Places now live inside a campaign**, `assets/campagnes/<name>/`, and the
  campaign is what you compose and will share: a folder of rooms whose
  `campagne.json` descriptor says which one a run starts in. The folder exists to
  scope names, as a place's folder already does for its rooms — without it, two
  authors both call a room "parking".
- **A place's files carry fixed names.** The room set is `jeu.json` and rooms
  live in `pieces/`, where two free names at the same level did not say which was
  a palette and which a plan. An existing place is brought over by renaming its
  room set and moving its rooms.
- **What a gem is worth has left `assets/objets/manifeste.json`** for the
  progression manifest: a value lives next to what it feeds, and this one feeds
  the thresholds. The progression manifest names the object it describes, and the
  resource check requires that object to exist.
- **The demonstration place is nine times the area** — ninety-eight cells a side
  instead of thirty-two. The old map fitted entirely within one screen and was
  crossed in six and a half seconds. Nine blocks and a perimeter wall make up the
  new one, and it is the first place in the repository to lay down more than one
  room.
- **A place covers its extent exactly once**, and the loader refuses both ways
  out of that, naming the first offending cell: a cell no room lays down is
  walked over without being drawn, and two overlapping rooms are settled by the
  order of the placements, which nothing announces.
- **`u` and `v` are a room's origin cell**, never its rank in a lattice: rooms of
  different sizes compose within one place, which is what the shipped place's
  perimeter does. The design document showed a `grille` field and rank positions
  that nothing ever read.
- **Shots aim where the target will be.** A projectile takes half a second to
  cover its range, during which a Badaud moves twelve times its own radius:
  anything crossing was missed, and only what came straight at the player was
  reliably hit. Misses remain — a creature pushed aside or sliding along a wall
  changes course mid-flight — but they now come from the situation rather than
  from the aim lagging behind.
- **The shipped place's pressure curve is tuned.** It opens at two points per
  second and rises by about a third per phase, roughly following how the weapon
  grows. Peaks double the rate instead of tripling it.
- **The shipped place now summons a single creature**, the Badaud. Seven kinds
  that no colour told apart made difficulty unreadable: there was no way to know
  whether a harder minute came from the curve or from a creature that had arrived
  unannounced. The other six return along with what will make them
  recognisable; the curve itself does not move by a single figure.

### Fixed

- **A zero duration is rejected at load time.** A `plancher_ms` of zero granted a
  level and three cards every tick, a `cadence_ms` of zero a weapon firing every
  frame: both loaded without a word, in the only two hand-written manifests. The
  rejection names the key and says what the value would produce.
- **A low-pressure phase bought nothing.** The bound on carried-over budget could
  fall below the price of the phase's cheapest creature: the budget rose, hit the
  bound, and nothing ever appeared — no rejection at load, no message, just an
  empty room. The bound now limits accumulation without ever keeping it from
  reaching a single purchase.
- **A shot could pass through an obstacle.** Its passability was only measured at
  the arrival point: a step grazing the corner where four cells meet entered one
  of them and left again without ever being seen inside it.
- **The panel was hard to read.** It had no ground of its own and sat straight on
  the scenery, where text vanished over light floor; the level was written in the
  dimmed tint besides, the one used for explanatory lines.

### Good to know

- **Version 0.3.0 was built on 3 September 2026 and never published.** Its tag
  remains in the repository; no archive was ever released from it.
- **The shipped place summons a single creature**, the Badaud. The other six
  profiles exist and are waiting for what will make them recognisable on screen:
  a horde of seven kinds that no colour tells apart makes difficulty unreadable.
- **The pressure curve has not been judged.** A play session settled what it was
  meant to settle — movement and shooting —, not late-run balance: played against
  a single kind of creature, it says nothing about what the full horde will be
  worth. A run may therefore fail to end.

## [0.2.1] — 2026-09-01 — Ce que la relecture a corrigé

**Rien ne change en jeu.** Un binaire 0.2.1 se comporte exactement comme un
0.2.0 : cette version corrige un défaut interne qu'on ne voit pas et remet la
documentation d'accord avec le code. Aucun lieu n'est à reprendre.

### Corrigé

- Le champ de flux allouait de la mémoire au fil de ses reconstructions : ses
  files naissaient vides et grandissaient à mesure que le joueur traversait le
  lieu, là où le budget d'allocation l'interdit. Elles reçoivent leur place au
  montage.
- Le contrôle des mentions de licence sautait tous les fichiers JSON, en
  confiant les manifestes à un contrôle qui ne regarde que `assets/` : ceux du
  reste du dépôt n'étaient donc lus par personne.

### Modifié

- La documentation dit ce que le code fait — l'ordre d'un pas de simulation, le
  domicile des plafonds de bassin, la clé du tri en profondeur, la structure
  d'une entité et l'origine d'une tuile. L'agrandissement du tampon vers la
  fenêtre est inscrit à l'étape des écrans de réglages, où il se décide.
- Les identifiants publics et les clés de journalisation sont en anglais, ce que
  les règles du dépôt demandaient déjà.

***

**Nothing changes in game.** A 0.2.1 binary behaves exactly like a 0.2.0 one:
this release fixes an internal defect no one can see and brings the
documentation back in line with the code. No level needs revisiting.

### Fixed

- The flow field allocated memory across its rebuilds: its queues started empty
  and grew as the player crossed the level, where the allocation budget forbids
  it. They now get their room at setup.
- The licence notice check skipped every JSON file, leaving manifests to a check
  that only looks at `assets/`: those elsewhere in the repository were read by
  no one.

### Changed

- The documentation states what the code does — the order of a simulation step,
  where the pool caps live, the depth sort key, an entity's layout and a tile's
  origin. Scaling the buffer to the window is recorded under the settings
  screens step, where it gets decided.
- Public identifiers and logging keys are in English, as the repository's own
  rules already required.

## [0.2.0] — 2026-08-31 — La horde à l'écran

**Cette version se regarde et se traverse, elle ne se joue pas encore.** La
fenêtre s'ouvre sur un lieu, on s'y déplace au clavier, et la horde converge en
contournant les obstacles — mais on ne peut pas mourir, donc pas recommencer.

Il n'y a donc ni écran de mort, ni porte, ni caisse, ni arme à ramasser, ni
montée de niveau, ni enchaînement de lieux. Le décor est fait de rectangles
colorés : les images existent dans le binaire, mais rien ne les charge encore.
C'est le comportement attendu, pas un défaut.

### Ajouté

- **Le lieu s'affiche, et se parcourt au clavier** : les flèches, ou le carré de
  touches à gauche — WASD sur un clavier américain, ZQSD sur un clavier français,
  puisque c'est la place d'une touche qui compte et non la lettre inscrite
  dessus. La caméra suit le joueur, se bloque aux bords du lieu plutôt que de
  découvrir du vide, et centre une fois pour toutes un lieu qui tient dans
  l'écran.
- Le sol montre ce que la simulation lit d'une case — franchissable, coûteuse ou
  mur — et non le décor, qui viendra avec les images.
- **La horde est à l'écran et converge**, en contournant les obstacles. Elle est
  semée au lancement et n'arrive pas par vagues : le scénario de pression viendra
  plus tard. Les créatures se recouvrent dans l'ordre de leur profondeur, et le
  joueur passe devant celles qui partagent la sienne.

### À savoir

- **La fenêtre agrandit l'image d'un facteur qui n'est pas entier**, si on la
  redimensionne : les pixels deviennent alors inégaux. Le tampon interne, lui,
  est bien fixe. L'agrandissement en entier est un réglage d'affichage, et il
  arrivera avec les écrans de réglages.

***

**This release can be looked at and walked through, it cannot be played yet.**
The window opens on a level, you move through it with the keyboard, and the
horde closes in around obstacles — but you cannot die, so you cannot start over.

There is therefore no death screen, no door, no crate, no weapon to pick up, no
level-up, no chain of levels. The scenery is coloured rectangles: the images are
in the binary, but nothing loads them yet. This is the expected behaviour, not a
defect.

### Added

- **The level is drawn, and can be walked with the keyboard**: the arrow keys, or
  the square of keys on the left — WASD on a US layout, ZQSD on a French one,
  since what counts is where a key sits and not the letter printed on it. The
  camera follows the player, stops at the level edges rather than revealing empty
  space, and centres a level small enough to fit the screen once and for all.
- The floor shows what the simulation reads from a tile — walkable, costly or
  wall — and not the scenery, which comes with the images.
- **The horde is on screen and closes in**, going around obstacles. It is seeded
  at startup and does not arrive in waves: the pressure schedule comes later.
  Creatures overlap in depth order, and the player is drawn in front of those
  sharing the same depth.

### Worth knowing

- **The window scales the image by a factor that is not an integer** if you
  resize it: pixels then come out uneven. The internal buffer itself is fixed.
  Integer scaling is a display setting, and it will arrive with the settings
  screens.

## [0.1.0] — 2026-08-31 — La simulation nue

**Cette version ne se joue pas et ne montre rien.** Le binaire charge ses
ressources, monte le monde et s'arrête sur « à implémenter : étape 2 » : c'est le
comportement attendu, pas un défaut. L'étape 1 de la feuille de route est la
simulation, sans rendu par construction ; la 0.2.0 sera la première à ouvrir une
fenêtre.

Il n'y a donc ni affichage, ni clavier, ni porte, ni caisse, ni arme à ramasser,
ni montée de niveau, ni enchaînement de lieux. Ce qui existe se mesure en test.

### Ajouté

- Le format des lieux et des pièces, en JSON, `version_format` 1. Un lieu ne
  porte que des identifiants de pièces et leurs positions, et ne peut embarquer
  ni image ni son. La clé `$comment` y est admise partout et sert de
  commentaire ; toute autre clé inconnue fait refuser le fichier.
- Le coût de traversée d'une tuile de décor, `cout_traversee`. La flaque, le sol
  sale et le sol fissuré ralentissent qui les traverse — le joueur comme les
  créatures, qui les contournent plutôt que d'y passer. Il est exigé sur ce qui
  se franchit et refusé sur ce qui bloque.
- Un lieu livré, `assets/lieux/place/`, chargé au lancement par le même chemin
  qu'un lieu écrit par un tiers. Les ressources sont embarquées dans le binaire :
  l'exécutable se suffit à lui-même.
- **Un lieu est un dossier** : son `lieu.json`, son jeu de pièces et ses pièces.
  Les noms de pièces sont ainsi locaux à un lieu, et deux auteurs peuvent nommer
  chacun la leur `carrefour`. Le nom du dossier et le champ `identifiant` doivent
  s'accorder, faute de quoi le lieu est refusé.
- La simulation : bassins d'entités préalloués et références qui survivent à la
  suppression, champ de flux pondéré par les coûts de terrain, grille de densité
  qui desserre la horde, et tir automatique visant le plus proche à portée. Trois
  cents poursuivants convergent vers une cible mobile en contournant les
  obstacles, sans une allocation par tick.
- Le manifeste des armes, `assets/armes/manifeste.json`, le seul de `assets/`
  tenu à la main. Le tireur porte les valeurs de son tir — cadence, portée,
  dégâts, nombre de projectiles, vitesse — et un projectile ne porte que son
  apparence.

***

**This release does not play and shows nothing.** The binary loads its assets,
builds the world and stops on "à implémenter : étape 2": that is the expected
behaviour, not a defect. Step 1 of the roadmap is the simulation, without
rendering by construction; 0.2.0 will be the first to open a window.

So there is no display, no keyboard, no door, no crate, no weapon to pick up, no
level-up and no chaining of places. What exists is measured in tests.

### Added

- The level and room format, in JSON, `version_format` 1. A level carries only
  room identifiers and their positions, and can embed neither images nor sounds.
  The `$comment` key is allowed anywhere and serves as a comment; any other
  unknown key causes the file to be rejected.
- The traversal cost of a decor tile, `cout_traversee`. Puddles, dirty ground
  and cracked ground slow down whoever crosses them — the player as much as the
  creatures, which go around rather than through. It is required on whatever can
  be crossed and rejected on whatever blocks.
- A shipped level, `assets/lieux/place/`, loaded at startup through the same path
  as a level written by a third party. Assets are embedded in the binary: the
  executable stands on its own.
- **A level is a directory**: its `lieu.json`, its room set and its rooms. Room
  names are thus local to a level, and two authors can each name theirs
  `carrefour`. The directory name and the `identifiant` field must agree, or the
  level is rejected.
- The simulation: preallocated entity pools with references that survive
  removal, a flow field weighted by terrain costs, a density grid that loosens
  the horde, and automatic fire aiming at the nearest target in range. Three
  hundred pursuers close in on a moving target while going around obstacles,
  without a single allocation per tick.
- The weapon manifest, `assets/armes/manifeste.json`, the only one in `assets/`
  kept by hand. The shooter carries the values of its shot — cooldown, range,
  damage, projectile count, speed — and a projectile carries only its appearance.
