package rudp

import (
	"net"
	"sync/atomic"
	"time"

	dcep "github.com/pion/datachannel"
	"github.com/pion/logging"
	"github.com/pion/sctp"
)

// Server ...
type Server struct {
	conn          *serverConn
	bufferSize    int
	assoc         *sctp.Association
	closed        atomic.Value // bool
	onClosed      func()
	loggerFactory logging.LoggerFactory
	log           logging.LeveledLogger
}

type serverConfig struct {
	conn                net.PacketConn
	remAddr             net.Addr
	bufferSize          int
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
		bufferSize:    config.bufferSize,
		onClosed:      config.onClosed,
		loggerFactory: config.loggerFactory,
		log:           log,
	}

	s.closed.Store(false)

	go func() {
		s.log.Debug("handlshake started")
		var err error
		s.assoc, err = sctp.Server(sctp.Config{
			LoggerFactory:        s.loggerFactory,
			MaxReceiveBufferSize: uint32(s.bufferSize),
			NetConn:              s.conn,
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
	if s.closed.Load().(bool) {
		return
	}
	s.log.Debugf("Server: handleInboud: %d bytes", len(data))
	s.conn.handleInbound(data)
}

// AcceptChannel ...
func (s *Server) AcceptChannel() (Channel, error) {
	s.log.Debug("accept stream")
	cfg := dcep.Config{LoggerFactory: s.loggerFactory}
	dcepCh, err := dcep.Accept(s.assoc, &cfg)
	if err != nil {
		return nil, err
	}

	dc := &dataChannel{
		dc: dcepCh,
		config: Config{
			ChannelType:          cfg.ChannelType,
			Negotiated:           cfg.Negotiated,
			Priority:             cfg.Priority,
			ReliabilityParameter: cfg.ReliabilityParameter,
			Label:                cfg.Label,
			Protocol:             cfg.Protocol,
		},
	}

	return dc, nil
}

// Close ...
func (s *Server) Close() error {
	var err error
	if !s.closed.Load().(bool) {
		err = s.conn.Close()
		s.closed.Store(true)
		time.AfterFunc(8*time.Second, func() {
			s.onClosed()
		})
	}
	return err
}

// LocalAddr ...
func (s *Server) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

// RemoteAddr ...
func (s *Server) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}
