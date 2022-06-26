package rudp

import (
	"net"

	dcep "github.com/pion/datachannel"
	"github.com/pion/logging"
	"github.com/pion/sctp"
)

// DialConfig ...
type DialConfig struct {
	Network       string
	LocalAddr     *net.UDPAddr
	RemoteAddr    *net.UDPAddr
	BufferSize    int
	LoggerFactory logging.LoggerFactory
}

// Client ...
type Client struct {
	conn          *net.UDPConn
	assoc         *sctp.Association
	loggerFactory logging.LoggerFactory
	log           logging.LeveledLogger
}

// Dial ...
func Dial(config *DialConfig) (*Client, error) {
	log := config.LoggerFactory.NewLogger("rudp-c")
	conn, err := net.DialUDP(config.Network, config.LocalAddr, config.RemoteAddr)
	if err != nil {
		return nil, err
	}

	if config.BufferSize > 0 {
		log.Debugf("setting buffer size to %d\n", config.BufferSize)
		err = conn.SetReadBuffer(config.BufferSize)
		if err != nil {
			return nil, err
		}
		err = conn.SetWriteBuffer(config.BufferSize)
		if err != nil {
			return nil, err
		}
	}

	log.Debug("instantiating SCTP client")
	assoc, err := sctp.Client(sctp.Config{
		LoggerFactory:        config.LoggerFactory,
		MaxReceiveBufferSize: uint32(config.BufferSize), // 0: defaults to 1MB
		NetConn:              conn,
	})
	if err != nil {
		return nil, err
	}

	log.Debug("creating new client")
	c := &Client{
		conn:          conn,
		assoc:         assoc,
		loggerFactory: config.LoggerFactory,
		log:           log,
	}

	return c, nil
}

// OpenChannel ...
func (c *Client) OpenChannel(chID uint16, cfg Config) (Channel, error) {
	c.log.Debugf("opening channel %d", chID)

	dc, err := dcep.Dial(c.assoc, chID, &dcep.Config{
		ChannelType:          cfg.ChannelType,
		Negotiated:           cfg.Negotiated,
		Priority:             cfg.Priority,
		ReliabilityParameter: cfg.ReliabilityParameter,
		Label:                cfg.Label,
		Protocol:             cfg.Protocol,
		LoggerFactory:        c.loggerFactory,
	})
	if err != nil {
		return nil, err
	}

	return dc, nil
}

// Close ...
func (c *Client) Close() error {
	return c.assoc.Close()
}
