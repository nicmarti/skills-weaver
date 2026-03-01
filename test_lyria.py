#!/usr/bin/env python3
"""
Test Lyria RealTime - Musique d'ambiance pour SkillsWeaver
Appuyer sur Entrée pour changer de scène, Ctrl+C pour quitter.
"""

import asyncio
import sys
import os

try:
    import pyaudio
    HAS_PYAUDIO = True
except ImportError:
    HAS_PYAUDIO = False
    print("⚠️  pyaudio non disponible - mode sauvegarde WAV activé")

from google import genai
from google.genai import types

# Clé API - utilise la variable d'environnement si disponible, sinon fallback
GEMINI_API_KEY = os.environ.get("GEMINI_API_KEY", "")

SAMPLE_RATE = 44100
CHANNELS = 2


def list_audio_devices():
    """Affiche tous les périphériques audio disponibles."""
    p = pyaudio.PyAudio()
    print("\n📻 Périphériques audio disponibles :")
    print("-" * 50)
    for i in range(p.get_device_count()):
        info = p.get_device_info_by_index(i)
        if info['maxOutputChannels'] > 0:
            marker = " ← BlackHole" if "BlackHole" in info['name'] else ""
            marker += " ← Multi-Output" if "Multi-Output" in info['name'] or "multi" in info['name'].lower() else ""
            print(f"  [{i}] {info['name']}{marker}")
    print("-" * 50)
    p.terminate()


def find_device_index(pa, name_fragment):
    """Trouve l'index d'un périphérique par fragment de nom."""
    for i in range(pa.get_device_count()):
        info = pa.get_device_info_by_index(i)
        if name_fragment.lower() in info['name'].lower() and info['maxOutputChannels'] > 0:
            return i
    return None

# Scènes D&D avec leurs paramètres Lyria
SCENES = [
    {
        "name": "🏰 Donjon Mystérieux",
        "prompts": [
            {"text": "dark medieval dungeon, tense atmosphere, low strings, mysterious ambient, fantasy RPG", "weight": 1.0},
        ],
        "bpm": 65,
        "temperature": 1.1,
    },
    {
        "name": "⚔️  Combat Épique",
        "prompts": [
            {"text": "epic battle music, fast drums, heroic orchestra, intense combat, fantasy adventure", "weight": 1.0},
        ],
        "bpm": 145,
        "temperature": 1.0,
    },
    {
        "name": "🍺 Taverne Animée",
        "prompts": [
            {"text": "lively medieval tavern, folk music, lutes and flutes, cheerful festive atmosphere", "weight": 1.0},
        ],
        "bpm": 118,
        "temperature": 0.9,
    },
    {
        "name": "🌲 Forêt Paisible",
        "prompts": [
            {"text": "peaceful enchanted forest, gentle flutes, nature ambiance, soft adventure music, birds", "weight": 1.0},
        ],
        "bpm": 75,
        "temperature": 1.0,
    },
    {
        "name": "🛒 Marché Matinal",
        "prompts": [
            {"text": "busy medieval market morning, upbeat folk music, lively outdoor, merchants, festive", "weight": 1.0},
        ],
        "bpm": 108,
        "temperature": 0.95,
    },
    {
        "name": "🌙 Exploration Nocturne",
        "prompts": [
            {"text": "night exploration, tense stealth music, dark ambient, infiltration, mysterious warehouse", "weight": 1.0},
        ],
        "bpm": 58,
        "temperature": 1.2,
    },
        {
            "name": "🌙 Ambiance sur le port",
            "prompts": [
                {"text": "ambiance jeu de roles medieval fantastique, sur le port, des bateaux, des marins, des gens qui achetent des poissons, découverte, aventure", "weight": 1.0},
            ],
            "bpm": 58,
            "temperature": 1.2,
        },
]


