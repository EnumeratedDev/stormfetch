package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var systemConfigDir = "/etc/"

var configPath = ""
var fetchScriptPath = ""

var TimeTaken = false

var config = StormfetchConfig{
	Ascii:            "auto",
	FetchScript:      "auto",
	ShowFSType:       false,
	HiddenPartitions: make([]string, 0),
	HiddenGPUS:       make([]int, 0),
}

type StormfetchConfig struct {
	Ascii             string   `yaml:"distro_ascii"`
	DistroName        string   `yaml:"distro_name"`
	FetchScript       string   `yaml:"fetch_script"`
	AnsiiColors       []int    `yaml:"ansii_colors"`
	ForceConfigAnsii  bool     `yaml:"force_config_ansii"`
	ShowFSType        bool     `yaml:"show_fs_type"`
	HiddenPartitions  []string `yaml:"hidden_partitions"`
	HiddenFilesystems []string `yaml:"hidden_filesystems"`
	HiddenGPUS        []int    `yaml:"hidden_gpus"`
}

func main() {
	readConfig()
	readFlags()
	runStormfetch()
}

func readConfig() {
	// Get home directory
	userConfigDir, _ := os.UserConfigDir()
	// Find valid config directory
	if _, err := os.Stat(path.Join(userConfigDir, "stormfetch/config.yaml")); err == nil {
		configPath = path.Join(userConfigDir, "stormfetch/config.yaml")
	} else if _, err := os.Stat(path.Join(systemConfigDir, "stormfetch/config.yaml")); err == nil {
		configPath = path.Join(systemConfigDir, "stormfetch/config.yaml")
	} else {
		log.Fatalf("Config file not found: %s", err.Error())
	}
	// Parse config
	bytes, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		log.Fatal(err)
	}
	if config.FetchScript == "" {
		log.Fatalf("Fetch script path is empty")
	} else if config.FetchScript != "auto" {
		stat, err := os.Stat(config.FetchScript)
		if err != nil {
			log.Fatalf("Fetch script file not found: %s", err.Error())
		} else if stat.IsDir() {
			log.Fatalf("Fetch script path points to a directory")
		}
	}
	if _, err := os.Stat(path.Join(userConfigDir, "stormfetch/fetch_script.sh")); err == nil {
		fetchScriptPath = path.Join(userConfigDir, "stormfetch/fetch_script.sh")
	} else if _, err := os.Stat(path.Join(systemConfigDir, "stormfetch/fetch_script.sh")); err == nil {
		fetchScriptPath = path.Join(systemConfigDir, "stormfetch/fetch_script.sh")
	} else {
		log.Fatalf("Fetch script file not found: %s", err.Error())
	}
}

func readFlags() {
	flag.StringVar(&config.Ascii, "ascii", config.Ascii, "Set distro ascii")
	flag.StringVar(&config.DistroName, "distro-name", config.DistroName, "Set distro name")
	flag.BoolVar(&TimeTaken, "time-taken", false, "Show time taken for fetched information")
	flag.Parse()
}

