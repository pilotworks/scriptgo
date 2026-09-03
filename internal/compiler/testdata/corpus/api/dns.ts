import {
    getServers,
    setServers,
    getDefaultResultOrder,
    setDefaultResultOrder,
    cancel,
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
    Resolver,
    promises,
    LookupAddress,
    MxRecord,
    NaptrRecord,
    SoaRecord,
    SrvRecord,
    CaaRecord,
    AnyRecord
} from "node:dns";

// @api: dns.getServers
// @expect: getServers passed
const servers = getServers();
if (servers.length > 0 && servers[0] === "127.0.0.1") {
    console.log("getServers passed");
}

// @api: dns.setServers
// @expect: setServers passed
setServers(["1.1.1.1", "8.8.8.8"]);
const updatedServers = getServers();
if (updatedServers.length === 2 && updatedServers[0] === "1.1.1.1") {
    console.log("setServers passed");
}

// @api: dns.getDefaultResultOrder
// @api: dns.setDefaultResultOrder
// @expect: resultOrder passed
setDefaultResultOrder("ipv4first");
if (getDefaultResultOrder() === "ipv4first") {
    console.log("resultOrder passed");
}

// @api: dns.cancel
// @expect: cancel passed
cancel();
console.log("cancel passed");

// @api: dns.lookup
// @expect: lookup passed
lookup("localhost", (err: unknown, address: string, family: number): void => {
    if (address === "127.0.0.1" && family === 4) {
        console.log("lookup passed");
    }
});

// @api: dns.lookupService
// @expect: lookupService passed
lookupService("127.0.0.1", 80, (err: unknown, hostname: string, service: string): void => {
    if (hostname === "localhost" && service === "http") {
        console.log("lookupService passed");
    }
});

// @api: dns.resolve
// @expect: resolve passed
resolve("localhost", (err: unknown, addresses: string[]): void => {
    if (addresses.length > 0) {
        console.log("resolve passed");
    }
});

// @api: dns.resolve4
// @expect: resolve4 passed
resolve4("localhost", (err: unknown, addresses: string[]): void => {
    if (addresses[0] === "127.0.0.1") {
        console.log("resolve4 passed");
    }
});

// @api: dns.resolve6
// @expect: resolve6 passed
resolve6("localhost", (err: unknown, addresses: string[]): void => {
    if (addresses[0] === "::1") {
        console.log("resolve6 passed");
    }
});

// @api: dns.resolveAny
// @expect: resolveAny passed
resolveAny("localhost", (err: unknown, records: AnyRecord[]): void => {
    if (records.length > 0 && records[0].type === "A") {
        console.log("resolveAny passed");
    }
});

// @api: dns.resolveCname
// @expect: resolveCname passed
resolveCname("example.com", (err: unknown, addresses: string[]): void => {
    if (addresses.length > 0) {
        console.log("resolveCname passed");
    }
});

// @api: dns.resolveCaa
// @expect: resolveCaa passed
resolveCaa("example.com", (err: unknown, records: CaaRecord[]): void => {
    if (records.length > 0 && records[0].issue === "letsencrypt.org") {
        console.log("resolveCaa passed");
    }
});

// @api: dns.resolveMx
// @expect: resolveMx passed
resolveMx("example.com", (err: unknown, records: MxRecord[]): void => {
    if (records.length > 0 && records[0].priority === 10) {
        console.log("resolveMx passed");
    }
});

// @api: dns.resolveNaptr
// @expect: resolveNaptr passed
resolveNaptr("example.com", (err: unknown, records: NaptrRecord[]): void => {
    if (records.length > 0 && records[0].flags === "s") {
        console.log("resolveNaptr passed");
    }
});

// @api: dns.resolveNs
// @expect: resolveNs passed
resolveNs("example.com", (err: unknown, addresses: string[]): void => {
    if (addresses.length > 0) {
        console.log("resolveNs passed");
    }
});

// @api: dns.resolvePtr
// @expect: resolvePtr passed
resolvePtr("127.0.0.1", (err: unknown, addresses: string[]): void => {
    if (addresses.length > 0 && addresses[0] === "localhost") {
        console.log("resolvePtr passed");
    }
});

// @api: dns.resolveSoa
// @expect: resolveSoa passed
resolveSoa("example.com", (err: unknown, record: SoaRecord): void => {
    if (record.serial > 0) {
        console.log("resolveSoa passed");
    }
});

// @api: dns.resolveSrv
// @expect: resolveSrv passed
resolveSrv("example.com", (err: unknown, records: SrvRecord[]): void => {
    if (records.length > 0) {
        console.log("resolveSrv passed");
    }
});

// @api: dns.resolveTlsa
// @expect: resolveTlsa passed
resolveTlsa("example.com", (err: unknown, records: string[]): void => {
    console.log("resolveTlsa passed");
});

// @api: dns.resolveTxt
// @expect: resolveTxt passed
resolveTxt("example.com", (err: unknown, records: string[][]): void => {
    if (records.length > 0) {
        console.log("resolveTxt passed");
    }
});

// @api: dns.reverse
// @expect: reverse passed
reverse("127.0.0.1", (err: unknown, hostnames: string[]): void => {
    if (hostnames.length > 0 && hostnames[0] === "localhost") {
        console.log("reverse passed");
    }
});

// @api: dns.dns.Resolver
// @api: dns.Resolver.setLocalAddress
// @api: dns.Resolver.cancel
// @api: dns.Resolver.Resolver
// @expect: Resolver passed
const res = new Resolver();
res.setLocalAddress("127.0.0.1", "::1");
res.cancel();
if (res instanceof Resolver) {
    console.log("Resolver passed");
}

// @api: dns.dnsPromises.Resolver
// @expect: promises Resolver passed
const pRes = new promises.Resolver();
if (pRes instanceof promises.Resolver) {
    console.log("promises Resolver passed");
}
