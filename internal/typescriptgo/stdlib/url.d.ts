export interface URLSearchParams {
    append(name: string, value: string): void;
    delete(name: string): void;
    get(name: string): string | null;
    getAll(name: string): string[];
    has(name: string): boolean;
    set(name: string, value: string): void;
    sort(): void;
    toString(): string;
    readonly size: number;
}

export interface URLSearchParamsConstructor {
    new(init?: string): URLSearchParams;
    readonly prototype: URLSearchParams;
}

export var URLSearchParams: URLSearchParamsConstructor;

export interface URL {
    href: string;
    origin: string;
    protocol: string;
    username: string;
    password: string;
    host: string;
    hostname: string;
    port: string;
    pathname: string;
    search: string;
    searchParams: URLSearchParams;
    hash: string;
    toString(): string;
    toJSON(): string;
}

export interface URLConstructor {
    new(url: string, base?: string | URL): URL;
    readonly prototype: URL;
    canParse(url: string, base?: string | URL): boolean;
}

export var URL: URLConstructor;
