//go:build !linux

package agent

import "xui-next/internal/domain"

func CollectMetrics() domain.NodeMetric {
	return domain.NodeMetric{XrayOK: true}
}
