# distributed-sync-chat

## Cliente Go (`sync-chat-client-go`)

Usa [Fyne](https://fyne.io) para la interfaz gráfica, que internamente
necesita **cgo** (compilar código C) y las librerías gráficas del sistema
(OpenGL/X11). Sin eso, `go build`/`go run` falla o no descarga bien las
dependencias. Antes de compilar, instala lo de tu sistema:

**macOS**
```bash
xcode-select --install
```

**Linux (Debian/Ubuntu)**
```bash
sudo apt install gcc libgl1-mesa-dev xorg-dev
```

**Linux (Fedora)**
```bash
sudo dnf install gcc libXcursor-devel libXrandr-devel mesa-libGL-devel libXi-devel libXinerama-devel
```

**Windows**
Windows no trae compilador de C. La forma más simple es instalar [MSYS2](https://www.msys2.org),
abrir la terminal "MSYS2 UCRT64" y correr:
```bash
pacman -S mingw-w64-ucrt-x86_64-gcc
```
y agregar `C:\msys64\ucrt64\bin` al PATH del sistema.

También hace falta **Go 1.22 o más nuevo** (`go version` para revisar).

Para compilar y correr:
```bash
cd sync-chat-client-go
go mod download
go run .
```

`go mod download` necesita internet la primera vez (baja los paquetes de
Fyne). Si tu compañero está detrás de una red corporativa que bloquea
`proxy.golang.org`, prueba con:
```bash
go env -w GOPROXY=https://goproxy.io,direct
```

> Nota: el `go.mod` antes pedía `go 1.26.5` (la versión que tenía instalada
> yo), mucho más nueva de lo que el código realmente necesita (Fyne solo
> pide 1.22). Si tu Go es más viejo que eso, `go` intenta bajar sola una
> versión nueva por internet — y si esa descarga falla es indistinguible
> de "no compila ni descarga las dependencias". Ya está bajado a 1.22.

## Cliente Python (`sync-chat-client-py`)

Usa `tkinter`, que viene en la librería estándar de Python — no hay que
instalar nada con `pip`. La única condición es que Python tenga tkinter
disponible:

- **Windows / macOS** (instalador oficial de [python.org](https://python.org)): ya lo incluye.
- **Linux**: en varias distros hay que instalarlo aparte:
  ```bash
  sudo apt install python3-tk      # Debian/Ubuntu
  sudo dnf install python3-tkinter # Fedora
  ```

Se necesita **Python 3.10+**. Para correr:
```bash
cd sync-chat-client-py
python3 main.py
```
