// @expect: ["A","B","C","D"]
// @expect: 4
class GenericGraph<T> {
    private adj: Map<T, T[]> = new Map();

    addVertex(v: T): void {
        if (!this.adj.has(v)) {
            this.adj.set(v, []);
        }
    }

    addEdge(u: T, v: T): void {
        this.addVertex(u);
        this.addVertex(v);
        const list = this.adj.get(u)!;
        list.push(v);
    }

    bfs(start: T): T[] {
        const visited = new Set<T>();
        const queue: T[] = [start];
        const result: T[] = [];
        visited.add(start);

        let head = 0;
        while (head < queue.length) {
            const vertex = queue[head];
            head++;
            result.push(vertex);

            const neighbors = this.adj.get(vertex) || [];
            for (let i = 0; i < neighbors.length; i++) {
                const neighbor = neighbors[i];
                if (!visited.has(neighbor)) {
                    visited.add(neighbor);
                    queue.push(neighbor);
                }
            }
        }
        return result;
    }

    vertexCount(): number {
        return this.adj.size;
    }
}

const graph = new GenericGraph<string>();
graph.addEdge("A", "B");
graph.addEdge("A", "C");
graph.addEdge("B", "D");
graph.addEdge("C", "D");

const bfsOrder = graph.bfs("A");
console.log(JSON.stringify(bfsOrder));
console.log(graph.vertexCount());
