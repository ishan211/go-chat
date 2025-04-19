// server.go
package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"
)

type Client struct {
	username      string
	conn          net.Conn
	send          chan string
	connected     bool
	history       []string
	status        string
	lastTypedTime time.Time
}

var (
	clients   = make(map[string]*Client)
	clientsMu sync.Mutex
	messageID = 1
)

func broadcast(sender string, msg string, privateTo string) {
	timestamp := time.Now().Format("15:04:05")
	formatted := fmt.Sprintf("[%s] [%s] #%d: %s", timestamp, sender, messageID, msg)

	clientsMu.Lock()
	defer clientsMu.Unlock()

	senderClient, senderExists := clients[sender]
	if sender != "Server" && (!senderExists || senderClient == nil) {
		fmt.Printf("❌ broadcast(): sender %s not found\n", sender)
		return
	}

	if privateTo != "" {
		if receiver, ok := clients[privateTo]; ok && receiver.connected {
			receiver.send <- "[PM] " + formatted
			receiver.history = append(receiver.history, "[PM] "+formatted)
			senderClient.send <- "[PM to " + privateTo + "] " + formatted
			senderClient.history = append(senderClient.history, "[PM to "+privateTo+"] "+formatted)
		} else {
			senderClient.send <- "❌ User not found or offline"
		}
		return
	}

	for _, client := range clients {
		if client.connected {
			client.send <- formatted
			client.history = append(client.history, formatted)
		}
	}

	messageID++
}

func handleClient(client *Client) {
	defer func() {
		clientsMu.Lock()
		client.connected = false
		client.status = "disconnected"
		clientsMu.Unlock()
		client.conn.Close()
		broadcast("Server", fmt.Sprintf("%s has disconnected.", client.username), "")
	}()

	scanner := bufio.NewScanner(client.conn)
	for scanner.Scan() {
		text := scanner.Text()
		fmt.Println("[Server] Received from", client.username+":", text)

		if strings.HasPrefix(text, "/msg ") {
			parts := strings.SplitN(text, " ", 3)
			if len(parts) < 3 {
				client.send <- "❌ Usage: /msg username message"
				continue
			}
			broadcast(client.username, parts[2], parts[1])
			continue
		}

		switch {
		case strings.HasPrefix(text, "/status "):
			newStatus := strings.TrimSpace(strings.TrimPrefix(text, "/status "))
			if newStatus == "typing" && time.Since(client.lastTypedTime) > 2*time.Second {
				continue
			}
			client.status = newStatus

		case text == "/users":
			clientsMu.Lock()
			var others []string
			for name, c := range clients {
				displayStatus := c.status
				if displayStatus == "typing" && time.Since(c.lastTypedTime) > 2*time.Second {
					displayStatus = "available"
				}
				if name != client.username {
					others = append(others, fmt.Sprintf("%s [%s]", name, displayStatus))
				}
			}
			sort.SliceStable(others, func(i, j int) bool {
				return strings.Split(others[i], " ")[1] < strings.Split(others[j], " ")[1]
			})
			selfStatus := client.status
			if selfStatus == "typing" && time.Since(client.lastTypedTime) > 2*time.Second {
				selfStatus = "available"
			}
			fullList := append([]string{fmt.Sprintf("%s [%s]", client.username, selfStatus)}, others...)
			client.send <- "All users: " + strings.Join(fullList, ", ")
			clientsMu.Unlock()

		default:
			broadcast(client.username, text, "")
		}

		client.lastTypedTime = time.Now()
	}
}

func clientWriter(client *Client) {
	for msg := range client.send {
		fmt.Fprintln(client.conn, msg)
	}
}

func getLocalIP() string {
	conn, err := net.Dial("udp", "10.255.255.255:1")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

func main() {
	tlsDir := "../tls"
	certPath := tlsDir + "/cert.pem"
	keyPath := tlsDir + "/key.pem"

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		fmt.Println("🔐 Generating TLS certificate...")
		os.MkdirAll(tlsDir, 0755)
		cmd := exec.Command("openssl", "req", "-x509", "-newkey", "rsa:2048",
			"-keyout", keyPath, "-out", certPath,
			"-days", "365", "-nodes", "-subj", "/CN=localhost")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatal("❌ Failed to generate TLS cert:", err)
		}
	}

	cert, _ := tls.LoadX509KeyPair(certPath, keyPath)
	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	ln, err := tls.Listen("tcp", ":9000", config)
	if err != nil {
		panic(err)
	}

	ip := getLocalIP()
	fmt.Printf("\n📡 Server running at: %s:9000\n", ip)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		go func(conn net.Conn) {
			fmt.Fprint(conn, "Enter username: ")
			scanner := bufio.NewScanner(conn)
			if !scanner.Scan() {
				conn.Close()
				return
			}
			username := scanner.Text()

			clientsMu.Lock()
			client, exists := clients[username]
			if exists {
				if client.connected {
					fmt.Fprintln(conn, "❌ User already online")
					clientsMu.Unlock()
					conn.Close()
					return
				}
				client.conn = conn
				client.send = make(chan string)
				client.connected = true
			} else {
				client = &Client{
					username:  username,
					conn:      conn,
					send:      make(chan string),
					connected: true,
					history:   []string{},
					status:    "available",
				}
				clients[username] = client
			}
			clientsMu.Unlock()

			go clientWriter(client)

			for _, msg := range client.history {
				client.send <- "[History] " + msg
			}

			broadcast("Server", fmt.Sprintf("%s has joined (%s).", username, client.status), "")
			handleClient(client)
		}(conn)
	}
}
