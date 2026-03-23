// Package relay provides a multi-hop fallback path for peers that cannot
// establish a direct WireGuard session (e.g. behind symmetric NAT).
//
// It uses friend/real.RealPacketDelivery from github.com/opd-ai/toxcore to
// route WireGuard datagrams through intermediate Tox friend connections.
package relay
