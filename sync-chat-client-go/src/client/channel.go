package client

import (
	"bufio"
	"encoding/json"
	"net"
	"sync"

	"chat-client/src/model"
)

// Channel es UN socket TCP. En este protocolo cada conversación tiene el
// suyo: el canal grupal es un Channel, y cada chat privado abre OTRO
// Channel independiente (con su propio JOIN{from,to}). Channel no sabe
// si es "el grupal" o "un privado con X": eso lo decide quien lo usa.
//
// "No bloqueante" se resuelve como en el cajero-cliente: una goroutine
// dedicada hace el ReadBytes bloqueante a nivel de SO mientras el resto
// del programa sigue vivo, y publica cada mensaje ya parseado en un
// channel de Go. Quien consume nunca se bloquea en el socket.
type Channel struct {
	conn   net.Conn
	writer *bufio.Writer
	wmu    sync.Mutex // Send puede llamarse desde la goroutine de UI mientras readLoop sigue leyendo

	incoming chan model.Message
}

// Dial abre el socket y arranca su goroutine de lectura.
// Todavía no manda JOIN: eso lo decide el llamador (grupal vs. privado).
func Dial(address string) (*Channel, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	ch := &Channel{
		conn:     conn,
		writer:   bufio.NewWriter(conn),
		incoming: make(chan model.Message, 64),
	}
	go ch.readLoop(bufio.NewReader(conn))
	return ch, nil
}

// readLoop es la única goroutine que lee este socket.
// Termina sola (cerrando incoming) cuando el servidor cierra la conexión.
func (ch *Channel) readLoop(reader *bufio.Reader) {
	defer close(ch.incoming)
	for {
		line, err := reader.ReadBytes('\n')
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

func (ch *Channel) Close() {
	if ch.conn != nil {
		ch.conn.Close()
	}
}
