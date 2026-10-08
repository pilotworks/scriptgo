package lowering

import (
	"github.com/pilotworks/scriptgo/internal/ir"
)

func registerNetworkIntrinsics(m map[string]BuiltinIntrinsic) {
	// DNS Intrinsics
	registerCallIntrinsic(m, []string{"__scriptgo.dnsLookup"}, CategoryNodeModule, "__dns.lookup", []ir.Type{ir.TypeString, ir.TypeNumber}, ir.TypeObject, 1, 2)
	registerCallIntrinsic(m, []string{"__scriptgo.dnsLookupService"}, CategoryNodeModule, "__dns.lookupService", []ir.Type{ir.TypeString, ir.TypeNumber}, ir.TypeObject, 2, 2)
	registerCallIntrinsic(m, []string{"__scriptgo.dnsReverse"}, CategoryNodeModule, "__dns.reverse", []ir.Type{ir.TypeString}, ir.TypeStringArray, 1, 1)
	registerCallIntrinsic(m, []string{"__scriptgo.dnsResolveStrings"}, CategoryNodeModule, "__dns.resolveStrings", []ir.Type{ir.TypeString, ir.TypeString}, ir.TypeStringArray, 2, 2)

	// Net TCP Intrinsics
	registerCallIntrinsic(m, []string{"__scriptgo.netSocketCreate"}, CategoryNodeModule, "__net.socketCreate", []ir.Type{ir.TypeNumber, ir.TypeNumber}, ir.TypeNumber, 0, 2)
	registerCallIntrinsic(m, []string{"__scriptgo.netSocketConnect"}, CategoryNodeModule, "__net.socketConnect", []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeNumber}, ir.TypeVoid, 3, 3)
	registerCallIntrinsic(m, []string{"__scriptgo.netSocketWrite"}, CategoryNodeModule, "__net.socketWrite", []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeNumber}, ir.TypeNumber, 3, 3)
	registerCallIntrinsic(m, []string{"__scriptgo.netSocketRead"}, CategoryNodeModule, "__net.socketRead", []ir.Type{ir.TypeNumber, ir.TypeNumber}, ir.TypeString, 2, 2)
	registerCallIntrinsic(m, []string{"__scriptgo.netSocketClose"}, CategoryNodeModule, "__net.socketClose", []ir.Type{ir.TypeNumber}, ir.TypeVoid, 1, 1)
	registerCallIntrinsic(m, []string{"__scriptgo.netSocketSetNoDelay"}, CategoryNodeModule, "__net.socketSetNoDelay", []ir.Type{ir.TypeNumber, ir.TypeNumber}, ir.TypeVoid, 2, 2)
	registerCallIntrinsic(m, []string{"__scriptgo.netSocketSetKeepAlive"}, CategoryNodeModule, "__net.socketSetKeepAlive", []ir.Type{ir.TypeNumber, ir.TypeNumber, ir.TypeNumber}, ir.TypeVoid, 3, 3)
	registerCallIntrinsic(m, []string{"__scriptgo.netServerListen"}, CategoryNodeModule, "__net.serverListen", []ir.Type{ir.TypeString, ir.TypeNumber, ir.TypeNumber}, ir.TypeNumber, 3, 3)
	registerCallIntrinsic(m, []string{"__scriptgo.netServerAccept"}, CategoryNodeModule, "__net.serverAccept", []ir.Type{ir.TypeNumber}, ir.TypeObject, 1, 1)

	// Dgram UDP Intrinsics
	registerCallIntrinsic(m, []string{"__scriptgo.dgramSocketCreate"}, CategoryNodeModule, "__dgram.socketCreate", []ir.Type{ir.TypeNumber}, ir.TypeNumber, 0, 1)
	registerCallIntrinsic(m, []string{"__scriptgo.dgramBind"}, CategoryNodeModule, "__dgram.bind", []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeNumber}, ir.TypeVoid, 3, 3)
	registerCallIntrinsic(m, []string{"__scriptgo.dgramSend"}, CategoryNodeModule, "__dgram.send", []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeNumber, ir.TypeNumber, ir.TypeString}, ir.TypeNumber, 5, 5)
	registerCallIntrinsic(m, []string{"__scriptgo.dgramRecv"}, CategoryNodeModule, "__dgram.recv", []ir.Type{ir.TypeNumber, ir.TypeNumber}, ir.TypeObject, 2, 2)
	registerCallIntrinsic(m, []string{"__scriptgo.dgramSetBroadcast"}, CategoryNodeModule, "__dgram.setBroadcast", []ir.Type{ir.TypeNumber, ir.TypeNumber}, ir.TypeVoid, 2, 2)
	registerCallIntrinsic(m, []string{"__scriptgo.dgramSetMulticastTTL"}, CategoryNodeModule, "__dgram.setMulticastTTL", []ir.Type{ir.TypeNumber, ir.TypeNumber}, ir.TypeVoid, 2, 2)
	registerCallIntrinsic(m, []string{"__scriptgo.dgramSetMulticastLoopback"}, CategoryNodeModule, "__dgram.setMulticastLoopback", []ir.Type{ir.TypeNumber, ir.TypeNumber}, ir.TypeVoid, 2, 2)
	registerCallIntrinsic(m, []string{"__scriptgo.dgramSetRecvBufferSize"}, CategoryNodeModule, "__dgram.setRecvBufferSize", []ir.Type{ir.TypeNumber, ir.TypeNumber}, ir.TypeVoid, 2, 2)
	registerCallIntrinsic(m, []string{"__scriptgo.dgramSetSendBufferSize"}, CategoryNodeModule, "__dgram.setSendBufferSize", []ir.Type{ir.TypeNumber, ir.TypeNumber}, ir.TypeVoid, 2, 2)
	registerCallIntrinsic(m, []string{"__scriptgo.dgramSetTTL"}, CategoryNodeModule, "__dgram.setTTL", []ir.Type{ir.TypeNumber, ir.TypeNumber}, ir.TypeVoid, 2, 2)
	registerCallIntrinsic(m, []string{"__scriptgo.dgramSetMulticastInterface"}, CategoryNodeModule, "__dgram.setMulticastInterface", []ir.Type{ir.TypeNumber, ir.TypeString}, ir.TypeVoid, 2, 2)
	registerCallIntrinsic(m, []string{"__scriptgo.dgramAddMembership"}, CategoryNodeModule, "__dgram.addMembership", []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeString}, ir.TypeVoid, 3, 3)
	registerCallIntrinsic(m, []string{"__scriptgo.dgramDropMembership"}, CategoryNodeModule, "__dgram.dropMembership", []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeString}, ir.TypeVoid, 3, 3)
	registerCallIntrinsic(m, []string{"__scriptgo.dgramAddSourceSpecificMembership"}, CategoryNodeModule, "__dgram.addSourceSpecificMembership", []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeString, ir.TypeString}, ir.TypeVoid, 4, 4)
	registerCallIntrinsic(m, []string{"__scriptgo.dgramDropSourceSpecificMembership"}, CategoryNodeModule, "__dgram.dropSourceSpecificMembership", []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeString, ir.TypeString}, ir.TypeVoid, 4, 4)
	registerCallIntrinsic(m, []string{"__scriptgo.dgramConnect"}, CategoryNodeModule, "__dgram.connect", []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeNumber}, ir.TypeVoid, 3, 3)
	registerCallIntrinsic(m, []string{"__scriptgo.dgramDisconnect"}, CategoryNodeModule, "__dgram.disconnect", []ir.Type{ir.TypeNumber}, ir.TypeVoid, 1, 1)
	registerCallIntrinsic(m, []string{"__scriptgo.dgramClose"}, CategoryNodeModule, "__dgram.close", []ir.Type{ir.TypeNumber}, ir.TypeVoid, 1, 1)

	// TTY Intrinsics
	registerCallIntrinsic(m, []string{"tty.isatty", "__scriptgo.ttyIsatty", "isatty"}, CategoryNodeModule, "__tty.isatty", []ir.Type{ir.TypeNumber}, ir.TypeBool, 1, 1)
	registerCallIntrinsic(m, []string{"tty.getWindowSize", "__scriptgo.ttyGetWindowSize"}, CategoryNodeModule, "__tty.getWindowSize", []ir.Type{ir.TypeNumber}, ir.TypeNumberArray, 1, 1)
	registerCallIntrinsic(m, []string{"tty.setRawMode", "__scriptgo.ttySetRawMode"}, CategoryNodeModule, "__tty.setRawMode", []ir.Type{ir.TypeNumber, ir.TypeBool}, ir.TypeBool, 2, 2)
	registerCallIntrinsic(m, []string{"tty.read", "__scriptgo.ttyRead"}, CategoryNodeModule, "__tty.read", []ir.Type{ir.TypeNumber, ir.TypeNumber}, ir.TypeString, 1, 2)
	registerCallIntrinsic(m, []string{"tty.readLine", "__scriptgo.ttyReadLine"}, CategoryNodeModule, "__tty.readLine", []ir.Type{ir.TypeNumber}, ir.TypeString, 1, 1)
	registerCallIntrinsic(m, []string{"tty.write", "__scriptgo.ttyWrite"}, CategoryNodeModule, "__tty.write", []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeNumber}, ir.TypeNumber, 2, 3)

	// TLS intrinsics. The TypeScript node:tls adapter owns option handling and
	// object/event semantics; these calls expose only the native TLS ABI.
	registerCallIntrinsic(m, []string{"__scriptgo.tlsContextCreate"}, CategoryNodeModule, "__tls.contextCreate", []ir.Type{ir.TypeString, ir.TypeBool}, ir.TypeNumber, 7, 7)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsSocketCreate"}, CategoryNodeModule, "__tls.socketCreate", []ir.Type{ir.TypeNumber, ir.TypeBool}, ir.TypeNumber, 2, 2)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsSocketConnect"}, CategoryNodeModule, "__tls.socketConnect", []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeNumber, ir.TypeString, ir.TypeBool, ir.TypeUint8Array}, ir.TypeNumber, 6, 6)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsSocketAdopt"}, CategoryNodeModule, "__tls.socketAdopt", []ir.Type{ir.TypeNumber, ir.TypeNumber, ir.TypeString, ir.TypeBool, ir.TypeBool, ir.TypeBool, ir.TypeUint8Array}, ir.TypeNumber, 7, 7)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsSocketWrite"}, CategoryNodeModule, "__tls.socketWrite", []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeNumber}, ir.TypeNumber, 3, 3)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsSocketWriteBytes"}, CategoryNodeModule, "__tls.socketWriteBytes", []ir.Type{ir.TypeNumber, ir.TypeUint8Array, ir.TypeBuffer}, ir.TypeNumber, 2, 2)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsSocketRead"}, CategoryNodeModule, "__tls.socketRead", []ir.Type{ir.TypeNumber, ir.TypeNumber}, ir.TypeString, 2, 2)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsPairWrite"}, CategoryNodeModule, "__tls.pairWrite", []ir.Type{ir.TypeNumber, ir.TypeString}, ir.TypeNumber, 4, 4)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsPairWriteBytes"}, CategoryNodeModule, "__tls.pairWriteBytes", []ir.Type{ir.TypeNumber, ir.TypeUint8Array, ir.TypeBuffer}, ir.TypeNumber, 3, 3)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsPairRead"}, CategoryNodeModule, "__tls.pairRead", []ir.Type{ir.TypeNumber}, ir.TypeString, 3, 3)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsSocketClose"}, CategoryNodeModule, "__tls.socketClose", []ir.Type{ir.TypeNumber}, ir.TypeVoid, 1, 1)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsSocketInfo"}, CategoryNodeModule, "__tls.socketInfo", []ir.Type{ir.TypeNumber, ir.TypeString}, ir.TypeString, 2, 2)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsSocketNumber"}, CategoryNodeModule, "__tls.socketNumber", []ir.Type{ir.TypeNumber, ir.TypeString}, ir.TypeNumber, 2, 2)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsSocketBool"}, CategoryNodeModule, "__tls.socketBool", []ir.Type{ir.TypeNumber, ir.TypeString}, ir.TypeBool, 2, 2)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsExportKeyingMaterial"}, CategoryNodeModule, "__tls.exportKeyingMaterial", []ir.Type{ir.TypeNumber, ir.TypeString, ir.TypeString, ir.TypeUint8Array}, ir.TypeString, 4, 4)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsSocketSetOption"}, CategoryNodeModule, "__tls.socketSetOption", []ir.Type{ir.TypeNumber, ir.TypeString}, ir.TypeVoid, 3, 3)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsSocketSetServername"}, CategoryNodeModule, "__tls.socketSetServername", []ir.Type{ir.TypeNumber, ir.TypeString}, ir.TypeVoid, 2, 2)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsSocketSetSession"}, CategoryNodeModule, "__tls.socketSetSession", []ir.Type{ir.TypeNumber, ir.TypeUint8Array}, ir.TypeVoid, 2, 2)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsSocketSetKeyCert"}, CategoryNodeModule, "__tls.socketSetKeyCert", []ir.Type{ir.TypeNumber, ir.TypeString}, ir.TypeVoid, 3, 3)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsSocketRenegotiate"}, CategoryNodeModule, "__tls.socketRenegotiate", []ir.Type{ir.TypeNumber}, ir.TypeBool, 1, 1)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsPairCreate"}, CategoryNodeModule, "__tls.pairCreate", []ir.Type{ir.TypeNumber, ir.TypeBool}, ir.TypeNumber, 4, 4)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsServerListen"}, CategoryNodeModule, "__tls.serverListen", []ir.Type{ir.TypeNumber, ir.TypeBool, ir.TypeString}, ir.TypeNumber, 6, 6)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsServerAccept"}, CategoryNodeModule, "__tls.serverAccept", []ir.Type{ir.TypeNumber}, ir.TypeNumber, 1, 1)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsServerClose"}, CategoryNodeModule, "__tls.serverClose", []ir.Type{ir.TypeNumber}, ir.TypeVoid, 1, 1)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsServerInfo"}, CategoryNodeModule, "__tls.serverInfo", []ir.Type{ir.TypeNumber, ir.TypeString}, ir.TypeString, 2, 2)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsServerSetContext"}, CategoryNodeModule, "__tls.serverSetContext", []ir.Type{ir.TypeNumber, ir.TypeBool}, ir.TypeVoid, 4, 4)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsServerAddContext"}, CategoryNodeModule, "__tls.serverAddContext", []ir.Type{ir.TypeNumber, ir.TypeString}, ir.TypeVoid, 3, 3)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsServerSetTicketKeys"}, CategoryNodeModule, "__tls.serverSetTicketKeys", []ir.Type{ir.TypeNumber, ir.TypeString}, ir.TypeVoid, 2, 2)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsX509ParsePem"}, CategoryNodeModule, "__tls.x509ParsePem", []ir.Type{ir.TypeString}, ir.TypeString, 1, 1)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsX509ParseBytes"}, CategoryNodeModule, "__tls.x509ParseBytes", []ir.Type{ir.TypeUint8Array, ir.TypeBuffer}, ir.TypeString, 1, 1)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsCiphers"}, CategoryNodeModule, "__tls.ciphers", nil, ir.TypeString, 0, 0)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsRootCertificates"}, CategoryNodeModule, "__tls.rootCertificates", nil, ir.TypeString, 0, 0)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsSystemCertificates"}, CategoryNodeModule, "__tls.systemCertificates", nil, ir.TypeString, 0, 0)
	registerCallIntrinsic(m, []string{"__scriptgo.tlsExtraCertificates"}, CategoryNodeModule, "__tls.extraCertificates", nil, ir.TypeString, 0, 0)
}
