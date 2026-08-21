export class Stats {
    size: number;
    mtimeMs: number;
    birthtimeMs: number;
    mode: number;
    constructor(size: number, mtimeMs: number, birthtimeMs: number, mode: number);
    isFile(): boolean;
    isDirectory(): boolean;
    isSymbolicLink(): boolean;
}
export interface MkdirOptions {
    recursive?: boolean;
}
export interface RmOptions {
    recursive?: boolean;
    force?: boolean;
}
export function readFileSync(path: string, encoding?: string): string;
export function writeFileSync(path: string, data: string): void;
export function existsSync(path: string): boolean;
export function unlinkSync(path: string): void;
export function statSync(path: string): Stats;
export function readdirSync(path: string): string[];
export function copyFileSync(src: string, dest: string): void;
export function renameSync(oldPath: string, newPath: string): void;
export function appendFileSync(path: string, data: string): void;
export function mkdirSync(path: string, options?: MkdirOptions): void;
export function rmSync(path: string, options?: RmOptions): void;
export class FSPromises {
    constructor();
    async readFile(path: string, encoding?: string): Promise<string>;
    async writeFile(path: string, data: string): Promise<void>;
    async stat(path: string): Promise<Stats>;
    async readdir(path: string): Promise<string[]>;
    async mkdir(path: string, options?: MkdirOptions): Promise<void>;
    async unlink(path: string): Promise<void>;
    async copyFile(src: string, dest: string): Promise<void>;
    async rename(oldPath: string, newPath: string): Promise<void>;
    async appendFile(path: string, data: string): Promise<void>;
    async rm(path: string, options?: RmOptions): Promise<void>;
}
export const promises: FSPromises;
