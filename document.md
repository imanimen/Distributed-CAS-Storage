# CAS - Complete Learning Guide

This document provides both a project overview and explains the networking concepts used in this codebase.

---

# Part A: Project Learning Guide

## What is CAS?

CAS (Content-Addressable Storage) is a peer-to-peer file storage system written in Go. It combines two main components:

1. **Storage Layer** - Files are stored based on their content hash (SHA1), creating a unique address for each file
2. **Network Layer** - Nodes communicate over TCP to exchange data and messages

The key insight: instead of storing files by user-defined names, you store them by their content hash. If two files have the same content, they get the same address - this naturally provides deduplication.

## How the System Works

### 1. Starting the Application (main.go)

When you run `make run`:

```go
func main() {
    tcpOptions := p2p.TCPTransportOption{
        ListenAddr:    ":3000",      // Listen on port 3000
        HandshakeFunc: p2p.NOPHandshakeFunc,  // No validation
        Decoder:       p2p.DefaultDecoder{},   // Raw bytes
        OnPeer:        OnPeer,        // Callback when peer connects
    }
    tr := p2p.NewTCPTransport(tcpOptions)

    if err := tr.ListenAndAccept(); err != nil {
        log.Fatal(err)
    }

    // Read messages from other peers
    go func() {
        for {
            message := <-tr.Consume()
            fmt.Printf("Message: %v\n", message)
        }
    }()

    select {}  // Block forever
}
```

This creates a TCP server that:
- Listens on port 3000 for incoming connections
- Accepts any peer (no handshake validation)
- Reads raw byte messages from peers
- Prints received messages

### 2. Data Flow: Key to File

When storing a file in CAS:

```
User provides key: "momsbestpicture"
        │
        ▼
┌───────────────────┐
│ SHA1 Hash         │
│ 6804429f741...    │
└───────────────────┘
        │
        ▼
┌───────────────────┐
│ Split into chunks │
│ of 5 characters  │
└───────────────────┘
        │
        ▼
┌───────────────────┐
│ Create directory  │
│ 68044/29f74/      │
│ 181a6/3c50c/...   │
└───────────────────┘
        │
        ▼
┌───────────────────┐
│ Store file at     │
│ path/filename     │
└───────────────────┘
```

### 3. Network Communication Flow

```
┌─────────────┐                    ┌─────────────┐
│   Node A    │                    │   Node B    │
│  :3000      │◄─── TCP Connect ──►│  :3000      │
└─────────────┘                    └─────────────┘
       │                                 │
       │  1. Accept()                   │
       │  2. NewTCPPeer(conn)           │
       │  3. HandshakeFunc(peer)        │
       │  4. OnPeer(peer)               │
       │  5. Decode() loop              │
       │                                │
       └──────── RPC {From, Payload} ──┘
```

### 4. File-by-File Code Walkthrough

#### store.go - The Storage Engine

| File | Purpose |
|------|---------|
| `CASPathTransformFunc(key)` | Takes a string key, SHA1 hashes it, splits into 5-char directories |
| `Store` struct | Holds storage configuration (root folder, path transform) |
| `Read(key)` | Opens file by key, returns io.Reader |
| `writeStream(key, r)` | Takes key + io.Reader, writes to disk |
| `Delete(key)` | Removes file and directory |
| `Exists(key)` | Checks if file exists |

#### p2p/ - The Networking Layer

| File | Purpose |
|------|---------|
| `transport.go` | Defines `Peer` and `Transport` interfaces |
| `tcp_transport.go` | TCP implementation of Transport interface |
| `message.go` | `RPC` struct: `From` (address) + `Payload` (bytes) |
| `encoding.go` | `Decoder` interface - converts bytes to RPC |
| `handshaker.go` | `HandshakeFunc` - validate incoming connections |

---

# Part B: Networking Concepts

## 1. Content-Addressable Storage (CAS)

**What it is**: A storage method where data is retrieved based on its content, not its location.

**How it works here**:
- A key (like "myfile.txt") is hashed using SHA1
- The hash determines where the file is stored
- Same content = same hash = same address = automatic deduplication

**Code reference**: `store.go:CASPathTransformFunc`

```go
func CASPathTransformFunc(key string) PathKey {
    hash := sha1.Sum([]byte(key))
    hashString := hex.EncodeToString(hash[:])
    
    // Split into 5-character chunks
    blockSize := 5
    sliceLength := len(hashString) / blockSize
    paths := make([]string, sliceLength)
    
    for i := 0; i < sliceLength; i++ {
        from, to := i*blockSize, (i*blockSize)+blockSize
        paths[i] = hashString[from:to]
    }
    
    return PathKey{
        PathName: strings.Join(paths, "/"),
        FileName: hashString,
    }
}
```

## 2. Peer-to-Peer (P2P) Networking

**What it is**: A decentralized network model where each node (peer) can act as both client and server.

**How it works here**:
- Each node runs a TCP server on a specific port
- Nodes can connect to each other directly
- No central server required - any node can communicate with any other node

**Key components**:
- `Peer` interface - represents a remote node
- `TCPPeer` - implementation for TCP connections
- `OnPeer` callback - triggered when a new peer connects

## 3. TCP Transport Layer

**What it is**: The underlying network protocol for reliable, ordered, error-checked communication.

