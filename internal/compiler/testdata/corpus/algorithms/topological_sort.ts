// @expect: ["D","B","A","C","E"]
// @expect: true
function topologicalSort(numNodes: number, edges: number[][]): number[] {
    const inDegree: number[] = [];
    const adj: number[][] = [];
    for (let i = 0; i < numNodes; i++) {
        inDegree.push(0);
        adj.push([]);
    }

    for (let i = 0; i < edges.length; i++) {
        const u = edges[i][0];
        const v = edges[i][1];
        adj[u].push(v);
        inDegree[v] = inDegree[v] + 1;
    }

    const queue: number[] = [];
    for (let i = 0; i < numNodes; i++) {
        if (inDegree[i] === 0) {
            queue.push(i);
        }
    }

    const order: number[] = [];
    let head = 0;
    while (head < queue.length) {
        const u = queue[head];
        head++;
        order.push(u);

        const neighbors = adj[u];
        for (let i = 0; i < neighbors.length; i++) {
            const v = neighbors[i];
            inDegree[v] = inDegree[v] - 1;
            if (inDegree[v] === 0) {
                queue.push(v);
            }
        }
    }

    return order;
}

const nodeNames = ["A", "B", "C", "D", "E"];
// D(3) -> B(1) -> A(0) -> C(2) -> E(4)
const edges = [
    [3, 1],
    [1, 0],
    [0, 2],
    [2, 4],
    [3, 0]
];

const result = topologicalSort(5, edges);
const namedResult = result.map((idx) => nodeNames[idx]);
console.log(JSON.stringify(namedResult));
console.log(result.length === 5);
