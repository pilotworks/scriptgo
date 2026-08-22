// ScriptGo Corpus: Path Standard Builtin APIs
// Consolidated test suite with inline assertions.

import * as path from "node:path";

// @api: path.basename
// @expect: baz.txt
console.log(path.basename("/foo/bar/baz.txt"));

// @api: path.dirname
// @expect: /foo/bar
console.log(path.dirname("/foo/bar/baz.txt"));

// @api: path.extname
// @expect: .txt
console.log(path.extname("/foo/bar/baz.txt"));

// @api: path.join
// @expect: /foo/bar/baz
console.log(path.join("/foo", "bar", "baz"));
