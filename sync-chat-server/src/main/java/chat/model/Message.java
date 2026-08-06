package chat.model;

import java.util.List;

/**
 * El ÚNICO objeto que viaja por el socket.
 *
 * Los nombres de estos campos son las llaves del JSON, así que
 * cambiarlos rompe la comunicación con los clientes Go y Python
 * SIN QUE NADA DEJE DE COMPILAR. Es el precio de estar distribuido.
 *
 * Campos no aplicables van en null: Gson los omite del JSON.
 */
public class Message {

    private MessageType type;
    private long seq;              // SOLO lo asigna el servidor
    private String from;
    private String to;             // destinatario, en conversaciones privadas
    private String body;
    private List<String> users;    // SOLO lo asigna el servidor

    /** Gson necesita un constructor sin argumentos para deserializar. */
    public Message() {}

    private Message(MessageType type, String from, String to, String body) {
        this.type = type;
        this.from = from;
        this.to   = to;
        this.body = body;
    }

    // --- Fábricas: solo las formas válidas del protocolo -----------------

    public static Message chat(MessageType type, String from, String to, String body) {
        return new Message(type, from, to, body);
    }

    public static Message users(List<String> users) {
        Message m = new Message(MessageType.USERS, null, null, null);
        m.users = users;
        return m;
    }

    public static Message openPrivate(String from) {
        return new Message(MessageType.OPEN_PRIVATE, from, null, null);
    }

    public static Message error(String body) {
        return new Message(MessageType.ERROR, null, null, body);
    }

    // --- Clave de conversación ------------------------------------------

    /**
     * Identifica a qué conversación pertenece este mensaje.
     *
     * Grupal  -> "GROUP"
     * Privada -> los dos nombres ORDENADOS y unidos por '|'
     *
     * El orden alfabético es lo que hace que userx→usery y usery→userx
     * produzcan la misma clave, y por tanto se emparejen en la misma
     * conversación. Sin ordenar, serían dos conversaciones distintas.
     */
    public String conversationKey() {
        if (to == null || to.isBlank()) return "GROUP";
        return from.compareTo(to) < 0 ? from + "|" + to
                                      : to + "|" + from;
    }

    // --- Acceso ----------------------------------------------------------

    public MessageType getType()   { return type; }
    public long        getSeq()    { return seq; }
    public String      getFrom()   { return from; }
    public String      getTo()     { return to; }
    public String      getBody()   { return body; }
    public List<String> getUsers() { return users; }

    /** Lo llama ÚNICAMENTE el coordinador, dentro de la región crítica. */
    public void setSeq(long seq) { this.seq = seq; }
}