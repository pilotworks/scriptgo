// @expect: host: localhost, port: 8080, secure: true, timeout: 5000
// @expect: host: remote.io, port: 443, secure: false, timeout: 3000
interface Config {
    server?: {
        host?: string;
        port?: number;
    };
    security?: {
        enabled?: boolean;
    };
    timeout?: number;
}

function printConfig(cfg: Config): void {
    const {
        server: {
            host: serverHost = "localhost",
            port: serverPort = 8080
        } = {},
        security: {
            enabled: isSecure = true
        } = {},
        timeout = 5000
    } = cfg;

    console.log("host: " + serverHost + ", port: " + serverPort + ", secure: " + isSecure + ", timeout: " + timeout);
}

printConfig({});
printConfig({
    server: { host: "remote.io", port: 443 },
    security: { enabled: false },
    timeout: 3000
});
