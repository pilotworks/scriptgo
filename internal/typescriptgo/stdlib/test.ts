// ScriptGo Standard Library: node:test

import { AbortSignal } from "node:events";
import { Readable } from "node:stream";

declare namespace __scriptgo {
  function exit(code: number): void;
}

export interface TestOptions {
  concurrency?: number | boolean;
  only?: boolean;
  signal?: AbortSignal;
  skip?: boolean | string;
  timeout?: number;
  todo?: boolean | string;
}

export interface SuiteOptions {
  concurrency?: number | boolean;
  only?: boolean;
  signal?: AbortSignal;
  skip?: boolean | string;
  timeout?: number;
  todo?: boolean | string;
}

export interface RunOptions {
  concurrency?: number | boolean;
  files?: string[];
  only?: boolean;
  setup?: Function;
  timeout?: number;
  signal?: unknown;
  testNamePatterns?: unknown;
}

export type TestFn = (
  t: TestContext,
  done?: (err?: unknown) => void,
) => void | Promise<void>;
export type SuiteFn = (s: SuiteContext) => void | Promise<void>;
export type HookFn = (s: SuiteContext | TestContext) => void | Promise<void>;

export interface MockFunctionOptions {
  times?: number;
}

export interface MockCall {
  arguments: unknown[];
  result: unknown;
  error: unknown;
  target: unknown;
  this: unknown;
}

export class MockFunctionContext {
  calls: MockCall[] = [];
  private _target: Function;
  private _impl: Function | null = null;
  private _onceImpls: Function[] = [];
  private _onceCalls: number[] = [];
  private _restoreFn: () => void;

  constructor(target: Function, restoreFn: () => void) {
    this._target = target;
    this._restoreFn = restoreFn;
  }

  callCount(): number {
    return this.calls.length;
  }

  mockImplementation(fn: Function): void {
    this._impl = fn;
  }

  mockImplementationOnce(fn: Function, onCall: number = -1): void {
    this._onceImpls.push(fn);
    this._onceCalls.push(onCall);
  }

  resetCalls(): void {
    this.calls = [];
  }

  restore(): void {
    this._restoreFn();
  }

  _invoke(thisArg: unknown, args: unknown[]): unknown {
    let impl = this._impl;
    const callIndex = this.calls.length;

    for (let i = 0; i < this._onceImpls.length; i++) {
      const targetCall = this._onceCalls[i];
      if (targetCall < 0 || targetCall === callIndex) {
        impl = this._onceImpls[i];
        this._onceImpls.splice(i, 1);
        this._onceCalls.splice(i, 1);
        break;
      }
    }

    let result: unknown = undefined;
    let error: unknown = undefined;
    try {
      if (impl) {
        result = Reflect.apply(impl, thisArg, args);
      } else if (this._target) {
        result = Reflect.apply(this._target, thisArg, args);
      }
    } catch (err: unknown) {
      error = err;
      this.calls.push({
        arguments: args,
        result: undefined,
        error,
        target: this._target,
        this: thisArg,
      });
      throw err;
    }

    this.calls.push({
      arguments: args,
      result,
      error: undefined,
      target: this._target,
      this: thisArg,
    });
    return result;
  }
}

export interface MockAccess {
  type: "get" | "set";
  value: unknown;
  stack?: Error;
}

export class MockPropertyContext {
  accesses: MockAccess[] = [];
  private _target: Record<string, unknown>;
  private _prop: string;
  private _origVal: unknown;
  private _val: unknown;
  private _onceVals: unknown[] = [];
  private _onceAccesses: number[] = [];

  constructor(target: Record<string, unknown>, prop: string, val?: unknown) {
    this._target = target;
    this._prop = prop;
    this._origVal = target[prop];
    this._val = val !== undefined ? val : this._origVal;
    target[prop] = this._val;
  }

  accessCount(): number {
    return this.accesses.length;
  }

  mockImplementation(value: unknown): void {
    this._val = value;
    this._target[this._prop] = value;
  }

  mockImplementationOnce(value: unknown, onAccess: number = -1): void {
    this._onceVals.push(value);
    this._onceAccesses.push(onAccess);
  }

  resetAccesses(): void {
    this.accesses = [];
  }

