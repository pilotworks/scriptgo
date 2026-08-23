// ScriptGo Corpus: Array Standard Builtin APIs
// Consolidated test suite with inline assertions.
export {};

// @api: array.at
// @expect: 10
// @expect: 30
const a_array_at_0: number[] = [10, 20, 30];
console.log(a_array_at_0.at(0));
console.log(a_array_at_0.at(-1));

// @api: array.concat
// @expect: 1,2,3,4
const a_array_concat_1: number[] = [1, 2]; const b_array_concat_1: number[] = [3, 4]; console.log(a_array_concat_1.concat(b_array_concat_1).join(","));

// @api: array.copyWithin
// @expect: 4,2,3,4,5
const a_array_copyWithin_2: number[] = [1, 2, 3, 4, 5];
a_array_copyWithin_2.copyWithin(0, 3, 4);
console.log(a_array_copyWithin_2.join(","));

// @api: array.entries
// @expect: [0, a]
// @expect: [1, b]
const a_array_entries_3: string[] = ["a", "b"];
for (const entry of a_array_entries_3.entries()) {
  console.log(entry);
}

// @api: array.every
// @expect: true
// @expect: false
const a_array_every_4: number[] = [1, 2, 3];
console.log(a_array_every_4.every((x: number) => x > 0));
console.log(a_array_every_4.every((x: number) => x > 1));

// @api: array.fill
// @expect: 0,0,0
const a_array_fill_5: number[] = [1, 2, 3];
a_array_fill_5.fill(0);
console.log(a_array_fill_5.join(","));

// @api: array.filter
// @expect: 20,30
const arr_array_filter_6: number[] = [10, 20, 30]; const res_array_filter_6 = arr_array_filter_6.filter((x: number): boolean => x > 15); console.log(res_array_filter_6.join(","));

// @api: array.find
// @expect: 20
const a_array_find_7: number[] = [10, 20, 30];
console.log(a_array_find_7.find((x: number) => x > 15));

// @api: array.findIndex
// @expect: 1
// @expect: -1
const a_array_findIndex_8: number[] = [10, 20, 30];
console.log(a_array_findIndex_8.findIndex((x: number) => x === 20));
console.log(a_array_findIndex_8.findIndex((x: number) => x === 99));

// @api: array.findLast
// @expect: 130
const a_array_findLast_9: number[] = [5, 12, 50, 130, 44];
const found_array_findLast_9 = a_array_findLast_9.findLast((n: number) => n > 45);
console.log(found_array_findLast_9);

// @api: array.findLastIndex
// @expect: 3
const a_array_findLastIndex_10: number[] = [5, 12, 50, 130, 44];
const idx_array_findLastIndex_10 = a_array_findLastIndex_10.findLastIndex((n: number) => n > 45);
console.log(idx_array_findLastIndex_10);

// @api: array.flat
// @expect: 1,2,3
const a_array_flat_11: number[] = [1, 2, 3];
const f_array_flat_11 = a_array_flat_11.flat();
console.log(f_array_flat_11.join(","));

// @api: array.flatMap
// @expect: 2,4,6
const a_array_flatMap_12: number[] = [1, 2, 3];
const res_array_flatMap_12 = a_array_flatMap_12.flatMap((x: number) => x * 2);
console.log(res_array_flatMap_12.join(","));

// @api: array.forEach
// @expect: 1
// @expect: 2
const arr_array_forEach_13: number[] = [1, 2]; arr_array_forEach_13.forEach((x: number) => { console.log(x); });

// @api: array.from
// @expect: h,e,l,l,o
const s_array_from_14 = "hello";
const arr_array_from_14 = Array.from(s_array_from_14);
console.log(arr_array_from_14.join(","));

// @api: array.fromAsync
// @expect: 2,4,6
const a_array_fromAsync_15: number[] = [1, 2, 3];
const arr_array_fromAsync_15 = await Array.fromAsync(a_array_fromAsync_15, (x: number) => x * 2);
console.log(arr_array_fromAsync_15.join(","));

// @api: array.includes
// @expect: true
// @expect: false
const arr_array_includes_16: number[] = [10, 20, 30]; console.log(arr_array_includes_16.includes(20)); console.log(arr_array_includes_16.includes(99));

// @api: array.indexOf
// @expect: 1
// @expect: -1
const arr_array_indexOf_17: number[] = [10, 20, 30]; console.log(arr_array_indexOf_17.indexOf(20)); console.log(arr_array_indexOf_17.indexOf(99));

// @api: array.isArray
// @expect: true
const a_array_isArray_18: number[] = [1, 2, 3];
console.log(Array.isArray(a_array_isArray_18));

// @api: array.join
// @expect: a-b-c
// @expect: a,b,c
const arr_array_join_19: string[] = ["a", "b", "c"]; console.log(arr_array_join_19.join("-")); console.log(arr_array_join_19.join());

// @api: array.keys
// @expect: 0
// @expect: 1
// @expect: 2
const a_array_keys_20: string[] = ["a", "b", "c"];
for (const k of a_array_keys_20.keys()) {
  console.log(k);
}

// @api: array.lastIndexOf
// @expect: 3
// @expect: -1
const a_array_lastIndexOf_21: number[] = [2, 5, 9, 2];
console.log(a_array_lastIndexOf_21.lastIndexOf(2));
console.log(a_array_lastIndexOf_21.lastIndexOf(7));

