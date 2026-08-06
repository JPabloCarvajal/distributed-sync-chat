package chat.server;

import java.util.ArrayList;
import java.util.List;
import java.util.Set;
import java.util.TreeSet;
import java.util.concurrent.CopyOnWriteArrayList;

/**
 * Los canales abiertos.
 *
 * RECURSO COMPARTIDO: lo modifican los hilos lectores (al entrar o salir
 * un usuario) y lo recorre el coordinador (en cada difusión). Si un hilo
 * lo modificara mientras otro lo recorre, saltaría
 * ConcurrentModificationException y ese hilo moriría.
 *
 * Se usa CopyOnWriteArrayList: cada escritura crea una copia interna,
 * así que recorrer NUNCA falla. Es ideal aquí porque se lee muchísimo
 * (cada mensaje) y se escribe poquísimo (entrar / salir).
 */
public class ClientRegistry {

    private final List<ClientSession> sessions = new CopyOnWriteArrayList<>();

    public void add(ClientSession session)    { sessions.add(session); }
    public void remove(ClientSession session) { sessions.remove(session); }

    /** Los canales de una conversación: destinatarios de un mensaje. */
    public List<ClientSession> byConversation(String key) {
        List<ClientSession> result = new ArrayList<>();
        for (ClientSession s : sessions)
            if (s.isAlive() && key.equals(s.getConversationKey()))
                result.add(s);
        return result;
    }

    /** Nombres conectados, sin repetir y ordenados: el panel derecho. */
    public List<String> usernames() {
        Set<String> names = new TreeSet<>();
        for (ClientSession s : sessions)
            if (s.isAlive() && s.getUsername() != null)
                names.add(s.getUsername());
        return new ArrayList<>(names);
    }

    /** El canal GRUPAL de un usuario: por ahí van los avisos de control. */
    public ClientSession groupChannelOf(String username) {
        for (ClientSession s : sessions)
            if (s.isAlive()
                && username.equals(s.getUsername())
                && "GROUP".equals(s.getConversationKey()))
                return s;
        return null;
    }
}