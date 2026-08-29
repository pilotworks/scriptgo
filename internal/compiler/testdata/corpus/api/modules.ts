import { Module } from "node:module";

// @api: modules.id
// @expect: mod_id: /tmp/test.js
const m = new Module("/tmp/test.js");
console.log("mod_id: " + m.id);

// @api: modules.filename
// @expect: mod_filename: /tmp/test.js
console.log("mod_filename: " + m.filename);

// @api: modules.loaded
// @expect: mod_loaded: true
console.log("mod_loaded: " + m.loaded);

// @api: modules.parent
// @expect: mod_parent: true
console.log("mod_parent: " + (m.parent === null));

// @api: modules.children
// @expect: mod_children: 0
console.log("mod_children: " + m.children.length);

// @api: modules.exports
// @expect: mod_exports: true
console.log("mod_exports: " + (typeof m.exports === "object"));

// @api: modules.paths
// @expect: mod_paths: 0
console.log("mod_paths: " + m.paths.length);

// @api: modules.path
// @expect: mod_path: 
console.log("mod_path: " + m.path);

// @api: modules.isPreloading
// @expect: mod_isPreloading: false
console.log("mod_isPreloading: " + m.isPreloading);

// @api: modules.require
// @expect: mod_require: true
const reqResult = m.require("test");
console.log("mod_require: " + (typeof reqResult === "object"));
