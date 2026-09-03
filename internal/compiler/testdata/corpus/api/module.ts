import {
    builtinModules,
    isBuiltin
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
