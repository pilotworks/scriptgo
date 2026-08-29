// Node.js Trace Events module (node:trace_events)

let _enabledCategories: string[] = ["node", "node.async_hooks", "v8"];

export interface TracingOptions {
    categories?: string[];
}

export class Tracing {
    categories: string = "";
    enabled: boolean = false;

    constructor(categories: string[] = []) {
        this.categories = categories.join(",");
        this.enabled = false;
    }

    enable(): void {
        this.enabled = true;
    }

    disable(): void {
        this.enabled = false;
    }
}

export function createTracing(options: TracingOptions = {}): Tracing {
    if (options && options.categories) {
        return new Tracing(options.categories);
    }
    return new Tracing([]);
}

export function getEnabledCategories(): string {
    return _enabledCategories.join(",");
}

export default {
    createTracing,
    getEnabledCategories,
    Tracing,
};
