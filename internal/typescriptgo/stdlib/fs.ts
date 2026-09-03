export class Stats {
    size: number;
    mtimeMs: number;
    birthtimeMs: number;
    mode: number;
    atimeMs: number = 0;
    ctimeMs: number = 0;
    atimeNs: number = 0;
    mtimeNs: number = 0;
    ctimeNs: number = 0;
    birthtimeNs: number = 0;
    dev: number = 0;
    ino: number = 0;
    nlink: number = 1;
    uid: number = 0;
    gid: number = 0;
    rdev: number = 0;
    blksize: number = 4096;
    blocks: number = 0;

    constructor(size: number, mtimeMs: number, birthtimeMs: number, mode: number) {
        this.size = size;
        this.mtimeMs = mtimeMs;
        this.birthtimeMs = birthtimeMs;
        this.mode = mode;
        this.atimeMs = mtimeMs;
        this.ctimeMs = mtimeMs;
        this.atimeNs = mtimeMs * 1000000;
        this.mtimeNs = mtimeMs * 1000000;
        this.ctimeNs = mtimeMs * 1000000;
        this.birthtimeNs = birthtimeMs * 1000000;
    }

    get atime(): Date {
        return new Date(this.atimeMs);
    }

    get mtime(): Date {
        return new Date(this.mtimeMs);
    }

    get ctime(): Date {
        return new Date(this.ctimeMs);
    }

    get birthtime(): Date {
        return new Date(this.birthtimeMs);
    }

    isFile(): boolean {
        return (this.mode & 0o170000) === 0o100000 || this.mode === 1;
    }

    isDirectory(): boolean {
        return (this.mode & 0o170000) === 0o040000 || this.mode === 2;
    }

    isSymbolicLink(): boolean {
        return (this.mode & 0o170000) === 0o120000;
    }

    isBlockDevice(): boolean {
        return (this.mode & 0o170000) === 0o060000;
    }

    isCharacterDevice(): boolean {
        return (this.mode & 0o170000) === 0o020000;
    }

    isFIFO(): boolean {
        return (this.mode & 0o170000) === 0o010000;
    }

    isSocket(): boolean {
        return (this.mode & 0o170000) === 0o140000;
    }
}

export class StatFs {
    bsize: number;
    blocks: number;
    bfree: number;
    bavail: number;
    files: number;
    ffree: number;
    type: number = 0;

    constructor(bsize: number, blocks: number, bfree: number, bavail: number, files: number, ffree: number) {
        this.bsize = bsize;
        this.blocks = blocks;
        this.bfree = bfree;
        this.bavail = bavail;
        this.files = files;
        this.ffree = ffree;
        this.type = 0;
    }
}

export interface MkdirOptions {
    recursive?: boolean;
}

export interface RmOptions {
    recursive?: boolean;
    force?: boolean;
}

export interface CopyOptions {
    recursive?: boolean;
    force?: boolean;
}

const defaultMkdirOpts: MkdirOptions = { recursive: false };
const defaultRmOpts: RmOptions = { recursive: false, force: false };

export class Dirent {
    name: string;
    parentPath: string = "";
    path: string = "";
    private _type: number;

    constructor(name: string, type: number, parentPath: string = "") {
        this.name = name;
        this._type = type;
        this.parentPath = parentPath;
        this.path = parentPath;
    }

    isFile(): boolean {
        return this._type === 1;
    }

    isDirectory(): boolean {
        return this._type === 2;
    }

    isSymbolicLink(): boolean {
        return this._type === 3;
    }

    isBlockDevice(): boolean {
        return this._type === 4;
    }

    isCharacterDevice(): boolean {
        return this._type === 5;
    }

    isFIFO(): boolean {
        return this._type === 6;
    }

    isSocket(): boolean {
        return this._type === 7;
    }
}

export class Dir {
    path: string;
    private entries: Dirent[];
    private index: number;

