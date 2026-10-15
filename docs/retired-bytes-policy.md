# Optional retired-bytes protection

Set `FNS_RETIRED_BYTES_POLICY` to a persisted JSON policy file to reject known retired content. Leave it unset to disable this optional policy. A configured missing or invalid file prevents initialization; changes require restart.

```json
{
  "version": 1,
  "entries": [
    {
      "uid": 1,
      "vault_id": 1,
      "path": "Archive/example.md",
      "sha256": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
    }
  ]
}
```

Each entry matches one exact user, vault ID, path, and SHA-256 of actual bytes. List multiple digests when necessary. Changed bytes remain eligible for normal synchronization. The policy applies to files and notes, including rename destinations and conflict handling.

Explicit Restore remains available. Its authorization is appended and synced to the adjacent `.audit.jsonl` file before the record changes. The policy directory must therefore be writable. A successful Restore does not exempt later stale uploads.

File and note Delete requests accept optional `expectedSHA256` and `expectedUpdatedTimestamp` conditions. Mismatches return error code 530 without deleting the current record. Callers that omit conditions retain the existing behavior.

HTTP note creation or update must include a string `content` field. An explicit empty string is valid; omitted or null content is rejected before any service mutation.

Mutation locks operate within one service process. Multiple replicas and external database or blob writers require coordination. This protection does not automatically clean up previously revived records.
