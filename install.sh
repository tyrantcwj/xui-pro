#!/usr/bin/env bash
set -euo pipefail

REPO="${XUI_PRO_REPO:-tyrantcwj/xui-pro}"
INSTALL_DIR="${XUI_PRO_INSTALL_DIR:-/usr/local/xui-pro}"
BIN_DIR="${XUI_PRO_BIN_DIR:-/usr/local/bin}"
CONFIG_DIR="${XUI_PRO_CONFIG_DIR:-/etc/xui-pro}"
SERVICE_DIR="/etc/systemd/system"
VERSION="${XUI_PRO_VERSION:-latest}"
GO_VERSION="${XUI_PRO_GO_VERSION:-1.22.12}"
FORCE_SOURCE="${XUI_PRO_SOURCE:-0}"
MODE="${1:-master}"

usage() {
  cat <<'EOF'
Usage:
  install.sh master [--listen :8080]
  install.sh agent --master http://xui.ityc.cc:8080 [--token TOKEN] [--node-id NODE_ID] [--country COUNTRY]

Examples:
  bash <(curl -Ls https://raw.githubusercontent.com/tyrantcwj/xui-pro/main/install.sh) master
  XUI_PRO_SOURCE=1 bash <(curl -Ls https://raw.githubusercontent.com/tyrantcwj/xui-pro/main/install.sh) master
  bash <(curl -Ls https://raw.githubusercontent.com/tyrantcwj/xui-pro/main/install.sh) agent --master http://xui.ityc.cc:8080 --token xxx --country China

Environment:
  XUI_PRO_VERSION=latest|v0.1.0
  XUI_PRO_REPO=tyrantcwj/xui-pro
  XUI_PRO_INSTALL_DIR=/usr/local/xui-pro
  XUI_PRO_GO_VERSION=1.22.12
  XUI_PRO_SOURCE=1
EOF
}

need_root() {
  if [ "$(id -u)" -ne 0 ]; then
    echo "Please run as root." >&2
    exit 1
  fi
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *) echo "Unsupported architecture: $(uname -m)" >&2; exit 1 ;;
  esac
}

download_url() {
  local arch="$1"
  if [ "$VERSION" = "latest" ]; then
    echo "https://github.com/${REPO}/releases/latest/download/xui-pro-linux-${arch}.tar.gz"
  else
    echo "https://github.com/${REPO}/releases/download/${VERSION}/xui-pro-linux-${arch}.tar.gz"
  fi
}

source_url() {
  echo "https://codeload.github.com/${REPO}/tar.gz/refs/heads/main"
}

parse_args() {
  XUI_LISTEN="${XUI_LISTEN:-:8080}"
  XUI_MASTER="${XUI_MASTER:-http://xui.ityc.cc:8080}"
  XUI_AGENT_TOKEN="${XUI_AGENT_TOKEN:-}"
  XUI_NODE_ID="${XUI_NODE_ID:-}"
  XUI_NODE_NAME="${XUI_NODE_NAME:-}"
  XUI_NODE_COUNTRY="${XUI_NODE_COUNTRY:-${XUI_NODE_REGION:-unknown}}"
  XUI_NODE_REGION="${XUI_NODE_REGION:-unknown}"
  XUI_NODE_ENDPOINT="${XUI_NODE_ENDPOINT:-}"
  XUI_SSH_USER="${XUI_SSH_USER:-root}"

  shift || true
  while [ "$#" -gt 0 ]; do
    case "$1" in
      --listen) XUI_LISTEN="$2"; shift 2 ;;
      --master) XUI_MASTER="$2"; shift 2 ;;
      --token) XUI_AGENT_TOKEN="$2"; shift 2 ;;
      --node-id) XUI_NODE_ID="$2"; shift 2 ;;
      --node-name) XUI_NODE_NAME="$2"; shift 2 ;;
      --country) XUI_NODE_COUNTRY="$2"; XUI_NODE_REGION="$2"; shift 2 ;;
      --region) XUI_NODE_COUNTRY="$2"; XUI_NODE_REGION="$2"; shift 2 ;;
      --endpoint) XUI_NODE_ENDPOINT="$2"; shift 2 ;;
      --ssh-user) XUI_SSH_USER="$2"; shift 2 ;;
      -h|--help) usage; exit 0 ;;
      *) echo "Unknown argument: $1" >&2; usage; exit 1 ;;
    esac
  done
}

install_files() {
  local arch="$1"
  local tmp
  tmp="$(mktemp -d)"
  trap "rm -rf '$tmp'" EXIT

  if [ "$FORCE_SOURCE" = "1" ]; then
    echo "Source build was requested. Building xui-pro from main..."
    build_from_source "$arch" "$tmp"
    return
  fi

  echo "Downloading xui-pro ${VERSION} for linux-${arch}..."
  if ! curl -fL "$(download_url "$arch")" -o "$tmp/xui-pro.tar.gz"; then
    echo "Release asset was not found. Falling back to source build..."
    build_from_source "$arch" "$tmp"
    return
  fi
  install_package "$tmp"
}

install_package() {
  local tmp="$1"
  mkdir -p "$INSTALL_DIR" "$CONFIG_DIR"
  tar -xzf "$tmp/xui-pro.tar.gz" -C "$tmp"
  cp -f "$tmp/xuid" "$INSTALL_DIR/xuid"
  cp -f "$tmp/xui-agent" "$INSTALL_DIR/xui-agent"
  cp -f "$tmp/xui-pro" "$BIN_DIR/xui-pro"
  chmod +x "$INSTALL_DIR/xuid" "$INSTALL_DIR/xui-agent" "$BIN_DIR/xui-pro"

  if [ -d "$tmp/reality" ]; then
    cp -R "$tmp/reality" "$INSTALL_DIR/"
  fi
}

