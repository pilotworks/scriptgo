// ScriptGo Corpus: Language: Expressions & Operators
// Consolidated test suite with inline assertions.

// --- Context Case: language_expressions_arithmetic ---
// @expect: 12
// @expect: 2
const quotient_expressions_arithmetic_0: number = (20 + 4) / 2;
const remainder_expressions_arithmetic_0: number = 17 % 5;
console.log(quotient_expressions_arithmetic_0);
console.log(remainder_expressions_arithmetic_0);

// --- Context Case: language_expressions_bitwise-operators ---
// @expect: 8
// @expect: 15
// @expect: 7
// @expect: -15
// @expect: 16
// @expect: 8
// @expect: 4294967295
// @expect: -4
const a_expressions_bitwise_operators_1: number = 14;
const b_expressions_bitwise_operators_1: number = 9;
console.log(a_expressions_bitwise_operators_1 & b_expressions_bitwise_operators_1);
console.log(a_expressions_bitwise_operators_1 | b_expressions_bitwise_operators_1);
console.log(a_expressions_bitwise_operators_1 ^ b_expressions_bitwise_operators_1);
console.log(~a_expressions_bitwise_operators_1);
console.log(1 << 4);
console.log(32 >> 2);
console.log(-1 >>> 0);
console.log(-16 >> 2);

// --- Context Case: language_expressions_bitwise-shifts-masks ---
// @expect: 120
// @expect: 86
// @expect: 52
// @expect: 18
// @expect: -4
// @expect: 1073741820
// @expect: 85
// @expect: -1
// @expect: -2147483648
// @expect: -1
// @expect: 1
const a_expressions_bitwise_shifts_masks_2: number = 0x12345678;
const mask_expressions_bitwise_shifts_masks_2: number = 0x000000FF;

console.log(a_expressions_bitwise_shifts_masks_2 & mask_expressions_bitwise_shifts_masks_2);
console.log((a_expressions_bitwise_shifts_masks_2 >> 8) & mask_expressions_bitwise_shifts_masks_2);
console.log((a_expressions_bitwise_shifts_masks_2 >> 16) & mask_expressions_bitwise_shifts_masks_2);
console.log((a_expressions_bitwise_shifts_masks_2 >> 24) & mask_expressions_bitwise_shifts_masks_2);

const negativeVal_expressions_bitwise_shifts_masks_2: number = -16;
console.log(negativeVal_expressions_bitwise_shifts_masks_2 >> 2);
console.log(negativeVal_expressions_bitwise_shifts_masks_2 >>> 2);

const combined_expressions_bitwise_shifts_masks_2: number = (0x0F | 0xF0) ^ 0xAA;
console.log(combined_expressions_bitwise_shifts_masks_2);

const inverted_expressions_bitwise_shifts_masks_2: number = ~0;
console.log(inverted_expressions_bitwise_shifts_masks_2);

const shiftWrap_expressions_bitwise_shifts_masks_2: number = 1 << 31;
console.log(shiftWrap_expressions_bitwise_shifts_masks_2);
console.log(shiftWrap_expressions_bitwise_shifts_masks_2 >> 31);
console.log(shiftWrap_expressions_bitwise_shifts_masks_2 >>> 31);

// --- Context Case: language_expressions_bool-literal ---
// @expect: true
const enabled_expressions_bool_literal_3: boolean = true;
console.log(enabled_expressions_bool_literal_3);

// --- Context Case: language_expressions_break-continue ---
// @expect: 1
// @expect: 2
// @expect: 4
// @expect: 5
let i_expressions_break_continue_4: number = 0;
while (i_expressions_break_continue_4 < 10) {
  i_expressions_break_continue_4 = i_expressions_break_continue_4 + 1;
  if (i_expressions_break_continue_4 === 3) {
    continue;
  }
  if (i_expressions_break_continue_4 === 6) {
    break;
  }
  console.log(i_expressions_break_continue_4);
}

// --- Context Case: language_expressions_comparison-operators ---
// @expect: true
// @expect: true
// @expect: true
// @expect: true
// @expect: true
// @expect: true
// @expect: true
// @expect: true
const a_expressions_comparison_operators_5: number = 42;
const b_expressions_comparison_operators_5: number = 42;
const c_expressions_comparison_operators_5: number = 10;
console.log(a_expressions_comparison_operators_5 === b_expressions_comparison_operators_5);
console.log(a_expressions_comparison_operators_5 !== c_expressions_comparison_operators_5);
console.log(a_expressions_comparison_operators_5 == b_expressions_comparison_operators_5);
console.log(a_expressions_comparison_operators_5 != c_expressions_comparison_operators_5);
console.log(a_expressions_comparison_operators_5 > c_expressions_comparison_operators_5);
console.log(c_expressions_comparison_operators_5 < a_expressions_comparison_operators_5);
console.log(a_expressions_comparison_operators_5 >= b_expressions_comparison_operators_5);
console.log(c_expressions_comparison_operators_5 <= a_expressions_comparison_operators_5);

