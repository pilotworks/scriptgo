// @expect: handle:73
// @expect: mode:true
// @expect: appended:9

interface PairOptions {
    socket?: unknown;
    rejectUnauthorized?: boolean;
    __tlsHandle?: number;
    __tlsPair?: boolean;
    __tlsMode?: number;
}

function describe(options: PairOptions): string {
    return "handle:" + options.__tlsHandle + "\nmode:" + options.__tlsPair;
}

const options: PairOptions = {
    __tlsHandle: 73,
    __tlsPair: true,
};

console.log(describe(options));
options.__tlsMode = 9;
console.log("appended:" + options.__tlsMode);
