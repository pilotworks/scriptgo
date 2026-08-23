export class URLSearchParams {
    constructor(init?: unknown);
    append(name: string, value: string): void;
    delete(name: string, value?: string): void;
    get(name: string): string | null;
    getAll(name: string): string[];
    has(name: string, value?: string): boolean;
    set(name: string, value: string): void;
    sort(): void;
    entries(): string[][];
    keys(): string[];
    values(): string[];
    forEach(fn: (value: string, name: string, parent: URLSearchParams) => void, thisArg?: unknown): void;
    toString(): string;
    readonly size: number;
}

export class URL {
    constructor(input: string, base?: string);
    href: string;
    origin: string;
    protocol: string;
    host: string;
    hostname: string;
    port: string;
    pathname: string;
    search: string;
    hash: string;
    username: string;
    password: string;
    readonly searchParams: URLSearchParams;
    toString(): string;
    toJSON(): string;
    static canParse(url: string, base?: string): boolean;
    static parse(input: string, base?: string): URL | null;
    static createObjectURL(blob: unknown): string;
    static revokeObjectURL(id: string): void;
}

export class Url {
    auth: string;
    hash: string;
    host: string;
    hostname: string;
    href: string;
    path: string;
    pathname: string;
    port: string;
    protocol: string;
    query: string;
    search: string;
    slashes: boolean;
}

export interface UrlObjectInput {
    href?: string;
    protocol?: string;
    slashes?: boolean;
    auth?: string;
    host?: string;
    hostname?: string;
    port?: string | number;
    pathname?: string;
    path?: string;
    search?: string;
    hash?: string;
}

export interface HttpOptionsResult {
    protocol: string;
    hostname: string;
    hash: string;
    search: string;
    pathname: string;
    path: string;
    href: string;
    port: number | null;
    auth: string | null;
}

export function parse(urlString: string, parseQueryString?: boolean, slashesDenoteHost?: boolean): Url;
export function format(urlObject: URL, options?: unknown): string;
export function resolve(from: string, to: string): string;
export function domainToASCII(domain: string): string;
export function domainToUnicode(domain: string): string;
export function fileURLToPath(url: string, options?: unknown): string;
export function fileURLToPathBuffer(url: string, options?: unknown): string;
export function pathToFileURL(path: string, options?: unknown): URL;
export function urlToHttpOptions(url: URL): HttpOptionsResult;
