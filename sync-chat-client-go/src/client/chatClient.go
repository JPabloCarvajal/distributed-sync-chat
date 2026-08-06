package client

import (
	"sync"

	"chat-client/src/model"
)

// ChatClient es el orquestador de dominio (el "ATMClient" del cajero,
// para chat): no sabe de JSON ni de sockets (eso es Channel), y no sabe de
// ventanas ni de widgets (eso es la UI).
//
// Solo el canal GRUPAL vive aquí, porque es el único que existe siempre
// (plano de control: JOIN inicial, GROUP, USERS, OPEN_PRIVATE, ERROR).
// Cada chat privado es un Channel aparte que ChatClient sabe ABRIR
// (OpenPrivate), pero cuya vida entera (leer, mostrar, cerrar) la dueña
// es la ventana privada que lo usa — así se respeta "cerrar una ventana
// privada solo mata SU socket".
type ChatClient struct {
	address  string
	username string
	group    *Channel

	mu                  sync.Mutex
	groupHandlers       []func(model.Message)
	usersHandlers       []func([]string)
	errorHandlers       []func(model.Message)
	openPrivateHandlers []func(peer string, ch *Channel)
	lastUsers           []string // último USERS recibido, para reproducirlo
}

func NewChatClient() *ChatClient {
	return &ChatClient{}
}

func (c *ChatClient) Username() string {
	return c.username
}

// Connect abre el canal grupal: JOIN sin 'to'. A partir de aquí queda vivo
// hasta Close(). Guarda la dirección porque cada privado abre OTRO socket
// contra el mismo servidor.
func (c *ChatClient) Connect(address, username string) error {
	ch := NewChannel(address)
	if err := ch.Open(); err != nil {
		return err
	}
	if err := ch.Send(model.Message{Type: model.TypeJoin, From: username}); err != nil {
		ch.Close()
		return err
	}

	c.address = address
	c.username = username
	c.group = ch

	go c.dispatchGroup()
	return nil
}

func (c *ChatClient) SendGroup(body string) error {
	return c.group.Send(model.Message{Type: model.TypeGroup, From: c.username, Body: body})
}

// OpenPrivate abre un socket NUEVO y dedicado con peer: JOIN{from, to}.
// Lo usa la UI en los dos lados del flujo C: cuando el usuario local hace
// clic sobre alguien (inicia), y cuando llega un OPEN_PRIVATE (responde).
func (c *ChatClient) OpenPrivate(peer string) (*Channel, error) {
	ch := NewChannel(c.address)
	if err := ch.Open(); err != nil {
		return nil, err
	}
	if err := ch.Send(model.Message{Type: model.TypeJoin, From: c.username, To: peer}); err != nil {
		ch.Close()
		return nil, err
	}
	return ch, nil
}

// OnGroupMessage registra un handler para cada mensaje de GRUPO que llegue.
func (c *ChatClient) OnGroupMessage(h func(model.Message)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.groupHandlers = append(c.groupHandlers, h)
}

// OnUsers registra un handler para la lista de usuarios conectados.
// Si el USERS ya llegó antes de registrarse (carrera de arranque), se
// reproduce el último conocido.
func (c *ChatClient) OnUsers(h func([]string)) {
	c.mu.Lock()
	c.usersHandlers = append(c.usersHandlers, h)
	last := c.lastUsers
	c.mu.Unlock()

	if last != nil {
		go h(last) // en goroutine: el handler llama a fyne.Do
	}
}

// OnError registra un handler para ERROR llegado por el canal grupal.
func (c *ChatClient) OnError(h func(model.Message)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.errorHandlers = append(c.errorHandlers, h)
}

// OnOpenPrivate registra un handler que se dispara ya con el canal privado
// ABIERTO Y EMPAREJADO (JOIN ya enviado): la UI solo tiene que pintar la
// ventana encima.
func (c *ChatClient) OnOpenPrivate(h func(peer string, ch *Channel)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.openPrivateHandlers = append(c.openPrivateHandlers, h)
}

// dispatchGroup es la única goroutine que consume el canal grupal.
// Nunca toca widgets: solo llama handlers, responsabilidad de la UI.
func (c *ChatClient) dispatchGroup() {
	for msg := range c.group.Incoming() {
		switch msg.Type {
		case model.TypeGroup:
			c.notifyGroup(msg)
		case model.TypeUsers:
			c.notifyUsers(msg.Users)
		case model.TypeOpenPrivate:
			// en goroutine aparte: abrir el socket privado implica un Dial
			// (bloqueante) y no puede frenar la lectura del canal grupal.
			go c.handleOpenPrivate(msg.From)
		case model.TypeError:
			c.notifyError(msg)
		}
	}
}

func (c *ChatClient) handleOpenPrivate(peer string) {
	ch, err := c.OpenPrivate(peer)
	if err != nil {
		return
	}

	c.mu.Lock()
	handlers := append([]func(string, *Channel){}, c.openPrivateHandlers...)
	c.mu.Unlock()

	for _, h := range handlers {
		h(peer, ch)
	}
}

func (c *ChatClient) notifyGroup(msg model.Message) {
	c.mu.Lock()
	handlers := append([]func(model.Message){}, c.groupHandlers...)
	c.mu.Unlock()

	for _, h := range handlers {
		h(msg)
	}
}

func (c *ChatClient) notifyUsers(users []string) {
	c.mu.Lock()
	c.lastUsers = users
	handlers := append([]func([]string){}, c.usersHandlers...)
	c.mu.Unlock()

	for _, h := range handlers {
		h(users)
	}
}

func (c *ChatClient) notifyError(msg model.Message) {
	c.mu.Lock()
	handlers := append([]func(model.Message){}, c.errorHandlers...)
	c.mu.Unlock()

	for _, h := range handlers {
		h(msg)
	}
}

// Close cierra el canal grupal. Los privados los cierra cada ventana.
func (c *ChatClient) Close() {
	if c.group != nil {
		c.group.Close()
	}
}
