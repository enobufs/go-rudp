package rudp

import (
	"io"
	"net"
	"time"

	"github.com/pion/logging"
)

type serverConn struct {
	conn    net.PacketConn
	remAddr net.Addr
	readCh  chan []byte
	log     logging.LeveledLogger
}

func newServerConn(conn net.PacketConn, remAddr net.Addr, log logging.LeveledLogger) *serverConn {
	return &serverConn{
		conn:    conn,
		remAddr: remAddr,
		readCh:  make(chan []byte, 1024),
		log:     log,
	}
}

func (c *serverConn) Read(b []byte) (n int, err error) {
	data, ok := <-c.readCh
	if !ok {
		return 0, io.EOF
	}

	c.log.Debugf("serverConn: read %d bytes", len(data))
	if len(b) < len(data) {
		return 0, io.ErrShortBuffer
	}
	bytesCopied := copy(b, data)
	c.log.Debugf("serverConn: bytesCopied=%d", bytesCopied)
	return bytesCopied, nil
}

func (c *serverConn) Write(b []byte) (n int, err error) {
	return c.conn.WriteTo(b, c.remAddr)
}

func (c *serverConn) Close() error {
	close(c.readCh)
	return nil
}

func (c *serverConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *serverConn) RemoteAddr() net.Addr {
	return c.remAddr
}

func (c *serverConn) SetDeadline(t time.Time) error {
	return nil // unused
}

func (c *serverConn) SetReadDeadline(t time.Time) error {
	return nil // unused
}

func (c *serverConn) SetWriteDeadline(t time.Time) error {
	return nil // unused
}

func (c *serverConn) handleInbound(data []byte) {
	c.log.Debugf("serverConn: handleInboud: %d bytes", len(data))
	buf := make([]byte, len(data))
	copy(buf, data)
	c.readCh <- buf // possible race with Close()
}
