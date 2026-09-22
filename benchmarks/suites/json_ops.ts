// JSON Operations Benchmark: High-volume serialization, deserialization & roundtrip

interface UserProfile {
    id: number;
    username: string;
    email: string;
    active: boolean;
    score: number;
}

function run(): void {
    const recordCount = 50;
    const users: UserProfile[] = [];

    for (let i = 0; i < recordCount; i++) {
        users.push({
            id: 1000 + i,
            username: "user_" + i,
            email: "user_" + i + "@example.com",
            active: i % 2 === 0,
            score: (i * 13) % 100 + 0.5,
        });
    }

    const iterations = 100;
    let serializedBytes = 0;
    let roundtripBytes = 0;

    for (let it = 0; it < iterations; it++) {
        // 1. JSON.stringify on typed static shapes
        const jsonStr = JSON.stringify(users);
        serializedBytes += jsonStr.length;

        // 2. JSON.parse on payload and roundtrip stringify
        const parsed: unknown = JSON.parse(jsonStr);
        const roundtripStr = JSON.stringify(parsed);
        roundtripBytes += roundtripStr.length;
    }

    console.log(
        "JSON ops completed, records:",
        recordCount,
        "iterations:",
        iterations,
        "serializedBytes:",
        serializedBytes,
        "roundtripBytes:",
        roundtripBytes
    );
}

run();
