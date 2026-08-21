function createFormatter(prefix: string, suffix: string, scale: number, uppercase: boolean): (val: number) => string {
  return (val: number): string => {
    const scaled = val * scale;
    let res = prefix + scaled + suffix;
    if (uppercase) {
      res = res.toUpperCase();
    }
    return res;
  };
}

const fmtUSD = createFormatter("USD $", " total", 1, false);
const fmtEUR = createFormatter("EUR ", " net", 0.9, false);
const fmtJPY = createFormatter("jpy ", " yen", 150, true);

console.log(fmtUSD(100));
console.log(fmtUSD(250.5));
console.log(fmtEUR(100));
console.log(fmtJPY(10));
