export class Stats {
    size: number;
    mtimeMs: number;
    birthtimeMs: number;
    mode: number;

    constructor(size: number, mtimeMs: number, birthtimeMs: number, mode: number) {
        this.size = size;
        this.mtimeMs = mtimeMs;
        this.birthtimeMs = birthtimeMs;
        this.mode = mode;
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
}

export interface MkdirOptions {
    recursive?: boolean;
}

export interface RmOptions {
    recursive?: boolean;
    force?: boolean;
}

const defaultMkdirOpts: MkdirOptions = { recursive: false };
const defaultRmOpts: RmOptions = { recursive: false, force: false };

declare namespace __scriptgo {
    function readFileSync(path: string, encoding?: string): string;
    function writeFileSync(path: string, data: string): void;
    function existsSync(path: string): boolean;
    function unlinkSync(path: string): void;
    function statSync(path: string): Stats;
    function readdirSync(path: string): string[];
    function copyFileSync(src: string, dest: string): void;
    function renameSync(oldPath: string, newPath: string): void;
    function appendFileSync(path: string, data: string): void;
    function mkdirSync(path: string, recursive?: boolean): void;
    function rmSync(path: string, recursive?: boolean, force?: boolean): void;
}

export function readFileSync(path: string, encoding?: string): string {
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

export function readdirSync(path: string): string[] {
    return __scriptgo.readdirSync(path);
}

export function copyFileSync(src: string, dest: string): void {
    __scriptgo.copyFileSync(src, dest);
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

export class FSPromises {
    private _tag: number = 0;

    constructor() {}

    async readFile(path: string, encoding: string = "utf8"): Promise<string> {
        return readFileSync(path, encoding);
    }

    async writeFile(path: string, data: string): Promise<void> {
        writeFileSync(path, data);
    }

    async stat(path: string): Promise<Stats> {
        return statSync(path);
    }

    async readdir(path: string): Promise<string[]> {
        return readdirSync(path);
    }

    async mkdir(path: string, options: MkdirOptions = defaultMkdirOpts): Promise<void> {
        mkdirSync(path, options);
    }

    async unlink(path: string): Promise<void> {
        unlinkSync(path);
    }

    async copyFile(src: string, dest: string): Promise<void> {
        copyFileSync(src, dest);
    }

    async rename(oldPath: string, newPath: string): Promise<void> {
        renameSync(oldPath, newPath);
    }

    async appendFile(path: string, data: string): Promise<void> {
        appendFileSync(path, data);
    }

    async rm(path: string, options: RmOptions = defaultRmOpts): Promise<void> {
        rmSync(path, options);
    }
}

export const promises: FSPromises = new FSPromises();
