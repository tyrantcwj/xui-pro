package master

import (
	"errors"
	"sync"
	"time"

	"xui-next/internal/domain"
)

type Store struct {
	mu        sync.RWMutex
	nodes     map[string]domain.Node
	metrics   map[string]domain.NodeMetric
	configs   map[string]domain.DesiredConfig
	overrides map[string]domain.Node
}

func NewStore() *Store {
	return &Store{
		nodes:     map[string]domain.Node{},
		metrics:   map[string]domain.NodeMetric{},
		configs:   map[string]domain.DesiredConfig{},
		overrides: map[string]domain.Node{},
	}
}

func (s *Store) UpsertNode(n domain.Node) domain.Node {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	existing, ok := s.nodes[n.ID]
	if !ok {
		n.CreatedAt = now
	} else {
		n.CreatedAt = existing.CreatedAt
	}
	if n.Country == "" {
		n.Country = n.Region
	}
	if n.Region == "" {
		n.Region = n.Country
	}
	if n.SSHUser == "" {
		n.SSHUser = "root"
	}
	if override, ok := s.overrides[n.ID]; ok {
		n = applyNodeOverride(n, override)
	}
	n.LastSeen = now
	n.Status = domain.NodeStatusOnline
	s.nodes[n.ID] = n
	return n
}

func (s *Store) SaveHeartbeat(req domain.HeartbeatRequest) domain.Node {
	n := s.UpsertNode(req.Node)
	s.mu.Lock()
	req.Metrics.NodeID = n.ID
	req.Metrics.SeenAt = n.LastSeen
	s.metrics[n.ID] = req.Metrics
	s.mu.Unlock()
	return n
}

func (s *Store) Nodes() []domain.Node {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]domain.Node, 0, len(s.nodes))
	for _, n := range s.nodes {
		if m, ok := s.metrics[n.ID]; ok {
			metric := m
			n.Metrics = &metric
		}
		out = append(out, n)
	}
	return out
}

func (s *Store) UpdateNode(id string, patch domain.Node) (domain.Node, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n, ok := s.nodes[id]
	if !ok {
		return domain.Node{}, false
	}
	if patch.Name != "" {
		n.Name = patch.Name
	}
	if patch.Country != "" {
		n.Country = patch.Country
		n.Region = patch.Country
	}
	if patch.Endpoint != "" {
		n.Endpoint = patch.Endpoint
	}
	if patch.SSHUser != "" {
		n.SSHUser = patch.SSHUser
	}
	s.overrides[id] = mergeNodeOverride(s.overrides[id], patch)
	s.nodes[id] = n
	return n, true
}

func applyNodeOverride(n domain.Node, override domain.Node) domain.Node {
	if override.Name != "" {
		n.Name = override.Name
	}
	if override.Country != "" {
		n.Country = override.Country
		n.Region = override.Country
	}
	if override.Endpoint != "" {
		n.Endpoint = override.Endpoint
	}
	if override.SSHUser != "" {
		n.SSHUser = override.SSHUser
	}
	return n
}

func mergeNodeOverride(existing domain.Node, patch domain.Node) domain.Node {
	if patch.Name != "" {
		existing.Name = patch.Name
	}
	if patch.Country != "" {
		existing.Country = patch.Country
		existing.Region = patch.Country
	}
	if patch.Endpoint != "" {
		existing.Endpoint = patch.Endpoint
	}
	if patch.SSHUser != "" {
		existing.SSHUser = patch.SSHUser
	}
	return existing
}

func (s *Store) DesiredConfig(nodeID string) (domain.DesiredConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cfg, ok := s.configs[nodeID]
	if !ok {
		return domain.DesiredConfig{}, errors.New("no desired config")
	}
	return cfg, nil
}

func (s *Store) SetDesiredConfig(nodeID string, cfg domain.DesiredConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configs[nodeID] = cfg
}
