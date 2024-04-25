package server

import (
	"log"
	"net"

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
	ErrorsCh     chan peer.Errors
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
		ErrorsCh:     make(chan peer.Errors),
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
		case err := <-s.ErrorsCh:
            _ = s.handleErrors(err)
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

func (s *Server) handleErrors(err peer.Errors) error {
	return resp.NewWriter(err.Peer.Conn).WriteError(err.Err)
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
	peer := peer.NewPeer(conn, s.MsgCh, s.RemovePeerCh, s.ErrorsCh)
	s.AddPeerCh <- peer
	if err := peer.ReadLoop(); err != nil {
		log.Println("Peer read error:", err)
	}
}
