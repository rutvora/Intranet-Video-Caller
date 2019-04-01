import Video.WebcamCapture;

public class Main {
    public static void main(String[] args) {
//        new StreamReceiver().start();
        WebcamCapture webcamCapture = new WebcamCapture();
        webcamCapture.startCapture();
        try {
            Thread.sleep(1000);
        } catch (InterruptedException e) {
            e.printStackTrace();
        }
        try {
            webcamCapture.exit();
        } catch (WebcamCapture.WebcamExitException e) {
            e.printStackTrace();
        }
//        new StreamSender().start();
    }
}
