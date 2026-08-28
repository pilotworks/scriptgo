// @expect: Found 3 SCCs
// @expect: 4
// @expect: 3
// @expect: 0 1 2
class TarjanSCC {
    private v: number;
    private adj: number[][];
    private index: number;
    private indices: number[];
    private lowlink: number[];
    private onStack: boolean[];
    private stack: number[];
    private sccs: number[][];

    constructor(v: number) {
        this.v = v;
        this.adj = [];
        this.index = 0;
        this.indices = [];
        this.lowlink = [];
        this.onStack = [];
        this.stack = [];
        this.sccs = [];

        for (let i = 0; i < v; i++) {
            this.adj.push([]);
            this.indices.push(-1);
            this.lowlink.push(-1);
            this.onStack.push(false);
        }
    }

    addEdge(u: number, w: number): void {
        this.adj[u].push(w);
    }

    findSCCs(): number[][] {
        for (let i = 0; i < this.v; i++) {
            if (this.indices[i] === -1) {
                this.strongConnect(i);
            }
        }
        return this.sccs;
    }

    private strongConnect(u: number): void {
        this.indices[u] = this.index;
        this.lowlink[u] = this.index;
        this.index++;
        this.stack.push(u);
        this.onStack[u] = true;

        for (let i = 0; i < this.adj[u].length; i++) {
            const v = this.adj[u][i];
            if (this.indices[v] === -1) {
                this.strongConnect(v);
                this.lowlink[u] = Math.min(this.lowlink[u], this.lowlink[v]);
            } else if (this.onStack[v]) {
                this.lowlink[u] = Math.min(this.lowlink[u], this.indices[v]);
            }
        }

        if (this.lowlink[u] === this.indices[u]) {
            const scc: number[] = [];
            while (true) {
                const w = this.stack.pop()!;
                this.onStack[w] = false;
                scc.push(w);
                if (w === u) break;
            }
            scc.sort((a, b) => a - b);
            this.sccs.push(scc);
        }
    }
}

const g = new TarjanSCC(5);
g.addEdge(1, 0);
g.addEdge(0, 2);
g.addEdge(2, 1);
g.addEdge(0, 3);
g.addEdge(3, 4);

const sccs = g.findSCCs();
console.log(`Found ${sccs.length} SCCs`);
for (const scc of sccs) {
    console.log(scc.join(" "));
}