// --- Context Case: language_expressions_complex-switch ---
// @expect: success
// @expect: success
// @expect: redirection
// @expect: client_error
// @expect: server_error
// @expect: unknown
// @expect: 10
// @expect: 15
// @expect: 20
// @expect: 0
function categorizeHttpStatus_expressions_complex_switch_6(code: number): string {
  let category_expressions_complex_switch_6 = "";
  switch (code) {
    case 200:
    case 201:
    case 204:
      category_expressions_complex_switch_6 = "success";
      break;
    case 301:
    case 302:
    case 304:
      category_expressions_complex_switch_6 = "redirection";
      break;
    case 400:
    case 401:
    case 403:
    case 404:
      category_expressions_complex_switch_6 = "client_error";
      break;
    case 500:
    case 502:
    case 503:
      category_expressions_complex_switch_6 = "server_error";
      break;
    default:
      category_expressions_complex_switch_6 = "unknown";
      break;
  }
  return category_expressions_complex_switch_6;
}

function processCommand_expressions_complex_switch_6(cmd: string): number {
  let priority_expressions_complex_switch_6 = 0;
  switch (cmd) {
    case "start":
      priority_expressions_complex_switch_6 = 10;
      break;
    case "stop":
      priority_expressions_complex_switch_6 = 20;
      break;
    case "restart":
      priority_expressions_complex_switch_6 = 15;
      break;
    default:
      priority_expressions_complex_switch_6 = 0;
      break;
  }
  return priority_expressions_complex_switch_6;
}

console.log(categorizeHttpStatus_expressions_complex_switch_6(200));
console.log(categorizeHttpStatus_expressions_complex_switch_6(204));
console.log(categorizeHttpStatus_expressions_complex_switch_6(301));
console.log(categorizeHttpStatus_expressions_complex_switch_6(404));
console.log(categorizeHttpStatus_expressions_complex_switch_6(500));
console.log(categorizeHttpStatus_expressions_complex_switch_6(418));

console.log(processCommand_expressions_complex_switch_6("start"));
console.log(processCommand_expressions_complex_switch_6("restart"));
console.log(processCommand_expressions_complex_switch_6("stop"));
console.log(processCommand_expressions_complex_switch_6("status"));

// --- Context Case: language_expressions_compound-assignments ---
// @expect: 30
// @expect: 15
// @expect: 3
// @expect: 8
// @expect: 13
// @expect: 14
// @expect: 8
// @expect: 4
let a_expressions_compound_assignments_7: number = 10;
a_expressions_compound_assignments_7 *= 3;
console.log(a_expressions_compound_assignments_7);
a_expressions_compound_assignments_7 /= 2;
console.log(a_expressions_compound_assignments_7);
a_expressions_compound_assignments_7 %= 4;
console.log(a_expressions_compound_assignments_7);
let b_expressions_compound_assignments_7: number = 12;
b_expressions_compound_assignments_7 &= 10;
console.log(b_expressions_compound_assignments_7);
b_expressions_compound_assignments_7 |= 5;
console.log(b_expressions_compound_assignments_7);
b_expressions_compound_assignments_7 ^= 3;
console.log(b_expressions_compound_assignments_7);
let c_expressions_compound_assignments_7: number = 1;
c_expressions_compound_assignments_7 <<= 3;
console.log(c_expressions_compound_assignments_7);
c_expressions_compound_assignments_7 >>= 1;
console.log(c_expressions_compound_assignments_7);

// --- Context Case: language_expressions_compound-bitwise ---
// @expect: 4
// @expect: 6
// @expect: 5
// @expect: 16
// @expect: 4
// @expect: 1073741820
// @expect: 42
// @expect: 100
// @expect: 50
// @expect: 50
let a_expressions_compound_bitwise_8 = 12; // 0b1100
let b_expressions_compound_bitwise_8 = 5;  // 0b0101

a_expressions_compound_bitwise_8 &= b_expressions_compound_bitwise_8;
console.log(a_expressions_compound_bitwise_8); // 4 (0b0100)

