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
    function availableParallelism(): number;
    function hostname(): string;
    function loadavg(): string;
    function cpus(): string;
    function networkInterfaces(): string;
    function userInfo(): string;
    function machine(): string;
    function version(): string;
    function getPriority(pid: number): number;
    function setPriority(pid: number, priority: number): void;
}

export interface CpuInfo {
    model: string;
    speed: number;
    times: {
        user: number;
        nice: number;
        sys: number;
        idle: number;
        irq: number;
    };
}

export interface NetworkInterfaceInfo {
    address: string;
    netmask: string;
    family: string;
    mac: string;
    internal: boolean;
    cidr: string | null;
    scopeid?: number;
}

export interface UserInfo<T = string> {
    username: T;
    uid: number;
    gid: number;
    shell: T;
    homedir: T;
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
    return __scriptgo.machine();
}

export function version(): string {
    return __scriptgo.version();
}

export function availableParallelism(): number {
    return __scriptgo.availableParallelism();
}

export function endianness(): "BE" | "LE" {
    const u16 = new Uint16Array([0x1234]);
    const u8 = new Uint8Array(u16.buffer);
    return u8[0] === 0x34 ? "LE" : "BE";
}

export function hostname(): string {
    return __scriptgo.hostname();
}

export function loadavg(): number[] {
    return JSON.parse(__scriptgo.loadavg()) as number[];
}

export function cpus(): CpuInfo[] {
    return JSON.parse(__scriptgo.cpus()) as CpuInfo[];
}

export function networkInterfaces(): Record<string, NetworkInterfaceInfo[]> {
    return JSON.parse(__scriptgo.networkInterfaces()) as Record<string, NetworkInterfaceInfo[]>;
}

export function userInfo(options?: { encoding?: string }): UserInfo<string> {
    return JSON.parse(__scriptgo.userInfo()) as UserInfo<string>;
}

export function getPriority(pid: number = 0): number {
    return __scriptgo.getPriority(pid);
}

export function setPriority(priority: number): void;
export function setPriority(pid: number, priority: number): void;
export function setPriority(pidOrPriority: number, priority: number = 999): void {
    if (priority === 999) {
        __scriptgo.setPriority(0, pidOrPriority);
    } else {
        __scriptgo.setPriority(pidOrPriority, priority);
    }
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
    availableParallelism,
    endianness,
    hostname,
    loadavg,
    cpus,
    networkInterfaces,
    userInfo,
    getPriority,
    setPriority,
};
