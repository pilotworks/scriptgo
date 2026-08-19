declare namespace __scriptgo {
    function exit(code: number): void;
    function cwd(): string;
    const argv: string[];
}

export function exit(code: number): void {
    __scriptgo.exit(code);
}

export function cwd(): string {
    return __scriptgo.cwd();
}

export const argv: string[] = __scriptgo.argv;
