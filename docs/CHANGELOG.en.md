# CHANGELOG

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

## v1.7.2
> *2026/01/09*

### ✨ Added
- **WebGui**: Added configuration settings function and related interfaces.
- **Admin**: Added Administrator ID setting.

---

## v1.7.1
> *2026/01/09*

### ✨ Added
- **Sync**: Added offline device note editing and merging functionality (requires Plugin v1.7+).

---

## v1.6.3
> *2026/01/08*

### 🚀 Optimized
- **WebGui**: Optimized note list search functionality.
- **WebGui**: Added icon display.
- **WebGui**: Added attachment display and refresh buttons to note vaults.

### 🛠️ Fixed
- **Stability**: Fixed exceptions that may occur during concurrent queries.

---

## v1.6.1
> *2026/01/07*

### 🚀 Optimized
- **Performance**: Optimized large note vault synchronization efficiency and data processing (requires Plugin v1.6+).
- **Cache**: Added browser caching mechanism for static content.

> [!CAUTION]
> This version involves database structure optimization. It is recommended to delete the original DB file in the `storage/database` directory of the server; note modification history will be regenerated.

---

## v1.5.4
> *2026/01/06*

### 🛠️ Fixed
- **Attachment**: Fixed occasional errors when uploading attachments.

---

## v1.5.3
> *2026/01/06*

### 🚀 Optimized
- **WebGui**: Deferred loading for editing functions to improve home page loading speed.

---

## v1.5.2
> *2026/01/05*

### 🛠️ Fixed
- **Sync**: Fixed inaccurate progress display for synchronization tasks.

---

## v1.5.1
> *2026/01/04*

### 🛠️ Fixed
- **Logic**: Fixed issue where notes could not be deleted normally after being renamed.
- **Stability**: Fixed issue where large-scale note synchronization caused WebSocket connections to reset.
- **I18n**: Fixed incorrect interface language in WebGui.

---

## v1.5.0
> *2026/01/04*

### ✨ Added
- **Recycle Bin**: Added note recycle bin function.
- **WebGui**: Added user status detection.
- **WebGui**: Added registration closure detection to the registration page.
- **WebGui**: Added keyboard shortcut support for operation confirmation.

### 🚀 Optimized
- **WebGui**: Improved note editing page experience.
- **Database**: Optimized and resolved database concurrent access issues.

### 🛠️ Fixed
- **Script**: Fixed issue where the shortcut script might overwrite the configuration file.

---

## v1.4.7
> *2026/01/03*

### 🛠️ Fixed
- **Database**: Attempted to resolve SQLite concurrency issues and corrected internal error codes.

---

## v1.4.6
> *2026/01/03*

### 🛠️ Fixed
- **Docker**: Fixed issue where running under Docker reported that the `temp` directory did not exist.

---

## v1.4.5
> *2026/01/03*

### 🛠️ Fixed
- **Sync**: Fixed issue where attachments could not be synchronized during the first or full sync (requires Plugin v1.5.14+).

---

## v1.4.4
> *2026/01/02*

### 🛠️ Fixed
- **Access**: Fixed issue where Emoji titles could not be accessed.

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
- **WebGui**: Added a red secondary confirmation pop-up for note deletion to prevent accidental deletion.

---

## v1.4.1
> *2025/12/31*

### 🚀 Optimized
- **API**: Added ETag browser caching mechanism to the note resource (images, etc.) download interface to improve loading speed.

---

## v1.4.0
> *2025/12/31*

### ✨ Added
- **WebGui**: Added maximization button to improve full-screen editing experience.
- **WebGui**: Added support for displaying Obsidian embedded images, PDFs, and other attachments in the note view page.
- **API**: Added resource download interface.

---

## v1.3.8
> *2025/12/31*

### 🚀 Optimized
- **Server**: Established a content hash version repository for notes to facilitate subsequent traceability, comparison, and merging operations.

---

## v1.3.7
> *2025/12/30*

### 🛠️ Fixed
- **Stability**: Added crash recovery mechanism for tasks and upgrade scripts to avoid a single task error causing the entire service to crash.
- **Stability**: Fixed Panic issues caused by nil pointers in various previous layers.

---

## v1.3.6
> *2025/12/30*

### 🛠️ Fixed
- **Task Management**: Fixed error problems existing in the task manager.

---

## v1.3.5
> *2025/12/30*

### 🚀 Optimized
- **WebGui**: Optimized the display effect of the note view page.
- **Scripts**: Optimized one-click installation/management scripts.

---

## v1.3.4
> *2025/12/30*

