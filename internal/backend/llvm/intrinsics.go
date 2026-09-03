package llvm

// intrinsicObjectDescriptor returns the runtime field layout for intrinsic
// results whose TypeScript consumers use dynamic property lookup.
func intrinsicObjectDescriptor(callee string) string {
	switch callee {
	case "__dns.lookup":
		return "address:family"
	case "__dns.lookupService":
		return "hostname:service"
	case "__text_encoder.encode_into":
		return "TextEncoderEncodeIntoResult"
	default:
		return ""
	}
}
