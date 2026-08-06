package chat.model;

/**
 * Los tipos de mensaje del protocolo.
 *
 * Este enum NO viaja como objeto: se serializa como texto ("GROUP",
 * "PRIVATE"...). Es el discriminador que permite decidir qué hacer
 * con un mensaje entrante.
 */
public enum MessageType {

    /** C→S. Declara quién soy y a qué conversación pertenece este canal. */
    JOIN,

    /** C↔S. Mensaje de la conversación grupal. */
    GROUP,

    /** C↔S. Mensaje de una conversación privada. */
    PRIVATE,

    /** S→C. Lista actualizada de conectados. Alimenta el panel derecho. */
    USERS,

    /** S→C. Aviso: alguien quiere hablarte en privado, abre una ventana. */
    OPEN_PRIVATE,

    /** S→C. Algo salió mal. */
    ERROR
}