  restore(): void {
    this._target[this._prop] = this._origVal;
  }
}

export interface MockTimerEntry {
  id: number;
  fireTime: number;
  callback: Function;
  isInterval: boolean;
  delay: number;
}

export class MockTimers {
  private _enabled: boolean = false;
  private _now: number = 0;
  private _queue: MockTimerEntry[] = [];
  private _nextId: number = 1;

  enable(options?: { apis?: string[] }): void {
    this._enabled = true;
    this._now = 0;
    this._queue = [];
  }

  reset(): void {
    this._enabled = false;
    this._now = 0;
    this._queue = [];
  }

  tick(ms: number): void {
    if (!this._enabled) return;
    this._now += ms;
    this._queue.sort((a, b) => a.fireTime - b.fireTime);
    const remaining: MockTimerEntry[] = [];
    for (const t of this._queue) {
      if (t.fireTime <= this._now) {
        t.callback();
        if (t.isInterval) {
          t.fireTime += t.delay;
          remaining.push(t);
        }
      } else {
        remaining.push(t);
      }
    }
    this._queue = remaining;
  }

  runAll(): void {
    if (!this._enabled) return;
    while (this._queue.length > 0) {
      this._queue.sort((a, b) => a.fireTime - b.fireTime);
      const t = this._queue.shift();
      if (!t) break;
      this._now = t.fireTime;
      t.callback();
    }
  }
}

export type MockFunction = ((...args: unknown[]) => unknown) & { mock: MockFunctionContext };

export class MockTracker {
  private _restores: (() => void)[] = [];
  timers: MockTimers = new MockTimers();

  fn(
    original?: Function,
    implementation?: Function,
    options?: MockFunctionOptions,
  ): MockFunction {
    const target = (implementation || original || (() => {})) as Function;
    const ctx = new MockFunctionContext(target, () => {});

    const proxy = function (a?: unknown, b?: unknown, c?: unknown, d?: unknown) {
      const args: unknown[] = [];
      if (a !== undefined) args.push(a);
      if (b !== undefined) args.push(b);
      if (c !== undefined) args.push(c);
      if (d !== undefined) args.push(d);
      return ctx._invoke(undefined, args);
    };

    const callableMock = {
      _fn: proxy,
      mock: ctx,
    };
    return (callableMock as object) as MockFunction;
  }

  method(
    object: Record<string, unknown>,
    methodName: string,
    implementation?: Function,
    options?: MockFunctionOptions,
  ): MockFunctionContext {
    const original = object[methodName] as Function;
    const ctx = new MockFunctionContext(original, () => {
      object[methodName] = original;
    });
    if (implementation) {
      ctx.mockImplementation(implementation);
    }
    const fn = function (a?: unknown, b?: unknown, c?: unknown, d?: unknown) {
      const args: unknown[] = [];
      if (a !== undefined) args.push(a);
      if (b !== undefined) args.push(b);
      if (c !== undefined) args.push(c);
      if (d !== undefined) args.push(d);
      return ctx._invoke(undefined, args);
    };
    const callableMock = {
      _fn: fn,
      mock: ctx,
    };
    object[methodName] = callableMock;
    this._restores.push(() => ctx.restore());
    return ctx;
  }

  property(
    object: Record<string, unknown>,
    propertyName: string,
    value?: unknown,
  ): MockPropertyContext {
    const ctx = new MockPropertyContext(object, propertyName, value);
    this._restores.push(() => ctx.restore());
    return ctx;
  }

  reset(): void {
    this._restores = [];
  }

  restoreAll(): void {
    for (let i = this._restores.length - 1; i >= 0; i--) {
      this._restores[i]();
    }
    this._restores = [];
  }
}

export const mock = new MockTracker();

export class SuiteContext {
  name: string;
  signal: AbortSignal;
  filePath?: string;

  constructor(name: string, signal?: AbortSignal) {
    this.name = name;
    this.signal = signal || new AbortSignal();
  }
}

export class TestContext {
  name: string;
  signal: AbortSignal;
  filePath?: string;
  mock: MockTracker = new MockTracker();

  private _node: TestNode;

  constructor(name: string, node: TestNode, signal?: AbortSignal) {
    this.name = name;
    this._node = node;
    this.signal = signal || new AbortSignal();
  }

