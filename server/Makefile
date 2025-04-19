# server/Makefile
build:
	@echo "Initializing go module if needed..."
	@test -f go.mod || go mod init server
	go build -o chat-server chat_server.go utils.go

run: build
	./chat-server

gen-cert:
	@if [ ! -f cert.pem ] || [ ! -f key.pem ]; then \
		echo "Generating TLS certificate..."; \
		openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 365 -nodes -subj "/CN=localhost"; \
	else \
		echo "cert.pem and key.pem already exist, skipping."; \
	fi
