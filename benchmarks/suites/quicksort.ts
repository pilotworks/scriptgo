// Quicksort Benchmark: In-place sorting of 100,000 float64 elements

function partition(arr: number[], low: number, high: number): number {
    const pivot = arr[high];
    let i = low - 1;
    for (let j = low; j < high; j++) {
        if (arr[j] <= pivot) {
            i++;
            const tmp = arr[i];
            arr[i] = arr[j];
            arr[j] = tmp;
        }
    }
    const tmp = arr[i + 1];
    arr[i + 1] = arr[high];
    arr[high] = tmp;
    return i + 1;
}

function quicksort(arr: number[], low: number, high: number): void {
    if (low < high) {
        const pi = partition(arr, low, high);
        quicksort(arr, low, pi - 1);
        quicksort(arr, pi + 1, high);
    }
}

function run(): void {
    const n = 100000;
    const arr: number[] = [];
    // Deterministic LCG random sequence
    let seed = 123456789;
    for (let i = 0; i < n; i++) {
        seed = (seed * 1103515245 + 12345) & 0x7fffffff;
        arr.push(seed % 1000000);
    }

    quicksort(arr, 0, arr.length - 1);

    // Sanity check
    for (let i = 1; i < arr.length; i++) {
        if (arr[i] < arr[i - 1]) {
            throw new Error("Sorting failed at index " + i);
        }
    }
    console.log("Sorted", arr.length, "elements successfully.");
}

run();
