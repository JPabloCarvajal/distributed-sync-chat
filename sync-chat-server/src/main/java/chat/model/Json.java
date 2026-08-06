package chat.model;

import com.google.gson.Gson;

/**
 * Frontera de SERIALIZACIÓN: objeto <-> texto.
 *
 * Los objetos no cruzan procesos. Lo que viaja son bytes. Aquí es donde
 * un Message se destruye en texto, y donde un texto se convierte en un
 * Message NUEVO (otro objeto, otra memoria, otro proceso).
 *
 * Esta clase existe para que sea el ÚNICO sitio del servidor que sabe
 * que el formato es JSON. Si mañana fuera otro, solo cambia aquí.
 */
public final class Json {

    private static final Gson GSON = new Gson();

    private Json() {}

    public static String toWire(Message msg) {
        return GSON.toJson(msg);
    }

    /** Devuelve null si el texto no es un Message válido: todo lo que
     *  llega por la red es NO CONFIABLE hasta que se valida. */
    public static Message fromWire(String line) {
        try {
            Message m = GSON.fromJson(line, Message.class);
            return (m != null && m.getType() != null) ? m : null;
        } catch (Exception e) {
            return null;
        }
    }
}