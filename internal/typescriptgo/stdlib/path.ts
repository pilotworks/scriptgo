// ScriptGo Standard Library: node:path

export const sep = "/";
export const delimiter = ":";

function trimTrailingSeparators(path: string): string {
    return path.length === 0 ? path : (path.lastIndexOf("/") === path.length - 1 ? path.slice(0, path.length - 1) : path);
}

function joinPair(first: string, second: string): string {
    const left: string = trimTrailingSeparators(first);
    const secondStart: number = second.lastIndexOf("/") === 0 ? 1 : 0;
    const right: string = trimTrailingSeparators(second.slice(secondStart, second.length));
    return left.length === 0 ? (right.length === 0 ? "." : right) : (right.length === 0 ? left : left + "/" + right);
}

function joinRest(current: string, rest: string[], index: number): string {
    if (index === rest.length) {
        return current;
    }
    return joinRest(joinPair(current, rest[index]), rest, index + 1);
}

export function join(first: string, second: string = "", ...rest: string[]): string {
    if (second === "" && rest.length === 0) {
        return normalize(first);
    }
    return normalize(joinRest(joinPair(first, second), rest, 0));
}

export function dirname(path: string): string {
    const trimmed: string = trimTrailingSeparators(path);
    const index: number = trimmed.lastIndexOf("/");
    return trimmed.length === 0 ? "/" : (index < 0 ? "." : (index === 0 ? "/" : trimmed.slice(0, index)));
}

export function basename(path: string, ext?: string): string {
    const trimmed: string = trimTrailingSeparators(path);
    const index: number = trimmed.lastIndexOf("/");
    const base: string = index < 0 ? trimmed : trimmed.slice(index + 1, trimmed.length);
    if (ext !== undefined && ext.length > 0 && base.endsWith(ext)) {
        return base.slice(0, base.length - ext.length);
    }
    return base;
}

export function extname(path: string): string {
    const name: string = basename(path);
    const index: number = name.lastIndexOf(".");
    return index <= 0 ? "" : name.slice(index, name.length);
}

export function isAbsolute(path: string): boolean {
    return path.length > 0 && path.charCodeAt(0) === 47; // '/'
}

export function normalize(path: string): string {
    if (path.length === 0) return ".";
    const isAbs = isAbsolute(path);
    const trailingSlash = path.charCodeAt(path.length - 1) === 47;
    const segments = path.split("/");
    const result: string[] = [];

    for (let i = 0; i < segments.length; i++) {
        const seg = segments[i];
        if (seg === "" || seg === ".") continue;
        if (seg === "..") {
            if (result.length > 0 && result[result.length - 1] !== "..") {
                result.pop();
            } else if (!isAbs) {
                result.push("..");
            }
        } else {
            result.push(seg);
        }
    }

    let joined = result.join("/");
    if (isAbs) {
        joined = "/" + joined;
    } else if (joined.length === 0) {
        joined = ".";
    }
    if (trailingSlash && !joined.endsWith("/")) {
        joined = joined + "/";
    }
    return joined;
}

export function resolve(...paths: string[]): string {
    let resolved = "";
    let resolvedAbsolute = false;

    for (let i = paths.length - 1; i >= 0 && !resolvedAbsolute; i--) {
        const path = paths[i];
        if (path.length === 0) continue;
        resolved = resolved.length === 0 ? path : path + "/" + resolved;
        resolvedAbsolute = isAbsolute(path);
    }

    if (!resolvedAbsolute) {
        resolved = "/" + resolved;
    }

    return normalize(resolved);
}

export function relative(from: string, to: string): string {
    const fromAbs = resolve(from);
    const toAbs = resolve(to);
    if (fromAbs === toAbs) return "";

    const fromParts = fromAbs.split("/").filter((s) => s.length > 0);
    const toParts = toAbs.split("/").filter((s) => s.length > 0);

    let same = 0;
    while (same < fromParts.length && same < toParts.length && fromParts[same] === toParts[same]) {
        same++;
    }

    const result: string[] = [];
    for (let i = same; i < fromParts.length; i++) {
        result.push("..");
    }
    for (let i = same; i < toParts.length; i++) {
        result.push(toParts[i]);
    }
    return result.join("/");
}

export class ParsedPath {
    root: string;
    dir: string;
    base: string;
    ext: string;
    name: string;

