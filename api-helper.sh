#!/bin/bash

# Fast Note Sync API Helper
# Usage: ./api-helper.sh <command> [args...]

BASE_URL="${FNS_URL:-http://localhost:9000}"
VAULT="${FNS_VAULT:-fastNoteSyncTest}"
TOKEN_FILE="/tmp/.fns_token"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'

# Get or refresh token
get_token() {
    if [[ -f "$TOKEN_FILE" ]]; then
        cat "$TOKEN_FILE"
    else
        echo "Not logged in. Run: $0 login <user> <pass>" >&2
        exit 1
    fi
}

# API call helper
api() {
    local method="$1"
    local endpoint="$2"
    shift 2
    local token=$(get_token)

    curl -s -X "$method" "${BASE_URL}/api${endpoint}" \
        -H "Authorization: Bearer $token" \
        -H "Content-Type: application/json" \
        "$@"
}

case "$1" in
    login)
        if [[ -z "$2" || -z "$3" ]]; then
            echo "Usage: $0 login <username> <password>"
            exit 1
        fi
        RESPONSE=$(curl -s -X POST "${BASE_URL}/api/user/login" \
            -H "Content-Type: application/x-www-form-urlencoded" \
            -d "credentials=$2&password=$3")
        TOKEN=$(echo "$RESPONSE" | jq -r '.data.token // empty')
        if [[ -n "$TOKEN" && "$TOKEN" != "null" ]]; then
            echo "$TOKEN" > "$TOKEN_FILE"
            echo -e "${GREEN}Logged in as $2${NC}"
        else
            echo -e "${RED}Login failed${NC}"
            echo "$RESPONSE" | jq .
            exit 1
        fi
        ;;

    logout)
        rm -f "$TOKEN_FILE"
        echo "Logged out"
        ;;

    health)
        curl -s "${BASE_URL}/api/health" | jq .
        ;;

    vaults)
        api GET "/vault" | jq -r '.data[].vault'
        ;;

    notes|list)
        api GET "/notes?vault=${VAULT}&pageSize=100" | jq -r '.data.list[].path'
        ;;

    get)
        if [[ -z "$2" ]]; then
            echo "Usage: $0 get <path>"
            exit 1
        fi
        api GET "/note?vault=${VAULT}&path=$(printf '%s' "$2" | jq -sRr @uri)" | jq .
        ;;

    content)
        if [[ -z "$2" ]]; then
            echo "Usage: $0 content <path>"
            exit 1
        fi
        api GET "/note?vault=${VAULT}&path=$(printf "%s" "$2" | jq -sRr @uri)" | jq -r '.data.content'
        ;;

    create)
        if [[ -z "$2" || -z "$3" ]]; then
            echo "Usage: $0 create <path> <content>"
            exit 1
        fi
        api POST "/note" -d "$(jq -n --arg v "$VAULT" --arg p "$2" --arg c "$3" \
            '{vault: $v, path: $p, content: $c}')" | jq .
        ;;

    append)
        if [[ -z "$2" || -z "$3" ]]; then
            echo "Usage: $0 append <path> <content>"
            exit 1
        fi
        api POST "/note/append" -d "$(jq -n --arg v "$VAULT" --arg p "$2" --arg c "$3" \
            '{vault: $v, path: $p, content: $c}')" | jq .
        ;;

    prepend)
        if [[ -z "$2" || -z "$3" ]]; then
            echo "Usage: $0 prepend <path> <content>"
            exit 1
        fi
        api POST "/note/prepend" -d "$(jq -n --arg v "$VAULT" --arg p "$2" --arg c "$3" \
            '{vault: $v, path: $p, content: $c}')" | jq .
        ;;

    replace)
        if [[ -z "$2" || -z "$3" || -z "$4" ]]; then
            echo "Usage: $0 replace <path> <find> <replace> [--all] [--regex]"
            exit 1
        fi
        ALL="false"
        REGEX="false"
        [[ "$5" == "--all" || "$6" == "--all" ]] && ALL="true"
        [[ "$5" == "--regex" || "$6" == "--regex" ]] && REGEX="true"
        api POST "/note/replace" -d "$(jq -n --arg v "$VAULT" --arg p "$2" \
            --arg f "$3" --arg r "$4" --argjson a "$ALL" --argjson re "$REGEX" \
            '{vault: $v, path: $p, find: $f, replace: $r, all: $a, regex: $re}')" | jq .
        ;;

    move)
        if [[ -z "$2" || -z "$3" ]]; then
            echo "Usage: $0 move <source> <destination> [--overwrite]"
            exit 1
        fi
        OW="false"
        [[ "$4" == "--overwrite" ]] && OW="true"
        api POST "/note/move" -d "$(jq -n --arg v "$VAULT" --arg p "$2" \
            --arg d "$3" --argjson o "$OW" \
            '{vault: $v, path: $p, destination: $d, overwrite: $o}')" | jq .
        ;;

    frontmatter)
        if [[ -z "$2" || -z "$3" ]]; then
            echo "Usage: $0 frontmatter <path> '<json updates>'"
            echo "Example: $0 frontmatter note.md '{\"title\": \"New Title\"}'"
            exit 1
        fi
        api PATCH "/note/frontmatter" -d "$(jq -n --arg v "$VAULT" --arg p "$2" \
            --argjson u "$3" '{vault: $v, path: $p, updates: $u}')" | jq .
        ;;

    backlinks)
        if [[ -z "$2" ]]; then
            echo "Usage: $0 backlinks <path>"
            exit 1
        fi
        api GET "/note/backlinks?vault=${VAULT}&path=$(printf "%s" "$2" | jq -sRr @uri)" | jq .
        ;;

    outlinks)
        if [[ -z "$2" ]]; then
            echo "Usage: $0 outlinks <path>"
            exit 1
        fi
        api GET "/note/outlinks?vault=${VAULT}&path=$(printf "%s" "$2" | jq -sRr @uri)" | jq .
        ;;

    history)
        if [[ -z "$2" ]]; then
            echo "Usage: $0 history <path>"
            exit 1
        fi
        api GET "/note/histories?vault=${VAULT}&path=$(printf "%s" "$2" | jq -sRr @uri)" | jq .
        ;;

    delete)
        if [[ -z "$2" ]]; then
            echo "Usage: $0 delete <path>"
            exit 1
        fi
        api DELETE "/note?vault=${VAULT}&path=$(printf "%s" "$2" | jq -sRr @uri)" | jq .
        ;;

    folders)
        api GET "/folders?vault=${VAULT}&path=$(printf '%s' "${2:-}" | jq -sRr @uri)" | jq .
        ;;

    folder)
        if [[ -z "$2" ]]; then
            echo "Usage: $0 folder <path>"
            exit 1
        fi
        api GET "/folder?vault=${VAULT}&path=$(printf '%s' "$2" | jq -sRr @uri)" | jq .
        ;;

    folder-notes)
        if [[ -z "$2" ]]; then
            echo "Usage: $0 folder-notes <path>"
            exit 1
        fi
        api GET "/folder/notes?vault=${VAULT}&path=$(printf '%s' "$2" | jq -sRr @uri)" | jq .
        ;;

    folder-files)
        if [[ -z "$2" ]]; then
            echo "Usage: $0 folder-files <path>"
            exit 1
        fi
        api GET "/folder/files?vault=${VAULT}&path=$(printf '%s' "$2" | jq -sRr @uri)" | jq .
        ;;

    tree)
        DEPTH="${2:-}"
        if [[ -n "$DEPTH" ]]; then
            api GET "/folder/tree?vault=${VAULT}&depth=${DEPTH}" | jq .
        else
            api GET "/folder/tree?vault=${VAULT}" | jq .
        fi
        ;;

    search)
        if [[ -z "$2" ]]; then
            echo "Usage: $0 search <term>"
            exit 1
        fi
        echo -e "${CYAN}Searching for: $2${NC}"
        api GET "/notes?vault=${VAULT}&pageSize=100" | jq -r '.data.list[].path' | while IFS= read -r path; do
            content=$(api GET "/note?vault=${VAULT}&path=$(printf '%s' "$path" | jq -sRr @uri)" | jq -r '.data.content // empty')
            if echo "$content" | grep -qi "$2"; then
                echo -e "\n${GREEN}=== $path ===${NC}"
                echo "$content" | grep -i "$2"
            fi
        done
        ;;

    raw)
        # Raw API call: ./api-helper.sh raw GET "/endpoint?params"
        if [[ -z "$2" || -z "$3" ]]; then
            echo "Usage: $0 raw <METHOD> <endpoint>"
            exit 1
        fi
        api "$2" "$3" | jq .
        ;;

    *)
        echo "Fast Note Sync API Helper"
        echo ""
        echo "Environment:"
        echo "  FNS_URL   - API URL (default: http://localhost:9000)"
        echo "  FNS_VAULT - Vault name (default: fastNoteSyncTest)"
        echo ""
        echo "Commands:"
        echo "  login <user> <pass>     Login and store token"
        echo "  logout                  Clear stored token"
        echo "  health                  Health check"
        echo "  vaults                  List vaults"
        echo "  notes                   List notes in vault"
        echo "  get <path>              Get note (full JSON)"
        echo "  content <path>          Get note content only"
        echo "  create <path> <content> Create/update note"
        echo "  append <path> <content> Append to note"
        echo "  prepend <path> <content> Prepend to note"
        echo "  replace <path> <find> <replace> [--all] [--regex]"
        echo "  move <src> <dest> [--overwrite]"
        echo "  frontmatter <path> '<json>'"
        echo "  backlinks <path>        Get backlinks"
        echo "  outlinks <path>         Get outlinks"
        echo "  history <path>          Get version history"
        echo "  delete <path>           Delete note"
        echo "  folders [path]          List child folders (root if no path)"
        echo "  folder <path>           Get folder info"
        echo "  folder-notes <path>     List notes in folder"
        echo "  folder-files <path>     List files in folder"
        echo "  tree [depth]            Get folder tree (optional depth limit)"
        echo "  search <term>           Search all notes"
        echo "  raw <METHOD> <endpoint> Raw API call"
        ;;
esac
