# CAS Project Documentation

## Project Overview

CAS (Content Addressable Storage) is a peer-to-peer storage system written in Go. It provides:
- A TCP-based transport layer for network communication between peers
- A file storage system with SHA1-based content addressable path transformation
- Support for handshake procedures and custom encoding/decoding

---

## File Reference

### main.go

**Purpose**: Application entry point - initializes and starts the P2P TCP transport server.

**Location**: `/Users/imanimen/Development/go/cas/main.go`

**Functions**:

| Function | Signature | Description |
|----------|-----------|-------------|
| OnPeer | `func(peer p2p.Peer) error` | Callback function triggered when a new peer connects. Currently prints peer info. |
| main | `func()` | Initializes TCP transport with options, starts listening, and consumes messages from the channel. |

**TCPTransportOption Configuration**:
- `ListenAddr`: ":3000" - Port to listen on
- `HandshakeFunc`: p2p.NOPHandshakeFunc - Handshake validation function
- `Decoder`: p2p.DefaultDecoder{} - Message decoder
- `OnPeer`: OnPeer callback function

**Commit History**:
- `6f18bc8` - fx: OnPeer method to make the application work

---

### store.go

**Purpose**: Provides content-addressable file storage with SHA1-based path transformation. Splits hashes into directory paths for efficient file organization.

**Location**: `/Users/imanimen/Development/go/cas/store.go`

**Types**:

| Type | Description |
|------|-------------|
| `PathTransformFunc` | Function type: `func(string) PathKey` - transforms a key into a path structure |
| `PathKey` | Struct containing `PathName` (nested directory path) and `FileName` (full hash) |
| `StoreOptions` | Configuration struct with `PathTransformFunc` |
| `Store` | Main storage struct holding StoreOptions |

**Constants/Variables**:

| Name | Type | Description |
|------|------|-------------|
| `DefaultPathTransformFunc` | `func(key string) string` | Default identity function (returns key as-is) |
| `CASPathTransformFunc` | `PathTransformFunc` | SHA1-based path transformation (default implementation) |

**Path Transformation Algorithm** (CASPathTransformFunc):
1. Generate SHA1 hash of the key
2. Convert hash to hexadecimal string
3. Split into blocks of 5 characters
4. Join with "/" to create nested directory structure

Example: `momsbestpicture` → `6804429f74181a63c50c3d81d733a12f14a353ff` → `68044/29f74/181a6/3c50c/3d81d/733a1/2f14a/353ff`

**Methods**:

| Method | Signature | Description |
|--------|-----------|-------------|
| CASPathTransformFunc | `func(key string) PathKey` | Transforms a key into PathKey using SHA1 hash |
| NewStore | `func(options StoreOptions) *Store` | Creates a new Store instance |
| FullPath | `func(p PathKey) string` | Returns full path: `PathName/FileName` |
| Read | `func(s *Store, key string) (io.Reader, error)` | Reads file content by key |
| writeStream | `func(s *Store, key string, r io.Reader) error` | Writes data from reader to disk |
| readStream | `func(s *Store, key string) (io.ReadCloser, error)` | Opens file for reading |
| Delete | `func(s *Store, key string) error` | Deletes file and directory structure |
| Exists | `func(s *Store, key string) bool` | Checks if file exists |
| FirstPathName | `func(p PathKey) string` | Returns first directory in path |

**Commit History**:
- `a9a19c0` - chore: filename method and other things
- `2d38a5d` - chore: store is now storing
- `11747c8` - feat: store implementation

---

### store_test.go

**Purpose**: Unit tests for the storage functionality.

**Location**: `/Users/imanimen/Development/go/cas/store_test.go`

**Test Functions**:

| Function | Description |
|----------|-------------|
| `TestPathTransformFunc` | Tests SHA1 path transformation with key "momsbestpicture", verifies PathName and FileName |
| `TestStore` | Integration test: writes data, reads it back, verifies content, then deletes |

**Commit History**:
- `4d95b9f` - chore: latest commits -> delete file + tests

---

### p2p/transport.go

**Purpose**: Defines core interfaces for P2P communication.

**Location**: `/Users/imanimen/Development/go/cas/p2p/transport.go`

**Interfaces**:

| Interface | Description |
|-----------|-------------|
| `Peer` | Represents a remote node. Method: `Close() error` |
| `Transport` | Handles communication between nodes. Methods: `ListenAndAccept() error`, `Consume() chan<- RPC` |

**Commit History**:
- `cc4c68c` - chore: interfaces

---

