// WHATWG URLPattern Standard API Implementation for ScriptGo

export interface URLPatternComponentResult {
    input: string;
    groups: Record<string, string | undefined>;
}

export interface URLPatternResult {
    inputs: (string | URLPatternInput)[];
    protocol: URLPatternComponentResult;
    username: URLPatternComponentResult;
    password: URLPatternComponentResult;
    hostname: URLPatternComponentResult;
    port: URLPatternComponentResult;
    pathname: URLPatternComponentResult;
    search: URLPatternComponentResult;
    hash: URLPatternComponentResult;
}

export interface URLPatternInput {
    protocol?: string;
    username?: string;
    password?: string;
    hostname?: string;
    port?: string;
    pathname?: string;
    search?: string;
    hash?: string;
    baseURL?: string;
}

export interface URLPatternOptions {
    ignoreCase?: boolean;
}

interface ParsedURLParts {
    protocol: string;
    username: string;
    password: string;
    hostname: string;
    port: string;
    pathname: string;
    search: string;
    hash: string;
}

function parseURLParts(urlStr: string): ParsedURLParts {
    let protocol = "";
    let username = "";
    let password = "";
    let hostname = "";
    let port = "";
    let pathname = "";
    let search = "";
    let hash = "";

    let rest = urlStr;
    const protoIdx = rest.indexOf("://");
    if (protoIdx !== -1) {
        protocol = rest.substring(0, protoIdx);
        rest = rest.substring(protoIdx + 3);

        let pathStart = rest.indexOf("/");
        const qStart = rest.indexOf("?");
        const hStart = rest.indexOf("#");
        if (pathStart === -1 || (qStart !== -1 && qStart < pathStart)) {
            pathStart = qStart;
        }
        if (pathStart === -1 || (hStart !== -1 && hStart < pathStart)) {
            pathStart = hStart;
        }

        let authority = rest;
        if (pathStart !== -1) {
            authority = rest.substring(0, pathStart);
            rest = rest.substring(pathStart);
        } else {
            authority = rest;
            rest = "/";
        }

        const atIdx = authority.indexOf("@");
        if (atIdx !== -1) {
            const userinfo = authority.substring(0, atIdx);
            authority = authority.substring(atIdx + 1);
            const colonIdx = userinfo.indexOf(":");
            if (colonIdx !== -1) {
                username = userinfo.substring(0, colonIdx);
                password = userinfo.substring(colonIdx + 1);
            } else {
                username = userinfo;
            }
        }

        const colonIdx = authority.lastIndexOf(":");
        if (colonIdx !== -1) {
            hostname = authority.substring(0, colonIdx);
            port = authority.substring(colonIdx + 1);
        } else {
            hostname = authority;
        }
    }

    const hashIdx = rest.indexOf("#");
    if (hashIdx !== -1) {
        hash = rest.substring(hashIdx + 1);
        rest = rest.substring(0, hashIdx);
    }

    const searchIdx = rest.indexOf("?");
    if (searchIdx !== -1) {
        search = rest.substring(searchIdx + 1);
        rest = rest.substring(0, searchIdx);
    }

    pathname = rest;
    if (pathname.length === 0) {
        pathname = "/";
    }

    return { protocol, username, password, hostname, port, pathname, search, hash };
}

function isIdentChar(ch: string): boolean {
    const code = ch.charCodeAt(0);
    return (code >= 97 && code <= 122) || (code >= 65 && code <= 90) || (code >= 48 && code <= 57) || code === 95;
}

function isRegexSpecial(ch: string): boolean {
    return ch === "." || ch === "+" || ch === "?" || ch === "^" || ch === "$" ||
        ch === "(" || ch === ")" || ch === "[" || ch === "]" || ch === "{" ||
        ch === "}" || ch === "|";
}

class ComponentPattern {
    raw: string;
    regex: RegExp;
    groupNames: string[];
    hasCustomRegex: boolean;
    isWildcard: boolean;

