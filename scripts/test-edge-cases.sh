#!/bin/bash

# Fast Note Sync Service - Edge Case Tests // Fast Note Sync Service - 边界情况测试
# Critical tests for edge cases and potential bugs // 针对边界情况和潜在 bug 的关键测试

BASE_URL="${1:-http://localhost:9000}"
API_URL="$BASE_URL/api"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

PASSED=0
FAILED=0
WARNINGS=0

print_header() {
    echo -e "\n${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
}

print_test() {
    echo -e "\n${YELLOW}▶ TEST: $1${NC}"
}
# Success/Failure printers // 成功/失败打印器

print_success() {
    echo -e "${GREEN}✓ PASS: $1${NC}"
    ((PASSED++))
}

print_failure() {
    echo -e "${RED}✗ FAIL: $1${NC}"
    ((FAILED++))
}

print_warning() {
    echo -e "${YELLOW}⚠ WARNING: $1${NC}"
    ((WARNINGS++))
}

print_info() {
    echo -e "${CYAN}ℹ INFO: $1${NC}"
}

check_code() {
    local response="$1"
    local expected="$2"
    local test_name="$3"
    local code=$(echo "$response" | jq -r '.code // empty' 2>/dev/null)

    if [[ "$code" == "$expected" ]]; then
        print_success "$test_name"
        return 0
    else
        print_failure "$test_name (expected code=$expected, got code=$code)"
        return 1
    fi
}

# ============================================================================
# SETUP - Create test user and vault // 设置 - 创建测试用户和库

TEST_USER="edgetest_$(date +%s)"
TEST_EMAIL="${TEST_USER}@test.com"
TEST_PASS="TestPass123!"

# Register and login
RESPONSE=$(curl -s -X POST "$API_URL/user/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\": \"$TEST_EMAIL\", \"username\": \"$TEST_USER\", \"password\": \"$TEST_PASS\", \"confirmPassword\": \"$TEST_PASS\"}")

TOKEN=$(echo "$RESPONSE" | jq -r '.data.token')
if [[ -z "$TOKEN" || "$TOKEN" == "null" ]]; then
    echo -e "${RED}Failed to get token. Exiting.${NC}"
    exit 1
fi
AUTH="Authorization: Bearer $TOKEN"

VAULT="EdgeTestVault_$(date +%s)"
curl -s -X POST "$API_URL/vault" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\"}" > /dev/null

echo -e "${GREEN}Setup complete. Vault: $VAULT${NC}"

# ============================================================================
# 1. MOVE OPERATIONS - Edge Cases // 1. 移动操作 - 边界情况

# Create two notes for move testing
print_test "Setup: Create source and destination notes"
curl -s -X POST "$API_URL/note" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"move-source.md\", \"content\": \"# Source Note\\n\\nOriginal content.\"}" > /dev/null
curl -s -X POST "$API_URL/note" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"move-dest.md\", \"content\": \"# Destination Note\\n\\nThis should be overwritten.\"}" > /dev/null

# Edit destination multiple times to create version history
for i in 1 2 3; do
    curl -s -X POST "$API_URL/note" -H "$AUTH" -H "Content-Type: application/json" \
        -d "{\"vault\": \"$VAULT\", \"path\": \"move-dest.md\", \"content\": \"# Destination Note v$i\\n\\nEdit $i\"}" > /dev/null
done
print_success "Created notes with version history"

# Check destination version before move
print_test "Check destination note version before move"
RESPONSE=$(curl -s "$API_URL/note?vault=$VAULT&path=move-dest.md" -H "$AUTH")
DEST_VERSION=$(echo "$RESPONSE" | jq -r '.data.version')
echo "Destination version before move: $DEST_VERSION"
if [[ "$DEST_VERSION" -ge 2 ]]; then
    print_success "Destination has version history (v$DEST_VERSION)"
else
    print_warning "Expected version >= 2, got $DEST_VERSION"
fi

