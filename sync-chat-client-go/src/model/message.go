package model

// Type identifica qué clase de mensaje viaja por el socket.
// Un único struct Message se usa para TODO: el campo Type dice cómo
// interpretar el resto.
type Type string

const (
	TypeJoin        Type = "JOIN"         // C->S: declara quién soy y a qué conversación pertenece ESTE canal
	TypeGroup       Type = "GROUP"        // C<->S: mensaje de la conversación grupal
	TypePrivate     Type = "PRIVATE"      // C<->S: mensaje de una conversación privada
	TypeUsers       Type = "USERS"        // S->C: lista de conectados (panel derecho)
	TypeOpenPrivate Type = "OPEN_PRIVATE" // S->C: alguien quiere hablarme en privado, debo abrir canal + ventana
	TypeError       Type = "ERROR"        // S->C: algo salió mal
)

// Message es el único tipo que cruza el socket, en ambas direcciones.
// Los campos que no aplican a un Type quedan vacíos y no se serializan.
//
// {"type":"JOIN","from":"userx"}                              -> canal grupal
// {"type":"JOIN","from":"userx","to":"usery"}                 -> canal privado, clave "userx|usery"
// {"type":"GROUP","from":"userx","body":"hola"}
// {"type":"PRIVATE","from":"userx","to":"usery","body":"oye"}
// {"type":"USERS","seq":41,"users":["user1","usery","userx"]}
// {"type":"OPEN_PRIVATE","seq":43,"from":"userx"}
// {"type":"ERROR","body":"..."}
type Message struct {
	Type  Type     `json:"type"`
	Seq   int64    `json:"seq,omitempty"`  // solo lo pone el servidor
	From  string   `json:"from,omitempty"`
	To    string   `json:"to,omitempty"`
	Body  string   `json:"body,omitempty"`
	Users []string `json:"users,omitempty"` // solo lo pone el servidor
}
