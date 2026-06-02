package xray

import (
	"encoding/json"
	"fmt"

	"xui-next/internal/domain"
)

type Config struct {
	Log       map[string]any   `json:"log,omitempty"`
	API       map[string]any   `json:"api,omitempty"`
	Stats     map[string]any   `json:"stats,omitempty"`
	Policy    map[string]any   `json:"policy,omitempty"`
	Inbounds  []map[string]any `json:"inbounds"`
	Outbounds []map[string]any `json:"outbounds"`
	Routing   map[string]any   `json:"routing,omitempty"`
}

func BuildConfig(inbounds []domain.Inbound) ([]byte, error) {
	cfg := Config{
		Log: map[string]any{"loglevel": "warning"},
		API: map[string]any{
			"tag":      "api",
			"services": []string{"StatsService"},
		},
		Stats: map[string]any{},
		Policy: map[string]any{
			"levels": map[string]any{
				"0": map[string]any{
					"statsUserUplink":   true,
					"statsUserDownlink": true,
				},
			},
			"system": map[string]any{
				"statsInboundUplink":    true,
				"statsInboundDownlink":  true,
				"statsOutboundUplink":   true,
				"statsOutboundDownlink": true,
			},
		},
		Inbounds: []map[string]any{
			{
				"tag":      "api",
				"listen":   "127.0.0.1",
				"port":     62789,
				"protocol": "dokodemo-door",
				"settings": map[string]any{"address": "127.0.0.1"},
			},
		},
		Outbounds: []map[string]any{
			{"tag": "direct", "protocol": "freedom"},
			{"tag": "blocked", "protocol": "blackhole"},
		},
		Routing: map[string]any{
			"rules": []map[string]any{
				{"type": "field", "inboundTag": []string{"api"}, "outboundTag": "api"},
			},
		},
	}
	for _, inbound := range inbounds {
		if !inbound.Enabled {
			continue
		}
		item, err := buildInbound(inbound)
		if err != nil {
			return nil, err
		}
		cfg.Inbounds = append(cfg.Inbounds, item)
	}
	return json.MarshalIndent(cfg, "", "  ")
}

func buildInbound(inbound domain.Inbound) (map[string]any, error) {
	if inbound.Port <= 0 {
		return nil, fmt.Errorf("inbound %s port is required", inbound.ID)
	}
	settings, err := decodeObject(inbound.Settings, map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("inbound %s settings: %w", inbound.ID, err)
	}
	stream, err := decodeObject(inbound.StreamSettings, map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("inbound %s streamSettings: %w", inbound.ID, err)
	}
	sniffing, err := decodeObject(inbound.Sniffing, map[string]any{"enabled": true, "destOverride": []string{"http", "tls", "quic"}})
	if err != nil {
		return nil, fmt.Errorf("inbound %s sniffing: %w", inbound.ID, err)
	}
	tag := inbound.Tag
	if tag == "" {
		tag = inbound.ID
	}
	out := map[string]any{
		"tag":      tag,
		"port":     inbound.Port,
		"protocol": inbound.Protocol,
		"settings": settings,
		"sniffing": sniffing,
	}
	if inbound.Listen != "" {
		out["listen"] = inbound.Listen
	}
	if len(stream) > 0 {
		out["streamSettings"] = stream
	}
	return out, nil
}

func decodeObject(raw string, fallback map[string]any) (map[string]any, error) {
	if raw == "" {
		return fallback, nil
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	if out == nil {
		return map[string]any{}, nil
	}
	return out, nil
}