# Test: Move without overwrite (should fail)
print_test "Move to existing path WITHOUT overwrite flag"
RESPONSE=$(curl -s -X POST "$API_URL/note/move" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"move-source.md\", \"destination\": \"move-dest.md\", \"overwrite\": false}")
echo "Response: $RESPONSE"
check_code "$RESPONSE" "460" "Move blocked when destination exists"

# Test: Move WITH overwrite
print_test "Move to existing path WITH overwrite=true"
RESPONSE=$(curl -s -X POST "$API_URL/note/move" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"move-source.md\", \"destination\": \"move-dest.md\", \"overwrite\": true}")
echo "Response: $RESPONSE"
check_code "$RESPONSE" "1" "Move with overwrite succeeded"

# CRITICAL: Check if destination's old version history is preserved
print_test "CRITICAL: Check version history after move+overwrite"
RESPONSE=$(curl -s "$API_URL/note/histories?vault=$VAULT&path=move-dest.md" -H "$AUTH")
echo "Response: $RESPONSE"
HISTORY_COUNT=$(echo "$RESPONSE" | jq -r '.data.pager.totalRows')
echo "History entries after move: $HISTORY_COUNT"
# Note: History is created with delay, so may be 0 immediately
if [[ "$HISTORY_COUNT" -eq "0" ]]; then
    print_warning "No immediate history - history may be created async after delay"
else
    print_info "History count: $HISTORY_COUNT"
fi

# Test: Move to same path (no-op) // 测试：移动到相同路径（无操作）
print_test "Move note to itself (same source and destination)"
curl -s -X POST "$API_URL/note" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"self-move.md\", \"content\": \"Self move test\"}" > /dev/null
RESPONSE=$(curl -s -X POST "$API_URL/note/move" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"self-move.md\", \"destination\": \"self-move.md\", \"overwrite\": false}")
echo "Response: $RESPONSE"
# This might succeed (no-op) or fail - depends on implementation
CODE=$(echo "$RESPONSE" | jq -r '.code')
if [[ "$CODE" == "1" ]]; then
    print_info "Move to self succeeded (treated as no-op)"
elif [[ "$CODE" == "460" ]]; then
    print_info "Move to self blocked by conflict check"
else
    print_warning "Unexpected code $CODE for move-to-self"
fi

# Test: Move non-existent file
print_test "Move non-existent file"
RESPONSE=$(curl -s -X POST "$API_URL/note/move" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"does-not-exist.md\", \"destination\": \"anywhere.md\"}")
echo "Response: $RESPONSE"
check_code "$RESPONSE" "429" "Move non-existent file returns 429"

# ============================================================================
# 2. REPLACE OPERATIONS - Multiple Matches // 2. 替换操作 - 多次匹配

# Create note with repeated content // 创建包含重复内容的笔记
print_test "Create note with multiple identical strings"
RESPONSE=$(curl -s -X POST "$API_URL/note" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"replace-multi.md\", \"content\": \"# Replace Test\\n\\nfoo bar foo baz foo qux foo\"}")
check_code "$RESPONSE" "1" "Created note with 4x 'foo'"

# Test: Replace first only (all=false) // 测试：仅替换第一个（all=false）
print_test "Replace with all=false (should replace only first occurrence)"
RESPONSE=$(curl -s -X POST "$API_URL/note/replace" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"replace-multi.md\", \"find\": \"foo\", \"replace\": \"REPLACED\", \"all\": false}")
echo "Response: $RESPONSE"
MATCH_COUNT=$(echo "$RESPONSE" | jq -r '.data.matchCount')
CONTENT=$(echo "$RESPONSE" | jq -r '.data.note.content')
echo "Match count: $MATCH_COUNT"
echo "New content: $CONTENT"

# Count remaining 'foo' occurrences // 计算剩余的 'foo' 出现次数
REMAINING=$(echo "$CONTENT" | grep -o "foo" | wc -l)
if [[ "$REMAINING" -eq 3 ]]; then
    print_success "Only first occurrence replaced (3 'foo' remaining)"
