package main

import (
	"fmt"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/jackmordaunt/ghw"
	"os"
	"os/exec"
	"slices"
	"strings"
)

func GetCPUModel() string {
	cpu, err := ghw.CPU()
	if err != nil {
		return ""
	}
	if len(cpu.Processors) == 0 {
		return ""
	}
	return cpu.Processors[0].Model
}

func GetCPUThreads() int {
	cpu, err := ghw.CPU()
	if err != nil {
		return 0
	}
	return int(cpu.TotalThreads)
}

func GetGPUModels() (ret []string) {
	cmd := exec.Command("sh", "-c", "lspci -v -m | grep 'VGA' -A6 | grep '^Device:'")
	bytes, err := cmd.Output()
	if err != nil {
		return nil
	}

	for i, gpu := range strings.Split(string(bytes), "\n") {
		if slices.Contains(config.HiddenGPUS, i+1) {
			continue
		}
		if gpu == "" {
			continue
		}
		gpu = strings.TrimPrefix(strings.TrimSpace(gpu), "Device:\t")
		ret = append(ret, gpu)
	}

	return ret
}

func GetMotherboardModel() string {
	bytes, err := os.ReadFile("/sys/devices/virtual/dmi/id/board_name")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(bytes))
}

func GetMonitorResolution() []string {
	var monitors []string
	if GetDisplayProtocol() != "" {
		err := glfw.Init()
		if err != nil {
			panic(err)
		}
		for _, monitor := range glfw.GetMonitors() {
			mode := monitor.GetVideoMode()
			monitors = append(monitors, fmt.Sprintf("%dx%d %dHz", mode.Width, mode.Height, mode.RefreshRate))
		}
		defer glfw.Terminate()
	}
	return monitors
}
