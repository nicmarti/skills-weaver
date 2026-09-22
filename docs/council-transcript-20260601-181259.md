# Council Transcript — Outil de Debug/Cohérence SkillsWeaver

**Date :** 2026-06-01
**Sujet :** Architecture et plan de développement d'un outil de debug/analyse de cohérence pour le moteur de JdR D&D 5e SkillsWeaver.

---

## Question originale (utilisateur)

> J'aimerai implémenter un outil pour debug une aventure. L'objectif est de permettre au concepteur du jeu, mais aussi à Claude Code, d'analyser une partie de jeu en cours, de vérifier la cohérence de l'aventure. Cela veut dire qu'il faut définir « qu'est-ce qu'une aventure cohérente ». Au-delà de la cohérence, le squelette de l'aventure et le déroulé général doivent être tous les deux cohérents. On peut d'abord compléter le système de « logs » existant. Forme : un onglet spécial ou une URL accessible pour examiner visuellement une aventure, utile aussi au développement futur. Il y a des incohérences réelles dans les fichiers `journal-session-XXX.json` (suite à des corrections de bugs).

---

## Question cadrée (transmise aux 5 conseillers)

Concevoir un outil de debug/analyse de cohérence pour le moteur SkillsWeaver (Go + Gin/HTMX/SSE web + agent DM IA sw-dm). Objectif : permettre au concepteur ET à Claude Code d'analyser une partie en cours et de vérifier la cohérence à deux niveaux — **(1) le SQUELETTE** (campaign-plan : actes, antagoniste, foreshadows, threads narratifs) et **(2) le DÉROULÉ** (journal session par session, progression, NPCs, inventaire) — en interne ET l'un vis-à-vis de l'autre. Forme : onglet/URL dédiée dans sw-web, lisible aussi par Claude Code.

**Contexte technique existant :** `journal.go` (JournalEntry + journal-meta.json/journal-session-N.json, 15 catégories), `campaign_plan.go` (Acts, Foreshadows liés, Progression, Pacing), `validation.go` existant (`ValidateCampaignPlan → ValidationResult{Errors,Warnings,Score 0-100}`), sessions.json/party.json/inventory.json/npcs-generated.json/state.json/agent-states.json, logs sw-dm-session-N.log, web en html/template + HTMX + SSE, data/world/ (canon).

**Corruption réelle déjà détectée :** IDs de journal fragmentés/troués (81,104,106,127 manquants), chronologie inversée entre session-0 et session-1, session-0 utilisé comme fourre-tout, next_id désynchronisé, categories[] vide dans une aventure.

**Sous-questions :** (a) CLI vs web vs package partagé `internal/coherence` ? (b) structure des règles (intégrité/narratif/gameplay) extensible ? (c) visualisation ? (d) rapport JSON machine pour Claude Code ? (e) enrichir le système de logs ? (f) découpage en phases ?

---

## Réponses des conseillers

