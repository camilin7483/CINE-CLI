#!/usr/bin/env bash
set -euo pipefail

# ──────────────────────────────────────────────
#  cine-cli — Installer
#  Automated setup: deps → build → configure
# ──────────────────────────────────────────────

BINARY="cine"
BIN_DIR="${HOME}/.local/bin"
CONFIG_DIR="${HOME}/.config/cine-cli"
DATA_DIR="${HOME}/.local/share/cine-cli"
REPO="camilin7483/CINE-CLI"
GO_VERSION_MIN="1.24"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; NC='\033[0m'
info()  { echo -e "${CYAN}::${NC} $1"; }
ok()    { echo -e "${GREEN}✓${NC} $1"; }
warn()  { echo -e "${YELLOW}⚠${NC} $1"; }
err()   { echo -e "${RED}✗${NC} $1"; }

cleanup() { rm -rf /tmp/cine-cli-install; }
trap cleanup EXIT

# ── OS Detection ──────────────────────────────
detect_os() {
  case "$(uname -s)" in
    Linux)
      if   grep -qi 'arch' /etc/os-release 2>/dev/null; then echo "arch"
      elif grep -qi 'debian\|ubuntu' /etc/os-release 2>/dev/null; then echo "debian"
      elif grep -qi 'fedora' /etc/os-release 2>/dev/null; then echo "fedora"
      elif grep -qi 'opensuse\|suse' /etc/os-release 2>/dev/null; then echo "suse"
      else echo "linux"; fi ;;
    Darwin) echo "macos" ;;
    *)      echo "unknown" ;;
  esac
}

# ── Deps Install ──────────────────────────────
install_deps() {
  local os=$1
  info "Installing system dependencies..."

  case "$os" in
    arch)
      sudo pacman -S --noconfirm --needed go mpv yt-dlp 2>/dev/null ||
        warn "Could not install deps, install manually: sudo pacman -S go mpv yt-dlp"
      ;;
    debian)
      sudo apt-get update -qq &&
      sudo apt-get install -y -qq golang-go mpv yt-dlp 2>/dev/null ||
        sudo apt-get install -y -qq golang-go mpv 2>/dev/null ||
        warn "Install manually: sudo apt install golang-go mpv; pip3 install yt-dlp"
      ;;
    fedora)
      sudo dnf install -y golang mpv yt-dlp 2>/dev/null ||
        warn "Install manually: sudo dnf install golang mpv yt-dlp"
      ;;
    macos)
      if ! command -v brew &>/dev/null; then
        warn "Homebrew not found. Install from https://brew.sh"
        return
      fi
      brew install go mpv yt-dlp 2>/dev/null ||
        warn "brew install failed, install manually: brew install go mpv yt-dlp"
      ;;
    *)
      warn "Unknown OS. Install deps manually: go, mpv, yt-dlp"
      ;;
  esac
}

# ── Go Version Check ──────────────────────────
check_go() {
  if ! command -v go &>/dev/null; then
    err "Go not found! Install Go $GO_VERSION_MIN+ first."
    return 1
  fi
  local ver; ver=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+')
  if awk "BEGIN {exit !($ver < $GO_VERSION_MIN)}"; then
    err "Go $ver detected, need $GO_VERSION_MIN+. Upgrade Go."
    return 1
  fi
  ok "Go $ver found"
}

# ── MPV Check ─────────────────────────────────
check_mpv() {
  if command -v mpv &>/dev/null; then
    ok "mpv found: $(mpv --version 2>&1 | head -1)"
  else
    warn "mpv not found — required for playback"
  fi
}

# ── TMDB Key Prompt ───────────────────────────
setup_tmdb() {
  local cfg_file="${CONFIG_DIR}/config.yaml"
  if [ -f "$cfg_file" ] && grep -q 'tmdb_api_key' "$cfg_file" 2>/dev/null &&
     ! grep -q 'tmdb_api_key: ""' "$cfg_file" 2>/dev/null; then
    ok "TMDB API key already configured"
    return
  fi

  echo ""
  info "TMDB API key required (free). Get yours at:"
  info "  https://www.themoviedb.org/settings/api"
  echo ""
  read -rp "  Paste your TMDB API key (or press Enter to skip): " key
  echo ""

  mkdir -p "$CONFIG_DIR"
  if [ -n "$key" ]; then
    cat > "$cfg_file" << CFGEOF
provider: vidsrc
player: mpv
quality: ""
language: en-US
tmdb_api_key: "${key}"
data_dir: ${DATA_DIR}
mpv_args:
  - "--hwdec=auto"
  - "--volume=70"
cache_ttl: 3600
theme: auto
max_results: 50
CFGEOF
    ok "Config written to ${cfg_file}"
  else
    warn "Skipping TMDB config. Edit ${CONFIG_DIR}/config.yaml later"
  fi
}

# ── Build ─────────────────────────────────────
build_cine() {
  local src="$1"
  info "Building cine-cli (this may take a moment)..."

  local out="/tmp/cine-cli-install/${BINARY}"
  mkdir -p /tmp/cine-cli-install

  if ! (cd "$src" && go build -ldflags="-s -w" -o "$out" ./cmd/cine 2>&1); then
    err "Build failed!"
    return 1
  fi

  local size; size=$(du -h "$out" | cut -f1)
  ok "Built ${BINARY} (${size})"
  echo "$out"
}

