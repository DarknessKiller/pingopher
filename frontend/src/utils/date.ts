export function computeRange(range: string) {
  const start = new Date();
  const now = new Date();

  switch (range) {
    case "30m":
      start.setMinutes(now.getMinutes() - 30);
      break;
    case "1h":
      start.setHours(now.getHours() - 1);
      break;
    case "12h":
      start.setHours(now.getHours() - 12);
      break;
    case "24h":
      start.setHours(now.getHours() - 24);
      break;
    case "7d":
      start.setDate(now.getDate() - 7);
      break;
    case "30d":
      start.setDate(now.getDate() - 30);
      break;
    default:
      start.setTime(0);
  }

  return {
    startAt: start.toISOString(),
    endAt: now.toISOString(),
  };
}
