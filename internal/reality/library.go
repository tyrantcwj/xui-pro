package reality

import (
	"crypto/tls"
	"encoding/json"
	"net"
	"os"
	"sort"
	"time"

	"xui-next/internal/domain"
)

type Library struct {
	domains []domain.RealityDomain
	timeout time.Duration
}

func Load(path string) (*Library, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var domains []domain.RealityDomain
	if err := json.Unmarshal(data, &domains); err != nil {
		return nil, err
	}
	return &Library{domains: domains, timeout: 4 * time.Second}, nil
}

func (l *Library) Domains(region string) []domain.RealityDomain {
	out := make([]domain.RealityDomain, 0, len(l.domains))
	for _, d := range l.domains {
		if region == "" || d.Region == region || d.Region == "global" {
			out = append(out, d)
		}
	}
	return out
}

func (l *Library) Recommend(region string, limit int) []domain.RealityDomain {
	candidates := l.Domains(region)
	for i := range candidates {
		candidates[i] = l.probe(candidates[i])
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].TLSOK != candidates[j].TLSOK {
			return candidates[i].TLSOK
		}
		return candidates[i].LatencyMS < candidates[j].LatencyMS
	})
	if limit <= 0 || limit > len(candidates) {
		limit = len(candidates)
	}
	return candidates[:limit]
}

func (l *Library) probe(d domain.RealityDomain) domain.RealityDomain {
	start := time.Now()
	addr := net.JoinHostPort(d.Domain, "443")
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: l.timeout}, "tcp", addr, &tls.Config{
		ServerName: d.SNI,
		MinVersion: tls.VersionTLS12,
	})
	d.LastChecked = time.Now().UTC().Format(time.RFC3339)
	d.LatencyMS = time.Since(start).Milliseconds()
	if err != nil {
		d.TLSOK = false
		d.LastError = err.Error()
		return d
	}
	_ = conn.Close()
	d.TLSOK = true
	return d
}
