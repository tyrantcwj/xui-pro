package master

import (
	"encoding/json"
	"net/http"
	"strings"

	"xui-next/internal/domain"
	"xui-next/internal/reality"
)

type Server struct {
	store   *Store
	reality *reality.Library
}

func NewServer(store *Store, reality *reality.Library) *Server {
	return &Server{store: store, reality: reality}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.health)
	mux.HandleFunc("GET /api/nodes", s.nodes)
	mux.HandleFunc("POST /api/nodes/register", s.registerNode)
	mux.HandleFunc("POST /api/nodes/", s.nodeAction)
	mux.HandleFunc("GET /api/reality/domains", s.realityDomains)
	mux.HandleFunc("POST /api/reality/recommend", s.realityRecommend)
	return withJSON(mux)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) nodes(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.store.Nodes())
}

func (s *Server) registerNode(w http.ResponseWriter, r *http.Request) {
	var n domain.Node
	if !decodeJSON(w, r, &n) {
		return
	}
	if n.ID == "" {
		http.Error(w, "node id is required", http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, s.store.UpsertNode(n))
}

func (s *Server) nodeAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/nodes/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}
	nodeID, action := parts[0], parts[1]
	switch action {
	case "heartbeat":
		var req domain.HeartbeatRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		req.Node.ID = nodeID
		writeJSON(w, http.StatusOK, s.store.SaveHeartbeat(req))
	case "desired-config":
		cfg, err := s.store.DesiredConfig(nodeID)
		if err != nil {
			writeJSON(w, http.StatusOK, domain.DesiredConfig{Version: "empty"})
			return
		}
		writeJSON(w, http.StatusOK, cfg)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) realityDomains(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.reality.Domains(r.URL.Query().Get("region")))
}

func (s *Server) realityRecommend(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Region string `json:"region"`
		Limit  int    `json:"limit"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	writeJSON(w, http.StatusOK, s.reality.Recommend(req.Region, req.Limit))
}

func withJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
