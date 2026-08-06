package chat.sync;

import chat.model.Message;
import java.util.Collection;
import java.util.function.Consumer;

/**
 * Contrato del COORDINADOR (Tanenbaum §6.3.2, algoritmo centralizado).
 *
 * Implementa la exclusión mutua sobre el recurso compartido del sistema:
 * el ORDEN DE LA CONVERSACIÓN. Solo un mensaje a la vez puede ocupar
 * la siguiente posición.
 *
 * Existen dos implementaciones intercambiables:
 *   - CentralizedSequencer : con región crítica (correcta)
 *   - UnsafeSequencer      : sin región crítica (para demostrar el fallo)
 */
public interface ISequencer {

    /**
     * Asigna número de secuencia y entrega el mensaje a cada destinatario,
     * como UNA SOLA operación indivisible.
     *
     * @param msg      mensaje a numerar
     * @param destinos a quién va (uno, dos o todos)
     * @param deliver  qué hacer con cada destinatario
     * @return el número asignado
     */
    <T> long stampAndDeliver(Message msg, Collection<T> destinos, Consumer<T> deliver);
}