export class URLSearchParams {
    constructor(init?: string);
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
    forEach(fn?: (value: string, name: string, parent: URLSearchParams)): void;
    toString(): string;
}
export class URL {
    constructor(input: string, base?: string);
    toString(): string;
    toJSON(): string;
    static canParse(url: string, base?: string): boolean;
    static parse(input: string, base?: string): URL | null;
    static createObjectURL(blob?: string): string;
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
export function parse(urlString: string, parseQueryString?: boolean, slashesDenoteHost?: boolean): Url;
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
export function format(urlObject: URL): string;
export function resolve(from: string, to: string): string;
export function domainToASCII(domain: string): string;
export function domainToUnicode(domain: string): string;
export function fileURLToPath(url: string): string;
export function fileURLToPathBuffer(url: string): string;
export function pathToFileURL(path: string): URL;
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
export function urlToHttpOptions(url: URL): HttpOptionsResult;
