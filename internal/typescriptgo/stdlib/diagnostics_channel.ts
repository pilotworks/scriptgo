// Node.js Diagnostics Channel module (node:diagnostics_channel)

export class Channel {
    name: string = "";
    private _subscribers: ((msg: unknown, name: string) => void)[] = [];

    constructor(name: string) {
        this.name = name;
        this._subscribers = [];
    }

    hasSubscribers(): boolean {
        return this._subscribers.length > 0;
    }

    publish(message: unknown): void {
        for (const sub of this._subscribers) {
            sub(message, this.name);
        }
    }

    subscribe(subscription: (msg: unknown, name: string) => void): void {
        this._subscribers.push(subscription);
    }

    unsubscribe(subscription: (msg: unknown, name: string) => void): boolean {
        const idx = this._subscribers.indexOf(subscription);
        if (idx !== -1) {
            this._subscribers.splice(idx, 1);
            return true;
        }
        return false;
    }
}

export class TracingChannel {
    name: string = "";
    start: Channel;
    end: Channel;
    asyncStart: Channel;
    asyncEnd: Channel;
    error: Channel;

    constructor(nameOrChannels: unknown) {
        if (typeof nameOrChannels === "string") {
            const prefix = nameOrChannels as string;
            this.name = prefix;
            this.start = new Channel(`${prefix}:start`);
            this.end = new Channel(`${prefix}:end`);
            this.asyncStart = new Channel(`${prefix}:asyncStart`);
            this.asyncEnd = new Channel(`${prefix}:asyncEnd`);
            this.error = new Channel(`${prefix}:error`);
        } else {
            this.name = "";
            this.start = new Channel("start");
            this.end = new Channel("end");
            this.asyncStart = new Channel("asyncStart");
            this.asyncEnd = new Channel("asyncEnd");
            this.error = new Channel("error");
        }
    }

    traceSync<R>(fn: () => R, context: unknown = {}): R {
        this.start.publish(context);
        try {
            const res: R = fn();
            this.end.publish(context);
            return res;
        } catch (err) {
            this.error.publish(err);
            throw err;
        }
    }

    tracePromise<R>(fn: () => Promise<R>, context: unknown = {}): Promise<R> {
        this.start.publish(context);
        try {
            const p = fn();
            return p.then((res: R) => {
                this.end.publish(context);
                return res;
            });
        } catch (err) {
            this.error.publish(err);
            throw err;
        }
    }

    traceCallback<R>(fn: () => R, position: number, context: unknown = {}): R {
        this.start.publish(context);
        const res: R = fn();
        this.end.publish(context);
        return res;
    }
}

const _channels: Map<string, Channel> = new Map();

export function channel(name: string): Channel {
    let ch = _channels.get(name);
    if (!ch) {
        ch = new Channel(name);
        _channels.set(name, ch);
    }
    return ch;
}

export function tracingChannel(nameOrChannels: unknown): TracingChannel {
    return new TracingChannel(nameOrChannels);
}

export function hasSubscribers(name: string): boolean {
    return channel(name).hasSubscribers();
}

export function subscribe(name: string, subscription: (msg: unknown, name: string) => void): void {
    channel(name).subscribe(subscription);
}

export function unsubscribe(name: string, subscription: (msg: unknown, name: string) => void): boolean {
    return channel(name).unsubscribe(subscription);
}

export function start(name: string, context: unknown): void {
    channel(`${name}:start`).publish(context);
}

export function end(name: string, context: unknown): void {
    channel(`${name}:end`).publish(context);
}

export function asyncStart(name: string, context: unknown): void {
    channel(`${name}:asyncStart`).publish(context);
}

export function asyncEnd(name: string, context: unknown): void {
    channel(`${name}:asyncEnd`).publish(context);
}

export function error(name: string, context: unknown): void {
    channel(`${name}:error`).publish(context);
}

export default {
    Channel,
    TracingChannel,
    channel,
    tracingChannel,
    hasSubscribers,
    subscribe,
    unsubscribe,
    start,
    end,
    asyncStart,
    asyncEnd,
    error,
};
