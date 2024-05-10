// https://youtu.be/LMrxfWB6sbQ?si=NEJTXXNYtDbt52DN&t=715

package main

import (
	"fmt"
	"log"
	"log/slog"
	"net"
)

// DefListenAddr represents default TCP port redis client is listening on
const DefListenAddr = ":6379"

type Config struct {
	ListenAddress string
}

type Server struct {
	Config
	ln             net.Listener
	peers          map[*Peer]bool
	addPeerChannel chan *Peer
	quitChannel    chan struct{}
	msgChannel     chan []byte
}

func newServer(cfg Config) *Server {
	if len(cfg.ListenAddress) == 0 {
		cfg.ListenAddress = DefListenAddr
	}

	return &Server{
		Config:         cfg,
		peers:          make(map[*Peer]bool),
		addPeerChannel: make(chan *Peer),
		quitChannel:    make(chan struct{}),
		msgChannel:     make(chan []byte),
	}
}

func (s *Server) start() error {
	ln, err := net.Listen("tcp", s.ListenAddress)
	if err != nil {
		return err
	}
	s.ln = ln
	go s.loop()

	slog.Info("serving!", "listenAddress", s.ListenAddress)

	return s.startAcceptingLoop()
}

func (s *Server) handleRawMsg(raw []byte) error {
	fmt.Printf("raw: %s\n", string(raw))
	return nil
}

func (s *Server) loop() {
	for {
		select {
		case rawMsg := <-s.msgChannel:
			if err := s.handleRawMsg(rawMsg); err != nil {
				slog.Error("raw msg err", "err", err.Error())
				return
			}
		case <-s.quitChannel:
			return
		case peer := <-s.addPeerChannel:
			s.peers[peer] = true
		}
	}
}

func (s *Server) startAcceptingLoop() error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			slog.Error("accept conn err", "err", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	peer := NewPeer(conn, s.msgChannel)
	s.addPeerChannel <- peer

	slog.Info("new peer connected", "remoteAddr", conn.RemoteAddr())
	err := peer.readLoop()
	if err != nil {
		slog.Error("peer reading err", "err", err.Error(), "remoteAddr", conn.RemoteAddr())
	}
}

func main() {
	srv := newServer(Config{})
	if err := srv.start(); err != nil {
		log.Fatalf("could not start a server: %s\n", err)
	}
}
