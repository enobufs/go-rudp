package rudp

import (
	"fmt"
	"net"
	"sync"

	"github.com/pion/logging"
	"github.com/pion/sctp"
)

const (
	defaultServerPort = 40916
)

// Server ...
type Server struct {
	conn          *serverConn
	assoc         *sctp.Association
	loggerFactory logging.LoggerFactory
	log           logging.LeveledLogger
}

func newServer(conn net.PacketConn, remAddr net.Addr, loggerFactory logging.LoggerFactory, onHandshakeComplete func()) (*Server, error) {
	svrConn := newServerConn(conn, remAddr, loggerFactory.NewLogger("rudp-s"))
	s := &Server{
		conn:          svrConn,
		loggerFactory: loggerFactory,
		log:           loggerFactory.NewLogger("Server"),
	}

	go func() {
		s.log.Debug("handlshake started")
		var err error
		s.assoc, err = sctp.Server(sctp.Config{
			LoggerFactory: s.loggerFactory,
			NetConn:       s.conn,
		})
		if err != nil {
			s.log.Error(err.Error())
			return
		}
		onHandshakeComplete()
	}()

	return s, nil
}

func (s *Server) handleInbound(data []byte) {
	s.log.Debugf("Server: handleInboud: %d bytes", len(data))
	s.conn.handleInbound(data)
}

// AcceptChannel ...
func (s *Server) AcceptChannel() (*Channel, error) {
	s.log.Debug("accept stream")
	stream, err := s.assoc.AcceptStream()
	if err != nil {
		return nil, err
	}
	/*
		buf := make([]byte, 64*1024)
		for {
			n, err := stream.Read(buf)
			s.log.Debugf("Received: %s", string(buf[:n]))
			if err != nil {
				fmt.Println("Error: ", err)
			}
		}
	*/
	return &Channel{
		stream: stream,
		log:    s.log,
	}, nil
}

// Close ...
func (s *Server) Close() error {
	return s.conn.Close()
}

// ListenConfig ...
type ListenConfig struct {
	Network       string
	LocalAddr     net.Addr
	Backlog       int
	LoggerFactory logging.LoggerFactory
}

// Listener ...
type Listener struct {
	conn          *net.UDPConn
	serverMap     map[string]*Server
	acceptCh      chan *Server
	closeCh       chan struct{}
	mutex         sync.RWMutex
	log           logging.LeveledLogger
	loggerFactory logging.LoggerFactory
}

// Listen ...
func Listen(config *ListenConfig) (*Listener, error) {
	loggerFactory := config.LoggerFactory
	log := loggerFactory.NewLogger("rudp-l")

	var locAddr *net.UDPAddr
	if config.LocalAddr == nil {
		locAddr = &net.UDPAddr{
			IP:   nil,
			Port: defaultServerPort,
		}
	} else {
		locAddr = config.LocalAddr.(*net.UDPAddr)
		if locAddr.Port == 0 {
			locAddr.Port = defaultServerPort
		}
	}

	conn, err := net.ListenUDP(config.Network, locAddr)
	if err != nil {
		return nil, err
	}

	log.Infof("listening on %s", conn.LocalAddr().String())

	backlog := config.Backlog
	if backlog == 0 {
		backlog = 8
	}

	l := &Listener{
		conn:          conn,
		serverMap:     map[string]*Server{},
		acceptCh:      make(chan *Server, backlog),
		closeCh:       make(chan struct{}),
		log:           log,
		loggerFactory: loggerFactory,
	}

	go l.run()

	return l, nil
}

func (l *Listener) run() {
	defer close(l.closeCh)
	defer close(l.acceptCh)

	buf := make([]byte, 64*1024)
	for {
		n, from, err := l.conn.ReadFrom(buf)
		if err != nil {
			break
		}
		l.log.Debugf("l.conn received %d bytes from %s", n, from.String())
		l.handleInbound(buf[:n], from)
	}
}

func (l *Listener) handleInbound(data []byte, from net.Addr) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	svr, ok := l.serverMap[from.String()]
	if !ok {
		l.log.Debugf("creating a new server for %s", from.String())
		var err error
		svr, err = newServer(l.conn, from, l.loggerFactory, func() {
			select {
			case l.acceptCh <- svr:
			default:
			}
		})
		if err != nil {
			l.log.Warn(err.Error())
			return
		}
		l.serverMap[from.String()] = svr
	}

	svr.handleInbound(data)
}

// Accept ...
func (l *Listener) Accept() (*Server, error) {
	svr, ok := <-l.acceptCh
	if !ok {
		return nil, fmt.Errorf("listener closed")
	}
	return svr, nil
}

// Close ...
func (l *Listener) Close() error {
	err := l.conn.Close()
	<-l.closeCh
	return err
}
