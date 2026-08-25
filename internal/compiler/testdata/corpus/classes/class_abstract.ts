// @expect: Console: hello
// @expect: Console: [TIMESTAMP] world
abstract class BaseLogger {
    abstract logMessage(msg: string): void;

    logWithTimestamp(msg: string): void {
        this.logMessage("[TIMESTAMP] " + msg);
    }
}

class ConsoleLogger extends BaseLogger {
    logMessage(msg: string): void {
        console.log("Console: " + msg);
    }
}

const logger = new ConsoleLogger();
logger.logMessage("hello");
logger.logWithTimestamp("world");
