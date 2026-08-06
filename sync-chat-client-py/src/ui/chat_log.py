"""ChatLog: la caja blanca de mensajes de la imagen. Líneas que se apilan
hacia abajo y hacen autoscroll. La comparten la ventana de grupo y cada
privada, igual que chatLog en el cliente de Go."""

from __future__ import annotations

import tkinter as tk
from tkinter import scrolledtext


class ChatLog(scrolledtext.ScrolledText):
    def __init__(self, master: tk.Widget):
        super().__init__(master, state="disabled", wrap="word")

    def append(self, line: str) -> None:
        self.configure(state="normal")
        self.insert("end", line + "\n")
        self.configure(state="disabled")
        self.see("end")