else
    print_failure "Expected 3 'foo' remaining, got $REMAINING"
fi

# Reset and test all=true // 重置并测试 all=true
curl -s -X POST "$API_URL/note" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"replace-multi.md\", \"content\": \"foo bar foo baz foo qux foo\"}" > /dev/null

print_test "Replace with all=true (should replace all occurrences)"
RESPONSE=$(curl -s -X POST "$API_URL/note/replace" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"replace-multi.md\", \"find\": \"foo\", \"replace\": \"REPLACED\", \"all\": true}")
echo "Response: $RESPONSE"
MATCH_COUNT=$(echo "$RESPONSE" | jq -r '.data.matchCount')
CONTENT=$(echo "$RESPONSE" | jq -r '.data.note.content')
REMAINING=$(echo "$CONTENT" | grep -o "foo" | wc -l)
REPLACED=$(echo "$CONTENT" | grep -o "REPLACED" | wc -l)

if [[ "$REPLACED" -eq 4 && "$REMAINING" -eq 0 ]]; then
    print_success "All 4 occurrences replaced"
else
    print_failure "Expected 4 replacements, got $REPLACED (foo remaining: $REMAINING)"
fi

# Test: Replace with empty string (deletion) // 测试：用空字符串替换（删除）
print_test "Replace with empty string (deletion)"
curl -s -X POST "$API_URL/note" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"replace-delete.md\", \"content\": \"Keep DELETE_ME this text\"}" > /dev/null
RESPONSE=$(curl -s -X POST "$API_URL/note/replace" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"replace-delete.md\", \"find\": \"DELETE_ME \", \"replace\": \"\", \"all\": true}")
CONTENT=$(echo "$RESPONSE" | jq -r '.data.note.content')
if [[ "$CONTENT" == "Keep this text" ]]; then
    print_success "Replacement with empty string works"
else
    print_failure "Expected 'Keep this text', got '$CONTENT'"
fi

# Test: Regex with capture groups // 测试：带捕获组的正则表达式
print_test "Regex replace with capture groups"
curl -s -X POST "$API_URL/note" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"regex-capture.md\", \"content\": \"Date: 2024-01-15 and 2024-02-20\"}" > /dev/null
RESPONSE=$(curl -s -X POST "$API_URL/note/replace" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"regex-capture.md\", \"find\": \"(\\\\d{4})-(\\\\d{2})-(\\\\d{2})\", \"replace\": \"\$2/\$3/\$1\", \"regex\": true, \"all\": true}")
echo "Response: $RESPONSE"
CONTENT=$(echo "$RESPONSE" | jq -r '.data.note.content')
echo "New content: $CONTENT"
if [[ "$CONTENT" == *"01/15/2024"* && "$CONTENT" == *"02/20/2024"* ]]; then
    print_success "Regex capture groups work correctly"
else
    print_warning "Capture group replacement may not work as expected"
fi

# Test: Invalid regex // 测试：无效的正则表达式
print_test "Invalid regex pattern"
RESPONSE=$(curl -s -X POST "$API_URL/note/replace" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"regex-capture.md\", \"find\": \"[invalid(regex\", \"replace\": \"x\", \"regex\": true}")
echo "Response: $RESPONSE"
check_code "$RESPONSE" "462" "Invalid regex returns error 462"

# ============================================================================
# 3. VERSION HISTORY TRACKING // 3. 版本历史追踪

# Create a note and edit it multiple times // 创建一个笔记并多次编辑
print_test "Create note and track version increments"
RESPONSE=$(curl -s -X POST "$API_URL/note" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"version-test.md\", \"content\": \"Version 0\"}")
V0=$(echo "$RESPONSE" | jq -r '.data.version')
echo "Initial version: $V0"

RESPONSE=$(curl -s -X POST "$API_URL/note" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"version-test.md\", \"content\": \"Version 1\"}")
V1=$(echo "$RESPONSE" | jq -r '.data.version')
echo "After edit 1: $V1"

