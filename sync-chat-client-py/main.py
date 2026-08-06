import tkinter as tk

from src.client.chat_client import ChatClient
from src.ui.login_window import LoginWindow


def main() -> None:
    root = tk.Tk()
    LoginWindow(root, ChatClient())
    root.mainloop()


if __name__ == "__main__":
    main()
