// Node.js Module system (node:module / node:modules)

export const builtinModules: string[] = [
    "assert",
    "buffer",
    "child_process",
    "console",
    "crypto",
    "dgram",
    "dns",
    "events",
    "formdata",
    "fs",
    "http",
    "https",
    "net",
    "os",
    "path",
    "process",
    "punycode",
    "querystring",
    "readline",
    "readline/promises",
    "stream",
    "string_decoder",
    "test",
    "timers",
    "tls",
    "tty",
    "url",
    "urlpattern",
    "util",
    "zlib"
];

export function isBuiltin(moduleName: string): boolean {
    const clean = moduleName.startsWith("node:") ? moduleName.slice(5) : moduleName;
    return builtinModules.indexOf(clean) !== -1;
}

export default {
    builtinModules,
    isBuiltin,
};
