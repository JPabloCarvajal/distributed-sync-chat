package main

import (
	"chat-client/src/client"
	"chat-client/src/ui"

	"fyne.io/fyne/v2/app"
)

func main() {
	a := app.NewWithID("co.chat-client.app")
	ui.ShowLoginWindow(a, client.NewChatClient())
	a.Run()
}
