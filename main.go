package main

import (
	"context"
	"fmt"
	"github.com/tinarao/redis/client"
	"log"
	"log/slog"
	"net"
	"time"
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
	kv             *KV
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
		kv:             NewKV(),
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

func (s *Server) handleRawMsg(rawMessage []byte) error {
	cmd, err := parseCmd(string(rawMessage))
	if err != nil {
		return err
	}
	switch v := cmd.(type) {
	case SetCommand:
		return s.kv.Set(v.key, v.val)
	}

	return nil
}

func (s *Server) loop() {
	for {
		select {
		case rawMsg := <-s.msgChannel:
			if err := s.handleRawMsg(rawMsg); err != nil {
				slog.Error("err while handling a message", "err", err.Error())
				continue
			}
		case <-s.quitChannel:
			continue
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
	go func() {
		if err := srv.start(); err != nil {
			log.Fatalf("could not start a server: %s\n", err)
		}
	}()

	time.Sleep(time.Second)

	cl := client.New("localhost:6379")

	for {
		var key, val string
		fmt.Print("Введите ключ: ")
		fmt.Scan(&key)
		fmt.Print("Введите значение: ")
		fmt.Scan(&val)

		if err := cl.Set(context.Background(), key, val); err != nil {
			log.Fatal(err)
		}
		time.Sleep(time.Second)

		fmt.Println(srv.kv.data)
	}

	select {} // block
}
