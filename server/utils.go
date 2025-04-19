package main

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
)

var (
	historyFile = "history.log"
	statusFile  = "status.json"
	historyMu   sync.Mutex
)

func AppendToHistory(line string) {
	historyMu.Lock()
	defer historyMu.Unlock()
	f, err := os.OpenFile(historyFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(line + "\n")
}

func LoadHistory() {
	_, err := os.Stat(historyFile)
	if os.IsNotExist(err) {
		os.Create(historyFile)
	}
}

func ReadFullHistory() []string {
	historyMu.Lock()
	defer historyMu.Unlock()

	var lines []string
	f, err := os.Open(historyFile)
	if err != nil {
		return lines
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

func SaveStatuses(clients map[string]*Client) {
	data := map[string]string{}
	for name, c := range clients {
		data[name] = c.status
	}
	jsonData, _ := json.MarshalIndent(data, "", "  ")
	os.WriteFile(statusFile, jsonData, 0644)
}

func LoadStatuses() map[string]string {
	data := map[string]string{}
	bytes, err := os.ReadFile(statusFile)
	if err != nil {
		return data
	}
	json.Unmarshal(bytes, &data)
	return data
}
