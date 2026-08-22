// ScriptGo Corpus: Language: Types, Generics & Primitives
// Consolidated test suite with inline assertions.

// @expect: 3
// @expect: 10n
// @expect: 20n
// @expect: 30n
// @expect: 99n
// @expect: 4
// @expect: 4
// @expect: 40n
// @expect: 3
// @expect: 10,99,30
// @expect: true
// @expect: false
// @expect: true
// @expect: true
// @expect: 16
// @expect: 20
// @expect: 4
// @expect: 4
// @expect: 40
// @expect: 3
// @expect: 1
// @expect: -1
// @expect: true
// @expect: false
// @expect: 2
// @expect: 20
// @expect: 30
// @expect: 2,4,6,8,10
// @expect: 4,8,12,16,20
// @expect: 8,12,16
// @expect: 8,12,16,100,200
// @expect: 200,100,16,12,8
// @expect: 10
// @expect: 40
// @expect: 10-20-30-40
// @expect: 10
// @expect: 3
// @expect: 4
// @expect: 5
// @expect: 40,30,20,5
// @expect: 40 30 20 5 100 200
// @expect: c
// @expect: b:c
// @expect: a:d
// @expect: 42
// @expect: 20
// @expect: a
// @expect: true
// @expect: false
// @expect: true
// @expect: false
// @expect: 25
// @expect: 30
// @expect: 4
// @expect: apple | banana | cherry | date
// @expect: apple,banana,cherry,date,elderberry
// @expect: elderberry
// @expect: apple
// @expect: avocado,banana,cherry,date
// @expect: avocado;banana;cherry
// @expect: AVOCADO BANANA CHERRY DATE
// @expect: true
// @expect: false
// @expect: 2
// @expect: -1
// @expect: 100n
// @expect: 20n
// @expect: 120n
// @expect: 80n
// @expect: 2000n
// @expect: 5n
// @expect: 0n
// @expect: 4n
// @expect: 116n
// @expect: 112n
// @expect: 80n
// @expect: 10n
// @expect: true
// @expect: false
// @expect: true
// @expect: false
// @expect: -20n
// @expect: -21n
// @expect: 42n
// @expect: 999n
// @expect: 100
// @expect: 100n
// @expect: 101n
// @expect: 102n
// @expect: 3
// @expect: 20
// @expect: alpha
// @expect: gamma
// @expect: 3
// @expect: 1
// @expect: 100
// @expect: hello
// @expect: 123
// @expect: scriptgo box
// @expect: 456
// @expect: updated box
// @expect: 42
// @expect: hello generics
// @expect: 99
// @expect: 123
// @expect: 456
// @expect: 555
// @expect: 100
// @expect: third
// @expect: Alice
// @expect: Widget
// @expect: true
// @expect: false
// @expect: 42
// @expect: ScriptGo
// @expect: boxed
// @expect: hello
// @expect: 42
// @expect: 100
// @expect: true
// @expect: 10
// @expect: 20
// @expect: 5
// @expect: 6
// @expect: 15
// @expect: 120
// @expect: start:abcd
// @expect: 4
// @expect: true
// @expect: 3
// @expect: 30
// @expect: 30
// @expect: 2
// @expect: false
// @expect: beta
// @expect: beta
// @expect: alpha
// @expect: undefined
// @expect: true
// @expect: 777
// @expect: interface test
// @expect: count
// @expect: 42
// @expect: 3
// @expect: hello
// @expect: 30
// @expect: Alice
// @expect: 30
// @expect: true
// @expect: 15
// @expect: 35
// @expect: 0
// @expect: 2
// @expect: 10
// @expect: 20
// @expect: hello world
// @expect: 42
// @expect: scriptgo
// @expect: 123 - apple
// @expect: ^[0-9]+$
// @expect:
// @expect: true
// @expect: false
// @expect: true
// @expect: false
// @expect: 10
// @expect: -1
// @expect: The quick brown dog
// @expect: aXbXcX
// @expect: true
// @expect: 2
// @expect: 12345
// @expect: 12345
// @expect: 2
// @expect: 999
// @expect: 999
// @expect: false
// @expect: true
// @expect: true
// @expect: app.key
// @expect: undefined
// @expect: foo
// @expect: Symbol(foo)
// @expect: undefined
// @expect: Symbol()
// @expect: true
// @expect: Symbol(Symbol.iterator)
// @expect: symbol
// @expect: true
// @expect: false
// @expect: app.version
// @expect: app.author
// @expect: false
// @expect: Symbol(app.version)
// @expect: Symbol(unique)
// @expect: 1
// @expect: 3
// @expect: 127
// @expect: -128
// @expect: -56
// @expect: 1
// @expect: 0
// @expect: 120
// @expect: 121
// @expect: 255
// @expect: 2
// @expect: 4
// @expect: 32767
// @expect: -32768
// @expect: 2
// @expect: 4
// @expect: 65535
// @expect: 100
// @expect: 4
// @expect: 8
// @expect: 4294967295
// @expect: 500
// @expect: 4
// @expect: 8
// @expect: 1.5
// @expect: -2.5
// @expect: 8
// @expect: 16
// @expect: 9007199254740993n
// @expect: -9007199254740993n
// @expect: 8
// @expect: 16
// @expect: 1844674407370955161n
// @expect: 42n
// @expect: true
// @expect: true
// @expect: true
// @expect: false
// @expect: 16
// @expect: false
// @expect: 8
// @expect: 8
// @expect: 4
// @expect: true
// @expect: 99
// @expect: 16
// @expect: 0
// @expect: -12
// @expect: 250
// @expect: 4660
// @expect: 13330
// @expect: 22136
// @expect: 30806
// @expect: -100000
// @expect: 1.5
// @expect: 3.141592653589793
// @expect: 9007199254740993n
// @expect: true
// @expect: 3
// @expect: 24
// @expect: 8
// @expect: 3.14159
// @expect: 2.71828
// @expect: 42.5
// @expect: 3
// @expect: 12
// @expect: 4
// @expect: 100000
// @expect: -50000
// @expect: 42
// @expect: 0
// @expect: 7
// @expect: 7
// @expect: 7
// @expect: 0
// @expect: 3
// @expect: 1
// @expect: 7
// @expect: 42
// @expect: 3
// @expect: 42
// @expect: 42
// @expect: 42
// @expect: 7
// @expect: 4
// @expect: 4
// @expect: 0
// @expect: 1
// @expect: 10
// @expect: 255
// @expect: 44
// @expect: 42
// @expect: 25
// @expect: 24
// @expect: 314
// @expect: ascending
// @expect: descending
// @expect: OK
// @expect: Not Found
// @expect: Server Error
// @expect: Circle(r=5)
// @expect: 78
// @expect: Rectangle(w=10, h=4)
// @expect: 40
// @expect: Square(s=6)
// @expect: 36

