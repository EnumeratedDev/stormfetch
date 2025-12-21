package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Build-time variables
var systemConfigDir = "/etc/"

// Flag variables
var ShowModuleTimeTaken = false

var configPath = ""

func main() {
	readConfig()
	parseFlags()
	initializeModuleMap()
	run()
}

func parseFlags() {
	flag.StringVar(&config.Ascii, "ascii", config.Ascii, "Set distro ascii")
	flag.StringVar(&config.DistroName, "distro-name", config.DistroName, "Set distro name")
	flag.BoolVar(&ShowModuleTimeTaken, "time-taken", false, "Show time taken to execute each module")
	flag.Parse()
}

func run() {
	// Fetch ascii art and remove header
	asciiArt := GetDistroAsciiArt()
	asciiArtHeader := ""
	if strings.HasPrefix(asciiArt, "#/") {
		asciiArtHeader = strings.SplitN(asciiArt, "\n", 2)[0]
		asciiArt = strings.SplitN(asciiArt, "\n", 2)[1]
	}
	asciiArtNoColor := asciiArt

	// Setup color map and replace colors in ascii art
	colorMap := setupColorMap(asciiArtHeader)
	for key, value := range colorMap {
		asciiArt = strings.ReplaceAll(asciiArt, "%"+strconv.Itoa(key), value)
		asciiArtNoColor = strings.ReplaceAll(asciiArtNoColor, "%"+strconv.Itoa(key), "")
	}

	// Execute modules in order
	modulesText := make([]string, 0)
	for _, moduleConfig := range config.Modules {
		module, ok := Modules[moduleConfig.Name]
		if !ok {
			continue
		}

		// Set module config options
		if moduleConfig.Format != "" {
			module.Format = moduleConfig.Format
		}
		if moduleConfig.Data != nil {
			module.Data = moduleConfig.Data
		}

		// Execute module
		start := time.Now().UnixMilli()
		text := module.Execute(module)
		end := time.Now().UnixMilli()

		// Show time taken
		if ShowModuleTimeTaken {
			fmt.Printf("Module '%s' took %d milliseconds\n", module.Name, end-start)
		}

		// Replace colors in returned string
		textNoColor := text
		for key, value := range colorMap {
			text = strings.ReplaceAll(text, "%"+strconv.Itoa(key), value)
			textNoColor = strings.ReplaceAll(textNoColor, "%"+strconv.Itoa(key), value)
		}

		// Continue if text length is 0
		if len(textNoColor) == 0 {
			continue
		}

		// Add text to slice
		for _, line := range strings.Split(strings.TrimSpace(text), "\n") {
			modulesText = append(modulesText, line)
		}
	}

	// Get longest line in ascii art
	maxWidth := 0
	for _, line := range strings.Split(asciiArtNoColor, "\n") {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	// Split ascii art into each lien
	asciiArtSplit := strings.Split(asciiArt, "\n")
	asciiArtNoColorSplit := strings.Split(asciiArtNoColor, "\n")

	// Get amount of lines to print
	lineCount := max(len(asciiArtSplit), len(modulesText))

	// Combine ascii art and module text
	final := strings.Builder{}
	for i := range lineCount {
		// Write ascii art
		currentLineLength := 0
		if i < len(asciiArtSplit) {
			final.WriteString(asciiArtSplit[i])
			currentLineLength += len(asciiArtNoColorSplit[i])
		}

		// Write blank space between ascii art and module text
		for i := currentLineLength; i < maxWidth+3; i++ {
			final.WriteString(" ")
		}

		// Write module text
		if i < len(modulesText) {
			final.WriteString(modulesText[i])
		}

		final.WriteString("\n")
	}

	fmt.Println(strings.TrimRight(final.String(), "\n") + "\033[0m")
}
