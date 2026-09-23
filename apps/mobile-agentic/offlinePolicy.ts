export function shouldQueueOffline(error: unknown): boolean {
  if (typeof error !== "object" || error === null || !("status" in error)) {
    return true;
  }
  return typeof (error as { status?: unknown }).status !== "number";
}
