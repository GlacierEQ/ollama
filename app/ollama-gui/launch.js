// This script checks if ollama server is running and starts it if necessary
// Then launches the GUI application

const { spawn, exec } = require('child_process');
const fs = require('fs');
const path = require('path');
const http = require('http');

// Get the directory where this script is running
const appDir = path.dirname(process.execPath);

// Function to check if Ollama server is running
function checkOllamaServer() {
    return new Promise((resolve) => {
        const req = http.request({
            hostname: 'localhost',
            port: 11434,
            path: '/api/health',
            method: 'GET',
            timeout: 1000
        }, (res) => {
            if (res.statusCode === 200) {
                resolve(true);
            } else {
                resolve(false);
            }
        });

        req.on('error', () => {
            resolve(false);
        });

        req.end();
    });
}

// Function to start Ollama server
function startOllamaServer() {
    const ollamaPath = path.join(appDir, 'ollama.exe');

    if (fs.existsSync(ollamaPath)) {
        console.log('Starting Ollama server...');
        const ollamaProcess = spawn(ollamaPath, ['serve'], {
            detached: true,
            stdio: 'ignore'
        });

        // Unref the process to let it run independently
        ollamaProcess.unref();

        return new Promise((resolve) => {
            // Wait for server to start (up to 10 seconds)
            let attempts = 0;
            const checkInterval = setInterval(async () => {
                attempts++;
                const running = await checkOllamaServer();

                if (running) {
                    clearInterval(checkInterval);
                    console.log('Ollama server started successfully');
                    resolve(true);
                } else if (attempts >= 20) { // 20 * 500ms = 10 seconds
                    clearInterval(checkInterval);
                    console.log('Timeout waiting for Ollama server to start');
                    resolve(false);
                }
            }, 500);
        });
    } else {
        console.error('Ollama executable not found at:', ollamaPath);
        return Promise.resolve(false);
    }
}

// Main execution
async function main() {
    // Check if server is running, start it if not
    const serverRunning = await checkOllamaServer();

    if (!serverRunning) {
        await startOllamaServer();
    } else {
        console.log('Ollama server is already running');
    }
}

// Start the application
main().catch(console.error);