# ── Install ───────────────────────────────────
install_binary() {
  local src_bin="$1"
  mkdir -p "$BIN_DIR"
  cp "$src_bin" "${BIN_DIR}/${BINARY}"
  chmod +x "${BIN_DIR}/${BINARY}"
  ok "Installed to ${BIN_DIR}/${BINARY}"
}

# ── Shell Completion ──────────────────────────
setup_completion() {
  local bin="${BIN_DIR}/${BINARY}"
  if ! command -v "$bin" &>/dev/null; then
    warn "Binary not in PATH, skipping completion"
    return
  fi

  local shell_type
  shell_type=$(basename "$SHELL")

  case "$shell_type" in
    zsh)
      mkdir -p "${HOME}/.zfunc"
      "$bin" completion zsh > "${HOME}/.zfunc/_cine" 2>/dev/null
      if ! grep -q 'fpath+=.*\.zfunc' "${HOME}/.zshrc" 2>/dev/null; then
        echo 'fpath+=("${HOME}/.zfunc")' >> "${HOME}/.zshrc"
        echo 'autoload -Uz compinit && compinit' >> "${HOME}/.zshrc"
      fi
      ok "Zsh completion installed"
      ;;
    bash)
      "$bin" completion bash > /tmp/cine-completion.bash 2>/dev/null
      local bashrc="${HOME}/.bashrc"
      if [ -f "$bashrc" ] && ! grep -q 'cine-completion' "$bashrc" 2>/dev/null; then
        echo "source /tmp/cine-completion.bash" >> "$bashrc"
      fi
      ok "Bash completion installed"
      ;;
    fish)
      "$bin" completion fish > "${HOME}/.config/fish/completions/cine.fish" 2>/dev/null
      ok "Fish completion installed"
      ;;
  esac
}

# ── PATH Check ────────────────────────────────
check_path() {
  if [[ ":$PATH:" != *":${BIN_DIR}:"* ]]; then
    warn "${BIN_DIR} is not in PATH"
    info "  Add this to your shell config:"
    info "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc"
  fi
}

# ── Data Dirs ─────────────────────────────────
setup_data_dirs() {
  mkdir -p "${DATA_DIR}/cache"
  mkdir -p "${HOME}/Downloads/cine-cli"
  mkdir -p "${CONFIG_DIR}/plugins"
  ok "Data directories created"
}

# ── Verify ────────────────────────────────────
verify() {
  if command -v "${BIN_DIR}/${BINARY}" &>/dev/null; then
    ok "cine-cli installed successfully!"
    echo ""
    echo -e "  ${CYAN}Usage:${NC}"
    echo -e "    ${YELLOW}cine${NC}              Launch the TUI"
    echo -e "    ${YELLOW}cine watch <title>${NC} Search and play"
    echo -e "    ${YELLOW}cine trending${NC}     Browse trending"
    echo -e "    ${YELLOW}cine --help${NC}       Show all commands"
    echo ""
  else
    err "Installation failed — ${BINARY} not found in PATH"
    return 1
  fi
}

# ══════════════════════════════════════════════
#  MAIN
# ══════════════════════════════════════════════

echo ""
echo -e "${CYAN}╔══════════════════════════════════╗${NC}"
echo -e "${CYAN}║       cine-cli  Installer        ║${NC}"
echo -e "${CYAN}║  github.com/${REPO}               ║${NC}"
echo -e "${CYAN}╚══════════════════════════════════╝${NC}"
echo ""

OS=$(detect_os)
info "Detected OS: ${OS}"

# ── Flags ─────────────────────────────────────
SKIP_DEPS=false
SOURCE_DIR=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --skip-deps) SKIP_DEPS=true ;;
    --from-source) SOURCE_DIR="$2"; shift ;;
    --help)
      echo "Usage: ./install.sh [--skip-deps] [--from-source <path>]"
      echo ""
      echo "  --skip-deps         Skip system dependency installation"
      echo "  --from-source <dir> Build from local source instead of cloning"
      echo "  --help              Show this help"
      exit 0
      ;;
  esac
  shift
done

# 1. System deps
if [ "$SKIP_DEPS" = false ]; then
  install_deps "$OS"
else
  info "Skipping dependency installation (--skip-deps)"
fi

# 2. Check Go + mpv
check_go || exit 1
check_mpv

# 3. Get source
if [ -n "$SOURCE_DIR" ]; then
  SRC="$SOURCE_DIR"
  ok "Using local source: ${SRC}"
elif [ -d "$(dirname "$0")/cmd/cine" ]; then
  SRC="$(cd "$(dirname "$0")" && pwd)"
  ok "Using script directory source: ${SRC}"
else
  info "Cloning repository..."
  SRC="/tmp/cine-cli-install/src"
  git clone --depth=1 "https://github.com/${REPO}.git" "$SRC" 2>/dev/null ||
    { err "Failed to clone repo"; exit 1; }
  ok "Repository cloned"
fi

# 4. Build
BUILT=$(build_cine "$SRC") || exit 1

# 5. Install binary
install_binary "$BUILT"

# 6. Create data directories
setup_data_dirs

# 7. TMDB config
setup_tmdb

# 8. Shell completion (optional)
echo ""
info "Shell completion (optional)?"
read -rp "  Install shell completions? [y/N]: " comp
if [[ "$comp" =~ ^[Yy] ]]; then
  setup_completion
fi

# 9. PATH check
check_path

# 10. Verify
verify
