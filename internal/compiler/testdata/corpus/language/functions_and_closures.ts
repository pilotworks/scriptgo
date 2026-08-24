// ScriptGo Corpus: Language: Functions & Closures
// Consolidated test suite with inline assertions.

// --- Context Case: language_closures_array-methods ---
// @expect: 2, 4, 6, 8, 10
// @expect: 2, 4
// @expect: 4
// @expect: 5
// @expect: 15
// @expect: 3
// @expect: true
// @expect: true
const nums_closures_array_methods_0 = [1, 2, 3, 4, 5];
const factor_closures_array_methods_0 = 2;
const doubled_closures_array_methods_0 = nums_closures_array_methods_0.map((x) => x * factor_closures_array_methods_0);
console.log(doubled_closures_array_methods_0.join(", "));

const evens_closures_array_methods_0 = nums_closures_array_methods_0.filter((x) => x % 2 === 0);
console.log(evens_closures_array_methods_0.join(", "));

nums_closures_array_methods_0.forEach((x) => {
    if (x > 3) {
        console.log(x);
    }
});

const sum_closures_array_methods_0 = nums_closures_array_methods_0.reduce((acc, x) => acc + x, 0);
console.log(sum_closures_array_methods_0);

const found_closures_array_methods_0 = nums_closures_array_methods_0.find((x) => x === 3);
console.log(found_closures_array_methods_0);

console.log(nums_closures_array_methods_0.some((x) => x > 4));
console.log(nums_closures_array_methods_0.every((x) => x > 0));

// --- Context Case: language_closures_basic ---
// @expect: 42
function apply_closures_basic_1(fn: (x: number) => number, val: number): number {
    return fn(val);
}

const double_closures_basic_1 = (x: number): number => x * 2;
console.log(apply_closures_basic_1(double_closures_basic_1, 21));

// --- Context Case: language_closures_capture-variable ---
// @expect: 30
// @expect: Hello, ScriptGo!
const multiplier_closures_capture_variable_2 = 3;
const mult_closures_capture_variable_2 = (x: number): number => x * multiplier_closures_capture_variable_2;
console.log(mult_closures_capture_variable_2(10));

const greeting_closures_capture_variable_2 = "Hello";
const greet_closures_capture_variable_2 = (name: string): string => `${greeting_closures_capture_variable_2}, ${name}!`;
console.log(greet_closures_capture_variable_2("ScriptGo"));

// --- Context Case: language_closures_currying-compose ---
// @expect: 15
// @expect: 20
// @expect: 30
// @expect: 25
// @expect: 256
function add_closures_currying_compose_3(a: number): (b: number) => number {
  return (b: number): number => a + b;
}

function multiply_closures_currying_compose_3(a: number): (b: number) => number {
  return (b: number): number => a * b;
}

function compose_closures_currying_compose_3(f: (x: number) => number, g: (x: number) => number): (x: number) => number {
  return (x: number): number => f(g(x));
}

const add5_closures_currying_compose_3 = add_closures_currying_compose_3(5);
const double_closures_currying_compose_3 = multiply_closures_currying_compose_3(2);
const add5ThenDouble_closures_currying_compose_3 = compose_closures_currying_compose_3(double_closures_currying_compose_3, add5_closures_currying_compose_3);
const doubleThenAdd5_closures_currying_compose_3 = compose_closures_currying_compose_3(add5_closures_currying_compose_3, double_closures_currying_compose_3);

console.log(add5_closures_currying_compose_3(10));
console.log(double_closures_currying_compose_3(10));
console.log(add5ThenDouble_closures_currying_compose_3(10));
console.log(doubleThenAdd5_closures_currying_compose_3(10));

// 3-stage pipeline
const square_closures_currying_compose_3 = (x: number): number => x * x;
const pipeline_closures_currying_compose_3 = compose_closures_currying_compose_3(square_closures_currying_compose_3, add5ThenDouble_closures_currying_compose_3);
console.log(pipeline_closures_currying_compose_3(3));

