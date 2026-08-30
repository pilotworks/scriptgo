// Node.js V8 module (node:v8)

export interface HeapStatistics {
    total_heap_size: number;
    total_heap_size_executable: number;
    total_physical_size: number;
    total_available_size: number;
    used_heap_size: number;
    heap_size_limit: number;
    malloced_memory: number;
    peak_malloced_memory: number;
    does_zap_garbage: number;
    number_of_native_contexts: number;
    number_of_detached_contexts: number;
    total_global_handles_size: number;
    used_global_handles_size: number;
    external_memory: number;
}

export interface HeapSpaceStatistics {
    space_name: string;
    space_size: number;
    space_used_size: number;
    space_available_size: number;
    physical_space_size: number;
}

export interface HeapCodeStatistics {
    code_and_metadata_size: number;
    bytecode_and_metadata_size: number;
    external_script_source_size: number;
    cpu_profiler_metadata_size: number;
}

export function cachedDataVersionTag(): number {
    return 1;
}

export function getHeapStatistics(): HeapStatistics {
    return {
        total_heap_size: 1048576,
        total_heap_size_executable: 524288,
        total_physical_size: 1048576,
        total_available_size: 104857600,
        used_heap_size: 524288,
        heap_size_limit: 1073741824,
        malloced_memory: 0,
        peak_malloced_memory: 0,
        does_zap_garbage: 0,
        number_of_native_contexts: 1,
        number_of_detached_contexts: 0,
        total_global_handles_size: 0,
        used_global_handles_size: 0,
        external_memory: 0,
    };
}

export function getHeapSpaceStatistics(): HeapSpaceStatistics[] {
    return [
        {
            space_name: "new_space",
            space_size: 1048576,
            space_used_size: 524288,
            space_available_size: 524288,
            physical_space_size: 1048576,
        }
    ];
}

export function getHeapCodeStatistics(): HeapCodeStatistics {
    return {
        code_and_metadata_size: 0,
        bytecode_and_metadata_size: 0,
        external_script_source_size: 0,
        cpu_profiler_metadata_size: 0,
    };
}

export function getCppHeapStatistics(): Record<string, unknown> {
    return {};
}

export function getHeapSnapshot(options?: unknown): unknown {
    return {};
}

export function writeHeapSnapshot(filename?: string, options?: unknown): string {
    return filename || "/tmp/heapdump.heapsnapshot";
}

export function setFlagsFromString(flags: string): void {}

export function queryObjects(prototype: unknown, options?: unknown): number {
    return 0;
}

export function stopCoverage(): void {}
export function takeCoverage(): void {}

export function setHeapSnapshotNearHeapLimit(limit: number): void {}

export function isStringOneByteRepresentation(str: string): boolean {
    return true;
}

export interface GCProfileResult {
    version: number;
}

export class GCProfiler {
    constructor() {}

    start(): void {}

    stop(): GCProfileResult {
        return { version: 1 };
    }
}

export class Serializer {
    private _encoded: string = "";
    private _raw: number[] = [];
    private _headerWritten: boolean = false;

    constructor() {}

    writeHeader(): void {
        this._headerWritten = true;
    }

    writeValue(value: unknown): boolean {
        this._headerWritten = true;
        if (value === undefined) {
            this._encoded = "u";
            return true;
        }
        const json = JSON.stringify(value);
        this._encoded = "j" + json;
        return true;
    }

    releaseBuffer(): Uint8Array {
        const encoder = new TextEncoder();
        const payload = encoder.encode(this._encoded);
        const lengthText = String(payload.length).padStart(8, "0");
        const header = encoder.encode("SGV8" + lengthText);
        const result = new Uint8Array(header.length + payload.length + this._raw.length);
        result.set(header, 0);
        result.set(payload, header.length);
        result.set(this._raw, header.length + payload.length);
        return result;
    }

    transferArrayBuffer(id: number, arrayBuffer: unknown): void {}

    writeUint32(value: number): void {
        const v = Math.trunc(value);
        this._raw.push(v & 0xff, (v >>> 8) & 0xff, (v >>> 16) & 0xff, (v >>> 24) & 0xff);
    }

    writeUint64(hi: number, lo: number): void {
        this.writeUint32(lo);
        this.writeUint32(hi);
    }

    writeDouble(value: number): void {
        const numbers = new Float64Array([value]);
        const bytes = new Uint8Array(numbers.buffer);
        for (let i = 0; i < bytes.length; i++) {
            this._raw.push(bytes[i]);
        }
    }

    writeRawBytes(buffer: Uint8Array): void {
        for (let i = 0; i < buffer.length; i++) {
            this._raw.push(buffer[i]);
        }
    }

    _writeHostObject(object: unknown): void {}

    _getDataCloneError(message: string): Error {
        return new Error(message);
    }

