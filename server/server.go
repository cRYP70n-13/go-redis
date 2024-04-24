package server

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"redis-clone/keyval"
	"redis-clone/peer"
	"redis-clone/proto"
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

	go s.loop()

	log.Println("Server is running on", s.ListenAddress)

	return s.acceptLoop()
}

// TODO: Atm this is not really gracefully shutting down the server we have to make it work.
func (s *Server) gracefullyShutdown() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	s.DoneCh <- struct{}{}
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
	case proto.PingCommand:
		return pingCommandHandler(msg)
	case proto.ConfigGetCommand:
		return configCommandGetHandler(msg)
	case proto.ExistCommand:
		return existCommandHandler(s, v, msg)
	case proto.DelCommand:
		return delCommandHandler(s, v, msg)
	case proto.IncrCommand:
		return incrCommandHandler(s, v, msg)
	case proto.DecrCommand:
		return decrCommandHandler(s, v, msg)
	default:
		return unhandledCommand(msg)
	}
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