// --- Context Case: language_closures_higher-order-compose ---
// @expect: 15
// @expect: 25
// @expect: 25
function makeAdder_closures_higher_order_compose_4(x: number): (y: number) => number {
    return (y: number): number => x + y;
}

function applyTwice_closures_higher_order_compose_4(f: (n: number) => number, val: number): number {
    return f(f(val));
}

const add5_closures_higher_order_compose_4 = makeAdder_closures_higher_order_compose_4(5);
console.log(add5_closures_higher_order_compose_4(10));
console.log(add5_closures_higher_order_compose_4(20));

const add10_closures_higher_order_compose_4 = makeAdder_closures_higher_order_compose_4(10);
console.log(applyTwice_closures_higher_order_compose_4(add10_closures_higher_order_compose_4, 5));

// --- Context Case: language_closures_multi-capture ---
// @expect: USD $100 total
// @expect: USD $250.5 total
// @expect: EUR 90 net
// @expect: JPY 1500 YEN
function createFormatter_closures_multi_capture_5(prefix: string, suffix: string, scale: number, uppercase: boolean): (val: number) => string {
  return (val: number): string => {
    const scaled_closures_multi_capture_5 = val * scale;
    let res_closures_multi_capture_5 = prefix + scaled_closures_multi_capture_5 + suffix;
    if (uppercase) {
      res_closures_multi_capture_5 = res_closures_multi_capture_5.toUpperCase();
    }
    return res_closures_multi_capture_5;
  };
}

const fmtUSD_closures_multi_capture_5 = createFormatter_closures_multi_capture_5("USD $", " total", 1, false);
const fmtEUR_closures_multi_capture_5 = createFormatter_closures_multi_capture_5("EUR ", " net", 0.9, false);
const fmtJPY_closures_multi_capture_5 = createFormatter_closures_multi_capture_5("jpy ", " yen", 150, true);

console.log(fmtUSD_closures_multi_capture_5(100));
console.log(fmtUSD_closures_multi_capture_5(250.5));
console.log(fmtEUR_closures_multi_capture_5(100));
console.log(fmtJPY_closures_multi_capture_5(10));

// --- Context Case: language_functions_arrow-functions ---
// @expect: 30
// @expect: 42
// @expect: Hello, TypeScript!
const add_functions_arrow_functions_6 = (a: number, b_functions_arrow_functions_6: number): number => a + b_functions_arrow_functions_6;
const multiply_functions_arrow_functions_6 = (a: number, b_functions_arrow_functions_6: number): number => {
    return a * b_functions_arrow_functions_6;
};
const greet_functions_arrow_functions_6 = (name: string): string => `Hello, ${name}!`;

console.log(add_functions_arrow_functions_6(10, 20));
console.log(multiply_functions_arrow_functions_6(6, 7));
console.log(greet_functions_arrow_functions_6("TypeScript"));

// --- Context Case: language_functions_default-and-rest ---
// @expect: [WARN] Server starting on port 8080
// @expect: [ERROR] Database connection lost
// @expect: 15
// @expect: 110
function buildLog_functions_default_and_rest_7(level: string = "INFO", ...messages: string[]): string {
  let joined_functions_default_and_rest_7: string = "";
  for (let i_functions_default_and_rest_7 = 0; i_functions_default_and_rest_7 < messages.length; i_functions_default_and_rest_7 = i_functions_default_and_rest_7 + 1) {
    if (i_functions_default_and_rest_7 > 0) {
      joined_functions_default_and_rest_7 = joined_functions_default_and_rest_7 + " ";
    }
    joined_functions_default_and_rest_7 = joined_functions_default_and_rest_7 + messages[i_functions_default_and_rest_7];
  }
  return "[" + level + "] " + joined_functions_default_and_rest_7;
}

function sumWithBase_functions_default_and_rest_7(base: number = 100, ...nums: number[]): number {
  let total_functions_default_and_rest_7: number = base;
  for (let i_functions_default_and_rest_7 = 0; i_functions_default_and_rest_7 < nums.length; i_functions_default_and_rest_7 = i_functions_default_and_rest_7 + 1) {
    total_functions_default_and_rest_7 = total_functions_default_and_rest_7 + nums[i_functions_default_and_rest_7];
  }
  return total_functions_default_and_rest_7;
}

console.log(buildLog_functions_default_and_rest_7("WARN", "Server", "starting", "on", "port", "8080"));
console.log(buildLog_functions_default_and_rest_7("ERROR", "Database", "connection", "lost"));
console.log(sumWithBase_functions_default_and_rest_7(0, 1, 2, 3, 4, 5));
console.log(sumWithBase_functions_default_and_rest_7(50, 10, 20, 30));

// --- Context Case: language_functions_function-expressions ---
// @expect: 42
// @expect: Hello, ScriptGo!
const multiply_functions_function_expressions_8 = function(a: number, b: number): number {
    return a * b;
};

const greet_functions_function_expressions_8 = function(name: string): string {
    return "Hello, " + name + "!";
};

console.log(multiply_functions_function_expressions_8(6, 7));
console.log(greet_functions_function_expressions_8("ScriptGo"));

// --- Context Case: language_functions_higher-order-pipeline ---
// @expect: 10
// @expect: 15
// @expect: 20
// @expect: 30
// @expect: 20
// @expect: 30
function createMultiplier_functions_higher_order_pipeline_9(factor: number): (x: number) => number {
    return (x: number): number => x * factor;
}

function applyTransform_functions_higher_order_pipeline_9(val: number, fn: (n: number) => number): number {
    return fn(val);
}

const double_functions_higher_order_pipeline_9 = createMultiplier_functions_higher_order_pipeline_9(2);
const triple_functions_higher_order_pipeline_9 = createMultiplier_functions_higher_order_pipeline_9(3);

console.log(double_functions_higher_order_pipeline_9(5));
console.log(triple_functions_higher_order_pipeline_9(5));
console.log(applyTransform_functions_higher_order_pipeline_9(10, double_functions_higher_order_pipeline_9));
console.log(applyTransform_functions_higher_order_pipeline_9(10, triple_functions_higher_order_pipeline_9));

function compose_functions_higher_order_pipeline_9(f: (n: number) => number, g: (n: number) => number): (n: number) => number {
    return (x: number): number => f(g(x));
}

const addTen_functions_higher_order_pipeline_9 = (n: number): number => n + 10;
const doubleThenAddTen_functions_higher_order_pipeline_9 = compose_functions_higher_order_pipeline_9(addTen_functions_higher_order_pipeline_9, double_functions_higher_order_pipeline_9);
const addTenThenDouble_functions_higher_order_pipeline_9 = compose_functions_higher_order_pipeline_9(double_functions_higher_order_pipeline_9, addTen_functions_higher_order_pipeline_9);

console.log(doubleThenAddTen_functions_higher_order_pipeline_9(5));
console.log(addTenThenDouble_functions_higher_order_pipeline_9(5));

// --- Context Case: language_functions_mutual-recursion ---
// @expect: true
// @expect: true
// @expect: false
// @expect: false
// @expect: true
// @expect: false
// @expect: true
// @expect: true
function isEven_functions_mutual_recursion_10(n: number): boolean {
    if (n === 0) {
        return true;
    }
    if (n < 0) {
        return isEven_functions_mutual_recursion_10(-n);
    }
    return isOdd_functions_mutual_recursion_10(n - 1);
}

function isOdd_functions_mutual_recursion_10(n: number): boolean {
    if (n === 0) {
        return false;
    }
    if (n < 0) {
        return isOdd_functions_mutual_recursion_10(-n);
    }
    return isEven_functions_mutual_recursion_10(n - 1);
}

console.log(isEven_functions_mutual_recursion_10(0));
console.log(isEven_functions_mutual_recursion_10(4));
console.log(isEven_functions_mutual_recursion_10(7));
console.log(isOdd_functions_mutual_recursion_10(0));
console.log(isOdd_functions_mutual_recursion_10(3));
console.log(isOdd_functions_mutual_recursion_10(8));
console.log(isEven_functions_mutual_recursion_10(-6));
console.log(isOdd_functions_mutual_recursion_10(-5));

