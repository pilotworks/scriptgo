declare namespace __scriptgo {
    function exit(code: number): void;
    function cwd(): string;
}

export function exit(code: number): void {
    __scriptgo.exit(code);
}

export function cwd(): string {
    return __scriptgo.cwd();
}

export const argv: string[] = ["scriptgo"];
export const env: Record<string, string | undefined> = { NODE_ENV: "production" };


