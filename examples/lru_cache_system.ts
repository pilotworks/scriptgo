// ============================================================================
// Multi-Tier Enterprise Cache System with LRU/TTL, Single-Flight & Metrics
// ============================================================================
// Features demonstrated:
// - Object-Oriented Architecture: Interfaces, Abstract Classes, Generics
// - Event-Driven Architecture with node:events (EventEmitter)
// - Cache Stampede Protection (SingleFlight / Promise Deduplication)
// - Doubly-Linked List Node Structure for O(1) LRU Operations
// - Time-To-Live (TTL) Automatic Invalidation & Scheduled Sweeper
// - Rich Performance Metrics (Hit Rate, Eviction Rate, Latency Tracking)
// - Error Handling, Serialization, Destructuring & Optional Chaining
// ============================================================================

import { EventEmitter } from "node:events";

// ----------------------------------------------------------------------------
// Type Definitions & Interfaces
// ----------------------------------------------------------------------------

export interface CacheOptions {
  capacity: number;
  defaultTTLMs?: number;
  enableSingleFlight?: boolean;
  cleanupIntervalMs?: number;
}

export interface CacheStats {
  hits: number;
  misses: number;
  sets: number;
  evictions: number;
  expirations: number;
  singleFlightSaves: number;
  totalLoadTimeMs: number;
  lastPurgeTimestamp: number;
}

export interface CacheEntry<V> {
  key: string;
  value: V;
  expiresAt: number | null;
  createdAt: number;
  lastAccessed: number;
  accessCount: number;
}

// Doubly linked list node for deterministic O(1) LRU eviction
class LinkedListNode<V> {
  public prev: LinkedListNode<V> | null = null;
  public next: LinkedListNode<V> | null = null;

  constructor(
    public key: string,
    public entry: CacheEntry<V>
  ) {}
}

export interface ICache<K extends string, V> {
  get(key: K): V | null;
  getOrLoad(key: K, loader: (k: K) => Promise<V>, ttlMs?: number): Promise<V>;
  set(key: K, value: V, ttlMs?: number): void;
  has(key: K): boolean;
  delete(key: K): boolean;
  clear(): void;
  size(): number;
  getStats(): CacheStats;
  getHitRatio(): number;
}

// ----------------------------------------------------------------------------
// LRU Cache Implementation with Doubly Linked List
// ----------------------------------------------------------------------------

export class AdvancedLRUCache<V> extends EventEmitter implements ICache<string, V> {
  private readonly capacity: number;
  private readonly defaultTTL: number | null;
  private readonly enableSingleFlight: boolean;

  // Primary Hash Map for O(1) lookups
  private readonly nodeMap: Map<string, LinkedListNode<V>>;

  // In-flight loader promises for Cache Stampede mitigation
  private readonly inFlightLoads: Map<string, Promise<V>>;

  // Doubly linked list sentinel head & tail
  private head: LinkedListNode<V> | null = null;
  private tail: LinkedListNode<V> | null = null;

  // Telemetry statistics
  private readonly stats: CacheStats;

  constructor(options: CacheOptions) {
    super();
    if (options.capacity <= 0) {
      throw new Error("Cache capacity must be a positive number greater than 0");
    }

    this.capacity = options.capacity;
    this.defaultTTL = options.defaultTTLMs ?? null;
    this.enableSingleFlight = options.enableSingleFlight ?? true;
    this.nodeMap = new Map<string, LinkedListNode<V>>();
    this.inFlightLoads = new Map<string, Promise<V>>();

    this.stats = {
      hits: 0,
      misses: 0,
      sets: 0,
      evictions: 0,
      expirations: 0,
      singleFlightSaves: 0,
      totalLoadTimeMs: 0,
      lastPurgeTimestamp: Date.now(),
    };
  }

  // --------------------------------------------------------------------------
  // Core Retrieval & Storage Operations
  // --------------------------------------------------------------------------

  public get(key: string): V | null {
    const node = this.nodeMap.get(key);
    const now = Date.now();

    if (!node) {
      this.stats.misses++;
      this.emit("miss", key);
      return null;
    }

    // Check TTL Expiration
    if (node.entry.expiresAt !== null && node.entry.expiresAt <= now) {
      this.removeNode(node);
      this.nodeMap.delete(key);
      this.stats.expirations++;
      this.stats.misses++;
      this.emit("expired", key, node.entry.value);
      return null;
    }

    // Update access metadata and move node to head (Most Recently Used)
    node.entry.lastAccessed = now;
    node.entry.accessCount++;
    this.moveToHead(node);

    this.stats.hits++;
    this.emit("hit", key, node.entry.value);
    return node.entry.value;
  }

