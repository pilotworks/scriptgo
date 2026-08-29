import {
    isMainThread,
    isInternalThread,
    parentPort,
    threadId,
    threadName,
    workerData,
    SHARE_ENV,
    resourceLimits,
    getEnvironmentData,
    setEnvironmentData,
    markAsUntransferable,
    isMarkedAsUntransferable,
    markAsUncloneable,
    moveMessagePortToContext,
    postMessageToThread,
    receiveMessageOnPort,
    MessagePort,
    MessageChannel,
    BroadcastChannel,
    Worker
} from "node:worker_threads";

// @api: worker_threads.isMainThread
// @expect: wt_main: true
console.log("wt_main: " + isMainThread);

// @api: worker_threads.isInternalThread
// @expect: wt_internal: false
console.log("wt_internal: " + isInternalThread);

// @api: worker_threads.parentPort
// @expect: wt_parentPort: null
console.log("wt_parentPort: " + parentPort);

// @api: worker_threads.threadId
// @expect: wt_threadId: 0
console.log("wt_threadId: " + threadId);

// @api: worker_threads.threadName
// @expect: wt_threadName: 
console.log("wt_threadName: " + threadName);

// @api: worker_threads.workerData
// @expect: wt_workerData: undefined
console.log("wt_workerData: " + workerData);

// @api: worker_threads.SHARE_ENV
// @expect: wt_shareEnv: true
console.log("wt_shareEnv: " + (typeof SHARE_ENV === "symbol"));

// @api: worker_threads.resourceLimits
// @expect: wt_resLimits: true
console.log("wt_resLimits: " + (typeof resourceLimits === "object"));

// @api: worker_threads.setEnvironmentData
// @api: worker_threads.getEnvironmentData
// @expect: wt_envData: 42
setEnvironmentData("key", 42);
console.log("wt_envData: " + getEnvironmentData("key"));

// @api: worker_threads.markAsUntransferable
// @expect: wt_markUntrans: true
markAsUntransferable({});
console.log("wt_markUntrans: true");

// @api: worker_threads.isMarkedAsUntransferable
// @expect: wt_isMarked: false
console.log("wt_isMarked: " + isMarkedAsUntransferable({}));

// @api: worker_threads.markAsUncloneable
// @expect: wt_markUnclone: true
markAsUncloneable({});
console.log("wt_markUnclone: true");

// @api: worker_threads.worker_threads.MessageChannel
// @api: worker_threads.MessageChannel
// @api: new worker_threads.MessageChannel
// @expect: wt_mc_inst: true
const mc = new MessageChannel();
console.log("wt_mc_inst: " + (mc instanceof MessageChannel));

// @api: worker_threads.worker_threads.MessagePort
// @api: worker_threads.MessagePort
// @api: new worker_threads.MessagePort
// @expect: wt_mp_inst: true
const port = mc.port1;
console.log("wt_mp_inst: " + (port instanceof MessagePort));

// @api: MessagePort.start
// @expect: wt_mp_start: true
port.start();
console.log("wt_mp_start: true");

// @api: MessagePort.hasRef
// @expect: wt_mp_hasRef: true
console.log("wt_mp_hasRef: " + port.hasRef());

// @api: MessagePort.unref
// @expect: wt_mp_unref: false
port.unref();
console.log("wt_mp_unref: " + port.hasRef());

// @api: MessagePort.ref
// @expect: wt_mp_ref: true
port.ref();
console.log("wt_mp_ref: " + port.hasRef());

// @api: MessagePort.postMessage
// @expect: wt_mp_post: true
port.postMessage({ data: 1 });
console.log("wt_mp_post: true");

// @api: MessagePort.close
// @expect: wt_mp_close: true
port.close();
console.log("wt_mp_close: true");

// @api: worker_threads.moveMessagePortToContext
// @expect: wt_movePort: true
console.log("wt_movePort: " + (moveMessagePortToContext(mc.port2, {}) === mc.port2));

// @api: worker_threads.postMessageToThread
// @expect: wt_postToThread: true
postMessageToThread(1, {});
console.log("wt_postToThread: true");

