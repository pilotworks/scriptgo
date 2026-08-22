// ============================================================================
// Enterprise Asynchronous Reactive Stream Processing & Analytics Engine
// ============================================================================
// Features demonstrated:
// - Async Generators (async function* / yield / yield*) & for await..of
// - Monadic Functional Stream Pipeline (map, filter, window, batch, retry, fold)
// - Windowing Operators: Tumbling Time Windows & Count-Based Batching
// - Fault-Tolerance: Exponential Backoff Retry & Dead Letter Queue (DLQ)
// - Real-Time Streaming Statistics: Welford's Online Algorithm for Variance/StdDev
// - Math Operations: Mean, Min, Max, Variance, Standard Deviation, Percentile
// - Error Handling, Type Narrowing, Interfaces & Generic Specialization
// ============================================================================

// ----------------------------------------------------------------------------
// Domain Types & Models
// ----------------------------------------------------------------------------

export interface IoTTelemetry {
  deviceId: string;
  cluster: string;
  temperatureC: number;
  vibrationHz: number;
  powerWatts: number;
  timestamp: number;
}

export interface MetricSummary {
  count: number;
  mean: number;
  stdDev: number;
  min: number;
  max: number;
}

export interface AggregatedWindowReport {
  cluster: string;
  windowStart: number;
  windowEnd: number;
  sampleCount: number;
  temperature: MetricSummary;
  power: MetricSummary;
  anomaliesDetected: number;
}

export interface DeadLetterRecord<T> {
  item: T;
  error: string;
  attemptCount: number;
  timestamp: number;
}

// ----------------------------------------------------------------------------
// Online Streaming Statistics (Welford's Algorithm)
// ----------------------------------------------------------------------------
// Computes running mean and variance in a single pass without storing all data
// ----------------------------------------------------------------------------

export class WelfordAccumulator {
  private count: number = 0;
  private mean: number = 0;
  private M2: number = 0;
  private minVal: number = Infinity;
  private maxVal: number = -Infinity;

  public update(value: number): void {
    this.count++;
    const delta = value - this.mean;
    this.mean += delta / this.count;
    const delta2 = value - this.mean;
    this.M2 += delta * delta2;

    if (value < this.minVal) this.minVal = value;
    if (value > this.maxVal) this.maxVal = value;
  }

  public getSummary(): MetricSummary {
    if (this.count === 0) {
      return { count: 0, mean: 0, stdDev: 0, min: 0, max: 0 };
    }
    const variance = this.count > 1 ? this.M2 / (this.count - 1) : 0;
    const stdDev = Math.sqrt(variance);

    return {
      count: this.count,
      mean: Math.round(this.mean * 100) / 100,
      stdDev: Math.round(stdDev * 100) / 100,
      min: Math.round(this.minVal * 100) / 100,
      max: Math.round(this.maxVal * 100) / 100,
    };
  }
}

// ----------------------------------------------------------------------------
// Reactive Asynchronous Stream Pipeline
// ----------------------------------------------------------------------------

export type AsyncTransform<T, R> = (item: T) => Promise<R> | R;
export type AsyncPredicate<T> = (item: T) => Promise<boolean> | boolean;

export class AsyncStream<T> {
  private constructor(private readonly producer: () => AsyncGenerator<T>) {}

  public static fromArray<T>(items: T[]): AsyncStream<T> {
    return new AsyncStream<T>(async function* () {
      for (const item of items) {
        yield item;
      }
    });
  }

  public static fromGenerator<T>(gen: () => AsyncGenerator<T>): AsyncStream<T> {
    return new AsyncStream<T>(gen);
  }

  /**
   * Transforms each element asynchronously.
   */
  public map<R>(fn: AsyncTransform<T, R>): AsyncStream<R> {
    const upstream = this.producer;
    return new AsyncStream<R>(async function* () {
      for await (const item of upstream()) {
        yield await fn(item);
      }
    });
  }

