package embedding

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type GPUStats struct {
	Name                       string  `json:"name"`
	UtilizationPercent         float64 `json:"utilization_percent"`
	MemoryUsedMiB              float64 `json:"memory_used_mib"`
	MemoryTotalMiB             float64 `json:"memory_total_mib"`
	TemperatureC               float64 `json:"temperature_c"`
	PowerDrawW                 float64 `json:"power_draw_w,omitempty"`
	PowerLimitW                float64 `json:"power_limit_w,omitempty"`
	GraphicsClockMHz           float64 `json:"graphics_clock_mhz,omitempty"`
	SoftwareThermalSlowdown    string  `json:"software_thermal_slowdown,omitempty"`
	HardwareThermalSlowdown    string  `json:"hardware_thermal_slowdown,omitempty"`
	HardwarePowerBrakeSlowdown string  `json:"hardware_power_brake_slowdown,omitempty"`
}

func ReadGPUStats(ctx context.Context) (GPUStats, error) {
	path := findNvidiaSMI()
	if path == "" {
		return GPUStats{}, errors.New("nvidia-smi not found")
	}
	fields := strings.Join([]string{
		"name", "utilization.gpu", "memory.used", "memory.total", "temperature.gpu",
		"power.draw", "power.limit", "clocks.current.graphics",
	}, ",")
	output, err := exec.CommandContext(ctx, path, "--query-gpu="+fields, "--format=csv,noheader,nounits").Output()
	if err != nil {
		return GPUStats{}, fmt.Errorf("query NVIDIA GPU telemetry: %w", err)
	}
	rows, err := csv.NewReader(strings.NewReader(strings.TrimSpace(string(output)))).ReadAll()
	if err != nil || len(rows) == 0 || len(rows[0]) < 8 {
		return GPUStats{}, errors.New("parse NVIDIA GPU telemetry")
	}
	row := rows[0]
	stats := GPUStats{Name: strings.TrimSpace(row[0])}
	stats.UtilizationPercent = parseGPUFloat(row[1])
	stats.MemoryUsedMiB = parseGPUFloat(row[2])
	stats.MemoryTotalMiB = parseGPUFloat(row[3])
	stats.TemperatureC = parseGPUFloat(row[4])
	stats.PowerDrawW = parseGPUFloat(row[5])
	stats.PowerLimitW = parseGPUFloat(row[6])
	stats.GraphicsClockMHz = parseGPUFloat(row[7])

	reasons := strings.Join([]string{
		"clocks_event_reasons.sw_thermal_slowdown",
		"clocks_event_reasons.hw_thermal_slowdown",
		"clocks_event_reasons.hw_power_brake_slowdown",
	}, ",")
	if reasonOutput, reasonErr := exec.CommandContext(ctx, path, "--query-gpu="+reasons, "--format=csv,noheader,nounits").Output(); reasonErr == nil {
		if reasonRows, readErr := csv.NewReader(strings.NewReader(strings.TrimSpace(string(reasonOutput)))).ReadAll(); readErr == nil && len(reasonRows) > 0 && len(reasonRows[0]) >= 3 {
			stats.SoftwareThermalSlowdown = strings.TrimSpace(reasonRows[0][0])
			stats.HardwareThermalSlowdown = strings.TrimSpace(reasonRows[0][1])
			stats.HardwarePowerBrakeSlowdown = strings.TrimSpace(reasonRows[0][2])
		}
	}
	return stats, nil
}

func parseGPUFloat(value string) float64 {
	value = strings.TrimSpace(strings.TrimSuffix(value, "%"))
	parsed, _ := strconv.ParseFloat(value, 64)
	return parsed
}