  skip(message?: string): void {
    this._node.skipped = true;
    this._node.skipMessage = message;
  }

  todo(message?: string): void {
    this._node.isTodo = true;
    this._node.todoMessage = message;
  }

  runOnly(enable: boolean): void {
    this._node.only = enable;
  }

  diagnostic(message: string): void {
    console.log("# " + message);
  }

  before(fn: HookFn): void {
    this._node.beforeHooks.push(fn);
  }

  after(fn: HookFn): void {
    this._node.afterHooks.push(fn);
  }

  beforeEach(fn: HookFn): void {
    this._node.beforeEachHooks.push(fn);
  }

  afterEach(fn: HookFn): void {
    this._node.afterEachHooks.push(fn);
  }

  async test(
    nameOrOptionsOrFn?: string | TestOptions | TestFn,
    optionsOrFn?: TestOptions | TestFn,
    fn?: TestFn,
  ): Promise<void> {
    return this._node.runSubtest(nameOrOptionsOrFn, optionsOrFn, fn);
  }
}

function formatDuration(ms: number): string {
  if (process.env["NODE_TEST_DETERMINISTIC"] === "1") {
    return "0.000ms";
  }
  return `${ms.toFixed(3)}ms`;
}

class TestNode {
  name: string;
  isSuite: boolean = false;
  fn?: TestFn | SuiteFn;
  options: TestOptions;
  parent: TestNode | null = null;
  children: TestNode[] = [];
  beforeHooks: HookFn[] = [];
  afterHooks: HookFn[] = [];
  beforeEachHooks: HookFn[] = [];
  afterEachHooks: HookFn[] = [];
  skipped: boolean = false;
  skipMessage?: string;
  isTodo: boolean = false;
  todoMessage?: string;
  only: boolean = false;
  passed: boolean = false;
  failed: boolean = false;
  error: unknown = null;
  durationMs: number = 0;
  executed: boolean = false;

  constructor(
    name: string,
    isSuite: boolean,
    fn?: TestFn | SuiteFn,
    options?: TestOptions,
  ) {
    this.name = name;
    this.isSuite = isSuite;
    this.fn = fn;
    this.options = options || {};
    this.children = [];
    this.beforeHooks = [];
    this.afterHooks = [];
    this.beforeEachHooks = [];
    this.afterEachHooks = [];
    if (this.options.skip !== undefined) {
      this.skipped = true;
      if (typeof this.options.skip === "string") {
        this.skipMessage = this.options.skip;
      }
    }
    if (this.options.todo !== undefined) {
      this.isTodo = true;
      if (typeof this.options.todo === "string") {
        this.todoMessage = this.options.todo;
      }
    }
    if (this.options.only) {
      this.only = true;
    }
  }

  async runSubtest(
    nameOrOptionsOrFn?: string | TestOptions | TestFn,
    optionsOrFn?: TestOptions | TestFn,
    fn?: TestFn,
  ): Promise<void> {
    let subName = "subtest";
    let subOpts: TestOptions = {};
    let subFn: TestFn | undefined;

    if (typeof nameOrOptionsOrFn === "string") {
      subName = nameOrOptionsOrFn;
      if (typeof optionsOrFn === "function") {
        subFn = optionsOrFn as TestFn;
      } else if (optionsOrFn && typeof optionsOrFn === "object") {
        subOpts = optionsOrFn as TestOptions;
        subFn = fn;
      }
    } else if (typeof nameOrOptionsOrFn === "function") {
      subFn = nameOrOptionsOrFn as TestFn;
    } else if (nameOrOptionsOrFn && typeof nameOrOptionsOrFn === "object") {
      subOpts = nameOrOptionsOrFn as TestOptions;
      if (typeof optionsOrFn === "function") {
        subFn = optionsOrFn as TestFn;
      }
    }

    const child = new TestNode(subName, false, subFn, subOpts);
    child.parent = this;
    this.children.push(child);
    await child.execute();
  }

