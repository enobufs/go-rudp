package rudp

import (
	//"fmt"
	"io"
	"net"
	//"strconv"
	//"time"

	"github.com/pion/logging"
	"github.com/pion/sctp"
)

// DialConfig ...
type DialConfig struct {
	Network       string
	LocalAddr     *net.UDPAddr
	RemoteAddr    *net.UDPAddr
	LoggerFactory logging.LoggerFactory
}

// Client ...
type Client struct {
	conn  *net.UDPConn
	assoc *sctp.Association
	log   logging.LeveledLogger
}

// Dial ...
func Dial(config *DialConfig) (*Client, error) {
	loggerFactory := config.LoggerFactory
	log := loggerFactory.NewLogger("rudp-c")
	conn, err := net.DialUDP(config.Network, config.LocalAddr, config.RemoteAddr)
	if err != nil {
		return nil, err
	}

	log.Debug("instantiating SCTP client")
	assoc, err := sctp.Client(sctp.Config{
		LoggerFactory: config.LoggerFactory,
		NetConn:       conn,
	})
	if err != nil {
		return nil, err
	}

	log.Debug("creating new client")
	c := &Client{
		conn:  conn,
		assoc: assoc,
		log:   log,
	}

	return c, nil
}

// OpenChannel ...
func (c *Client) OpenChannel(ch uint16) (io.ReadWriteCloser, error) {
	c.log.Debugf("opening channel %d", ch)
	stream, err := c.assoc.OpenStream(ch, sctp.PayloadTypeWebRTCBinary)
	if err != nil {
		return nil, err
	}

	return &Channel{
		stream: stream,
		log:    c.log,
	}, nil
}
