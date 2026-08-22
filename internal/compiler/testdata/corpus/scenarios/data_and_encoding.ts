// ScriptGo Corpus: Scenario: Data, Binary & Encoding
// Consolidated test suite with inline assertions.

import * as crypto from "crypto";

// --- Context Case: scenarios_buffer_base64 ---
// @expect: SGVsbG8sIHNjcmlwdGdvIQ==
// @expect: Hello, scriptgo!
const encoded_buffer_base64_0: string = btoa("Hello, scriptgo!");
console.log(encoded_buffer_base64_0);
const decoded_buffer_base64_0: string = atob(encoded_buffer_base64_0);
console.log(decoded_buffer_base64_0);

// --- Context Case: scenarios_buffer_base64_long ---
// @expect: QUJDREVGR0hJSktMTU5PUA==
// @expect: ABCDEFGHIJKLMNOP
const encoded_buffer_base64_long_1: string = btoa("ABCDEFGHIJKLMNOP");
console.log(encoded_buffer_base64_long_1);
const decoded_buffer_base64_long_1: string = atob(encoded_buffer_base64_long_1);
console.log(decoded_buffer_base64_long_1);

// --- Context Case: scenarios_buffer_binary_rw ---
// @expect: 250
// @expect: -10
// @expect: 4660
// @expect: 13330
// @expect: 305419896
// @expect: -123456
const buf_buffer_binary_rw_2: Buffer = Buffer.alloc(16);

buf_buffer_binary_rw_2.writeUInt8(250, 0);
console.log(buf_buffer_binary_rw_2.readUInt8(0));

buf_buffer_binary_rw_2.writeInt8(-10, 1);
console.log(buf_buffer_binary_rw_2.readInt8(1));

buf_buffer_binary_rw_2.writeUInt16LE(4660, 2);
console.log(buf_buffer_binary_rw_2.readUInt16LE(2));
console.log(buf_buffer_binary_rw_2.readUInt16BE(2));

buf_buffer_binary_rw_2.writeUInt32LE(305419896, 4);
console.log(buf_buffer_binary_rw_2.readUInt32LE(4));

buf_buffer_binary_rw_2.writeInt32BE(-123456, 8);
console.log(buf_buffer_binary_rw_2.readInt32BE(8));

// --- Context Case: scenarios_buffer_from_encodings ---
// @expect: 5
// @expect: hello
// @expect: hello
// @expect: 68656c6c6f
// @expect: hello
// @expect: aGVsbG8=
// @expect: 5
// @expect: 5
// @expect: 5
const bUtf8_buffer_from_encodings_3: Buffer = Buffer.from("hello", "utf8");
console.log(bUtf8_buffer_from_encodings_3.length);
console.log(bUtf8_buffer_from_encodings_3.toString());

const bHex_buffer_from_encodings_3: Buffer = Buffer.from("68656c6c6f", "hex");
console.log(bHex_buffer_from_encodings_3.toString());
console.log(bHex_buffer_from_encodings_3.toString("hex"));

const bB64_buffer_from_encodings_3: Buffer = Buffer.from("aGVsbG8=", "base64");
console.log(bB64_buffer_from_encodings_3.toString());
console.log(bB64_buffer_from_encodings_3.toString("base64"));

console.log(Buffer.byteLength("hello"));
console.log(Buffer.byteLength("68656c6c6f", "hex"));
console.log(Buffer.byteLength("aGVsbG8=", "base64"));

// --- Context Case: scenarios_crypto_basic ---
// @expect: 36
// @expect: 8
const id1_crypto_basic_4: string = crypto.randomUUID();
console.log(id1_crypto_basic_4.length);
console.log(id1_crypto_basic_4.indexOf("-"));

// --- Context Case: scenarios_json_nested ---
// @expect: {"city":"San Francisco","zip":94105}
// @expect: {"id":1,"name":"Bob","company":{"name":"Tech Corp","address":{"city":"San Francisco","zip":94105}}}
interface Address_json_nested_5 {
  city: string;
  zip: number;
}

