// Node.js SQLite module (node:sqlite)

export const constants = {
    SQLITE_OPEN_READONLY: 1,
    SQLITE_OPEN_READWRITE: 2,
    SQLITE_OPEN_CREATE: 4,
    SQLITE_CHANGESETAPPLY_NOSAVEPOINT: 1,
    SQLITE_CHANGESETAPPLY_INVERT: 2,
};

export async function backup(options?: unknown): Promise<void> {}

export class Session {
    constructor() {}

    changeset(): Uint8Array {
        return new Uint8Array(0);
    }

    patchset(): Uint8Array {
        return new Uint8Array(0);
    }
}

export interface StatementResult {
    changes: number;
    lastInsertRowid: number;
}

export class StatementSync {
    private _sql: string = "";

    constructor(sql: string = "") {
        this._sql = sql;
    }

    all(namedParameters?: unknown, ...anonymousParameters: unknown[]): unknown[] {
        return [];
    }

    columns(): unknown[] {
        return [];
    }

    get(namedParameters?: unknown, ...anonymousParameters: unknown[]): unknown {
        return undefined;
    }

    iterate(namedParameters?: unknown, ...anonymousParameters: unknown[]): unknown[] {
        return [];
    }

    run(namedParameters?: unknown, ...anonymousParameters: unknown[]): StatementResult {
        return { changes: 0, lastInsertRowid: 0 };
    }

    setAllowBareNamedParameters(enabled: boolean): void {}
    setAllowUnknownNamedParameters(enabled: boolean): void {}
    setReturnArrays(enabled: boolean): void {}
    setReadBigInts(enabled: boolean): void {}

    expandedSQL(): string {
        return this._sql;
    }

    sourceSQL(): string {
        return this._sql;
    }
}

export class DatabaseSync {
    isOpen: boolean = true;
    isTransaction: boolean = false;
    location: string = ":memory:";

    constructor(location: string = ":memory:", options?: unknown) {
        this.location = location;
        this.isOpen = true;
        this.isTransaction = false;
    }

    open(): void {
        this.isOpen = true;
    }

    close(): void {
        this.isOpen = false;
    }

    [Symbol.dispose](): void {
        this.close();
    }

    aggregate(name: string, options: unknown): void {}

    loadExtension(path: string): void {}

    enableLoadExtension(enabled: boolean): void {}

    exec(sql: string): void {}

    function(name: string, options?: unknown, fn?: unknown): void {}

    prepare(sql: string): StatementSync {
        return new StatementSync(sql);
    }

    createSession(options?: unknown): Session {
        return new Session();
    }

    applyChangeset(changeset: Uint8Array, options?: unknown): boolean {
        return true;
    }
}

export default {
    constants,
    backup,
    Session,
    StatementSync,
    DatabaseSync,
};
