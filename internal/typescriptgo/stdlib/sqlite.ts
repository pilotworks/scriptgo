// Node.js SQLite module (node:sqlite)

export type SQLInputValue = null | number | bigint | string | Uint8Array;
export type SQLOutputValue = null | number | bigint | string | Uint8Array;

export interface DatabaseSyncOptions {
    open?: boolean;
    enableForeignKeyConstraints?: boolean;
    enableDoubleQuotedStringLiterals?: boolean;
    readOnly?: boolean;
    allowExtension?: boolean;
    timeout?: number;
    readBigInts?: boolean;
    returnArrays?: boolean;
    allowBareNamedParameters?: boolean;
    allowUnknownNamedParameters?: boolean;
}

export interface CreateSessionOptions {
    table?: string;
    db?: string;
}

export interface ApplyChangesetOptions {
    filter?: (tableName: string) => boolean;
    onConflict?: (conflictType: number) => number;
}

export interface FunctionOptions {
    deterministic?: boolean;
    directOnly?: boolean;
    useBigIntArguments?: boolean;
    varargs?: boolean;
}

export interface AggregateOptions<T extends SQLInputValue = SQLInputValue> {
    deterministic?: boolean;
    directOnly?: boolean;
    useBigIntArguments?: boolean;
    varargs?: boolean;
    start?: T | (() => T);
    step?: (accumulator: T, ...args: SQLOutputValue[]) => T;
    result?: (accumulator: T) => SQLInputValue;
    inverse?: (accumulator: T, ...args: SQLOutputValue[]) => T;
}

export interface StatementColumnMetadata {
    column: string | null;
    database: string | null;
    name: string;
    table: string | null;
    type: string | null;
}

export interface StatementResult {
    changes: number;
    lastInsertRowid: number;
}

export interface BackupProgress {
    totalPages: number;
    remainingPages: number;
}

export interface BackupOptions {
    rate?: number;
    progress?: (progress: BackupProgress) => void;
}

declare namespace __scriptgo {
    function sqliteOpen(location: string, flags?: number): object;
    function sqliteClose(db: object): void;
    function sqliteExec(db: object, sql: string): void;
    function sqlitePrepare(db: object, sql: string): object;
    function sqliteRun(stmt: object, firstParam?: unknown, restParams?: unknown): StatementResult;
    function sqliteGet(stmt: object, firstParam?: unknown, restParams?: unknown): object;
    function sqliteAll(stmt: object, firstParam?: unknown, restParams?: unknown): object[];
    function sqliteColumns(stmt: object): StatementColumnMetadata[];
    function sqliteExpandedSQL(stmt: object): string;
    function sqliteFinalize(stmt: object): void;
    function sqliteStmtConfig(stmt: object, bare: number, unk: number, retArr: number, readBi: number): void;
    function sqliteEnableLoadExtension(db: object, enabled: number): void;
    function sqliteLoadExtension(db: object, path: string): void;
    function sqliteCreateSession(db: object, table?: string): object;
    function sqliteSessionChangeset(sess: object): Uint8Array;
    function sqliteSessionPatchset(sess: object): Uint8Array;
    function sqliteSessionClose(sess: object): void;
    function sqliteApplyChangeset(db: object, changeset: Uint8Array, onConflict?: number): boolean;
    function sqliteLocation(db: object, dbName?: string): string;
    function sqliteIsTransaction(db: object): boolean;
    function sqliteCreateFunction(db: object, name: string, deterministic: number, directOnly: number, fn?: unknown): void;
    function sqliteCreateAggregate(db: object, name: string, deterministic: number, directOnly: number, options?: unknown): void;
    function sqliteBackup(db: object, destPath: string): number;
}

export const constants = {
    SQLITE_OPEN_READONLY: 1,
    SQLITE_OPEN_READWRITE: 2,
    SQLITE_OPEN_CREATE: 4,
    SQLITE_CHANGESETAPPLY_NOSAVEPOINT: 1,
    SQLITE_CHANGESETAPPLY_INVERT: 2,
    SQLITE_CHANGESET_OMIT: 0,
    SQLITE_CHANGESET_REPLACE: 1,
    SQLITE_CHANGESET_ABORT: 2,
    SQLITE_CHANGESET_DATA: 1,
    SQLITE_CHANGESET_NOTFOUND: 2,
    SQLITE_CHANGESET_CONFLICT: 3,
    SQLITE_CHANGESET_CONSTRAINT: 4,
    SQLITE_CHANGESET_FOREIGN_KEY: 5,
};

