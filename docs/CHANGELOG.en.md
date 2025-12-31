# CHANGELOG

All notable changes to this project will be documented in this file.

This project adheres to [Keep a Changelog](https://keepachangelog.com/en/0.3.0/) standards.

---

## v1.4.0
> *2025/12/31*

### ✨ Added
- **WebGui**: Added maximum button to enhance full-screen editing experience.
- **WebGui**: Added support for displaying Obsidian embedded images, PDFs, and other attachments in the note view page.
- **API**: Added resource download interface.

---

## v1.3.8
> *2025/12/31*

### 🚀 Optimized
- **Server**: Established a content hash versioning library for notes to facilitate future traceability, comparison, and merging operations.

---

## v1.3.7
> *2025/12/30*

### 🛠️ Fixed
- **Stability**: Added crash recovery mechanism for tasks and upgrade scripts to prevent a single task error from crashing the entire service.
- **Stability**: Fixed panic issues caused by nil pointers at various levels.

---

## v1.3.6
> *2025/12/30*

### 🛠️ Fixed
- **Task Management**: Fixed reporting issues in the task manager.

---

## v1.3.5
> *2025/12/30*

### 🚀 Optimized
- **WebGui**: Optimized the display of the note view page.
- **Scripts**: Optimized the one-click installation/management script.

---

## v1.3.4
> *2025/12/30*

### 🛠️ Fixed
- **Sync**: Fixed an issue where sync command processing errors caused files to be incorrectly synchronization-created to all clients.
- **Scripts**: Fixed an issue where `Ctrl+C` in the one-click script would also close the started services.

---

## v1.3.3
> *2025/12/29*

### 🛠️ Fixed
- **Sync**: Resolved potential update confusion issues when using multiple note repositories for a single user.

---

## v1.3.2
> *2025/12/28*

### ✨ Added
- **Multi-language**: Added support for multi-language environments.

### 🚀 Optimized
- **WebGui**: Optimized the display of note history difference comparisons.

---

## v1.3.1
> *2025/12/28*

### 🚀 Optimized
- **Logic**: Optimized the logic processing flow when modifying note titles.

---

## v1.3.0
> *2025/12/28*

### ✨ Added
- **WebGui**: Added settings to allow users to control WebGui font settings.

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
- **WebGui**: Fixed a display bug caused by empty historical version content.

---

## v1.2.3
> *2025/12/27*

### ✨ Added
- **API**: Added note history related interfaces and functions.

### 🚀 Optimized
- **Database**: Optimized database query efficiency.
- **WebGui**: Modified WebGui display fonts and fixed various display bugs.

### 🛠️ Fixed
- **Stability**: Fixed issues occurring during high concurrency access.

---

## v1.2.2
> *2025/12/27*

### 🛠️ Fixed
- **WebGui**: Fixed a blank page issue caused by empty note history.

---

## v1.2.1
> *2025/12/27*

### ✨ Added
- **API**: Added note history related interfaces and functions.

### 🚀 Optimized
- **Database**: Optimized database query efficiency.
- **Stability**: Resolved stability issues during high concurrency access.

---

## v1.0.4
> *2025/12/26*

### 🛠️ Fixed
- **WebGui**: Fixed a blank display issue caused by build exceptions in the WebGui page.

---

## v1.0.3
> *2025/12/25*

### 🛠️ Fixed
- **WebGui**: Resolved layout display issues caused by excessively long note titles.

---

## v1.0.2
> *2025/12/25*

### 🚀 Optimized
- **Attachments**: Optimized attachment upload logic, significantly reducing upload time.

### 🛠️ Fixed
- **CI/CD**: Fixed update restriction issues in GitHub Actions.

---

## v1.0.1
> *2025/12/23*

### 🛠️ Fixed
- **Permissions**: Fixed permission issues during upload on some systems.

---

## v1.0.0
> *2025/12/22*

### ✨ Added
- **Sync**: Added configuration file synchronization related functions and interfaces.

### 🚀 Optimized
- **Scripts**: Optimized script output display.

### 🛠️ Fixed
- **Scripts**: Fixed script management execution failures.

---

## v0.11.5
> *2025/12/19*

### 🛠️ Fixed
- **Docker**: Fixed the issue where Docker images could not be executed.

---

## v0.11.4
> *2025/12/18*

### ✨ Added
- **Auth**: Added version information downlink functionality in the verification authorization interface.

---

## v0.11.3
> *2025/12/16*

### ✨ Added
- **Cleanup**: Added automatic cleanup tasks at startup and Session automatic cleanup logic.

### 🛠️ Fixed
- **Stability**: Fixed abnormal exit issues due to closed connections under high concurrency.

---

## v0.11.2
> *2025/12/15*

### 🛠️ Fixed
- **Stability**: Fixed abnormal program exit due to closed connections under concurrency.

---

## v0.11.1
> *2025/12/14*

### ✨ Added
- **Architecture**: Added prefixes to messages for future business expansion.

---

## v0.10.2
> *2025/12/12*

### ✨ Added
- **Settings**: Added upload and download chunk size settings (default 512KB).

---

## v0.10.1
> *2025/12/12*

### ✨ Added
- **Features**: Added binary file download functionality.
- **Features**: Added WebSocket chunked download functionality.
- **Features**: Added version control management.

---

## v0.9.6
> *2025/12/11*

- Initial version (records started).
