// @expect: 25
// @expect: 77
// @expect: 100
class Temperature {
    private _celsius: number;

    constructor(c: number) {
        this._celsius = c;
    }

    get celsius(): number {
        return this._celsius;
    }

    set celsius(c: number) {
        this._celsius = c;
    }

    get fahrenheit(): number {
        return this._celsius * 1.8 + 32;
    }

    set fahrenheit(f: number) {
        this._celsius = (f - 32) / 1.8;
    }
}

const t = new Temperature(25);
console.log(t.celsius);
console.log(t.fahrenheit);
t.fahrenheit = 212;
console.log(t.celsius);
