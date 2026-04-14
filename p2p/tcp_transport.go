package p2p

import (
	"fmt"
	"net"
	"sync"
)

// TCPPeer represents the remote node over a TCP
// established connection.
type TCPPeer struct {
	// conn is underlying connection of the peer
	conn net.Conn

	// if we dial a connection -> outbound == true
	// if we accept and retrieve a connection -> outbound == false (inbound)
	outbound bool
}

type TCPTransportOption struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

// NewTCPPeer creates a new TCPPeer with the given connection.
// The outbound parameter indicates whether this peer initiated the connection (true)
// or accepted it (false).
func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

type TCPTransport struct {
	TCPTransportOption
	listener net.Listener
	rpcChan  chan RPC

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

type Temp struct{}

// NewTCPTransport creates a new TCPTransport with the given options.
func NewTCPTransport(options TCPTransportOption) *TCPTransport {
	return &TCPTransport{
		TCPTransportOption: options,
		rpcChan:            make(chan RPC),
	}
}

// ListenAndAccept starts the TCP listener on the configured address
// and begins accepting incoming connections in a goroutine.
// It returns an error if the listener fails to start.
func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}
	go t.acceptor()

	return nil
}

// Close closes the underlying TCP connection for this peer.
func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

// Consume returns a read-only channel that receives incoming RPC messages
// from remote peers in the network.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcChan
}

// acceptor runs in a goroutine and accepts incoming TCP connections.
// For each new connection, it spawns a connector goroutine to handle the peer.
func (t *TCPTransport) acceptor() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP listener accept error: %v\n", err)
		}
		fmt.Printf("New incoming connection %+v\n", conn)

		go t.connector(conn)
	}
}

// connector handles an established TCP connection from a remote peer.
// It performs the handshake, invokes the OnPeer callback, and starts
// reading incoming RPC messages in a loop.
func (t *TCPTransport) connector(conn net.Conn) {
	var err error

	defer func() {
		fmt.Printf("Dropping connection %s", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, true)

	if err := t.HandshakeFunc(peer); err != nil {
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	// read loop
	rpc := RPC{}
	for {
		err := t.Decoder.Decode(conn, &rpc)
		if err != nil {
			fmt.Printf("TCP error: %s\n", err)
			return
		}

		rpc.From = conn.RemoteAddr()
		t.rpcChan <- rpc
	}

}
