package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

type PackageManager struct {
	Name           string
	ExecutableName string
	GetPackages    func(...any) int
	FunctionInput  []any
}

var PackageManagers = []PackageManager{
	{Name: "dpkg", ExecutableName: "dpkg", GetPackages: pmFileLines, FunctionInput: []any{"/var/lib/dpkg/status", "Status: install ok installed"}},
	{Name: "pacman", ExecutableName: "pacman", GetPackages: pmDirectoryElements, FunctionInput: []any{"/var/lib/pacman/local/", true}},
	{Name: "rpm", ExecutableName: "rpm", GetPackages: pmShellCommandLines, FunctionInput: []any{"rpm -qa"}},
	{Name: "xbps", ExecutableName: "xbps-query", GetPackages: pmFileLines, FunctionInput: []any{"/var/db/xbps/pkgdb-0.38.plist", "<string>installed</string>"}},
	{Name: "bpm", ExecutableName: "bpm", GetPackages: pmDirectoryElements, FunctionInput: []any{"/var/lib/bpm/installed/"}},
	{Name: "portage", ExecutableName: "emerge", GetPackages: pmPortage},
	{Name: "flatpak", ExecutableName: "flatpak", GetPackages: pmFlatpak},
	{Name: "snap", ExecutableName: "snap", GetPackages: pmSnap},
}

func (pm *PackageManager) CountPackages() int {
	// Return 0 if package manager is not found
	if _, err := exec.LookPath(pm.ExecutableName); err != nil {
		return 0
	}

	return pm.GetPackages(pm.FunctionInput...)
}

func GetInstalledPackages() (ret string) {
	for _, pm := range PackageManagers {
		count := pm.CountPackages()
		if count > 0 {
			if ret == "" {
				ret += fmt.Sprintf("%d (%s)", count, pm.Name)
			} else {
				ret += fmt.Sprintf(" %d (%s)", count, pm.Name)
			}
		}
	}

	return ret
}

func pmDirectoryElements(args ...any) int {
	directory := args[0].(string)
	dirsOnly := false
	if len(args) >= 2 {
		dirsOnly = args[1].(bool)
	}
	total := 0

	dirEntries, _ := os.ReadDir(directory)
	for _, entry := range dirEntries {
		if !entry.IsDir() && dirsOnly {
			continue
		}
		total++
	}

	return total
}

func pmFileLines(args ...any) int {
	filepath := args[0].(string)
	mustContain := ""
	if len(args) >= 2 {
		mustContain = args[1].(string)
	}
	total := 0

	content, err := os.ReadFile(filepath)
	if err == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(content)), "\n") {
			if mustContain != "" && !strings.Contains(line, mustContain) {
				continue
			}

			total++
		}
	}

	return total
}

func pmShellCommandLines(args ...any) int {
	command := args[0].(string)

	output := runCommand(command, "/bin/sh")

	return len(strings.Split(output, "\n"))
}

func pmShellCommandOutput(args ...any) int {
	command := args[0].(string)

	output := runCommand(command, "/bin/sh")

	packageCount, _ := strconv.Atoi(output)
	return packageCount
}

func pmPortage(args ...any) int {
	portageDir := "/var/db/pkg"
	total := 0

	dirEntries, _ := os.ReadDir(portageDir)
	for _, repo := range dirEntries {
		if !repo.IsDir() {
			continue
		}

		packages, _ := os.ReadDir(path.Join(portageDir, repo.Name()))
		total += len(packages)
	}

	return total
}

func pmFlatpak(args ...any) int {
	arch := GetArch()
	flatpakDir := "/var/lib/flatpak"
	total := 0

	// Count applications
	apps, err := os.ReadDir(path.Join(flatpakDir, "app"))
	if err == nil {
		for _, app := range apps {
			if strings.HasSuffix(app.Name(), ".Locale") || strings.HasSuffix(app.Name(), ".Debug") {
				continue
			}

			dirEntries, _ := os.ReadDir(path.Join(flatpakDir, "app", app.Name(), arch))

			total += len(dirEntries)
		}
	}

	// Count runtimes
	runtimes, err := os.ReadDir(path.Join(flatpakDir, "runtime"))
	if err == nil {
		for _, runtime := range runtimes {
			if strings.HasSuffix(runtime.Name(), ".Locale") || strings.HasSuffix(runtime.Name(), ".Debug") {
				continue
			}

			dirEntries, _ := os.ReadDir(path.Join(flatpakDir, "runtime", runtime.Name(), arch))

			total += len(dirEntries)
		}
	}

	return total
}

func pmSnap(args ...any) int {
	total := pmDirectoryElements("/snap", true)
	if total > 0 {
		return total - 1
	}

	total = pmDirectoryElements("/var/lib/snapd/snap", true)
	if total > 0 {
		return total - 1
	}

	return 0
}
