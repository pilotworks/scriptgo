// @expect: 10
// @expect: 20
// @expect: 30
// @expect: 40
// @expect: 50
class MinHeap {
    private heap: number[] = [];

    push(val: number): void {
        this.heap.push(val);
        this.bubbleUp(this.heap.length - 1);
    }

    pop(): number | undefined {
        if (this.heap.length === 0) return undefined;
        const top = this.heap[0];
        const bottom = this.heap.pop()!;
        if (this.heap.length > 0) {
            this.heap[0] = bottom;
            this.bubbleDown(0);
        }
        return top;
    }

    peek(): number | undefined {
        return this.heap[0];
    }

    size(): number {
        return this.heap.length;
    }

    private bubbleUp(index: number): void {
        while (index > 0) {
            const parent = Math.floor((index - 1) / 2);
            if (this.heap[parent] <= this.heap[index]) break;
            const tmp = this.heap[parent];
            this.heap[parent] = this.heap[index];
            this.heap[index] = tmp;
            index = parent;
        }
    }

    private bubbleDown(index: number): void {
        while (true) {
            let smallest = index;
            const left = 2 * index + 1;
            const right = 2 * index + 2;

            if (left < this.heap.length && this.heap[left] < this.heap[smallest]) {
                smallest = left;
            }
            if (right < this.heap.length && this.heap[right] < this.heap[smallest]) {
                smallest = right;
            }
            if (smallest === index) break;

            const tmp = this.heap[index];
            this.heap[index] = this.heap[smallest];
            this.heap[smallest] = tmp;
            index = smallest;
        }
    }
}

const pq = new MinHeap();
pq.push(30);
pq.push(10);
pq.push(50);
pq.push(20);
pq.push(40);

while (pq.size() > 0) {
    console.log(pq.pop());
}