// --- Context Case: language_arrays_bigint-array ---
const bigs_arrays_bigint_array_0: bigint[] = [10n, 20n, 30n];
console.log(bigs_arrays_bigint_array_0.length);
console.log(bigs_arrays_bigint_array_0[0]);
console.log(bigs_arrays_bigint_array_0[1]);
console.log(bigs_arrays_bigint_array_0[2]);

bigs_arrays_bigint_array_0[1] = 99n;
console.log(bigs_arrays_bigint_array_0[1]);

console.log(bigs_arrays_bigint_array_0.push(40n));
console.log(bigs_arrays_bigint_array_0.length);
console.log(bigs_arrays_bigint_array_0.pop()!);
console.log(bigs_arrays_bigint_array_0.length);
console.log(bigs_arrays_bigint_array_0.join(","));

// --- Context Case: language_arrays_bool-array-index ---
const flags_arrays_bool_array_index_1: boolean[] = [true, false, true];
console.log(flags_arrays_bool_array_index_1[0]);
console.log(flags_arrays_bool_array_index_1[1]);
console.log(flags_arrays_bool_array_index_1[2]);

flags_arrays_bool_array_index_1[1] = true;
console.log(flags_arrays_bool_array_index_1[1]);

// --- Context Case: language_arrays_index-expression ---
const values_arrays_index_expression_2: number[] = [4, 8, 15, 16, 23, 42];
const position_arrays_index_expression_2: number = 2 + 1;
console.log(values_arrays_index_expression_2[position_arrays_index_expression_2]);

// --- Context Case: language_arrays_index-number ---
const values_arrays_index_number_3: number[] = [10, 20];
console.log(values_arrays_index_number_3[1]);

// --- Context Case: language_arrays_methods ---
const nums_arrays_methods_4: number[] = [10, 20, 30];
console.log(nums_arrays_methods_4.push(40));
console.log(nums_arrays_methods_4.length);
console.log(nums_arrays_methods_4.pop());
console.log(nums_arrays_methods_4.length);
console.log(nums_arrays_methods_4.indexOf(20));
console.log(nums_arrays_methods_4.indexOf(99));
console.log(nums_arrays_methods_4.includes(30));
console.log(nums_arrays_methods_4.includes(99));

const sub_arrays_methods_4: number[] = nums_arrays_methods_4.slice(1, 3);
console.log(sub_arrays_methods_4.length);
console.log(sub_arrays_methods_4[0]);
console.log(sub_arrays_methods_4[1]);

// --- Context Case: language_arrays_methods-chaining ---
const nums_arrays_methods_chaining_5: number[] = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];

const evens_arrays_methods_chaining_5: number[] = nums_arrays_methods_chaining_5.filter((n: number): boolean => n % 2 === 0);
console.log(evens_arrays_methods_chaining_5.join(","));

const doubled_arrays_methods_chaining_5: number[] = evens_arrays_methods_chaining_5.map((n: number): number => n * 2);
console.log(doubled_arrays_methods_chaining_5.join(","));

const sliced_arrays_methods_chaining_5: number[] = doubled_arrays_methods_chaining_5.slice(1, 4);
console.log(sliced_arrays_methods_chaining_5.join(","));

const more_arrays_methods_chaining_5: number[] = [100, 200];
const combined_arrays_methods_chaining_5: number[] = sliced_arrays_methods_chaining_5.concat(more_arrays_methods_chaining_5);
console.log(combined_arrays_methods_chaining_5.join(","));

const reversed_arrays_methods_chaining_5: number[] = combined_arrays_methods_chaining_5.reverse();
console.log(reversed_arrays_methods_chaining_5.join(","));

// --- Context Case: language_arrays_methods_expansion ---
const nums_arrays_methods_expansion_6: number[] = [10, 20, 30, 40];
console.log(nums_arrays_methods_expansion_6.at(0)!);
console.log(nums_arrays_methods_expansion_6.at(-1)!);
console.log(nums_arrays_methods_expansion_6.join("-"));

const first_arrays_methods_expansion_6: number = nums_arrays_methods_expansion_6.shift()!;
console.log(first_arrays_methods_expansion_6);
console.log(nums_arrays_methods_expansion_6.length);

const newLen_arrays_methods_expansion_6: number = nums_arrays_methods_expansion_6.unshift(5);
console.log(newLen_arrays_methods_expansion_6);
console.log(nums_arrays_methods_expansion_6.at(0)!);

const rev_arrays_methods_expansion_6: number[] = nums_arrays_methods_expansion_6.reverse();
console.log(rev_arrays_methods_expansion_6.join(","));

const more_arrays_methods_expansion_6: number[] = [100, 200];
const combined_arrays_methods_expansion_6: number[] = nums_arrays_methods_expansion_6.concat(more_arrays_methods_expansion_6);
console.log(combined_arrays_methods_expansion_6.join(" "));

const strs_arrays_methods_expansion_6: string[] = ["a", "b", "c", "d"];
console.log(strs_arrays_methods_expansion_6.at(-2)!);
const spliced_arrays_methods_expansion_6: string[] = strs_arrays_methods_expansion_6.splice(1, 2);
console.log(spliced_arrays_methods_expansion_6.join(":"));
console.log(strs_arrays_methods_expansion_6.join(":"));

// --- Context Case: language_arrays_mutation ---
const numbers_arrays_mutation_7: number[] = [1, 2, 3];
numbers_arrays_mutation_7[1] = 42;
console.log(numbers_arrays_mutation_7[1]);

// --- Context Case: language_arrays_native-number-array ---
const values_arrays_native_number_array_8: number[] = [10, 20];
console.log(values_arrays_native_number_array_8[1]);

// --- Context Case: language_arrays_native-string-array ---
const values_arrays_native_string_array_9: string[] = ['a', 'b'];
console.log(values_arrays_native_string_array_9[0]);

// --- Context Case: language_arrays_predicates ---
const nums_arrays_predicates_10: number[] = [10, 15, 20, 25, 30];

const hasOdd_arrays_predicates_10: boolean = nums_arrays_predicates_10.some((n: number): boolean => n % 2 !== 0);
console.log(hasOdd_arrays_predicates_10);

const hasNegative_arrays_predicates_10: boolean = nums_arrays_predicates_10.some((n: number): boolean => n < 0);
console.log(hasNegative_arrays_predicates_10);

const allPositive_arrays_predicates_10: boolean = nums_arrays_predicates_10.every((n: number): boolean => n > 0);
console.log(allPositive_arrays_predicates_10);

const allEven_arrays_predicates_10: boolean = nums_arrays_predicates_10.every((n: number): boolean => n % 2 === 0);
console.log(allEven_arrays_predicates_10);

