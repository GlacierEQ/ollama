# Ollama Project

## Project Setup

This project requires the following tools to be installed and available in your system PATH:

1. Go (https://golang.org/doc/install)
2. GCC (MinGW-w64 for Windows: http://mingw-w64.org/doku.php/download)
3. CMake (https://cmake.org/download/)

Please install these tools before proceeding with the project setup.

## Installation Steps

1. Install Go:
   - Download and install Go from https://golang.org/doc/install
   - Add Go to your system PATH

2. Install GCC:
   - For Windows, download and install MinGW-w64 from http://mingw-w64.org/doku.php/download
   - Add the MinGW-w64 bin directory to your system PATH

3. Install CMake:
   - Download and install CMake from https://cmake.org/download/
   - Add CMake to your system PATH

4. Verify installations:
   - Open a new command prompt or terminal
   - Run the following commands to verify the installations:
     ```
     go version
     gcc --version
     cmake --version
     ```

5. Clone the repository:
   ```
   git clone https://github.com/yourusername/ollama.git
   cd ollama
   ```

6. Initialize Go modules:
   ```
   go mod tidy
   ```

7. Build the project:
   ```
   go build
   ```

## Running the Project

After completing the setup, you can run the project using:

```
go run main.go
```

For more detailed instructions on using the project, please refer to the documentation in the `docs` directory.