### Le Contrarian
Tu traites ça comme un problème d'outillage alors que les données prouvent que c'est un problème de *writes*. IDs fragmentés, chronologie inversée, session-0 fourre-tout, next_id désynchronisé. Un analyseur de cohérence ne corrige rien — il génère juste un plus joli rapport de la corruption que ton code continue de produire. Fixe le write-path du journal (allocation d'ID atomique, routage de session, sync du meta) D'ABORD, ou l'outil mesure du bruit.

Deuxième faille : la « cohérence » entre squelette et déroulé n'est pas calculable. Les règles d'intégrité sont déterministes — automatise-les. Mais « l'arc de l'antagoniste a-t-il avancé ? » exige de juger l'intention narrative. Ton score 0-100 va confondre « aucun ID manquant » avec « l'histoire a du sens ». Ne fusionne JAMAIS intégrité et narratif en un seul score.

Troisième : qui consomme ça, et quand ? Un onglet de debug que personne n'ouvre est un poids mort. Il doit tourner *automatiquement* et hurler à la régression.

Questions évitées : tes deux aventures sont en désaccord sur `categories[]` et la contiguïté des IDs — laquelle est correcte ? Tu ne peux pas valider contre un schéma non gelé. Quel est le coût des faux positifs ?

Verdict : plus petite chose viable — CLI `sw-validate` sur un schéma *gelé*, en CI, intégrité-seule, piloté par exit-code. La cohérence narrative est un jouet de recherche ; livre-la en dernier, si jamais.

### Le Penseur Premiers Principes
Tu confonds deux problèmes de formes opposées. « Cohérence » suppose des données fiables. Mais le journal est structurellement cassé : ce n'est pas une aventure incohérente, c'est une **base de données corrompue**. Aucune analyse narrative ne signifie quoi que ce soit par-dessus de mauvais octets.

Deux primitives en couches :
1. **Intégrité** (bien formé et canonique ?) — déterministe, doit passer en premier ; causé par les writers, pas les readers. Un checker read-only qui trouve des trous à chaque fois est du théâtre. Rends l'allocation d'ID/routage/ordre corrects-par-construction au write-path. Alors le check devient un garde-fou de régression.
2. **Cohérence** (l'histoire bien formée a du sens ?) — n'a de sens qu'au-dessus de l'intégrité, floue/de jugement — c'est là que l'LLM/Claude Code intervient, pas un moteur Go.

Réponse : le check de cohérence n'est pas la bonne primitive en premier. Construis l'intégrité (renomme le package `integrity`) + le fix du write-path. Émets du JSON (Claude est ton vrai moteur de cohérence — ne hardcode pas le narratif en Go). Onglet web = viewer, phase 3. Bonne question : « pourquoi mes données me mentent-elles, et qu'est-ce qui rend une source autoritaire ? »

### L'Expansionniste
Construis le checker comme package partagé `internal/coherence` d'abord — CLI et web sont de fins renderers. Mais ce n'est pas un outil de debug : c'est un **moteur de qualité narrative et un flywheel d'entraînement du DM IA**.

Le vrai prix : chaque règle de cohérence (foreshadow abandonné, thread délaissé, importance NPC jamais payée, stagnation de pacing) devient un **signal réinjecté dans l'agent sw-dm** — nourris les violations dans le briefing world-keeper à `start_session`. Le DM s'auto-corrige en cours de campagne. Un moat qu'aucun autre projet d'IA-DM n'a.

Plus gros coups : le **graphe de foreshadows** (planté vs tiré) est de l'or YouTube, screenshotable. Le **Score de Cohérence dans le temps** = battement de cœur de la campagne. Le **diff squelette-vs-déroulé** est un produit réutilisable. **Règles comme données** (JSON/YAML) pour que Claude Code écrive de nouvelles règles par l'exemple.

Phase 1 : package + JSON + 5 règles. Phase 2 : réinjection dans la boucle agent. Phase 3 : le graphe. Ne construis pas un linter — construis la chose qui rend le DM plus malin à chaque session.

### L'Outsider
« Cohérent » fait trop de travail. Tu empiles trois contrôles sans rapport : (1) **intégrité des données** (linter déterministe), (2) **drift plan-vs-réalité** (diff mécanique), (3) **qualité narrative** (jugement IA/humain). Les mélanger = un outil mou qui fait les trois mal. Nomme-les séparément.

Qui ouvre ça, et quand ? Non dit. Créateur solo débuggant vs MJ préparant = surfaces différentes. Tu ne peux pas concevoir la visualisation avant d'y répondre.

Le but humain-visuel et le but IA-lisible ne sont **PAS le même outil**. L'IA veut des faits structurés (JSON). L'humain veut une forme saisissable d'un coup d'œil (timeline, drapeaux rouges). Le moteur d'analyse est partagé ; les renderers diffèrent. Moteur une fois, émets JSON, affiche une page par-dessus.

Ordre : intégrité d'abord (tu ne peux pas juger la cohérence sur des données corrompues — garbage in) ; linter d'intégrité + meilleur logging = même fix (la corruption venait d'écritures bâclées) ; puis JSON ; puis drift/qualité ; onglet web en dernier. Ne construis pas le joli onglet en premier.

### L'Exécuteur
Construis `internal/coherence/` comme package partagé — pas un nouveau binaire, pas web-first.

Lundi matin :
1. `internal/coherence/report.go` — `Report{Skeleton adventure.ValidationResult; DataIntegrity []Finding; Flow []Finding; OverallScore int}` et `Finding{Severity string; Rule string; Message string; Context map[string]any}`.
2. `internal/coherence/analyze.go` — `Analyze(advDir string) (*Report, error)` charge journal-meta + sessions, appelle `ValidateCampaignPlan()`, fourre dans `Report.Skeleton`.
3. `internal/coherence/integrity.go` — porte la corruption déjà trouvée : `checkJournalIDGaps`, `checkChronoOrder`, `checkNextIDSync`, `checkEmptyCategories`. ROI maximal.
4. Route Gin `GET /play/:slug/coherence` → partial HTMX itérant les Findings dans un `<table>`. Copie le handler `/info` et change le corps.

À zapper : règles narratives/gameplay, SSE, visualisations, extension de logs. Livre intégrité + squelette cette semaine.

Pour (d) JSON machine : gratuit — `?format=json`, `c.JSON(200, report)`. Même struct. C'est ton feed Claude Code.

Premier pas maintenant : `integrity.go` avec `checkJournalIDGaps`. Tu sais qu'il trouve 81,104,106,127. Écris-la, assure contre l'aventure réelle, test vert, fini.

---

## Peer reviews (anonymisation révélée)

**Mapping :** A = Expansionniste · B = Contrarian · C = Exécuteur · D = Premiers Principes · E = Outsider

### Relecture 1
1. **C la plus forte** — seule à transformer le prompt en étapes lundi-matin exécutables (structs concrètes, IDs corrompus réels à asserter, raccourci copie-/info-handler, JSON gratuit). Répond aux 6 sous-questions en restant livrable.
2. **A le plus gros angle mort** — ignore que les données sont corrompues ; bâtir un « DM auto-correcteur » sur des octets-déchets amplifie le bruit. Court vers les graphes YouTube avant de fixer l'allocation d'ID atomique.
3. **Tous ont manqué :** quelle source fait foi quand squelette et déroulé divergent, et le chemin de migration/réparation des saves déjà corrompus. Aucun ne spécifie un outil de réparation → rapport rouge en permanence sur les legacy saves, signal inutile.

### Relecture 2
1. **C la plus forte** — seule à devenir du code runnable lundi (types concrets, réutilise `ValidateCampaignPlan()`, porte les corruptions observées, réutilise le pattern `/info`, JSON gratuit). D et E ont le bon insight mais restent au niveau principe.
2. **A le plus gros angle mort** — impossible de bâtir un « flywheel de qualité narrative » quand les IDs et la chronologie mentent ; réintroduit le score 0-100 unique que B/D/E rejettent.
3. **Tous ont sauté la question source-autoritaire/réparation.** D la gesticule mais personne ne spécifie la *remédiation* : une fois les trous détectés, l'outil répare-t-il les aventures corrompues existantes ? Détection sans réparation laisse la corruption en place pour toujours.

### Relecture 3
1. **C la plus forte** — seule à répondre concrètement aux 6 questions (package partagé, structs nommées, checks d'intégrité réels, route HTMX clonant `/info`, JSON gratuit, premier pas testable). D et E diagnostiquent bien mais refusent de construire.
2. **A le plus gros angle mort** — ignore entièrement la corruption ; nourrir le briefing world-keeper pendant que le journal a des trous = DM s'auto-corrige sur du garbage. Reframe un outil de debug en pitch produit « YouTube gold / flywheel ».
3. **Tous ont manqué :** quelle source fait foi en cas de désaccord, et qui répare les données déjà corrompues. Aucun ne mentionne que `validation.go` et un nouveau checker d'intégrité vont se chevaucher, ni la réconciliation des deux aventures aux schémas divergents.

### Relecture 4
1. **C la plus forte** — seule à mapper directement sur le code existant (réutilise `ValidateCampaignPlan()`, porte la corruption documentée, copie le pattern `/info` éprouvé, test vert contre une aventure réelle dès le jour 1). Sépare intégrité et narratif sans philosopher.
2. **A le plus gros angle mort** — ignore que la source est corrompue. Optimise pour un moat narratif et des démos YouTube pendant que la fondation est en feu.
3. **Tous ont manqué la question migration/réparation.** Les deux aventures existantes sont *déjà* corrompues contre un schéma non gelé. Personne ne spécifie comment *réparer* (un `sw-adventure repair` one-shot), ni qui arbitre le désaccord de schéma. Détection sans remédiation = corpus rouge en permanence.

### Relecture 5
1. **C la plus forte** — seule à traduire le design en Go livrable (types concrets, layout fichiers, réutilisation, route HTMX clonée de `/info`, JSON, premier test vert contre données corrompues réelles). Absorbe les insights justes de B/D/E.
2. **A le plus gros angle mort** — romance la couche narrative/flywheel IA en ignorant que les données sont déjà corrompues par le write-path. Nourrir les briefings de violations bruitées rendrait le DM plus bête.
3. **Tous ont manqué remédiation/backfill.** Aucun ne propose de *réparer* (un mode `--fix`/migrate, ou comment re-homer sans risque les orphelins de session-0). Aussi non traité : quelle source fait foi quand journal-meta et fichiers de session divergent.

---

## Synthèse du Chairman

### Là où le council converge (signal fort)
- **« Cohérence » est un mot-valise dangereux.** 4/5 séparent spontanément : (1) intégrité des données (déterministe), (2) drift squelette-vs-déroulé (diff mécanique), (3) qualité narrative (jugement → IA). Ne jamais fusionner intégrité et narratif dans un score 0-100 unique.
- **L'intégrité d'abord, non négociable.** Aucune analyse narrative n'a de sens sur des octets corrompus. Le journal troué = base corrompue, pas aventure incohérente.
- **La corruption vient du write-path.** Un checker read-only qui retrouve les mêmes trous est du théâtre ; les bugs sont dans le code qui écrit (allocation d'ID, routage de session, sync du meta).
- **Architecture : un moteur partagé `internal/coherence`, plusieurs front-ends.** CLI et web = fins renderers. Émettre JSON, afficher une page par-dessus.
- **Le rapport JSON est quasi gratuit** et c'est le vrai canal pour Claude Code (`?format=json`, même struct).

### Là où le council s'affronte
- **Narratif : jamais ou plus tard ?** Le Contrarian le relègue à « jamais ». Premiers Principes + Expansionniste : pas en Go, parce que Claude/l'IA EST le moteur de cohérence narrative. Résolution : pas « jamais » — « pas en Go ». Le Go produit des faits ; l'IA juge en consommant le JSON.
- **CLI-en-CI vs onglet web.** Pas vraiment un conflit (moteur partagé sert les deux), mais le Contrarian gagne sur un point : un onglet que personne n'ouvre est mort — il faut un déclenchement automatique (CI ou post-session).
- **Vision vs fondations.** L'Expansionniste a raison sur la direction (flywheel, graphe), tort sur l'ordre : auto-corriger un DM sur des données mensongères le rend plus bête.

### Angle mort révélé par le peer-review
**Les 5 ont tous oublié la remédiation.** Tous détectent et préviennent ; les deux aventures existantes sont DÉJÀ corrompues. Sans un `sw-adventure repair` / mode `--fix` (re-homing des orphelins de session-0, resync de `next_id`), le rapport reste **rouge en permanence** → signal inutile. Corollaire : désigner la **source autoritaire** en cas de contradiction et **geler le schéma** avant de valider.

### Recommandation
Un seul moteur partagé `internal/coherence` avec **trois couches nommées et non fusionnées** :
1. **Intégrité** (déterministe, Go) : porter les corruptions réelles en checks. ROI max, testable dès lundi.
2. **Drift squelette↔déroulé** (diff mécanique, Go) : réutiliser `ValidateCampaignPlan()` + croiser foreshadows plantés/résolus, threads, NPCs plan vs `npcs-generated.json`, sessions cibles des actes vs jouées.
3. **Qualité narrative** (PAS en Go) : le moteur émet du JSON ; Claude Code juge. Réinjection world-keeper plus tard.

Struct `Report{Skeleton, DataIntegrity, Drift, score-PAR-COUCHE}` + `Finding{Severity, Rule, Message, Context}`. Exposer (a) route web `GET /play/:slug/coherence` (partial HTMX cloné de `/info`) + `?format=json`, et (b) binaire fin `sw-validate` exit-code pour la CI. **En parallèle de la couche 1, livrer `sw-adventure repair`** (re-home session-0, resync next_id, gel du schéma).

**On ne fait PAS maintenant :** graphe de foreshadows, SSE, flywheel agent, score global unique, règles narratives en Go.

### La seule chose à faire en premier
Créer `internal/coherence/integrity.go` avec `checkJournalIDGaps(advDir) []Finding`, et écrire le test qui **assure** qu'il retrouve les IDs manquants connus (`81, 104, 106, 127`) sur `le-voyageur-de-tuncmor`. Test vert = la fondation déterministe tient. Tout le reste se branche dessus.
