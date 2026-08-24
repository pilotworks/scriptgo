/**
 * Graph & Pathfinding Algorithms Engine
 * 
 * Demonstrates:
 * - Generic Graph Data Structure with Weighted Edges
 * - Dijkstra's Shortest Path Algorithm
 * - A* Pathfinding on 2D Grid with Manhattan Heuristic
 * - Kruskal's Minimum Spanning Tree (MST) using Disjoint Set Union (DSU)
 * - Topological Sort (Kahn's Algorithm) for DAG dependency resolution
 * - Map, Set, generic classes, and priority queue implementations
 */

export class MinPriorityQueue<T> {
    private elements: { item: T; priority: number }[] = [];

    push(item: T, priority: number): void {
        this.elements.push({ item, priority });
        this.bubbleUp(this.elements.length - 1);
    }

    pop(): T | undefined {
        if (this.elements.length === 0) return undefined;
        const top = this.elements[0].item;
        const bottom = this.elements.pop()!;
        if (this.elements.length > 0) {
            this.elements[0] = bottom;
            this.sinkDown(0);
        }
        return top;
    }

    isEmpty(): boolean {
        return this.elements.length === 0;
    }

    size(): number {
        return this.elements.length;
    }

    private bubbleUp(index: number): void {
        let current = index;
        while (current > 0) {
            const parent = Math.floor((current - 1) / 2);
            if (this.elements[current].priority < this.elements[parent].priority) {
                const tmp = this.elements[current];
                this.elements[current] = this.elements[parent];
                this.elements[parent] = tmp;
                current = parent;
            } else {
                break;
            }
        }
    }

    private sinkDown(index: number): void {
        let current = index;
        const len = this.elements.length;
        while (true) {
            let left = 2 * current + 1;
            let right = 2 * current + 2;
            let smallest = current;

            if (left < len && this.elements[left].priority < this.elements[smallest].priority) {
                smallest = left;
            }
            if (right < len && this.elements[right].priority < this.elements[smallest].priority) {
                smallest = right;
            }

            if (smallest !== current) {
                const tmp = this.elements[current];
                this.elements[current] = this.elements[smallest];
                this.elements[smallest] = tmp;
                current = smallest;
            } else {
                break;
            }
        }
    }
}

export interface Edge<T> {
    from: T;
    to: T;
    weight: number;
}

export class Graph<T> {
    private adjacencyList: Map<T, Edge<T>[]> = new Map();

    addVertex(vertex: T): void {
        if (!this.adjacencyList.has(vertex)) {
            this.adjacencyList.set(vertex, []);
        }
    }

    addEdge(from: T, to: T, weight: number = 1, bidirectional: boolean = false): void {
        this.addVertex(from);
        this.addVertex(to);
        this.adjacencyList.get(from)!.push({ from, to, weight });
        if (bidirectional) {
            this.adjacencyList.get(to)!.push({ from: to, to: from, weight });
        }
    }

    getVertices(): T[] {
        return Array.from(this.adjacencyList.keys());
    }

    getNeighbors(vertex: T): Edge<T>[] {
        return this.adjacencyList.get(vertex) || [];
    }

    getAllEdges(): Edge<T>[] {
        const edges: Edge<T>[] = [];
        for (const list of this.adjacencyList.values()) {
            for (const edge of list) {
                edges.push(edge);
            }
        }
        return edges;
    }

    /**
     * Dijkstra's Single Source Shortest Path
     */
    dijkstra(start: T, end: T): { distance: number; path: T[] } | null {
        const distances: Map<T, number> = new Map();
        const previous: Map<T, T | null> = new Map();
        const pq = new MinPriorityQueue<T>();

        for (const vertex of this.adjacencyList.keys()) {
            distances.set(vertex, Infinity);
            previous.set(vertex, null);
        }

        distances.set(start, 0);
        pq.push(start, 0);

        while (!pq.isEmpty()) {
            const current = pq.pop()!;
            const currentDist = distances.get(current)!;

            if (current === end) {
                // Reconstruct path
                const path: T[] = [];
                let curr: T | null = end;
                while (curr !== null) {
                    path.unshift(curr);
                    curr = previous.get(curr) || null;
                }
                return { distance: currentDist, path };
            }

            for (const edge of this.getNeighbors(current)) {
                const alt = currentDist + edge.weight;
                if (alt < distances.get(edge.to)!) {
                    distances.set(edge.to, alt);
                    previous.set(edge.to, current);
                    pq.push(edge.to, alt);
                }
            }
        }

        return null; // unreachable
    }

