// @expect: Standard Console::Logger: This is an informational message.
// @expect: ---
// @expect: File::Logger: This is a debug level message.
// @expect: Standard Console::Logger: This is a debug level message.
// @expect: ---
// @expect: Error Console::Logger: This is an error information.
// @expect: File::Logger: This is an error information.
// @expect: Standard Console::Logger: This is an error information.
abstract class AbstractLogger {
    protected level: number;
    protected nextLogger: AbstractLogger | null = null;

    constructor(level: number) {
        this.level = level;
    }

    setNextLogger(nextLogger: AbstractLogger): AbstractLogger {
        this.nextLogger = nextLogger;
        return nextLogger;
    }

    logMessage(level: number, message: string): void {
        if (this.level <= level) {
            this.write(message);
        }
        if (this.nextLogger !== null) {
            this.nextLogger.logMessage(level, message);
        }
    }

    abstract write(message: string): void;
}

class InfoLogger extends AbstractLogger {
    constructor(level: number) {
        super(level);
    }

    write(message: string): void {
        console.log("Standard Console::Logger: " + message);
    }
}

class ErrorLogger extends AbstractLogger {
    constructor(level: number) {
        super(level);
    }

    write(message: string): void {
        console.log("Error Console::Logger: " + message);
    }
}

class FileLogger extends AbstractLogger {
    constructor(level: number) {
        super(level);
    }

    write(message: string): void {
        console.log("File::Logger: " + message);
    }
}

const INFO = 1;
const DEBUG = 2;
const ERROR = 3;

const errorLogger = new ErrorLogger(ERROR);
const fileLogger = new FileLogger(DEBUG);
const consoleLogger = new InfoLogger(INFO);

errorLogger.setNextLogger(fileLogger);
fileLogger.setNextLogger(consoleLogger);

consoleLogger.logMessage(INFO, "This is an informational message.");
console.log("---");
fileLogger.logMessage(DEBUG, "This is a debug level message.");
console.log("---");
errorLogger.logMessage(ERROR, "This is an error information.");
