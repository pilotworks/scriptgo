// @expect: 0
// @expect: 2
// @expect: 4
// @expect: 6
// @expect: 12
enum FileAccess {
    None = 0,
    Read = 1 << 1,
    Write = 1 << 2,
    ReadWrite = Read | Write,
    Custom = Read + 10
}

console.log(FileAccess.None);
console.log(FileAccess.Read);
console.log(FileAccess.Write);
console.log(FileAccess.ReadWrite);
console.log(FileAccess.Custom);