    _getSharedArrayBufferId(sharedArrayBuffer: unknown): number {
        return 0;
    }

    _setTreatArrayBufferViewsAsHostObjects(flag: boolean): void {}
}

export class Deserializer {
    private _buffer: Uint8Array;
    private _payload: string = "";
    private _cursor: number = 0;
    private _headerRead: boolean = false;

    constructor(buffer: Uint8Array = new Uint8Array(0)) {
        this._buffer = buffer;
    }

    readHeader(): boolean {
        if (this._buffer.length < 12) {
            return false;
        }
        const decoder = new TextDecoder();
        const headerBytes = this._buffer.slice(0, 12);
        const header = decoder.decode(headerBytes);
        if (header.slice(0, 4) !== "SGV8") {
            return false;
        }
        const payloadLength = Number(header.slice(4, 12));
        if (!Number.isFinite(payloadLength) || payloadLength < 0 || 12 + payloadLength > this._buffer.length) {
            return false;
        }
        const payloadBytes = this._buffer.slice(12, 12 + payloadLength);
        this._payload = decoder.decode(payloadBytes);
        this._cursor = 12 + payloadLength;
        this._headerRead = true;
        return true;
    }

    readValue(): unknown {
        if (!this._headerRead && !this.readHeader()) {
            return undefined;
        }
        if (this._payload === "" || this._payload.charAt(0) === "u") {
            return undefined;
        }
        if (this._payload.charAt(0) !== "j") {
            return undefined;
        }
        return JSON.parse(this._payload.slice(1));
    }

    transferArrayBuffer(id: number, arrayBuffer: unknown): void {}

    getWireFormatVersion(): number {
        return 13;
    }

    readUint32(): number {
        if (this._cursor + 4 > this._buffer.length) {
            return 0;
        }
        const b0 = this._buffer[this._cursor];
        const b1 = this._buffer[this._cursor + 1];
        const b2 = this._buffer[this._cursor + 2];
        const b3 = this._buffer[this._cursor + 3];
        this._cursor += 4;
        return (b0 | (b1 << 8) | (b2 << 16) | (b3 << 24)) >>> 0;
    }

    readUint64(): number[] {
        const lo = this.readUint32();
        const hi = this.readUint32();
        return [hi, lo];
    }

    readDouble(): number {
        if (this._cursor + 8 > this._buffer.length) {
            return 0.0;
        }
        const bytes = this._buffer.slice(this._cursor, this._cursor + 8);
        this._cursor += 8;
        const numbers = new Float64Array(bytes.buffer);
        return numbers[0];
    }

    readRawBytes(length: number): Uint8Array {
        const size = Math.max(0, Math.trunc(length));
        const result = new Uint8Array(size);
        for (let i = 0; i < size && this._cursor < this._buffer.length; i++) {
            result[i] = this._buffer[this._cursor++];
        }
        return result;
    }

    _readHostObject(): unknown {
        return undefined;
    }
}

export class DefaultSerializer extends Serializer {}
export class DefaultDeserializer extends Deserializer {}

export function serialize(value: unknown): Uint8Array {
    const s = new Serializer();
    s.writeHeader();
    s.writeValue(value);
    return s.releaseBuffer();
}

export function deserialize(buffer: Uint8Array): unknown {
    const d = new Deserializer(buffer);
    d.readHeader();
    return d.readValue();
}

export function onInit(callback: unknown): void {}
export function onSettled(callback: unknown): void {}
export function onBefore(callback: unknown): void {}
export function onAfter(callback: unknown): void {}

export function createHook(callbacks: unknown): Record<string, unknown> {
    return {
        enable: () => {},
        disable: () => {},
    };
}

export function init(): void {}
export function before(): void {}
export function after(): void {}
export function settled(): void {}

export function addSerializeCallback(callback: unknown): void {}
export function addDeserializeCallback(callback: unknown): void {}
export function setDeserializeMainFunction(readFunc: unknown): void {}

export function isBuildingSnapshot(): boolean {
    return false;
}

export default {
    cachedDataVersionTag,
    getHeapStatistics,
    getHeapSpaceStatistics,
    getHeapCodeStatistics,
    getCppHeapStatistics,
    getHeapSnapshot,
    writeHeapSnapshot,
    setFlagsFromString,
    queryObjects,
    stopCoverage,
    takeCoverage,
    setHeapSnapshotNearHeapLimit,
    isStringOneByteRepresentation,
    GCProfiler,
    Serializer,
    Deserializer,
    DefaultSerializer,
    DefaultDeserializer,
    serialize,
    deserialize,
    onInit,
    onSettled,
    onBefore,
    onAfter,
    createHook,
    init,
    before,
    after,
    settled,
    addSerializeCallback,
    addDeserializeCallback,
    setDeserializeMainFunction,
    isBuildingSnapshot,
};