a_expressions_compound_bitwise_8 |= 2;
console.log(a_expressions_compound_bitwise_8); // 6 (0b0110)

a_expressions_compound_bitwise_8 ^= 3;
console.log(a_expressions_compound_bitwise_8); // 5 (0b0101)

let c_expressions_compound_bitwise_8 = 1;
c_expressions_compound_bitwise_8 <<= 4;
console.log(c_expressions_compound_bitwise_8); // 16

c_expressions_compound_bitwise_8 >>= 2;
console.log(c_expressions_compound_bitwise_8); // 4

let neg_expressions_compound_bitwise_8 = -16;
neg_expressions_compound_bitwise_8 >>>= 2;
console.log(neg_expressions_compound_bitwise_8); // 1073741820

// Logical assignment operators
let x_expressions_compound_bitwise_8: number = 0;
let y_expressions_compound_bitwise_8: number = 42;

x_expressions_compound_bitwise_8 ||= y_expressions_compound_bitwise_8;
console.log(x_expressions_compound_bitwise_8); // 42

x_expressions_compound_bitwise_8 &&= 100;
console.log(x_expressions_compound_bitwise_8); // 100

let n_expressions_compound_bitwise_8: number | null = null;
n_expressions_compound_bitwise_8 ??= 50;
console.log(n_expressions_compound_bitwise_8); // 50

n_expressions_compound_bitwise_8 ??= 999;
console.log(n_expressions_compound_bitwise_8); // 50

// --- Context Case: language_expressions_do-while ---
// @expect: 1
// @expect: 2
// @expect: 3
let count_expressions_do_while_9: number = 0;
do {
  count_expressions_do_while_9 = count_expressions_do_while_9 + 1;
  console.log(count_expressions_do_while_9);
} while (count_expressions_do_while_9 < 3);

// --- Context Case: language_expressions_exception-unwinding ---
// @expect: ->stepA->stepB->stepC(ok)[finallyC](B_afterC)[finallyB](A_afterB)[finallyA]
// @expect: ->stepA->stepB->stepC[finallyC][catchB:range out of bounds][finallyB](A_afterB)[finallyA]
class UnwindTracker_expressions_exception_unwinding_10 {
  trail: string = "";

  stepC(fail: boolean): void {
    this.trail = this.trail + "->stepC";
    try {
      if (fail) {
        throw "range out of bounds";
      }
      this.trail = this.trail + "(ok)";
    } finally {
      this.trail = this.trail + "[finallyC]";
    }
  }

  stepB(fail: boolean): void {
    this.trail = this.trail + "->stepB";
    try {
      this.stepC(fail);
      this.trail = this.trail + "(B_afterC)";
    } catch (err_expressions_exception_unwinding_10) {
      this.trail = this.trail + "[catchB:" + err_expressions_exception_unwinding_10 + "]";
    } finally {
      this.trail = this.trail + "[finallyB]";
    }
  }

  stepA(fail: boolean): void {
    this.trail = this.trail + "->stepA";
    try {
      this.stepB(fail);
      this.trail = this.trail + "(A_afterB)";
    } finally {
      this.trail = this.trail + "[finallyA]";
    }
  }
}

const tracker1_expressions_exception_unwinding_10 = new UnwindTracker_expressions_exception_unwinding_10();
tracker1_expressions_exception_unwinding_10.stepA(false);
console.log(tracker1_expressions_exception_unwinding_10.trail);

const tracker2_expressions_exception_unwinding_10 = new UnwindTracker_expressions_exception_unwinding_10();
tracker2_expressions_exception_unwinding_10.stepA(true);
console.log(tracker2_expressions_exception_unwinding_10.trail);

// --- Context Case: language_expressions_exponentiation ---
// @expect: 1024
// @expect: 27
// @expect: 1
// @expect: 16
const base_expressions_exponentiation_11: number = 2;
const exp_expressions_exponentiation_11: number = 10;
console.log(base_expressions_exponentiation_11 ** exp_expressions_exponentiation_11);
console.log(3 ** 3);
console.log(2 ** 0);
let x_expressions_exponentiation_11: number = 4;
x_expressions_exponentiation_11 **= 2;
console.log(x_expressions_exponentiation_11);

// --- Context Case: language_expressions_for-loop ---
// @expect: 24
let result_expressions_for_loop_12: number = 1;
for (let i_expressions_for_loop_12: number = 1; i_expressions_for_loop_12 <= 4; i_expressions_for_loop_12 += 1) {
    result_expressions_for_loop_12 = result_expressions_for_loop_12 * i_expressions_for_loop_12;
}
console.log(result_expressions_for_loop_12);

