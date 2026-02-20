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

func NewTCPTransport(options TCPTransportOption) *TCPTransport {
	return &TCPTransport{
		TCPTransportOption: options,
		rpcChan:            make(chan RPC),
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}
	go t.acceptor()

	return nil
}

// Close implements the Peer interface
func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

// Consume implements the Transport Interface
// which will return read-only channel for reading the incoming new message received from
// another Peer in the network
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcChan
}

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
