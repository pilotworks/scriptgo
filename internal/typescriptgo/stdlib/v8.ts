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
    constructor() {}

    writeHeader(): void {}

    writeValue(value: unknown): boolean {
        return true;
    }

    releaseBuffer(): Uint8Array {
        return new Uint8Array(0);
    }

    transferArrayBuffer(id: number, arrayBuffer: unknown): void {}

    writeUint32(value: number): void {}

    writeUint64(hi: number, lo: number): void {}

    writeDouble(value: number): void {}

    writeRawBytes(buffer: Uint8Array): void {}

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
    constructor(buffer?: unknown) {}

    readHeader(): boolean {
        return true;
    }

    readValue(): unknown {
        return undefined;
    }

    transferArrayBuffer(id: number, arrayBuffer: unknown): void {}

    getWireFormatVersion(): number {
        return 13;
    }

    readUint32(): number {
        return 0;
    }

    readUint64(): number[] {
        return [0, 0];
    }

    readDouble(): number {
        return 0.0;
    }

    readRawBytes(length: number): Uint8Array {
        return new Uint8Array(length);
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
