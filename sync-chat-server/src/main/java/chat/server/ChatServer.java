package chat.server;

import java.net.ServerSocket;
import java.net.Socket;

/**
 * PUERTO DE CONTROL: escucha y acepta conexiones.
 *
 * El socket de escucha NUNCA transporta mensajes de chat: solo espera
 * handshakes TCP. Cada accept() devuelve un socket NUEVO y distinto,
 * identificado por la 4-tupla (IP cliente, puerto cliente, IP servidor,
 * 1802). Por eso N clientes comparten el puerto 1802 sin pisarse.
 *
 * Este hilo SOLO acepta. La conversación la llevan los dos hilos que
 * arranca cada ClientSession. Si aquí se conversara, mientras se atiende
 * a un cliente nadie escucharía y los demás quedarían colgados.
 */
public class ChatServer {

    private final int port;
    private final IChatService service;

    public ChatServer(int port, IChatService service) {
        this.port = port;
        this.service = service;
    }

    public void start() throws Exception {
        try (ServerSocket listener = new ServerSocket(port)) {
            System.out.println("Puerto de control escuchando en :" + port);

            while (true) {
                Socket socket = listener.accept();   // aquí NACE el canal
                System.out.println("Canal abierto con " + socket.getRemoteSocketAddress());

                try {
                    new ClientSession(socket).start(service);   // 2 hilos propios
                } catch (Exception e) {
                    System.out.println("no se pudo abrir el canal: " + e.getMessage());
                    try { socket.close(); } catch (Exception ignored) {}
                }
            }
        }
    }
}