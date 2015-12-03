@echo off

if exist "%cd%\.gopath\" rd /s /q "%cd%\.gopath\"
md "%cd%\.gopath\src\github.com\10gen\"
mklink /J "%cd%\.gopath\src\github.com\10gen\sqlproxy" "%cd%" >nul 2>&1
set GOPATH=%cd%\.gopath;%cd%\vendor