// --- Context Case: language_functions_optional-param ---
// @expect: hello alice
// @expect: hello stranger
function greet_functions_optional_param_11(name: string | null): void {
    if (name != null) {
        console.log("hello " + name);
    } else {
        console.log("hello stranger");
    }
}
greet_functions_optional_param_11("alice");
greet_functions_optional_param_11(null);

// --- Context Case: language_functions_recursion-factorial-fibonacci ---
// @expect: 1
// @expect: 120
// @expect: 5040
// @expect: 0
// @expect: 1
// @expect: 8
// @expect: 55
function factorial_functions_recursion_factorial_fibonacci_12(n: number): number {
    if (n <= 1) {
        return 1;
    }
    return n * factorial_functions_recursion_factorial_fibonacci_12(n - 1);
}

function fibonacci_functions_recursion_factorial_fibonacci_12(n: number): number {
    if (n <= 0) {
        return 0;
    }
    if (n === 1) {
        return 1;
    }
    return fibonacci_functions_recursion_factorial_fibonacci_12(n - 1) + fibonacci_functions_recursion_factorial_fibonacci_12(n - 2);
}

console.log(factorial_functions_recursion_factorial_fibonacci_12(1));
console.log(factorial_functions_recursion_factorial_fibonacci_12(5));
console.log(factorial_functions_recursion_factorial_fibonacci_12(7));

console.log(fibonacci_functions_recursion_factorial_fibonacci_12(0));
console.log(fibonacci_functions_recursion_factorial_fibonacci_12(1));
console.log(fibonacci_functions_recursion_factorial_fibonacci_12(6));
console.log(fibonacci_functions_recursion_factorial_fibonacci_12(10));

// --- Context Case: language_functions_rest-parameters ---
// @expect: start:a:b:c
function concatAll_functions_rest_parameters_13(prefix: string, ...items: string[]): string {
    let result_functions_rest_parameters_13: string = prefix;
    for (let i_functions_rest_parameters_13: number = 0; i_functions_rest_parameters_13 < items.length; i_functions_rest_parameters_13 += 1) {
        result_functions_rest_parameters_13 += ":" + items[i_functions_rest_parameters_13];
    }
    return result_functions_rest_parameters_13;
}

console.log(concatAll_functions_rest_parameters_13("start", "a", "b", "c"));

// --- Context Case: language_functions_return-value ---
// @expect: 42
function answer_functions_return_value_14(): number {
  return 42;
}
console.log(answer_functions_return_value_14());

// --- Context Case: language_functions_single-argument ---
// @expect: 42
function double_functions_single_argument_15(value: number): number {
  return value * 2;
}
console.log(double_functions_single_argument_15(21));

// --- Context Case: language_functions_two-arguments ---
// @expect: 42
function combine_functions_two_arguments_16(left: number, right: number): number {
  return left + right;
}
console.log(combine_functions_two_arguments_16(19, 23));

// --- Context Case: language_closures_mutating_shared_state ---
// @expect: 10
// @expect: 11
// @expect: 16
// @expect: 16
// @expect: 100
// @expect: 101
// @expect: 16
interface MutatingCounter {
    inc: () => number;
    add: (amount: number) => number;
    get: () => number;
}

function createMutatingCounter(initial: number): MutatingCounter {
    let count = initial;
    return {
        inc: (): number => {
            count = count + 1;
            return count;
        },
        add: (amount: number): number => {
            count = count + amount;
            return count;
        },
        get: (): number => count
    };
}

const counter1 = createMutatingCounter(10);
console.log(counter1.get());
console.log(counter1.inc());
console.log(counter1.add(5));
console.log(counter1.get());

const counter2 = createMutatingCounter(100);
console.log(counter2.get());
console.log(counter2.inc());
console.log(counter1.get());

