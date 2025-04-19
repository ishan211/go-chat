// client.go
package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var app *tview.Application
var messageView *tview.TextView
var inputField *tview.InputField
var userList *tview.List
var conn net.Conn
var username string
var lastTypedAt time.Time

func connectToServer(user string) error {
	var err error
	conn, err = tls.Dial("tcp", "localhost:9000", &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return err
	}
	fmt.Fprintf(conn, "%s\n", user)
	return nil
}

func listenForMessages() {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msg := scanner.Text()
		app.QueueUpdateDraw(func() {
			if strings.HasPrefix(msg, "All users:") {
				updateUserList(msg)
			} else {
				messageView.Write([]byte(msg + "\n"))
			}
		})
	}
}

func updateUserList(line string) {
	parts := strings.Split(strings.TrimPrefix(line, "All users: "), ", ")
	userList.Clear()
	for _, entry := range parts {
		entry = strings.TrimSpace(entry)
		if entry != "" {
			userList.AddItem(entry, "", 0, nil)
		}
	}
}


func sendMessage(text string) {
	if conn == nil {
		return
	}
	fmt.Fprintln(conn, text)
	lastTypedAt = time.Now()
	go func() {
		time.Sleep(500 * time.Millisecond)
		if time.Since(lastTypedAt) < 2*time.Second {
			fmt.Fprintln(conn, "/status typing")
		}
	}()
}

func buildChatUI() tview.Primitive {
	messageView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() { app.Draw() })

	inputField = tview.NewInputField().
		SetLabel("Message: ").
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				text := inputField.GetText()
				inputField.SetText("")
				if text != "" {
					sendMessage(text)
				}
			}
		})

	list := tview.NewList()
	list.ShowSecondaryText(false)
	list.SetTitle("Users")
	list.SetBorder(true)
	userList = list

	chatLayout := tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(messageView, 0, 1, false).
			AddItem(inputField, 3, 0, true), 0, 3, true).
		AddItem(userList, 30, 1, false)

	return chatLayout
}

func main() {
	app = tview.NewApplication()

	var loginInput *tview.InputField
	loginInput = tview.NewInputField().
		SetLabel("Username: ").
		SetDoneFunc(func(key tcell.Key) {
			username = loginInput.GetText()
			if username == "" {
				return
			}
			if err := connectToServer(username); err != nil {
				log.Fatal(err)
			}
			go listenForMessages()
			app.SetRoot(buildChatUI(), true).SetFocus(inputField)
		})

	loginUI := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().SetText("Enter Username to Connect:"), 1, 0, false).
		AddItem(loginInput, 1, 0, true)

	if err := app.SetRoot(loginUI, true).Run(); err != nil {
		log.Fatal(err)
	}
}
