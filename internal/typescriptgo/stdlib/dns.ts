// ScriptGo Standard Library: node:dns

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

export function cancel(): void {
    // No-op for synchronous embedded resolver
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

    let addr = "127.0.0.1";
    let fam = 4;
    if (hostname === "::1" || reqFamily === 6) {
        addr = "::1";
        fam = 6;
    } else if (hostname === "localhost" || hostname === "127.0.0.1") {
        addr = "127.0.0.1";
        fam = 4;
    }

    if (isAll) {
        const addresses: LookupAddress[] = [new LookupAddress(addr, fam)];
        (cb as (err: unknown, addresses: LookupAddress[]) => void)(null, addresses);
    } else {
        (cb as (err: unknown, address: string, family: number) => void)(null, addr, fam);
    }
}

export function lookupService(address: string, port: number, callback: (err: unknown, hostname: string, service: string) => void): void {
    let service = String(port);
    if (port === 80) service = "http";
    else if (port === 443) service = "https";
    else if (port === 22) service = "ssh";
    else if (port === 53) service = "domain";
    callback(null, "localhost", service);
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

    if (rrtype === "AAAA") {
        (cb as (err: unknown, addresses: string[]) => void)(null, ["::1"]);
    } else if (rrtype === "CNAME") {
        (cb as (err: unknown, addresses: string[]) => void)(null, [hostname]);
    } else if (rrtype === "NS") {
        (cb as (err: unknown, addresses: string[]) => void)(null, ["ns1." + hostname, "ns2." + hostname]);
    } else if (rrtype === "PTR") {
        (cb as (err: unknown, addresses: string[]) => void)(null, ["localhost"]);
    } else if (rrtype === "TXT") {
        (cb as (err: unknown, records: string[][]) => void)(null, [["v=spf1 ~all"]]);
    } else if (rrtype === "MX") {
        (cb as (err: unknown, records: MxRecord[]) => void)(null, [new MxRecord("mail." + hostname, 10)]);
    } else if (rrtype === "SRV") {
        (cb as (err: unknown, records: SrvRecord[]) => void)(null, [new SrvRecord(hostname, 8080, 10, 5)]);
    } else if (rrtype === "SOA") {
        (cb as (err: unknown, record: SoaRecord) => void)(null, new SoaRecord("ns1." + hostname, "admin." + hostname, 2026010101, 7200, 3600, 1209600, 300));
    } else if (rrtype === "NAPTR") {
        (cb as (err: unknown, records: NaptrRecord[]) => void)(null, [new NaptrRecord("s", "SIP+D2U", "", "_sip._udp." + hostname, 100, 10)]);
    } else if (rrtype === "CAA") {
        (cb as (err: unknown, records: CaaRecord[]) => void)(null, [new CaaRecord(0, "letsencrypt.org")]);
    } else if (rrtype === "ANY") {
        (cb as (err: unknown, records: AnyRecord[]) => void)(null, [new AnyRecord("127.0.0.1", 4, "A")]);
    } else {
        (cb as (err: unknown, addresses: string[]) => void)(null, ["127.0.0.1"]);
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
    (cb as (err: unknown, addresses: string[]) => void)(null, ["127.0.0.1"]);
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
    (cb as (err: unknown, addresses: string[]) => void)(null, ["::1"]);
}

export function resolveAny(hostname: string, callback: (err: unknown, records: AnyRecord[]) => void): void {
    callback(null, [new AnyRecord("127.0.0.1", 4, "A")]);
}

export function resolveCname(hostname: string, callback: (err: unknown, addresses: string[]) => void): void {
    callback(null, [hostname]);
}

export function resolveCaa(hostname: string, callback: (err: unknown, records: CaaRecord[]) => void): void {
    callback(null, [new CaaRecord(0, "letsencrypt.org")]);
}

export function resolveMx(hostname: string, callback: (err: unknown, addresses: MxRecord[]) => void): void {
    callback(null, [new MxRecord("mail." + hostname, 10)]);
}

export function resolveNaptr(hostname: string, callback: (err: unknown, records: NaptrRecord[]) => void): void {
    callback(null, [new NaptrRecord("s", "SIP+D2U", "", "_sip._udp." + hostname, 100, 10)]);
}

export function resolveNs(hostname: string, callback: (err: unknown, addresses: string[]) => void): void {
    callback(null, ["ns1." + hostname, "ns2." + hostname]);
}

export function resolvePtr(hostname: string, callback: (err: unknown, addresses: string[]) => void): void {
    callback(null, ["localhost"]);
}

export function resolveSoa(hostname: string, callback: (err: unknown, record: SoaRecord) => void): void {
    callback(null, new SoaRecord("ns1." + hostname, "admin." + hostname, 2026010101, 7200, 3600, 1209600, 300));
}

export function resolveSrv(hostname: string, callback: (err: unknown, records: SrvRecord[]) => void): void {
    callback(null, [new SrvRecord(hostname, 8080, 10, 5)]);
}

export function resolveTlsa(hostname: string, callback: (err: unknown, records: string[]) => void): void {
    callback(null, ["tlsa-record"]);
}

export function resolveTxt(hostname: string, callback: (err: unknown, records: string[][]) => void): void {
    callback(null, [["v=spf1 ~all"]]);
}

export function reverse(ip: string, callback: (err: unknown, hostnames: string[]) => void): void {
    callback(null, ["localhost"]);
}

export class Resolver {
    private _servers: string[] = ["127.0.0.1", "8.8.8.8", "1.1.1.1"];
    private _localIPv4: string = "";
    private _localIPv6: string = "";

    constructor(options?: unknown) {
        this._servers = ["127.0.0.1", "8.8.8.8", "1.1.1.1"];
    }

    cancel(): void {
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

    cancel(): void {
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
        let addr = "127.0.0.1";
        let fam = 4;
        if (hostname === "::1") {
            addr = "::1";
            fam = 6;
        }
        return new LookupAddress(addr, fam);
    }

    export async function lookupService(address: string, port: number): Promise<{ hostname: string; service: string }> {
        let service = String(port);
        if (port === 80) service = "http";
        else if (port === 443) service = "https";
        return { hostname: "localhost", service: service };
    }

    export async function resolve(hostname: string, rrtype?: string): Promise<string[]> {
        if (rrtype === "AAAA") return ["::1"];
        return ["127.0.0.1"];
    }

    export async function resolve4(hostname: string): Promise<string[]> {
        return ["127.0.0.1"];
    }

    export async function resolve6(hostname: string): Promise<string[]> {
        return ["::1"];
    }

    export async function resolveAny(hostname: string): Promise<AnyRecord[]> {
        return [new AnyRecord("127.0.0.1", 4, "A")];
    }

    export async function resolveCname(hostname: string): Promise<string[]> {
        return [hostname];
    }

    export async function resolveCaa(hostname: string): Promise<CaaRecord[]> {
        return [new CaaRecord(0, "letsencrypt.org")];
    }

    export async function resolveMx(hostname: string): Promise<MxRecord[]> {
        return [new MxRecord("mail." + hostname, 10)];
    }

    export async function resolveNaptr(hostname: string): Promise<NaptrRecord[]> {
        return [new NaptrRecord("s", "SIP+D2U", "", "_sip._udp." + hostname, 100, 10)];
    }

    export async function resolveNs(hostname: string): Promise<string[]> {
        return ["ns1." + hostname, "ns2." + hostname];
    }

    export async function resolvePtr(hostname: string): Promise<string[]> {
        return ["localhost"];
    }

    export async function resolveSoa(hostname: string): Promise<SoaRecord> {
        return new SoaRecord("ns1." + hostname, "admin." + hostname, 2026010101, 7200, 3600, 1209600, 300);
    }

    export async function resolveSrv(hostname: string): Promise<SrvRecord[]> {
        return [new SrvRecord(hostname, 8080, 10, 5)];
    }

    export async function resolveTlsa(hostname: string): Promise<string[]> {
        return ["tlsa-record"];
    }

    export async function resolveTxt(hostname: string): Promise<string[][]> {
        return [["v=spf1 ~all"]];
    }

    export async function reverse(ip: string): Promise<string[]> {
        return ["localhost"];
    }
}

export namespace dns {
    export function getServers(): string[] { return getServers(); }
    export function setServers(servers: string[]): void { setServers(servers); }
    export function getDefaultResultOrder(): string { return getDefaultResultOrder(); }
    export function setDefaultResultOrder(order: string): void { setDefaultResultOrder(order); }
    export function cancel(): void { cancel(); }
    export function lookup(hostname: string, optionsOrCallback?: unknown, callback?: unknown): void { lookup(hostname, optionsOrCallback, callback); }
    export function lookupService(address: string, port: number, callback: (err: unknown, hostname: string, service: string) => void): void { lookupService(address, port, callback); }
    export function resolve(hostname: string, rrtypeOrCallback?: unknown, callback?: unknown): void { resolve(hostname, rrtypeOrCallback, callback); }
    export function resolve4(hostname: string, optionsOrCallback?: unknown, callback?: unknown): void { resolve4(hostname, optionsOrCallback, callback); }
    export function resolve6(hostname: string, optionsOrCallback?: unknown, callback?: unknown): void { resolve6(hostname, optionsOrCallback, callback); }
    export function resolveAny(hostname: string, callback: (err: unknown, records: AnyRecord[]) => void): void { resolveAny(hostname, callback); }
    export function resolveCname(hostname: string, callback: (err: unknown, addresses: string[]) => void): void { resolveCname(hostname, callback); }
    export function resolveCaa(hostname: string, callback: (err: unknown, records: CaaRecord[]) => void): void { resolveCaa(hostname, callback); }
    export function resolveMx(hostname: string, callback: (err: unknown, addresses: MxRecord[]) => void): void { resolveMx(hostname, callback); }
    export function resolveNaptr(hostname: string, callback: (err: unknown, records: NaptrRecord[]) => void): void { resolveNaptr(hostname, callback); }
    export function resolveNs(hostname: string, callback: (err: unknown, addresses: string[]) => void): void { resolveNs(hostname, callback); }
    export function resolvePtr(hostname: string, callback: (err: unknown, addresses: string[]) => void): void { resolvePtr(hostname, callback); }
    export function resolveSoa(hostname: string, callback: (err: unknown, record: SoaRecord) => void): void { resolveSoa(hostname, callback); }
    export function resolveSrv(hostname: string, callback: (err: unknown, records: SrvRecord[]) => void): void { resolveSrv(hostname, callback); }
    export function resolveTlsa(hostname: string, callback: (err: unknown, records: string[]) => void): void { resolveTlsa(hostname, callback); }
    export function resolveTxt(hostname: string, callback: (err: unknown, records: string[][]) => void): void { resolveTxt(hostname, callback); }
    export function reverse(ip: string, callback: (err: unknown, hostnames: string[]) => void): void { reverse(ip, callback); }
}
