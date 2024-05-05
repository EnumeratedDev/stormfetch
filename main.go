package main

import (
	"fmt"
	yaml "gopkg.in/yaml.v3"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

var asciiPath string = ""
var configPath string = ""

var config StormfetchConfig = StormfetchConfig{}

type StormfetchConfig struct {
	Ascii string `yaml:"distro_ascii"`
	Fetch string `yaml:"fetch_script"`
}

type DistroInfo struct {
	ID        string
	LongName  string
	ShortName string
}

func main() {
	readConfig()
}

func readConfig() {
	// Get home directory
	homedir, _ := os.UserConfigDir()
	// Find valid config directory
	if _, err := os.Stat(path.Join(homedir, "stormfetch/config.yaml")); err == nil {
		configPath = path.Join(homedir, "stormfetch/config.yaml")
	} else if _, err := os.Stat("/etc/stormfetch/config.yaml"); err == nil {
		configPath = "/etc/stormfetch/config.yaml"
	} else {
		log.Fatalf("Config file not found: %s", err.Error())
	}
	// Find valid ascii directory
	if stat, err := os.Stat(path.Join(homedir, "stormfetch/ascii/")); err == nil && stat.IsDir() {
		asciiPath = path.Join(homedir, "stormfetch/ascii/")
	} else if stat, err := os.Stat("/etc/stormfetch/ascii/"); err == nil && stat.IsDir() {
		asciiPath = "/etc/stormfetch/ascii/"
	} else {
		log.Fatal("Ascii directory not found")
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

	// Write fetch script to file
	temp, err := os.CreateTemp("/tmp", "stormfetch")
	if err != nil {
		return
	}
	err = os.WriteFile(temp.Name(), []byte(config.Fetch), 644)
	if err != nil {
		return
	}
	//Execute fetch script
	cmd := exec.Command("/bin/sh", configPath)
	workdir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Dir = workdir
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "DISTRO_LONG_NAME="+getDistroInfo().LongName)
	cmd.Env = append(cmd.Env, "DISTRO_SHORT_NAME="+getDistroInfo().ShortName)
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	err = os.Remove(temp.Name())
	if err != nil {
		return
	}
	// Print Distro Information
	ascii := getDistroAscii()
	maxWidth := 0
	for _, line := range strings.Split(ascii, "\n") {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}
	for lineIndex, line := range strings.Split(ascii, "\n") {
		for i := len(line); i < maxWidth+5; i++ {
			line = line + " "
		}
		if lineIndex < len(strings.Split(string(out), "\n")) {
			line = line + strings.Split(string(out), "\n")[lineIndex]
		}
		fmt.Println(line)
	}
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
	case "debian":
		return DistroInfo{
			ID:        "debian",
			LongName:  releaseMap["PRETTY_NAME"],
			ShortName: releaseMap["NAME"],
		}
	case "ubuntu":
		return DistroInfo{
			ID:        "ubuntu",
			LongName:  releaseMap["PRETTY_NAME"],
			ShortName: releaseMap["NAME"],
		}
	case "arch":
		return DistroInfo{
			ID:        "arch",
			LongName:  releaseMap["PRETTY_NAME"],
			ShortName: releaseMap["NAME"],
		}
	default:
		return DistroInfo{
			ID:        "unknown",
			LongName:  "Unknown",
			ShortName: "Unknown",
		}
	}
}

func getDistroAscii() string {
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
	if _, err := os.Stat(path.Join(asciiPath, id)); err == nil {
		bytes, err := os.ReadFile(path.Join(asciiPath, id))
		if err != nil {
			return defaultAscii
		}
		return string(bytes)
	} else {
		return defaultAscii
	}
}
