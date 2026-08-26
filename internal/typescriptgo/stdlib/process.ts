declare namespace __scriptgo {
    function exit(code: number): void;
    function cwd(): string;
    function argv(): string[];
    function pid(): number;
    function ppid(): number;
    function version(): string;
    function platform(): string;
    function arch(): string;
    function uptime(): number;
}

export function exit(code: number): void {
    __scriptgo.exit(code);
}

export function cwd(): string {
    return __scriptgo.cwd();
}

export const argv: string[] = __scriptgo.argv();
export const env: Record<string, string | undefined> = {};
export const pid = __scriptgo.pid();
export const ppid = __scriptgo.ppid();
export const version = __scriptgo.version();
export const versions: Record<string, string> = {
    scriptgo: __scriptgo.version().replace(/^v/, ""),
};
export const platform = __scriptgo.platform();
export const arch = __scriptgo.arch();

export function nextTick(callback: (...args: unknown[]) => void, ...args: unknown[]): void {
    callback(...args);
}

export function hrtime(time?: [number, number]): [number, number] {
    const now = Date.now();
    const seconds = Math.floor(now / 1000);
    const nanos = (now % 1000) * 1000000;
    if (time !== undefined && time.length >= 2) {
        let secDiff = seconds - time[0];
        let nanoDiff = nanos - time[1];
        if (nanoDiff < 0) {
            secDiff = secDiff - 1;
            nanoDiff = nanoDiff + 1000000000;
        }
        return [secDiff, nanoDiff];
    }
    return [seconds, nanos];
}

export function uptime(): number {
    return __scriptgo.uptime();
}

export default {
    exit,
    cwd,
    argv,
    env,
    pid,
    ppid,
    version,
    versions,
    platform,
    arch,
    nextTick,
    hrtime,
    uptime,
};


