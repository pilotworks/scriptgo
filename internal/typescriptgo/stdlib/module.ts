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
    "https",
    "net",
    "os",
    "path",
    "process",
    "punycode",
    "querystring",
    "readline",
    "repl",
    "stream",
    "string_decoder",
    "timers",
    "tls",
    "tty",
    "url",
    "util",
    "v8",
    "vm",
    "wasi",
    "worker_threads",
    "zlib"
];

export const compileCacheStatus = {
    ENABLED: 0,
    ALREADY_ENABLED: 1,
    FAILED: 2,
    DISABLED: 3,
};

export const constants = {
    compileCacheStatus,
};

export function isBuiltin(moduleName: string): boolean {
    const clean = moduleName.startsWith("node:") ? moduleName.slice(5) : moduleName;
    return builtinModules.indexOf(clean) !== -1;
}

export function createRequire(filename: string): (id: string) => unknown {
    return (id: string) => {
        return {};
    };
}

export function enableCompileCache(cacheDir?: string): { status: number; directory?: string; message?: string } {
    return {
        status: compileCacheStatus.ENABLED,
        directory: cacheDir || "/tmp/node-compile-cache",
    };
}

export function getCompileCacheDir(): string | undefined {
    return "/tmp/node-compile-cache";
}

export function flushCompileCache(): void {
}

export function findPackageJSON(specifier: string, base?: string): string | undefined {
    return undefined;
}

export function register(specifier: string, parentURL?: string, options?: unknown): void {
}

export function registerHooks(options: unknown): void {
}

export function stripTypeScriptTypes(code: string, options?: unknown): string {
    return code;
}

export function syncBuiltinESMExports(): void {
}

export function getSourceMapsSupport(): boolean {
    return false;
}

export function setSourceMapsSupport(enabled: boolean): void {
}

export interface SourceMapEntry {
    generatedLine: number;
    generatedColumn: number;
    originalSource: string;
    originalLine: number;
    originalColumn: number;
    name: string;
}

export interface SourceMapOrigin {
    name: string;
    source: string;
    line: number;
    column: number;
}

export class SourceMap {
    payload: Record<string, unknown> = {};

    constructor(payload: unknown = null) {
        this.payload = {};
    }

    findEntry(line: number, column: number): SourceMapEntry {
        return {
            generatedLine: line,
            generatedColumn: column,
            originalSource: "",
            originalLine: line,
            originalColumn: column,
            name: "",
        };
    }

    findOrigin(line: number, column: number): SourceMapOrigin {
        return {
            name: "",
            source: "",
            line: line,
            column: column,
        };
    }
}

export function findSourceMap(path: string, error?: unknown): SourceMap | undefined {
    return undefined;
}

export class Module {
    id: string = "";
    filename: string = "";
    loaded: boolean = true;
    parent: unknown = null;
    children: unknown[] = [];
    exports: unknown = {};
    paths: string[] = [];
    path: string = "";
    isPreloading: boolean = false;

    constructor(id: string = "", parent: unknown = null) {
        this.id = id;
        this.filename = id;
        this.loaded = true;
        this.parent = parent;
        this.children = [];
        this.exports = {};
        this.paths = [];
        this.path = "";
        this.isPreloading = false;
    }

    require(id: string): unknown {
        return {};
    }
}

export default {
    builtinModules,
    compileCacheStatus,
    constants,
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
    findSourceMap,
    Module,
};
