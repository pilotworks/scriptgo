// @expect: 10,20
// @expect: 30,40
type Config = {
    point: {
        x: number;
        y: number;
    };
};

const draw = ({ point: { x, y } }: Config) => {
    console.log(x + "," + y);
};

draw({ point: { x: 10, y: 20 } });
draw({ point: { x: 30, y: 40 } });
