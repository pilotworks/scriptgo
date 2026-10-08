package llvm

import "strings"

// runtimeDeclarations lists the native runtime ABI declaration blocks in
// the order they appear in every emitted module.
var runtimeDeclarations = []string{
	declarationsConsole,
	declarationsFail,
	declarationsLLVMIntrinsics,
	declarationsNumber,
	declarationsArray,
	declarationsObject,
	declarationsJson,
	declarationsString,
	declarationsRegex,
	declarationsSymbol,
	declarationsFs,
	declarationsProcess,
	declarationsChild,
	declarationsDns,
	declarationsNet,
	declarationsDgram,
	declarationsTls,
	declarationsCrypto,
	declarationsZlib,
	declarationsDate,
	declarationsWeb,
	declarationsSqlite,
	declarationsException,
	declarationsClosure,
	declarationsArraybuffer,
	declarationsBuffer,
	declarationsDataview,
	declarationsMap,
	declarationsSet,
	declarationsText,
	declarationsConsole2,
	declarationsJson2,
	declarationsConsole3,
	declarationsWebsocket,
	declarationsIntl,
	declarationsOs,
	declarationsProcess2,
	declarationsTty,
	declarationsClosure2,
	declarationsObject2,
}

// writeRuntimeDeclarations emits the runtime ABI declarations shared by all modules.
func writeRuntimeDeclarations(out *strings.Builder) {
	for _, block := range runtimeDeclarations {
		out.WriteString(block)
	}
}
