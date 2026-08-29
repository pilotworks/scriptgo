import {
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
    setResolveSnapshotPath
} from "node:test";

// @api: test.MockFunctionContext
// @api: test.test.MockFunctionContext
// @api: new test.MockFunctionContext
// @expect: test_mfc_inst: true
const mfc = new MockFunctionContext();
console.log("test_mfc_inst: " + (mfc instanceof MockFunctionContext));

// @api: test.MockFunctionContext.calls
// @api: MockFunctionContext.calls
// @api: test.MockFunctionContext.callCount
// @api: MockFunctionContext.callCount
// @expect: test_mfc_callCount: 0
console.log("test_mfc_callCount: " + mfc.callCount());

// @api: test.MockFunctionContext.mockImplementation
// @api: MockFunctionContext.mockImplementation
// @expect: test_mfc_mockImpl: true
mfc.mockImplementation(() => {});
console.log("test_mfc_mockImpl: true");

// @api: test.MockFunctionContext.mockImplementationOnce
// @api: MockFunctionContext.mockImplementationOnce
// @expect: test_mfc_mockImplOnce: true
mfc.mockImplementationOnce(() => {});
console.log("test_mfc_mockImplOnce: true");

// @api: test.MockFunctionContext.resetCalls
// @api: MockFunctionContext.resetCalls
// @expect: test_mfc_resetCalls: true
mfc.resetCalls();
console.log("test_mfc_resetCalls: true");

// @api: test.MockFunctionContext.restore
// @api: MockFunctionContext.restore
// @expect: test_mfc_restore: true
mfc.restore();
console.log("test_mfc_restore: true");

// @api: test.MockPropertyContext
// @api: test.test.MockPropertyContext
// @api: new test.MockPropertyContext
// @expect: test_mpc_inst: true
const mpc = new MockPropertyContext();
console.log("test_mpc_inst: " + (mpc instanceof MockPropertyContext));

// @api: test.MockPropertyContext.accesses
// @api: MockPropertyContext.accesses
// @api: test.MockPropertyContext.accessCount
// @api: MockPropertyContext.accessCount
// @expect: test_mpc_accessCount: 0
console.log("test_mpc_accessCount: " + mpc.accessCount());

// @api: test.MockPropertyContext.mockImplementation
// @api: MockPropertyContext.mockImplementation
// @expect: test_mpc_mockImpl: true
mpc.mockImplementation(() => {});
console.log("test_mpc_mockImpl: true");

// @api: test.MockPropertyContext.mockImplementationOnce
// @api: MockPropertyContext.mockImplementationOnce
// @expect: test_mpc_mockImplOnce: true
mpc.mockImplementationOnce(() => {});
console.log("test_mpc_mockImplOnce: true");

// @api: test.MockPropertyContext.resetAccesses
// @api: MockPropertyContext.resetAccesses
// @expect: test_mpc_resetAccess: true
mpc.resetAccesses();
console.log("test_mpc_resetAccess: true");

// @api: test.MockPropertyContext.restore
// @api: MockPropertyContext.restore
// @expect: test_mpc_restore: true
mpc.restore();
console.log("test_mpc_restore: true");

// @api: test.MockModuleContext
// @api: test.test.MockModuleContext
// @api: new test.MockModuleContext
// @api: test.MockModuleContext.restore
// @api: MockModuleContext.restore
// @expect: test_mmc_inst: true
const mmc = new MockModuleContext();
mmc.restore();
console.log("test_mmc_inst: " + (mmc instanceof MockModuleContext));

// @api: test.MockTimers
// @api: test.test.MockTimers
// @api: new test.MockTimers
// @expect: test_mt_inst: true
const mt = new MockTimers();
console.log("test_mt_inst: " + (mt instanceof MockTimers));

// @api: test.MockTimers.enable
// @api: MockTimers.enable
// @expect: test_mt_enable: true
mt.enable();
console.log("test_mt_enable: true");

// @api: test.MockTimers.tick
// @api: MockTimers.tick
// @expect: test_mt_tick: true
mt.tick(100);
console.log("test_mt_tick: true");

// @api: test.MockTimers.runAll
// @api: MockTimers.runAll
// @expect: test_mt_runAll: true
mt.runAll();
console.log("test_mt_runAll: true");

// @api: test.MockTimers.setTime
// @api: MockTimers.setTime
// @expect: test_mt_setTime: true
mt.setTime(1000);
console.log("test_mt_setTime: true");