    /**
     * Topological Sort using Kahn's Algorithm
     */
    topologicalSort(): T[] {
        const inDegree: Map<T, number> = new Map();
        for (const v of this.adjacencyList.keys()) {
            inDegree.set(v, 0);
        }

        for (const edges of this.adjacencyList.values()) {
            for (const edge of edges) {
                inDegree.set(edge.to, (inDegree.get(edge.to) || 0) + 1);
            }
        }

        const queue: T[] = [];
        for (const [v, deg] of inDegree.entries()) {
            if (deg === 0) {
                queue.push(v);
            }
        }

        const result: T[] = [];
        while (queue.length > 0) {
            const node = queue.shift()!;
            result.push(node);

            for (const edge of this.getNeighbors(node)) {
                const newDeg = inDegree.get(edge.to)! - 1;
                inDegree.set(edge.to, newDeg);
                if (newDeg === 0) {
                    queue.push(edge.to);
                }
            }
        }

        if (result.length !== this.adjacencyList.size) {
            throw new Error("Graph contains a cycle; topological sort not possible");
        }

        return result;
    }
}

// ==========================================
// Disjoint Set Union (DSU) for Kruskal's MST
// ==========================================

export class DisjointSet<T> {
    private parent: Map<T, T> = new Map();
    private rank: Map<T, number> = new Map();

    makeSet(item: T): void {
        this.parent.set(item, item);
        this.rank.set(item, 0);
    }

    find(item: T): T {
        if (this.parent.get(item) !== item) {
            this.parent.set(item, this.find(this.parent.get(item)!)); // Path compression
        }
        return this.parent.get(item)!;
    }

    union(x: T, y: T): boolean {
        const rootX = this.find(x);
        const rootY = this.find(y);

        if (rootX === rootY) return false;

        const rankX = this.rank.get(rootX)!;
        const rankY = this.rank.get(rootY)!;

        if (rankX < rankY) {
            this.parent.set(rootX, rootY);
        } else if (rankX > rankY) {
            this.parent.set(rootY, rootX);
        } else {
            this.parent.set(rootY, rootX);
            this.rank.set(rootX, rankX + 1);
        }

        return true;
    }
}

export function kruskalMST<T>(graph: Graph<T>): { mstEdges: Edge<T>[]; totalWeight: number } {
    const dsu = new DisjointSet<T>();
    const vertices = graph.getVertices();
    for (let i = 0; i < vertices.length; i++) {
        dsu.makeSet(vertices[i]);
    }

    const allEdges = graph.getAllEdges();
    // Sort edges by weight ascending
    allEdges.sort((a, b) => a.weight - b.weight);

    const mstEdges: Edge<T>[] = [];
    let totalWeight = 0;

    for (let i = 0; i < allEdges.length; i++) {
        const edge = allEdges[i];
        if (dsu.union(edge.from, edge.to)) {
            mstEdges.push(edge);
            totalWeight += edge.weight;
        }
    }

    return { mstEdges, totalWeight };
}

// ==========================================
// A* Grid Pathfinding
// ==========================================

export interface GridPoint {
    r: number;
    c: number;
}

export class GridAStar {
    private rows: number;
    private cols: number;
    private grid: number[][]; // 0: traversable, 1: obstacle

    constructor(grid: number[][]) {
        this.grid = grid;
        this.rows = grid.length;
        this.cols = grid[0].length;
    }

    private manhattan(a: GridPoint, b: GridPoint): number {
        return Math.abs(a.r - b.r) + Math.abs(a.c - b.c);
    }

    private pointKey(p: GridPoint): string {
        return `${p.r},${p.c}`;
    }

