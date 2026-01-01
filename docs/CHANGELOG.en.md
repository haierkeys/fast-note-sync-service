# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/0.3.0/).

---

## v1.4.2
> *2026/01/01*

### ✨ Added
- **WebGui**: Added a red double-confirmation dialog for note deletion to prevent accidental deletion.

---

## v1.4.1
> *2025/12/31*

### 🚀 Optimized
- **API**: Added ETag browser caching mechanism for note resource (images, etc.) download interface to improve loading speed.

---

## v1.4.0
> *2025/12/31*

### ✨ Added
- **WebGui**: Added a maximize button to enhance the full-screen editing experience.
- **WebGui**: Added support for normal display of Obsidian embedded images, PDFs, and other attachments in the note view page.
- **API**: Added a resource download interface.

---

## v1.3.8
> *2025/12/31*

### 🚀 Optimized
- **Server**: Established a content hash repository for notes to facilitate subsequent tracing, comparison, and merging operations.

---

## v1.3.7
> *2025/12/30*

### 🛠️ Fixed
- **Stability**: Added a crash recovery mechanism for tasks and upgrade scripts to prevent a single task error from causing the entire service to crash.
- **Stability**: Fixed Panic issues caused by nil pointers occurring at various levels.

---

## v1.3.6
> *2025/12/30*

### 🛠️ Fixed
- **Task Management**: Fixed error issues in the task manager.

---

## v1.3.5
> *2025/12/30*

### 🚀 Optimized
- **WebGui**: Optimized the display effect of the note view page.
- **Script**: Optimized the one-click installation/management script.

---

## v1.3.4
> *2025/12/30*

### 🛠️ Fixed
- **Sync**: Fixed an issue where sync command processing errors caused files to be incorrectly synced and created on all clients.
- **Script**: Fixed an issue where the one-click script caused started services to be closed simultaneously upon `Ctrl+C`.

---

## v1.3.3
> *2025/12/29*

### 🛠️ Fixed
- **Sync**: Resolved update confusion issues that might occur with multiple note repositories for a single user.

---

## v1.3.2
> *2025/12/28*

### ✨ Added
- **Multi-language**: Added support for multi-language environments.

### 🚀 Optimized
- **WebGui**: Optimized the display effect of note history difference comparison.

---

## v1.3.1
> *2025/12/28*

### 🚀 Optimized
- **Logic Processing**: Optimized the logic processing flow when changing note titles.

---

## v1.3.0
> *2025/12/28*

### ✨ Added
- **WebGui**: Added settings to allow users to control WebGui font settings.

---

## v1.2.6
> *2025/12/27*

### 🚀 Optimized
- **WebGui**: Optimized font loading logic to avoid interface lag caused by font loading.

---

## v1.2.5
> *2025/12/27*

### ✨ Added
- **Client**: Added support for recording client names.

### 🚀 Optimized
- **Cleanup Logic**: Added sync cleanup logic after note renaming.

---

## v1.2.4
> *2025/12/27*

### 🛠️ Fixed
- **WebGui**: Fixed a display bug caused when history version content is empty.

---

## v1.2.3
> *2025/12/27*

### ✨ Added
- **API**: Added note history-related interfaces and functions.

### 🚀 Optimized
- **Database**: Optimized database query efficiency.
- **WebGui**: Modified WebGui display fonts and fixed various display bugs.

### 🛠️ Fixed
- **Stability**: Fixed issues occurring during high-concurrency access.

---

## v1.2.2
> *2025/12/27*

### 🛠️ Fixed
- **WebGui**: Fixed the problem of blank pages when note history is empty.

---

## v1.2.1
> *2025/12/27*

### ✨ Added
- **API**: Added note history-related interfaces and functions.

### 🚀 Optimized
- **Database**: Optimized database query efficiency.
- **Stability**: Resolved stability issues during large-scale concurrent access.

---

## v1.0.4
> *2025/12/26*

### 🛠️ Fixed
- **WebGui**: Fixed the problem of blank display caused by WebGui page build exceptions.

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
- **CI/CD**: Fixed GitHub Action update limitation issues.

---

## v1.0.1
> *2025/12/23*

### 🛠️ Fixed
- **Permissions**: Fixed issues on some systems during upload due to insufficient permissions.

---

## v1.0.0
> *2025/12/22*

### ✨ Added
- **Sync**: Added configuration file synchronization related functions and interfaces.

### 🚀 Optimized
- **Script**: Optimized the output of the display script.

### 🛠️ Fixed
- **Script**: Fixed script control execution failure issues.

---

## v0.11.5
> *2025/12/19*

### 🛠️ Fixed
- **Docker**: Fixed the issue where Docker images cannot be executed.

---

## v0.11.4
> *2025/12/18*

### ✨ Added
- **Authentication**: Added the function of issuing version information in the authentication interface.

---

## v0.11.3
> *2025/12/16*

### ✨ Added
- **Cleanup**: Added automatic cleanup tasks upon program startup and Session automatic cleanup logic.

### 🛠️ Fixed
- **Stability**: Fixed abnormal exit issues due to connection closure under high concurrency.

---

## v0.11.2
> *2025/12/15*

### 🛠️ Fixed
- **Stability**: Fixed program abnormal exit due to connection closure under concurrency.

---

## v0.11.1
> *2025/12/14*

### ✨ Added
- **Architecture**: Added prefixes to messages to facilitate subsequent business function expansion.

---

## v0.10.2
> *2025/12/12*

### ✨ Added
- **Settings**: Added upload/download chunk settings (default 512KB).

---

## v0.10.1
> *2025/12/12*

### ✨ Added
- **Features**: Added binary file download function.
- **Features**: Added WebSocket chunked download function.
- **Features**: Added version control management.

---

## v0.9.6
> *2025/12/11*

- Initial version (recording starts).
