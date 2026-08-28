// @expect: Sorted: 11, 12, 22, 25, 34, 64, 90
// @expect: Sorted: 11, 12, 22, 25, 34, 64, 90
interface SortStrategy {
    sort(dataset: number[]): number[];
}

class BubbleSortStrategy implements SortStrategy {
    sort(dataset: number[]): number[] {
        const arr: number[] = [];
        for (let i = 0; i < dataset.length; i++) {
            arr.push(dataset[i]);
        }
        for (let i = 0; i < arr.length; i++) {
            for (let j = 0; j < arr.length - i - 1; j++) {
                if (arr[j] > arr[j + 1]) {
                    const temp = arr[j];
                    arr[j] = arr[j + 1];
                    arr[j + 1] = temp;
                }
            }
        }
        return arr;
    }
}

class InsertionSortStrategy implements SortStrategy {
    sort(dataset: number[]): number[] {
        const arr: number[] = [];
        for (let i = 0; i < dataset.length; i++) {
            arr.push(dataset[i]);
        }
        for (let i = 1; i < arr.length; i++) {
            const key = arr[i];
            let j = i - 1;
            while (j >= 0 && arr[j] > key) {
                arr[j + 1] = arr[j];
                j--;
            }
            arr[j + 1] = key;
        }
        return arr;
    }
}

class SorterContext {
    private strategy: SortStrategy;

    constructor(strategy: SortStrategy) {
        this.strategy = strategy;
    }

    setStrategy(strategy: SortStrategy): void {
        this.strategy = strategy;
    }

    sortData(data: number[]): void {
        const result = this.strategy.sort(data);
        let out = "";
        for (let i = 0; i < result.length; i++) {
            if (i > 0) out += ", ";
            out += result[i];
        }
        console.log("Sorted: " + out);
    }
}

const data = [64, 34, 25, 12, 22, 11, 90];
const sorter = new SorterContext(new BubbleSortStrategy());
sorter.sortData(data);

sorter.setStrategy(new InsertionSortStrategy());
sorter.sortData(data);
