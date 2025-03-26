# Build the Ollama GUI application

param (
    [string]$Version = "0.0.0"
)

# Ensure we're in the project root
$ProjectRoot = Split-Path -Parent (Split-Path -Parent $PSCommandPath)
Set-Location $ProjectRoot

# Check for required tools
$wails = Get-Command wails -ErrorAction SilentlyContinue
if (-not $wails) {
    Write-Host "Wails is required to build the GUI application. Installing..."
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
}

# Check for NSIS for installer creation
$makensis = Get-Command makensis -ErrorAction SilentlyContinue
if (-not $makensis) {
    Write-Host "NSIS is required to create the installer. Please install it from https://nsis.sourceforge.io/Download"
    Write-Host "Continuing with just the executable build..."
}

# Build the GUI application
Write-Host "Building Ollama GUI..."
Set-Location "$ProjectRoot\app\ollama-gui"
wails build -trimpath -o "ollama-gui.exe" -v -ldflags "-X 'main.Version=$Version'"

# Copy the executable to the distribution directory
if (!(Test-Path "$ProjectRoot\dist")) {
    New-Item -ItemType Directory -Path "$ProjectRoot\dist" -Force
}

Write-Host "Copying GUI executable to dist directory..."
if (Test-Path "$ProjectRoot\app\ollama-gui\build\bin\ollama-gui.exe") {
    Copy-Item "$ProjectRoot\app\ollama-gui\build\bin\ollama-gui.exe" "$ProjectRoot\dist\ollama-gui.exe" -Force
}

# Check if we have ollama.exe in the dist directory
if (Test-Path "$ProjectRoot\dist\ollama.exe") {
    Write-Host "Found ollama.exe, will include in the package"
    
    # Create a package with both executables
    if (!(Test-Path "$ProjectRoot\dist\ollama-package")) {
        New-Item -ItemType Directory -Path "$ProjectRoot\dist\ollama-package" -Force
    }
    
    Copy-Item "$ProjectRoot\dist\ollama.exe" "$ProjectRoot\dist\ollama-package\ollama.exe" -Force
    Copy-Item "$ProjectRoot\dist\ollama-gui.exe" "$ProjectRoot\dist\ollama-package\ollama-gui.exe" -Force
    
    # Copy lib directory if it exists
    if (Test-Path "$ProjectRoot\dist\lib") {
        Copy-Item "$ProjectRoot\dist\lib" "$ProjectRoot\dist\ollama-package\" -Recurse -Force
    }
    
    # Create a zip file
    $zipPath = "$ProjectRoot\dist\ollama-gui-$Version.zip"
    if (Test-Path $zipPath) {
        Remove-Item $zipPath -Force
    }
    
    Write-Host "Creating zip package..."
    Compress-Archive -Path "$ProjectRoot\dist\ollama-package\*" -DestinationPath $zipPath -Force
    
    # Create installer if NSIS is available
    if ($makensis) {
        Write-Host "Creating installer..."
        
        # Copy necessary files for NSIS
        Copy-Item "$ProjectRoot\app\ollama-gui\build\installer.nsi" "$ProjectRoot\dist\ollama-package\installer.nsi" -Force
        
        # Create a simple license file if it doesn't exist
        if (!(Test-Path "$ProjectRoot\dist\ollama-package\license.txt")) {
            @"
Ollama GUI License

Copyright (c) 2023 Ollama Inc.
All rights reserved.

This software is provided as-is, without warranty of any kind.
"@ | Out-File -FilePath "$ProjectRoot\dist\ollama-package\license.txt" -Encoding utf8 -Force
        }
        
        # Run NSIS to create the installer
        Set-Location "$ProjectRoot\dist\ollama-package"
        & makensis installer.nsi
        
        if (Test-Path "$ProjectRoot\dist\ollama-package\OllamaGUISetup.exe") {
            Copy-Item "$ProjectRoot\dist\ollama-package\OllamaGUISetup.exe" "$ProjectRoot\dist\OllamaGUISetup.exe" -Force
            Write-Host "Installer created: $ProjectRoot\dist\OllamaGUISetup.exe"
        } else {
            Write-Host "Failed to create installer"
        }
    }
}

Write-Host "GUI build complete: $ProjectRoot\dist\ollama-gui.exe"
if (Test-Path "$ProjectRoot\dist\ollama-gui-$Version.zip") {
    Write-Host "Package created: $ProjectRoot\dist\ollama-gui-$Version.zip"
}
if (Test-Path "$ProjectRoot\dist\OllamaGUISetup.exe") {
    Write-Host "Installer created: $ProjectRoot\dist\OllamaGUISetup.exe"
}
