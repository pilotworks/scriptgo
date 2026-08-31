// Node.js VM module (node:vm)
// Static Tier: Pure TypeScript ambient declarations for compile-time checking and tooling.
// Note: Dynamic runtime execution requires the '--dynamic' compilation tier (QuickJS-ng).

export const DONT_CONTEXTIFY: number;

export const constants: {
    readonly USE_MAIN_CONTEXT_DEFAULT_LOADER: number;
    readonly DONT_CONTEXTIFY: number;
};

export interface ScriptOptions {
    filename?: string;
    lineOffset?: number;
    columnOffset?: number;
    cachedData?: Uint8Array | ArrayBufferView;
    produceCachedData?: boolean;
    importModuleDynamically?: ((specifier: string, script: Script, importAssertions?: Record<string, string>) => unknown) | number;
}

export interface RunningScriptOptions {
    displayErrors?: boolean;
    timeout?: number;
    breakOnSigint?: boolean;
}

export interface ContextCodeGenerationOptions {
    strings?: boolean;
    wasm?: boolean;
}

export interface RunningScriptInNewContextOptions extends RunningScriptOptions {
    contextName?: string;
    contextOrigin?: string;
    contextCodeGeneration?: ContextCodeGenerationOptions;
    microtaskMode?: "afterEvaluate";
}

export interface CreateContextOptions {
    name?: string;
    origin?: string;
    codeGeneration?: ContextCodeGenerationOptions;
    microtaskMode?: "afterEvaluate";
    importModuleDynamically?: ((specifier: string, script: Script, importAssertions?: Record<string, string>) => unknown) | number;
}

export interface CompileFunctionOptions {
    filename?: string;
    lineOffset?: number;
    columnOffset?: number;
    cachedData?: Uint8Array | ArrayBufferView;
    produceCachedData?: boolean;
    parsingContext?: Record<string, unknown>;
    contextExtensions?: Record<string, unknown>[];
    importModuleDynamically?: ((specifier: string, script: Script, importAssertions?: Record<string, string>) => unknown) | number;
}

export interface MeasureMemoryOptions {
    mode?: "summary" | "detailed";
    execution?: "default" | "eager";
}

export interface MemoryMeasurementTotal {
    jsMemoryEstimate: number;
    jsMemoryRange: [number, number];
}

export interface MemoryMeasurement {
    total: MemoryMeasurementTotal;
    [key: string]: unknown;
}

export type ModuleStatus = "unlinked" | "linking" | "linked" | "evaluating" | "evaluated" | "errored";

export type ModuleLinker = (
    specifier: string,
    referencingModule: Module,
    extra?: { attributes?: Record<string, string>; assert?: Record<string, string> }
) => Module | Promise<Module>;

export interface SourceTextModuleOptions {
    identifier?: string;
    context?: Record<string, unknown>;
    lineOffset?: number;
    columnOffset?: number;
    initializeImportMeta?: (meta: unknown, module: SourceTextModule) => void;
    importModuleDynamically?: ((specifier: string, module: Module, importAssertions?: Record<string, string>) => unknown) | number;
}

export interface SyntheticModuleOptions {
    identifier?: string;
    context?: Record<string, unknown>;
}

export class Script {
    cachedDataRejected: boolean;
    sourceMapURL: string | undefined;

    constructor(code: string, options?: ScriptOptions | string);

    createCachedData(): Uint8Array;
    runInContext(contextifiedObject: unknown, options?: RunningScriptOptions): unknown;
    runInNewContext(contextObject?: unknown, options?: RunningScriptInNewContextOptions): unknown;
    runInThisContext(options?: RunningScriptOptions): unknown;
}

export class Module {
    error: unknown;
    identifier: string;
    namespace: Record<string, unknown>;
    status: ModuleStatus;

    constructor();

    evaluate(options?: { timeout?: number; breakOnSigint?: boolean }): Promise<unknown>;
    link(linker: ModuleLinker | unknown): Promise<void>;
}

export class SourceTextModule extends Module {
    dependencySpecifiers: string[];
    moduleRequests: unknown[];
    linkRequests: unknown[];

    constructor(code: string, options?: SourceTextModuleOptions);

    createCachedData(): Uint8Array;
    instantiate(): void;
}

export class SyntheticModule extends Module {
    constructor(
        exportNames?: string[],
        evaluateCallback?: (this: SyntheticModule) => void,
        options?: SyntheticModuleOptions
    );

    setExport(name: string, value: unknown): void;
}

export function compileFunction(
    code: string,
    params?: string[],
    options?: CompileFunctionOptions
): Function;

export function createContext(
    contextObject?: unknown,
    options?: CreateContextOptions
): Record<string, unknown>;

export function isContext(object: unknown): boolean;

export function measureMemory(options?: MeasureMemoryOptions): Promise<MemoryMeasurement>;

export function runInContext(
    code: string,
    contextifiedObject: unknown,
    options?: RunningScriptOptions
): unknown;

export function runInNewContext(
    code: string,
    contextObject?: unknown,
    options?: RunningScriptInNewContextOptions
): unknown;

export function runInThisContext(
    code: string,
    options?: RunningScriptOptions
): unknown;

declare const _default: {
    DONT_CONTEXTIFY: typeof DONT_CONTEXTIFY;
    constants: typeof constants;
    Script: typeof Script;
    Module: typeof Module;
    SourceTextModule: typeof SourceTextModule;
    SyntheticModule: typeof SyntheticModule;
    compileFunction: typeof compileFunction;
    createContext: typeof createContext;
    isContext: typeof isContext;
    measureMemory: typeof measureMemory;
    runInContext: typeof runInContext;
    runInNewContext: typeof runInNewContext;
    runInThisContext: typeof runInThisContext;
};
export default _default;
