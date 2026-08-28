// @expect: 0 5 8 9
// @expect: INF 0 3 4
// @expect: INF INF 0 1
// @expect: INF INF INF 0
const INF = 999999;

function floydWarshall(graph: number[][]): number[][] {
    const v = graph.length;
    const dist: number[][] = [];

    for (let i = 0; i < v; i++) {
        dist[i] = [];
        for (let j = 0; j < v; j++) {
            dist[i][j] = graph[i][j];
        }
    }

    for (let k = 0; k < v; k++) {
        for (let i = 0; i < v; i++) {
            for (let j = 0; j < v; j++) {
                if (dist[i][k] + dist[k][j] < dist[i][j]) {
                    dist[i][j] = dist[i][k] + dist[k][j];
                }
            }
        }
    }

    return dist;
}

const g = [
    [0, 5, INF, 10],
    [INF, 0, 3, INF],
    [INF, INF, 0, 1],
    [INF, INF, INF, 0]
];

const res = floydWarshall(g);
for (let i = 0; i < res.length; i++) {
    let row = "";
    for (let j = 0; j < res[i].length; j++) {
        row += (res[i][j] >= INF ? "INF" : res[i][j].toString()) + " ";
    }
    console.log(row.trim());
}