const foundFirstOverTwenty_arrays_predicates_10: number = nums_arrays_predicates_10.find((n: number): boolean => n > 20)!;
console.log(foundFirstOverTwenty_arrays_predicates_10);

const foundFirstDivBySix_arrays_predicates_10: number = nums_arrays_predicates_10.find((n: number): boolean => n % 6 === 0)!;
console.log(foundFirstDivBySix_arrays_predicates_10);

// --- Context Case: language_arrays_string-transformations ---
const words_arrays_string_transformations_11: string[] = ["apple", "banana", "cherry", "date"];

console.log(words_arrays_string_transformations_11.length);
console.log(words_arrays_string_transformations_11.join(" | "));

words_arrays_string_transformations_11.push("elderberry");
console.log(words_arrays_string_transformations_11.join(","));

const removedLast_arrays_string_transformations_11: string = words_arrays_string_transformations_11.pop()!;
console.log(removedLast_arrays_string_transformations_11);

const removedFirst_arrays_string_transformations_11: string = words_arrays_string_transformations_11.shift()!;
console.log(removedFirst_arrays_string_transformations_11);

words_arrays_string_transformations_11.unshift("avocado");
console.log(words_arrays_string_transformations_11.join(","));

const longWords_arrays_string_transformations_11: string[] = words_arrays_string_transformations_11.filter((w: string): boolean => w.length > 5);
console.log(longWords_arrays_string_transformations_11.join(";"));

const upperWords_arrays_string_transformations_11: string[] = words_arrays_string_transformations_11.map((w: string): string => w.toUpperCase());
console.log(upperWords_arrays_string_transformations_11.join(" "));

console.log(words_arrays_string_transformations_11.includes("banana"));
console.log(words_arrays_string_transformations_11.includes("mango"));
console.log(words_arrays_string_transformations_11.indexOf("cherry"));
console.log(words_arrays_string_transformations_11.indexOf("nonexistent"));

// --- Context Case: language_bigint_literals ---
let a_bigint_literals_12 = 100n;
let b_bigint_literals_12 = 20n;
console.log(a_bigint_literals_12);
console.log(b_bigint_literals_12);
console.log(a_bigint_literals_12 + b_bigint_literals_12);
console.log(a_bigint_literals_12 - b_bigint_literals_12);
console.log(a_bigint_literals_12 * b_bigint_literals_12);
console.log(a_bigint_literals_12 / b_bigint_literals_12);
console.log(a_bigint_literals_12 % b_bigint_literals_12);
console.log(a_bigint_literals_12 & b_bigint_literals_12);
console.log(a_bigint_literals_12 | b_bigint_literals_12);
console.log(a_bigint_literals_12 ^ b_bigint_literals_12);
console.log(b_bigint_literals_12 << 2n);
console.log(b_bigint_literals_12 >> 1n);
console.log(a_bigint_literals_12 === 100n);
console.log(a_bigint_literals_12 !== 100n);
console.log(a_bigint_literals_12 > b_bigint_literals_12);
console.log(a_bigint_literals_12 < b_bigint_literals_12);
console.log(-b_bigint_literals_12);
console.log(~b_bigint_literals_12);
let c_bigint_literals_12 = BigInt(42);
console.log(c_bigint_literals_12);
let d_bigint_literals_12 = BigInt("999");
console.log(d_bigint_literals_12);
console.log(a_bigint_literals_12.toString());
console.log(a_bigint_literals_12++);
console.log(a_bigint_literals_12);
console.log(++a_bigint_literals_12);

// --- Context Case: language_generics_array-types ---
const numbers_generics_array_types_13: Array<number> = [10, 20, 30];
const names_generics_array_types_13: Array<string> = ["alpha", "beta", "gamma"];
const readonlyNums_generics_array_types_13: ReadonlyArray<number> = [1, 2, 3];

console.log(numbers_generics_array_types_13.length);
console.log(numbers_generics_array_types_13[1]);
console.log(names_generics_array_types_13[0]);
console.log(names_generics_array_types_13[2]);
console.log(readonlyNums_generics_array_types_13.length);
console.log(readonlyNums_generics_array_types_13[0]);

// --- Context Case: language_generics_arrow-functions ---
const wrap_generics_arrow_functions_14 = <T>(val: T): T[] => [val];

const nums_generics_arrow_functions_14 = wrap_generics_arrow_functions_14(100);
console.log(nums_generics_arrow_functions_14[0]);

const strs_generics_arrow_functions_14 = wrap_generics_arrow_functions_14("hello");
console.log(strs_generics_arrow_functions_14[0]);

// --- Context Case: language_generics_classes ---
class Box_generics_classes_15<T> {
  value: T;

  constructor(value: T) {
    this.value = value;
  }

  getValue(): T {
    return this.value;
  }

  setValue(newValue: T): void {
    this.value = newValue;
  }
}

const numBox_generics_classes_15 = new Box_generics_classes_15<number>(123);
const strBox_generics_classes_15 = new Box_generics_classes_15<string>("scriptgo box");

console.log(numBox_generics_classes_15.getValue());
console.log(strBox_generics_classes_15.getValue());

numBox_generics_classes_15.setValue(456);
strBox_generics_classes_15.setValue("updated box");

console.log(numBox_generics_classes_15.getValue());
console.log(strBox_generics_classes_15.getValue());

// --- Context Case: language_generics_functions ---
function identity_generics_functions_16<T>(value: T): T {
  return value;
}

function pickFirst_generics_functions_16<T, U>(first: T, second: U): T {
  return first;
}

function pickSecond_generics_functions_16<T, U>(first: T, second: U): U {
  return second;
}

function wrapInArray_generics_functions_16<T>(item: T): T[] {
  return [item];
}

function getFirst_generics_functions_16<T>(arr: T[]): T {
  return arr[0];
}

function getLast_generics_functions_16<T>(arr: T[]): T {
  return arr[arr.length - 1];
}

const n_generics_functions_16 = identity_generics_functions_16<number>(42);
const s_generics_functions_16 = identity_generics_functions_16<string>("hello generics");
const inferredNum_generics_functions_16 = identity_generics_functions_16(99);

console.log(n_generics_functions_16);
console.log(s_generics_functions_16);
console.log(inferredNum_generics_functions_16);
console.log(pickFirst_generics_functions_16<number, string>(123, "ignored"));
console.log(pickSecond_generics_functions_16<string, number>("ignored", 456));
console.log(wrapInArray_generics_functions_16<number>(555)[0]);
console.log(getFirst_generics_functions_16<number>([100, 200, 300]));
console.log(getLast_generics_functions_16<string>(["first", "second", "third"]));

// --- Context Case: language_generics_generic-constraints ---
interface HasId_generics_generic_constraints_17 {
  id: number;
  name: string;
}

