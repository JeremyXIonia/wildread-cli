@echo off
call "C:\Program Files\Microsoft Visual Studio\2022\Community\VC\Auxiliary\Build\vcvars64.bat" > NUL
cd /d D:\workspace-latest\cli-read
go test ./store/... -v
