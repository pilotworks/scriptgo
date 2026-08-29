import { isSea, getAsset, getAssetAsBlob, getRawAsset, getAssetKeys } from "node:sea";

// @api: single-executable-applications.isSea
// @expect: sea_isSea: false
console.log("sea_isSea: " + isSea());

// @api: single-executable-applications.getAssetKeys
// @expect: sea_keys: 0
const keys = getAssetKeys();
console.log("sea_keys: " + keys.length);

// @api: single-executable-applications.getAsset
// @expect: sea_asset: 0
const asset = getAsset("nonexistent");
console.log("sea_asset: " + (asset as ArrayBuffer).byteLength);

// @api: single-executable-applications.getAssetAsBlob
// @expect: sea_blob: true
const blob = getAssetAsBlob("nonexistent");
console.log("sea_blob: " + (blob === null));

// @api: single-executable-applications.getRawAsset
// @expect: sea_raw: 0
const raw = getRawAsset("nonexistent");
console.log("sea_raw: " + raw.byteLength);