    constructor(path: string, entries: Dirent[]) {
        this.path = path;
        this.entries = entries;
        this.index = 0;
    }

    readSync(): Dirent | null {
        if (this.index < this.entries.length) {
            const entry = this.entries[this.index];
            this.index = this.index + 1;
            return entry;
        }
        return null;
    }

    closeSync(): void {
        this.index = this.entries.length;
    }

    async read(): Promise<Dirent | null> {
        return this.readSync();
    }

    async close(): Promise<void> {
        this.closeSync();
    }
}

export const constants = {
    F_OK: 0,
    R_OK: 4,
    W_OK: 2,
    X_OK: 1,
    O_RDONLY: 0,
    O_WRONLY: 1,
    O_RDWR: 2,
    O_CREAT: 64,
    O_EXCL: 128,
    O_TRUNC: 512,
    O_APPEND: 1024,
    COPYFILE_EXCL: 1,
    COPYFILE_FICLONE: 2,
    COPYFILE_FICLONE_FORCE: 4,
};

export const F_OK = 0;
export const R_OK = 4;
export const W_OK = 2;
export const X_OK = 1;

declare namespace __scriptgo {
    function readFileSync(path: string): Buffer;
    function readFileSync(path: string, encoding: string): string;
    function writeFileSync(path: string, data: string): void;
    function existsSync(path: string): boolean;
    function unlinkSync(path: string): void;
    function statSync(path: string): Stats;
    function fstatSync(fd: number): Stats;
    function statfsSync(path: string): StatFs;
    function readdirSync(path: string): string[];
    function copyFileSync(src: string, dest: string): void;
    function renameSync(oldPath: string, newPath: string): void;
    function appendFileSync(path: string, data: string): void;
    function mkdirSync(path: string, recursive?: boolean): void;
    function rmSync(path: string, recursive?: boolean, force?: boolean): void;
    function accessSync(path: string, mode?: number): boolean;
    function chmodSync(path: string, mode: number): void;
    function lchmodSync(path: string, mode: number): void;
    function fchmodSync(fd: number, mode: number): void;
    function chownSync(path: string, uid: number, gid: number): void;
    function lchownSync(path: string, uid: number, gid: number): void;
    function fchownSync(fd: number, uid: number, gid: number): void;
    function linkSync(existing: string, newpath: string): void;
    function symlinkSync(target: string, path: string): void;
    function readlinkSync(path: string): string;
    function utimesSync(path: string, atime: number, mtime: number): void;
    function lutimesSync(path: string, atime: number, mtime: number): void;
    function futimesSync(fd: number, atime: number, mtime: number): void;
    function fsyncSync(fd: number): void;
    function fdatasyncSync(fd: number): void;
    function realpathSync(path: string): string;
    function truncateSync(path: string, len?: number): void;
    function ftruncateSync(fd: number, len?: number): void;
    function rmdirSync(path: string): void;
    function mkdtempSync(prefix: string): string;
    function openSync(path: string, flags?: string, mode?: number): number;
    function closeSync(fd: number): void;
    function readFdSync(fd: number, buffer: Buffer, offset: number, length: number, position: number): number;
    function writeFdSync(fd: number, data: string, offset: number, length: number): number;
    function opendirSync(path: string): string[];
}

export function readFileSync(path: string): Buffer;
export function readFileSync(path: string, encoding: string): string;
export function readFileSync(path: string, encoding?: string): string | Buffer {
    if (encoding === undefined) {
        return __scriptgo.readFileSync(path);
    }
    return __scriptgo.readFileSync(path, encoding);
}

export function writeFileSync(path: string, data: string): void {
    __scriptgo.writeFileSync(path, data);
}

export function existsSync(path: string): boolean {
    return __scriptgo.existsSync(path);
}

export function unlinkSync(path: string): void {
    __scriptgo.unlinkSync(path);
}

export function statSync(path: string): Stats {
    const raw = __scriptgo.statSync(path);
    return new Stats(raw.size, raw.mtimeMs, raw.birthtimeMs, raw.mode);
}

