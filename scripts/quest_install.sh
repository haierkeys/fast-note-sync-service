#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

# ===========================================
# Fast-Note Sync Service 管理脚本 (Premium)
# ===========================================

REPO="haierkeys/fast-note-sync-service"
BIN_BASE="fast-note-sync-service"
INSTALL_DIR="/opt/fast-note"
BIN_PATH="$INSTALL_DIR/$BIN_BASE"
LINK_BIN="/usr/local/bin/fast-note"
INSTALLER_LINK="/usr/local/bin/fast-note-installer"
INSTALLER_SELF_PATH="/opt/fast-note/fast-note-installer.sh"
SERVICE_NAME="fast-note.service"
LOG_FILE="/var/log/fast-note.log"
TMPDIR="${TMPDIR:-/tmp}"
GITHUB_RAW="https://github.com/$REPO/releases/download"
GITHUB_API="https://api.github.com/repos/$REPO/releases"
SUDO=""

# --- 颜色系统 ---
_RED=$(tput setaf 1)
_GREEN=$(tput setaf 2)
_YELLOW=$(tput setaf 3)
_BLUE=$(tput setaf 4)
_MAGENTA=$(tput setaf 5)
_CYAN=$(tput setaf 6)
_BOLD=$(tput bold)
_ITALIC=$(tput sitm)
_DIM=$(tput dim)
_RESET=$(tput sgr0)

# --- 视觉装饰 ---
_INFO="  [i] "
_SUCCESS="  [+] "
_WARN="  [!] "
_ERROR="  [-] "
_STEP="  >> "

draw_banner() {
    clear
    load_version
    local ver_display="${_YELLOW}$L_NOT_INSTALLED${_RESET}"
    if [ -n "$INSTALLED_VER" ]; then
        local clean_v="${INSTALLED_VER#v}"
        ver_display="${_GREEN}$BIN_BASE v${clean_v}${_RESET}"
    fi

    cat <<EOF
${_CYAN}${_BOLD}
    ______           __     _   __      __          _____
   / ____/___ ______/ /_   / | / /___  / /____     / ___/__  ______  _____
  / /_  / __  / ___/ __/  /  |/ / __ \/ __/ _ \    \__ \/ / / / __ \/ ___/
 / __/ / /_/ (__  ) /_   / /|  / /_/ / /_/  __/   ___/ / /_/ / / / / /__
/_/    \__,_/____/\__/  /_/ |_/\____/\__/\___/   /____/\__, /_/ /_/\___/
                                                      /____/

       Fast Note Sync Service Manager Script
   ================================================
   $L_CUR_VER: $ver_display
${_RESET}
EOF
  echo -e "\n"
}

msg() { echo -e "${_BOLD}$1${_RESET}"; }
info() { echo -e "${_INFO}${_CYAN}$1${_RESET}"; }
success() { echo -e "${_SUCCESS}${_GREEN}$1${_RESET}"; }
warn() { echo -e "${_WARN}${_YELLOW}$1${_RESET}"; }
error() { echo -e "${_ERROR}${_RED}$1${_RESET}"; }
step() { echo -e "${_STEP}${_BLUE}$1${_RESET}"; }

# --- Language Support ---
LANG_CONF="$HOME/.fast-note-sync.lang"
VERSION_CONF="$INSTALL_DIR/.version"
CURRENT_LANG="en" # Default to English
INSTALLED_VER=""

save_lang() {
    echo "$CURRENT_LANG" > "$LANG_CONF" 2>/dev/null || true
}

