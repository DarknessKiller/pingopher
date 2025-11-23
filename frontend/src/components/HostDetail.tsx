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
];

const HostDetail: React.FC<HostDetailProps> = ({ host }) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<ParsedHistory[]>([]);
  const [range, setRange] = useState("30m");

  useEffect(() => {
    setRange("30m");
  }, [host]);

  useEffect(() => {
    let isMounted = true;
    let timeoutId: ReturnType<typeof setTimeout>;

    const fetchHistory = async (showLoading = true) => {
      if (!host) return;
      if (showLoading) setLoading(true);

      const { startAt, endAt } = computeRange(range);

      try {
        const response = await getHostHistory(host.id, startAt, endAt);
        if (!isMounted) return;
        const results = response.data.results || [];

        const parsedData: ParsedHistory[] = results.map((item: Result) => ({
          latencyValue: parseInt(item.latency.replace(" ms", ""), 10),
          time: new Date(item.timestamp).toLocaleString(),
          dns: item.dns,
        }));

        setData(parsedData);
      } catch (error) {
        if (!isMounted) return;
        console.error(error);
        message.error((error as Error).message);
      } finally {
        if (isMounted) {
          if (showLoading) setLoading(false);
          timeoutId = setTimeout(() => fetchHistory(false), 10000);
        }
      }
    };

    fetchHistory(true);

    return () => {
      isMounted = false;
      clearTimeout(timeoutId);
    };
  }, [host, range]);

  useEffect(() => {
    if (!containerRef.current) return;

    const old = containerRef.current.querySelector("canvas");
    if (old) old.remove();

    if (data.length === 0) return;

    const line = new Line(containerRef.current, {
      data,
      xField: "time",
      yField: "latencyValue",
      seriesField: "dns",
      yAxis: { label: { formatter: (v) => `${v} ms` } },
      legend: { position: "top" },
      smooth: true,
    });

    line.render();
    return () => line.destroy();
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
