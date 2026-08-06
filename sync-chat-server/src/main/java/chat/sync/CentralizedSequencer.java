package chat.sync;

import chat.model.Message;
import java.util.Collection;
import java.util.concurrent.Semaphore;
import java.util.function.Consumer;

/**
 * COORDINADOR CENTRALIZADO con exclusión mutua.
 *
 * Tanenbaum §6.3.2: "la manera más directa de lograr la exclusión mutua
 * es simular lo que se hace en un sistema de un procesador. Se elige un
 * proceso como coordinador."
 *
 * Este servidor es ese coordinador. Cada cliente tiene un hilo lector que
 * lo representa; el semáforo pone esos hilos en fila.
 *
 * Correspondencia con el algoritmo del libro:
 *   PETICIÓN                 -> turno.acquire()
 *   cola de peticiones       -> cola interna del semáforo
 *   AUTORIZACIÓN             -> acquire() retorna
 *   región crítica           -> numerar + entregar
 *   LIBERACIÓN               -> turno.release()
 *   "un proceso a la vez"    -> 1 permiso
 *   "las peticiones son autorizadas en el orden en que se reciben"
 *                            -> fair = true
 */
public class CentralizedSequencer implements ISequencer {

    /** El recurso compartido: la siguiente posición de la conversación. */
    private long seq = 0;

    /**
     * Semáforo binario JUSTO.
     *   1     -> un solo permiso: un hilo a la vez en la región crítica
     *   true  -> equidad: los hilos entran en el orden en que llegaron
     *
     * La equidad NO es opcional aquí: es la propiedad de justicia que el
     * libro atribuye al algoritmo centralizado. 'synchronized' no la
     * garantiza; la JVM puede despertar los hilos en cualquier orden.
     */
    private final Semaphore turno = new Semaphore(1, true);

    @Override
    public <T> long stampAndDeliver(Message msg, Collection<T> destinos, Consumer<T> deliver) {
        try {
            turno.acquire();                  // PETICIÓN + AUTORIZACIÓN

            // ─── REGIÓN CRÍTICA ──────────────────────────────────────
            long n = ++seq;                   // tomar la posición
            msg.setSeq(n);
            for (T destino : destinos)        // y comprometerla en TODOS
                deliver.accept(destino);      // los buzones, sin soltar
            return n;
            // ─────────────────────────────────────────────────────────

        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            return -1;
        } finally {
            turno.release();                  // LIBERACIÓN
        }
    }
}