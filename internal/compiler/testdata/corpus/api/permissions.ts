import { has } from "node:permissions";

// @api: permissions.has
// @expect: perm_has_fs_read: true
// @expect: perm_has_fs_write: true
console.log("perm_has_fs_read: " + has("fs.read"));
console.log("perm_has_fs_write: " + has("fs.write", "/tmp"));
