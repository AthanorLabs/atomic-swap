// main
package main

import (
	"embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// 'wails dev' should properly launch vite to serve the site
// for live development without needing to separately launch
// 'npm run dev' or your flavor such as pnpm in the frontend
// directory separatelye

// The comment below chooses what gets packaged with
// the application.

//go:embed all:frontend/build
var assets embed.FS

// FileLoader ...
type FileLoader struct {
	http.Handler
}

// NewFileLoader ...
func NewFileLoader() *FileLoader {
	return &FileLoader{}
}

func (h *FileLoader) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var err error
	requestedFilename := strings.TrimPrefix(req.URL.Path, "/")
	println("Requesting file:", requestedFilename)
	fileData, err := os.ReadFile(filepath.Clean(requestedFilename))
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		_, _ = res.Write([]byte(fmt.Sprintf("Could not load file %s", requestedFilename)))
	}

	_, _ = res.Write(fileData)
}

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "AtomicSwap",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: NewFileLoader(),
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
