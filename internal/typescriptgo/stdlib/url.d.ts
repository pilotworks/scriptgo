export class URLSearchParams {
    constructor(init?: string);
    append(name: string, value: string): void;
    delete(name: string): void;
    get(name: string): string | null;
    getAll(name: string): string[];
    has(name: string): boolean;
    set(name: string, value: string): void;
    sort(): void;
    toString(): string;
}
export class URL {
    constructor(input: string, base?: string);
    toString(): string;
    toJSON(): string;
    static canParse(url: string, base?: string): boolean;
}
