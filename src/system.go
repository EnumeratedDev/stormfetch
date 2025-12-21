package main

import (
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"

	"github.com/mitchellh/go-ps"
)

type DistroInfo struct {
	ID        string
	LongName  string
	ShortName string
}

func GetDistroInfo() DistroInfo {
	info := DistroInfo{
		ID:        "unknown",
		LongName:  "Unknown",
		ShortName: "Unknown",
	}
	if strings.TrimSpace(config.DistroName) != "" {
		info.LongName = strings.TrimSpace(config.DistroName)
		info.ShortName = strings.TrimSpace(config.DistroName)
	}

	// Detect release file location
	var releaseFile string
	if _, err := os.Stat("/bedrock/etc/os-release"); os.Getenv("BEDROCK_RESTRICT") == "" && err == nil {
		// Using Bedrock Linux
		releaseFile = "/bedrock/etc/os-release"
	} else if _, err := os.Stat("/etc/os-release"); err == nil {
		// Using a regular linux distribution
		releaseFile = "/etc/os-release"
	} else {
		return info
	}

	releaseMap, err := ReadKeyValueFile(releaseFile)
	if err != nil {
		return info
	}

	if id, ok := releaseMap["ID"]; ok {
		info.ID = id
	}
	if longName, ok := releaseMap["PRETTY_NAME"]; ok && info.LongName == "Unknown" {
		info.LongName = longName
	}
	if shortName, ok := releaseMap["NAME"]; ok && info.ShortName == "Unknown" {
		info.ShortName = shortName
	}
	return info
}

func GetDistroAsciiArt() string {
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
		id = GetDistroInfo().ID
	} else {
		id = config.Ascii
	}
	userConfDir, err := os.UserConfigDir()
	if err != nil {
		if _, err := os.Stat(path.Join(systemConfigDir, "stormfetch/ascii/", id)); err == nil {
			bytes, err := os.ReadFile(path.Join(systemConfigDir, "stormfetch/ascii/", id))
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
	} else if _, err := os.Stat(path.Join(systemConfigDir, "stormfetch/ascii/", id)); err == nil {
		bytes, err := os.ReadFile(path.Join(systemConfigDir, "stormfetch/ascii/", id))
		if err != nil {
			return defaultAscii
		}
		return strings.TrimRight(string(bytes), "\n\t ")
	} else {
		return defaultAscii
	}
}

func GetArch() string {
	uname := syscall.Utsname{}
	err := syscall.Uname(&uname)
	if err != nil {
		return "unknown"
	}

	var byteString [65]byte
	var indexLength int
	for ; uname.Machine[indexLength] != 0; indexLength++ {
		byteString[indexLength] = uint8(uname.Machine[indexLength])
	}
	return string(byteString[:indexLength])
}

func GetKernel() (string, string) {
	uname := syscall.Utsname{}
	err := syscall.Uname(&uname)
	if err != nil {
		return "unknown", "unknown"
	}

	var kernelNameByteString [65]byte
	var kernelNameLength int
	for ; uname.Sysname[kernelNameLength] != 0; kernelNameLength++ {
		kernelNameByteString[kernelNameLength] = uint8(uname.Sysname[kernelNameLength])
	}

	var kernelReleaseByteString [65]byte
	var kernelReleaseLength int
	for ; uname.Release[kernelReleaseLength] != 0; kernelReleaseLength++ {
		kernelReleaseByteString[kernelReleaseLength] = uint8(uname.Release[kernelReleaseLength])
	}

	return string(kernelNameByteString[:kernelNameLength]), string(kernelReleaseByteString[:kernelReleaseLength])
}

func GetInitSystem() string {
	runCommand := func(command string) string {
		cmd := exec.Command("/bin/sh", "-c", command)
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

	process, err := ps.FindProcess(1)
	if err != nil {
		return ""
	}

	// Special cases
	// OpenRC check
	if _, err := os.Stat("/usr/sbin/openrc"); err == nil {
		return "OpenRC " + runCommand("openrc --version | awk '{print $3}'")
	}

	// Default PID 1 process name checking
	switch process.Executable() {
	case "systemd":
		return "Systemd " + runCommand("systemctl --version | head -n1 | awk '{print $2}'")
	case "runit":
		return "Runit"
	case "dinit":
		return "Dinit " + runCommand("dinit --version | head -n1 | awk '{print substr($3, 1, length($3)-1)}'")
	case "enit":
		return "Enit " + runCommand("enit --version | awk '{print $3}'")
	default:
		return process.Executable()
	}
}

func GetLibc() string {
	checkLibcOutput, err := exec.Command("ldd", "/usr/bin/ls").Output()
	if err != nil {
		return "Unknown"
	}

	if strings.Contains(string(checkLibcOutput), "ld-musl") {
		// Using Musl Libc
		output, _ := exec.Command("ldd").CombinedOutput()
		return "Musl " + strings.TrimPrefix(strings.Split(strings.TrimSpace(string(output)), "\n")[1], "Version ")
	} else {
		// Using Glibc
		cmd := exec.Command("ldd", "--version")
		output, err := cmd.Output()
		if err != nil {
			return "Glibc"
		}
		outputSplit := strings.Split(strings.Split(strings.TrimSpace(string(output)), "\n")[0], " ")
		ver := outputSplit[len(outputSplit)-1]
		return "Glibc " + ver
	}
}
