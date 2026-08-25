// @expect: 5000
// @expect: 3
class ConfigManager {
    static defaultTimeout: number;
    static defaultRetries: number;

    static {
        ConfigManager.defaultTimeout = 5000;
        ConfigManager.defaultRetries = 3;
    }
}

console.log(ConfigManager.defaultTimeout);
console.log(ConfigManager.defaultRetries);
