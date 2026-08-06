package client

import (
	"bufio"
	"encoding/json"
	"net"
	"sync"

	"chat-client/src/model"
)

// Channel es UN socket TCP: el mismo rol que ATMProxy en el cajero, pero
// aquí la conversación no es ping-pong (una petición, una respuesta) sino
// que el servidor puede escribir en cualquier momento. Por eso, además de
// Open/Send/Close (igual que ATMProxy), hace falta una goroutine de
// lectura continua que deje cada mensaje en Incoming().
//
// En este protocolo cada conversación tiene SU PROPIO Channel: el canal
// grupal es uno, y cada chat privado abre otro independiente (con su
// propio JOIN{from,to}). Channel no sabe si es "el grupal" o "un privado
// con X": eso lo decide quien lo usa.
type Channel struct {
	address string
	conn    net.Conn
	reader  *bufio.Reader
	writer  *bufio.Writer
	wmu     sync.Mutex // Send puede llamarse desde la goroutine de UI mientras readLoop sigue leyendo

	incoming chan model.Message
}

// NewChannel NO conecta: solo recuerda a quién hay que llamar (igual que NewATMProxy).
func NewChannel(address string) *Channel {
	return &Channel{address: address}
}

// Open abre el socket y arranca la goroutine de lectura.
// Todavía no manda JOIN: eso lo decide el llamador (grupal vs. privado).
func (ch *Channel) Open() error {
	conn, err := net.Dial("tcp", ch.address)
	if err != nil {
		return err
	}
	ch.conn = conn
	ch.reader = bufio.NewReader(conn)
	ch.writer = bufio.NewWriter(conn)
	ch.incoming = make(chan model.Message, 64)

	go ch.readLoop()
	return nil
}

// readLoop es la única goroutine que lee este socket. "No bloqueante" se
// resuelve como en el cajero-cliente: el ReadBytes bloqueante a nivel de
// SO corre aparte, y el resto del programa sigue vivo consumiendo
// Incoming(). Termina sola (cerrando el channel) cuando el servidor cierra
// la conexión.
func (ch *Channel) readLoop() {
	defer close(ch.incoming)
	for {
		line, err := ch.reader.ReadBytes('\n')
		if err != nil {
			return
		}
		var msg model.Message
		if err := json.Unmarshal(line, &msg); err != nil {
			continue // línea corrupta: se descarta, no se tumba la conexión
		}
		ch.incoming <- msg
	}
}

// Incoming expone el canal de solo-lectura de mensajes ya parseados.
func (ch *Channel) Incoming() <-chan model.Message {
	return ch.incoming
}

func (ch *Channel) Send(msg model.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	ch.wmu.Lock()
	defer ch.wmu.Unlock()
	if _, err := ch.writer.Write(data); err != nil {
		return err
	}
	if err := ch.writer.WriteByte('\n'); err != nil {
		return err
	}
	return ch.writer.Flush()
}

// Close libera el extremo del cliente (igual que ATMProxy.Close).
func (ch *Channel) Close() {
	if ch.conn != nil {
		ch.conn.Close()
		ch.conn = nil
	}
}
