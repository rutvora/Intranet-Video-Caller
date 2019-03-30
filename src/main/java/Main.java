public class Main {
    public static void main(String[] args) {
        new StreamClient().start();
        try {
            Thread.sleep(1000);
        } catch (InterruptedException e) {
            e.printStackTrace();
        }
        new StreamServer().start();
    }
}
