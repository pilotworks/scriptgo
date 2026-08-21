const encoded: string = btoa("ABCDEFGHIJKLMNOP");
console.log(encoded);
const decoded: string = atob(encoded);
console.log(decoded);
