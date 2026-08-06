"""Channel: UN socket TCP. Mismo rol que el Channel del cliente de Go
(y que ATMProxy en el cajero): el constructor solo recuerda la dirección,
open() conecta, close() libera.

Aquí la conversación no es ping-pong (una petición, una respuesta) sino que
el servidor puede escribir en cualquier momento, así que hace falta un
hilo de lectura continuo. "No bloqueante" se resuelve con el patrón
clásico de hilos: un hilo dedicado hace el recv() bloqueante a nivel de SO
mientras el resto del programa sigue vivo, y publica cada mensaje ya
parseado en una queue.Queue (thread-safe por diseño: es justamente el
mecanismo de sincronización entre el hilo lector y quien consume).

En este protocolo cada conversación tiene SU PROPIO Channel: el canal
grupal es uno, y cada chat privado abre otro independiente (con su propio
JOIN{from,to}). Channel no sabe si es "el grupal" o "un privado con X":
eso lo decide quien lo usa.
"""

from __future__ import annotations

import json
import queue
import socket
import threading
from typing import Optional

from src.model.message import Message


class Channel:
    def __init__(self, address: str):
        host, port = address.split(":")
        self._host = host
        self._port = int(port)
        self._sock: Optional[socket.socket] = None
        self._write_lock = threading.Lock()  # send() puede llamarse desde el hilo de UI mientras _read_loop sigue leyendo
        self.incoming: "queue.Queue[Optional[Message]]" = queue.Queue()

    def open(self) -> None:
        self._sock = socket.create_connection((self._host, self._port))
        threading.Thread(target=self._read_loop, daemon=True).start()

    def _read_loop(self) -> None:
        """Única hilo que lee este socket. Termina solo (dejando un
        centinela None en incoming) cuando el servidor cierra la conexión.
        """
        buffer = b""
        try:
            while True:
                chunk = self._sock.recv(4096)
                if not chunk:
                    return
                buffer += chunk
                while b"\n" in buffer:
                    line, buffer = buffer.split(b"\n", 1)
                    if not line:
                        continue
                    try:
                        self.incoming.put(Message.from_json(line.decode("utf-8")))
                    except (json.JSONDecodeError, UnicodeDecodeError):
                        continue  # línea corrupta: se descarta, no se tumba la conexión
        except OSError:
            return
        finally:
            self.incoming.put(None)

    def send(self, message: Message) -> None:
        data = (message.to_json() + "\n").encode("utf-8")
        with self._write_lock:
            self._sock.sendall(data)

    def close(self) -> None:
        if self._sock is not None:
            try:
                self._sock.close()
            finally:
                self._sock = None