RESPONSE=$(curl -s -X POST "$API_URL/note/append" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"version-test.md\", \"content\": \"\\nAppended\"}")
V2=$(echo "$RESPONSE" | jq -r '.data.version')
echo "After append: $V2"

RESPONSE=$(curl -s -X POST "$API_URL/note/prepend" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"version-test.md\", \"content\": \"Prepended\\n\"}")
V3=$(echo "$RESPONSE" | jq -r '.data.version')
echo "After prepend: $V3"

RESPONSE=$(curl -s -X POST "$API_URL/note/replace" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"version-test.md\", \"find\": \"Version\", \"replace\": \"Ver\"}")
V4=$(echo "$RESPONSE" | jq -r '.data.note.version')
echo "After replace: $V4"

if [[ "$V0" == "0" && "$V1" == "1" && "$V2" == "2" && "$V3" == "3" && "$V4" == "4" ]]; then
    print_success "Version increments correctly: 0 -> 1 -> 2 -> 3 -> 4"
else
    print_failure "Version sequence unexpected: $V0 -> $V1 -> $V2 -> $V3 -> $V4"
fi

# ============================================================================
# 4. FRONTMATTER EDGE CASES // 4. Frontmatter 边界情况

# Test: Patch frontmatter on note WITHOUT frontmatter // 测试：在没有 Frontmatter 的笔记上修补 Frontmatter
print_test "Patch frontmatter on note without existing frontmatter"
curl -s -X POST "$API_URL/note" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"no-frontmatter.md\", \"content\": \"# Just a heading\\n\\nNo frontmatter here.\"}" > /dev/null
RESPONSE=$(curl -s -X PATCH "$API_URL/note/frontmatter" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"no-frontmatter.md\", \"updates\": {\"title\": \"Added Title\"}}")
echo "Response: $RESPONSE"
CONTENT=$(echo "$RESPONSE" | jq -r '.data.content')
if [[ "$CONTENT" == "---"* ]]; then
    print_success "Frontmatter was created on note without it"
else
    print_warning "Frontmatter may not have been added"
fi
echo "Content preview: $(echo "$CONTENT" | head -c 200)"

# Test: Remove non-existent key // 测试：删除不存在的键
print_test "Remove frontmatter key that doesn't exist"
curl -s -X POST "$API_URL/note" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"fm-remove.md\", \"content\": \"---\\ntitle: Test\\n---\\nBody\"}" > /dev/null
RESPONSE=$(curl -s -X PATCH "$API_URL/note/frontmatter" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"fm-remove.md\", \"remove\": [\"nonexistent\", \"also_missing\"]}")
echo "Response: $RESPONSE"
check_code "$RESPONSE" "1" "Removing non-existent keys doesn't error"

# Test: Add nested YAML structure // 测试：添加嵌套的 YAML 结构
print_test "Add nested YAML structure to frontmatter"
RESPONSE=$(curl -s -X PATCH "$API_URL/note/frontmatter" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"fm-remove.md\", \"updates\": {\"metadata\": {\"author\": \"test\", \"tags\": [\"a\", \"b\"]}}}")
echo "Response: $RESPONSE"
CONTENT=$(echo "$RESPONSE" | jq -r '.data.content')
if [[ "$CONTENT" == *"metadata"* ]]; then
    print_success "Nested YAML structure added"
else
    print_warning "Nested structure may not work"
fi

# ============================================================================
# 5. LINK INDEXING - Backlinks & Outlinks // 5. 链接索引 - 反向链接与正向链接

# Create notes with various link formats // 创建包含各种链接格式的笔记
print_test "Create notes with wiki links for link testing"
curl -s -X POST "$API_URL/note" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"target-note.md\", \"content\": \"# Target Note\\n\\nThis is the target.\"}" > /dev/null

curl -s -X POST "$API_URL/note" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"linker1.md\", \"content\": \"# Linker 1\\n\\nLinks to [[target-note]] here.\"}" > /dev/null

