// Node.js Test Runner module (node:test)

export class MockFunctionContext {
    calls: unknown[] = [];

    constructor() {
        this.calls = [];
    }

    callCount(): number {
        return this.calls.length;
    }

    mockImplementation(fn: unknown): void {}
    mockImplementationOnce(fn: unknown, onCall?: number): void {}
    resetCalls(): void {
        this.calls = [];
    }
    restore(): void {}
}

export class MockPropertyContext {
    accesses: unknown[] = [];

    constructor() {
        this.accesses = [];
    }

    accessCount(): number {
        return this.accesses.length;
    }

    mockImplementation(fn: unknown): void {}
    mockImplementationOnce(fn: unknown, onCall?: number): void {}
    resetAccesses(): void {
        this.accesses = [];
    }
    restore(): void {}
}

export class MockModuleContext {
    constructor() {}
    restore(): void {}
}

export class MockTimers {
    constructor() {}

    enable(options?: unknown): void {}
    reset(): void {}
    [Symbol.dispose](): void {
        this.reset();
    }
    tick(ms: number): void {}
    runAll(): void {}
    setTime(time: number): void {}
}

export class MockTracker {
    timers: MockTimers = new MockTimers();

    constructor() {
        this.timers = new MockTimers();
    }

    fn(original?: unknown, implementation?: unknown, options?: unknown): unknown {
        return original || (() => {});
    }

    getter(object: unknown, property: string, implementation?: unknown, options?: unknown): MockPropertyContext {
        return new MockPropertyContext();
    }

    method(object: unknown, method: string, implementation?: unknown, options?: unknown): MockFunctionContext {
        return new MockFunctionContext();
    }

    module(specifier: string, options?: unknown): MockModuleContext {
        return new MockModuleContext();
    }

    property(object: unknown, property: string, value?: unknown, options?: unknown): MockPropertyContext {
        return new MockPropertyContext();
    }

    reset(): void {}
    restoreAll(): void {}

    setter(object: unknown, property: string, implementation?: unknown, options?: unknown): MockPropertyContext {
        return new MockPropertyContext();
    }
}

export const mock: MockTracker = new MockTracker();

export class TestsStream {
    constructor() {}
}

export class TestContext {
    name: string = "test";
    fullName: string = "test";
    filePath: string = "";
    passed: boolean = true;
    error: unknown = null;
    attempt: number = 1;
    signal: unknown = null;
    assert: Record<string, unknown> = {};

    constructor(name: string = "test") {
        this.name = name;
        this.fullName = name;
        this.filePath = "";
        this.passed = true;
        this.error = null;
        this.attempt = 1;
        this.signal = null;
        this.assert = {};
    }

    before(fn: unknown): void {}
    beforeEach(fn: unknown): void {}
    after(fn: unknown): void {}
    afterEach(fn: unknown): void {}
    diagnostic(message: string): void {}
    plan(count: number): void {}
    runOnly(shouldRunOnlyTests: boolean): void {}
    skip(message?: string): void {}
    todo(message?: string): void {}

    test(name?: unknown, options?: unknown, fn?: unknown): void {
        const callback = typeof name === "function" ? name : (typeof options === "function" ? options : fn);
        if (typeof callback === "function") {
            (callback as Function)(this);
        }
    }

    async waitFor(condition: unknown, options?: unknown): Promise<void> {}
}

export class SuiteContext {
    name: string = "suite";
    fullName: string = "suite";
    filePath: string = "";
    signal: unknown = null;

    constructor(name: string = "suite") {
        this.name = name;
        this.fullName = name;
        this.filePath = "";
        this.signal = null;
    }
}

export function test(name?: unknown, options?: unknown, fn?: unknown): void {
    const callback = typeof name === "function" ? name : (typeof options === "function" ? options : fn);
    if (typeof callback === "function") {
        const t = new TestContext(typeof name === "string" ? name : "test");
        (callback as Function)(t);
    }
}

export function suite(name?: unknown, options?: unknown, fn?: unknown): void {
    const callback = typeof name === "function" ? name : (typeof options === "function" ? options : fn);
    if (typeof callback === "function") {
        const s = new SuiteContext(typeof name === "string" ? name : "suite");
        (callback as Function)(s);
    }
}

export function describe(name?: unknown, options?: unknown, fn?: unknown): void {
    suite(name, options, fn);
}

export function it(name?: unknown, options?: unknown, fn?: unknown): void {
    test(name, options, fn);
}

export function skip(name?: unknown, options?: unknown, fn?: unknown): void {}
export function todo(name?: unknown, options?: unknown, fn?: unknown): void {}
export function only(name?: unknown, options?: unknown, fn?: unknown): void {
    test(name, options, fn);
}

export function before(fn: unknown): void {}
export function after(fn: unknown): void {}
export function beforeEach(fn: unknown): void {}
export function afterEach(fn: unknown): void {}

export function run(options?: unknown): TestsStream {
    return new TestsStream();
}

export function register(name: string, fn: unknown): void {}
export function setDefaultSnapshotSerializers(serializers: unknown[]): void {}
export function setResolveSnapshotPath(fn: unknown): void {}

export default {
    MockFunctionContext,
    MockPropertyContext,
    MockModuleContext,
    MockTimers,
    MockTracker,
    mock,
    TestsStream,
    TestContext,
    SuiteContext,
    test,
    suite,
    describe,
    it,
    skip,
    todo,
    only,
    before,
    after,
    beforeEach,
    afterEach,
    run,
    register,
    setDefaultSnapshotSerializers,
    setResolveSnapshotPath,
};
