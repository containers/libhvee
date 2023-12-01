
. ./win-lib.ps1

# Targets
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

# TODO: add validate target
switch ($target) {
    {$_ -in '', 'binaries'} {
        Binaries
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
        Write-Host "Example: Run all tests "
        Write-Host " .\winmake localtest"
        Write-Host
    }
}
