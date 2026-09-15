// ScriptGo Corpus: Dynamic async Promise requiring an unavailable host job
// @dynamic
// @run.err: SG5004: Dynamic Promise requires an unavailable host job
import { neverSettles } from "./dynamic.js";

neverSettles().then((v: number) => console.log(v));
