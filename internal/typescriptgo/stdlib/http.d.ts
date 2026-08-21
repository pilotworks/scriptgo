export class Headers {
    _keys: string[];
    _values: string[];
    constructor(init?: Headers | null);
    append(name: string, value: string): void;
    delete(name: string): void;
    get(name: string): string | null;
    has(name: string): boolean;
    set(name: string, value: string): void;
    forEach(callback?: (value: string, name: string, parent: Headers)): void;
    entries(): string[][];
    keys(): string[];
    values(): string[];
}
export interface RequestInit {
    method?: string;
    headers?: unknown;
    body?: string | null;
}
export class Request {
    url: string;
    method: string;
    headers: Headers;
    body: string;
    constructor(input: unknown, init?: RequestInit);
}
export interface ResponseInit {
    status?: number;
    statusText?: string;
    headers?: unknown;
}
export class Response {
    ok: boolean;
    status: number;
    statusText: string;
    headers: Headers;
    url: string;
    _body: string;
    constructor(body?: string, init?: ResponseInit);
    async text(): Promise<string>;
    async json<T = unknown>(): Promise<T>;
    async arrayBuffer(): Promise<ArrayBuffer>;
    static json(data: string, init?: ResponseInit): Response;
    static error(): Response;
    static redirect(url: string, status?: number): Response;
}
export class FetchResponseData {
    status: number;
    statusText: string;
    headers: string[];
    body: string;
    constructor(status: number, statusText: string, headers: string[], body: string);
}
export async function fetch(input: unknown, init?: RequestInit): Promise<Response>;
export const METHODS: string[];
export const STATUS_CODES;
export function getStatusText(code: number): string;
