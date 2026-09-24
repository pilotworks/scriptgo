import { Blob, File } from "node:buffer";

export type FormDataEntryValue = File | string;

export class FormDataEntry {
    name: string;
    value: FormDataEntryValue;

    constructor(name: string, value: FormDataEntryValue) {
        this.name = name;
        this.value = value;
    }
}

export class FormDataIterator<T = unknown> {
    private _values: T[];
    private _index: number = 0;

    constructor(values: T[]) {
        this._values = values;
        this._index = 0;
    }

    next(): { value: T | undefined; done: boolean } {
        if (this._index < this._values.length) {
            const value = this._values[this._index];
            this._index = this._index + 1;
            return { value: value, done: false };
        }
        return { value: undefined, done: true };
    }

    [Symbol.iterator](): FormDataIterator<T> {
        return this;
    }
}

export class FormData {
    private _entries: FormDataEntry[] = [];

    constructor() {
    }

    append(name: string, value: unknown, fileName?: string): void {
        const entryValue = this._toEntryValue(value, fileName);
        this._entries.push(new FormDataEntry(String(name), entryValue));
    }

    delete(name: string): void {
        const next: FormDataEntry[] = [];
        const strName = String(name);
        for (let i = 0; i < this._entries.length; i++) {
            if (this._entries[i].name !== strName) {
                next.push(this._entries[i]);
            }
        }
        this._entries = next;
    }

    get(name: string): FormDataEntryValue | null {
        const strName = String(name);
        for (let i = 0; i < this._entries.length; i++) {
            if (this._entries[i].name === strName) {
                return this._entries[i].value;
            }
        }
        return null;
    }

    getAll(name: string): FormDataEntryValue[] {
        const strName = String(name);
        const res: FormDataEntryValue[] = [];
        for (let i = 0; i < this._entries.length; i++) {
            if (this._entries[i].name === strName) {
                res.push(this._entries[i].value);
            }
        }
        return res;
    }

    has(name: string): boolean {
        const strName = String(name);
        for (let i = 0; i < this._entries.length; i++) {
            if (this._entries[i].name === strName) {
                return true;
            }
        }
        return false;
    }

    set(name: string, value: unknown, fileName?: string): void {
        const strName = String(name);
        const entryValue = this._toEntryValue(value, fileName);
        let found = false;
        const next: FormDataEntry[] = [];
        for (let i = 0; i < this._entries.length; i++) {
            if (this._entries[i].name === strName) {
                if (!found) {
                    next.push(new FormDataEntry(strName, entryValue));
                    found = true;
                }
            } else {
                next.push(this._entries[i]);
            }
        }
        if (!found) {
            next.push(new FormDataEntry(strName, entryValue));
        }
        this._entries = next;
    }

    forEach(callback: (value: FormDataEntryValue, key: string, parent: FormData) => void, thisArg?: unknown): void {
        for (let i = 0; i < this._entries.length; i++) {
            callback(this._entries[i].value, this._entries[i].name, this);
        }
    }

    entries(): FormDataIterator<[string, FormDataEntryValue]> {
        const res: [string, FormDataEntryValue][] = [];
        for (let i = 0; i < this._entries.length; i++) {
            res.push([this._entries[i].name, this._entries[i].value]);
        }
        return new FormDataIterator<[string, FormDataEntryValue]>(res);
    }

    keys(): FormDataIterator<string> {
        const res: string[] = [];
        for (let i = 0; i < this._entries.length; i++) {
            res.push(this._entries[i].name);
        }
        return new FormDataIterator<string>(res);
    }

    values(): FormDataIterator<FormDataEntryValue> {
        const res: FormDataEntryValue[] = [];
        for (let i = 0; i < this._entries.length; i++) {
            res.push(this._entries[i].value);
        }
        return new FormDataIterator<FormDataEntryValue>(res);
    }

    [Symbol.iterator](): FormDataIterator<[string, FormDataEntryValue]> {
        return this.entries();
    }

    private _toEntryValue(value: unknown, fileName?: string): FormDataEntryValue {
        if (value instanceof File) {
            if (fileName !== undefined && fileName !== null && fileName !== "") {
                return new File([value as File], fileName, {
                    type: (value as File).type,
                    lastModified: (value as File).lastModified
                });
            }
            return value as File;
        }
        if (value instanceof Blob) {
            const name = (fileName !== undefined && fileName !== null && fileName !== "") ? fileName : "blob";
            return new File([value as Blob], name, { type: (value as Blob).type });
        }
        return String(value);
    }
}