**How it works here**:
- `net.Listen("tcp", ":3000")` - creates a TCP listener
- `listener.Accept()` - waits for incoming connections
- `net.Conn` - represents an active connection

**Code reference**: `p2p/tcp_transport.go`

```go
func (t *TCPTransport) ListenAndAccept() error {
    var err error
    t.listener, err = net.Listen("tcp", t.ListenAddr)
    if err != nil {
        return err
    }
    go t.acceptor()  // Handle connections in background
    return nil
}

func (t *TCPTransport) acceptor() {
    for {
        conn, err := t.listener.Accept()
        if err != nil {
            fmt.Printf("TCP listener accept error: %v\n", err)
        }
        go t.connector(conn)  // Handle each connection
    }
}
```

**Key concepts**:
- **Listener**: Server-side, waits for connections (`net.Listen`)
- **Connection**: Active communication channel (`net.Conn`)
- **Goroutines**: Each connection handled concurrently with `go t.connector(conn)`

## 4. RPC (Remote Procedure Call)

**What it is**: A way to send messages between distributed systems.

**How it works here**:
- Messages are represented as `RPC` struct
- Contains sender address (`From`) and message data (`Payload`)
- No actual "procedure calls" - just raw data exchange

**Code reference**: `p2p/message.go`

```go
type RPC struct {
    From    net.Addr  // Sender's network address
    Payload []byte    // Message data
}
```

## 5. Handshake Protocol

**What it is**: Initial negotiation between peers before full communication begins.

**How it works here**:
- `HandshakeFunc` is called when a new connection is established
- Can validate the remote peer
- Can reject connection by returning an error

**Code reference**: `p2p/handshaker.go`

```go
type HandshakeFunc func(Peer) error

// No-op handshake - always accepts
func NOPHandshakeFunc(Peer) error { return nil }

// Example custom handshake:
func MyHandshakeFunc(peer Peer) error {
    // Check peer identity, version, etc.
    return nil  // or return ErrInvalidHandShake
}
```

**Connection flow**:
```
Accept connection → NewTCPPeer(conn) → HandshakeFunc(peer) → OnPeer(peer) → Read loop
```

## 6. Message Encoding/Decoding

**What it is**: Converting data structures to/from bytes for network transmission.

**How it works here**: The `Decoder` interface defines how incoming bytes become `RPC` structs.

**Code reference**: `p2p/encoding.go`

### DefaultDecoder (Raw Bytes)
```go
type DefaultDecoder struct{}

func (dec DefaultDecoder) Decode(r io.Reader, msg *RPC) error {
    buf := make([]byte, 1028)
    n, err := r.Read(buf)
    if err != nil {
        return err
    }
    msg.Payload = buf[:n]
    return nil
}
```

### GOBDecoder (Go's Binary Encoding)
```go
type GOBDecoder struct{}

func (dec GOBDecoder) Decode(r io.Reader, msg *RPC) error {
    return gob.NewDecoder(r).Decode(msg)
}
```

**Usage**:
```go
tcpOptions := p2p.TCPTransportOption{
    // Use raw bytes
    Decoder: p2p.DefaultDecoder{},
    // OR use structured encoding
    Decoder: p2p.GOBDecoder{},
}
```

## 7. Go Interfaces

**What it is**: A type that defines a contract without implementing it.

**Why it matters here**: The code uses interfaces to make components pluggable:

```go
// Transport is an interface - any implementation works
type Transport interface {
    ListenAndAccept() error
    Consume() chan<- RPC
}

// TCPTransport implements Transport
var _ Transport = (*TCPTransport)(nil)

// Decoder is an interface - plug in different encodings
type Decoder interface {
    Decode(io.Reader, *RPC) error
}
```

**Benefits**:
- Swap TCP for UDP without changing other code
- Use different message encoders without modification
- Easy to test with mock implementations

## 8. Goroutines and Channels

**What it is**: Go's concurrency model - lightweight threads and communication channels.

**How it works here**:

```go
// acceptor runs in background, accepts connections
go t.acceptor()

// connector reads in a loop
for {
    err := t.Decoder.Decode(conn, &rpc)
    // ...
}

// Messages sent via channel
t.rpcChan <- rpc

// Main loop reads from channel
for msg := range tr.Consume() {
    fmt.Printf("Message: %v\n", msg)
}
```

---

# Summary

This project demonstrates:

| Concept | Implementation |
|---------|----------------|
| CAS | SHA1-based path transformation in `store.go` |
| P2P | TCPTransport accepts connections from any peer |
| TCP | `net.Listen`, `net.Conn` in `tcp_transport.go` |
| RPC | `RPC` struct with `From` and `Payload` fields |
| Handshake | `HandshakeFunc` validates new connections |
| Decoding | `Decoder` interface with two implementations |
| Interfaces | `Peer`, `Transport`, `Decoder` abstractions |
| Concurrency | Goroutines for acceptor/connector, channels for messaging |

---

# Next Steps

To extend this project, consider:

1. **Add file transfer** - Send actual file content over RPC
2. **Implement DHT** - Distributed hash table for finding files
3. **Add encryption** - Encrypt payloads for security
4. **Protocol versioning** - Add version to handshake
5. **Connection pooling** - Reuse connections instead of creating new ones
6. **Message types** - Add message types (request, response, error)