    constructor(rawPattern: string, isPath: boolean, ignoreCase: boolean = false) {
        this.raw = rawPattern;
        this.groupNames = [];
        this.hasCustomRegex = false;
        this.isWildcard = false;

        if (rawPattern === "*") {
            this.isWildcard = true;
            this.groupNames.push("0");
            this.regex = RegExp("^(.*)$", ignoreCase ? "i" : "");
            return;
        }

        if (rawPattern === "") {
            this.regex = RegExp("^$", ignoreCase ? "i" : "");
            return;
        }

        let regexStr = "^";
        let anonIdx = 0;
        let i = 0;
        const n = rawPattern.length;

        while (i < n) {
            const ch = rawPattern.charAt(i);
            if (ch === "\\") {
                if (i + 1 < n) {
                    regexStr += "\\" + rawPattern.charAt(i + 1);
                    i += 2;
                } else {
                    regexStr += "\\\\";
                    i++;
                }
                continue;
            }

            if (ch === "*") {
                regexStr += "(.*)";
                this.groupNames.push(String(anonIdx));
                anonIdx = anonIdx + 1;
                i++;
                continue;
            }

            if (ch === "{") {
                const closeIdx = rawPattern.indexOf("}", i + 1);
                if (closeIdx !== -1) {
                    const inner = rawPattern.substring(i + 1, closeIdx);
                    const isOptional = (closeIdx + 1 < n && rawPattern.charAt(closeIdx + 1) === "?");
                    let subRegex = "";
                    let j = 0;
                    while (j < inner.length) {
                        const subCh = inner.charAt(j);
                        if (subCh === ":") {
                            let nStart = j + 1;
                            let nEnd = nStart;
                            while (nEnd < inner.length && isIdentChar(inner.charAt(nEnd))) {
                                nEnd++;
                            }
                            const pName = inner.substring(nStart, nEnd);
                            j = nEnd;
                            const dReg = isPath ? "[^/]+" : "[^/?#&]+";
                            subRegex += "(" + dReg + ")";
                            this.groupNames.push(pName);
                            continue;
                        }
                        if (isRegexSpecial(subCh)) {
                            subRegex += "\\" + subCh;
                        } else {
                            subRegex += subCh;
                        }
                        j++;
                    }
                    if (isOptional) {
                        regexStr += "(?:" + subRegex + ")?";
                        i = closeIdx + 2;
                    } else {
                        regexStr += "(?:" + subRegex + ")";
                        i = closeIdx + 1;
                    }
                    continue;
                }
            }

            if (ch === ":") {
                let nameStart = i + 1;
                let nameEnd = nameStart;
                while (nameEnd < n && isIdentChar(rawPattern.charAt(nameEnd))) {
                    nameEnd++;
                }
                const paramName = rawPattern.substring(nameStart, nameEnd);
                i = nameEnd;

                let customReg = "";
                if (i < n && rawPattern.charAt(i) === "(") {
                    let parenDepth = 1;
                    let parenEnd = i + 1;
                    while (parenEnd < n && parenDepth > 0) {
                        if (rawPattern.charAt(parenEnd) === "\\") {
                            parenEnd += 2;
                        } else if (rawPattern.charAt(parenEnd) === "(") {
                            parenDepth++;
                            parenEnd++;
                        } else if (rawPattern.charAt(parenEnd) === ")") {
                            parenDepth--;
                            parenEnd++;
                        } else {
                            parenEnd++;
                        }
                    }
                    customReg = rawPattern.substring(i + 1, parenEnd - 1);
                    this.hasCustomRegex = true;
                    i = parenEnd;
                }

                let modifier = "";
                if (i < n && (rawPattern.charAt(i) === "?" || rawPattern.charAt(i) === "+" || rawPattern.charAt(i) === "*")) {
                    modifier = rawPattern.charAt(i);
                    i++;
                }

                const innerRegex = (customReg.length > 0)
                    ? customReg
                    : (isPath ? "[^/]+" : "[^/?#&]+");

                if (modifier === "?") {
                    if (regexStr.endsWith("/")) {
                        regexStr = regexStr.substring(0, regexStr.length - 1) + "(?:/(" + innerRegex + "))?";
                    } else {
                        regexStr += "(?:(" + innerRegex + "))?";
                    }
                } else if (modifier === "+") {
                    regexStr += "((?:" + innerRegex + ")+)";
                } else if (modifier === "*") {
                    regexStr += "((?:" + innerRegex + ")*)";
                } else {
                    regexStr += "(" + innerRegex + ")";
                }

                this.groupNames.push(paramName);
                continue;
            }

            if (isRegexSpecial(ch)) {
                regexStr += "\\" + ch;
            } else {
                regexStr += ch;
            }
            i++;
        }

        regexStr += "$";
        this.regex = RegExp(regexStr, ignoreCase ? "i" : "");
    }