curl -s -X POST "$API_URL/note" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"linker2.md\", \"content\": \"# Linker 2\\n\\nAlso links to [[target-note|with alias]] and [[target-note#heading]].\"}" > /dev/null

curl -s -X POST "$API_URL/note" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"circular-a.md\", \"content\": \"Links to [[circular-b]]\"}" > /dev/null

curl -s -X POST "$API_URL/note" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"circular-b.md\", \"content\": \"Links to [[circular-a]]\"}" > /dev/null

print_success "Created test notes with links"

# Test: Get backlinks // 测试：获取反向链接
print_test "Get backlinks for target-note"
RESPONSE=$(curl -s "$API_URL/note/backlinks?vault=$VAULT&path=target-note.md" -H "$AUTH")
echo "Response: $RESPONSE"
BACKLINK_COUNT=$(echo "$RESPONSE" | jq -r '.data | length')
echo "Backlink count: $BACKLINK_COUNT"
if [[ "$BACKLINK_COUNT" -ge 1 ]]; then
    print_success "Found $BACKLINK_COUNT backlinks"
else
    print_warning "Expected backlinks but got $BACKLINK_COUNT - link indexing may have issues"
fi

# Test: Get outlinks with alias // 测试：获取带别名的出站链接
print_test "Get outlinks (check alias handling)"
RESPONSE=$(curl -s "$API_URL/note/outlinks?vault=$VAULT&path=linker2.md" -H "$AUTH")
echo "Response: $RESPONSE"
check_code "$RESPONSE" "1" "Get outlinks succeeded"

# Test: Circular links // 测试：循环链接
print_test "Test circular link handling"
RESPONSE=$(curl -s "$API_URL/note/backlinks?vault=$VAULT&path=circular-a.md" -H "$AUTH")
echo "Backlinks to circular-a: $RESPONSE"
RESPONSE=$(curl -s "$API_URL/note/backlinks?vault=$VAULT&path=circular-b.md" -H "$AUTH")
echo "Backlinks to circular-b: $RESPONSE"
print_info "Circular links should not cause infinite loops"

# ============================================================================
print_header "6. APPEND/PREPEND EDGE CASES"
# 6. APPEND/PREPEND EDGE CASES // 6. 追加/前置 边界情况

# Test: Append to empty note // 测试：追加到空笔记
print_test "Append to empty note"
curl -s -X POST "$API_URL/note" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"empty-note.md\", \"content\": \"\"}" > /dev/null
RESPONSE=$(curl -s -X POST "$API_URL/note/append" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"empty-note.md\", \"content\": \"First content\"}")
CONTENT=$(echo "$RESPONSE" | jq -r '.data.content')
if [[ "$CONTENT" == "First content" ]]; then
    print_success "Append to empty note works"
else
    print_failure "Append to empty note failed"
fi

# Test: Prepend with frontmatter preservation // 测试：前置内容并保留 Frontmatter
print_test "Prepend to note with frontmatter (should go after frontmatter)"
curl -s -X POST "$API_URL/note" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"fm-prepend.md\", \"content\": \"---\\ntitle: Test\\n---\\nOriginal body\"}" > /dev/null
RESPONSE=$(curl -s -X POST "$API_URL/note/prepend" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"fm-prepend.md\", \"content\": \"PREPENDED\\n\"}")
CONTENT=$(echo "$RESPONSE" | jq -r '.data.content')
echo "Content after prepend:"
echo "$CONTENT"
# Check frontmatter is still first // 检查 Frontmatter 是否仍在开头
if [[ "$CONTENT" == "---"* && "$CONTENT" == *"PREPENDED"* ]]; then
    print_success "Frontmatter preserved, content prepended after it"
else
    print_warning "Check prepend behavior with frontmatter"
fi

