function add(x: number, y: number): number {
    return x + y;
}

let a = 1;
let b = a++;
console.log(b);
console.log(a);

let c = ++a;
console.log(c);
console.log(a);

let d = a--;
console.log(d);
console.log(a);

let e = --a;
console.log(e);
console.log(a);

let arr = [10, 20, 30];
let i = 0;
console.log(arr[i++]);
console.log(i);
console.log(arr[0]++);
console.log(arr[0]);
console.log(--arr[0]);
console.log(arr[0]);

let obj = { count: 5 };
console.log(obj.count++);
console.log(obj.count);
console.log(++obj.count);
console.log(obj.count);
console.log(obj.count--);
console.log(obj.count);
console.log(--obj.count);
console.log(obj.count);

let n = 2;
console.log(add(n++, ++n));
console.log(n);

let loopCount = 3;
while (loopCount-- > 0) {
    console.log(loopCount);
}

for (let k = 0; k < 3; k++) {
    console.log(k);
}
