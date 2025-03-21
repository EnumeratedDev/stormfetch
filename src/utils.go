package main

import (
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"
)

func FormatBytes(bytes uint64) string {
	var suffixes [6]string
	suffixes[0] = "B"
	suffixes[1] = "KiB"
	suffixes[2] = "MiB"
	suffixes[3] = "GiB"
	suffixes[4] = "TiB"
	suffixes[5] = "PiB"

	bf := float64(bytes)
	for _, unit := range suffixes {
		if math.Abs(bf) < 1024.0 {
			return fmt.Sprintf("%3.1f %s", bf, unit)
		}
		bf /= 1024.0
	}
	return fmt.Sprintf("%.1fYiB", bf)
}

func StripAnsii(str string) string {
	const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
	var re = regexp.MustCompile(ansi)
	return re.ReplaceAllString(str, "")
}

func ReadKeyValueFile(filepath string) (map[string]string, error) {
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
