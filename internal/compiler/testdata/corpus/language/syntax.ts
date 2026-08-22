// ScriptGo Corpus: Language: Syntax & Control Flow
// Consolidated test suite with inline assertions.

// --- Context Case: language_in_operator ---
// @expect: true
// @expect: true
// @expect: false
// @expect: true
// @expect: false
// @expect: true
// @expect: true
// @expect: false
// @expect: false
// @expect: true
// @expect: true
// @expect: false
class Item_in_operator_0 {
    id: number;
    title: string;
    constructor(id: number, title: string) {
        this.id = id;
        this.title = title;
    }
}

let user_in_operator_0 = { name: "Alice", age: 30 };
console.log("name" in user_in_operator_0);
console.log("age" in user_in_operator_0);
console.log("email" in user_in_operator_0);

let k1_in_operator_0 = "name";
console.log(k1_in_operator_0 in user_in_operator_0);

let k2_in_operator_0 = "score";
console.log(k2_in_operator_0 in user_in_operator_0);

let list_in_operator_0 = [10, 20, 30];
console.log(0 in list_in_operator_0);
console.log(2 in list_in_operator_0);
console.log(3 in list_in_operator_0);
console.log(-1 in list_in_operator_0);

let item_in_operator_0 = new Item_in_operator_0(1, "Book");
console.log("id" in item_in_operator_0);
console.log("title" in item_in_operator_0);
console.log("price" in item_in_operator_0);

// --- Context Case: language_labeled_statement ---
// @expect: 12
// @expect: 8
// @expect: 6
let sum_labeled_statement_1 = 0;
outer: for (let i_labeled_statement_1 = 0; i_labeled_statement_1 < 5; i_labeled_statement_1++) {
    for (let j_labeled_statement_1 = 0; j_labeled_statement_1 < 5; j_labeled_statement_1++) {
        if (i_labeled_statement_1 === 2 && j_labeled_statement_1 === 2) {
            break outer;
        }
        sum_labeled_statement_1 += 1;
    }
}
console.log(sum_labeled_statement_1);

let contCount_labeled_statement_1 = 0;
loopA: for (let i_labeled_statement_1 = 0; i_labeled_statement_1 < 4; i_labeled_statement_1++) {
    for (let j_labeled_statement_1 = 0; j_labeled_statement_1 < 4; j_labeled_statement_1++) {
        if (j_labeled_statement_1 === 2) {
            continue loopA;
        }
        contCount_labeled_statement_1 += 1;
    }
}
console.log(contCount_labeled_statement_1);

let whileSum_labeled_statement_1 = 0;
let x_labeled_statement_1 = 0;
outerWhile: while (x_labeled_statement_1 < 3) {
    x_labeled_statement_1++;
    let y_labeled_statement_1 = 0;
    while (y_labeled_statement_1 < 3) {
        y_labeled_statement_1++;
        if (x_labeled_statement_1 === 2) {
            continue outerWhile;
        }
        whileSum_labeled_statement_1 += 1;
    }
}
console.log(whileSum_labeled_statement_1);

// --- Context Case: language_optional_call ---
// @expect: Hello, Alice
// @expect: Welcome, Bob!
// @expect: 42
function greet_optional_call_2(name: string): string {
    return "Hello, " + name;
}

class Greeter_optional_call_2 {
    greeting: string;
    constructor(greeting: string) {
        this.greeting = greeting;
    }
    say(name: string): string {
        return this.greeting + ", " + name + "!";
    }
}

console.log(greet_optional_call_2?.("Alice"));

const greeter_optional_call_2 = new Greeter_optional_call_2("Welcome");
console.log(greeter_optional_call_2?.say?.("Bob"));

const fn_optional_call_2 = (x: number) => x * 2;
console.log(fn_optional_call_2?.(21));

// --- Context Case: language_postfix_prefix_update ---
// @expect: 1
// @expect: 2
// @expect: 3
// @expect: 3
// @expect: 3
// @expect: 2
// @expect: 1
// @expect: 1
// @expect: 10
// @expect: 1
// @expect: 10
// @expect: 11
// @expect: 10
// @expect: 10
// @expect: 5
// @expect: 6
// @expect: 7
// @expect: 7
// @expect: 7
// @expect: 6
// @expect: 5
// @expect: 5
// @expect: 6
// @expect: 4
// @expect: 2
// @expect: 1
// @expect: 0
// @expect: 0
// @expect: 1
// @expect: 2
function add_postfix_prefix_update_3(x: number, y: number): number {
    return x + y;
}

