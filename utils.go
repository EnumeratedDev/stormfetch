package main

import (
	"bufio"
	"fmt"
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xinerama"
	"github.com/jackmordaunt/ghw"
	"github.com/mitchellh/go-ps"
	"log"
	"math"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"syscall"
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
		releaseMap, err = ReadKeyValueFile("/etc/os-release")
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
	for _, graphics := range gpu.GraphicsCards {
		if graphics.DeviceInfo != nil {
			return graphics.DeviceInfo.Product.Name
		}
	}
	return ""
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
	if GetDisplayProtocol() == "X11" {
		conn, err := xgb.NewConn()
		if err != nil {
			return nil
		}
		err = xinerama.Init(conn)
		if err != nil {
			return nil
		}
		reply, _ := xinerama.QueryScreens(conn).Reply()
		conn.Close()
		for _, screen := range reply.ScreenInfo {
			monitors = append(monitors, strconv.Itoa(int(screen.Width))+"x"+strconv.Itoa(int(screen.Height)))
		}
	}
	return monitors
}

type partition struct {
	Device     string
	MountPoint string
	TotalSize  uint64
	UsedSize   uint64
	FreeSize   uint64
}

func getMountedPartitions() []partition {
	entries, err := os.ReadDir("/dev/disk/by-partuuid")
	if err != nil {
		return nil
	}

	bytes, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return nil
	}
	mountsRaw := strings.Split(string(bytes), "\n")
	mounts := make(map[string]string)
	for i := 0; i < len(mountsRaw); i++ {
		split := strings.Split(mountsRaw[i], " ")
		if len(split) <= 2 {
			continue
		}
		mounts[split[0]] = split[1]
	}

	var partitions []partition
	for _, entry := range entries {
		link, err := filepath.EvalSymlinks(filepath.Join("/dev/disk/by-partuuid/", entry.Name()))
		if err != nil {
			continue
		}
		if _, ok := mounts[link]; !ok {
			continue
		}
		p := partition{
			link,
			mounts[link],
			0,
			0,
			0,
		}
		buf := new(syscall.Statfs_t)
		err = syscall.Statfs(p.MountPoint, buf)
		if err != nil {
			log.Fatal(err)
		}
		totalBlocks := buf.Blocks
		freeBlocks := buf.Bfree
		usedBlocks := totalBlocks - freeBlocks
		blockSize := uint64(buf.Bsize)

		p.TotalSize = totalBlocks * blockSize
		p.FreeSize = freeBlocks * blockSize
		p.UsedSize = usedBlocks * blockSize

		partitions = append(partitions, p)
	}
	return partitions
}

func FormatBytes(bytes uint64) string {
	var suffixes [6]string
	suffixes[0] = "B"
	suffixes[1] = "KiB"
	suffixes[2] = "MiB"
	suffixes[3] = "GiB"
	suffixes[4] = "TiB"
	suffixes[5] = "PiB"

	bf := float64(bytes)
	for _, unit := range suffixes {
		if math.Abs(bf) < 1024.0 {
			return fmt.Sprintf("%3.1f %s", bf, unit)
		}
		bf /= 1024.0
	}
	return fmt.Sprintf("%.1fYiB", bf)
}

func StripAnsii(str string) string {
	const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
	var re = regexp.MustCompile(ansi)
	return re.ReplaceAllString(str, "")
}

func ReadKeyValueFile(filepath string) (map[string]string, error) {
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
