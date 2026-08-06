package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// chatLog es la caja blanca de mensajes de la imagen: líneas que se
// apilan hacia abajo y hacen autoscroll. La usan tanto la ventana de
// grupo como cada ventana privada, así que vive aparte para no repetir
// este layout en las dos.
type chatLog struct {
	box  *fyne.Container
	view *container.Scroll
}

func newChatLog() *chatLog {
	box := container.NewVBox()
	return &chatLog{box: box, view: container.NewVScroll(box)}
}

func (l *chatLog) append(line string) {
	l.box.Add(widget.NewLabel(line))
	l.box.Refresh()
	l.view.ScrollToBottom()
}
