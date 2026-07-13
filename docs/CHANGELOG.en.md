# CHANGELOG

All notable changes to this project will be documented in this file.

The project adheres to [Keep a Changelog](https://keepachangelog.com/en/0.3.0/) guidelines.

---

## v3.6.0
> *2026/07/09*

### ✨ Added

- **Sync**: Added a sliding-window pipeline mode for the sync protocol. Upload batches and download pages, previously confirmed one at a time (stop-and-wait), can now have multiple batches/pages in flight concurrently, greatly cutting full-sync time for large vaults (measured 6.64x faster upload, 3.19x faster download). The whole mechanism is capability-negotiated and bidirectionally backward-compatible: any combination of old/new client and old/new server automatically falls back to the previous one-at-a-time mode — no forced upgrade required.
- **Config**: Added `pipeline-window-up` (default `8`) and `pipeline-window-down` (default `4`), controlling how many upload batches / download pages may be in flight at once. Setting either explicitly to `0` instantly rolls the connection back to stop-and-wait mode, serving as a runtime safety switch for the pipeline feature (also readable/writable from the admin panel).

### 🚀 Optimized

- **Sync**: Raised the default download sync chunk size `sync-down-chunk-num` from `50` to `200`, reducing the number of page round-trips in stop-and-wait scenarios.
- **Sync**: Full/incremental sync now reads note content on demand and backfills it only when a page is actually sent, avoiding serial blind reads and full in-memory materialization for large vaults.
- **Database**: Added a `(vault_id, path_hash)` composite index to the `note` table, eliminating point-query full table scans during full uploads.
- **Database**: Removed several redundant duplicate queries across the FTS-write, update-check, and batch-delete paths (`upsertFTS` reuses the record the caller just wrote; `NoteModify` merges its duplicate queries; batch note deletion drops the per-item existence check).
- **Sync**: The write-queue worker now drains already-queued operations in non-blocking batches, reducing per-operation scheduling overhead.
- **Sync**: `sync_log` writes moved from one bare goroutine per record to a bounded channel with a single batching worker, avoiding a burst of goroutines during large uploads.
- **Sync**: Broadcast fan-out now sends concurrently outside the lock, so a slow device no longer delays message delivery to the same user's other online devices.
- **Performance**: Note diffing now enables line-level preprocessing dynamically based on base text length, speeding up merges of large notes.
- **Performance**: Folder counts are now computed with a single `GROUP BY` aggregate query instead of 2 `COUNT` queries per folder (N+1); file path migration checks and the `git-sync` enablement lookup now use in-process caches on these hot paths, cutting repeated I/O and DB queries.

### 🛠️ Fixed

- **Sync**: Fixed a data-loss bug where concurrent writes to the same note/file could misuse `singleflight`, silently dropping the later write while still returning a success acknowledgment.
- **Sync**: Fixed `os.Rename` failures during note/attachment rename being silently swallowed, which left the database and disk state inconsistent.
- **Stability**: Added panic recovery to background bare goroutines (backup, git sync, etc.) so a single task panic no longer crashes the whole server process.
- **Sync**: The upload batch protocol now deduplicates by batch index, preventing client retransmissions from causing duplicate writes and miscounts on the server.
- **Sync**: When a download-page `PageAck` doesn't match the expected page, the server now resends the current page as a fallback, preventing the client from stalling until the 10-minute session TTL.
- **Concurrency**: Fixed unlocked reads/writes of WebSocket client info fields, which could race under parallel processing.
- **Stability**: Added a minimum-idle-time guard to the DB connection pool's LRU eviction, preventing still-in-use connections from being closed by mistake.
- **Config**: Fixed a half-broken rollback safety switch — explicitly setting `pipeline-window-up`/`pipeline-window-down` to `0` used to be silently overwritten back to its default by the config's second defaults-fill pass; an explicit `0` is now correctly recognized and preserved.
- **Sync**: Fixed a zero-value ambiguity where page 0 of the paginated envelope field `pageIndex` was indistinguishable from a non-paginated broadcast message in Protobuf mode.

---

## v3.1.1
> *2026/05/25*

### 🚀 Optimized

- **WebGUI**: Optimized theme menu interaction. Clicking the theme button now triggers a clean popup menu, supporting "Follow System", "Auto", "Light", and "Dark" selections. The button icon dynamically changes according to the currently active theme.
- **WebGUI**: Adjusted default theme settings. For new users who haven't manually configured a theme, the default is changed from "Auto" to "Follow System" for a more natural user experience.
- **WebGUI**: Optimized hover and click effects for the "Theme", "Color", and "Language" buttons in the top navigation bar, making operations more responsive and natural.
- **WebGUI**: Optimized button tooltips. Hovering over top navigation buttons now displays brief functional explanations. These tooltips are automatically disabled on mobile devices to prevent accidental triggers by touch gestures.

---

## v3.1.0
> *2026/05/24*

### ✨ Added

- **Sync**: Added static mock access endpoints for attachments, allowing Nginx proxies to access static resources of specific vaults and whitelisted file types without service tokens.
- **Update**: Upgraded version detection mechanism to show the complete changelog of all historical versions between the current version and the latest version upon detection.

---

## v3.0.6
> *2026/05/23*

### ✨ Added

- **MCP**: Added `file_write` tool, allowing third-party AI clients to write attachments or files directly using Base64 encoding and sync real-time to all terminals (please avoid uploading extremely large files to prevent OOM).
- **Network**: Added custom response headers configuration, supporting unified customized response headers for all HTTP services and WebSocket handshake stages (refer to the `server.custom-response-headers` option in the latest `config.yaml`).

### 🚀 Optimized

- **WebGUI**: Optimized note list drag-and-drop navigation. Users can now move notes or attachments to the root directory or any subdirectories by dragging.
- **WebGUI**: Optimized Obsidian properties rendering. The note editor and shared note pages now properly display Obsidian note properties.
- **WebGUI**: Optimized sharing entry. Added a quick "Share Note" button entry directly in the note editing page.

### 🛠️ Fixed

- **Autostart**: Fixed permission issues preventing macOS autostart services from running normally and resolved execution errors in the quick installation script.
- **Config**: Fixed configuration inheritance issue where explicitly setting default-enabled boolean configurations (such as `webgui-login-token-bind-ip`, cloud storage enablers, and WebSocket parallelization/compression) to `false` in the configuration file was still overwritten by default values.
- **WebGUI**: Fixed file access errors in WebGUI where loading images or attachments failed because direct browser requests lacked specific client headers and got blocked.
- **WebGUI**: Fixed a bug where WebGUI showed redundant paths in the note title edit bar.

---

## v3.0.5
> *2026/05/23*

### ✨ Added

- **MCP**: Added `file_write` tool, allowing third-party AI clients to write attachments or files directly using Base64 encoding and sync real-time to all terminals.
- **Network**: Added custom response headers configuration, supporting unified customized response headers for all HTTP services and WebSocket handshake stages.

### 🚀 Optimized

- **WebGUI**: Optimized note list drag-and-drop navigation. Users can now move notes or attachments to the root directory or any subdirectories by dragging.
- **WebGUI**: Optimized Obsidian properties rendering. The note editor and shared note pages now properly display Obsidian note properties.
- **WebGUIGUI**: Optimized sharing entry. Added a quick "Share Note" button entry directly in the note editing page.

### 🛠️ Fixed

- **Autostart**: Fixed permission issues preventing macOS autostart services from running normally and resolved execution errors in the quick installation script.
- **Config**: Fixed configuration inheritance issue where explicitly setting default-enabled boolean configurations to `false` was still overwritten by default values.
- **WebGUI**: Fixed file access errors in WebGUI where loading images or attachments failed due to missing client headers.
- **WebGUI**: Fixed a bug where WebGUI showed redundant paths in the note title edit bar.

---

## v3.0.4
> *2026/05/21*

### ✨ Added

- **Sync**: Added an active push notification mechanism after file uploads. Active pushes will notify all online clients after uploads finish, regardless of REST API or WebSocket upload channels.

### 🚀 Optimized

- **Auth**: Optimized WebGUI login token configuration. Added separate configuration support for token validity duration and client IP binding, allowing IP binding to be disabled for Cloudflare Tunnel scenarios.
- **Experience**: Optimized feedback channels. Added GitHub `bug report` and `feature request` templates for easier issue reporting.

### 🛠️ Fixed

- **Logging**: Fixed log redundancy. Normal WebSocket disconnections, timeouts, and network drops will no longer be logged as error logs.
- **Sync**: Fixed file upload completion criteria. Added comprehensive checks using both byte count and chunk count to avoid completion flows getting stuck.
- **Sync**: Fixed sync conflicts. Old duplicate upload sessions on the same path are now terminated before starting new sessions, reducing sync loops and system resources.

---

## v3.0.3
> *2026/05/18*

### ✨ Added

- **Auth**: Added support for token-specific note vault restrictions. You can now define allowed note vaults for each token to automatically reject unauthorized and cross-vault data requests.

### 🛠️ Fixed

- **Stability**: Fixed a service crash issue triggered by retrieving system metrics on specific platforms like Windows.

---

## v3.0.2
> *2026/05/16*

### ✨ Added

- **Update**: Added update release channel selection, allowing users to switch update sources between "Stable" and "Beta" releases.
- **Auth**: Added display of connected clients' recent activity in the authorization token manager, showing the last active timestamp of each connected client.
- **Admin**: Added client eviction functionality, allowing administrators to forcibly terminate a specific WebSocket client connection in the client manager.

### 🚀 Optimized

- **WebGUI**: Optimized sponsor list UI display.

---

## v3.0.0
> *2026/05/14*

> [!WARNING]
> **Important Notes**: This update introduces a brand new token architecture, which will invalidate all existing authorizations in older versions. After upgrading, you will need to re-authorize all client devices.

### 🔐 Core Update: Brand New Stateful Token Permission System

This release completely overhauls the underlying security architecture, transitioning from a single stateless token to a **stateful multi-dimensional (protocol, client, capability)** validation token:
- **Multi-Dimensional Validation**:
  - **Protocol Control**: Supports independent authorization and restriction for `REST API`, `WebSocket`, `MCP`, and other access protocols and contexts.
  - **Client Restriction**: Supports binding specific devices using wildcards (e.g., `obsidian*`) to achieve "one device, one token".
  - **Content Control**: Granular atomic permissions down to note read/write (`note_r/w`), attachment management (`file_r/w`), and system config (`config_r/w`).
- **Generation Check & Token Rotation**:
  - Introduced `Nonce`-based generation mechanism. Tokens can be rotated when leakage risks arise, instantly invalidating previous generation tokens.
- **Dynamic Environment Binding**:
  - Supports binding static or wildcard **IP addresses** and **User-Agent** to limit token access under insecure network environments.
- **Token Access Metrics**:
  - The server records access behavior for all tokens, supporting real-time monitoring of active protocols, source IPs, and client versions.

### ✨ Added

- **Backup**: Added encryption password configuration for backup tasks. Both full and incremental backups now support flexible encryption strategies ("no password", "fixed manual password", or "system-generated random password") to enhance backup data storage security.

### 🛠️ Fixed

- **Cloud Storage**: Upgraded Alibaba Cloud OSS v2 SDK driver, fixing storage configuration connection test failures and region identification errors.

### 🚀 Optimized

- **Performance**: Optimized file handling performance. Adjusted hash calculation file size threshold to 10MB, and significantly reduced file system I/O overhead by deferring physical directory creations.
- **Protocol**: Standardized synchronization protocol. Forced the inclusion of the `vault` field in all WebSocket return messages to ensure multi-vault sync state consistency.
- **Security**: Hardened backup security. Masked sensitive password fields in backup logs by default, supporting manual decryption to view.

---

## v2.14.4
> *2026/05/10*

### 🚀 Optimized

- **Performance**: Adjusted hash calculation file size threshold to 10MB.
- **Sync**: Optimized attachment downloads. The server now actively transmits file hashes during downloads to prevent clients from repeating hash calculations.

---

## v2.14.3
> *2026/05/09*

### ✨ Added

- **MCP**: Added `StreamableHTTP` transmission protocol support for MCP, resolving the `Streaming unsupported` compatibility issue.

---

## v2.14.1
> *2026/05/09*

### ✨ Added

- **MCP**: Added `StreamableHTTP` transmission protocol support for MCP.

---

## v2.14.0
> *2026/05/08*

### ✨ Added

- **Sync**: Enhanced Git autocommit sync, adding synchronization support for Obsidian and custom folder configurations.
- **Backup**: Hardened backup security, integrating random password generation when exporting backup ZIP archives.
- **Architecture**: Separated page and API services. The management backend and shared pages can now run on independent ports to meet deployment requirements in different environments.

### 🛠️ Fixed

- **Security**: Optimized upgrade parameter filtering, applying strict sanitization to management APIs to prevent unexpected risks.
- **Security**: Optimized API security, removing sensitive data output from public interfaces.

### 🚀 Optimized

- **Architecture**: Optimized router modularization, refactoring the single router file into functional chunk-based router structures.

---

## v2.13.8
> *2026/05/07*

### ✨ Added

- **Security**: Added a safety mechanism to automatically forbid new user registrations when no administrator account is configured.

### 🛠️ Fixed

- **Metrics**: Fixed file statistics drift. Fixed an issue where the file count was inaccurate because the `Rename` flag was not properly reset on reused records during note renaming or moving.
- **MCP**: Fixed MCP SSE connection compatibility, adjusting path returns of the MCP SSE endpoint from relative to absolute paths to resolve connection timeouts on certain clients.

### 🚀 Optimized

- **Privacy**: Blocked search engine crawlers from indexing all FNS service pages.

---

## v2.13.7
> *2026/05/05*

### ✨ Added

- **API**: Added attachment upload endpoint `POST /api/file` with `multipart/form-data` upload and auto-sync support.

### 🛠️ Fixed

- **Network**: Fixed CORS middleware blocker, allowing cross-origin requests for the `/api/health` endpoint to avoid issues with monitoring integrations.

### 🚀 Optimized

- **Sync**: Added background auto-deduplication tasks for notes and attachments, resolving historical record redundancy caused by hash mismatches.

---

## v2.13.6
> *2026/05/03*

### 🛠️ Fixed

- **Sync**: Fixed synchronization time windows and broadcast message integrity in multi-client incremental sync.

---

## v2.13.5
> *2026/05/02*

### 🛠️ Fixed

- **Network**: Fixed a CORS connection issue under Android systems where devices failed to establish connections.

---

## v2.13.4
> *2026/05/02*

### 🛠️ Fixed

- **Network**: Fixed mobile API failures where direct requests using the `capacitor://` protocol header were blocked by the server.

---

## v2.13.3
> *2026/05/02*

### 🛠️ Fixed

- **Database**: Fixed PostgreSQL FTS (Full-Text Search) compatible index token field length insufficiency.

---

## v2.13.2
> *2026/04/27*

### 🛠️ Fixed

- **Network**: Fixed potential security vulnerabilities in CORS cross-origin requests.

---

## v2.13.1
> *2026/04/27*

### 🚀 Optimized

- **Admin**: Optimized vault deletion logic. Deleting a note vault now immediately purges all physically associated file resources.

### 🛠️ Fixed

- **Sync**: Fixed rename operations failing to properly reset the modification time (`mtime`) flag.

---

## v2.13.0
> *2026/04/26*

### 🚀 Optimized

- **Sync**: Optimized message acknowledgment (ACK) mechanisms.
- **Sync**: Optimized hash algorithm performance. Files larger than 100MB are now calculated using a split-segment technique, dramatically lowering CPU load during large uploads.

---

## v2.12.9
> *2026/04/24*

### 🛠️ Fixed

- **Security**: Fixed potential SQL injection vulnerabilities in the third-party library `pgx < 5.9.2` due to incorrect processing of dollar-sign quotes.

---

## v2.12.8
> *2026/04/24*

### ✨ Added

- **Update**: Added active notifications to connected clients when a new server release is available.

---

## v2.12.7
> *2026/04/23*

### 🛠️ Fixed

- **WebGUI**: Fixed text overlapping issues in the note revision history diff view.

---

## v2.12.6
> *2026/04/22*

### ✨ Added

- **Auth**: Added administrative authorization checking.
- **Build**: Optimized build pipeline, compiling the changelog directly into the server executable binary.

---

## v2.12.5
> *2026/04/21*

### 🛠️ Fixed

- **Logging**: Fixed incorrect time format displays in sync update logs.

---

## v2.12.4
> *2026/04/21*

### ✨ Added

- **Logging**: Added support for client type and version logging in both sync update logs and note history.
- **MCP**: Fixed MCP sessions failing to record client identification.

---

## v2.12.2
> *2026/04/21*

### 🛠️ Fixed

- **Logging**: Fixed missing client fields and vault name display anomalies in note update logs.

---

## v2.12.1
> *2026/04/21*

### 🛠️ Fixed

- **Logging**: Fixed missing client identification fields in update logs.

---

## v2.12.0
> *2026/04/21*

### ✨ Added

- **Logging**: Added note vault update logging.

### 🚀 Optimized

- **WebGUI**: Optimized note vault configuration page UI, updating navigation icons.

---

## v2.11.13
> *2026/04/20*

### ✨ Added

- **MCP**: Added support for HEAD request checks in MCP protocol context.

---

## v2.11.12
> *2026/04/19*

### 🚀 Optimized

- **Concurrency**: Added write concurrency deduplication for notes, attachments, and configurations to achieve atomic writes without relying solely on slow database transaction isolation.
- **API**: Implemented RESTful interfaces for note property updates along with API documentation updates.

---

## v2.11.11
> *2026/04/17*

### ✨ Added

- **Architecture**: Added build and running support for the ARMv7 Little-endian target platform.

---

## v2.11.10
> *2026/04/17*

### 🛠️ Fixed

- **Share**: Fixed embedded images in notes failing to load in online shared and read views under certain routing scenarios.

---

## v2.11.9
> *2026/04/16*

### 🚀 Optimized

- **MCP**: Optimized MCP SSE long connection keep-alive heartbeat mechanisms.
- **WebGUI**: Optimized server online state detection and update mechanics.
- **WebGUI**: Unified API headers for all RESTful requests to simplify centralized authorization.
- **Search**: Removed regex search capability for notes to reduce project complexity.

---

## v2.11.8
> *2026/04/15*

### 🛠️ Fixed

- **MCP**: Fixed MCP default vault configurations failing to take effect.

### 🚀 Optimized

- **MCP**: Optimized TCP connections, introducing server-side heartbeats to prevent unexpected client disconnects.

---

## v2.11.7
> *2026/04/14*

### 🛠️ Fixed

- **Backup**: Fixed backup export, skipping missing file anomalies automatically to prevent archiving job terminations.
- **Sync**: Solved attachment download races, fixing duplicate downloads under highly concurrent scenarios.
- **Database**: Fixed SQL query errors, adding missing `GROUP BY` fields in the search interface.
- **MCP**: Fixed MCP SSE disconnection anomalies by adjusting middleware boundaries to avoid connections being killed by `ContextTimeout`.

---

## v2.11.6
> *2026/04/12*

### 🛠️ Fixed

- **Sync**: Fixed duplicate downloads of attachments under rare synchronization states.

---

## v2.11.5
> *2026/04/12*

### 🚀 Optimized

- **Search**: Migrated Full-text search engine to a Go-native inverted index.

### 🛠️ Fixed

- **Database**: Fixed model field type overflow issues under multi-database deployments.

---

## v2.11.4
> *2026/04/09*

### 🛠️ Fixed

- **Share**: Fixed synchronization issues in shared status.
- **Sync**: Integrated automatic merging of shared records upon note renaming.

---

## v2.11.3
> *2026/04/02*

### ✨ Added

- **Share**: Added incremental sync API endpoints for shared notes.

---

## v2.11.1
> *2026/04/02*

### 🛠️ Fixed

- **Test**: Fixed data testing coverage issues.
- **Admin**: Fixed sensitive password obfuscation issues where cloud storage passwords could not be accessed.

---

## v2.11.0
> *2026/04/01*

### ✨ Added

- **Database**: Added support for PostgreSQL and MySQL databases, greatly enhancing concurrent throughput capabilities.

### 🚀 Optimized

- **WebGUI**: Optimized settings views in the administration UI.

---

## v2.10.2
> *2026/03/28*

### ✨ Added

- **MCP**: Added configurations to set a default note vault for MCP.

### 🛠️ Fixed

- **Performance**: Fixed Out-of-Memory (OOM) failures when loading large vaults into memory.

---

## v2.10.1
> *2026/03/27*

### 🛠️ Fixed

- **WebGUI**: Fixed API error handling inside the administration backend.

---

## v2.10.0
> *2026/03/27*

### ✨ Added

- **MCP**: Native Model Context Protocol (MCP) support. The server can now be connected as an MCP backend to AI environments like Cherry Studio or Cursor, granting AI reading and writing capabilities to private notes and syncing across endpoints.

---

## v2.9.3
> *2026/03/24*

### 🛠️ Fixed

- **Share**: Fixed inaccurate counts in active shared lists.

---

## v2.9.2
> *2026/03/24*

### 🛠️ Fixed

- **Share**: Fixed empty titles in short URL redirections when direct redirection masks were enabled.

---

## v2.9.1
> *2026/03/24*

### 🛠️ Fixed

- **Share**: Fixed redirection errors and duplicate generation issues in short URLs.

---

## v2.9.0
> *2026/03/24*

### ✨ Added

- **Share**: Unified administration workflows by merging note vault settings and shared list views.
- **Share**: Added password authentication prompts for shared note views.
- **Share**: Supported short URL generation for shared documents.

---

## v2.8.0
> *2026/03/22*

### ✨ Added

- **Share**: Added standalone administration view for shared resources.
- **WebGUI**: Implemented sharing creation mechanics from within the Web UI.

---

## v2.7.3
> *2026/03/19*

### 🛠️ Fixed

- **Logic**: Fixed shutdown sequence cleanup deficiencies.

---

## v2.7.2
> *2026/03/15*

### 🛠️ Fixed

- **Sync**: Fixed WebDAV mirror sync failures.
- **Automation**: Fixed history days parameter restoration errors in Git automated synchronization.

### 🚀 Optimized

- **Automation**: Configured git signatures to load dynamically from config preferences.

---

## v2.7.1
> *2026/03/13*

### ✨ Added

- **Share**: Added API interfaces to cancel shares.

---

## v2.7.0
> *2026/03/12*

### ✨ Added

- **Share**: Introduced fundamental note-sharing functions.

---

## v2.6.0
> *2026/03/07*

### ✨ Added

- **WebGUI**: Implemented on-the-fly rendering capability for Canvas elements in the note editor.
- **WebGUI**: Implemented online administration for note vault configuration files.
- **Admin**: Implemented active WebSocket connections manager.

### 🚀 Optimized

- **WebGUI**: Redesigned renaming mechanisms for files and folders.

---

## v2.5.5
> *2026/03/05*

### 🚀 Optimized

- **Sync**: Optimized sync mechanisms, resolving offline deletions failing to broadcast under certain conditions.

---

## v2.5.4
> *2026/03/03*

### ✨ Added

- **Settings**: Added update check toggle.
- **WebGUI**: Supported side-by-side editing and previewing modes.

### 🛠️ Fixed

- **WebGUI**: Fixed responsive styling adjustments on mobile devices.

---

## v2.5.3
> *2026/03/03*

### 🚀 Optimized

- **Release**: Built for CNB Registry validation testing.

---

## v2.5.2
> *2026/03/02*

### 🚀 Optimized

- **Sync**: Moved WebSocket tuning variables into `config.yaml` for flexible operations.

---

## v2.5.1
> *2026/03/02*

### 🚀 Optimized

- **Sync**: Resolved error logging anomalies during massive parallel attachment uploads.

---

## v2.5.0
> *2026/03/02*

### ✨ Added

- **WebGUI**: Upgraded editor baseline to CodeMirror 6.
- **WebGUI**: Implemented support for Obsidian Callouts rendering styles.
- **WebGUI**: Supported navigation triggers for Obsidian internal wikilinks.

### 🚀 Optimized

- **Sync**: Avoided writing empty placeholders during upload phases; deferred folder generation calls to avoid performance stalls in active syncing.

---

## v2.4.4
> *2026/03/01*

### 🛠️ Fixed

- **Performance**: Resolved high GPU loads caused by active loop animations in WebGUI.

### 🚀 Optimized

- **Performance**: Increased default write queue thresholds to 1000 items.

---

## v2.4.3
> *2026/03/01*

### 🛠️ Fixed

- **Sync**: Fixed execution errors on Git synchronizations.
- **Metrics**: Rectified connection count estimation.
- **Sync**: Fixed sync marker initialization bugs during history rollbacks.

---

## v2.4.1
> *2026/03/01*

### 🛠️ Fixed

- **Update**: Resolved configuration migrations blocking rolling updates.

---

## v2.4.0
> *2026/03/01*

### ✨ Added

- **WebGUI**: Integrated resource gzip/brotli asset optimizations.
- **WebGUI**: Integrated CodeMirror 6 editor base.
- **Trash**: Added cleanups and purging functions for trashed vaults.

---

## v2.3.0
> *2026/02/28*

### ✨ Added

- **Tunnel**: Integrated cloud tunnel setups (Ngrok / Cloudflare Tunnel) to support secure public HTTP/WebSocket accesses without proxy configurations.

---

## v2.2.1
> *2026/02/27*

### 🚀 Optimized

- **Sync**: Restrained duplicate runs for automated backup jobs.
- **Security**: Disabled administrative ports on `9001` by default.

---

## v2.2.0
> *2026/02/26*

### 🚀 Optimized

- **Sync**: Reduced message packet sizes during active sync, enhancing connectivity resilience (requires plugin v1.16+).

---

## v2.1.4
> *2026/02/26*

### 🛠️ Fixed

- **Sync**: Provided quick stabilization fix for large-packet connections dropped by middleboxes.

---

## v2.1.3
> *2026/02/25*

### 🛠️ Fixed

- **WebDAV**: Fixed authentication handling under WebDAV.
- **Backup**: Fixed automated jobs overriding storage databases with empty states during failed read procedures.

---

## v2.1.2 / v2.1.1 / v2.1.0
> *2026/02/25*

### ✨ Added

- **Update**: Integrated online version upgrades.

### 🛠️ Fixed

- **i18n**: Resolved locale preferences failing to load or stick.

---

## v2.0.10
> *2026/02/24*

### 🛠️ Fixed

- **Cloud Storage**: Addressed credential validation problems inside S3/OSS adapters.

---

## v2.0.9
> *2026/02/24*

### 🛠️ Fixed

- **CI/CD**: Addressed version flag errors inside automated GitHub Action packaging.

---

## v2.0.8
> *2026/02/24*

### 🛠️ Fixed

- **Update**: Fixed upgrade validation alert triggers.

---

## v2.0.7
> *2026/02/24*

### 🛠️ Fixed

- **Sync**: Solved path hashing bugs occurring on special character paths, fixing redundant directory synchronization, file duplication, and deletion faults.

---

## v2.0.6 / v2.0.5
> *2026/02/24*

### 🛠️ Fixed

- **Backup**: Prevented backup routines from running continuously triggered by file write operations.

### 🚀 Optimized

- **Automation**: Implemented synchronization backing for unitialized repository spaces.

---

## v2.0.4
> *2026/02/24*

### 🛠️ Fixed

- **Sync**: Addressed missing creation times inside uploaded assets (requires client 1.15+).

---

## v2.0.3
> *2026/02/23*

### 🛠️ Fixed

- **WebGUI**: Solved document rendering freezes in edit windows.
- **i18n**: Restored missing translation dictionaries.

### 🚀 Optimized

- **WebGUI**: Implemented theme automation switching, adding support for customized application icons.

---

## v2.0.2
> *2026/02/23*

### 🛠️ Fixed

- **Database**: Addressed table migration constraints.

---

## v2.0.1
> *2026/02/23*

### 🚀 Optimized

- **System**: Shipped major version 2.0 release, bringing robust performance gains and fundamental stability enhancements.

---

## v1.16.2
> *2026/02/14*

### 🚀 Optimized

- **WebGui**: Adjusted WebGui interface position.
- **Features**: Added display of server information.
- **Performance**: Optimized list height for zero-copy access to fix issues with low display height in various lists.
- **Sync**: Optimized note/attachment sync logic.

---

## v1.16.1
> *2026/02/14*

### 🚀 Optimized

- **WebGui**: Adjusted WebGui interface position.
- **Features**: Added display of server information.

---

## v1.15.11
> *2026/02/14*

### 🚀 Optimized

- **WebGui**: Optimized WebGui interface and added URL support.

---

## v1.15.10
> *2026/02/14*

### 🚀 Optimized

- **Architecture**: Adjusted service toolkit.
- **API**: Adjusted API response structure.

---

## v1.15.9
> *2026/02/14*

### ✨ Added

- **Tools**: Added access entry for fns docs and ws debug tools.

---

## v1.15.8
> *2026/02/13*

### 🛠️ Fixed

- **Stability**: Fixed minor BUG in time processing.

---

## v1.15.7
> *2026/02/13*

### 🛠️ Fixed

- **Sync**: Fixed issue with offline deletion not clearing local hash table.

---

## v1.15.6
> *2026/02/13*

### 🛠️ Fixed

- **Scripts**: Fixed fns shortcut script running issue on macOS.
- **Logging**: Fixed log printing content.

---

## v1.15.5
> *2026/02/12*

### 🚀 Optimized

- **CI/CD**: Adjusted GitHub Action to use go mod version for building and publishing.

---

## v1.15.4
> *2026/02/12*

### ✨ Added

- **Sync**: Added feature to clear note configuration related messages.

---

## v1.15.3
> *2026/02/10*

### 🛠️ Fixed

- **Folder**: Added fallback solution for duplicate folders and startup task to clear duplicates.

---

## v1.15.2
> *2026/02/09*

### 🚀 Optimized

- **Database**: Optimized DB performance and structure, performed batch formatting.

---

## v1.15.1
> *2026/02/07*

### ✨ Added

- **Folder**: Added folder management features, including models and related logic.
- **Sync**: Fixed potential data race issues and optimized note/attachment renaming.

---

## v1.14.1
> *2026/01/31*

### ✨ Added

- **Trash**: Added trash and batch recovery for attachment management.

### 🛠️ Fixed

- **Stability**: Fixed issue where resources were not created correctly due to identical modified time and content in attachments/config files.

### 🚀 Optimized

- **API**: Optimized attachment view/download interfaces with zero-copy access.
- **WebGui**: Fixed low display height issues in various lists.

---

## v1.14.0
> *2026/01/31*

### ✨ Added

- **Trash**: Added trash for attachment management.
- **WebGui**: Added display of server information.
- **Sync**: Added note and attachment renaming features.

### 🛠️ Fixed

- **Stability**: Fixed potential data race issues.

---

## v1.13.0
> *2026/01/30*

### ✨ Added
- **Sync**: Added offline deletion synchronization for attachments, notes, and configs.
- **Sync**: Added auto-download of missing files in incremental sync mode.

---

## v1.12.0
> *2026/01/29*

### 🚀 Optimized
- **Language**: Translated/updated all code comments and documentation to bilingual (CN/EN) or English.
- **API**: Improved internationalization (i18n) for API response messages.
- **Stability**: Fixed automatic resource prefix issues.
- **API**: Added API extensions: edit operations, backlinks, and health checks.

---

## v1.11.3
> *2026/01/27*

### 🛠️ Fixed
- **Attachment**: Fixed attachment download timeout (30s) error; now configurable, default is 1 hour.

---

## v1.11.2
> *2026/01/27*

### ✨ Added
- **WebGui**: Added Obsidian SSO auto-authorization mechanism.

### 🚀 Optimized
- **WebGui**: Improved authorization configuration UI.

---

## v1.11.1
> *2026/01/26*

### 🚀 Optimized
- **Release**: Adjusted version release workflow.

---

## v1.11.0
> *2026/01/26*

### ✨ Added
- **Feature**: Added version detection and version information retrieval features.

---

## v1.10.8
> *2026/01/26*

### ✨ Added
- **API**: Added attachment status detection interface.

---

## v1.10.7
> *2026/01/25*

### 🛠️ Fixed
- **Stability**: Fixed server crash caused by consistency checks during file uploads.

---

## v1.10.6
> *2026/01/24*

### ✨ Added
- **WebGui**: Added pagination for the attachment management page.

---

## v1.10.5
> *2026/01/23*

### 🛠️ Fixed
- **Trash**: Fixed issues when restoring notes/versions from the trash and history.

---

## v1.10.4
> *2026/01/23*

### 🛠️ Fixed
- **Attachment**: Fixed connection drops during attachment uploads and lowered error logging level for shard upload failures.

---

## v1.10.3
> *2026/01/20*

### 🚀 Optimized
- **WebGui**: Replaced zoom effect in note vault list with a selected shadow effect.

### 🛠️ Fixed
- **WebGui**: Fixed a bug where note vaults with special characters in their names were inaccessible.

---

## v1.10.2
> *2026/01/20*

### 🛠️ Fixed
- **Admin**: Fixed bugs preventing new user registration and the ability to disable user registration.

---

## v1.10.1
> *2026/01/20*

### 🛠️ Fixed
- **Admin**: Fixed issues with new user registration.

---

## v1.10.0
> *2026/01/19*

### ✨ Added
- **Attachment**: Added attachment management functionality.
- **Auth**: Added configuration for Token expiration time.
- **Share**: Added interfaces for sharing functionality.
- **Docs**: Added Swagger API documentation.

### 🚀 Optimized
- **WebGui**: Adjusted WebGui deployment path.
- **API**: Refined API error messages.

### 🛠️ Fixed
- **WebGui**: Fixed notice issues caused by WebGui auto-translation.

---

## v1.9.1
> *2026/01/14*

### 🚀 Optimized
- **WebGui**: Added blue color scheme and optimized editor display.

---

## v1.9.0
> *2026/01/14*

### ✨ Added
- **WebGui**: Complete UI refactor (contributed by @ZyphrZero).
- **WebGui**: Replaced editor with Vditor, supporting rich text and Markdown real-time rendering.
- **WebGui**: Supported custom note search, list field sorting, and color themes.
- **WebGui**: Added dark mode, online version detection, and trash restoration.
- **Settings**: Added historical version retention and save delay settings.

### 🚀 Optimized
- **Security**: Optimized service token encryption obfuscation characters.

---

## v1.8.1
> *2026/01/12*

### 🔄 Changed
- **Architecture**: Introduced DDD layered architecture (contributed by @ZyphrZero), removed global variables, and implemented Dependency Injection pattern.

### 🚀 Optimized
- **Sync**: Optimized offline note merging with line-level conflict detection and 3-way merge.
- **Performance**: Added Worker Pool and Per-User Write Queue to solve SQLite concurrency lock issues.
- **WebSocket**: Optimized Context lifecycle management and enhanced TraceID tracking.

### 🛠️ Fixed
- **Logic**: Fixed a bug where note renaming could lead to note loss and errors.

---

## v1.7.3
> *2026/01/09*

### 🛠️ Fixed
- **Database**: Added友好 error message for database creation failures.

---

## v1.7.2
> *2026/01/09*

### ✨ Added
- **WebGui**: Added configuration settings functionality and related interfaces.
- **Admin**: Added Admin ID setting.

---

## v1.7.1
> *2026/01/09*

### ✨ Added
- **Sync**: Added offline device note editing merge functionality (requires plugin v1.7+).

---

## v1.6.3
> *2026/01/08*

### 🚀 Optimized
- **WebGui**: Optimized note list search.
- **WebGui**: Added icon display.
- **WebGui**: Added attachment display and refresh button in note vault.

### 🛠️ Fixed
- **Stability**: Fixed potential exceptions during concurrent queries.

---

## v1.6.1
> *2026/01/07*

### 🚀 Optimized
- **Performance**: Optimized sync efficiency and data processing for large note vaults (requires plugin v1.6+).
- **Cache**: Added browser caching mechanism for static content.

> [!CAUTION]
> This version involves database structure optimization. It is recommended to delete the DB file under `storage/database` on the server; note modification history will be regenerated.

---

## v1.5.4
> *2026/01/06*

### 🛠️ Fixed
- **Attachment**: Fixed occasional errors when uploading attachments.

---

## v1.5.3
> *2026/01/06*

### 🚀 Optimized
- **WebGui**: Lazy-loaded editing features to improve home page loading speed.

---

## v1.5.2
> *2026/01/05*

### 🛠️ Fixed
- **Sync**: Fixed inaccurate sync task progress display.

---

## v1.5.1
> *2026/01/04*

### 🛠️ Fixed
- **Logic**: Fixed a bug where notes couldn't be deleted properly after renaming.
- **Stability**: Fixed WebSocket connection resets during large-scale note synchronization.
- **i18n**: Fixed WebGui API language errors.

---

## v1.5.0
> *2026/01/04*

### ✨ Added
- **Trash**: Added note trash bin feature.
- **WebGui**: Added user status detection.
- **WebGui**: Added registration closed detection on the sign-up page.
- **WebGui**: Added keyboard shortcut support for operation confirmations.

### 🚀 Optimized
- **WebGui**: Improved note editing user experience.
- **Database**: Optimized and resolved database concurrent access issues.

### 🛠️ Fixed
- **Script**: Fixed a bug where shortcut scripts might overwrite configuration files.

---

## v1.4.7
> *2026/01/03*

### 🛠️ Fixed
- **Database**: Attempted to solve SQLite concurrency issues and corrected internal error codes.

---

## v1.4.6
> *2026/01/03*

### 🛠️ Fixed
- **Docker**: Fixed an issue where the `temp` directory did not exist in Docker environments.

---

## v1.4.5
> *2026/01/03*

### 🛠️ Fixed
- **Sync**: Fixed an issue where attachments couldn't be synced during initial or full sync (requires plugin v1.5.14+).

---

## v1.4.4
> *2026/01/02*

### 🛠️ Fixed
- **Access**: Fixed accessibility issues with titles containing Emojis.

### ✨ Added
- **Docs**: Added help file.

---

## v1.4.3
> *2026/01/02*

### 🔄 Changed
- **Vault**: Note vault deletion operation changed to soft delete.

---

## v1.4.2
> *2026/01/01*

### ✨ Added
- **WebGui**: Added a red confirmation popup for note deletions to prevent accidental deletion.

---

## v1.4.1
> *2025/12/31*

### 🚀 Optimized
- **API**: Added ETag browser caching for note resource (images, etc.) download interface to improve loading speed.

---

## v1.4.0
> *2025/12/31*

### ✨ Added
- **WebGui**: Added maximize button to enhance full-screen editing experience.
- **WebGui**: Supported display of Obsidian embedded images, PDFs, and other attachments in note view.
- **API**: Added resource download interface.

---

## v1.3.8
> *2025/12/31*

### 🚀 Optimized
- **Server**: Established a content hash version repository for notes to facilitate future tracing, comparison, and merging.

---

## v1.3.7
> *2025/12/30*

### 🛠️ Fixed
- **Stability**: Added panic recovery for tasks and upgrade scripts to prevent service crashes.
- **Stability**: Fixed Nil Pointer Panic issues in various layers.

---

## v1.3.6
> *2025/12/30*

### 🛠️ Fixed
- **Task Management**: Fixed errors in the task manager.

---

## v1.3.5
> *2025/12/30*

### 🚀 Optimized
- **WebGui**: Optimized note viewing display.
- **Script**: Optimized one-click installation/management script.

---

## v1.3.4
> *2025/12/30*

### 🛠️ Fixed
- **Sync**: Fixed sync command processing errors leading to incorrect file synchronization across clients.
- **Script**: Fixed one-click scripts closing the service upon `Ctrl+C`.

---

## v1.3.3
> *2025/12/29*

### 🛠️ Fixed
- **Sync**: Resolved potential update confusion across multiple note vaults for a single user.

---

## v1.3.2
> *2025/12/28*

### ✨ Added
- **i18n**: Added support for multi-language environments.

### 🚀 Optimized
- **WebGui**: Optimized note version diff display.

---

## v1.3.1
> *2025/12/28*

### 🚀 Optimized
- **Logic**: Optimized logic for note title modification.

---

## v1.3.0
> *2025/12/28*

### ✨ Added
- **WebGui**: Added setting for users to control WebGui font settings.

---

## v1.2.6
> *2025/12/27*

### 🚀 Optimized
- **WebGui**: Optimized font loading logic to avoid UI stuttering.

---

## v1.2.5
> *2025/12/27*

### ✨ Added
- **Client**: Added record support for client names.

### 🚀 Optimized
- **Cleanup**: Added sync cleanup logic after note renaming.

---

## v1.2.4
> *2025/12/27*

### 🛠️ Fixed
- **WebGui**: Fixed display bug when history version content is empty.

---

## v1.2.3
> *2025/12/27*

### ✨ Added
- **API**: Added note history related interfaces and functions.

### 🚀 Optimized
- **Database**: Optimized database query efficiency.
- **WebGui**: Changed WebGui display font and fixed various display bugs.

### 🛠️ Fixed
- **Stability**: Fixed issues during high concurrent access.

---

## v1.2.2
> *2025/12/27*

### 🛠️ Fixed
- **WebGui**: Fixed blank page issues caused by empty note history.

---

## v1.2.1
> *2025/12/27*

### ✨ Added
- **API**: Added note history related interfaces and functions.

### 🚀 Optimized
- **Database**: Optimized database query efficiency.
- **Stability**: Resolved stability issues during high concurrent access.

---

## v1.0.4
> *2025/12/26*

### 🛠️ Fixed
- **WebGui**: Fixed blank display issues caused by WebGui build exceptions.

---

## v1.0.3
> *2025/12/25*

### 🛠️ Fixed
- **WebGui**: Resolved layout issues caused by long note titles.

---

## v1.0.2
> *2025/12/25*

### 🚀 Optimized
- **Attachment**: Optimized attachment upload logic, significantly reducing upload time.

### 🛠️ Fixed
- **CI/CD**: Corrected GitHub Action update limits.

---

## v1.0.1
> *2025/12/23*

### 🛠️ Fixed
- **Permission**: Fixed permission issues during upload on some systems.

---

## v1.0.0
> *2025/12/22*

### ✨ Added
- **Sync**: Added configuration file synchronization features and interfaces.

### 🚀 Optimized
- **Script**: Optimized script output display.

### 🛠️ Fixed
- **Script**: Fixed script execution control failures.

---

## v0.11.5
> *2025/12/19*

### 🛠️ Fixed
- **Docker**: Fixed Docker image execution issues.

---

## v0.11.4
> *2025/12/18*

### ✨ Added
- **Auth**: Added version information downlink in the authorization validation interface.

---

## v0.11.3
> *2025/12/16*

### ✨ Added
- **Cleanup**: Added auto-cleanup tasks on startup and Session auto-cleanup logic.

### 🛠️ Fixed
- **Stability**: Fixed abnormal exit issues during high concurrency due to connection closures.

---

## v0.11.2
> *2025/12/15*

### 🛠️ Fixed
- **Stability**: Fixed abnormal exit issues during concurrency due to connection closures.

---

## v0.11.1
> *2025/12/14*

### ✨ Added
- **Architecture**: Added prefix to messages for future business expansion.

---

## v0.10.2
> *2025/12/12*

### ✨ Added
- **Settings**: Added shard settings for upload/download (default 512KB).

---

## v0.10.1
> *2025/12/12*

### ✨ Added
- **Feature**: Added binary file download feature.
- **Feature**: Added WebSocket chunked download feature.
- **Feature**: Added version control management.

---

## v0.9.6
> *2025/12/11*

- Initial release (recording started).
