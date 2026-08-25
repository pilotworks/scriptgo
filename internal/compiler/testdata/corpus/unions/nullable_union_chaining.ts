// @expect: Key=secret, Timeout=5000
// @expect: Key=NO_KEY, Timeout=3000
// @expect: Key=NO_KEY, Timeout=3000
type Config = {
    apiKey?: string | null;
    timeout?: number | null;
};

function getConfigSummary(cfg: Config): string {
    const key = cfg.apiKey ?? "NO_KEY";
    const to = cfg.timeout ?? 3000;
    return "Key=" + key + ", Timeout=" + to;
}

console.log(getConfigSummary({ apiKey: "secret", timeout: 5000 }));
console.log(getConfigSummary({ apiKey: null, timeout: null }));
console.log(getConfigSummary({}));
