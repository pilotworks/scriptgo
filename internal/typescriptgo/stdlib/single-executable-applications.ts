// Node.js Single Executable Applications module (node:sea / node:single-executable-applications)

let _seaAssets: Map<string, string> = new Map<string, string>();

export function isSea(): boolean {
    return false;
}

export function getAsset(key: string, encoding: string | null = null): ArrayBuffer | string {
    const val = _seaAssets.get(key);
    if (val !== undefined) {
        if (encoding !== null) {
            return val;
        }
        const buf = new Uint8Array(val.length);
        for (let i = 0; i < val.length; i++) {
            buf[i] = val.charCodeAt(i);
        }
        return buf.buffer as ArrayBuffer;
    }
    if (encoding !== null) {
        return "";
    }
    return new ArrayBuffer(0);
}

export function getAssetAsBlob(key: string, options: unknown = null): unknown {
    return null;
}

export function getRawAsset(key: string): ArrayBuffer {
    return new ArrayBuffer(0);
}

export function getAssetKeys(): string[] {
    return [];
}

export default {
    isSea,
    getAsset,
    getAssetAsBlob,
    getRawAsset,
    getAssetKeys,
};
