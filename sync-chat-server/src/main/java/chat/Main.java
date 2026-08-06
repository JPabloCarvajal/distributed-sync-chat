package chat;

import chat.server.ChatServer;
import chat.server.ChatService;
import chat.server.ClientRegistry;
import chat.sync.CentralizedSequencer;
import chat.sync.ISequencer;

/**
 * Cableado del servidor. Aquí se decide QUÉ implementación se inyecta;
 * ninguna clase de abajo conoce a las otras por su tipo concreto.
 */
public class Main {

    private static final int PORT = 1802;

    public static void main(String[] args) throws Exception {

        // 1. El coordinador: exclusión mutua sobre el orden (§6.3.2)
        ISequencer coordinator = new CentralizedSequencer();

        // 2. El registro de canales abiertos
        ClientRegistry registry = new ClientRegistry();

        // 3. La lógica: decide destinatarios, delega el orden
        ChatService service = new ChatService(registry, coordinator);

        // 4. El transporte: acepta conexiones y lanza hilos
        new ChatServer(PORT, service).start();
    }
}   