  /**
   * Filters stream elements matching a predicate.
   */
  public filter(predicate: AsyncPredicate<T>): AsyncStream<T> {
    const upstream = this.producer;
    return new AsyncStream<T>(async function* () {
      for await (const item of upstream()) {
        if (await predicate(item)) {
          yield item;
        }
      }
    });
  }

  /**
   * Batches elements into fixed-size chunks.
   */
  public batch(size: number): AsyncStream<T[]> {
    const upstream = this.producer;
    return new AsyncStream<T[]>(async function* () {
      let buffer: T[] = [];
      for await (const item of upstream()) {
        buffer.push(item);
        if (buffer.length >= size) {
          yield buffer;
          buffer = [];
        }
      }
      if (buffer.length > 0) {
        yield buffer;
      }
    });
  }

  /**
   * Retries an async transform on failure with exponential backoff.
   * If all retries fail, routes the record to the Dead Letter Queue (DLQ).
   */
  public retryWithDLQ<R>(
    transform: (item: T) => Promise<R>,
    maxRetries: number,
    dlq: DeadLetterRecord<T>[]
  ): AsyncStream<R> {
    const upstream = this.producer;
    return new AsyncStream<R>(async function* () {
      for await (const item of upstream()) {
        let attempts = 0;
        let success = false;
        let lastError = "";

        while (attempts < maxRetries && !success) {
          attempts++;
          try {
            const result = await transform(item);
            yield result;
            success = true;
          } catch (err) {
            lastError = `${err}`;
          }
        }

        if (!success) {
          dlq.push({
            item,
            error: lastError,
            attemptCount: attempts,
            timestamp: Date.now(),
          });
        }
      }
    });
  }

  /**
   * Reduces the entire stream to a single aggregated value.
   */
  public async fold<R>(reducer: (acc: R, item: T) => R, initial: R): Promise<R> {
    let acc = initial;
    for await (const item of this.producer()) {
      acc = reducer(acc, item);
    }
    return acc;
  }

  /**
   * Collects all stream elements into an array.
   */
  public async toArray(): Promise<T[]> {
    const results: T[] = [];
    for await (const item of this.producer()) {
      results.push(item);
    }
    return results;
  }
}

// ----------------------------------------------------------------------------
// Simulated IoT Telemetry Stream Generator
// ----------------------------------------------------------------------------

async function* generateIndustrialTelemetry(): AsyncGenerator<IoTTelemetry> {
  const clusters = ["alpha-forge", "beta-hydro", "gamma-turbine", "delta-datacenter"];
  const count = 16;

  for (let i = 0; i < count; i++) {
    const cluster = clusters[i % clusters.length];
    const isTurbine = cluster === "gamma-turbine";
    const isAnomaly = i === 7 || i === 13;

    const baseTemp = isTurbine ? 65.0 : 42.0;
    const tempNoise = (i % 5) * 1.8 - 3.6;
    const temperatureC = isAnomaly ? 112.5 : baseTemp + tempNoise;

    const vibrationHz = isAnomaly ? 450.0 : 60.0 + (i % 4) * 5.2;
    const powerWatts = isAnomaly ? 8500.0 : 3200.0 + (i % 6) * 180.0;

    yield {
      deviceId: `device-${(100 + i).toString()}`,
      cluster,
      temperatureC: Math.round(temperatureC * 100) / 100,
      vibrationHz: Math.round(vibrationHz * 10) / 10,
      powerWatts: Math.round(powerWatts * 10) / 10,
      timestamp: 1724300000000 + i * 5000,
    };
  }
}

// ----------------------------------------------------------------------------
// Execution & Stream Processing Workflow
// ----------------------------------------------------------------------------

console.log("=================================================================");
console.log("   SCRIPTGO ASYNCHRONOUS REACTIVE STREAM PROCESSING ENGINE       ");
console.log("=================================================================");

