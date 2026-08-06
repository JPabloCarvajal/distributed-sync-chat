package chat.server;

import chat.model.Message;
import chat.sync.ISequencer;

import java.util.List;

/**
 * Decide QUIÉN recibe cada mensaje. No sabe de sockets ni de concurrencia.
 *
 * Reparto de responsabilidades:
 *   ChatService  -> a quién va      (lógica de chat)
 *   ISequencer   -> en qué orden    (exclusión mutua, §6.3.2)  
 *   ClientSession-> cómo se entrega (buzón + socket)
 */
public class ChatService implements IChatService {

    private final ClientRegistry registry;
    private final ISequencer coordinator;

    public ChatService(ClientRegistry registry, ISequencer coordinator) {
        this.registry = registry;
        this.coordinator = coordinator;
    }

    @Override
    public void onMessage(ClientSession session, Message msg) {
        switch (msg.getType()) {
            case JOIN            -> handleJoin(session, msg);
            case GROUP, PRIVATE  -> handleChat(session, msg);
            default              -> { /* tipos que solo el servidor emite */ }
        }
    }

    /** Un canal declara quién es y a qué conversación pertenece. */
    private void handleJoin(ClientSession session, Message msg) {
        if (msg.getFrom() == null || msg.getFrom().isBlank()) {
            session.enqueue(Message.error("nombre de usuario requerido"));
            return;
        }

        String key = msg.conversationKey();
        session.identify(msg.getFrom(), key);
        registry.add(session);

        System.out.printf("JOIN  %s en [%s]%n", msg.getFrom(), key);

        if ("GROUP".equals(key)) {
            broadcastUsers();
        } else if (!hasChannel(key, msg.getTo())) {
            // El otro NO tiene todavía canal en esta conversación:
            // soy quien inicia, hay que invitarlo.
            //
            // Si YA lo tiene, este JOIN es su respuesta a una invitación
            // previa. Invitarlo otra vez provocaría que él nos invitara
            // a nosotros, y así infinitamente.
            notifyPeer(msg.getFrom(), msg.getTo());
        }
    }

    /** Un mensaje de conversación: se numera y se reparte. */
    private void handleChat(ClientSession session, Message msg) {
        if (session.getConversationKey() == null) return;   // no hizo JOIN

        List<ClientSession> destinos =
                registry.byConversation(session.getConversationKey());

        coordinator.stampAndDeliver(msg, destinos, s -> s.enqueue(msg));
    }

    @Override
    public void onDisconnect(ClientSession session) {
        registry.remove(session);
        if (session.getUsername() == null) return;

        System.out.printf("SALE  %s de [%s]%n",
                session.getUsername(), session.getConversationKey());

        if ("GROUP".equals(session.getConversationKey()))
            broadcastUsers();
    }

    // --- Avisos del servidor: van por los canales GRUPALES ---------------

    /** Difunde la lista de conectados: alimenta el panel derecho. */
    private void broadcastUsers() {
        Message msg = Message.users(registry.usernames());
        List<ClientSession> destinos = registry.byConversation("GROUP");
        coordinator.stampAndDeliver(msg, destinos, s -> s.enqueue(msg));
    }

    /**
     * Avisa a 'peer' de que 'from' quiere hablarle en privado.
     * Va por su canal GRUPAL, porque el privado aún no existe:
     * el servidor no puede empujar por un socket que nadie ha abierto.
     */
    private void notifyPeer(String from, String peer) {
        ClientSession target = registry.groupChannelOf(peer);
        if (target == null) return;                    // no está conectado

        Message msg = Message.openPrivate(from);
        coordinator.stampAndDeliver(msg, List.of(target), s -> s.enqueue(msg));
    }

    /** ¿Tiene ya 'username' un canal en esta conversación? */
    private boolean hasChannel(String key, String username) {
        for (ClientSession s : registry.byConversation(key))
            if (username.equals(s.getUsername())) return true;
        return false;
    }
}