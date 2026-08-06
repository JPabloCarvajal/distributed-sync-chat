package ui

import (
	"chat-client/src/client"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ShowLoginWindow pide servidor y usuario. Al conectar, abre la ventana de
// grupo (la ventana principal de la imagen) y cierra esta.
func ShowLoginWindow(a fyne.App, c *client.ChatClient) {
	w := a.NewWindow("Chat - Conectar")
	w.Resize(fyne.NewSize(360, 200))

	address := widget.NewEntry()
	address.SetText("127.0.0.1:1802")
	username := widget.NewEntry()
	username.SetPlaceHolder("usuario")

	status := widget.NewLabel("")

	form := widget.NewForm(
		widget.NewFormItem("Servidor", address),
		widget.NewFormItem("Usuario", username),
	)

	var connect *widget.Button
	connect = widget.NewButton("Conectar", func() {
		name := username.Text
		if name == "" {
			status.SetText("El usuario no puede estar vacío")
			return
		}

		connect.Disable()
		if err := c.Connect(address.Text, name); err != nil {
			status.SetText("No se pudo conectar: " + err.Error())
			connect.Enable()
			return
		}

		ShowGroupWindow(a, c)
		w.Close()
	})

	w.SetContent(container.NewVBox(form, connect, status))
	w.Show()
}
