# Plan d'Améliorations - SkillsWeaver

> Analyse effectuée le 21 décembre 2025
> Version du projet : Post-Phase 9 (Journal Illustrator)

## Sommaire

- [Criticité CRITIQUE](#criticité-critique) - Bloque le gameplay
- [Criticité HAUTE](#criticité-haute) - Gameplay incomplet
- [Criticité MOYENNE](#criticité-moyenne) - Polish et cohérence
- [Criticité BASSE](#criticité-basse) - Nice-to-have

---

## Criticité CRITIQUE

Ces problèmes cassent le gameplay et doivent être résolus en priorité.

---

### CRIT-01: Convention AC Ambiguë

**Statut**: [x] Terminé (21 déc 2025)

**Problème**:
Le système utilise deux conventions contradictoires pour la Classe d'Armure (AC) :
- `rules-keeper.md` et le code personnage utilisent AC montante (moderne) : sans armure = 11, plates = 17
- Les règles BFRPG originales utilisent AC descendante : sans armure = 10, plates = 2

Exemple d'incohérence :
- `monsters.json` : `"armor_class": 13` pour un gobelin
- `equipment.json` : `"ac_bonus": 6` pour plates (qui s'additionne à quoi ?)
- Calcul d'attaque : confus entre "toucher AC 13" et "battre AC 13"

**Impact**: Les attaques des monstres contre les PJ et vice-versa utilisent une logique incohérente.

**Solutions proposées**:

1. **Option A - Garder AC montante (recommandé)** : Plus intuitif pour les joueurs modernes
   - Documenter clairement : "AC 11 = base sans armure, +DEX, +bonus armure"
   - Vérifier que tous les monstres utilisent la même convention
   - Formule attaque : d20 + bonus >= AC cible

2. **Option B - Revenir à BFRPG pur** : Fidélité aux règles originales
   - Convertir tous les AC de monstres
   - AC descend : meilleur = plus bas
   - Formule attaque : d20 >= THAC0 - AC cible

**Fichiers à modifier**:
- `data/monsters.json` - Vérifier/convertir tous les AC
- `data/equipment.json` - Clarifier ac_bonus
- `.claude/agents/rules-keeper.md` - Documenter la convention choisie
- `internal/character/character.go` - Implémenter calcul AC correct

**Prompt**:
```
Implémente la convention AC montante cohérente dans SkillsWeaver :

1. Vérifie et documente que tous les monstres dans data/monsters.json utilisent AC montante (plus haut = mieux protégé)

2. Clarifie dans equipment.json que ac_bonus s'ajoute à AC base 11

3. Dans rules-keeper.md, ajoute une section claire :
   - AC de base sans armure : 11
   - Formule : AC = 11 + bonus DEX + bonus armure
   - Pour toucher : d20 + bonus attaque >= AC cible

4. Dans internal/character/character.go, implémente CalculateAC() qui :
   - Prend l'AC base (11)
   - Ajoute le modificateur DEX
   - Ajoute le bonus d'armure équipée (si implémenté)

Assure-toi que la logique est cohérente entre PJ et monstres.
```

**Résolution** (21 déc 2025):
- ✅ Vérifié : `monsters.json` utilise déjà la convention AC montante (13-20)
- ✅ Ajouté `_armor_class_system` dans `equipment.json` documentant la convention
- ✅ Mis à jour `rules-keeper.md` avec formule explicite et exemples
- ✅ `CalculateArmorClass()` existait déjà dans `character.go`
- ✅ Ajouté 10 tests unitaires dans `character_test.go` (TestCalculateArmorClass, TestCalculateArmorClassWithNonArmorItems)

---

### CRIT-02: Équipements Magicien Manquants

**Statut**: [x] Terminé (21 déc 2025)

**Problème**:
Dans `equipment.json`, le starting_equipment pour magic-user référence :
```json
"required": ["backpack", "bedroll", "rations_iron", "waterskin", "spellbook", "ink_quill"]
```

Mais `spellbook` et `ink_quill` n'existent pas dans la section `adventuring_gear`.

**Impact**: La création de magiciens de niveau 1 référence des équipements inexistants.

**Solution**:
Ajouter les items manquants à `adventuring_gear` dans `equipment.json`.

**Fichiers à modifier**:
- `data/equipment.json`

**Prompt**:
```
Ajoute les équipements manquants pour les magiciens dans data/equipment.json :

Dans la section "adventuring_gear", ajoute :
1. spellbook (Grimoire) - coût 50 gp, poids 3 lb
2. ink_quill (Plume et encre) - coût 5 gp, poids 0.1 lb

Vérifie ensuite que le starting_equipment de magic-user référence des items existants.
```

**Résolution** (21 déc 2025):
- ✅ Ajouté `spellbook` (Grimoire, 50 gp, 3 lb) dans `equipment.json`
- ✅ Ajouté `ink_quill` (Plume et encre, 5 gp, 0.1 lb) dans `equipment.json`
- ✅ Ajouté test `TestMagicUserEquipment` vérifiant l'existence et les propriétés
- ✅ Ajouté test `TestStartingEquipmentReferencesExist` vérifiant toutes les classes

---

### CRIT-03: Calcul AC Personnage Non Implémenté

**Statut**: [x] Terminé (21 déc 2025) - Résolu avec CRIT-01

**Problème**:
La structure Character a un champ `ArmorClass` mais il n'est jamais calculé dynamiquement :
- Pas de prise en compte de l'équipement
- Pas d'ajout du modificateur DEX
- Valeur codée en dur ou aléatoire

Exemple : `thorin.json` montre `"armor_class": 11` (juste le DEX, sans armure)

**Impact**: Les personnages n'ont pas d'AC correcte, rendant le combat impossible.

**Solution**:
Implémenter une méthode `CalculateAC()` dans le package character.

**Fichiers à modifier**:
- `internal/character/character.go`
- `internal/character/character_test.go`

**Prompt**:
```
Implémente le calcul de Classe d'Armure dans internal/character/character.go :

1. Ajoute une méthode CalculateAC() à la structure Character :
   - AC de base = 11
   - Ajoute le modificateur Dexterity
   - Pour l'instant, ne gère pas l'équipement (sera fait plus tard)
   - Minimum AC = 1

2. Modifie la création de personnage pour appeler CalculateAC() automatiquement

3. Ajoute des tests dans character_test.go :
   - Personnage avec DEX 18 (+3) devrait avoir AC 14
   - Personnage avec DEX 3 (-3) devrait avoir AC 8
   - Personnage avec DEX 10 (0) devrait avoir AC 11

4. Mets à jour la méthode d'export pour inclure l'AC calculée
```

**Résolution** (21 déc 2025):
- ✅ `CalculateArmorClass()` existait déjà dans `character.go:203-216`
- ✅ Prend en compte AC base 11 + DEX modifier + armor bonus
- ✅ 10 tests unitaires ajoutés lors de CRIT-01 (TestCalculateArmorClass)
- ℹ️ Note : Les objets magiques ne sont pas encore gérés (amélioration future)

---

### CRIT-04: Level Limits Confus (0 = quoi ?)

**Statut**: [x] Terminé (21 déc 2025)

**Problème**:
Dans `races.json`, les `level_limits` utilisent 0 de façon ambiguë :
```json
"level_limits": {
  "fighter": 6,
  "magic-user": 9,
  "thief": 0  // 0 = pas de limite ou interdit ?
}
```

Les vrais règles BFRPG : les voleurs n'ont pas de limite de niveau pour aucune race.

**Impact**: Confusion sur ce que signifie 0 - "illimité" ou "classe interdite".

**Solution**:
Utiliser une convention claire :
- `-1` ou valeur absente = illimité
- `0` = classe interdite pour cette race

**Fichiers à modifier**:
- `data/races.json`
- `internal/data/loader.go` (interprétation)
- Documentation

**Prompt**:
```
Clarifie la convention des level_limits dans SkillsWeaver :

1. Dans data/races.json, applique la convention :
   - Valeur positive (ex: 6) = niveau maximum
   - Valeur -1 = pas de limite de niveau
   - Absence de la classe = classe interdite pour cette race

2. Mets à jour toutes les races avec cette convention :
   - Human : toutes classes à -1 (illimité)
   - Elf : fighter 6, magic-user 9, thief -1, cleric absent (interdit)
   - Dwarf : fighter 7, cleric 6, thief -1, magic-user absent (interdit)
   - Halfling : fighter 4, thief -1, autres absentes

3. Dans internal/data/loader.go, ajoute une fonction :
   func (r *Race) GetLevelLimit(className string) (limit int, allowed bool)
   - Retourne -1, true si illimité
   - Retourne N, true si limité à N
   - Retourne 0, false si classe interdite

4. Documente cette convention dans le header de races.json
```

**Résolution** (21 déc 2025):
- ✅ Analysé `races.json` : la convention actuelle utilise 0 = "illimité" (correct pour BFRPG)
- ✅ Ajouté `_level_limits_convention` documentation dans `races.json` expliquant :
  - 0 = Unlimited (pas de limite de niveau)
  - Nombre positif = Niveau maximum
  - Format string "6/9" = Multi-classe
  - Classe absente de level_limits = classe interdite
- ✅ Amélioré commentaires de `GetLevelLimit()` dans `loader.go` avec documentation des valeurs de retour
- ✅ Étendu `TestGetLevelLimit` de 5 à 18 cas de test couvrant tous les scénarios

---

### CRIT-05: Calcul PV Niveau 1 Non Implémenté

**Statut**: [x] Terminé (21 déc 2025)

**Problème**:
Le calcul des points de vie au niveau 1 n'est pas clairement implémenté :
- La règle BFRPG : Dé de Vie de la classe + modificateur CON (minimum 1)
- Le code actuel semble donner le max du dé + CON, mais pas documenté

**Impact**: Incertitude sur les PV des personnages créés.

**Solution**:
Implémenter et documenter clairement la méthode de calcul des PV.

**Fichiers à modifier**:
- `internal/character/character.go`
- `internal/character/character_test.go`

**Prompt**:
```
Implémente le calcul des Points de Vie niveau 1 dans internal/character/character.go :

1. Ajoute ou clarifie la méthode CalculateHitPoints(class *Class) :
   - Au niveau 1, on lance le Dé de Vie de la classe (d8 guerrier, d6 clerc, d4 magicien/voleur)
   - On ajoute le modificateur Constitution
   - Minimum 1 PV (même avec CON très basse)
   - MaxHitPoints = HitPoints au niveau 1

2. Option de création :
   - --max-hp : donne le maximum du dé + CON (pour niveau 1)
   - Par défaut : lance le dé + CON

3. Ajoute des tests :
   - Guerrier CON 16 (+2) : PV entre 3 et 10 (d8+2)
   - Magicien CON 6 (-1) : PV entre 1 et 3 (d4-1, min 1)
   - Avec --max-hp : Guerrier = 8 + CON mod

4. Documente dans character-generator skill la méthode utilisée
```

**Résolution** (21 déc 2025):
- ✅ Ajouté paramètre `maxHP bool` à `RollHitPoints()` dans `character.go`
- ✅ Deux méthodes supportées :
  - `maxHP=false` : lance le dé de vie aléatoirement (règle BFRPG standard)
  - `maxHP=true` : prend le maximum du dé (variante pour survie)
- ✅ Ajouté `--max-hp` flag dans CLI `sw-character create`
- ✅ Documentation complète du calcul avec exemples
- ✅ 6 nouveaux tests ajoutés :
  - `TestRollHitPointsMaxHP` : 4 classes avec max HP
  - `TestRollHitPointsRandomRoll` : vérification des plages de valeurs
  - `TestRollHitPointsMinimumOneHP` : vérification minimum 1 PV
  - `TestRollHitPointsRandomWithLowCON` : clamp avec CON négatif
- ✅ Mis à jour `rules-keeper.md` avec section "Points de Vie au Niveau 1"

---

## Criticité HAUTE

Ces problèmes rendent le gameplay incomplet mais ne le cassent pas totalement.

---

### HIGH-01: Système de Sorts Manquant

**Statut**: [x] Complété (22 déc 2025)

**Problème**:
Le jeu n'a pas de système de sorts implémenté :
- Pas de `spells.json` avec la liste des sorts
- Pas de structure Spell dans le code
- Les magiciens et clercs ne peuvent pas vraiment lancer de sorts
- `classes.json` a `spells_per_level` mais rien pour les utiliser

**Impact**: Les classes de lanceurs de sorts sont inutilisables en combat.

**Résolution** (22 déc 2025):
- ✅ Créé `data/spells.json` avec 41 sorts BFRPG niveaux 1-2 :
  - 8 sorts cléricaux niveau 1 (Soins légers, Détection du mal, Détection de la magie, Lumière, Protection contre le mal, Purification nourriture/eau, Délivrance de la peur, Résistance au froid)
  - 9 sorts cléricaux niveau 2 (Bénédiction, Charme-animal, Détection des pièges, Immobilisation de personne, Résistance au feu, Silence 15' de rayon, Communication avec les animaux, Marteau spirituel)
  - 13 sorts arcaniques niveau 1 (Charme-personne, Détection de la magie, Disque flottant, Verrouillage, Lumière, Projectile magique, Bouche magique, Protection contre le mal, Lecture des langues, Lecture de la magie, Bouclier, Sommeil, Ventriloquie)
  - 12 sorts arcaniques niveau 2 (Lumière éternelle, Détection du mal, Détection de l'invisible, Invisibilité, Déblocage, Lévitation, Localisation d'objet, Lecture des pensées, Image miroir, Force fantasmagorique, Toile d'araignée, Verrou magique)
- ✅ Structure JSON complète avec : id, name_en, name_fr, level, type, reversible, range, duration, description bilingue, save, damage si applicable
- ✅ Ajouté à Character struct dans `internal/character/character.go` :
  - `KnownSpells []string` - IDs des sorts connus
  - `PreparedSpells []string` - sorts préparés pour la journée
  - `SpellSlots map[int]int` - slots disponibles par niveau
  - `SpellSlotsUsed map[int]int` - slots utilisés par niveau
- ✅ Ajouté méthodes dans character.go :
  - `InitializeSpellSlots(gd *data.GameData) bool` - initialise les slots selon classe/niveau
  - `CanCastSpells(gd *data.GameData) bool` - vérifie si la classe peut lancer des sorts
  - `GetSpellType(gd *data.GameData) string` - retourne "arcane" ou "divine"
- ✅ Mis à jour `ToMarkdown()` pour afficher la section Magie avec slots et sorts
- ✅ Mis à jour `.claude/agents/rules-keeper.md` avec documentation complète :
  - Tables des emplacements de sorts par niveau
  - Règles de préparation et de lancement
  - Listes complètes des 41 sorts avec portée, durée et effet
- ✅ Code compile sans erreurs

**Fichiers modifiés**:
- `data/spells.json` (nouveau - 41 sorts)
- `internal/character/character.go` (structure + méthodes)
- `.claude/agents/rules-keeper.md` (documentation sorts)

---

### HIGH-02: Compétences Voleur Incomplètes

**Statut**: [ ] À faire

**Problème**:
Dans `classes.json`, les `thief_skills` ne sont documentées que pour les niveaux 1-3 :
```json
"thief_skills": {
  "1": { "open_locks": 25, ... },
  "2": { "open_locks": 30, ... },
  "3": { "open_locks": 35, ... }
}
```

La progression s'arrête au niveau 3.

**Impact**: Les voleurs au-delà du niveau 3 n'ont pas de compétences définies.

**Solution**:
Étendre les thief_skills jusqu'au niveau 10 avec progression +5% par niveau.

**Fichiers à modifier**:
- `data/classes.json`

**Prompt**:
```
Complète les compétences de voleur dans data/classes.json :

Étends thief_skills du niveau 4 au niveau 10.

Progression BFRPG (environ +5% par niveau) :
- open_locks : 25, 30, 35, 40, 45, 50, 55, 60, 65, 70
- remove_traps : 20, 25, 30, 35, 40, 45, 50, 55, 60, 65
- pick_pockets : 30, 35, 40, 45, 50, 55, 60, 65, 70, 75
- move_silently : 25, 30, 35, 40, 45, 50, 55, 60, 65, 70
- climb_walls : 80, 81, 82, 83, 84, 85, 86, 87, 88, 89
- hide : 10, 15, 20, 25, 30, 35, 40, 45, 50, 55
- listen : 30, 34, 38, 42, 46, 50, 54, 58, 62, 66

Ajoute les niveaux 4 à 10 dans la structure existante.
```

---

### HIGH-03: Initiative et Combat Non Implémentés

**Statut**: [x] Complété

**Problème**:
Aucun système de combat n'est implémenté :
- Pas de fonction pour rouler l'initiative
- Pas de structure pour tracker un combat en cours
- Pas de résolution d'attaque automatique

**Impact**: Le MJ doit tout faire manuellement, le système ne peut pas assister.

**Solution**:
Créer un système de combat basique dans le package adventure ou nouveau package combat.

**Fichiers à créer/modifier**:
- `internal/combat/combat.go` (nouveau)
- `internal/dice/dice.go` (ajouter Initiative)
- `.claude/agents/dungeon-master.md`

**Résolution** (21 déc 2025):
- ✅ Ajouté `Initiative(dexMod int)` dans `dice.go` : 1d6 + modificateur DEX
- ✅ Ajouté `AttackRoll(bonus int)` dans `dice.go` : d20 + bonus d'attaque
- ✅ Ajouté `NaturalRoll()`, `IsCriticalHit()`, `IsCriticalMiss()` pour les critiques
- ✅ Créé package `internal/combat/` avec système complet :
  - `Combatant` : nom, initiative, DEX mod, AC, HP, bonus attaque, dégâts, délai d'action
  - `Combat` : round, combattants, ordre de tour, tour actuel
  - `AttackResult` : résultat détaillé d'une attaque avec critiques
  - Fonctions : NewCombat, AddCombatant, RollInitiative, GetTurnOrder, NextTurn, NewRound, DelayAction, ActOnInitiative, Attack, Heal, TakeDamage, IsOver, GetWinner, Status
- ✅ 4 tests ajoutés dans `dice_test.go` (Initiative, AttackRoll, Critiques)
- ✅ 16 tests créés dans `combat_test.go`
- ✅ Documentation mise à jour dans `rules-keeper.md` avec règles BFRPG détaillées
- ⏳ Commandes CLI sw-adventure combat à implémenter séparément si nécessaire

**Prompt** (référence):
```
Crée un système de combat basique pour SkillsWeaver :

1. Dans internal/dice/dice.go, ajoute :
   - func (r *Roller) Initiative() int // 1d6
   - func (r *Roller) AttackRoll(bonus int) (roll int, natural int)

2. Crée internal/combat/combat.go avec :

   type Combatant struct {
     Name      string
     Initiative int
     AC        int
     HP        int
     MaxHP     int
     AttackBonus int
     IsEnemy   bool
   }

   type Combat struct {
     Round      int
     Combatants []Combatant
     CurrentTurn int
   }

   func NewCombat() *Combat
   func (c *Combat) AddCombatant(name string, ac, hp, attackBonus int, isEnemy bool)
   func (c *Combat) RollInitiative() // Roule pour tous
   func (c *Combat) GetTurnOrder() []Combatant
   func (c *Combat) NextTurn() *Combatant
   func (c *Combat) Attack(attacker, defender string, damage int) (hit bool, damageDealt int)

3. Ajoute une commande sw-adventure combat :
   - sw-adventure combat start "Aventure"
   - sw-adventure combat add "Goblin" --ac=13 --hp=4 --enemy
   - sw-adventure combat initiative
   - sw-adventure combat status

4. Documente dans rules-keeper.md la séquence de combat BFRPG
```

---

### HIGH-04: Système de Guérison Absent

**Statut**: [ ] À faire

**Problème**:
Aucun système de guérison n'est implémenté :
- Pas de fonction pour soigner des PV
- Pas de règle de repos (court/long)
- Les potions de soin existent dans treasure.json mais sans effet défini

**Impact**: Les personnages blessés ne peuvent pas récupérer.

**Solution**:
Implémenter des fonctions de guérison dans le package character.

**Fichiers à modifier**:
- `internal/character/character.go`
- `.claude/agents/rules-keeper.md`

**Prompt**:
```
Implémente un système de guérison dans SkillsWeaver :

1. Dans internal/character/character.go, ajoute :

   // Soigne un montant de PV (ne dépasse pas MaxHP)
   func (c *Character) Heal(amount int) int // retourne PV actuels

   // Subit des dégâts (minimum 0 PV)
   func (c *Character) TakeDamage(amount int) int // retourne PV restants

   // Repos court (1 heure) - récupère 1 PV si niveau 1-3, 2 si 4+
   func (c *Character) ShortRest() int

   // Repos long (8 heures) - récupère niveau PV
   func (c *Character) LongRest() int

   // Est mort ?
   func (c *Character) IsDead() bool // PV <= 0

2. Ajoute des tests pour chaque fonction

3. Dans rules-keeper.md, documente :
   - Repos court : 1h, récupère 1-2 PV selon niveau
   - Repos long : 8h, récupère niveau PV
   - Soins magiques : instantanés, effet complet
   - À 0 PV : inconscient, à -10 : mort

4. Mets à jour sw-character avec :
   - sw-character heal "Aldric" 5
   - sw-character damage "Aldric" 3
   - sw-character rest "Aldric" --short
   - sw-character rest "Aldric" --long
```

---

### HIGH-05: Progression XP et Level Up

**Statut**: [ ] À faire

**Problème**:
Le système ne gère pas la progression :
- Character a Level et XP mais pas de fonction LevelUp()
- Pas de vérification automatique du seuil XP
- Pas de calcul des nouveaux PV, sorts, etc.

**Impact**: Les personnages ne progressent jamais même après des combats.

**Solution**:
Implémenter un système de progression dans le package character.

**Fichiers à modifier**:
- `internal/character/character.go`
- `internal/character/character_test.go`
- `data/classes.json` (vérifier XP tables)

**Prompt**:
```
Implémente la progression de niveau dans SkillsWeaver :

1. Dans internal/character/character.go :

   // Ajoute de l'XP et vérifie le level up
   func (c *Character) GainXP(amount int, classData *Class) bool // true si level up

   // Monte de niveau
   func (c *Character) LevelUp(classData *Class) error

   Le level up doit :
   - Incrémenter Level
   - Rouler le nouveau dé de vie + CON mod (min 1)
   - Ajouter aux MaxHP et HP actuels
   - Mettre à jour AttackBonus selon la table de classe
   - Mettre à jour les SavingThrows
   - Pour voleur : mettre à jour les skills
   - Pour lanceurs : ajouter des spell slots

2. Ajoute une commande sw-character :
   - sw-character xp "Aldric" 500  // ajoute 500 XP
   - sw-character levelup "Aldric" // force le level up si XP suffisant

3. Vérifie que classes.json a les tables XP complètes :
   - xp_for_level: {"2": 2000, "3": 4000, ...}

4. Ajoute des tests :
   - Guerrier gagne 2000 XP -> passe niveau 2
   - PV augmentent de d8 + CON
   - Attack bonus passe à +2
```

---

### HIGH-06: Noms Halfelins Manquants

**Statut**: [ ] À faire

**Problème**:
Le fichier `names.json` ne contient pas de noms pour les halfelins :
- Section "halfling" absente ou vide
- Impossible de générer des noms de halfelins avec sw-names

**Impact**: Une race jouable n'a pas de support pour la génération de noms.

**Solution**:
Ajouter une section complète de noms halfelins style Hobbit/Tolkien.

**Fichiers à modifier**:
- `data/names.json`

**Prompt**:
```
Ajoute les noms de halfelins dans data/names.json :

Crée une section "halfling" avec la même structure que les autres races :
{
  "halfling": {
    "male": {
      "first": [...],  // 30-40 prénoms masculins
      "last": [...]    // 30-40 noms de famille
    },
    "female": {
      "first": [...],  // 30-40 prénoms féminins
      "last": [...]    // mêmes noms de famille
    }
  }
}

Style : Hobbit/Tolkien (Bilbo, Frodo, Sam, Pippin, Merry...)
- Prénoms courts, souvent terminés en -o, -i, -a
- Noms de famille descriptifs ou liés à la nature : Sacquet, Brandebouc, Touque, Gamegie, Bolgeur, Fierpied

Exemples de prénoms masculins : Bilbo, Frodo, Sam, Pippin, Merry, Hamfast, Drogo, Otho, Lotho, Fredegar, Folco, Griffo, Polo, Porto, Ponto, Largo, Fosco, Bingo...

Exemples de prénoms féminins : Rosa, Petunia, Primula, Dora, Lobelia, Angelica, Belladonna, Esmeralda, Melilot, Menegilda, Pearl, Ruby, Diamond...

Vérifie que sw-names generate halfling fonctionne après l'ajout.
```

---

### HIGH-07: Tables Turn Undead Incomplètes

**Statut**: [ ] À faire

**Problème**:
Dans `classes.json`, la table `turn_undead` pour les clercs est incomplète :
- Seulement quelques types de morts-vivants
- Pas de progression au-delà du niveau 7
- Pas de valeur "D" (destruction automatique)

**Impact**: Les clercs de haut niveau n'ont pas de règles pour repousser les morts-vivants.

**Solution**:
Compléter la table turn_undead avec tous les types et niveaux.

**Fichiers à modifier**:
- `data/classes.json`

**Prompt**:
```
Complète la table turn_undead dans data/classes.json :

La table doit couvrir :
- Niveaux de clerc : 1 à 10
- Types de morts-vivants : skeleton, zombie, ghoul, wight, wraith, mummy, spectre, vampire, liche

Format (valeur à atteindre sur 2d6, "D" = destruction automatique, "-" = impossible) :

"turn_undead": {
  "skeleton": {"1": 7, "2": "D", "3": "D", "4": "D", ...},
  "zombie": {"1": 9, "2": 7, "3": "D", "4": "D", ...},
  "ghoul": {"1": 11, "2": 9, "3": 7, "4": "D", ...},
  "wight": {"1": "-", "2": 11, "3": 9, "4": 7, ...},
  "wraith": {"1": "-", "2": "-", "3": 11, "4": 9, ...},
  "mummy": {"1": "-", "2": "-", "3": "-", "4": 11, ...},
  "spectre": {"1": "-", "2": "-", "3": "-", "4": "-", "5": 11, ...},
  "vampire": {"1": "-", "2": "-", "3": "-", "4": "-", "5": "-", "6": 11, ...},
  "liche": {"1": "-", ..., "9": 11, "10": 9}
}

Règle BFRPG : Plus le clerc est de haut niveau, plus il repousse facilement les morts-vivants faibles (D = destruction) et peut affecter les plus puissants.
```

---

### HIGH-08: Tables de Rencontres Limitées

**Statut**: [x] Complété

**Problème**:
`monsters.json` n'a que 5-6 tables de rencontres :
- dungeon_level_1 à 4
- forest
- undead_crypt

Manquent : caverne, mer/côte, ville, désert, montagne, marais...

**Impact**: Variété limitée pour le MJ dans les environnements.

**Solution**:
Ajouter 6+ nouvelles tables de rencontres avec créatures appropriées.

**Fichiers à modifier**:
- `data/monsters.json`

**Résolution** (21 déc 2025):
- ✅ Ajouté 22 nouveaux monstres au bestiaire (stats BFRPG officielles) :
  - Animaux : crocodile, crocodile_large, lion, hyena, giant_frog, giant_snake_constrictor, giant_snake_venomous, hawk, vulture, camel, sea_serpent
  - Monstres : giant_scorpion, giant_crab, mummy, lizardman
  - Humains : bandit, pirate, cultist, thug, city_guard, merchant
- ✅ Ajouté 9 nouvelles tables de rencontres :
  - `cave` : Caverne (kobolds, gobelins, chauve-souris, ours, troll)
  - `mountain` : Montagne (loups, harpies, ogres, dragon)
  - `ruins` : Ruines (morts-vivants, méduse, cockatrice)
  - `road` : Route (bandits, loups, gobelins, gnolls)
  - `urban` : Ville (brigands, gardes, cultistes, marchands)
  - `desert` : Désert (scorpions, momies, hyènes, lions)
  - `coastal` : Côte (crabes, pirates, crocodiles, serpent de mer)
  - `swamp` : Marais (hommes-lézards, grenouilles, crocodiles, trolls)
- ✅ Total : 15 tables de rencontres (6 existantes + 9 nouvelles)
- ✅ Total : 55 monstres (33 existants + 22 nouveaux)
- ✅ Sources : [The Free Bestiary BFRPG](https://clayadavis.gitlab.io/osr-bestiary/bestiary/bfrpg/)

**Prompt** (référence):
```
Ajoute des tables de rencontres dans data/monsters.json :

Dans la section "encounter_tables", ajoute :

1. "cave" (Caverne) :
   - giant_rat, giant_bat, giant_spider, kobold, goblin, orc, troll, bear

2. "mountain" (Montagne) :
   - wolf, dire_wolf, bear, giant_bat, harpy, ogre, troll, dragon_red_young

3. "swamp" (Marais) :
   - giant_rat, giant_centipede, lizardfolk (à ajouter), troll, green_slime, crocodile (à ajouter)

4. "coastal" (Côte) :
   - giant_crab (à ajouter), merfolk (à ajouter), shark (à ajouter), pirate (humain)

5. "urban" (Ville) :
   - rat, thug (humain), guard (humain), thief (humain), beggar (humain), cultist (humain)

6. "desert" (Désert) :
   - giant_scorpion (à ajouter), mummy, skeleton, vulture (à ajouter), sand_worm (à ajouter)

Pour chaque table, utilise le format existant :
"cave": ["giant_rat", "giant_bat", "kobold", "goblin", "giant_spider", "orc", "troll", "bear"]

Note : certains monstres devront être ajoutés au bestiaire principal.
```

---

## Criticité MOYENNE

Ces problèmes affectent la cohérence et le polish mais n'empêchent pas le jeu.

---

### MED-01: Documentation Agents vs Skills Floue

**Statut**: [x] Résolu

**Problème**:
La distinction entre agents et skills n'est pas claire :
- `character-creator.md` (agent) et `character-generator` (skill) font des choses similaires
- Le workflow réel (qui appelle quoi) n'est pas documenté
- `dungeon-master.md` n'utilise jamais explicitement certains skills

**Impact**: Confusion pour les utilisateurs et développeurs.

**Solution**:
Clarifier et documenter la hiérarchie et les responsabilités.

**Résolution**:
1. Section "Architecture : Skills vs Agents" ajoutée à CLAUDE.md avec :
   - Définitions claires (Skills = CLI, Agents = Personas)
   - Diagramme ASCII de la hiérarchie
   - Exemples de workflows (création de personnage, session de jeu)
2. Chaque agent liste maintenant ses skills dans une table "Skills Utilisés"
3. Chaque skill indique quels agents l'utilisent dans "Utilisé par"

**Fichiers modifiés**:
- `CLAUDE.md` (section Architecture)
- `.claude/agents/dungeon-master.md` (7 skills)
- `.claude/agents/character-creator.md` (3 skills)
- `.claude/agents/rules-keeper.md` (2 skills)
- `.claude/skills/*/SKILL.md` (9 skills)

---

### MED-02: Système d'Encombrement Non Implémenté

**Statut**: [ ] À faire

**Problème**:
`rules-keeper.md` mentionne l'encombrement mais il n'est pas codé :
- Equipment a des poids mais Character ne les track pas
- Pas de calcul de mouvement réduit
- Pas d'alerte quand surchargé

**Impact**: Les joueurs peuvent porter infiniment d'équipement.

**Solution**:
Implémenter le calcul d'encombrement dans character.

**Fichiers à modifier**:
- `internal/character/character.go`
- `.claude/agents/rules-keeper.md`

**Prompt**:
```
Implémente le système d'encombrement dans SkillsWeaver :

1. Dans internal/character/character.go :

   // Ajoute des champs
   type Character struct {
     // ... existants
     CarriedWeight float64  // poids total en livres
     Encumbrance   string   // "light", "medium", "heavy", "overloaded"
     Movement      int      // vitesse en pieds
   }

   // Calcule l'encombrement
   func (c *Character) CalculateEncumbrance() {
     // Basé sur FOR :
     // Light: jusqu'à FOR × 5 lb → 40' mouvement
     // Medium: FOR × 5 à FOR × 10 lb → 30' mouvement
     // Heavy: FOR × 10 à FOR × 15 lb → 20' mouvement
     // Overloaded: > FOR × 15 lb → 10' mouvement, -4 attaque
   }

2. Quand l'équipement change, recalculer automatiquement

3. Dans rules-keeper.md, documente les seuils et effets

4. Ajoute une commande :
   - sw-character encumbrance "Aldric"
```

---

### MED-03: Système d'Alignement Absent

**Statut**: [ ] À faire

**Problème**:
BFRPG utilise un système d'alignement simple (Bon, Neutre, Mauvais) mais :
- Pas de champ Alignment dans Character
- Pas de choix lors de la création
- Pas d'impact sur le jeu

**Impact**: Aspect roleplay manquant, certains sorts/objets dépendent de l'alignement.

**Solution**:
Ajouter l'alignement au système de personnages.

**Fichiers à modifier**:
- `internal/character/character.go`
- `cmd/character/main.go`
- `.claude/agents/character-creator.md`

**Prompt**:
```
Ajoute le système d'alignement dans SkillsWeaver :

1. Dans internal/character/character.go :

   type Character struct {
     // ... existants
     Alignment string  // "lawful", "neutral", "chaotic"
   }

   Valeurs acceptées :
   - "lawful" (Loyal/Bon)
   - "neutral" (Neutre)
   - "chaotic" (Chaotique/Mauvais)

2. Dans cmd/character/main.go :
   - Ajoute flag --alignment=lawful|neutral|chaotic
   - Par défaut : "neutral"

3. Dans character-creator.md, ajoute une étape :
   "Choisissez l'alignement de votre personnage :
   - Loyal : respecte l'ordre, protège les innocents
   - Neutre : équilibre, pragmatisme
   - Chaotique : liberté personnelle, peut être égoïste"

4. Dans rules-keeper.md, documente :
   - Certains sorts affectent selon l'alignement (Protection contre le mal)
   - Certains objets magiques réagissent à l'alignement
```

---

### MED-04: Occupations PNJ Incomplètes

**Statut**: [ ] À faire

**Problème**:
`npc-traits.json` n'a pas de catégories pour :
- Magiciens/lanceurs de sorts
- Guerriers professionnels
- Aventuriers

**Impact**: Impossible de générer des PNJ de ces types.

**Solution**:
Ajouter les catégories manquantes.

**Fichiers à modifier**:
- `data/npc-traits.json`

**Prompt**:
```
Complète les occupations dans data/npc-traits.json :

Ajoute les catégories manquantes :

"occupation": {
  // ... existantes ...

  "arcane": [
    "magicien", "apprenti mage", "alchimiste", "enchanteur",
    "nécromancien", "illusionniste", "sage", "bibliothécaire",
    "herboriste mystique", "astrologue"
  ],

  "martial": [
    "mercenaire", "chevalier", "écuyer", "maître d'armes",
    "gladiateur", "archer", "garde du corps", "vétéran",
    "chasseur de primes", "duelliste"
  ],

  "explorer": [
    "aventurier", "explorateur", "cartographe", "guide",
    "chasseur de trésors", "archéologue", "pilleur de tombes",
    "éclaireur", "vagabond", "ermite"
  ]
}

Assure-toi que sw-npc generate --occupation=arcane fonctionne.
```

---

### MED-05: Résistances/Faiblesses des Monstres

**Statut**: [ ] À faire

**Problème**:
Les monstres ont des capacités spéciales en texte mais pas structurées :
- Vampire : "destroyed by sunlight" en description
- Troll : "regeneration stopped by fire" en description
- Pas de champs `resistances`, `vulnerabilities`, `immunities`

**Impact**: Le système ne peut pas automatiser les interactions spéciales.

**Solution**:
Ajouter des champs structurés pour les résistances.

**Fichiers à modifier**:
- `data/monsters.json`
- `internal/monster/monster.go`

**Prompt**:
```
Ajoute les résistances et faiblesses aux monstres dans data/monsters.json :

Pour chaque monstre approprié, ajoute :

"resistances": ["type1", "type2"],     // dégâts réduits
"immunities": ["type1", "type2"],      // dégâts ignorés
"vulnerabilities": ["type1", "type2"], // dégâts doublés

Types de dégâts : "fire", "cold", "lightning", "acid", "poison", "necrotic", "radiant", "normal_weapons", "magic"

Exemples :

Troll :
"resistances": [],
"immunities": [],
"vulnerabilities": ["fire", "acid"],
"regeneration": 3,
"regeneration_blocked_by": ["fire", "acid"]

Vampire :
"resistances": ["normal_weapons"],
"immunities": ["poison", "necrotic"],
"vulnerabilities": ["radiant", "fire"],
"special_vulnerabilities": ["sunlight", "running_water", "stake_to_heart"]

Skeleton :
"resistances": ["piercing"],
"immunities": ["poison", "necrotic"],
"vulnerabilities": ["bludgeoning"]

Mets à jour internal/monster/monster.go pour parser ces nouveaux champs.
```

---

### MED-06: Export Journal Markdown

**Statut**: [ ] À faire

**Problème**:
Le journal d'aventure est stocké en JSON mais :
- Pas d'export formaté pour impression
- Pas d'intégration avec les images générées
- Format brut difficile à lire

**Impact**: Les joueurs ne peuvent pas avoir un beau récap de leurs aventures.

**Solution**:
Ajouter une commande d'export Markdown du journal.

**Fichiers à modifier**:
- `cmd/adventure/main.go`
- `internal/adventure/journal.go`

**Prompt**:
```
Ajoute l'export Markdown du journal dans SkillsWeaver :

1. Dans internal/adventure/journal.go :

   func (a *Adventure) ExportJournalMarkdown() string {
     // Génère un document Markdown formaté :
     // # Journal de [Nom Aventure]
     //
     // ## Session 1 - [Date]
     //
     // ### Événements
     // - [emoji] [timestamp] Description
     //
     // ### Résumé
     // [summary de la session]
     //
     // ---
     //
     // ## Session 2 - [Date]
     // ...
   }

2. Si des images existent dans data/adventures/[nom]/images/ :
   - Les inclure avec ![description](chemin_relatif)
   - Matcher par ID d'entrée de journal

3. Ajoute une commande :
   - sw-adventure export-journal "Aventure" > journal.md
   - sw-adventure export-journal "Aventure" --output=journal.md

4. Format avec :
   - Emojis pour les types d'événements
   - Timestamps relatifs ("Il y a 2 heures" ou date absolue)
   - Séparation claire entre sessions
```

---

### MED-07: Giant Bat Taille Incorrecte

**Statut**: [x] Résolu - Valeurs correctes selon BFRPG

**Problème**:
Dans `monsters.json`, la chauve-souris géante est "medium" :
```json
{
  "id": "giant_bat",
  "size": "medium",
  "hit_dice": "2d8"
}
```

~~Une chauve-souris, même géante, devrait être "small" avec 1d8.~~

**Résolution**:
Après vérification contre le fichier Excel BFRPG officiel (Mike Roop r3) et le bestiaire web, les valeurs actuelles sont **correctes** :
- HD: 2d8 (conforme BFRPG)
- Size: medium (envergure 4.5m selon description BFRPG)
- Save: Fighter 2
- XP: 75

Aucune modification nécessaire.

**Fichiers à modifier**:
- `data/monsters.json`

**Prompt**:
```
Corrige la chauve-souris géante dans data/monsters.json en prenant comme reference la page https://clayadavis.gitlab.io/osr-bestiary/bestiary/bfrpg/core/bat-and-bat-giant/

Modifie giant_bat :
- "size": "small" (au lieu de "medium")
- "hit_dice": "1d8" (au lieu de "2d8")
- "hit_points_avg": 4 (au lieu de 9)

Ou, si tu veux garder une version plus dangereuse, renomme en "dire_bat" :
- "id": "dire_bat"
- "name_fr": "Chauve-souris sinistre"
- "size": "medium"
- "hit_dice": "2d8"

Et crée une version normale :
- "id": "giant_bat"
- "name_fr": "Chauve-souris géante"
- "size": "small"
- "hit_dice": "1d8"
```

---

## Criticité BASSE

Ces améliorations sont nice-to-have et enrichiraient l'expérience.

---

### LOW-01: Traits de Personnalité Automatisés

**Statut**: [ ] À faire

**Problème**:
Les PNJ ont des traits mais les PJ n'en ont pas :
- Pas de génération de background
- Pas de traits de personnalité
- Pas de liens ou relations

**Impact**: Moins de profondeur roleplay pour les PJ.

**Solution**:
Ajouter une génération optionnelle de traits pour les PJ.

**Fichiers à modifier**:
- `internal/character/character.go`
- `data/character-traits.json` (nouveau)

**Prompt**:
```
Ajoute des traits de personnalité optionnels aux PJ :

1. Crée data/character-traits.json avec :
   - personality_traits: ["courageux", "prudent", "curieux", ...]
   - ideals: ["justice", "liberté", "pouvoir", ...]
   - bonds: ["famille", "mentor", "village natal", ...]
   - flaws: ["impulsif", "cupide", "naïf", ...]

2. Dans Character struct, ajoute :
   - PersonalityTrait string
   - Ideal string
   - Bond string
   - Flaw string

3. Dans sw-character create, ajoute :
   - --generate-traits : génère des traits aléatoires
   - --trait="courageux" : spécifie manuellement

4. L'agent character-creator peut proposer de générer ces traits
```

---

### LOW-02: Scénarios Prédéfinis

**Statut**: [ ] À faire

**Problème**:
Le MJ doit tout créer de zéro :
- Pas de modules d'aventure inclus
- Pas d'exemples de donjons
- Pas de PNJ pré-créés

**Impact**: Charge de travail importante pour commencer à jouer.

**Solution**:
Créer un ou deux scénarios d'introduction.

**Fichiers à créer**:
- `data/scenarios/starter-dungeon/` (nouveau dossier)

**Prompt**:
```
Crée un scénario d'introduction dans data/scenarios/starter-dungeon/ :

Structure :
data/scenarios/starter-dungeon/
├── adventure.json     # Métadonnées du scénario
├── introduction.md    # Texte d'accroche pour le MJ
├── rooms.json         # Description des salles
├── encounters.json    # Rencontres prédéfinies
├── npcs.json          # PNJ du scénario
└── treasure.json      # Trésors spécifiques

Scénario : "La Crypte du Gobelin Roi"
- Niveau recommandé : 1-2
- Durée : 2-3 sessions
- 5-8 salles
- Boss : Gobelin Roi (goblin avec stats améliorées)
- Trésor final : arme magique +1 et 200 po

Inclus :
- 2-3 PNJ (donneur de quête, prisonnier à sauver)
- 3-4 types de rencontres
- 1 piège
- 1 énigme simple
```

---

### LOW-03: Gestion des Compagnons

**Statut**: [ ] À faire

**Problème**:
Les personnages ne peuvent pas avoir :
- Familiers (magiciens)
- Animaux de compagnie
- Henchmen (suivants)

**Impact**: Options de gameplay limitées.

**Solution**:
Ajouter une structure Companion au personnage.

**Fichiers à modifier**:
- `internal/character/character.go`

**Prompt**:
```
Ajoute un système de compagnons aux personnages :

1. Dans internal/character/character.go :

   type Companion struct {
     Name       string
     Type       string  // "familiar", "animal", "henchman"
     HP         int
     MaxHP      int
     AC         int
     Attack     string  // ex: "1d4"
     Special    string  // capacité spéciale
     Loyalty    int     // 1-12, pour henchmen
   }

   type Character struct {
     // ... existants
     Companions []Companion
   }

2. Ajoute des commandes :
   - sw-character add-companion "Aldric" "Rufus" --type=animal
   - sw-character list-companions "Aldric"
   - sw-character remove-companion "Aldric" "Rufus"

3. Règles BFRPG pour henchmen :
   - Jet de réaction initial
   - Morale basée sur CHA du maître
   - Partage de l'XP
```

---

### LOW-04: Challenge Rating des Rencontres

**Statut**: [ ] À faire

**Problème**:
Pas de système pour évaluer la difficulté d'une rencontre :
- Pas de CR sur les monstres
- Pas de calcul de difficulté vs niveau du groupe
- MJ doit deviner si c'est équilibré

**Impact**: Rencontres potentiellement déséquilibrées.

**Solution**:
Ajouter un système simple de difficulté.

**Fichiers à modifier**:
- `data/monsters.json`
- `internal/monster/monster.go`

**Prompt**:
```
Ajoute un système de difficulté des rencontres :

1. Dans monsters.json, ajoute pour chaque monstre :
   "challenge_rating": 0.25  // 1/4, 1/2, 1, 2, 3, etc.

   Règle simplifiée basée sur DV :
   - 1/4 DV ou moins : CR 0.125
   - 1/2 DV : CR 0.25
   - 1 DV : CR 0.5
   - 2 DV : CR 1
   - 3-4 DV : CR 2
   - 5-6 DV : CR 3
   - etc.

2. Dans internal/monster/monster.go :

   func CalculateEncounterDifficulty(monsters []Monster, partyLevel int, partySize int) string {
     // Retourne "Easy", "Medium", "Hard", "Deadly"
     // Basé sur CR total vs capacité du groupe
   }

3. Ajoute une commande :
   - sw-monster difficulty goblin goblin goblin orc --party-level=1 --party-size=4
   - Output: "Medium encounter for a party of 4 level 1 characters"
```

---

### LOW-05: Système de Pièges

**Statut**: [ ] À faire

**Problème**:
Aucun système de pièges :
- Pas de données de pièges
- Pas de règles de détection/désamorçage
- Voleurs ne peuvent pas utiliser leurs compétences

**Impact**: Élément classique du donjon manquant.

**Solution**:
Créer un système de pièges simple.

**Fichiers à créer**:
- `data/traps.json` (nouveau)
- `internal/trap/trap.go` (nouveau)

**Prompt**:
```
Crée un système de pièges pour SkillsWeaver :

1. Crée data/traps.json :

{
  "traps": [
    {
      "id": "pit_trap",
      "name_fr": "Fosse",
      "name_en": "Pit Trap",
      "detection_dc": 14,
      "disarm_dc": 12,
      "damage": "1d6",
      "damage_type": "falling",
      "description_fr": "Une trappe s'ouvre sous vos pieds...",
      "trigger": "pressure_plate"
    },
    {
      "id": "poison_needle",
      "name_fr": "Aiguille empoisonnée",
      "detection_dc": 16,
      "disarm_dc": 14,
      "damage": "1d4",
      "damage_type": "poison",
      "save": "poison",
      "save_dc": 12
    },
    // ... 8-10 pièges classiques
  ]
}

2. Crée internal/trap/trap.go avec :
   - func LoadTraps() []Trap
   - func (t *Trap) AttemptDetection(skill int) bool
   - func (t *Trap) AttemptDisarm(skill int) bool
   - func (t *Trap) Trigger() (damage int, effect string)

3. Ajoute CLI :
   - sw-trap list
   - sw-trap show pit_trap
   - sw-trap random
```

---

### LOW-06: Validation des Données au Démarrage

**Statut**: [x] Complété (22 déc 2025)

**Problème**:
Les fichiers JSON ne sont pas validés :
- Références croisées non vérifiées
- Erreurs silencieuses si données manquantes
- Pas d'avertissement si incohérence

**Impact**: Bugs difficiles à diagnostiquer en session.

**Solution**:
Créer une fonction de validation globale.

**Fichiers à modifier**:
- `internal/data/loader.go`
- `cmd/*/main.go`

**Résolution**:
Implémenté un système de validation complet :

1. **internal/data/loader.go** :
   - `ValidationError` struct avec File, Field, Message, Severity
   - `ValidateGameData()` fonction principale
   - `validateRaceClassRefs()` valide les classes autorisées (supporte multi-classe)
   - `validateStartingEquipment()` valide les références d'équipement
   - `isValidClassRef()` gère la notation multi-classe (ex: "fighter/magic-user")
   - `itemExists()` vérifie l'existence des items

2. **cmd/validate/main.go** - Nouveau CLI sw-validate :
   - Validation de races.json (allowed_classes)
   - Validation de equipment.json (starting_equipment)
   - Validation de monsters.json (treasure_type A-U ou 'none')
   - Validation de names.json (couverture des races)
   - Validation de spells.json (spell_lists)
   - Output texte ou JSON (--json)
   - Code de sortie 0 (succès) ou 1 (erreurs)

3. **Corrections de données** :
   - monsters.json: kobold treasure_type "P, Q" → "P"
   - monsters.json: gelatinous_cube treasure_type "V" → "none"

Note: L'option --validate pour les autres CLI n'a pas été implémentée car sw-validate couvre le besoin de validation.

**Prompt**:
```
Implémente une validation des données au démarrage :

1. Dans internal/data/loader.go :

   func ValidateGameData(gd *GameData) []ValidationError {
     var errors []ValidationError

     // Vérifie que chaque starting_equipment référence des items existants
     // Vérifie que chaque race a des classes valides
     // Vérifie que chaque monstre a un treasure_type valide
     // Vérifie que les noms couvrent toutes les races
     // etc.

     return errors
   }

   type ValidationError struct {
     File    string
     Field   string
     Message string
     Severity string // "error", "warning"
   }

2. Ajoute une commande :
   - sw-validate
   - Affiche tous les problèmes trouvés

3. Option pour les autres CLI :
   - --validate : valide avant d'exécuter
   - Affiche warnings mais continue si seulement warnings
```

---

## Statistiques

| Criticité | Nombre | % du total |
|-----------|--------|------------|
| CRITIQUE | 5 | 16% |
| HAUTE | 8 | 26% |
| MOYENNE | 7 | 23% |
| BASSE | 6 | 19% |
| **TOTAL** | **26** | 100% |

## Ordre de Résolution Suggéré

1. **Sprint 1 - Fondations** : CRIT-01 à CRIT-05
2. **Sprint 2 - Gameplay Core** : HIGH-01, HIGH-03, HIGH-04, HIGH-05
3. **Sprint 3 - Complétion** : HIGH-02, HIGH-06, HIGH-07, HIGH-08
4. **Sprint 4 - Polish** : MED-01 à MED-07
5. **Sprint 5 - Enrichissement** : LOW-01 à LOW-06

---

*Document généré le 21 décembre 2025*
