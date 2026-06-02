//go:build linux

package agent

import (
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"xui-next/internal/domain"
)

func CollectMetrics() domain.NodeMetric {
	return domain.NodeMetric{
		CPU:    cpuPercent(),
		Memory: memoryPercent(),
		Disk:   diskPercent("/"),
		XrayOK: true,
	}
}

func cpuPercent() float64 {
	first, ok := readCPU()
	if !ok {
		return 0
	}
	time.Sleep(120 * time.Millisecond)
	second, ok := readCPU()
	if !ok {
		return 0
	}
	idle := float64(second.idle - first.idle)
	total := float64(second.total - first.total)
	if total <= 0 {
		return 0
	}
	return round((1 - idle/total) * 100)
}

type cpuSample struct {
	idle  uint64
	total uint64
}

func readCPU() (cpuSample, bool) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return cpuSample{}, false
	}
	line := strings.SplitN(string(data), "\n", 2)[0]
	fields := strings.Fields(line)
	if len(fields) < 5 || fields[0] != "cpu" {
		return cpuSample{}, false
	}
	var nums []uint64
	for _, f := range fields[1:] {
		v, err := strconv.ParseUint(f, 10, 64)
		if err != nil {
			return cpuSample{}, false
		}
		nums = append(nums, v)
	}
	var total uint64
	for _, v := range nums {
		total += v
	}
	idle := nums[3]
	if len(nums) > 4 {
		idle += nums[4]
	}
	return cpuSample{idle: idle, total: total}, true
}

func memoryPercent() float64 {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0
	}
	values := map[string]uint64{}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSuffix(fields[0], ":")
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err == nil {
			values[key] = value
		}
	}
	total := values["MemTotal"]
	available := values["MemAvailable"]
	if total == 0 {
		return 0
	}
	return round((1 - float64(available)/float64(total)) * 100)
}

func diskPercent(path string) float64 {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0
	}
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	if total == 0 {
		return 0
	}
	return round((1 - float64(free)/float64(total)) * 100)
}

func round(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}
