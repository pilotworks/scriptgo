// @expect: 3 1.2599210498948732 1.4422495703074083 0.1 10 -2 -3 4.641588833612774e-104 -1.7031839360032603e-108 4.641588833612779e+102 0.7937005259840998 1.9129311827723892 23.11204184662531 4.640041121316107e-67 4 5 0.6933612743506347 
// @expect: 339607982 0 -0 Infinity NaN 3
// Math.cbrt matches V8's fdlibm-derived result (the C library's cbrt may
// differ by an ulp, e.g. glibc gives 3.0000000000000004 for 27).
const values = [27, 2, 3, 0.001, 1000, -8, -27, 1e-310, -5e-324, 1e308, 0.5, 7, 12345.678, 9.99e-200, 64, 125, 1 / 3];
let line = "";
for (const v of values) line += Math.cbrt(Number(String(v))) + " ";
console.log(line);
let h = 0;
for (let i = 1; i < 20000; i++) {
  const v = i * 0.37 + i * i * 1e-3;
  const c = Math.cbrt(v);
  h = (h * 31 + Math.floor((c - Math.floor(c)) * 1e15)) % 1000000007;
}
console.log(h, Math.cbrt(0), Math.cbrt(-0), Math.cbrt(Infinity), Math.cbrt(NaN), Math.cbrt(27));
