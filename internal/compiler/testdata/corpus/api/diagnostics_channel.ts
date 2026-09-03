import {
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
    error
} from "node:diagnostics_channel";

// @api: diagnostics_channel.diagnostics_channel.Channel
// @api: diagnostics_channel.Channel
// @api: new diagnostics_channel.Channel
// @expect: dc_chan_inst: true
const ch = channel("my-chan");
console.log("dc_chan_inst: " + (ch instanceof Channel));

// @api: diagnostics_channel.subscribe
// @api: Channel.subscribe
// @expect: dc_sub: true
const listener = (msg: unknown) => {};
ch.subscribe(listener);
console.log("dc_sub: true");

// @api: diagnostics_channel.hasSubscribers
// @expect: dc_hasSub: true
console.log("dc_hasSub: " + hasSubscribers("my-chan"));

// @api: Channel.publish
// @expect: dc_pub: true
ch.publish({ data: 123 });
console.log("dc_pub: true");

// @api: diagnostics_channel.unsubscribe
// @api: Channel.unsubscribe
// @expect: dc_unsub: true
console.log("dc_unsub: " + ch.unsubscribe(listener));

// @api: diagnostics_channel.diagnostics_channel.TracingChannel
// @api: diagnostics_channel.TracingChannel
// @api: new diagnostics_channel.TracingChannel
// @expect: dc_tc_inst: true
const tc = tracingChannel("my-trace");
console.log("dc_tc_inst: " + (tc instanceof TracingChannel));

// @api: TracingChannel.traceSync
// @expect: dc_tc_traceSync: 100
const syncRes = tc.traceSync(() => 100);
console.log("dc_tc_traceSync: " + syncRes);


// @api: diagnostics_channel.start
// @expect: dc_start: true
start("my-chan", {});
console.log("dc_start: true");

// @api: diagnostics_channel.end
// @expect: dc_end: true
end("my-chan", {});
console.log("dc_end: true");

// @api: diagnostics_channel.asyncStart
// @expect: dc_asyncStart: true
asyncStart("my-chan", {});
console.log("dc_asyncStart: true");

// @api: diagnostics_channel.asyncEnd
// @expect: dc_asyncEnd: true
asyncEnd("my-chan", {});
console.log("dc_asyncEnd: true");

// @api: diagnostics_channel.error
// @expect: dc_error: true
error("my-chan", {});
console.log("dc_error: true");

// @api: TracingChannel.tracePromise
// @expect: dc_tc_tracePromise: 300
tc.tracePromise(async () => 300).then((res: number) => {
    console.log("dc_tc_tracePromise: " + res);
});