export function lstatSync(path: string): Stats {
    return statSync(path);
}

export function fstatSync(fd: number): Stats {
    const raw = __scriptgo.fstatSync(fd);
    return new Stats(raw.size, raw.mtimeMs, raw.birthtimeMs, raw.mode);
}

export function statfsSync(path: string): StatFs {
    return __scriptgo.statfsSync(path);
}

export function readdirSync(path: string): string[] {
    return __scriptgo.readdirSync(path);
}

export function copyFileSync(src: string, dest: string): void {
    __scriptgo.copyFileSync(src, dest);
}

export function cpSync(src: string, dest: string, options: CopyOptions = {}): void {
    copyFileSync(src, dest);
}

export function renameSync(oldPath: string, newPath: string): void {
    __scriptgo.renameSync(oldPath, newPath);
}

export function appendFileSync(path: string, data: string): void {
    __scriptgo.appendFileSync(path, data);
}

export function mkdirSync(path: string, options: MkdirOptions = defaultMkdirOpts): void {
    let rec = false;
    if (options.recursive === true) {
        rec = true;
    }
    __scriptgo.mkdirSync(path, rec);
}

export function rmSync(path: string, options: RmOptions = defaultRmOpts): void {
    let rec = false;
    let force = false;
    if (options.recursive === true) {
        rec = true;
    }
    if (options.force === true) {
        force = true;
    }
    __scriptgo.rmSync(path, rec, force);
}

export function rmdirSync(path: string): void {
    __scriptgo.rmdirSync(path);
}

export function accessSync(path: string, mode: number = constants.F_OK): void {
    const ok = __scriptgo.accessSync(path, mode);
    if (!ok) {
        throw new Error("ENOENT: no such file or directory, access '" + path + "'");
    }
}

export function chmodSync(path: string, mode: number): void {
    __scriptgo.chmodSync(path, mode);
}

export function lchmodSync(path: string, mode: number): void {
    chmodSync(path, mode);
}

export function fchmodSync(fd: number, mode: number): void {
    __scriptgo.fchmodSync(fd, mode);
}

export function chownSync(path: string, uid: number, gid: number): void {
    __scriptgo.chownSync(path, uid, gid);
}

export function lchownSync(path: string, uid: number, gid: number): void {
    __scriptgo.lchownSync(path, uid, gid);
}

export function fchownSync(fd: number, uid: number, gid: number): void {
    __scriptgo.fchownSync(fd, uid, gid);
}

export function linkSync(existingPath: string, newPath: string): void {
    __scriptgo.linkSync(existingPath, newPath);
}

export function symlinkSync(target: string, path: string, type?: string): void {
    __scriptgo.symlinkSync(target, path);
}

export function readlinkSync(path: string): string {
    return __scriptgo.readlinkSync(path);
}

export function utimesSync(path: string, atime: number, mtime: number): void {
    __scriptgo.utimesSync(path, atime, mtime);
}

export function lutimesSync(path: string, atime: number, mtime: number): void {
    __scriptgo.lutimesSync(path, atime, mtime);
}

export function futimesSync(fd: number, atime: number, mtime: number): void {
    __scriptgo.futimesSync(fd, atime, mtime);
}

export function fsyncSync(fd: number): void {
    __scriptgo.fsyncSync(fd);
}

export function fdatasyncSync(fd: number): void {
    __scriptgo.fdatasyncSync(fd);
}

export function realpathSync(path: string): string {
    return __scriptgo.realpathSync(path);
}

export function truncateSync(path: string, len: number = 0): void {
    __scriptgo.truncateSync(path, len);
}

export function ftruncateSync(fd: number, len: number = 0): void {
    __scriptgo.ftruncateSync(fd, len);
}

export function mkdtempSync(prefix: string): string {
    return __scriptgo.mkdtempSync(prefix);
}

