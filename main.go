package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/tidwall/resp"
)

const DefaultConfigAddr = ":5001"

type Config struct {
	ListenAddress string
}

type Server struct {
	Config
	Peers        map[*Peer]bool
	Listener     net.Listener
	AddPeerCh    chan *Peer
	RemovePeerCh chan *Peer
	DoneCh       chan struct{}
	MsgCh        chan Message
	Kv           *KV
}

type Message struct {
	Cmd  Command
	Peer *Peer
}

func NewServer(cfg Config) *Server {
	if cfg.ListenAddress == "" {
		cfg.ListenAddress = DefaultConfigAddr
	}

	return &Server{
		Config:       cfg,
		Peers:        make(map[*Peer]bool),
		AddPeerCh:    make(chan *Peer),
		DoneCh:       make(chan struct{}),
		MsgCh:        make(chan Message),
		RemovePeerCh: make(chan *Peer),
		Kv:           NewKeyVal(),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.ListenAddress)
	if err != nil {
		return err
	}

	s.Listener = ln

	// go s.gracefullyShutdown()
	go s.loop()

	log.Println("Server is running on", s.ListenAddress)

	return s.acceptLoop()
}

// TODO: Atm this is not really gracefully shutting down the server we have to make it work.
func (s *Server) gracefullyShutdown() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("Received termination signal. Shutting down...")
	s.Listener.Close()
	close(s.DoneCh)
	os.Exit(0)
}

// loop continuously listens for messages, adds or removes peers, or exits.
func (s *Server) loop() {
	for {
		select {
		case msg := <-s.MsgCh:
			if err := s.handleMessage(msg); err != nil {
				log.Println("Error handling message:", err)
			}
		case peer := <-s.AddPeerCh:
			s.Peers[peer] = true
			log.Println("New peer connected:", peer.conn.RemoteAddr())
		case peerToRemove := <-s.RemovePeerCh:
			delete(s.Peers, peerToRemove)
			log.Println("Peer disconnected:", peerToRemove.conn.RemoteAddr())
		case <-s.DoneCh:
			return
		}
	}
}

// acceptLoop accepts incoming connections and handles them.
func (s *Server) acceptLoop() error {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}

		go s.handleConn(conn)
	}
}

func (s *Server) handleMessage(msg Message) error {
	switch v := msg.Cmd.(type) {
	case ClientCommand:
		return clientCommandHandler(msg)
	case SetCommand:
		return setCommandHandler(s, v, msg)
	case GetCommand:
		return getCommandHandler(s, v, msg)
	case HelloCommand:
		return helloCommandHandler(msg)
	}

	return nil
}

func clientCommandHandler(msg Message) error {
	return resp.
		NewWriter(msg.Peer.conn).
		WriteString("OK")
}

func helloCommandHandler(msg Message) error {
	spec := map[string]string{
		"server":  "redis",
		"role":    "master",
		"version": "6.0.0",
		"mode":    "standalone",
		"proto":   "3",
	}
	resMap := writeRespMap(spec)
	_, err := msg.Peer.Send(resMap)
	if err != nil {
		return fmt.Errorf("error sending response to peer: %s", err)
	}
	return nil
}

func getCommandHandler(s *Server, v GetCommand, msg Message) error {
	val, ok := s.Kv.Get(v.key)
	if !ok {
		return fmt.Errorf("key not found")
	}

	if err := resp.
		NewWriter(msg.Peer.conn).
		WriteString(string(val)); err != nil {
		return err
	}

	return nil
}

func setCommandHandler(s *Server, v SetCommand, msg Message) error {
	if err := s.Kv.Set(v.key, v.value); err != nil {
		return err
	}
	// FIXME: We have a bug with our OWN WRITTEN CLIENT here
	// When we send get request to get the value associated with the key
	// we get the OK message which is not fine we have to send the value
	// but with the official redis client this is working fine
	if err := resp.
		NewWriter(msg.Peer.conn).
		WriteString("OK"); err != nil {
		return err
	}
	return nil
}

// handleConn handles incoming connections.
func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	peer := NewPeer(conn, s.MsgCh, s.RemovePeerCh)
	s.AddPeerCh <- peer
	if err := peer.readLoop(); err != nil {
		log.Println("Peer read error:", err)
	}
}

func main() {
	server := NewServer(Config{})
	log.Fatal(server.Start())
}
