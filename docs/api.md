# API Draft

## Master API

### `GET /api/health`

Returns service health.

```json
{ "status": "ok" }
```

### `GET /api/nodes`

Lists registered VPS nodes.

### `POST /api/nodes/register`

Registers or refreshes a node.

```json
{
  "id": "hk-01",
  "name": "Hong Kong 01",
  "region": "asia",
  "endpoint": "10.0.0.2",
  "version": "next-dev"
}
```

### `POST /api/nodes/{id}/heartbeat`

Ingests node health and traffic metrics.

```json
{
  "node": { "id": "hk-01", "name": "Hong Kong 01", "region": "asia" },
  "metrics": {
    "cpu": 12.4,
    "memory": 41.2,
    "disk": 55.9,
    "up": 1024,
    "down": 4096,
    "xrayOk": true
  }
}
```

### `POST /api/nodes/{id}/desired-config`

Returns the desired Xray config package for an agent. Empty configs return:

```json
{ "version": "empty", "hash": "", "config": "" }
```

### `GET /api/reality/domains?region=asia`

Lists candidate Reality domains. `global` domains are included with regional
results.

### `POST /api/reality/recommend`

Probes candidate domains from the master side and returns the best options.

```json
{ "region": "asia", "limit": 5 }
```

## Next API Additions

- `POST /api/auth/login`
- `POST /api/enrollments`
- `POST /api/inbound-profiles`
- `POST /api/deployments`
- `GET /api/events`
- `GET /api/audit`