  getBeforeEachHooks(): HookFn[] {
    const hooks: HookFn[] = [];
    let curr: TestNode | null = this.parent;
    const chain: TestNode[] = [];
    while (curr) {
      chain.unshift(curr);
      curr = curr.parent;
    }
    for (const n of chain) {
      hooks.push(...n.beforeEachHooks);
    }
    hooks.push(...this.beforeEachHooks);
    return hooks;
  }

  getAfterEachHooks(): HookFn[] {
    const hooks: HookFn[] = [];
    hooks.push(...this.afterEachHooks);
    let curr: TestNode | null = this.parent;
    while (curr) {
      hooks.push(...curr.afterEachHooks);
      curr = curr.parent;
    }
    return hooks;
  }

  async runHooks(hooks: HookFn[], ctx: SuiteContext | TestContext): Promise<void> {
    for (let i = 0; i < hooks.length; i++) {
      await hooks[i](ctx);
    }
  }

  async runChildren(): Promise<void> {
    for (let i = 0; i < this.children.length; i++) {
      await this.children[i].execute();
    }
  }

  async runSuiteFn(sCtx: SuiteContext): Promise<void> {
    if (this.fn) {
      await (this.fn as SuiteFn)(sCtx);
    }
  }

  async executeSuite(): Promise<void> {
    const indent = this.getIndent();
    const start = Date.now();
    console.log(`${indent}▶ ${this.name}`);
    const prevSuite = currentSuite;
    currentSuite = this;

    const sCtx = new SuiteContext(this.name);
    await this.runSuiteFn(sCtx);
    await this.runHooks(this.beforeHooks, sCtx);
    await this.runChildren();
    await this.runHooks(this.afterHooks, sCtx);

    currentSuite = prevSuite;
    this.durationMs = Date.now() - start;
    this.passed = true;
    console.log(`${indent}✔ ${this.name} (${formatDuration(this.durationMs)})`);
  }

  async invokeFn(tCtx: TestContext): Promise<void> {
    if (this.fn) {
      await this.fn(tCtx);
    }
  }

  async runTestSafely(tCtx: TestContext): Promise<unknown> {
    try {
      await this.invokeFn(tCtx);
      return null;
    } catch (err: unknown) {
      return err;
    }
  }

  async executeTest(): Promise<void> {
    const indent = this.getIndent();
    const start = Date.now();
    const tCtx = new TestContext(this.name, this);

    await this.runHooks(this.getBeforeEachHooks(), tCtx);
    const testError = await this.runTestSafely(tCtx);
    await this.runHooks(this.getAfterEachHooks(), tCtx);

    this.durationMs = Date.now() - start;
    if (testError) {
      if (this.isTodo) {
        const todoSuffix = this.todoMessage ? ` # TODO ${this.todoMessage}` : " # TODO";
        console.log(`${indent}✔ ${this.name} (${formatDuration(this.durationMs)})${todoSuffix}`);
        statsTodo += 1;
      } else {
        this.failed = true;
        this.error = testError;
        statsFail += 1;
        console.log(`${indent}✖ ${this.name} (${formatDuration(this.durationMs)})`);
        failedNodes.push(this);
      }
    } else {
      this.passed = true;
      if (this.isTodo) {
        const todoSuffix = this.todoMessage ? ` # TODO ${this.todoMessage}` : " # TODO";
        console.log(`${indent}✔ ${this.name} (${formatDuration(this.durationMs)})${todoSuffix}`);
        statsTodo += 1;
      } else {
        console.log(`${indent}✔ ${this.name} (${formatDuration(this.durationMs)})`);
        statsPass += 1;
      }
    }
  }

  async execute(): Promise<void> {
    if (this.executed) return;
    this.executed = true;

    if (this.isSuite) {
      await this.executeSuite();
    } else if (this.skipped) {
      const indent = this.getIndent();
      const skipSuffix = this.skipMessage ? ` # SKIP ${this.skipMessage}` : " # SKIP";
      console.log(`${indent}﹣ ${this.name} (0.000ms)${skipSuffix}`);
      statsSkipped += 1;
    } else {
      await this.executeTest();
    }
  }

  getIndent(): string {
    let depth = 0;
    let p = this.parent;
    while (p) {
      depth++;
      p = p.parent;
    }
    let s = "";
    for (let i = 0; i < depth; i++) {
      s += "  ";
    }
    return s;
  }
}