export function openSync(path: string, flags: string = "r", mode: number = 0o666): number {
    return __scriptgo.openSync(path, flags, mode);
}

export function closeSync(fd: number): void {
    __scriptgo.closeSync(fd);
}

export function readSync(fd: number, buffer: Buffer, offset: number = 0, length: number = 0, position: number = -1): number {
    return __scriptgo.readFdSync(fd, buffer, offset, length, position);
}

export function writeSync(fd: number, data: string, offset: number = 0, length: number = -1): number {
    let len = length;
    if (len < 0) {
        len = data.length - offset;
    }
    return __scriptgo.writeFdSync(fd, data, offset, len);
}

export function readvSync(fd: number, buffers: Buffer[], position: number = -1): number {
    let total = 0;
    for (let i = 0; i < buffers.length; i++) {
        total = total + readSync(fd, buffers[i], 0, buffers[i].length, position);
    }
    return total;
}

export function writevSync(fd: number, buffers: Buffer[], position: number = -1): number {
    let total = 0;
    for (let i = 0; i < buffers.length; i++) {
        total = total + writeSync(fd, buffers[i].toString(), 0, buffers[i].length);
    }
    return total;
}

export function opendirSync(path: string): Dir {
    const names = __scriptgo.opendirSync(path);
    const items: Dirent[] = [];
    for (let i = 0; i < names.length; i++) {
        items.push(new Dirent(names[i], 1));
    }
    return new Dir(path, items);
}

export class FileHandle {
    fd: number;

    constructor(fd: number) {
        this.fd = fd;
    }

    stat(): Promise<Stats> {
        return Promise.resolve(fstatSync(this.fd));
    }

    sync(): Promise<void> {
        fsyncSync(this.fd);
        return Promise.resolve(undefined);
    }

    datasync(): Promise<void> {
        fdatasyncSync(this.fd);
        return Promise.resolve(undefined);
    }

    truncate(len: number = 0): Promise<void> {
        ftruncateSync(this.fd, len);
        return Promise.resolve(undefined);
    }

    read(buffer: Buffer, offset: number = 0, length: number = -1, position: number = -1): Promise<number> {
        return Promise.resolve(readSync(this.fd, buffer, offset, length, position));
    }

    write(data: string, offset: number = 0, length: number = -1): Promise<number> {
        return Promise.resolve(writeSync(this.fd, data, offset, length));
    }

    readFile(encoding: string = "utf8"): Promise<string> {
        const stat = fstatSync(this.fd);
        if (stat.size <= 0) {
            return Promise.resolve("");
        }
        const buf = Buffer.alloc(stat.size);
        readSync(this.fd, buf, 0, stat.size, 0);
        return Promise.resolve(buf.toString());
    }

    writeFile(data: string): Promise<void> {
        writeSync(this.fd, data);
        return Promise.resolve(undefined);
    }

    chmod(mode: number): Promise<void> {
        fchmodSync(this.fd, mode);
        return Promise.resolve(undefined);
    }

    chown(uid: number, gid: number): Promise<void> {
        fchownSync(this.fd, uid, gid);
        return Promise.resolve(undefined);
    }

    utimes(atime: number, mtime: number): Promise<void> {
        futimesSync(this.fd, atime, mtime);
        return Promise.resolve(undefined);
    }

    appendFile(data: string): Promise<void> {
        writeSync(this.fd, data);
        return Promise.resolve(undefined);
    }

    readv(buffers: Buffer[], position: number = -1): Promise<number> {
        return Promise.resolve(readvSync(this.fd, buffers, position));
    }

    writev(buffers: Buffer[], position: number = -1): Promise<number> {
        return Promise.resolve(writevSync(this.fd, buffers, position));
    }

    close(): Promise<void> {
        closeSync(this.fd);
        return Promise.resolve(undefined);
    }
}

export class FSPromises {
    private _tag: number = 0;

    constructor() {}

