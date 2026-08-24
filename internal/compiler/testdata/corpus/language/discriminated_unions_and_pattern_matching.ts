// @expect: === Discriminated Unions & Pattern Matching Test ===
// @expect: Circle(r=5, color=red, area=78.54)
// @expect: Rectangle(w=10, h=4, color=blue, area=40.00)
// @expect: Triangle(b=6, h=8, color=green, area=24.00)
// @expect: Polygon(numVertices=4, color=purple, area=12.00)
// @expect: 
// @expect: Testing Result Pattern:
// @expect: 100 / 4: OK: 25.0000
// @expect: 100 / 0: ERR: Division by zero is not allowed
// @expect: 
// @expect: Testing Nested Destructuring with Defaults:
// @expect: Config 1: Host=127.0.0.1:8080, SSL=true, Cert=/etc/ssl/custom.crt, Ciphers=[AES256,CHACHA20], MaxConn=1000, Timeout=5000ms
// @expect: Config 2: Host=0.0.0.0:443, SSL=false, Cert=default.crt, Ciphers=[AES256,CHACHA20], MaxConn=50000, Timeout=30000ms

// Discriminated Unions, Tagged Variants, Narrowing & Destructuring

type Shape =
    | { kind: "circle"; radius: number; color: string }
    | { kind: "rectangle"; width: number; height: number; color: string }
    | { kind: "triangle"; base: number; height: number; color: string }
    | { kind: "polygon"; vertices: number[][]; color: string };

function calculateArea(shape: Shape): number {
    switch (shape.kind) {
        case "circle":
            return Math.PI * shape.radius * shape.radius;
        case "rectangle":
            return shape.width * shape.height;
        case "triangle":
            return 0.5 * shape.base * shape.height;
        case "polygon": {
            // Shoelace formula for polygon area
            const pts = shape.vertices;
            let sum = 0;
            const n = pts.length;
            for (let i = 0; i < n; i++) {
                const j = (i + 1) % n;
                sum += pts[i][0] * pts[j][1];
                sum -= pts[j][0] * pts[i][1];
            }
            return Math.abs(sum) * 0.5;
        }
    }
}

function describeShape(shape: Shape): string {
    const area = calculateArea(shape);
    switch (shape.kind) {
        case "circle": {
            const { radius, color } = shape;
            return `Circle(r=${radius}, color=${color}, area=${area.toFixed(2)})`;
        }
        case "rectangle": {
            const { width, height, color } = shape;
            return `Rectangle(w=${width}, h=${height}, color=${color}, area=${area.toFixed(2)})`;
        }
        case "triangle": {
            const { base, height, color } = shape;
            return `Triangle(b=${base}, h=${height}, color=${color}, area=${area.toFixed(2)})`;
        }
        case "polygon": {
            const { vertices, color } = shape;
            return `Polygon(numVertices=${vertices.length}, color=${color}, area=${area.toFixed(2)})`;
        }
    }
}

// Result / Either Pattern with Tagged Union
type Result<T, E> =
    | { success: true; data: T }
    | { success: false; error: E };

function divideSafe(a: number, b: number): Result<number, string> {
    if (b === 0) {
        return { success: false, error: "Division by zero is not allowed" };
    }
    return { success: true, data: a / b };
}

function processResult(res: Result<number, string>): string {
    if (res.success) {
        return `OK: ${res.data.toFixed(4)}`;
    } else {
        return `ERR: ${res.error}`;
    }
}

// Complex Nested Destructuring with Default Fallbacks
interface ServerConfig {
    host: string;
    port: number;
    security: {
        ssl: boolean;
        certPath?: string;
        ciphers?: string[];
    };
    limits?: {
        maxConnections?: number;
        timeoutMs?: number;
    };
}

function parseConfig(cfg: ServerConfig): string {
    const {
        host,
        port,
        security: { ssl, certPath = "default.crt", ciphers = ["AES256", "CHACHA20"] },
        limits: { maxConnections = 1000, timeoutMs = 5000 } = {}
    } = cfg;

    return `Host=${host}:${port}, SSL=${ssl}, Cert=${certPath}, Ciphers=[${ciphers.join(",")}], MaxConn=${maxConnections}, Timeout=${timeoutMs}ms`;
}

function main(): void {
    console.log("=== Discriminated Unions & Pattern Matching Test ===");

    const shapes: Shape[] = [
        { kind: "circle", radius: 5, color: "red" },
        { kind: "rectangle", width: 10, height: 4, color: "blue" },
        { kind: "triangle", base: 6, height: 8, color: "green" },
        { kind: "polygon", vertices: [[0, 0], [4, 0], [4, 3], [0, 3]], color: "purple" }
    ];

    for (let i = 0; i < shapes.length; i++) {
        console.log(describeShape(shapes[i]));
    }

    console.log("\nTesting Result Pattern:");
    console.log("100 / 4:", processResult(divideSafe(100, 4)));
    console.log("100 / 0:", processResult(divideSafe(100, 0)));

    console.log("\nTesting Nested Destructuring with Defaults:");
    const config1: ServerConfig = {
        host: "127.0.0.1",
        port: 8080,
        security: {
            ssl: true,
            certPath: "/etc/ssl/custom.crt"
        }
    };
    console.log("Config 1:", parseConfig(config1));

    const config2: ServerConfig = {
        host: "0.0.0.0",
        port: 443,
        security: {
            ssl: false
        },
        limits: {
            maxConnections: 50000,
            timeoutMs: 30000
        }
    };
    console.log("Config 2:", parseConfig(config2));
}

main();
