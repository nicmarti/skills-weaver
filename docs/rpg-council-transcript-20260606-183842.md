# Conseil de la Table — Transcript complet

**Date :** 2026-06-06 18:38:42
**Objet :** Analyse du rapport de cohérence de l'aventure `les-cendres-du-commandant-oublie` (4 sessions jouées)
**5e siège :** MJ Pragmatique (pas de swap — dimension table/gameplay)
**Sources analysées (session web live → analyse de fichiers, pas de navigateur) :**
- `data/adventures/les-cendres-du-commandant-oublie/coherence-narrative.json`
- `data/adventures/les-cendres-du-commandant-oublie/foreshadows.json`
- `adventure.json`

---

## Question d'origine

> Invoque le conseil de JDR afin d'analyser le résultat de cohérence et la critique de scénario de la dernière aventure (`/play/les-cendres-du-commandant-oublie/coherence`). Propose quelques actions/modifications pour prendre en compte le retour, améliorer l'aventure et éviter les soucis identifiés sur l'histoire.

## Question cadrée (envoyée aux 5 conseillers)

Aventure « Les Cendres du Commandant Oublié » (4 sessions, terminée). Le rapport de cohérence (lentilles world-keeper, rules-keeper, scenario-critic + synthèse) identifie :

- **Constat central (3 lentilles convergent) :** l'antagoniste **Volker Eisenherz n'a jamais été une présence jouée** — menace administrative, nommé par tous les PNJ, aperçu une fois de loin (S2, `fsh_004`), confirmé « à 2 jours de marche » (S4), absent de la résolution. La révélation finale est rapportée par Gorren *après* la livraison — narrée, pas jouée.
- **Règles :** effondrement des compétences en S4 (10 jets / 7 scènes) ; Intuition (WIS) jamais lancée en 4 sessions malgré 10+ PNJ à agenda caché ; 2 jets de sauvegarde sur 239 (0,8 %) ; interrogatoire-clé résolu sans mécanique (RAW = test contesté) ; biais combat dans l'XP (S3 social = 150 XP vs combat S1-S2 = 350 XP) ; XP S4 non loggés ; capacités de classe (Attaque Sournoise, Fougue d'Action, Channel Divinity, Inspiration Bardique) jamais confirmées dans les logs ; 46 jets pour 3 éclaireurs CR ~1/2.
- **Scénario/pacing :** asymétrie structurelle (Acte 1 surchargé d'exposition, Acte 2 le plus riche, Acte 3 compressé en une session) ; PNJ qui déversent un bloc d'exposition à la première apparition ; fils non payés (objet métallique de Gorren, ouverture du 3e caisson) ; `fsh_003` et `fsh_004` toujours statut `active`.

**Question :** quelles actions/modifications concrètes — (a) pour cette aventure, (b) pour le moteur/personas (dungeon-master, rules-keeper, world-keeper, campaign-plan, briefing `start_session`) — afin d'éviter ces problèmes à l'avenir ? Quel est le changement à plus fort levier ?

---

## Mapping d'anonymisation (révélé)

| Lettre | Conseiller |
|--------|-----------|
| A | L'Architecte Narratif |
| B | Le Briseur de Jeu |
| C | Le Joueur à la Table |
| D | Le MJ Pragmatique |
| E | Le Gardien des Règles |

---

## Réponses des conseillers

### 🎭 Le Joueur à la Table

Volker n'a jamais existé pour le joueur. C'est ça, le crime. Quatre sessions à courir après un nom — jamais un visage, jamais une réplique, jamais un jet d'opposition contre lui. La satisfaction promise (« on va faire tomber Volker ») a été livrée par **Gorren en voix off, après coup**. On m'a raconté ma propre victoire. C'est le contraire d'un climax : c'est un reçu.

Et S4 confirme que le moteur me retire la manette au pire moment. La scène d'interrogatoire — l'info la plus stratégique de l'aventure — résolue sans un jet ? Ça veut dire que mon perso, mes stats, mon Attaque Sournoise, mon Inspiration Bardique n'ont **aucune prise** sur le moment qui compte. J'aurais pu être un caillou. Intuition jamais lancée avec 10 PNJ menteurs : le DM a décidé seul de ce que je « ressens ». Pas d'agentivité, pas de doute, pas de tension.

**Le changement à plus fort levier — un seul :** une règle dure dans le persona dungeon-master + check au `start_session` de l'acte final : l'antagoniste nommé DOIT apparaître en scène jouée, avec au moins un échange direct PJ↔antagoniste avant la résolution, et la confrontation finale ne peut PAS être narrée en récit rapporté. Si le campaign-plan a un antagoniste, le briefing de l'acte 3 doit le flagger : « Volker n'a jamais été rencontré — programmer la scène face-à-face cette session. » Pas de payoff en voix off. Jamais.

En second levier, pour le ressenti : interdire la résolution narrative des scènes-clés. Enquête, interrogatoire, négociation à enjeu → **jet obligatoire**, avec échec possible. Et que rules-keeper rappelle au DM d'inviter les capacités signature (« quelqu'un veut tenter une Intuition ? »). Je veux *agir* sur l'histoire, pas la recevoir.