    exec(input: string): { matched: boolean; groups: Record<string, string | undefined> } {
        if (this.isWildcard) {
            const groups: Record<string, string | undefined> = {};
            groups["0"] = input;
            return { matched: true, groups: groups };
        }
        if (!this.regex.test(input)) {
            const emptyGroups: Record<string, string | undefined> = {};
            return { matched: false, groups: emptyGroups };
        }
        const m = this.regex.exec(input) as string[];
        const groups: Record<string, string | undefined> = {};
        for (let i = 0; i < this.groupNames.length; i++) {
            const gName = this.groupNames[i];
            let val: string | undefined = undefined;
            if (m !== null && i + 1 < m.length && m[i + 1] !== undefined && m[i + 1] !== null) {
                val = m[i + 1];
            }
            groups[gName] = val;
        }
        return { matched: true, groups: groups };
    }
}

export class URLPattern {
    readonly protocol: string;
    readonly username: string;
    readonly password: string;
    readonly hostname: string;
    readonly port: string;
    readonly pathname: string;
    readonly search: string;
    readonly hash: string;
    readonly hasRegExpGroups: boolean;

    private _protocolPattern: ComponentPattern;
    private _usernamePattern: ComponentPattern;
    private _passwordPattern: ComponentPattern;
    private _hostnamePattern: ComponentPattern;
    private _portPattern: ComponentPattern;
    private _pathnamePattern: ComponentPattern;
    private _searchPattern: ComponentPattern;
    private _hashPattern: ComponentPattern;