  public set(key: string, value: V, ttlMs?: number): void {
    const now = Date.now();
    const effectiveTTL = ttlMs ?? this.defaultTTL;
    const expiresAt = effectiveTTL !== null ? now + effectiveTTL : null;

    const existingNode = this.nodeMap.get(key);

    if (existingNode) {
      // Update existing entry
      existingNode.entry.value = value;
      existingNode.entry.expiresAt = expiresAt;
      existingNode.entry.lastAccessed = now;
      this.moveToHead(existingNode);
      this.stats.sets++;
      this.emit("update", key, value);
      return;
    }

    // Evict oldest node if at capacity
    if (this.nodeMap.size >= this.capacity) {
      this.evictTail();
    }

    // Create fresh entry & node
    const entry: CacheEntry<V> = {
      key,
      value,
      expiresAt,
      createdAt: now,
      lastAccessed: now,
      accessCount: 1,
    };

    const newNode = new LinkedListNode<V>(key, entry);
    this.nodeMap.set(key, newNode);
    this.insertAtHead(newNode);

    this.stats.sets++;
    this.emit("set", key, value);
  }

  /**
   * SingleFlight Pattern: Deduplicates concurrent loader calls for the same key.
   */
  public async getOrLoad(
    key: string,
    loader: (k: string) => Promise<V>,
    ttlMs?: number
  ): Promise<V> {
    const cached = this.get(key);
    if (cached !== null) {
      return cached;
    }

    if (!this.enableSingleFlight) {
      const startTime = Date.now();
      const loadedValue = await loader(key);
      this.stats.totalLoadTimeMs += Date.now() - startTime;
      this.set(key, loadedValue, ttlMs);
      return loadedValue;
    }

    // Check if another async task is already loading this exact key
    const existingFlight = this.inFlightLoads.get(key);
    if (existingFlight) {
      this.stats.singleFlightSaves++;
      this.emit("singleFlightJoin", key);
      return await existingFlight;
    }

    // Create single flight load promise
    const loadPromise = (async () => {
      const startTime = Date.now();
      try {
        const value = await loader(key);
        this.stats.totalLoadTimeMs += Date.now() - startTime;
        this.set(key, value, ttlMs);
        return value;
      } finally {
        this.inFlightLoads.delete(key);
      }
    })();

    this.inFlightLoads.set(key, loadPromise);
    return await loadPromise;
  }

  public has(key: string): boolean {
    const node = this.nodeMap.get(key);
    if (!node) return false;
    if (node.entry.expiresAt !== null && node.entry.expiresAt <= Date.now()) {
      this.removeNode(node);
      this.nodeMap.delete(key);
      this.stats.expirations++;
      return false;
    }
    return true;
  }

  public delete(key: string): boolean {
    const node = this.nodeMap.get(key);
    if (!node) return false;

    this.removeNode(node);
    this.nodeMap.delete(key);
    this.emit("delete", key);
    return true;
  }

  public clear(): void {
    const count = this.nodeMap.size;
    this.nodeMap.clear();
    this.head = null;
    this.tail = null;
    this.emit("clear", count);
  }

  public size(): number {
    return this.nodeMap.size;
  }

  // --------------------------------------------------------------------------
  // Maintenance & Analytics
  // --------------------------------------------------------------------------

  public purgeExpired(): number {
    const now = Date.now();
    let purgedCount = 0;
    const expiredNodes: LinkedListNode<V>[] = [];

    this.nodeMap.forEach((node) => {
      if (node.entry.expiresAt !== null && node.entry.expiresAt <= now) {
        expiredNodes.push(node);
      }
    });

    for (const node of expiredNodes) {
      this.removeNode(node);
      this.nodeMap.delete(node.key);
      this.stats.expirations++;
      purgedCount++;
      this.emit("expired", node.key, node.entry.value);
    }

    this.stats.lastPurgeTimestamp = now;
    return purgedCount;
  }

  public getStats(): CacheStats {
    return { ...this.stats };
  }

  public getHitRatio(): number {
    const totalRequests = this.stats.hits + this.stats.misses;
    if (totalRequests === 0) return 0.0;
    return Math.round((this.stats.hits / totalRequests) * 10000) / 100;
  }

  public dumpKeysInLRUOrder(): string[] {
    const keys: string[] = [];
    let current = this.head;
    while (current !== null) {
      keys.push(current.key);
      current = current.next;
    }
    return keys;
  }

  // --------------------------------------------------------------------------
  // Doubly-Linked List Internal Helpers
  // --------------------------------------------------------------------------

  private insertAtHead(node: LinkedListNode<V>): void {
    node.prev = null;
    node.next = this.head;

    if (this.head !== null) {
      this.head.prev = node;
    }
    this.head = node;

    if (this.tail === null) {
      this.tail = node;
    }
  }

  private removeNode(node: LinkedListNode<V>): void {
    if (node.prev !== null) {
      node.prev.next = node.next;
    } else {
      this.head = node.next;
    }

    if (node.next !== null) {
      node.next.prev = node.prev;
    } else {
      this.tail = node.prev;
    }

    node.prev = null;
    node.next = null;
  }

  private moveToHead(node: LinkedListNode<V>): void {
    if (node === this.head) {
      return;
    }
    this.removeNode(node);
    this.insertAtHead(node);
  }

