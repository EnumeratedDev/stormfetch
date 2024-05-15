package main

import (
	"bufio"
	"fmt"
	"github.com/jackmordaunt/ghw"
	"github.com/mitchellh/go-ps"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type DistroInfo struct {
	ID        string
	LongName  string
	ShortName string
}

func getDistroInfo() DistroInfo {
	distroID := ""
	var releaseMap = make(map[string]string)
	if _, err := os.Stat("/etc/os-release"); err == nil {
		releaseMap, err = readKeyValueFile("/etc/os-release")
		if err != nil {
			return DistroInfo{
				ID:        "unknown",
				LongName:  "Unknown",
				ShortName: "Unknown",
			}
		}
		if value, ok := releaseMap["ID"]; ok {
			distroID = value
		}
	}

	switch distroID {
	default:
		if id, ok := releaseMap["ID"]; ok {
			if longName, ok := releaseMap["PRETTY_NAME"]; ok {
				if shortName, ok := releaseMap["NAME"]; ok {
					return DistroInfo{
						ID:        id,
						LongName:  longName,
						ShortName: shortName,
					}
				}
			}
		}
		return DistroInfo{
			ID:        "unknown",
			LongName:  "Unknown",
			ShortName: "Unknown",
		}
	}
}

func getDistroAsciiArt() string {
	defaultAscii :=
		`    .--.
   |o_o |
   |:_/ |
  //   \ \
 (|     | )
/'\_   _/'\
\___)=(___/ `
	var id string
	if config.Ascii == "auto" {
		id = getDistroInfo().ID
	} else {
		id = config.Ascii
	}
	userConfDir, err := os.UserConfigDir()
	if err != nil {
		if _, err := os.Stat(path.Join("/etc/stormfetch/ascii/", id)); err == nil {
			bytes, err := os.ReadFile(path.Join("/etc/stormfetch/ascii/", id))
			if err != nil {
				return defaultAscii
			}
			return string(bytes)
		} else {
			return defaultAscii
		}
	}
	if _, err := os.Stat(path.Join(userConfDir, "stormfetch/ascii/", id)); err == nil {
		bytes, err := os.ReadFile(path.Join(userConfDir, "stormfetch/ascii/", id))
		if err != nil {
			return defaultAscii
		}
		return string(bytes)
	} else if _, err := os.Stat(path.Join("/etc/stormfetch/ascii/", id)); err == nil {
		bytes, err := os.ReadFile(path.Join("/etc/stormfetch/ascii/", id))
		if err != nil {
			return defaultAscii
		}
		return string(bytes)
	} else {
		return defaultAscii
	}
}

func getCPUName() string {
	cpu, err := ghw.CPU()
	if err != nil {
		return ""
	}
	if len(cpu.Processors) == 0 {
		return ""
	}
	return cpu.Processors[0].Model
}

func getCPUThreads() int {
	cpu, err := ghw.CPU()
	if err != nil {
		return 0
	}
	return int(cpu.TotalThreads)
}

func getGPUName() string {
	null, _ := os.Open(os.DevNull)
	serr := os.Stderr
	os.Stderr = null
	gpu, err := ghw.GPU()
	defer null.Close()
	os.Stderr = serr
	if err != nil {
		return ""
	}
	if len(gpu.GraphicsCards) == 0 {
		return ""
	}
	return gpu.GraphicsCards[0].DeviceInfo.Product.Name
}

type Memory struct {
	MemTotal     int
	MemFree      int
	MemAvailable int
}

func GetMemoryInfo() *Memory {
	toInt := func(raw string) int {
		if raw == "" {
			return 0
		}
		res, err := strconv.Atoi(raw)
		if err != nil {
			panic(err)
		}
		return res
	}

	parseLine := func(raw string) (key string, value int) {
		text := strings.ReplaceAll(raw[:len(raw)-2], " ", "")
		keyValue := strings.Split(text, ":")
		return keyValue[0], toInt(keyValue[1])
	}

	if _, err := os.Stat("/proc/meminfo"); err != nil {
		return nil
	}
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	bufio.NewScanner(file)
	scanner := bufio.NewScanner(file)
	res := Memory{}
	for scanner.Scan() {
		key, value := parseLine(scanner.Text())
		switch key {
		case "MemTotal":
			res.MemTotal = value / 1024
		case "MemFree":
			res.MemFree = value / 1024
		case "MemAvailable":
			res.MemAvailable = value / 1024
		}
	}
	return &res
}

func GetShell() string {
	runCommand := func(command string) string {
		cmd := exec.Command("/bin/bash", "-c", command)
		workdir, err := os.Getwd()
		if err != nil {
			return ""
		}
		cmd.Dir = workdir
		cmd.Env = os.Environ()
		out, err := cmd.Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	}
	file, err := os.ReadFile("/etc/passwd")
	if err != nil {
		return ""
	}
	str := string(file)
	shell := ""

	for _, line := range strings.Split(str, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		userInfo := strings.Split(line, ":")
		if userInfo[2] == strconv.Itoa(os.Getuid()) {
			shell = userInfo[6]
		}
	}
	shellName := filepath.Base(shell)
	switch shellName {
	case "dash":
		return "Dash"
	case "bash":
		return "Bash " + runCommand("echo $BASH_VERSION")
	case "zsh":
		return "Zsh " + runCommand("$SHELL --version | awk '{print $2}'")
	case "fish":
		return "Fish " + runCommand("$SHELL --version | awk '{print $3}'")
	default:
		return "Unknown"
	}
}

func GetDEWM() string {
	processes, err := ps.Processes()
	if err != nil {
		log.Fatal(err)
	}
	var executables []string
	for _, process := range processes {
		executables = append(executables, process.Executable())
	}

	processExists := func(process string) bool {
		return slices.Contains(executables, process)
	}
	runCommand := func(command string) string {
		cmd := exec.Command("/bin/bash", "-c", command)
		workdir, err := os.Getwd()
		if err != nil {
			return ""
		}
		cmd.Dir = workdir
		cmd.Env = os.Environ()
		out, err := cmd.Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	}

	if processExists("plasmashell") {
		return "KDE Plasma " + runCommand("plasmashell --version | awk '{print $2}'")
	} else if processExists("gnome-session") {
		return "Gnome " + runCommand("gnome-shell --version | awk '{print $3}'")
	} else if processExists("xfce4-session") {
		return "XFCE " + runCommand("xfce4-session --version | grep xfce4-session | awk '{print $2}'")
	} else if processExists("cinnamon") {
		return "Cinnamon " + runCommand("cinnamon --version | awk '{print $3}'")
	} else if processExists("mate-panel") {
		return "MATE " + runCommand("mate-about --version | awk '{print $4}'")
	} else if processExists("lxsession") {
		return "LXDE"
	} else if processExists("sway") {
		return "Sway " + runCommand("sway --version | awk '{print $3}'")
	} else if processExists("bspwm") {
		return "Bspwm " + runCommand("bspwm -v")
	} else if processExists("icewm-session") {
		return "IceWM " + runCommand("icewm --version | awk '{print $2}'")
	}
	return ""
}

func GetDisplayProtocol() string {
	protocol := os.Getenv("XDG_SESSION_TYPE")
	if protocol == "x11" {
		return "X11"
	} else if protocol == "wayland" {
		return "Wayland"
	}
	return ""
}

func getMonitorResolution() []string {
	var monitors []string
	runCommand := func(command string) string {
		cmd := exec.Command("/bin/bash", "-c", command)
		workdir, err := os.Getwd()
		if err != nil {
			return ""
		}
		cmd.Dir = workdir
		cmd.Env = os.Environ()
		out, err := cmd.Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	}

	if GetDisplayProtocol() == "X11" {
		if _, err := os.Stat("/usr/bin/xrandr"); err != nil {
			return monitors
		}
		connections := strings.Split(runCommand("xrandr --query | grep -w \"connected\" | awk '{print $1}'"), "\n")
		for i, con := range connections {
			Xaxis := runCommand(fmt.Sprintf("xrandr --current | grep -m%d '*' | tail -n1 | uniq | awk '{print $1}' | cut -d 'x' -f1", i+1))
			Yaxis := runCommand(fmt.Sprintf("xrandr --current | grep -m%d '*' | tail -n1 | uniq | awk '{print $1}' | cut -d 'x' -f2", i+1))
			monitors = append(monitors, con+" ("+Xaxis+"x"+Yaxis+")")
		}
	}
	return monitors
}

func stripAnsii(str string) string {
	const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
	var re = regexp.MustCompile(ansi)
	return re.ReplaceAllString(str, "")
}

func readKeyValueFile(filepath string) (map[string]string, error) {
	ret := make(map[string]string)
	if _, err := os.Stat(filepath); err != nil {
		return nil, err
	}
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	str := string(bytes)
	lines := strings.Split(str, "\n")
	for _, line := range lines {
		if len(strings.Split(line, "=")) >= 2 {
			key := strings.SplitN(line, "=", 2)[0]
			value := strings.SplitN(line, "=", 2)[1]
			if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				value = value[1 : len(value)-1]
			}
			ret[key] = value
		}
	}
	return ret, nil
}
