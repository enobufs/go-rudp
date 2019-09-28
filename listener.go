package rudp

import (
	"fmt"
	"net"
	"sync"

	"github.com/pion/logging"
)

const (
	defaultServerPort = 40916
)

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
		svr, err = newServer(&serverConfig{
			conn:          l.conn,
			remAddr:       from,
			loggerFactory: l.loggerFactory,
			onHandshakeComplete: func() {
				select {
				case l.acceptCh <- svr:
				default:
				}
			},
			onClosed: func() {
				l.mutex.Lock()
				defer l.mutex.Unlock()
				delete(l.serverMap, from.String())
			},
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

// LocalAddr ...
func (l *Listener) LocalAddr() net.Addr {
	return l.conn.LocalAddr()
}
