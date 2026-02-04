#!/bin/bash

# Folder API Test Script
# Tests all folder-related and general API endpoints, saves outputs as JSON

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/.env"

URL="${url:-http://localhost:9000}"
TOKEN="${local_test_auth_token}"
VAULT="fastnotsyncTest"
OUT_DIR="$SCRIPT_DIR/test-outputs"

mkdir -p "$OUT_DIR"

PASS=0
FAIL=0

run_test() {
    local name="$1"
    local method="$2"
    local endpoint="$3"
    local body="${4:-}"
    local filename="$OUT_DIR/${name}.json"

    if [[ -n "$body" ]]; then
        RESPONSE=$(curl -s -X "$method" "${URL}/api${endpoint}" \
            -H "token: $TOKEN" \
            -H "Content-Type: application/json" \
            -d "$body")
    else
        RESPONSE=$(curl -s -X "$method" "${URL}/api${endpoint}" \
            -H "token: $TOKEN")
    fi

    echo "$RESPONSE" | jq . > "$filename" 2>/dev/null

    CODE=$(echo "$RESPONSE" | jq -r '.code // empty')
    if [[ "$CODE" == "1" || "$CODE" == "2" ]]; then
        echo "  PASS  $name"
        PASS=$((PASS + 1))
    else
        echo "  FAIL  $name (code=$CODE)"
        FAIL=$((FAIL + 1))
    fi
}

echo "=== Fast Note Sync API Tests ==="
echo "URL: $URL"
echo "Vault: $VAULT"
echo "Output: $OUT_DIR/"
echo ""

# --- Health & Version (no auth) ---
echo "--- Health & Version ---"
run_test "health" GET "/health"
run_test "version" GET "/version"

# --- User ---
echo "--- User ---"
run_test "user-info" GET "/user/info"

# --- Vaults ---
echo "--- Vaults ---"
run_test "vault-list" GET "/vault"

# --- Notes ---
echo "--- Notes ---"
run_test "notes-list" GET "/notes?vault=$VAULT&pageSize=100"
run_test "note-get" GET "/note?vault=$VAULT&path=test.md"

# --- Note Edit Operations ---
echo "--- Note Edit Operations ---"

# Create a test note
run_test "note-create" POST "/note" \
    "{\"vault\":\"$VAULT\",\"path\":\"_test/api-test-note.md\",\"content\":\"# Test Note\\n\\nOriginal content.\\n\\nLine to replace.\"}"

# Append
run_test "note-append" POST "/note/append" \
    "{\"vault\":\"$VAULT\",\"path\":\"_test/api-test-note.md\",\"content\":\"\\n\\nAppended content.\"}"

# Prepend
run_test "note-prepend" POST "/note/prepend" \
    "{\"vault\":\"$VAULT\",\"path\":\"_test/api-test-note.md\",\"content\":\"Prepended line.\\n\\n\"}"

# Replace
run_test "note-replace" POST "/note/replace" \
    "{\"vault\":\"$VAULT\",\"path\":\"_test/api-test-note.md\",\"find\":\"Line to replace\",\"replace\":\"Line was replaced\",\"all\":true}"

# Frontmatter
run_test "note-frontmatter" PATCH "/note/frontmatter" \
    "{\"vault\":\"$VAULT\",\"path\":\"_test/api-test-note.md\",\"updates\":{\"tags\":[\"test\",\"api\"],\"status\":\"draft\"}}"

# Get the note after edits
run_test "note-after-edits" GET "/note?vault=$VAULT&path=_test/api-test-note.md"

# Move
run_test "note-move" POST "/note/move" \
    "{\"vault\":\"$VAULT\",\"path\":\"_test/api-test-note.md\",\"destination\":\"_test/api-test-moved.md\"}"

# --- Links ---
echo "--- Links ---"
# Pick a note that likely has backlinks
run_test "backlinks" GET "/note/backlinks?vault=$VAULT&path=projects/test-backlinks/folder-a/note.md"
run_test "outlinks" GET "/note/outlinks?vault=$VAULT&path=projects/test-backlinks/folder-a/note.md"

# --- History ---
echo "--- History ---"
run_test "note-history" GET "/note/histories?vault=$VAULT&path=_test/api-test-moved.md"

# --- Folders ---
echo "--- Folders ---"
run_test "folder-list" GET "/folders?vault=$VAULT"
run_test "folder-get" GET "/folder?vault=$VAULT&path=projects"
run_test "folder-notes" GET "/folder/notes?vault=$VAULT&path=projects"
run_test "folder-files" GET "/folder/files?vault=$VAULT&path=projects"

# --- Folder Tree (our new endpoint) ---
echo "--- Folder Tree ---"
run_test "folder-tree" GET "/folder/tree?vault=$VAULT"
run_test "folder-tree-depth1" GET "/folder/tree?vault=$VAULT&depth=1"

# --- Cleanup ---
echo "--- Cleanup ---"
run_test "note-delete" DELETE "/note?vault=$VAULT&path=_test/api-test-moved.md"

# --- Summary ---
echo ""
echo "=== Results ==="
echo "  Passed: $PASS"
echo "  Failed: $FAIL"
echo "  Output: $OUT_DIR/"
echo ""
echo "JSON files:"
ls -1 "$OUT_DIR/"
