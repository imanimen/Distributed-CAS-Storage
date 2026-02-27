# CAS - Content Addressable Storage

A distributed peer-to-peer file storage system written in Go. Files are stored using SHA1 content-addressable paths, and nodes communicate over a TCP transport layer.

## Features

- **Content-Addressable Storage** - Files are stored in a SHA1-derived directory tree, ensuring deduplication and integrity
- **P2P Networking** - TCP-based transport with pluggable handshake and message encoding
- **Streaming I/O** - Read and write files via Go's `io.Reader`/`io.Writer` interfaces
- **Extensible Encoding** - Ships with a raw byte decoder (`DefaultDecoder`) and a GOB decoder (`GOBDecoder`)

## Architecture

```
main.go                       Entry point - wires transport + store
│
├── store.go                  CAS storage engine (core)
│   └── PathTransformFunc     SHA1 key → nested directory path
│
└── p2p/                      Networking layer
    ├── transport.go          Peer & Transport interfaces
    ├── tcp_transport.go      TCP implementation
    ├── message.go            RPC message struct
    ├── encoding.go           Decoder interface + implementations
    └── handshaker.go         Handshake function types
```

## How CAS Path Transform Works

A key is hashed with SHA1, and the hex digest is split into 5-character segments to form a directory tree:

```
"momsbestpicture"
  → SHA1: 6804429f74181a63c50c3d81d733a12f14a353ff
  → Path: 68044/29f74/181a6/3c50c/3d81d/733a1/2f14a/353ff
```

This distributes files across directories to avoid filesystem bottlenecks.

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
tcpOptions := p2p.TCPTransportOption{
    ListenAddr:    ":3000",
    HandshakeFunc: p2p.NOPHandshakeFunc,
    Decoder:       p2p.DefaultDecoder{},
    OnPeer: func(peer p2p.Peer) error {
        fmt.Printf("peer connected: %v\n", peer)
        return nil
    },
}
tr := p2p.NewTCPTransport(tcpOptions)
tr.ListenAndAccept()

// read incoming messages
for msg := range tr.Consume() {
    fmt.Printf("%s: %s\n", msg.From, msg.Payload)
}
```

### Content-Addressable Store

```go
store := NewStore(StoreOptions{
    Root:              "mydata",
    PathTransformFunc: CASPathTransformFunc,
})

// write
store.writeStream("photo.jpg", bytes.NewReader(data))

// read
reader, _ := store.Read("photo.jpg")

// check existence
store.Exists("photo.jpg")

// delete
store.Delete("photo.jpg")
```

## Project Structure

```
.
├── main.go               Application entry point
├── store.go              CAS storage engine
├── store_test.go         Storage tests
├── Makefile              Build/run/test targets
├── go.mod
├── go.sum
└── p2p/
    ├── transport.go          Peer and Transport interfaces
    ├── tcp_transport.go      TCPPeer, TCPTransport, connection lifecycle
    ├── tcp_transport_test.go Transport tests
    ├── message.go            RPC struct (From, Payload)
    ├── encoding.go           DefaultDecoder (raw bytes), GOBDecoder
    └── handshaker.go         HandshakeFunc, NOPHandshakeFunc
```

## License

This project is for educational and experimental purposes.
