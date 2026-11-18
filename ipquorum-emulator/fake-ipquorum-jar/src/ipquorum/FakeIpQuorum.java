package ipquorum;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.io.OutputStream;
import java.net.ConnectException;
import java.net.InetSocketAddress;
import java.net.Socket;
import java.net.SocketTimeoutException;
import java.nio.charset.StandardCharsets;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.List;
import java.util.Properties;
import java.util.concurrent.CountDownLatch;

/**
 * FAKE ip_quorum.jar for testing. NOT the IBM IP quorum application.
 *
 * It prints log output in the format of the real ip_quorum.jar and keeps TCP
 * connections to the nodes listed in the embedded quorum.properties. The wire
 * protocol is a simplified line-based protocol shared with the emulator's mock
 * nodes (see ipquorum-emulator/README.md). The real protocol is not public.
 *
 * Tuning (system properties, e.g. via JAVA_OPTS):
 *   -Dfake.retry.ms=2000      delay between reconnect attempts
 *   -Dfake.connect.ms=3000    TCP connect timeout
 *   -Dfake.handshake.ms=5000  handshake reply timeout
 *   -Dfake.keepalive.ms=2000  PING interval
 */
public final class FakeIpQuorum {
    static final int PROTOCOL = 3;
    static final String CONFIG_RESOURCE = "/quorum.properties";
    private static final DateTimeFormatter TS = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss:SSS");

    private static final long RETRY_MS = Long.getLong("fake.retry.ms", 2000L);
    private static final int CONNECT_MS = Integer.getInteger("fake.connect.ms", 3000);
    private static final int HANDSHAKE_MS = Integer.getInteger("fake.handshake.ms", 5000);
    private static final long KEEPALIVE_MS = Long.getLong("fake.keepalive.ms", 2000L);

    private static volatile boolean running = true;
    private static final Object OUT = new Object();

    private FakeIpQuorum() {}

    public static void main(String[] args) throws Exception {
        console("=== IP quorum ===");
        fine("Quorum", "CONFIG", "=== IP quorum ===");
        console("Name set to null.");
        fine("Quorum", "CONFIG", "Name set to null.");

        List<String> nodes;
        try {
            nodes = loadNodes();
        } catch (IOException | RuntimeException e) {
            console("Unable to parse the configuration: " + e.getMessage());
            fine("Quorum", "SEVERE", "Unable to parse the configuration: " + e.getMessage());
            System.exit(1);
            return;
        }

        for (String n : nodes) {
            fine("Quorum", "FINE", "Node " + n + " (0, 0)");
        }
        console("Successfully parsed the configuration, found " + nodes.size() + " nodes.");
        fine("Quorum", "CONFIG", "Successfully parsed the configuration, found " + nodes.size() + " nodes.");

        Runtime.getRuntime().addShutdownHook(new Thread(() -> {
            running = false;
            console("Shutting down.");
        }));

        CountDownLatch done = new CountDownLatch(nodes.size());
        int id = 13;
        for (String n : nodes) {
            final int tid = ++id;
            Thread t = new Thread(() -> {
                try {
                    new NodeClient(n, tid).run();
                } finally {
                    done.countDown();
                }
            }, "node-" + n);
            t.setDaemon(true);
            t.start();
        }
        done.await();
    }

    private static List<String> loadNodes() throws IOException {
        Properties props = new Properties();
        try (InputStream in = FakeIpQuorum.class.getResourceAsStream(CONFIG_RESOURCE)) {
            if (in == null) {
                throw new IOException("quorum.properties not found in the jar");
            }
            props.load(in);
        }
        String raw = props.getProperty("nodes", "").trim();
        if (raw.isEmpty()) {
            throw new IOException("no nodes configured");
        }
        List<String> nodes = new ArrayList<>();
        for (String n : raw.split(",")) {
            n = n.trim();
            if (!n.isEmpty()) {
                nodes.add(n);
            }
        }
        return nodes;
    }

    static void console(String msg) {
        synchronized (OUT) {
            System.out.println(msg);
        }
    }

    static void fine(String source, String level, String msg) {
        synchronized (OUT) {
            System.out.println(LocalDateTime.now().format(TS) + " " + source + " " + level + ": " + msg);
        }
    }

    /** Maintains the connection to one node and reconnects when it is lost. */
    private static final class NodeClient {
        private final String address;
        private final String host;
        private final int port;
        private final int tid;

        NodeClient(String address, int tid) {
            this.address = address;
            this.tid = tid;
            int i = address.lastIndexOf(':');
            this.host = address.substring(0, i);
            this.port = Integer.parseInt(address.substring(i + 1));
        }

        private void info(String msg) {
            fine(host, "[" + tid + "] INFO", msg);
        }

        private void fineMsg(String msg) {
            fine(host, "[" + tid + "] FINE", msg);
        }

        private void warn(String msg) {
            fine(host, "[" + tid + "] WARNING", msg);
        }

        void run() {
            while (running) {
                try {
                    session();
                } catch (SocketTimeoutException e) {
                    console("Handshake timed out for " + address);
                    warn("Timed out waiting for the handshake reply: " + e.getMessage());
                } catch (ConnectException e) {
                    console("Failed to open socket to " + address);
                    warn("Failed to open socket: " + e.getMessage());
                } catch (IOException e) {
                    console("Lost connection to " + address);
                    warn("Connection lost: " + e.getMessage());
                }
                sleep(RETRY_MS);
            }
        }

        private void session() throws IOException {
            console("Trying to open socket");
            info("Trying to open socket");
            try (Socket s = new Socket()) {
                s.connect(new InetSocketAddress(host, port), CONNECT_MS);
                s.setSoTimeout(HANDSHAKE_MS);
                OutputStream out = s.getOutputStream();
                BufferedReader in = new BufferedReader(new InputStreamReader(s.getInputStream(), StandardCharsets.UTF_8));

                console("Handshaking");
                info("Handshaking");
                String hello = "HELLO protocol=" + PROTOCOL + " features_bitmap=0";
                fineMsg("<Msg [protocol=" + PROTOCOL + ", sequence=0, command=HANDSHAKE_REQUEST, length=1] Data [features_bitmap=0]");
                send(out, hello);
                String reply = in.readLine();
                if (reply == null) {
                    throw new IOException("connection closed during the handshake");
                }
                if (!reply.startsWith("HELLO-ACK")) {
                    throw new IOException("unexpected handshake reply: " + reply);
                }
                fineMsg(">Msg [protocol=" + PROTOCOL + ", sequence=0, command=HANDSHAKE_RESPONSE, length=1] Data [" + reply + "]");

                console("Creating UID");
                console("Waiting for UID");
                console("*Connecting");
                console("Connected to " + host);
                info("Connected to " + host);

                s.setSoTimeout((int) Math.max(KEEPALIVE_MS * 3, HANDSHAKE_MS));
                while (running) {
                    sleep(KEEPALIVE_MS);
                    if (!running) {
                        break;
                    }
                    send(out, "PING");
                    String pong = in.readLine();
                    if (pong == null) {
                        throw new IOException("connection closed by " + address);
                    }
                    if (!pong.equals("PONG")) {
                        throw new IOException("unexpected keepalive reply: " + pong);
                    }
                }
            }
        }

        private static void send(OutputStream out, String line) throws IOException {
            out.write((line + "\n").getBytes(StandardCharsets.UTF_8));
            out.flush();
        }
    }

    private static void sleep(long ms) {
        try {
            Thread.sleep(ms);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            running = false;
        }
    }
}
