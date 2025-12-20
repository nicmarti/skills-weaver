# SkillsWeaver - Moteur de Jeu de Rôle avec Claude Code

## Description

**SkillsWeaver** est un moteur de jeu de rôle interactif basé sur les règles de **Basic Fantasy RPG** (BFRPG), orchestré par Claude Code. Il utilise des skills et des sous-agents pour gérer les différentes mécaniques du jeu.

Le préfixe `sw-` identifie toutes les commandes CLI du projet.

## But du Projet

Créer une expérience de jeu de rôle complète où Claude Code agit comme :
- **Lanceur de dés** via des scripts Go
- **Créateur de personnages** guidant le joueur
- **Maître du Jeu** pour les sessions de jeu

## Structure du Projet

```
skillsweaver/
├── .claude/
│   ├── skills/              # Skills Claude Code
│   │   ├── dice-roller/     # Lancer de dés
│   │   ├── character-generator/ # Création de personnages
│   │   ├── adventure-manager/   # Gestion des aventures
│   │   ├── name-generator/      # Génération de noms
│   │   ├── npc-generator/       # Génération de PNJ
│   │   ├── image-generator/     # Génération d'images
│   │   ├── monster-manual/      # Bestiaire
│   │   └── treasure-generator/  # Génération de trésors
│   └── agents/              # Sous-agents spécialisés
│       ├── character-creator.md
│       ├── rules-keeper.md
│       └── dungeon-master.md
├── cmd/
│   ├── dice/                # CLI sw-dice
│   ├── character/           # CLI sw-character
│   ├── adventure/           # CLI sw-adventure
│   ├── names/               # CLI sw-names
│   ├── npc/                 # CLI sw-npc
│   ├── image/               # CLI sw-image
│   ├── monster/             # CLI sw-monster
│   └── treasure/            # CLI sw-treasure
├── internal/
│   ├── dice/                # Package lancer de dés
│   ├── data/                # Chargement données JSON
│   ├── character/           # Package personnages
│   ├── adventure/           # Package aventures/campagnes
│   ├── names/               # Package génération de noms
│   ├── npc/                 # Package génération de PNJ
│   ├── image/               # Package génération d'images
│   ├── monster/             # Package bestiaire
│   └── treasure/            # Package trésors
├── data/
│   ├── names.json           # Dictionnaires de noms
│   ├── npc-traits.json      # Traits pour les PNJ
│   ├── monsters.json        # Bestiaire BFRPG
│   ├── treasure.json        # Tables de trésors BFRPG
│   ├── characters/          # Personnages sauvegardés
│   ├── adventures/          # Aventures sauvegardées
│   └── images/              # Images générées
├── ai/                      # Documentation et plans
└── CLAUDE.md                # Ce fichier
```

## Outils Disponibles

### CLI sw-dice

Lancer des dés avec notation standard RPG :

```bash
# Compiler
go build -o sw-dice ./cmd/dice

# Utiliser
./sw-dice roll d20              # Lance 1d20
./sw-dice roll 2d6+3            # Lance 2d6, ajoute 3
./sw-dice roll 4d6kh3           # Lance 4d6, garde les 3 plus hauts
./sw-dice roll d20 --advantage  # Avantage (2d20, garde le plus haut)
./sw-dice stats                 # Génère 6 caractéristiques (4d6kh3)
./sw-dice stats --classic       # Méthode classique (3d6)
```

### Skill dice-roller

La skill `dice-roller` permet à Claude de lancer des dés automatiquement pendant une session. Elle est découverte automatiquement quand on parle de jets de dés.

### CLI sw-character

Créer et gérer des personnages BFRPG :

```bash
# Compiler
go build -o sw-character ./cmd/character

# Créer un personnage
./sw-character create "Aldric" --race=human --class=fighter
./sw-character create "Lyra" --race=elf --class=magic-user --method=classic

# Gérer
./sw-character list              # Liste tous les personnages
./sw-character show "Aldric"     # Affiche la fiche
./sw-character delete "Aldric"   # Supprime
./sw-character export "Aldric" --format=json
```

### Skill character-generator

La skill `character-generator` permet à Claude de créer des personnages en guidant le joueur étape par étape.

### CLI sw-adventure

Gérer des aventures et campagnes BFRPG :