let a_postfix_prefix_update_3 = 1;
let b_postfix_prefix_update_3 = a_postfix_prefix_update_3++;
console.log(b_postfix_prefix_update_3);
console.log(a_postfix_prefix_update_3);

let c_postfix_prefix_update_3 = ++a_postfix_prefix_update_3;
console.log(c_postfix_prefix_update_3);
console.log(a_postfix_prefix_update_3);

let d_postfix_prefix_update_3 = a_postfix_prefix_update_3--;
console.log(d_postfix_prefix_update_3);
console.log(a_postfix_prefix_update_3);

let e_postfix_prefix_update_3 = --a_postfix_prefix_update_3;
console.log(e_postfix_prefix_update_3);
console.log(a_postfix_prefix_update_3);

let arr_postfix_prefix_update_3 = [10, 20, 30];
let i_postfix_prefix_update_3 = 0;
console.log(arr_postfix_prefix_update_3[i_postfix_prefix_update_3++]);
console.log(i_postfix_prefix_update_3);
console.log(arr_postfix_prefix_update_3[0]++);
console.log(arr_postfix_prefix_update_3[0]);
console.log(--arr_postfix_prefix_update_3[0]);
console.log(arr_postfix_prefix_update_3[0]);

let obj_postfix_prefix_update_3 = { count: 5 };
console.log(obj_postfix_prefix_update_3.count++);
console.log(obj_postfix_prefix_update_3.count);
console.log(++obj_postfix_prefix_update_3.count);
console.log(obj_postfix_prefix_update_3.count);
console.log(obj_postfix_prefix_update_3.count--);
console.log(obj_postfix_prefix_update_3.count);
console.log(--obj_postfix_prefix_update_3.count);
console.log(obj_postfix_prefix_update_3.count);

let n_postfix_prefix_update_3 = 2;
console.log(add_postfix_prefix_update_3(n_postfix_prefix_update_3++, ++n_postfix_prefix_update_3));
console.log(n_postfix_prefix_update_3);

let loopCount_postfix_prefix_update_3 = 3;
while (loopCount_postfix_prefix_update_3-- > 0) {
    console.log(loopCount_postfix_prefix_update_3);
}

for (let k_postfix_prefix_update_3 = 0; k_postfix_prefix_update_3 < 3; k_postfix_prefix_update_3++) {
    console.log(k_postfix_prefix_update_3);
}

// --- Context Case: language_syntax_complex-nested-loops-labels ---
// @expect: breaking inner at j=4
// @expect: breaking inner at j=4
// @expect: skipping outer at i=2, j=2
// @expect: breaking inner at j=4
// @expect: breaking outer at i=4, j=3
// @expect: 242
let sum_syntax_complex_nested_loops_labels_4: number = 0;

outerLoop: for (let i_syntax_complex_nested_loops_labels_4: number = 0; i_syntax_complex_nested_loops_labels_4 < 5; i_syntax_complex_nested_loops_labels_4++) {
    let j_syntax_complex_nested_loops_labels_4: number = 0;
    innerLoop: while (j_syntax_complex_nested_loops_labels_4 < 5) {
        j_syntax_complex_nested_loops_labels_4++;
        if (i_syntax_complex_nested_loops_labels_4 === 2 && j_syntax_complex_nested_loops_labels_4 === 2) {
            console.log("skipping outer at i=2, j=2");
            continue outerLoop;
        }
        if (i_syntax_complex_nested_loops_labels_4 === 4 && j_syntax_complex_nested_loops_labels_4 === 3) {
            console.log("breaking outer at i=4, j=3");
            break outerLoop;
        }
        if (j_syntax_complex_nested_loops_labels_4 === 4) {
            console.log("breaking inner at j=4");
            break innerLoop;
        }
        sum_syntax_complex_nested_loops_labels_4 += (i_syntax_complex_nested_loops_labels_4 * 10) + j_syntax_complex_nested_loops_labels_4;
    }
}