# Test: Append with wiki links (should they be indexed?) // 测试：追加带 Wiki 链接的内容（是否应被索引？）
print_test "Append content with wiki links"
RESPONSE=$(curl -s -X POST "$API_URL/note/append" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"fm-prepend.md\", \"content\": \"\\n\\nNew link: [[new-target]]\"}")
check_code "$RESPONSE" "1" "Appended content with link"
# Check if outlinks are updated // 检查出站链接是否已更新
RESPONSE=$(curl -s "$API_URL/note/outlinks?vault=$VAULT&path=fm-prepend.md" -H "$AUTH")
OUTLINKS=$(echo "$RESPONSE" | jq -r '.data | length')
echo "Outlinks after append: $OUTLINKS"
if [[ "$OUTLINKS" -ge 1 ]]; then
    print_success "Links in appended content are indexed"
else
    print_warning "Links from append may not be indexed immediately"
fi

# ============================================================================
print_header "7. createOnly EDGE CASES"
# 7. createOnly EDGE CASES // 7. createOnly 边界情况

# Test: createOnly on soft-deleted note // 测试：在软删除的笔记上使用 createOnly
print_test "createOnly on soft-deleted note"
# Create and delete a note // 创建并删除一个笔记
curl -s -X POST "$API_URL/note" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"deleted-note.md\", \"content\": \"Will be deleted\"}" > /dev/null
curl -s -X DELETE "$API_URL/note?vault=$VAULT&path=deleted-note.md" -H "$AUTH" > /dev/null

# Try to create with createOnly // 尝试使用 createOnly 创建
RESPONSE=$(curl -s -X POST "$API_URL/note" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"deleted-note.md\", \"content\": \"New content\", \"createOnly\": true}")
echo "Response: $RESPONSE"
CODE=$(echo "$RESPONSE" | jq -r '.code')
if [[ "$CODE" == "1" ]]; then
    print_success "createOnly succeeds on soft-deleted note (recreates it)"
elif [[ "$CODE" == "430" ]]; then
    print_info "createOnly blocks even for soft-deleted notes"
else
    print_warning "Unexpected behavior: code $CODE"
fi

# ============================================================================
print_header "8. PATH EDGE CASES"
# 8. PATH EDGE CASES // 8. 路径边界情况

# Test: Unicode in path // 测试：路径中的 Unicode 字符
print_test "Unicode characters in path"
RESPONSE=$(curl -s -X POST "$API_URL/note" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"日本語/笔记.md\", \"content\": \"Unicode path test\"}")
echo "Response: $RESPONSE"
check_code "$RESPONSE" "1" "Unicode path accepted"

# Retrieve it back // 检索回来
RESPONSE=$(curl -s "$API_URL/note?vault=$VAULT&path=%E6%97%A5%E6%9C%AC%E8%AA%9E/%E7%AC%94%E8%AE%B0.md" -H "$AUTH")
check_code "$RESPONSE" "1" "Unicode path retrieved"

# Test: Path with spaces // 测试：带空格的路径
print_test "Path with spaces"
RESPONSE=$(curl -s -X POST "$API_URL/note" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"folder with spaces/note with spaces.md\", \"content\": \"Spaces test\"}")
check_code "$RESPONSE" "1" "Path with spaces accepted"

# Test: Very long path // 测试：非常长的路径
print_test "Very long path (200+ characters)"
LONG_PATH="very/long/path/$(printf 'x%.0s' {1..150}).md"
RESPONSE=$(curl -s -X POST "$API_URL/note" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"$LONG_PATH\", \"content\": \"Long path test\"}")
echo "Response: $RESPONSE"
CODE=$(echo "$RESPONSE" | jq -r '.code')
if [[ "$CODE" == "1" ]]; then
    print_success "Long path accepted"
else
    print_info "Long path rejected (may be intended): code $CODE"
fi

# Test: Path traversal attempt // 测试：路径遍历尝试
print_test "Path traversal attempt (../)"
RESPONSE=$(curl -s -X POST "$API_URL/note" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"../../../etc/passwd\", \"content\": \"Should fail\"}")
echo "Response: $RESPONSE"
# This should either fail or sanitize the path // 这应该要么失败要么净化路径
print_info "Path traversal handling: $(echo "$RESPONSE" | jq -r '.code')"

