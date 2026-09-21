// Mandelbrot Benchmark: Intense floating-point computation, escape-time algorithm, bitwise checksum

function mandelbrot(): number {
    const width = 500;
    const height = 500;
    const maxIter = 100;
    let checksum = 0;

    for (let y = 0; y < height; y++) {
        const ci = (y * 2.0) / height - 1.0;
        for (let x = 0; x < width; x++) {
            const cr = (x * 3.0) / width - 2.0;
            let zr = 0.0;
            let zi = 0.0;
            let iter = 0;

            while (iter < maxIter && (zr * zr + zi * zi) <= 4.0) {
                const zrNew = zr * zr - zi * zi + cr;
                zi = 2.0 * zr * zi + ci;
                zr = zrNew;
                iter++;
            }

            checksum = (checksum + iter) & 0x7fffffff;
        }
    }

    return checksum;
}

function run(): void {
    const result = mandelbrot();
    console.log("Mandelbrot completed, checksum:", result);
}

run();
