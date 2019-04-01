package Network;

import java.io.IOException;
import java.net.*;

public class StreamSender {
    public void start() {
        String IPAddr = null; //TODO
        StreamSenderHelper helper = new StreamSenderHelper(IPAddr);
        Thread streamServer = new Thread(helper);
        streamServer.start();
    }

    private class StreamSenderHelper implements Runnable {
        DatagramSocket UDPSocket;
        InetAddress ip;

        StreamSenderHelper(String IPAddr) {
            try {
                ip = InetAddress.getByName(IPAddr);
            } catch (UnknownHostException e) {
                e.printStackTrace();
            }
        }

        @Override
        public void run() {
            try {
                UDPSocket = new DatagramSocket();
            } catch (SocketException e) {
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
