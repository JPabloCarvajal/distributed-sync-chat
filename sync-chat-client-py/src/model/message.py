"""Message: el único tipo que cruza el socket, en ambas direcciones.
Equivalente exacto de model.Message en el cliente de Go.

{"type":"JOIN","from":"userx"}                              -> canal grupal
{"type":"JOIN","from":"userx","to":"usery"}                 -> canal privado, clave "userx|usery"
{"type":"GROUP","from":"userx","body":"hola"}
{"type":"PRIVATE","from":"userx","to":"usery","body":"oye"}
{"type":"USERS","seq":41,"users":["user1","usery","userx"]}
{"type":"OPEN_PRIVATE","seq":43,"from":"userx"}
{"type":"ERROR","body":"..."}
"""

from __future__ import annotations

import json
from dataclasses import dataclass
from typing import Optional


class Type:
    JOIN = "JOIN"                  # C->S: declara quién soy y a qué conversación pertenece ESTE canal
    GROUP = "GROUP"                # C<->S: mensaje de la conversación grupal
    PRIVATE = "PRIVATE"            # C<->S: mensaje de una conversación privada
    USERS = "USERS"                # S->C: lista de conectados (panel derecho)
    OPEN_PRIVATE = "OPEN_PRIVATE"  # S->C: alguien quiere hablarme en privado, debo abrir canal + ventana
    ERROR = "ERROR"                # S->C: algo salió mal


@dataclass
class Message:
    type: str
    seq: Optional[int] = None            # solo lo pone el servidor
    from_: Optional[str] = None
    to: Optional[str] = None
    body: Optional[str] = None
    users: Optional[list[str]] = None    # solo lo pone el servidor

    def to_json(self) -> str:
        data = {"type": self.type}
        if self.seq is not None:
            data["seq"] = self.seq
        if self.from_ is not None:
            data["from"] = self.from_
        if self.to is not None:
            data["to"] = self.to
        if self.body is not None:
            data["body"] = self.body
        if self.users is not None:
            data["users"] = self.users
        return json.dumps(data)

    @staticmethod
    def from_json(line: str) -> "Message":
        data = json.loads(line)
        return Message(
            type=data.get("type", ""),
            seq=data.get("seq"),
            from_=data.get("from"),
            to=data.get("to"),
            body=data.get("body"),
            users=data.get("users"),
        )
