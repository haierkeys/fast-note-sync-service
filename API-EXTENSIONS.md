# API Extensions for Fast Note Sync Service

This document describes the new API endpoints added in the `feature/api-extensions` branch.

## New Endpoints

### Health Check
```
GET /api/health
```
Returns server health status including database connectivity and uptime.

**Response:**
```json
{
  "status": "healthy",
  "version": "1.11.1",
  "uptime": 123.45,
  "database": "connected"
}
```

### Note Operations

#### Patch Frontmatter
```
PATCH /api/note/frontmatter?vault=<vault>&path=<path>
```
Update or remove YAML frontmatter fields without modifying note body.

**Body:**
```json
{
  "updates": {"title": "New Title", "tags": ["a", "b"]},
  "remove": ["oldField"]
}
```

#### Append Content
```
POST /api/note/append?vault=<vault>&path=<path>
```
Append content to the end of a note.

**Body:**
```json
{
  "content": "\n\n## New Section\nAppended content"
}
```

#### Prepend Content
```
POST /api/note/prepend?vault=<vault>&path=<path>
```
Prepend content after frontmatter (if present) or at the beginning.

**Body:**
```json
{
  "content": "Prepended content\n\n"
}
```

#### Find and Replace
```
POST /api/note/replace?vault=<vault>&path=<path>
```
Find and replace text in a note. Supports regex.

**Body:**
```json
{
  "find": "old text",
  "replace": "new text",
  "regex": false,
  "all": true,
  "failIfNoMatch": false
}
```

**Response includes `matchCount`:**
```json
{
  "matchCount": 3,
  "note": { ... }
}
```

#### Move Note
```
POST /api/note/move?vault=<vault>&path=<path>
```
Move/rename a note to a new path.

**Body:**
```json
{
  "destination": "new/path/note.md",
  "overwrite": false
}
```

### Link Operations

#### Get Backlinks
```
GET /api/note/backlinks?vault=<vault>&path=<path>
```
Get all notes that link TO this note.

**Response:**
```json
{
  "data": [
    {
      "path": "other-note.md",
      "linkText": "alias",
      "context": "...surrounding text [[link]]..."
    }
  ]
}
```

#### Get Outlinks
```
GET /api/note/outlinks?vault=<vault>&path=<path>
```
Get all links FROM this note.

**Response:**
```json
{
  "data": [
    {
      "path": "target-note",
      "linkText": "display text",
      "context": "...[[target-note|display text]]..."
    }
  ]
}
```

### Note Creation

#### Create Only (Don't Update)
```
POST /api/note?vault=<vault>&path=<path>
```
With `createOnly: true`, returns error 430 if note already exists.

**Body:**
```json
{
  "content": "...",
  "createOnly": true
}
```

## New Error Codes

| Code | Message |
|------|---------|
| 460 | Destination note already exists |
| 461 | No match found |
| 462 | Invalid regex pattern |

## Configuration

New config option in `config.yaml` (also editable via Admin Settings API):
```yaml
app:
  default-api-folder: ""  # Optional: prepend this folder to note paths without /
```

## Known Limitations

### Backlinks

1. **Exact path matching only**: Backlinks match the stored link path exactly. Obsidian's "shortest path when possible" resolution is not fully replicated server-side.

2. **Extension handling**: Links are stored without `.md` extension (e.g., `[[Note1]]` stores "Note1"). Queries with full paths (e.g., "Note1.md") are normalized by stripping `.md`.

3. **Heading anchors**: Links with `#heading` (e.g., `[[note#section]]`) are stored as-is. Querying backlinks for "note.md" won't find links to "note#section".

4. **Relative paths**: Links using relative paths (e.g., `[[../folder/note]]`) are stored as-is. Cross-folder resolution is not performed.

### Link Indexing

- Links are indexed when notes are saved via the API
- Existing notes need to be re-saved to populate the link index
- Link parsing uses regex: `\[\[([^\]|]+)(?:\|([^\]]+))?\]\]`

### Version History

- Version numbers increment on each content change
- History entries are created asynchronously (with delay)
- Move operations migrate history from source to destination note

## Testing

Run the test scripts:
```bash
# Basic API tests (41 tests)
./test-api.sh

# Edge case tests
./test-edge-cases.sh
```

## Files Changed

### New Files
- `internal/routers/api_router/handler_health.go`
- `internal/domain/domain_note_link.go`
- `internal/model/note_link.gen.go`
- `internal/dao/note_link_repository.go`
- `internal/service/note_link_service.go`
- `pkg/util/frontmatter.go`
- `pkg/util/link_parser.go`
- `pkg/util/path.go`
- `test-api.sh`
- `test-edge-cases.sh`

### Modified Files
- `config/config.yaml`
- `internal/app/app.go`
- `internal/app/config.go`
- `internal/dao/note_repository.go`
- `internal/domain/repository.go`
- `internal/dto/note_dto.go`
- `internal/model/model.go`
- `internal/routers/api_router/handler_note.go`
- `internal/routers/router.go`
- `internal/service/note_service.go`
- `pkg/code/common.go`
