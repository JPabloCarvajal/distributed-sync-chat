# distributed-sync-chat

Chat con un servidor central (Java) y dos clientes (Go y Python) a elección.

## Servidor (`sync-chat-server`)

Requiere **Java 25+** y Maven.
```bash
cd sync-chat-server
mvn package
mvn compile exec:java  

java -jar target/sync-chat-server-1.0.jar
```
Escucha por defecto en el puerto `1802`.

## Cliente Go (`sync-chat-client-go`)

Requiere **Go 1.22+** y un compilador de C (lo usa la librería gráfica Fyne):

- **macOS**: `xcode-select --install`
- **Linux (Debian/Ubuntu)**: `sudo apt install gcc libgl1-mesa-dev xorg-dev`
- **Linux (Fedora)**: `sudo dnf install gcc libXcursor-devel libXrandr-devel mesa-libGL-devel libXi-devel libXinerama-devel`
- **Windows**: instalar [MSYS2](https://www.msys2.org), abrir "MSYS2 UCRT64" y correr `pacman -S mingw-w64-ucrt-x86_64-gcc`, luego agregar `C:\msys64\ucrt64\bin` al PATH.

Para correr:
```bash
cd sync-chat-client-go
go run .
```

## Cliente Python (`sync-chat-client-py`)

Requiere **Python 3.10+** con `tkinter` (viene incluido en Windows/macOS; en Linux hay que instalarlo aparte):
```bash
sudo apt install python3-tk      # Debian/Ubuntu
sudo dnf install python3-tkinter # Fedora
```

Para correr (no necesita `pip install`, solo librería estándar):
```bash
cd sync-chat-client-py
python3 main.py
```
