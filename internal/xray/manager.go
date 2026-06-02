package xray

import (
	"os"
	"os/exec"
	"strings"

	"xui-next/internal/domain"
)

type Manager struct {
	ConfigPath  string
	ServiceName string
}

type Status struct {
	ConfigPath  string `json:"configPath"`
	ServiceName string `json:"serviceName"`
	Running     bool   `json:"running"`
	LastError   string `json:"lastError,omitempty"`
}

func NewManager(configPath, serviceName string) *Manager {
	if configPath == "" {
		configPath = "/usr/local/xui-pro/xray/config.json"
	}
	if serviceName == "" {
		serviceName = "xray"
	}
	return &Manager{ConfigPath: configPath, ServiceName: serviceName}
}

func (m *Manager) Render(inbounds []domain.Inbound) ([]byte, error) {
	return BuildConfig(inbounds)
}

func (m *Manager) Apply(inbounds []domain.Inbound) Status {
	data, err := BuildConfig(inbounds)
	if err != nil {
		return Status{ConfigPath: m.ConfigPath, ServiceName: m.ServiceName, LastError: err.Error()}
	}
	if err := os.MkdirAll(dirName(m.ConfigPath), 0755); err != nil {
		return Status{ConfigPath: m.ConfigPath, ServiceName: m.ServiceName, LastError: err.Error()}
	}
	if err := os.WriteFile(m.ConfigPath, data, 0644); err != nil {
		return Status{ConfigPath: m.ConfigPath, ServiceName: m.ServiceName, LastError: err.Error()}
	}
	status := m.Status()
	if err := exec.Command("systemctl", "restart", m.ServiceName).Run(); err != nil {
		status.LastError = "config written, restart failed: " + err.Error()
		return status
	}
	return m.Status()
}

func (m *Manager) Status() Status {
	running := exec.Command("systemctl", "is-active", "--quiet", m.ServiceName).Run() == nil
	return Status{ConfigPath: m.ConfigPath, ServiceName: m.ServiceName, Running: running}
}

func dirName(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx <= 0 {
		return "."
	}
	return path[:idx]
}
