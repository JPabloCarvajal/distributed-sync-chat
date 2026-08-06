"""LoginWindow: pide servidor y usuario. Al conectar, construye la
ventana de grupo (la principal de la imagen) DENTRO de la misma ventana
raíz de Tk y reemplaza este formulario — equivalente a cerrar la ventana
de login y abrir la de grupo en el cliente de Go."""

from __future__ import annotations

import tkinter as tk
from tkinter import ttk

from src.client.chat_client import ChatClient
from src.ui.group_window import GroupWindow


class LoginWindow:
    def __init__(self, root: tk.Tk, client: ChatClient):
        self._root = root
        self._client = client

        root.title("Chat - Conectar")
        root.geometry("360x180")

        frame = ttk.Frame(root, padding=16)
        frame.pack(fill="both", expand=True)
        frame.columnconfigure(1, weight=1)

        ttk.Label(frame, text="Servidor").grid(row=0, column=0, sticky="w", pady=4)
        self._address = ttk.Entry(frame)
        self._address.insert(0, "127.0.0.1:1802")
        self._address.grid(row=0, column=1, sticky="ew", pady=4)

        ttk.Label(frame, text="Usuario").grid(row=1, column=0, sticky="w", pady=4)
        self._username = ttk.Entry(frame)
        self._username.grid(row=1, column=1, sticky="ew", pady=4)

        self._connect_btn = ttk.Button(frame, text="Conectar", command=self._connect)
        self._connect_btn.grid(row=2, column=0, columnspan=2, pady=8, sticky="ew")

        self._status = ttk.Label(frame, text="", foreground="red")
        self._status.grid(row=3, column=0, columnspan=2, sticky="w")

    def _connect(self) -> None:
        username = self._username.get().strip()
        if not username:
            self._status.config(text="El usuario no puede estar vacío")
            return

        address = self._address.get().strip()
        self._connect_btn.config(state="disabled")
        try:
            self._client.connect(address, username)
        except OSError as exc:
            self._status.config(text=f"No se pudo conectar: {exc}")
            self._connect_btn.config(state="normal")
            return

        for child in self._root.winfo_children():
            child.destroy()
        GroupWindow(self._root, self._client)
