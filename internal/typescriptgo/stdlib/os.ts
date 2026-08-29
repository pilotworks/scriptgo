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

export class CpuTimes {
    user: number;
    nice: number;
    sys: number;
    idle: number;
    irq: number;

    constructor(user: number, nice: number, sys: number, idle: number, irq: number) {
        this.user = user;
        this.nice = nice;
        this.sys = sys;
        this.idle = idle;
        this.irq = irq;
    }
}

export class CpuInfo {
    model: string;
    speed: number;
    times: CpuTimes;

    constructor(model: string, speed: number, times: CpuTimes) {
        this.model = model;
        this.speed = speed;
        this.times = times;
    }
}

export class NetworkInterfaceInfo {
    address: string;
    netmask: string;
    family: string;
    mac: string;
    internal: boolean;
    cidr: string;

    constructor(address: string, netmask: string, family: string, mac: string, internal: boolean, cidr: string) {
        this.address = address;
        this.netmask = netmask;
        this.family = family;
        this.mac = mac;
        this.internal = internal;
        this.cidr = cidr;
    }
}

export class UserInfo {
    uid: number;
    gid: number;
    username: string;
    homedir: string;
    shell: string;

    constructor(uid: number, gid: number, username: string, homedir: string, shell: string) {
        this.uid = uid;
        this.gid = gid;
        this.username = username;
        this.homedir = homedir;
        this.shell = shell;
    }
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

export function hostname(): string {
    return "localhost";
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

export function endianness(): string {
    return "LE";
}

export function loadavg(): number[] {
    return [0, 0, 0];
}

export function availableParallelism(): number {
    return 4;
}

export function cpus(): CpuInfo[] {
    const res: CpuInfo[] = [];
    const count = availableParallelism();
    for (let i = 0; i < count; i++) {
        res.push(new CpuInfo("Native CPU", 2400, new CpuTimes(100, 0, 100, 1000, 0)));
    }
    return res;
}

export function networkInterfaces(): Record<string, NetworkInterfaceInfo[]> {
    const lo: NetworkInterfaceInfo[] = [
        new NetworkInterfaceInfo("127.0.0.1", "255.0.0.0", "IPv4", "00:00:00:00:00:00", true, "127.0.0.1/8")
    ];
    return {
        "lo0": lo,
    };
}

export function userInfo(): UserInfo {
    return new UserInfo(501, 20, "user", homedir(), "/bin/sh");
}

export function getPriority(pid: number = 0): number {
    return 0;
}

export function setPriority(pid: number = 0, priority: number = 0): void {}

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
    hostname,
    machine,
    version,
    endianness,
    loadavg,
    cpus,
    networkInterfaces,
    userInfo,
    getPriority,
    setPriority,
};