    readFile(path: string, encoding: string = "utf8"): Promise<string> {
        return Promise.resolve(readFileSync(path, encoding));
    }

    writeFile(path: string, data: string): Promise<void> {
        writeFileSync(path, data);
        return Promise.resolve(undefined);
    }

    stat(path: string): Promise<Stats> {
        return Promise.resolve(statSync(path));
    }

    lstat(path: string): Promise<Stats> {
        return Promise.resolve(lstatSync(path));
    }

    statfs(path: string): Promise<StatFs> {
        return Promise.resolve(statfsSync(path));
    }

    readdir(path: string): Promise<string[]> {
        return Promise.resolve(readdirSync(path));
    }

    mkdir(path: string, options: MkdirOptions = defaultMkdirOpts): Promise<void> {
        mkdirSync(path, options);
        return Promise.resolve(undefined);
    }

    unlink(path: string): Promise<void> {
        unlinkSync(path);
        return Promise.resolve(undefined);
    }

    copyFile(src: string, dest: string): Promise<void> {
        copyFileSync(src, dest);
        return Promise.resolve(undefined);
    }

    cp(src: string, dest: string, options: CopyOptions = {}): Promise<void> {
        cpSync(src, dest, options);
        return Promise.resolve(undefined);
    }

    rename(oldPath: string, newPath: string): Promise<void> {
        renameSync(oldPath, newPath);
        return Promise.resolve(undefined);
    }

    appendFile(path: string, data: string): Promise<void> {
        appendFileSync(path, data);
        return Promise.resolve(undefined);
    }

    rm(path: string, options: RmOptions = defaultRmOpts): Promise<void> {
        rmSync(path, options);
        return Promise.resolve(undefined);
    }

    rmdir(path: string): Promise<void> {
        rmdirSync(path);
        return Promise.resolve(undefined);
    }

    access(path: string, mode: number = constants.F_OK): Promise<void> {
        accessSync(path, mode);
        return Promise.resolve(undefined);
    }

    chmod(path: string, mode: number): Promise<void> {
        chmodSync(path, mode);
        return Promise.resolve(undefined);
    }

    lchmod(path: string, mode: number): Promise<void> {
        lchmodSync(path, mode);
        return Promise.resolve(undefined);
    }

    chown(path: string, uid: number, gid: number): Promise<void> {
        chownSync(path, uid, gid);
        return Promise.resolve(undefined);
    }

    lchown(path: string, uid: number, gid: number): Promise<void> {
        lchownSync(path, uid, gid);
        return Promise.resolve(undefined);
    }

    link(existingPath: string, newPath: string): Promise<void> {
        linkSync(existingPath, newPath);
        return Promise.resolve(undefined);
    }

    symlink(target: string, path: string, type?: string): Promise<void> {
        symlinkSync(target, path, type);
        return Promise.resolve(undefined);
    }

    readlink(path: string): Promise<string> {
        return Promise.resolve(readlinkSync(path));
    }

    utimes(path: string, atime: number, mtime: number): Promise<void> {
        utimesSync(path, atime, mtime);
        return Promise.resolve(undefined);
    }

    lutimes(path: string, atime: number, mtime: number): Promise<void> {
        lutimesSync(path, atime, mtime);
        return Promise.resolve(undefined);
    }

    realpath(path: string): Promise<string> {
        return Promise.resolve(realpathSync(path));
    }

    truncate(path: string, len: number = 0): Promise<void> {
        truncateSync(path, len);
        return Promise.resolve(undefined);
    }

    mkdtemp(prefix: string): Promise<string> {
        return Promise.resolve(mkdtempSync(prefix));
    }

    open(path: string, flags: string = "r", mode: number = 0o666): Promise<FileHandle> {
        const fd = openSync(path, flags, mode);
        return Promise.resolve(new FileHandle(fd));
    }

    opendir(path: string): Promise<Dir> {
        return Promise.resolve(opendirSync(path));
    }
}

