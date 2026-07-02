const OFFSETS: Record<string, (d: Date) => void> = {
  "30m": (d) => d.setMinutes(d.getMinutes() - 30),
  "1h": (d) => d.setHours(d.getHours() - 1),
  "12h": (d) => d.setHours(d.getHours() - 12),
  "24h": (d) => d.setHours(d.getHours() - 24),
  "7d": (d) => d.setDate(d.getDate() - 7),
  "30d": (d) => d.setDate(d.getDate() - 30),
};

export function computeRange(range: string) {
  const start = new Date();
  const apply = OFFSETS[range];
  if (apply) {
    apply(start);
  } else {
    start.setTime(0);
  }
  return { startAt: start.toISOString(), endAt: new Date().toISOString() };
}
