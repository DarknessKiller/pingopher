import React, { useEffect, useRef, useState } from "react";
import { Line } from "@antv/g2plot";
import { getHostHistory, type Host, type Result } from "../api";
import { Spin, Empty, message, Select } from "antd";
import { computeRange } from "../utils/date";

const { Option } = Select;

interface HostDetailProps {
  host: Host;
}

interface ParsedHistory {
  latencyValue: number;
  time: string;
  dns?: string;
}

const TIME_RANGES = [
  { label: "Last 30 minutes", value: "30m" },
  { label: "Last 1 hour", value: "1h" },
  { label: "Last 12 hours", value: "12h" },
  { label: "Last 24 hours", value: "24h" },
  { label: "Last 7 days", value: "7d" },
  { label: "Last 30 days", value: "30d" },
] as const;

const HostDetail: React.FC<HostDetailProps> = ({ host }) => {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const chartRef = useRef<Line | null>(null);

  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<ParsedHistory[]>([]);
  const [range, setRange] =
    useState<(typeof TIME_RANGES)[number]["value"]>("30m");

  // Reset time range when host changes
  useEffect(() => {
    setRange("30m");
    setData([]);
  }, [host]);

  // ---- Async polling effect ----
  useEffect(() => {
    if (!host) return;

    const controller = new AbortController();
    let refreshTimer: NodeJS.Timeout | null = null;

    const fetchHistoryAsync = async () => {
      setLoading(true);
      const { startAt, endAt } = computeRange(range);

      try {
        const response = await getHostHistory(host.id, startAt, endAt, {
          signal: controller.signal,
        });

        if (controller.signal.aborted) return;

        const results: Result[] = response.data?.results ?? [];
        const parsed: ParsedHistory[] = results.map((item) => ({
          latencyValue: parseInt(item.latency.replace(" ms", ""), 10),
          time: new Date(item.timestamp).toLocaleString(),
          dns: item.dns,
        }));

        setData(parsed);
      } catch (err) {
        if (!controller.signal.aborted) {
          console.error(err);
          message.error((err as Error).message);
        }
      } finally {
        if (!controller.signal.aborted) setLoading(false);
      }
    };

    const runPolling = async () => {
      await fetchHistoryAsync();
      if (!controller.signal.aborted) {
        refreshTimer = setTimeout(runPolling, 10_000);
      }
    };

    runPolling();

    return () => {
      controller.abort();
      if (refreshTimer) clearTimeout(refreshTimer);
    };
  }, [host, range]);

  // ---- Initialize Chart Once ----
  useEffect(() => {
    if (!containerRef.current || chartRef.current) return;

    chartRef.current = new Line(containerRef.current, {
      data: [],
      xField: "time",
      yField: "latencyValue",
      seriesField: "dns",
      yAxis: { label: { formatter: (v) => `${v} ms` } },
      legend: { position: "top" },
      smooth: true,
      lineStyle: { lineWidth: 2 },
    });

    chartRef.current.render();

    return () => {
      chartRef.current?.destroy();
      chartRef.current = null;
    };
  }, []);

  // ---- Update chart when data changes ----
  useEffect(() => {
    if (!chartRef.current) return;

    chartRef.current.changeData(data.length ? data : []);
  }, [data]);

  return (
    <div>
      <div style={{ marginTop: 10 }}>
        <h4 style={{ marginBottom: 8 }}>Latency History</h4>

        <Select
          value={range}
          onChange={setRange}
          style={{ width: 200, marginBottom: 16 }}
        >
          {TIME_RANGES.map((r) => (
            <Option key={r.value} value={r.value}>
              {r.label}
            </Option>
          ))}
        </Select>
      </div>

      <div style={{ position: "relative", height: 400 }}>
        {loading && (
          <div
            style={{
              position: "absolute",
              top: "50%",
              left: "50%",
              transform: "translate(-50%, -50%)",
              zIndex: 1,
            }}
          >
            <Spin size="large" />
          </div>
        )}

        {!loading && data.length === 0 && (
          <Empty
            description="No history data available"
            style={{ marginTop: 100 }}
          />
        )}

        <div ref={containerRef} style={{ height: "100%" }} />
      </div>
    </div>
  );
};

export default HostDetail;
