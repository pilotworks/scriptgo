// ScriptGo Corpus: Date Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: date.UTC
// @expect: 1577836800000
console.log(Date.UTC(2020, 0, 1));

// @api: date.constructor
// @expect: true
const d1_date_constructor_1 = new Date(1577836800000);
const d2_date_constructor_1 = new Date("2020-01-01T00:00:00.000Z");
console.log(d1_date_constructor_1.getTime() === d2_date_constructor_1.getTime());

// @api: date.getDate
// @expect: true
const d_date_getDate_2 = new Date(1577836800000); console.log(d_date_getDate_2.getDate() >= 1 && d_date_getDate_2.getDate() <= 31);

// @api: date.getDay
// @expect: true
const d_date_getDay_3 = new Date(1577836800000); console.log(d_date_getDay_3.getDay() >= 0 && d_date_getDay_3.getDay() <= 6);

// @api: date.getFullYear
// @expect: true
const d_date_getFullYear_4 = new Date(1577836800000); console.log(d_date_getFullYear_4.getFullYear() >= 2019 && d_date_getFullYear_4.getFullYear() <= 2020);

// @api: date.getHours
// @expect: true
const d_date_getHours_5 = new Date(1577836800000); console.log(d_date_getHours_5.getHours() >= 0 && d_date_getHours_5.getHours() <= 23);

// @api: date.getMilliseconds
// @expect: 123
const d_date_getMilliseconds_6 = new Date(1577836800123); console.log(d_date_getMilliseconds_6.getMilliseconds());

// @api: date.getMinutes
// @expect: true
const d_date_getMinutes_7 = new Date(1577836800000); console.log(d_date_getMinutes_7.getMinutes() >= 0 && d_date_getMinutes_7.getMinutes() <= 59);

// @api: date.getMonth
// @expect: true
const d_date_getMonth_8 = new Date(1577836800000); console.log(d_date_getMonth_8.getMonth() >= 0 && d_date_getMonth_8.getMonth() <= 11);

// @api: date.getSeconds
// @expect: 45
const d_date_getSeconds_9 = new Date(1577836845000); console.log(d_date_getSeconds_9.getSeconds());

// @api: date.getTime
// @expect: 1000
const d_date_getTime_10 = new Date(1000); console.log(d_date_getTime_10.getTime());

// @api: date.getTimezoneOffset
// @expect: number
const d_date_getTimezoneOffset_11 = new Date(1577836800000); console.log(typeof d_date_getTimezoneOffset_11.getTimezoneOffset());

// @api: date.getUTCDate
// @expect: 1
const d_date_getUTCDate_12 = new Date(1577836800000); console.log(d_date_getUTCDate_12.getUTCDate());

// @api: date.getUTCDay
// @expect: 3
const d_date_getUTCDay_13 = new Date(1577836800000); console.log(d_date_getUTCDay_13.getUTCDay());

// @api: date.getUTCFullYear
// @expect: 2020
const d_date_getUTCFullYear_14 = new Date(1577836800000); console.log(d_date_getUTCFullYear_14.getUTCFullYear());

// @api: date.getUTCHours
// @expect: 0
const d_date_getUTCHours_15 = new Date(1577836800000); console.log(d_date_getUTCHours_15.getUTCHours());

// @api: date.getUTCMilliseconds
// @expect: 456
const d_date_getUTCMilliseconds_16 = new Date(1577836800456); console.log(d_date_getUTCMilliseconds_16.getUTCMilliseconds());

// @api: date.getUTCMinutes
// @expect: 0
const d_date_getUTCMinutes_17 = new Date(1577836800000); console.log(d_date_getUTCMinutes_17.getUTCMinutes());

// @api: date.getUTCMonth
// @expect: 0
const d_date_getUTCMonth_18 = new Date(1577836800000); console.log(d_date_getUTCMonth_18.getUTCMonth());

// @api: date.getUTCSeconds
// @expect: 30
const d_date_getUTCSeconds_19 = new Date(1577836830000); console.log(d_date_getUTCSeconds_19.getUTCSeconds());

// @api: date.now
// @expect: true
const t_date_now_20 = Date.now(); console.log(t_date_now_20 > 0);

// @api: date.parse
// @expect: true
const t_date_parse_21 = Date.parse("2026-01-01T00:00:00Z"); console.log(t_date_parse_21 > 0);

// @api: date.setDate
// @expect: 15
const d_date_setDate_22 = new Date(1577836800000); d_date_setDate_22.setDate(15); console.log(d_date_setDate_22.getDate());

// @api: date.setFullYear
// @expect: 2025
const d_date_setFullYear_23 = new Date(1577836800000); d_date_setFullYear_23.setFullYear(2025); console.log(d_date_setFullYear_23.getFullYear());