    findPath(start: GridPoint, goal: GridPoint): GridPoint[] | null {
        const openSet = new MinPriorityQueue<GridPoint>();
        const cameFrom: Map<string, GridPoint> = new Map();
        const gScore: Map<string, number> = new Map();
        const fScore: Map<string, number> = new Map();

        const startKey = this.pointKey(start);
        const goalKey = this.pointKey(goal);

        gScore.set(startKey, 0);
        fScore.set(startKey, this.manhattan(start, goal));
        openSet.push(start, fScore.get(startKey)!);

        const directions = [
            { r: -1, c: 0 },
            { r: 1, c: 0 },
            { r: 0, c: -1 },
            { r: 0, c: 1 }
        ];

        while (!openSet.isEmpty()) {
            const current = openSet.pop()!;
            const currentKey = this.pointKey(current);

            if (current.r === goal.r && current.c === goal.c) {
                // Reconstruct path
                const path: GridPoint[] = [current];
                let currKey = currentKey;
                while (cameFrom.has(currKey)) {
                    const prev = cameFrom.get(currKey)!;
                    path.unshift(prev);
                    currKey = this.pointKey(prev);
                }
                return path;
            }

            const currentG = gScore.get(currentKey)!;

            for (let i = 0; i < directions.length; i++) {
                const nr = current.r + directions[i].r;
                const nc = current.c + directions[i].c;

                if (nr < 0 || nr >= this.rows || nc < 0 || nc >= this.cols) continue;
                if (this.grid[nr][nc] === 1) continue; // Obstacle

                const neighbor: GridPoint = { r: nr, c: nc };
                const neighborKey = this.pointKey(neighbor);
                const tentativeG = currentG + 1;

                if (tentativeG < (gScore.get(neighborKey) ?? Infinity)) {
                    cameFrom.set(neighborKey, current);
                    gScore.set(neighborKey, tentativeG);
                    const f = tentativeG + this.manhattan(neighbor, goal);
                    fScore.set(neighborKey, f);
                    openSet.push(neighbor, f);
                }
            }
        }

        return null; // No path found
    }
}

// ==========================================
// Demonstration
// ==========================================

function main(): void {
    console.log("=== Graph & Pathfinding Algorithms Demo ===");

    // 1. Dijkstra Shortest Path
    const cityMap = new Graph<string>();
    cityMap.addEdge("A", "B", 4, true);
    cityMap.addEdge("A", "C", 2, true);
    cityMap.addEdge("B", "C", 1, true);
    cityMap.addEdge("B", "D", 5, true);
    cityMap.addEdge("C", "D", 8, true);
    cityMap.addEdge("C", "E", 10, true);
    cityMap.addEdge("D", "E", 2, true);
    cityMap.addEdge("D", "Z", 6, true);
    cityMap.addEdge("E", "Z", 3, true);

    const dijkstraResult = cityMap.dijkstra("A", "Z");
    if (dijkstraResult) {
        console.log(`Dijkstra Shortest Path A -> Z:`);
        console.log(`  Path: ${dijkstraResult.path.join(" -> ")}`);
        console.log(`  Distance: ${dijkstraResult.distance}`);
    }

    // 2. Kruskal Minimum Spanning Tree
    const mstResult = kruskalMST(cityMap);
    console.log(`\nKruskal Minimum Spanning Tree:`);
    console.log(`  Total MST Weight: ${mstResult.totalWeight}`);
    console.log(`  MST Edges:`);
    for (let i = 0; i < mstResult.mstEdges.length; i++) {
        const e = mstResult.mstEdges[i];
        console.log(`    ${e.from} - ${e.to} (w: ${e.weight})`);
    }

    // 3. Topological Sort (Task Build Pipeline)
    const taskGraph = new Graph<string>();
    taskGraph.addEdge("compile_frontend", "package_bundle");
    taskGraph.addEdge("compile_backend", "package_bundle");
    taskGraph.addEdge("fetch_dependencies", "compile_frontend");
    taskGraph.addEdge("fetch_dependencies", "compile_backend");
    taskGraph.addEdge("lint_code", "compile_frontend");
    taskGraph.addEdge("lint_code", "compile_backend");
    taskGraph.addEdge("package_bundle", "run_e2e_tests");
    taskGraph.addEdge("package_bundle", "docker_build");

    const buildOrder = taskGraph.topologicalSort();
    console.log(`\nTopological Sort (Build Pipeline):`);
    console.log(`  Execution Order: ${buildOrder.join(" -> ")}`);

    // 4. A* Grid Pathfinding
    const grid = [
        [0, 0, 0, 0, 0],
        [0, 1, 1, 1, 0],
        [0, 0, 0, 1, 0],
        [1, 1, 0, 1, 0],
        [0, 0, 0, 0, 0]
    ];

    const astar = new GridAStar(grid);
    const startPt: GridPoint = { r: 0, c: 0 };
    const goalPt: GridPoint = { r: 4, c: 4 };
    const path = astar.findPath(startPt, goalPt);

    console.log(`\nA* Path from (0,0) to (4,4):`);
    if (path) {
        console.log(`  Steps (${path.length}): ${path.map(p => `(${p.r},${p.c})`).join(" -> ")}`);
    } else {
        console.log(`  No path found!`);
    }
}

main();
