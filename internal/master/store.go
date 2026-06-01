package master

import (
	"errors"
	"sync"
	"time"

	"xui-next/internal/domain"
)

type Store struct {
	mu      sync.RWMutex
	nodes   map[string]domain.Node
	metrics map[string]domain.NodeMetric
	configs map[string]domain.DesiredConfig
}

func NewStore() *Store {
	return &Store{
		nodes:   map[string]domain.Node{},
		metrics: map[string]domain.NodeMetric{},
		configs: map[string]domain.DesiredConfig{},
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
		out = append(out, n)
	}
	return out
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
