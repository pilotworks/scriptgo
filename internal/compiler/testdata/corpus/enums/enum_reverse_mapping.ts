// @expect: 1
// @expect: Pending
// @expect: 2
// @expect: Active
// @expect: 3
// @expect: Completed
enum Status {
    Pending = 1,
    Active = 2,
    Completed = 3
}

console.log(Status.Pending);
console.log(Status[1]);
console.log(Status.Active);
console.log(Status[2]);
console.log(Status.Completed);
console.log(Status[3]);
