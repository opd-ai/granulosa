# Goal-Achievement Assessment

## Project Context

- **What it claims to do**: Decentralized WireGuard mesh VPN over the Tox protocol, providing Tailscale-style peer-to-peer networking with no coordination server, no DERP relays, and no accounts. Uses Tox Kademlia DHT for peer discovery and toxcore's built-in NAT hole-punching.

- **Target audience**: Developers and privacy-conscious users who want a fully decentralized VPN mesh with anonymity-network support (.onion, .i2p, .loki, .nym endpoints).

- **Architecture**: 
  - `tunnel/` — ToxBind implementing `conn.Bind` over `transport.Transport`
  - `mesh/` — Peer registry and WireGuard device synchronization
  - `discovery/` — DHT wrapper using `dht.RoutingTable` and `dht.Maintainer`
  - `ipalloc/` — Deterministic ULA IPv6 allocation (SHA-256 of public key)
  - `relay/` — Multi-hop fallback via `friend/real.RealPacketDelivery`
  - `cmd/granulosa/` — CLI: init, join, peers, status

- **Existing CI/quality gates**: None. No GitHub Actions, no Makefile, no linter configuration.

## Goal-Achievement Summary

| Stated Goal | Status | Evidence | Gap Description |
|-------------|--------|----------|-----------------|
| Decentralized WireGuard mesh VPN | ❌ Missing | 0 functions, 0 structs across all packages | Only doc.go scaffolding exists; no implementation |
| DHT-only peer discovery | ❌ Missing | `internal/discovery/` contains only doc.go | DHTPeerDiscovery not implemented |
| Pure Go (no cgo) | ⚠️ Partial | go.mod declares Go 1.24.0; no cgo imports | Achievable but untestable without code |
| Standard net.Listener/net.Conn interfaces | ❌ Missing | `internal/mesh/` contains only doc.go | Mesh API not implemented |
| Anonymity-network aware | ❌ Missing | No transport integration code | Depends on toxcore MultiTransport (not integrated) |
| ToxBind satisfying conn.Bind | ❌ Missing | `internal/tunnel/` contains only doc.go | Core binding not implemented |
| Userspace TUN via netstack | ❌ Missing | No wireguard/tun/netstack integration | WireGuard device not created |
| Deterministic IPv6 ULA allocation | ❌ Missing | `internal/ipalloc/` contains only doc.go | SHA-256 allocation not implemented |
| CLI commands (init, join, peers, status) | ❌ Missing | `cmd/granulosa/` contains only doc.go | No main.go, no cobra/cli setup |
| NAT traversal via toxcore | ❌ Missing | No code invoking transport.NATTraversal | Hole-punching not implemented |
| Multi-hop relay fallback | ❌ Missing | `internal/relay/` contains only doc.go | RealPacketDelivery not integrated |

**Overall: 0/11 goals fully achieved**

The README explicitly states "Pre-implementation — architecture and roadmap defined; code scaffolding in place." This assessment confirms that statement: the project has thorough architectural documentation but zero implementation.

## Metrics Summary

| Metric | Value | Assessment |
|--------|-------|------------|
| Total Lines of Code | 0 | No functional code |
| Total Functions | 0 | Documentation only |
| Total Structs | 0 | No types defined |
| Total Interfaces | 0 | No contracts defined |
| Total Packages | 7 | All are scaffold-only |
| Test Coverage | 0% | No tests |
| Documentation Coverage | 100% (packages) | All packages have doc comments |

## Roadmap

### Priority 1: Implement Core Tunnel (Phase 1 per ARCHITECTURE.md)

The ToxBind integration is the foundation—everything else depends on it.

- [ ] **Declare external dependencies** in go.mod:
  - `github.com/opd-ai/toxcore` (DHT, transport, Noise-IK)
  - `golang.zx2c4.com/wireguard` (WireGuard-Go)
  - `golang.zx2c4.com/wireguard/tun/netstack` (userspace TUN)
  - `golang.org/x/crypto` (HKDF key derivation)
  
- [ ] **Implement `internal/tunnel/toxbind.go`**:
  - Define `ToxBind` struct implementing `conn.Bind` interface
  - Use `transport.Transport.Send()` for egress
  - Use `transport.Transport.RegisterHandler()` for ingress demultiplexing
  - Wrap WireGuard datagrams in `transport.Packet`
  - Force MTU to 1280 bytes per ARCHITECTURE.md risk mitigation
  
- [ ] **Implement `internal/tunnel/device.go`**:
  - Create WireGuard device via `tun.CreateTUNFromNetstack()`
  - Configure device with static peer keys for initial testing
  - Wire ToxBind to device's socket layer
  
- [ ] **Implement HKDF key derivation** to prevent cross-protocol key reuse:
  - Derive separate Tox-layer session keys from shared static keys
  - Document key derivation boundary in tunnel/doc.go

- [ ] **Validation**: Two nodes exchange ICMP through the WireGuard tunnel

### Priority 2: Implement Mesh Discovery (Phase 2 per ARCHITECTURE.md)

Without discovery, peers must be statically configured.

- [ ] **Implement `internal/discovery/dht.go`**:
  - Create `DHTPeerDiscovery` struct wrapping `dht.RoutingTable`
  - Implement `FindClosestNodes()` for each WireGuard peer key
  - Emit `transport.NetworkAddress` updates via channel
  