interface Company_json_nested_5 {
  name: string;
  address: Address_json_nested_5;
}

interface Employee_json_nested_5 {
  id: number;
  name: string;
  company: Company_json_nested_5;
}

const emp_json_nested_5: Employee_json_nested_5 = {
  id: 1,
  name: "Bob",
  company: {
    name: "Tech Corp",
    address: {
      city: "San Francisco",
      zip: 94105,
    },
  },
};

console.log(JSON.stringify(emp_json_nested_5.company.address));
console.log(JSON.stringify(emp_json_nested_5));

// --- Context Case: scenarios_json_nested_structures ---
// @expect: {"host":"localhost","port":8080,"enabled":true,"tags":["api","v1","production"]}
// @expect: [10,20,30]
// @expect: ["alpha","beta","gamma"]
// @expect: true
// @expect: false
// @expect: 12345
// @expect: "hello world"
interface ServerConfig_json_nested_structures_6 {
    host: string;
    port: number;
    enabled: boolean;
    tags: string[];
}

const config_json_nested_structures_6: ServerConfig_json_nested_structures_6 = {
    host: "localhost",
    port: 8080,
    enabled: true,
    tags: ["api", "v1", "production"],
};

const jsonStr_json_nested_structures_6: string = JSON.stringify(config_json_nested_structures_6);
console.log(jsonStr_json_nested_structures_6);

console.log(JSON.stringify([10, 20, 30]));
console.log(JSON.stringify(["alpha", "beta", "gamma"]));
console.log(JSON.stringify(true));
console.log(JSON.stringify(false));
console.log(JSON.stringify(12345));
console.log(JSON.stringify("hello world"));

// --- Context Case: scenarios_json_object ---
// @expect: {"id":101,"name":"Alice","active":true,"scores":[95,100],"tags":["admin","dev"]}
interface UserProfile_json_object_7 {
    id: number;
    name: string;
    active: boolean;
    scores: number[];
    tags: string[];
}

const profile_json_object_7: UserProfile_json_object_7 = {
    id: 101,
    name: "Alice",
    active: true,
    scores: [95, 100],
    tags: ["admin", "dev"],
};

console.log(JSON.stringify(profile_json_object_7));

// --- Context Case: scenarios_util_text_encoding ---
// @expect: utf-8
// @expect: 21
// @expect: 72
// @expect: 83
// @expect: utf-8
// @expect: false
// @expect: false
// @expect: Hello, ScriptGo! 🚀
// @expect: 10
// @expect: 10
// @expect: Hello Worl
// @expect: 0
const encoder_util_text_encoding_8 = new TextEncoder();
console.log(encoder_util_text_encoding_8.encoding);

const encoded_util_text_encoding_8 = encoder_util_text_encoding_8.encode("Hello, ScriptGo! 🚀");
console.log(encoded_util_text_encoding_8.length);
console.log(encoded_util_text_encoding_8[0]);
console.log(encoded_util_text_encoding_8[7]);

const decoder_util_text_encoding_8 = new TextDecoder();
console.log(decoder_util_text_encoding_8.encoding);
console.log(decoder_util_text_encoding_8.fatal);
console.log(decoder_util_text_encoding_8.ignoreBOM);

const decoded_util_text_encoding_8 = decoder_util_text_encoding_8.decode(encoded_util_text_encoding_8);
console.log(decoded_util_text_encoding_8);

const dest_util_text_encoding_8 = new Uint8Array(10);
const res_util_text_encoding_8 = encoder_util_text_encoding_8.encodeInto("Hello World", dest_util_text_encoding_8);
console.log(res_util_text_encoding_8.read);
console.log(res_util_text_encoding_8.written);
console.log(decoder_util_text_encoding_8.decode(dest_util_text_encoding_8));

const empty_util_text_encoding_8 = encoder_util_text_encoding_8.encode("");
console.log(empty_util_text_encoding_8.length);
console.log(decoder_util_text_encoding_8.decode(empty_util_text_encoding_8));
// @expect:
