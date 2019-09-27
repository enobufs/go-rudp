package rudp

import (
	"net"
	"testing"

	"github.com/pion/logging"
	"github.com/stretchr/testify/assert"
)

func TestConnectionSimple(t *testing.T) {
	loggerFactory := logging.NewDefaultLoggerFactory()
	log := loggerFactory.NewLogger("test")
	l, err := Listen(&ListenConfig{
		Network:       "udp4",
		LoggerFactory: loggerFactory,
	})
	assert.NoError(t, err, "should succeed")
	defer func() {
		err = l.Close()
		assert.NoError(t, err, "should succeed")
	}()

	channelID := uint16(123)

	serverDoneCh := make(chan struct{})
	clientDoneCh := make(chan struct{})
	msgs := []string{
		"Hello",
		"Good to see you",
		"Bye",
	}

	go func() {
		s, err2 := l.Accept()
		log.Debugf("l.Accept() returned: %v", err2)
		if !assert.NoError(t, err2, "should succeed") {
			close(serverDoneCh)
			return
		}
		defer s.Close()

		serverCh, err2 := s.AcceptChannel()
		if !assert.NoError(t, err2, "should succeed") {
			return
		}
		defer serverCh.Close()

		assert.Equal(t, channelID, serverCh.ID(), "should match")

		buf := make([]byte, 64*1024)
		var nReceived int
		for {
			n, err2 := serverCh.Read(buf)
			if err2 != nil {
				break
			}

			msg := string(buf[:n])
			log.Debugf("server received %s", msg)
			assert.Equal(t, msgs[nReceived], msg, "should match")
			nReceived++

			n, err2 = serverCh.Write([]byte(msg))
			assert.NoError(t, err2, "should succeed")
			assert.Equal(t, len(msg), n, "should match")
		}

		log.Debug("server done")
		close(serverDoneCh)
	}()

	c, err := Dial(&DialConfig{
		Network: "udp4",
		RemoteAddr: &net.UDPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: defaultServerPort,
		},
		LoggerFactory: loggerFactory,
	})
	assert.NoError(t, err, "should succeed")

	go func() {
		clientCh, err2 := c.OpenChannel(channelID)
		if !assert.NoError(t, err2, "should succeed") {
			return
		}

		buf := make([]byte, 64*1024)
		for i := 0; i < len(msgs); i++ {
			msg := msgs[i]
			log.Debugf("client sending: %s", msg)
			clientCh.Write([]byte(msg))
			n, err2 := clientCh.Read(buf)
			assert.NoError(t, err2, "should succeed")
			assert.Equal(t, msg, string(buf[:n]), "should match")
			log.Debugf("client received: %s", string(buf[:n]))
		}

		log.Debug("closing client channel")
		err2 = clientCh.Close()
		assert.NoError(t, err2, "should succeed")

		log.Debug("client done")
		close(clientDoneCh)
	}()

	<-clientDoneCh
	<-serverDoneCh
}
