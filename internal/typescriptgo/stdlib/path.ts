function trimTrailingSeparators(path: string): string {
    return path.length == 0 ? path : (path.lastIndexOf("/") == path.length - 1 ? path.slice(0, path.length - 1) : path);
}

function joinPair(first: string, second: string): string {
	const left: string = trimTrailingSeparators(first);
	const secondStart: number = second.lastIndexOf("/") == 0 ? 1 : 0;
	const right: string = trimTrailingSeparators(second.slice(secondStart, second.length));
	return left.length == 0 ? (right.length == 0 ? "." : right) : (right.length == 0 ? left : left + "/" + right);
}

function joinRest(current: string, rest: string[], index: number): string {
	if (index == rest.length) {
		return current;
	}
	return joinRest(joinPair(current, rest[index]), rest, index + 1);
}

export function join(first: string, second: string, ...rest: string[]): string {
	return joinRest(joinPair(first, second), rest, 0);
}

export function dirname(path: string): string {
    const trimmed: string = trimTrailingSeparators(path);
    const index: number = trimmed.lastIndexOf("/");
    return trimmed.length == 0 ? "/" : (index < 0 ? "." : (index == 0 ? "/" : trimmed.slice(0, index)));
}

export function basename(path: string): string {
    const trimmed: string = trimTrailingSeparators(path);
    const index: number = trimmed.lastIndexOf("/");
    return index < 0 ? trimmed : trimmed.slice(index + 1, trimmed.length);
}

export function extname(path: string): string {
    const name: string = basename(path);
    const index: number = name.lastIndexOf(".");
    return index <= 0 ? "" : name.slice(index, name.length);
}
