@echo off

:loop
cls

@REM gocritic check -enable="#performance" ./...
@REM gocritic check -enableAll -disable="#experimental,#opinionated,#commentedOutCode" ./...

SET filename=GoShell
go build -buildvcs=false -race -o %filename%.exe

IF %ERRORLEVEL% EQU 0 %filename%.exe -nofiles

pause
goto loop