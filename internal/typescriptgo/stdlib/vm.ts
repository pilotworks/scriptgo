// Node.js VM module (node:vm)

export const DONT_CONTEXTIFY = 1;

export const constants = {
    USE_MAIN_CONTEXT_DEFAULT_LOADER: 0,
};

export class Script {
    cachedDataRejected: boolean = false;
    sourceMapURL: string | undefined = undefined;
    private _code: string = "";

    constructor(code: string, options?: unknown) {
        this._code = code;
        this.cachedDataRejected = false;
        this.sourceMapURL = undefined;
    }

    createCachedData(): Uint8Array {
        return new Uint8Array(0);
    }

    runInContext(contextifiedObject: unknown, options?: unknown): unknown {
        return undefined;
    }

    runInNewContext(contextObject?: unknown, options?: unknown): unknown {
        return undefined;
    }

    runInThisContext(options?: unknown): unknown {
        return undefined;
    }
}

export class Module {
    error: unknown = null;
    identifier: string = "";
    namespace: Record<string, unknown> = {};
    status: string = "unlinked";

    constructor() {
        this.error = null;
        this.identifier = "";
        this.namespace = {};
        this.status = "unlinked";
    }

    async evaluate(options?: unknown): Promise<unknown> {
        this.status = "evaluated";
        return undefined;
    }

    async link(linker: unknown): Promise<void> {
        this.status = "linked";
    }
}

export class SourceTextModule extends Module {
    dependencySpecifiers: string[] = [];
    moduleRequests: unknown[] = [];
    linkRequests: unknown[] = [];

    constructor(code: string, options?: unknown) {
        super();
        this.dependencySpecifiers = [];
        this.moduleRequests = [];
        this.linkRequests = [];
    }

    createCachedData(): Uint8Array {
        return new Uint8Array(0);
    }

    instantiate(): void {
        this.status = "instantiated";
    }
}

export class SyntheticModule extends Module {
    constructor(exportNames: string[] = [], evaluateCallback?: unknown, options?: unknown) {
        super();
    }

    setExport(name: string, value: unknown): void {}
}

export function compileFunction(code: string, params?: string[], options?: unknown): Function {
    return () => {};
}

export function createContext(contextObject: unknown = {}, options?: unknown): Record<string, unknown> {
    if (contextObject && typeof contextObject === "object") {
        return contextObject as Record<string, unknown>;
    }
    return {};
}

export function isContext(object: unknown): boolean {
    return object !== null && typeof object === "object";
}

export async function measureMemory(options?: unknown): Promise<Record<string, unknown>> {
    return {
        total: { jsMemoryEstimate: 0, jsMemoryRange: [0, 0] },
    };
}

export function runInContext(code: string, contextifiedObject: unknown, options?: unknown): unknown {
    return undefined;
}

export function runInNewContext(code: string, contextObject?: unknown, options?: unknown): unknown {
    return undefined;
}

export function runInThisContext(code: string, options?: unknown): unknown {
    return undefined;
}

export default {
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
    runInThisContext,
};
