"""ChatClient: el orquestador de dominio (mismo rol que ChatClient en el
cliente de Go). No sabe de sockets (eso es Channel) ni de widgets (eso es
la UI).

Solo el canal GRUPAL vive aquí, porque es el único que existe siempre
(plano de control: JOIN inicial, GROUP, USERS, OPEN_PRIVATE, ERROR). Cada
chat privado es un Channel aparte que ChatClient sabe ABRIR (open_private),
pero cuya vida entera (leer, mostrar, cerrar) la dueña es la ventana
privada que lo usa — así se respeta "cerrar una ventana privada solo mata
SU socket".
"""

from __future__ import annotations

import threading
from typing import Callable, Optional

from src.client.proxy import Channel
from src.model.message import Message, Type


class ChatClient:
    def __init__(self) -> None:
        self._address = ""
        self.username = ""
        self._group: Optional[Channel] = None
        self._last_users: list[str] | None = None

        self._lock = threading.Lock()
        self._group_handlers: list[Callable[[Message], None]] = []
        self._users_handlers: list[Callable[[list[str]], None]] = []
        self._error_handlers: list[Callable[[Message], None]] = []
        self._open_private_handlers: list[Callable[[str, Channel], None]] = []

    def connect(self, address: str, username: str) -> None:
        """Abre el canal grupal: JOIN sin 'to'. A partir de aquí queda vivo
        hasta close(). Guarda la dirección porque cada privado abre OTRO
        socket contra el mismo servidor."""
        channel = Channel(address)
        channel.open()
        channel.send(Message(type=Type.JOIN, from_=username))

        self._address = address
        self.username = username
        self._group = channel

        threading.Thread(target=self._dispatch_group, daemon=True).start()

    def send_group(self, body: str) -> None:
        self._group.send(Message(type=Type.GROUP, from_=self.username, body=body))

    def open_private(self, peer: str) -> Channel:
        """Abre un socket NUEVO y dedicado con peer: JOIN{from, to}. Lo usa
        la UI en los dos lados del flujo C: cuando el usuario local hace
        clic sobre alguien (inicia), y cuando llega un OPEN_PRIVATE
        (responde)."""
        channel = Channel(self._address)
        channel.open()
        channel.send(Message(type=Type.JOIN, from_=self.username, to=peer))
        return channel

    def on_group_message(self, handler: Callable[[Message], None]) -> None:
        with self._lock:
            self._group_handlers.append(handler)

    def on_users(self, handler: Callable[[list[str]], None]) -> None:
        """Si el USERS ya llegó antes de registrarse (carrera de arranque),
        se reproduce el último conocido."""
        with self._lock:
            self._users_handlers.append(handler)
            last = self._last_users

        if last is not None:
            handler(last)

    def on_error(self, handler: Callable[[Message], None]) -> None:
        with self._lock:
            self._error_handlers.append(handler)

    def on_open_private(self, handler: Callable[[str, Channel], None]) -> None:
        """Se dispara ya con el canal privado ABIERTO Y EMPAREJADO (JOIN ya
        enviado): la UI solo tiene que pintar la ventana encima."""
        with self._lock:
            self._open_private_handlers.append(handler)

    def _dispatch_group(self) -> None:
        """Único hilo que consume el canal grupal. Nunca toca widgets: solo
        llama handlers, responsabilidad de la UI."""
        while True:
            msg = self._group.incoming.get()
            if msg is None:
                return

            if msg.type == Type.GROUP:
                self._notify(self._group_handlers, msg)
            elif msg.type == Type.USERS:
                users = msg.users or []
                with self._lock:
                    self._last_users = users
                self._notify(self._users_handlers, users)   
            elif msg.type == Type.OPEN_PRIVATE:
                # en hilo aparte: abrir el socket privado implica un connect
                # (bloqueante) y no puede frenar la lectura del canal grupal.
                threading.Thread(
                    target=self._handle_open_private, args=(msg.from_,), daemon=True
                ).start()
            elif msg.type == Type.ERROR:
                self._notify(self._error_handlers, msg)

    def _handle_open_private(self, peer: str) -> None:
        try:
            channel = self.open_private(peer)
        except OSError:
            return
        self._notify(self._open_private_handlers, peer, channel)

    def _notify(self, handlers: list[Callable], *args) -> None:
        with self._lock:
            snapshot = list(handlers)
        for handler in snapshot:
            handler(*args)

    def close(self) -> None:
        """Cierra el canal grupal. Los privados los cierra cada ventana."""
        if self._group is not None:
            self._group.close()