// @api: worker_threads.receiveMessageOnPort
// @expect: wt_recvPort: undefined
console.log("wt_recvPort: " + receiveMessageOnPort(port));

// @api: worker_threads.worker_threads.BroadcastChannel
// @api: worker_threads.BroadcastChannel
// @api: new worker_threads.BroadcastChannel
// @expect: wt_bc_inst: true
const bc = new BroadcastChannel("my-broadcast");
console.log("wt_bc_inst: " + (bc instanceof BroadcastChannel));

// @api: BroadcastChannel.onmessage
// @api: BroadcastChannel.onmessageerror
// @expect: wt_bc_handlers: true
bc.onmessage = (e: unknown) => {};
bc.onmessageerror = (e: unknown) => {};
console.log("wt_bc_handlers: true");

// @api: BroadcastChannel.postMessage
// @expect: wt_bc_post: true
bc.postMessage("hello");
console.log("wt_bc_post: true");

// @api: BroadcastChannel.ref
// @expect: wt_bc_ref: true
bc.ref();
console.log("wt_bc_ref: true");

// @api: BroadcastChannel.unref
// @expect: wt_bc_unref: true
bc.unref();
console.log("wt_bc_unref: true");

// @api: BroadcastChannel.close
// @expect: wt_bc_close: true
bc.close();
console.log("wt_bc_close: true");

// @api: worker_threads.worker_threads.Worker
// @api: worker_threads.Worker
// @api: new worker_threads.Worker
// @expect: wt_worker_inst: true
const w = new Worker("./worker.js");
console.log("wt_worker_inst: " + (w instanceof Worker));

// @api: Worker.threadId
// @expect: wt_w_threadId: 1
console.log("wt_w_threadId: " + w.threadId);

// @api: Worker.threadName
// @expect: wt_w_threadName: 
console.log("wt_w_threadName: " + w.threadName);

// @api: Worker.stdin
// @expect: wt_w_stdin: null
console.log("wt_w_stdin: " + w.stdin);

// @api: Worker.stdout
// @expect: wt_w_stdout: null
console.log("wt_w_stdout: " + w.stdout);

// @api: Worker.stderr
// @expect: wt_w_stderr: null
console.log("wt_w_stderr: " + w.stderr);

// @api: Worker.performance
// @expect: wt_w_perf: true
console.log("wt_w_perf: " + (typeof w.performance === "object"));

// @api: Worker.resourceLimits
// @expect: wt_w_resLimits: true
console.log("wt_w_resLimits: " + (typeof w.resourceLimits === "object"));

// @api: Worker.postMessage
// @expect: wt_w_post: true
w.postMessage({ hello: "world" });
console.log("wt_w_post: true");

// @api: Worker.ref
// @expect: wt_w_ref: true
w.ref();
console.log("wt_w_ref: true");

// @api: Worker.unref
// @expect: wt_w_unref: true
w.unref();
console.log("wt_w_unref: true");

// @api: Worker.cpuUsage
// @expect: wt_w_cpu: 0
console.log("wt_w_cpu: " + w.cpuUsage().user);

// @api: Worker.startCpuProfile
// @expect: wt_w_startProfile: undefined
console.log("wt_w_startProfile: " + w.startCpuProfile());

// @api: Worker.terminate
// @api: Worker.getHeapSnapshot
// @api: Worker.getHeapStatistics
// @api: Worker.[Symbol.asyncDispose]
// @expect: wt_w_term: 0
// @expect: wt_w_heapSnap: true
// @expect: wt_w_heapStats: true
// @expect: wt_w_dispose: true
const runAsync = async () => {
    const code = await w.terminate();
    console.log("wt_w_term: " + code);

    const stream = await w.getHeapSnapshot();
    console.log("wt_w_heapSnap: " + (typeof stream === "object"));

    const stats = await w.getHeapStatistics();
    console.log("wt_w_heapStats: " + (typeof stats === "object"));

    await w[Symbol.asyncDispose]();
    console.log("wt_w_dispose: true");
};
runAsync();
