package main

import (
	"embed"
	"os/exec"
	"runtime"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
)

//go:embed all:frontend/dist
var assets embed.FS

// The following program implements a proxy that forwards players from one local address to a remote address.
func main() {
	// Run the loopback excempt command
	checkNetIsolation()
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:         "Tedac",
		Width:         905,
		Height:        500,
		Frameless:     true,
		DisableResize: true,
		Assets:        assets,
		OnStartup:     app.startup,
		Bind:          []any{app},
	})
	if err != nil {
		panic(err)
	}
}

// checkNetIsolation checks if a loopback exempt is in place to allow the
// hosting device to join the server. This is only relevant on Windows.
func checkNetIsolation() {
	if runtime.GOOS != "windows" {
		return
	}
	data, _ := exec.Command("CheckNetIsolation", "LoopbackExempt", "-s", `-n="microsoft.minecraftuwp_8wekyb3d8bbwe"`).CombinedOutput()
}