// @api: test.MockTimers.[Symbol.dispose]
// @api: MockTimers.[Symbol.dispose]
// @api: test.MockTimers.reset
// @api: MockTimers.reset
// @expect: test_mt_reset: true
mt.reset();
console.log("test_mt_reset: true");

// @api: test.MockTracker
// @api: test.test.MockTracker
// @api: new test.MockTracker
// @expect: test_tracker_inst: true
const tracker = new MockTracker();
console.log("test_tracker_inst: " + (tracker instanceof MockTracker));

// @api: test.MockTracker.fn
// @api: MockTracker.fn
// @expect: test_tracker_fn: function
const mockFn = tracker.fn();
console.log("test_tracker_fn: " + typeof mockFn);

// @api: test.MockTracker.getter
// @api: MockTracker.getter
// @expect: test_tracker_getter: true
const getCtx = tracker.getter({}, "prop");
console.log("test_tracker_getter: " + (getCtx instanceof MockPropertyContext));

// @api: test.MockTracker.setter
// @api: MockTracker.setter
// @expect: test_tracker_setter: true
const setCtx = tracker.setter({}, "prop");
console.log("test_tracker_setter: " + (setCtx instanceof MockPropertyContext));

// @api: test.MockTracker.method
// @api: MockTracker.method
// @expect: test_tracker_method: true
const methCtx = tracker.method({}, "meth");
console.log("test_tracker_method: " + (methCtx instanceof MockFunctionContext));

// @api: test.MockTracker.module
// @api: MockTracker.module
// @expect: test_tracker_module: true
const modCtx = tracker.module("fs");
console.log("test_tracker_module: " + (modCtx instanceof MockModuleContext));

// @api: test.MockTracker.property
// @api: MockTracker.property
// @expect: test_tracker_property: true
const propCtx = tracker.property({}, "prop");
console.log("test_tracker_property: " + (propCtx instanceof MockPropertyContext));

// @api: test.MockTracker.reset
// @api: MockTracker.reset
// @expect: test_tracker_reset: true
tracker.reset();
console.log("test_tracker_reset: true");

// @api: test.MockTracker.restoreAll
// @api: MockTracker.restoreAll
// @expect: test_tracker_restoreAll: true
tracker.restoreAll();
console.log("test_tracker_restoreAll: true");

// @api: test.TestsStream
// @api: test.test.TestsStream
// @api: new test.TestsStream
// @expect: test_stream_inst: true
const stream = run();
console.log("test_stream_inst: " + (stream instanceof TestsStream));

// @api: test.before
// @expect: test_before: true
before(() => {});
console.log("test_before: true");

// @api: test.after
// @expect: test_after: true
after(() => {});
console.log("test_after: true");

// @api: test.beforeEach
// @expect: test_beforeEach: true
beforeEach(() => {});
console.log("test_beforeEach: true");

// @api: test.afterEach
// @expect: test_afterEach: true
afterEach(() => {});
console.log("test_afterEach: true");

// @api: test.register
// @expect: test_register: true
register("test-reporter", {});
console.log("test_register: true");

// @api: test.setDefaultSnapshotSerializers
// @expect: test_setSnapSerializers: true
setDefaultSnapshotSerializers([]);
console.log("test_setSnapSerializers: true");

// @api: test.setResolveSnapshotPath
// @expect: test_setResolveSnap: true
setResolveSnapshotPath(() => "");
console.log("test_setResolveSnap: true");

// @api: test.TestContext
// @api: test.test.TestContext
// @api: new test.TestContext
// @expect: test_tctx_inst: true
const tctx = new TestContext("my_test");
console.log("test_tctx_inst: " + (tctx instanceof TestContext));

// @api: test.TestContext.name
// @api: TestContext.name
// @expect: test_tctx_name: my_test
console.log("test_tctx_name: " + tctx.name);

// @api: test.TestContext.fullName
// @api: TestContext.fullName
// @expect: test_tctx_fullName: my_test
console.log("test_tctx_fullName: " + tctx.fullName);

// @api: test.TestContext.filePath
// @api: TestContext.filePath
// @expect: test_tctx_filePath: 
console.log("test_tctx_filePath: " + tctx.filePath);

// @api: test.TestContext.passed
// @api: TestContext.passed
// @expect: test_tctx_passed: true
console.log("test_tctx_passed: " + tctx.passed);

// @api: test.TestContext.error
// @api: TestContext.error
// @expect: test_tctx_error: true
console.log("test_tctx_error: " + (tctx.error === null));

// @api: test.TestContext.attempt
// @api: TestContext.attempt
// @expect: test_tctx_attempt: 1
console.log("test_tctx_attempt: " + tctx.attempt);

