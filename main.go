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
	Distro string `yaml:"distro_id"`
	Ascii  string `yaml:"distro_ascii"`
	Fetch  string `yaml:"fetch_script"`
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
	cmd.Environ()
	getDistroInfo()
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
	// Print Distro Ascii
	fmt.Println(getDistroAscii())
	// Print Fetch Script Output
	fmt.Println(string(out))
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
	if config.Distro == "auto" {
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
	fmt.Println(path.Join(asciiPath, getDistroInfo().ID))
	if _, err := os.Stat(path.Join(asciiPath, getDistroInfo().ID)); err == nil {
		bytes, err := os.ReadFile(path.Join(asciiPath, getDistroInfo().ID))
		if err != nil {
			return ""
		}
		return string(bytes)
	} else {
		return defaultAscii
	}
}