func SetupFetchEnv(showTimeTaken bool) []string {
	var env = make(map[string]string)
	setVariable := func(key string, setter func() string) {
		start := time.Now().UnixMilli()
		env[key] = setter()
		end := time.Now().UnixMilli()
		if showTimeTaken {
			fmt.Printf("Setting '%s' took %d milliseconds\n", key, end-start)
		}
	}
	setVariable("PACKAGES", func() string { return GetInstalledPackages() })
	setVariable("DISTRO_LONG_NAME", func() string { return GetDistroInfo().LongName })
	setVariable("DISTRO_SHORT_NAME", func() string { return GetDistroInfo().ShortName })
	setVariable("CPU_MODEL", func() string { return GetCPUModel() })
	setVariable("MOTHERBOARD", func() string { return GetMotherboardModel() })
	setVariable("CPU_THREADS", func() string { return strconv.Itoa(GetCPUThreads()) })
	start := time.Now().UnixMilli()
	memory := GetMemoryInfo()
	end := time.Now().UnixMilli()
	if showTimeTaken {
		fmt.Printf("Setting '%s' took %d milliseconds\n", "MEM_*", end-start)
	}
	if memory != nil {
		env["MEM_TOTAL"] = strconv.Itoa(memory.MemTotal)
		env["MEM_USED"] = strconv.Itoa(memory.MemTotal - memory.MemAvailable)
		env["MEM_FREE"] = strconv.Itoa(memory.MemAvailable)
	}
	start = time.Now().UnixMilli()
	partitions := GetMountedPartitions(config.HiddenPartitions, config.HiddenFilesystems)
	end = time.Now().UnixMilli()
	if showTimeTaken {
		fmt.Printf("Setting '%s' took %d milliseconds\n", "PARTITION_*", end-start)
	}
	if len(partitions) != 0 {
		env["MOUNTED_PARTITIONS"] = strconv.Itoa(len(partitions))
		for i, part := range partitions {
			env["PARTITION"+strconv.Itoa(i+1)+"_DEVICE"] = part.Device
			env["PARTITION"+strconv.Itoa(i+1)+"_MOUNTPOINT"] = part.MountPoint
			if part.Label != "" {
				env["PARTITION"+strconv.Itoa(i+1)+"_LABEL"] = part.Label
			}
			if part.FileystemType != "" && config.ShowFSType {
				env["PARTITION"+strconv.Itoa(i+1)+"_TYPE"] = part.FileystemType
			}
			env["PARTITION"+strconv.Itoa(i+1)+"_TOTAL_SIZE"] = FormatBytes(part.TotalSize)
			env["PARTITION"+strconv.Itoa(i+1)+"_USED_SIZE"] = FormatBytes(part.UsedSize)
			env["PARTITION"+strconv.Itoa(i+1)+"_FREE_SIZE"] = FormatBytes(part.FreeSize)
		}
	}
	setVariable("DE_WM", func() string { return GetDEWM() })
	setVariable("USER_SHELL", func() string { return GetShell() })
	setVariable("DISPLAY_PROTOCOL", func() string { return GetDisplayProtocol() })
	setVariable("LIBC", func() string { return GetLibc() })
	setVariable("INIT_SYSTEM", func() string { return GetInitSystem() })
	setVariable("LOCAL_IPV4", func() string { return GetLocalIP() })
	start = time.Now().UnixMilli()
	monitors := GetMonitorResolution()
	end = time.Now().UnixMilli()
	if showTimeTaken {
		fmt.Printf("Setting '%s' took %d milliseconds\n", "MONITOR_*", end-start)
	}
	if len(monitors) != 0 {
		env["CONNECTED_MONITORS"] = strconv.Itoa(len(monitors))
		for i, monitor := range monitors {
			env["MONITOR"+strconv.Itoa(i+1)] = monitor
		}
	}
	start = time.Now().UnixMilli()
	gpus := GetGPUModels()
	end = time.Now().UnixMilli()
	if showTimeTaken {
		fmt.Printf("Setting '%s' took %d milliseconds\n", "GPU_*", end-start)
	}
	if len(gpus) != 0 {
		env["CONNECTED_GPUS"] = strconv.Itoa(len(gpus))
		for i, gpu := range gpus {
			if gpu == "" {
				continue
			}
			env["GPU"+strconv.Itoa(i+1)] = gpu
		}
	}

	var ret = make([]string, len(env))
	i := 0
	for key, value := range env {
		ret[i] = fmt.Sprintf("%s=%s", key, value)
		i++
	}
	return ret
}

func runStormfetch() {
	// Fetch ascii art
	asciiArt := GetDistroAsciiArt()

	// Setup color map
	colorMap := setupColorMap(asciiArt)
	if len(colorMap) > 0 {
		asciiArt = os.Expand(asciiArt, func(s string) string {
			return colorMap[s]
		})

		asciiArt = strings.SplitN(asciiArt, "\n", 2)[1]
	}

	asciiSplit := strings.Split(asciiArt, "\n")
	asciiNoColor := StripAnsii(asciiArt)

	// Execute fetch script
	cmd := exec.Command("/bin/bash", fetchScriptPath)
	cmd.Dir = path.Dir(fetchScriptPath)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, SetupFetchEnv(TimeTaken)...)
	cmd.Env = append(cmd.Env, "C0=\033[0m")
	for key, value := range colorMap {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}
	out, err := cmd.Output()
	if err != nil {
		log.Fatalf("Error: Could not run fetch script: %s", err)
	}
	// Print Distro Information
	maxWidth := 0
	for _, line := range strings.Split(asciiNoColor, "\n") {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}
	final := ""
	y := len(asciiSplit)
	if len(asciiSplit) < len(strings.Split(string(out), "\n")) {
		y = len(strings.Split(string(out), "\n"))
	}
	for lineIndex := 0; lineIndex < y; lineIndex++ {
		line := ""
		for i := 0; i < maxWidth+5; i++ {
			line = line + " "
		}
		lastAsciiColor := ""
		if lineIndex < len(asciiSplit) {
			line = asciiSplit[lineIndex]
			lineVisibleLength := len(strings.Split(asciiNoColor, "\n")[lineIndex])
			if lineIndex != 0 {
				r := regexp.MustCompile("\033[38;5;[0-9]+m")
				matches := r.FindAllString(asciiSplit[lineIndex-1], -1)
				if len(matches) != 0 {
					lastAsciiColor = r.FindAllString(asciiSplit[lineIndex-1], -1)[len(matches)-1]
				}
			}
			for i := lineVisibleLength; i < maxWidth+5; i++ {
				line = line + " "
			}
			asciiSplit[lineIndex] = lastAsciiColor + line
		}
		str := string(out)
		if lineIndex < len(strings.Split(str, "\n")) {
			line = line + colorMap["C0"] + strings.Split(str, "\n")[lineIndex]
		}
		final += lastAsciiColor + line + "\n"
	}
	final = strings.TrimRight(final, "\n\t ")
	fmt.Println(final + "\033[0m")
}
