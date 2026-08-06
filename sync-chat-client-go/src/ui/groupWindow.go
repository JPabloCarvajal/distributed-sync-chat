package ui

import (
	"fmt"

	"chat-client/src/client"
	"chat-client/src/model"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ShowGroupWindow es la ventana principal de la imagen: título con el
// usuario, chat de grupo a la izquierda, lista de usuarios conectados a la
// derecha e input + Send abajo.
//
// Un clic en un usuario de la lista ABRE UN SOCKET NUEVO (flujo C, paso 1):
// JOIN{from: yo, to: peer}, y con ese canal ya conectado pinta la ventana
// privada. Si llega un OPEN_PRIVATE (alguien más nos invita, flujo C paso 2)
// ChatClient ya dejó el canal abierto y emparejado; aquí solo se pinta la
// ventana. En ningún caso se reutiliza una ventana/canal existente: cada
// apertura -por clic propio o por invitación- crea uno nuevo.
func ShowGroupWindow(a fyne.App, c *client.ChatClient) {
	w := a.NewWindow(c.Username())
	w.Resize(fyne.NewSize(700, 480))

	log := newChatLog()

	users := []string{}
	userList := widget.NewList(
		func() int { return len(users) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(users[id])
		},
	)
	userList.OnSelected = func(id widget.ListItemID) {
		peer := users[id]
		userList.UnselectAll()
		if peer == c.Username() {
			return
		}
		if !privates.claim(peer) {
			return // ya hay ventana abierta con él
		}

		go func() {
			ch, err := c.OpenPrivate(peer)
			if err != nil {
				privates.release(peer)
				fyne.Do(func() { dialog.ShowError(err, w) })
				return
			}
			fyne.Do(func() { ShowPrivateWindow(a, c, peer, ch) })
		}()
	}

	input := widget.NewEntry()
	input.SetPlaceHolder("Escribe un mensaje...")
	send := widget.NewButton("Send", func() {
		body := input.Text
		if body == "" {
			return
		}
		c.SendGroup(body)
		input.SetText("")
	})
	input.OnSubmitted = func(string) { send.OnTapped() }

	c.OnGroupMessage(func(msg model.Message) {
		fyne.Do(func() {
			log.append(fmt.Sprintf("[%d] %s: %s", msg.Seq, msg.From, msg.Body))
		})
	})

	c.OnUsers(func(list []string) {
		fyne.Do(func() {
			users = list
			userList.Refresh()
		})
	})

	c.OnError(func(msg model.Message) {
		fyne.Do(func() {
			log.append(fmt.Sprintf("[ERROR] %s", msg.Body))
		})
	})

	// Flujo C, paso 2: alguien más quiere hablar conmigo. ChatClient ya
	// hizo el JOIN de emparejamiento; aquí solo abrimos la ventana.
	c.OnOpenPrivate(func(peer string, ch *client.Channel) {
		if !privates.claim(peer) {
			ch.Close() // ya tenemos ventana con él: este canal sobra
			return
		}
		fyne.Do(func() { ShowPrivateWindow(a, c, peer, ch) })
	})

	sidebar := container.NewBorder(nil, nil, nil, nil, userList)
	center := container.NewHSplit(log.view, sidebar)
	center.Offset = 0.75

	title := widget.NewLabelWithStyle(c.Username(), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	bottom := container.NewBorder(nil, nil, nil, send, input)

	w.SetContent(container.NewBorder(title, bottom, nil, nil, center))

	w.Show()
}
