package main

import (
	"embed"
	"fmt"
)

//go:embed all:frontend/dist
var assets embed.FS

// The following program implements a proxy that forwards players from one local address to a remote address.
func main() {
	// Create an instance of the app structure
	app := NewApp()
	err := app.Connect("zeqa.net:19132")
	if err != nil {
		panic(err)
	}
	info, err := app.ProxyingInfo()
	if err != nil {
		panic(err)
	}
	fmt.Println(info.LocalAddress)
	select {}

	// Create application with options
	//err := wails.Run(&options.App{
	//	Title:         "Tedac",
	//	Width:         905,
	//	Height:        525,
	//	Frameless:     true,
	//	DisableResize: true,
	//	Assets:        assets,
	//	OnStartup:     app.startup,
	//	Bind:          []any{app},
	//})
	//if err != nil {
	//	panic(err)
	//}
}