```bash
# Compiler
go build -o sw-adventure ./cmd/adventure

# Créer une aventure
./sw-adventure create "La Mine Perdue" "Une aventure dans les montagnes"

# Gérer le groupe
./sw-adventure add-character "La Mine Perdue" "Aldric"
./sw-adventure party "La Mine Perdue"

# Sessions de jeu
./sw-adventure start-session "La Mine Perdue"
./sw-adventure log "La Mine Perdue" combat "Combat contre 3 gobelins"
./sw-adventure add-gold "La Mine Perdue" 50 "Trésor gobelin"
./sw-adventure end-session "La Mine Perdue" "Premier niveau exploré"

# Consulter
./sw-adventure status "La Mine Perdue"    # Statut complet
./sw-adventure journal "La Mine Perdue"   # Journal de l'aventure
./sw-adventure sessions "La Mine Perdue"  # Historique des sessions
./sw-adventure inventory "La Mine Perdue" # Inventaire partagé
```

### Skill adventure-manager

La skill `adventure-manager` permet à Claude de gérer les aventures, suivre les sessions et maintenir le journal automatique.

### CLI sw-names

Générer des noms de personnages fantasy :

```bash
# Compiler
go build -o sw-names ./cmd/names

# Générer des noms par race
./sw-names generate dwarf                    # Nom de nain
./sw-names generate elf --gender=f           # Nom d'elfe féminin
./sw-names generate human --count=5          # 5 noms humains
./sw-names generate halfling --first-only    # Prénom de halfelin

# Générer des noms de PNJ
./sw-names npc innkeeper                     # Nom de tavernier
./sw-names npc merchant                      # Nom de marchand
./sw-names npc villain                       # Nom de méchant

# Lister les options
./sw-names list                              # Toutes les options
```

### Skill name-generator

La skill `name-generator` permet à Claude de générer des noms pour les joueurs et les PNJ selon la race et le type.

### CLI sw-npc

Générer des PNJ complets :

```bash
# Compiler
go build -o sw-npc ./cmd/npc

# Générer un PNJ complet
./sw-npc generate                              # PNJ aléatoire
./sw-npc generate --race=dwarf --gender=m      # Nain masculin
./sw-npc generate --occupation=authority       # Figure d'autorité
./sw-npc generate --attitude=hostile           # PNJ hostile

# Génération rapide
./sw-npc quick --count=5                       # 5 PNJ en une ligne

# Formats de sortie
./sw-npc generate --format=md                  # Markdown (défaut)
./sw-npc generate --format=json                # JSON
./sw-npc generate --format=short               # Une ligne
```

### Skill npc-generator

La skill `npc-generator` permet à Claude de créer des PNJ complets avec apparence, personnalité, motivations et secrets.

### CLI sw-image

Générer des images heroic fantasy via fal.ai FLUX.1 :

```bash
# Compiler
go build -o sw-image ./cmd/image

# Prérequis: variable d'environnement FAL_KEY
export FAL_KEY="votre_clé_fal_ai"

# Portrait de personnage existant
./sw-image character "Aldric" --style=epic

# Portrait de PNJ
./sw-image npc --race=dwarf --gender=m --occupation=skilled

# Scène d'aventure
./sw-image scene "Combat contre des gobelins" --type=battle

# Monstre
./sw-image monster dragon --style=dark_fantasy

# Objet magique
./sw-image item weapon "épée flamboyante"

# Lieu
./sw-image location dungeon "Les Mines Perdues"

# Prompt personnalisé
./sw-image custom "Un groupe d'aventuriers dans une taverne"

# Lister les options
./sw-image list
```

### Skill image-generator

La skill `image-generator` permet à Claude de générer des illustrations fantasy pour enrichir l'expérience de jeu : portraits, scènes, monstres, objets et lieux.

### CLI sw-monster

Consulter le bestiaire et générer des rencontres :

```bash
# Compiler
go build -o sw-monster ./cmd/monster

# Consulter un monstre
./sw-monster show goblin              # Fiche complète
./sw-monster show dragon_red_adult    # Dragon rouge adulte
./sw-monster search undead            # Recherche par type

# Lister les monstres
./sw-monster list                     # Tous les monstres
./sw-monster list --type=humanoid    # Par type
./sw-monster types                    # Types disponibles

# Générer une rencontre
./sw-monster encounter dungeon_level_1  # Niveau 1
./sw-monster encounter --level=3        # Par niveau de groupe
./sw-monster encounter forest           # En forêt

# Créer des ennemis avec PV
./sw-monster roll orc --count=4       # 4 orcs avec PV aléatoires
./sw-monster roll goblin --count=6    # 6 gobelins
```

