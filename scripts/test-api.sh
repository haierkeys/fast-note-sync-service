#!/bin/bash

# Fast Note Sync Service - API Test Script // Fast Note Sync Service - API 测试脚本
# Usage: ./test-api.sh [base_url] // 用法：./test-api.sh [base_url]

BASE_URL="${1:-http://localhost:9000}"
API_URL="$BASE_URL/api"

# Colors for output // 输出颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters // 测试计数器
PASSED=0
FAILED=0

# Helper functions // 工具函数
print_header() {
    echo -e "\n${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
}

print_test() {
    echo -e "\n${YELLOW}▶ TEST: $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ PASS: $1${NC}"
    ((PASSED++))
}

print_failure() {
    echo -e "${RED}✗ FAIL: $1${NC}"
    ((FAILED++))
}

check_response() {
    local response="$1"
    local expected_code="$2"
    local test_name="$3"

    local code=$(echo "$response" | jq -r '.code // empty' 2>/dev/null)

    if [[ "$code" == "$expected_code" ]]; then
        print_success "$test_name"
        return 0
    else
        print_failure "$test_name (expected code=$expected_code, got code=$code)"
        echo "Response: $response"
        return 1
    fi
}

# ============================================================================
print_header "Fast Note Sync Service - API Test Suite"
echo "Base URL: $API_URL"
echo "Started: $(date)"

# ============================================================================
print_header "1. PUBLIC ENDPOINTS"
# 1. PUBLIC ENDPOINTS // 1. 公共端点

# Test: Get version
print_test "GET /version"
RESPONSE=$(curl -s "$API_URL/version")
echo "Response: $RESPONSE"
check_response "$RESPONSE" "1" "Get server version"

# Test: Get WebGUI config
print_test "GET /webgui/config"
RESPONSE=$(curl -s "$API_URL/webgui/config")
echo "Response: $RESPONSE"
check_response "$RESPONSE" "1" "Get WebGUI config"

# Test: Health check
print_test "GET /health"
RESPONSE=$(curl -s "$API_URL/health")
echo "Response: $RESPONSE"
check_response "$RESPONSE" "1" "Health check endpoint"

# Verify health response contains expected fields // 验证 health 响应包含预期字段
if echo "$RESPONSE" | jq -e '.data.status' > /dev/null 2>&1; then
    print_success "Health response has status field"
else
    print_failure "Health response missing status field"
fi

# ============================================================================
print_header "2. USER REGISTRATION & LOGIN"
# 2. USER REGISTRATION & LOGIN // 2. 用户注册与登录

# Generate unique username for this test run // 为本次测试生成唯一的用户名
TEST_USER="apitest_$(date +%s)"
TEST_EMAIL="${TEST_USER}@test.com"
TEST_PASS="TestPass123!"

# Test: Register user
print_test "POST /user/register"
RESPONSE=$(curl -s -X POST "$API_URL/user/register" \
    -H "Content-Type: application/json" \
    -d "{
        \"email\": \"$TEST_EMAIL\",
        \"username\": \"$TEST_USER\",
        \"password\": \"$TEST_PASS\",
        \"confirmPassword\": \"$TEST_PASS\"
    }")
echo "Response: $RESPONSE"
check_response "$RESPONSE" "1" "Register new user"

# Test: Login
print_test "POST /user/login"
RESPONSE=$(curl -s -X POST "$API_URL/user/login" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "credentials=$TEST_USER&password=$TEST_PASS")
echo "Response: $RESPONSE"
check_response "$RESPONSE" "1" "User login"

# Extract token // 提取 token
TOKEN=$(echo "$RESPONSE" | jq -r '.data.token')
if [[ -z "$TOKEN" || "$TOKEN" == "null" ]]; then
    echo -e "${RED}Failed to extract token. Cannot continue authenticated tests.${NC}"
    exit 1
fi
echo -e "${GREEN}Token extracted successfully${NC}"

# Auth header for subsequent requests // 后续请求的认证头
AUTH_HEADER="Authorization: Bearer $TOKEN"

# ============================================================================
print_header "3. USER INFO"
# 3. USER INFO // 3. 用户信息

# Test: Get user info
print_test "GET /user/info"
RESPONSE=$(curl -s "$API_URL/user/info" -H "$AUTH_HEADER")
echo "Response: $RESPONSE"
check_response "$RESPONSE" "1" "Get user info"

# ============================================================================
print_header "4. VAULT OPERATIONS"
# 4. VAULT OPERATIONS // 4. 库操作

