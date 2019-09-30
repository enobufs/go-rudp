package rudp

import (
	"net"

	"github.com/pion/logging"
	"github.com/pion/sctp"
)

// Server ...
type Server struct {
	conn          *serverConn
	assoc         *sctp.Association
	onClosed      func()
	loggerFactory logging.LoggerFactory
	log           logging.LeveledLogger
}

type serverConfig struct {
	conn                net.PacketConn
	remAddr             net.Addr
	onHandshakeComplete func()
	onClosed            func()
	loggerFactory       logging.LoggerFactory
}

func newServer(config *serverConfig) (*Server, error) {
	log := config.loggerFactory.NewLogger("rudp-s")
	svrConn := newServerConn(
		config.conn,
		config.remAddr,
		log)

	s := &Server{
		conn:          svrConn,
		onClosed:      config.onClosed,
		loggerFactory: config.loggerFactory,
		log:           log,
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
		config.onHandshakeComplete()
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
	return &Channel{
		stream: stream,
		log:    s.log,
	}, nil
}

// Close ...
func (s *Server) Close() error {
	return s.conn.Close()
}

// LocalAddr ...
func (s *Server) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

// RemoteAddr ...
func (s *Server) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}
