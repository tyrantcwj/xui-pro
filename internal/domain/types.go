package domain

import "time"

type NodeStatus string

const (
	NodeStatusPending NodeStatus = "pending"
	NodeStatusOnline  NodeStatus = "online"
	NodeStatusOffline NodeStatus = "offline"
)

type Node struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Region    string      `json:"region"`
	Country   string      `json:"country"`
	Endpoint  string      `json:"endpoint"`
	PublicIP  string      `json:"publicIp"`
	SSHUser   string      `json:"sshUser"`
	Version   string      `json:"version"`
	Status    NodeStatus  `json:"status"`
	LastSeen  time.Time   `json:"lastSeen"`
	CreatedAt time.Time   `json:"createdAt"`
	Metrics   *NodeMetric `json:"metrics,omitempty"`
}

type NodeMetric struct {
	NodeID string    `json:"nodeId"`
	CPU    float64   `json:"cpu"`
	Memory float64   `json:"memory"`
	Disk   float64   `json:"disk"`
	Up     int64     `json:"up"`
	Down   int64     `json:"down"`
	XrayOK bool      `json:"xrayOk"`
	SeenAt time.Time `json:"seenAt"`
}

type HeartbeatRequest struct {
	Node    Node       `json:"node"`
	Metrics NodeMetric `json:"metrics"`
}

type DesiredConfig struct {
	Version string `json:"version"`
	Hash    string `json:"hash"`
	Config  string `json:"config"`
}

type Inbound struct {
	ID             string    `json:"id"`
	Remark         string    `json:"remark"`
	Protocol       string    `json:"protocol"`
	Listen         string    `json:"listen"`
	Port           int       `json:"port"`
	Enabled        bool      `json:"enabled"`
	Up             int64     `json:"up"`
	Down           int64     `json:"down"`
	Total          int64     `json:"total"`
	ExpiryTime     int64     `json:"expiryTime"`
	Settings       string    `json:"settings"`
	StreamSettings string    `json:"streamSettings"`
	Sniffing       string    `json:"sniffing"`
	Tag            string    `json:"tag"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type RealityDomain struct {
	Domain      string `json:"domain"`
	Region      string `json:"region"`
	Category    string `json:"category"`
	Port        int    `json:"port"`
	SNI         string `json:"sni"`
	Notes       string `json:"notes,omitempty"`
	LatencyMS   int64  `json:"latencyMs,omitempty"`
	TLSOK       bool   `json:"tlsOk,omitempty"`
	LastError   string `json:"lastError,omitempty"`
	LastChecked string `json:"lastChecked,omitempty"`
}
