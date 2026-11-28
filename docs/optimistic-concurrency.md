# Optimistic concurrency for note writes

REST and MCP note mutations support **git-style non-fast-forward protection**: a client sends the
revision it read, and the server applies the write only if that revision is still current. A stale
write is rejected with `HTTP 409` and the note is left untouched, so no client's edit is ever
silently discarded.

This is opt-in per request. Requests that do not send `revision` keep the previous
(unconditional, last-write-wins) behaviour — see [Compatibility](#compatibility).

## Revision

Every note carries a `revision`: an integer that starts at `1` on creation and increases by exactly
one on every mutation of that note:

| Mutation | Advances `revision` |
|---|---|
| create | yes (`1`) |
| content write (`POST /api/note`) | yes |
| append / prepend / find-replace / frontmatter patch | yes |
| rename / move (the retired **and** the new record; reusing a tombstone target also rewrites it) | yes |
| delete (soft) | yes |
| restore from the recycle bin | yes |
| restore a historical version (`PUT /api/note/history/restore`) | yes |
| recursive folder delete (MCP `folder_delete`, which soft-deletes every contained note) | yes, per contained note |
| a legacy write without `revision` | yes |
| mtime-only sync bookkeeping (`UpdateMtime`, `UpdateActionMtime`) | no |
| delayed history snapshot (`UpdateSnapshot`) | no |

Any of the above therefore invalidates a revision read before it — including the indirect ones
(history restore, folder cascade), which are easy to miss when a client caches revisions.

`revision` is not the same value as `version` (the latter is maintained by the note-history
subsystem and is not monotonic with respect to note mutations), and it is not the content hash.
A content hash cannot express rename/move, delete/restore or ABA (content returning to an earlier
byte-identical state); a revision can.

Read `revision` from any of:

* `GET /api/note` → `data.revision`
* `GET /api/notes` → `data.list[].revision`
* `GET /api/notes?isRecycle=true` → tombstone revision (needed to restore conditionally)
* MCP `note_get` / `note_list` → `structuredContent.note.revision`
* MCP mutation tool results → `structuredContent.note.revision`

## Conditional write

Send `revision` alongside the mutation:

```http
POST /api/note
token: <token>
Content-Type: application/json

{"vault":"Projects","path":"Projects/Plan.md","content":"# Plan\n","revision":7}
```

Supported on:

| Endpoint | Where the revision goes |
|---|---|
| `POST /api/note` | body `revision` |
| `POST /api/note/append` | body `revision` |
| `POST /api/note/prepend` | body `revision` |
| `POST /api/note/replace` | body `revision` |
| `PATCH /api/note/frontmatter` | body `revision` |
| `POST /api/note/rename` | body `revision` (the **source** note's revision) |
| `DELETE /api/note` | query `revision` |
| `PUT /api/note/restore` | body `revision` (the tombstone's revision) |

The same optional argument exists on the MCP tools `note_create_or_update`, `note_append`,
`note_prepend`, `note_replace`, `note_patch_frontmatter`, `note_delete`, `note_restore`,
`note_rename`.

On success the response body contains the new `revision` (previous + 1).

### Rejected write

```
HTTP/1.1 409 Conflict

{
  "code": 452,
  "status": false,
  "message": "Note revision conflict, the note was modified by another client",
  "data": {
    "path": "Projects/Plan.md",
    "revision": 9,
    "contentHash": "-1544506754",
    "mtime": 1795820000000,
    "deleted": false,
    "content": "<current stored content>"
  }
}
```

`data` describes the note exactly as it is stored now, so the client can reconcile and retry in a
single round trip without a second read. `data.content` is **empty when `data.deleted` is true** —
a tombstone's body is never echoed; restore it first if you need it. Nothing from the rejected
request was applied: the content file, the full-text index and the history snapshots are unchanged.

Two application codes exist:

| Code | HTTP | Meaning |
|---|---|---|
| `452` | `409` | the supplied revision is no longer current (or the note no longer exists) |
| `453` | `409` | the note is in the recycle bin; content writes cannot revive it, restore it first |

> Note: the service otherwise answers `HTTP 200` for application-level errors and clients read the
> `code` field. Conditional-write rejections are the deliberate exception so that any HTTP client,
> proxy or agent sees the failure in the status line. `453` can also reach a client that did *not*
> send a revision (see below).

MCP renders the same payload as a tool error (`isError: true`) whose `content[0].text` is the JSON
body above.

## Deleted notes

A soft-deleted note keeps its row and its revision. Ordinary content writes never revive it:

* conditional write → `453` (stale revision on a tombstone reports `452` first if the revision is
  wrong)
* legacy write without `revision` → `453`

Revival is an explicit, separately authorised operation:

```
PUT /api/note/restore   {"vault":"Projects","path":"Projects/Plan.md","revision":11}
```

where `11` is the revision read from `GET /api/notes?isRecycle=true`. The restore advances the
revision again, so a client that read the note before deletion cannot restore an older state.

The WebSocket transport is exempt from this rule: the Obsidian plugin protocol has always allowed
recreating a path whose server record is a tombstone, and it keeps that behaviour.

## Client algorithm

```
loop:
  note = GET /api/note(...)            # note.revision, note.content
  save(note.content with my changes)
  res  = POST /api/note(..., revision = note.revision)
  if res.code == 1:                    # accepted
      note.revision = res.data.revision
      break
  if res.code in (452, 453):           # someone else won, or the note was deleted
      theirs = res.data                # current content + revision, no extra read needed
      content = merge(theirs.content, my_changes)   # or abort and tell the user
      note.revision = theirs.revision
      continue                         # retry with the new revision
  raise
```

Never retry with the old revision, and never re-send a rejected payload blindly — that is exactly
the silent-overwrite behaviour this contract exists to prevent.

## Atomicity

The comparison and the mutation are one SQL statement:

```sql
UPDATE note
   SET content_hash = ?, path = ?, ..., rev = rev + 1, updated_timestamp = ?
 WHERE id = ? AND rev = ?
```

`rev = rev + 1` is always evaluated by the database (never as a client-computed literal), and a
zero row count is the conflict. The check therefore cannot be interleaved: it holds with the write
queue enabled or disabled, with several service instances sharing one database, and on SQLite,
MySQL and PostgreSQL. Read-modify-write helpers (append/prepend/replace/frontmatter) additionally
hold the per-path in-process mutex across their read and their write, so the content they compose is
always derived from the revision they then verify.

A rejected write performs no filesystem write, no full-text upsert and no history entry.

### Known limitation: a storage failure after a successful compare-and-swap

The guarantee above covers *rejections*. It does not cover an I/O failure that happens **after** the
CAS statement has committed: `ExecuteWrite` serialises writes but does not open a database
transaction, so the order is `UPDATE note … → SaveContentToFile(content.txt) → upsertFTS`. If the
file write fails (read-only directory, full disk, crash), the row keeps the new `revision` and
`content_hash` while `content.txt` still holds the previous bytes, the API reports that mismatched
pair, and a client whose retry is refused sees a conflict payload whose `contentHash` does not match
its `content`. This is a pre-existing storage-ordering defect of the service (it predates this
contract and affects legacy writes identically); reversing the order would only move the window.
It is reported separately from this contract (see the accompanying upstream bug report) and should
not be worked around by clients.

## Compatibility

* **Reads** gained one field (`revision`). Additive.
* **Writes without `revision`** behave exactly as before: last writer wins. This is *not*
  concurrency-safe and is documented as such; it exists so plugin/WebSocket traffic and older
  scripts keep working. Because every mutation (including a legacy one) advances the revision,
  a conditional client is still protected against an interleaved legacy write.
* **Writes without `revision` to a deleted note** are now refused with `453` instead of silently
  resurrecting the note. Use `PUT /api/note/restore`.
* **WebSocket / Obsidian plugin**: untouched. The protocol has no new fields; `baseHash` +
  `offlineSyncStrategy` conflict handling works as before, and `rev` simply advances. The WS
  transport ignores the `revision` field entirely — for `NoteModify`, `NoteDelete` and `NoteRename`
  alike — so a frame that carries one cannot accidentally acquire REST/MCP semantics.
* **Attachments** (`/api/file*`): not covered by this contract; file uploads remain
  last-write-wins. Tracked as a follow-up.

## Migration

`note.rev` is added by the existing per-database auto-migration
(`INTEGER NOT NULL DEFAULT 0`, populated as `1` for new notes). Pre-existing rows start at `0`,
which is a valid baseline: read the note, send `revision: 0`, and the first write is accepted and
moves it to `1`. No backfill, no downtime, no data rewrite. Rolling back the service binary leaves
an unused column that older builds ignore.
