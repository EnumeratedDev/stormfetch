package main

import (
	"log"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

type StormfetchConfig struct {
	Ascii            string                   `yaml:"distro_ascii"`
	DistroName       string                   `yaml:"distro_name"`
	Modules          []stormfetchModuleConfig `yaml:"modules"`
	AnsiiColors      []int                    `yaml:"ansii_colors"`
	ForceConfigAnsii bool                     `yaml:"force_config_ansii"`
}

var config = StormfetchConfig{
	Ascii:   "auto",
	Modules: make([]stormfetchModuleConfig, 0),
}

func readConfig() {
	// Get home directory
	userConfigDir, _ := os.UserConfigDir()

	// Find valid config directory
	if _, err := os.Stat(path.Join(userConfigDir, "stormfetch/config.yaml")); err == nil {
		configPath = path.Join(userConfigDir, "stormfetch/config.yaml")
	} else if _, err := os.Stat(path.Join(systemConfigDir, "stormfetch/config.yaml")); err == nil {
		configPath = path.Join(systemConfigDir, "stormfetch/config.yaml")
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
}
