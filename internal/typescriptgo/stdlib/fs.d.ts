export interface Stats {
    size: number;
    mtimeMs: number;
    birthtimeMs: number;
    mode: number;
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

export namespace promises {
    function readFile(path: string, encoding?: string): Promise<string>;
    function writeFile(path: string, data: string): Promise<void>;
    function stat(path: string): Promise<Stats>;
    function readdir(path: string): Promise<string[]>;
    function mkdir(path: string, options?: MkdirOptions): Promise<void>;
    function unlink(path: string): Promise<void>;
    function copyFile(src: string, dest: string): Promise<void>;
    function rename(oldPath: string, newPath: string): Promise<void>;
    function appendFile(path: string, data: string): Promise<void>;
    function rm(path: string, options?: RmOptions): Promise<void>;
}