// @api: array.length
// @expect: 3
const a_array_length_22: number[] = [1, 2, 3];
console.log(a_array_length_22.length);

// @api: array.map
// @expect: 2,4,6
const arr_array_map_23: number[] = [1, 2, 3]; const res_array_map_23 = arr_array_map_23.map((x: number): number => x * 2); console.log(res_array_map_23.join(","));

// @api: array.of
// @expect: 1,2,3
const a_array_of_24 = Array.of(1, 2, 3);
console.log(a_array_of_24.join(","));

// @api: array.pop
// @expect: 3
// @expect: 1,2
const arr_array_pop_25: number[] = [1, 2, 3]; console.log(arr_array_pop_25.pop()); console.log(arr_array_pop_25.join(","));

// @api: array.push
// @expect: 3
// @expect: 1,2,3
const arr_array_push_26: number[] = [1, 2]; console.log(arr_array_push_26.push(3)); console.log(arr_array_push_26.join(","));

// @api: array.reduce
// @expect: 10
const arr_array_reduce_27: number[] = [1, 2, 3, 4]; const sum_array_reduce_27 = arr_array_reduce_27.reduce((acc: number, curr_array_reduce_27: number): number => acc + curr_array_reduce_27, 0); console.log(sum_array_reduce_27);

// @api: array.reduceRight
// @expect: 10
const a_array_reduceRight_28: number[] = [1, 2, 3, 4];
const sum_array_reduceRight_28 = a_array_reduceRight_28.reduceRight((acc: number, val_array_reduceRight_28: number) => acc + val_array_reduceRight_28, 0);
console.log(sum_array_reduceRight_28);

// @api: array.reverse
// @expect: 3,2,1
const a_array_reverse_29: number[] = [1, 2, 3];
a_array_reverse_29.reverse();
console.log(a_array_reverse_29.join(","));

// @api: array.shift
// @expect: 1
// @expect: 2,3
const arr_array_shift_30: number[] = [1, 2, 3]; console.log(arr_array_shift_30.shift()); console.log(arr_array_shift_30.join(","));

// @api: array.slice
// @expect: 20,30
const arr_array_slice_31: number[] = [10, 20, 30, 40]; console.log(arr_array_slice_31.slice(1, 3).join(","));

// @api: array.some
// @expect: true
// @expect: false
const a_array_some_32: number[] = [1, 2, 3];
console.log(a_array_some_32.some((x: number) => x > 2));
console.log(a_array_some_32.some((x: number) => x > 5));

// @api: array.sort
// @expect: 1,1,3,4,5,9
const a_array_sort_33: number[] = [3, 1, 4, 1, 5, 9];
a_array_sort_33.sort();
console.log(a_array_sort_33.join(","));

// @api: array.splice
// @expect: 2,3
// @expect: 1,4
const arr_array_splice_34: number[] = [1, 2, 3, 4]; const rem_array_splice_34 = arr_array_splice_34.splice(1, 2); console.log(rem_array_splice_34.join(",")); console.log(arr_array_splice_34.join(","));

// @api: array.toLocaleString
// @expect: 1,2,3
const a_array_toLocaleString_35: number[] = [1, 2, 3];
console.log(a_array_toLocaleString_35.toLocaleString());

// @api: array.toReversed
// @expect: 3,2,1
// @expect: 1,2,3
const a_array_toReversed_36: number[] = [1, 2, 3];
const r_array_toReversed_36 = a_array_toReversed_36.toReversed();
console.log(r_array_toReversed_36.join(","));
console.log(a_array_toReversed_36.join(","));

// @api: array.toSorted
// @expect: 1,2,3
// @expect: 3,1,2
const a_array_toSorted_37: number[] = [3, 1, 2];
const s_array_toSorted_37 = a_array_toSorted_37.toSorted();
console.log(s_array_toSorted_37.join(","));
console.log(a_array_toSorted_37.join(","));

// @api: array.toSpliced
// @expect: 1,4
// @expect: 1,2,3,4
const a_array_toSpliced_38: number[] = [1, 2, 3, 4];
const s_array_toSpliced_38 = a_array_toSpliced_38.toSpliced(1, 2);
console.log(s_array_toSpliced_38.join(","));
console.log(a_array_toSpliced_38.join(","));

// @api: array.toString
// @expect: 1,2,3
const a_array_toString_39: number[] = [1, 2, 3];
console.log(a_array_toString_39.toString());

// @api: array.unshift
// @expect: 3
// @expect: 1,2,3
const arr_array_unshift_40: number[] = [2, 3]; console.log(arr_array_unshift_40.unshift(1)); console.log(arr_array_unshift_40.join(","));

// @api: array.values
// @expect: 10
// @expect: 20
// @expect: 30
const a_array_values_41: number[] = [10, 20, 30];
for (const v of a_array_values_41.values()) {
  console.log(v);
}

// @api: array.with
// @expect: 1,99,3
// @expect: 1,2,3
const a_array_with_42: number[] = [1, 2, 3];
const w_array_with_42 = a_array_with_42.with(1, 99);
console.log(w_array_with_42.join(","));
console.log(a_array_with_42.join(","));

// @api: array.constructor
// @expect: 1,2,3
const a_array_constructor_43: number[] = new Array(1, 2, 3);
console.log(a_array_constructor_43.join(","));

