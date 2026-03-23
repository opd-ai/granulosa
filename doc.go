// Package granulosa implements a decentralised WireGuard VPN mesh that tunnels
// WireGuard traffic over the Tox protocol.
//
// Peer discovery, NAT traversal, and anonymity-network routing are provided by
// [github.com/opd-ai/toxcore]. The WireGuard userspace device is provided by
// [golang.zx2c4.com/wireguard] and runs entirely in userspace — no kernel
// module or elevated privileges are required.
//
// # Quick start
//
//	cfg := granulosa.Config{
//	    PrivateKey:       myKey,
//	    BootstrapPeers:   bootstrapAddrs,
//	}
//	mesh, err := granulosa.New(cfg)
//	if err != nil { log.Fatal(err) }
//	defer mesh.Close()
//
//	ln, err := mesh.Listen("tcp", ":8080")
//	// ln is a standard net.Listener backed by the mesh virtual interface.
//
// See ARCHITECTURE.md for the full technical design.
package granulosa