console.log(sum_syntax_complex_nested_loops_labels_4);

// --- Context Case: language_syntax_debugger ---
// @expect: start
// @expect: 42
// @expect: done
function compute_syntax_debugger_5(a: number, b: number): number {
    debugger;
    const sum_syntax_debugger_5 = a + b;
    debugger;
    return sum_syntax_debugger_5;
}

debugger;
console.log("start");
const result_syntax_debugger_5 = compute_syntax_debugger_5(20, 22);
debugger;
console.log(result_syntax_debugger_5);
console.log("done");

// --- Context Case: language_syntax_default-params ---
// @expect: Hello, Alice!
// @expect: Good morning, Bob!
// @expect: 45
// @expect: 11
// @expect: 10
function greet_syntax_default_params_6(name: string, greeting: string = "Hello"): string {
    return greeting + ", " + name + "!";
}

function compute_syntax_default_params_6(a: number, b: number = 20, c: number = 5): number {
    return a * b + c;
}

console.log(greet_syntax_default_params_6("Alice"));
console.log(greet_syntax_default_params_6("Bob", "Good morning"));
console.log(compute_syntax_default_params_6(2));
console.log(compute_syntax_default_params_6(2, 3));
console.log(compute_syntax_default_params_6(2, 3, 4));

// --- Context Case: language_syntax_destructuring ---
// @expect: 10
// @expect: 25
// @expect: 100
// @expect: 200
// @expect: 2
// @expect: 300
// @expect: 400
class Point_syntax_destructuring_7 {
    x: number;
    y: number;
    constructor(x: number, y: number) {
        this.x = x;
        this.y = y;
    }
}

const pt_syntax_destructuring_7: Point_syntax_destructuring_7 = new Point_syntax_destructuring_7(10, 25);
const { x, y } = pt_syntax_destructuring_7;
console.log(x);
console.log(y);

const numbers_syntax_destructuring_7: number[] = [100, 200, 300, 400];
const [first, second, ...rest] = numbers_syntax_destructuring_7;
console.log(first);
console.log(second);
console.log(rest.length);
console.log(rest[0]);
console.log(rest[1]);

// --- Context Case: language_syntax_enums ---
// @expect: 0
// @expect: 1
// @expect: 10
// @expect: 11
// @expect: RED
// @expect: GREEN
// @expect: BLUE
enum Status_syntax_enums_8 {
    Active,
    Inactive,
    Pending = 10,
    Done,
}

enum Color_syntax_enums_8 {
    Red = "RED",
    Green = "GREEN",
    Blue = "BLUE",
}

console.log(Status_syntax_enums_8.Active);
console.log(Status_syntax_enums_8.Inactive);
console.log(Status_syntax_enums_8.Pending);
console.log(Status_syntax_enums_8.Done);
console.log(Color_syntax_enums_8.Red);
console.log(Color_syntax_enums_8.Green);
console.log(Color_syntax_enums_8.Blue);

// --- Context Case: language_syntax_enums-reverse ---
// @expect: 0
// @expect: 1
// @expect: 2
// @expect: Red
// @expect: Green
// @expect: Blue
// @expect: 200
// @expect: OK
// @expect: NotFound
// @expect: INFO
// @expect: ERROR
enum Color_syntax_enums_reverse_9 {
  Red,
  Green,
  Blue,
}

enum HttpStatus_syntax_enums_reverse_9 {
  OK = 200,
  NotFound = 404,
}

enum LogLevel_syntax_enums_reverse_9 {
  Info = "INFO",
  Error = "ERROR",
}

console.log(Color_syntax_enums_reverse_9.Red);
console.log(Color_syntax_enums_reverse_9.Green);
console.log(Color_syntax_enums_reverse_9.Blue);

console.log(Color_syntax_enums_reverse_9[0]);
console.log(Color_syntax_enums_reverse_9[1]);
console.log(Color_syntax_enums_reverse_9[2]);

console.log(HttpStatus_syntax_enums_reverse_9.OK);
console.log(HttpStatus_syntax_enums_reverse_9[200]);
console.log(HttpStatus_syntax_enums_reverse_9[404]);

