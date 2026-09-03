// ScriptGo Standard Library: node:dns

declare namespace __scriptgo {
    function dnsLookup(hostname: string, family?: number): { address: string; family: number };
    function dnsLookupService(address: string, port: number): { hostname: string; service: string };
    function dnsReverse(ip: string): string[];
    function dnsResolveStrings(hostname: string, rrtype: string): string[];
}

export const NODATA = "ENODATA";
export const FORMERR = "EFORMERR";
export const SERVFAIL = "ESERVFAIL";
export const NOTFOUND = "ENOTFOUND";
export const NOTIMP = "ENOTIMP";
export const REFUSED = "EREFUSED";
export const BADQUERY = "EBADQUERY";
export const BADNAME = "EBADNAME";
export const BADFAMILY = "EBADFAMILY";
export const BADRESP = "EBADRESP";
export const CONNREFUSED = "ECONNREFUSED";
export const TIMEOUT = "ETIMEOUT";
export const EOF = "EOF";
export const FILE = "EFILE";
export const NOMEM = "ENOMEM";
export const DESTRUCTION = "EDESTRUCTION";
export const BADSTR = "EBADSTR";
export const BADFLAGS = "EBADFLAGS";
export const NONAME = "ENONAME";
export const BADHINTS = "EBADHINTS";
export const NOTINITIALIZED = "ENOTINITIALIZED";
export const LOADIPHLPAPI = "ELOADIPHLPAPI";
export const ADDRGETNETWORKPARAMS = "EADDRGETNETWORKPARAMS";
export const CANCELLED = "ECANCELLED";

export class LookupAddress {
    address: string = "";
    family: number = 4;

    constructor(address: string, family: number) {
        this.address = address;
        this.family = family;
    }
}

let _dnsResultOrder: string = "verbatim";

export function getDefaultResultOrder(): string {
    return _dnsResultOrder;
}

export function setDefaultResultOrder(order: string): void {
    if (order === "ipv4first" || order === "verbatim") {
        _dnsResultOrder = order;
    }
}

export function lookup(hostname: string, optionsOrCallback?: unknown, callback?: unknown): void {
    let cb: Function = () => {};
    let isAll = false;
    let reqFamily = 0;

    if (typeof optionsOrCallback === "function") {
        cb = optionsOrCallback as Function;
    } else if (typeof callback === "function") {
        cb = callback as Function;
        if (typeof optionsOrCallback === "object" && optionsOrCallback !== null) {
            const opt = optionsOrCallback as { all?: boolean; family?: number };
            if (opt.all) isAll = true;
            if (typeof opt.family === "number") reqFamily = opt.family as number;
        } else if (typeof optionsOrCallback === "number") {
            reqFamily = optionsOrCallback as number;
        }
    } else {
        return;
    }

    if (reqFamily === 0 && _dnsResultOrder === "ipv4first") {
        reqFamily = 4;
    }

    try {
        const res = __scriptgo.dnsLookup(hostname, reqFamily);
        if (isAll) {
            const addresses: LookupAddress[] = [new LookupAddress(res.address, res.family)];
            (cb as (err: unknown, addresses: LookupAddress[]) => void)(null, addresses);
        } else {
            (cb as (err: unknown, address: string, family: number) => void)(null, res.address, res.family);
        }
    } catch (err) {
        if (isAll) {
            (cb as (err: unknown, addresses: LookupAddress[]) => void)(err, []);
        } else {
            (cb as (err: unknown, address: string, family: number) => void)(err, "", 0);
        }
    }
}

export function lookupService(address: string, port: number, callback: (err: unknown, hostname: string, service: string) => void): void {
    try {
        const res = __scriptgo.dnsLookupService(address, port);
        callback(null, res.hostname, res.service);
    } catch (err) {
        callback(err, "", "");
    }
}

export function resolve(hostname: string, rrtypeOrCallback?: unknown, callback?: unknown): void {
    let cb: Function = () => {};
    let rrtype = "A";

    if (typeof rrtypeOrCallback === "function") {
        cb = rrtypeOrCallback as Function;
    } else if (typeof callback === "function") {
        cb = callback as Function;
        if (typeof rrtypeOrCallback === "string") {
            rrtype = rrtypeOrCallback as string;
        }
    } else {
        return;
    }
    try {
        const addrs = __scriptgo.dnsResolveStrings(hostname, rrtype);
        (cb as (err: unknown, addresses: string[]) => void)(null, addrs);
    } catch (err) {
        (cb as (err: unknown, addresses: string[]) => void)(err, []);
    }
}

