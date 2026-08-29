import {
    isMaster,
    isPrimary,
    isWorker,
    schedulingPolicy,
    settings,
    worker,
    workers,
    setupMaster,
    setupPrimary,
    fork,
    disconnect,
    Worker
} from "node:cluster";

// @api: cluster.isMaster
// @expect: cl_master: true
console.log("cl_master: " + isMaster);

// @api: cluster.isPrimary
// @expect: cl_primary: true
console.log("cl_primary: " + isPrimary);

// @api: cluster.isWorker
// @expect: cl_worker: false
console.log("cl_worker: " + isWorker);

// @api: cluster.schedulingPolicy
// @expect: cl_sched: 2
console.log("cl_sched: " + schedulingPolicy);

// @api: cluster.settings
// @expect: cl_settings: true
console.log("cl_settings: " + (typeof settings === "object"));

// @api: cluster.worker
// @expect: cl_curWorker: undefined
console.log("cl_curWorker: " + worker);

// @api: cluster.setupMaster
// @expect: cl_setupMaster: true
setupMaster({});
console.log("cl_setupMaster: true");

// @api: cluster.setupPrimary
// @expect: cl_setupPrimary: true
setupPrimary({});
console.log("cl_setupPrimary: true");

// @api: cluster.fork
// @api: cluster.cluster.Worker
// @api: cluster.Worker
// @api: new cluster.Worker
// @expect: cl_w_inst: true
const w = fork();
console.log("cl_w_inst: " + (w instanceof Worker));

// @api: cluster.workers
// @expect: cl_workers: true
console.log("cl_workers: " + (typeof workers === "object"));

// @api: Worker.id
// @expect: cl_w_id: 1
console.log("cl_w_id: " + w.id);

// @api: Worker.process
// @expect: cl_w_proc: true
console.log("cl_w_proc: " + (typeof w.process === "object"));

// @api: Worker.isConnected
// @expect: cl_w_connected: true
console.log("cl_w_connected: " + w.isConnected());

// @api: Worker.isDead
// @expect: cl_w_dead: false
console.log("cl_w_dead: " + w.isDead());

// @api: Worker.send
// @expect: cl_w_send: true
console.log("cl_w_send: " + w.send({ cmd: "ping" }));

// @api: Worker.disconnect
// @expect: cl_w_disconn: true
w.disconnect();
console.log("cl_w_disconn: true");

// @api: Worker.exitedAfterDisconnect
// @expect: cl_w_exited: true
console.log("cl_w_exited: " + w.exitedAfterDisconnect);

// @api: Worker.kill
// @expect: cl_w_kill: true
w.kill("SIGTERM");
console.log("cl_w_kill: " + w.isDead());

// @api: cluster.disconnect
// @expect: cl_disconnect: true
disconnect(() => {
    console.log("cl_disconnect: true");
});
