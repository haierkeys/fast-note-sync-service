#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

# ===========================================
# fast-note 一键安装/管理脚本（增强版）
# - 自动安装自身到 /usr/local/bin/fast-note-installer
# - 自动下载对应平台的 release 资产并安装
# - 支持全部卸载 (full-uninstall)
# - 支持 systemd 或 nohup 后台运行
# 用法:
#   bash <(curl -fsSL <raw-url>)           # 交互菜单
#   bash <(curl -fsSL <raw-url>) install   # 安装最新 release
#   bash <(curl -fsSL <raw-url>) install 0.8.6  # 安装指定版本
#   bash <(curl -fsSL <raw-url>) install-self   # 将安装脚本本身复制到 /usr/local/bin
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

# colors
_RED="$(tput setaf 1 2>/dev/null || echo '')"
_GREEN="$(tput setaf 2 2>/dev/null || echo '')"
_YELLOW="$(tput setaf 3 2>/dev/null || echo '')"
_BLUE="$(tput setaf 4 2>/dev/null || echo '')"
_RESET="$(tput sgr0 2>/dev/null || echo '')"

ensure_root() {
    if [ "$EUID" -ne 0 ]; then
        if command -v sudo >/dev/null 2>&1; then
            SUDO="sudo"
        else
            echo "${_RED}需要 root 权限或安装 sudo 后重试${_RESET}" >&2
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
    
    echo "${_BLUE}尝试下载: $url${_RESET}"
    if curl -fSL -o "$out" "$url"; then
        echo "$out"
        return 0
    fi
    
    echo "${_YELLOW}直接构造 URL 下载失败，尝试通过 GitHub API 查找资产...${_RESET}"
    
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
        echo "${_RED}无法获取 release 信息，请检查网络或仓库是否存在 release。${_RESET}" >&2
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
            echo "${_BLUE}找到资产: $asset_url${_RESET}"
            curl -L --fail -o "$out" "$asset_url"
            echo "$out"
            return 0
        fi
    else
        # fallback to grep/sed extraction
        local asset_url
        asset_url="$(echo "$release_json" | grep -oE '"browser_download_url"[[:space:]]*:[[:space:]]*"[^"]+' | sed -E 's/.*:"([^"]+)$/\1/' | grep "$os" | grep "$arch" | head -n1 || true)"
        if [ -n "$asset_url" ]; then
            echo "${_BLUE}找到资产: $asset_url${_RESET}"
            curl -L --fail -o "$out" "$asset_url"
            echo "$out"
            return 0
        fi
    fi
    
    echo "${_RED}未能找到合适的资产 (os:$os arch:$arch)。请检查 Release 命名规则是否为: ${BIN_BASE}-<version>-<os>-<arch>.tar.gz${_RESET}" >&2
    return 3
}

install_binary_from_tar() {
    local tarball="$1"
    ensure_root
    mkdir -p "$INSTALL_DIR"
    echo "${_BLUE}解压 $tarball 到 $INSTALL_DIR ...${_RESET}"
    tar -xzf "$tarball" -C "$INSTALL_DIR" || { echo "${_RED}解压失败${_RESET}" >&2; return 1; }
    
    # 如果解压后文件名已为二进制名，直接使用；否则尝试找到第一个可执行文件并重命名
    if [ -f "$INSTALL_DIR/$BIN_BASE" ]; then
        mv -f "$INSTALL_DIR/$BIN_BASE" "$BIN_PATH" 2>/dev/null || true
        chmod +x "$BIN_PATH"
    else
        local exe
        exe="$(find "$INSTALL_DIR" -maxdepth 2 -type f -perm -111 | head -n1 || true)"
        if [ -n "$exe" ]; then
            mv -f "$exe" "$BIN_PATH"
            chmod +x "$BIN_PATH"
        else
            echo "${_RED}未在压缩包中找到可执行文件。请确认 tar 包根目录包含可执行二进制。${_RESET}" >&2
            return 2
        fi
    fi
    
    # create symlink
    $SUDO ln -sf "$BIN_PATH" "$LINK_BIN" || true
    echo "${_GREEN}已在 $LINK_BIN 创建快捷命令 (指向 $BIN_PATH)${_RESET}"
}

create_systemd_service() {
    ensure_root
    if command -v systemctl >/dev/null 2>&1; then
        echo "${_BLUE}创建 systemd 服务 /etc/systemd/system/$SERVICE_NAME${_RESET}"
    cat <<EOF | $SUDO tee /etc/systemd/system/$SERVICE_NAME >/dev/null
[Unit]
Description=fast-note sync service
After=network.target

[Service]
Type=simple
ExecStart=$BIN_PATH
Restart=on-failure
RestartSec=5
StandardOutput=append:$LOG_FILE
StandardError=append:$LOG_FILE
User=root
Environment=HOME=/root

[Install]
WantedBy=multi-user.target
EOF
        $SUDO systemctl daemon-reload || true
        $SUDO systemctl enable --now "$SERVICE_NAME" || true
        echo "${_GREEN}systemd 服务已启用并尝试启动。${_RESET}"
    else
        echo "${_YELLOW}未检测到 systemd，跳过服务创建。将采用 nohup 启动。${_RESET}"
    fi
}

start_service() {
    ensure_root
    if command -v systemctl >/dev/null 2>&1; then
        $SUDO systemctl start "$SERVICE_NAME" || true
    else
        $SUDO bash -c "nohup $BIN_PATH >> $LOG_FILE 2>&1 &"
    fi
    echo "${_GREEN}启动命令已执行（视 systemd 状态而定）。${_RESET}"
}

stop_service() {
    ensure_root
    if command -v systemctl >/dev/null 2>&1; then
        $SUDO systemctl stop "$SERVICE_NAME" || true
    else
        pkill -f "$BIN_PATH" || true
    fi
    echo "${_GREEN}停止命令已执行。${_RESET}"
}

status_service() {
    if command -v systemctl >/dev/null 2>&1; then
        systemctl status "$SERVICE_NAME" --no-pager || true
    else
        local pids
        pids="$(pgrep -f "$BIN_PATH" || true)"
        if [ -n "$pids" ]; then
            echo "${_GREEN}fast-note 运行中，PID: $pids${_RESET}"
        else
            echo "${_YELLOW}fast-note 未运行${_RESET}"
        fi
    fi
}

uninstall_service() {
    ensure_root
    echo "${_BLUE}停止服务并移除二进制和链接...${_RESET}"
    if command -v systemctl >/dev/null 2>&1; then
        $SUDO systemctl stop "$SERVICE_NAME" || true
        $SUDO systemctl disable "$SERVICE_NAME" || true
        $SUDO rm -f /etc/systemd/system/"$SERVICE_NAME" || true
        $SUDO systemctl daemon-reload || true
    fi
    $SUDO rm -f "$LINK_BIN" || true
    $SUDO rm -f "$BIN_PATH" || true
    echo "${_GREEN}已卸载服务与二进制（保留配置与日志）。${_RESET}"
}

full_uninstall() {
    ensure_root
    echo "${_RED}执行全部卸载，将删除安装目录、日志、安装脚本以及软链。${_RESET}"
    read -rp "确认全部卸载并删除所有文件吗？这将删除 $INSTALL_DIR, $LOG_FILE, $INSTALLER_SELF_PATH, $INSTALLER_LINK (y/N): " yn
    yn="${yn:-N}"
    if [[ ! "$yn" =~ ^[Yy]$ ]]; then
        echo "已取消全部卸载。"
        return 0
    fi
    
    # stop and remove systemd
    if command -v systemctl >/dev/null 2>&1; then
        $SUDO systemctl stop "$SERVICE_NAME" || true
        $SUDO systemctl disable "$SERVICE_NAME" || true
        $SUDO rm -f /etc/systemd/system/"$SERVICE_NAME" || true
        $SUDO systemctl daemon-reload || true
    fi
    
    $SUDO rm -rf "$INSTALL_DIR" || true
    $SUDO rm -f "$LINK_BIN" || true
    $SUDO rm -f "$INSTALLER_SELF_PATH" || true
    $SUDO rm -f "$INSTALLER_LINK" || true
    $SUDO rm -f "$LOG_FILE" || true
    
    echo "${_GREEN}全部卸载完成。已删除: $INSTALL_DIR, $LOG_FILE, 安装脚本与软链。${_RESET}"
}

install_self() {
    ensure_root
    # 如果当前脚本是通过 stdin 执行（bash <(curl ...))，则 $0 不是脚本路径；需要重新下载 raw 并写入
    local src_url_hint
    src_url_hint="${1:-}"
    if [ -n "$src_url_hint" ]; then
        echo "${_BLUE}从指定 URL 下载安装脚本: $src_url_hint${_RESET}"
        $SUDO mkdir -p "$(dirname "$INSTALLER_SELF_PATH")"
        $SUDO curl -fsSL "$src_url_hint" -o "$INSTALLER_SELF_PATH" || { echo "${_RED}下载安装脚本失败${_RESET}"; return 1; }
    else
        # 尝试从 /proc/self/fd/0 或 $0 内容写入
        if [ -f "$0" ]; then
            echo "${_BLUE}复制当前脚本 $0 到 $INSTALLER_SELF_PATH${_RESET}"
            $SUDO mkdir -p "$(dirname "$INSTALLER_SELF_PATH")"
            $SUDO cp -f "$0" "$INSTALLER_SELF_PATH"
        else
            echo "${_YELLOW}脚本可能是通过 stdin 执行，尝试从 GitHub raw URL 自动下载（需要网络）。${_RESET}"
            # 构造 raw url based on common path; 用户应传入 src_url_hint 更可靠
            echo "${_YELLOW}请使用参数 install-self <raw-url> 以指定 raw 文件地址（推荐）。${_RESET}"
            return 2
        fi
    fi
    
    $SUDO chmod +x "$INSTALLER_SELF_PATH"
    $SUDO ln -sf "$INSTALLER_SELF_PATH" "$INSTALLER_LINK"
    echo "${_GREEN}安装脚本已复制到 $INSTALLER_SELF_PATH 并创建软链 $INSTALLER_LINK${_RESET}"
}

show_menu() {
  cat <<EOF
${_BLUE}============================================${_RESET}
${_GREEN} fast-note 管理脚本（增强版）${_RESET}
  1) 安装/更新 (install)
  2) 卸载 (uninstall)
  3) 全部卸载 (full-uninstall)
  4) 启动 (start)
  5) 停止 (stop)
  6) 状态 (status)
  7) 安装脚本到系统 (install-self)
  8) 退出 (quit)
