import {
    Channel,
    channel,
    tracingChannel,
    hasSubscribers,
    unsubscribe
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