interface HasExtra_generics_generic_constraints_17 extends HasId_generics_generic_constraints_17 {
  role: string;
}

function getEntityName_generics_generic_constraints_17<T extends HasId_generics_generic_constraints_17>(entity: T): string {
  return entity.name;
}

function compareEntityIds_generics_generic_constraints_17<T extends HasId_generics_generic_constraints_17, U extends HasId_generics_generic_constraints_17>(a: T, b: U): boolean {
  return a.id === b.id;
}

class UserAccount_generics_generic_constraints_17 implements HasExtra_generics_generic_constraints_17 {
  constructor(public id: number, public name: string, public role: string) {}
}

class ProductItem_generics_generic_constraints_17 implements HasId_generics_generic_constraints_17 {
  constructor(public id: number, public name: string) {}
}

const user_generics_generic_constraints_17 = new UserAccount_generics_generic_constraints_17(101, "Alice", "Admin");
const product_generics_generic_constraints_17 = new ProductItem_generics_generic_constraints_17(101, "Widget");
const product2_generics_generic_constraints_17 = new ProductItem_generics_generic_constraints_17(202, "Gadget");

console.log(getEntityName_generics_generic_constraints_17<UserAccount_generics_generic_constraints_17>(user_generics_generic_constraints_17));
console.log(getEntityName_generics_generic_constraints_17<ProductItem_generics_generic_constraints_17>(product_generics_generic_constraints_17));
console.log(compareEntityIds_generics_generic_constraints_17<UserAccount_generics_generic_constraints_17, ProductItem_generics_generic_constraints_17>(user_generics_generic_constraints_17, product_generics_generic_constraints_17));
console.log(compareEntityIds_generics_generic_constraints_17<ProductItem_generics_generic_constraints_17, ProductItem_generics_generic_constraints_17>(product_generics_generic_constraints_17, product2_generics_generic_constraints_17));

// --- Context Case: language_generics_generic-nested ---
interface Container_generics_generic_nested_18<T> {
    value: T;
}

class Box_generics_generic_nested_18<T> {
    item: T;
    constructor(item: T) {
        this.item = item;
    }
    getItem(): T {
        return this.item;
    }
}

function wrap_generics_generic_nested_18<T>(val: T): Container_generics_generic_nested_18<T> {
    const c_generics_generic_nested_18: Container_generics_generic_nested_18<T> = { value: val };
    return c_generics_generic_nested_18;
}

const numBox_generics_generic_nested_18 = new Box_generics_generic_nested_18<number>(42);
console.log(numBox_generics_generic_nested_18.getItem());

const strBox_generics_generic_nested_18 = new Box_generics_generic_nested_18<string>("ScriptGo");
console.log(strBox_generics_generic_nested_18.getItem());

const wrapped_generics_generic_nested_18 = wrap_generics_generic_nested_18<string>("boxed");
console.log(wrapped_generics_generic_nested_18.value);

// --- Context Case: language_generics_generic-pairs ---
function swap_generics_generic_pairs_19<T, U>(first: T, second: U): [U, T] {
  return [second, first];
}

function mapPair_generics_generic_pairs_19<T, U>(first: T, second: T, fn: (val: T) => U): [U, U] {
  return [fn(first), fn(second)];
}

const p1_generics_generic_pairs_19 = swap_generics_generic_pairs_19<number, string>(42, "hello");
console.log(p1_generics_generic_pairs_19[0]);
console.log(p1_generics_generic_pairs_19[1]);

const p2_generics_generic_pairs_19 = swap_generics_generic_pairs_19<boolean, number>(true, 100);
console.log(p2_generics_generic_pairs_19[0]);
console.log(p2_generics_generic_pairs_19[1]);

const doubleNum_generics_generic_pairs_19 = (x: number): number => x * 2;
const mappedNums_generics_generic_pairs_19 = mapPair_generics_generic_pairs_19<number, number>(5, 10, doubleNum_generics_generic_pairs_19);
console.log(mappedNums_generics_generic_pairs_19[0]);
console.log(mappedNums_generics_generic_pairs_19[1]);

const strLen_generics_generic_pairs_19 = (s: string): number => s.length;
const mappedStrings_generics_generic_pairs_19 = mapPair_generics_generic_pairs_19<string, number>("apple", "banana", strLen_generics_generic_pairs_19);
console.log(mappedStrings_generics_generic_pairs_19[0]);
console.log(mappedStrings_generics_generic_pairs_19[1]);

// --- Context Case: language_generics_generic-reduce ---
function customReduce_generics_generic_reduce_20<T, R>(arr: T[], initial: R, reducer: (acc_generics_generic_reduce_20: R, item: T) => R): R {
  let acc_generics_generic_reduce_20: R = initial;
  for (let i_generics_generic_reduce_20 = 0; i_generics_generic_reduce_20 < arr.length; i_generics_generic_reduce_20 = i_generics_generic_reduce_20 + 1) {
    acc_generics_generic_reduce_20 = reducer(acc_generics_generic_reduce_20, arr[i_generics_generic_reduce_20]);
  }
  return acc_generics_generic_reduce_20;
}

const numbers_generics_generic_reduce_20: number[] = [1, 2, 3, 4, 5];
const sum_generics_generic_reduce_20 = customReduce_generics_generic_reduce_20<number, number>(numbers_generics_generic_reduce_20, 0, (acc_generics_generic_reduce_20: number, x_generics_generic_reduce_20: number) => acc_generics_generic_reduce_20 + x_generics_generic_reduce_20);
console.log(sum_generics_generic_reduce_20);

const product_generics_generic_reduce_20 = customReduce_generics_generic_reduce_20<number, number>(numbers_generics_generic_reduce_20, 1, (acc_generics_generic_reduce_20: number, x_generics_generic_reduce_20: number) => acc_generics_generic_reduce_20 * x_generics_generic_reduce_20);
console.log(product_generics_generic_reduce_20);

const words_generics_generic_reduce_20: string[] = ["a", "b", "c", "d"];
const joined_generics_generic_reduce_20 = customReduce_generics_generic_reduce_20<string, string>(words_generics_generic_reduce_20, "start:", (acc_generics_generic_reduce_20: string, s_generics_generic_reduce_20: string) => acc_generics_generic_reduce_20 + s_generics_generic_reduce_20);
console.log(joined_generics_generic_reduce_20);

const totalLength_generics_generic_reduce_20 = customReduce_generics_generic_reduce_20<string, number>(words_generics_generic_reduce_20, 0, (acc_generics_generic_reduce_20: number, s_generics_generic_reduce_20: string) => acc_generics_generic_reduce_20 + s_generics_generic_reduce_20.length);
console.log(totalLength_generics_generic_reduce_20);

// --- Context Case: language_generics_generic-stack ---
class Stack_generics_generic_stack_21<T> {
  private items: T[];

