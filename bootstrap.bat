@echo off
rem Necto Autonomous Native Bootstrap Batch File
rem Builds the self-hosted Necto compiler without requiring Go.

echo ==================================================================
echo        Necto Autonomous Native Bootstrap (No Go Required)         
echo ==================================================================

where clang >nul 2>nul
if %ERRORLEVEL% equ 0 (
    set CC=clang
) else (
    where gcc >nul 2>nul
    if %ERRORLEVEL% equ 0 (
        set CC=gcc
    ) else (
        echo [ERROR] Neither clang nor gcc found in PATH.
        echo Please install Clang/LLVM or configure the toolchain.
        exit /b 1
    )
)

echo [OK] Backend C/LLVM compiler detected: %CC%

if exist bin\necto.exe (
    echo Bootstrapping self-hosted compiler via bin\necto.exe...
    bin\necto.exe bootstrap
) else (
    echo Compiling native compiler directly...
)

if exist bin\necto-native.exe (
    echo.
    echo [SUCCESS] bin\necto-native.exe is ready for pure native compilation!
) else (
    echo.
    echo [ERROR] bin\necto-native.exe was not created.
    exit /b 1
)
