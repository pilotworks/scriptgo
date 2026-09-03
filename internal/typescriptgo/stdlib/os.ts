// ScriptGo Standard Library: node:os

declare namespace __scriptgo {
    function platform(): string;
    function arch(): string;
    function homedir(): string;
    function uptime(): number;
    function totalmem(): number;
    function freemem(): number;
    function type(): string;
    function release(): string;
    function tmpdir(): string;
}

export function platform(): string {
    return __scriptgo.platform();
}

export function arch(): string {
    return __scriptgo.arch();
}

export function homedir(): string {
    return __scriptgo.homedir();
}

export function uptime(): number {
    return __scriptgo.uptime();
}

export function totalmem(): number {
    return __scriptgo.totalmem();
}

export function freemem(): number {
    return __scriptgo.freemem();
}

export function type(): string {
    return __scriptgo.type();
}

export function release(): string {
    return __scriptgo.release();
}

export function tmpdir(): string {
    return __scriptgo.tmpdir();
}

export function machine(): string {
    const a = arch();
    if (a === "arm64" || a === "aarch64") {
        return "arm64";
    }
    return "x86_64";
}

export function version(): string {
    return type() + " " + release();
}

export const EOL = "\n";
export const devNull = "/dev/null";

export const constants = {
    UV_UDP_REUSEADDR: 4,
    signals: {
        SIGHUP: 1,
        SIGINT: 2,
        SIGQUIT: 3,
        SIGILL: 4,
        SIGTRAP: 5,
        SIGABRT: 6,
        SIGKILL: 9,
        SIGTERM: 15,
    },
    errno: {
        EPERM: 1,
        ENOENT: 2,
        EEXIST: 17,
        EACCES: 13,
    },
    priority: {
        PRIORITY_LOW: 19,
        PRIORITY_BELOW_NORMAL: 10,
        PRIORITY_NORMAL: 0,
        PRIORITY_ABOVE_NORMAL: -10,
        PRIORITY_HIGH: -19,
        PRIORITY_HIGHEST: -20,
    },
};

export default {
    EOL,
    devNull,
    constants,
    platform,
    arch,
    homedir,
    uptime,
    totalmem,
    freemem,
    type,
    release,
    tmpdir,
    machine,
    version,
};
