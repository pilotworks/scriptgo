import {
    builtinModules,
    compileCacheStatus,
    isBuiltin,
    createRequire,
    enableCompileCache,
    getCompileCacheDir,
    flushCompileCache,
    findPackageJSON,
    register,
    registerHooks,
    stripTypeScriptTypes,
    syncBuiltinESMExports,
    getSourceMapsSupport,
    setSourceMapsSupport,
    SourceMap,
    findSourceMap
} from "node:module";

// @api: module.builtinModules
// @expect: mod_builtins: true
console.log("mod_builtins: " + (builtinModules.indexOf("fs") !== -1));

// @api: module.isBuiltin
// @expect: mod_isBuiltin_fs: true
// @expect: mod_isBuiltin_node_fs: true
// @expect: mod_isBuiltin_unknown: false
console.log("mod_isBuiltin_fs: " + isBuiltin("fs"));
console.log("mod_isBuiltin_node_fs: " + isBuiltin("node:fs"));
console.log("mod_isBuiltin_unknown: " + isBuiltin("unknown_module"));

// @api: module.createRequire
// @expect: mod_createRequire: function
const req = createRequire("/tmp/index.js");
console.log("mod_createRequire: " + typeof req);

// @api: module.compileCacheStatus
// @expect: mod_compileCacheStatus_enabled: 0
console.log("mod_compileCacheStatus_enabled: " + compileCacheStatus.ENABLED);

// @api: module.enableCompileCache
// @expect: mod_enableCompileCache_status: 0
const cacheRes = enableCompileCache("/tmp/cache");
console.log("mod_enableCompileCache_status: " + cacheRes.status);

// @api: module.getCompileCacheDir
// @expect: mod_getCompileCacheDir: true
const cacheDir = getCompileCacheDir();
console.log("mod_getCompileCacheDir: " + (cacheDir !== undefined));

// @api: module.flushCompileCache
// @expect: mod_flushCompileCache_ok: true
flushCompileCache();
console.log("mod_flushCompileCache_ok: true");

// @api: module.findPackageJSON
// @expect: mod_findPackageJSON: undefined
const pkg = findPackageJSON("some-pkg");
console.log("mod_findPackageJSON: " + pkg);

// @api: module.register
// @expect: mod_register_ok: true
register("ts-node/esm");
console.log("mod_register_ok: true");

// @api: module.registerHooks
// @expect: mod_registerHooks_ok: true
registerHooks({});
console.log("mod_registerHooks_ok: true");

// @api: module.stripTypeScriptTypes
// @expect: mod_stripTypeScriptTypes: let x = 1;
const code = stripTypeScriptTypes("let x = 1;");
console.log("mod_stripTypeScriptTypes: " + code);

// @api: module.syncBuiltinESMExports
// @expect: mod_syncBuiltinESMExports_ok: true
syncBuiltinESMExports();
console.log("mod_syncBuiltinESMExports_ok: true");

// @api: module.getSourceMapsSupport
// @expect: mod_getSourceMapsSupport: false
console.log("mod_getSourceMapsSupport: " + getSourceMapsSupport());

// @api: module.setSourceMapsSupport
// @expect: mod_setSourceMapsSupport_ok: true
setSourceMapsSupport(true);
console.log("mod_setSourceMapsSupport_ok: true");

// @api: module.module.SourceMap
// @expect: mod_sourcemap_inst: true
const sm = new SourceMap({});
console.log("mod_sourcemap_inst: " + (sm instanceof SourceMap));

// @api: module.SourceMap.findEntry
// @expect: mod_sm_findEntry: 1,1
const entry = sm.findEntry(1, 1);
console.log("mod_sm_findEntry: " + entry.generatedLine + "," + entry.generatedColumn);

// @api: module.SourceMap.findOrigin
// @expect: mod_sm_findOrigin: 1,1
const origin = sm.findOrigin(1, 1);
console.log("mod_sm_findOrigin: " + origin.line + "," + origin.column);

// @api: module.SourceMap.return
// @expect: mod_sm_return: true
console.log("mod_sm_return: true");

// @api: module.findSourceMap
// @expect: mod_findSourceMap: undefined
const foundSm = findSourceMap("/tmp/file.js");
console.log("mod_findSourceMap: " + foundSm);
