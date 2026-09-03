// ScriptGo Corpus: Node.js Path Module (Strict 1:1 Parity Tests)
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

// @api: path.isAbsolute
// @expect: true
console.log(path.isAbsolute("/foo/bar"));

// @api: path.normalize
// @expect: /foo/baz
console.log(path.normalize("/foo/bar/../baz"));

// @api: path.resolve
// @expect: /foo/bar/baz
console.log(path.resolve("/foo/bar", "./baz"));

// @api: path.relative
// @expect: ../baz
console.log(path.relative("/foo/bar", "/foo/baz"));

// @api: path.parse
// @expect: true
const parsed = path.parse("/foo/bar/baz.txt");
console.log(parsed.base === "baz.txt" && parsed.ext === ".txt" && parsed.name === "baz");

// @api: path.format
// @expect: /foo/bar/baz.txt
console.log(path.format({ dir: "/foo/bar", base: "baz.txt" }));

// @api: path.toNamespacedPath
// @expect: /foo/bar
console.log(path.toNamespacedPath("/foo/bar"));

// @api: path.sep
// @expect: /
console.log(path.sep);

// @api: path.delimiter
// @expect: :
console.log(path.delimiter);

// @api: path.posix
// @expect: true
console.log(path.posix.sep === "/");

// @api: path.win32
// @expect: true
console.log(path.win32.sep === "\\");
