#!/usr/bin/env python3
"""
Parse D&D 5e spells from sorts_et_magie.md and generate spells.json

Usage: python3 scripts/parse_spells_d5e.py
"""

import re
import json
import sys

# School mapping (French to English IDs)
SCHOOLS = {
    "Abjuration": "abjuration",
    "Divination": "divination",
    "Enchantement": "enchantment",
    "Évocation": "evocation",
    "Illusion": "illusion",
    "Invocation": "conjuration",  # French uses "Invocation" for Conjuration
    "Nécromancie": "necromancy",
    "Transmutation": "transmutation"
}

# Class mapping (French to English IDs)
CLASSES = {
    "Barde": "bard",
    "Clerc": "cleric",
    "Druide": "druid",
    "Ensorceleur": "sorcerer",
    "Guerrier": "fighter",
    "Magicien": "wizard",
    "Moine": "monk",
    "Occultiste": "warlock",
    "Paladin": "paladin",
    "Rôdeur": "ranger",
    "Roublard": "rogue",
    "Barbare": "barbarian"
}

# Level keywords (French)
LEVEL_KEYWORDS = {
    "sort mineur": 0,
    "mineur": 0,
    "1er niveau": 1,
    "2e niveau": 2,
    "3e niveau": 3,
    "4e niveau": 4,
    "5e niveau": 5,
    "6e niveau": 6,
    "7e niveau": 7,
    "8e niveau": 8,
    "9e niveau": 9
}

def parse_spell_header(line):
    """Extract spell name from header: # **Name**"""
    match = re.match(r'^#+ \*\*(.*?)\*\*$', line.strip())
    if match:
        return match.group(1).strip()
    return None

def parse_school_level_classes(line):
    """Parse: *School du Niveau (Classe1, Classe2)*"""
    line = line.strip()
    if not line.startswith('*') or not line.endswith('*'):
        return None, None, []

    # Remove asterisks
    content = line[1:-1].strip()

    # Extract school
    school = None
    for fr_school, en_school in SCHOOLS.items():
        if content.startswith(fr_school):
            school = en_school
            break

    # Extract level
    level = None
    for keyword, lvl in LEVEL_KEYWORDS.items():
        if keyword in content.lower():
            level = lvl
            break

    # Extract classes from parentheses
    classes = []
    match = re.search(r'\((.*?)\)', content)
    if match:
        class_str = match.group(1)
        for fr_class in class_str.split(','):
            fr_class = fr_class.strip()
            if fr_class in CLASSES:
                classes.append(CLASSES[fr_class])

    return school, level, classes

def parse_property(line, keyword):
    """Extract property value: **Keyword :** value"""
    pattern = rf'\*\*{re.escape(keyword)}\s*:\*\*\s*(.+)$'
    match = re.search(pattern, line)
    if match:
        return match.group(1).strip()
    return None

def parse_components(comp_str):
    """Parse components: V, S, M (materials) -> array + optional material"""
    components = []
    material = None

    # Extract V, S, M
    if 'V' in comp_str:
        components.append('V')
    if 'S' in comp_str:
        components.append('S')
    if 'M' in comp_str:
        components.append('M')
        # Extract material from parentheses
        match = re.search(r'\(([^)]+)\)', comp_str)
        if match:
            material = match.group(1).strip()

    return components, material

def create_spell_id(name):
    """Create spell ID from French name: Projectile magique -> projectile_magique"""
    spell_id = name.lower()
    # Remove accents
    replacements = {
        'à': 'a', 'â': 'a', 'ä': 'a',
        'é': 'e', 'è': 'e', 'ê': 'e', 'ë': 'e',
        'î': 'i', 'ï': 'i',
        'ô': 'o', 'ö': 'o',
        'ù': 'u', 'û': 'u', 'ü': 'u',
        'ç': 'c',
        'æ': 'ae', 'œ': 'oe'
    }
    for fr, en in replacements.items():
        spell_id = spell_id.replace(fr, en)

    # Replace spaces and special chars with underscores
    spell_id = re.sub(r'[^a-z0-9]+', '_', spell_id)
    spell_id = spell_id.strip('_')

    return spell_id

