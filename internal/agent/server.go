package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"xui-next/internal/domain"
)

type Agent struct {
	MasterURL string
	Node      domain.Node
	client    *http.Client
}

func New(masterURL string, node domain.Node) *Agent {
	return &Agent{
		MasterURL: masterURL,
		Node:      node,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (a *Agent) Register() error {
	return a.post("/api/nodes/register", a.Node, nil)
}

func (a *Agent) Heartbeat() error {
	req := domain.HeartbeatRequest{
		Node:    a.Node,
		Metrics: CollectMetrics(),
	}
	return a.post("/api/nodes/"+a.Node.ID+"/heartbeat", req, nil)
}

func (a *Agent) DesiredConfig() (domain.DesiredConfig, error) {
	var cfg domain.DesiredConfig
	err := a.post("/api/nodes/"+a.Node.ID+"/desired-config", map[string]string{}, &cfg)
	return cfg, err
}

func (a *Agent) post(path string, body any, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	resp, err := a.client.Post(a.MasterURL+path, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("master returned %s", resp.Status)
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func NodeFromEnv() domain.Node {
	hostname, _ := os.Hostname()
	country := env("XUI_NODE_COUNTRY", env("XUI_NODE_REGION", "unknown"))
	return domain.Node{
		ID:       env("XUI_NODE_ID", hostname),
		Name:     env("XUI_NODE_NAME", hostname),
		Region:   country,
		Country:  country,
		Endpoint: env("XUI_NODE_ENDPOINT", hostname),
		SSHUser:  env("XUI_SSH_USER", "root"),
		Version:  "next-dev",
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