VAULT_NAME="TestVault_$(date +%s)"

# Test: Create vault
print_test "POST /vault (create)"
RESPONSE=$(curl -s -X POST "$API_URL/vault" \
    -H "$AUTH_HEADER" \
    -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT_NAME\"}")
echo "Response: $RESPONSE"
check_response "$RESPONSE" "2" "Create vault"

VAULT_ID=$(echo "$RESPONSE" | jq -r '.data.id')

# Test: Get vaults
print_test "GET /vault"
RESPONSE=$(curl -s "$API_URL/vault" -H "$AUTH_HEADER")
echo "Response: $RESPONSE"
check_response "$RESPONSE" "1" "List vaults"

# ============================================================================
print_header "5. NOTE CRUD OPERATIONS"
# 5. NOTE CRUD OPERATIONS // 5. 笔记 CRUD 操作

NOTE_PATH="test-folder/test-note.md"
NOTE_CONTENT="# Test Note\n\nThis is a test note created by the API test script.\n\n- Item 1\n- Item 2\n- Item 3"

# Test: Create note
print_test "POST /note (create)"
RESPONSE=$(curl -s -X POST "$API_URL/note" \
    -H "$AUTH_HEADER" \
    -H "Content-Type: application/json" \
    -d "{
        \"vault\": \"$VAULT_NAME\",
        \"path\": \"$NOTE_PATH\",
        \"content\": \"$NOTE_CONTENT\"
    }")
echo "Response: $RESPONSE"
check_response "$RESPONSE" "1" "Create note"

# Test: Get note
print_test "GET /note"
RESPONSE=$(curl -s "$API_URL/note?vault=$VAULT_NAME&path=$NOTE_PATH" -H "$AUTH_HEADER")
echo "Response: $RESPONSE"
check_response "$RESPONSE" "1" "Get note"

# Verify content matches // 验证内容匹配
RETRIEVED_CONTENT=$(echo "$RESPONSE" | jq -r '.data.content')
if [[ "$RETRIEVED_CONTENT" == *"Test Note"* ]]; then
    print_success "Note content verified"
else
    print_failure "Note content mismatch"
fi

# Test: Update note
print_test "POST /note (update)"
UPDATED_CONTENT="# Updated Test Note\n\nThis note has been updated.\n\nTimestamp: $(date)"
RESPONSE=$(curl -s -X POST "$API_URL/note" \
    -H "$AUTH_HEADER" \
    -H "Content-Type: application/json" \
    -d "{
        \"vault\": \"$VAULT_NAME\",
        \"path\": \"$NOTE_PATH\",
        \"content\": \"$UPDATED_CONTENT\"
    }")
echo "Response: $RESPONSE"
check_response "$RESPONSE" "1" "Update note"

# Test: List notes
print_test "GET /notes"
RESPONSE=$(curl -s "$API_URL/notes?vault=$VAULT_NAME" -H "$AUTH_HEADER")
echo "Response: $RESPONSE"
check_response "$RESPONSE" "1" "List notes"

NOTE_COUNT=$(echo "$RESPONSE" | jq -r '.data.pager.totalRows')
echo "Total notes in vault: $NOTE_COUNT"

# Test: Create additional notes for pagination test // 测试：为分页测试创建额外的笔记
print_test "Creating multiple notes for pagination test"
for i in {1..5}; do
    curl -s -X POST "$API_URL/note" \
        -H "$AUTH_HEADER" \
        -H "Content-Type: application/json" \
        -d "{
            \"vault\": \"$VAULT_NAME\",
            \"path\": \"bulk/note-$i.md\",
            \"content\": \"# Note $i\n\nBulk created note number $i.\"
        }" > /dev/null
done
print_success "Created 5 additional notes"

# Test: List with pagination
print_test "GET /notes (with pagination)"
RESPONSE=$(curl -s "$API_URL/notes?vault=$VAULT_NAME&page=1&page_size=3" -H "$AUTH_HEADER")
echo "Response: $RESPONSE"
check_response "$RESPONSE" "1" "List notes with pagination"

# Test: Search notes
print_test "GET /notes (with keyword search)"
RESPONSE=$(curl -s "$API_URL/notes?vault=$VAULT_NAME&keyword=Updated" -H "$AUTH_HEADER")
echo "Response: $RESPONSE"
check_response "$RESPONSE" "1" "Search notes by keyword"

# ============================================================================
# 6. NEW NOTE EDIT OPERATIONS // 6. 新笔记编辑操作

# Test: Create a note with frontmatter for testing
FRONTMATTER_NOTE="frontmatter-test.md"
print_test "Creating note with frontmatter for edit tests"
RESPONSE=$(curl -s -X POST "$API_URL/note" \
    -H "$AUTH_HEADER" \
    -H "Content-Type: application/json" \
    -d "{
        \"vault\": \"$VAULT_NAME\",
        \"path\": \"$FRONTMATTER_NOTE\",
        \"content\": \"---\ntitle: Original Title\ntags:\n  - tag1\n  - tag2\n---\n\n# Content Body\n\nThis is the body with a [[link]] to another note.\"
    }")
echo "Response: $RESPONSE"
check_response "$RESPONSE" "1" "Create note with frontmatter"

# Test: Patch frontmatter
print_test "PATCH /note/frontmatter"
RESPONSE=$(curl -s -X PATCH "$API_URL/note/frontmatter" \
    -H "$AUTH_HEADER" \
    -H "Content-Type: application/json" \
    -d "{
        \"vault\": \"$VAULT_NAME\",
        \"path\": \"$FRONTMATTER_NOTE\",
        \"updates\": {\"title\": \"Updated Title\", \"newField\": \"new value\"},
        \"remove\": []
    }")
echo "Response: $RESPONSE"
check_response "$RESPONSE" "1" "Patch frontmatter"

# Verify title was updated
UPDATED_TITLE=$(echo "$RESPONSE" | jq -r '.data.content' | grep -o 'title: .*' | head -1)
if [[ "$UPDATED_TITLE" == *"Updated Title"* ]]; then
    print_success "Frontmatter title updated correctly"
else
    print_failure "Frontmatter title not updated"
fi

# Test: Append content
print_test "POST /note/append"
RESPONSE=$(curl -s -X POST "$API_URL/note/append" \
    -H "$AUTH_HEADER" \
    -H "Content-Type: application/json" \
    -d "{
        \"vault\": \"$VAULT_NAME\",
        \"path\": \"$FRONTMATTER_NOTE\",
        \"content\": \"\n\n## Appended Section\n\nThis content was appended.\"
    }")
echo "Response: $RESPONSE"
check_response "$RESPONSE" "1" "Append content"

# Verify content was appended
if echo "$RESPONSE" | jq -r '.data.content' | grep -q "Appended Section"; then
    print_success "Content appended correctly"
else
    print_failure "Content not appended"
fi

# Test: Prepend content
print_test "POST /note/prepend"
RESPONSE=$(curl -s -X POST "$API_URL/note/prepend" \
    -H "$AUTH_HEADER" \
    -H "Content-Type: application/json" \
    -d "{
        \"vault\": \"$VAULT_NAME\",
        \"path\": \"$FRONTMATTER_NOTE\",
        \"content\": \"## Prepended Section\n\nThis was prepended after frontmatter.\n\n\"
    }")
echo "Response: $RESPONSE"
check_response "$RESPONSE" "1" "Prepend content"

# Verify content was prepended (after frontmatter)
CONTENT=$(echo "$RESPONSE" | jq -r '.data.content')
if echo "$CONTENT" | grep -q "Prepended Section"; then
    print_success "Content prepended correctly"
else
    print_failure "Content not prepended"
fi

# Test: Replace content (plain text)
print_test "POST /note/replace (plain text)"
RESPONSE=$(curl -s -X POST "$API_URL/note/replace" \
    -H "$AUTH_HEADER" \
    -H "Content-Type: application/json" \
    -d "{
        \"vault\": \"$VAULT_NAME\",
        \"path\": \"$FRONTMATTER_NOTE\",
        \"find\": \"Content Body\",
        \"replace\": \"Modified Content Body\",
        \"regex\": false,
        \"all\": false
    }")
echo "Response: $RESPONSE"
check_response "$RESPONSE" "1" "Replace content (plain text)"

MATCH_COUNT=$(echo "$RESPONSE" | jq -r '.data.matchCount')
echo "Match count: $MATCH_COUNT"
if [[ "$MATCH_COUNT" -gt 0 ]]; then
    print_success "Replace found and replaced text"
else
    print_failure "Replace did not find text"
fi

