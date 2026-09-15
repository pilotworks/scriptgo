// ScriptGo Standard Library: WebSocket RFC 6455 Client Implementation
// Compliant with WHATWG WebSocket & W3C EventTarget standards.

import { Event, EventTarget } from "events";

export class CloseEventInit {
    code?: number;
    reason?: string;
    wasClean?: boolean;
}

export class MessageEventInit<T = unknown> {
    data?: T;
    origin?: string;
    lastEventId?: string;
}

export class CloseEvent extends Event {
    readonly code: number;
    readonly reason: string;
    readonly wasClean: boolean;

    constructor(type: string, eventInitDict: CloseEventInit = new CloseEventInit()) {
        super(type, { bubbles: false, cancelable: false });
        this.code = eventInitDict.code !== undefined ? eventInitDict.code : 1000;
        this.reason = eventInitDict.reason !== undefined ? eventInitDict.reason : "";
        this.wasClean = eventInitDict.wasClean !== undefined ? eventInitDict.wasClean : (this.code === 1000);
    }
}

export class MessageEvent<T = unknown> extends Event {
    readonly data: T;
    readonly origin: string;
    readonly lastEventId: string;

    constructor(type: string, eventInitDict: MessageEventInit<T> = new MessageEventInit<T>()) {
        super(type, { bubbles: false, cancelable: false });
        this.data = eventInitDict.data as T;
        this.origin = eventInitDict.origin !== undefined ? eventInitDict.origin : "";
        this.lastEventId = eventInitDict.lastEventId !== undefined ? eventInitDict.lastEventId : "";
    }
}

export class WebSocketPollData {
    eventType: number;
    data: string;
    code: number;
    reason: string;

    constructor(eventType: number, data: string, code: number, reason: string) {
        this.eventType = eventType;
        this.data = data;
        this.code = code;
        this.reason = reason;
    }
}

declare namespace __scriptgo {
    function websocketConnect(url: string, protocol?: string): number;
    function websocketSendText(handle: number, data: string): number;
    function websocketClose(handle: number, code?: number, reason?: string): void;
    function websocketPoll(handle: number): WebSocketPollData;
    function websocketReadyState(handle: number): number;
}

export class WebSocket extends EventTarget {
    static readonly CONNECTING: number = 0;
    static readonly OPEN: number = 1;
    static readonly CLOSING: number = 2;
    static readonly CLOSED: number = 3;

    readonly CONNECTING: number = 0;
    readonly OPEN: number = 1;
    readonly CLOSING: number = 2;
    readonly CLOSED: number = 3;

    readonly url: string;
    readonly protocol: string;
    readonly extensions: string = "";
    readonly bufferedAmount: number = 0;
    binaryType: string = "blob";

    onclose: ((this: WebSocket, ev: CloseEvent) => unknown) | null = null;
    onerror: ((this: WebSocket, ev: Event) => unknown) | null = null;
    onmessage: ((this: WebSocket, ev: MessageEvent) => unknown) | null = null;
    onopen: ((this: WebSocket, ev: Event) => unknown) | null = null;

    private _handle: number;
    private _timerId: number = 0;

    constructor(url: string | URL, protocols?: string | string[]) {
        super();
        const urlStr = typeof url === "string" ? url : (url ? url.toString() : "");
        this.url = urlStr;
        let proto = "";
        if (typeof protocols === "string") {
            proto = protocols;
        } else if (protocols !== undefined && protocols !== null && Array.isArray(protocols) && protocols.length > 0) {
            proto = protocols[0];
        }
        this.protocol = proto;

        this._handle = __scriptgo.websocketConnect(urlStr, proto);
        this._startPolling();
    }

    get readyState(): number {
        return __scriptgo.websocketReadyState(this._handle);
    }

    send(data: string | ArrayBufferLike | Blob | ArrayBufferView): void {
        if (this.readyState !== WebSocket.OPEN) {
            throw new Error("InvalidStateError: WebSocket is not open");
        }
        const text = typeof data === "string" ? data : String(data);
        __scriptgo.websocketSendText(this._handle, text);
    }

    close(code: number = 1000, reason: string = ""): void {
        if (this.readyState === WebSocket.CLOSING || this.readyState === WebSocket.CLOSED) {
            return;
        }
        __scriptgo.websocketClose(this._handle, code, reason);
    }

    private _startPolling(): void {
        const check = () => {
            const poll = __scriptgo.websocketPoll(this._handle);
            if (poll.eventType === 1) {
                // 1 = open
                const ev = new Event("open");
                if (this.onopen) {
                    this.onopen.call(this, ev);
                }
                this.dispatchEvent(ev);
            } else if (poll.eventType === 2) {
                // 2 = message
                const init = new MessageEventInit<string>();
                init.data = poll.data;
                init.origin = this.url;
                const msgEv = new MessageEvent("message", init);
                if (this.onmessage) {
                    this.onmessage.call(this, msgEv);
                }
                this.dispatchEvent(msgEv);
            } else if (poll.eventType === 3) {
                // 3 = close
                if (this._timerId !== 0) {
                    clearInterval(this._timerId);
                    this._timerId = 0;
                }
                const init = new CloseEventInit();
                init.code = poll.code;
                init.reason = poll.reason;
                init.wasClean = poll.code === 1000;
                const closeEv = new CloseEvent("close", init);
                if (this.onclose) {
                    this.onclose.call(this, closeEv);
                }
                this.dispatchEvent(closeEv);
            } else if (poll.eventType === 4) {
                // 4 = error
                const errEv = new Event("error");
                if (this.onerror) {
                    this.onerror.call(this, errEv);
                }
                this.dispatchEvent(errEv);
            }

            if (this.readyState === WebSocket.CLOSED) {
                if (this._timerId !== 0) {
                    clearInterval(this._timerId);
                    this._timerId = 0;
                }
            }
        };

        this._timerId = setInterval(check, 10) as unknown as number;
    }
}

export default {
    WebSocket,
    CloseEvent,
    MessageEvent,
};
