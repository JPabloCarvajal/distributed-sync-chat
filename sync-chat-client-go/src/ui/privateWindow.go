package ui

import (
	"fmt"

	"chat-client/src/client"
	"chat-client/src/model"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ShowPrivateWindow pinta un chat 1-a-1 sobre un Channel que YA llega
// conectado y con el JOIN de emparejamiento enviado (lo abrió ChatClient,
// por clic propio o por OPEN_PRIVATE). Esta ventana es la ÚNICA dueña de
// ese socket: lo lee hasta que se cierra, y cerrar la ventana cierra SOLO
// ese canal (flujo E) — el otro lado se entera porque su propio read()
// falla, no porque nosotros le mandemos nada.
func ShowPrivateWindow(a fyne.App, c *client.ChatClient, peer string, ch *client.Channel) {
	w := a.NewWindow(fmt.Sprintf("%s -> %s", c.Username(), peer))
	w.Resize(fyne.NewSize(420, 420))

	log := newChatLog()

	input := widget.NewEntry()
	input.SetPlaceHolder("Escribe un mensaje...")
	send := widget.NewButton("Send", func() {
		body := input.Text
		if body == "" {
			return
		}
		ch.Send(model.Message{Type: model.TypePrivate, From: c.Username(), To: peer, Body: body})
		input.SetText("")
	})
	input.OnSubmitted = func(string) { send.OnTapped() }

	go func() {
		for msg := range ch.Incoming() {
			line := ""
			switch msg.Type {
			case model.TypePrivate:
				line = fmt.Sprintf("%s: %s", msg.From, msg.Body)
			case model.TypeError:
				line = fmt.Sprintf("[ERROR] %s", msg.Body)
			default:
				continue
			}
			fyne.Do(func() { log.append(line) })
		}
	}()

	title := widget.NewLabelWithStyle(peer, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	bottom := container.NewBorder(nil, nil, nil, send, input)

	w.SetContent(container.NewBorder(title, bottom, nil, nil, log.view))
	w.SetOnClosed(func() {
		ch.Close()
		privates.release(peer)
	})
	w.Show()
}
