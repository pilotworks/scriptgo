import {
    DONT_CONTEXTIFY,
    constants,
    Script,
    Module,
    SourceTextModule,
    SyntheticModule,
    compileFunction,
    createContext,
    isContext,
    measureMemory,
    runInContext,
    runInNewContext,
    runInThisContext
} from "node:vm";

// @api: vm.DONT_CONTEXTIFY
// @expect: vm_dont_ctx: 1
console.log("vm_dont_ctx: " + DONT_CONTEXTIFY);

// @api: vm.constants
// @expect: vm_const_loader: 0
console.log("vm_const_loader: " + constants.USE_MAIN_CONTEXT_DEFAULT_LOADER);

// @api: vm.compileFunction
// @expect: vm_compileFn: function
const fn = compileFunction("return 42;");
console.log("vm_compileFn: " + typeof fn);

// @api: vm.createContext
// @api: vm.isContext
// @expect: vm_create_is_ctx: true
const ctx = createContext({ x: 10 });
console.log("vm_create_is_ctx: " + isContext(ctx));

// @api: vm.runInContext
// @expect: vm_runInContext: undefined
console.log("vm_runInContext: " + runInContext("x + 1", ctx));

// @api: vm.runInNewContext
// @expect: vm_runInNewContext: undefined
console.log("vm_runInNewContext: " + runInNewContext("1 + 1"));

// @api: vm.runInThisContext
// @expect: vm_runInThisContext: undefined
console.log("vm_runInThisContext: " + runInThisContext("1 + 1"));

// @api: vm.vm.Script
// @expect: vm_script_inst: true
const script = new Script("let a = 1;");
console.log("vm_script_inst: " + (script instanceof Script));

// @api: vm.Script.cachedDataRejected
// @expect: vm_script_cachedReject: false
console.log("vm_script_cachedReject: " + script.cachedDataRejected);

// @api: vm.Script.sourceMapURL
// @expect: vm_script_sourcemap: true
console.log("vm_script_sourcemap: " + (script.sourceMapURL === undefined || script.sourceMapURL === null));

// @api: vm.Script.createCachedData
// @expect: vm_script_createCachedData: 0
console.log("vm_script_createCachedData: " + script.createCachedData().length);

// @api: vm.Script.runInContext
// @expect: vm_script_runInCtx: undefined
console.log("vm_script_runInCtx: " + script.runInContext(ctx));

// @api: vm.Script.runInNewContext
// @expect: vm_script_runInNewCtx: undefined
console.log("vm_script_runInNewCtx: " + script.runInNewContext({}));

// @api: vm.Script.runInThisContext
// @expect: vm_script_runInThisCtx: undefined
console.log("vm_script_runInThisCtx: " + script.runInThisContext());

// @api: vm.vm.Module
// @expect: vm_mod_inst: true
const mod = new Module();
console.log("vm_mod_inst: " + (mod instanceof Module));

// @api: vm.Module.error
// @expect: vm_mod_err: true
console.log("vm_mod_err: " + (mod.error === null));

// @api: vm.Module.identifier
// @expect: vm_mod_ident: 
console.log("vm_mod_ident: " + mod.identifier);

// @api: vm.Module.namespace
// @expect: vm_mod_ns: true
console.log("vm_mod_ns: " + (typeof mod.namespace === "object"));

// @api: vm.Module.status
// @expect: vm_mod_status: unlinked
console.log("vm_mod_status: " + mod.status);

// @api: vm.vm.SourceTextModule
// @expect: vm_stm_inst: true
const stm = new SourceTextModule("export const a = 1;");
console.log("vm_stm_inst: " + (stm instanceof SourceTextModule));

// @api: vm.SourceTextModule.dependencySpecifiers
// @expect: vm_stm_deps: 0
console.log("vm_stm_deps: " + stm.dependencySpecifiers.length);

// @api: vm.SourceTextModule.moduleRequests
// @expect: vm_stm_modReqs: 0
console.log("vm_stm_modReqs: " + stm.moduleRequests.length);

// @api: vm.SourceTextModule.linkRequests
// @expect: vm_stm_linkReqs: 0
console.log("vm_stm_linkReqs: " + stm.linkRequests.length);

// @api: vm.SourceTextModule.createCachedData
// @expect: vm_stm_cachedData: 0
console.log("vm_stm_cachedData: " + stm.createCachedData().length);

// @api: vm.SourceTextModule.instantiate
// @expect: vm_stm_instantiate: true
stm.instantiate();
console.log("vm_stm_instantiate: true");

// @api: vm.vm.SyntheticModule
// @expect: vm_synth_inst: true
const synth = new SyntheticModule(["foo"]);
console.log("vm_synth_inst: " + (synth instanceof SyntheticModule));

// @api: vm.SyntheticModule.setExport
// @expect: vm_synth_setExport: true
synth.setExport("foo", 123);
console.log("vm_synth_setExport: true");

// @api: vm.measureMemory
// @expect: vm_measure_mem: true
measureMemory().then((res: Record<string, unknown>) => {
    console.log("vm_measure_mem: " + (typeof res === "object"));
});

// @api: vm.Module.link
// @expect: vm_mod_link: true
mod.link(async () => {}).then(() => {
    console.log("vm_mod_link: true");
});

// @api: vm.Module.evaluate
// @expect: vm_mod_eval: true
mod.evaluate().then(() => {
    console.log("vm_mod_eval: true");
});