// @api: date.setHours
// @expect: 10
const d_date_setHours_24 = new Date(1577836800000); d_date_setHours_24.setHours(10); console.log(d_date_setHours_24.getHours());

// @api: date.setMilliseconds
// @expect: 789
const d_date_setMilliseconds_25 = new Date(1577836800000); d_date_setMilliseconds_25.setMilliseconds(789); console.log(d_date_setMilliseconds_25.getMilliseconds());

// @api: date.setMinutes
// @expect: 42
const d_date_setMinutes_26 = new Date(1577836800000); d_date_setMinutes_26.setMinutes(42); console.log(d_date_setMinutes_26.getMinutes());

// @api: date.setMonth
// @expect: 5
const d_date_setMonth_27 = new Date(1577836800000); d_date_setMonth_27.setMonth(5); console.log(d_date_setMonth_27.getMonth());

// @api: date.setSeconds
// @expect: 25
const d_date_setSeconds_28 = new Date(1577836800000); d_date_setSeconds_28.setSeconds(25); console.log(d_date_setSeconds_28.getSeconds());

// @api: date.setTime
// @expect: 1577836800000
const d_date_setTime_29 = new Date(0); d_date_setTime_29.setTime(1577836800000); console.log(d_date_setTime_29.getTime());

// @api: date.setUTCDate
// @expect: 20
const d_date_setUTCDate_30 = new Date(1577836800000); d_date_setUTCDate_30.setUTCDate(20); console.log(d_date_setUTCDate_30.getUTCDate());

// @api: date.setUTCFullYear
// @expect: 2030
const d_date_setUTCFullYear_31 = new Date(1577836800000); d_date_setUTCFullYear_31.setUTCFullYear(2030); console.log(d_date_setUTCFullYear_31.getUTCFullYear());

// @api: date.setUTCHours
// @expect: 18
const d_date_setUTCHours_32 = new Date(1577836800000); d_date_setUTCHours_32.setUTCHours(18); console.log(d_date_setUTCHours_32.getUTCHours());

// @api: date.setUTCMilliseconds
// @expect: 321
const d_date_setUTCMilliseconds_33 = new Date(1577836800000); d_date_setUTCMilliseconds_33.setUTCMilliseconds(321); console.log(d_date_setUTCMilliseconds_33.getUTCMilliseconds());

// @api: date.setUTCMinutes
// @expect: 35
const d_date_setUTCMinutes_34 = new Date(1577836800000); d_date_setUTCMinutes_34.setUTCMinutes(35); console.log(d_date_setUTCMinutes_34.getUTCMinutes());

// @api: date.setUTCMonth
// @expect: 10
const d_date_setUTCMonth_35 = new Date(1577836800000); d_date_setUTCMonth_35.setUTCMonth(10); console.log(d_date_setUTCMonth_35.getUTCMonth());

// @api: date.setUTCSeconds
// @expect: 50
const d_date_setUTCSeconds_36 = new Date(1577836800000); d_date_setUTCSeconds_36.setUTCSeconds(50); console.log(d_date_setUTCSeconds_36.getUTCSeconds());

// @api: date.toDateString
// @expect: string
const d_date_toDateString_37 = new Date(1577836800000); console.log(typeof d_date_toDateString_37.toDateString());

// @api: date.toISOString
// @expect: 1970-01-01T00:00:00.000Z
const d_date_toISOString_38 = new Date(0); console.log(d_date_toISOString_38.toISOString());

// @api: date.toJSON
// @expect: 2020-01-01T00:00:00.000Z
const d_date_toJSON_39 = new Date(1577836800000); console.log(d_date_toJSON_39.toJSON());

// @api: date.toLocaleDateString
// @expect: string
const d_date_toLocaleDateString_40 = new Date(1577836800000); console.log(typeof d_date_toLocaleDateString_40.toLocaleDateString());

// @api: date.toLocaleString
// @expect: string
const d_date_toLocaleString_41 = new Date(1577836800000); console.log(typeof d_date_toLocaleString_41.toLocaleString());

// @api: date.toLocaleTimeString
// @expect: string
const d_date_toLocaleTimeString_42 = new Date(1577836800000); console.log(typeof d_date_toLocaleTimeString_42.toLocaleTimeString());

// @api: date.toString
// @expect: true
const d_date_toString_43 = new Date(1700000000000);
console.log(d_date_toString_43.toString().length > 0);

// @api: date.toTimeString
// @expect: string
const d_date_toTimeString_45 = new Date(1577836800000); console.log(typeof d_date_toTimeString_45.toTimeString());

// @api: date.toUTCString
// @expect: string
const d_date_toUTCString_46 = new Date(1577836800000); console.log(typeof d_date_toUTCString_46.toUTCString());

// @api: date.valueOf
// @expect: 1577836800000
const d_date_valueOf_47 = new Date(1577836800000); console.log(d_date_valueOf_47.valueOf());
