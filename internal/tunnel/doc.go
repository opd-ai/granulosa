// Package tunnel wires a WireGuard-Go userspace device to the Tox transport
// layer via ToxBind, a conn.Bind implementation backed by transport.Transport.
//
// ToxBind never references concrete UDP socket types; all networking is
// expressed through the net.Addr / transport.Transport interfaces in keeping
// with the toxcore convention.
package tunnel
