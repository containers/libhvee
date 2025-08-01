
. ./win-lib.ps1

# Targets
function Validate {
    Install-GolangciLint
    Run-Command "./bin/golangci-lint.exe run --skip-dirs `"test/e2e`""
}

function Install-GolangciLint {
    if (Test-Path -Path "./bin/golangci-lint.exe" -PathType Leaf) {
        Write-Host "golangci-lint already exists"
        return
    }
    
    Write-Host "Installing golangci-lint"
    New-Item -ItemType Directory -Force -Path "./bin" | Out-Null
    
    # Set version and download golangci-lint
    $version = "2.2.1"
    $url = "https://github.com/golangci/golangci-lint/releases/download/v$version/golangci-lint-$version-windows-amd64.zip"
    $zipPath = "./bin/golangci-lint.zip"
    
    try {
        Write-Host "Downloading golangci-lint v$version"
        Invoke-WebRequest -Uri $url -OutFile $zipPath
        Check-Exit 2 "Download golangci-lint"
        
        Write-Host "Extracting golangci-lint"
        Expand-Archive -Path $zipPath -DestinationPath "./bin/temp" -Force
        Check-Exit 2 "Extract golangci-lint"
        
        # Move the executable to the bin directory
        $extractedExe = "./bin/temp/golangci-lint-$version-windows-amd64/golangci-lint.exe"
        Move-Item -Path $extractedExe -Destination "./bin/golangci-lint.exe"
        Check-Exit 2 "Move golangci-lint executable"
        
        # Clean up
        Remove-Item -Path $zipPath -Force -ErrorAction Ignore
        Remove-Item -Path "./bin/temp" -Recurse -Force -ErrorAction Ignore
        
        Write-Host "golangci-lint installed successfully"
    }
    catch {
        Write-Host "Failed to install golangci-lint: $_"
        throw
    }
}

function Binaries{
    New-Item -ItemType Directory -Force -Path "./bin"

    Run-Command "go build -o bin ./cmd/kvpctl"
    Run-Command "go build -o bin ./cmd/dumpvms"
    Run-Command "go build -o bin ./cmd/createvm"
    Run-Command "go build -o bin ./cmd/updatevm"
}

function Make-Clean{
     Remove-Item ./bin -Recurse -Force -ErrorAction Ignore -Confirm:$false
}

function Local-Test {
    param (
    [string]$files
    );
    Build-Ginkgo
    if ($files) {
         $files = " --focus-file $files "
    }

    Run-Command "./test/tools/build/ginkgo.exe  $files ./test/e2e/. "
}

# Helpers
function Build-Ginkgo{
    if (Test-Path -Path ./test/tools/build/ginkgo.exe -PathType Leaf) {
        return
    }
    Write-Host "Building Ginkgo"
    Push-Location ./test/tools
    Run-Command "go build -o build/ginkgo.exe ./vendor/github.com/onsi/ginkgo/v2/ginkgo"
    Pop-Location
}


# Init script
$target = $args[0]

switch ($target) {
    {$_ -in '', 'binaries'} {
        Binaries
    }
    'validate' {
        Validate
    }
    'localtest' {
        if ($args.Count -gt 1) {
            $files = $args[1]
        }
        Local-Test  -files $files
    }
    'clean' {
        Make-Clean
    }
    default {
        Write-Host "Usage: " $MyInvocation.MyCommand.Name "<target> [options]"
        Write-Host
        Write-Host "Example: Build binaries "
        Write-Host " .\winmake binaries"
        Write-Host
        Write-Host "Example: Run validation "
        Write-Host " .\winmake validate"
        Write-Host
        Write-Host "Example: Run all tests "
        Write-Host " .\winmake localtest"
        Write-Host
    }
}