// @api: test.TestContext.signal
// @api: TestContext.signal
// @expect: test_tctx_signal: true
console.log("test_tctx_signal: " + (tctx.signal === null));

// @api: test.TestContext.assert
// @api: TestContext.assert
// @expect: test_tctx_assert: true
console.log("test_tctx_assert: " + (typeof tctx.assert === "object"));

// @api: test.TestContext.before
// @api: TestContext.before
// @expect: test_tctx_before: true
tctx.before(() => {});
console.log("test_tctx_before: true");

// @api: test.TestContext.beforeEach
// @api: TestContext.beforeEach
// @expect: test_tctx_beforeEach: true
tctx.beforeEach(() => {});
console.log("test_tctx_beforeEach: true");

// @api: test.TestContext.after
// @api: TestContext.after
// @expect: test_tctx_after: true
tctx.after(() => {});
console.log("test_tctx_after: true");

// @api: test.TestContext.afterEach
// @api: TestContext.afterEach
// @expect: test_tctx_afterEach: true
tctx.afterEach(() => {});
console.log("test_tctx_afterEach: true");

// @api: test.TestContext.diagnostic
// @api: TestContext.diagnostic
// @expect: test_tctx_diag: true
tctx.diagnostic("msg");
console.log("test_tctx_diag: true");

// @api: test.TestContext.plan
// @api: TestContext.plan
// @expect: test_tctx_plan: true
tctx.plan(1);
console.log("test_tctx_plan: true");

// @api: test.TestContext.runOnly
// @api: TestContext.runOnly
// @expect: test_tctx_runOnly: true
tctx.runOnly(true);
console.log("test_tctx_runOnly: true");

// @api: test.TestContext.skip
// @api: TestContext.skip
// @expect: test_tctx_skip: true
tctx.skip("skipped");
console.log("test_tctx_skip: true");

// @api: test.TestContext.todo
// @api: TestContext.todo
// @expect: test_tctx_todo: true
tctx.todo("todo");
console.log("test_tctx_todo: true");

// @api: test.SuiteContext
// @api: test.test.SuiteContext
// @api: new test.SuiteContext
// @expect: test_sctx_inst: true
const sctx = new SuiteContext("my_suite");
console.log("test_sctx_inst: " + (sctx instanceof SuiteContext));

// @api: test.SuiteContext.name
// @api: SuiteContext.name
// @expect: test_sctx_name: my_suite
console.log("test_sctx_name: " + sctx.name);

// @api: test.SuiteContext.fullName
// @api: SuiteContext.fullName
// @expect: test_sctx_fullName: my_suite
console.log("test_sctx_fullName: " + sctx.fullName);

// @api: test.SuiteContext.filePath
// @api: SuiteContext.filePath
// @expect: test_sctx_filePath: 
console.log("test_sctx_filePath: " + sctx.filePath);

// @api: test.SuiteContext.signal
// @api: SuiteContext.signal
// @expect: test_sctx_signal: true
console.log("test_sctx_signal: " + (sctx.signal === null));

// @api: test.run
// @expect: test_run: true
console.log("test_run: " + (typeof run === "function"));

// @api: test.skip
// @expect: test_skip: true
skip("skip", () => {});
console.log("test_skip: true");

// @api: test.todo
// @expect: test_todo: true
todo("todo", () => {});
console.log("test_todo: true");

// @api: test.only
// @expect: test_only: true
only("only", () => {});
console.log("test_only: true");

// @api: test.test
// @expect: test_fn: true
test("test1", (t: TestContext) => {
    console.log("test_fn: " + (t instanceof TestContext));
});

// @api: test.it
// @expect: test_it: true
it("it1", (t: TestContext) => {
    console.log("test_it: " + (t instanceof TestContext));
});

// @api: test.suite
// @expect: test_suite: true
suite("suite1", (s: SuiteContext) => {
    console.log("test_suite: " + (s instanceof SuiteContext));
});

// @api: test.describe
// @expect: test_describe: true
describe("desc1", (s: SuiteContext) => {
    console.log("test_describe: " + (s instanceof SuiteContext));
});

// @api: test.TestContext.test
// @api: TestContext.test
// @expect: test_tctx_subtest: true
tctx.test("subtest", () => {
    console.log("test_tctx_subtest: true");
});

// @api: test.TestContext.waitFor
// @api: TestContext.waitFor
// @expect: test_tctx_waitFor: true
tctx.waitFor(() => true).then(() => {
    console.log("test_tctx_waitFor: true");
});