// Global Runner State
let currentSuite: TestNode | null = null;
const rootNodes: TestNode[] = [];
const failedNodes: TestNode[] = [];
let autoRunScheduled = false;
let isRunning = false;
let activeStream: TestsStream | null = null;
let abortRequested = false;

let statsTests = 0;
let statsSuites = 0;
let statsPass = 0;
let statsFail = 0;
let statsCancelled = 0;
let statsSkipped = 0;
let statsTodo = 0;
let statsDurationMs = 0;

function scheduleAutoRun(): void {
  if (autoRunScheduled) return;
  autoRunScheduled = true;
  queueMicrotask(async () => {
    if (!isRunning) {
      await executeAll();
    }
  });
}

function countNodes(nodes: TestNode[]): void {
  if (!nodes) return;
  for (let i = 0; i < nodes.length; i++) {
    const n = nodes[i];
    if (n.isSuite) {
      statsSuites += 1;
      if (n.children && n.children.length > 0) {
        countNodes(n.children);
      }
    } else {
      statsTests += 1;
      if (n.children && n.children.length > 0) {
        countNodes(n.children);
      }
    }
  }
}

async function executeAll(options?: RunOptions, stream?: TestsStream): Promise<void> {
  if (isRunning) return;
  isRunning = true;
  if (stream) {
    activeStream = stream;
  }
  const start = Date.now();

  if (options) {
    if (options.signal) {
      const sig = options.signal as { aborted?: boolean; addEventListener?: (event: string, listener: () => void) => void };
      if (sig.aborted) {
        abortRequested = true;
      }
      if (sig.addEventListener) {
        sig.addEventListener("abort", () => {
          abortRequested = true;
        });
      }
    }
    if (options.timeout !== undefined && options.timeout > 0) {
      setTimeout(() => {
        abortRequested = true;
      }, options.timeout);
    }
  }

  let onlyMode = false;
  if (options) {
    if (options.only) {
      onlyMode = true;
    }
  }

  countNodes(rootNodes);

  for (let i = 0; i < rootNodes.length; i++) {
    if (abortRequested) {
      statsCancelled += 1;
      break;
    }
    const node = rootNodes[i];
    if (onlyMode && !node.only) {
      continue;
    }
    await node.execute();
  }

  statsDurationMs = Date.now() - start;

  console.log(`ℹ tests ${statsTests}`);
  console.log(`ℹ suites ${statsSuites}`);
  console.log(`ℹ pass ${statsPass}`);
  console.log(`ℹ fail ${statsFail}`);
  console.log(`ℹ cancelled ${statsCancelled}`);
  console.log(`ℹ skipped ${statsSkipped}`);
  console.log(`ℹ todo ${statsTodo}`);

  if (activeStream) {
    activeStream.emit("test:summary", {
      tests: statsTests,
      suites: statsSuites,
      pass: statsPass,
      fail: statsFail,
      cancelled: statsCancelled,
      skipped: statsSkipped,
      todo: statsTodo,
      durationMs: statsDurationMs,
    });
  }

  if (statsFail > 0) {
    console.log(`\n✖ failing tests:`);
    for (let i = 0; i < failedNodes.length; i++) {
      const f = failedNodes[i];
      console.log(`✖ ${f.name}`);
      if (f.error) {
        console.log(`  ${f.error}`);
      }
    }
    __scriptgo.exit(1);
  }
}

// Top level functions
function createTest(
  nameOrOptionsOrFn?: string | TestOptions | TestFn,
  optionsOrFn?: TestOptions | TestFn,
  fn?: TestFn,
  isOnly: boolean = false,
  isSkip: boolean = false,
  isTodo: boolean = false,
): Promise<void> {
  let name = "test";
  let opts: TestOptions = {};
  let testFn: TestFn | undefined;

  if (typeof nameOrOptionsOrFn === "string") {
    name = nameOrOptionsOrFn;
    if (typeof optionsOrFn === "function") {
      testFn = optionsOrFn as TestFn;
    } else if (optionsOrFn && typeof optionsOrFn === "object") {
      opts = optionsOrFn as TestOptions;
      testFn = fn;
    }
  } else if (typeof nameOrOptionsOrFn === "function") {
    testFn = nameOrOptionsOrFn as TestFn;
  } else if (nameOrOptionsOrFn && typeof nameOrOptionsOrFn === "object") {
    opts = nameOrOptionsOrFn as TestOptions;
    if (typeof optionsOrFn === "function") {
      testFn = optionsOrFn as TestFn;
    }
  }

  if (isOnly) opts.only = true;
  if (isSkip) opts.skip = true;
  if (isTodo) opts.todo = true;

  const node = new TestNode(name, false, testFn, opts);
  if (currentSuite) {
    node.parent = currentSuite;
    currentSuite.children.push(node);
  } else {
    rootNodes.push(node);
    scheduleAutoRun();
  }

  return Promise.resolve();
}

