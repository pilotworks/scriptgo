export async function answer() {
  return await Promise.resolve(42);
}

export async function fail() {
  return await Promise.reject("recovered");
}