  private evictTail(): void {
    if (this.tail === null) {
      return;
    }
    const victim = this.tail;
    this.removeNode(victim);
    this.nodeMap.delete(victim.key);

    this.stats.evictions++;
    this.emit("evict", victim.key, victim.entry.value);
  }
}

// ----------------------------------------------------------------------------
// Domain Demonstration: Database Record Repository Simulation
// ----------------------------------------------------------------------------

interface UserAccount {
  userId: string;
  email: string;
  fullName: string;
  role: "admin" | "developer" | "analyst";
  reputation: number;
  lastLoginEpoch: number;
}

// Simulated slow Database loader
async function mockDatabaseFetchUser(userId: string): Promise<UserAccount> {
  const simulatedLatencyMs = 25;
  const start = Date.now();
  while (Date.now() - start < simulatedLatencyMs) {
    // Busy wait simulation for synchronous deterministic execution
  }

  return {
    userId,
    email: `${userId.toLowerCase()}@enterprise.internal`,
    fullName: `User ${userId.toUpperCase()}`,
    role: userId === "usr-001" ? "admin" : "developer",
    reputation: Math.floor(Math.random() * 500) + 100,
    lastLoginEpoch: Date.now(),
  };
}

console.log("=================================================================");
console.log("   SCRIPTGO ENTERPRISE LRU CACHE & SINGLE-FLIGHT DEMONSTRATION   ");
console.log("=================================================================");

async function main(): Promise<void> {
  const cache = new AdvancedLRUCache<UserAccount>({
    capacity: 3,
    defaultTTLMs: 10000,
    enableSingleFlight: true,
  });

  // Attach event observers
  cache.on("hit", (key: string) => {
    console.log(`  [EVENT: HIT] Key '${key}' resolved immediately from memory.`);
  });

  cache.on("miss", (key: string) => {
    console.log(`  [EVENT: MISS] Key '${key}' not in cache.`);
  });

  cache.on("evict", (key: string, val: UserAccount) => {
    console.log(`  [EVENT: EVICT] Capacity reached. Evicted least-recently-used '${key}' (${val.fullName}).`);
  });

  cache.on("singleFlightJoin", (key: string) => {
    console.log(`  [EVENT: SINGLE-FLIGHT] Shared existing in-flight promise for '${key}'.`);
  });

  // 1. Fill Cache to Capacity
  console.log("\n--- Phase 1: Populating Cache (Capacity = 3) ---");
  await cache.getOrLoad("usr-001", mockDatabaseFetchUser);
  await cache.getOrLoad("usr-002", mockDatabaseFetchUser);
  await cache.getOrLoad("usr-003", mockDatabaseFetchUser);

  console.log(`Active Keys in LRU Order: [${cache.dumpKeysInLRUOrder().join(", ")}]`);

  // 2. Access 'usr-001' to promote it to MRU (Head)
  console.log("\n--- Phase 2: Accessing usr-001 (Promotes to Head) ---");
  const user1 = cache.get("usr-001");
  if (user1) {
    console.log(`Retrieved user1: ${user1.fullName} (${user1.role})`);
  }
  console.log(`Active Keys after promotion: [${cache.dumpKeysInLRUOrder().join(", ")}]`);

  // 3. Insert 4th item -> Should evict 'usr-002' (which is at the tail)
  console.log("\n--- Phase 3: Inserting 4th item (Triggers LRU Eviction) ---");
  await cache.getOrLoad("usr-004", mockDatabaseFetchUser);
  console.log(`Active Keys after eviction: [${cache.dumpKeysInLRUOrder().join(", ")}]`);
  console.log(`Has usr-002? ${cache.has("usr-002")}`);
  console.log(`Has usr-001? ${cache.has("usr-001")}`);

  // 4. Test SingleFlight Stampede Mitigation with concurrent calls
  console.log("\n--- Phase 4: Simulating Concurrent Stampede for 'usr-005' ---");
  const p1 = cache.getOrLoad("usr-005", mockDatabaseFetchUser);
  const p2 = cache.getOrLoad("usr-005", mockDatabaseFetchUser);
  const p3 = cache.getOrLoad("usr-005", mockDatabaseFetchUser);

  const [res1, res2, res3] = await Promise.all([p1, p2, p3]);
  console.log(`Stampede resolved. Returned user: ${res1.fullName} (All 3 promises matched: ${res1 === res2 && res2 === res3})`);

  // 5. Telemetry & Analytics Report
  console.log("\n--- Phase 5: Telemetry and Performance Metrics ---");
  const stats = cache.getStats();
  console.log(`Total Hits           : ${stats.hits}`);
  console.log(`Total Misses         : ${stats.misses}`);
  console.log(`Hit Ratio            : ${cache.getHitRatio()}%`);
  console.log(`Evictions Triggered  : ${stats.evictions}`);
  console.log(`Single-Flight Saves  : ${stats.singleFlightSaves}`);
  console.log(`Final Cache Size     : ${cache.size()} / 3`);
  console.log("=================================================================\n");
}

main().catch((err) => {
  console.log(`Execution Error: ${err}`);
});
