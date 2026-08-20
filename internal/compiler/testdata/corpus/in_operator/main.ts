class Item {
    id: number;
    title: string;
    constructor(id: number, title: string) {
        this.id = id;
        this.title = title;
    }
}

let user = { name: "Alice", age: 30 };
console.log("name" in user);
console.log("age" in user);
console.log("email" in user);

let k1 = "name";
console.log(k1 in user);

let k2 = "score";
console.log(k2 in user);

let list = [10, 20, 30];
console.log(0 in list);
console.log(2 in list);
console.log(3 in list);
console.log(-1 in list);

let item = new Item(1, "Book");
console.log("id" in item);
console.log("title" in item);
console.log("price" in item);
