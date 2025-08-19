#!/bin/bash
#
# A script to install the latest release of dbdump from GitHub.
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/manydesigns/dbdump/main/install.sh | bash
#
# This script will:
# 1. Detect the user's OS and architecture.
# 2. Fetch the latest release from GitHub.
# 3. Download the correct release asset.
# 4. Unpack the binary and move it to /usr/local/bin.
# 5. Make the binary executable.

set -e

# --- Configuration ---
GITHUB_REPO="ManyDesigns/dbdump"
BINARY_NAME="dbdump"
INSTALL_DIR="/usr/local/bin"

# --- Helper Functions ---

# Function to print informational messages.
msg() {
  echo -e "\033[32mINFO:\033[0m $1"
}

# Function to print error messages and exit.
err() {
  echo -e "\033[31mERROR:\033[0m $1" >&2
  exit 1
}

# Check for required tools before starting.
check_dependencies() {
  for cmd in curl tar gzip; do
    if ! command -v "$cmd" &>/dev/null; then
      err "'$cmd' is not installed, but is required. Please install it and try again."
    fi
  done
}

# Detect the operating system and architecture.
detect_os_and_arch() {
  OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
  ARCH="$(uname -m)"

  case "$OS" in
    linux) OS="linux" ;;
    darwin) OS="darwin" ;;
    *) err "Unsupported operating system: $OS" ;;
  esac

  case "$ARCH" in
    x86_64 | amd64) ARCH="amd64" ;;
    aarch64 | arm64) ARCH="arm64" ;;
    *) err "Unsupported architecture: $ARCH" ;;
  esac
}

# Fetch the latest release version from the GitHub API.
get_latest_release_version() {
  msg "Fetching the latest release version..."
  local api_url="https://api.github.com/repos/${GITHUB_REPO}/releases/latest"

  # Use curl with grep and sed to extract the tag name, avoiding a dependency on jq.
  VERSION=$(curl -s "$api_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

  if [ -z "$VERSION" ]; then
    err "Could not fetch the latest release version. Please check that the GITHUB_REPO variable is set correctly."
  fi
  msg "The latest version is $VERSION"
}

# Download and install the binary.
download_and_install() {
  local asset_filename="${BINARY_NAME}-${OS}-${ARCH}.tar.gz"
  local download_url="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${asset_filename}"
  local tmp_dir=$(mktemp -d)

  # Ensure the temporary directory is cleaned up on exit.
  trap 'rm -rf "$tmp_dir"' EXIT

  msg "Downloading from $download_url"
  if ! curl -L "$download_url" -o "${tmp_dir}/${asset_filename}"; then
    err "Failed to download the release asset. Please check the URL and your network connection."
  fi

  msg "Extracting the binary..."
  tar -xzf "${tmp_dir}/${asset_filename}" -C "$tmp_dir"

  msg "Installing '${BINARY_NAME}' to '${INSTALL_DIR}'..."
  if [ -w "$INSTALL_DIR" ]; then
    mv "${tmp_dir}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
  else
    msg "Write permissions are required for ${INSTALL_DIR}. Using sudo..."
    sudo mv "${tmp_dir}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
  fi

  msg "'${BINARY_NAME}' has been installed successfully!"
  msg "You can now run '${BINARY_NAME}' from your terminal."
}

# --- Main Logic ---

main() {
  if [ "$GITHUB_REPO" == "YOUR_USER/YOUR_REPO" ]; then
    err "Please edit the script and set the GITHUB_REPO variable to your repository."
  fi

  check_dependencies
  detect_os_and_arch
  get_latest_release_version
  download_and_install
}

# --- Run the Script ---
main
