// Node.js Performance Hooks module (node:perf_hooks)

export class PerformanceEntry {
    name: string = "";
    entryType: string = "";
    startTime: number = 0;
    duration: number = 0;

    constructor(name: string = "", entryType: string = "", startTime: number = 0, duration: number = 0) {
        this.name = name;
        this.entryType = entryType;
        this.startTime = startTime;
        this.duration = duration;
    }
}

export class PerformanceMark extends PerformanceEntry {
    detail: unknown = null;

    constructor(name: string, options?: { detail?: unknown; startTime?: number }) {
        super(name, "mark", options?.startTime || 0, 0);
        this.detail = options?.detail || null;
    }
}

export class PerformanceMeasure extends PerformanceEntry {
    detail: unknown = null;

    constructor(name: string, options?: { detail?: unknown; startTime?: number; duration?: number }) {
        super(name, "measure", options?.startTime || 0, options?.duration || 0);
        this.detail = options?.detail || null;
    }
}

export class PerformanceNodeEntry extends PerformanceEntry {
    detail: unknown = null;
    flags: number = 0;
    kind: number = 0;

    constructor(name: string = "") {
        super(name, "node", 0, 0);
        this.detail = null;
        this.flags = 0;
        this.kind = 0;
    }
}

export class PerformanceNodeTiming extends PerformanceEntry {
    bootstrapComplete: number = 0;
    environment: number = 0;
    idleTime: number = 0;
    loopExit: number = 0;
    loopStart: number = 0;
    nodeStart: number = 0;
    v8Start: number = 0;

    constructor() {
        super("node", "node", 0, 0);
        this.bootstrapComplete = 0;
        this.environment = 0;
        this.idleTime = 0;
        this.loopExit = 0;
        this.loopStart = 0;
        this.nodeStart = 0;
        this.v8Start = 0;
    }

    return(): void {}
}

export class PerformanceResourceTiming extends PerformanceEntry {
    workerStart: number = 0;
    redirectStart: number = 0;
    redirectEnd: number = 0;
    fetchStart: number = 0;
    domainLookupStart: number = 0;
    domainLookupEnd: number = 0;
    connectStart: number = 0;
    connectEnd: number = 0;
    secureConnectionStart: number = 0;
    requestStart: number = 0;
    responseEnd: number = 0;
    transferSize: number = 0;
    encodedBodySize: number = 0;
    decodedBodySize: number = 0;

    constructor(name: string = "") {
        super(name, "resource", 0, 0);
        this.workerStart = 0;
        this.redirectStart = 0;
        this.redirectEnd = 0;
        this.fetchStart = 0;
        this.domainLookupStart = 0;
        this.domainLookupEnd = 0;
        this.connectStart = 0;
        this.connectEnd = 0;
        this.secureConnectionStart = 0;
        this.requestStart = 0;
        this.responseEnd = 0;
        this.transferSize = 0;
        this.encodedBodySize = 0;
        this.decodedBodySize = 0;
    }

    toJSON(): unknown {
        return {};
    }
}

export class PerformanceObserverEntryList {
    private _entries: PerformanceEntry[] = [];

    constructor(entries: PerformanceEntry[] = []) {
        this._entries = entries;
    }

    getEntries(): PerformanceEntry[] {
        return this._entries;
    }

    getEntriesByType(type: string): PerformanceEntry[] {
        const res: PerformanceEntry[] = [];
        for (const e of this._entries) {
            if (e.entryType === type) {
                res.push(e);
            }
        }
        return res;
    }

    getEntriesByName(name: string, type?: string): PerformanceEntry[] {
        const res: PerformanceEntry[] = [];
        for (const e of this._entries) {
            if (e.name === name && (!type || e.entryType === type)) {
                res.push(e);
            }
        }
        return res;
    }
}

export class PerformanceObserver {
    static supportedEntryTypes: string[] = ["mark", "measure", "node", "resource"];
    private _callback: (list: PerformanceObserverEntryList, observer: PerformanceObserver) => void;

    constructor(callback: (list: PerformanceObserverEntryList, observer: PerformanceObserver) => void) {
        this._callback = callback;
    }

    observe(options: { entryTypes?: string[]; type?: string; buffered?: boolean }): void {}

    disconnect(): void {}

    takeRecords(): PerformanceEntry[] {
        return [];
    }
}

export class Histogram {
    min: number = 0;
    max: number = 0;
    mean: number = 0;
    stddev: number = 0;
    count: number = 0;
    exceeds: number = 0;
    minBigInt: bigint = 0n;
    maxBigInt: bigint = 0n;
    countBigInt: bigint = 0n;
    exceedsBigInt: bigint = 0n;
    percentiles: Map<number, number> = new Map();
    percentilesBigInt: Map<bigint, bigint> = new Map();

    reset(): void {
        this.count = 0;
        this.min = 0;
        this.max = 0;
        this.mean = 0;
        this.stddev = 0;
        this.exceeds = 0;
    }

    percentile(percentile: number): number {
        return 0;
    }

    percentileBigInt(percentile: bigint): bigint {
        return 0n;
    }
}

export class IntervalHistogram extends Histogram {
    enable(): boolean {
        return true;
    }

    disable(): boolean {
        return true;
    }
}

export class RecordableHistogram extends Histogram {
    record(val: number | bigint): void {
        this.count++;
    }

    recordDelta(): void {
        this.count++;
    }

    add(other: RecordableHistogram): void {
        this.count += other.count;
    }
}

export function createHistogram(options?: unknown): RecordableHistogram {
    return new RecordableHistogram();
}

export function monitorEventLoopDelay(options?: unknown): IntervalHistogram {
    return new IntervalHistogram();
}

export class Performance {
    nodeTiming: PerformanceNodeTiming = new PerformanceNodeTiming();
    timeOrigin: number = 0;

    now(): number {
        return 0;
    }

    mark(name: string, options?: { detail?: unknown; startTime?: number }): PerformanceMark {
        return new PerformanceMark(name, options);
    }

    measure(name: string, startMarkOrOptions?: unknown, endMark?: string): PerformanceMeasure {
        return new PerformanceMeasure(name);
    }

    clearMarks(name?: string): void {}
    clearMeasures(name?: string): void {}
    clearResourceTimings(): void {}
    getEntries(): PerformanceEntry[] {
        return [];
    }
    getEntriesByName(name: string, type?: string): PerformanceEntry[] {
        return [];
    }
    getEntriesByType(type: string): PerformanceEntry[] {
        return [];
    }
}

export const performance: Performance = new Performance();

export default {
    PerformanceEntry,
    PerformanceMark,
    PerformanceMeasure,
    PerformanceNodeEntry,
    PerformanceNodeTiming,
    PerformanceResourceTiming,
    PerformanceObserverEntryList,
    PerformanceObserver,
    Histogram,
    IntervalHistogram,
    RecordableHistogram,
    createHistogram,
    monitorEventLoopDelay,
    performance,
};