export class Session {
    private _handle: object | null = null;

    constructor(handle: object | null = null) {
        this._handle = handle;
    }

    changeset(): Uint8Array {
        if (!this._handle) return new Uint8Array(0);
        return __scriptgo.sqliteSessionChangeset(this._handle);
    }

    patchset(): Uint8Array {
        if (!this._handle) return new Uint8Array(0);
        return __scriptgo.sqliteSessionPatchset(this._handle);
    }

    close(): void {
        if (this._handle) {
            __scriptgo.sqliteSessionClose(this._handle);
            this._handle = null;
        }
    }
}

export class StatementSync {
    private _handle: object | null = null;
    private _sql: string = "";
    private _allowBare: number = 0;
    private _allowUnk: number = 0;
    private _returnArrays: number = 0;
    private _readBigInts: number = 0;

    constructor(handle: object | null, sql: string = "") {
        this._handle = handle;
        this._sql = sql;
    }

    all(namedParameters?: Record<string, SQLInputValue> | SQLInputValue, ...anonymousParameters: SQLInputValue[]): object[] {
        if (!this._handle) return [];
        return __scriptgo.sqliteAll(this._handle, namedParameters, anonymousParameters);
    }

    columns(): StatementColumnMetadata[] {
        if (!this._handle) return [];
        return __scriptgo.sqliteColumns(this._handle);
    }

    get(namedParameters?: Record<string, SQLInputValue> | SQLInputValue, ...anonymousParameters: SQLInputValue[]): object | undefined {
        if (!this._handle) return undefined;
        return __scriptgo.sqliteGet(this._handle, namedParameters, anonymousParameters);
    }

    iterate(namedParameters?: Record<string, SQLInputValue> | SQLInputValue, ...anonymousParameters: SQLInputValue[]): object[] {
        return this.all(namedParameters, ...anonymousParameters);
    }

    run(namedParameters?: Record<string, SQLInputValue> | SQLInputValue, ...anonymousParameters: SQLInputValue[]): StatementResult {
        if (!this._handle) return { changes: 0, lastInsertRowid: 0 };
        return __scriptgo.sqliteRun(this._handle, namedParameters, anonymousParameters);
    }

    setAllowBareNamedParameters(enabled: boolean): void {
        this._allowBare = enabled ? 1 : 0;
        this._updateConfig();
    }

    setAllowUnknownNamedParameters(enabled: boolean): void {
        this._allowUnk = enabled ? 1 : 0;
        this._updateConfig();
    }

    setReturnArrays(enabled: boolean): void {
        this._returnArrays = enabled ? 1 : 0;
        this._updateConfig();
    }

    setReadBigInts(enabled: boolean): void {
        this._readBigInts = enabled ? 1 : 0;
        this._updateConfig();
    }

    private _updateConfig(): void {
        if (this._handle) {
            __scriptgo.sqliteStmtConfig(this._handle, this._allowBare, this._allowUnk, this._returnArrays, this._readBigInts);
        }
    }

    expandedSQL(): string {
        if (!this._handle) return this._sql;
        return __scriptgo.sqliteExpandedSQL(this._handle);
    }

    sourceSQL(): string {
        return this._sql;
    }

    finalize(): void {
        if (this._handle) {
            __scriptgo.sqliteFinalize(this._handle);
            this._handle = null;
        }
    }
}

export class DatabaseSync {
    isOpen: boolean = true;
    location: string = ":memory:";
    private _handle: object | null = null;

