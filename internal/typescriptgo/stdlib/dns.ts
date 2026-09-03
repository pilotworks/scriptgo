// ScriptGo Standard Library: node:dns

declare namespace __scriptgo {
    function dnsLookup(hostname: string, family?: number): { address: string; family: number };
    function dnsLookupService(address: string, port: number): { hostname: string; service: string };
    function dnsReverse(ip: string): string[];
    function dnsResolveStrings(hostname: string, rrtype: string): string[];
    function dnsResolveTxt(hostname: string): string[];
    function dnsResolveMx(hostname: string): { exchanges: string[]; priorities: number[] };
    function dnsResolveSrv(hostname: string): { names: string[]; ports: number[]; priorities: number[]; weights: number[] };
    function dnsResolveSoa(hostname: string): { nsname: string; hostmaster: string; serial: number; refresh: number; retry: number; expire: number; minttl: number };
    function dnsResolveCaa(hostname: string): { criticals: number[]; issues: string[] };
    function dnsResolveNaptr(hostname: string): { flags: string[]; services: string[]; regexps: string[]; replacements: string[]; orders: number[]; preferences: number[] };
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

export class MxRecord {
    exchange: string = "";
    priority: number = 0;

    constructor(exchange: string, priority: number) {
        this.exchange = exchange;
        this.priority = priority;
    }
}

export class NaptrRecord {
    flags: string = "";
    service: string = "";
    regexp: string = "";
    replacement: string = "";
    order: number = 0;
    preference: number = 0;

    constructor(flags: string, service: string, regexp: string, replacement: string, order: number, preference: number) {
        this.flags = flags;
        this.service = service;
        this.regexp = regexp;
        this.replacement = replacement;
        this.order = order;
        this.preference = preference;
    }
}

export class SoaRecord {
    nsname: string = "";
    hostmaster: string = "";
    serial: number = 0;
    refresh: number = 0;
    retry: number = 0;
    expire: number = 0;
    minttl: number = 0;

    constructor(nsname: string, hostmaster: string, serial: number, refresh: number, retry: number, expire: number, minttl: number) {
        this.nsname = nsname;
        this.hostmaster = hostmaster;
        this.serial = serial;
        this.refresh = refresh;
        this.retry = retry;
        this.expire = expire;
        this.minttl = minttl;
    }
}

export class SrvRecord {
    name: string = "";
    port: number = 0;
    priority: number = 0;
    weight: number = 0;

    constructor(name: string, port: number, priority: number, weight: number) {
        this.name = name;
        this.port = port;
        this.priority = priority;
        this.weight = weight;
    }
}

export class CaaRecord {
    critical: number = 0;
    issue: string = "";

    constructor(critical: number, issue: string) {
        this.critical = critical;
        this.issue = issue;
    }
}

export class AnyRecord {
    address: string = "";
    family: number = 4;
    type: string = "A";

    constructor(address: string, family: number, type: string) {
        this.address = address;
        this.family = family;
        this.type = type;
    }
}

let _dnsServers: string[] = ["127.0.0.1", "8.8.8.8", "1.1.1.1"];
let _dnsResultOrder: string = "verbatim";

export function getServers(): string[] {
    return _dnsServers;
}

export function setServers(servers: string[]): void {
    _dnsServers = servers;
}

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

