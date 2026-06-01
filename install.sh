#!/usr/bin/env bash
set -euo pipefail

REPO="${XUI_PRO_REPO:-tyrantcwj/xui-pro}"
INSTALL_DIR="${XUI_PRO_INSTALL_DIR:-/usr/local/xui-pro}"
BIN_DIR="${XUI_PRO_BIN_DIR:-/usr/local/bin}"
CONFIG_DIR="${XUI_PRO_CONFIG_DIR:-/etc/xui-pro}"
SERVICE_DIR="/etc/systemd/system"
VERSION="${XUI_PRO_VERSION:-latest}"
MODE="${1:-master}"

usage() {
  cat <<'EOF'
Usage:
  install.sh master [--listen :8080]
  install.sh agent --master https://panel.example.com [--token TOKEN] [--node-id NODE_ID] [--region REGION]

Examples:
  bash <(curl -Ls https://raw.githubusercontent.com/tyrantcwj/xui-pro/main/install.sh) master
  bash <(curl -Ls https://raw.githubusercontent.com/tyrantcwj/xui-pro/main/install.sh) agent --master https://panel.example.com --token xxx

Environment:
  XUI_PRO_VERSION=latest|v0.1.0
  XUI_PRO_REPO=tyrantcwj/xui-pro
  XUI_PRO_INSTALL_DIR=/usr/local/xui-pro
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

parse_args() {
  XUI_LISTEN="${XUI_LISTEN:-:8080}"
  XUI_MASTER="${XUI_MASTER:-}"
  XUI_AGENT_TOKEN="${XUI_AGENT_TOKEN:-}"
  XUI_NODE_ID="${XUI_NODE_ID:-}"
  XUI_NODE_REGION="${XUI_NODE_REGION:-unknown}"

  shift || true
  while [ "$#" -gt 0 ]; do
    case "$1" in
      --listen) XUI_LISTEN="$2"; shift 2 ;;
      --master) XUI_MASTER="$2"; shift 2 ;;
      --token) XUI_AGENT_TOKEN="$2"; shift 2 ;;
      --node-id) XUI_NODE_ID="$2"; shift 2 ;;
      --region) XUI_NODE_REGION="$2"; shift 2 ;;
      -h|--help) usage; exit 0 ;;
      *) echo "Unknown argument: $1" >&2; usage; exit 1 ;;
    esac
  done
}

install_files() {
  local arch="$1"
  local tmp
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT

  echo "Downloading xui-pro ${VERSION} for linux-${arch}..."
  curl -fL "$(download_url "$arch")" -o "$tmp/xui-pro.tar.gz"
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

write_master_env() {
  cat > "$CONFIG_DIR/master.env" <<EOF
XUI_LISTEN=${XUI_LISTEN}
XUI_REALITY_LIBRARY=${INSTALL_DIR}/reality/domains.json
EOF
}

write_agent_env() {
  if [ -z "$XUI_MASTER" ]; then
    echo "agent mode requires --master https://panel.example.com" >&2
    exit 1
  fi
  if [ -z "$XUI_NODE_ID" ]; then
    XUI_NODE_ID="$(hostname)"
  fi
  cat > "$CONFIG_DIR/agent.env" <<EOF
XUI_MASTER=${XUI_MASTER}
XUI_AGENT_TOKEN=${XUI_AGENT_TOKEN}
XUI_NODE_ID=${XUI_NODE_ID}
XUI_NODE_NAME=${XUI_NODE_ID}
XUI_NODE_REGION=${XUI_NODE_REGION}
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