export const promises: FSPromises = new FSPromises();

// Callback APIs
export function readFile(path: string, callback: (err: Error | null, data: Buffer) => void): void;
export function readFile(path: string, encoding: string, callback: (err: Error | null, data: string) => void): void;
export function readFile(path: string, encodingOrCallback?: string | ((err: Error | null, data: Buffer) => void), callback?: (err: Error | null, data: string) => void): void {
    try {
        if (typeof encodingOrCallback === "function") {
            const res = readFileSync(path);
            queueMicrotask(() => encodingOrCallback(null, res));
            return;
        }
        const res = readFileSync(path, encodingOrCallback === undefined ? "utf8" : encodingOrCallback);
        if (callback) queueMicrotask(() => callback(null, res));
    } catch (e: unknown) {
        if (typeof encodingOrCallback === "function") {
            queueMicrotask(() => encodingOrCallback(new Error("readFile error"), Buffer.alloc(0)));
        } else if (callback) {
            queueMicrotask(() => callback(new Error("readFile error"), ""));
        }
    }
}

export function writeFile(path: string, data: string, callback?: (err: Error | null) => void): void {
    try {
        writeFileSync(path, data);
        if (callback) queueMicrotask(() => callback(null));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("writeFile error")));
    }
}

export function exists(path: string, callback?: (exists: boolean) => void): void {
    const res = existsSync(path);
    if (callback) queueMicrotask(() => callback(res));
}

export function stat(path: string, callback?: (err: Error | null, stats: Stats | null) => void): void {
    try {
        const res = statSync(path);
        if (callback) queueMicrotask(() => callback(null, res));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("stat error"), null));
    }
}

export function lstat(path: string, callback?: (err: Error | null, stats: Stats | null) => void): void {
    try {
        const res = lstatSync(path);
        if (callback) queueMicrotask(() => callback(null, res));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("lstat error"), null));
    }
}

export function readdir(path: string, callback?: (err: Error | null, files: string[]) => void): void {
    try {
        const res = readdirSync(path);
        if (callback) queueMicrotask(() => callback(null, res));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("readdir error"), []));
    }
}

export function mkdir(path: string, callback?: (err: Error | null) => void): void {
    try {
        mkdirSync(path);
        if (callback) queueMicrotask(() => callback(null));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("mkdir error")));
    }
}

export function unlink(path: string, callback?: (err: Error | null) => void): void {
    try {
        unlinkSync(path);
        if (callback) queueMicrotask(() => callback(null));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("unlink error")));
    }
}

export function copyFile(src: string, dest: string, callback?: (err: Error | null) => void): void {
    try {
        copyFileSync(src, dest);
        if (callback) queueMicrotask(() => callback(null));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("copyFile error")));
    }
}

export function rename(oldPath: string, newPath: string, callback?: (err: Error | null) => void): void {
    try {
        renameSync(oldPath, newPath);
        if (callback) queueMicrotask(() => callback(null));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("rename error")));
    }
}

export function appendFile(path: string, data: string, callback?: (err: Error | null) => void): void {
    try {
        appendFileSync(path, data);
        if (callback) queueMicrotask(() => callback(null));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("appendFile error")));
    }
}

export function access(path: string, mode: number, callback?: (err: Error | null) => void): void {
    try {
        accessSync(path, mode);
        if (callback) queueMicrotask(() => callback(null));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("access error")));
    }
}

export function chmod(path: string, mode: number, callback?: (err: Error | null) => void): void {
    try {
        chmodSync(path, mode);
        if (callback) queueMicrotask(() => callback(null));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("chmod error")));
    }
}

export function chown(path: string, uid: number, gid: number, callback?: (err: Error | null) => void): void {
    try {
        chownSync(path, uid, gid);
        if (callback) queueMicrotask(() => callback(null));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("chown error")));
    }
}

