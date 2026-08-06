package ui

import "sync"

// privateRegistry recuerda con qué usuarios hay ventana privada abierta,
// para no crear dos con el mismo.
//
// Recurso compartido: lo tocan la goroutine de UI (clic en la lista) y la
// goroutine que atiende OPEN_PRIVATE. Por eso lleva mutex.
type privateRegistry struct {
	mu   sync.Mutex
	open map[string]bool
}

var privates = &privateRegistry{open: map[string]bool{}}

// claim reserva 'peer'. Devuelve false si ya había ventana con él.
// Reservar y comprometer sin hueco en medio: mismo razonamiento que la
// región crítica del servidor.
func (r *privateRegistry) claim(peer string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.open[peer] {
		return false
	}
	r.open[peer] = true
	return true
}

func (r *privateRegistry) release(peer string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.open, peer)
}
