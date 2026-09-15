// ScriptGo Corpus: Dynamic async Promise with timer host job
// @dynamic
// @expect: immediate: done
// @expect: delayed: 42
import { delayValue, immediateValue } from "./dynamic.js";

delayValue(20, 42).then((v: number) => {
  console.log("delayed: " + v);
});

immediateValue("done").then((s: string) => {
  console.log("immediate: " + s);
});