load_lang() {
    local force_load="${1:-}"
    # Read from file only if forced or if it's the first time
    if [ "$force_load" = "init" ] && [ -f "$LANG_CONF" ]; then
        CURRENT_LANG=$(cat "$LANG_CONF" 2>/dev/null | tr -d '[:space:]' || echo "en")
    fi

    if [ "$CURRENT_LANG" = "zh" ]; then
        L_MENU_1="安装 / 升级服务"
        L_MENU_1_D="下载服务程序并自动安装管理工具至系统"
        L_MENU_2="启动服务"
        L_MENU_2_D="在后台启动同步服务"
        L_MENU_3="停止服务"
        L_MENU_3_D="终止正在运行的服务进程"
        L_MENU_4="服务状态"
        L_MENU_4_D="检查运行状态并预览最新日志"
        L_MENU_5="全部卸载"
        L_MENU_5_D="彻底移除程序、配置及所有日志文件"
        L_MENU_6="安装脚本到系统"
        L_MENU_6_D="将管理工具添加到全局快捷命令 fast-note-installer"
        L_MENU_0="退出"
        L_MENU_L="Switch to English (切换至英文)"
        L_SELECT="请选择"
        L_INPUT_VER="输入版本 (留空使用 latest)"
        L_INPUT_URL="输入脚本 URL (留空复制本地"
        L_ENTER_URL="输入脚本 URL [默认:"

        L_ERR_ROOT="需要 root 权限或安装 sudo 后重试"
        L_TRY_DL="尝试下载"
        L_DL_FAIL_API="直接下载失败，尝试通过 API 查找..."
        L_ERR_NO_REL="无法获取 release 信息"
        L_FOUND_ASSET="找到资产"
        L_ERR_NO_ASSET="未能找到合适的资产"
        L_EXTRACTING="正在解压资产到"
        L_ERR_EXTRACT="解压失败"
        L_ERR_NO_EXE="未在压缩包中找到可执行文件"
        L_LINK_CREATED="已创建快捷命令"

        L_SVC_RUNNING="服务已经在运行中"
        L_STARTING="正在启动服务..."
        L_START_SUCCESS="服务已成功启动"
        L_LOG_PREVIEW="实时日志预览"
        L_START_FAIL="启动失败！请检查日志详情"
        L_STOPPING="正在发送停止信号..."
        L_STOP_SUCCESS="服务已停止"
        L_STATUS="状态"
        L_STATUS_RUN="运行中"
        L_STATUS_STOP="已停止"
        L_LOG_RECENT="最近 20 行日志预览"

        L_UN_WARN="准备执行全部卸载！将删除目录、日志及所有配置。"
        L_UN_CONFIRM="确认执行全部卸载吗？"
        L_UN_CANCEL="已取消卸载"
        L_CLEAN_PROC="清理残留进程..."
        L_CLEAN_FILES="移除安装目录与文件..."
        L_UN_DONE="全部卸载完成"

        L_DL_SCRIPT="从指定 URL 下载安装脚本..."
        L_ERR_DL_SCRIPT="下载安装脚本失败"
        L_CP_SCRIPT="复制当前脚本到系统目录..."
        L_ST_DL_SCRIPT="脚本通过 stdin 执行，正在尝试从 GitHub 自动获取..."
        L_INST_DONE="安装脚本已就绪"

        L_PRE_DL="准备下载 fast-note"
        L_INST_ALL_DONE="安装/升级流程已完成"
        L_INST_TIP="提示: 输入 fast-note-installer 或选择菜单启动服务。"
        L_INVALID="无效选项，请重新选择"
        L_USAGE="用法"
        L_CUR_VER="当前版本"
        L_NOT_INSTALLED="未安装"
    else
        L_MENU_1="Install / Update Service"
        L_MENU_1_D="Download service and install manager to system"
        L_MENU_2="Start Service"
        L_MENU_2_D="Start the sync service in background"
        L_MENU_3="Stop Service"
        L_MENU_3_D="Terminate the running service process"
        L_MENU_4="Service Status"
        L_MENU_4_D="Check status and preview recent logs"
        L_MENU_5="Uninstall All"
        L_MENU_5_D="Remove program, config, and all logs"
        L_MENU_6="Install Self to System"
        L_MENU_6_D="Add this tool to global commands (fast-note-installer)"
        L_MENU_0="Quit"
        L_MENU_L="切换至中文 (Switch to Chinese)"
        L_SELECT="Please select"
        L_INPUT_VER="Enter version (leave blank for latest)"
        L_INPUT_URL="Enter script URL (leave blank to copy local"
        L_ENTER_URL="Enter script URL [Default:"

        L_ERR_ROOT="Root privileges or sudo required"
        L_TRY_DL="Trying to download"
        L_DL_FAIL_API="Direct download failed, trying via API..."
        L_ERR_NO_REL="Failed to get release info"
        L_FOUND_ASSET="Asset found"
        L_ERR_NO_ASSET="No suitable asset found"
        L_EXTRACTING="Extracting asset to"
        L_ERR_EXTRACT="Extraction failed"
        L_ERR_NO_EXE="Executable not found in package"
        L_LINK_CREATED="Symbolic link created"

        L_SVC_RUNNING="Service is already running"
        L_STARTING="Starting service..."
        L_START_SUCCESS="Service started successfully"
        L_LOG_PREVIEW="Real-time log preview"
        L_START_FAIL="Start failed! Please check logs"
        L_STOPPING="Sending stop signal..."
        L_STOP_SUCCESS="Service stopped"
        L_STATUS="Status"
        L_STATUS_RUN="Running"
        L_STATUS_STOP="Stopped"
        L_LOG_RECENT="Recent 20 lines of log"

        L_UN_WARN="Preparing full uninstall! All data will be deleted."
        L_UN_CONFIRM="Confirm full uninstall?"
        L_UN_CANCEL="Uninstall cancelled"
        L_CLEAN_PROC="Cleaning up processes..."
        L_CLEAN_FILES="Removing directories and files..."
        L_UN_DONE="Full uninstall completed"

        L_DL_SCRIPT="Downloading script from URL..."
        L_ERR_DL_SCRIPT="Failed to download script"
        L_CP_SCRIPT="Copying current script to system..."
        L_ST_DL_SCRIPT="Running via stdin, trying to fetch from GitHub..."
        L_INST_DONE="Installer is ready"

        L_PRE_DL="Preparing to download fast-note"
        L_INST_ALL_DONE="Install/Update process completed"
        L_INST_TIP="Tip: Type fast-note-installer or use menu to start service."
        L_INVALID="Invalid option, please try again"
        L_USAGE="Usage"
        L_CUR_VER="Installed Version"
        L_NOT_INSTALLED="Not installed"
    fi
}
load_lang "init" # Initial load

