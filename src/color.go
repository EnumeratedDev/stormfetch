package main

import (
	"fmt"
	"strings"
)

func setupColorMap(asciiArtHeader string) map[int]string {
	colorMap := make(map[int]string)
	colorMap[0] = "\033[0m"

	// Return if header is empty
	if asciiArtHeader == "" {
		return colorMap
	}

	// Read colors from ascii art header
	ansiColors := strings.Split(strings.TrimPrefix(asciiArtHeader, "#/"), ";")
	for i, ansiColor := range ansiColors {
		colorMap[i+1] = fmt.Sprintf("\033[38;5;%sm", ansiColor)
	}

	return colorMap
}
