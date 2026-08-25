import React, { useEffect, useRef, useState } from "react";
import { Line, type Datum } from "@antv/g2plot";
import { getHostHistory, type Host, type Result } from "../api";
import { useToast } from "./Toast";
import { computeRange } from "../utils/date";
import { processHistoryResults, type ParsedHistory, type DowntimeEvent } from "../utils/historyParser";
import { CheckCircleIcon, XCircleIcon, ClockIcon } from "./icons";

interface HostDetailProps {
  host: Host;
}

const TIME_RANGES = [
  { label: "Last 30 minutes", value: "30m" },
  { label: "Last 1 hour", value: "1h" },
  { label: "Last 12 hours", value: "12h" },
  { label: "Last 24 hours", value: "24h" },
  { label: "Last 7 days", value: "7d" },
  { label: "Last 30 days", value: "30d" },
] as const;

const APPLE_PALETTE = ["#007aff", "#34c759", "#ff9500", "#ff3b30", "#af52de", "#5ac8fa", "#ffcc00", "#ff2d55"];

const HostDetail: React.FC<HostDetailProps> = ({ host }) => {
  const toast = useToast();
  const containerRef = useRef<HTMLDivElement | null>(null);
  const chartRef = useRef<Line | null>(null);

  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<ParsedHistory[]>([]);
  const [downtimes, setDowntimes] = useState<DowntimeEvent[]>([]);
  const [uptimePercent, setUptimePercent] = useState<number>(100);
  const [dnsColors, setDnsColors] = useState<Record<string, string>>({});
  const [range, setRange] = useState<(typeof TIME_RANGES)[number]["value"]>("30m");

  useEffect(() => {
    if (!host) return;
    const controller = new AbortController();

    const fetchData = async () => {
      setLoading(true);
      const { startAt, endAt } = computeRange(range);
      try {
        const response = await getHostHistory(host.id, startAt, endAt, { signal: controller.signal });
        if (controller.signal.aborted) return;
        const results: Result[] = response.data?.results ?? [];
        const processed = processHistoryResults(results, host);
        setUptimePercent(processed.uptimePercent);
        setDnsColors(processed.dnsColors);
        setData(processed.parsed);
        setDowntimes(processed.downtimes);
      } catch (err) {
        if (!controller.signal.aborted) {
          console.error(err);
          toast.error((err as Error).message);
        }
      } finally {
        if (!controller.signal.aborted) setLoading(false);
      }
    };

    fetchData();
    return () => { controller.abort(); };
  }, [host, range]);

  useEffect(() => {
    if (!containerRef.current || chartRef.current) return;

    chartRef.current = new Line(containerRef.current, {
      data: [],
      xField: "time",
      yField: "latencyValue",
      seriesField: "dns",
      color: APPLE_PALETTE,
      yAxis: {
        label: { formatter: (v) => `${v} ms` },
      },
      tooltip: {
        formatter: (datum: Datum) => ({
          name: String(datum.dns),
          value: datum.latencyValue !== null ? `${datum.latencyValue} ms` : "-",
        }),
      },
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

  useEffect(() => {
    if (chartRef.current) {
      const annotations = downtimes.map((dt) => ({
        type: "region" as const,
        start: [dt.formattedStart, "min"] as [string, string],
        end: [
          dt.formattedEnd || (data.length > 0 ? data[data.length - 1].time : dt.formattedStart),
          "max",
        ] as [string, string],
        style: { fill: "var(--danger)", fillOpacity: 0.15 },
      }));
      chartRef.current.update({
        data,
        annotations,
        color: (datum: Datum) => dnsColors[datum.dns] || APPLE_PALETTE[0],
      });
    }
  }, [data, downtimes, dnsColors]);

  return (
    <div>
      {/* Stats */}
      <div className="stat-row">
        <div className="card card-sm">
          <div className="stat-card">
            <div className={`stat-card-icon ${uptimePercent >= 95 ? "success" : "danger"}`}>
              {uptimePercent >= 95 ? <CheckCircleIcon size={20} /> : <XCircleIcon size={20} />}
            </div>
            <div>
              <div className="stat-card-value" style={{ color: uptimePercent >= 95 ? "var(--success)" : "var(--danger)" }}>
                {uptimePercent.toFixed(2)}%
              </div>
              <div className="stat-card-label">Uptime</div>
            </div>
          </div>
        </div>
        <div className="card card-sm">
          <div className="stat-card">
            <div className={`stat-card-icon ${downtimes.length === 0 ? "success" : "danger"}`}>
              <ClockIcon size={20} />
            </div>
            <div>
              <div className="stat-card-value" style={{ color: downtimes.length === 0 ? "var(--success)" : "var(--danger)" }}>
                {downtimes.length}
              </div>
              <div className="stat-card-label">Incidents</div>
            </div>
          </div>
        </div>
      </div>

      {/* Chart */}
      <div className="chart-header">
        <h4>Latency History</h4>
        <select className="form-select" style={{ width: 200 }} value={range} onChange={(e) => setRange(e.target.value as typeof range)}>
          {TIME_RANGES.map((r) => (
            <option key={r.value} value={r.value}>{r.label}</option>
          ))}
        </select>
      </div>

      <div className="chart-container">
        {loading && (
          <div className="overlay-center">
            <div className="spinner spinner-lg" />
          </div>
        )}
        {!loading && data.length === 0 && (
          <div className="overlay-center">
            <div className="empty-state">
              <ClockIcon size={32} />
              <span>No history data available</span>
            </div>
          </div>
        )}
        <div ref={containerRef} style={{ height: "100%" }} />
      </div>

      {/* Downtime Timeline */}
      {downtimes.length > 0 && (
        <div>
          <h4 style={{ marginBottom: "var(--space-md)", fontWeight: 600 }}>Downtime Events</h4>
          <div className="timeline">
            {downtimes.map((dt, idx) => (
              <div key={idx} className={`timeline-item ${dt.end ? "down" : "partial"}`}>
                <div className="timeline-item-time">
                  {dt.formattedStart} {dt.formattedEnd ? `– ${dt.formattedEnd}` : "– Now"}
                </div>
                <div className="timeline-item-meta">
                  <span className="tag tag-info">{dt.dns}</span>
                  <span className="tag tag-danger">{dt.reason}</span>
                  <span className="timeline-item-duration">Duration: {dt.durationString}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

export default HostDetail;
