// @expect: 42 handled
// @expect: 7
// Methods named then/catch on non-Promise objects are ordinary methods.
class Task {
  then(): number { return 42; }
  catch(): string { return "handled"; }
}
const task = new Task();
console.log(task.then(), task.catch());
const pending: Promise<number> = Promise.resolve(7);
pending.then((value) => { console.log(value); });