def parse_spells(input_file):
    """Parse all spells from markdown file"""
    spells = []
    current_spell = None
    description_lines = []
    in_description = False

    with open(input_file, 'r', encoding='utf-8') as f:
        lines = f.readlines()

    for i, line in enumerate(lines):
        # Check for spell header (main spell entries start with # **)
        name = parse_spell_header(line)
        if name and i > 180:  # Skip headers before spell list (around line 187)
            # Save previous spell if exists
            if current_spell:
                current_spell['description_fr'] = ' '.join(description_lines).strip()
                # Extract upcast if present
                desc = current_spell['description_fr']
                upcast_match = re.search(r'\*Emplacement de niveau supérieur[.\s]*\*\s*(.+?)(?=\n\n|\Z)', desc, re.DOTALL)
                if upcast_match:
                    current_spell['upcast'] = upcast_match.group(1).strip()
                    # Remove upcast from description
                    current_spell['description_fr'] = re.sub(r'\*Emplacement de niveau supérieur[.\s]*\*.*', '', desc, flags=re.DOTALL).strip()

                spells.append(current_spell)

            # Start new spell
            current_spell = {
                'id': create_spell_id(name),
                'name': name,
                'name_en': None,  # TODO: Add English names
                'level': None,
                'school': None,
                'casting_time': None,
                'range': None,
                'components': [],
                'duration': None,
                'concentration': False,
                'ritual': False,
                'classes': [],
                'description_fr': None,
                'description_en': None,
                'upcast': None
            }
            description_lines = []
            in_description = False
            continue

        if not current_spell:
            continue

        # Parse school, level, classes (line after spell name)
        if current_spell['school'] is None:
            school, level, classes = parse_school_level_classes(line)
            if school:
                current_spell['school'] = school
                current_spell['level'] = level if level is not None else 0
                current_spell['classes'] = classes
                continue

        # Parse properties
        casting_time = parse_property(line, 'Temps d\'incantation')
        if casting_time:
            current_spell['casting_time'] = casting_time
            # Check for ritual
            if 'rituel' in casting_time.lower():
                current_spell['ritual'] = True
            continue

        range_val = parse_property(line, 'Portée')
        if range_val:
            current_spell['range'] = range_val
            continue

        comp_str = parse_property(line, 'Composantes')
        if comp_str:
            components, material = parse_components(comp_str)
            current_spell['components'] = components
            if material:
                current_spell['material'] = material
            continue

        duration = parse_property(line, 'Durée')
        if duration:
            current_spell['duration'] = duration
            # Check for concentration
            if 'concentration' in duration.lower():
                current_spell['concentration'] = True
            continue

        # Collect description text (skip empty lines, section headers)
        if line.strip() and not line.startswith('#') and not line.startswith('**') and current_spell['duration'] is not None:
            in_description = True
            description_lines.append(line.strip())

    # Save last spell
    if current_spell and current_spell['school'] is not None:
        current_spell['description_fr'] = ' '.join(description_lines).strip()
        desc = current_spell['description_fr']
        upcast_match = re.search(r'\*Emplacement de niveau supérieur[.\s]*\*\s*(.+?)(?=\n\n|\Z)', desc, re.DOTALL)
        if upcast_match:
            current_spell['upcast'] = upcast_match.group(1).strip()
            current_spell['description_fr'] = re.sub(r'\*Emplacement de niveau supérieur[.\s]*\*.*', '', desc, flags=re.DOTALL).strip()
        spells.append(current_spell)

    # Filter out entries without school or level (section headers)
    spells = [s for s in spells if s['school'] is not None and s['level'] is not None]

    return spells

def main():
    input_file = 'docs/markdown-new/sorts_et_magie.md'
    output_file = 'data/5e/spells.json'

    print(f"Parsing spells from {input_file}...")
    spells = parse_spells(input_file)

    print(f"Found {len(spells)} spells")

    # Show first few spells for verification
    print("\nFirst 3 spells:")
    for spell in spells[:3]:
        print(f"  - {spell['name']} (ID: {spell['id']}, Level: {spell['level']}, School: {spell['school']})")
        print(f"    Classes: {', '.join(spell['classes'])}")
        print(f"    Components: {', '.join(spell['components'])}")
        if spell['concentration']:
            print(f"    Concentration: Yes")
        if spell['ritual']:
            print(f"    Ritual: Yes")
        print()

    # Create output structure
    output = {'spells': spells}

    # Ensure data/5e directory exists
    import os
    os.makedirs('data/5e', exist_ok=True)

    # Write JSON
    with open(output_file, 'w', encoding='utf-8') as f:
        json.dump(output, f, ensure_ascii=False, indent=2)

    print(f"✓ Wrote {len(spells)} spells to {output_file}")

    # Statistics
    by_level = {}
    for spell in spells:
        lvl = spell['level']
        by_level[lvl] = by_level.get(lvl, 0) + 1

    print("\nSpells by level:")
    for lvl in sorted(by_level.keys()):
        print(f"  Level {lvl}: {by_level[lvl]} spells")

if __name__ == '__main__':
    main()
