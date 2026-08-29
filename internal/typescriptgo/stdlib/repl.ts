// Node.js REPL module (node:repl)

export const builtinModules: string[] = [
    "assert",
    "buffer",
    "child_process",
    "console",
    "crypto",
    "dgram",
    "dns",
    "events",
    "fs",
    "http",
    "https",
    "net",
    "os",
    "path",
    "process",
    "punycode",
    "querystring",
    "readline",
    "repl",
    "stream",
    "string_decoder",
    "timers",
    "tls",
    "tty",
    "url",
    "util",
    "v8",
    "vm",
    "wasi",
    "worker_threads",
    "zlib"
];

export class REPLServer {
    prompt: string = "> ";
    terminal: boolean = true;
    useColors: boolean = true;
    useGlobal: boolean = false;
    context: Record<string, unknown> = {};

    constructor(options: unknown = null) {
        if (typeof options === "string") {
            this.prompt = options;
        } else if (options !== null && typeof options === "object") {
            const opt = options as Record<string, unknown>;
            if (typeof opt["prompt"] === "string") {
                this.prompt = opt["prompt"] as string;
            }
        }
        this.context = {};
    }

    defineCommand(keyword: string, cmd: unknown): void {
    }

    displayPrompt(preserveCursor: boolean = false): void {
    }

    clearBufferedCommand(): void {
    }

    setupHistory(historyPath: string, callback: unknown = null): void {
        if (typeof callback === "function") {
            (callback as Function)(null, this);
        }
    }
}

export function start(options: unknown = null): REPLServer {
    return new REPLServer(options);
}

export default {
    start,
    builtinModules,
    REPLServer,
};
