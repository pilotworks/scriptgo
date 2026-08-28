// ScriptGo Standard Library: node:punycode

export class Ucs2 {
    decode(str: string): number[] {
        const output: number[] = [];
        for (let i = 0; i < str.length; i++) {
            output.push(str.charCodeAt(i));
        }
        return output;
    }

    encode(codePoints: number[]): string {
        let res = "";
        for (let i = 0; i < codePoints.length; i++) {
            res += String.fromCharCode(codePoints[i]);
        }
        return res;
    }
}

export const ucs2 = new Ucs2();

function basicToDigit(codePoint: number): number {
    if (codePoint >= 48 && codePoint <= 57) {
        return codePoint - 22;
    }
    if (codePoint >= 65 && codePoint <= 90) {
        return codePoint - 65;
    }
    if (codePoint >= 97 && codePoint <= 122) {
        return codePoint - 97;
    }
    return 36;
}

function digitToBasic(digit: number): number {
    if (digit < 26) {
        return digit + 97; // 'a'..'z'
    }
    return digit + 22; // '0'..'9'
}

function adapt(delta: number, numPoints: number, firstTime: boolean): number {
    const base = 36;
    const tmin = 1;
    const tmax = 26;
    const skew = 38;
    const damp = 700;

    let k = 0;
    let d = firstTime ? Math.floor(delta / damp) : Math.floor(delta / 2);
    d += Math.floor(d / numPoints);
    while (d > ((base - tmin) * tmax) / 2) {
        d = Math.floor(d / (base - tmin));
        k += base;
    }
    return Math.floor(k + ((base - tmin + 1) * d) / (d + skew));
}

function arrayInsert(arr: number[], index: number, item: number): void {
    const len = arr.length;
    arr.push(item);
    for (let j = len; j > index; j--) {
        arr[j] = arr[j - 1];
    }
    arr[index] = item;
}

export function decode(input: string): string {
    const base = 36;
    const tmin = 1;
    const tmax = 26;
    const initialBias = 72;
    const initialN = 128;
    const delimiter = "-";

    const output: number[] = [];
    const inputLength = input.length;
    let n = initialN;
    let i = 0;
    let bias = initialBias;

    let basic = input.lastIndexOf(delimiter);
    if (basic < 0) {
        basic = 0;
    }

    for (let j = 0; j < basic; ++j) {
        if (input.charCodeAt(j) >= 0x80) {
            throw new RangeError("Illegal input >= 0x80");
        }
        output.push(input.charCodeAt(j));
    }

    let index = basic > 0 ? basic + 1 : 0;
    while (index < inputLength) {
        const oldi = i;
        let w = 1;
        for (let k = base; true; k += base) {
            if (index >= inputLength) {
                throw new RangeError("Invalid punycode input");
            }
            const digit = basicToDigit(input.charCodeAt(index));
            index++;
            if (digit >= base) {
                throw new RangeError("Invalid punycode digit");
            }
            i += digit * w;
            const t = k <= bias ? tmin : (k >= bias + tmax ? tmax : k - bias);
            if (digit < t) {
                break;
            }
            w = w * (base - t);
        }
        const outLen = output.length + 1;
        bias = adapt(i - oldi, outLen, oldi === 0);
        n += Math.floor(i / outLen);
        i %= outLen;
        arrayInsert(output, i, n);
        i++;
    }

    return ucs2.encode(output);
}

export function encode(input: string): string {
    const maxInt = 2147483647;
    const base = 36;
    const tmin = 1;
    const tmax = 26;
    const initialBias = 72;
    const initialN = 128;
    const delimiter = "-";

    const inputCodePoints = ucs2.decode(input);
    const inputLength = inputCodePoints.length;

    let n = initialN;
    let delta = 0;
    let bias = initialBias;
    let output = "";

    for (let j = 0; j < inputLength; ++j) {
        const currentValue = inputCodePoints[j];
        if (currentValue < 0x80) {
            output += String.fromCharCode(currentValue);
        }
    }

    const basicLength = output.length;
    let handledCPCount = basicLength;

    if (basicLength > 0) {
        output += delimiter;
    }

    while (handledCPCount < inputLength) {
        let m = maxInt;
        for (let j = 0; j < inputLength; ++j) {
            const currentValue = inputCodePoints[j];
            if (currentValue >= n && currentValue < m) {
                m = currentValue;
            }
        }

        delta += (m - n) * (handledCPCount + 1);
        n = m;

        for (let j = 0; j < inputLength; ++j) {
            const currentValue = inputCodePoints[j];
            if (currentValue < n) {
                delta++;
            }
            if (currentValue === n) {
                let q = delta;
                for (let k = base; true; k += base) {
                    const t = k <= bias ? tmin : (k >= bias + tmax ? tmax : k - bias);
                    if (q < t) {
                        break;
                    }
                    const qMinusT = q - t;
                    const baseMinusT = base - t;
                    output += String.fromCharCode(digitToBasic(t + (qMinusT % baseMinusT)));
                    q = Math.floor(qMinusT / baseMinusT);
                }
                output += String.fromCharCode(digitToBasic(q));
                bias = adapt(delta, handledCPCount + 1, handledCPCount === basicLength);
                delta = 0;
                handledCPCount++;
            }
        }
        delta++;
        n++;
    }

    return output;
}

function hasNonAscii(str: string): boolean {
    for (let i = 0; i < str.length; i++) {
        if (str.charCodeAt(i) > 0x7F) {
            return true;
        }
    }
    return false;
}

export function toASCII(domain: string): string {
    const labels = domain.split(".");
    const result: string[] = [];
    for (let i = 0; i < labels.length; i++) {
        const label = labels[i];
        if (hasNonAscii(label)) {
            result.push("xn--" + encode(label));
        } else {
            result.push(label);
        }
    }
    return result.join(".");
}

export function toUnicode(domain: string): string {
    const labels = domain.split(".");
    const result: string[] = [];
    for (let i = 0; i < labels.length; i++) {
        let label = labels[i];
        if (label.startsWith("xn--") || label.startsWith("XN--")) {
            result.push(decode(label.substring(4)));
        } else {
            result.push(label);
        }
    }
    return result.join(".");
}

export const version = "2.1.0";

export class Punycode {
    version: string = "2.1.0";
    ucs2: Ucs2 = ucs2;

    decode(input: string): string {
        return decode(input);
    }

    encode(input: string): string {
        return encode(input);
    }

    toASCII(domain: string): string {
        return toASCII(domain);
    }

    toUnicode(domain: string): string {
        return toUnicode(domain);
    }
}

export const punycode = new Punycode();

export default punycode;
