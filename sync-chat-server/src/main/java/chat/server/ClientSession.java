package chat.server;

import chat.model.Json;
import chat.model.Message;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.io.PrintWriter;
import java.net.Socket;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.LinkedBlockingQueue;

/**
 * UN CANAL con un cliente: su socket, sus dos hilos y su buzón.
 *
 * DOS HILOS, no uno. Si el mismo hilo leyera y escribiera, mientras
 * está dormido en readLine() esperando a su usuario no podría
 * entregarle lo que otros escribieron. Un hilo, una dirección.
 *
 *   hilo lector   -> readLine() bloquea hasta que el usuario escribe
 *   hilo escritor -> take()     bloquea hasta que hay algo que enviar
 *
 * El BUZÓN (BlockingQueue) desacopla la región crítica de la red:
 * el coordinador deposita en memoria (microsegundos, nunca bloquea)
 * y este hilo escritor se encarga del socket, fuera del turno.
 * Sin esto, un cliente con red lenta congelaría a todos.
 */
public class ClientSession {

    private final Socket socket;
    private final BufferedReader in;
    private final PrintWriter out;

    /** Buzón de salida. Ya es seguro entre hilos: no necesita candado. */
    private final BlockingQueue<Message> outbox = new LinkedBlockingQueue<>();

    /** Quién es y a qué conversación pertenece. Los fija el JOIN. */
    private String username;
    private String conversationKey;

    private volatile boolean alive = true;

    public ClientSession(Socket socket) throws Exception {
        this.socket = socket;
        this.in  = new BufferedReader(new InputStreamReader(socket.getInputStream(), "UTF-8"));
        this.out = new PrintWriter(new java.io.OutputStreamWriter(socket.getOutputStream(), "UTF-8"), true);
    }

    // --- Identidad -------------------------------------------------------

    public String getUsername()        { return username; }
    public String getConversationKey() { return conversationKey; }
    public boolean isAlive()           { return alive; }
    public String remoteAddress()      { return socket.getRemoteSocketAddress().toString(); }

    public void identify(String username, String conversationKey) {
        this.username = username;
        this.conversationKey = conversationKey;
    }

    // --- Entrega ---------------------------------------------------------

    /**
     * Deposita en el buzón. Lo llama el COORDINADOR, dentro de la región
     * crítica. Debe ser rápido y no bloquear: por eso es una cola en
     * memoria y no una escritura al socket.
     */
    public void enqueue(Message msg) {
        if (alive) outbox.offer(msg);
    }

    // --- Los dos hilos ---------------------------------------------------

    /** Arranca ambos hilos. Los nombra para poder verlos en el depurador. */
    public void start(IChatService service) {
        Thread reader = new Thread(() -> readLoop(service), "lector-" + remoteAddress());
        Thread writer = new Thread(this::writeLoop,          "escritor-" + remoteAddress());
        reader.start();
        writer.start();
    }

    /** PRODUCTOR: recibe del socket y entrega al servicio. */
    private void readLoop(IChatService service) {
        try {
            String line;
            while (alive && (line = in.readLine()) != null) {   // bloquea aquí
                Message msg = Json.fromWire(line);
                if (msg == null) continue;                      // texto inválido: ignorar
                service.onMessage(this, msg);
            }
        } catch (Exception e) {
            // socket cerrado o error de red: se sale del bucle
        } finally {
            service.onDisconnect(this);
            close();
        }
    }

    /** CONSUMIDOR: saca del buzón y escribe al socket. */
    private void writeLoop() {
        try {
            while (alive) {
                Message msg = outbox.take();                    // bloquea aquí
                out.println(Json.toWire(msg));                  // el '\n' es el framing
                if (out.checkError()) break;                    // el cliente se fue
            }
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        } finally {
            close();
        }
    }

    // --- Cierre ----------------------------------------------------------

    /** Idempotente: puede llamarse desde ambos hilos sin problema. */
    public void close() {
        if (!alive) return;
        alive = false;
        outbox.offer(new Message());     // despierta al escritor si dormía
        try { socket.close(); } catch (Exception ignored) {}
    }
}