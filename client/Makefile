# client/Makefile
build:
	@echo "Initializing go module if needed..."
	@test -f go.mod || go mod init client
	@go list -m github.com/rivo/tview >/dev/null 2>&1 || go get github.com/rivo/tview github.com/gdamore/tcell/v2
	go build -o chat-client chat_tview_client.go

run: build
	./chat-client
