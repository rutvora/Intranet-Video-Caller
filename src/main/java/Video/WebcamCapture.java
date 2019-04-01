package Video;

import com.github.sarxos.webcam.Webcam;

import javax.imageio.ImageIO;
import java.io.File;
import java.io.IOException;

public class WebcamCapture {

    private static WebcamCaptureHelper webcamCaptureHelper = null;

    public void startCapture() {
        if (webcamCaptureHelper != null) return;
        webcamCaptureHelper = new WebcamCaptureHelper();
        Thread captureThread = new Thread(webcamCaptureHelper);
        captureThread.start();
    }

    public void exit() throws WebcamExitException {
        if (webcamCaptureHelper == null) throw new WebcamExitException("Webcam Not Capturing");
        webcamCaptureHelper.exit = true;
    }

    public class WebcamExitException extends Exception {
        WebcamExitException(String s) {
            super(s);
        }
    }

    private class WebcamCaptureHelper implements Runnable {

        volatile boolean exit = false;
        volatile int limit;
        private int count = 0;

        @Override
        public void run() {
            Webcam webcam = Webcam.getDefault();
            webcam.open();
            while (!exit) {
                try {
                    System.out.println("Capturing.....");
                    ImageIO.write(webcam.getImage(), "JPG", new File(count + ".jpg"));
                    count++;
                    if (count == limit) {
                        String commands = null; //TODO: Replace with actual command string
                        Process ffmpeg = Runtime.getRuntime().exec(commands);

                    }
                    //TODO: Call JNI (after keyframe distance number of files)
                } catch (IOException e) {
                    e.printStackTrace();
                }
            }

            //TODO: Cleanup temp files

        }
    }

}
