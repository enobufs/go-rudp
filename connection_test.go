package rudp

import (
	"net"
	"testing"
	"time"

	"github.com/pion/logging"
	"github.com/stretchr/testify/assert"
)

type echoServerConfig struct {
	t             *testing.T
	channelID     uint16
	loggerFactory logging.LoggerFactory
	log           logging.LeveledLogger
}

type echoClientConfig struct {
	t             *testing.T
	id            int
	channelID     uint16
	loggerFactory logging.LoggerFactory
	log           logging.LeveledLogger
}

func echoServer(cfg *echoServerConfig) (chan<- struct{}, <-chan struct{}) {
	shutDownCh := make(chan struct{})
	doneCh := make(chan struct{})

	// make sure the server is already listening when this function returns
	l, err := Listen(&ListenConfig{
		Network:       "udp4",
		LoggerFactory: cfg.loggerFactory,
	})
	assert.NoError(cfg.t, err, "should succeed")

	go func() {
		defer close(doneCh)

		for {
			s, err := l.Accept()
			if err != nil {
				break
			}

			go func() {
				defer l.Close()
				defer s.Close()

				// assume 1 channel per client
				serverCh, err := s.AcceptChannel()
				if !assert.NoError(cfg.t, err, "should succeed") {
					return
				}

				go func() {
					<-shutDownCh
					serverCh.Close()
				}()

				assert.Equal(cfg.t, cfg.channelID, serverCh.ID(), "should match")

				buf := make([]byte, 64*1024)
				var nReceived int
				for {
					n, err := serverCh.Read(buf)
					if err != nil {
						break
					}

					msg := string(buf[:n])
					cfg.log.Debugf("server received %s", msg)
					nReceived++

					n, err = serverCh.Write([]byte(msg))
					assert.NoError(cfg.t, err, "should succeed")
					assert.Equal(cfg.t, len(msg), n, "should match")
				}
			}()
		}

		cfg.log.Debug("server done")
	}()

	return shutDownCh, doneCh
}

func echoClient(cfg *echoClientConfig) <-chan struct{} {
	doneCh := make(chan struct{})

	go func() {
		defer close(doneCh)
		c, err := Dial(&DialConfig{
			Network: "udp4",
			RemoteAddr: &net.UDPAddr{
				IP:   net.ParseIP("127.0.0.1"),
				Port: defaultServerPort,
			},
			LoggerFactory: cfg.loggerFactory,
		})
		assert.NoError(cfg.t, err, "should succeed")
		defer func() {
			err := c.Close()
			assert.NoError(cfg.t, err, "should succeed")
		}()

		clientCh, err := c.OpenChannel(cfg.channelID)
		if !assert.NoError(cfg.t, err, "should succeed") {
			cfg.t.FailNow()
		}
		defer func() {
			clientCh.Close()
			time.Sleep(200 * time.Millisecond)
		}()

		msgs := []string{
			"Hello 1",
			"Hello 2",
			"Hello 3",
			"Hello 4",
		}

		buf := make([]byte, 64*1024)
		for i := 0; i < len(msgs); i++ {
			msg := msgs[i]
			cfg.log.Debugf("client(%d) sending: %s", cfg.id, msg)
			clientCh.Write([]byte(msg))
			n, err := clientCh.Read(buf)
			if err != nil {
				continue
			}
			assert.Equal(cfg.t, msg, string(buf[:n]), "should match")
			cfg.log.Debugf("client(%d) received: %s", cfg.id, string(buf[:n]))
		}

		cfg.log.Debugf("client(%d) done", cfg.id)
	}()

	return doneCh
}

func TestConnectionEchoSingle(t *testing.T) {
	loggerFactory := logging.NewDefaultLoggerFactory()
	log := loggerFactory.NewLogger("test")
	channelID := uint16(123)

	shutDownCh, serverDoneCh := echoServer(&echoServerConfig{
		t:             t,
		channelID:     channelID,
		loggerFactory: loggerFactory,
		log:           log,
	})

	clientDoneCh := echoClient(&echoClientConfig{
		t:             t,
		channelID:     channelID,
		loggerFactory: loggerFactory,
		log:           log,
	})

	<-clientDoneCh
	close(shutDownCh)
	<-serverDoneCh
}

func TestConnectionEchoMulti(t *testing.T) {
	loggerFactory := logging.NewDefaultLoggerFactory()
	log := loggerFactory.NewLogger("test")
	channelID := uint16(123)

	shutDownCh, serverDoneCh := echoServer(&echoServerConfig{
		t:             t,
		channelID:     channelID,
		loggerFactory: loggerFactory,
		log:           log,
	})

	var clientDoneChs []<-chan struct{}

	for i := 0; i < 4; i++ {
		clientDoneCh := echoClient(&echoClientConfig{
			t:             t,
			id:            i,
			channelID:     channelID,
			loggerFactory: loggerFactory,
			log:           log,
		})
		clientDoneChs = append(clientDoneChs, clientDoneCh)
	}

	for _, clientDoneCh := range clientDoneChs {
		<-clientDoneCh
	}
	close(shutDownCh)
	<-serverDoneCh
}
