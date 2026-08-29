// Node.js Errors & Environment Variables module (node:errors, node:environment_variables)

export class SystemError extends Error {
    name: string = "SystemError";
    message: string = "";
    address?: string;
    code: string = "";
    dest?: string;
    errno: number = 0;
    info?: unknown;
    path?: string;
    port?: number;
    syscall: string = "";

    constructor(message?: string) {
        super(message);
        this.name = "SystemError";
        this.message = message || "";
        this.code = "";
        this.errno = 0;
        this.syscall = "";
    }
}

export class AssertionError extends Error {
    name: string = "AssertionError";
    message: string = "";
    actual?: unknown;
    expected?: unknown;
    operator?: string;
    generatedMessage?: boolean;
    code: string = "ERR_ASSERTION";

    constructor(options?: unknown) {
        super("AssertionError");
        this.name = "AssertionError";
        this.message = "AssertionError";
        this.code = "ERR_ASSERTION";
    }
}

export class CustomError extends Error {
    code?: string;
    constructor(message?: string) {
        super(message);
    }
}

export const env: Record<string, string | undefined> = {};

export default {
    AssertionError,
    SystemError,
    CustomError,
    env,
};
