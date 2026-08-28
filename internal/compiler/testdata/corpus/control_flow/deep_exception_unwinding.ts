// @expect: L3_TRY;L2_TRY;L1_TRY;L1_POST;L1_FINALLY;L2_POST;L2_FINALLY;L3_POST;L3_FINALLY;
// @expect: L3_TRY;L2_TRY;L1_TRY;L1_CATCH: ERR_AT_L1;L1_FINALLY;L2_POST;L2_FINALLY;L3_POST;L3_FINALLY;
function complexUnwinding(level: number, shouldThrow: boolean): string {
    let trace = "";

    try {
        trace += `L${level}_TRY;`;
        if (level > 1) {
            trace += complexUnwinding(level - 1, shouldThrow);
        } else if (shouldThrow) {
            throw new Error(`ERR_AT_L${level}`);
        }
        trace += `L${level}_POST;`;
    } catch (err) {
        trace += `L${level}_CATCH: ${(err as Error).message};`;
        if (level === 2) {
            throw new Error(`RETHROWN_FROM_L2`);
        }
    } finally {
        trace += `L${level}_FINALLY;`;
    }

    return trace;
}

try {
    const res1 = complexUnwinding(3, false);
    console.log(res1);
} catch (e) {
    console.log(`OUTER_CATCH: ${(e as Error).message}`);
}

try {
    const res2 = complexUnwinding(3, true);
    console.log(res2);
} catch (e) {
    console.log(`OUTER_CATCH: ${(e as Error).message}`);
}
