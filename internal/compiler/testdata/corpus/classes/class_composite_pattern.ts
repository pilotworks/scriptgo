// @expect: Total size: 3050
// @expect: + [root] (3050B)
// @expect:   + [docs] (2550B)
// @expect:     - readme.md (150B)
// @expect:     - spec.pdf (2400B)
// @expect:   - logo.png (500B)
abstract class FileSystemComponent {
    protected name: string;

    constructor(name: string) {
        this.name = name;
    }

    abstract getSize(): number;
    abstract display(indent: string): string;
}

class FileItem extends FileSystemComponent {
    private size: number;

    constructor(name: string, size: number) {
        super(name);
        this.size = size;
    }

    getSize(): number {
        return this.size;
    }

    display(indent: string): string {
        return `${indent}- ${this.name} (${this.size}B)\n`;
    }
}

class DirectoryItem extends FileSystemComponent {
    private children: FileSystemComponent[] = [];

    constructor(name: string) {
        super(name);
    }

    add(component: FileSystemComponent): void {
        this.children.push(component);
    }

    getSize(): number {
        let total = 0;
        for (let i = 0; i < this.children.length; i++) {
            total += this.children[i].getSize();
        }
        return total;
    }

    display(indent: string): string {
        let out = `${indent}+ [${this.name}] (${this.getSize()}B)\n`;
        for (let i = 0; i < this.children.length; i++) {
            out += this.children[i].display(indent + "  ");
        }
        return out;
    }
}

const root = new DirectoryItem("root");
const docs = new DirectoryItem("docs");
const file1 = new FileItem("readme.md", 150);
const file2 = new FileItem("spec.pdf", 2400);
const file3 = new FileItem("logo.png", 500);

docs.add(file1);
docs.add(file2);
root.add(docs);
root.add(file3);

console.log(`Total size: ${root.getSize()}`);
console.log(root.display("").trim());
