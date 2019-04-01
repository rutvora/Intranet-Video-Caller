package Network;

import java.io.IOException;
import java.net.DatagramPacket;
import java.net.DatagramSocket;
import java.net.SocketException;

public class StreamReceiver {
    public void start() {
        StreamReceiverHelper helper = new StreamReceiverHelper();
        Thread streamClient = new Thread(helper);
        streamClient.start();
    }

    private class StreamReceiverHelper implements Runnable {

        DatagramPacket datagramPacket = null;
        private DatagramSocket UDPSocket;
        private byte[] receive;

        @Override
        public void run() {
            try {
                UDPSocket = new DatagramSocket(8000);
                receive = new byte[UDPSocket.getReceiveBufferSize()];
            } catch (SocketException e) {
                e.printStackTrace();
            }

            while (true) {
                datagramPacket = new DatagramPacket(receive, receive.length);
                try {
                    UDPSocket.receive(datagramPacket);
                    System.out.println(datagramPacket.getLength());
                } catch (IOException e) {
                    e.printStackTrace();
                }
                break;
            }

        }
    }
}
