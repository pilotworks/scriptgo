export class EventEmitter {
    static defaultMaxListeners: number;
    static listenerCount(emitter: EventEmitter, event: string): number;
    addListener(event: string, listener: Function): EventEmitter;
    on(event: string, listener: Function): EventEmitter;
    once(event: string, listener: Function): EventEmitter;
    prependListener(event: string, listener: Function): EventEmitter;
    prependOnceListener(event: string, listener: Function): EventEmitter;
    removeListener(event: string, listener: Function): EventEmitter;
    off(event: string, listener: Function): EventEmitter;
    removeAllListeners(event?: string): EventEmitter;
    setMaxListeners(n: number): EventEmitter;
    getMaxListeners(): number;
    listenerCount(event: string): number;
    listeners(event: string): Function[];
    rawListeners(event: string): Function[];
    eventNames(): string[];
    emit(event: string, arg1?: unknown, arg2?: unknown, arg3?: unknown, arg4?: unknown): boolean;
}
