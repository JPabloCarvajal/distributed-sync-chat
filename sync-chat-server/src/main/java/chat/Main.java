package chat;

import com.google.gson.Gson;
import java.util.concurrent.Semaphore;

public class Main {
    public static void main(String[] args) {
        Gson gson = new Gson();
        Semaphore turno = new Semaphore(1, true);

        System.out.println("Java " + System.getProperty("java.version"));
        System.out.println("Gson: " + gson.toJson(new int[]{1, 2, 3}));
        System.out.println("Semaforo justo: " + turno.isFair());
    }
}
