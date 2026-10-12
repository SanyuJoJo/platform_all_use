@echo off
setlocal

set "SCRIPT_DIR=%~dp0"
for %%I in ("%SCRIPT_DIR%..") do set "BIN_DIR=%%~fI"
set "PLUGIN_DIR=%BIN_DIR%\plugins"
set "DATA_DIR=%BIN_DIR%\data"
set "LOG_DIR=%DATA_DIR%\logs"
set "PID_DIR=%DATA_DIR%\pids"
set "HOST_BIN=%BIN_DIR%\host.exe"

rem 默认配置文件为 bin\config.json
if "%HOST_CONFIG%"=="" set "HOST_CONFIG=%BIN_DIR%\config.json"

if not exist "%LOG_DIR%"          mkdir "%LOG_DIR%"
if not exist "%PID_DIR%"          mkdir "%PID_DIR%"
if not exist "%DATA_DIR%\backups" mkdir "%DATA_DIR%\backups"

if "%~1"==""         goto :fg
if /I "%~1"=="fg"    goto :fg
if /I "%~1"=="start" goto :start
if /I "%~1"=="stop"  goto :stop
if /I "%~1"=="status" goto :status
goto :usage

:fg
if not exist "%HOST_BIN%" (
    echo [start] 未找到主程序: %HOST_BIN%
    echo         请先运行: make all
    exit /b 1
)
echo [start] 前台启动 host（Ctrl+C 退出）
echo [start] 配置文件: %HOST_CONFIG%
echo [start] 日志: %LOG_DIR%\host.log
echo -------------------------------------------------------------
powershell -NoProfile -Command "$env:HOST_CONFIG='%HOST_CONFIG%'; & '%HOST_BIN%' 2>&1 | Tee-Object -FilePath '%LOG_DIR%\host.log' -Append"
goto :eof

:start
if not exist "%HOST_BIN%" (
    echo [start] 未找到主程序: %HOST_BIN%
    exit /b 1
)
echo [start] 后台启动 host
echo [start] 配置文件: %HOST_CONFIG%
powershell -NoProfile -Command "$env:HOST_CONFIG='%HOST_CONFIG%'; Start-Process -FilePath '%HOST_BIN%' -WindowStyle Hidden -RedirectStandardOutput '%LOG_DIR%\host.log' -RedirectStandardError '%LOG_DIR%\host.err.log'"
echo [start] host 已启动
goto :eof

:stop
taskkill /IM host.exe /F >nul 2>&1
taskkill /IM echo.exe /F >nul 2>&1
taskkill /IM python.exe /F >nul 2>&1
echo [stop] 已全部停止
goto :eof

:status
echo === 进程状态 ===
tasklist /FI "IMAGENAME eq host.exe" | find /I "host.exe" >nul && echo host   运行中 || echo host   已停止
tasklist /FI "IMAGENAME eq echo.exe" | find /I "echo.exe" >nul && echo echo   运行中 || echo echo   已停止
echo === 配置文件 ===
echo %HOST_CONFIG%
goto :eof

:usage
echo 用法: %~nx0 [fg^|start^|stop^|status]
echo 环境变量: HOST_CONFIG=^<配置文件路径^>
exit /b 1
