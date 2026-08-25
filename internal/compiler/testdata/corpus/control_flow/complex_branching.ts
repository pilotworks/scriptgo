// @expect: Allowed
// @expect: Need Guardian
// @expect: Allowed
// @expect: Denied
function checkAccess(age: number, hasTicket: boolean, isVip: boolean): string {
    if (isVip || (age >= 18 && hasTicket)) {
        return "Allowed";
    } else if (age < 18 && hasTicket) {
        return "Need Guardian";
    } else {
        return "Denied";
    }
}

console.log(checkAccess(25, true, false));
console.log(checkAccess(16, true, false));
console.log(checkAccess(16, false, true));
console.log(checkAccess(20, false, false));
