@echo off

@REM SET filename=explorer

:loop
cls
:: go get github.com/akavel/rsrc
:: rsrc -manifest main.manifest -o rsrc.syso
@REM go build -buildvcs=false -o %filename%.exe

@REM SET filename=explorer
@REM go build -buildvcs=false -o %filename%.exe

@REM rsrc -manifest GoShell.exe.manifest -ico=app.ico,add.ico,application_lightning.ico,application_edit.ico,application_error.ico -o rsrc.syso

go-winres make
SET filename=GoShell
go build -buildvcs=false -race -o %filename%.exe

pause
goto loop