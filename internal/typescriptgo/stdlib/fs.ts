declare namespace __scriptgo {
    function readFileSync(path: string, encoding?: string): string;
    function writeFileSync(path: string, data: string): void;
    function existsSync(path: string): boolean;
}

export function readFileSync(path: string, encoding?: string): string {
    return __scriptgo.readFileSync(path, encoding);
}

export function writeFileSync(path: string, data: string): void {
    __scriptgo.writeFileSync(path, data);
}

export function existsSync(path: string): boolean {
    return __scriptgo.existsSync(path);
}
