// @expect: Main Window
// @expect: 800
// @expect: 600
interface WindowSettings {
    title: string;
}

interface WindowSettings {
    width: number;
    height: number;
}

const win: WindowSettings = {
    title: "Main Window",
    width: 800,
    height: 600
};

console.log(win.title);
console.log(win.width);
console.log(win.height);
