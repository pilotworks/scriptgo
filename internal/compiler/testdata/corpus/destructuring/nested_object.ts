// @expect: Alice
// @expect: 30
// @expect: CityA
// @expect: 12345
const user = {
    name: "Alice",
    age: 30,
    address: {
        city: "CityA",
        zip: 12345,
    },
};

const { name, age, address: { city, zip } } = user;
console.log(name);
console.log(age);
console.log(city);
console.log(zip);
