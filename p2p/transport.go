package p2p

// Peer represents a remote node in the P2P network.
// It provides an interface for closing the connection to the peer.
type Peer interface {
	Close() error
}

// Transport is an interface for network communication between nodes.
// Implementations handle the specific protocol (TCP, UDP, WebSocket, etc.).
// ListenAndAccept starts listening for incoming connections.
// Consume returns a channel for receiving incoming RPC messages from peers.
type Transport interface {
	ListenAndAccept() error
	Consume() chan<- RPC
}