console.log(LogLevel_syntax_enums_reverse_9.Info);
console.log(LogLevel_syntax_enums_reverse_9.Error);

// --- Context Case: language_syntax_for-in ---
// @expect: 0
// @expect: 1
// @expect: 2
// @expect: host
// @expect: port
// 1. for...in on array
const arr_syntax_for_in_10: string[] = ["x", "y", "z"];
for (const key in arr_syntax_for_in_10) {
    console.log(key);
}

// 2. for...in on static object shape
class Config_syntax_for_in_10 {
    host: string;
    port: number;
    constructor(host: string, port: number) {
        this.host = host;
        this.port = port;
    }
}

const cfg_syntax_for_in_10: Config_syntax_for_in_10 = new Config_syntax_for_in_10("localhost", 8080);
for (const prop in cfg_syntax_for_in_10) {
    console.log(prop);
}

// --- Context Case: language_syntax_for-of-advanced ---
// @expect: a
// @expect: b
// @expect: c
// @expect: true
// @expect: false
// @expect: true
// @expect: 1
// @expect: A
// @expect: 2
// @expect: B
// @expect: 80
// 1. Iterate over string characters
const text_syntax_for_of_advanced_11: string = "abc";
for (const ch of text_syntax_for_of_advanced_11) {
    console.log(ch);
}

// 2. Iterate over boolean array
const flags_syntax_for_of_advanced_11: boolean[] = [true, false, true];
for (const f of flags_syntax_for_of_advanced_11) {
    console.log(f);
}

// 3. Iterate over object array
class Item_syntax_for_of_advanced_11 {
    id: number;
    name: string;
    constructor(id: number, name: string) {
        this.id = id;
        this.name = name;
    }
}

const items_syntax_for_of_advanced_11: Item_syntax_for_of_advanced_11[] = [new Item_syntax_for_of_advanced_11(1, "A"), new Item_syntax_for_of_advanced_11(2, "B")];
for (const item of items_syntax_for_of_advanced_11) {
    console.log(item.id);
    console.log(item.name);
}

// 4. Continue and break in for...of
const nums_syntax_for_of_advanced_11: number[] = [10, 20, 30, 40, 50];
let sum_syntax_for_of_advanced_11: number = 0;
for (const n of nums_syntax_for_of_advanced_11) {
    if (n === 20) {
        continue;
    }
    if (n === 50) {
        break;
    }
    sum_syntax_for_of_advanced_11 += n;
}
console.log(sum_syntax_for_of_advanced_11); // 10 + 30 + 40 = 80

// --- Context Case: language_syntax_for-of-destructuring ---
// @expect: 1
// @expect: 2
// @expect: 10
// @expect: 20
// @expect: apple
// @expect: 5
// @expect: banana
// @expect: 10
class Point_syntax_for_of_destructuring_12 {
    x: number;
    y: number;
    constructor(x: number, y: number) {
        this.x = x;
        this.y = y;
    }
}

const points_syntax_for_of_destructuring_12: Point_syntax_for_of_destructuring_12[] = [new Point_syntax_for_of_destructuring_12(1, 2), new Point_syntax_for_of_destructuring_12(10, 20)];
for (const { x, y } of points_syntax_for_of_destructuring_12) {
    console.log(x);
    console.log(y);
}

const pairs: [string, number][] = [["apple", 5], ["banana", 10]];
for (const [fruit, count] of pairs) {
    console.log(fruit);
    console.log(count);
}

// --- Context Case: language_syntax_nullish ---
// @expect: 3000
// @expect: 8080
// @expect: 8080
// @expect: App
// @expect: 5000
// @expect: 11
class Config_syntax_nullish_13 {
    title: string;
    port: number;
    constructor(title: string, port: number) {
        this.title = title;
        this.port = port;
    }
}

function getPort_syntax_nullish_13(port: string | null | undefined): string {
    const fallback_syntax_nullish_13: string = "8080";
    return port ?? fallback_syntax_nullish_13;
}

console.log(getPort_syntax_nullish_13("3000"));
console.log(getPort_syntax_nullish_13(null));
console.log(getPort_syntax_nullish_13(undefined));

