// wizard.js — drives the "fortune teller" adventure-creation wizard.
// The deck (questions + cards) is injected server-side as JSON; card->meaning
// mapping stays on the server, so this file only tracks which cards were chosen.
(function () {
    "use strict";

    const deckEl = document.getElementById("deck-data");
    let deck = { questions: [], cards: [] };
    try {
        deck = JSON.parse(deckEl.getAttribute("data-deck"));
    } catch (e) {
        console.error("Impossible de lire le jeu de cartes", e);
    }

    const cardById = {};
    (deck.cards || []).forEach((c) => { cardById[c.id] = c; });

    // Debug mode (?debug=1): exercise the full draw but create nothing.
    const DEBUG = new URLSearchParams(window.location.search).get("debug") === "1";
    if (DEBUG) {
        const badge = document.createElement("div");
        badge.className = "wizard-debug-badge";
        badge.textContent = "DEBUG — aucune aventure ne sera créée";
        document.body.appendChild(badge);
    }

    const state = {
        qIndex: 0,
        draws: {},          // questionId -> [cardId, cardId, cardId]
        selections: {},     // questionId -> cardId
    };

    // --- helpers -----------------------------------------------------------

    function showStep(name) {
        document.querySelectorAll(".wizard-step").forEach((s) => {
            s.hidden = s.getAttribute("data-step") !== name;
        });
    }

    // Fisher-Yates shuffle (no Date/Math.random concerns in browser).
    function pickThree(ids) {
        const pool = ids.slice();
        for (let i = pool.length - 1; i > 0; i--) {
            const j = Math.floor(Math.random() * (i + 1));
            [pool[i], pool[j]] = [pool[j], pool[i]];
        }
        return pool.slice(0, 3);
    }

    function renderProgress() {
        const el = document.getElementById("question-progress");
        const total = deck.questions.length;
        el.textContent = "Carte " + (state.qIndex + 1) + " sur " + total;
    }

    function renderQuestion() {
        const q = deck.questions[state.qIndex];
        document.getElementById("question-prompt").textContent = q.prompt;
        renderProgress();

        // Stable draw per question (computed once).
        if (!state.draws[q.id]) {
            state.draws[q.id] = pickThree(q.card_ids);
        }

        const row = document.getElementById("card-row");
        row.innerHTML = "";
        state.draws[q.id].forEach((cardId) => {
            const card = cardById[cardId];
            if (!card) return;

            const el = document.createElement("button");
            el.className = "tarot-card";
            el.type = "button";

            const img = document.createElement("img");
            img.className = "tarot-art";
            img.src = "/static/cards/" + card.art;
            img.alt = card.name;
            img.onerror = function () {
                img.style.display = "none";
                el.classList.add("no-art");
            };

            const name = document.createElement("div");
            name.className = "tarot-name";
            name.textContent = card.name;

            const flavor = document.createElement("div");
            flavor.className = "tarot-flavor";
            flavor.textContent = card.flavor || "";

            el.appendChild(img);
            el.appendChild(name);
            el.appendChild(flavor);
            el.addEventListener("click", () => selectCard(q.id, cardId, el));
            row.appendChild(el);
        });

        showStep("questions");
    }

    function selectCard(questionId, cardId, el) {
        state.selections[questionId] = cardId;
        // Briefly highlight the chosen card, then advance.
        document.querySelectorAll(".tarot-card").forEach((c) => c.classList.remove("chosen"));
        el.classList.add("chosen");

        setTimeout(() => {
            if (state.qIndex < deck.questions.length - 1) {
                state.qIndex++;
                renderQuestion();
            } else {
                showStep("characters");
            }
        }, 550);
    }

    function selectedCharacters() {
        return Array.from(document.querySelectorAll(".hero-check:checked")).map((c) => c.value);
    }

    // Ask the LLM for an evocative title coherent with the reading.
    // onlyIfEmpty=true skips when the player already typed a name.
    async function suggestName(onlyIfEmpty) {
        const nameInput = document.getElementById("wizard-name");
        if (onlyIfEmpty && nameInput.value.trim() !== "") return;

        const hint = document.getElementById("wizard-name-hint");
        const btn = document.getElementById("wizard-name-suggest");
        hint.hidden = false;
        btn.disabled = true;
        btn.classList.add("spinning");

        try {
            const res = await fetch("/adventure/suggest-name", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    card_ids: orderedCardIds(),
                    theme: document.getElementById("wizard-theme").value.trim(),
                }),
            });
            if (res.ok) {
                const data = await res.json();
                if (data.name) nameInput.value = data.name;
            }
        } catch (e) {
            /* best-effort: leave the field as-is */
        } finally {
            hint.hidden = true;
            btn.disabled = false;
            btn.classList.remove("spinning");
        }
    }

    // --- submission --------------------------------------------------------

    function orderedCardIds() {
        // One card per question, in question order (matches server validation).
        return deck.questions.map((q) => state.selections[q.id]).filter(Boolean);
    }

    async function submit() {
        const nameInput = document.getElementById("wizard-name");
        const errEl = document.getElementById("wizard-error");
        const name = nameInput.value.trim();
        if (!name && !DEBUG) {
            errEl.textContent = "Donne un nom à ton aventure.";
            errEl.hidden = false;
            nameInput.focus();
            return;
        }
        errEl.hidden = true;

        const payload = {
            name: name,
            theme: document.getElementById("wizard-theme").value.trim(),
            duration: document.getElementById("wizard-duration").value,
            card_ids: orderedCardIds(),
            characters: selectedCharacters(),
            debug: DEBUG,
        };

        showOverlay(DEBUG ? "Lecture des cartes (mode debug)..." : "La voyante consulte le destin...");
        if (!DEBUG) startPhaseA();

        let res;
        try {
            res = await fetch("/adventure", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(payload),
            });
        } catch (e) {
            return failOverlay("La connexion au destin a échoué. Réessaie.");
        }

        if (!res.ok) {
            let msg = "La voyante n'a pu achever sa lecture.";
            try { const j = await res.json(); if (j.error) msg = j.error; } catch (e) {}
            return failOverlay(msg);
        }

        const data = await res.json();
        if (DEBUG) {
            renderDebug(data);
            return;
        }
        await revealDestiny(data.slug, data.redirect);
    }

    function renderDebug(data) {
        clearStepTimers();
        document.getElementById("wizard-spinner").hidden = true;
        document.getElementById("wizard-steps").hidden = true;
        document.getElementById("destiny-card").hidden = true;
        document.getElementById("wizard-overlay-text").textContent =
            "🔮 Lecture des cartes (aucune aventure créée)";

        const lines = [];
        lines.push("Nom         : " + (data.name || "(vide)"));
        lines.push("Durée       : " + (data.duration || "-"));
        lines.push("Héros       : " + ((data.selected_characters || []).join(", ") || "(aucun)"));
        lines.push("");
        lines.push("=== Cartes choisies ===");
        (data.chosen_cards || []).forEach((c) => {
            lines.push("• [" + c.question + "] " + c.name);
            lines.push("    « " + c.flavor + " »");
            lines.push("    params: " + JSON.stringify(c.params));
        });
        lines.push("");
        lines.push("=== Brief agrégé (paramètres cachés) ===");
        Object.keys(data.brief || {}).forEach((k) => {
            if (data.brief[k]) lines.push("  " + k + " : " + data.brief[k]);
        });
        lines.push("");
        lines.push("=== Cohérence ===");
        lines.push("  Aventures récentes : " + ((data.coherence.recent_adventures || []).join(", ") || "(aucune)"));
        lines.push("  PNJ à éviter       : " + ((data.coherence.npc_names_to_avoid || []).join(", ") || "(aucun)"));
        lines.push("");
        lines.push("=== Fragment injecté (lecture de la voyante) ===");
        lines.push(data.prompt_section || "(vide)");
        lines.push("");
        lines.push("=== PROMPT COMPLET envoyé au générateur de plan de campagne ===");
        lines.push("(le fragment ci-dessus est inséré dans ce prompt ; le LLM renvoie un");
        lines.push(" campaign-plan.json structuré en actes/PNJ/lieux/foreshadows)");
        lines.push("");
        lines.push(data.campaign_prompt || "(indisponible — persona DM introuvable ?)");

        const panel = document.getElementById("wizard-debug");
        panel.hidden = false;
        panel.innerHTML = "";

        const pre = document.createElement("pre");
        pre.className = "wizard-debug-pre";
        pre.textContent = lines.join("\n");
        panel.appendChild(pre);

        const actions = document.createElement("div");
        actions.className = "wizard-actions";
        const close = document.createElement("button");
        close.className = "btn btn-secondary";
        close.textContent = "Fermer";
        close.addEventListener("click", () => {
            document.getElementById("wizard-overlay").hidden = true;
            panel.hidden = true;
        });
        const restart = document.createElement("button");
        restart.className = "btn btn-primary";
        restart.textContent = "Nouveau tirage";
        restart.addEventListener("click", () => window.location.reload());
        actions.appendChild(close);
        actions.appendChild(restart);
        panel.appendChild(actions);
    }

    // --- staged loading ----------------------------------------------------
    // The creation request is a single opaque blocking call (campaign plan +
    // NPCs), so we can't get real sub-progress from it. Instead we surface the
    // work as a narrative checklist with two REAL sync points: the POST
    // returning (steps 0-2 done) and the destiny image being ready (step 3).
    const LOADING_STEPS = [
        { icon: "🔮", label: "La voyante mêle les cartes du destin" },
        { icon: "📜", label: "Les fils de ton histoire se tissent" },
        { icon: "🎭", label: "Les visages que tu croiseras s'éveillent" },
        { icon: "🎴", label: "Le destin prend forme dans la boule de cristal" },
        { icon: "✨", label: "Ton destin est scellé" },
    ];
    // Steps 0..PHASE_A_LAST resolve when the POST returns; step 3 when the
    // image is ready; step 4 is the final sealed state.
    const PHASE_A_LAST = 2;

    let stepTimers = [];

    function clearStepTimers() {
        stepTimers.forEach((t) => clearTimeout(t));
        stepTimers = [];
    }

    function renderSteps() {
        const list = document.getElementById("wizard-steps");
        list.innerHTML = "";
        LOADING_STEPS.forEach((step, i) => {
            const li = document.createElement("li");
            li.className = "wizard-step-item is-pending";
            li.id = "wizard-step-" + i;

            const icon = document.createElement("span");
            icon.className = "wizard-step-icon";
            icon.textContent = step.icon;

            const label = document.createElement("span");
            label.className = "wizard-step-label";
            label.textContent = step.label;

            li.appendChild(icon);
            li.appendChild(label);
            list.appendChild(li);
        });
        list.hidden = false;
    }

    // Steps before `active` are done, `active` is active, the rest pending.
    function paintSteps(active, allDone) {
        LOADING_STEPS.forEach((_, i) => {
            const li = document.getElementById("wizard-step-" + i);
            if (!li) return;
            li.classList.remove("is-pending", "is-active", "is-done");
            if (allDone || i < active) {
                li.classList.add("is-done");
            } else if (i === active) {
                li.classList.add("is-active");
            } else {
                li.classList.add("is-pending");
            }
        });
    }

    // Walk the active marker through phase-A beats. Never advances past
    // PHASE_A_LAST while the POST is still in flight, so we never claim a step
    // "done" before the server has actually finished it.
    function startPhaseA() {
        renderSteps();
        paintSteps(0, false);
        stepTimers.push(setTimeout(() => paintSteps(1, false), 1800));
        stepTimers.push(setTimeout(() => paintSteps(PHASE_A_LAST, false), 9000));
    }

    // POST returned: phase A is genuinely complete — move to the vision.
    function enterVision() {
        clearStepTimers();
        paintSteps(3, false);
    }

    function sealDestiny() {
        clearStepTimers();
        paintSteps(LOADING_STEPS.length, true);
    }

    // --- overlay / destiny card -------------------------------------------

    function showOverlay(text) {
        document.getElementById("wizard-overlay-text").textContent = text;
        document.getElementById("destiny-card").hidden = true;
        document.getElementById("wizard-spinner").hidden = false;
        document.getElementById("wizard-overlay").hidden = false;
    }

    function failOverlay(text) {
        clearStepTimers();
        const errEl = document.getElementById("wizard-error");
        document.getElementById("wizard-overlay").hidden = true;
        document.getElementById("wizard-steps").hidden = true;
        errEl.textContent = text;
        errEl.hidden = false;
        showStep("final");
    }

    function sleep(ms) {
        return new Promise((resolve) => setTimeout(resolve, ms));
    }

    async function revealDestiny(slug, redirect) {
        enterVision(); // POST returned: plan + NPCs are done, now awaiting the image.
        document.getElementById("wizard-overlay-text").textContent = "Le destin se révèle...";

        const deadline = Date.now() + 15000;
        while (Date.now() < deadline) {
            try {
                const r = await fetch("/adventure/destiny/" + slug, { cache: "no-store" });
                if (r.ok) {
                    const j = await r.json();
                    if (j.ready && j.url) {
                        const img = document.getElementById("destiny-card");
                        img.src = j.url;
                        img.hidden = false;
                        document.getElementById("wizard-spinner").hidden = true;
                        sealDestiny();
                        document.getElementById("wizard-overlay-text").textContent =
                            "Ton destin est scellé.";
                        await sleep(2800);
                        window.location = redirect;
                        return;
                    }
                }
            } catch (e) { /* keep polling */ }
            await sleep(1500);
        }
        // Image not ready in time — seal the checklist anyway and continue.
        sealDestiny();
        window.location = redirect;
    }

    // --- wiring ------------------------------------------------------------

    document.getElementById("wizard-start").addEventListener("click", () => {
        state.qIndex = 0;
        renderQuestion();
    });
    document.getElementById("characters-next").addEventListener("click", () => {
        showStep("final");
        suggestName(true); // pre-fill the title, only if the player hasn't typed one
    });
    document.getElementById("wizard-name-suggest").addEventListener("click", () => suggestName(false));
    document.getElementById("wizard-submit").addEventListener("click", submit);

    showStep("intro");
})();