def print_help():
    print("\n" + "="*50)
    print("  LYRIA REALTIME - TEST SKILLSWEAVER")
    print("="*50)
    for i, scene in enumerate(SCENES):
        print(f"  {i+1}. {scene['name']}")
    print()
    print("  Commandes :")
    print("  [1-6]  → Changer de scène directement")
    print("  [n]    → Scène suivante")
    print("  [q]    → Quitter")
    print("="*50 + "\n")


async def transition_to_scene(session, scene, previous_scene=None):
    """Transition fluide vers une nouvelle scène via fondu croisé."""
    print(f"\n🎵 → {scene['name']} (BPM: {scene['bpm']})")

    # Si on avait une scène précédente, faire un fondu croisé
    if previous_scene:
        print("   Fondu croisé en cours...")
        mixed_prompts = []
        # Ancienne scène avec poids décroissant
        for p in previous_scene["prompts"]:
            mixed_prompts.append(types.WeightedPrompt(text=p["text"], weight=0.3))
        # Nouvelle scène avec poids fort
        for p in scene["prompts"]:
            mixed_prompts.append(types.WeightedPrompt(text=p["text"], weight=1.0))

        await session.set_weighted_prompts(prompts=mixed_prompts)
        await asyncio.sleep(3)  # Laisser la transition se faire

    # Définir la scène finale
    final_prompts = [
        types.WeightedPrompt(text=p["text"], weight=p["weight"])
        for p in scene["prompts"]
    ]
    await session.set_weighted_prompts(prompts=final_prompts)
    await session.set_music_generation_config(
        config=types.LiveMusicGenerationConfig(
            bpm=scene["bpm"],
            temperature=scene["temperature"],
        )
    )
    print(f"   ✓ Scène active : {scene['name']}")


async def audio_player(session, stream):
    """Tâche de fond : reçoit et joue les chunks audio."""
    bytes_received = 0
    try:
        async for message in session.receive():
            if message.server_content and message.server_content.audio_chunks:
                for chunk in message.server_content.audio_chunks:
                    if stream:
                        stream.write(chunk.data)
                    bytes_received += len(chunk.data)
    except asyncio.CancelledError:
        pass
    except Exception as e:
        print(f"\n⚠️  Erreur audio : {e}")


async def audio_collector(session, buffer):
    """Mode WAV : collecte les chunks en mémoire."""
    try:
        async for message in session.receive():
            if message.server_content and message.server_content.audio_chunks:
                for chunk in message.server_content.audio_chunks:
                    buffer.extend(chunk.data)
    except asyncio.CancelledError:
        pass


async def keyboard_listener(scene_queue):
    """Écoute les touches clavier de façon asynchrone."""
    loop = asyncio.get_event_loop()
    while True:
        key = await loop.run_in_executor(None, sys.stdin.readline)
        key = key.strip().lower()
        if key:
            await scene_queue.put(key)


