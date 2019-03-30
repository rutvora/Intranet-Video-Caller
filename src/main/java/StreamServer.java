import java.io.IOException;
import java.net.*;

class StreamServer {
    void start() {
        StreamServerHelper helper = new StreamServerHelper();
        Thread streamServer = new Thread(helper);
        streamServer.start();
    }

    private class StreamServerHelper implements Runnable {
        DatagramSocket UDPSocket;
        InetAddress ip;

        @Override
        public void run() {
            try {
                UDPSocket = new DatagramSocket();
                ip = InetAddress.getLocalHost();
            } catch (SocketException | UnknownHostException e) {
                e.printStackTrace();
            }
            byte[] buff = "12345678".getBytes();
            DatagramPacket datagramPacket = new DatagramPacket(buff, buff.length, ip, 8000);
            try {
                UDPSocket.send(datagramPacket);
            } catch (IOException e) {
                e.printStackTrace();
            }

        }
    }
}
