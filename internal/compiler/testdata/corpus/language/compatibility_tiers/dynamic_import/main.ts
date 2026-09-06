async function load(): Promise<void> {
  await import("./dependency");
}

load();
