// ScriptGo Standard Library: node:string_decoder

function bufferFromBytes(bytes: number[]): Buffer {
    const len = bytes.length;
    const buf = Buffer.alloc(len);
    for (let i = 0; i < len; i++) {
        buf.writeUInt8(bytes[i], i);
    }
    return buf;
}

export class StringDecoder {
    encoding: string;
    private _buffered: number[] = [];

    constructor(encoding: string = "utf8") {
        let enc = encoding.toLowerCase();
        if (enc === "utf8" || enc === "utf-8") {
            this.encoding = "utf8";
        } else if (enc === "base64" || enc === "base64url") {
            this.encoding = enc;
        } else if (enc === "hex") {
            this.encoding = "hex";
        } else if (enc === "ascii") {
            this.encoding = "ascii";
        } else if (enc === "latin1" || enc === "binary") {
            this.encoding = "latin1";
        } else if (enc === "utf16le" || enc === "ucs2" || enc === "ucs-2") {
            this.encoding = "utf16le";
        } else {
            this.encoding = encoding;
        }
    }

    private _getBytes(data: Buffer | Uint8Array): number[] {
        const res: number[] = [];
        const len = data.length;
        for (let i = 0; i < len; i++) {
            res.push(data[i]);
        }
        return res;
    }

    write(data: Buffer | Uint8Array): string {
        const incoming = this._getBytes(data);
        if (incoming.length === 0) {
            return "";
        }

        const total: number[] = [];
        for (let i = 0; i < this._buffered.length; i++) {
            total.push(this._buffered[i]);
        }
        for (let i = 0; i < incoming.length; i++) {
            total.push(incoming[i]);
        }
        this._buffered = [];

        if (this.encoding === "utf8") {
            return this._writeUtf8(total);
        } else if (this.encoding === "base64" || this.encoding === "base64url") {
            return this._writeBase64(total);
        } else if (this.encoding === "hex") {
            return this._writeHex(total);
        } else if (this.encoding === "utf16le") {
            return this._writeUtf16(total);
        } else {
            const buf = bufferFromBytes(total);
            return buf.toString(this.encoding);
        }
    }

    private _writeUtf8(bytes: number[]): string {
        const totalLen = bytes.length;
        let i = 0;
        while (i < totalLen) {
            const b = bytes[i];
            let seqLen = 1;
            if (b < 0x80) {
                seqLen = 1;
            } else if ((b & 0xE0) === 0xC0) {
                seqLen = 2;
            } else if ((b & 0xF0) === 0xE0) {
                seqLen = 3;
            } else if ((b & 0xF8) === 0xF0) {
                seqLen = 4;
            } else {
                seqLen = 1;
            }

            if (i + seqLen <= totalLen) {
                i += seqLen;
            } else {
                break;
            }
        }

        const completeBytes: number[] = [];
        for (let k = 0; k < i; k++) {
            completeBytes.push(bytes[k]);
        }
        for (let k = i; k < totalLen; k++) {
            this._buffered.push(bytes[k]);
        }

        if (completeBytes.length === 0) {
            return "";
        }
        const buf = bufferFromBytes(completeBytes);
        return buf.toString("utf8");
    }

    private _writeBase64(bytes: number[]): string {
        const rem = bytes.length % 3;
        const completeLen = bytes.length - rem;
        const completeBytes: number[] = [];
        for (let k = 0; k < completeLen; k++) {
            completeBytes.push(bytes[k]);
        }
        for (let k = completeLen; k < bytes.length; k++) {
            this._buffered.push(bytes[k]);
        }

        if (completeBytes.length === 0) {
            return "";
        }
        const buf = bufferFromBytes(completeBytes);
        return buf.toString(this.encoding);
    }

    private _writeHex(bytes: number[]): string {
        const buf = bufferFromBytes(bytes);
        return buf.toString("hex");
    }

    private _writeUtf16(bytes: number[]): string {
        const rem = bytes.length % 2;
        const completeLen = bytes.length - rem;
        const completeBytes: number[] = [];
        for (let k = 0; k < completeLen; k++) {
            completeBytes.push(bytes[k]);
        }
        for (let k = completeLen; k < bytes.length; k++) {
            this._buffered.push(bytes[k]);
        }

        if (completeBytes.length === 0) {
            return "";
        }
        const buf = bufferFromBytes(completeBytes);
        return buf.toString("utf16le");
    }

    end(data?: Buffer | Uint8Array): string {
        let res = "";
        if (data !== undefined) {
            res += this.write(data);
        }
        if (this._buffered.length > 0) {
            const buf = bufferFromBytes(this._buffered);
            res += buf.toString(this.encoding);
            this._buffered = [];
        }
        return res;
    }
}

export default {
    StringDecoder,
};
