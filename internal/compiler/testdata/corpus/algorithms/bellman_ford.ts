// @expect: No negative cycle
// @expect: Dist 0->0 = 0
// @expect: Dist 0->1 = -1
// @expect: Dist 0->2 = 2
// @expect: Dist 0->3 = -2
// @expect: Dist 0->4 = 1
type BellmanEdge = {
    u: number;
    v: number;
    w: number;
};

function bellmanFord(numVertices: number, edges: BellmanEdge[], src: number): number[] {
    const dist: number[] = [];
    const INF = 1e9;

    for (let i = 0; i < numVertices; i++) {
        dist.push(INF);
    }
    dist[src] = 0;

    for (let i = 1; i <= numVertices - 1; i++) {
        for (let j = 0; j < edges.length; j++) {
            const edge = edges[j];
            if (dist[edge.u] !== INF && dist[edge.u] + edge.w < dist[edge.v]) {
                dist[edge.v] = dist[edge.u] + edge.w;
            }
        }
    }

    // Check negative cycle
    let hasNegCycle = false;
    for (let j = 0; j < edges.length; j++) {
        const edge = edges[j];
        if (dist[edge.u] !== INF && dist[edge.u] + edge.w < dist[edge.v]) {
            hasNegCycle = true;
            break;
        }
    }

    if (hasNegCycle) {
        console.log("Negative cycle detected");
    } else {
        console.log("No negative cycle");
    }

    return dist;
}

const edges: BellmanEdge[] = [
    { u: 0, v: 1, w: -1 },
    { u: 0, v: 2, w: 4 },
    { u: 1, v: 2, w: 3 },
    { u: 1, v: 3, w: 2 },
    { u: 1, v: 4, w: 2 },
    { u: 3, v: 2, w: 5 },
    { u: 3, v: 1, w: 1 },
    { u: 4, v: 3, w: -3 }
];

const res = bellmanFord(5, edges, 0);
for (let i = 0; i < res.length; i++) {
    console.log("Dist 0->" + i + " = " + res[i]);
}
