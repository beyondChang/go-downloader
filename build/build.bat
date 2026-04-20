@echo off
setlocal

cd /d "%~dp0"

echo [1/5] Check go-winres...
where go-winres >nul 2>&1 || (
    echo go-winres not found, installing...
    go install github.com/tc-hib/go-winres@latest
    if errorlevel 1 (
        echo Install go-winres failed!
        pause
        exit /b 1
    )
)

echo [2/5] Generate resource file...
go-winres make
if errorlevel 1 (
    echo Generate resource failed!
    pause
    exit /b 1
)

for %%f in (rsrc_*.syso) do copy /y "%%f" "%~dp0..\%%f" >nul
del rsrc_*.syso 2>nul

cd /d "%~dp0.."

echo [3/5] Tidy modules...
go mod tidy
if errorlevel 1 (
    echo Go mod tidy failed!
    pause
    exit /b 1
)

echo [4/5] Build program...
set APPNAME=go-downloader
go build -ldflags="-s -w -H=windowsgui" -o %APPNAME%.exe .
if errorlevel 1 (
    echo Build failed!
    pause
    exit /b 1
)

del rsrc_*.syso 2>nul

echo [5/5] Check UPX...
where upx >nul 2>&1 && (
    echo Compressing with UPX...
    upx -9 %APPNAME%.exe
    if errorlevel 1 echo UPX compress failed, skip.
) || (
    echo UPX not found, skip.
)

echo.
echo Done: %APPNAME%.exe
pause
