import { WASI } from "node:wasi";

// @api: new wasi.WASI
// @expect: wasi_inst: true
const wasi = new WASI({ version: "preview1" });
console.log("wasi_inst: " + (wasi instanceof WASI));

// @api: WASI.wasiImport
// @expect: wasi_import_ok: true
console.log("wasi_import_ok: " + (typeof wasi.wasiImport === "object"));

// @api: WASI.getImportObject
// @expect: wasi_getImportObject: true
const importObj = wasi.getImportObject();
console.log("wasi_getImportObject: " + (typeof importObj === "object"));

// @api: WASI.start
// @expect: wasi_start_ok: true
wasi.start(null);
console.log("wasi_start_ok: true");

// @api: WASI.initialize
// @expect: wasi_initialize_ok: true
wasi.initialize(null);
console.log("wasi_initialize_ok: true");
