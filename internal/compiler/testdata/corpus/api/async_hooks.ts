import {
    createHook,
    AsyncHook,
    executionAsyncId,
    triggerAsyncId,
    executionAsyncResource,
    AsyncLocalStorage,
    AsyncResource
} from "node:async_hooks";

// @api: async_hooks.createHook
// @api: async_hooks.async_hooks.AsyncHook
// @api: new async_hooks.AsyncHook
// @expect: ah_hook_inst: true
const hook = createHook({});
console.log("ah_hook_inst: " + (hook instanceof AsyncHook));

// @api: AsyncHook.enable
// @expect: ah_hook_enable: true
hook.enable();
console.log("ah_hook_enable: true");

// @api: AsyncHook.disable
// @expect: ah_hook_disable: true
hook.disable();
console.log("ah_hook_disable: true");

// @api: AsyncHook.executionAsyncResource
// @expect: ah_hook_execRes: true
console.log("ah_hook_execRes: " + (hook.executionAsyncResource() === null));

// @api: AsyncHook.executionAsyncId
// @expect: ah_hook_execId: 1
console.log("ah_hook_execId: " + hook.executionAsyncId());

// @api: AsyncHook.triggerAsyncId
// @expect: ah_hook_trigId: 0
console.log("ah_hook_trigId: " + hook.triggerAsyncId());

// @api: AsyncHook.return
// @expect: ah_hook_ret: true
hook.return();
console.log("ah_hook_ret: true");

// @api: async_hooks.async_hooks.AsyncLocalStorage
// @api: async_hooks.AsyncLocalStorage
// @api: async_context.async_context.AsyncLocalStorage
// @api: async_context.AsyncLocalStorage
// @api: new async_context.AsyncLocalStorage
// @expect: ah_als_inst: true
const als = new AsyncLocalStorage<number>();
console.log("ah_als_inst: " + (als instanceof AsyncLocalStorage));

// @api: AsyncLocalStorage.enterWith
// @api: AsyncLocalStorage.getStore
// @expect: ah_als_store: 42
als.enterWith(42);
console.log("ah_als_store: " + als.getStore());

// @api: AsyncLocalStorage.run
// @expect: ah_als_run: 100
const runRes = als.run(100, () => als.getStore());
console.log("ah_als_run: " + runRes);

// @api: AsyncLocalStorage.exit
// @expect: ah_als_exit: undefined
const exitRes = als.exit(() => als.getStore());
console.log("ah_als_exit: " + exitRes);

// @api: AsyncLocalStorage.disable
// @expect: ah_als_disable: undefined
als.disable();
console.log("ah_als_disable: " + als.getStore());

// @api: async_hooks.async_hooks.AsyncResource
// @api: async_hooks.AsyncResource
// @api: async_context.async_context.AsyncResource
// @api: async_context.AsyncResource
// @api: new async_context.AsyncResource
// @expect: ah_ar_inst: true
const ar = new AsyncResource("testResource");
console.log("ah_ar_inst: " + (ar instanceof AsyncResource));

// @api: AsyncResource.asyncId
// @expect: ah_ar_asyncId: 1
console.log("ah_ar_asyncId: " + ar.asyncId());

// @api: AsyncResource.triggerAsyncId
// @expect: ah_ar_trigId: 0
console.log("ah_ar_trigId: " + ar.triggerAsyncId());

// @api: AsyncResource.bind
// @expect: ah_ar_bind: true
const boundFn = ar.bind((x: number) => x * 2);
console.log("ah_ar_bind: " + (typeof boundFn === "function"));

// @api: AsyncResource.runInAsyncScope
// @expect: ah_ar_runScope: 84
const scopeRes = ar.runInAsyncScope((x: number) => x * 2, null, 42);
console.log("ah_ar_runScope: " + scopeRes);

// @api: AsyncResource.emitDestroy
// @expect: ah_ar_emitDestroy: true
ar.emitDestroy();
console.log("ah_ar_emitDestroy: true");
