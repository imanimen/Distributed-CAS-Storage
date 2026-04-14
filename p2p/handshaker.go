package p2p

import "errors"

// ErrInvalidHandShake is returned when the handshake between the local and remote node fails.
var ErrInvalidHandShake = errors.New("invalid handshake")

// HandshakeFunc is a function type that performs a handshake with a peer.
// It returns an error if the handshake fails.
type HandshakeFunc func(Peer) error

// NOPHandshakeFunc is a no-op handshake function that always succeeds.
// Use this when no handshake validation is required.
func NOPHandshakeFunc(Peer) error { return nil }
