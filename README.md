# XUI Next

XUI Next is a modernized rewrite plan and starter implementation inspired by
`sing-web/x-ui`.

The old project is a single-node Go + Gin + SQLite panel with embedded Vue 2
templates. This version keeps Go for low resource usage, but splits the product
into:

- `xuid`: master control plane and API server
- `xui-agent`: lightweight node agent for each VPS
- `web`: Vue 3 SPA dashboard
- `reality`: curated Reality domain library

## Goals

- Manage many VPS nodes from one master panel.
- Provide a responsive Vue 3 dashboard with dark mode and visual metrics.
- Keep agents small and safe: heartbeat, metrics, Xray config apply, restart.
- Add Reality domain recommendations by region, latency, TLS health, and SNI
  suitability.
- Keep deployment simple: one binary for master, one binary for agent.

## Current Artifact

This folder is an implementation seed rather than a full production fork. It
contains runnable-shape Go code and Vue 3 source structure, but this Codex
environment does not have usable `go` or `node`, so compilation was not run
locally.

## Suggested First Run

```bash
cd xui-next
go run ./cmd/xuid
```

In another terminal:

```bash
XUI_MASTER=http://127.0.0.1:8080 XUI_NODE_ID=hk-01 go run ./cmd/xui-agent
```

Frontend development:

```bash
cd web
npm install
npm run dev
```

## Roadmap

1. Replace in-memory store with SQLite migrations.
2. Add JWT/OIDC login and per-node enrollment tokens.
3. Implement signed config apply between master and agent.
4. Port Xray inbound/outbound generation from the upstream project.
5. Add Reality probing worker and scheduled domain library refresh.
6. Build SPA assets and embed them into `xuid`.
