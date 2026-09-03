package llvm

// intrinsicObjectDescriptor returns the runtime field layout for intrinsic
// results whose TypeScript consumers use dynamic property lookup.
func intrinsicObjectDescriptor(callee string) string {
	switch callee {
	case "__dns.lookup":
		return "address:family"
	case "__dns.lookupService":
		return "hostname:service"
	case "__dns.resolveMx":
		return "exchanges:priorities"
	case "__dns.resolveSrv":
		return "names:ports:priorities:weights"
	case "__dns.resolveSoa":
		return "nsname:hostmaster:serial:refresh:retry:expire:minttl"
	case "__dns.resolveCaa":
		return "criticals:issues"
	case "__dns.resolveNaptr":
		return "flags:services:regexps:replacements:orders:preferences"
	case "__text_encoder.encode_into":
		return "TextEncoderEncodeIntoResult"
	default:
		return ""
	}
}
