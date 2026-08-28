// @expect: Dist to 0: 0
// @expect: Dist to 1: 7
// @expect: Dist to 2: 9
// @expect: Dist to 3: 20
// @expect: Dist to 4: 26
// @expect: Dist to 5: 11
type Edge = {
    to: number;
    weight: number;
};

type Graph = {
    numVertices: number;
    adj: Edge[][];
};

function createGraph(n: number): Graph {
    const adj: Edge[][] = [];
    for (let i = 0; i < n; i++) {
        adj.push([]);
    }
    return { numVertices: n, adj };
}

function addEdge(g: Graph, u: number, v: number, weight: number): void {
    g.adj[u].push({ to: v, weight });
}

function dijkstra(g: Graph, src: number): number[] {
    const dist: number[] = [];
    const visited: boolean[] = [];

    for (let i = 0; i < g.numVertices; i++) {
        dist.push(1e9);
        visited.push(false);
    }

    dist[src] = 0;

    for (let i = 0; i < g.numVertices; i++) {
        let u = -1;
        let minDist = 1e9;

        for (let v = 0; v < g.numVertices; v++) {
            if (!visited[v] && dist[v] < minDist) {
                minDist = dist[v];
                u = v;
            }
        }

        if (u === -1 || minDist === 1e9) {
            break;
        }

        visited[u] = true;

        const edges = g.adj[u];
        for (let j = 0; j < edges.length; j++) {
            const edge = edges[j];
            if (!visited[edge.to] && dist[u] + edge.weight < dist[edge.to]) {
                dist[edge.to] = dist[u] + edge.weight;
            }
        }
    }

    return dist;
}

const g = createGraph(6);
addEdge(g, 0, 1, 7);
addEdge(g, 0, 2, 9);
addEdge(g, 0, 5, 14);
addEdge(g, 1, 2, 10);
addEdge(g, 1, 3, 15);
addEdge(g, 2, 3, 11);
addEdge(g, 2, 5, 2);
addEdge(g, 3, 4, 6);
addEdge(g, 4, 5, 9);

const distances = dijkstra(g, 0);
for (let i = 0; i < distances.length; i++) {
    console.log("Dist to " + i + ": " + distances[i]);
}
