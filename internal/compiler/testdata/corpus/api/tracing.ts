import { createTracing, getEnabledCategories } from "node:trace_events";

// @api: tracing.getEnabledCategories
// @expect: tracing_getEnabledCategories: true
const enabledCats = getEnabledCategories();
console.log("tracing_getEnabledCategories: " + (enabledCats.indexOf("v8") !== -1));

// @api: tracing.createTracing
// @expect: tracing_create_ok: true
const tr = createTracing({ categories: ["node.perf"] });
console.log("tracing_create_ok: " + (tr !== null));

// @api: tracing.categories
// @expect: tracing_categories: node.perf
console.log("tracing_categories: " + tr.categories);

// @api: tracing.enabled
// @expect: tracing_enabled: false
console.log("tracing_enabled: " + tr.enabled);

// @api: tracing.enable
// @expect: tracing_enable_ok: true
tr.enable();
console.log("tracing_enable_ok: " + tr.enabled);

// @api: tracing.disable
// @expect: tracing_disable_ok: false
tr.disable();
console.log("tracing_disable_ok: " + tr.enabled);
