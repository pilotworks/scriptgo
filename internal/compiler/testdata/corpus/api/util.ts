import {
    format,
    inspect,
    isDeepStrictEqual,
    types,
    isArray,
    isBoolean,
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
    MIMEType,
    MIMEParams,
    parseEnv,
    diff
} from "node:util";

// @api: util.format
// @expect: hello 42
console.log(format("%s %d", "hello", 42));

// @api: util.inspect
// @expect: 123
// @expect: "hello"
console.log(inspect(123));
console.log(inspect("hello"));

// @api: util.isDeepStrictEqual
// @expect: true
// @expect: false
console.log(isDeepStrictEqual(1, 1));
console.log(isDeepStrictEqual(1, 2));

// @api: util.types
// @expect: true
// @expect: true
// @expect: true
// @expect: true
console.log(types.isDate(new Date(0)));
console.log(types.isRegExp(/abc/));
console.log(types.isNativeError(new Error("msg")));
console.log(types.isPromise(Promise.resolve(1)));

// @api: util type predicates
// @expect: true
// @expect: true
// @expect: true
// @expect: true
// @expect: true
// @expect: true
// @expect: true
// @expect: true
// @expect: true
// @expect: true
console.log(isArray([1, 2]));
console.log(isBoolean(true));
console.log(isDate(new Date(0)));
console.log(isError(new Error("err")));
console.log(isFunction(() => {}));
console.log(isNull(null));
console.log(isNullOrUndefined(undefined));
console.log(isNumber(123));
console.log(isString("abc"));
console.log(isUndefined(undefined));

// @api: util.deprecate
// @expect: [DEPRECATION] deprecated fn
// @expect: 6
const fn = deprecate((x: number): number => x + 1, "deprecated fn");
console.log(fn(5));

// @api: util.toUSVString
// @expect: hello
console.log(toUSVString("hello"));

// @api: util.stripVTControlCharacters
// @expect: hello
console.log(stripVTControlCharacters("hello"));

// @api: util.styleText
// @expect: styled
console.log(stripVTControlCharacters(styleText("red", "styled")));

// @api: util.getSystemErrorName
// @expect: ENOENT
console.log(getSystemErrorName(-2));

// @api: util.getSystemErrorMap
// @expect: true
const map = getSystemErrorMap();
console.log(map.size > 0);

// @api: util.MIMEType & MIMEParams
// @expect: text
// @expect: html
// @expect: text/html
// @expect: utf-8
const mime = new MIMEType("text/html; charset=utf-8");
console.log(mime.type);
console.log(mime.subtype);
console.log(mime.essence);
console.log(mime.params.get("charset"));

// @api: util.parseEnv
// @expect: bar
// @expect: qux
const env = parseEnv("FOO=bar\nBAZ=qux\n# comment");
console.log(env.get("FOO"));
console.log(env.get("BAZ"));