    constructor(input?: unknown, baseURLOrOptions?: unknown, optionsArg?: URLPatternOptions) {
        let baseURL: string | undefined = undefined;
        let options: URLPatternOptions | undefined = optionsArg;

        if (typeof baseURLOrOptions === "string") {
            baseURL = baseURLOrOptions as string;
        } else if (baseURLOrOptions && typeof baseURLOrOptions === "object") {
            options = baseURLOrOptions as URLPatternOptions;
        }

        const ignoreCase = (options !== undefined && options !== null && options.ignoreCase === true);

        let proto = "*";
        let user = "*";
        let pass = "*";
        let host = "*";
        let port = "*";
        let path = "*";
        let search = "*";
        let hash = "*";

        if (typeof input === "string") {
            const str = input as string;
            let fullStr = str;
            if (baseURL !== undefined && baseURL !== null && baseURL.length > 0) {
                if (str.indexOf("://") === -1) {
                    const baseParts = parseURLParts(baseURL);
                    proto = baseParts.protocol.length > 0 ? baseParts.protocol : "*";
                    user = baseParts.username.length > 0 ? baseParts.username : "*";
                    pass = baseParts.password.length > 0 ? baseParts.password : "*";
                    host = baseParts.hostname.length > 0 ? baseParts.hostname : "*";
                    port = baseParts.port.length > 0 ? baseParts.port : "*";

                    const relParts = parseURLParts(str);
                    path = relParts.pathname;
                    search = str.indexOf("?") !== -1 ? relParts.search : "";
                    hash = str.indexOf("#") !== -1 ? relParts.hash : "";
                    fullStr = "";
                }
            }
            if (fullStr.length > 0) {
                if (fullStr.indexOf("://") !== -1) {
                    const parts = parseURLParts(fullStr);
                    proto = parts.protocol;
                    user = parts.username;
                    pass = parts.password;
                    host = parts.hostname;
                    port = parts.port;
                    path = parts.pathname;
                    search = fullStr.indexOf("?") !== -1 ? parts.search : "";
                    hash = fullStr.indexOf("#") !== -1 ? parts.hash : "";
                } else {
                    throw new TypeError("Failed to construct 'URLPattern': Invalid URL pattern.");
                }
            }
        } else if (input && typeof input === "object") {
            const obj = input as URLPatternInput;
            if (obj.baseURL !== undefined && obj.baseURL !== null && obj.baseURL.length > 0) {
                const baseParts = parseURLParts(obj.baseURL);
                if (obj.protocol !== undefined) proto = obj.protocol; else if (baseParts.protocol.length > 0) proto = baseParts.protocol;
                if (obj.username !== undefined) user = obj.username; else if (baseParts.username.length > 0) user = baseParts.username;
                if (obj.password !== undefined) pass = obj.password; else if (baseParts.password.length > 0) pass = baseParts.password;
                if (obj.hostname !== undefined) host = obj.hostname; else if (baseParts.hostname.length > 0) host = baseParts.hostname;
                if (obj.port !== undefined) port = obj.port; else if (baseParts.port.length > 0) port = baseParts.port;
                if (obj.pathname !== undefined) path = obj.pathname; else if (baseParts.pathname.length > 0) path = baseParts.pathname;
                if (obj.search !== undefined) search = obj.search; else if (baseParts.search.length > 0) search = baseParts.search;
                if (obj.hash !== undefined) hash = obj.hash; else if (baseParts.hash.length > 0) hash = baseParts.hash;
            } else {
                if (obj.protocol !== undefined) proto = obj.protocol;
                if (obj.username !== undefined) user = obj.username;
                if (obj.password !== undefined) pass = obj.password;
                if (obj.hostname !== undefined) host = obj.hostname;
                if (obj.port !== undefined) port = obj.port;
                if (obj.pathname !== undefined) path = obj.pathname;
                if (obj.search !== undefined) search = obj.search;
                if (obj.hash !== undefined) hash = obj.hash;
            }
        }

        if (proto.endsWith(":")) {
            proto = proto.substring(0, proto.length - 1);
        }
        if (search.startsWith("?")) {
            search = search.substring(1);
        }
        if (hash.startsWith("#")) {
            hash = hash.substring(1);
        }

        this.protocol = proto;
        this.username = user;
        this.password = pass;
        this.hostname = host;
        this.port = port;
        this.pathname = path;
        this.search = search;
        this.hash = hash;

        this._protocolPattern = new ComponentPattern(proto, false, ignoreCase);
        this._usernamePattern = new ComponentPattern(user, false, ignoreCase);
        this._passwordPattern = new ComponentPattern(pass, false, ignoreCase);
        this._hostnamePattern = new ComponentPattern(host, false, ignoreCase);
        this._portPattern = new ComponentPattern(port, false, ignoreCase);
        this._pathnamePattern = new ComponentPattern(path, true, ignoreCase);
        this._searchPattern = new ComponentPattern(search, false, ignoreCase);
        this._hashPattern = new ComponentPattern(hash, false, ignoreCase);

        this.hasRegExpGroups = (
            this._protocolPattern.hasCustomRegex ||
            this._usernamePattern.hasCustomRegex ||
            this._passwordPattern.hasCustomRegex ||
            this._hostnamePattern.hasCustomRegex ||
            this._portPattern.hasCustomRegex ||
            this._pathnamePattern.hasCustomRegex ||
            this._searchPattern.hasCustomRegex ||
            this._hashPattern.hasCustomRegex
        );
    }

    test(input?: unknown, baseURL?: string): boolean {
        return this.exec(input, baseURL) !== null;
    }

