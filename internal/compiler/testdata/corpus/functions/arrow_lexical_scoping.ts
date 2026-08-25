// @expect: 3
class Timer {
    seconds: number = 0;

    start() {
        const tick = () => {
            this.seconds++;
        };
        tick();
        tick();
        tick();
    }
}

const timer = new Timer();
timer.start();
console.log(timer.seconds);