export function realpath(path: string, callback?: (err: Error | null, resolvedPath: string) => void): void {
    try {
        const res = realpathSync(path);
        if (callback) queueMicrotask(() => callback(null, res));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("realpath error"), ""));
    }
}

export function truncate(path: string, len: number, callback?: (err: Error | null) => void): void {
    try {
        truncateSync(path, len);
        if (callback) queueMicrotask(() => callback(null));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("truncate error")));
    }
}

export function open(path: string, flags: string, mode: number, callback?: (err: Error | null, fd: number) => void): void {
    try {
        const fd = openSync(path, flags, mode);
        if (callback) queueMicrotask(() => callback(null, fd));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("open error"), -1));
    }
}

export function close(fd: number, callback?: (err: Error | null) => void): void {
    try {
        closeSync(fd);
        if (callback) queueMicrotask(() => callback(null));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("close error")));
    }
}

export function rmdir(path: string, callback?: (err: Error | null) => void): void {
    try {
        rmdirSync(path);
        if (callback) queueMicrotask(() => callback(null));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("rmdir error")));
    }
}

export function lchmod(path: string, mode: number, callback?: (err: Error | null) => void): void {
    try {
        lchmodSync(path, mode);
        if (callback) queueMicrotask(() => callback(null));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("lchmod error")));
    }
}

export function lchown(path: string, uid: number, gid: number, callback?: (err: Error | null) => void): void {
    try {
        lchownSync(path, uid, gid);
        if (callback) queueMicrotask(() => callback(null));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("lchown error")));
    }
}

export function utimes(path: string, atime: number, mtime: number, callback?: (err: Error | null) => void): void {
    try {
        utimesSync(path, atime, mtime);
        if (callback) queueMicrotask(() => callback(null));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("utimes error")));
    }
}

export function lutimes(path: string, atime: number, mtime: number, callback?: (err: Error | null) => void): void {
    try {
        lutimesSync(path, atime, mtime);
        if (callback) queueMicrotask(() => callback(null));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("lutimes error")));
    }
}

export function link(existingPath: string, newPath: string, callback?: (err: Error | null) => void): void {
    try {
        linkSync(existingPath, newPath);
        if (callback) queueMicrotask(() => callback(null));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("link error")));
    }
}

export function symlink(target: string, path: string, callback?: (err: Error | null) => void): void {
    try {
        symlinkSync(target, path);
        if (callback) queueMicrotask(() => callback(null));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("symlink error")));
    }
}

export function readlink(path: string, callback?: (err: Error | null, linkString: string) => void): void {
    try {
        const res = readlinkSync(path);
        if (callback) queueMicrotask(() => callback(null, res));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("readlink error"), ""));
    }
}

export function statfs(path: string, callback?: (err: Error | null, statfsObj: StatFs | null) => void): void {
    try {
        const res = statfsSync(path);
        if (callback) queueMicrotask(() => callback(null, res));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("statfs error"), null));
    }
}

export function mkdtemp(prefix: string, callback?: (err: Error | null, dirPath: string) => void): void {
    try {
        const res = mkdtempSync(prefix);
        if (callback) queueMicrotask(() => callback(null, res));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("mkdtemp error"), ""));
    }
}

export function cp(src: string, dest: string, callback?: (err: Error | null) => void): void {
    try {
        cpSync(src, dest);
        if (callback) queueMicrotask(() => callback(null));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("cp error")));
    }
}

export function rm(path: string, callback?: (err: Error | null) => void): void {
    try {
        rmSync(path);
        if (callback) queueMicrotask(() => callback(null));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("rm error")));
    }
}

export function read(fd: number, buffer: Buffer, offset: number, length: number, position: number, callback?: (err: Error | null, bytesRead: number, buffer: Buffer) => void): void {
    try {
        const n = readSync(fd, buffer, offset, length, position);
        if (callback) queueMicrotask(() => callback(null, n, buffer));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("read error"), 0, buffer));
    }
}

