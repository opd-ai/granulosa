// Package discovery wraps dht.RoutingTable and dht.Maintainer to provide
// automatic WireGuard peer endpoint resolution over the Tox DHT.
//
// DHTPeerDiscovery calls dht.RoutingTable.FindClosestNodes for each known
// WireGuard peer key and emits transport.NetworkAddress updates to the mesh
// layer over a channel.  Bootstrap is handled by dht.GossipBootstrap and
// dht.LANDiscovery for zero-config local-network operation.
package discovery