    if (rrtype === "MX") {
        resolveMx(hostname, cb as (err: unknown, addresses: MxRecord[]) => void);
    } else if (rrtype === "TXT") {
        resolveTxt(hostname, cb as (err: unknown, records: string[][]) => void);
    } else if (rrtype === "SRV") {
        resolveSrv(hostname, cb as (err: unknown, records: SrvRecord[]) => void);
    } else if (rrtype === "SOA") {
        resolveSoa(hostname, cb as (err: unknown, record: SoaRecord) => void);
    } else if (rrtype === "NAPTR") {
        resolveNaptr(hostname, cb as (err: unknown, records: NaptrRecord[]) => void);
    } else if (rrtype === "CAA") {
        resolveCaa(hostname, cb as (err: unknown, records: CaaRecord[]) => void);
    } else if (rrtype === "ANY") {
        resolveAny(hostname, cb as (err: unknown, records: AnyRecord[]) => void);
    } else {
        try {
            const addrs = __scriptgo.dnsResolveStrings(hostname, rrtype);
            (cb as (err: unknown, addresses: string[]) => void)(null, addrs);
        } catch (err) {
            (cb as (err: unknown, addresses: string[]) => void)(err, []);
        }
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

export function resolveAny(hostname: string, callback: (err: unknown, records: AnyRecord[]) => void): void {
    try {
        const fam = _dnsResultOrder === "ipv4first" ? 4 : 0;
        const res = __scriptgo.dnsLookup(hostname, fam);
        callback(null, [new AnyRecord(res.address, res.family, res.family === 6 ? "AAAA" : "A")]);
    } catch (err) {
        callback(err, []);
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

export function resolveCaa(hostname: string, callback: (err: unknown, records: CaaRecord[]) => void): void {
    try {
        const res = __scriptgo.dnsResolveCaa(hostname);
        const records: CaaRecord[] = [];
        for (let i = 0; i < res.criticals.length; i++) {
            records.push(new CaaRecord(res.criticals[i], res.issues[i]));
        }
        callback(null, records);
    } catch (err) {
        callback(err, []);
    }
}

export function resolveMx(hostname: string, callback: (err: unknown, addresses: MxRecord[]) => void): void {
    try {
        const res = __scriptgo.dnsResolveMx(hostname);
        const records: MxRecord[] = [];
        for (let i = 0; i < res.exchanges.length; i++) {
            records.push(new MxRecord(res.exchanges[i], res.priorities[i]));
        }
        callback(null, records);
    } catch (err) {
        callback(err, []);
    }
}

export function resolveNaptr(hostname: string, callback: (err: unknown, records: NaptrRecord[]) => void): void {
    try {
        const res = __scriptgo.dnsResolveNaptr(hostname);
        const records: NaptrRecord[] = [];
        for (let i = 0; i < res.flags.length; i++) {
            records.push(new NaptrRecord(res.flags[i], res.services[i], res.regexps[i], res.replacements[i], res.orders[i], res.preferences[i]));
        }
        callback(null, records);
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

export function resolveSoa(hostname: string, callback: (err: unknown, record: SoaRecord) => void): void {
    try {
        const res = __scriptgo.dnsResolveSoa(hostname);
        callback(null, new SoaRecord(res.nsname, res.hostmaster, res.serial, res.refresh, res.retry, res.expire, res.minttl));
    } catch (err) {
        callback(err, new SoaRecord("", "", 0, 0, 0, 0, 0));
    }
}

export function resolveSrv(hostname: string, callback: (err: unknown, records: SrvRecord[]) => void): void {
    try {
        const res = __scriptgo.dnsResolveSrv(hostname);
        const records: SrvRecord[] = [];
        for (let i = 0; i < res.names.length; i++) {
            records.push(new SrvRecord(res.names[i], res.ports[i], res.priorities[i], res.weights[i]));
        }
        callback(null, records);
    } catch (err) {
        callback(err, []);
    }
}

export function resolveTlsa(hostname: string, callback: (err: unknown, records: string[]) => void): void {
    try {
        const addrs = __scriptgo.dnsResolveStrings(hostname, "TLSA");
        callback(null, addrs);
    } catch (err) {
        callback(err, []);
    }
}

export function resolveTxt(hostname: string, callback: (err: unknown, records: string[][]) => void): void {
    try {
        const res = __scriptgo.dnsResolveTxt(hostname);
        const records: string[][] = [];
        for (let i = 0; i < res.length; i++) {
            records.push([res[i]]);
        }
        callback(null, records);
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

export class Resolver {
    private _servers: string[] = ["127.0.0.1", "8.8.8.8", "1.1.1.1"];
    private _localIPv4: string = "";
    private _localIPv6: string = "";

    constructor(options?: unknown) {
        this._servers = ["127.0.0.1", "8.8.8.8", "1.1.1.1"];
    }

    setLocalAddress(ipv4?: string, ipv6?: string): void {
        if (ipv4) this._localIPv4 = ipv4;
        if (ipv6) this._localIPv6 = ipv6;
    }

    getServers(): string[] {
        const res: string[] = [];
        for (let i = 0; i < this._servers.length; i++) {
            res.push(this._servers[i]);
        }
        return res;
    }

    setServers(servers: string[]): void {
        const next: string[] = [];
        for (let i = 0; i < servers.length; i++) {
            next.push(servers[i]);
        }
        this._servers = next;
    }

    resolve(hostname: string, rrtypeOrCallback?: unknown, callback?: unknown): void {
        resolve(hostname, rrtypeOrCallback, callback);
    }

    resolve4(hostname: string, optionsOrCallback?: unknown, callback?: unknown): void {
        resolve4(hostname, optionsOrCallback, callback);
    }

    resolve6(hostname: string, optionsOrCallback?: unknown, callback?: unknown): void {
        resolve6(hostname, optionsOrCallback, callback);
    }

    resolveAny(hostname: string, callback: (err: unknown, records: AnyRecord[]) => void): void {
        resolveAny(hostname, callback);
    }

    resolveCname(hostname: string, callback: (err: unknown, addresses: string[]) => void): void {
        resolveCname(hostname, callback);
    }

    resolveCaa(hostname: string, callback: (err: unknown, records: CaaRecord[]) => void): void {
        resolveCaa(hostname, callback);
    }

    resolveMx(hostname: string, callback: (err: unknown, addresses: MxRecord[]) => void): void {
        resolveMx(hostname, callback);
    }

    resolveNaptr(hostname: string, callback: (err: unknown, records: NaptrRecord[]) => void): void {
        resolveNaptr(hostname, callback);
    }

    resolveNs(hostname: string, callback: (err: unknown, addresses: string[]) => void): void {
        resolveNs(hostname, callback);
    }

    resolvePtr(hostname: string, callback: (err: unknown, addresses: string[]) => void): void {
        resolvePtr(hostname, callback);
    }

    resolveSoa(hostname: string, callback: (err: unknown, record: SoaRecord) => void): void {
        resolveSoa(hostname, callback);
    }

    resolveSrv(hostname: string, callback: (err: unknown, records: SrvRecord[]) => void): void {
        resolveSrv(hostname, callback);
    }

    resolveTlsa(hostname: string, callback: (err: unknown, records: string[]) => void): void {
        resolveTlsa(hostname, callback);
    }

    resolveTxt(hostname: string, callback: (err: unknown, records: string[][]) => void): void {
        resolveTxt(hostname, callback);
    }

    reverse(ip: string, callback: (err: unknown, hostnames: string[]) => void): void {
        reverse(ip, callback);
    }
}

export class PromiseResolver {
    private _servers: string[] = ["127.0.0.1", "8.8.8.8", "1.1.1.1"];
    private _localIPv4: string = "";
    private _localIPv6: string = "";

    constructor(options?: unknown) {
        this._servers = ["127.0.0.1", "8.8.8.8", "1.1.1.1"];
    }

    setLocalAddress(ipv4?: string, ipv6?: string): void {
        if (ipv4) this._localIPv4 = ipv4;
        if (ipv6) this._localIPv6 = ipv6;
    }

    getServers(): string[] {
        const res: string[] = [];
        for (let i = 0; i < this._servers.length; i++) {
            res.push(this._servers[i]);
        }
        return res;
    }

    setServers(servers: string[]): void {
        const next: string[] = [];
        for (let i = 0; i < servers.length; i++) {
            next.push(servers[i]);
        }
        this._servers = next;
    }
}

export namespace promises {
    export class Resolver extends PromiseResolver {}

    export async function getServers(): Promise<string[]> {
        return getServers();
    }

    export async function setServers(servers: string[]): Promise<void> {
        setServers(servers);
    }

    export async function getDefaultResultOrder(): Promise<string> {
        return getDefaultResultOrder();
    }

    export async function setDefaultResultOrder(order: string): Promise<void> {
        setDefaultResultOrder(order);
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

    export async function resolveAny(hostname: string): Promise<AnyRecord[]> {
        const res = __scriptgo.dnsLookup(hostname, 0);
        return [new AnyRecord(res.address, res.family, res.family === 6 ? "AAAA" : "A")];
    }

    export async function resolveCname(hostname: string): Promise<string[]> {
        return __scriptgo.dnsResolveStrings(hostname, "CNAME");
    }

    export async function resolveCaa(hostname: string): Promise<CaaRecord[]> {
        const res = __scriptgo.dnsResolveCaa(hostname);
        const records: CaaRecord[] = [];
        for (let i = 0; i < res.criticals.length; i++) {
            records.push(new CaaRecord(res.criticals[i], res.issues[i]));
        }
        return records;
    }

    export async function resolveMx(hostname: string): Promise<MxRecord[]> {
        const res = __scriptgo.dnsResolveMx(hostname);
        const records: MxRecord[] = [];
        for (let i = 0; i < res.exchanges.length; i++) {
            records.push(new MxRecord(res.exchanges[i], res.priorities[i]));
        }
        return records;
    }

    export async function resolveNaptr(hostname: string): Promise<NaptrRecord[]> {
        const res = __scriptgo.dnsResolveNaptr(hostname);
        const records: NaptrRecord[] = [];
        for (let i = 0; i < res.flags.length; i++) {
            records.push(new NaptrRecord(res.flags[i], res.services[i], res.regexps[i], res.replacements[i], res.orders[i], res.preferences[i]));
        }
        return records;
    }

    export async function resolveNs(hostname: string): Promise<string[]> {
        return __scriptgo.dnsResolveStrings(hostname, "NS");
    }

    export async function resolvePtr(hostname: string): Promise<string[]> {
        return __scriptgo.dnsResolveStrings(hostname, "PTR");
    }

    export async function resolveSoa(hostname: string): Promise<SoaRecord> {
        const res = __scriptgo.dnsResolveSoa(hostname);
        return new SoaRecord(res.nsname, res.hostmaster, res.serial, res.refresh, res.retry, res.expire, res.minttl);
    }

    export async function resolveSrv(hostname: string): Promise<SrvRecord[]> {
        const res = __scriptgo.dnsResolveSrv(hostname);
        const records: SrvRecord[] = [];
        for (let i = 0; i < res.names.length; i++) {
            records.push(new SrvRecord(res.names[i], res.ports[i], res.priorities[i], res.weights[i]));
        }
        return records;
    }

    export async function resolveTlsa(hostname: string): Promise<string[]> {
        return __scriptgo.dnsResolveStrings(hostname, "TLSA");
    }

    export async function resolveTxt(hostname: string): Promise<string[][]> {
        const res = __scriptgo.dnsResolveTxt(hostname);
        const records: string[][] = [];
        for (let i = 0; i < res.length; i++) {
            records.push([res[i]]);
        }
        return records;
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
    MxRecord,
    NaptrRecord,
    SoaRecord,
    SrvRecord,
    CaaRecord,
    AnyRecord,
    Resolver,
    lookup,
    lookupService,
    resolve,
    resolve4,
    resolve6,
    resolveAny,
    resolveCname,
    resolveCaa,
    resolveMx,
    resolveNaptr,
    resolveNs,
    resolvePtr,
    resolveSoa,
    resolveSrv,
    resolveTlsa,
    resolveTxt,
    reverse,
    getServers,
    setServers,
    getDefaultResultOrder,
    setDefaultResultOrder,
    promises,
};
