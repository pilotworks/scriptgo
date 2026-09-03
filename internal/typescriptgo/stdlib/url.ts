class URLSearchParamEntry {
    name: string;
    value: string;

    constructor(name: string, value: string) {
        this.name = name;
        this.value = value;
    }
}

export class URLSearchParams {
    private _entries: URLSearchParamEntry[] = [];

    constructor(init: string = "") {
        if (init.length > 0) {
            let str = init;
            if (str.indexOf("?") === 0) {
                str = str.slice(1, str.length);
            }
            if (str.length > 0) {
                const pairs = str.split("&");
                for (let i = 0; i < pairs.length; i++) {
                    const pair = pairs[i];
                    if (pair.length > 0) {
                        const eqIdx = pair.indexOf("=");
                        if (eqIdx >= 0) {
                            const k = pair.slice(0, eqIdx);
                            const v = pair.slice(eqIdx + 1, pair.length);
                            this._entries.push(new URLSearchParamEntry(k, v));
                        } else {
                            this._entries.push(new URLSearchParamEntry(pair, ""));
                        }
                    }
                }
            }
        }
    }

    append(name: string, value: string): void {
        this._entries.push(new URLSearchParamEntry(name, value));
    }

    delete(name: string, value?: string): void {
        const next: URLSearchParamEntry[] = [];
        const hasVal = value !== undefined && value !== "undefined" && value !== null && value !== "null";
        for (let i = 0; i < this._entries.length; i++) {
            if (hasVal) {
                if (this._entries[i].name !== name || this._entries[i].value !== value) {
                    next.push(this._entries[i]);
                }
            } else {
                if (this._entries[i].name !== name) {
                    next.push(this._entries[i]);
                }
            }
        }
        this._entries = next;
    }

    get(name: string): string | null {
        for (let i = 0; i < this._entries.length; i++) {
            if (this._entries[i].name === name) {
                return this._entries[i].value;
            }
        }
        return null;
    }

    getAll(name: string): string[] {
        const res: string[] = [];
        for (let i = 0; i < this._entries.length; i++) {
            if (this._entries[i].name === name) {
                res.push(this._entries[i].value);
            }
        }
        return res;
    }

    has(name: string, value?: string): boolean {
        const hasVal = value !== undefined && value !== "undefined" && value !== null && value !== "null";
        for (let i = 0; i < this._entries.length; i++) {
            if (this._entries[i].name === name) {
                if (hasVal) {
                    if (this._entries[i].value === value) {
                        return true;
                    }
                } else {
                    return true;
                }
            }
        }
        return false;
    }

    set(name: string, value: string): void {
        let found = false;
        const next: URLSearchParamEntry[] = [];
        for (let i = 0; i < this._entries.length; i++) {
            if (this._entries[i].name === name) {
                if (!found) {
                    next.push(new URLSearchParamEntry(name, value));
                    found = true;
                }
            } else {
                next.push(this._entries[i]);
            }
        }
        if (!found) {
            next.push(new URLSearchParamEntry(name, value));
        }
        this._entries = next;
    }

    sort(): void {
        const len = this._entries.length;
        for (let i = 0; i < len - 1; i++) {
            for (let j = 0; j < len - i - 1; j++) {
                if (this._entries[j].name > this._entries[j + 1].name) {
                    const temp = this._entries[j];
                    this._entries[j] = this._entries[j + 1];
                    this._entries[j + 1] = temp;
                }
            }
        }
    }

    entries(): string[][] {
        const res: string[][] = [];
        for (let i = 0; i < this._entries.length; i++) {
            res.push([this._entries[i].name, this._entries[i].value]);
        }
        return res;
    }

    keys(): string[] {
        const res: string[] = [];
        for (let i = 0; i < this._entries.length; i++) {
            res.push(this._entries[i].name);
        }
        return res;
    }

    values(): string[] {
        const res: string[] = [];
        for (let i = 0; i < this._entries.length; i++) {
            res.push(this._entries[i].value);
        }
        return res;
    }

    forEach(fn: (value: string, name: string, parent: URLSearchParams) => void, thisArg: unknown = null): void {
        for (let i = 0; i < this._entries.length; i++) {
            fn(this._entries[i].value, this._entries[i].name, this);
        }
    }

    toString(): string {
        let res = "";
        for (let i = 0; i < this._entries.length; i++) {
            if (i > 0) {
                res = res + "&";
            }
            res = res + this._entries[i].name + "=" + this._entries[i].value;
        }
        return res;
    }

    get size(): number {
        return this._entries.length;
    }
}

export class URL {
    private _protocol: string = "";
    private _username: string = "";
    private _password: string = "";
    private _hostname: string = "";
    private _port: string = "";
    private _pathname: string = "/";
    private _search: string = "";
    private _hash: string = "";
    private _searchParams: URLSearchParams;