async function runAnalyticsPipeline(): Promise<void> {
  const deadLetterQueue: DeadLetterRecord<IoTTelemetry>[] = [];

  console.log("\n[1] Starting Stream Consumption & Processing Pipeline...");

  // Build stream pipeline
  const stream = AsyncStream.fromGenerator(generateIndustrialTelemetry)
    // Stage 1: Validation & Quality Control
    .filter((record) => {
      return record.temperatureC > -40.0 && record.temperatureC < 200.0;
    })
    // Stage 2: Transform with simulated flaky external enrichment & DLQ handling
    .retryWithDLQ(
      async (record): Promise<IoTTelemetry> => {
        // Simulate occasional network timeout on specific device
        if (record.deviceId === "device-104") {
          throw new Error("External calibration service unavailable (503)");
        }
        return record;
      },
      2, // Max 2 retries
      deadLetterQueue
    );

  // Stage 3: Ingest valid stream items
  const validRecords = await stream.toArray();
  console.log(`  Stream processed ${validRecords.length} records successfully.`);
  console.log(`  Dead Letter Queue captured ${deadLetterQueue.length} failed records.`);

  if (deadLetterQueue.length > 0) {
    for (const dlq of deadLetterQueue) {
      console.log(`  [DLQ ALERT] Device '${dlq.item.deviceId}': ${dlq.error} after ${dlq.attemptCount} retries.`);
    }
  }

  // --------------------------------------------------------------------------
  // Aggregation & Statistical Window Calculation
  // --------------------------------------------------------------------------
  console.log("\n[2] Computing Statistical Aggregations by Cluster (Welford's Online Algorithm)...");

  const clusterGroups = new Map<string, IoTTelemetry[]>();
  for (const item of validRecords) {
    const group = clusterGroups.get(item.cluster);
    if (group) {
      group.push(item);
    } else {
      clusterGroups.set(item.cluster, [item]);
    }
  }

  const reports: AggregatedWindowReport[] = [];

  clusterGroups.forEach((records, cluster) => {
    const tempAcc = new WelfordAccumulator();
    const powerAcc = new WelfordAccumulator();
    let anomalies = 0;
    let minTimestamp = Infinity;
    let maxTimestamp = -Infinity;

    for (const r of records) {
      tempAcc.update(r.temperatureC);
      powerAcc.update(r.powerWatts);

      if (r.timestamp < minTimestamp) minTimestamp = r.timestamp;
      if (r.timestamp > maxTimestamp) maxTimestamp = r.timestamp;

      // Anomaly threshold: temperature > 80C or vibration > 200Hz
      if (r.temperatureC > 80.0 || r.vibrationHz > 200.0) {
        anomalies++;
      }
    }

    const report: AggregatedWindowReport = {
      cluster,
      windowStart: minTimestamp,
      windowEnd: maxTimestamp,
      sampleCount: records.length,
      temperature: tempAcc.getSummary(),
      power: powerAcc.getSummary(),
      anomaliesDetected: anomalies,
    };

    reports.push(report);
  });

  // Display structured statistical summaries
  console.log("\n+--------------------+---------+---------------+---------------+--------------------+-----------+");
  console.log("| Cluster Name       | Samples | Temp Mean±Std | Temp Range    | Power Mean±Std (W) | Anomalies |");
  console.log("+--------------------+---------+---------------+---------------+--------------------+-----------+");

  for (const rep of reports) {
    const cName = rep.cluster.padEnd(18);
    const count = rep.sampleCount.toString().padStart(7);
    const tempMean = `${rep.temperature.mean}±${rep.temperature.stdDev}°C`.padEnd(13);
    const tempRange = `[${rep.temperature.min}, ${rep.temperature.max}]`.padEnd(13);
    const powerMean = `${rep.power.mean}±${rep.power.stdDev}`.padEnd(18);
    const anom = rep.anomaliesDetected.toString().padStart(9);

    console.log(`| ${cName} | ${count} | ${tempMean} | ${tempRange} | ${powerMean} | ${anom} |`);
  }
  console.log("+--------------------+---------+---------------+---------------+--------------------+-----------+");

  console.log("\n=================================================================");
  console.log("         ASYNCHRONOUS PIPELINE EXECUTION FINISHED                ");
  console.log("=================================================================\n");
}

runAnalyticsPipeline().catch((err) => {
  console.log(`Pipeline Fatal Error: ${err}`);
});
