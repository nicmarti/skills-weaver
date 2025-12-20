> Oui commençons par le bestiaire. \
> Est-ce qu'un sous-agent avec un modele simple serait la bonne solution? Ou penses-tu qu'il faut créer des
skills ? 
> Voyons les 2 approches et je te dirai ce que je souhaite.

⏺ Excellente question ! Voici les deux approches comparées.

  ---
Approche 1 : Sous-agent "Bestiary Keeper"

Fonctionnement

Un agent conversationnel qui répond aux questions sur les monstres.

User: "Quels monstres pour un groupe niveau 2 ?"
Agent: "Je recommande 3-4 gobelins (15 PV chacun, CA 14)..."

Avantages

- Flexible : peut répondre à des questions complexes
- Contextuel : peut adapter les rencontres à la situation narrative
- Créatif : peut inventer des variantes ou combiner des monstres

Inconvénients

- Coût API : chaque requête consomme des tokens
- Lent : ~2-3 secondes par réponse
- Non-déterministe : peut donner des stats différentes à chaque fois
- Pas de persistance : données en mémoire uniquement

Fichiers

.claude/agents/bestiary-keeper.md   # Prompt avec données intégrées

  ---
Approche 2 : Skill + CLI (comme les autres)

Fonctionnement

Un outil Go avec données JSON, utilisable par Claude ou directement.

./monster show goblin              # Fiche complète
./monster encounter --level=2      # Génère une rencontre
./monster list --type=undead       # Liste les morts-vivants

Avantages

- Rapide : instantané, pas d'appel API
- Déterministe : stats toujours cohérentes
- Intégré : utilisable par le dungeon-master agent
- Persistant : données dans data/monsters.json
- Extensible : facile d'ajouter des monstres

Inconvénients

- Rigide : ne répond qu'aux commandes prévues
- Développement : plus de code à écrire

Fichiers

data/monsters.json                      # Données des monstres
internal/monster/monster.go             # Package Go
cmd/monster/main.go                     # CLI
.claude/skills/monster-manual/SKILL.md  # Skill

  ---
Comparaison

| Critère        | Sous-agent | Skill + CLI   |
  |----------------|------------|---------------|
| Vitesse        | ~2-3s      | Instantané    |
| Coût           | Tokens API | Gratuit       |
| Cohérence      | Variable   | Garantie      |
| Flexibilité    | Haute      | Moyenne       |
| Intégration MJ | Directe    | Via commandes |
| Effort dev     | Faible     | Moyen         |

  ---
Ma recommandation

Skill + CLI (Approche 2) car :
1. Les stats de monstres doivent être fixes et fiables (pas d'hallucination sur les PV)
2. Le dungeon-master agent peut appeler ./monster show goblin quand il en a besoin
3. Cohérent avec l'architecture existante (dice, character, npc...)
4. Permet de générer des rencontres équilibrées algorithmiquement

Le sous-agent serait utile pour du lore (histoires, comportements), mais pas pour les stats de combat.

