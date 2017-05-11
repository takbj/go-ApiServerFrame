@echo off
rem set GOPATH=%GOPATH%;%~dp0
set GOROOT=D:\Go

set GO=D:\Go\bin\go.exe

set GOOS=windows
set GOARCH=amd64


set VERSION=v.1.0.0
echo version=%VERSION%
set LDFLAGS=" -w -s -X main._VERSION_=%VERSION%"

echo start install sever ...
%GO% install -ldflags %LDFLAGS% apiServer/serverMain
::-gcflags " -N -l"
::%GO% install -gcflags " -N -l" ucserver/uc_main

del config\\*.json /Q
copy src\\apiServer\\config\\*.json config\\
echo complate
pause

bin\serverMain.exe
