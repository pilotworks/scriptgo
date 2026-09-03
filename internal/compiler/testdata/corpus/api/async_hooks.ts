import {
    AsyncLocalStorage
} from "node:async_hooks";

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

