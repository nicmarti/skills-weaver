# Dungeons - Moteur de Jeu de Rôle avec Claude Code

## Description

Ce projet est un moteur de jeu de rôle interactif basé sur les règles de **Basic Fantasy RPG** (BFRPG), orchestré par Claude Code. Il utilise des skills et des sous-agents pour gérer les différentes mécaniques du jeu.

## But du Projet

Créer une expérience de jeu de rôle complète où Claude Code agit comme :
- **Lanceur de dés** via des scripts Go
- **Créateur de personnages** guidant le joueur
- **Maître du Jeu** pour les sessions de jeu

## Structure du Projet

```
dungeons/
├── .claude/
│   ├── skills/              # Skills Claude Code
│   │   ├── dice-roller/     # Lancer de dés
│   │   ├── character-generator/ # Création de personnages
│   │   ├── adventure-manager/   # Gestion des aventures
│   │   ├── name-generator/      # Génération de noms
│   │   ├── npc-generator/       # Génération de PNJ
│   │   ├── image-generator/     # Génération d'images
│   │   └── monster-manual/      # Bestiaire
│   └── agents/              # Sous-agents spécialisés
│       ├── character-creator.md
│       ├── rules-keeper.md
│       └── dungeon-master.md
├── cmd/
│   ├── dice/                # CLI pour les dés
│   ├── character/           # CLI pour les personnages
│   ├── adventure/           # CLI pour les aventures
│   ├── names/               # CLI pour les noms
│   ├── npc/                 # CLI pour les PNJ
│   ├── image/               # CLI pour les images
│   └── monster/             # CLI pour le bestiaire
├── internal/
│   ├── dice/                # Package lancer de dés
│   ├── data/                # Chargement données JSON
│   ├── character/           # Package personnages
│   ├── adventure/           # Package aventures/campagnes
│   ├── names/               # Package génération de noms
│   ├── npc/                 # Package génération de PNJ
│   ├── image/               # Package génération d'images
│   └── monster/             # Package bestiaire
├── data/
│   ├── names.json           # Dictionnaires de noms
│   ├── npc-traits.json      # Traits pour les PNJ
│   ├── monsters.json        # Bestiaire BFRPG
│   ├── characters/          # Personnages sauvegardés
│   ├── adventures/          # Aventures sauvegardées
│   └── images/              # Images générées
├── ai/                      # Documentation et plans
└── CLAUDE.md                # Ce fichier
```

## Outils Disponibles

### CLI Dice

Lancer des dés avec notation standard RPG :

```bash
# Compiler
go build -o dice ./cmd/dice

# Utiliser
./dice roll d20              # Lance 1d20
./dice roll 2d6+3            # Lance 2d6, ajoute 3
./dice roll 4d6kh3           # Lance 4d6, garde les 3 plus hauts
./dice roll d20 --advantage  # Avantage (2d20, garde le plus haut)
./dice stats                 # Génère 6 caractéristiques (4d6kh3)
./dice stats --classic       # Méthode classique (3d6)
```

### Skill dice-roller

La skill `dice-roller` permet à Claude de lancer des dés automatiquement pendant une session. Elle est découverte automatiquement quand on parle de jets de dés.

### CLI Character

Créer et gérer des personnages BFRPG :

```bash
# Compiler
go build -o character ./cmd/character

# Créer un personnage
./character create "Aldric" --race=human --class=fighter
./character create "Lyra" --race=elf --class=magic-user --method=classic

# Gérer
./character list              # Liste tous les personnages
./character show "Aldric"     # Affiche la fiche
./character delete "Aldric"   # Supprime
./character export "Aldric" --format=json
```

### Skill character-generator

La skill `character-generator` permet à Claude de créer des personnages en guidant le joueur étape par étape.

### CLI Adventure

Gérer des aventures et campagnes BFRPG :

```bash
# Compiler
go build -o adventure ./cmd/adventure

# Créer une aventure
./adventure create "La Mine Perdue" "Une aventure dans les montagnes"

# Gérer le groupe
./adventure add-character "La Mine Perdue" "Aldric"
./adventure party "La Mine Perdue"

# Sessions de jeu
./adventure start-session "La Mine Perdue"
./adventure log "La Mine Perdue" combat "Combat contre 3 gobelins"
./adventure add-gold "La Mine Perdue" 50 "Trésor gobelin"
./adventure end-session "La Mine Perdue" "Premier niveau exploré"

# Consulter
./adventure status "La Mine Perdue"    # Statut complet
./adventure journal "La Mine Perdue"   # Journal de l'aventure
./adventure sessions "La Mine Perdue"  # Historique des sessions
./adventure inventory "La Mine Perdue" # Inventaire partagé
```

