# Architecture Overview

`granulosa` implements a fully decentralised WireGuard VPN mesh by using the Tox
protocol stack as its packet-carrier layer. The core technique is to satisfy
WireGuard-Go's pluggable socket abstraction (`conn.Bind`) with a custom
`ToxBind` implementation that routes encrypted WireGuard datagrams through
`transport.Transport`, inheriting Tox's NAT traversal, anonymity-network
support, and DHT-based peer discovery at no extra cost.

## WireGuard ↔ Tox Integration

The WireGuard device is created via
`tun.CreateTUNFromNetstack()` (`golang.zx2c4.com/wireguard/tun/netstack`),
producing a userspace TUN backed by gVisor's netstack — no kernel module or
root privileges required. `ToxBind` implements `conn.Bind` using
`transport.Transport.Send()` for egress and
`transport.Transport.RegisterHandler()` for ingress demultiplexing.
Outbound WireGuard UDP datagrams are wrapped in Tox `transport.Packet` values
and delivered via `transport.Transport`; inbound Tox packets carrying
WireGuard payloads are unwrapped and injected into the WireGuard device's
receive queue. `transport.NoiseTransport` (wrapping the underlying carrier
with `noise.IKHandshake`) provides Tox-layer encryption, giving double
encapsulation: WireGuard's own Noise-IK session runs on top.

## Peer Identity Mapping

WireGuard peers are identified by 32-byte Curve25519 public keys; so are Tox
DHT nodes. Granulosa uses a **1:1 mapping**: the WireGuard peer public key is
the Tox node ID. No separate identity layer is needed, and peer authentication
in WireGuard and in DHT routing reference the same key material.
⚠ Keys must remain logically separate: the Tox-layer session key is
HKDF-derived from the shared static key to prevent cross-protocol key reuse
(`golang.org/x/crypto/hkdf`).

## Peer Discovery Flow

1. On startup, the local node calls `dht.RoutingTable.FindClosestNodes(peerKey)`
   using each known WireGuard peer's public key as the DHT target.
2. `dht.Maintainer` keeps k-buckets fresh; `dht.GossipBootstrap` propagates
   new peers across the mesh; `dht.LANDiscovery` provides zero-config
   bootstrap on local networks.
3. Each resolved `transport.NetworkAddress` (IPv4/IPv6/.onion/.i2p) is written
   to the WireGuard device as a peer endpoint update via `wgctrl`.

## NAT Traversal

`transport.NATTraversal.PunchHole()` initiates UDP hole-punching before each
new WireGuard session. When `transport.NATTraversal.DetectNATType()` returns
`NATTypeSymmetric`, traffic falls back through `transport.MultiTransport`
using a Tor or I2P path — no DERP server required.

## Mesh IP Allocation

Each node derives its IPv6 mesh address deterministically:

```
meshIP = fc00:: | SHA-256(tox_public_key)[0:14]
```

This yields a `fc00::/7` (ULA) address. The collision probability for a
10 000-node mesh is below 10⁻²⁴; no coordination server is required.

---

# Package Structure

```
github.com/opd-ai/granulosa
├── go.mod
├── doc.go                  # package granulosa — public Mesh API
├── cmd/
│   └── granulosa/          # CLI: init, join, status, peers
├── internal/
│   ├── tunnel/             # ToxBind (conn.Bind), WireGuard device lifecycle
│   ├── mesh/               # Peer registry, endpoint updates, keep-alive
│   ├── discovery/          # DHT wrapper: dht.RoutingTable + dht.Maintainer
│   ├── ipalloc/            # Deterministic ULA IPv6 allocation (SHA-256)
│   └── relay/              # Multi-hop relay via friend/real.RealPacketDelivery
```

**External dependencies** (all pure Go, no cgo):

| Dependency | Version | Role |
|---|---|---|
| `github.com/opd-ai/toxcore` | latest | DHT, transport, Noise-IK, async messaging |
| `golang.zx2c4.com/wireguard` | v0.0.0 | WireGuard-Go userspace device |
| `golang.zx2c4.com/wireguard/tun/netstack` | same | Userspace TUN via gVisor netstack |
| `golang.org/x/crypto` | latest | HKDF key derivation |
| `golang.org/x/net` | latest | `netip` address types |

**Key interfaces**

- `tunnel.ToxBind` implements `conn.Bind` over `transport.Transport`; it
  never references concrete UDP socket types, preserving toxcore's
  _no-concrete-network-types_ convention.
