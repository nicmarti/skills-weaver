# Map Prompt Generation Guidelines

## Purpose

Generate detailed French prompts for 2D fantasy map creation using fal.ai flux-2. These prompts will be used to create visual maps for tabletop RPG sessions.

## Target Audience

The maps are for:
- Dungeon Masters (narration and campaign tracking)
- Players (visual reference during gameplay)
- Campaign documentation (world-building)

## Template Structures by Map Type

### City Maps

**Structure**: `[Perspective] + [Kingdom Style] + [Geographic Context] + [Districts] + [POIs] + [Infrastructure] + [Realism]`

**Example**:
```
Cette carte montre la ville portuaire de Cordova en vue aérienne. Style architectural valdorin maritime avec influences italiennes, bâtiments aux toits bleus et or. La ville est au bord de la Mer Alyséenne, un fleuve se jette dans la mer. Le sud est bordé de falaises, l'ouest s'ouvre sur des plaines. Quartiers : port commercial avec docks et entrepôts, district marchand avec places et boutiques, quartier résidentiel avec maisons colorées. Points d'intérêt : La Taverne du Voile Écarlate (bâtiment rouge vif), les Docks Marchands (quais animés), L'Auberge du Crâne d'Or (édifice doré), Villa de Valorian (manoir en périphérie). Infrastructure : phare, chantier naval, murailles côtières. La ville a une forme organique suivant le relief, rues sinueuses, chemins cohérents reliant les quartiers, places publiques naturelles.
```

**Key Elements**:
- Perspective: Always "vue aérienne" (aerial view)
- Kingdom style: Match architectural theme (valdorin/karvath/lumenciel/astrene)
- Geography: Coastline, rivers, hills, plains - use cardinal directions
- Districts: 3-4 major quarters with purpose
- POIs: Name specific buildings/landmarks from game data
- Infrastructure: Ports, walls, gates, market squares
- Realism: Organic shape, natural street layout, coherent paths

### Regional Maps

**Structure**: `[Bird's Eye View] + [Kingdom Territory] + [Geographic Features] + [Settlements] + [Roads] + [Terrain] + [Borders] + [Cartographic Elements]`

**Example**:
```
Cette carte montre la région de la Côte Occidentale vue du ciel, style carte géographique. Territoire du Royaume de Valdorine. Région côtière prospère dominée par le commerce maritime. Villes principales : Cordova (capitale, grand port), Port-de-Lune (chantiers navals), Havre-d'Argent (centre financier). Terrain : côte accidentée au sud avec falaises, plaines fertiles à l'ouest, collines boisées au nord. Routes commerciales terrestres reliant les trois villes, avec chemins secondaires vers villages. Route maritime longeant la côte. Frontières du royaume Valdorine clairement marquées. Échelle large montrant 200 km. Avec légende cartographique, symboles géographiques (montagnes triangulaires, rivières bleues, forêts vertes), style carte médiévale fantasy avec parchemin vieilli.
```

**Key Elements**:
- Scale: Specify region size (50-500 km)
- Multiple settlements: Show cities, towns, villages
- Trade routes: Land and sea connections
- Terrain variety: Mountains, forests, plains, coasts
- Borders: Kingdom boundaries
- Cartographic style: Medieval fantasy map aesthetic
- Legend: Symbols for different geographic features

### Dungeon Maps

**Structure**: `[Top-Down] + [Level] + [Room Layout] + [Corridors] + [Hazards] + [Architecture] + [Grid]`

**Example**:
```
Plan de donjon en vue du dessus : La Crypte des Ombres, Niveau 1. Salles de différentes tailles : grande salle d'entrée (10m x 15m), crypte centrale (12m x 12m), chapelle latérale (6m x 8m), salles de garde (4m x 6m chacune). Couloirs étroits (2m de large) reliant les salles. Portes en bois renforcé clairement indiquées. Pièges marqués : fosse dissimulée (symbole X), lames pendulaires (symbole !), dalle piégée (symbole triangle). Portes secrètes en lignes pointillées. Escaliers vers niveau 2 au nord-est. Architecture de pierre médiévale naine, piliers carrés dans la crypte centrale, torches fixées aux murs tous les 5m. Grille au sol avec carrés de 1,5m pour placement des figurines. Échelle indiquée en mètres. Style plan de D&D classique, noir et blanc avec ombrage gris pour profondeur des salles.
```

**Key Elements**:
- Level specification: Which floor (1, 2, 3, etc.)
- Room dimensions: Specific measurements
- Corridors: Width and connections
- Doors: Regular, reinforced, secret (dotted lines)
- Traps: Marked with symbols (X, !, triangle)
- Stairs: Connections to other levels
- Architecture: Stone, wood, pillars, torches
- Grid: 1.5m squares (D&D standard)
- Style: Black & white, isometric shading

