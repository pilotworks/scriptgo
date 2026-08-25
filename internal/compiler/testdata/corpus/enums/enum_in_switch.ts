// @expect: Red Color
// @expect: Green Color
// @expect: Blue Color
enum Color {
    Red,
    Green,
    Blue
}

function getColorName(c: Color): string {
    switch (c) {
        case Color.Red:
            return "Red Color";
        case Color.Green:
            return "Green Color";
        case Color.Blue:
            return "Blue Color";
        default:
            return "Other";
    }
}

console.log(getColorName(Color.Red));
console.log(getColorName(Color.Green));
console.log(getColorName(Color.Blue));