### p2p/tcp_transport.go

**Purpose**: TCP implementation of the Transport interface for peer-to-peer communication.

**Location**: `/Users/imanimen/Development/go/cas/p2p/tcp_transport.go`

**Types**:

| Type | Description |
|------|-------------|
| `TCPPeer` | Represents a remote node over TCP connection. Has `conn` (net.Conn) and `outbound` bool |
| `TCPTransportOption` | Configuration: ListenAddr, HandshakeFunc, Decoder, OnPeer callback |
| `TCPTransport` | Main transport struct with listener, rpcChan, peers map |
| `Temp` | Unused placeholder type |

**Methods**:

| Method | Signature | Description |
|--------|-----------|-------------|
| NewTCPPeer | `func(conn net.Conn, outbound bool) *TCPPeer` | Creates new TCP peer |
| NewTCPTransport | `func(options TCPTransportOption) *TCPTransport` | Creates new TCP transport |
| ListenAndAccept | `func(t *TCPTransport) error` | Starts TCP listener and runs acceptor goroutine |
| Close | `func(p *TCPPeer) error` | Implements Peer interface - closes connection |
| Consume | `func(t *TCPTransport) <-chan RPC` | Returns channel for incoming messages |
| acceptor | `func(t *TCPTransport)` | Goroutine accepting incoming connections |
| connector | `func(t *TCPTransport, conn net.Conn)` | Goroutine handling connection lifecycle: handshake → OnPeer → read loop |

**Flow**:
1. `ListenAndAccept()` starts listener and runs `acceptor()`
2. `acceptor()` accepts connections and spawns `connector()` goroutines
3. `connector()` performs handshake, calls OnPeer callback, then enters read loop
4. Read loop decodes messages and sends to `rpcChan`

**Commit History**:
- `2f27260` - feat: reader is now on
- `b9571bc` - chroe: dropping connection logic
- `4e47305` - chore: consume method to handle the messages on the channel
- `c11823f` - HOTFIX: TCP option configurations
- `e979df0` - chore: tcp peer connection is now working
- `b4e253b` - chore: test + listenAndAccept function to handle connector + acceptor
- `554255f` - chore(tests): test the TCP transport
- `cc4c68c` - chore: interfaces (partial)

---

### p2p/tcp_transport_test.go

**Purpose**: Unit tests for TCP transport.

**Location**: `/Users/imanimen/Development/go/cas/p2p/tcp_transport_test.go`

**Test Functions**:

| Function | Description |
|----------|-------------|
| `TestTCPTransport` | Creates TCP transport, verifies ListenAddr, starts ListenAndAccept |

**Dependencies**: Uses `github.com/stretchr/testify/assert`

**Commit History**:
- `554255f` - chore(tests): test the TCP transport

---

### p2p/message.go

**Purpose**: Defines the RPC message structure for peer-to-peer communication.

**Location**: `/Users/imanimen/Development/go/cas/p2p/message.go`

**Types**:

| Type | Description |
|------|-------------|
| `RPC` | Struct holding `From` (net.Addr) and `Payload` ([]byte) - arbitrary data sent between peers |

**Commit History**:
- `f7b2cb1` - chore: from Message to RPC, we have come a long way
- `cefdf90` - chore: new property in message struct is From
- `2be32c7` - chore: finally the messages are showing but as bytes
- `9238f9a` - chore: reading bytes in the first place
- `5e4bcea` - chore: add Message.go

---

### p2p/encoding.go

**Purpose**: Provides message decoding implementations for P2P communication.

**Location**: `/Users/imanimen/Development/go/cas/p2p/encoding.go`

**Interfaces**:

| Interface | Description |
|-----------|-------------|
| `Decoder` | Interface with `Decode(io.Reader, *RPC) error` method |

**Types**:

| Type | Description |
|------|-------------|
| `GOBDecoder` | Uses Go's encoding/gob for structured message decoding |
| `DefaultDecoder` | Reads raw bytes (1028 byte buffer) - returns payload as-is |

**Commit History**:
- `88d5612` - chore: Gob Decoder implemented and new TCP options
- `85748da` - chore: decoder
- `0abebc0` - chore: handshaker & decoder

---

### p2p/handshaker.go

**Purpose**: Defines handshake functionality for peer connection validation.

**Location**: `/Users/imanimen/Development/go/cas/p2p/handshaker.go`

**Types/Variables**:

| Name | Type | Description |
|------|------|-------------|
| `ErrInvalidHandShake` | `error` | Error returned when handshake fails |
| `HandshakeFunc` | `func(Peer) error` | Function type for handshake validation |
| `NOPHandshakeFunc` | `HandshakeFunc` | No-op handshake that always succeeds |