  constructor() {
    this.items = [];
  }

  push(item: T): void {
    this.items.push(item);
  }

  pop(): T | undefined {
    if (this.items.length === 0) {
      return undefined;
    }
    return this.items.pop();
  }

  peek(): T | undefined {
    if (this.items.length === 0) {
      return undefined;
    }
    return this.items[this.items.length - 1];
  }

  isEmpty(): boolean {
    return this.items.length === 0;
  }

  size(): number {
    return this.items.length;
  }
}

// Test with numbers
const numStack_generics_generic_stack_21 = new Stack_generics_generic_stack_21<number>();
console.log(numStack_generics_generic_stack_21.isEmpty());
numStack_generics_generic_stack_21.push(10);
numStack_generics_generic_stack_21.push(20);
numStack_generics_generic_stack_21.push(30);
console.log(numStack_generics_generic_stack_21.size());
console.log(numStack_generics_generic_stack_21.peek());
console.log(numStack_generics_generic_stack_21.pop());
console.log(numStack_generics_generic_stack_21.size());
console.log(numStack_generics_generic_stack_21.isEmpty());

// Test with strings
const strStack_generics_generic_stack_21 = new Stack_generics_generic_stack_21<string>();
strStack_generics_generic_stack_21.push("alpha");
strStack_generics_generic_stack_21.push("beta");
console.log(strStack_generics_generic_stack_21.peek());
console.log(strStack_generics_generic_stack_21.pop());
console.log(strStack_generics_generic_stack_21.pop());
console.log(strStack_generics_generic_stack_21.pop());
console.log(strStack_generics_generic_stack_21.isEmpty());

// --- Context Case: language_generics_interfaces ---
interface Container_generics_interfaces_22<T> {
  get(): T;
}

class ItemHolder_generics_interfaces_22<T> implements Container_generics_interfaces_22<T> {
  item: T;

  constructor(item: T) {
    this.item = item;
  }

  get(): T {
    return this.item;
  }
}

const holder1_generics_interfaces_22 = new ItemHolder_generics_interfaces_22<number>(777);
const holder2_generics_interfaces_22 = new ItemHolder_generics_interfaces_22<string>("interface test");

console.log(holder1_generics_interfaces_22.get());
console.log(holder2_generics_interfaces_22.get());

// --- Context Case: language_generics_multi-param ---
function makePair_generics_multi_param_23<T, U>(a: T, b: U): [T, U] {
    return [a, b];
}

const p_generics_multi_param_23 = makePair_generics_multi_param_23("count", 42);
console.log(p_generics_multi_param_23[0]);
console.log(p_generics_multi_param_23[1]);

// --- Context Case: language_generics_type-aliases ---
type NumArray_generics_type_aliases_24 = Array<number>;
type StrArray_generics_type_aliases_24 = Array<string>;
type MyList_generics_type_aliases_24<T> = T[];

const a_generics_type_aliases_24: NumArray_generics_type_aliases_24 = [1, 2, 3];
const b_generics_type_aliases_24: StrArray_generics_type_aliases_24 = ["hello", "world"];
const c_generics_type_aliases_24: MyList_generics_type_aliases_24<number> = [10, 20, 30];

console.log(a_generics_type_aliases_24.length);
console.log(b_generics_type_aliases_24[0]);
console.log(c_generics_type_aliases_24[2]);

// --- Context Case: language_inference_anonymous-object ---
const user_inference_anonymous_object_25 = { name: "Alice", age: 30, active: true };
console.log(user_inference_anonymous_object_25.name);
console.log(user_inference_anonymous_object_25.age);
console.log(user_inference_anonymous_object_25.active);

const point1_inference_anonymous_object_25 = { x: 10, y: 20 };
const point2_inference_anonymous_object_25 = { x: 5, y: 15 };
console.log(point1_inference_anonymous_object_25.x + point2_inference_anonymous_object_25.x);
console.log(point1_inference_anonymous_object_25.y + point2_inference_anonymous_object_25.y);

// --- Context Case: language_inference_empty-array ---
const emptyNums_inference_empty_array_26: number[] = [];
console.log(emptyNums_inference_empty_array_26.length);
emptyNums_inference_empty_array_26.push(10);
emptyNums_inference_empty_array_26.push(20);
console.log(emptyNums_inference_empty_array_26.length);
console.log(emptyNums_inference_empty_array_26[0]);
console.log(emptyNums_inference_empty_array_26[1]);

const emptyStrings_inference_empty_array_26: string[] = [];
emptyStrings_inference_empty_array_26.push("hello");
emptyStrings_inference_empty_array_26.push("world");
console.log(emptyStrings_inference_empty_array_26.join(" "));

// --- Context Case: language_inference_generic-inference ---
function identity_inference_generic_inference_27<T>(x: T): T {
    return x;
}

function pair_inference_generic_inference_27<A, B>(first: A, second: B): string {
    return `${first} - ${second}`;
}

const num_inference_generic_inference_27 = identity_inference_generic_inference_27(42);
console.log(num_inference_generic_inference_27);

const str_inference_generic_inference_27 = identity_inference_generic_inference_27("scriptgo");
console.log(str_inference_generic_inference_27);

const p_inference_generic_inference_27 = pair_inference_generic_inference_27(123, "apple");
console.log(p_inference_generic_inference_27);

// --- Context Case: language_regex_literals ---
let re_regex_literals_28 = /^[0-9]+$/;
console.log(re_regex_literals_28.source);
console.log(re_regex_literals_28.flags);
console.log(re_regex_literals_28.test("12345"));
console.log(re_regex_literals_28.test("123a"));

let reI_regex_literals_28 = /hello/i;
console.log(reI_regex_literals_28.test("HELLO"));
console.log(reI_regex_literals_28.test("world"));

let text_regex_literals_28 = "The quick brown fox";
console.log(text_regex_literals_28.search(/brown/));
console.log(text_regex_literals_28.search(/cat/));

let replaced_regex_literals_28 = text_regex_literals_28.replace(/fox/, "dog");
console.log(replaced_regex_literals_28);

let repGlobal_regex_literals_28 = "a1b2c3".replace(/[0-9]/g, "X");
console.log(repGlobal_regex_literals_28);

let newRe_regex_literals_28 = new RegExp("abc", "i");
console.log(newRe_regex_literals_28.test("ABC"));

// --- Context Case: language_regex_literals_regex-methods ---
const re_regex_literals_regex_methods_29 = /([0-9]+)/;
const text_regex_literals_regex_methods_29 = "order 12345 placed";

const match_regex_literals_regex_methods_29 = text_regex_literals_regex_methods_29.match(re_regex_literals_regex_methods_29)!;
console.log(match_regex_literals_regex_methods_29.length);
console.log(match_regex_literals_regex_methods_29[0]);
console.log(match_regex_literals_regex_methods_29[1]);

