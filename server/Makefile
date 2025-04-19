# server/Makefile
build:
	@echo Initializing Go module...
	if not exist go.mod go mod init server
	@echo Building server...
	go build -o server.exe server.go utils.go

run: build
	@echo Running server...
	server.exe

rebuild:
	@echo Cleaning server...
	del /f /q server.exe go.mod go.sum history.log status.json 2>nul
	del /f /q ..\tls\cert.pem ..\tls\key.pem 2>nul
	@make run

start:
	@if not exist server.exe ( make rebuild ) else ( make run )