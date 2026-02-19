# Sypher-mini build script (PowerShell)
# Usage: .\build.ps1 [build|clean|rebuild|test|run|extensions]

param(
    [Parameter(Position=0)]
    [ValidateSet("build", "clean", "rebuild", "test", "run", "extensions")]
    [string]$Target = "build"
)

$BuildDir = "build"
$BinaryName = "sypher"
$ExtensionsDir = "extensions"

function Do-Extensions {
    Get-ChildItem -Path $ExtensionsDir -Directory | ForEach-Object {
        $pkg = Join-Path $_.FullName "package.json"
        if (Test-Path $pkg) {
            Write-Host "Building extension: $($_.Name)"
            Push-Location $_.FullName
            npm install
            if ($LASTEXITCODE -ne 0) { Pop-Location; throw "npm install failed" }
            npm run build
            if ($LASTEXITCODE -ne 0) { Pop-Location; throw "npm run build failed" }
            Pop-Location
        }
    }
    Write-Host "Extensions build complete"
}

function Do-Build {
    Do-Extensions
    New-Item -ItemType Directory -Force -Path $BuildDir | Out-Null
    go build -o "$BuildDir/$BinaryName" ./cmd/sypher
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Build complete: $BuildDir/$BinaryName"
    }
}

function Do-Clean {
    if (Test-Path $BuildDir) {
        Remove-Item -Recurse -Force $BuildDir -ErrorAction SilentlyContinue
    }
    go clean -cache
    Get-ChildItem -Path $ExtensionsDir -Directory | ForEach-Object {
        $pkg = Join-Path $_.FullName "package.json"
        if (Test-Path $pkg) {
            $nodeModules = Join-Path $_.FullName "node_modules"
            $dist = Join-Path $_.FullName "dist"
            if (Test-Path $nodeModules) { Remove-Item -Recurse -Force $nodeModules -ErrorAction SilentlyContinue }
            if (Test-Path $dist) { Remove-Item -Recurse -Force $dist -ErrorAction SilentlyContinue }
        }
    }
    Write-Host "Clean complete"
}

switch ($Target) {
    "build"      { Do-Build }
    "clean"      { Do-Clean }
    "rebuild"    { Do-Clean; Do-Build }
    "extensions" { Do-Extensions }
    "test"       { go test ./... }
    "run"        { Do-Build; & "./$BuildDir/$BinaryName" gateway }
}
