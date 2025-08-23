package main

import (
	"fmt"
	"strconv"
	"strings"
)

func setupColorMap(asciiArt string) map[string]string {
	colorMap := make(map[string]string)

	// Read colors from ascii art
	if strings.HasPrefix(asciiArt, "#/") {
		firstLine := strings.Split(asciiArt, "\n")[0]
		ansiColors := strings.Split(strings.TrimPrefix(firstLine, "#/"), ";")

		for i, ansiColor := range ansiColors {
			colorMap["C"+strconv.Itoa(i+1)] = fmt.Sprintf("\033[38;5;%sm", ansiColor)
		}
	}

	return colorMap
}
