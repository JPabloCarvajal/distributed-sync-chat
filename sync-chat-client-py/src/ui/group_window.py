"""GroupWindow: la ventana principal de la imagen. Título con el usuario,
chat de grupo a la izquierda, lista de usuarios conectados a la derecha e
input + Send abajo.

Un clic en un usuario ABRE UN SOCKET NUEVO (flujo C, paso 1):
JOIN{from: yo, to: peer}, y con ese canal ya conectado pinta la ventana
privada. Si llega un OPEN_PRIVATE (alguien más nos invita, flujo C paso 2)
ChatClient ya dejó el canal abierto y emparejado; aquí solo se pinta la
ventana. En ningún caso se reutiliza una ventana/canal existente: cada
apertura -por clic propio o por invitación- crea uno nuevo.
"""
from src.ui.private_registry import privates
from __future__ import annotations

import threading
import tkinter as tk
from tkinter import messagebox, ttk

from src.client.chat_client import ChatClient
from src.client.proxy import Channel
from src.model.message import Message
from src.ui.chat_log import ChatLog
from src.ui.private_window import PrivateWindow


class GroupWindow:
    def __init__(self, root: tk.Tk, client: ChatClient):
        self._root = root
        self._client = client
        self._users: list[str] = []

        root.title(client.username)
        root.geometry("700x480")
        root.protocol("WM_DELETE_WINDOW", self._on_close)

        ttk.Label(
            root, text=client.username, font=("TkDefaultFont", 12, "bold")
        ).pack(anchor="w", padx=8, pady=4)

        center = ttk.Frame(root)
        center.pack(fill="both", expand=True, padx=8)

        self._log = ChatLog(center)
        self._log.pack(side="left", fill="both", expand=True)

        sidebar = ttk.Frame(center, width=140)
        sidebar.pack(side="right", fill="y")
        sidebar.pack_propagate(False)

        self._user_list = tk.Listbox(sidebar)
        self._user_list.pack(fill="both", expand=True)
        self._user_list.bind("<<ListboxSelect>>", self._on_user_selected)

        bottom = ttk.Frame(root)
        bottom.pack(fill="x", padx=8, pady=8)

        self._input = ttk.Entry(bottom)
        self._input.pack(side="left", fill="x", expand=True)
        self._input.bind("<Return>", lambda _event: self._send())

        ttk.Button(bottom, text="Send", command=self._send).pack(side="left", padx=(6, 0))

        client.on_group_message(self._handle_group_message)
        client.on_users(self._handle_users)
        client.on_error(self._handle_error)
        client.on_open_private(self._handle_open_private)

    def _send(self) -> None:
        body = self._input.get().strip()
        if not body:
            return
        self._client.send_group(body)
        self._input.delete(0, "end")

    def _on_user_selected(self, _event: tk.Event) -> None:
        selection = self._user_list.curselection()
        if not selection:
            return
        peer = self._users[selection[0]]
        self._user_list.selection_clear(0, "end")
        if peer == self._client.username:
            return
        if not privates.claim(peer):
            return  # ya hay ventana abierta con él

        def open_channel() -> None:
            try:
                channel = self._client.open_private(peer)
            except OSError as exc:
                privates.release(peer)
                self._root.after(0, lambda: messagebox.showerror("Error", str(exc)))
                return
            self._root.after(0, lambda: PrivateWindow(self._root, self._client, peer, channel))

        threading.Thread(target=open_channel, daemon=True).start()

    def _handle_group_message(self, msg: Message) -> None:
        self._root.after(0, lambda: self._log.append(f"{msg.from_}: {msg.body}"))

    def _handle_users(self, users: list[str]) -> None:
        def update() -> None:
            self._users = users
            self._user_list.delete(0, "end")
            for user in users:
                self._user_list.insert("end", user)

        self._root.after(0, update)

    def _handle_error(self, msg: Message) -> None:
        self._root.after(0, lambda: self._log.append(f"[ERROR] {msg.body}"))

    def _handle_open_private(self, peer: str, channel: Channel) -> None:
        if not privates.claim(peer):
            channel.close()  # ya tenemos ventana con él: este canal sobra
            return
        self._root.after(0, lambda: PrivateWindow(self._root, self._client, peer, channel))

    def _on_close(self) -> None:
        self._channel.close()
        privates.release(self._peer)
        self._window.destroy()
