package main

import (
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/jackmordaunt/ghw"
)

type CPU struct {
	Vendor  string
	Model   string
	Cores   int
	Threads int
}

type Monitor struct {
	Width       int
	Height      int
	RefreshRate int
}

func GetCPUs(hiddenCPUs []int) []CPU {
	ret := make([]CPU, 0)

	cpus, err := ghw.CPU()
	if err != nil {
		return ret
	}

	for i, cpu := range cpus.Processors {
		if slices.Contains(hiddenCPUs, i+1) {
			continue
		}

		ret = append(ret, CPU{
			Vendor:  cpu.Vendor,
			Model:   cpu.Model,
			Cores:   int(cpu.NumCores),
			Threads: int(cpu.NumThreads),
		})
	}

	return ret
}

func GetGPUModels(hiddenGPUS []int) (ret []string) {
	cmd := exec.Command("sh", "-c", "lspci -v -m | grep 'VGA' -A6 | grep '^Device:'")
	bytes, err := cmd.Output()
	if err != nil {
		return nil
	}

	for i, gpu := range strings.Split(string(bytes), "\n") {
		if slices.Contains(hiddenGPUS, i+1) {
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

func GetMonitors() []Monitor {
	var monitors []Monitor
	if GetDisplayProtocol() != "" {
		err := glfw.Init()
		if err != nil {
			panic(err)
		}

		for _, monitor := range glfw.GetMonitors() {
			mode := monitor.GetVideoMode()

			monitors = append(monitors, Monitor{
				Width:       mode.Width,
				Height:      mode.Height,
				RefreshRate: mode.RefreshRate,
			})
		}
		defer glfw.Terminate()
	}
	return monitors
}