// --- Context Case: language_expressions_for-of ---
// @expect: 10
// @expect: 20
// @expect: 30
const numbers_expressions_for_of_13: number[] = [10, 20, 30];
for (const n of numbers_expressions_for_of_13) {
  console.log(n);
}

// --- Context Case: language_expressions_if-else ---
// @expect: passed
let resultStatus_expressions_if_else_14: string = "pending";
const score_expressions_if_else_14: number = 85;
if (score_expressions_if_else_14 >= 80) {
    resultStatus_expressions_if_else_14 = "passed";
} else {
    resultStatus_expressions_if_else_14 = "failed";
}
console.log(resultStatus_expressions_if_else_14);

// --- Context Case: language_expressions_logical-operators ---
// @expect: true
// @expect: false
// @expect: true
// @expect: false
// @expect: false
// @expect: true
const t_expressions_logical_operators_15: boolean = true;
const f_expressions_logical_operators_15: boolean = false;
console.log(t_expressions_logical_operators_15 && true);
console.log(t_expressions_logical_operators_15 && f_expressions_logical_operators_15);
console.log(f_expressions_logical_operators_15 || t_expressions_logical_operators_15);
console.log(f_expressions_logical_operators_15 || false);
console.log(!t_expressions_logical_operators_15);
console.log(!f_expressions_logical_operators_15);

// --- Context Case: language_expressions_nested-loops-labeled ---
// @expect: skip middle at i=1, j=2
// @expect: break outer at i=2, j=2
// @expect: Total hits: 16
// @expect: Early break: true
// @expect: odd: 1
// @expect: odd: 3
// @expect: odd: 5
let hitCount_expressions_nested_loops_labeled_16 = 0;
let earlyBreak_expressions_nested_loops_labeled_16 = false;

outer: for (let i_expressions_nested_loops_labeled_16 = 0; i_expressions_nested_loops_labeled_16 < 4; i_expressions_nested_loops_labeled_16 = i_expressions_nested_loops_labeled_16 + 1) {
  middle: for (let j_expressions_nested_loops_labeled_16 = 0; j_expressions_nested_loops_labeled_16 < 4; j_expressions_nested_loops_labeled_16 = j_expressions_nested_loops_labeled_16 + 1) {
    if (i_expressions_nested_loops_labeled_16 === 1 && j_expressions_nested_loops_labeled_16 === 2) {
      console.log("skip middle at i=" + i_expressions_nested_loops_labeled_16 + ", j=" + j_expressions_nested_loops_labeled_16);
      continue outer;
    }
    if (i_expressions_nested_loops_labeled_16 === 2 && j_expressions_nested_loops_labeled_16 === 2) {
      console.log("break outer at i=" + i_expressions_nested_loops_labeled_16 + ", j=" + j_expressions_nested_loops_labeled_16);
      earlyBreak_expressions_nested_loops_labeled_16 = true;
      break outer;
    }
    inner: for (let k_expressions_nested_loops_labeled_16 = 0; k_expressions_nested_loops_labeled_16 < 2; k_expressions_nested_loops_labeled_16 = k_expressions_nested_loops_labeled_16 + 1) {
      hitCount_expressions_nested_loops_labeled_16 = hitCount_expressions_nested_loops_labeled_16 + 1;
    }
  }
}

console.log("Total hits: " + hitCount_expressions_nested_loops_labeled_16);
console.log("Early break: " + earlyBreak_expressions_nested_loops_labeled_16);

// Labeled while loop test
let counter_expressions_nested_loops_labeled_16 = 0;
loopA: while (counter_expressions_nested_loops_labeled_16 < 10) {
  counter_expressions_nested_loops_labeled_16 = counter_expressions_nested_loops_labeled_16 + 1;
  if (counter_expressions_nested_loops_labeled_16 % 2 === 0) {
    continue loopA;
  }
  if (counter_expressions_nested_loops_labeled_16 > 6) {
    break loopA;
  }
  console.log("odd: " + counter_expressions_nested_loops_labeled_16);
}

// --- Context Case: language_expressions_nested-ternary ---
// @expect: A
// @expect: B
// @expect: C
// @expect: D
// @expect: F
// @expect: 10
// @expect: 15
// @expect: 20
function classifyScore_expressions_nested_ternary_17(score: number): string {
    return score >= 90 ? "A" :
           score >= 80 ? "B" :
           score >= 70 ? "C" :
           score >= 60 ? "D" : "F";
}

