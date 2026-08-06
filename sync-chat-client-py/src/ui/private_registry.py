"""Recuerda con qué usuarios hay ventana privada abierta, para no crear
dos con el mismo.

Recurso compartido: lo tocan el hilo de UI (clic en la lista) y el hilo
que atiende OPEN_PRIVATE. Por eso lleva Lock.
"""

from __future__ import annotations

import threading


class PrivateRegistry:
    def __init__(self) -> None:
        self._lock = threading.Lock()
        self._open: set[str] = set()

    def claim(self, peer: str) -> bool:
        """Reserva 'peer'. False si ya había ventana con él."""
        with self._lock:
            if peer in self._open:
                return False
            self._open.add(peer)
            return True

    def release(self, peer: str) -> None:
        with self._lock:
            self._open.discard(peer)


privates = PrivateRegistry()