**Commit History**:
- `cfd5184` - chore: handshake func
- `9680de2` - HOTIFX: shakeHands func
- `0abebc0` - chore: handshaker & decoder

---

## Git Commit History

| Commit | Description |
|--------|-------------|
| `4d95b9f` | chore: latest commits -> delete file + tests |
| `bb3a071` | HOTFIX: replace deprecated lib with io.ReadAll() |
| `2f27260` | feat: reader is now on |
| `a9a19c0` | chore: filename method and other things |
| `2d38a5d` | chore: store is now storing |
| `11747c8` | feat: store implementation |
| `6f18bc8` | fx: OnPeer method to make the application work |
| `b9571bc` | chore: dropping connection logic |
| `4e47305` | chore: consume method to handle the messages on the channel |
| `5793579` | fix: test is now working |
| `f7b2cb1` | chore: from Message to RPC, we have come a long way |
| `cefdf90` | chore: new property in message struct is From |
| `2be32c7` | chore: finally the messages are showing but as bytes |
| `9238f9a` | chore: reading bytes in the first place |
| `5e4bcea` | chore: add Message.go |
| `c11823f` | HOTFIX: TCP option configurations |
| `88d5612` | chore: Gob Decoder implemented and new TCP options |
| `cfd5184` | chore: handshake func |
| `9680de2` | HOTIFX: shakeHands func |
| `85748da` | chore: decoder |
| `0abebc0` | chore: handshaker & decoder |
| `816b525` | build workflow |
| `e979df0` | chore: tcp peer connection is now working |
| `b4e253b` | chore: test + listenAndAccept function to handle connector + acceptor |
| `554255f` | chore(tests): test the TCP transport |
| `cc4c68c` | chore: interfaces |
| `42c7140` | first commit |

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         main.go                             │
│                    (Entry Point)                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  TCPTransportOption                                  │   │
│  │  - ListenAddr: ":3000"                              │   │
│  │  - HandshakeFunc                                    │   │
│  │  - Decoder                                          │   │
│  │  - OnPeer callback                                  │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    p2p/transport.go                         │
│  ┌─────────────┐              ┌─────────────┐              │
│  │   Peer      │              │  Transport  │              │
│  │  (interface)│              │ (interface) │              │
│  └─────────────┘              └─────────────┘              │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                   p2p/tcp_transport.go                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │  TCPPeer    │  │TCPTransport │  │ TCPTransportOption  │ │
│  │  - conn     │  │ - listener │  │ - ListenAddr       │ │
│  │  - outbound │  │ - rpcChan  │  │ - HandshakeFunc    │ │
│  └─────────────┘  │ - peers    │  │ - Decoder          │ │
│                   └─────────────┘  │ - OnPeer          │ │
│                                      └─────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                            │
          ┌─────────────────┼─────────────────┐
          ▼                 ▼                 ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────────┐
│ p2p/encoding.go │ │p2p/handshaker.go│ │   p2p/message.go    │
│  - Decoder      │ │ - HandshakeFunc │ │     - RPC           │
│  - GOBDecoder   │ │ - NOPHandshake  │ │  - From: net.Addr  │
│  - DefaultDec   │ └─────────────────┘ │  - Payload: []byte │
└─────────────────┘                     └─────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                        store.go                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Store                                                │  │
│  │  - PathTransformFunc: CASPathTransformFunc (SHA1)   │  │
│  │                                                      │  │
│  │  Methods:                                            │  │
│  │  - Read(key) → io.Reader                            │  │
│  │  - writeStream(key, io.Reader) → error              │  │
│  │  - Delete(key) → error                              │  │
│  │  - Exists(key) → bool                               │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  Path Transformation Example:                               │
│  "momsbestpicture" → SHA1 → "6804429f74181a63..."          │
│  → Split(5) → "68044/29f74/181a6/3c50c/3d81d/733a1/..."   │
└─────────────────────────────────────────────────────────────┘
```

---

## Usage Example

```go
package main

import (
    "fmt"
    "log"

    "github.com/imanimen/cas/p2p"
)

func OnPeer(peer p2p.Peer) error {
    fmt.Printf("New peer connected: %v\n", peer)
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

    // Consume messages
    go func() {
        for {
            message := <-tr.Consume()
            fmt.Printf("Message from %v: %v\n", message.From, message.Payload)
        }
    }()

    select {} // Block forever
}
```
