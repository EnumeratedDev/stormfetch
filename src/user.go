package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/mitchellh/go-ps"
)

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
	case "nu":
		return "Nushell " + runCommand("$SHELL --version")
	default:
		return "Unknown"
	}
}

func GetDEWM() string {
	processes, err := ps.Processes()
	if err != nil {
		log.Fatalf("Error: could not get processes: %s", err)
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
		return "XFCE " + runCommand("xfce4-session --version | head -n1 | awk '{print $2}'")
	} else if processExists("cinnamon") {
		return "Cinnamon " + runCommand("cinnamon --version | awk '{print $3}'")
	} else if processExists("mate-panel") {
		return "MATE " + runCommand("mate-about --version | awk '{print $4}'")
	} else if processExists("lxsession") {
		return "LXDE"
	} else if processExists("lxqt-session") {
		return "LXQt " + runCommand("lxqt-session --version | head -n1 | awk '{print $2}'")
	} else if processExists("i3") || processExists("i3-with-shmlog") {
		return "i3 " + runCommand("i3 --version | awk '{print $3}'")
	} else if processExists("sway") {
		if runCommand("sway --version | awk '{print $1}'") == "swayfx" {
			return "SwayFX " + runCommand("sway --version | awk '{print $3}'")
		} else {
			return "Sway " + runCommand("sway --version | awk '{print $3}'")
		}
	} else if processExists("bspwm") {
		return "Bspwm " + runCommand("bspwm -v")
	} else if processExists("Hyprland") {
		return "Hyprland " + runCommand("hyprctl version | sed -n 3p | awk '{print $2}' | tr -d 'v,'")
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
