public class LongRunning {
    public static void main(String[] args) throws InterruptedException {
        System.out.println("Starting application...");
        System.out.flush();
        
        while (true) {
            Thread.sleep(2000);
            System.out.println("Processing...");
            System.out.flush();
        }
    }
}