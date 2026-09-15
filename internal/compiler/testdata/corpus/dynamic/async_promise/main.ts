// ScriptGo Corpus: Dynamic async Promise boundary
// @dynamic
// @expect: 42
// @expect: recovered
import { answer, fail } from "./dynamic.js";

answer().then((value: number) => console.log(value));
fail().catch(() => console.log("recovered"));