    constructor(location: string = ":memory:", options?: DatabaseSyncOptions) {
        this.location = location;
        if (options && options.open === false) {
            this.isOpen = false;
            this._handle = null;
            return;
        }
        const flags = (options && options.readOnly) ? constants.SQLITE_OPEN_READONLY : (constants.SQLITE_OPEN_READWRITE | constants.SQLITE_OPEN_CREATE);
        this._handle = __scriptgo.sqliteOpen(location, flags);
        this.isOpen = this._handle !== null;
        if (this._handle && options) {
            if (options.enableForeignKeyConstraints === false) {
                this.exec("PRAGMA foreign_keys = OFF;");
            } else {
                this.exec("PRAGMA foreign_keys = ON;");
            }
            if (options.allowExtension) {
                this.enableLoadExtension(true);
            }
        }
    }

    _getHandle(): object | null {
        return this._handle;
    }

    get isTransaction(): boolean {
        if (!this._handle) return false;
        return __scriptgo.sqliteIsTransaction(this._handle);
    }

    open(): void {
        if (!this.isOpen || !this._handle) {
            this._handle = __scriptgo.sqliteOpen(this.location, constants.SQLITE_OPEN_READWRITE | constants.SQLITE_OPEN_CREATE);
            this.isOpen = this._handle !== null;
        }
    }

    close(): void {
        if (this._handle) {
            __scriptgo.sqliteClose(this._handle);
            this._handle = null;
        }
        this.isOpen = false;
    }

    [Symbol.dispose](): void {
        this.close();
    }

    aggregate(name: string, options?: AggregateOptions): void {
        const det = (options && options.deterministic) ? 1 : 0;
        const dir = (options && options.directOnly) ? 1 : 0;
        if (this._handle) {
            __scriptgo.sqliteCreateAggregate(this._handle, name, det, dir, options);
        }
    }

    loadExtension(path: string): void {
        if (this._handle) {
            __scriptgo.sqliteLoadExtension(this._handle, path);
        }
    }

    enableLoadExtension(enabled: boolean): void {
        if (this._handle) {
            __scriptgo.sqliteEnableLoadExtension(this._handle, enabled ? 1 : 0);
        }
    }

    exec(sql: string): void {
        if (this._handle) {
            __scriptgo.sqliteExec(this._handle, sql);
        }
    }

    function(name: string, optionsOrFn?: FunctionOptions | ((...args: SQLOutputValue[]) => SQLInputValue), fn?: (...args: SQLOutputValue[]) => SQLInputValue): void {
        let options: FunctionOptions | undefined;
        let func: unknown;
        if (typeof optionsOrFn === "function") {
            func = optionsOrFn;
        } else if (typeof optionsOrFn === "object" && optionsOrFn !== null) {
            options = optionsOrFn as FunctionOptions;
            func = fn;
        }
        const det = (options && options.deterministic) ? 1 : 0;
        const dir = (options && options.directOnly) ? 1 : 0;
        if (this._handle) {
            __scriptgo.sqliteCreateFunction(this._handle, name, det, dir, func);
        }
    }

    prepare(sql: string): StatementSync {
        if (!this._handle) {
            throw new Error("sqlite: database is not open");
        }
        const stmtHandle = __scriptgo.sqlitePrepare(this._handle, sql);
        return new StatementSync(stmtHandle, sql);
    }

    createSession(options?: CreateSessionOptions): Session {
        if (!this._handle) {
            return new Session(null);
        }
        const table = (options && options.table) ? options.table : "";
        const sessHandle = __scriptgo.sqliteCreateSession(this._handle, table);
        return new Session(sessHandle);
    }

    applyChangeset(changeset: Uint8Array, options?: ApplyChangesetOptions): boolean {
        if (!this._handle) return false;
        return __scriptgo.sqliteApplyChangeset(this._handle, changeset, 0);
    }
}

export async function backup(sourceDb?: DatabaseSync, destination?: string, options?: BackupOptions): Promise<number> {
    if (sourceDb) {
        const handle = sourceDb._getHandle();
        const dest = destination ? destination : "";
        if (handle && dest.length > 0) {
            const pages = __scriptgo.sqliteBackup(handle, dest);
            return Promise.resolve(pages);
        }
    }
    return Promise.resolve(0);
}

export default {
    constants,
    backup,
    Session,
    StatementSync,
    DatabaseSync,
};
