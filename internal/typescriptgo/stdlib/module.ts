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
    "fs",
    "http",
    "net",
    "os",
    "path",
    "process",
    "punycode",
    "querystring",
    "stream",
    "string_decoder",
    "timers",
    "tls",
    "url",
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
