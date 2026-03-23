# granulosa

**Decentralised WireGuard mesh VPN over the Tox protocol.**

`granulosa` tunnels WireGuard traffic through
[toxcore](https://github.com/opd-ai/toxcore), giving you a Tailscale-style
peer-to-peer VPN mesh with no coordination server, no DERP relays, and no
accounts.  Peer discovery uses the Tox Kademlia DHT; NAT traversal reuses
toxcore's built-in hole-punching; anonymity-network paths (Tor, I2P, Lokinet,
Nym) are supported for free through toxcore's `transport.MultiTransport`.

## Status

Pre-implementation — architecture and roadmap defined; code scaffolding in
place.  See [ARCHITECTURE.md](ARCHITECTURE.md) for the full technical design.

## Design goals

* **No central infrastructure** — DHT-only peer discovery via `dht.RoutingTable`
* **Pure Go** — no cgo, no kernel modules required
* **Standard interfaces** — exposes `net.Listener` / `net.Conn` to applications
* **Anonymity-network aware** — `.onion` / `.i2p` / `.loki` / `.nym` endpoints
  work out of the box

## Module

```
github.com/opd-ai/granulosa
```

Requires Go 1.24.0 or later.

## License

MIT — see [LICENSE](LICENSE).