export function resolve4(hostname: string, optionsOrCallback?: unknown, callback?: unknown): void {
    let cb: Function = () => {};
    if (typeof optionsOrCallback === "function") {
        cb = optionsOrCallback as Function;
    } else if (typeof callback === "function") {
        cb = callback as Function;
    } else {
        return;
    }
    try {
        const addrs = __scriptgo.dnsResolveStrings(hostname, "A");
        (cb as (err: unknown, addresses: string[]) => void)(null, addrs);
    } catch (err) {
        (cb as (err: unknown, addresses: string[]) => void)(err, []);
    }
}

export function resolve6(hostname: string, optionsOrCallback?: unknown, callback?: unknown): void {
    let cb: Function = () => {};
    if (typeof optionsOrCallback === "function") {
        cb = optionsOrCallback as Function;
    } else if (typeof callback === "function") {
        cb = callback as Function;
    } else {
        return;
    }
    try {
        const addrs = __scriptgo.dnsResolveStrings(hostname, "AAAA");
        (cb as (err: unknown, addresses: string[]) => void)(null, addrs);
    } catch (err) {
        (cb as (err: unknown, addresses: string[]) => void)(err, []);
    }
}
export function resolveCname(hostname: string, callback: (err: unknown, addresses: string[]) => void): void {
    try {
        const addrs = __scriptgo.dnsResolveStrings(hostname, "CNAME");
        callback(null, addrs);
    } catch (err) {
        callback(err, []);
    }
}


export function resolveNs(hostname: string, callback: (err: unknown, addresses: string[]) => void): void {
    try {
        const addrs = __scriptgo.dnsResolveStrings(hostname, "NS");
        callback(null, addrs);
    } catch (err) {
        callback(err, []);
    }
}

export function resolvePtr(hostname: string, callback: (err: unknown, addresses: string[]) => void): void {
    try {
        const addrs = __scriptgo.dnsResolveStrings(hostname, "PTR");
        callback(null, addrs);
    } catch (err) {
        callback(err, []);
    }
}

export function reverse(ip: string, callback: (err: unknown, hostnames: string[]) => void): void {
    try {
        const names = __scriptgo.dnsReverse(ip);
        callback(null, names);
    } catch (err) {
        callback(err, []);
    }
}


export namespace promises {
    export async function getDefaultResultOrder(): Promise<string> {
        return _dnsResultOrder;
    }

    export async function setDefaultResultOrder(order: string): Promise<void> {
        if (order === "ipv4first" || order === "verbatim") {
            _dnsResultOrder = order;
        }
    }

    export async function lookup(hostname: string, options?: unknown): Promise<LookupAddress> {
        let reqFamily = 0;
        if (typeof options === "object" && options !== null) {
            const opt = options as { family?: number };
            if (typeof opt.family === "number") reqFamily = opt.family as number;
        } else if (typeof options === "number") {
            reqFamily = options as number;
        }
        const res = __scriptgo.dnsLookup(hostname, reqFamily);
        return new LookupAddress(res.address, res.family);
    }

    export async function lookupService(address: string, port: number): Promise<{ hostname: string; service: string }> {
        const res = __scriptgo.dnsLookupService(address, port);
        return { hostname: res.hostname, service: res.service };
    }

    export async function resolve(hostname: string, rrtype: string = "A"): Promise<string[]> {
        return __scriptgo.dnsResolveStrings(hostname, rrtype);
    }

    export async function resolve4(hostname: string): Promise<string[]> {
        return __scriptgo.dnsResolveStrings(hostname, "A");
    }

    export async function resolve6(hostname: string): Promise<string[]> {
        return __scriptgo.dnsResolveStrings(hostname, "AAAA");
    }

    export async function resolveCname(hostname: string): Promise<string[]> {
        return __scriptgo.dnsResolveStrings(hostname, "CNAME");
    }

    export async function resolveNs(hostname: string): Promise<string[]> {
        return __scriptgo.dnsResolveStrings(hostname, "NS");
    }

    export async function resolvePtr(hostname: string): Promise<string[]> {
        return __scriptgo.dnsResolveStrings(hostname, "PTR");
    }

    export async function reverse(ip: string): Promise<string[]> {
        return __scriptgo.dnsReverse(ip);
    }
}

export default {
    NODATA,
    FORMERR,
    SERVFAIL,
    NOTFOUND,
    NOTIMP,
    REFUSED,
    BADQUERY,
    BADNAME,
    BADFAMILY,
    BADRESP,
    CONNREFUSED,
    TIMEOUT,
    EOF,
    FILE,
    NOMEM,
    DESTRUCTION,
    BADSTR,
    BADFLAGS,
    NONAME,
    BADHINTS,
    NOTINITIALIZED,
    LOADIPHLPAPI,
    ADDRGETNETWORKPARAMS,
    CANCELLED,
    LookupAddress,
    lookup,
    lookupService,
    resolve,
    resolve4,
    resolve6,
    resolveCname,
    resolveNs,
    resolvePtr,
    reverse,
    getDefaultResultOrder,
    setDefaultResultOrder,
    promises,
};
