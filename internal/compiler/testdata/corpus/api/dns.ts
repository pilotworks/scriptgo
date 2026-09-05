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
lookup("localhost", (err: unknown, address: string, family: number): void => {
    if (address === "127.0.0.1" && family === 4) {
    }
});

// @api: dns.lookupService
lookupService("127.0.0.1", 80, (err: unknown, hostname: string, service: string): void => {
    if (hostname === "localhost" && service === "http") {
    }
});

// @api: dns.resolve
resolve("localhost", (err: unknown, addresses: string[]): void => {
    if (err !== null || Array.isArray(addresses)) {
    }
});

// @api: dns.resolve4
resolve4("localhost", (err: unknown, addresses: string[]): void => {
    if (err !== null || Array.isArray(addresses)) {
    }
});

// @api: dns.resolve6
resolve6("localhost", (err: unknown, addresses: string[]): void => {
    if (err !== null || Array.isArray(addresses)) {
    }
});

// @api: dns.resolveCname
resolveCname("example.com", (err: unknown, addresses: string[]): void => {
    if (Array.isArray(addresses)) {
    }
});

// @api: dns.resolveNs
resolveNs("example.com", (err: unknown, addresses: string[]): void => {
    if (err !== null || Array.isArray(addresses)) {
    }
});

// @api: dns.resolvePtr
resolvePtr("127.0.0.1", (err: unknown, addresses: string[]): void => {
    if (err !== null || Array.isArray(addresses)) {
    }
});

// @api: dns.reverse
reverse("127.0.0.1", (err: unknown, hostnames: string[]): void => {
    if (hostnames.length > 0 && hostnames[0] === "localhost") {
    }
});
