// ScriptGo Corpus: Http Standard Builtin APIs
// Consolidated test suite with inline assertions.

import {
    METHODS,
    STATUS_CODES,
    maxHeaderSize,
    validateHeaderName,
    validateHeaderValue
} from "node:http";

// @api: METHODS
// @expect: true
// @expect: GET
console.log(METHODS.length > 0);
console.log(METHODS[6]);

// @api: STATUS_CODES
// @expect: OK
// @expect: Not Found
console.log(STATUS_CODES["200"]);
console.log(STATUS_CODES["404"]);

// @api: http.validateHeaderName
// @expect: valid
validateHeaderName("Content-Type");
console.log("valid");

// @api: http.validateHeaderValue
// @expect: val-ok
validateHeaderValue("Content-Type", "application/json");
console.log("val-ok");

// @api: maxHeaderSize
// @expect: 16384
console.log(maxHeaderSize);
