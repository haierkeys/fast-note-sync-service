#!/bin/bash
set -e

# Globals
TEMP="$(mktemp -d)"

function cleanup() {
    rm -rf "$TEMP"
    # Only print new line if we are not returning a value (e.g. for simple commands)
}
trap cleanup EXIT

function on_interrupt() {
    echo -e "\n\nOperation cancelled by user (Ctrl+C). Exiting..."
    exit 130
}
trap on_interrupt INT
ARCH=$(dpkg --print-architecture)
OS="linux"

# Function to fetch latest Go versions
# Usage: fetch_versions <count>
function fetch_versions() {
    local count=$1
    if [ -z "$count" ]; then count=10; fi
    
    # Fetch existing versions from go.dev
    # Using curl and grep to parse JSON mode from go.dev
    curl -s 'https://go.dev/dl/?mode=json' | grep -Eo 'go[0-9]+\.[0-9]+(\.[0-9]+)?' | sort -V -u -r | head -n "$count"
}

function install_go() {
    local target_ver="$1" # Optional version argument
    
    if [ -n "$target_ver" ]; then
        # Direct install mode
        local selected_version="$target_ver"
        if [[ ! "$selected_version" =~ ^go ]]; then
            selected_version="go$selected_version"
        fi
        do_install "$selected_version"
        return
    fi
    
    echo "-------------------------------------"
    echo "          Install Go"
    echo "-------------------------------------"
    
    echo "Fetching latest 10 Go versions..." >&2
    local versions
    versions=$(fetch_versions 10)
    
    if [ -z "$versions" ]; then
        echo "Failed to fetch Go versions."
        return
    fi
    
    echo "Available Versions:"
    local i=1
    declare -A VERSION_MAP
    for v in $versions; do
        echo "$i) $v"
        VERSION_MAP[$i]=$v
        ((i++))
    done
    echo "0) Back to Main Menu"
    
    local choice
    read -p "Enter version number to install (default: 1): " choice
    
    if [ -z "$choice" ]; then choice=1; fi
    
    if [ "$choice" == "0" ]; then
        return
    fi
    
    local selected_version=""
    if [[ "$choice" =~ ^[0-9]+$ ]] && [ "${VERSION_MAP[$choice]+isset}" ]; then
        selected_version="${VERSION_MAP[$choice]}"
    else
        # Allow custom version input
        selected_version="$choice"
        if [[ ! "$selected_version" =~ ^go ]]; then
            selected_version="go$selected_version"
        fi
    fi
    
    do_install "$selected_version"
    
    read -p "Press Enter to continue..."
}

function do_install() {
    local selected_version="$1"
    echo "Selected version: $selected_version"
    
    # Check if this exact version is already installed
    if command -v go &> /dev/null; then
        local current_version
        current_version=$(go version | awk '{print $3}')
        if [ "$current_version" == "$selected_version" ]; then
            # If running interactively (no args passed to script originally), ask.
            # But here we complicate things if direct install.
            # For simplicity: always ask confirm if tty.
            if [ -t 0 ]; then
                read -p "Go version $selected_version is already installed. Overwrite? [y/N]: " confirm
                if [[ ! "$confirm" =~ ^[yY] ]]; then
                    echo "Installation cancelled."
                    return
                fi
            else
                echo "Go version $selected_version is already installed. Overwriting..."
            fi
        fi
    fi
    
    # Download and Install logic here
    local download_url="https://go.dev/dl/${selected_version}.${OS}-${ARCH}.tar.gz"
    
    echo "Downloading $download_url ..."
    if wget --progress=dot:mega "$download_url" -O "$TEMP/go-linux.tar.gz"; then
        echo "Removing old installation..."
        rm -rf /usr/local/go
        
        echo "Extracting..."
        tar -C /usr/local -xzf "$TEMP/go-linux.tar.gz"
        
        setup_env
        
        echo "-------------------------------------"
        echo "Go $selected_version installed successfully."
        /usr/local/go/bin/go version
        echo "-------------------------------------"
    else
        echo "Download failed."
        exit 1
    fi
}

function uninstall_go() {
    echo "-------------------------------------"
    echo "          Uninstall Go"
    echo "-------------------------------------"
    
    if [ ! -d "/usr/local/go" ]; then
        echo "Go does not seem to be installed in /usr/local/go."
        if [ -t 0 ]; then read -p "Press Enter to continue..." ; fi
        return
    fi
    
    if [ -t 0 ]; then
        echo "This will remove:"
        echo "  - /usr/local/go"
        read -p "Are you sure? [y/N]: " confirm
        if [[ ! "$confirm" =~ ^[yY] ]]; then
            echo "Uninstallation cancelled."
            return
        fi
    fi
    
    rm -rf /usr/local/go
    echo "Go uninstalled."
    
    if [ -t 0 ]; then read -p "Press Enter to continue..." ; fi
}

function list_versions() {
    echo "-------------------------------------"
    echo "       Available Go Versions"
    echo "-------------------------------------"
    local versions
    versions=$(fetch_versions 10)
    
    local i=1
    for v in $versions; do
        echo "$i) $v"
        ((i++))
    done
    echo "-------------------------------------"
}
# Separate function for wait to reuse logic
function wait_enter() {
    if [ -t 0 ]; then
        read -p "Press Enter to return to menu..."
    fi
}

function setup_env() {
    mkdir -p /go/bin /go/src /go/pkg
    
    # Define env vars for current session
    export GO_HOME=/usr/local/go
    export GOPATH=/go
    export PATH=${GOPATH}/bin:${GO_HOME}/bin/:$PATH
}

function show_menu() {
    while true; do
        clear
        echo "====================================="
        echo "       Go Manager (go_install.sh)"
        echo "====================================="
        echo "1. Install Go"
        echo "2. Uninstall Go"
        echo "3. List Available Versions (Top 10)"
        echo "0. Exit"
        echo "====================================="
        read -p "Enter your choice: " main_choice
        
        case $main_choice in
            1) install_go ;;
            2) uninstall_go ;;
            3) list_versions; wait_enter ;;
            0) echo "Goodbye!"; exit 0 ;;
            *) echo "Invalid choice. Please try again."; sleep 1 ;;
        esac
    done
}

# Main Dispatcher
CMD="${1:-menu}"

case "$CMD" in
    install)
        # Shift to get potential version argument
        shift
        install_go "$1"
    ;;
    uninstall)
        uninstall_go
    ;;
    list)
        list_versions
    ;;
    menu)
        show_menu
    ;;
    *)
        # If the first argument doesn't match above commands, default to menu
        # This allows simple `bash <(curl...)` to work
        show_menu
    ;;
esac