# CAS - Content Addressable Storage

A distributed peer-to-peer file storage system written in Go. Files are stored using SHA1 content-addressable paths, and nodes communicate over a TCP transport layer.

## Features

- **Content-Addressable Storage** - Files are stored in a SHA1-derived directory tree, ensuring deduplication and integrity
- **P2P Networking** - TCP-based transport with pluggable handshake and message encoding
- **Streaming I/O** - Read and write files via Go's `io.Reader`/`io.Writer` interfaces
- **Extensible Encoding** - Ships with a raw byte decoder (`DefaultDecoder`) and a GOB decoder (`GOBDecoder`)
- **Pluggable Handshake** - Custom handshake functions for peer validation

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         main.go                             │
│              (Entry Point - initializes P2P)               │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                       p2p/ (Networking Layer)                │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   Peer      │  │  Transport  │  │      RPC            │  │
│  │  interface  │  │  interface  │  │   (message struct)  │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
│         │                │                    │              │
│         ▼                ▼                    ▼              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │                 tcp_transport.go                      │  │
│  │   TCPPeer ←── TCPTransport ←── TCPTransportOption    │  │
│  └──────────────────────────────────────────────────────┘  │
│         │                │                    │              │
│         ▼                ▼                    ▼              │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────┐   │
│  │handshaker.go│ │ encoding.go │ │     message.go      │   │
│  │HandshakeFunc│ │  Decoder    │ │  RPC{From,Payload}  │   │
│  └─────────────┘ └─────────────┘ └─────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                       store.go (Storage Layer)                │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Store                                                │   │
│  │  - Root: "casnetwork" (default)                      │   │
│  │  - PathTransformFunc: SHA1-based hierarchy          │   │
│  │                                                      │   │
│  │  Methods:                                             │   │
│  │  - Read(key) → io.Reader                             │   │
│  │  - writeStream(key, io.Reader)                       │   │
│  │  - Delete(key)                                       │   │
│  │  - Exists(key) → bool                                │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## How CAS Path Transform Works

A key is hashed with SHA1, and the hex digest is split into 5-character segments to form a directory tree:

```
"momsbestpicture"
  → SHA1: 6804429f74181a63c50c3d81d733a12f14a353ff
  → Path: 68044/29f74/181a6/3c50c/3d81d/733a1/2f14a/353ff
```

This distributes files across directories to avoid filesystem bottlenecks and enables content-based deduplication.

## Getting Started

### Prerequisites

- Go 1.25+

### Build & Run

```sh
make run
```

This builds the binary to `bin/cas` and starts a node listening on `:3000`.

### Test

```sh
make test
```

### Send a message to a running node

```sh
echo "hello" | nc localhost 3000
```

## Usage

### P2P Transport

```go
package main

import (
    "fmt"
    "log"

    "github.com/imanimen/cas/p2p"
)

func OnPeer(peer p2p.Peer) error {
    fmt.Printf("peer connected: %v\n", peer)
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

    // read incoming messages
    for msg := range tr.Consume() {
        fmt.Printf("%s: %s\n", msg.From, msg.Payload)
    }
}
```

### Content-Addressable Store

```go
package main

import (
    "bytes"
    "fmt"
)

func main() {
    store := NewStore(StoreOptions{
        Root:              "mydata",
        PathTransformFunc: CASPathTransformFunc,
    })

    data := []byte("hello world")

    // write
    if err := store.writeStream("myfile.txt", bytes.NewReader(data)); err != nil {
        fmt.Println("Error:", err)
    }

    // read
    reader, _ := store.Read("myfile.txt")
    fmt.Println("Content:", reader)

    // check existence
    fmt.Println("Exists:", store.Exists("myfile.txt"))

    // delete
    store.Delete("myfile.txt")
}
```

## Project Structure

```
.
├── main.go                   Application entry point
├── store.go                  CAS storage engine
├── store_test.go             Storage tests
├── README.md                 Project documentation
├── DOCUMENTATION.md          Detailed documentation
├── Makefile                  Build/run/test targets
├── go.mod
├── go.sum
└── p2p/
    ├── transport.go          Peer and Transport interfaces
    ├── tcp_transport.go      TCPPeer, TCPTransport, connection lifecycle
    ├── tcp_transport_test.go Transport tests
    ├── message.go            RPC struct (From, Payload)
    ├── encoding.go           Decoder interface: DefaultDecoder, GOBDecoder
    └── handshaker.go         HandshakeFunc, NOPHandshakeFunc, ErrInvalidHandShake
```

## Core Interfaces

### Peer Interface
```go
type Peer interface {
    Close() error
}
```

### Transport Interface
```go
type Transport interface {
    ListenAndAccept() error
    Consume() chan<- RPC
}
```

### Decoder Interface
```go
type Decoder interface {
    Decode(io.Reader, *RPC) error
}
```

## License

This project is for educational and experimental purposes.