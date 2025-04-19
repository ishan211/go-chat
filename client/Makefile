# client/Makefile
build:
	@echo Initializing Go module...
	if not exist go.mod go mod init client
	@echo Installing dependencies...
	go get github.com/rivo/tview github.com/gdamore/tcell/v2
	@echo Building client...
	go build -o client.exe client.go

run: build
	@echo Running client...
	client.exe

rebuild:
	@echo Cleaning up...
	del /f /q client.exe go.mod go.sum 2>nul
	@make run

