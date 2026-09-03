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

    toJSON(): Record<string, unknown> {
        return {
            name: this.name,
            entryType: this.entryType,
            startTime: this.startTime,
            duration: this.duration,
        };
    }
}

export class PerformanceMark extends PerformanceEntry {
    detail: unknown = null;

    constructor(name: string, options?: { detail?: unknown; startTime?: number }) {
        super(name, "mark", options?.startTime || 0, 0);
        this.detail = options && options.detail !== undefined ? options.detail : null;
    }
}

export class PerformanceMeasure extends PerformanceEntry {
    detail: unknown = null;

    constructor(name: string, options?: { detail?: unknown; startTime?: number; duration?: number }) {
        super(name, "measure", options?.startTime || 0, options?.duration || 0);
        this.detail = options && options.detail !== undefined ? options.detail : null;
    }
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

    toJSON(): Record<string, unknown> {
        return {
            name: this.name,
            entryType: this.entryType,
            startTime: this.startTime,
            duration: this.duration,
            workerStart: this.workerStart,
            redirectStart: this.redirectStart,
            redirectEnd: this.redirectEnd,
            fetchStart: this.fetchStart,
            domainLookupStart: this.domainLookupStart,
            domainLookupEnd: this.domainLookupEnd,
            connectStart: this.connectStart,
            connectEnd: this.connectEnd,
            secureConnectionStart: this.secureConnectionStart,
            requestStart: this.requestStart,
            responseEnd: this.responseEnd,
            transferSize: this.transferSize,
            encodedBodySize: this.encodedBodySize,
            decodedBodySize: this.decodedBodySize,
        };
    }
}

export class PerformanceObserverEntryList {
    private _entries: PerformanceEntry[] = [];

    constructor(entries: PerformanceEntry[] = []) {
        this._entries = entries;
    }