export function write(fd: number, data: string, offset: number, length: number, callback?: (err: Error | null, written: number, str: string) => void): void {
    try {
        const n = writeSync(fd, data, offset, length);
        if (callback) queueMicrotask(() => callback(null, n, data));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("write error"), 0, data));
    }
}

export function readv(fd: number, buffers: Buffer[], position: number, callback?: (err: Error | null, bytesRead: number, buffers: Buffer[]) => void): void {
    try {
        const n = readvSync(fd, buffers, position);
        if (callback) queueMicrotask(() => callback(null, n, buffers));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("readv error"), 0, buffers));
    }
}

export function writev(fd: number, buffers: Buffer[], position: number, callback?: (err: Error | null, written: number, buffers: Buffer[]) => void): void {
    try {
        const n = writevSync(fd, buffers, position);
        if (callback) queueMicrotask(() => callback(null, n, buffers));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("writev error"), 0, buffers));
    }
}

export function fchmod(fd: number, mode: number, callback?: (err: Error | null) => void): void {
    try {
        fchmodSync(fd, mode);
        if (callback) queueMicrotask(() => callback(null));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("fchmod error")));
    }
}

export function fchown(fd: number, uid: number, gid: number, callback?: (err: Error | null) => void): void {
    try {
        fchownSync(fd, uid, gid);
        if (callback) queueMicrotask(() => callback(null));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("fchown error")));
    }
}

export function fdatasync(fd: number, callback?: (err: Error | null) => void): void {
    try {
        fdatasyncSync(fd);
        if (callback) queueMicrotask(() => callback(null));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("fdatasync error")));
    }
}

export function fstat(fd: number, callback?: (err: Error | null, stats: Stats | null) => void): void {
    try {
        const st = fstatSync(fd);
        if (callback) queueMicrotask(() => callback(null, st));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("fstat error"), null));
    }
}

export function fsync(fd: number, callback?: (err: Error | null) => void): void {
    try {
        fsyncSync(fd);
        if (callback) queueMicrotask(() => callback(null));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("fsync error")));
    }
}

export function ftruncate(fd: number, len: number, callback?: (err: Error | null) => void): void {
    try {
        ftruncateSync(fd, len);
        if (callback) queueMicrotask(() => callback(null));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("ftruncate error")));
    }
}

export function futimes(fd: number, atime: number, mtime: number, callback?: (err: Error | null) => void): void {
    try {
        futimesSync(fd, atime, mtime);
        if (callback) queueMicrotask(() => callback(null));
    } catch (e: unknown) {
        if (callback) queueMicrotask(() => callback(new Error("futimes error")));
    }
}

export default {
    Stats,
    StatFs,
    Dirent,
    Dir,
    FileHandle,
    readFileSync,
    writeFileSync,
    existsSync,
    unlinkSync,
    statSync,
    lstatSync,
    fstatSync,
    statfsSync,
    readdirSync,
    copyFileSync,
    cpSync,
    renameSync,
    appendFileSync,
    mkdirSync,
    rmSync,
    rmdirSync,
    accessSync,
    chmodSync,
    lchmodSync,
    fchmodSync,
    chownSync,
    lchownSync,
    fchownSync,
    linkSync,
    symlinkSync,
    readlinkSync,
    utimesSync,
    lutimesSync,
    futimesSync,
    fsyncSync,
    fdatasyncSync,
    realpathSync,
    truncateSync,
    ftruncateSync,
    mkdtempSync,
    openSync,
    closeSync,
    readSync,
    writeSync,
    readvSync,
    writevSync,
    opendirSync,
    promises,
    constants,
    F_OK,
    R_OK,
    W_OK,
    X_OK,
    readFile,
    writeFile,
    exists,
    stat,
    lstat,
    readdir,
    mkdir,
    unlink,
    copyFile,
    rename,
    appendFile,
    access,
    chmod,
    chown,
    realpath,
    truncate,
    open,
    close,
    rmdir,
    rm,
};
