// Node.js WASI module (node:wasi)

export class WASI {
    wasiImport: unknown = {};

    constructor(options: unknown = null) {
        this.wasiImport = {};
    }

    getImportObject(): unknown {
        return {
            wasi_snapshot_preview1: this.wasiImport,
        };
    }

    start(instance: unknown): void {
    }

    initialize(instance: unknown): void {
    }
}

export default {
    WASI,
};
