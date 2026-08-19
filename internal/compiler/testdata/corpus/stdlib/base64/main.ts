const encoded: string = btoa("Hello, scriptgo!");
console.log(encoded);
const decoded: string = atob(encoded);
console.log(decoded);
