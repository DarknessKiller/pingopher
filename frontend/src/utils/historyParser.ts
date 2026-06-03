import { type Host, type Result } from "../api";

export interface ParsedHistory {
  latencyValue: number | null;
  time: string;
  dns?: string;
}

export interface DowntimeEvent {
  start: Date;
  end: Date | null;
  formattedStart: string;
  formattedEnd: string | null;
  reason: string;
  durationString: string;
  dns: string;
}

export const COLOR_PALETTE = [
  '#5B8FF9', '#5AD8A6', '#5D7092', '#F6BD16', '#E8684A', 
  '#6DC8EC', '#9270CA', '#FF9D4D', '#269A99', '#FF99C3'
];

export const formatDuration = (ms: number) => {
  if (ms < 1000) return "< 1s";
  const sec = Math.floor(ms / 1000);
  if (sec < 60) return `${sec}s`;
  const min = Math.floor(sec / 60);
  if (min < 60) return `${min}m ${sec % 60}s`;
  const hrs = Math.floor(min / 60);
  return `${hrs}h ${min % 60}m`;
};

export const checkStatusCode = (status: number, patterns: string[] | undefined): boolean => {
  if (!patterns || patterns.length === 0) return true;
  for (let p of patterns) {
    p = p.trim().toLowerCase();
    if (!p) continue;
    if (p === status.toString()) return true;
    if (p.includes("-")) {
      const parts = p.split("-");
      if (parts.length === 2) {
        const from = parseInt(parts[0], 10);
        const to = parseInt(parts[1], 10);
        if (status >= from && status <= to) return true;
      }
    }
    if (p.includes("x") || p.includes("*")) {
      const regexStr = p.replace(/x|\*/g, "\\d");
      const regex = new RegExp(`^${regexStr}$`);
      if (regex.test(status.toString())) return true;
    }
  }
  return false;
};

export interface ProcessedHistoryData {
  dnsColors: Record<string, string>;
  downtimes: DowntimeEvent[];
  parsed: ParsedHistory[];
  uptimePercent: number;
}

export const processHistoryResults = (results: Result[], host: Host): ProcessedHistoryData => {
  if (results.length === 0) {
    return { dnsColors: {}, downtimes: [], parsed: [], uptimePercent: 100 };
  }

  // Ensure chronological order
  results.sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());

  // Assign colors dynamically
  const dnsNames = Array.from(new Set(results.map(r => r.dns || "System DNS"))).sort();
  const colorMap: Record<string, string> = {};
  dnsNames.forEach((n, i) => {
    colorMap[n] = COLOR_PALETTE[i % COLOR_PALETTE.length];
  });

  // Group by DNS
  const resultsByDns: Record<string, Result[]> = {};
  for (const r of results) {
    const key = r.dns || "System DNS";
    if (!resultsByDns[key]) resultsByDns[key] = [];
    resultsByDns[key].push(r);
  }

  const rawDowntimes: DowntimeEvent[] = [];

  for (const [dnsName, dnsResults] of Object.entries(resultsByDns)) {
    let currentDowntime: DowntimeEvent | null = null;
    
    for (const item of dnsResults) {
      const isDown = !!item.errorMsg || (item.statusCode !== 0 && !checkStatusCode(item.statusCode, host.acceptedStatusCodes));
      const time = new Date(item.timestamp);
      const formattedTime = time.toLocaleString();

      if (isDown) {
        if (!currentDowntime) {
          currentDowntime = {
            start: time,
            end: null,
            formattedStart: formattedTime,
            formattedEnd: null,
            reason: item.errorMsg || `Status Code: ${item.statusCode}`,
            durationString: "Ongoing",
            dns: dnsName, 
          };
        }
      } else {
        if (currentDowntime) {
          currentDowntime.end = time;
          currentDowntime.formattedEnd = formattedTime;
          const durationMs = time.getTime() - currentDowntime.start.getTime();
          currentDowntime.durationString = formatDuration(durationMs);
          rawDowntimes.push({ ...currentDowntime });
          currentDowntime = null;
        }
      }
    }

    if (currentDowntime) {
      const now = new Date();
      const durationMs = now.getTime() - currentDowntime.start.getTime();
      currentDowntime.durationString = `${formatDuration(durationMs)} (Ongoing)`;
      rawDowntimes.push(currentDowntime);
    }
  }

  // Merge
  const newDowntimes: DowntimeEvent[] = [];
  const mergeThresholdMs = (host.pingInterval || 60) * 1000 * 1.5;

  const rawByDns: Record<string, DowntimeEvent[]> = {};
  for (const dt of rawDowntimes) {
    if (!rawByDns[dt.dns]) rawByDns[dt.dns] = [];
    rawByDns[dt.dns].push(dt);
  }

  for (const [, dts] of Object.entries(rawByDns)) {
      const merged: DowntimeEvent[] = [];
      for (const dt of dts) {
        if (merged.length === 0) {
          merged.push(dt);
        } else {
          const prev = merged[merged.length - 1];
          if (prev.end) {
            const gapMs = dt.start.getTime() - prev.end.getTime();
            if (gapMs <= mergeThresholdMs) {
              prev.end = dt.end;
              prev.formattedEnd = dt.formattedEnd;
              if (dt.end) {
                const durMs = dt.end.getTime() - prev.start.getTime();
                prev.durationString = formatDuration(durMs);
              } else {
                const durMs = new Date().getTime() - prev.start.getTime();
                prev.durationString = `${formatDuration(durMs)} (Ongoing)`;
              }
              continue;
            }
          }
          merged.push(dt);
        }
      }
      newDowntimes.push(...merged);
  }

  let uptimePercent = 100;
  if (results.length > 0) {
    let totalDowntimeMs = 0;
    let totalMonitoredMs = 0;
    const now = new Date().getTime();

    for (const dnsResults of Object.values(resultsByDns)) {
      if (dnsResults.length > 0) {
        const firstPingMs = new Date(dnsResults[0].timestamp).getTime();
        totalMonitoredMs += Math.max(0, now - firstPingMs);
      }
    }

    for (const dt of newDowntimes) {
      const startMs = dt.start.getTime();
      const endMs = dt.end ? dt.end.getTime() : now;
      totalDowntimeMs += Math.max(0, endMs - startMs);
    }

    if (totalMonitoredMs > 0) {
      const actualUptimeMs = Math.max(0, totalMonitoredMs - totalDowntimeMs);
      uptimePercent = Math.min(100, (actualUptimeMs / totalMonitoredMs) * 100);
    }
  }

  newDowntimes.sort((a, b) => b.start.getTime() - a.start.getTime());

  const parsed: ParsedHistory[] = results.map((item) => {
    const isDown = !!item.errorMsg || (item.statusCode !== 0 && !checkStatusCode(item.statusCode, host.acceptedStatusCodes));
    return {
      latencyValue: isDown ? null : (parseInt(item.latency.replace(" ms", ""), 10) || 0),
      time: new Date(item.timestamp).toLocaleString(),
      dns: item.dns || "System DNS",
    };
  });
  parsed.sort((a, b) => new Date(a.time).getTime() - new Date(b.time).getTime());

  return {
    dnsColors: colorMap,
    downtimes: newDowntimes,
    parsed,
    uptimePercent
  };
};
