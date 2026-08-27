// @expect: 1:odd
// @expect: 2:even
// @expect: 3:odd
// @expect: 4:special
// @expect: loop ended at 4
function processItems(): void {
    let count = 0;
    for (let i = 1; i <= 5; i++) {
        let label = "";
        switch (i) {
            case 1:
            case 3:
                label = "odd";
                break;
            case 2:
                label = "even";
                break;
            case 4:
                label = "special";
                break;
            default:
                label = "other";
                break;
        }

        console.log(i + ":" + label);

        if (i === 4) {
            count = i;
            break;
        }
    }
    console.log("loop ended at " + count);
}

processItems();
