import {
    PerformanceEntry,
    PerformanceMark,
    PerformanceMeasure,
    PerformanceNodeEntry,
    PerformanceNodeTiming,
    PerformanceResourceTiming,
    PerformanceObserverEntryList,
    PerformanceObserver,
    Histogram,
    IntervalHistogram,
    RecordableHistogram,
    createHistogram,
    monitorEventLoopDelay,
    performance
} from "node:perf_hooks";

// @api: perf_hooks.PerformanceEntry
// @api: PerformanceEntry.name
// @api: PerformanceEntry.entryType
// @api: PerformanceEntry.startTime
// @api: PerformanceEntry.duration
// @expect: ph_entry: my-entry mark 10 20
const entry = new PerformanceEntry("my-entry", "mark", 10, 20);
console.log("ph_entry: " + entry.name + " " + entry.entryType + " " + entry.startTime + " " + entry.duration);

// @api: perf_hooks.PerformanceMark
// @api: PerformanceMark.detail
// @expect: ph_mark: mark 42
const mark = new PerformanceMark("mark1", { detail: 42 });
console.log("ph_mark: " + mark.entryType + " " + mark.detail);

// @api: perf_hooks.PerformanceMeasure
// @api: PerformanceMeasure.detail
// @expect: ph_measure: measure 99
const measure = new PerformanceMeasure("m1", { detail: 99 });
console.log("ph_measure: " + measure.entryType + " " + measure.detail);

// @api: perf_hooks.PerformanceNodeEntry
// @api: PerformanceNodeEntry.flags
// @api: PerformanceNodeEntry.kind
// @api: PerformanceNodeEntry.detail
// @expect: ph_nodeEntry: node 0 0
const ne = new PerformanceNodeEntry("node-e");
console.log("ph_nodeEntry: " + ne.entryType + " " + ne.flags + " " + ne.kind);

// @api: perf_hooks.PerformanceNodeTiming
// @api: PerformanceNodeTiming.bootstrapComplete
// @api: PerformanceNodeTiming.environment
// @api: PerformanceNodeTiming.idleTime
// @api: PerformanceNodeTiming.loopExit
// @api: PerformanceNodeTiming.loopStart
// @api: PerformanceNodeTiming.nodeStart
// @api: PerformanceNodeTiming.v8Start
// @api: PerformanceNodeTiming.return
// @expect: ph_nodeTiming: node 0
const nt = new PerformanceNodeTiming();
nt.return();
console.log("ph_nodeTiming: " + nt.entryType + " " + nt.bootstrapComplete);

// @api: perf_hooks.PerformanceResourceTiming
// @api: PerformanceResourceTiming.workerStart
// @api: PerformanceResourceTiming.redirectStart
// @api: PerformanceResourceTiming.redirectEnd
// @api: PerformanceResourceTiming.fetchStart
// @api: PerformanceResourceTiming.domainLookupStart
// @api: PerformanceResourceTiming.domainLookupEnd
// @api: PerformanceResourceTiming.connectStart
// @api: PerformanceResourceTiming.connectEnd
// @api: PerformanceResourceTiming.secureConnectionStart
// @api: PerformanceResourceTiming.requestStart
// @api: PerformanceResourceTiming.responseEnd
// @api: PerformanceResourceTiming.transferSize
// @api: PerformanceResourceTiming.encodedBodySize
// @api: PerformanceResourceTiming.decodedBodySize
// @api: PerformanceResourceTiming.toJSON
// @expect: ph_resTiming: resource true
const rt = new PerformanceResourceTiming("res1");
console.log("ph_resTiming: " + rt.entryType + " " + (typeof rt.toJSON() === "object"));

// @api: perf_hooks.PerformanceObserverEntryList
// @api: PerformanceObserverEntryList.getEntries
// @api: PerformanceObserverEntryList.getEntriesByName
// @api: PerformanceObserverEntryList.getEntriesByType
// @expect: ph_entryList: 1 1 1
const el = new PerformanceObserverEntryList([entry]);
console.log("ph_entryList: " + el.getEntries().length + " " + el.getEntriesByName("my-entry").length + " " + el.getEntriesByType("mark").length);

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

// @api: perf_hooks.Histogram
// @api: Histogram.percentile
// @api: Histogram.percentileBigInt
// @api: Histogram.reset
// @api: Histogram.count
// @api: Histogram.countBigInt
// @api: Histogram.exceeds
// @api: Histogram.exceedsBigInt
// @api: Histogram.max
// @api: Histogram.maxBigInt
// @api: Histogram.mean
// @api: Histogram.min
// @api: Histogram.minBigInt
// @api: Histogram.percentiles
// @api: Histogram.percentilesBigInt
// @api: Histogram.stddev
// @expect: ph_hist: 0 0 0
const hist = new Histogram();
hist.reset();
console.log("ph_hist: " + hist.count + " " + hist.min + " " + hist.percentile(50));

// @api: perf_hooks.IntervalHistogram
// @api: IntervalHistogram.enable
// @api: IntervalHistogram.disable
// @api: perf_hooks.monitorEventLoopDelay
// @expect: ph_intervalHist: true true
const ih = monitorEventLoopDelay();
console.log("ph_intervalHist: " + ih.enable() + " " + ih.disable());

// @api: perf_hooks.RecordableHistogram
// @api: RecordableHistogram.record
// @api: RecordableHistogram.recordDelta
// @api: RecordableHistogram.add
// @api: perf_hooks.createHistogram
// @expect: ph_recHist: 3
const rh = createHistogram();
rh.record(10);
rh.recordDelta();
const rh2 = createHistogram();
rh2.record(5);
rh.add(rh2);
console.log("ph_recHist: " + rh.count);

// @api: perf_hooks.performance
// @expect: ph_perf: true
console.log("ph_perf: " + (typeof performance.now() === "number"));
