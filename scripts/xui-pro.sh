#!/usr/bin/env bash
set -euo pipefail

cmd="${1:-status}"

case "$cmd" in
  start) systemctl start xui-pro.service ;;
  stop) systemctl stop xui-pro.service ;;
  restart) systemctl restart xui-pro.service ;;
  status) systemctl status xui-pro.service --no-pager ;;
  logs) journalctl -u xui-pro.service -f ;;
  agent-start) systemctl start xui-pro-agent.service ;;
  agent-stop) systemctl stop xui-pro-agent.service ;;
  agent-restart) systemctl restart xui-pro-agent.service ;;
  agent-status) systemctl status xui-pro-agent.service --no-pager ;;
  agent-logs) journalctl -u xui-pro-agent.service -f ;;
  version)
    /usr/local/xui-pro/xuid --version 2>/dev/null || echo "xui-pro development build"
    ;;
  *)
    cat <<'EOF'
Usage: xui-pro <command>

Master:
  start | stop | restart | status | logs

Agent:
  agent-start | agent-stop | agent-restart | agent-status | agent-logs

Other:
  version
EOF
    exit 1
    ;;
esac
