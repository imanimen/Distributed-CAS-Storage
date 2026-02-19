package main

import (
	"log"

	"github.com/imanimen/cas/p2p"
)

func main() {
	tcpOptions := p2p.TCPTransportOption{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.GOBDecoder{},
	}
	tr := p2p.NewTCPTransport(tcpOptions)

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}
}
