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

var configPath = ""
var fetchScriptPath = ""

var config = StormfetchConfig{}

type StormfetchConfig struct {
	Ascii       string `yaml:"distro_ascii"`
	FetchScript string `yaml:"fetch_script"`
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
	configDir, _ := os.UserConfigDir()
	// Find valid config directory
	if _, err := os.Stat(path.Join(configDir, "stormfetch/config.yaml")); err == nil {
		configPath = path.Join(configDir, "stormfetch/config.yaml")
	} else if _, err := os.Stat("/etc/stormfetch/config.yaml"); err == nil {
		configPath = "/etc/stormfetch/config.yaml"
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
	if _, err := os.Stat(path.Join(configDir, "stormfetch/fetch_script.sh")); err == nil {
		fetchScriptPath = path.Join(configDir, "stormfetch/fetch_script.sh")
	} else if _, err := os.Stat("/etc/stormfetch/fetch_script.sh"); err == nil {
		fetchScriptPath = "/etc/stormfetch/fetch_script.sh"
	} else {
		log.Fatalf("Fetch script file not found: %s", err.Error())
	}
	//Execute fetch script
	cmd := exec.Command("/bin/bash", fetchScriptPath)
	cmd.Dir = path.Dir(fetchScriptPath)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "DISTRO_LONG_NAME="+getDistroInfo().LongName)
	cmd.Env = append(cmd.Env, "DISTRO_SHORT_NAME="+getDistroInfo().ShortName)
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
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