function createSuite(
  nameOrOptionsOrFn?: string | SuiteOptions | SuiteFn,
  optionsOrFn?: SuiteOptions | SuiteFn,
  fn?: SuiteFn,
  isOnly: boolean = false,
  isSkip: boolean = false,
  isTodo: boolean = false,
): Promise<void> {
  let name = "suite";
  let opts: SuiteOptions = {};
  let suiteFn: SuiteFn | undefined;

  if (typeof nameOrOptionsOrFn === "string") {
    name = nameOrOptionsOrFn;
    if (typeof optionsOrFn === "function") {
      suiteFn = optionsOrFn as SuiteFn;
    } else if (optionsOrFn && typeof optionsOrFn === "object") {
      opts = optionsOrFn as SuiteOptions;
      suiteFn = fn;
    }
  } else if (typeof nameOrOptionsOrFn === "function") {
    suiteFn = nameOrOptionsOrFn as SuiteFn;
  } else if (nameOrOptionsOrFn && typeof nameOrOptionsOrFn === "object") {
    opts = nameOrOptionsOrFn as SuiteOptions;
    if (typeof optionsOrFn === "function") {
      suiteFn = optionsOrFn as SuiteFn;
    }
  }

  if (isOnly) opts.only = true;
  if (isSkip) opts.skip = true;
  if (isTodo) opts.todo = true;

  const node = new TestNode(name, true, suiteFn, opts);
  if (currentSuite) {
    node.parent = currentSuite;
    currentSuite.children.push(node);
  } else {
    rootNodes.push(node);
    scheduleAutoRun();
  }

  return Promise.resolve();
}

// test function & modifiers
export function test(
  nameOrOptionsOrFn?: string | TestOptions | TestFn,
  optionsOrFn?: TestOptions | TestFn,
  fn?: TestFn,
): Promise<void> {
  return createTest(nameOrOptionsOrFn, optionsOrFn, fn, false, false, false);
}

export namespace test {
  export function only(
    nameOrOptionsOrFn?: string | TestOptions | TestFn,
    optionsOrFn?: TestOptions | TestFn,
    fn?: TestFn,
  ): Promise<void> {
    return createTest(nameOrOptionsOrFn, optionsOrFn, fn, true, false, false);
  }

  export function skip(
    nameOrOptionsOrFn?: string | TestOptions | TestFn,
    optionsOrFn?: TestOptions | TestFn,
    fn?: TestFn,
  ): Promise<void> {
    return createTest(nameOrOptionsOrFn, optionsOrFn, fn, false, true, false);
  }

  export function todo(
    nameOrOptionsOrFn?: string | TestOptions | TestFn,
    optionsOrFn?: TestOptions | TestFn,
    fn?: TestFn,
  ): Promise<void> {
    return createTest(nameOrOptionsOrFn, optionsOrFn, fn, false, false, true);
  }
}

// it alias
export function it(
  nameOrOptionsOrFn?: string | TestOptions | TestFn,
  optionsOrFn?: TestOptions | TestFn,
  fn?: TestFn,
): Promise<void> {
  return createTest(nameOrOptionsOrFn, optionsOrFn, fn, false, false, false);
}

export namespace it {
  export function only(
    nameOrOptionsOrFn?: string | TestOptions | TestFn,
    optionsOrFn?: TestOptions | TestFn,
    fn?: TestFn,
  ): Promise<void> {
    return createTest(nameOrOptionsOrFn, optionsOrFn, fn, true, false, false);
  }

