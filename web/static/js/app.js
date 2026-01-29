// SkillsWeaver Web Client
// Handles SSE streaming and game interaction

(function() {
    'use strict';

    // Only run on game pages
    if (typeof slug === 'undefined') {
        return;
    }

    const conversationEl = document.getElementById('conversation');
    const messageForm = document.getElementById('message-form');
    const messageInput = document.getElementById('message-input');
    const sendBtn = document.getElementById('send-btn');
    const toolStatus = document.getElementById('tool-status');
    const toolNameEl = toolStatus?.querySelector('.tool-name');

    let eventSource = null;
    let currentDmMessage = null;
    let currentRawText = ''; // Accumulate raw text for markdown rendering

    // Initialize
    function init() {
        if (messageForm) {
            messageForm.addEventListener('submit', handleSubmit);
        }
        initMinimap();
        scrollToBottom();
    }

    // Parse markdown table to HTML
    function parseTable(tableText) {
        const lines = tableText.trim().split('\n');
        if (lines.length < 2) return null;

        // Check if it looks like a table (has | characters)
        if (!lines[0].includes('|')) return null;

        const parseRow = (line) => {
            return line.split('|')
                .map(cell => cell.trim())
                .filter((cell, i, arr) => i > 0 && i < arr.length - 1 || cell !== '');
        };

        const headers = parseRow(lines[0]);

        // Check for separator line (|---|---|)
        let dataStartIndex = 1;
        if (lines[1] && lines[1].match(/^\|?[\s\-:]+\|/)) {
            dataStartIndex = 2;
        }

        let html = '<table class="markdown-table"><thead><tr>';
        headers.forEach(h => {
            html += `<th>${escapeHtml(h)}</th>`;
        });
        html += '</tr></thead><tbody>';

        for (let i = dataStartIndex; i < lines.length; i++) {
            if (!lines[i].trim()) continue;
            const cells = parseRow(lines[i]);
            if (cells.length === 0) continue;
            html += '<tr>';
            cells.forEach(cell => {
                // Apply inline formatting to cell content
                let cellHtml = escapeHtml(cell);
                cellHtml = cellHtml.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
                cellHtml = cellHtml.replace(/\*([^*]+?)\*/g, '<em>$1</em>');
                html += `<td>${cellHtml}</td>`;
            });
            html += '</tr>';
        }
        html += '</tbody></table>';
        return html;
    }

    // Simple markdown parser
    function renderMarkdown(text) {
        // Split into lines for processing
        const lines = text.split('\n');
        let html = '';
        let inTable = false;
        let tableLines = [];
        let inCodeBlock = false;

        for (let i = 0; i < lines.length; i++) {
            let line = lines[i];

            // Code blocks
            if (line.trim().startsWith('```')) {
                if (inCodeBlock) {
                    html += '</code></pre>';
                    inCodeBlock = false;
                } else {
                    html += '<pre><code>';
                    inCodeBlock = true;
                }
                continue;
            }

            if (inCodeBlock) {
                html += escapeHtml(line) + '\n';
                continue;
            }

            // Table detection
            if (line.trim().startsWith('|') || (line.includes('|') && line.trim().match(/^\|?[^|]+\|/))) {
                if (!inTable) {
                    inTable = true;
                    tableLines = [];
                }
                tableLines.push(line);
                continue;
            } else if (inTable) {
                // End of table
                const tableHtml = parseTable(tableLines.join('\n'));
                if (tableHtml) {
                    html += tableHtml;
                } else {
                    // Fallback: render as regular lines
                    tableLines.forEach(tl => {
                        html += renderLine(tl) + '<br>';
                    });
                }
                inTable = false;
                tableLines = [];
            }

            // Regular line processing
            html += renderLine(line);

            // Add line break if not a block element
            if (!line.trim().startsWith('#') && line.trim() !== '') {
                html += '<br>';
            } else if (line.trim() === '') {
                html += '<br>';
            }
        }

        // Handle remaining table at end
        if (inTable && tableLines.length > 0) {
            const tableHtml = parseTable(tableLines.join('\n'));
            if (tableHtml) {
                html += tableHtml;
            }
        }

        if (inCodeBlock) {
            html += '</code></pre>';
        }

        return html;
    }

    // Render a single line with inline formatting
    function renderLine(line) {
        let html = escapeHtml(line);

        // Headers: ### Header
        if (html.match(/^#{1,4}\s/)) {
            const level = html.match(/^(#+)/)[1].length;
            const headerTag = `h${Math.min(level + 1, 6)}`;
            html = `<${headerTag}>${html.replace(/^#+\s*/, '')}</${headerTag}>`;
            return html;
        }

        // Bold: **text** or __text__
        html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
        html = html.replace(/__(.+?)__/g, '<strong>$1</strong>');

        // Italic: *text* or _text_
        html = html.replace(/(?<![*\w])\*([^*]+?)\*(?![*\w])/g, '<em>$1</em>');
        html = html.replace(/(?<![_\w])_([^_]+?)_(?![_\w])/g, '<em>$1</em>');

        // Lists: - item or * item (at start of line)
        if (html.match(/^[\-\*]\s/)) {
            html = '<li>' + html.replace(/^[\-\*]\s*/, '') + '</li>';
        }

        return html;
    }

    // Handle form submission
    async function handleSubmit(e) {
        e.preventDefault();

        const message = messageInput.value.trim();
        if (!message) return;

        // Add user message to conversation
        addUserMessage(message);

        // Clear input and disable form
        messageInput.value = '';
        setLoading(true);

        try {
            // Send message to server
            const response = await fetch(`/play/${slug}/message`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                },
                body: `message=${encodeURIComponent(message)}`
            });

            if (!response.ok) {
                const data = await response.json();
                throw new Error(data.error || 'Failed to send message');
            }

            // Connect to SSE stream
            connectSSE();

        } catch (error) {
            console.error('Error:', error);
            addErrorMessage(error.message);
            setLoading(false);
        }
    }

    // Connect to SSE stream
    function connectSSE() {
        if (eventSource) {
            eventSource.close();
        }

        eventSource = new EventSource(`/play/${slug}/stream`);

        eventSource.addEventListener('text', handleTextEvent);
        eventSource.addEventListener('tool_start', handleToolStart);
        eventSource.addEventListener('tool_complete', handleToolComplete);
        eventSource.addEventListener('agent_start', handleAgentStart);
        eventSource.addEventListener('agent_complete', handleAgentComplete);
        eventSource.addEventListener('image', handleImageEvent);
        eventSource.addEventListener('location_update', handleLocationUpdate);
        eventSource.addEventListener('map_generated', handleMapGenerated);
        eventSource.addEventListener('error', handleErrorEvent);
        eventSource.addEventListener('complete', handleComplete);
        eventSource.addEventListener('done', handleDone);

        eventSource.onerror = function(e) {
            console.error('SSE error:', e);
            eventSource.close();
            setLoading(false);
        };
    }

    // Handle text chunks
    function handleTextEvent(e) {
        // Text is JSON-encoded to preserve newlines
        let text;
        try {
            text = JSON.parse(e.data);
        } catch (err) {
            // Fallback if not JSON
            text = e.data;
        }

        if (!currentDmMessage) {
            currentDmMessage = createDmMessage();
            currentRawText = '';
        }

        // Accumulate raw text
        currentRawText += text;

        // Render markdown and update display
        const textEl = currentDmMessage.querySelector('.dm-text');
        if (textEl) {
            textEl.innerHTML = renderMarkdown(currentRawText);
        }

        scrollToBottom();
    }

    // Handle tool start
    function handleToolStart(e) {
        try {
            const data = JSON.parse(e.data);
            showToolStatus(data.tool_name);

            // Add tool notification to conversation
            if (currentDmMessage) {
                const notification = document.createElement('div');
                notification.className = 'tool-notification';
                notification.textContent = `[${data.tool_name}...]`;
                currentDmMessage.appendChild(notification);
                scrollToBottom();
            }
        } catch (err) {
            console.error('Error parsing tool_start:', err);
        }
    }

    // Handle tool complete
    function handleToolComplete(e) {
        try {
            const data = JSON.parse(e.data);
            hideToolStatus();

            // Update tool notification
            if (currentDmMessage) {
                const notifications = currentDmMessage.querySelectorAll('.tool-notification:not(.complete)');
                const lastNotification = notifications[notifications.length - 1];
                if (lastNotification) {
                    lastNotification.textContent = `[${data.display}]`;
                    lastNotification.classList.add('complete');
                }
            }
        } catch (err) {
            console.error('Error parsing tool_complete:', err);
        }
    }

    // Handle agent invocation start
    function handleAgentStart(e) {
        try {
            const data = JSON.parse(e.data);
            showToolStatus(`Consulting ${data.agent_name}...`);

            // Add agent notification
            if (currentDmMessage) {
                const notification = document.createElement('div');
                notification.className = 'tool-notification agent';
                notification.textContent = `[Consulting ${data.agent_name}...]`;
                currentDmMessage.appendChild(notification);
                scrollToBottom();
            }
        } catch (err) {
            console.error('Error parsing agent_start:', err);
        }
    }

    // Handle agent invocation complete
    function handleAgentComplete(e) {
        try {
            const data = JSON.parse(e.data);
            hideToolStatus();

            // Update agent notification
            if (currentDmMessage) {
                const notifications = currentDmMessage.querySelectorAll('.tool-notification.agent:not(.complete)');
                const lastNotification = notifications[notifications.length - 1];
                if (lastNotification) {
                    const duration = data.duration_ms ? ` (${(data.duration_ms / 1000).toFixed(1)}s)` : '';
                    lastNotification.textContent = `[${data.agent_name} responded${duration}]`;
                    lastNotification.classList.add('complete');
                }
            }
        } catch (err) {
            console.error('Error parsing agent_complete:', err);
        }
    }

    // Handle image event
    function handleImageEvent(e) {
        try {
            const data = JSON.parse(e.data);

            if (currentDmMessage && data.image_path) {
                // Convert path to web URL
                const imagePath = data.image_path;
                // Extract relative path from data/adventures/<slug>/images/...
                const match = imagePath.match(/images\/.+$/);
                if (match) {
                    const img = document.createElement('img');
                    img.className = 'message-image';
                    img.src = `/play/${slug}/images/${match[0].replace('images/', '')}`;
                    img.alt = data.prompt || 'Generated image';
                    img.loading = 'lazy';
                    currentDmMessage.appendChild(img);
                    scrollToBottom();
                }
            }

            // Check if this is a map image (contains map type keywords)
            if (data.image_path && (
                data.image_path.includes('_city_') ||
                data.image_path.includes('_region_') ||
                data.image_path.includes('_dungeon_') ||
                data.image_path.includes('_tactical_')
            )) {
                // Map was generated - extract location from filename
                // Filename format: location_maptype_scale_model.png
                // Example: cordova_city_medium_flux-pro-11.png
                const filename = data.image_path.split('/').pop();
                const parts = filename.split('_');

                // Extract location name (all parts before the map type)
                let locationName = '';
                for (let i = 0; i < parts.length; i++) {
                    const part = parts[i];
                    if (part === 'city' || part === 'region' || part === 'dungeon' || part === 'tactical') {
                        break;
                    }
                    if (locationName) locationName += ' ';
                    locationName += part;
                }

                // Capitalize location name
                locationName = locationName.split('-').map(word =>
                    word.charAt(0).toUpperCase() + word.slice(1)
                ).join(' ');

                // Refresh minimap after short delay to ensure file is saved
                setTimeout(() => {
                    updateMinimap(locationName || 'Unknown');
                }, 1500); // 1.5 second delay to ensure map file is written
            }
        } catch (err) {
            console.error('Error parsing image event:', err);
        }
    }

    // Handle error event
    function handleErrorEvent(e) {
        try {
            const data = JSON.parse(e.data);
            addErrorMessage(data.error);
        } catch (err) {
            addErrorMessage('An error occurred');
        }
        hideToolStatus();
    }

    // Handle complete event
    function handleComplete(e) {
        // Final render of accumulated text
        if (currentDmMessage && currentRawText) {
            const textEl = currentDmMessage.querySelector('.dm-text');
            if (textEl) {
                textEl.innerHTML = renderMarkdown(currentRawText);
            }
        }

        currentDmMessage = null;
        currentRawText = '';
        hideToolStatus();
        setLoading(false);

        // Trigger adventure info refresh
        document.body.dispatchEvent(new CustomEvent('refreshInfo'));
    }

    // Handle done event (stream closed)
    function handleDone(e) {
        if (eventSource) {
            eventSource.close();
            eventSource = null;
        }

        // Final render
        if (currentDmMessage && currentRawText) {
            const textEl = currentDmMessage.querySelector('.dm-text');
            if (textEl) {
                textEl.innerHTML = renderMarkdown(currentRawText);
            }
        }

        currentDmMessage = null;
        currentRawText = '';
        hideToolStatus();
        setLoading(false);
    }

    // ============================================
    // MINI-MAP FUNCTIONS
    // ============================================

    // Handle location update SSE event
    function handleLocationUpdate(e) {
        const data = JSON.parse(e.data);
        updateMinimap(data.location);
    }

    // Handle map generated SSE event
    function handleMapGenerated(e) {
        const data = JSON.parse(e.data);
        // Refresh minimap after short delay to ensure file is saved
        setTimeout(() => {
            updateMinimap(data.location);
        }, 1000);
    }

    // Fetch and update minimap data
    async function updateMinimap(location) {
        try {
            const response = await fetch(`/play/${slug}/minimap`);
            if (!response.ok) {
                console.error('Failed to fetch minimap data');
                return;
            }
            const data = await response.json();
            renderMinimap(data);
        } catch (err) {
            console.error('Error updating minimap:', err);
        }
    }

    // Render minimap with data
    function renderMinimap(data) {
        const breadcrumb = document.getElementById('minimap-breadcrumb');
        const container = document.getElementById('minimap-container');

        if (!breadcrumb || !container) return;

        // Update breadcrumb
        breadcrumb.innerHTML = data.hierarchy.map((item, index) => {
            const isLast = index === data.hierarchy.length - 1;
            const classes = isLast ? 'breadcrumb-item active' : 'breadcrumb-item';
            return `<span class="${classes}" data-level="${index}">${item}</span>`;
        }).join('<span class="breadcrumb-separator">›</span>');

        // Add breadcrumb click handlers
        breadcrumb.querySelectorAll('.breadcrumb-item:not(.active)').forEach(item => {
            item.addEventListener('click', () => {
                const level = parseInt(item.dataset.level);
                const location = data.hierarchy[level];
                updateMinimap(location);
            });
        });

        // Update map image or placeholder
        if (data.map_available && data.map_url) {
            container.innerHTML = `<img src="${data.map_url}" alt="Carte de ${data.location}" class="minimap-image" id="minimap-image" loading="lazy">`;
            const mapImage = document.getElementById('minimap-image');
            if (mapImage) {
                mapImage.addEventListener('click', () => {
                    openMinimapLightbox(data.map_url, data.location);
                });
            }
        } else {
            container.innerHTML = `
                <div class="minimap-placeholder">
                    <span class="minimap-icon">🗺️</span>
                    <p>Carte non disponible</p>
                    <button class="btn-generate-map" id="btn-generate-map">Générer Carte</button>
                </div>`;
            const generateBtn = document.getElementById('btn-generate-map');
            if (generateBtn) {
                generateBtn.addEventListener('click', () => {
                    requestMapGeneration(data.location, data.map_type || 'city');
                });
            }
        }
    }

    // Open minimap in lightbox (reuse existing lightbox)
    function openMinimapLightbox(imageUrl, caption) {
        const lightbox = document.getElementById('lightbox');
        const lightboxImage = document.getElementById('lightbox-image');
        if (lightbox && lightboxImage) {
            lightboxImage.src = imageUrl;
            lightboxImage.alt = caption;
            lightbox.classList.add('active');
        }
    }

    // Request map generation via DM agent
    function requestMapGeneration(location, mapType) {
        if (!messageInput || !messageForm) return;
        const message = `generate_map ${mapType} "${location}"`;
        messageInput.value = message;
        // Trigger form submission
        const submitEvent = new Event('submit', { cancelable: true, bubbles: true });
        messageForm.dispatchEvent(submitEvent);
    }

    // Initialize minimap on page load
    async function initMinimap() {
        const locationEl = document.querySelector('.info-value.location');
        const location = locationEl ? locationEl.textContent.trim() : 'Unknown';
        await updateMinimap(location);
    }

    // Add user message to conversation
    function addUserMessage(text) {
        const messageEl = document.createElement('div');
        messageEl.className = 'message message-user';
        messageEl.textContent = text;
        conversationEl.appendChild(messageEl);
        scrollToBottom();
    }

    // Create a new DM message container
    function createDmMessage() {
        const messageEl = document.createElement('div');
        messageEl.className = 'message message-dm';

        const textEl = document.createElement('div');
        textEl.className = 'dm-text';
        messageEl.appendChild(textEl);

        conversationEl.appendChild(messageEl);
        return messageEl;
    }

    // Add error message
    function addErrorMessage(text) {
        const messageEl = document.createElement('div');
        messageEl.className = 'message message-dm';
        messageEl.innerHTML = `<div class="error-toast"><span class="error-icon">&#x26A0;</span><span class="error-text">${escapeHtml(text)}</span></div>`;
        conversationEl.appendChild(messageEl);
        scrollToBottom();
    }

    // Show tool status
    function showToolStatus(name) {
        if (toolStatus && toolNameEl) {
            toolNameEl.textContent = name;
            toolStatus.classList.remove('hidden');
        }
    }

    // Hide tool status
    function hideToolStatus() {
        if (toolStatus) {
            toolStatus.classList.add('hidden');
        }
    }

    // Set loading state
    function setLoading(loading) {
        if (sendBtn) {
            sendBtn.disabled = loading;
            sendBtn.classList.toggle('loading', loading);
        }
        if (messageInput) {
            messageInput.disabled = loading;
        }
    }

    // Scroll to bottom of conversation
    function scrollToBottom() {
        if (conversationEl) {
            conversationEl.scrollTop = conversationEl.scrollHeight;
        }
    }

    // Escape HTML
    function escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // Initialize when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();

// ============================================================
// Character Sheet Modal Module
// ============================================================
(function() {
    'use strict';

    // Only run on game pages
    if (typeof slug === 'undefined') {
        return;
    }

    const characterModal = document.getElementById('character-modal');
    const characterModalOverlay = characterModal?.querySelector('.character-modal-overlay');
    const characterModalClose = document.getElementById('character-modal-close');
    const characterModalContent = document.getElementById('character-modal-content');
    const partyMembers = document.querySelectorAll('.party-member');

    // Initialize character sheet functionality
    function initCharacterSheet() {
        if (!characterModal) return;

        // Add click handlers to party members
        partyMembers.forEach(member => {
            member.addEventListener('click', () => {
                const characterName = member.dataset.character;
                if (characterName) {
                    loadCharacterSheet(characterName);
                }
            });
        });

        // Close modal handlers
        if (characterModalClose) {
            characterModalClose.addEventListener('click', closeCharacterModal);
        }
        if (characterModalOverlay) {
            characterModalOverlay.addEventListener('click', closeCharacterModal);
        }

        // Keyboard navigation
        document.addEventListener('keydown', handleCharacterModalKeyboard);
    }

    // Load character sheet via AJAX
    async function loadCharacterSheet(characterName) {
        if (!characterModalContent) return;

        // Show modal with loading state
        characterModalContent.innerHTML = '<div class="cs-loading">Chargement...</div>';
        characterModal.classList.add('active');
        document.body.style.overflow = 'hidden';

        try {
            const response = await fetch(`/play/${slug}/character/${encodeURIComponent(characterName)}`);
            if (!response.ok) {
                throw new Error('Failed to load character sheet');
            }

            const html = await response.text();
            characterModalContent.innerHTML = html;
        } catch (error) {
            console.error('Character sheet error:', error);
            characterModalContent.innerHTML = `
                <div class="cs-error">
                    <p>Erreur lors du chargement de la fiche</p>
                    <p class="cs-error-details">${escapeHtml(error.message)}</p>
                </div>
            `;
        }
    }

    // Close character modal
    function closeCharacterModal() {
        if (!characterModal) return;
        characterModal.classList.remove('active');
        document.body.style.overflow = '';
    }

    // Handle keyboard events for modal
    function handleCharacterModalKeyboard(e) {
        if (!characterModal?.classList.contains('active')) return;

        if (e.key === 'Escape') {
            closeCharacterModal();
        }
    }

    // Utility: escape HTML
    function escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // Initialize when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initCharacterSheet);
    } else {
        initCharacterSheet();
    }
})();

// ============================================================
// Gallery Module
// ============================================================
(function() {
    'use strict';

    // Only run on game pages with gallery
    if (typeof slug === 'undefined') {
        return;
    }

    const galleryPanel = document.getElementById('gallery-panel');
    const galleryToggle = document.getElementById('gallery-toggle');
    const galleryGrid = document.getElementById('gallery-grid');
    const galleryTabs = document.querySelectorAll('.gallery-tab');
    const gameContainer = document.querySelector('.game-container');

    const lightbox = document.getElementById('lightbox');
    const lightboxImage = document.getElementById('lightbox-image');
    const lightboxCaption = document.getElementById('lightbox-caption');
    const lightboxClose = document.getElementById('lightbox-close');
    const lightboxPrev = document.getElementById('lightbox-prev');
    const lightboxNext = document.getElementById('lightbox-next');
    const lightboxOverlay = lightbox?.querySelector('.lightbox-overlay');

    let allImages = [];
    let filteredImages = [];
    let currentImageIndex = 0;
    let currentCategory = 'all';

    // Initialize gallery
    function initGallery() {
        if (!galleryPanel) return;

        // Toggle panel
        if (galleryToggle) {
            galleryToggle.addEventListener('click', toggleGallery);
        }

        // Tabs
        galleryTabs.forEach(tab => {
            tab.addEventListener('click', () => {
                const category = tab.dataset.category;
                setActiveTab(category);
                filterImages(category);
            });
        });

        // Lightbox controls
        if (lightboxClose) {
            lightboxClose.addEventListener('click', closeLightbox);
        }
        if (lightboxOverlay) {
            lightboxOverlay.addEventListener('click', closeLightbox);
        }
        if (lightboxPrev) {
            lightboxPrev.addEventListener('click', showPrevImage);
        }
        if (lightboxNext) {
            lightboxNext.addEventListener('click', showNextImage);
        }

        // Keyboard navigation
        document.addEventListener('keydown', handleKeyboard);

        // Load images
        loadGalleryImages();

        // Refresh gallery when new images are generated
        document.body.addEventListener('refreshInfo', loadGalleryImages);
    }

    // Toggle gallery panel visibility
    function toggleGallery() {
        galleryPanel.classList.toggle('collapsed');
        gameContainer.classList.toggle('gallery-collapsed');
    }

    // Set active tab
    function setActiveTab(category) {
        currentCategory = category;
        galleryTabs.forEach(tab => {
            tab.classList.toggle('active', tab.dataset.category === category);
        });
    }

    // Load images from API
    async function loadGalleryImages() {
        if (!galleryGrid) return;

        galleryGrid.innerHTML = '<div class="gallery-loading">Chargement...</div>';

        try {
            const response = await fetch(`/play/${slug}/gallery`);
            if (!response.ok) {
                throw new Error('Failed to load gallery');
            }

            const data = await response.json();
            allImages = data.images || [];
            filterImages(currentCategory);
        } catch (error) {
            console.error('Gallery error:', error);
            galleryGrid.innerHTML = '<div class="gallery-empty">Erreur de chargement</div>';
        }
    }

    // Filter images by category
    function filterImages(category) {
        if (category === 'all') {
            filteredImages = allImages;
        } else {
            filteredImages = allImages.filter(img => img.category === category);
        }
        renderGallery();
    }

    // Render gallery grid
    function renderGallery() {
        if (!galleryGrid) return;

        if (filteredImages.length === 0) {
            galleryGrid.innerHTML = '<div class="gallery-empty">Aucune image</div>';
            return;
        }

        galleryGrid.innerHTML = filteredImages.map((img, index) => `
            <div class="gallery-item" data-index="${index}">
                <img src="${img.thumbnail}" alt="${escapeAttr(img.title)}" loading="lazy">
                <span class="gallery-item-badge ${img.category}">${img.category === 'session' ? 'S' + img.session : 'Carte'}</span>
                <div class="gallery-item-overlay">
                    <span class="gallery-item-title">${escapeHtml(img.title)}</span>
                </div>
            </div>
        `).join('');

        // Add click handlers
        galleryGrid.querySelectorAll('.gallery-item').forEach(item => {
            item.addEventListener('click', () => {
                const index = parseInt(item.dataset.index, 10);
                openLightbox(index);
            });
        });
    }

    // Open lightbox
    function openLightbox(index) {
        if (!lightbox || !filteredImages[index]) return;

        currentImageIndex = index;
        const img = filteredImages[index];

        lightboxImage.src = img.url;
        lightboxImage.alt = img.title;
        lightboxCaption.textContent = img.title;

        lightbox.classList.add('active');
        document.body.style.overflow = 'hidden';

        updateNavButtons();
    }

    // Close lightbox
    function closeLightbox() {
        if (!lightbox) return;

        lightbox.classList.remove('active');
        document.body.style.overflow = '';
    }

    // Show previous image
    function showPrevImage() {
        if (currentImageIndex > 0) {
            openLightbox(currentImageIndex - 1);
        }
    }

    // Show next image
    function showNextImage() {
        if (currentImageIndex < filteredImages.length - 1) {
            openLightbox(currentImageIndex + 1);
        }
    }

    // Update navigation buttons visibility
    function updateNavButtons() {
        if (lightboxPrev) {
            lightboxPrev.style.visibility = currentImageIndex > 0 ? 'visible' : 'hidden';
        }
        if (lightboxNext) {
            lightboxNext.style.visibility = currentImageIndex < filteredImages.length - 1 ? 'visible' : 'hidden';
        }
    }

    // Keyboard navigation
    function handleKeyboard(e) {
        if (!lightbox?.classList.contains('active')) return;

        switch (e.key) {
            case 'Escape':
                closeLightbox();
                break;
            case 'ArrowLeft':
                showPrevImage();
                break;
            case 'ArrowRight':
                showNextImage();
                break;
        }
    }

    // Utility: escape HTML
    function escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // Utility: escape attribute value
    function escapeAttr(text) {
        return text.replace(/"/g, '&quot;').replace(/'/g, '&#39;');
    }

    // Initialize when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initGallery);
    } else {
        initGallery();
    }
})();