${_BLUE}============================================${_RESET}
EOF
    
    while true; do
        read -rp "请选择 [1-8]: " opt
        case "$opt" in
            1) read -rp "输入版本（留空使用 latest）： " v; v="${v:-latest}"; install_cmd "$v";;
            2) uninstall_service;;
            3) full_uninstall;;
            4) start_service;;
            5) stop_service;;
            6) status_service;;
            7) read -rp "输入安装脚本 raw URL（可留空，若留空需在可访问文件系统的场景下执行）: " raw; install_self "$raw";;
            8|q) exit 0;;
            *) echo "无效选项";;
        esac
    done
}

install_cmd() {
    local ver="${1:-latest}"
    local os arch tarball
    os="$(detect_os)"
    arch="$(_arch_map)"
    echo "${_BLUE}准备安装 fast-note 版本: $ver (OS:$os ARCH:$arch)${_RESET}"
    if [ "$ver" = "latest" ]; then
        ver="$(get_latest_tag || echo latest)"
    fi
    tarball="$(download_release_asset "$ver" "$os" "$arch")" || { echo "${_RED}下载失败${_RESET}"; return 1; }
    install_binary_from_tar "$tarball"
    create_systemd_service
    echo "${_GREEN}安装完成。可使用: sudo $LINK_BIN (或 systemctl start $SERVICE_NAME)${_RESET}"
}

# main dispatcher
cmd="${1:-menu}"
case "$cmd" in
    install)
        ensure_root
        install_cmd "${2:-latest}"
    ;;
    uninstall)
        uninstall_service
    ;;
    full-uninstall|full_uninstall)
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
        uninstall_service
        install_cmd "latest"
    ;;
    install-self)
        # 若通过 stdin 执行并带有第二参数 raw URL，则会下载到 INSTALLER_SELF_PATH
        install_self "${2:-}"
    ;;
    menu)
        show_menu
    ;;
    *)
        echo "未知命令 $cmd"
        echo "用法: install|uninstall|full-uninstall|start|stop|status|update|install-self|menu"
        exit 1
    ;;
esac
