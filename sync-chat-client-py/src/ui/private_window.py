"""PrivateWindow: chat 1-a-1 sobre un Channel que YA llega conectado y con
el JOIN de emparejamiento enviado (lo abrió ChatClient, por clic propio o
por OPEN_PRIVATE). Esta ventana es la ÚNICA dueña de ese socket: lo lee
hasta que se cierra, y cerrar la ventana cierra SOLO ese canal (flujo E) —
el otro lado se entera porque su propio recv() falla, no porque nosotros
le mandemos nada."""

from __future__ import annotations

import threading
import tkinter as tk
from tkinter import ttk

from src.client.chat_client import ChatClient
from src.client.proxy import Channel
from src.model.message import Message, Type
from src.ui.chat_log import ChatLog
from src.ui.private_registry import privates


class PrivateWindow:
    def __init__(self, root: tk.Tk, client: ChatClient, peer: str, channel: Channel):
        self._client = client
        self._peer = peer
        self._channel = channel

        window = tk.Toplevel(root)
        self._window = window
        window.title(f"{client.username} -> {peer}")
        window.geometry("420x420")
        window.protocol("WM_DELETE_WINDOW", self._on_close)

        ttk.Label(window, text=peer, font=("TkDefaultFont", 12, "bold")).pack(
            anchor="w", padx=8, pady=4
        )

        self._log = ChatLog(window)
        self._log.pack(fill="both", expand=True, padx=8)

        bottom = ttk.Frame(window)
        bottom.pack(fill="x", padx=8, pady=8)

        self._input = ttk.Entry(bottom)
        self._input.pack(side="left", fill="x", expand=True)
        self._input.bind("<Return>", lambda _event: self._send())

        ttk.Button(bottom, text="Send", command=self._send).pack(side="left", padx=(6, 0))

        threading.Thread(target=self._read_loop, daemon=True).start()

    def _send(self) -> None:
        body = self._input.get().strip()
        if not body:
            return
        self._channel.send(
            Message(type=Type.PRIVATE, from_=self._client.username, to=self._peer, body=body)
        )
        self._input.delete(0, "end")

    def _read_loop(self) -> None:
        while True:
            msg = self._channel.incoming.get()
            if msg is None:
                return

            if msg.type == Type.PRIVATE:
                line = f"{msg.from_}: {msg.body}"
            elif msg.type == Type.ERROR:
                line = f"[ERROR] {msg.body}"
            else:
                continue

            self._window.after(0, lambda line=line: self._log.append(line))

    def _on_close(self) -> None:
        self._channel.close()
        privates.release(self._peer)
        self._window.destroy()
