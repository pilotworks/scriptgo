interface ProcessEnv {
    [key: string]: string | undefined;
}

interface Process {
    argv: string[];
    env: ProcessEnv;
    exit(code?: number): void;
    cwd(): string;
}

declare var process: Process;
