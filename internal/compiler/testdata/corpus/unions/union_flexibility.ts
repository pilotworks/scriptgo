// @expect: TRUTHY: 100
// @expect: TRUTHY: hello
// @expect: FALSY: 0
// @expect: FALSY: 
// @expect: GOT_NULL
// @expect: STRING_UPPER: SCRIPTGO
// @expect: NUM_DOUBLED: 100
// @expect: BOOL_VAL: true

function checkTruthiness(condition: number | string | boolean | null) {
    if (condition) {
        console.log("TRUTHY:", condition);
    } else {
        if (condition === null) {
            console.log("GOT_NULL");
        } else {
            console.log("FALSY:", condition);
        }
    }
}

function processNarrowed(val: number | string | boolean) {
    if (typeof val === "string") {
        console.log("STRING_UPPER:", val.toUpperCase());
    } else if (typeof val === "number") {
        console.log("NUM_DOUBLED:", val * 2);
    } else {
        console.log("BOOL_VAL:", val);
    }
}

function forwardBroad(val: number | string | boolean | null) {
    if (val !== null) {
        processNarrowed(val);
    }
}

checkTruthiness(100);
checkTruthiness("hello");
checkTruthiness(0);
checkTruthiness("");
checkTruthiness(null);

forwardBroad("scriptgo");
forwardBroad(50);
forwardBroad(true);
