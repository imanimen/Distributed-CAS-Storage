package main

import (
	"fmt"
	"log"

	"github.com/imanimen/cas/p2p"
)

func main() {
	tcpOptions := p2p.TCPTransportOption{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		OnPeer: func(peer p2p.Peer) error {
			return fmt.Errorf("failed to on peer func %v\n", peer)
		},
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
