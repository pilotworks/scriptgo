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
        resolved = path + "/" + resolved;
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

export const posix = {
    join,
    dirname,
    basename,
    extname,
    isAbsolute,
    normalize,
    resolve,
    relative,
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
    sep,
    delimiter,
    posix,
    win32,
};
