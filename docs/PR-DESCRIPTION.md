# PR: Folder Tree Endpoint + Folder API Bug Fixes

## Summary

Adds a new `GET /api/folder/tree` endpoint that returns the complete folder hierarchy with note/file counts, and fixes duplicate folder bugs in the existing folder API endpoints.

## New Feature: Folder Tree Endpoint

```
GET /api/folder/tree?vault=<name>&depth=<optional>
```

Returns the full folder tree structure with note and file counts per folder. Supports optional `depth` parameter to limit tree depth.

**Response example:**
```json
{
  "folders": [
    {
      "path": "projects",
      "name": "projects",
      "noteCount": 3,
      "fileCount": 0,
      "children": [
        {
          "path": "projects/golf-email-series",
          "name": "golf-email-series",
          "noteCount": 6,
          "fileCount": 0,
          "children": [...]
        }
      ]
    }
  ],
  "rootNoteCount": 26,
  "rootFileCount": 2
}
```

## Bug Fixes

### 1. Folder AutoMigrate missing (upstream bug)

`folder_repository.go` was using `UseQuery()` instead of `UseQueryWithOnceFunc()`, so the folder table was never auto-created on fresh databases. Added `folder()` helper method with `model.AutoMigrate(g, "Folder")`, matching the pattern used in `note_repository.go`.

### 2. Duplicate folders in API responses

**Root cause:** `EnsurePathFID` has a check-then-create race condition. When multiple notes sync concurrently (even from a single device), each goroutine independently checks if a folder exists and creates it if not. Without atomicity, multiple goroutines can all see "not found" and all insert a record for the same path. Confirmed via direct DB inspection — e.g. 3 rows for "projects" path in a single-device setup.

**Query-side fix applied to all affected endpoints:**

| Endpoint | Fix |
|----------|-----|
| `GET /api/folders` (List) | Resolves all folder IDs per path via `GetAllByPathHash`, queries children across all matching parent FIDs, deduplicates results by PathHash |
| `GET /api/folder/notes` (ListNotes) | Resolves all folder IDs per path, uses `FID IN (...)` query to find notes across all duplicate folder records |
| `GET /api/folder/files` (ListFiles) | Same approach as ListNotes for files |
| `GET /api/folder/tree` (GetTree) | Path-based deduplication, merges note/file counts across all duplicate folder records |

**Note:** The `ListByUpdatedTimestamp` method (used by the sync endpoint) already had PathHash deduplication — the same pattern was missing from List, ListNotes, and ListFiles.

**Root cause not fixed in this PR.** A proper fix would require either a `singleflight` keyed by `(vaultID, path)` in `EnsurePathFID`, or a `UNIQUE` constraint on `(vault_id, path_hash)` in the folder table. See comment on `EnsurePathFID` in `folder_service.go` for details. The query-side fixes make the API correct regardless of duplicate rows.

### 3. Swagger docs regenerated

Updated generated Swagger files to include the new `/api/folder/tree` endpoint.

## Files Changed

| File | Change |
|------|--------|
| `internal/dto/folder_dto.go` | Added FolderTreeRequest, FolderTreeNode, FolderTreeResponse DTOs |
| `internal/domain/repository.go` | Added `GetAllByPathHash` (FolderRepo), `ListByFIDs`/`ListByFIDsCount` (NoteRepo, FileRepo) |
| `internal/dao/folder_repository.go` | Fixed AutoMigrate, implemented `GetAllByPathHash` |
| `internal/dao/note_repository.go` | Implemented `ListByFIDs`, `ListByFIDsCount` |
| `internal/dao/file_repository.go` | Implemented `ListByFIDs`, `ListByFIDsCount` |
| `internal/service/folder_service.go` | Added `GetTree`, fixed `List`/`ListNotes`/`ListFiles` dedup, documented race condition |
| `internal/routers/api_router/handler_folder.go` | Added `Tree` handler with Swagger annotations |
| `internal/routers/router.go` | Registered `/folder/tree` route |
| `docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml` | Regenerated Swagger docs |

## Testing

23/23 API tests pass (test-folder-api.sh), including:
- All existing folder CRUD endpoints
- New folder tree endpoint (full tree + depth-limited)
- Note/file listing within folders
- Backlinks, outlinks, note edit operations

## Breaking Changes

None. All changes are additive. Existing API behavior is preserved (with duplicates now correctly deduplicated).
