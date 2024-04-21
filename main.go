package main

import (
	"bytes"
	"fmt"
	"log"
	"log/slog"
	"net"
	"reflect"
)

const DEFAULT_CONFIG_ADDR = ":5001"

type Config struct {
	listenAddress string
}

// Server struct is the representation of our server with the necessary config.
type Server struct {
	Config

	peers map[*Peer]bool

	ln           net.Listener
	addPeerCh    chan *Peer
	removePeerCh chan *Peer
	doneCh       chan struct{}
	msgCh        chan Message

	Kv *KV
}

type Message struct {
	cmd  Command
	peer *Peer
}

// NewServer will create a new instance of our server with some basic defaults.
func NewServer(cfg Config) *Server {
	if len(cfg.listenAddress) == 0 {
		cfg.listenAddress = DEFAULT_CONFIG_ADDR
	}

	return &Server{
		Config:       cfg,
		peers:        make(map[*Peer]bool),
		addPeerCh:    make(chan *Peer),
		doneCh:       make(chan struct{}),
		msgCh:        make(chan Message),
		removePeerCh: make(chan *Peer),
		Kv:           NewKeyVal(),
	}
}

// Start will start our server with the config provided in the constructor.
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddress)
	if err != nil {
		return err
	}

	s.ln = ln

	go s.loop()

	slog.Info("server running", "listenAddr", s.listenAddress)

	return s.acceptLoop()
}

// loop with just loop forever listening on the msg channel in case we
// receive anything we send it to the handler if it's a message,
// and if it's a peer we add it to the map of peers that our server is holding
// in case we received done msg or quit we just return to break the loop
func (s *Server) loop() {
	for {
		select {
		case msg := <-s.msgCh:
			if err := s.handleMessage(msg); err != nil {
				slog.Error("handle raw message error", "err", err)
			}
		case peer := <-s.addPeerCh:
			s.peers[peer] = true
			slog.Info("new peer connected: ", "remoteAddr", peer.conn.RemoteAddr())
		case peerToRemove := <-s.removePeerCh:
			// FIXME: a possible race condition is here
			delete(s.peers, peerToRemove)
			slog.Info("peer disconnected", "remoteAdd", peerToRemove.conn.RemoteAddr())
		case <-s.doneCh:
			return
		}
	}
}

// acceptLoop is accepting tcp connections and making each one of them a peer
func (s *Server) acceptLoop() error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			slog.Error("accept error", "err", err)
			continue
		}

		go s.handleConn(conn)
	}
}

// handleMessage parses the command we receive in our connection and then executes the
// necessary function e.g: GET, SET ...
func (s *Server) handleMessage(msg Message) error {
	slog.Info("we got the command", "type", reflect.TypeOf(msg.cmd))
	switch v := msg.cmd.(type) {
	case SetCommand:
		if err := s.Kv.Set(v.key, v.value); err != nil {
			return err
		}
		_, err := msg.peer.Send([]byte("+OK\r\n"))
		if err != nil {
			return fmt.Errorf("peer send error %s", err)
		}
		slog.Info("We successfully set the key to the value", "key", string(v.key), "value", string(v.value))
	case GetCommand:
		val, ok := s.Kv.Get(v.key)
		if !ok {
			// TODO: Respond to the redis client with proper errors so we handle them on the client side
			// And we also need to clean up this code make correct writers and stuff...
			return fmt.Errorf("key not found")
		}
		buf := &bytes.Buffer{}
		buf.WriteString("+" + string(val) + "\r\n")
		_, err := msg.peer.Send(buf.Bytes())
		if err != nil {
			return fmt.Errorf("peer send error %s", err)
		}
	case HelloCommand:
		spec := map[string]string{
			"server":  "redis",
			"role":    "master",
			"version": "6.0.0",
			"mode":    "standalone",
			"proto":   "3",
		}
		resMap := writeRespMap(spec)
		_, err := msg.peer.Send(resMap)
		if err != nil {
			slog.Error("peer send error", "err", err)
			return fmt.Errorf("peer send error %s", err)
		}
		slog.Info("response sent successfully to the client", "resMap", resMap)
	}

	return nil
}

// handleConn will create a new peer for each connection that we are handling
// and send that peer over a channel to our server then triggers the read loop for ongoing peer's connection.
func (s *Server) handleConn(conn net.Conn) {
	peer := NewPeer(conn, s.msgCh, s.removePeerCh)
	s.addPeerCh <- peer
	if err := peer.readLoop(); err != nil {
		slog.Error("peer read error", "err", err, "remoteAddr", peer.conn.RemoteAddr())
	}
}

func main() {
	server := NewServer(Config{})
	log.Fatal(server.Start())
}