- [ ] **Implement bootstrap mechanisms**:
  - Integrate `dht.GossipBootstrap` for mesh propagation
  - Integrate `dht.LANDiscovery` for zero-config local networks
  
- [ ] **Implement `internal/mesh/mesh.go`**:
  - Track active peers and their `transport.NetworkAddress` endpoints
  - Update WireGuard device peer endpoints via wgctrl
  - Expose `net.Listener` / `net.Conn` via netstack virtual interface
  
- [ ] **Implement NAT traversal**:
  - Call `transport.NATTraversal.PunchHole()` before new sessions
  - Detect NAT type via `transport.NATTraversal.DetectNATType()`
  - Fall back to MultiTransport for symmetric NAT

- [ ] **Validation**: Peers join with only a bootstrap address; others discovered automatically

### Priority 3: Implement IPv6 Allocation and Relay (Phase 3)

These enable production-ready operation.

- [ ] **Implement `internal/ipalloc/alloc.go`**:
  - Compute `fc00:: | SHA-256(tox_public_key)[0:14]` for mesh IP
  - Validate ULA range compliance
  - Return `netip.Addr` for integration with WireGuard config
  
- [ ] **Implement `internal/relay/relay.go`**:
  - Create relay path via `friend/real.RealPacketDelivery`
  - Route WireGuard datagrams through Tox friend connections
  - Enable fallback for symmetric NAT scenarios

- [ ] **Validation**: Mesh works behind symmetric NAT using relay

### Priority 4: Implement CLI (Phase 1-2)

Users need a way to interact with the mesh.

- [ ] **Create `cmd/granulosa/main.go`**:
  - Implement `init` command: generate identity key-pair
  - Implement `join` command: read config, bootstrap, start mesh
  - Implement `peers` command: list connected peers with mesh IPs
  - Implement `status` command: show tunnel statistics
  
- [ ] **Define configuration format**:
  - XDG_CONFIG_HOME/granulosa location
  - YAML or TOML for private key, bootstrap peers, peer allowlist

- [ ] **Validation**: User can initialize, join mesh, and see peers

### Priority 5: Public API and Documentation (Phase 3-4)

The root package should expose a stable public API.

- [ ] **Implement `mesh.go` in root package**:
  - Define `Config` struct per doc.go example
  - Implement `New(Config) (*Mesh, error)` constructor
  - Implement `Mesh.Listen(network, address)` returning `net.Listener`
  - Implement `Mesh.Dial(network, address)` returning `net.Conn`
  - Implement `Mesh.Close()` for graceful shutdown

- [ ] **Add comprehensive examples**:
  - Example server using `mesh.Listen()`
  - Example client using `mesh.Dial()`
  - Example peer discovery callback

- [ ] **Validation**: Applications can use granulosa as a library

### Priority 6: Quality Infrastructure (Phase 4)

Production hardening requires CI and testing.

- [ ] **Add GitHub Actions workflow**:
  - `go build ./...` on push
  - `go test -race ./...` on push
  - `go vet ./...` on push
  - Consider staticcheck or golangci-lint

- [ ] **Add unit tests for core components**:
  - `tunnel/toxbind_test.go` — test Bind interface compliance
  - `ipalloc/alloc_test.go` — test deterministic IP generation
  - `discovery/dht_test.go` — test peer resolution
  
- [ ] **Add integration tests**:
  - Two-node tunnel establishment
  - NAT traversal scenarios
  - Relay fallback scenarios

- [ ] **Validation**: CI passes, test coverage >70% for critical paths

### Priority 7: Production Features (Phase 4)

Advanced features for real-world deployment.

- [ ] **Key rotation without session interruption**:
  - Hook into toxcore's async epoch rotation
  - Implement graceful handoff during WireGuard rekey
  
- [ ] **Observability**:
  - Prometheus metrics: peer_count, tunnel_bytes_total, dht_bucket_health
  - Structured logging with slog
  
- [ ] **Persistent configuration**:
  - Save/load peer state across restarts
  - Peer blocklist support

- [ ] **Validation**: Mesh operates reliably in production scenarios

## Competitive Context

| Capability | Tailscale | Meshguard | Granulosa (Target) |
|------------|-----------|-----------|-------------------|
| Central coordinator | Yes (control server) | No | No (DHT only) |
| NAT traversal | DERP relays | SWIM gossip | Tox hole-punch + MultiTransport |
| Anonymity networks | No | No | Yes (.onion/.i2p/.loki/.nym) |
| Identity | SSO/IdP | Public key | Public key |
| Pure Go | Yes | Yes | Yes (target) |

Granulosa's unique value proposition is anonymity-network support inherited from toxcore—this should be emphasized once the core tunnel works.

## Risk Mitigation Notes (from ARCHITECTURE.md)

1. **MTU mismatch**: Force WireGuard MTU to 1280B; add reassembly in ToxBind
2. **DHT latency**: Cache routing table entries; pre-resolve on startup
3. **Key reuse**: HKDF-derive separate subkeys per layer
4. **Symmetric NAT (~15%)**: Fall back to Tor/I2P via MultiTransport

## Appendix: Metrics Report

```
Files Processed: 7
Total Lines of Code: 0
Total Functions: 0
Total Structs: 0
Total Interfaces: 0
Test Coverage: 0%
Naming Score: 98.6% (1 minor violation: package/directory name mismatch at root)
```

All packages contain only documentation scaffolding. Implementation should follow the priority order above to maximize progress toward stated goals.
