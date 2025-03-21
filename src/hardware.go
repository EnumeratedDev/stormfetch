package main

import (
	"fmt"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/jackmordaunt/ghw"
	"os"
	"os/exec"
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

func GetGPUModels() []string {
	var ret []string
	cmd := exec.Command("/bin/bash", "-c", "lspci -v -m | grep 'VGA' -A6 | grep '^Device:' | sed 's/^Device://' | awk '{$1=$1};1'")
	bytes, err := cmd.Output()
	if err != nil {
		return nil
	}
	for _, name := range strings.Split(string(bytes), "\n") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		ret = append(ret, name)
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