### 🛠️ Fixed
- **Sync**: Fixed issue where synchronization command processing errors caused files to be incorrectly synchronized and created on all clients.
- **Scripts**: Fixed issue where the one-click script caused the already started service to be closed synchronously when `Ctrl+C` was pressed.

---

## v1.3.3
> *2025/12/29*

### 🛠️ Fixed
- **Sync**: Resolved the confusion update problem that might occur when a single user has multiple note repositories.

---

## v1.3.2
> *2025/12/28*

### ✨ Added
- **i18n**: Added support for multi-language environments.

### 🚀 Optimized
- **WebGui**: Optimized the display effect of note history difference comparison.

---

## v1.3.1
> *2025/12/28*

### 🚀 Optimized
- **Logic**: Optimized the logic processing flow when changing note titles.

---

## v1.3.0
> *2025/12/28*

### ✨ Added
- **WebGui**: Added settings option to allow users to control WebGui font settings.

---

## v1.2.6
> *2025/12/27*

### 🚀 Optimized
- **WebGui**: Optimized font loading logic to avoid interface lagging caused by font loading.

---

## v1.2.5
> *2025/12/27*

### ✨ Added
- **Client**: Added support for recording client names.

### 🚀 Optimized
- **Cleanup**: Added sync cleanup logic after note renaming.

---

## v1.2.4
> *2025/12/27*

### 🛠️ Fixed
- **WebGui**: Fixed display bug when the history version content is empty.

---

## v1.2.3
> *2025/12/27*

### ✨ Added
- **API**: Added interfaces and functions related to note history.

### 🚀 Optimized
- **Database**: Optimized database query efficiency.
- **WebGui**: Changed WebGui display font and fixed various display bugs.

### 🛠️ Fixed
- **Stability**: Fixed issues appearing during high-frequency concurrent access.

---

## v1.2.2
> *2025/12/27*

### 🛠️ Fixed
- **WebGui**: Corrected the blank page issue caused by empty note history.

---

## v1.2.1
> *2025/12/27*

### ✨ Added
- **API**: Added interfaces and functions related to note history.

### 🚀 Optimized
- **Database**: Optimized database query efficiency.
- **Stability**: Resolved stability issues during large amounts of concurrent access.

---

## v1.0.4
> *2025/12/26*

### 🛠️ Fixed
- **WebGui**: Corrected blank display problem caused by WebGui page build exceptions.

---

## v1.0.3
> *2025/12/25*

### 🛠️ Fixed
- **WebGui**: Resolved layout display issues caused by overly long note titles.

---

## v1.0.2
> *2025/12/25*

### 🚀 Optimized
- **Attachment**: Optimized attachment upload logic, significantly reducing upload time.

### 🛠️ Fixed
- **CI/CD**: Corrected GitHub Action update restriction issue.

---

## v1.0.1
> *2025/12/23*

### 🛠️ Fixed
- **Permission**: Fixed issues caused by insufficient permissions during upload on some systems.

---

## v1.0.0
> *2025/12/22*

### ✨ Added
- **Sync**: Added functions and interfaces related to configuration file synchronization.

### 🚀 Optimized
- **Scripts**: Optimized the output of display scripts.

### 🛠️ Fixed
- **Scripts**: Corrected issue where script control execution failed.

---

## v0.11.5
> *2025/12/19*

### 🛠️ Fixed
- **Docker**: Corrected Docker image execution failure issue.

---

## v0.11.4
> *2025/12/18*

### ✨ Added
- **Auth**: Added version information delivery function to the authorization verification interface.

---

## v0.11.3
> *2025/12/16*

### ✨ Added
- **Cleanup**: Added automatic cleanup tasks at program startup and session automatic cleanup logic.

### 🛠️ Fixed
- **Stability**: Corrected abnormal exit issues caused by closed connections under high concurrency.

---

## v0.11.2
> *2025/12/15*

### 🛠️ Fixed
- **Stability**: Corrected abnormal program exit caused by closed connections under high concurrency.

---

## v0.11.1
> *2025/12/14*

### ✨ Added
- **Architecture**: Added prefix to messages to facilitate subsequent business expansion.

---

## v0.10.2
> *2025/12/12*

### ✨ Added
- **Settings**: Added upload/download chunk setting (default 512KB).

---

## v0.10.1
> *2025/12/12*

### ✨ Added
- **Feature**: Added binary file download function.
- **Feature**: Added WebSocket chunked download function.
- **Feature**: Added version control management.

---

## v0.9.6
> *2025/12/11*

- Initial release (recording started).
