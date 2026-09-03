declare namespace __scriptgo {
    function sharedArrayBufferNew(byteLength: number): SharedArrayBuffer;
    function atomicsIsLockFree(size: number): boolean;
    function atomicsAdd(typedArray: Int32Array | Uint32Array, index: number, value: number): number;
    function atomicsSub(typedArray: Int32Array | Uint32Array, index: number, value: number): number;
    function atomicsAnd(typedArray: Int32Array | Uint32Array, index: number, value: number): number;
    function atomicsOr(typedArray: Int32Array | Uint32Array, index: number, value: number): number;
    function atomicsXor(typedArray: Int32Array | Uint32Array, index: number, value: number): number;
    function atomicsLoad(typedArray: Int32Array | Uint32Array, index: number): number;
    function atomicsStore(typedArray: Int32Array | Uint32Array, index: number, value: number): number;
    function atomicsExchange(typedArray: Int32Array | Uint32Array, index: number, value: number): number;
    function atomicsCompareExchange(typedArray: Int32Array | Uint32Array, index: number, expected: number, replacement: number): number;
    function atomicsWait(typedArray: Int32Array, index: number, value: number, timeout?: number): "ok" | "not-equal" | "timed-out";
    function atomicsNotify(typedArray: Int32Array, index: number, count?: number): number;
}

export class SharedArrayBuffer {
    readonly byteLength: number;

    constructor(byteLength: number) {
        return __scriptgo.sharedArrayBufferNew(byteLength);
    }
}

export const Atomics = {
    isLockFree(size: number): boolean {
        return __scriptgo.atomicsIsLockFree(size);
    },
    add(typedArray: Int32Array | Uint32Array, index: number, value: number): number {
        return __scriptgo.atomicsAdd(typedArray, index, value);
    },
    sub(typedArray: Int32Array | Uint32Array, index: number, value: number): number {
        return __scriptgo.atomicsSub(typedArray, index, value);
    },
    and(typedArray: Int32Array | Uint32Array, index: number, value: number): number {
        return __scriptgo.atomicsAnd(typedArray, index, value);
    },
    or(typedArray: Int32Array | Uint32Array, index: number, value: number): number {
        return __scriptgo.atomicsOr(typedArray, index, value);
    },
    xor(typedArray: Int32Array | Uint32Array, index: number, value: number): number {
        return __scriptgo.atomicsXor(typedArray, index, value);
    },
    load(typedArray: Int32Array | Uint32Array, index: number): number {
        return __scriptgo.atomicsLoad(typedArray, index);
    },
    store(typedArray: Int32Array | Uint32Array, index: number, value: number): number {
        return __scriptgo.atomicsStore(typedArray, index, value);
    },
    exchange(typedArray: Int32Array | Uint32Array, index: number, value: number): number {
        return __scriptgo.atomicsExchange(typedArray, index, value);
    },
    compareExchange(typedArray: Int32Array | Uint32Array, index: number, expected: number, replacement: number): number {
        return __scriptgo.atomicsCompareExchange(typedArray, index, expected, replacement);
    },
    wait(typedArray: Int32Array, index: number, value: number, timeout?: number): "ok" | "not-equal" | "timed-out" {
        return __scriptgo.atomicsWait(typedArray, index, value, timeout);
    },
    notify(typedArray: Int32Array, index: number, count?: number): number {
        return __scriptgo.atomicsNotify(typedArray, index, count);
    }
};

export default Atomics;