    exec(input?: unknown, baseURL?: string): URLPatternResult | null {
        let candidate: ParsedURLParts = {
            protocol: "",
            username: "",
            password: "",
            hostname: "",
            port: "",
            pathname: "",
            search: "",
            hash: ""
        };

        const inputsList: (string | URLPatternInput)[] = [];

        if (typeof input === "string") {
            const str = input as string;
            if (baseURL !== undefined && baseURL !== null && baseURL.length > 0) {
                inputsList.push(str);
                inputsList.push(baseURL);
                if (str.indexOf("://") === -1) {
                    const baseParts = parseURLParts(baseURL);
                    const relParts = parseURLParts(str);
                    candidate = {
                        protocol: baseParts.protocol,
                        username: baseParts.username,
                        password: baseParts.password,
                        hostname: baseParts.hostname,
                        port: baseParts.port,
                        pathname: relParts.pathname,
                        search: str.indexOf("?") !== -1 ? relParts.search : "",
                        hash: str.indexOf("#") !== -1 ? relParts.hash : ""
                    };
                } else {
                    candidate = parseURLParts(str);
                }
            } else {
                inputsList.push(str);
                candidate = parseURLParts(str);
                if (str.indexOf("://") === -1) {
                    candidate.protocol = "";
                    candidate.hostname = "";
                    candidate.port = "";
                    candidate.username = "";
                    candidate.password = "";
                }
            }
        } else if (input && typeof input === "object") {
            const obj = input as URLPatternInput;
            inputsList.push(obj);
            if (obj.baseURL !== undefined && obj.baseURL !== null && obj.baseURL.length > 0) {
                const baseParts = parseURLParts(obj.baseURL);
                candidate = {
                    protocol: obj.protocol !== undefined ? obj.protocol : baseParts.protocol,
                    username: obj.username !== undefined ? obj.username : baseParts.username,
                    password: obj.password !== undefined ? obj.password : baseParts.password,
                    hostname: obj.hostname !== undefined ? obj.hostname : baseParts.hostname,
                    port: obj.port !== undefined ? obj.port : baseParts.port,
                    pathname: obj.pathname !== undefined ? obj.pathname : baseParts.pathname,
                    search: obj.search !== undefined ? obj.search : baseParts.search,
                    hash: obj.hash !== undefined ? obj.hash : baseParts.hash
                };
            } else {
                candidate = {
                    protocol: obj.protocol !== undefined ? obj.protocol : "",
                    username: obj.username !== undefined ? obj.username : "",
                    password: obj.password !== undefined ? obj.password : "",
                    hostname: obj.hostname !== undefined ? obj.hostname : "",
                    port: obj.port !== undefined ? obj.port : "",
                    pathname: obj.pathname !== undefined ? obj.pathname : "",
                    search: obj.search !== undefined ? obj.search : "",
                    hash: obj.hash !== undefined ? obj.hash : ""
                };
            }
        } else {
            inputsList.push({
                protocol: undefined,
                username: undefined,
                password: undefined,
                hostname: undefined,
                port: undefined,
                pathname: undefined,
                search: undefined,
                hash: undefined,
                baseURL: undefined
            });
        }

        if (candidate.protocol.endsWith(":")) {
            candidate.protocol = candidate.protocol.substring(0, candidate.protocol.length - 1);
        }
        if (candidate.search.startsWith("?")) {
            candidate.search = candidate.search.substring(1);
        }
        if (candidate.hash.startsWith("#")) {
            candidate.hash = candidate.hash.substring(1);
        }

        const pM = this._protocolPattern.exec(candidate.protocol);
        if (!pM.matched) return null;
        const uM = this._usernamePattern.exec(candidate.username);
        if (!uM.matched) return null;
        const pwM = this._passwordPattern.exec(candidate.password);
        if (!pwM.matched) return null;
        const hM = this._hostnamePattern.exec(candidate.hostname);
        if (!hM.matched) return null;
        const ptM = this._portPattern.exec(candidate.port);
        if (!ptM.matched) return null;
        const pathM = this._pathnamePattern.exec(candidate.pathname);
        if (!pathM.matched) return null;
        const sM = this._searchPattern.exec(candidate.search);
        if (!sM.matched) return null;
        const hsM = this._hashPattern.exec(candidate.hash);
        if (!hsM.matched) return null;

        return {
            inputs: inputsList,
            protocol: { input: candidate.protocol, groups: pM.groups },
            username: { input: candidate.username, groups: uM.groups },
            password: { input: candidate.password, groups: pwM.groups },
            hostname: { input: candidate.hostname, groups: hM.groups },
            port: { input: candidate.port, groups: ptM.groups },
            pathname: { input: candidate.pathname, groups: pathM.groups },
            search: { input: candidate.search, groups: sM.groups },
            hash: { input: candidate.hash, groups: hsM.groups }
        };
    }
}

export default {
    URLPattern
};
