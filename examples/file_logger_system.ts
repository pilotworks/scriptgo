import { writeFileSync, readFileSync, existsSync, unlinkSync } from "node:fs";
import { join } from "node:path";
import { tmpdir } from "node:os";

interface LogEntry {
  level: "INFO" | "WARN" | "ERROR";
  message: string;
  timestamp: string;
}

class JsonFileLogger {
  private logPath: string;

  constructor(filename: string) {
    this.logPath = join(tmpdir(), filename);
  }

  log(level: "INFO" | "WARN" | "ERROR", message: string): void {
    const entry: LogEntry = {
      level,
      message,
      timestamp: new Date().toISOString(),
    };
    const line = JSON.stringify(entry) + "\n";
    writeFileSync(this.logPath, line);
  }

  readLastLog(): string {
    if (existsSync(this.logPath)) {
      return readFileSync(this.logPath);
    }
    return "";
  }

  cleanup(): void {
    if (existsSync(this.logPath)) {
      unlinkSync(this.logPath);
    }
  }
}

console.log("=== JSON File Logger System ===");
const logger = new JsonFileLogger("scriptgo_app.log");
try {
  logger.log("INFO", "Application initialization started");
  const content = logger.readLastLog();
  console.log(`Logged entry: ${content.trim()}`);
} finally {
  logger.cleanup();
  console.log("Cleanup log file finished.");
}