### Tactical Battle Maps

**Structure**: `[Top-Down Grid] + [Terrain] + [Terrain Features] + [Cover] + [Obstacles] + [Elevation] + [Tactical Elements]`

**Example**:
```
Carte tactique de combat en vue du dessus. Combat dans une clairière forestière. Terrain : forêt dense avec clairière centrale (15m x 20m). Arbres denses sur les bords (couverture totale), sous-bois épais (couverture partielle), clairière dégagée au centre. Éléments de couverture : gros rochers (3m de diamètre, couverture totale), souches d'arbres (1.5m, couverture partielle), rondins tombés (obstacles bas). Ruisseau traversant la carte d'est en ouest (1m de large, terrain difficile). Variations d'élévation : berges du ruisseau (-0.5m), butte au nord (+2m, avantage de hauteur). Grille de combat avec carrés de 1.5m. Format carré 20x20 cases pour alignement des figurines. Style carte de combat D&D, couleurs distinctes : vert foncé pour forêt, vert clair pour clairière, bleu pour eau, brun pour terre. Lisible et pratique pour le jeu.
```

**Key Elements**:
- Grid: Always 1.5m squares
- Terrain type: Forest, mountain, plains, swamp, etc.
- Cover levels: Total, partial, none
- Obstacles: Rocks, trees, walls, debris
- Terrain types: Difficult (mud, water), normal, advantageous (high ground)
- Elevation: Mark height differences
- Format: Square (20x20 or similar) for miniature placement
- Colors: Distinct zones for clarity
- Practical: Optimized for gameplay

## Length Guidelines

**Target**: 100-200 words per prompt

- **Minimum**: 80 words (ensure sufficient detail)
- **Maximum**: 250 words (avoid overwhelming the image generator)
- **Sweet spot**: 150 words (balance detail and clarity)

## Best Practices

### Geographic Precision
- Use cardinal directions (nord, sud, est, ouest)
- Specify distances when relevant (200 km, 50m, etc.)
- Describe relative positions clearly

### Kingdom Architectural Styles

Match building styles to kingdom:
- **Valdorine**: Maritime, Italian-inspired, blue/gold colors, port architecture
- **Karvath**: Militaristic, Germanic fortresses, red/black/steel colors, walls and towers
- **Lumenciel**: Religious, Latin sacred architecture, white/gold, cathedrals and monasteries
- **Astrène**: Melancholic, Nordic influences, gray/silver, weathered stone

### Natural & Organic Layouts
- Cities: Irregular shapes following terrain
- Streets: Curved, not grid-like (except military fortresses)
- Growth patterns: Logical development from center
- Avoid: Perfect symmetry, rigid grids, monotone layouts

### POI Integration
- Name specific buildings from geography.json
- Add descriptive details (colors, sizes, unique features)
- Position logically (tavern near docks, villa in periphery, etc.)

### Cartographic Language
- Use proper map terminology
- Include scale indicators
- Add legend descriptions
- Mention map style explicitly (medieval fantasy, parchment, etc.)

### French Typography
- Add space before colons: "Terrain : forêt"
- Add space before semicolons: "salles ; couloirs"
- Use proper accents: "château", "forêt", "rivière"

## Translation from Game Data

When enriching prompts from geography.json:

### Location Data
```json
{
  "name": "Cordova",
  "type": "port majeur",
  "description": "Ville portuaire scintillante...",
  "key_locations": ["Taverne du Voile Écarlate", "Docks"]
}
```

**Becomes**:
```
Cette carte montre la ville portuaire de Cordova... Points d'intérêt : La Taverne du Voile Écarlate (taverne populaire), les Docks Marchands (quais animés)...
```

### Kingdom Data
```json
{
  "id": "valdorine",
  "colors": ["bleu", "or"],
  "motto": "L'argent n'a pas d'odeur"
}
```

**Injects**:
```
Style architectural valdorin maritime avec bâtiments aux toits bleus et or...
```

### Distance Data
```json
{
  "distances": {
    "Pierrebrune": "3 jours à pied (90-100 km)"
  }
}
```

**For regional maps**:
```
Cordova et Pierrebrune distants de 90 km, reliés par route commerciale...
```

## Common Mistakes to Avoid

❌ **Too Abstract**: "A beautiful city by the sea"
✅ **Specific**: "Ville portuaire de Cordova avec docks au sud, quartier marchand au centre, phare à l'ouest"

❌ **No Context**: "Map of a dungeon"
✅ **Contextual**: "Plan de la Crypte des Ombres, niveau 1, architecture naine avec salles carrées et piliers massifs"

