package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"net"

	"redis-clone/client"
)

const DEFAULT_CONFIG_ADDR = ":5001"

type Config struct {
	listenAddress string
}

type Server struct {
	Config
	peers     map[*Peer]bool
	ln        net.Listener
	addPeerCh chan *Peer
	doneCh    chan struct{}
	msgCh     chan Message

	Kv *KV
}

type Message struct {
	data []byte
	peer *Peer
}

func NewServer(cfg Config) *Server {
	if len(cfg.listenAddress) == 0 {
		cfg.listenAddress = DEFAULT_CONFIG_ADDR
	}

	return &Server{
		Config:    cfg,
		peers:     make(map[*Peer]bool),
		addPeerCh: make(chan *Peer),
		doneCh:    make(chan struct{}),
		msgCh:     make(chan Message),
		Kv:        NewKeyVal(),
	}
}

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

func (s *Server) loop() {
	for {
		select {
		case msg := <-s.msgCh:
			if err := s.handleMessage(msg); err != nil {
				slog.Error("handle raw message error", "err", err)
			}
		case peer := <-s.addPeerCh:
			s.peers[peer] = true
		case <-s.doneCh:
			return
		}
	}
}

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

func (s *Server) handleMessage(msg Message) error {
	cmd, err := parseCommand(string(msg.data))
	if err != nil {
		return err
	}

	switch v := cmd.(type) {
	case SetCommand:
		return s.Kv.Set(v.key, v.value)
	case GetCommand:
		fmt.Println(string(v.key), string(v.value))
		val, ok := s.Kv.Get(v.key)
		if !ok {
			return fmt.Errorf("key not found")
		}
		_, err := msg.peer.Send(val)
		if err != nil {
			slog.Error("peer send error", "err", err)
			return err
		}
	}

	return nil
}

func (s *Server) handleConn(conn net.Conn) {
	peer := NewPeer(conn, s.msgCh)
	s.addPeerCh <- peer
	if err := peer.readLoop(); err != nil {
		slog.Error("peer read error", "err", err, "remoteAddr", peer.conn.RemoteAddr())
	}
}

func main() {
	server := NewServer(Config{})
	go func() {
		log.Fatal(server.Start())
	}()

	client := client.New("localhost:5001")

	for i := 0; i < 20; i++ {
		if err := client.Set(context.Background(), fmt.Sprintf("Otmane_%d", i), fmt.Sprintf("Kimdil_%d", i)); err != nil {
			log.Fatal(err)
		}
	}

	value, err := client.Get(context.Background(), fmt.Sprintf("Otmane_%d", rand.Intn(20)))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(value)

	select {} // This is just blocking so our program won't exit
}
