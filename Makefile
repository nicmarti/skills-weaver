# SkillsWeaver - Makefile
# Moteur de jeu de rôle Basic Fantasy RPG avec Claude Code

# Configuration
BINARY_PREFIX = sw
BINARIES = $(BINARY_PREFIX)-dice $(BINARY_PREFIX)-character $(BINARY_PREFIX)-adventure \
           $(BINARY_PREFIX)-names $(BINARY_PREFIX)-npc $(BINARY_PREFIX)-image \
           $(BINARY_PREFIX)-monster $(BINARY_PREFIX)-treasure $(BINARY_PREFIX)-dm \
           $(BINARY_PREFIX)-equipment $(BINARY_PREFIX)-spell $(BINARY_PREFIX)-rebuild-journal \
           $(BINARY_PREFIX)-location-names

# Go commands
GOCMD = go
GOBUILD = $(GOCMD) build
GOTEST = $(GOCMD) test
GOVET = $(GOCMD) vet
GOFMT = $(GOCMD) fmt
GOMOD = $(GOCMD) mod

# Build flags
LDFLAGS = -s -w

.PHONY: all build test test-verbose test-coverage lint fmt clean install tidy help

# =============================================================================
# Build targets
# =============================================================================

all: build ## Build par défaut

build: $(BINARIES) ## Compile tous les binaires sw-*

$(BINARY_PREFIX)-dice: cmd/dice/main.go internal/dice/*.go
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $@ ./cmd/dice

$(BINARY_PREFIX)-character: cmd/character/main.go internal/character/*.go internal/data/*.go
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $@ ./cmd/character

$(BINARY_PREFIX)-adventure: cmd/adventure/main.go internal/adventure/*.go internal/ai/*.go
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $@ ./cmd/adventure

$(BINARY_PREFIX)-names: cmd/names/main.go internal/names/*.go
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $@ ./cmd/names

$(BINARY_PREFIX)-npc: cmd/npc/main.go internal/npc/*.go internal/names/*.go
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $@ ./cmd/npc

$(BINARY_PREFIX)-image: cmd/image/main.go internal/image/*.go
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $@ ./cmd/image

$(BINARY_PREFIX)-monster: cmd/monster/main.go internal/monster/*.go internal/dice/*.go
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $@ ./cmd/monster

$(BINARY_PREFIX)-treasure: cmd/treasure/main.go internal/treasure/*.go internal/dice/*.go
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $@ ./cmd/treasure

$(BINARY_PREFIX)-dm: cmd/dm/main.go internal/agent/*.go internal/dmtools/*.go internal/ui/*.go
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $@ ./cmd/dm

$(BINARY_PREFIX)-equipment: cmd/equipment/main.go internal/equipment/*.go
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $@ ./cmd/equipment

$(BINARY_PREFIX)-spell: cmd/spell/main.go internal/spell/*.go
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $@ ./cmd/spell

$(BINARY_PREFIX)-rebuild-journal: cmd/rebuild-journal/main.go internal/adventure/*.go
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $@ ./cmd/rebuild-journal

$(BINARY_PREFIX)-location-names: cmd/location-names/main.go internal/locations/*.go
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $@ ./cmd/location-names

# =============================================================================
# Maintenance targets
# =============================================================================

rebuild-journal: $(BINARY_PREFIX)-rebuild-journal ## Reconstruit les événements perdus du journal
	./$(BINARY_PREFIX)-rebuild-journal

# =============================================================================
# Test targets
# =============================================================================

test: ## Lance tous les tests
	$(GOTEST) ./...

test-verbose: ## Lance les tests avec détails
	$(GOTEST) -v ./...

test-coverage: ## Lance les tests avec rapport de couverture
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Rapport de couverture: coverage.html"

# =============================================================================
# Code quality targets
# =============================================================================

lint: ## Vérifie le code (go vet + staticcheck si installé)
	$(GOVET) ./...
	@if command -v staticcheck > /dev/null 2>&1; then \
		staticcheck ./...; \
	else \
		echo "staticcheck non installé. Installer avec: go install honnef.co/go/tools/cmd/staticcheck@latest"; \
	fi

fmt: ## Formate le code source
	$(GOFMT) ./...

tidy: ## Nettoie les dépendances go.mod
	$(GOMOD) tidy

# =============================================================================
# Utility targets
# =============================================================================

clean: ## Supprime les binaires et fichiers temporaires
	rm -f $(BINARIES)
	rm -f coverage.out coverage.html

install: build ## Installe les binaires dans GOPATH/bin
	@if [ -z "$(GOPATH)" ]; then \
		echo "GOPATH non défini, installation dans ~/go/bin"; \
		cp $(BINARIES) ~/go/bin/; \
	else \
		cp $(BINARIES) $(GOPATH)/bin/; \
	fi
	@echo "Binaires installés"

# =============================================================================
# Development helpers
# =============================================================================

run-dice: $(BINARY_PREFIX)-dice ## Compile et lance sw-dice
	./$(BINARY_PREFIX)-dice $(ARGS)

run-character: $(BINARY_PREFIX)-character ## Compile et lance sw-character
	./$(BINARY_PREFIX)-character $(ARGS)

run-adventure: $(BINARY_PREFIX)-adventure ## Compile et lance sw-adventure
	./$(BINARY_PREFIX)-adventure $(ARGS)

# =============================================================================
# Help
# =============================================================================

help: ## Affiche cette aide
	@echo "SkillsWeaver - Commandes disponibles:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Exemples:"
	@echo "  make                  # Compile tous les binaires"
	@echo "  make test             # Lance les tests"
	@echo "  make lint             # Vérifie le code"
	@echo "  make clean            # Nettoie les binaires"
	@echo "  make run-dice ARGS='roll 2d6+3'"
