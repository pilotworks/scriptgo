export type NextFunction = (err?: Error | null) => void;

export interface RequestLike {
    method: string;
    url: string;
    path: string;
    params: Record<string, string>;
    query: Record<string, string>;
    headers: Record<string, string>;
    body: unknown;
    get(name: string): string | undefined;
}

export interface ResponseLike {
    status(code: number): ResponseLike;
    setHeader(name: string, value: string): ResponseLike;
    getHeader(name: string): string | undefined;
    json(data: unknown): void;
    send(body: string): void;
    redirect(url: string, status?: number): void;
    end(): void;
}

export type Handler = (req: RequestLike, res: ResponseLike, next: NextFunction) => void;

export type ErrorHandler = (err: Error, req: RequestLike, res: ResponseLike, next: NextFunction) => void;

export interface RouteEntry {
    method: string;
    path: string;
    segments: string[];
    handler: Handler;
}
