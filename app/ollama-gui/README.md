# Ollama GUI

A simple graphical user interface for Ollama.

## Prerequisites

1. Install Go (1.18 or newer): https://go.dev/doc/install
2. Install Node.js (16 or newer): https://nodejs.org/
3. Install Wails: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

## Building the Application

### Development Mode

To run the application in development mode:

```bash
cd app/ollama-gui
wails dev
```

This will start the application in development mode with hot reloading.

### Building for Production

To build the application for production:

```bash
cd app/ollama-gui
wails build
```

This will create a production-ready executable in the `build/bin` directory.

## Architecture

This application acts as a frontend for the Ollama API server. It will:

1. Check if an Ollama server is already running
2. If not, it will start the Ollama server in the background
3. Connect to the server via the HTTP API
4. Provide a user-friendly interface for interacting with the Ollama models

## Features

- List available models
- Pull new models from Ollama library
- Chat with selected models
- Simple and intuitive user interface

## License

Same as the Ollama project.
