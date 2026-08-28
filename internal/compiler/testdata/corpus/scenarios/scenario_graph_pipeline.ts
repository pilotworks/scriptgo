// @expect: Execution Order:
// @expect: - lint
// @expect: - build
// @expect: - integration_test
// @expect: - unit_test
// @expect: - deploy
interface Task {
    id: string;
    dependencies: string[];
    cost: number;
}

class PipelineRunner {
    private tasks: Map<string, Task> = new Map();

    addTask(task: Task): void {
        this.tasks.set(task.id, task);
    }

    getExecutionOrder(): string[] {
        const inDegree: Map<string, number> = new Map();
        const adj: Map<string, string[]> = new Map();

        for (const [id] of this.tasks) {
            inDegree.set(id, 0);
            adj.set(id, []);
        }

        for (const [id, task] of this.tasks) {
            for (const dep of task.dependencies) {
                adj.get(dep)!.push(id);
                inDegree.set(id, inDegree.get(id)! + 1);
            }
        }

        const queue: string[] = [];
        for (const [id, deg] of inDegree) {
            if (deg === 0) queue.push(id);
        }
        queue.sort();

        const order: string[] = [];
        while (queue.length > 0) {
            const curr = queue.shift()!;
            order.push(curr);

            for (const next of adj.get(curr)!) {
                inDegree.set(next, inDegree.get(next)! - 1);
                if (inDegree.get(next) === 0) {
                    queue.push(next);
                    queue.sort();
                }
            }
        }

        if (order.length !== this.tasks.size) {
            throw new Error("Cyclic dependency detected!");
        }

        return order;
    }
}

const pipeline = new PipelineRunner();
pipeline.addTask({ id: "lint", dependencies: [], cost: 10 });
pipeline.addTask({ id: "build", dependencies: ["lint"], cost: 30 });
pipeline.addTask({ id: "unit_test", dependencies: ["build"], cost: 20 });
pipeline.addTask({ id: "integration_test", dependencies: ["build"], cost: 50 });
pipeline.addTask({ id: "deploy", dependencies: ["unit_test", "integration_test"], cost: 100 });

const order = pipeline.getExecutionOrder();
console.log("Execution Order:");
for (const step of order) {
    console.log(`- ${step}`);
}
