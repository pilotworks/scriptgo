// @dynamic
const target: any = {
  foo: "bar",
  count: 10,
  extra: "baz",
};

const handler: any = {
  get(t: any, prop: any, receiver: any) {
    if (prop === "magic") return 999;
    return t[prop];
  },
  set(t: any, prop: any, val: any) {
    t[prop] = val;
    return true;
  },
  has(t: any, prop: any) {
    if (prop === "secret") return true;
    return prop in t;
  },
  deleteProperty(t: any, prop: any) {
    delete t[prop];
    return true;
  },
  ownKeys(t: any) {
    return ["foo", "count", "extra"];
  },
};

const proxy: any = new Proxy(target, handler);

console.log(proxy.foo);
console.log(proxy.magic);

proxy.count = 42;
console.log(proxy.count);

console.log("secret" in proxy);
console.log("missing" in proxy);

const keys: any = Object.keys(proxy);
console.log(keys.length);
console.log(keys[0]);
console.log(keys[1]);
console.log(keys[2]);

delete proxy.foo;
console.log(proxy.foo);

const fnTarget: any = (a: number, b: number) => a + b;
const fnHandler: any = {
  apply(t: any, thisArg: any, args: any) {
    return Number(t(args[0], args[1])) * 2;
  },
};
const fnProxy: any = new Proxy(fnTarget, fnHandler);
console.log(fnProxy(10, 11));
