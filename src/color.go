package main

import (
	"fmt"
	"strings"
)

func setupColorMap(asciiArtHeader string) []string {
	colorMap := make([]string, 10)

	// Set default color map values
	for i := range 10 {
		colorMap[i] = "\033[0m"
	}

	// Return if header is empty
	if asciiArtHeader == "" {
		return colorMap
	}

	// Read colors from ascii art header
	ansiColors := strings.Split(strings.TrimPrefix(asciiArtHeader, "#/"), ";")
	for i := 0; i < 9 && i < len(ansiColors); i++ {
		colorMap[i+1] = fmt.Sprintf("\033[38;5;%sm", ansiColors[i])
	}

	return colorMap
}