  export function skip(
    nameOrOptionsOrFn?: string | TestOptions | TestFn,
    optionsOrFn?: TestOptions | TestFn,
    fn?: TestFn,
  ): Promise<void> {
    return createTest(nameOrOptionsOrFn, optionsOrFn, fn, false, true, false);
  }

  export function todo(
    nameOrOptionsOrFn?: string | TestOptions | TestFn,
    optionsOrFn?: TestOptions | TestFn,
    fn?: TestFn,
  ): Promise<void> {
    return createTest(nameOrOptionsOrFn, optionsOrFn, fn, false, false, true);
  }
}

// describe function & modifiers
export function describe(
  nameOrOptionsOrFn?: string | SuiteOptions | SuiteFn,
  optionsOrFn?: SuiteOptions | SuiteFn,
  fn?: SuiteFn,
): Promise<void> {
  return createSuite(nameOrOptionsOrFn, optionsOrFn, fn, false, false, false);
}

export namespace describe {
  export function only(
    nameOrOptionsOrFn?: string | SuiteOptions | SuiteFn,
    optionsOrFn?: SuiteOptions | SuiteFn,
    fn?: SuiteFn,
  ): Promise<void> {
    return createSuite(nameOrOptionsOrFn, optionsOrFn, fn, true, false, false);
  }

  export function skip(
    nameOrOptionsOrFn?: string | SuiteOptions | SuiteFn,
    optionsOrFn?: SuiteOptions | SuiteFn,
    fn?: SuiteFn,
  ): Promise<void> {
    return createSuite(nameOrOptionsOrFn, optionsOrFn, fn, false, true, false);
  }

  export function todo(
    nameOrOptionsOrFn?: string | SuiteOptions | SuiteFn,
    optionsOrFn?: SuiteOptions | SuiteFn,
    fn?: SuiteFn,
  ): Promise<void> {
    return createSuite(nameOrOptionsOrFn, optionsOrFn, fn, false, false, true);
  }
}

// suite alias
export function suite(
  nameOrOptionsOrFn?: string | SuiteOptions | SuiteFn,
  optionsOrFn?: SuiteOptions | SuiteFn,
  fn?: SuiteFn,
): Promise<void> {
  return createSuite(nameOrOptionsOrFn, optionsOrFn, fn, false, false, false);
}

export namespace suite {
  export function only(
    nameOrOptionsOrFn?: string | SuiteOptions | SuiteFn,
    optionsOrFn?: SuiteOptions | SuiteFn,
    fn?: SuiteFn,
  ): Promise<void> {
    return createSuite(nameOrOptionsOrFn, optionsOrFn, fn, true, false, false);
  }

  export function skip(
    nameOrOptionsOrFn?: string | SuiteOptions | SuiteFn,
    optionsOrFn?: SuiteOptions | SuiteFn,
    fn?: SuiteFn,
  ): Promise<void> {
    return createSuite(nameOrOptionsOrFn, optionsOrFn, fn, false, true, false);
  }

  export function todo(
    nameOrOptionsOrFn?: string | SuiteOptions | SuiteFn,
    optionsOrFn?: SuiteOptions | SuiteFn,
    fn?: SuiteFn,
  ): Promise<void> {
    return createSuite(nameOrOptionsOrFn, optionsOrFn, fn, false, false, true);
  }
}

// Hooks
export function before(fn: HookFn): void {
  if (currentSuite) {
    currentSuite.beforeHooks.push(fn);
  }
}

export function after(fn: HookFn): void {
  if (currentSuite) {
    currentSuite.afterHooks.push(fn);
  }
}

export function beforeEach(fn: HookFn): void {
  if (currentSuite) {
    currentSuite.beforeEachHooks.push(fn);
  }
}

export function afterEach(fn: HookFn): void {
  if (currentSuite) {
    currentSuite.afterEachHooks.push(fn);
  }
}

export class TestsStream extends Readable {
  constructor() {
    super();
  }
}

export function run(options?: RunOptions): TestsStream {
  const stream = new TestsStream();
  if (options && options.setup) {
    (options.setup as Function)(stream);
  }
  executeAll(options, stream).then(() => {
    stream.push(null);
  });
  return stream;
}

export default test;
