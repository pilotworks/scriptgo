// ScriptGo Corpus: Http Standard Builtin APIs
// Consolidated test suite with inline assertions.

import { METHODS } from "node:http";
import * as http from "node:http";

// @api: http.METHODS
// @expect: true
// @expect: GET
console.log(METHODS.length > 0);
console.log(METHODS[6]);

// @api: http.getStatusText
// @expect: OK
// @expect: Not Found
console.log(http.getStatusText(200));
console.log(http.getStatusText(404));