const execRes_regex_literals_regex_methods_29 = re_regex_literals_regex_methods_29.exec("test 999 done")!;
console.log(execRes_regex_literals_regex_methods_29.length);
console.log(execRes_regex_literals_regex_methods_29[0]);
console.log(execRes_regex_literals_regex_methods_29[1]);

// --- Context Case: language_symbol_primitive ---
const s1_symbol_primitive_30: symbol = Symbol("foo");
const s2_symbol_primitive_30: symbol = Symbol("foo");
console.log(s1_symbol_primitive_30 === s2_symbol_primitive_30);
console.log(s1_symbol_primitive_30 === s1_symbol_primitive_30);

const reg1_symbol_primitive_30: symbol = Symbol.for("app.key");
const reg2_symbol_primitive_30: symbol = Symbol.for("app.key");
console.log(reg1_symbol_primitive_30 === reg2_symbol_primitive_30);

const k1_symbol_primitive_30 = Symbol.keyFor(reg1_symbol_primitive_30);
console.log(k1_symbol_primitive_30);
const k2_symbol_primitive_30 = Symbol.keyFor(s1_symbol_primitive_30);
console.log(k2_symbol_primitive_30);

console.log(s1_symbol_primitive_30.description);
console.log(s1_symbol_primitive_30.toString());

const sEmpty_symbol_primitive_30: symbol = Symbol();
console.log(sEmpty_symbol_primitive_30.description);
console.log(sEmpty_symbol_primitive_30.toString());

const it1_symbol_primitive_30: symbol = Symbol.iterator;
const it2_symbol_primitive_30: symbol = Symbol.iterator;
console.log(it1_symbol_primitive_30 === it2_symbol_primitive_30);
console.log(it1_symbol_primitive_30.toString());

console.log(typeof s1_symbol_primitive_30);

// --- Context Case: language_symbol_primitive_symbol-registry-advanced ---
const s1_symbol_primitive_symbol_registry_advanced_31: symbol = Symbol.for("app.version");
const s2_symbol_primitive_symbol_registry_advanced_31: symbol = Symbol.for("app.version");
const s3_symbol_primitive_symbol_registry_advanced_31: symbol = Symbol.for("app.author");

console.log(s1_symbol_primitive_symbol_registry_advanced_31 === s2_symbol_primitive_symbol_registry_advanced_31);
console.log(s1_symbol_primitive_symbol_registry_advanced_31 === s3_symbol_primitive_symbol_registry_advanced_31);

const k1_symbol_primitive_symbol_registry_advanced_31: string = Symbol.keyFor(s1_symbol_primitive_symbol_registry_advanced_31)!;
const k3_symbol_primitive_symbol_registry_advanced_31: string = Symbol.keyFor(s3_symbol_primitive_symbol_registry_advanced_31)!;
console.log(k1_symbol_primitive_symbol_registry_advanced_31);
console.log(k3_symbol_primitive_symbol_registry_advanced_31);

const localSym1_symbol_primitive_symbol_registry_advanced_31: symbol = Symbol("unique");
const localSym2_symbol_primitive_symbol_registry_advanced_31: symbol = Symbol("unique");
console.log(localSym1_symbol_primitive_symbol_registry_advanced_31 === localSym2_symbol_primitive_symbol_registry_advanced_31);

console.log(s1_symbol_primitive_symbol_registry_advanced_31.toString());
console.log(localSym1_symbol_primitive_symbol_registry_advanced_31.toString());

// --- Context Case: language_typedarrays_all_types ---
// Int8Array
const i8_typedarrays_all_types_32 = new Int8Array(3);
i8_typedarrays_all_types_32[0] = 127;
i8_typedarrays_all_types_32[1] = -128;
i8_typedarrays_all_types_32[2] = 200;
console.log(i8_typedarrays_all_types_32.BYTES_PER_ELEMENT);
console.log(i8_typedarrays_all_types_32.byteLength);
console.log(i8_typedarrays_all_types_32[0]);
console.log(i8_typedarrays_all_types_32[1]);
console.log(i8_typedarrays_all_types_32[2]);

// Uint8ClampedArray
const u8c_typedarrays_all_types_32 = new Uint8ClampedArray(4);
u8c_typedarrays_all_types_32[0] = -10;
u8c_typedarrays_all_types_32[1] = 120.4;
u8c_typedarrays_all_types_32[2] = 120.6;
u8c_typedarrays_all_types_32[3] = 300;
console.log(u8c_typedarrays_all_types_32.BYTES_PER_ELEMENT);
console.log(u8c_typedarrays_all_types_32[0]);
console.log(u8c_typedarrays_all_types_32[1]);
console.log(u8c_typedarrays_all_types_32[2]);
console.log(u8c_typedarrays_all_types_32[3]);

// Int16Array & Uint16Array
const i16_typedarrays_all_types_32 = new Int16Array(2);
i16_typedarrays_all_types_32[0] = 32767;
i16_typedarrays_all_types_32[1] = -32768;
console.log(i16_typedarrays_all_types_32.BYTES_PER_ELEMENT);
console.log(i16_typedarrays_all_types_32.byteLength);
console.log(i16_typedarrays_all_types_32[0]);
console.log(i16_typedarrays_all_types_32[1]);

const u16_typedarrays_all_types_32 = new Uint16Array(2);
u16_typedarrays_all_types_32[0] = 65535;
u16_typedarrays_all_types_32[1] = 100;
console.log(u16_typedarrays_all_types_32.BYTES_PER_ELEMENT);
console.log(u16_typedarrays_all_types_32.byteLength);
console.log(u16_typedarrays_all_types_32[0]);
console.log(u16_typedarrays_all_types_32[1]);

// Uint32Array & Float32Array
const u32_typedarrays_all_types_32 = new Uint32Array(2);
u32_typedarrays_all_types_32[0] = 4294967295;
u32_typedarrays_all_types_32[1] = 500;
console.log(u32_typedarrays_all_types_32.BYTES_PER_ELEMENT);
console.log(u32_typedarrays_all_types_32.byteLength);
console.log(u32_typedarrays_all_types_32[0]);
console.log(u32_typedarrays_all_types_32[1]);

const f32_typedarrays_all_types_32 = new Float32Array(2);
f32_typedarrays_all_types_32[0] = 1.5;
f32_typedarrays_all_types_32[1] = -2.5;
console.log(f32_typedarrays_all_types_32.BYTES_PER_ELEMENT);
console.log(f32_typedarrays_all_types_32.byteLength);
console.log(f32_typedarrays_all_types_32[0]);
console.log(f32_typedarrays_all_types_32[1]);

