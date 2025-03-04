package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

// App struct
type App struct {
	ctx             context.Context
	serverRunning   bool
	serverProcess   *exec.Cmd
	serverCancelFn  context.CancelFunc
	ollamaExecutable string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.findOllamaExecutable()
	a.ensureOllamaIsRunning()
}

func (a *App) findOllamaExecutable() {
	var exeName string
	if runtime.GOOS == "windows" {
		exeName = "ollama.exe"
	} else {
		exeName = "ollama"
	}

	// First check if it's in the same directory as our app
	execPath, err := os.Executable()
	if err == nil {
		dirPath := filepath.Dir(execPath)
		candidatePath := filepath.Join(dirPath, exeName)
		if _, err := os.Stat(candidatePath); err == nil {
			a.ollamaExecutable = candidatePath
			return
		}
	}

	// Next check if it's in PATH
	path, err := exec.LookPath(exeName)
	if err == nil {
		a.ollamaExecutable = path
		return
	}

	// Fallback to specific locations
	commonPaths := []string{
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Ollama", exeName),
		"/usr/local/bin/ollama",
		"/usr/bin/ollama",
		"/opt/homebrew/bin/ollama",
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			a.ollamaExecutable = path
			return
		}
	}

	log.Println("Could not find the Ollama executable")
}

// ensureOllamaIsRunning makes sure the Ollama server is running
func (a *App) ensureOllamaIsRunning() {
	// First check if the server is already running
	client := &http.Client{
		Timeout: 1 * time.Second,
	}
	resp, err := client.Get("http://localhost:11434/api/health")
	if err == nil && resp.StatusCode == http.StatusOK {
		a.serverRunning = true
		return
	}

	// Start the server if it's not running and we have the executable
	if a.ollamaExecutable != "" {
		var ctx context.Context
		ctx, a.serverCancelFn = context.WithCancel(context.Background())
		a.serverProcess = exec.CommandContext(ctx, a.ollamaExecutable, "serve")
		
		// Run the process in the background
		err := a.serverProcess.Start()
		if err != nil {
			log.Printf("Failed to start Ollama server: %v", err)
			return
		}

		// Wait for the server to be ready
		for i := 0; i < 30; i++ { // Try for 30 seconds
			resp, err := client.Get("http://localhost:11434/api/health")
			if err == nil && resp.StatusCode == http.StatusOK {
				a.serverRunning = true
				break
			}
			time.Sleep(1 * time.Second)
		}
	}
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	// Gracefully stop the ollama server
	if a.serverCancelFn != nil {
		a.serverCancelFn()
	}
}

// GetModels returns a list of available models
func (a *App) GetModels() []map[string]interface{} {
	if !a.serverRunning {
		return []map[string]interface{}{}
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get("http://localhost:11434/api/tags")
	if err != nil {
		log.Printf("Error getting models: %v", err)
		return []map[string]interface{}{}
	}
	defer resp.Body.Close()

	var result struct {
		Models []map[string]interface{} `json:"models"`
	}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&result); err != nil {
		log.Printf("Error decoding response: %v", err)
		return []map[string]interface{}{}
	}

	return result.Models
}

// RunModel starts a chat with the specified model
func (a *App) RunModel(model string) string {
	if !a.serverRunning {
		return "Ollama server is not running"
	}

	// In a real implementation, you would connect to the model via the API
	// and return a session ID or reference for further interaction
	return fmt.Sprintf("Started session with model: %s", model)
}

// Chat sends a message to the model and returns the response
func (a *App) Chat(model, message string) string {
	if !a.serverRunning {
		return "Ollama server is not running"
	}

	// In a real implementation, you would send the message to the model API
	// and return the response
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": message},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Sprintf("Error preparing request: %v", err)
	}

	req, err := http.NewRequest("POST", "http://localhost:11434/api/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Sprintf("Error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&result); err != nil {
		return fmt.Sprintf("Error parsing response: %v", err)
	}

	return result.Message.Content
}

func main() {
	// Create application with options
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "Ollama GUI",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarDefault(),
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			About: &mac.AboutInfo{
				Title:   "Ollama GUI",
				Message: "© 2023 Ollama",
				Icon:    icon,
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
