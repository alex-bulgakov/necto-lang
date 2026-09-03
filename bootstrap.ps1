$ErrorActionPreference = "Stop"

Write-Host "==================================================================" -ForegroundColor Cyan
Write-Host "       Necto Autonomous Native Bootstrap (No Go Required)         " -ForegroundColor Cyan
Write-Host "==================================================================" -ForegroundColor Cyan

# 1. Discover C/LLVM Backend
$cc = $null
$candidates = @("clang", "gcc")
foreach ($c in $candidates) {
    $cmd = Get-Command "$c.exe" -ErrorAction SilentlyContinue
    if ($cmd) {
        $cc = $cmd.Source
        break
    }
}

if (-not $cc) {
    Write-Host "Error: Neither clang nor gcc found in PATH." -ForegroundColor Red
    Write-Host "To resolve: install LLVM/Clang or run 'necto toolchain install'."
    exit 1
}

Write-Host "Backend C/LLVM compiler detected: $cc" -ForegroundColor Green

# 2. Build or invoke bootstrap
if (Test-Path "bin\necto.exe") {
    Write-Host "Bootstrapping self-hosted compiler via bin\necto.exe..." -ForegroundColor Yellow
    & ".\bin\necto.exe" bootstrap
}

if (Test-Path "bin\necto-native.exe") {
    Write-Host "SUCCESS: bin\necto-native.exe is ready for pure native compilation!" -ForegroundColor Green
} else {
    Write-Host "Error: bin\necto-native.exe was not created." -ForegroundColor Red
    exit 1
}