console.log(classifyScore_expressions_nested_ternary_17(95));
console.log(classifyScore_expressions_nested_ternary_17(85));
console.log(classifyScore_expressions_nested_ternary_17(75));
console.log(classifyScore_expressions_nested_ternary_17(65));
console.log(classifyScore_expressions_nested_ternary_17(50));

function clamp_expressions_nested_ternary_17(val: number, min: number, max: number): number {
    return val < min ? min : val > max ? max : val;
}

console.log(clamp_expressions_nested_ternary_17(5, 10, 20));
console.log(clamp_expressions_nested_ternary_17(15, 10, 20));
console.log(clamp_expressions_nested_ternary_17(25, 10, 20));

// --- Context Case: language_expressions_nested-try-catch ---
// @expect: outer try
// @expect: inner try
// @expect: inner success
// @expect: inner finally
// @expect: outer finally
// @expect: outer try
// @expect: inner try
// @expect: caught inner:
// @expect: inner error
// @expect: inner finally
// @expect: caught outer:
// @expect: re-thrown from inner
// @expect: outer finally
function compute_expressions_nested_try_catch_18(x: number) {
  try {
    console.log("outer try");
    try {
      console.log("inner try");
      if (x === 1) {
        throw "inner error";
      }
      console.log("inner success");
    } catch (e1_expressions_nested_try_catch_18) {
      console.log("caught inner:");
      console.log(e1_expressions_nested_try_catch_18);
      throw "re-thrown from inner";
    } finally {
      console.log("inner finally");
    }
  } catch (e2_expressions_nested_try_catch_18) {
    console.log("caught outer:");
    console.log(e2_expressions_nested_try_catch_18);
  } finally {
    console.log("outer finally");
  }
}

function main_expressions_nested_try_catch_18() {
  compute_expressions_nested_try_catch_18(0);
  compute_expressions_nested_try_catch_18(1);
}

main_expressions_nested_try_catch_18();

// --- Context Case: language_expressions_nullable-string ---
// @expect: hello
// @expect: world
let s_expressions_nullable_string_19: string | null = "hello";
console.log(s_expressions_nullable_string_19);
const fallback_expressions_nullable_string_19: string | undefined = undefined;
console.log(fallback_expressions_nullable_string_19 ?? "world");

// --- Context Case: language_expressions_nullish-coalescing ---
// @expect: hello
// @expect: fallback
// @expect: fallback
// @expect: 32
const a_expressions_nullish_coalescing_20: string = "hello";
const b_expressions_nullish_coalescing_20: string = a_expressions_nullish_coalescing_20 ?? "fallback";
console.log(b_expressions_nullish_coalescing_20);

const undef_expressions_nullish_coalescing_20: string | undefined = undefined;
const c_expressions_nullish_coalescing_20: string = undef_expressions_nullish_coalescing_20 ?? "fallback";
console.log(c_expressions_nullish_coalescing_20);

const nul_expressions_nullish_coalescing_20: string | null = null;
const d_expressions_nullish_coalescing_20: string = nul_expressions_nullish_coalescing_20 ?? "fallback";
console.log(d_expressions_nullish_coalescing_20);

let num_expressions_nullish_coalescing_20: number = 256;
num_expressions_nullish_coalescing_20 >>>= 3;
console.log(num_expressions_nullish_coalescing_20);

// --- Context Case: language_expressions_number-literal ---
// @expect: 123.5
console.log(123.5);

// --- Context Case: language_expressions_object-mutation ---
// @expect: 50
class Box_expressions_object_mutation_22 {
  width: number = 10;
  height: number = 20;
}

const b_expressions_object_mutation_22: Box_expressions_object_mutation_22 = new Box_expressions_object_mutation_22();
b_expressions_object_mutation_22.width = 50;
console.log(b_expressions_object_mutation_22.width);

// --- Context Case: language_expressions_postfix-prefix-complex ---
// @expect: 11
// @expect: 11
// @expect: 12
// @expect: 11
// @expect: 19
// @expect: 19
// @expect: 18
// @expect: 19
// @expect: 7
// @expect: 3
// @expect: 22
let x_expressions_postfix_prefix_complex_23 = 10;
let y_expressions_postfix_prefix_complex_23 = ++x_expressions_postfix_prefix_complex_23;
console.log(x_expressions_postfix_prefix_complex_23); // 11
console.log(y_expressions_postfix_prefix_complex_23); // 11

let z_expressions_postfix_prefix_complex_23 = x_expressions_postfix_prefix_complex_23++;
console.log(x_expressions_postfix_prefix_complex_23); // 12
console.log(z_expressions_postfix_prefix_complex_23); // 11

