// ScriptGo Corpus: Util Standard Builtin APIs
// Consolidated test suite with 1:1 isolated assertions for official Node.js util APIs.

import util, {
    format,
    formatWithOptions,
    inspect,
    isDeepStrictEqual,
    types,
    isArray,
    isBoolean,
    isBuffer,
    isDate,
    isError,
    isFunction,
    isNull,
    isNullOrUndefined,
    isNumber,
    isObject,
    isPrimitive,
    isRegExp,
    isString,
    isSymbol,
    isUndefined,
    deprecate,
    _extend,
    toUSVString,
    stripVTControlCharacters,
    styleText,
    getSystemErrorName,
    getSystemErrorMap,
    getSystemErrorMessage,
    parseEnv,
    promisify,
    callbackify,
    MIMEType,
    MIMEParams
} from "node:util";

// @api: util.format
// @expect: format_res: hello 42
console.log("format_res: " + format("%s %d", "hello", 42));

// @api: util.formatWithOptions
// @expect: formatWithOptions_res: hello 42
console.log("formatWithOptions_res: " + formatWithOptions({}, "%s %d", "hello", 42));

// @api: util.inspect
// @expect: inspect_res: "hello"
console.log("inspect_res: " + inspect("hello"));

// @api: util.isDeepStrictEqual
// @expect: isDeepStrictEqual_res: true
console.log("isDeepStrictEqual_res: " + isDeepStrictEqual(1, 1));

// @api: util.types
// @expect: types_res: true
console.log("types_res: " + types.isDate(new Date(0)));

// @api: util.isArray
// @expect: isArray_res: true
console.log("isArray_res: " + isArray([1, 2]));

// @api: util.isBoolean
// @expect: isBoolean_res: true
console.log("isBoolean_res: " + isBoolean(true));

import { Buffer } from "node:buffer";

// @api: util.isBuffer
// @expect: isBuffer_res: true
console.log("isBuffer_res: " + isBuffer(Buffer.alloc(2)));

// @api: util.isDate
// @expect: isDate_res: true
console.log("isDate_res: " + isDate(new Date(0)));

// @api: util.isError
// @expect: isError_res: true
console.log("isError_res: " + isError(new Error("msg")));

// @api: util.isFunction
// @expect: isFunction_res: true
console.log("isFunction_res: " + isFunction(() => {}));

// @api: util.isNull
// @expect: isNull_res: true
console.log("isNull_res: " + isNull(null));

// @api: util.isNullOrUndefined
// @expect: isNullOrUndefined_res: true
console.log("isNullOrUndefined_res: " + isNullOrUndefined(undefined));

// @api: util.isNumber
// @expect: isNumber_res: true
console.log("isNumber_res: " + isNumber(42));

// @api: util.isObject
// @expect: isObject_res: true
console.log("isObject_res: " + isObject({ a: 1 }));

// @api: util.isPrimitive
// @expect: isPrimitive_res: true
console.log("isPrimitive_res: " + isPrimitive(123));

// @api: util.isRegExp
// @expect: isRegExp_res: true
console.log("isRegExp_res: " + isRegExp(/abc/));

// @api: util.isString
// @expect: isString_res: true
console.log("isString_res: " + isString("hello"));

// @api: util.isSymbol
// @expect: isSymbol_res: true
console.log("isSymbol_res: " + isSymbol(Symbol("test")));

// @api: util.isUndefined
// @expect: isUndefined_res: true
console.log("isUndefined_res: " + isUndefined(undefined));

// @api: util.deprecate
// @expect: [DEPRECATION] deprecated_test: test
// @expect: deprecate_res: 7
const depFn = deprecate((x: number): number => x + 2, "test", "deprecated_test");
console.log("deprecate_res: " + depFn(5));

// @api: util.toUSVString
// @expect: toUSVString_res: hello
console.log("toUSVString_res: " + toUSVString("hello"));

// @api: util.stripVTControlCharacters
// @expect: stripVTControlCharacters_res: hello
console.log("stripVTControlCharacters_res: " + stripVTControlCharacters("hello"));

// @api: util.styleText
// @expect: styleText_res: styled
console.log("styleText_res: " + stripVTControlCharacters(styleText("red", "styled")));

// @api: util.getSystemErrorName
// @expect: getSystemErrorName_res: ENOENT
console.log("getSystemErrorName_res: " + getSystemErrorName(-2));

// @api: util.getSystemErrorMap
// @expect: getSystemErrorMap_res: true
console.log("getSystemErrorMap_res: " + (getSystemErrorMap().size > 0));

// @api: util.getSystemErrorMessage
// @expect: getSystemErrorMessage_res: No such file or directory
console.log("getSystemErrorMessage_res: " + getSystemErrorMessage(-2));

// @api: util.parseEnv
// @expect: parseEnv_res: bar
const env = parseEnv("FOO=bar\n# comment");
console.log("parseEnv_res: " + env.get("FOO"));

// @api: util.promisify
// @expect: promisify_res: true
console.log("promisify_res: " + (typeof promisify === "function"));

// @api: util.callbackify
// @expect: callbackify_res: true
console.log("callbackify_res: " + (typeof callbackify === "function"));

// @api: util._extend
// @expect: _extend_res: true
const ext = _extend({ a: 1 }, { b: 2 });
console.log("_extend_res: " + (ext !== null));

// @api: util.MIMEType
// @expect: mimetype_inst: true
const mime = new MIMEType("text/html; charset=utf-8");
console.log("mimetype_inst: " + (mime !== null));

// @api: MIMEType.type
// @expect: mimetype_type: text
console.log("mimetype_type: " + mime.type);

// @api: MIMEType.subtype
// @expect: mimetype_subtype: html
console.log("mimetype_subtype: " + mime.subtype);

// @api: MIMEType.essence
// @expect: mimetype_essence: text/html
console.log("mimetype_essence: " + mime.essence);

// @api: MIMEType.params
// @expect: mimetype_params: utf-8
console.log("mimetype_params: " + mime.params.get("charset"));

// @api: MIMEType.toString
// @expect: mimetype_toString: text/html;charset=utf-8
console.log("mimetype_toString: " + mime.toString());

// @api: MIMEType.toJSON
// @expect: mimetype_toJSON: text/html;charset=utf-8
console.log("mimetype_toJSON: " + mime.toJSON());

// @api: util.MIMEParams
// @expect: mimeparams_inst: true
const mparams = new MIMEParams("a=1; b=2");
console.log("mimeparams_inst: " + (mparams !== null));

// @api: MIMEParams.get
// @expect: mimeparams_get: 1
console.log("mimeparams_get: " + mparams.get("a"));

// @api: MIMEParams.set
// @expect: mimeparams_set: 3
mparams.set("c", "3");
console.log("mimeparams_set: " + mparams.get("c"));

// @api: MIMEParams.has
// @expect: mimeparams_has: true
console.log("mimeparams_has: " + mparams.has("a"));

// @api: MIMEParams.delete
// @expect: mimeparams_delete: false
mparams.delete("a");
console.log("mimeparams_delete: " + mparams.has("a"));

// @api: MIMEParams.keys
// @expect: mimeparams_keys: true
console.log("mimeparams_keys: " + (mparams.keys().length > 0));

// @api: MIMEParams.values
// @expect: mimeparams_values: true
console.log("mimeparams_values: " + (mparams.values().length > 0));

// @api: MIMEParams.entries
// @expect: mimeparams_entries: true
console.log("mimeparams_entries: " + (mparams.entries().length > 0));

