package rudp

import (
	dcep "github.com/pion/datachannel"
)

type Config struct {
	ChannelType          dcep.ChannelType
	Negotiated           bool
	Priority             uint16
	ReliabilityParameter uint32
	Label                string
	Protocol             string
}

type Channel interface {
	Read(p []byte) (int, error)              // Read reads a packet of len(p) bytes as binary data
	Write(p []byte) (int, error)             // Write writes len(p) bytes from p as binary data
	Close() error                            // Close closes the DataChannel and the underlying SCTP stream.
	StreamIdentifier() uint16                // StreamIdentifier returns the Stream identifier associated to the stream.
	BufferedAmount() uint64                  // BufferedAmount returns the number of bytes of data currently queued to be sent over this stream.
	BufferedAmountLowThreshold() uint64      // BufferedAmountLowThreshold returns the number of bytes of buffered outgoing data that is considered "low." Defaults to 0.
	SetBufferedAmountLowThreshold(th uint64) // SetBufferedAmountLowThreshold is used to update the threshold. See BufferedAmountLowThreshold().
	OnBufferedAmountLow(f func())            // OnBufferedAmountLow sets the callback handler which would be called when the number of bytes of outgoing data buffered is lower than the threshold.
	Config() Config                          // Channel configuration
}

type dataChannel struct {
	dc     *dcep.DataChannel
	config Config
}

func (dc *dataChannel) Read(p []byte) (int, error) {
	return dc.dc.Read(p)
}

func (dc *dataChannel) Write(p []byte) (int, error) {
	return dc.dc.Write(p)
}

func (dc *dataChannel) Close() error {
	return dc.dc.Close()
}

func (dc *dataChannel) StreamIdentifier() uint16 {
	return dc.dc.StreamIdentifier()
}

func (dc *dataChannel) BufferedAmount() uint64 {
	return dc.dc.BufferedAmount()
}

func (dc *dataChannel) BufferedAmountLowThreshold() uint64 {
	return dc.dc.BufferedAmountLowThreshold()
}

func (dc *dataChannel) SetBufferedAmountLowThreshold(th uint64) {
	dc.dc.SetBufferedAmountLowThreshold(th)
}

func (dc *dataChannel) OnBufferedAmountLow(f func()) {
	dc.dc.OnBufferedAmountLow(f)
}

func (dc *dataChannel) Config() Config {
	return dc.config
}