let a_expressions_postfix_prefix_complex_23 = 20;
let b_expressions_postfix_prefix_complex_23 = --a_expressions_postfix_prefix_complex_23;
console.log(a_expressions_postfix_prefix_complex_23); // 19
console.log(b_expressions_postfix_prefix_complex_23); // 19

let c_expressions_postfix_prefix_complex_23 = a_expressions_postfix_prefix_complex_23--;
console.log(a_expressions_postfix_prefix_complex_23); // 18
console.log(c_expressions_postfix_prefix_complex_23); // 19

// In expressions
let m_expressions_postfix_prefix_complex_23 = 5;
let n_expressions_postfix_prefix_complex_23 = 2;
let res_expressions_postfix_prefix_complex_23 = (m_expressions_postfix_prefix_complex_23++) * (++n_expressions_postfix_prefix_complex_23) + (++m_expressions_postfix_prefix_complex_23);
console.log(m_expressions_postfix_prefix_complex_23); // 7
console.log(n_expressions_postfix_prefix_complex_23); // 3
console.log(res_expressions_postfix_prefix_complex_23); // 5 * 3 + 7 = 22

// --- Context Case: language_expressions_string-concat ---
// @expect: hello, world
// @expect: count: 42
// @expect: 100 ms
// @expect: isReady: true
// @expect: false is false
// @expect: total: 50
const greeting_expressions_string_concat_24: string = 'hello' + ', ' + 'world';
console.log(greeting_expressions_string_concat_24);

const numConcat1_expressions_string_concat_24: string = "count: " + 42;
const numConcat2_expressions_string_concat_24: string = 100 + " ms";
console.log(numConcat1_expressions_string_concat_24);
console.log(numConcat2_expressions_string_concat_24);

const boolConcat1_expressions_string_concat_24: string = "isReady: " + true;
const boolConcat2_expressions_string_concat_24: string = false + " is false";
console.log(boolConcat1_expressions_string_concat_24);
console.log(boolConcat2_expressions_string_concat_24);

let compoundStr_expressions_string_concat_24: string = "total: ";
compoundStr_expressions_string_concat_24 += 50;
console.log(compoundStr_expressions_string_concat_24);

// --- Context Case: language_expressions_string-equality ---
// @expect: true
// @expect: true
// @expect: true
// @expect: true
const s1_expressions_string_equality_25: string = "typescript";
const s2_expressions_string_equality_25: string = "typescript";
const s3_expressions_string_equality_25: string = "golang";
console.log(s1_expressions_string_equality_25 === s2_expressions_string_equality_25);
console.log(s1_expressions_string_equality_25 !== s3_expressions_string_equality_25);
console.log(s1_expressions_string_equality_25 == s2_expressions_string_equality_25);
console.log(s1_expressions_string_equality_25 != s3_expressions_string_equality_25);

// --- Context Case: language_expressions_switch-case ---
// @expect: red
const fruit_expressions_switch_case_26: string = "apple";
switch (fruit_expressions_switch_case_26) {
  case "banana":
    console.log("yellow");
    break;
  case "apple":
    console.log("red");
    break;
  default:
    console.log("unknown");
    break;
}

// --- Context Case: language_expressions_switch-fallthrough ---
// @expect: prime
// @expect: prime
// @expect: composite
// @expect: prime
// @expect: composite
// @expect: other
function classify_expressions_switch_fallthrough_27(digit: number): string {
  switch (digit) {
    case 2:
    case 3:
    case 5:
    case 7:
      return "prime";
    case 4:
    case 6:
    case 8:
    case 9:
      return "composite";
    default:
      return "other";
  }
}

console.log(classify_expressions_switch_fallthrough_27(2));
console.log(classify_expressions_switch_fallthrough_27(3));
console.log(classify_expressions_switch_fallthrough_27(4));
console.log(classify_expressions_switch_fallthrough_27(7));
console.log(classify_expressions_switch_fallthrough_27(9));
console.log(classify_expressions_switch_fallthrough_27(1));

// --- Context Case: language_expressions_template-strings ---
// @expect: User Alice has score 100, winner: true
const user_expressions_template_strings_28: string = "Alice";
const score_expressions_template_strings_28: number = 100;
const isWinner_expressions_template_strings_28: boolean = true;
console.log(`User ${user_expressions_template_strings_28} has score ${score_expressions_template_strings_28}, winner: ${isWinner_expressions_template_strings_28}`);