    constructor(input: string, base: string = "") {
        let full = input;
        if (base.length > 0) {
            full = URL._resolve(input, base);
        }
        this._parse(full);
        this._searchParams = new URLSearchParams(this._search);
    }

    private static _resolve(relative: string, base: string): string {
        if (relative.indexOf("http://") === 0 || relative.indexOf("https://") === 0 || relative.indexOf("file://") === 0 || relative.indexOf("ws://") === 0 || relative.indexOf("wss://") === 0) {
            return relative;
        }
        const b = new URL(base);
        if (relative.indexOf("?") === 0) {
            return b.origin + b.pathname + relative;
        }
        if (relative.indexOf("#") === 0) {
            return b.origin + b.pathname + b.search + relative;
        }
        if (relative.indexOf("/") === 0) {
            return b.origin + relative;
        }
        let basePath = b.pathname;
        const lastSlash = basePath.lastIndexOf("/");
        if (lastSlash >= 0) {
            basePath = basePath.slice(0, lastSlash + 1);
        } else {
            basePath = "/";
        }
        return b.origin + basePath + relative;
    }

    private _parse(urlStr: string): void {
        let rest = urlStr;

        // 1. Hash
        const hashIdx = rest.indexOf("#");
        if (hashIdx >= 0) {
            this._hash = rest.slice(hashIdx, rest.length);
            rest = rest.slice(0, hashIdx);
        } else {
            this._hash = "";
        }

        // 2. Search
        const searchIdx = rest.indexOf("?");
        if (searchIdx >= 0) {
            this._search = rest.slice(searchIdx, rest.length);
            rest = rest.slice(0, searchIdx);
        } else {
            this._search = "";
        }

        // 3. Scheme / Protocol
        const schemeIdx = rest.indexOf("://");
        if (schemeIdx >= 0) {
            this._protocol = rest.slice(0, schemeIdx + 1); // "http:"
            rest = rest.slice(schemeIdx + 3, rest.length);
        } else {
            const colonIdx = rest.indexOf(":");
            if (colonIdx >= 0) {
                this._protocol = rest.slice(0, colonIdx + 1);
                rest = rest.slice(colonIdx + 1, rest.length);
            } else {
                this._protocol = "";
            }
        }

        // 4. Authority & Path
        const slashIdx = rest.indexOf("/");
        let authority = rest;
        if (slashIdx >= 0) {
            authority = rest.slice(0, slashIdx);
            this._pathname = rest.slice(slashIdx, rest.length);
        } else {
            this._pathname = "/";
        }

        // 5. Userinfo in Authority
        const atIdx = authority.indexOf("@");
        if (atIdx >= 0) {
            const userinfo = authority.slice(0, atIdx);
            authority = authority.slice(atIdx + 1, authority.length);
            const userColon = userinfo.indexOf(":");
            if (userColon >= 0) {
                this._username = userinfo.slice(0, userColon);
                this._password = userinfo.slice(userColon + 1, userinfo.length);
            } else {
                this._username = userinfo;
                this._password = "";
            }
        } else {
            this._username = "";
            this._password = "";
        }

        // 6. Host & Port
        const portColon = authority.lastIndexOf(":");
        if (portColon >= 0) {
            this._hostname = authority.slice(0, portColon);
            this._port = authority.slice(portColon + 1, authority.length);
        } else {
            this._hostname = authority;
            this._port = "";
        }
    }

    get href(): string {
        let auth = "";
        if (this._username.length > 0) {
            auth = this._username;
            if (this._password.length > 0) {
                auth = auth + ":" + this._password;
            }
            auth = auth + "@";
        }
        let hostStr = this._hostname;
        if (this._port.length > 0) {
            hostStr = hostStr + ":" + this._port;
        }
        let res = this._protocol + "//" + auth + hostStr + this._pathname;
        const spStr = this._searchParams.toString();
        if (spStr.length > 0) {
            res = res + "?" + spStr;
        } else if (this._search.length > 0) {
            res = res + this._search;
        }
        if (this._hash.length > 0) {
            res = res + this._hash;
        }
        return res;
    }

    set href(val: string) {
        this._parse(val);
        this._searchParams = new URLSearchParams(this._search);
    }

    get origin(): string {
        if (this._protocol.length === 0) {
            return "";
        }
        let hostStr = this._hostname;
        if (this._port.length > 0) {
            hostStr = hostStr + ":" + this._port;
        }
        return this._protocol + "//" + hostStr;
    }

    get protocol(): string {
        return this._protocol;
    }

    set protocol(val: string) {
        let p = val;
        if (p.lastIndexOf(":") !== p.length - 1) {
            p = p + ":";
        }
        this._protocol = p;
    }

