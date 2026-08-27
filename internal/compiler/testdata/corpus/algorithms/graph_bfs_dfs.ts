// @expect: BFS: 0 1 2 3
// @expect: DFS: 0 1 2 3
class Graph {
    private adj: number[][] = [];

    constructor(numVertices: number) {
        for (let i = 0; i < numVertices; i++) {
            this.adj.push([]);
        }
    }

    addEdge(u: number, v: number): void {
        this.adj[u].push(v);
    }

    bfs(start: number): number[] {
        const visited: boolean[] = [];
        for (let i = 0; i < this.adj.length; i++) {
            visited.push(false);
        }

        const queue: number[] = [start];
        visited[start] = true;
        const result: number[] = [];

        while (queue.length > 0) {
            const current = queue.shift()!;
            result.push(current);

            for (const neighbor of this.adj[current]) {
                if (!visited[neighbor]) {
                    visited[neighbor] = true;
                    queue.push(neighbor);
                }
            }
        }
        return result;
    }

    dfs(start: number): number[] {
        const visited: boolean[] = [];
        for (let i = 0; i < this.adj.length; i++) {
            visited.push(false);
        }
        const result: number[] = [];
        this.dfsHelper(start, visited, result);
        return result;
    }

    private dfsHelper(u: number, visited: boolean[], result: number[]): void {
        visited[u] = true;
        result.push(u);
        for (const neighbor of this.adj[u]) {
            if (!visited[neighbor]) {
                this.dfsHelper(neighbor, visited, result);
            }
        }
    }
}

const g = new Graph(4);
g.addEdge(0, 1);
g.addEdge(0, 2);
g.addEdge(1, 2);
g.addEdge(2, 0);
g.addEdge(2, 3);
g.addEdge(3, 3);

console.log("BFS: " + g.bfs(0).join(" "));
console.log("DFS: " + g.dfs(0).join(" "));