// --- Context Case: language_expressions_ternary-conditional ---
// @expect: true
// @expect: passed
const score_expressions_ternary_conditional_29: number = 85;
const isPass_expressions_ternary_conditional_29: boolean = score_expressions_ternary_conditional_29 >= 50 ? true : false;
const message_expressions_ternary_conditional_29: string = isPass_expressions_ternary_conditional_29 ? "passed" : "failed";
console.log(isPass_expressions_ternary_conditional_29);
console.log(message_expressions_ternary_conditional_29);

// --- Context Case: language_expressions_try-catch ---
// @expect: no throw
// @expect: caught:
// @expect: custom error message
// @expect: recovered
function mayThrow_expressions_try_catch_30(flag: number) {
  if (flag === 1) {
    throw "custom error message";
  }
  console.log("no throw");
}

function main_expressions_try_catch_30() {
  try {
    mayThrow_expressions_try_catch_30(0);
    mayThrow_expressions_try_catch_30(1);
    console.log("unreachable");
  } catch (e_expressions_try_catch_30) {
    console.log("caught:");
    console.log(e_expressions_try_catch_30);
  }
  console.log("recovered");
}

main_expressions_try_catch_30();

// --- Context Case: language_expressions_try-catch-finally ---
// @expect: start try
// @expect: end try
// @expect: in finally
// @expect: start try
// @expect: in catch:
// @expect: boom
// @expect: in finally
function runTest_expressions_try_catch_finally_31(shouldThrow: number) {
  try {
    console.log("start try");
    if (shouldThrow === 1) {
      throw "boom";
    }
    console.log("end try");
  } catch (err_expressions_try_catch_finally_31) {
    console.log("in catch:");
    console.log(err_expressions_try_catch_finally_31);
  } finally {
    console.log("in finally");
  }
}

function main_expressions_try_catch_finally_31() {
  runTest_expressions_try_catch_finally_31(0);
  runTest_expressions_try_catch_finally_31(1);
}

main_expressions_try_catch_finally_31();

// --- Context Case: language_expressions_try-catch-native ---
// @expect: boom
// @expect: finally
try {
  throw "boom";
} catch (error_expressions_try_catch_native_32) {
  console.log(error_expressions_try_catch_native_32);
} finally {
  console.log("finally");
}

// --- Context Case: language_expressions_typeof ---
// @expect: number
// @expect: string
// @expect: boolean
// @expect: object
// @expect: number
// @expect: undefined
// @expect: object
const num_expressions_typeof_33: number = 42;
const str_expressions_typeof_33: string = "hello";
const flag_expressions_typeof_33: boolean = true;
const arr_expressions_typeof_33: number[] = [1, 2, 3];

console.log(typeof num_expressions_typeof_33);
console.log(typeof str_expressions_typeof_33);
console.log(typeof flag_expressions_typeof_33);
console.log(typeof arr_expressions_typeof_33);
console.log(typeof (1 + 2));
console.log(typeof undefined);
console.log(typeof null);

// --- Context Case: language_expressions_typeof-exhaustive ---
// @expect: number
// @expect: string
// @expect: boolean
// @expect: bigint
// @expect: symbol
// @expect: function
// @expect: object
// @expect: object
// @expect: number
// @expect: string
// @expect: boolean
// @expect: bigint
// @expect: symbol
const num_expressions_typeof_exhaustive_34 = 42;
const str_expressions_typeof_exhaustive_34 = "hello";
const boolVal_expressions_typeof_exhaustive_34 = true;
const bigIntVal_expressions_typeof_exhaustive_34 = 100n;
const sym_expressions_typeof_exhaustive_34 = Symbol("test");
const fn_expressions_typeof_exhaustive_34 = (x: number): number => x * 2;
const arr_expressions_typeof_exhaustive_34 = [1, 2, 3];
const obj_expressions_typeof_exhaustive_34 = { a: 1, b: "two" };

console.log(typeof num_expressions_typeof_exhaustive_34);
console.log(typeof str_expressions_typeof_exhaustive_34);
console.log(typeof boolVal_expressions_typeof_exhaustive_34);
console.log(typeof bigIntVal_expressions_typeof_exhaustive_34);
console.log(typeof sym_expressions_typeof_exhaustive_34);
console.log(typeof fn_expressions_typeof_exhaustive_34);
console.log(typeof arr_expressions_typeof_exhaustive_34);
console.log(typeof obj_expressions_typeof_exhaustive_34);

// Dynamic check on unknown/any
function checkType_expressions_typeof_exhaustive_34(val: unknown): string {
  return typeof val;
}

