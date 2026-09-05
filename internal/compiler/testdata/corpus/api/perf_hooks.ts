import {
    PerformanceMark,
    PerformanceObserver,
    createHistogram,
    performance
} from "node:perf_hooks";

// @api: perf_hooks.PerformanceMark
// @api: PerformanceMark.detail
// @expect: ph_mark: mark 42
const mark = new PerformanceMark("mark1", { detail: 42 });
console.log("ph_mark: " + mark.entryType + " " + mark.detail);

// @api: perf_hooks.PerformanceObserver
// @api: PerformanceObserver.supportedEntryTypes
// @api: PerformanceObserver.observe
// @api: PerformanceObserver.takeRecords
// @api: PerformanceObserver.disconnect
// @expect: ph_obs: true 0
const obs = new PerformanceObserver((list) => {});
obs.observe({ entryTypes: ["mark"] });
const records = obs.takeRecords();
obs.disconnect();
console.log("ph_obs: " + (PerformanceObserver.supportedEntryTypes.length > 0) + " " + records.length);

// @api: perf_hooks.RecordableHistogram
// @api: RecordableHistogram.record
// @api: RecordableHistogram.add
// @api: perf_hooks.createHistogram
// @expect: ph_recHist: 2
const rh = createHistogram();
rh.record(10);
const rh2 = createHistogram();
rh2.record(5);
rh.add(rh2);
console.log("ph_recHist: " + rh.count);

// @api: perf_hooks.performance
// @expect: ph_perf: true
console.log("ph_perf: " + (typeof performance.now() === "number"));
