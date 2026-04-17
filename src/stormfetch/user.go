package main

import (
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/mitchellh/go-ps"
)

type DEWM struct {
	Name string
	Type string
}

func GetShell() string {
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
	shellName := "Unknown"
	shellVersion := ""
	switch filepath.Base(shell) {
	case "bash":
		shellName = "Bash"
		shellVersion = runCommand("$SHELL --version | head -n1 | awk '{print $4}'", "/bin/sh")
	case "zsh":
		shellName = "Zsh"
		shellVersion = runCommand("$SHELL --version | awk '{print $2}'", "/bin/sh")
	case "fish":
		shellName = "Fish"
		shellVersion = runCommand("$SHELL --version | awk '{print $3}'", "/bin/sh")
	case "nu":
		shellName = "Nushell"
		shellVersion = runCommand("$SHELL --version", "/bin/sh")
	default:
		return "Unknown"
	}

	if shellVersion != "" {
		return shellName + " " + shellVersion
	} else {
		return shellName
	}
}

func GetDEWM() DEWM {
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
	if processExists("plasmashell") {
		dewm := DEWM{
			Name: "KDE Plasma",
			Type: "DE",
		}

		if version := runCommand("plasmashell --version | awk '{print $2}'", "/bin/sh"); version != "" {
			dewm.Name += " " + version
		}

		return dewm
	} else if processExists("gnome-session") {
		dewm := DEWM{
			Name: "Gnome",
			Type: "DE",
		}

		if version := runCommand("gnome-shell --version | awk '{print $3}'", "/bin/sh"); version != "" {
			dewm.Name += " " + version
		}

		return dewm
	} else if processExists("xfce4-session") {
		dewm := DEWM{
			Name: "XFCE",
			Type: "DE",
		}

		if version := runCommand("xfce4-session --version | head -n1 | awk '{print $2}'", "/bin/sh"); version != "" {
			dewm.Name += " " + version
		}

		return dewm
	} else if processExists("cinnamon") {
		dewm := DEWM{
			Name: "Cinnamon",
			Type: "DE",
		}

		if version := runCommand("cinnamon --version | awk '{print $3}'", "/bin/sh"); version != "" {
			dewm.Name += " " + version
		}

		return dewm
	} else if processExists("mate-panel") {
		dewm := DEWM{
			Name: "MATE",
			Type: "DE",
		}

		if version := runCommand("mate-about --version | awk '{print $4}'", "/bin/sh"); version != "" {
			dewm.Name += " " + version
		}

		return dewm
	} else if processExists("lxsession") {
		dewm := DEWM{
			Name: "LXDE",
			Type: "DE",
		}

		return dewm
	} else if processExists("lxqt-session") {
		dewm := DEWM{
			Name: "LXQt",
			Type: "DE",
		}

		if version := runCommand("lxqt-session --version | head -n1 | awk '{print $2}'", "/bin/sh"); version != "" {
			dewm.Name += " " + version
		}

		return dewm
	} else if processExists("i3") || processExists("i3-with-shmlog") {
		dewm := DEWM{
			Name: "i3",
			Type: "WM",
		}

		if version := runCommand("i3 --version | awk '{print $3}'", "/bin/sh"); version != "" {
			dewm.Name += " " + version
		}

		return dewm
	} else if processExists("sway") {
		dewm := DEWM{
			Name: "Sway",
			Type: "WM",
		}

		if runCommand("sway --version | awk '{print $1}'", "/bin/sh") == "swayfx" {
			dewm.Name = "SwayFX"
		} else {
			dewm.Name = "Sway"
		}

		if version := runCommand("sway --version | awk '{print $3}'", "/bin/sh"); version != "" {
			dewm.Name += " " + version
		}

		return dewm
	} else if processExists("bspwm") {
		dewm := DEWM{
			Name: "Bspwm",
			Type: "WM",
		}

		if version := runCommand("bspwm -v", "/bin/sh"); version != "" {
			dewm.Name += " " + version
		}

		return dewm
	} else if processExists("Hyprland") {
		dewm := DEWM{
			Name: "Hyprland",
			Type: "WM",
		}

		if version := runCommand("hyprctl version | sed -n 3p | awk '{print $2}' | tr -d 'v,'", "/bin/sh"); version != "" {
			dewm.Name += " " + version
		}

		return dewm
	} else if processExists("icewm-session") {
		dewm := DEWM{
			Name: "IceWM",
			Type: "WM",
		}

		if version := runCommand("icewm --version | awk '{print $2}'", "/bin/sh"); version != "" {
			dewm.Name += " " + version
		}

		return dewm
	}
	return DEWM{
		Name: "Unknown",
		Type: "Unknown",
	}
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
