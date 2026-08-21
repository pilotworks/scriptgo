export class Hash {
    constructor(algorithm: string);
    update(data: string): Hash;
    digest(encoding?: string): string;
}
export function createHash(algorithm: string): Hash;
export function randomUUID(): string;