// BigInt64Array & BigUint64Array
const bi64_typedarrays_all_types_32 = new BigInt64Array(2);
bi64_typedarrays_all_types_32[0] = 9007199254740993n;
bi64_typedarrays_all_types_32[1] = -9007199254740993n;
console.log(bi64_typedarrays_all_types_32.BYTES_PER_ELEMENT);
console.log(bi64_typedarrays_all_types_32.byteLength);
console.log(bi64_typedarrays_all_types_32[0]);
console.log(bi64_typedarrays_all_types_32[1]);

const bu64_typedarrays_all_types_32 = new BigUint64Array(2);
bu64_typedarrays_all_types_32[0] = 1844674407370955161n;
bu64_typedarrays_all_types_32[1] = 42n;
console.log(bu64_typedarrays_all_types_32.BYTES_PER_ELEMENT);
console.log(bu64_typedarrays_all_types_32.byteLength);
console.log(bu64_typedarrays_all_types_32[0]);
console.log(bu64_typedarrays_all_types_32[1]);

console.log(ArrayBuffer.isView(i8_typedarrays_all_types_32));
console.log(ArrayBuffer.isView(u8c_typedarrays_all_types_32));
console.log(ArrayBuffer.isView(bi64_typedarrays_all_types_32));
console.log(ArrayBuffer.isView("not a view"));

// --- Context Case: language_typedarrays_arraybuffer ---
const buf_typedarrays_arraybuffer_33 = new ArrayBuffer(16);
console.log(buf_typedarrays_arraybuffer_33.byteLength);
console.log(ArrayBuffer.isView(buf_typedarrays_arraybuffer_33));

const u8_typedarrays_arraybuffer_33 = new Uint8Array(buf_typedarrays_arraybuffer_33, 4, 8);
console.log(u8_typedarrays_arraybuffer_33.length);
console.log(u8_typedarrays_arraybuffer_33.byteLength);
console.log(u8_typedarrays_arraybuffer_33.byteOffset);
console.log(ArrayBuffer.isView(u8_typedarrays_arraybuffer_33));

u8_typedarrays_arraybuffer_33[0] = 99;
const u8All_typedarrays_arraybuffer_33 = new Uint8Array(buf_typedarrays_arraybuffer_33);
console.log(u8All_typedarrays_arraybuffer_33[4]);

// --- Context Case: language_typedarrays_dataview ---
const buf_typedarrays_dataview_34 = new ArrayBuffer(16);
const dv_typedarrays_dataview_34 = new DataView(buf_typedarrays_dataview_34, 0, 16);

console.log(dv_typedarrays_dataview_34.byteLength);
console.log(dv_typedarrays_dataview_34.byteOffset);

dv_typedarrays_dataview_34.setInt8(0, -12);
dv_typedarrays_dataview_34.setUint8(1, 250);
console.log(dv_typedarrays_dataview_34.getInt8(0));
console.log(dv_typedarrays_dataview_34.getUint8(1));

// 16-bit
dv_typedarrays_dataview_34.setInt16(2, 4660); // Big-Endian by default (0x1234)
console.log(dv_typedarrays_dataview_34.getInt16(2)); // 4660
console.log(dv_typedarrays_dataview_34.getInt16(2, true)); // 13330 (0x3412)

dv_typedarrays_dataview_34.setUint16(4, 22136, true); // Little-Endian write (0x5678)
console.log(dv_typedarrays_dataview_34.getUint16(4, true)); // 22136
console.log(dv_typedarrays_dataview_34.getUint16(4)); // 30806 (0x7856)

// 32-bit
dv_typedarrays_dataview_34.setInt32(6, -100000);
console.log(dv_typedarrays_dataview_34.getInt32(6));

// float
dv_typedarrays_dataview_34.setFloat32(0, 1.5, true);
console.log(dv_typedarrays_dataview_34.getFloat32(0, true));

dv_typedarrays_dataview_34.setFloat64(0, 3.141592653589793);
console.log(dv_typedarrays_dataview_34.getFloat64(0));

// BigInt
dv_typedarrays_dataview_34.setBigInt64(8, 9007199254740993n, true);
console.log(dv_typedarrays_dataview_34.getBigInt64(8, true));

console.log(ArrayBuffer.isView(dv_typedarrays_dataview_34));

// --- Context Case: language_typedarrays_float64 ---
const f64_typedarrays_float64_35 = new Float64Array(3);
f64_typedarrays_float64_35[0] = 3.14159;
f64_typedarrays_float64_35[1] = 2.71828;
f64_typedarrays_float64_35[2] = 42.5;

console.log(f64_typedarrays_float64_35.length);
console.log(f64_typedarrays_float64_35.byteLength);
console.log(f64_typedarrays_float64_35.BYTES_PER_ELEMENT);
console.log(f64_typedarrays_float64_35[0]);
console.log(f64_typedarrays_float64_35[1]);
console.log(f64_typedarrays_float64_35[2]);

// --- Context Case: language_typedarrays_int32 ---
const i32_typedarrays_int32_36 = new Int32Array(3);
i32_typedarrays_int32_36[0] = 100000;
i32_typedarrays_int32_36[1] = -50000;
i32_typedarrays_int32_36[2] = 42;

console.log(i32_typedarrays_int32_36.length);
console.log(i32_typedarrays_int32_36.byteLength);
console.log(i32_typedarrays_int32_36.BYTES_PER_ELEMENT);
console.log(i32_typedarrays_int32_36[0]);
console.log(i32_typedarrays_int32_36[1]);
console.log(i32_typedarrays_int32_36[2]);

// --- Context Case: language_typedarrays_methods ---
const src_typedarrays_methods_37 = new Uint8Array(6);
src_typedarrays_methods_37.fill(7, 1, 4);
console.log(src_typedarrays_methods_37[0]);
console.log(src_typedarrays_methods_37[1]);
console.log(src_typedarrays_methods_37[2]);
console.log(src_typedarrays_methods_37[3]);
console.log(src_typedarrays_methods_37[4]);

const sub_typedarrays_methods_37 = src_typedarrays_methods_37.subarray(1, 4);
console.log(sub_typedarrays_methods_37.length);
console.log(sub_typedarrays_methods_37.byteOffset);
console.log(sub_typedarrays_methods_37[0]);

sub_typedarrays_methods_37[0] = 42;
console.log(src_typedarrays_methods_37[1]); // verifies subarray shares underlying buffer

const sl_typedarrays_methods_37 = src_typedarrays_methods_37.slice(1, 4);
console.log(sl_typedarrays_methods_37.length);
console.log(sl_typedarrays_methods_37[0]);
sl_typedarrays_methods_37[0] = 99;
console.log(src_typedarrays_methods_37[1]); // verifies slice copies buffer (src[1] is still 42)

