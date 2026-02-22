# CHANGELOG

All notable changes to this project will be documented in this file.

The project adheres to [Keep a Changelog](https://keepachangelog.com/en/0.3.0/) guidelines.

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