    get host(): string {
        if (this._port.length > 0) {
            return this._hostname + ":" + this._port;
        }
        return this._hostname;
    }

    set host(val: string) {
        const colonIdx = val.lastIndexOf(":");
        if (colonIdx >= 0) {
            this._hostname = val.slice(0, colonIdx);
            this._port = val.slice(colonIdx + 1, val.length);
        } else {
            this._hostname = val;
            this._port = "";
        }
    }

    get hostname(): string {
        return this._hostname;
    }

    set hostname(val: string) {
        this._hostname = val;
    }

    get port(): string {
        return this._port;
    }

    set port(val: string) {
        this._port = val;
    }

    get pathname(): string {
        return this._pathname;
    }

    set pathname(val: string) {
        if (val.indexOf("/") !== 0) {
            this._pathname = "/" + val;
        } else {
            this._pathname = val;
        }
    }

    get search(): string {
        const spStr = this._searchParams.toString();
        if (spStr.length > 0) {
            return "?" + spStr;
        }
        return this._search;
    }

    set search(val: string) {
        let s = val;
        if (s.length > 0 && s.indexOf("?") !== 0) {
            s = "?" + s;
        }
        this._search = s;
        this._searchParams = new URLSearchParams(s);
    }

    get hash(): string {
        return this._hash;
    }

    set hash(val: string) {
        let h = val;
        if (h.length > 0 && h.indexOf("#") !== 0) {
            h = "#" + h;
        }
        this._hash = h;
    }

    get username(): string {
        return this._username;
    }

    set username(val: string) {
        this._username = val;
    }

    get password(): string {
        return this._password;
    }

    set password(val: string) {
        this._password = val;
    }

    get searchParams(): URLSearchParams {
        return this._searchParams;
    }

    toString(): string {
        return this.href;
    }

    toJSON(): string {
        return this.href;
    }

    static canParse(url: string, base: string = ""): boolean {
        if (url.length === 0) {
            return false;
        }
        if (url.indexOf("http://") === 0 || url.indexOf("https://") === 0 || url.indexOf("file://") === 0 || url.indexOf("ws://") === 0 || url.indexOf("wss://") === 0) {
            return true;
        }
        if (base.length > 0 && (base.indexOf("http://") === 0 || base.indexOf("https://") === 0 || base.indexOf("file://") === 0)) {
            return true;
        }
        return false;
    }

    static parse(input: string, base: string = ""): URL | null {
        if (URL.canParse(input, base)) {
            return new URL(input, base);
        }
        return null;
    }
}

export class Url {
    auth: string = "";
    hash: string = "";
    host: string = "";
    hostname: string = "";
    href: string = "";
    path: string = "";
    pathname: string = "";
    port: string = "";
    protocol: string = "";
    query: string = "";
    search: string = "";
    slashes: boolean = false;
}

export function parse(urlString: string, parseQueryString: boolean = false, slashesDenoteHost: boolean = false): Url {
    const u = new Url();
    u.href = urlString;
    const urlObj = new URL(urlString, "http://localhost");
    u.protocol = urlObj.protocol;
    u.hostname = urlObj.hostname;
    u.port = urlObj.port;
    u.host = urlObj.host;
    u.pathname = urlObj.pathname;
    u.search = urlObj.search;
    u.hash = urlObj.hash;
    u.path = urlObj.search.length > 0 ? urlObj.pathname + urlObj.search : urlObj.pathname;
    u.auth = urlObj.username.length > 0 ? (urlObj.password.length > 0 ? urlObj.username + ":" + urlObj.password : urlObj.username) : "";
    u.slashes = urlString.indexOf("//") >= 0;
    u.query = u.search.indexOf("?") === 0 ? u.search.slice(1, u.search.length) : u.search;
    return u;
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

export function format(urlObject: URL): string {
    return urlObject.toString();
}

export function resolve(from: string, to: string): string {
    const u = new URL(to, from);
    return u.href;
}

export function fileURLToPath(url: string): string {
    let str = url;
    if (str.indexOf("file://") === 0) {
        str = str.slice(7, str.length);
    }
    return str;
}

export function pathToFileURL(path: string): URL {
    let p = path;
    if (p.indexOf("/") !== 0) {
        p = "/" + p;
    }
    return new URL("file://" + p);
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

export function urlToHttpOptions(url: URL): HttpOptionsResult {
    return {
        protocol: url.protocol,
        hostname: url.hostname,
        hash: url.hash,
        search: url.search,
        pathname: url.pathname,
        path: url.pathname + url.search,
        href: url.href,
        port: url.port.length > 0 ? Number(url.port) : null,
        auth: url.username.length > 0 ? (url.password.length > 0 ? url.username + ":" + url.password : url.username) : null
    };
}
