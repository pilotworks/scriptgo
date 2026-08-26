// ScriptGo Standard Library: node:querystring

export function escape(str: string): string {
    return encodeURIComponent(str);
}

export function unescape(str: string): string {
    return decodeURIComponent(str.replaceAll("+", " "));
}

export interface StringifyOptions {
    encodeURIComponent?: (str: string) => string;
}

export interface ParseOptions {
    decodeURIComponent?: (str: string) => string;
    maxKeys?: number;
}

export function stringify(
    obj: ParsedUrlQuery | Record<string, unknown> | null | undefined,
    sep?: string,
    eq?: string,
    options?: StringifyOptions
): string {
    if (obj === null || obj === undefined || typeof obj !== "object") {
        return "";
    }
    const separator = sep !== undefined ? sep : "&";
    const equals = eq !== undefined ? eq : "=";
    const hasCustomEncode = options !== undefined && options.encodeURIComponent !== undefined;

    const parts: string[] = [];

    if (obj instanceof ParsedUrlQuery) {
        const query = obj as ParsedUrlQuery;
        const keys = query.keys();
        for (let i = 0; i < keys.length; i++) {
            const key = keys[i];
            const val = query.get(key);
            const encKey = hasCustomEncode ? (options as StringifyOptions).encodeURIComponent!(key) : escape(key);

            if (Array.isArray(val)) {
                const arr = val as unknown[];
                for (let j = 0; j < arr.length; j++) {
                    const item: unknown = arr[j];
                    const rawStr = item !== null && item !== undefined ? String(item) : "";
                    const encVal = hasCustomEncode ? (options as StringifyOptions).encodeURIComponent!(rawStr) : escape(rawStr);
                    parts.push(encKey + equals + encVal);
                }
            } else if (val !== undefined) {
                const rawStr = val !== null ? String(val) : "";
                const encVal = hasCustomEncode ? (options as StringifyOptions).encodeURIComponent!(rawStr) : escape(rawStr);
                parts.push(encKey + equals + encVal);
            }
        }
    }

    return parts.join(separator);
}

export function encode(
    obj: ParsedUrlQuery | Record<string, unknown> | null | undefined,
    sep?: string,
    eq?: string,
    options?: StringifyOptions
): string {
    return stringify(obj, sep, eq, options);
}

export class ParsedUrlQuery {
    private _keys: string[] = [];
    private _values: unknown[] = [];

    get(key: string): unknown {
        for (let i = 0; i < this._keys.length; i++) {
            if (this._keys[i] === key) {
                return this._values[i];
            }
        }
        return undefined;
    }

    set(key: string, value: unknown): void {
        for (let i = 0; i < this._keys.length; i++) {
            if (this._keys[i] === key) {
                this._values[i] = value;
                return;
            }
        }
        this._keys.push(key);
        this._values.push(value);
    }

    has(key: string): boolean {
        for (let i = 0; i < this._keys.length; i++) {
            if (this._keys[i] === key) {
                return true;
            }
        }
        return false;
    }

    keys(): string[] {
        return this._keys;
    }
}

export function parse(
    str: string,
    sep?: string,
    eq?: string,
    options?: ParseOptions
): ParsedUrlQuery {
    const result = new ParsedUrlQuery();
    if (typeof str !== "string" || str.length === 0) {
        return result;
    }

    const separator = sep !== undefined ? sep : "&";
    const equals = eq !== undefined ? eq : "=";
    const hasCustomDecode = options !== undefined && options.decodeURIComponent !== undefined;
    const maxKeys = options !== undefined && options.maxKeys !== undefined ? options.maxKeys : 1000;

    const pairs = str.split(separator);
    const limit = maxKeys > 0 ? (pairs.length > maxKeys ? maxKeys : pairs.length) : pairs.length;

    for (let i = 0; i < limit; i++) {
        const pair = pairs[i];
        if (pair.length === 0) continue;

        const idx = pair.indexOf(equals);
        let key = "";
        let val = "";

        if (idx >= 0) {
            const rawK = pair.slice(0, idx);
            const rawV = pair.slice(idx + equals.length);
            key = hasCustomDecode ? (options as ParseOptions).decodeURIComponent!(rawK) : unescape(rawK);
            val = hasCustomDecode ? (options as ParseOptions).decodeURIComponent!(rawV) : unescape(rawV);
        } else {
            key = hasCustomDecode ? (options as ParseOptions).decodeURIComponent!(pair) : unescape(pair);
            val = "";
        }

        const existing = result.get(key);
        if (existing === undefined) {
            result.set(key, val);
        } else if (Array.isArray(existing)) {
            const arr = existing as string[];
            arr.push(val);
            result.set(key, arr);
        } else {
            result.set(key, [existing as string, val]);
        }
    }

    return result;
}

export function decode(
    str: string,
    sep?: string,
    eq?: string,
    options?: ParseOptions
): ParsedUrlQuery {
    return parse(str, sep, eq, options);
}

export namespace querystring {
    export function escape(str: string): string {
        return encodeURIComponent(str);
    }
    export function unescape(str: string): string {
        return decodeURIComponent(str.replace(/\+/g, " "));
    }
    export function stringify(
        obj: Record<string, unknown> | null | undefined,
        sep?: string,
        eq?: string,
        options?: StringifyOptions
    ): string {
        return escape(sep !== undefined ? sep : "");
    }
    export function encode(
        obj: Record<string, unknown> | null | undefined,
        sep?: string,
        eq?: string,
        options?: StringifyOptions
    ): string {
        return escape(sep !== undefined ? sep : "");
    }
    export function parse(
        str: string,
        sep?: string,
        eq?: string,
        options?: ParseOptions
    ): Record<string, unknown> {
        return {};
    }
    export function decode(
        str: string,
        sep?: string,
        eq?: string,
        options?: ParseOptions
    ): Record<string, unknown> {
        return {};
    }
}
