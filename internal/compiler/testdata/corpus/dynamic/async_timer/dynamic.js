export async function delayValue(ms, value) {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve(value);
    }, ms);
  });
}

export async function immediateValue(value) {
  return new Promise((resolve) => {
    setImmediate(() => {
      resolve(value);
    });
  });
}