const cfg_syntax_nullish_13: Config_syntax_nullish_13 = new Config_syntax_nullish_13("App", 5000);
console.log(cfg_syntax_nullish_13?.title);
console.log(cfg_syntax_nullish_13?.port);

const str_syntax_nullish_13: string = "hello world";
console.log(str_syntax_nullish_13?.length);

// --- Context Case: language_syntax_spread ---
// @expect: 5
// @expect: 1
// @expect: 2
// @expect: 99
// @expect: 3
// @expect: 4
// @expect: 100
const a_syntax_spread_14: number[] = [1, 2];
const b_syntax_spread_14: number[] = [3, 4];
const combined_syntax_spread_14: number[] = [...a_syntax_spread_14, 99, ...b_syntax_spread_14];
console.log(combined_syntax_spread_14.length);
console.log(combined_syntax_spread_14[0]);
console.log(combined_syntax_spread_14[1]);
console.log(combined_syntax_spread_14[2]);
console.log(combined_syntax_spread_14[3]);
console.log(combined_syntax_spread_14[4]);

function sumAll_syntax_spread_14(...vals: number[]): number {
    let sum_syntax_spread_14: number = 0;
    for (const val of vals) {
        sum_syntax_spread_14 += val;
    }
    return sum_syntax_spread_14;
}

console.log(sumAll_syntax_spread_14(10, 20, 30, 40));

// --- Context Case: language_syntax_string-index-access ---
// @expect: H
// @expect: e
// @expect: l
// @expect: l
// @expect: o
// @expect: T
// @expect: S
// @expect: t
const greeting_syntax_string_index_access_15: string = "Hello";
console.log(greeting_syntax_string_index_access_15[0]);
console.log(greeting_syntax_string_index_access_15[1]);
console.log(greeting_syntax_string_index_access_15[2]);
console.log(greeting_syntax_string_index_access_15[3]);
console.log(greeting_syntax_string_index_access_15[4]);

const text_syntax_string_index_access_15: string = "TypeScript";
console.log(text_syntax_string_index_access_15[0]);
console.log(text_syntax_string_index_access_15[4]);
console.log(text_syntax_string_index_access_15[9]);

// --- Context Case: language_syntax_switch-advanced ---
// @expect: two
// @expect: 1
// @expect: start 1
// @expect: fallthrough to 2
// @expect: 60
class Counter_syntax_switch_advanced_16 {
    val: number;
    constructor(val: number) {
        this.val = val;
    }
    next(): number {
        this.val += 1;
        return 2;
    }
}

const c_syntax_switch_advanced_16: Counter_syntax_switch_advanced_16 = new Counter_syntax_switch_advanced_16(0);

// 1. Test single evaluation of switch expression
switch (c_syntax_switch_advanced_16.next()) {
    case 1:
        console.log("one");
        break;
    case 2:
        console.log("two");
        break;
    default:
        console.log("other");
        break;
}
console.log(c_syntax_switch_advanced_16.val); // Must be 1

// 2. Test fallthrough across non-empty statements
let tag_syntax_switch_advanced_16: number = 1;
switch (tag_syntax_switch_advanced_16) {
    case 1:
        console.log("start 1");
    case 2:
        console.log("fallthrough to 2");
        break;
    case 3:
        console.log("3");
        break;
    default:
        console.log("def");
        break;
}

// 3. Test break inside loop only breaks switch
let loopCount_syntax_switch_advanced_16: number = 0;
for (let i_syntax_switch_advanced_16: number = 0; i_syntax_switch_advanced_16 < 3; i_syntax_switch_advanced_16 += 1) {
    switch (i_syntax_switch_advanced_16) {
        case 0:
            loopCount_syntax_switch_advanced_16 += 10;
            break; // should break switch, not for-loop!
        case 1:
            loopCount_syntax_switch_advanced_16 += 20;
            break;
        default:
            loopCount_syntax_switch_advanced_16 += 30;
            break;
    }
}
console.log(loopCount_syntax_switch_advanced_16); // 10 + 20 + 30 = 60