- `mesh.Mesh` exposes `net.Listener` and `net.Conn` by delegating to the
  netstack virtual network interface, giving applications a standard Go
  networking surface.
- `discovery.DHTPeerDiscovery` wraps `dht.RoutingTable` and emits
  `transport.NetworkAddress` updates to `mesh.Mesh` over a channel.

---

# Implementation Roadmap

## Phase 1 — Core Tunnel (complexity: M)

*Toxcore packages: `transport/`, `noise/`*

- Implement `ToxBind` satisfying `conn.Bind` over `transport.Transport`
- Wire WireGuard-Go device with `tun.CreateTUNFromNetstack()`
- Integrate `transport.NoiseTransport` as Tox-layer session security
- Static peer config: manually specify Tox public keys + bootstrap addresses
- Validate with two nodes exchanging ICMP through the WireGuard tunnel

## Phase 2 — Mesh Discovery (complexity: M)

*Toxcore packages: `dht/`, `transport/`*

- Implement `DHTPeerDiscovery` wrapping `dht.RoutingTable.FindClosestNodes()`
- Bootstrap via `dht.GossipBootstrap` and `dht.LANDiscovery`
- Run `transport.NATTraversal.PunchHole()` before each new WireGuard session
- Push `transport.NetworkAddress` changes to WireGuard via `wgctrl`
- Peers join with only a bootstrap address; all others are discovered automatically

## Phase 3 — Features (complexity: L)

*Toxcore packages: `async/`, `friend/real/`, `dht/`*

- Deterministic IPv6 ULA allocation via `internal/ipalloc` (SHA-256 of key)
- Internal DNS via netstack's resolver for `<hex-pubkey>.granulosa` names
- Multi-hop relay through `friend/real.RealPacketDelivery` for symmetric NAT
- `async` epoch-based key rotation for forward-secret mesh control-plane messages

## Phase 4 — Production Hardening (complexity: M)

*Toxcore packages: `interfaces/`, `factory/`*

- Key rotation without session interruption (hooks into `async` epoch rotation)
- Prometheus metrics: peer count, tunnel throughput, DHT k-bucket health
- Persistent configuration and peer blocklists

---

# Feasibility Assessment

## Showstopper Risks

| Risk | Impact | Mitigation |
|---|---|---|
| Tox transport MTU (~1 400 B) vs WireGuard default MTU (1 420 B) | Packet loss / reassembly failures | Force WireGuard MTU to 1 280 B; add reassembly in `ToxBind` using `transport.PacketExtensions` |
| DHT lookup latency (100–500 ms first contact) | Slow initial peer connection | Cache `dht.RoutingTable` entries; speculative pre-resolve on app startup via `dht.Maintainer` |
| Curve25519 key reuse across WireGuard and Tox layers | Cross-protocol cryptographic weakness | HKDF-derive independent subkeys for each layer; enforce boundary in `tunnel.ToxBind` |
| Symmetric NAT peers (≈15 % of internet hosts) | No direct path available | Fall back to `transport.MultiTransport` Tor/I2P path; relay via `friend/real.RealPacketDelivery` |

## Limitations vs. Tailscale

| Feature | Tailscale | Granulosa |
|---|---|---|
| Relay fallback | DERP servers (centralised, fast) | Tox friend relay (decentralised, higher latency) |
| ACL / firewall policy | Tailnet admin console | Not planned (Phase 4 extension) |
| Managed DNS | MagicDNS | netstack mDNS (Phase 3) |
| Identity / SSO | Corporate IdP | None — key is identity |
| Mobile clients | Official iOS/Android | Not planned |

## Performance

Double encapsulation (WireGuard Noise-IK + Tox `noise.IKHandshake`) adds
~40–80 bytes per packet and two ChaCha20-Poly1305 operations (~2–4 µs on
modern hardware). Latency overhead is dominated by DHT lookup on first contact
(100–500 ms) and Tox overlay routing (~10–50 ms per hop).

## Toxcore Advantages

- `transport.MultiTransport` delivers built-in `.onion`/`.i2p`/`.loki`/`.nym`
  routing — granulosa inherits anonymity-network support without additional code.
- `noise.IKHandshake` (Noise_IK_25519_ChaChaPoly_SHA256) is reused for the
  Tox control plane, reducing implementation surface and audit scope.
- `transport.NATTraversal` covers ≥80 % of NAT topologies without any relay
  server, matching Tailscale's direct-path success rate.
- `async` forward-secure pre-key exchange enables secure mesh control-plane
  signalling with per-epoch key rotation and no PKI.
