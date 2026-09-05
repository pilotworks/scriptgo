// @expect: boom

async function failAfterAwait(): Promise<void> {
    await Promise.resolve(undefined);
    throw new Error("boom");
}

failAfterAwait().catch((error: Error) => console.log(error.message));
