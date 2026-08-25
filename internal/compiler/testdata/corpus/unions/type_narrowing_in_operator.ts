// @expect: Nemo is swimming
// @expect: Eagle is flying
type Fish = { swim: () => void; name: string };
type Bird = { fly: () => void; name: string };

function move(animal: Fish | Bird): string {
    if ("swim" in animal) {
        return animal.name + " is swimming";
    } else {
        return animal.name + " is flying";
    }
}

const f: Fish = { swim: () => {}, name: "Nemo" };
const b: Bird = { fly: () => {}, name: "Eagle" };

console.log(move(f));
console.log(move(b));