async def main():
    print_help()

    # Lister les périphériques si demandé
    if len(sys.argv) > 1 and sys.argv[1] == '--list-devices':
        if HAS_PYAUDIO:
            list_audio_devices()
        else:
            print("pyaudio requis pour lister les périphériques")
        return

    # Périphérique de sortie audio (None = défaut système)
    # Exemples : "BlackHole", "Speakers + BlackHole", None
    OUTPUT_DEVICE_NAME = os.environ.get("LYRIA_AUDIO_DEVICE", None)

    # Initialisation du client Gemini
    print("🔌 Connexion à Lyria RealTime...")
    client = genai.Client(
        api_key=GEMINI_API_KEY,
        http_options={'api_version': 'v1alpha'}
    )

    # Initialisation audio
    pa = None
    stream = None
    if HAS_PYAUDIO:
        try:
            pa = pyaudio.PyAudio()

            # Chercher le périphérique demandé
            device_index = None
            if OUTPUT_DEVICE_NAME:
                device_index = find_device_index(pa, OUTPUT_DEVICE_NAME)
                if device_index is None:
                    print(f"⚠️  Périphérique '{OUTPUT_DEVICE_NAME}' non trouvé → périphérique par défaut")
                    print("   Lance avec --list-devices pour voir les options disponibles")
                else:
                    info = pa.get_device_info_by_index(device_index)
                    print(f"🔊 Sortie audio : [{device_index}] {info['name']}")
            else:
                print("🔊 Sortie audio : périphérique système par défaut")

            stream = pa.open(
                format=pyaudio.paInt16,
                channels=CHANNELS,
                rate=SAMPLE_RATE,
                output=True,
                output_device_index=device_index,
                frames_per_buffer=1024,
            )
        except Exception as e:
            print(f"⚠️  pyaudio échoue ({e}) → mode WAV")
            if pa:
                pa.terminate()
            stream = None
            pa = None

    audio_buffer = bytearray()

    try:
        async with client.aio.live.music.connect(
            model='models/lyria-realtime-exp'
        ) as session:
            print("✓ Connecté à Lyria RealTime\n")

            # Démarrer avec la première scène
            current_scene_idx = 0
            current_scene = SCENES[current_scene_idx]
            await transition_to_scene(session, current_scene)
            await session.play()
            print("▶ Musique en cours... (attendre ~5s le premier son)\n")

            # Lancer la réception audio en arrière-plan
            if stream:
                audio_task = asyncio.create_task(audio_player(session, stream))
            else:
                audio_task = asyncio.create_task(audio_collector(session, audio_buffer))

            # File de commandes clavier
            scene_queue = asyncio.Queue()
            keyboard_task = asyncio.create_task(keyboard_listener(scene_queue))

            print("En attente de commande (tapez [n], [1-6] ou [q] + Entrée) :")

            try:
                while True:
                    try:
                        key = await asyncio.wait_for(scene_queue.get(), timeout=0.5)
                    except asyncio.TimeoutError:
                        continue

                    if key == 'q' or key == 'quit':
                        print("\n⏹  Arrêt...")
                        break
                    elif key == 'n':
                        prev = current_scene
                        current_scene_idx = (current_scene_idx + 1) % len(SCENES)
                        current_scene = SCENES[current_scene_idx]
                        await transition_to_scene(session, current_scene, prev)
                    elif key.isdigit() and 1 <= int(key) <= len(SCENES):
                        prev = current_scene
                        current_scene_idx = int(key) - 1
                        current_scene = SCENES[current_scene_idx]
                        await transition_to_scene(session, current_scene, prev)
                    else:
                        print(f"  Commande inconnue : '{key}' (tapez n, 1-6, ou q)")

            except KeyboardInterrupt:
                print("\n⏹  Interrompu par Ctrl+C")
            finally:
                keyboard_task.cancel()
                audio_task.cancel()
                try:
                    await asyncio.gather(audio_task, keyboard_task, return_exceptions=True)
                except Exception:
                    pass

    except Exception as e:
        print(f"\n❌ Erreur de connexion : {e}")
        print("\nVérifications :")
        print("  1. Ta clé API est valide et a accès à Lyria RealTime")
        print("  2. Le modèle lyria-realtime-exp est disponible dans ta région")
        print("  3. pip install google-genai --upgrade")
        raise
    finally:
        if stream:
            stream.stop_stream()
            stream.close()
        if pa:
            pa.terminate()

        # Sauvegarder en WAV si mode buffer
        if not stream and len(audio_buffer) > 0:
            import wave
            wav_file = "test_lyria_output.wav"
            with wave.open(wav_file, "wb") as wf:
                wf.setnchannels(CHANNELS)
                wf.setsampwidth(2)  # 16-bit
                wf.setframerate(SAMPLE_RATE)
                wf.writeframes(bytes(audio_buffer))
            duration = len(audio_buffer) / (SAMPLE_RATE * CHANNELS * 2)
            print(f"\n💾 Audio sauvegardé : {wav_file} ({duration:.1f}s)")


if __name__ == "__main__":
    asyncio.run(main())
