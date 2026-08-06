package chat.server;

import chat.model.Message;

/**
 * Contrato de qué hacer con lo que llega por un canal.
 *
 * Lo consume ClientSession (sus hilos lo invocan).
 * Lo provee ChatService.
 *
 * Gracias a esta interfaz, ClientSession NO sabe nada de chat: sabe
 * leer líneas, escribir líneas y avisar. Es el mismo principio que en
 * el ATM, donde TCPServer dependía de model.ATM y no de ATMService.
 */
public interface IChatService {

    /** Llegó un mensaje por este canal. */
    void onMessage(ClientSession session, Message msg);

    /** El canal se cerró: se cayó el socket o el usuario cerró la ventana. */
    void onDisconnect(ClientSession session);
}