package p2p

import "net"

// RPC represents a remote procedure call message sent between nodes in the network.
// It contains the source address and the payload data.
type RPC struct {
	From    net.Addr // The network address of the sender
	Payload []byte   // The message payload/data
}
