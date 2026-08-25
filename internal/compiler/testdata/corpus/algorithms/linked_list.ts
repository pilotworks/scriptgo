// @expect: 1->2->3->4
// @expect: 4->3->2->1
class ListNode<T> {
    val: T;
    next: ListNode<T> | null = null;
    constructor(val: T) {
        this.val = val;
    }
}

class LinkedList<T> {
    head: ListNode<T> | null = null;
    size: number = 0;

    append(val: T): void {
        const node = new ListNode(val);
        if (!this.head) {
            this.head = node;
        } else {
            let curr = this.head;
            while (curr.next) {
                curr = curr.next;
            }
            curr.next = node;
        }
        this.size++;
    }

    toArray(): T[] {
        const res: T[] = [];
        let curr = this.head;
        while (curr) {
            res.push(curr.val);
            curr = curr.next;
        }
        return res;
    }

    reverse(): void {
        let prev: ListNode<T> | null = null;
        let curr = this.head;
        while (curr) {
            const next = curr.next;
            curr.next = prev;
            prev = curr;
            curr = next;
        }
        this.head = prev;
    }
}

const list = new LinkedList<number>();
list.append(1);
list.append(2);
list.append(3);
list.append(4);

console.log(list.toArray().join("->"));
list.reverse();
console.log(list.toArray().join("->"));
