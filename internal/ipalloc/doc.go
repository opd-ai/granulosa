// Package ipalloc assigns each mesh peer a deterministic IPv6 ULA address
// derived from its 32-byte Tox public key:
//
//	meshIP = fc00:: | SHA-256(tox_public_key)[0:14]
//
// The address falls inside the fc00::/7 Unique Local Address range.
// Collision probability for a 10 000-node mesh is below 10⁻²⁴; no
// coordination server is required.
package ipalloc
