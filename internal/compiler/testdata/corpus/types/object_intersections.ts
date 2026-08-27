// @expect: ID: 101
// @expect: Name: Widget
// @expect: Price: 29.99
// @expect: In Stock: true
type Identifiable = { id: number };
type Named = { name: string };
type Priced = { price: number; inStock: boolean };

type Product = Identifiable & Named & Priced;

const item: Product = {
    id: 101,
    name: "Widget",
    price: 29.99,
    inStock: true
};

console.log("ID: " + item.id);
console.log("Name: " + item.name);
console.log("Price: " + item.price);
console.log("In Stock: " + item.inStock);
