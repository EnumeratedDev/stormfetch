package main

import (
	"fmt"
	"os/exec"
	"strings"
)

type PackageManager struct {
	Name               string
	ExecutableName     string
	PackageListCommand string
}

var PackageManagers = []PackageManager{
	{Name: "dpkg", ExecutableName: "dpkg", PackageListCommand: "dpkg-query -f '${Package}\\n' -W"},
	{Name: "pacman", ExecutableName: "pacman", PackageListCommand: "pacman -Q"},
	{Name: "rpm", ExecutableName: "rpm", PackageListCommand: "rpm -qa"},
	{Name: "xbps", ExecutableName: "xbps-query", PackageListCommand: "xbps-query -l"},
	{Name: "bpm", ExecutableName: "bpm", PackageListCommand: "bpm list -n"},
	{Name: "portage", ExecutableName: "emerge", PackageListCommand: "find /var/db/pkg/*/ -mindepth 1 -maxdepth 1"},
	{Name: "flatpak", ExecutableName: "flatpak", PackageListCommand: "flatpak list"},
	{Name: "snap", ExecutableName: "snap", PackageListCommand: "snap list | tail +2"},
}

func (pm *PackageManager) CountPackages() int {
	// Return 0 if package manager is not found
	if _, err := exec.LookPath(pm.ExecutableName); err != nil {
		return 0
	}

	output, err := exec.Command("/bin/sh", "-c", pm.PackageListCommand).Output()
	if err != nil {
		return 0
	}

	return strings.Count(string(output), "\n")
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
