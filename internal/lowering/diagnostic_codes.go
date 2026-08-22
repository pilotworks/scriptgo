package lowering

// SubsetCode is the stable scriptgo diagnostic namespace for Static lowering.
type SubsetCode string

const (
	// SG1xxx: Static semantic and type boundaries.
	CodeAnyBoundary       SubsetCode = "SG1001"
	CodeUnionNarrowing    SubsetCode = "SG1002"
	CodeGenericSpecialize SubsetCode = "SG1003"
	CodeFunctionValue     SubsetCode = "SG1004"
	CodeStructuralFlow    SubsetCode = "SG1005"
	CodeUnknownBoundary   SubsetCode = "SG1006"

	// SG2xxx: Static lowering and coverage fences.
	CodeStdlibNotLowered SubsetCode = "SG2001"
	CodeTupleNotLowered  SubsetCode = "SG2002"
	CodeDateNotLowered   SubsetCode = "SG2003"
	CodeMapSetLimit      SubsetCode = "SG2004"
	CodeLanguageLowering SubsetCode = "SG2005"

	// SG3xxx: Target capabilities.
	CodeTargetNetworking SubsetCode = "SG3001"
	CodeProcessSpawn     SubsetCode = "SG3002"
	CodeSignals          SubsetCode = "SG3003"
	CodeNativeFFI        SubsetCode = "SG3004"
	CodePlatformAPI      SubsetCode = "SG3005"
	CodeUnsafeDoubleCast SubsetCode = "SG3006"

	// SG4xxx: Semantic divergences and runtime safety outcomes.
	CodeDenseArrayTrap  SubsetCode = "SG4001"
	CodeCheckedCast     SubsetCode = "SG4002"
	CodeWidthCopy       SubsetCode = "SG4003"
	CodeRuntimeHardTrap SubsetCode = "SG4004"
	CodeWarnCheckedCast SubsetCode = "SG4005"

	// SG9xxx: Internal compiler and fallback diagnostics.
	CodeInternalFallback SubsetCode = "SG9001"
)
