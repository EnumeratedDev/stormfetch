package main

import (
	"log"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

type StormfetchConfig struct {
	Ascii                   string                   `yaml:"distro_ascii"`
	DisableAmdgpuIdsWarning bool                     `yaml:"disable_amdgpu_ids_warning"`
	Modules                 []stormfetchModuleConfig `yaml:"modules"`
	AnsiiColors             []int                    `yaml:"ansii_colors"`
	ForceConfigAnsii        bool                     `yaml:"force_config_ansii"`
}

var config = StormfetchConfig{
	Ascii:   "auto",
	Modules: make([]stormfetchModuleConfig, 0),
}

func readConfig() {
	if ConfigPath == "" {
		// Get home directory
		userConfigDir, _ := os.UserConfigDir()

		// Find valid config directory
		if _, err := os.Stat(path.Join(userConfigDir, "stormfetch/config.yml")); err == nil {
			ConfigPath = path.Join(userConfigDir, "stormfetch/config.yml")
		} else if _, err := os.Stat(path.Join(SystemConfigDir, "stormfetch/config.yml")); err == nil {
			ConfigPath = path.Join(SystemConfigDir, "stormfetch/config.yml")
		} else {
			log.Fatalf("Config file not found: %s", err.Error())
		}
	}

	// Parse config
	bytes, err := os.ReadFile(ConfigPath)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		log.Fatal(err)
	}
}
