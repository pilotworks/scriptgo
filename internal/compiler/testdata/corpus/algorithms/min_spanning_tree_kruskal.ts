// @expect: 31
// @expect: 4
class DisjointSetUnion {
    parent: number[];
    rank: number[];

    constructor(size: number) {
        this.parent = [];
        this.rank = [];
        for (let i = 0; i < size; i++) {
            this.parent.push(i);
            this.rank.push(0);
        }
    }

    find(i: number): number {
        if (this.parent[i] === i) {
            return i;
        }
        this.parent[i] = this.find(this.parent[i]);
        return this.parent[i];
    }

    union(i: number, j: number): boolean {
        const rootI = this.find(i);
        const rootJ = this.find(j);
        if (rootI === rootJ) {
            return false;
        }

        if (this.rank[rootI] < this.rank[rootJ]) {
            this.parent[rootI] = rootJ;
        } else if (this.rank[rootI] > this.rank[rootJ]) {
            this.parent[rootJ] = rootI;
        } else {
            this.parent[rootJ] = rootI;
            this.rank[rootI] = this.rank[rootI] + 1;
        }
        return true;
    }
}

interface Edge {
    u: number;
    v: number;
    weight: number;
}

function kruskalMST(numVertices: number, edges: Edge[]): [number, Edge[]] {
    // Sort edges by weight
    const sortedEdges = [...edges].sort((a, b) => a.weight - b.weight);
    const dsu = new DisjointSetUnion(numVertices);

    let totalWeight = 0;
    const mstEdges: Edge[] = [];

    for (let i = 0; i < sortedEdges.length; i++) {
        const edge = sortedEdges[i];
        if (dsu.union(edge.u, edge.v)) {
            totalWeight += edge.weight;
            mstEdges.push(edge);
            if (mstEdges.length === numVertices - 1) {
                break;
            }
        }
    }

    return [totalWeight, mstEdges];
}

const edges: Edge[] = [
    { u: 0, v: 1, weight: 10 },
    { u: 0, v: 2, weight: 6 },
    { u: 0, v: 3, weight: 5 },
    { u: 1, v: 3, weight: 15 },
    { u: 2, v: 3, weight: 4 },
    { u: 3, v: 4, weight: 18 },
    { u: 1, v: 4, weight: 12 }
];

const [minCost, mst] = kruskalMST(5, edges);
console.log(minCost);
console.log(mst.length);
