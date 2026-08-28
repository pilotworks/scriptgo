// @expect: undefined
// @expect: 42
// @expect: undefined
// @expect: undefined
// @expect: hello
// @expect: undefined
// @expect: undefined
// @expect: true
// @expect: false
// @expect: undefined
// @expect: undefined
// @expect: 100n
// @expect: undefined
// @expect: undefined
// @expect: 10,20
// @expect: undefined
// @expect: undefined
// @expect: 3
// @expect: undefined
// @expect: undefined
// @expect: called
// @expect: undefined
// @expect: undefined
// @expect: 123
// @expect: scriptgo
// @expect: undefined
// @expect: undefined
// @expect: 456
// @expect: true
// @expect: undefined
// @expect: undefined
// @expect: world
// @expect: false
// @expect: undefined
// @expect: undefined
// @expect: 789
// @expect: 1
// @expect: undefined
// @expect: undefined
// @expect: 1000
// @expect: quad
// @expect: true
// @expect: undefined
// @expect: null
// @expect: 55
// @expect: null
// @expect: null
// @expect: nullable
// @expect: null
// @expect: null
// @expect: true
// @expect: null
// @expect: null
// @expect: 999n
// @expect: null
// @expect: null
// @expect: 40
// @expect: null
// @expect: null
// @expect: 2
// @expect: null
// @expect: null
// @expect: 77
// @expect: text
// @expect: null
// @expect: null
// @expect: 88
// @expect: false
// @expect: null
// @expect: null
// @expect: 99
// @expect: all-null
// @expect: true
// @expect: null
// @expect: undefined
// @expect: 1
// @expect: null
// @expect: undefined
// @expect: undefined
// @expect: str-nud
// @expect: null
// @expect: undefined
// @expect: undefined
// @expect: true
// @expect: null
// @expect: undefined
// @expect: undefined
// @expect: 12345n
// @expect: null
// @expect: undefined
// @expect: undefined
// @expect: 50
// @expect: null
// @expect: undefined
// @expect: undefined
// @expect: 2
// @expect: null
// @expect: undefined
// @expect: undefined
// @expect: 42
// @expect: mixed
// @expect: null
// @expect: undefined
// @expect: undefined
// @expect: 500
// @expect: poly
// @expect: false
// @expect: null
// @expect: undefined

// Section 1: Single types with undefined (uninitialized and initialized)
let u_num: number | undefined;
console.log(u_num);
u_num = 42;
console.log(u_num);
u_num = undefined;
console.log(u_num);

let u_str: string | undefined;
console.log(u_str);
u_str = "hello";
console.log(u_str);
u_str = undefined;
console.log(u_str);

let u_bool: boolean | undefined;
console.log(u_bool);
u_bool = true;
console.log(u_bool);
u_bool = false;
console.log(u_bool);
u_bool = undefined;
console.log(u_bool);

let u_bigint: bigint | undefined;
console.log(u_bigint);
u_bigint = 100n;
console.log(u_bigint);
u_bigint = undefined;
console.log(u_bigint);

type Point = { x: number; y: number };
let u_obj: Point | undefined;
console.log(u_obj);
u_obj = { x: 10, y: 20 };
console.log(u_obj !== undefined ? u_obj.x + "," + u_obj.y : "none");
u_obj = undefined;
console.log(u_obj);

let u_arr: number[] | undefined;
console.log(u_arr);
u_arr = [1, 2, 3];
console.log(u_arr !== undefined ? u_arr.length : -1);
u_arr = undefined;
console.log(u_arr);

let u_fn: (() => string) | undefined;
console.log(u_fn);
u_fn = () => "called";
console.log(u_fn !== undefined ? u_fn() : "none");
u_fn = undefined;
console.log(u_fn);

// Section 2: 3-way unions with undefined
let u3_ns: number | string | undefined;
console.log(u3_ns);
u3_ns = 123;
console.log(u3_ns);
u3_ns = "scriptgo";
console.log(u3_ns);
u3_ns = undefined;
console.log(u3_ns);

let u3_nb: number | boolean | undefined;
console.log(u3_nb);
u3_nb = 456;
console.log(u3_nb);
u3_nb = true;
console.log(u3_nb);
u3_nb = undefined;
console.log(u3_nb);

let u3_sb: string | boolean | undefined;
console.log(u3_sb);
u3_sb = "world";
console.log(u3_sb);
u3_sb = false;
console.log(u3_sb);
u3_sb = undefined;
console.log(u3_sb);