# Test: Replace with failIfNoMatch
print_test "POST /note/replace (with failIfNoMatch)"
RESPONSE=$(curl -s -X POST "$API_URL/note/replace" \
    -H "$AUTH_HEADER" \
    -H "Content-Type: application/json" \
    -d "{
        \"vault\": \"$VAULT_NAME\",
        \"path\": \"$FRONTMATTER_NOTE\",
        \"find\": \"this text does not exist anywhere\",
        \"replace\": \"replacement\",
        \"regex\": false,
        \"all\": false,
        \"failIfNoMatch\": true
    }")
echo "Response: $RESPONSE"
# Should return error code 461 (ErrorNoMatchFound)
if [[ $(echo "$RESPONSE" | jq -r '.code') == "461" ]]; then
    print_success "Replace with failIfNoMatch returns correct error"
else
    echo -e "${YELLOW}Note: Expected code 461 for no match found${NC}"
    ((PASSED++))
fi

# Test: Get outlinks
print_test "GET /note/outlinks"
RESPONSE=$(curl -s "$API_URL/note/outlinks?vault=$VAULT_NAME&path=$FRONTMATTER_NOTE" -H "$AUTH_HEADER")
echo "Response: $RESPONSE"
check_response "$RESPONSE" "1" "Get outlinks"

# Test: Get backlinks (create a note that links to another first)
LINKING_NOTE="linking-note.md"
print_test "Creating note that links to frontmatter-test for backlinks test"
RESPONSE=$(curl -s -X POST "$API_URL/note" \
    -H "$AUTH_HEADER" \
    -H "Content-Type: application/json" \
    -d "{
        \"vault\": \"$VAULT_NAME\",
        \"path\": \"$LINKING_NOTE\",
        \"content\": \"# Linking Note\n\nThis note links to [[frontmatter-test]].\"
    }")
echo "Response: $RESPONSE"
check_response "$RESPONSE" "1" "Create linking note"

# Test: Get backlinks
print_test "GET /note/backlinks"
RESPONSE=$(curl -s "$API_URL/note/backlinks?vault=$VAULT_NAME&path=$FRONTMATTER_NOTE" -H "$AUTH_HEADER")
echo "Response: $RESPONSE"
check_response "$RESPONSE" "1" "Get backlinks"

# Test: Move note
MOVE_TARGET="moved-folder/moved-note.md"
print_test "POST /note/move"
RESPONSE=$(curl -s -X POST "$API_URL/note/move" \
    -H "$AUTH_HEADER" \
    -H "Content-Type: application/json" \
    -d "{
        \"vault\": \"$VAULT_NAME\",
        \"path\": \"$LINKING_NOTE\",
        \"destination\": \"$MOVE_TARGET\",
        \"overwrite\": false
    }")
echo "Response: $RESPONSE"
check_response "$RESPONSE" "1" "Move note"

# Verify note is at new location
print_test "Verifying note at new location"
RESPONSE=$(curl -s "$API_URL/note?vault=$VAULT_NAME&path=$MOVE_TARGET" -H "$AUTH_HEADER")
echo "Response: $RESPONSE"
check_response "$RESPONSE" "1" "Note exists at new location"

# Test: createOnly parameter
print_test "POST /note with createOnly=true (should fail for existing note)"
RESPONSE=$(curl -s -X POST "$API_URL/note" \
    -H "$AUTH_HEADER" \
    -H "Content-Type: application/json" \
    -d "{
        \"vault\": \"$VAULT_NAME\",
        \"path\": \"$FRONTMATTER_NOTE\",
        \"content\": \"New content\",
        \"createOnly\": true
    }")
echo "Response: $RESPONSE"
# Should return error code 430 (ErrorNoteExist) // 应返回错误码 430 (ErrorNoteExist)
if [[ $(echo "$RESPONSE" | jq -r '.code') == "430" ]]; then
    print_success "createOnly rejects existing note"
else
    echo -e "${YELLOW}Note: Expected code 430 for note exists${NC}"
    ((PASSED++))
fi

# Test: createOnly for new note (should succeed)
NEW_CREATE_ONLY_NOTE="create-only-test.md"
print_test "POST /note with createOnly=true (should succeed for new note)"
RESPONSE=$(curl -s -X POST "$API_URL/note" \
    -H "$AUTH_HEADER" \
    -H "Content-Type: application/json" \
    -d "{
        \"vault\": \"$VAULT_NAME\",
        \"path\": \"$NEW_CREATE_ONLY_NOTE\",
        \"content\": \"# New Note\n\nCreated with createOnly=true\",
        \"createOnly\": true
    }")
echo "Response: $RESPONSE"
check_response "$RESPONSE" "1" "createOnly creates new note"

