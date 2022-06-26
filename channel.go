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
	Read(p []byte) (int, error)                            // Read reads a packet of len(p) bytes as binary data
	ReadDataChannel(p []byte) (int, bool, error)           // ReadDataChannel reads a packet of len(p) bytes
	MessagesSent() uint32                                  // MessagesSent returns the number of messages sent
	MessagesReceived() uint32                              // MessagesReceived returns the number of messages received
	BytesSent() uint64                                     // BytesSent returns the number of bytes sent
	BytesReceived() uint64                                 // BytesReceived returns the number of bytes received
	StreamIdentifier() uint16                              // StreamIdentifier returns the Stream identifier associated to the stream.
	Write(p []byte) (int, error)                           // Write writes len(p) bytes from p as binary data
	WriteDataChannel(p []byte, isString bool) (int, error) // WriteDataChannel writes len(p) bytes from p
	Close() error                                          // Close closes the DataChannel and the underlying SCTP stream.
	BufferedAmount() uint64                                // BufferedAmount returns the number of bytes of data currently queued to be sent over this stream.
	BufferedAmountLowThreshold() uint64                    // BufferedAmountLowThreshold returns the number of bytes of buffered outgoing data that is considered "low." Defaults to 0.
	SetBufferedAmountLowThreshold(th uint64)               // SetBufferedAmountLowThreshold is used to update the threshold. See BufferedAmountLowThreshold().
	OnBufferedAmountLow(f func())                          // OnBufferedAmountLow sets the callback handler which would be called when the number of bytes of outgoing data buffered is lower than the threshold.
}