console.log(checkType_expressions_typeof_exhaustive_34(123));
console.log(checkType_expressions_typeof_exhaustive_34("world"));
console.log(checkType_expressions_typeof_exhaustive_34(false));
console.log(checkType_expressions_typeof_exhaustive_34(999n));
console.log(checkType_expressions_typeof_exhaustive_34(Symbol("dyn")));

// --- Context Case: language_expressions_typeof-narrowing ---
// @expect: 50
// @expect: hello world
const a_expressions_typeof_narrowing_35: string | number = 42;
if (typeof a_expressions_typeof_narrowing_35 === "number") {
    console.log(a_expressions_typeof_narrowing_35 + 8);
}
const b_expressions_typeof_narrowing_35: string | number = "hello";
if (typeof b_expressions_typeof_narrowing_35 === "string") {
    console.log(b_expressions_typeof_narrowing_35 + " world");
}

// --- Context Case: language_expressions_unary-minus ---
// @expect: -15
// @expect: 15
const x_expressions_unary_minus_36: number = 15;
const neg_expressions_unary_minus_36: number = -x_expressions_unary_minus_36;
const back_expressions_unary_minus_36: number = -neg_expressions_unary_minus_36;
console.log(neg_expressions_unary_minus_36);
console.log(back_expressions_unary_minus_36);

// --- Context Case: language_expressions_unknown-boxing ---
// @expect: 42
// @expect: hello world
// @expect: true
const uNum_expressions_unknown_boxing_37: unknown = 42;
const uStr_expressions_unknown_boxing_37: unknown = "hello world";
const uBool_expressions_unknown_boxing_37: unknown = true;

console.log(uNum_expressions_unknown_boxing_37);
console.log(uStr_expressions_unknown_boxing_37);
console.log(uBool_expressions_unknown_boxing_37);

// --- Context Case: language_expressions_unknown-checked-cast ---
// @expect: 150
// @expect: scriptgo
// @expect: false
const u1_expressions_unknown_checked_cast_38: unknown = 100;
const u2_expressions_unknown_checked_cast_38: unknown = "scriptgo";
const u3_expressions_unknown_checked_cast_38: unknown = false;

const n_expressions_unknown_checked_cast_38: number = u1_expressions_unknown_checked_cast_38 as number;
const s_expressions_unknown_checked_cast_38: string = u2_expressions_unknown_checked_cast_38 as string;
const b_expressions_unknown_checked_cast_38: boolean = u3_expressions_unknown_checked_cast_38 as boolean;

console.log(n_expressions_unknown_checked_cast_38 + 50);
console.log(s_expressions_unknown_checked_cast_38);
console.log(b_expressions_unknown_checked_cast_38);

// --- Context Case: language_expressions_unknown-functions ---
// @expect: 42
// @expect: 52
function wrap_expressions_unknown_functions_39(x: number): unknown {
    return x * 2;
}

function processUnknown_expressions_unknown_functions_39(u: unknown): number {
    return (u as number) + 10;
}

const res_expressions_unknown_functions_39: unknown = wrap_expressions_unknown_functions_39(21);
console.log(res_expressions_unknown_functions_39);
const finalNum_expressions_unknown_functions_39: number = processUnknown_expressions_unknown_functions_39(res_expressions_unknown_functions_39);
console.log(finalNum_expressions_unknown_functions_39);

// --- Context Case: language_expressions_unknown-typeof ---
// @expect: number
// @expect: string
// @expect: boolean
const val1_expressions_unknown_typeof_40: unknown = 123;
const val2_expressions_unknown_typeof_40: unknown = "typescript";
const val3_expressions_unknown_typeof_40: unknown = true;

console.log(typeof val1_expressions_unknown_typeof_40);
console.log(typeof val2_expressions_unknown_typeof_40);
console.log(typeof val3_expressions_unknown_typeof_40);

// --- Context Case: language_expressions_variable-mutation ---
// @expect: 12
let count_expressions_variable_mutation_41: number = 0;
count_expressions_variable_mutation_41 = count_expressions_variable_mutation_41 + 10;
count_expressions_variable_mutation_41 += 5;
count_expressions_variable_mutation_41 -= 3;
console.log(count_expressions_variable_mutation_41);

// --- Context Case: language_expressions_while-loop ---
// @expect: 15
let sum_expressions_while_loop_42: number = 0;
let n_expressions_while_loop_42: number = 1;
while (n_expressions_while_loop_42 <= 5) {
    sum_expressions_while_loop_42 += n_expressions_while_loop_42;
    n_expressions_while_loop_42 += 1;
}
console.log(sum_expressions_while_loop_42);
