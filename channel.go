package rudp

import (
	"github.com/pion/logging"
	"github.com/pion/sctp"
)

// Channel ...
type Channel struct {
	stream *sctp.Stream
	log    logging.LeveledLogger
}

// ID returns the channel ID
func (c *Channel) ID() uint16 {
	return c.stream.StreamIdentifier()
}

// Read ...
func (c *Channel) Read(data []byte) (int, error) {
	return c.stream.Read(data)
}

// Write ...
func (c *Channel) Write(data []byte) (int, error) {
	return c.stream.Write(data)
}

// Close ...
func (c *Channel) Close() error {
	c.log.Debugf("closing channel(%d)", c.ID())
	return c.stream.Close()
}