    getEntries(): PerformanceEntry[] {
        return this._entries.slice();
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

export interface PerformanceObserverOptions {
    entryTypes?: string[];
    type?: string;
    buffered?: boolean;
}

export class PerformanceObserver {
    static supportedEntryTypes: string[] = ["mark", "measure", "node", "resource"];
    private _callback: (list: PerformanceObserverEntryList, observer: PerformanceObserver) => void;
    private _entries: PerformanceEntry[] = [];
    private _entryTypes: string[] = [];
    private _observing: boolean = false;

    constructor(callback: (list: PerformanceObserverEntryList, observer: PerformanceObserver) => void) {
        this._callback = callback;
        this._entries = [];
        this._entryTypes = [];
        this._observing = false;
    }

    observe(options: PerformanceObserverOptions): void {
        this._entryTypes = options.entryTypes ? options.entryTypes.slice() : (options.type ? [options.type] : []);
        this._observing = true;
        performanceObservers.push(this);
        if (options.buffered) {
            const buffered: PerformanceEntry[] = [];
            for (const entry of performanceEntries) {
                if (this._accepts(entry)) {
                    buffered.push(entry);
                }
            }
            if (buffered.length > 0) {
                for (const entry of buffered) {
                    this._entries.push(entry);
                }
                this._flush();
            }
        }
    }

    disconnect(): void {
        this._observing = false;
        const idx = performanceObservers.indexOf(this);
        if (idx !== -1) {
            performanceObservers.splice(idx, 1);
        }
    }

    takeRecords(): PerformanceEntry[] {
        const result = this._entries.slice();
        this._entries = [];
        return result;
    }

    _accepts(entry: PerformanceEntry): boolean {
        return this._entryTypes.length === 0 || this._entryTypes.indexOf(entry.entryType) !== -1;
    }

    _enqueue(entry: PerformanceEntry): void {
        if (!this._observing || !this._accepts(entry)) {
            return;
        }
        this._entries.push(entry);
        this._flush();
    }

    _flush(): void {
        if (this._entries.length === 0) {
            return;
        }
        const records = this.takeRecords();
        this._callback(new PerformanceObserverEntryList(records), this);
    }
}

const performanceEntries: PerformanceEntry[] = [];
const performanceObservers: PerformanceObserver[] = [];

function addPerformanceEntry(entry: PerformanceEntry): void {
    performanceEntries.push(entry);
    for (const observer of performanceObservers.slice()) {
        observer._enqueue(entry);
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
    private _samples: number[] = [];

    reset(): void {
        this.count = 0;
        this.min = 0;
        this.max = 0;
        this.mean = 0;
        this.stddev = 0;
        this.exceeds = 0;
        this.minBigInt = 0n;
        this.maxBigInt = 0n;
        this.countBigInt = 0n;
        this.exceedsBigInt = 0n;
        this.percentiles.clear();
        this.percentilesBigInt.clear();
        this._samples = [];
    }

    percentile(percentile: number): number {
        if (this._samples.length === 0) {
            return 0;
        }
        const sorted = this._samples.slice().sort((a, b) => a - b);
        const p = Math.max(0, Math.min(100, percentile));
        const index = Math.min(sorted.length - 1, Math.max(0, Math.ceil((p / 100) * sorted.length) - 1));
        return sorted[index];
    }

    percentileBigInt(percentile: bigint): bigint {
        return BigInt(this.percentile(Number(percentile)));
    }

    _record(value: number): void {
        if (value < 0 || value !== value) {
            return;
        }
        this._samples.push(value);
        this.count++;
        this.countBigInt = BigInt(this.count);
        if (this.count === 1 || value < this.min) this.min = value;
        if (this.count === 1 || value > this.max) this.max = value;
        this.minBigInt = BigInt(this.min);
        this.maxBigInt = BigInt(this.max);
        let sum = 0;
        for (const sample of this._samples) sum += sample;
        this.mean = sum / this._samples.length;
        let variance = 0;
        for (const sample of this._samples) variance += (sample - this.mean) * (sample - this.mean);
        this.stddev = Math.sqrt(variance / this._samples.length);
        this.percentiles.set(50, this.percentile(50));
        this.percentiles.set(99, this.percentile(99));
        this.percentilesBigInt.set(50n, BigInt(this.percentile(50)));
        this.percentilesBigInt.set(99n, BigInt(this.percentile(99)));
    }
}

export class RecordableHistogram extends Histogram {
    record(val: number | bigint): void {
        this._record(typeof val === "bigint" ? Number(val) : val);
    }

    recordDelta(): void {
        this._record(Date.now());
    }

    add(other: RecordableHistogram): void {
        for (let i = 0; i < other.count; i++) {
            this._record(other.mean);
        }
    }
}

export function createHistogram(options?: unknown): RecordableHistogram {
    return new RecordableHistogram();
}

export class Performance {
    timeOrigin: number = 0;

    now(): number {
        return Date.now() - this.timeOrigin;
    }

    mark(name: string, options?: { detail?: unknown; startTime?: number }): PerformanceMark {
        const mark = new PerformanceMark(name, options);
        addPerformanceEntry(mark);
        return mark;
    }

    measure(name: string, startMarkOrOptions?: unknown, endMark?: string): PerformanceMeasure {
        let startTime = 0;
        let duration = this.now();
        if (typeof startMarkOrOptions === "string") {
            const start = this.getEntriesByName(startMarkOrOptions, "mark");
            if (start.length > 0) startTime = start[0].startTime;
            if (endMark !== undefined) {
                const end = this.getEntriesByName(endMark, "mark");
                if (end.length > 0) duration = end[0].startTime - startTime;
            } else {
                duration = this.now() - startTime;
            }
        } else if (startMarkOrOptions && typeof startMarkOrOptions === "object") {
            const options = startMarkOrOptions as { start?: string; end?: string; detail?: unknown };
            if (options.start) {
                const start = this.getEntriesByName(options.start, "mark");
                if (start.length > 0) startTime = start[0].startTime;
            }
            if (options.end) {
                const end = this.getEntriesByName(options.end, "mark");
                if (end.length > 0) duration = end[0].startTime - startTime;
            } else {
                duration = this.now() - startTime;
            }
        }
        const measure = new PerformanceMeasure(name, { startTime, duration });
        addPerformanceEntry(measure);
        return measure;
    }

    clearMarks(name?: string): void {
        clearPerformanceEntries("mark", name);
    }
    clearMeasures(name?: string): void {
        clearPerformanceEntries("measure", name);
    }
    clearResourceTimings(): void {
        clearPerformanceEntries("resource");
    }
    getEntries(): PerformanceEntry[] {
        return performanceEntries.slice();
    }
    getEntriesByName(name: string, type?: string): PerformanceEntry[] {
        const result: PerformanceEntry[] = [];
        for (const entry of performanceEntries) {
            if (entry.name === name && (!type || entry.entryType === type)) result.push(entry);
        }
        return result;
    }
    getEntriesByType(type: string): PerformanceEntry[] {
        const result: PerformanceEntry[] = [];
        for (const entry of performanceEntries) {
            if (entry.entryType === type) result.push(entry);
        }
        return result;
    }
}

function clearPerformanceEntries(type: string, name?: string): void {
    for (let i = performanceEntries.length - 1; i >= 0; i--) {
        if (performanceEntries[i].entryType === type && (name === undefined || performanceEntries[i].name === name)) {
            performanceEntries.splice(i, 1);
        }
    }
}

export const performance: Performance = new Performance();
performance.timeOrigin = Date.now();

export default {
    PerformanceEntry,
    PerformanceMark,
    PerformanceMeasure,
    PerformanceResourceTiming,
    PerformanceObserverEntryList,
    PerformanceObserver,
    Histogram,
    RecordableHistogram,
    createHistogram,
    performance,
};