    constructor(root: string, dir: string, base: string, ext: string, name: string) {
        this.root = root;
        this.dir = dir;
        this.base = base;
        this.ext = ext;
        this.name = name;
    }
}

export interface FormatInputPathObject {
    root?: string;
    dir?: string;
    base?: string;
    ext?: string;
    name?: string;
}

export function parse(path: string): ParsedPath {
    const isAbs = isAbsolute(path);
    const root = isAbs ? "/" : "";
    const dir = dirname(path);
    const base = basename(path);
    const ext = extname(path);
    const name = ext.length > 0 ? base.slice(0, base.length - ext.length) : base;
    return new ParsedPath(root, dir, base, ext, name);
}

export function format(pathObject: FormatInputPathObject): string {
    const dir = pathObject.dir || pathObject.root || "";
    const base = pathObject.base || ((pathObject.name || "") + (pathObject.ext || ""));
    if (dir.length === 0) {
        return base;
    }
    if (dir.endsWith("/")) {
        return dir + base;
    }
    return dir + "/" + base;
}

export function toNamespacedPath(path: string): string {
    return path;
}

export function win32ToNamespacedPath(path: string): string {
    if (typeof path !== "string" || path.length === 0) return path;
    if (path.startsWith("\\\\?\\") || path.startsWith("\\\\.\\")) {
        return path;
    }
    const normalized = path.replace(/\//g, "\\");
    if (normalized.length >= 2) {
        const c0 = normalized.charCodeAt(0);
        const isAlpha = (c0 >= 65 && c0 <= 90) || (c0 >= 97 && c0 <= 122);
        if (isAlpha && normalized.charCodeAt(1) === 58) {
            if (normalized.length === 2) {
                return "\\\\?\\" + normalized + "\\";
            }
            if (normalized.charCodeAt(2) === 92) {
                return "\\\\?\\" + normalized;
            }
        } else if (normalized.startsWith("\\\\")) {
            return "\\\\?\\UNC\\" + normalized.slice(2);
        }
    }
    return path;
}

export function matchesGlob(path: string, pattern: string): boolean {
    if (typeof path !== "string" || typeof pattern !== "string") {
        throw new TypeError("path and pattern must be strings");
    }
    if (pattern === "*") {
        return path.indexOf("/") === -1;
    }
    if (pattern === "**") {
        return true;
    }

    let rx = "^";
    let inClass = false;
    for (let i = 0; i < pattern.length; i++) {
        const c = pattern[i];
        if (inClass) {
            if (c === "]") {
                inClass = false;
                rx += "]";
            } else if (c === "\\") {
                rx += "\\\\";
            } else {
                rx += c;
            }
            continue;
        }
        if (c === "*") {
            if (i + 1 < pattern.length && pattern[i + 1] === "*") {
                i++;
                if (i + 1 < pattern.length && pattern[i + 1] === "/") {
                    i++;
                    rx += "(?:.*\\/)?";
                } else {
                    rx += ".*";
                }
            } else {
                rx += "[^\\/]*";
            }
        } else if (c === "?") {
            rx += "[^\\/]";
        } else if (c === "[") {
            inClass = true;
            rx += "[";
        } else if (c === "." || c === "(" || c === ")" || c === "+" || c === "^" || c === "$" || c === "{" || c === "}" || c === "|") {
            rx += "\\" + c;
        } else if (c === "\\") {
            rx += "\\\\";
        } else {
            rx += c;
        }
    }
    rx += "$";
    const reg = new RegExp(rx);
    return reg.test(path);
}

export const posix = {
    join,
    dirname,
    basename,
    extname,
    isAbsolute,
    normalize,
    resolve,
    relative,
    parse,
    format,
    toNamespacedPath,
    matchesGlob,
    sep: "/",
    delimiter: ":",
};

export const win32 = {
    join,
    dirname,
    basename,
    extname,
    isAbsolute,
    normalize,
    resolve,
    relative,
    parse,
    format,
    toNamespacedPath: win32ToNamespacedPath,
    matchesGlob,
    sep: "\\",
    delimiter: ";",
};

export default {
    join,
    dirname,
    basename,
    extname,
    isAbsolute,
    normalize,
    resolve,
    relative,
    parse,
    format,
    toNamespacedPath,
    matchesGlob,
    sep,
    delimiter,
    posix,
    win32,
};

