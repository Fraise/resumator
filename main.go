package main

import (
	"embed"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var appIcon []byte

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "Resumator",
		Width:     1280,
		Height:    820,
		MinWidth:  960,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		// Linux has no build-time icon resource like Windows/macOS, so the
		// taskbar/window icon must be passed explicitly. WebviewGpuPolicyNever
		// must be set explicitly too: it is only Wails' default when the
		// Linux options are nil, and GPU acceleration triggers
		// "Error 71 (Protocol error) dispatching to Wayland display"
		// (wails issue #2977, DMABUF on Wayland).
		Linux: &linux.Options{
			Icon:             appIcon,
			WebviewGpuPolicy: linux.WebviewGpuPolicyNever,
		},
		// WebKitGTK navigates the webview to any file dropped onto the
		// window (file:// URL); the app has no drop handling, so disable it.
		DragAndDrop: &options.DragAndDrop{DisableWebViewDrop: true},
		OnStartup:   app.startup,
		OnShutdown:  app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

// writeAll writes bytes to a file, truncating it first.
func writeAll(path string, data []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	// Close can report delayed write errors; don't swallow them.
	return f.Close()
}