// --- Context Case: language_syntax_switch-fallthrough-patterns ---
// @expect: bronze-silver-gold-tier
// @expect: silver-gold-tier
// @expect: gold-tier
// @expect: platinum-special
// @expect: unknown-tier
// @expect: 31
// @expect: 28
// @expect: 30
// @expect: 31
// @expect: -1
function evaluateTier_syntax_switch_fallthrough_patterns_17(tier: string): string {
    let result_syntax_switch_fallthrough_patterns_17: string = "";
    switch (tier) {
        case "bronze":
            result_syntax_switch_fallthrough_patterns_17 += "bronze-";
            // fallthrough
        case "silver":
            result_syntax_switch_fallthrough_patterns_17 += "silver-";
            // fallthrough
        case "gold":
            result_syntax_switch_fallthrough_patterns_17 += "gold-tier";
            break;
        case "platinum":
            result_syntax_switch_fallthrough_patterns_17 += "platinum-special";
            break;
        default:
            result_syntax_switch_fallthrough_patterns_17 = "unknown-tier";
            break;
    }
    return result_syntax_switch_fallthrough_patterns_17;
}

console.log(evaluateTier_syntax_switch_fallthrough_patterns_17("bronze"));
console.log(evaluateTier_syntax_switch_fallthrough_patterns_17("silver"));
console.log(evaluateTier_syntax_switch_fallthrough_patterns_17("gold"));
console.log(evaluateTier_syntax_switch_fallthrough_patterns_17("platinum"));
console.log(evaluateTier_syntax_switch_fallthrough_patterns_17("diamond"));

function getDaysInMonth_syntax_switch_fallthrough_patterns_17(month: number): number {
    let days_syntax_switch_fallthrough_patterns_17: number = 0;
    switch (month) {
        case 1:
        case 3:
        case 5:
        case 7:
        case 8:
        case 10:
        case 12:
            days_syntax_switch_fallthrough_patterns_17 = 31;
            break;
        case 4:
        case 6:
        case 9:
        case 11:
            days_syntax_switch_fallthrough_patterns_17 = 30;
            break;
        case 2:
            days_syntax_switch_fallthrough_patterns_17 = 28;
            break;
        default:
            days_syntax_switch_fallthrough_patterns_17 = -1;
            break;
    }
    return days_syntax_switch_fallthrough_patterns_17;
}

console.log(getDaysInMonth_syntax_switch_fallthrough_patterns_17(1));
console.log(getDaysInMonth_syntax_switch_fallthrough_patterns_17(2));
console.log(getDaysInMonth_syntax_switch_fallthrough_patterns_17(4));
console.log(getDaysInMonth_syntax_switch_fallthrough_patterns_17(12));
console.log(getDaysInMonth_syntax_switch_fallthrough_patterns_17(13));

// --- Context Case: language_syntax_tuple-mutation ---
// @expect: score
// @expect: 100
// @expect: new_score
// @expect: 200
const pair: [string, number] = ["score", 100];
console.log(pair[0]);
console.log(pair[1]);

pair[0] = "new_score";
pair[1] = 200;
console.log(pair[0]);
console.log(pair[1]);

// --- Context Case: language_syntax_tuples ---
// @expect: Alice
// @expect: 25
// @expect: 10
// @expect: 20
// @expect: point
const user: [string, number] = ["Alice", 25];
console.log(user[0]);
console.log(user[1]);

const coord: [number, number, string] = [10, 20, "point"];
console.log(coord[0]);
console.log(coord[1]);
console.log(coord[2]);

// --- Context Case: language_tagged_template ---
// @expect: User: Alice, Age: 30.
// @expect: Hello world (tagged)
function tag_tagged_template_20(strings: TemplateStringsArray, name: string, age: number): string {
    return strings[0] + name + strings[1] + age + strings[2];
}

const user_tagged_template_20 = "Alice";
const years_tagged_template_20 = 30;
const formatted_tagged_template_20 = tag_tagged_template_20`User: ${user_tagged_template_20}, Age: ${years_tagged_template_20}.`;
console.log(formatted_tagged_template_20);

function simpleTag_tagged_template_20(strings: TemplateStringsArray): string {
    return strings[0] + " (tagged)";
}

console.log(simpleTag_tagged_template_20`Hello world`);
