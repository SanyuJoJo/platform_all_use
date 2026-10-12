@echo off
rem 开发调试总入口（Windows）
rem 用法: dev.bat [start^|stop^|status]

set "ROOT=%~dp0"
cd /d "%ROOT%"

if "%~1"==""        make run & goto :eof
if /I "%~1"=="stop"   make stop & goto :eof
if /I "%~1"=="status" make status & goto :eof
make run
