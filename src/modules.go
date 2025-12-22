package main

import (
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Used to declare modules in config files
type stormfetchModuleConfig struct {
	Name   string         `yaml:"name"`
	Format string         `yaml:"format"`
	Data   map[string]any `yaml:"data"`
}

type StormfetchModule struct {
	Execute func(StormfetchModule) string
	stormfetchModuleConfig
}

var Modules map[string]StormfetchModule = make(map[string]StormfetchModule)

func (sm StormfetchModule) GetData(key string, defaultValue any) (any, bool) {
	if sm.Data == nil {
		return defaultValue, false
	}

	data, ok := sm.Data[key]
	if !ok {
		return defaultValue, false
	}

	if reflect.ValueOf(data).Kind() != reflect.ValueOf(defaultValue).Kind() {
		return defaultValue, false
	}

	return data, true
}

func RegisterModule(module StormfetchModule) bool {
	if _, ok := Modules[module.Name]; ok {
		return false
	}

	Modules[module.Name] = module
	return true
}

func initializeModuleMap() {
	// Distribution Module
	distributionModule := StormfetchModule{stormfetchModuleConfig: stormfetchModuleConfig{Name: "distribution", Format: "%3Distribution: %4$DISTRO_SHORT ($ARCH)"}, Execute: func(sm StormfetchModule) string {
		distroInfo := GetDistroInfo()
		return os.Expand(sm.Format, func(s string) string {
			switch s {
			case "DISTRO_ID":
				return distroInfo.ID
			case "DISTRO_SHORT":
				return distroInfo.ShortName
			case "DISTRO_LONG":
				return distroInfo.LongName
			case "ARCH":
				return GetArch()
			default:
				return ""
			}
		})
	}}
	RegisterModule(distributionModule)

	// Hostname module
	hostnameModule := StormfetchModule{stormfetchModuleConfig: stormfetchModuleConfig{Name: "hostname", Format: "%3Hostname: %4$HOSTNAME"}, Execute: func(sm StormfetchModule) string {
		hostname, _ := os.Hostname()
		return os.Expand(sm.Format, func(s string) string {
			switch s {
			case "HOSTNAME":
				return hostname
			default:
				return ""
			}
		})
	}}
	RegisterModule(hostnameModule)

	// Kernel module
	kernelModule := StormfetchModule{stormfetchModuleConfig: stormfetchModuleConfig{Name: "kernel", Format: "%3Kernel: %4$KERNEL_NAME $KERNEL_RELEASE"}, Execute: func(sm StormfetchModule) string {
		kernelName, kernelRelease := GetKernel()
		return os.Expand(sm.Format, func(s string) string {
			switch s {
			case "KERNEL_NAME":
				return kernelName
			case "KERNEL_RELEASE":
				return kernelRelease
			default:
				return ""
			}
		})
	}}
	RegisterModule(kernelModule)

	// Packages module
	packagesModule := StormfetchModule{stormfetchModuleConfig: stormfetchModuleConfig{Name: "packages", Format: "%3Packages: %4$PACKAGES"}, Execute: func(sm StormfetchModule) string {
		return os.Expand(sm.Format, func(s string) string {
			switch s {
			case "PACKAGES":
				return GetInstalledPackages()
			default:
				return ""
			}
		})
	}}
	RegisterModule(packagesModule)

	// Shell module
	shellModule := StormfetchModule{stormfetchModuleConfig: stormfetchModuleConfig{Name: "shell", Format: "%3Shell: %4$SHELL"}, Execute: func(sm StormfetchModule) string {
		return os.Expand(sm.Format, func(s string) string {
			switch s {
			case "SHELL":
				return GetShell()
			default:
				return ""
			}
		})
	}}
	RegisterModule(shellModule)

	// Init system module
	initSystemModule := StormfetchModule{stormfetchModuleConfig: stormfetchModuleConfig{Name: "init_system", Format: "%3Init: %4$INIT"}, Execute: func(sm StormfetchModule) string {
		return os.Expand(sm.Format, func(s string) string {
			switch s {
			case "INIT":
				return GetInitSystem()
			default:
				return ""
			}
		})
	}}
	RegisterModule(initSystemModule)

	// Libc module
	libcModule := StormfetchModule{stormfetchModuleConfig: stormfetchModuleConfig{Name: "libc", Format: "%3Libc: %4$LIBC"}, Execute: func(sm StormfetchModule) string {
		return os.Expand(sm.Format, func(s string) string {
			switch s {
			case "LIBC":
				return GetLibc()
			default:
				return ""
			}
		})
	}}
	RegisterModule(libcModule)

	// Motherboard module
	MotherboardModule := StormfetchModule{stormfetchModuleConfig: stormfetchModuleConfig{Name: "motherboard", Format: "%3Motherboard: %4$MOTHERBOARD"}, Execute: func(sm StormfetchModule) string {
		motherboard := GetMotherboardModel()

		// Return empty string if can't detect motherboard model
		if motherboard == "" {
			return ""
		}

		return os.Expand(sm.Format, func(s string) string {
			switch s {
			case "MOTHERBOARD":
				return motherboard
			default:
				return ""
			}
		})
	}}
	RegisterModule(MotherboardModule)

	// Motherboard module
	cpusModule := StormfetchModule{stormfetchModuleConfig: stormfetchModuleConfig{Name: "cpus", Format: "%3CPU: %4$CPU_MODEL ($CPU_THREADS threads)"}, Execute: func(sm StormfetchModule) string {
		hiddenCPUsInterface, _ := sm.GetData("hidden_cpus", make([]any, 0))

		// Convert interface slices to string slices
		hiddenCPUs := make([]int, 0)
		for _, value := range hiddenCPUsInterface.([]any) {
			hiddenCPUs = append(hiddenCPUs, value.(int))
		}

		builder := strings.Builder{}
		cpus := GetCPUs(hiddenCPUs)

		for i, cpu := range cpus {
			expanded := os.Expand(sm.Format, func(s string) string {
				switch s {
				case "CPU_NUM":
					return strconv.Itoa(i + 1)
				case "CPU_VENDOR":
					return cpu.Vendor
				case "CPU_MODEL":
					return cpu.Model
				case "CPU_CORES":
					return strconv.Itoa(cpu.Cores)
				case "CPU_THREADS":
					return strconv.Itoa(cpu.Threads)
				default:
					return ""
				}
			})

			builder.WriteString(expanded + "\n")
		}

		return builder.String()
	}}
	RegisterModule(cpusModule)

	// GPUs module
	gpusModule := StormfetchModule{stormfetchModuleConfig: stormfetchModuleConfig{Name: "gpus", Format: "%3GPU: %4$GPU_MODEL"}, Execute: func(sm StormfetchModule) string {
		hiddenGPUsInterface, _ := sm.GetData("hidden_gpus", make([]any, 0))

		// Convert interface slices to string slices
		hiddenGPUs := make([]int, 0)
		for _, value := range hiddenGPUsInterface.([]any) {
			hiddenGPUs = append(hiddenGPUs, value.(int))
		}

		builder := strings.Builder{}
		gpus := GetGPUModels(hiddenGPUs)

		for i, gpu := range gpus {
			expanded := os.Expand(sm.Format, func(s string) string {
				switch s {
				case "GPU_NUM":
					return strconv.Itoa(i + 1)
				case "GPU_MODEL":
					return gpu
				default:
					return ""
				}
			})

			builder.WriteString(expanded + "\n")
		}

		return builder.String()
	}}
	RegisterModule(gpusModule)

	// Memory module
	memoryModule := StormfetchModule{stormfetchModuleConfig: stormfetchModuleConfig{Name: "memory", Format: "%3Memory: %4$MEM_USED MiB / $MEM_TOTAL MiB"}, Execute: func(sm StormfetchModule) string {
		memoryInfo := GetMemoryInfo()

		return os.Expand(sm.Format, func(s string) string {
			switch s {
			case "MEM_TOTAL":
				return strconv.Itoa(memoryInfo.MemTotal)
			case "MEM_AVAILABLE":
				return strconv.Itoa(memoryInfo.MemAvailable)
			case "MEM_FREE":
				return strconv.Itoa(memoryInfo.MemFree)
			case "MEM_USED":
				return strconv.Itoa(memoryInfo.MemTotal - memoryInfo.MemAvailable)
			default:
				return ""
			}
		})
	}}
	RegisterModule(memoryModule)

	// Partitions module
	partitionsModule := StormfetchModule{stormfetchModuleConfig: stormfetchModuleConfig{Name: "partitions", Format: "%3Partition ${PART_AUTONAME} (${PART_FS}): %4${PART_USED} / ${PART_TOTAL}"}, Execute: func(sm StormfetchModule) string {
		hiddenPartitionsInterface, _ := sm.GetData("hidden_partitions", make([]any, 0))
		hiddenFilesystemsInterface, _ := sm.GetData("hidden_filesystems", make([]any, 0))
		alternativeNamesInterface, _ := sm.GetData("alternative_names", make(map[string]any, 0))

		// Convert interface slices to string slices
		hiddenPartitions := make([]string, 0)
		for _, value := range hiddenPartitionsInterface.([]any) {
			hiddenPartitions = append(hiddenPartitions, value.(string))
		}
		hiddenFilesystems := make([]string, 0)
		for _, value := range hiddenFilesystemsInterface.([]any) {
			hiddenFilesystems = append(hiddenFilesystems, value.(string))
		}

		// Convert interface map to map[string]string
		alternativeNames := make(map[string]string)
		for key, value := range alternativeNamesInterface.(map[string]any) {
			alternativeNames[key] = value.(string)
		}

		builder := strings.Builder{}
		partitions := GetMountedPartitions(hiddenPartitions, hiddenFilesystems)

		for i, partition := range partitions {
			partitionAutoname := ""
			if altName, ok := alternativeNames[partition.Device]; ok {
				partitionAutoname = altName
			} else if partition.Label != "" {
				partitionAutoname = partition.Label
			} else {
				partitionAutoname = partition.MountPoint
			}

			expanded := os.Expand(sm.Format, func(s string) string {
				switch s {
				case "PART_NUM":
					return strconv.Itoa(i + 1)
				case "PART_FS":
					return partition.FileystemType
				case "PART_DEVICE":
					return partition.Device
				case "PART_AUTONAME":
					return partitionAutoname
				case "PART_LABEL":
					return partition.Label
				case "PART_MOUNTPOINT":
					return partition.MountPoint
				case "PART_FREE":
					return FormatBytes(partition.FreeSize)
				case "PART_USED":
					return FormatBytes(partition.UsedSize)
				case "PART_TOTAL":
					return FormatBytes(partition.TotalSize)
				default:
					return ""
				}
			})

			builder.WriteString(expanded + "\n")
		}

		return builder.String()
	}}
	RegisterModule(partitionsModule)

	// Local IP module
	localIpModule := StormfetchModule{stormfetchModuleConfig: stormfetchModuleConfig{Name: "local_ip", Format: "%3Local IP: %4$LOCAL_IP"}, Execute: func(sm StormfetchModule) string {
		return os.Expand(sm.Format, func(s string) string {
			switch s {
			case "LOCAL_IP":
				return GetLocalIP()
			default:
				return ""
			}
		})
	}}
	RegisterModule(localIpModule)

	// DEWM module
	dewmModule := StormfetchModule{stormfetchModuleConfig: stormfetchModuleConfig{Name: "de_wm", Format: "%3${DEWM_TYPE}: %4${DEWM_NAME} ${DEWM_VERSION} ($DISPLAY_PROTOCOL)"}, Execute: func(sm StormfetchModule) string {
		// Return empty string if currently in TTY
		if os.Getenv("XDG_SESSION_TYPE") == "" || os.Getenv("XDG_SESSION_TYPE") == "tty" {
			return ""
		}

		dewm := GetDEWM()

		// Return empty string if can't detect DE/WM
		if dewm.Name == "Unknown" {
			return ""
		}

		return os.Expand(sm.Format, func(s string) string {
			switch s {
			case "DEWM_NAME":
				return dewm.Name
			case "DEWM_TYPE":
				return dewm.Type
			case "DEWM_VERSION":
				return dewm.Version
			case "DISPLAY_PROTOCOL":
				return GetDisplayProtocol()
			default:
				return ""
			}
		})
	}}
	RegisterModule(dewmModule)

	// Monitors module
	monitorsModule := StormfetchModule{stormfetchModuleConfig: stormfetchModuleConfig{Name: "monitors", Format: "%3Monitor: %4${MONITOR_WIDTH}x${MONITOR_HEIGHT} ${MONITOR_REFRESH_RATE}Hz"}, Execute: func(sm StormfetchModule) string {
		builder := strings.Builder{}
		monitors := GetMonitors()

		for i, monitor := range monitors {
			expanded := os.Expand(sm.Format, func(s string) string {
				switch s {
				case "MONITOR_NUM":
					return strconv.Itoa(i + 1)
				case "MONITOR_WIDTH":
					return strconv.Itoa(monitor.Width)
				case "MONITOR_HEIGHT":
					return strconv.Itoa(monitor.Height)
				case "MONITOR_REFRESH_RATE":
					return strconv.Itoa(monitor.RefreshRate)
				default:
					return ""
				}
			})

			builder.WriteString(expanded + "\n")
		}

		return builder.String()
	}}
	RegisterModule(monitorsModule)

	// Custom module
	customModule := StormfetchModule{stormfetchModuleConfig: stormfetchModuleConfig{Name: "custom"}, Execute: func(sm StormfetchModule) string {
		shell, _ := sm.GetData("shell", "/bin/sh")
		commandList, _ := sm.GetData("commands", make([]any, 0))

		// Exeucte all commands
		commandOutput := make(map[int]string)
		for i, value := range commandList.([]any) {
			command, ok := value.(string)
			if !ok {
				continue
			}

			commandOutput[i+1] = runCommand(command, shell.(string))
		}

		return os.Expand(sm.Format, func(s string) string {
			if len(s) <= 4 || !strings.HasPrefix(s, "CMD_") {
				return ""
			}

			commandIndexStr := strings.Split(s, "CMD_")[1]

			commandIndex, err := strconv.Atoi(commandIndexStr)
			if err != nil {
				return ""
			}

			return commandOutput[commandIndex]
		})
	}}
	RegisterModule(customModule)
}
