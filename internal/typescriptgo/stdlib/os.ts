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

export const EOL = "\n";
export const devNull = "/dev/null";
export const constants = {
    UV_UDP_REUSEADDR: 4,
    signals: {},
    errno: {},
    priority: {},
};

export function availableParallelism(): number {
    return 1;
}

export default {
    EOL,
    devNull,
    constants,
    availableParallelism,
    platform,
    arch,
    homedir,
    uptime,
    totalmem,
    freemem,
    type,
    release,
    tmpdir,
};

