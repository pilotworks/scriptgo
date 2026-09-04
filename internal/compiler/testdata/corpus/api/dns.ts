import {
    getDefaultResultOrder,
    setDefaultResultOrder,
    lookup,
    lookupService,
    resolve,
    resolve4,
    resolve6,
    resolveCname,
    resolveNs,
    resolvePtr,
    reverse,
    promises,
} from "node:dns";


// @api: dns.getDefaultResultOrder
// @api: dns.setDefaultResultOrder
// @expect: resultOrder passed
setDefaultResultOrder("ipv4first");
if (getDefaultResultOrder() === "ipv4first") {
    console.log("resultOrder passed");
}

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

// @api: dns.resolveCname
// @expect: resolveCname passed
resolveCname("example.com", (err: unknown, addresses: string[]): void => {
    if (Array.isArray(addresses)) {
        console.log("resolveCname passed");
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

// @api: dns.reverse
// @expect: reverse passed
reverse("127.0.0.1", (err: unknown, hostnames: string[]): void => {
    if (hostnames.length > 0 && hostnames[0] === "localhost") {
        console.log("reverse passed");
    }
});
