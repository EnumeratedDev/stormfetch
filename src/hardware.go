package main

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/jackmordaunt/ghw"
)

type CPU struct {
	Model   string
	Cores   int
	Threads int
}

type GPU struct {
	PCIAddress string
	Vendor     string
	Name       string
	Product    string
	Subsystem  string
	Driver     string
	VramTotal  string
	VramUsed   string
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

		// Remove unnecessary information from the CPU model
		model := cpu.Model
		stringsToRemove := []string{
			" CPU", " FPU", " APU", " Processor",
			" Dual-Core", " Quad-Core", " Six-Core", " Eight-Core", " Ten-Core",
			" 2-Core", " 4-Core", " 6-Core", " 8-Core", " 10-Core", " 12-Core", " 14-Core", " 16-Core",
		}
		for _, str := range stringsToRemove {
			model = strings.ReplaceAll(model, str, "")
		}
		model = strings.Split(model, "w/ Radeon ")[0]
		model = strings.Split(model, "with Radeon ")[0]
		model = strings.Split(model, "@")[0]
		model = strings.TrimSpace(model)

		ret = append(ret, CPU{
			Model:   model,
			Cores:   int(cpu.NumCores),
			Threads: int(cpu.NumThreads),
		})
	}

	return ret
}

func GetGPUModels(hiddenGPUs []int) []GPU {
	ret := make([]GPU, 0)

	// Set stderr to nil to avoid warnings
	stderr := os.Stderr
	os.Stderr = nil

	gpus, err := ghw.GPU()
	if err != nil {
		return ret
	}

	// Restore stderr
	os.Stderr = stderr

	for i, gpu := range gpus.GraphicsCards {
		if slices.Contains(hiddenGPUs, i+1) {
			continue
		}

		// Set alternative names for vendors
		var vendor string
		switch gpu.DeviceInfo.Vendor.ID {
		case "1002":
			vendor = "AMD"
		case "10de":
			vendor = "Nvidia"
		case "8086":
			vendor = "Intel"
		default:
			vendor = gpu.DeviceInfo.Vendor.Name
		}

		// Set GPU name
		name := ""

		// Use GPU name from amdgpu.ids database
		if vendor == "AMD" {
			fetchedName, err := fetchAmdGpuName(gpu.DeviceInfo.Product.ID, gpu.DeviceInfo.Revision)
			if err == nil && !config.DisableAmdgpuIdsWarning {
				name = fetchedName
			} else {
				fmt.Println("Warning: could not fetch GPU name from amdgpu.ids database! Error: " + err.Error())
				fmt.Println("         You can disable this warning in the configuration file")
			}
		}

		if name == "" {
			if gpu.DeviceInfo.Subsystem.Name == "" || gpu.DeviceInfo.Subsystem.Name == "unknown" {
				// Set GPU name to product name
				name = gpu.DeviceInfo.Product.Name
			} else {
				// Set GPU name to subsystem name
				name = gpu.DeviceInfo.Subsystem.Name
			}

			// Use GPU name in brackets
			leftBracket := strings.IndexByte(name, '[')
			rightBracket := strings.IndexByte(name, ']')
			if leftBracket != -1 && rightBracket != -1 {
				name = name[leftBracket+1 : rightBracket]
			}
		}

		// Get VRAM
		vramTotal := "Unknown"
		bytes, err := os.ReadFile("/sys/class/drm/card" + strconv.Itoa(gpu.Index) + "/device/mem_info_vram_total")
		if err == nil {
			vramUint, _ := strconv.ParseUint(strings.TrimSpace(string(bytes)), 10, 64)
			vramTotal = FormatBytes(vramUint)
		}
		vramUsed := "Unknown"
		bytes, err = os.ReadFile("/sys/class/drm/card" + strconv.Itoa(gpu.Index) + "/device/mem_info_vram_used")
		if err == nil {
			vramUint, _ := strconv.ParseUint(strings.TrimSpace(string(bytes)), 10, 64)
			vramUsed = FormatBytes(vramUint)
		}

		ret = append(ret, GPU{
			PCIAddress: gpu.Address,
			Vendor:     vendor,
			Name:       name,
			Product:    gpu.DeviceInfo.Product.Name,
			Subsystem:  gpu.DeviceInfo.Subsystem.Name,
			Driver:     gpu.DeviceInfo.Driver,
			VramTotal:  vramTotal,
			VramUsed:   vramUsed,
		})
	}

	return ret
}

func GetMotherboardModel() string {
	bytes, err := os.ReadFile("/sys/devices/virtual/dmi/id/board_name")
	if err != nil {
		return ""
	}

	// Remove duplicate whitespaces
	ret := strings.Join(strings.Fields(string(bytes)), " ")

	return ret
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