# --- Version Tracking ---
load_version() {
    if [ -f "$VERSION_CONF" ]; then
        INSTALLED_VER=$(cat "$VERSION_CONF" 2>/dev/null | tr -d '[:space:]' || echo "")
    else
        INSTALLED_VER=""
    fi
}

save_version() {
    local v="$1"
    if [ -d "$INSTALL_DIR" ]; then
        echo "$v" | $SUDO tee "$VERSION_CONF" >/dev/null 2>&1 || true
    fi
}

ensure_root() {
    if [ "$EUID" -ne 0 ]; then
        if command -v sudo >/dev/null 2>&1; then
            SUDO="sudo"
        else
            error "$L_ERR_ROOT"
            exit 1
        fi
    else
        SUDO=""
    fi
}

detect_os() {
    local os="$(uname -s | tr '[:upper:]' '[:lower:]')"
    case "$os" in
        linux) echo "linux" ;;
        darwin) echo "darwin" ;;
        mingw*|msys*|cygwin*) echo "windows" ;;
        *) echo "$os" ;;
    esac
}

_arch_map() {
    local a="$(uname -m)"
    case "$a" in
        x86_64|amd64) echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        armv7*|armv6*) echo "armv7" ;;
        *) echo "$a" ;;
    esac
}

# try get latest tag from GitHub API; fallback to "latest" string
get_latest_tag() {
    if command -v curl >/dev/null 2>&1; then
        local latest
        latest="$(curl -fsSL "$GITHUB_API/latest" 2>/dev/null | grep -oP '"tag_name"\s*:\s*"\K[^"]+' || true)"
        if [ -n "$latest" ]; then
            echo "$latest"
            return 0
        fi
    fi
    echo "latest"
}

# construct expected asset file name: fast-note-sync-service-<ver>-<os>-<arch>.tar.gz
asset_name_for() {
    local ver="$1" os="$2" arch="$3"
    # strip leading v if present in tag
    local clean_ver="${ver#v}"
    echo "${BIN_BASE}-${clean_ver}-${os}-${arch}.tar.gz"
}

# download using direct URL based on naming convention; if fails, try API lookup
download_release_asset() {
    local ver="$1" os="$2" arch="$3"
    local clean_ver="${ver#v}"
    local asset_name
    asset_name="$(asset_name_for "$ver" "$os" "$arch")"
    local out="$TMPDIR/$asset_name"
    local url="$GITHUB_RAW/${clean_ver}/${asset_name}"

    info "$L_TRY_DL: ${_BOLD}$url${_RESET}" >&2
    if curl -fSL -o "$out" "$url"; then
        echo "$out"
        return 0
    fi

    warn "$L_DL_FAIL_API" >&2

    # try API: find release by tag or latest
    local release_json
    if [ "$ver" = "latest" ] || [ -z "$ver" ]; then
        release_json="$(curl -fsSL "$GITHUB_API/latest" 2>/dev/null || true)"
    else
        # find release by tag
        release_json="$(curl -fsSL "$GITHUB_API/tags/$ver" 2>/dev/null || true)"
        if [ -z "$release_json" ]; then
            release_json="$(curl -fsSL "$GITHUB_API" 2>/dev/null | grep -A20 "\"tag_name\": \"$ver\"" -n || true)"
        fi
    fi

    if [ -z "$release_json" ]; then
        echo "${_RED}$L_ERR_NO_REL${_RESET}" >&2
        return 2
    fi

    # If jq exists use it
    if command -v jq >/dev/null 2>&1; then
        local asset_url
        asset_url="$(echo "$release_json" | jq -r --arg name "$asset_name" '.assets[] | select(.name==$name) | .browser_download_url' 2>/dev/null || true)"
        if [ -z "$asset_url" ]; then
            asset_url="$(echo "$release_json" | jq -r --arg os "$os" --arg arch "$arch" '.assets[] | select(.name|test($os) and .name|test($arch)) | .browser_download_url' 2>/dev/null | head -n1 || true)"
        fi
        if [ -n "$asset_url" ]; then
            info "$L_FOUND_ASSET: ${_BOLD}$asset_url${_RESET}" >&2
            curl -L --fail -o "$out" "$asset_url"
            echo "$out"
            return 0
        fi
    else
        # fallback to grep/sed extraction
        local asset_url
        asset_url="$(echo "$release_json" | grep -oE '"browser_download_url"[[:space:]]*:[[:space:]]*"[^"]+' | sed -E 's/.*:"([^"]+)$/\1/' | grep "$os" | grep "$arch" | head -n1 || true)"
        if [ -n "$asset_url" ]; then
            info "$L_FOUND_ASSET: ${_BOLD}$asset_url${_RESET}" >&2
            curl -L --fail -o "$out" "$asset_url"
            echo "$out"
            return 0
        fi
    fi

    error "$L_ERR_NO_ASSET (os:$os arch:$arch)" >&2
    return 3
}

install_binary_from_tar() {
    local tarball="$1"
    ensure_root
    mkdir -p "$INSTALL_DIR"
    step "$L_EXTRACTING $INSTALL_DIR ..."
    $SUDO tar -xzf "$tarball" -C "$INSTALL_DIR" || { error "$L_ERR_EXTRACT"; return 1; }

    if [ -f "$INSTALL_DIR/$BIN_BASE" ]; then
        $SUDO mv -f "$INSTALL_DIR/$BIN_BASE" "$BIN_PATH" 2>/dev/null || true
        $SUDO chmod +x "$BIN_PATH"
    else
        local exe
        exe="$(find "$INSTALL_DIR" -maxdepth 2 -type f -perm -111 | head -n1 || true)"
        if [ -n "$exe" ]; then
            $SUDO mv -f "$exe" "$BIN_PATH"
            $SUDO chmod +x "$BIN_PATH"
        else
            error "$L_ERR_NO_EXE"
            return 2
        fi
    fi

    $SUDO ln -sf "$BIN_PATH" "$LINK_BIN" || true
    success "$L_LINK_CREATED: ${_BOLD}$LINK_BIN${_RESET}"
}



start_service() {
    ensure_root
    if pgrep -f "$BIN_PATH" >/dev/null 2>&1; then
        warn "$L_SVC_RUNNING (PID: $(pgrep -f "$BIN_PATH"))"
        return
    fi
    step "$L_STARTING"
    # 保持 bash -c 包装以解决重定向权限问题
    $SUDO bash -c "cd $INSTALL_DIR && nohup ./$BIN_BASE run >> $LOG_FILE 2>&1 &"

    sleep 2
    if pgrep -f "$BIN_PATH" >/dev/null 2>&1; then
        success "$L_START_SUCCESS"
        info "$L_LOG_PREVIEW: ${_BOLD}tail -f $LOG_FILE${_RESET}"
    else
        error "$L_START_FAIL: ${_BOLD}sudo tail -n 20 $LOG_FILE${_RESET}"
    fi
}

stop_service() {
    ensure_root
    step "$L_STOPPING"
    $SUDO pkill -f "$BIN_PATH" || true
    success "$L_STOP_SUCCESS"
}

status_service() {
    local pids
    pids="$(pgrep -f "$BIN_PATH" || true)"
    if [ -n "$pids" ]; then
        success "$L_STATUS: ${_BOLD}$L_STATUS_RUN${_RESET} (PID: $pids)"
        echo -e "\n${_BLUE}${_BOLD}$L_LOG_RECENT ($LOG_FILE):${_RESET}"
        echo "${_CYAN}------------------------------------------------------------${_RESET}"
        $SUDO tail -n 20 "$LOG_FILE" 2>/dev/null || true
        echo "${_CYAN}------------------------------------------------------------${_RESET}"
    else
        warn "$L_STATUS: ${_BOLD}$L_STATUS_STOP${_RESET}"
    fi
}


full_uninstall() {
    ensure_root
    warn "$L_UN_WARN"
    read -rp "  $(echo -e "${_BOLD}$L_UN_CONFIRM [y/N]: ${_RESET}")" yn
    yn="${yn:-N}"
    if [[ ! "$yn" =~ ^[Yy]$ ]]; then
        info "$L_UN_CANCEL"
        return 0
    fi

    step "$L_CLEAN_PROC"
    $SUDO pkill -f "$BIN_PATH" || true

    step "$L_CLEAN_FILES"
    $SUDO rm -rf "$INSTALL_DIR" || true
    $SUDO rm -f "$LINK_BIN" || true
    $SUDO rm -f "$INSTALLER_SELF_PATH" || true
    $SUDO rm -f "$INSTALLER_LINK" || true
    $SUDO rm -f "$LOG_FILE" || true

    success "$L_UN_DONE"
}

install_self() {
    ensure_root
    local src_url_hint
    src_url_hint="${1:-}"
    if [ -n "$src_url_hint" ]; then
        step "$L_DL_SCRIPT"
        $SUDO mkdir -p "$(dirname "$INSTALLER_SELF_PATH")"
        $SUDO curl -fsSL "$src_url_hint" -o "$INSTALLER_SELF_PATH" || { error "$L_ERR_DL_SCRIPT"; return 1; }
    else
        if [ -f "$0" ]; then
            step "$L_CP_SCRIPT"
            $SUDO mkdir -p "$(dirname "$INSTALLER_SELF_PATH")"
            $SUDO cp -f "$0" "$INSTALLER_SELF_PATH"
        else
            warn "$L_ST_DL_SCRIPT"
            return 2
        fi
    fi

    $SUDO chmod +x "$INSTALLER_SELF_PATH"
    $SUDO ln -sf "$INSTALLER_SELF_PATH" "$INSTALLER_LINK"
    success "$L_INST_DONE: ${_BOLD}$INSTALLER_LINK${_RESET}"
}

show_menu() {
    draw_banner
    echo -e "  [1] ${_BOLD}$L_MENU_1${_RESET}"
    echo -e "      ${_CYAN}${_ITALIC}${_DIM}$L_MENU_1_D${_RESET}"
    echo -e "  [2] ${_BOLD}$L_MENU_2${_RESET}"
    echo -e "      ${_CYAN}${_ITALIC}${_DIM}$L_MENU_2_D${_RESET}"
    echo -e "  [3] ${_BOLD}$L_MENU_3${_RESET}"
    echo -e "      ${_CYAN}${_ITALIC}${_DIM}$L_MENU_3_D${_RESET}"
    echo -e "  [4] ${_BOLD}$L_MENU_4${_RESET}"
    echo -e "      ${_CYAN}${_ITALIC}${_DIM}$L_MENU_4_D${_RESET}"
    echo -e "  [5] ${_BOLD}$L_MENU_5${_RESET}"
    echo -e "      ${_CYAN}${_ITALIC}${_DIM}$L_MENU_5_D${_RESET}"
    echo -e "  [6] ${_BOLD}$L_MENU_6${_RESET}"
    echo -e "      ${_CYAN}${_ITALIC}${_DIM}$L_MENU_6_D${_RESET}"
    echo -e "  [L] ${_BOLD}$L_MENU_L${_RESET}"
    echo -e "  [0] ${_BOLD}$L_MENU_0${_RESET}"
    echo -e "\n${_BLUE} ================================================ ${_RESET}"

    while true; do
        read -rp "  $(echo -e "${_BOLD}$L_SELECT [0-6, L]: ${_RESET}")" opt
        case "$opt" in
            1) read -rp "  $(echo -e "${_BOLD}$L_INPUT_VER: ${_RESET}")" v; v="${v:-latest}"; install_cmd "$v";;
            2) start_service;;
            3) stop_service;;
            4) status_service;;
            5) full_uninstall;;
            6)
                if [ -f "$0" ]; then
                    read -rp "  $(echo -e "${_BOLD}$L_INPUT_URL $0): ${_RESET}")" raw
                else
                    default_raw="https://raw.githubusercontent.com/haierkeys/fast-note-sync-service/master/scripts/quest_install.sh"
                    read -rp "  $(echo -e "${_BOLD}$L_ENTER_URL $default_raw]: ${_RESET}")" raw
                    raw="${raw:-$default_raw}"
                fi
                install_self "$raw"
                ;;
            L|l)
                if [ "$CURRENT_LANG" = "en" ]; then CURRENT_LANG="zh"; else CURRENT_LANG="en"; fi
                save_lang
                load_lang
                draw_banner
                show_menu
                return
                ;;
            0|q) exit 0;;
            *) warn "$L_INVALID";;
        esac
    done
}