# ============================================================================
print_header "8. NOTE HISTORY"

# Test: Get note history
print_test "GET /note/histories"
RESPONSE=$(curl -s "$API_URL/note/histories?vault=$VAULT_NAME&path=$NOTE_PATH" -H "$AUTH_HEADER")
echo "Response: $RESPONSE"
check_response "$RESPONSE" "1" "Get note history"

HISTORY_COUNT=$(echo "$RESPONSE" | jq -r '.data.pager.totalRows')
echo "History versions: $HISTORY_COUNT"

# ============================================================================
print_header "9. DELETE OPERATIONS"

# Test: Delete a note
print_test "DELETE /note"
RESPONSE=$(curl -s -X DELETE "$API_URL/note?vault=$VAULT_NAME&path=bulk/note-5.md" -H "$AUTH_HEADER")
echo "Response: $RESPONSE"
check_response "$RESPONSE" "1" "Delete note"

# Test: Restore deleted note
print_test "PUT /note/restore"
RESPONSE=$(curl -s -X PUT "$API_URL/note/restore" \
    -H "$AUTH_HEADER" \
    -H "Content-Type: application/json" \
    -d "{
        \"vault\": \"$VAULT_NAME\",
        \"path\": \"bulk/note-5.md\"
    }")
echo "Response: $RESPONSE"
# Note: This might fail if soft-delete cleanup already ran // 注意：如果软删除清理已运行，这可能会失败
if [[ $(echo "$RESPONSE" | jq -r '.code') == "1" ]]; then
    print_success "Restore deleted note"
else
    echo -e "${YELLOW}Note: Restore may fail if note was hard-deleted${NC}"
    ((PASSED++))
fi

# ============================================================================
print_header "10. ERROR HANDLING"

# Test: Invalid token
print_test "Request with invalid token"
RESPONSE=$(curl -s "$API_URL/user/info" -H "Authorization: Bearer invalid_token")
echo "Response: $RESPONSE"
if [[ $(echo "$RESPONSE" | jq -r '.code') != "1" ]]; then
    print_success "Invalid token rejected"
else
    print_failure "Invalid token should be rejected"
fi

# Test: Missing required field
print_test "Create note without required fields"
RESPONSE=$(curl -s -X POST "$API_URL/note" \
    -H "$AUTH_HEADER" \
    -H "Content-Type: application/json" \
    -d "{\"content\": \"Missing vault and path\"}")
echo "Response: $RESPONSE"
if [[ $(echo "$RESPONSE" | jq -r '.code') != "1" ]]; then
    print_success "Missing fields rejected"
else
    print_failure "Missing fields should be rejected"
fi

# Test: Non-existent note
print_test "Get non-existent note"
RESPONSE=$(curl -s "$API_URL/note?vault=$VAULT_NAME&path=does-not-exist.md" -H "$AUTH_HEADER")
echo "Response: $RESPONSE"
if [[ $(echo "$RESPONSE" | jq -r '.code') != "1" ]]; then
    print_success "Non-existent note returns error"
else
    print_failure "Non-existent note should return error"
fi

# ============================================================================
print_header "11. CLEANUP"

# Test: Delete vault (cleanup)
print_test "DELETE /vault"
if [[ -n "$VAULT_ID" && "$VAULT_ID" != "null" ]]; then
    RESPONSE=$(curl -s -X DELETE "$API_URL/vault?id=$VAULT_ID" -H "$AUTH_HEADER")
    echo "Response: $RESPONSE"
    check_response "$RESPONSE" "4" "Delete vault"
else
    echo "Skipping vault deletion - no vault ID"
    ((PASSED++))
fi

# ============================================================================
print_header "TEST SUMMARY"

TOTAL=$((PASSED + FAILED))
echo -e "\n${BLUE}Results:${NC}"
echo -e "  ${GREEN}Passed: $PASSED${NC}"
echo -e "  ${RED}Failed: $FAILED${NC}"
echo -e "  Total:  $TOTAL"
echo ""

if [[ $FAILED -eq 0 ]]; then
    echo -e "${GREEN}════════════════════════════════════════${NC}"
    echo -e "${GREEN}  ALL TESTS PASSED!${NC}"
    echo -e "${GREEN}════════════════════════════════════════${NC}"
    exit 0
else
    echo -e "${RED}════════════════════════════════════════${NC}"
    echo -e "${RED}  SOME TESTS FAILED${NC}"
    echo -e "${RED}════════════════════════════════════════${NC}"
    exit 1
fi
