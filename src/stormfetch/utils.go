package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
)

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

func fetchAmdGpuName(productId, revision string) (string, error) {
	productId = strings.ToUpper(productId)
	revision = strings.ToUpper(revision[2:])

	// Get cache directory
	cachedir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	// Ensure amdgpu.ids file exists and download it if it doesn't
	if _, err := os.Stat(path.Join(cachedir, "amdgpu.ids")); err != nil {
		cmd := exec.Command("curl", "-o", path.Join(cachedir, "amdgpu.ids"), "https://gitlab.freedesktop.org/mesa/libdrm/-/raw/main/data/amdgpu.ids")
		err = cmd.Run()
		if err != nil {
			return "", fmt.Errorf("Could not fetch amdgpu.ids using curl")
		}
	}

	// Read amdgpu.ids file
	amdgpuIds, err := os.ReadFile(path.Join(cachedir, "amdgpu.ids"))
	if err != nil {
		return "", err
	}

	// Parse read data and find GPU
	for _, line := range strings.Split(string(amdgpuIds), "\n") {
		if len(line) < 2 || line[0] == '#' {
			continue
		}

		fields := strings.Split(line, ",\t")
		if fields[0] == productId && fields[1] == revision {
			return strings.TrimPrefix(fields[2], "AMD "), nil
		}
	}

	return "", nil
}

func runCommand(command string, shell string) string {
	cmd := exec.Command(shell, "-c", command)
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