ensure_go() {
  local arch="$1"
  local need_install=0
  if command -v go >/dev/null 2>&1; then
    local minor
    minor="$(go version | sed -n 's/.*go[0-9][0-9]*\.\([0-9][0-9]*\).*/\1/p' | head -n1)"
    if [ -z "$minor" ] || [ "$minor" -lt 22 ]; then
      need_install=1
    fi
  else
    need_install=1
  fi

  if [ "$need_install" -eq 0 ]; then
    return
  fi

  local toolchain_dir="/usr/local/xui-pro-toolchain"
  local go_tar="/tmp/go${GO_VERSION}.linux-${arch}.tar.gz"
  echo "Installing portable Go ${GO_VERSION} for source build..."
  mkdir -p "$toolchain_dir"
  curl -fL "https://go.dev/dl/go${GO_VERSION}.linux-${arch}.tar.gz" -o "$go_tar"
  rm -rf "$toolchain_dir/go"
  tar -xzf "$go_tar" -C "$toolchain_dir"
  export PATH="$toolchain_dir/go/bin:$PATH"
}

build_from_source() {
  local arch="$1"
  local tmp="$2"
  local src="$tmp/src"
  mkdir -p "$src" "$tmp/package"

  curl -fL "$(source_url)" -o "$tmp/source.tar.gz"
  tar -xzf "$tmp/source.tar.gz" -C "$src" --strip-components=1
  ensure_go "$arch"

  echo "Building xui-pro from source..."
  (cd "$src" && CGO_ENABLED=0 GOOS=linux GOARCH="$arch" go build -trimpath -ldflags "-s -w" -o "$tmp/package/xuid" ./cmd/xuid)
  (cd "$src" && CGO_ENABLED=0 GOOS=linux GOARCH="$arch" go build -trimpath -ldflags "-s -w" -o "$tmp/package/xui-agent" ./cmd/xui-agent)
  cp "$src/scripts/xui-pro.sh" "$tmp/package/xui-pro"
  chmod +x "$tmp/package/xuid" "$tmp/package/xui-agent" "$tmp/package/xui-pro"
  cp -R "$src/reality" "$tmp/package/reality"
  (cd "$tmp/package" && tar -czf "$tmp/xui-pro.tar.gz" .)
  install_package "$tmp"
}

write_master_env() {
  cat > "$CONFIG_DIR/master.env" <<EOF
XUI_LISTEN=${XUI_LISTEN}
XUI_REALITY_LIBRARY=${INSTALL_DIR}/reality/domains.json
EOF
}

write_agent_env() {
  if [ -z "$XUI_NODE_ID" ]; then
    XUI_NODE_ID="$(hostname)"
  fi
  if [ -z "$XUI_NODE_NAME" ]; then
    XUI_NODE_NAME="$XUI_NODE_ID"
  fi
  if [ -z "$XUI_NODE_ENDPOINT" ]; then
    XUI_NODE_ENDPOINT="$XUI_NODE_ID"
  fi
  cat > "$CONFIG_DIR/agent.env" <<EOF
XUI_MASTER=${XUI_MASTER}
XUI_AGENT_TOKEN=${XUI_AGENT_TOKEN}
XUI_NODE_ID=${XUI_NODE_ID}
XUI_NODE_NAME=${XUI_NODE_NAME}
XUI_NODE_COUNTRY=${XUI_NODE_COUNTRY}
XUI_NODE_REGION=${XUI_NODE_REGION}
XUI_NODE_ENDPOINT=${XUI_NODE_ENDPOINT}
XUI_SSH_USER=${XUI_SSH_USER}
EOF
}

write_services() {
  cat > "$SERVICE_DIR/xui-pro.service" <<EOF
[Unit]
Description=XUI Pro Master
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
EnvironmentFile=-${CONFIG_DIR}/master.env
ExecStart=${INSTALL_DIR}/xuid
Restart=on-failure
RestartSec=5s
LimitNOFILE=1048576

[Install]
WantedBy=multi-user.target
EOF

  cat > "$SERVICE_DIR/xui-pro-agent.service" <<EOF
[Unit]
Description=XUI Pro Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
EnvironmentFile=-${CONFIG_DIR}/agent.env
ExecStart=${INSTALL_DIR}/xui-agent
Restart=on-failure
RestartSec=5s
LimitNOFILE=1048576

[Install]
WantedBy=multi-user.target
EOF
}

enable_mode() {
  systemctl daemon-reload
  case "$MODE" in
    master)
      write_master_env
      systemctl enable --now xui-pro.service
      systemctl disable --now xui-pro-agent.service >/dev/null 2>&1 || true
      echo "XUI Pro master installed. Check: xui-pro status"
      ;;
    agent)
      write_agent_env
      systemctl enable --now xui-pro-agent.service
      echo "XUI Pro agent installed. Check: xui-pro agent-status"
      ;;
    *)
      echo "Unknown mode: $MODE" >&2
      usage
      exit 1
      ;;
  esac
}

main() {
  need_root
  parse_args "$@"
  arch="$(detect_arch)"
  install_files "$arch"
  write_services
  enable_mode
}

main "$@"
