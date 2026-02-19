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

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

type TCPTransport struct {
	listenAddress string
	listener      net.Listener
	shakeHands    HandshakeFunc
	decoder       Decoder

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

type Temp struct{}

func NewTCPTransport(listenAddr string) *TCPTransport {
	return &TCPTransport{
		shakeHands:    NOPHandshakeFunc,
		listenAddress: listenAddr,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.listenAddress)
	if err != nil {
		return err
	}
	go t.acceptor()

	return nil
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
	peer := NewTCPPeer(conn, true)

	if err := t.shakeHands(peer); err != nil {
		fmt.Printf("TCP handshake error: %v\n", err)
	}

	// read loop
	msg := &Temp{}
	for {
		if err := t.decoder.Decode(conn, msg); err != nil {
			fmt.Printf("TCP error: %s\n", err)
		}
	}

}
