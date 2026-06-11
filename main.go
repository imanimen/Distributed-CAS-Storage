package main

import (
	"fmt"
	"log"

	"github.com/imanimen/cas/p2p"
)

// OnPeer is a callback function that is invoked when a new peer connects.
// It receives the peer and can return an error to reject the connection.
func OnPeer(peer p2p.Peer) error {
	fmt.Printf("OnPeer %v\n", peer) // todo: change
	return nil
}

func main() {
	tcpOptions := p2p.TCPTransportOption{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		OnPeer:        OnPeer,
	}
	tr := p2p.NewTCPTransport(tcpOptions)

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			message := <-tr.Consume()
			fmt.Printf("Message: %v\n", message)
		}
	}()

	select {}
}