### Skill monster-manual

La skill `monster-manual` permet à Claude de consulter les stats des monstres et générer des rencontres équilibrées pendant les sessions de jeu.

### CLI sw-treasure

Générer des trésors selon les tables BFRPG :

```bash
# Compiler
go build -o sw-treasure ./cmd/treasure

# Générer un trésor
./sw-treasure generate R              # Trésor type R (Gobelin)
./sw-treasure generate A              # Trésor type A (Dragon)
./sw-treasure generate B --count=3    # 3 trésors type B

# Lister les types de trésors
./sw-treasure types                   # Tous les types A-U

# Détails d'un type
./sw-treasure info A                  # Probabilités du type A

# Lister les objets magiques
./sw-treasure items                   # Catégories disponibles
./sw-treasure items potions           # Toutes les potions
./sw-treasure items weapons           # Armes magiques
./sw-treasure items armor             # Armures magiques
```

### Skill treasure-generator

La skill `treasure-generator` permet à Claude de générer des trésors appropriés après les combats, en respectant les types de trésors assignés aux monstres.

## Sous-Agents Spécialisés

Les agents sont disponibles dans `.claude/agents/` :

### character-creator
Guide interactif pour créer des personnages étape par étape. Explique les races, classes, et aide à faire des choix cohérents.

### rules-keeper
Référence rapide des règles BFRPG. Répond aux questions sur le combat, la magie, les jets de sauvegarde et arbitre les situations.

### dungeon-master
Maître du Jeu complet. Narration immersive, gestion des rencontres, incarnation des PNJ, et tracking automatique via les commandes sw-adventure.

## Règles BFRPG

### Races Disponibles

| Race | Modificateurs | Classes Autorisées |
|------|--------------|-------------------|
| Humain | Aucun | Toutes |
| Elfe | +1 DEX, -1 CON | Guerrier (6), Magicien (9), Voleur |
| Nain | +1 CON, -1 CHA | Guerrier (7), Clerc (6), Voleur |
| Halfelin | +1 DEX, -1 FOR | Guerrier (4), Voleur |

### Classes Disponibles

| Classe | Dé de Vie | Armes | Armures |
|--------|-----------|-------|---------|
| Guerrier | d8 | Toutes | Toutes |
| Clerc | d6 | Contondantes | Toutes |
| Magicien | d4 | Dague, bâton | Aucune |
| Voleur | d4 | Toutes | Cuir |

## Commandes de Développement

```bash
# Compiler tous les outils SkillsWeaver
go build -o sw-dice ./cmd/dice
go build -o sw-character ./cmd/character
go build -o sw-adventure ./cmd/adventure
go build -o sw-names ./cmd/names
go build -o sw-npc ./cmd/npc
go build -o sw-image ./cmd/image
go build -o sw-monster ./cmd/monster
go build -o sw-treasure ./cmd/treasure

# Lancer les tests
go test ./...

# Tester le système de dés
go test ./internal/dice/... -v

# Tester le chargement des données
go test ./internal/data/... -v

# Tester le générateur de personnages
go test ./internal/character/... -v
```

## Plan d'Implémentation

Voir `ai/PLAN.md` pour le plan détaillé avec les phases :

1. **Phase 1** : Système de dés [TERMINEE]
2. **Phase 2** : Données BFRPG [TERMINEE]
3. **Phase 3** : Générateur de personnages [TERMINEE]
4. **Phase 3bis** : Gestionnaire d'aventures [TERMINEE]
5. **Phase 4** : Sous-agents spécialisés [TERMINEE]
6. **Phase 4bis** : Générateur de noms [TERMINEE]
7. **Phase 5** : Générateur de PNJ [TERMINEE]
8. **Phase 6** : Générateur d'images [TERMINEE]
9. **Phase 7** : Bestiaire BFRPG [TERMINEE]
10. **Phase 8** : Tables de trésors [TERMINEE]

## Ressources

- [Basic Fantasy RPG](https://www.basicfantasy.org/) - Règles complètes (gratuit)
- [SRD BFRPG](https://www.basicfantasy.org/srd/) - System Reference Document
