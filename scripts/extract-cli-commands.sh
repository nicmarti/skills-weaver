#!/bin/bash
# Script pour extraire les commandes CLI du log sw-dm
# Usage: ./scripts/extract-cli-commands.sh [adventure-name] [tool-name]

set -euo pipefail

ADVENTURES_DIR="data/adventures"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

show_usage() {
    echo "Usage: $0 [adventure-name] [tool-name]"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Show all CLI commands from all adventures"
    echo "  $0 la-crypte-des-ombres              # Show all CLI commands from specific adventure"
    echo "  $0 la-crypte-des-ombres generate_map # Show only generate_map commands"
    echo ""
    echo "Available tools:"
    echo "  - generate_map"
    echo "  - generate_npc"
    echo "  - generate_image"
    echo "  - generate_treasure"
    echo "  - roll_dice"
    echo "  - get_monster"
    exit 1
}

# Parse arguments
ADVENTURE_NAME="${1:-}"
TOOL_NAME="${2:-}"

if [ "$ADVENTURE_NAME" = "-h" ] || [ "$ADVENTURE_NAME" = "--help" ]; then
    show_usage
fi

# Find log files
if [ -n "$ADVENTURE_NAME" ]; then
    LOG_FILE="$ADVENTURES_DIR/$ADVENTURE_NAME/sw-dm.log"
    if [ ! -f "$LOG_FILE" ]; then
        echo -e "${YELLOW}Error: Log file not found: $LOG_FILE${NC}"
        exit 1
    fi
    LOG_FILES=("$LOG_FILE")
else
    # Find all log files
    mapfile -t LOG_FILES < <(find "$ADVENTURES_DIR" -name "sw-dm.log" 2>/dev/null)
    if [ ${#LOG_FILES[@]} -eq 0 ]; then
        echo -e "${YELLOW}No sw-dm.log files found in $ADVENTURES_DIR${NC}"
        exit 1
    fi
fi

echo -e "${GREEN}=== CLI Commands Extractor ===${NC}"
echo ""

# Process each log file
for LOG_FILE in "${LOG_FILES[@]}"; do
    ADVENTURE=$(basename "$(dirname "$LOG_FILE")")

    echo -e "${BLUE}Adventure: $ADVENTURE${NC}"
    echo ""

    # Extract commands with context
    if [ -n "$TOOL_NAME" ]; then
        # Filter by specific tool
        grep -B 6 "Equivalent CLI:" "$LOG_FILE" | \
        grep -A 6 "TOOL CALL: $TOOL_NAME" | \
        while IFS= read -r line; do
            if [[ "$line" =~ ^\[.*\]\ TOOL\ CALL:\ (.*)\ \(ID: ]]; then
                echo -e "${YELLOW}${BASH_REMATCH[1]}${NC}"
            elif [[ "$line" =~ ^[[:space:]]*Equivalent\ CLI: ]]; then
                CLI_CMD="${line#*Equivalent CLI:}"
                echo "$CLI_CMD"
                echo ""
            fi
        done
    else
        # Show all commands
        grep -B 6 "Equivalent CLI:" "$LOG_FILE" | \
        while IFS= read -r line; do
            if [[ "$line" =~ ^\[.*\]\ TOOL\ CALL:\ (.*)\ \(ID: ]]; then
                echo -e "${YELLOW}${BASH_REMATCH[1]}${NC}"
            elif [[ "$line" =~ ^[[:space:]]*Equivalent\ CLI: ]]; then
                CLI_CMD="${line#*Equivalent CLI:}"
                echo "$CLI_CMD"
                echo ""
            fi
        done
    fi

    echo ""
done

echo -e "${GREEN}=== Summary ===${NC}"
echo ""

# Count by tool type
for LOG_FILE in "${LOG_FILES[@]}"; do
    ADVENTURE=$(basename "$(dirname "$LOG_FILE")")

    if [ -n "$TOOL_NAME" ]; then
        COUNT=$(grep -c "TOOL CALL: $TOOL_NAME" "$LOG_FILE" || echo "0")
        echo -e "$ADVENTURE: ${GREEN}$COUNT${NC} $TOOL_NAME calls"
    else
        echo -e "$ADVENTURE:"
        grep "TOOL CALL:" "$LOG_FILE" | \
        sed 's/.*TOOL CALL: \([^ ]*\).*/\1/' | \
        sort | uniq -c | \
        while read -r count tool; do
            echo -e "  - $tool: ${GREEN}$count${NC} calls"
        done
    fi
done