Pour cette aventure : le fil de l'objet métallique de Gorren et la silhouette de Volker (`fsh_004`) sont encore actifs — c'est mon accroche pour rejouer. Donnez-moi une session 5 où je regarde enfin Volker dans les yeux.

### ⚖️ Le Gardien des Règles

Effondrement des compétences en S4 — symptôme le plus grave côté RAW. Une scène d'interrogatoire stratégique résolue sans jet viole directement le SRD : c'est un **test de caractéristique contesté** (Intimidation/Persuasion CHA vs Intuition/Tromperie WIS, SRD « Contests »). Idem, l'Intuition jamais lancée sur 10+ PNJ à agenda caché = la mécanique de détection de mensonge purement décorative.

**Levier #1 (le plus fort) — un tool `request_check` couplé à un gating dans `start_session`.** Ajouter dans `internal/dmtools/` un tool qui force le DM à déclarer, pour toute scène d'enjeu (interrogatoire, négociation, fouille, détection), la DC + la compétence + l'éventuel jet contesté AVANT de narrer l'issue. Le briefing `start_session` doit inclure un **bilan de couverture mécanique de la session précédente** : compétences/sauvegardes non sollicitées, capacités de classe jamais déclenchées. C'est le point d'appui : on transforme une consigne molle de persona en donnée de contexte mesurée.