# ============================================================================
# 9. CONCURRENT EDIT SIMULATION // 9. 并发编辑模拟

print_test "Rapid sequential edits (version conflict potential)"
# Create note // 创建笔记
curl -s -X POST "$API_URL/note" -H "$AUTH" -H "Content-Type: application/json" \
    -d "{\"vault\": \"$VAULT\", \"path\": \"rapid-edit.md\", \"content\": \"Initial\"}" > /dev/null

# Rapid edits
for i in {1..5}; do
    curl -s -X POST "$API_URL/note" -H "$AUTH" -H "Content-Type: application/json" \
        -d "{\"vault\": \"$VAULT\", \"path\": \"rapid-edit.md\", \"content\": \"Edit $i at $(date +%s%N)\"}" &
done
wait

RESPONSE=$(curl -s "$API_URL/note?vault=$VAULT&path=rapid-edit.md" -H "$AUTH")
FINAL_VERSION=$(echo "$RESPONSE" | jq -r '.data.version')
echo "Final version after 5 rapid edits: $FINAL_VERSION"
if [[ "$FINAL_VERSION" -ge 4 ]]; then
    print_success "Handled rapid edits (version: $FINAL_VERSION)"
else
    print_warning "Some edits may have been lost (version: $FINAL_VERSION)"
fi

# ============================================================================
print_header "10. FILE/ATTACHMENT API CHECK"
# 10. FILE/ATTACHMENT API CHECK // 10. 文件/附件 API 检查

print_test "Check for REST file upload endpoint"
# Try common file upload endpoints
RESPONSE=$(curl -s -X POST "$API_URL/file" -H "$AUTH" -F "file=@/dev/null" 2>/dev/null)
echo "POST /file response: $RESPONSE"
RESPONSE=$(curl -s -X POST "$API_URL/upload" -H "$AUTH" -F "file=@/dev/null" 2>/dev/null)
echo "POST /upload response: $RESPONSE"
RESPONSE=$(curl -s -X POST "$API_URL/attachment" -H "$AUTH" -F "file=@/dev/null" 2>/dev/null)
echo "POST /attachment response: $RESPONSE"
print_info "File uploads appear to be WebSocket-only (not REST API)"

# ============================================================================
print_header "11. CLEANUP"
# 11. CLEANUP // 11. 清理

# Get vault ID and delete
VAULT_RESPONSE=$(curl -s "$API_URL/vault" -H "$AUTH")
VAULT_ID=$(echo "$VAULT_RESPONSE" | jq -r ".data[] | select(.vault==\"$VAULT\") | .id")
if [[ -n "$VAULT_ID" && "$VAULT_ID" != "null" ]]; then
    curl -s -X DELETE "$API_URL/vault?id=$VAULT_ID" -H "$AUTH" > /dev/null
    print_success "Cleaned up test vault"
fi

# ============================================================================
print_header "TEST SUMMARY"

echo -e "\n${BLUE}Results:${NC}"
echo -e "  ${GREEN}Passed:   $PASSED${NC}"
echo -e "  ${RED}Failed:   $FAILED${NC}"
echo -e "  ${YELLOW}Warnings: $WARNINGS${NC}"
TOTAL=$((PASSED + FAILED))
echo -e "  Total:    $TOTAL"
echo ""

if [[ $FAILED -eq 0 ]]; then
    echo -e "${GREEN}════════════════════════════════════════${NC}"
    echo -e "${GREEN}  ALL CRITICAL TESTS PASSED!${NC}"
    if [[ $WARNINGS -gt 0 ]]; then
        echo -e "${YELLOW}  ($WARNINGS warnings to review)${NC}"
    fi
    echo -e "${GREEN}════════════════════════════════════════${NC}"
    exit 0
else
    echo -e "${RED}════════════════════════════════════════${NC}"
    echo -e "${RED}  $FAILED TESTS FAILED${NC}"
    echo -e "${RED}════════════════════════════════════════${NC}"
    exit 1
fi