const dest_typedarrays_methods_37 = new Uint8Array(8);
dest_typedarrays_methods_37.set(sub_typedarrays_methods_37, 2);
console.log(dest_typedarrays_methods_37[2]);
console.log(dest_typedarrays_methods_37[3]);

// --- Context Case: language_typedarrays_uint8 ---
const u8_typedarrays_uint8_38 = new Uint8Array(4);
u8_typedarrays_uint8_38[0] = 10;
u8_typedarrays_uint8_38[1] = 255;
u8_typedarrays_uint8_38[2] = 300; // truncated to 44 (300 & 0xff)
u8_typedarrays_uint8_38[3] = 42;

console.log(u8_typedarrays_uint8_38.length);
console.log(u8_typedarrays_uint8_38.byteLength);
console.log(u8_typedarrays_uint8_38.byteOffset);
console.log(u8_typedarrays_uint8_38.BYTES_PER_ELEMENT);
console.log(u8_typedarrays_uint8_38[0]);
console.log(u8_typedarrays_uint8_38[1]);
console.log(u8_typedarrays_uint8_38[2]);
console.log(u8_typedarrays_uint8_38[3]);

// --- Context Case: language_unions_discriminated-unions-extended ---
interface Square_unions_discriminated_unions_extended_39 {
  kind: "square";
  size: number;
}

interface Rectangle_unions_discriminated_unions_extended_39 {
  kind: "rectangle";
  width: number;
  height: number;
}

interface Circle_unions_discriminated_unions_extended_39 {
  kind: "circle";
  radius: number;
}

type Shape_unions_discriminated_unions_extended_39 = Square_unions_discriminated_unions_extended_39 | Rectangle_unions_discriminated_unions_extended_39 | Circle_unions_discriminated_unions_extended_39;

function area_unions_discriminated_unions_extended_39(s: Shape_unions_discriminated_unions_extended_39): number {
  if (s.kind === "square") {
    const sq_unions_discriminated_unions_extended_39 = s as Square_unions_discriminated_unions_extended_39;
    return sq_unions_discriminated_unions_extended_39.size * sq_unions_discriminated_unions_extended_39.size;
  }
  if (s.kind === "rectangle") {
    const rect_unions_discriminated_unions_extended_39 = s as Rectangle_unions_discriminated_unions_extended_39;
    return rect_unions_discriminated_unions_extended_39.width * rect_unions_discriminated_unions_extended_39.height;
  }
  const circ_unions_discriminated_unions_extended_39 = s as Circle_unions_discriminated_unions_extended_39;
  return 3.14 * circ_unions_discriminated_unions_extended_39.radius * circ_unions_discriminated_unions_extended_39.radius;
}

const s1_unions_discriminated_unions_extended_39: Square_unions_discriminated_unions_extended_39 = { kind: "square", size: 5 };
const s2_unions_discriminated_unions_extended_39: Rectangle_unions_discriminated_unions_extended_39 = { kind: "rectangle", width: 4, height: 6 };
const s3_unions_discriminated_unions_extended_39: Circle_unions_discriminated_unions_extended_39 = { kind: "circle", radius: 10 };

console.log(area_unions_discriminated_unions_extended_39(s1_unions_discriminated_unions_extended_39));
console.log(area_unions_discriminated_unions_extended_39(s2_unions_discriminated_unions_extended_39));
console.log(area_unions_discriminated_unions_extended_39(s3_unions_discriminated_unions_extended_39));

// --- Context Case: language_unions_literal-unions ---
type Direction_unions_literal_unions_40 = "asc" | "desc";
type HttpStatus_unions_literal_unions_40 = 200 | 404 | 500;

function sortOrder_unions_literal_unions_40(dir: Direction_unions_literal_unions_40): string {
    if (dir === "asc") {
        return "ascending";
    }
    return "descending";
}

function handleStatus_unions_literal_unions_40(status: HttpStatus_unions_literal_unions_40): string {
    switch (status) {
        case 200:
            return "OK";
        case 404:
            return "Not Found";
        case 500:
            return "Server Error";
        default:
            return "Unknown";
    }
}

console.log(sortOrder_unions_literal_unions_40("asc"));
console.log(sortOrder_unions_literal_unions_40("desc"));
console.log(handleStatus_unions_literal_unions_40(200));
console.log(handleStatus_unions_literal_unions_40(404));
console.log(handleStatus_unions_literal_unions_40(500));

// --- Context Case: language_unions_multi-variant-discriminated ---
interface Circle_unions_multi_variant_discriminated_41 {
    kind: "circle";
    radius: number;
}

interface Rectangle_unions_multi_variant_discriminated_41 {
    kind: "rectangle";
    width: number;
    height: number;
}

interface Square_unions_multi_variant_discriminated_41 {
    kind: "square";
    size: number;
}

type Shape_unions_multi_variant_discriminated_41 = Circle_unions_multi_variant_discriminated_41 | Rectangle_unions_multi_variant_discriminated_41 | Square_unions_multi_variant_discriminated_41;

function calculateArea_unions_multi_variant_discriminated_41(s: Shape_unions_multi_variant_discriminated_41): number {
    if (s.kind === "circle") {
        return Math.floor(Math.PI * s.radius * s.radius);
    }
    if (s.kind === "rectangle") {
        return s.width * s.height;
    }
    if (s.kind === "square") {
        return s.size * s.size;
    }
    return 0;
}

function describeShape_unions_multi_variant_discriminated_41(s: Shape_unions_multi_variant_discriminated_41): string {
    switch (s.kind) {
        case "circle":
            return `Circle(r=${s.radius})`;
        case "rectangle":
            return `Rectangle(w=${s.width}, h=${s.height})`;
        case "square":
            return `Square(s=${s.size})`;
    }
}

const c_unions_multi_variant_discriminated_41: Shape_unions_multi_variant_discriminated_41 = { kind: "circle", radius: 5 };
const r_unions_multi_variant_discriminated_41: Shape_unions_multi_variant_discriminated_41 = { kind: "rectangle", width: 10, height: 4 };
const sq_unions_multi_variant_discriminated_41: Shape_unions_multi_variant_discriminated_41 = { kind: "square", size: 6 };

console.log(describeShape_unions_multi_variant_discriminated_41(c_unions_multi_variant_discriminated_41));
console.log(calculateArea_unions_multi_variant_discriminated_41(c_unions_multi_variant_discriminated_41));

console.log(describeShape_unions_multi_variant_discriminated_41(r_unions_multi_variant_discriminated_41));
console.log(calculateArea_unions_multi_variant_discriminated_41(r_unions_multi_variant_discriminated_41));

console.log(describeShape_unions_multi_variant_discriminated_41(sq_unions_multi_variant_discriminated_41));
console.log(calculateArea_unions_multi_variant_discriminated_41(sq_unions_multi_variant_discriminated_41));