**Capacités de classe.** Le rules-keeper doit recevoir la composition du groupe et vérifier l'usage RAW : Attaque Sournoise (1d6/niveau impair, une fois/tour) sous-évaluée ici = dégâts faussés sur 4 sessions ; Fougue d'Action (action bonus d'attaque) perdue = combat allongé — d'où vraisemblablement les **46 jets pour 3 éclaireurs CR 1/2** (rencontre « facile », qui traîne faute d'optimisation des actions).

**XP — corriger le biais combat.** Le SRD prévoit explicitement l'**XP pour objectifs non-combat** (DMG « Noncombat Challenges »). 150 XP pour la session sociale la plus dense vs 350 combat est un barème cassé. Imposer dans le tool XP une attribution par **pilier** (combat / exploration / interaction) et **logger S4** (actuellement absent → niveau bloqué).

**(a) Rétro-correction :** oui. Recréditer XP S4 + rééquilibrer S3 fait franchir le palier — le groupe **doit être niveau 3**. C'est dû RAW, pas un cadeau.

**(b) Changement à plus fort levier :** le bilan de couverture mécanique injecté au `start_session`. Le reste découle de cette mesure.

### 📖 L'Architecte Narratif

L'antagoniste absent est le symptôme ; la cause est architecturale, et elle est dans ton domaine. Le `campaign-plan.json` tracke des foreshadows (planted/payoff) mais ne tracke jamais les **apparitions de l'antagoniste**. `fsh_004` a été planté en S2 et personne, ni le moteur ni le DM, n'a réclamé sa montée en présence. Volker est resté « active » parce que rien ne forçait son incarnation.

**Le changement à plus fort levier : un budget d'apparitions d'antagoniste dans le campaign-plan, vérifié par le briefing de `start_session`.**

(a) **Clore rétroactivement** — Une scène épilogue jouée, pas narrée : Gorren convoque le groupe, étale les preuves, et Volker apparaît enfin (interrogatoire mécanisé, jet social). Identifier l'objet métallique (clé du 3e caisson) et résoudre qui a rescellé. Passer `fsh_003`/`fsh_004` en `paid`. C'est une demi-session, pas un patch JSON.

(b) **Moteur/personas** — trois leviers concrets :

1. **Schéma campaign-plan** : ajouter à l'antagoniste `presence_schedule` (`offscreen → glimpsed → confronted`) avec session-cible par palier, et un `escalation_debt` qui monte tant que l'antagoniste n'a pas progressé d'un cran. Le briefing de `start_session` lit ce debt et l'injecte : *« Volker n'a pas été incarné depuis 2 sessions — fais-le AGIR cette session. »*
2. **Briefing anti-compression** : `start_session` calcule `sessions_restantes vs actes_restants`. Si Acte 3 doit tenir en 1 session, le briefing l'annonce et exige une session de respiration AVANT la résolution. Pas de dénouement en un souffle.
3. **Persona dungeon-master** : règle anti-exposition. Un PNJ-clé ne livre **jamais** son bloc complet à la première apparition — fragment, contradiction, ou silence. L'info se *construit*. Et règle foreshadow : aucun fsh ne meurt `active` — le briefing de la dernière session liste les fsh non payés comme dette à résoudre.

Le world-keeper devrait *réclamer* l'antagoniste, pas seulement le mentionner. Sinon on rejouera ce vide à chaque campagne.

### 💥 Le Briseur de Jeu

Le piège fatal, formulé d'emblée : **le rapport de cohérence mesure la complétude du LOGGING, pas la qualité du JEU**, et il appelle ces deux choses par le même mot. CLAUDE.md le dit noir sur blanc — combats loggés coup par coup, marqueurs de progression rarement écrits, `sessions.json` vide pour location/xp. Donc « Attaque Sournoise jamais confirmée » ne signifie PAS « jamais appliquée » : ça signifie « le moteur n'a aucun champ qui l'enregistre ». Tu instrumentes une absence de capteur comme une absence d'événement. C'est de la pareidolie sur des trous de log.

Le biais structurel : `roll_dice` est le SEUL événement fiablement tracé. Tout ce que le rapport « trouve » converge donc mécaniquement vers « pas assez de jets » — parce que c'est la seule chose visible. Le combat (46 jets) explose les compteurs ; la résolution narrative (interrogatoire) est invisible. Conclusion biaisée garantie : « biais combat », « effondrement compétences ». Le rapport ne fait que redécrire la couverture de son propre télescope.

La correction naïve est le vrai danger. Forcer plus de jets = jets sans enjeu (le DM lancera Intuition pour cocher la case). Forcer Volker en scène = antagoniste téléporté pour satisfaire une métrique. Tu transformerais un DM qui prend des **raccourcis légitimes** (résoudre vite, comprimer) en DM qui fait du **théâtre pour les logs**. Pire jeu, meilleur rapport. C'est l'effet pervers maximal.

Actions concrètes :

**(a) Cette aventure** : aucune « correction ». Tu ne peux pas distinguer non-joué de non-loggé a posteriori. Verdict honnête : « non mesurable ». Relire les transcripts bruts (pas le journal) pour Volker/interrogatoire — la seule source vérité.

**(b) Moteur/personas** : ne force RIEN côté DM. Corrige l'INSTRUMENTATION : logger les résolutions narratives (`narrative_resolution` avec enjeu), les capacités de classe déclenchées, les XP S4. Et tague chaque rapport : « basé sur logs `roll_dice` uniquement — silence ≠ absence ». Tant que le capteur ment, tout audit ment avec lui.

### 🛠️ Le MJ Pragmatique

L'antagoniste fantôme et les compétences effondrées sont deux symptômes, une seule cause : le briefing `start_session` ne RÉCLAME rien. Il informe, il n'exige pas. Chemin le plus court :

**(b) Moteur/personas — coût quasi nul :**

1. **Briefing `start_session` : ajouter 3 lignes impératives, pas un sous-système.** « Antagoniste actif ce chapitre : Volker. Avant la fin de session, il DOIT agir on-screen au moins une fois (sbire, lettre, sabotage, apparition). » C'est du `fmt.Sprintf` dans le code de briefing — pas un tool, pas une scène. Volker passe d'administratif à incarné pour le prix d'une string.
2. **Persona dungeon-master : une puce sur la résolution non-combat.** « Quand un PJ enquête/interroge/persuade, EXIGE un jet (Intuition, Investigation, Persuasion) avant de donner l'info. Les PNJ ne déversent pas l'exposition : ils résistent, mentent, négocient. » Règle les compétences mortes, l'Intuition jamais lancée et l'interrogatoire non mécanisé d'un coup. Zéro Go.
3. **Foreshadow nag dans le briefing.** Lister les `fsh_*` "active" depuis >2 sessions avec « à résoudre ou abandonner explicitement ». Déjà dans `foreshadows.json`, juste à afficher. Une boucle, une ligne.

**(a) Cette aventure :** rien à rejouer. Marquer `fsh_003`/`fsh_004` `resolved` ou `abandoned` à la main dans le JSON. Note de campagne. Fini.

**NE PAS FAIRE :** le sous-système « scène d'antagoniste » dédié, le moteur de pacing qui interdit de compresser l'acte final, un tool `force_skill_check`, le tracking automatique des capacités de classe (Sournoise/Fougue). Tout ça, c'est des centaines de lignes Go, de la dette, des tokens — pour discipliner un agent Opus qui obéit déjà à une instruction de prompt bien formulée. L'XP biaisé/non loggé : vrai bug, mais cosmétique sur 4 sessions — un `log_event` rappelé dans le persona suffit, pas un correcteur d'XP.

Le levier, c'est l'impératif dans le briefing, pas le code.

---

## Relectures par les pairs (anonymes)

### Relecteur 1
1. **Plus forte = A (Architecte Narratif).** Seule à relier le symptôme à la cause architecturale précise (le campaign-plan tracke les foreshadows mais jamais la présence de l'antagoniste) et propose le mécanisme manquant (`presence_schedule` + `escalation_debt`). Bon niveau d'abstraction, sans l'usine à gaz que D dénonce.
2. **Angle mort = E (Gardien des Règles).** Empile `request_check`, gating, bilan de couverture, XP par pilier sans entendre l'avertissement de B ; confond conformité RAW et qualité de jeu.
3. **Manque commun : qui déclenche la correction, et quand.** `start_session` arrive trop tard — l'antagoniste absent se constate en *fin* de session. Il manque un check de clôture (`end_session`) qui reporte l'`escalation_debt`. Et personne ne questionne le rapport lui-même comme déclencheur fiable.

### Relecteur 2
1. **Plus forte = B.** Seule à attaquer la validité de la prémisse, sur un fait documenté de CLAUDE.md (`roll_dice` seul fiablement tracé). Sa recommandation est la condition préalable aux autres.
2. **Angle mort = E.** Rétro-correction XP « niveau 3 RAW » à partir de logs déclarés non fiables. Confond non-loggé et non-appliqué.
3. **Manque commun : biais de version moteur.** Savoir si « Les Cendres » a tourné sur la version courante avant de patcher ; rejouer une aventure de référence pour valider ; horodater aventure/version.

### Relecteur 3
1. **Plus forte = B.** Questionne l'instrument plutôt que d'obéir au diagnostic. Corriger l'instrumentation avant de discipliner l'agent = seul ordre d'opérations correct.
2. **Angle mort = E.** Rétro-correction XP présume non-loggé = non-attribué ; `request_check` = over-engineering que D critique.
3. **Manque commun :** aucune ne fait de la relecture du transcript brut S4 le préalable bloquant aux corrections.

### Relecteur 4
1. **Plus forte = B.** Remet en cause la prémisse, fait documenté CLAUDE.md ; nomme le danger « pire jeu, meilleur rapport ».
2. **Angle mort = E.** Empile du code, exige une rétro-correction sur données non fiables ; manque le doute sur la source.
3. **Manque commun :** valider le diagnostic avant d'agir (relire le transcript brut) + biais inter-versions (patcher peut viser un bug déjà corrigé).

### Relecteur 5
1. **Plus forte = B.** Seule à attaquer la prémisse au lieu d'optimiser une métrique trompeuse ; sa correction est la condition préalable au reste.
2. **Angle mort = E (et la rétro-correction XP de A/E).** « Imposer niveau 3 RAW » sur des XP non fiablement loggés ; empile tools + gating + bilans (centaines de lignes Go).
3. **Manque commun :** personne ne propose de valider qu'une correction marche avant de la généraliser (rejouer une session de référence sur la version courante). Mesurer la qualité du jeu reste non résolu — tous corrigent des proxies.

---

## Verdict du président

### Là où le conseil s'accorde

1. **Le briefing `start_session` est le point de levier — mais il informe sans rien exiger.** MJ Pragmatique, Joueur et Gardien des Règles pointent le même endroit. La correction la moins chère vit là : transformer trois lignes descriptives en trois lignes impératives.
2. **La résolution narrative des scènes-clés est une faute de jeu, pas seulement de log.** Joueur, Gardien des Règles et Architecte convergent : l'interrogatoire sans jet et la confrontation Volker narrée « après coup » privent le joueur de prise. Le persona DM doit interdire de résoudre une scène-pivot en récit pur.
3. **Pour CETTE aventure, on ne rejoue presque rien.** Quasi-unanimité : marquer `fsh_003`/`fsh_004` à la main, au mieux un court épilogue joué.

### Là où le conseil s'affronte

Le vrai affrontement est **Gardien des Règles vs Briseur de Jeu**, et la relecture est brutale : quatre relecteurs sur cinq désignent B comme la voix la plus forte et E comme l'angle mort principal.

- **Camp E — corriger et instrumenter.** L'effondrement des compétences, l'Intuition jamais lancée, l'interrogatoire sans jet sont des symptômes mesurables. Réponse : tool `request_check`, gating, bilan de couverture mécanique, rétro-correction XP. Logique : sans contrainte mécanique, le DM continuera à narrer ce qui devrait être joué.
- **Camp B — l'instrument ment, ne force rien.** Le rapport mesure le *logging*, pas le *jeu*. CLAUDE.md : combats coup par coup, marqueurs rarement écrits, `sessions.json` vide, `roll_dice` seul fiable. Donc « Attaque Sournoise jamais confirmée » ≠ jamais appliquée. Forcer des jets = jets sans enjeu ; forcer Volker = antagoniste téléporté. **Meilleur rapport, pire jeu.**

Pourquoi des conseillers raisonnables divergent : E fait confiance au rapport comme mesure de l'état du jeu ; B refuse cette prémisse. Désaccord de **diagnostic**, pas de solution. La rétro-correction XP de E (relayée par A) s'effondre : corriger des XP « dus RAW » sur des logs que CLAUDE.md déclare non fiables, c'est calibrer une règle sur un instrument cassé. Cinq relecteurs sur cinq la rejettent. Le conseil tranche : **pas de rétro-correction XP.**

### Angles morts détectés par le conseil

- **Valider le diagnostic sur les transcripts bruts AVANT d'agir.** Le rapport à 3 lentilles est dérivé du journal/des logs — l'accusé rédige son propre PV.
- **Le briefing `start_session` arrive trop tard pour l'antagoniste absent.** Il manque un **check de clôture `end_session`** qui reporte l'`escalation_debt`.
- **Biais de version moteur** (CLAUDE.md) : ne pas patcher si l'aventure est pré-version-courante.
- **La qualité réelle du jeu reste non mesurée** : chacun corrige un *proxy*.

### La recommandation

**B a raison sur le diagnostic, le MJ Pragmatique a raison sur la cure. On n'instrumente pas à l'aveugle, et on ne code surtout pas un sous-système avant d'avoir vérifié que le problème existe sur la version actuelle.**

**Étape 0 — Vérifier le terrain (bloquant, coût nul).** (a) Lire le transcript brut de S4 — pas le journal — pour confirmer que les capacités/jets manquent *réellement* ; (b) vérifier la version moteur/personas sous laquelle l'aventure a tourné. Si le diagnostic ne survit pas, ou si l'aventure est pré-version-courante, **on s'arrête là.**

**Pour CETTE aventure (cheap, manuel).** Marquer `fsh_003`/`fsh_004` en `resolved` ou `abandoned`. Pas de rétro-correction XP. Épilogue joué optionnel (Gorren convoque, Volker en scène, interrogatoire mécanisé) **uniquement si le joueur le veut.**

**Pour le moteur — CHEAP (prompt/briefing, `fmt.Sprintf`), si l'étape 0 confirme :**
1. **Briefing impératif** : « Antagoniste actif : Volker. Il DOIT agir on-screen au moins une fois cette session. »
2. **Persona DM, une puce** : exiger un jet avant de livrer une info à enjeu ; les PNJ résistent, mentent, négocient ; ne jamais résoudre une scène-pivot en récit rapporté.
3. **Foreshadow/escalation nag** : lister au briefing les `fsh` `active` depuis > 2 sessions + injecter, depuis un check de clôture `end_session`, la dette si l'antagoniste n'a pas agi.

**Pour le moteur — CHER, à NE PAS faire :** tool `request_check`/`force_skill_check`, moteur de pacing, sous-système de scène d'antagoniste, tracking automatique des capacités, gating dur. Des centaines de lignes de Go pour discipliner un agent Opus qui obéit déjà à un bon prompt — et calibrées sur un instrument non fiable.

**Validation avant généralisation :** après les patchs prompt, rejouer une aventure de référence courte sur la version courante et relire son transcript brut. Si le briefing impératif + la puce persona suffisent, ne rien coder de plus.

Le changement à plus fort levier n'est donc **pas** un tool ni un check : c'est **l'impératif au briefing + sa boucle de dette à `end_session`**, gardé par une vérification du diagnostic en amont.

### La première chose à faire

Lire le **transcript brut de la session 4** de « Les Cendres du Commandant Oublié » (pas le journal, pas le rapport) et confirmer que l'antagoniste, l'interrogatoire et les capacités manquent vraiment au jeu — avant d'écrire une seule ligne de patch.
