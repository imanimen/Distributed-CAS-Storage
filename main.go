package main

import (
	"fmt"
	"log"

	"github.com/imanimen/cas/p2p"
)

func OnPeer(peer p2p.Peer) error {
	fmt.Printf("OnPeer %s\n", peer) // todo: chang
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
