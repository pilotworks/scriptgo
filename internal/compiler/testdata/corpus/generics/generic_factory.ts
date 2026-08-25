// @expect: GenericWidget
// @expect: ButtonWidget
class Widget {
    name: string = "GenericWidget";
}

class ButtonWidget extends Widget {
    name: string = "ButtonWidget";
}

function createInstance<T>(creator: () => T): T {
    return creator();
}

const w = createInstance(() => new Widget());
const b = createInstance(() => new ButtonWidget());

console.log(w.name);
console.log(b.name);
