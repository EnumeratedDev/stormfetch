package main

import (
	"fmt"
	yaml "gopkg.in/yaml.v3"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
)

var configPath = ""
var fetchScriptPath = ""

var config = StormfetchConfig{
	Ascii:             "auto",
	FetchScript:       "auto",
	AnsiiColors:       make([]int, 0),
	ForceConfigAnsii:  false,
	DependencyWarning: true,
}

type StormfetchConfig struct {
	Ascii             string `yaml:"distro_ascii"`
	FetchScript       string `yaml:"fetch_script"`
	AnsiiColors       []int  `yaml:"ansii_colors"`
	ForceConfigAnsii  bool   `yaml:"force_config_ansii"`
	DependencyWarning bool   `yaml:"dependency_warning"`
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
	// Show Dependency warning if enabled
	if config.DependencyWarning {
		dependencies := []string{"lshw", "xhost", "xdpyinfo"}
		var missing []string
		for _, depend := range dependencies {
			if _, err := os.Stat(path.Join("/usr/bin/", depend)); err != nil {
				missing = append(missing, depend)
			}
		}
		if len(missing) != 0 {
			fmt.Println("[WARNING] Stormfetch functionality may be limited due to the following dependencies not being installed:")
			for _, depend := range missing {
				fmt.Println(depend)
			}
			fmt.Println("You can disable this warning through your stormfetch config")
		}
	}
	// Fetch ascii art and apply colors
	colorMap := make(map[string]string)
	colorMap["C0"] = "C0=\033[0m"
	setColorMap := func() {
		for i, color := range config.AnsiiColors {
			if i > 6 {
				break
			}
			colorMap["C"+strconv.Itoa(i+1)] = fmt.Sprintf("\033[38;5;%dm", color)
		}
	}
	setColorMap()
	ascii := ""
	if strings.HasPrefix(getDistroAsciiArt(), "#/") {
		firstLine := strings.Split(getDistroAsciiArt(), "\n")[0]
		if !config.ForceConfigAnsii {
			ansiiColors := strings.Split(strings.TrimPrefix(firstLine, "#/"), ";")
			for i, color := range ansiiColors {
				atoi, err := strconv.Atoi(color)
				if err != nil {
					log.Fatal(err)
				}
				if i < len(config.AnsiiColors) {
					config.AnsiiColors[i] = atoi
				} else {
					config.AnsiiColors = append(config.AnsiiColors, atoi)
				}
			}
			setColorMap()
		}
		ascii = os.Expand(getDistroAsciiArt(), func(s string) string {
			return colorMap[s]
		})
		ascii = strings.TrimPrefix(ascii, firstLine+"\n")
	} else {
		ascii = os.Expand(getDistroAsciiArt(), func(s string) string {
			return colorMap[s]
		})
	}
	asciiSplit := strings.Split(ascii, "\n")
	asciiNoColor := stripAnsii(ascii)
	//Execute fetch script
	cmd := exec.Command("/bin/bash", fetchScriptPath)
	cmd.Dir = path.Dir(fetchScriptPath)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "C0=\033[0m")
	for key, value := range colorMap {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}
	cmd.Env = append(cmd.Env, "DISTRO_LONG_NAME="+getDistroInfo().LongName)
	cmd.Env = append(cmd.Env, "DISTRO_SHORT_NAME="+getDistroInfo().ShortName)
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	// Print Distro Information
	maxWidth := 0
	for _, line := range strings.Split(asciiNoColor, "\n") {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}
	final := ""
	for lineIndex, line := range asciiSplit {
		lineVisibleLength := len(strings.Split(asciiNoColor, "\n")[lineIndex])
		lastAsciiColor := ""
		if lineIndex != 0 {
			r := regexp.MustCompile("\033[38;5;[0-9]+m")
			matches := r.FindAllString(asciiSplit[lineIndex-1], -1)
			if len(matches) != 0 {
				lastAsciiColor = r.FindAllString(asciiSplit[lineIndex-1], -1)[len(matches)-1]
			}
		}
		for i := lineVisibleLength; i < maxWidth+5; i++ {
			line = line + " "
		}
		asciiSplit[lineIndex] = lastAsciiColor + line
		str := string(out)
		if lineIndex < len(strings.Split(str, "\n")) {
			line = line + strings.Split(str, "\n")[lineIndex]
		}
		final += lastAsciiColor + line + "\n"
	}
	fmt.Println(final)
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

func stripAnsii(str string) string {
	const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
	var re = regexp.MustCompile(ansi)
	return re.ReplaceAllString(str, "")
}
