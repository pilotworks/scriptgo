// @expect: 10.5
// @expect: 20.2
// @expect: 30.8
function getCoordinates(): [number, number, number] {
    return [10.5, 20.2, 30.8];
}

const [x, y, z] = getCoordinates();
console.log(x);
console.log(y);
console.log(z);
