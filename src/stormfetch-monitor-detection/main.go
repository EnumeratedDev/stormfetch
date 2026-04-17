package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-gl/glfw/v3.3/glfw"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			os.Exit(1)
		}
	}()

	err := glfw.Init()
	if err != nil {
		os.Exit(1)
	}

	for _, monitor := range glfw.GetMonitors() {
		mode := monitor.GetVideoMode()

		fmt.Print(monitor.GetName() + ",")
		fmt.Print(strconv.Itoa(mode.Width) + ",")
		fmt.Print(strconv.Itoa(mode.Height) + ",")
		fmt.Print(strconv.Itoa(mode.RefreshRate))
		fmt.Println()
	}
	defer glfw.Terminate()
}