let u3_no: number | Point | undefined;
console.log(u3_no);
u3_no = 789;
console.log(u3_no);
u3_no = { x: 1, y: 2 };
console.log(typeof u3_no === "object" && u3_no !== null ? u3_no.x : "other");
u3_no = undefined;
console.log(u3_no);

// 4-way union with undefined
let u4_nsb: number | string | boolean | undefined;
console.log(u4_nsb);
u4_nsb = 1000;
console.log(u4_nsb);
u4_nsb = "quad";
console.log(u4_nsb);
u4_nsb = true;
console.log(u4_nsb);
u4_nsb = undefined;
console.log(u4_nsb);

// Section 3: Null distributions
let n_num: number | null = null;
console.log(n_num);
n_num = 55;
console.log(n_num);
n_num = null;
console.log(n_num);

let n_str: string | null = null;
console.log(n_str);
n_str = "nullable";
console.log(n_str);
n_str = null;
console.log(n_str);

let n_bool: boolean | null = null;
console.log(n_bool);
n_bool = true;
console.log(n_bool);
n_bool = null;
console.log(n_bool);

let n_bi: bigint | null = null;
console.log(n_bi);
n_bi = 999n;
console.log(n_bi);
n_bi = null;
console.log(n_bi);

let n_obj: Point | null = null;
console.log(n_obj);
n_obj = { x: 30, y: 40 };
console.log(n_obj !== null ? n_obj.y : -1);
n_obj = null;
console.log(n_obj);

let n_arr: number[] | null = null;
console.log(n_arr);
n_arr = [10, 20];
console.log(n_arr !== null ? n_arr.length : -1);
n_arr = null;
console.log(n_arr);

let n_ns: number | string | null = null;
console.log(n_ns);
n_ns = 77;
console.log(n_ns);
n_ns = "text";
console.log(n_ns);
n_ns = null;
console.log(n_ns);

let n_nb: number | boolean | null = null;
console.log(n_nb);
n_nb = 88;
console.log(n_nb);
n_nb = false;
console.log(n_nb);
n_nb = null;
console.log(n_nb);

let n_nsb: number | string | boolean | null = null;
console.log(n_nsb);
n_nsb = 99;
console.log(n_nsb);
n_nsb = "all-null";
console.log(n_nsb);
n_nsb = true;
console.log(n_nsb);
n_nsb = null;
console.log(n_nsb);

// Section 4: Distributions with both null and undefined
let nud_num: number | null | undefined;
console.log(nud_num);
nud_num = 1;
console.log(nud_num);
nud_num = null;
console.log(nud_num);
nud_num = undefined;
console.log(nud_num);

let nud_str: string | null | undefined;
console.log(nud_str);
nud_str = "str-nud";
console.log(nud_str);
nud_str = null;
console.log(nud_str);
nud_str = undefined;
console.log(nud_str);

let nud_bool: boolean | null | undefined;
console.log(nud_bool);
nud_bool = true;
console.log(nud_bool);
nud_bool = null;
console.log(nud_bool);
nud_bool = undefined;
console.log(nud_bool);

let nud_bi: bigint | null | undefined;
console.log(nud_bi);
nud_bi = 12345n;
console.log(nud_bi);
nud_bi = null;
console.log(nud_bi);
nud_bi = undefined;
console.log(nud_bi);

let nud_obj: Point | null | undefined;
console.log(nud_obj);
nud_obj = { x: 50, y: 60 };
console.log(nud_obj ? nud_obj.x : "nil");
nud_obj = null;
console.log(nud_obj);
nud_obj = undefined;
console.log(nud_obj);

let nud_arr: string[] | null | undefined;
console.log(nud_arr);
nud_arr = ["a", "b"];
console.log(nud_arr ? nud_arr.length : -1);
nud_arr = null;
console.log(nud_arr);
nud_arr = undefined;
console.log(nud_arr);

let nud_ns: number | string | null | undefined;
console.log(nud_ns);
nud_ns = 42;
console.log(nud_ns);
nud_ns = "mixed";
console.log(nud_ns);
nud_ns = null;
console.log(nud_ns);
nud_ns = undefined;
console.log(nud_ns);

let nud_nsb: number | string | boolean | null | undefined;
console.log(nud_nsb);
nud_nsb = 500;
console.log(nud_nsb);
nud_nsb = "poly";
console.log(nud_nsb);
nud_nsb = false;
console.log(nud_nsb);
nud_nsb = null;
console.log(nud_nsb);
nud_nsb = undefined;
console.log(nud_nsb);
