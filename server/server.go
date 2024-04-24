package server

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"redis-clone/keyval"
	"redis-clone/peer"
	"redis-clone/proto"

	"github.com/tidwall/resp"
)

const DefaultConfigAddr = ":5001"

type Config struct {
	ListenAddress string
}

type Server struct {
	Config
	Peers        map[*peer.Peer]bool
	Listener     net.Listener
	AddPeerCh    chan *peer.Peer
	RemovePeerCh chan *peer.Peer
	DoneCh       chan struct{}
	MsgCh        chan peer.Message
	Kv           *keyval.KV
}


func NewServer(cfg Config) *Server {
	if cfg.ListenAddress == "" {
		cfg.ListenAddress = DefaultConfigAddr
	}

	return &Server{
		Config:       cfg,
		Peers:        make(map[*peer.Peer]bool),
		AddPeerCh:    make(chan *peer.Peer),
		RemovePeerCh: make(chan *peer.Peer),
		MsgCh:        make(chan peer.Message),
		DoneCh:       make(chan struct{}),
		Kv:           keyval.NewKeyVal(),
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
			log.Println("New peer connected:", peer.Conn.RemoteAddr())
		case peerToRemove := <-s.RemovePeerCh:
			delete(s.Peers, peerToRemove)
			log.Println("Peer disconnected:", peerToRemove.Conn.RemoteAddr())
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

func (s *Server) handleMessage(msg peer.Message) error {
	switch v := msg.Cmd.(type) {
	case proto.ClientCommand:
		return clientCommandHandler(msg)
	case proto.SetCommand:
		return setCommandHandler(s, v, msg)
	case proto.GetCommand:
		return getCommandHandler(s, v, msg)
	case proto.HelloCommand:
		return helloCommandHandler(msg)
    case proto.CommandCommand:
        return commandCommandHandler(msg)
	}

	return nil
}

func clientCommandHandler(msg peer.Message) error {
	return resp.
		NewWriter(msg.Peer.Conn).
		WriteString("OK")
}

func commandCommandHandler(msg peer.Message) error {
	spec := map[string]string{
		"server":  "redis",
		"role":    "master",
		"version": "6.0.0",
		"mode":    "standalone",
		"proto":   "3",
		"Author":  "Otmane",
	}
	resMap := proto.WriteRespMap(spec)
	_, err := msg.Peer.Send(resMap)
	if err != nil {
		return fmt.Errorf("error sending response to peer: %s", err)
	}
	return nil
}

func helloCommandHandler(msg peer.Message) error {
	spec := map[string]string{
		"server":  "redis",
		"role":    "master",
		"version": "6.0.0",
		"mode":    "standalone",
		"proto":   "3",
		"Author":  "Otmane",
	}
	resMap := proto.WriteRespMap(spec)
	_, err := msg.Peer.Send(resMap)
	if err != nil {
		return fmt.Errorf("error sending response to peer: %s", err)
	}
	return nil
}

func getCommandHandler(s *Server, v proto.GetCommand, msg peer.Message) error {
	val, ok := s.Kv.Get(v.Key)
	if !ok {
		return fmt.Errorf("key not found")
	}

	return resp.
		NewWriter(msg.Peer.Conn).
		WriteString(string(val))
}

func setCommandHandler(s *Server, v proto.SetCommand, msg peer.Message) error {
	if err := s.Kv.Set(v.Key, v.Value); err != nil {
		return err
	}
	// FIXME: We have a bug with our OWN WRITTEN CLIENT here
	// When we send get request to get the value associated with the key
	// we get the OK message which is not fine we have to send the value
	// but with the official redis client this is working fine
	return resp.
		NewWriter(msg.Peer.Conn).
		WriteString("OK")
}

// handleConn handles incoming connections.
func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	peer := peer.NewPeer(conn, s.MsgCh, s.RemovePeerCh)
	s.AddPeerCh <- peer
	if err := peer.ReadLoop(); err != nil {
		log.Println("Peer read error:", err)
	}
}