❌ **Monotone**: "City with square grid, uniform buildings"
✅ **Organic**: "Ville avec rues sinueuses suivant le relief, bâtiments de tailles variées, places publiques naturelles"

❌ **Missing Scale**: "Large regional map"
✅ **Scaled**: "Carte régionale, échelle 200 km, montrant 3 villes majeures et routes"

❌ **Generic Style**: "Fantasy map"
✅ **Styled**: "Carte style parchemin médiéval valdorin, avec rose des vents, légende cartographique, couleurs bleu et or"

❌ **No Kingdom Identity**: "Medieval city"
✅ **Kingdom-specific**: "Cité karvath militarisée avec murailles germaniques, tours de guet tous les 50m, architecture martiale rouge et noir"

## Quality Checklist

Before finalizing a map prompt, verify:

- [ ] 100-200 words length
- [ ] Perspective clearly stated (aerial, top-down, bird's eye)
- [ ] Kingdom architectural style mentioned (if applicable)
- [ ] Geographic context with cardinal directions
- [ ] 3+ specific POIs named (for city maps)
- [ ] Infrastructure details (ports, walls, roads, etc.)
- [ ] Organic/natural layout instructions
- [ ] Scale or dimensions specified
- [ ] Cartographic style described
- [ ] French typography correct (spaces before :;)
- [ ] No English words (except proper names if needed)
- [ ] Coherent and readable flow

## Output Format

The enriched prompt should be:
- **Single paragraph** (no line breaks mid-prompt)
- **Flowing text** (use periods and commas, not bullet lists)
- **Complete sentences**
- **Proper French grammar and typography**

## Examples by Type

### City Map (Valdorine)
```
Cette carte montre la ville portuaire de Cordova en vue aérienne, capitale du Royaume de Valdorine. Style architectural valdorin maritime avec influences italiennes, bâtiments aux façades colorées et toits en tuiles bleues et or. La ville s'étend le long de la Mer Alyséenne, avec un fleuve se jetant dans la mer au centre. Géographie : côte au sud et à l'est, fleuve traversant d'ouest en est, falaises au sud, plaines agricoles à l'ouest, collines boisées au nord. Quartiers : port commercial avec docks et entrepôts (sud-est), district marchand avec places et boutiques (centre), quartier résidentiel avec maisons colorées (nord), quartier noble avec villas (ouest). Points d'intérêt : La Taverne du Voile Écarlate (bâtiment rouge vif près des docks), les Docks Marchands (quais animés avec navires), L'Auberge du Crâne d'Or (édifice doré à trois étages), Manoir de Valorian le Doré (villa luxueuse en périphérie ouest), Chantier Naval (installations industrielles sud). Infrastructure : grand phare à l'entrée du port, murailles côtières, marché central avec fontaine, pont de pierre sur le fleuve, routes principales pavées. La ville a une forme organique suivant le relief et la côte, rues sinueuses dans le vieux quartier, avenues plus larges dans le quartier marchand, chemins cohérents reliant les différents quartiers, places publiques naturelles aux intersections importantes. Style carte fantasy détaillée, vue isométrique légère, couleurs vives pour l'ambiance méditerranéenne valdorine.
```

### Tactical Map (Forest)
```
Carte tactique de combat en vue du dessus pour une embuscade en forêt dense. Terrain : forêt mixte avec clairière centrale ovale (12m x 18m). Arbres denses sur tout le pourtour (pins et chênes, 2-3m de diamètre, couverture totale), sous-bois épais avec fougères (couverture partielle, terrain difficile), clairière avec herbe rase au centre (terrain normal). Éléments de couverture : trois gros rochers (3-4m de diamètre, couverture totale, peut grimper dessus), souches d'arbres abattus (1.5m, couverture partielle), rondins tombés (obstacles bas, peut sauter par-dessus). Ruisseau serpentant du nord-ouest au sud-est (1m de large, profond de 50cm, terrain difficile pour traverser). Petit pont de bois au centre (2m de large, peut être détruit). Variations d'élévation : berges du ruisseau en contrebas (-0.5m), butte rocheuse au nord (+2m, avantage de hauteur pour archers), pente douce à l'ouest. Grille de combat avec carrés de 1.5m clairement visible. Format carré 20x20 cases pour placement optimal des figurines. Style carte de combat D&D, couleurs distinctes : vert foncé pour forêt dense, vert moyen pour sous-bois, vert clair pour clairière, bleu pour eau, gris pour rochers, brun pour terre. Symboles tactiques : flèches pour pentes, lignes pour élévation. Lisible, pratique pour le jeu, zones de mouvement claires.
```

