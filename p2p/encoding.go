package p2p

import (
	"encoding/gob"
	"io"
)

// Decoder is an interface for decoding incoming messages from an io.Reader into an RPC.
type Decoder interface {
	Decode(io.Reader, *RPC) error
}

// GOBDecoder uses the Go encoding/gob package to decode messages.
type GOBDecoder struct{}

// Decode reads a message from the reader and decodes it into the provided RPC using GOB encoding.
func (dec GOBDecoder) Decode(r io.Reader, msg *RPC) error {
	return gob.NewDecoder(r).Decode(msg)
}

// DefaultDecoder is a simple decoder that reads raw bytes from the stream.
// It reads up to 1028 bytes into the RPC payload.
type DefaultDecoder struct{}

// Decode reads up to 1028 bytes from the reader and stores them in the RPC payload.
func (dec DefaultDecoder) Decode(r io.Reader, msg *RPC) error {
	buf := make([]byte, 1028)
	n, err := r.Read(buf)
	if err != nil {
		return err
	}

	msg.Payload = buf[:n]

	return nil
}