install_cmd() {
    local ver="${1:-latest}"
    local os arch tarball
    os="$(detect_os)"
    arch="$(_arch_map)"
    step "$L_PRE_DL ${_BOLD}$ver${_RESET} ($os/$arch)..."
    if [ "$ver" = "latest" ]; then
        ver="$(get_latest_tag || echo latest)"
    fi
    tarball="$(download_release_asset "$ver" "$os" "$arch")" || { error "$L_ERR_NO_REL"; return 1; }
    install_binary_from_tar "$tarball"
    save_version "$ver"
    install_self >/dev/null 2>&1 || true
    success "$L_INST_ALL_DONE"
    info "$L_INST_TIP"
}

# main dispatcher
cmd="${1:-menu}"
case "$cmd" in
    install)
        ensure_root
        install_cmd "${2:-latest}"
    ;;
    uninstall|full-uninstall|full_uninstall)
        full_uninstall
    ;;
    start)
        start_service
    ;;
    stop)
        stop_service
    ;;
    status)
        status_service
    ;;
    update)
        ensure_root
        stop_service
        install_cmd "latest"
    ;;
    install-self)
        install_self "${2:-}"
    ;;
    menu)
        show_menu
    ;;
    *)
        draw_banner
        echo -e "$L_USAGE: $0 {install|uninstall|full-uninstall|start|stop|status|update|install-self|menu}"
        exit 1
    ;;
esac