### Skill adventure-manager

La skill `adventure-manager` permet à Claude de gérer les aventures, suivre les sessions et maintenir le journal automatique.

### CLI Names

Générer des noms de personnages fantasy :

```bash
# Compiler
go build -o names ./cmd/names

# Générer des noms par race
./names generate dwarf                    # Nom de nain
./names generate elf --gender=f           # Nom d'elfe féminin
./names generate human --count=5          # 5 noms humains
./names generate halfling --first-only    # Prénom de halfelin

# Générer des noms de PNJ
./names npc innkeeper                     # Nom de tavernier
./names npc merchant                      # Nom de marchand
./names npc villain                       # Nom de méchant

# Lister les options
./names list                              # Toutes les options
```

### Skill name-generator

La skill `name-generator` permet à Claude de générer des noms pour les joueurs et les PNJ selon la race et le type.

### CLI NPC

Générer des PNJ complets :

```bash
# Compiler
go build -o npc ./cmd/npc

# Générer un PNJ complet
./npc generate                              # PNJ aléatoire
./npc generate --race=dwarf --gender=m      # Nain masculin
./npc generate --occupation=authority       # Figure d'autorité
./npc generate --attitude=hostile           # PNJ hostile

# Génération rapide
./npc quick --count=5                       # 5 PNJ en une ligne

# Formats de sortie
./npc generate --format=md                  # Markdown (défaut)
./npc generate --format=json                # JSON
./npc generate --format=short               # Une ligne
```

### Skill npc-generator

La skill `npc-generator` permet à Claude de créer des PNJ complets avec apparence, personnalité, motivations et secrets.

### CLI Image

Générer des images heroic fantasy via fal.ai FLUX.1 :

```bash
# Compiler
go build -o image ./cmd/image

# Prérequis: variable d'environnement FAL_KEY
export FAL_KEY="votre_clé_fal_ai"

# Portrait de personnage existant
./image character "Aldric" --style=epic

# Portrait de PNJ
./image npc --race=dwarf --gender=m --occupation=skilled

# Scène d'aventure
./image scene "Combat contre des gobelins" --type=battle

# Monstre
./image monster dragon --style=dark_fantasy

# Objet magique
./image item weapon "épée flamboyante"

# Lieu
./image location dungeon "Les Mines Perdues"

# Prompt personnalisé
./image custom "Un groupe d'aventuriers dans une taverne"

# Lister les options
./image list
```

### Skill image-generator

La skill `image-generator` permet à Claude de générer des illustrations fantasy pour enrichir l'expérience de jeu : portraits, scènes, monstres, objets et lieux.

### CLI Monster

Consulter le bestiaire et générer des rencontres :

```bash
# Compiler
go build -o monster ./cmd/monster

# Consulter un monstre
./monster show goblin              # Fiche complète
./monster show dragon_red_adult    # Dragon rouge adulte
./monster search undead            # Recherche par type

# Lister les monstres
./monster list                     # Tous les monstres
./monster list --type=humanoid    # Par type
./monster types                    # Types disponibles

# Générer une rencontre
./monster encounter dungeon_level_1  # Niveau 1
./monster encounter --level=3        # Par niveau de groupe
./monster encounter forest           # En forêt

# Créer des ennemis avec PV
./monster roll orc --count=4       # 4 orcs avec PV aléatoires
./monster roll goblin --count=6    # 6 gobelins
```

### Skill monster-manual

La skill `monster-manual` permet à Claude de consulter les stats des monstres et générer des rencontres équilibrées pendant les sessions de jeu.

## Sous-Agents Spécialisés

Les agents sont disponibles dans `.claude/agents/` :

### character-creator
Guide interactif pour créer des personnages étape par étape. Explique les races, classes, et aide à faire des choix cohérents.

### rules-keeper
Référence rapide des règles BFRPG. Répond aux questions sur le combat, la magie, les jets de sauvegarde et arbitre les situations.

### dungeon-master
Maître du Jeu complet. Narration immersive, gestion des rencontres, incarnation des PNJ, et tracking automatique via les commandes adventure.

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
# Compiler tout
go build ./...

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

## Ressources

- [Basic Fantasy RPG](https://www.basicfantasy.org/) - Règles complètes (gratuit)
- [SRD BFRPG](https://www.basicfantasy.org/srd/) - System Reference Document
