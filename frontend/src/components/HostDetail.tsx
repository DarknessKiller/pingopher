import React, { useEffect, useRef, useState } from "react";
import { Line } from "@antv/g2plot";
import { getHostHistory, type Host, type Result } from "../api";
import { Spin, Empty, message, Select, Row, Col, Statistic, Card, Timeline, Tag } from "antd";
import { CheckCircleOutlined, CloseCircleOutlined, ClockCircleOutlined } from "@ant-design/icons";
import { computeRange } from "../utils/date";
import { processHistoryResults, type ParsedHistory, type DowntimeEvent } from "../utils/historyParser";

const { Option } = Select;

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

const HostDetail: React.FC<HostDetailProps> = ({ host }) => {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const chartRef = useRef<Line | null>(null);

  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<ParsedHistory[]>([]);
  const [downtimes, setDowntimes] = useState<DowntimeEvent[]>([]);
  const [uptimePercent, setUptimePercent] = useState<number>(100);
  const [dnsColors, setDnsColors] = useState<Record<string, string>>({});

  const [range, setRange] = useState<(typeof TIME_RANGES)[number]["value"]>("30m");

  // Reset time range when host changes
  useEffect(() => {
    setRange("30m");
    setData([]);
    setDowntimes([]);
    setUptimePercent(100);
    setDnsColors({});
  }, [host]);

useEffect(() => {
  if (!host) return;

  const controller = new AbortController();

  const fetchData = async () => {
    setLoading(true);
    const { startAt, endAt } = computeRange(range);

    try {
      const response = await getHostHistory(host.id, startAt, endAt, {
        signal: controller.signal,
      });

      if (controller.signal.aborted) return;

      const results: Result[] = response.data?.results ?? [];
      
      // Delegate parsing logic to utility module
      const processed = processHistoryResults(results, host);

      setUptimePercent(processed.uptimePercent);
      setDnsColors(processed.dnsColors);
      setData(processed.parsed);
      setDowntimes(processed.downtimes);
    } catch (err) {
      if (!controller.signal.aborted) {
        console.error(err);
        message.error((err as Error).message);
      }
    } finally {
      if (!controller.signal.aborted) setLoading(false);
    }
  };

  fetchData();

  return () => {
    controller.abort();
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
      tooltip: {
        formatter: (datum: any) => {
          return { name: datum.dns, value: datum.latencyValue !== null ? `${datum.latencyValue} ms` : '-' };
        },
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

  // ---- Update chart when data changes ----
  useEffect(() => {
    if (chartRef.current) {
      const annotations = downtimes.map((dt) => ({
        type: 'region',
        start: [dt.formattedStart, 'min'],
        end: [dt.formattedEnd || (data.length > 0 ? data[data.length - 1].time : dt.formattedStart), 'max'],
        style: {
          fill: '#ff4d4f',
          fillOpacity: 0.15,
        },
      }));

      chartRef.current.update({
        data: data.length ? data : [],
        annotations: annotations as any,
        color: (datum: any) => dnsColors[datum.dns] || '#5B8FF9',
      });
    }
  }, [data, downtimes, dnsColors]);

  return (
    <div>
      <div style={{ marginTop: 10, display: "flex", justifyContent: "space-between", flexWrap: "wrap", alignItems: "center" }}>
        <h4 style={{ marginBottom: 16 }}>Latency History</h4>

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

      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={24} sm={12} md={8}>
          <Card size="small">
            <Statistic
              title="Uptime"
              value={uptimePercent}
              precision={2}
              suffix="%"
              valueStyle={{ color: uptimePercent >= 95 ? '#3f8600' : '#cf1322' }}
              prefix={uptimePercent >= 95 ? <CheckCircleOutlined /> : <CloseCircleOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={8}>
          <Card size="small">
            <Statistic
              title="Incidents"
              value={downtimes.length}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: downtimes.length === 0 ? '#3f8600' : '#cf1322' }}
            />
          </Card>
        </Col>
      </Row>

      <div style={{ position: "relative", height: 400, marginBottom: 24 }}>
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
          <div
            style={{
              position: "absolute",
              top: "50%",
              left: "50%",
              transform: "translate(-50%, -50%)",
              zIndex: 1,
            }}
          >
            <Empty description="No history data available" />
          </div>
        )}

        <div ref={containerRef} style={{ height: "100%" }} />
      </div>

      {downtimes.length > 0 && (
        <div style={{ marginTop: 24 }}>
          <h4 style={{ marginBottom: 16 }}>Downtime Events</h4>
          <div style={{ maxHeight: 300, overflowY: 'auto', paddingRight: 16 }}>
            <Timeline
              items={downtimes.map((dt, idx) => ({
                key: idx,
                color: dt.end ? 'red' : 'orange',
                children: (
                  <div>
                    <div style={{ fontWeight: 'bold' }}>
                      {dt.formattedStart} {dt.formattedEnd ? ` - ${dt.formattedEnd}` : ' - Now'}
                    </div>
                    <div style={{ marginTop: 4 }}>
                      <Tag color={dnsColors[dt.dns] || "geekblue"}>{dt.dns}</Tag>
                      <Tag color="error">{dt.reason}</Tag>
                      <span style={{ marginLeft: 8, color: '#888' }}>
                        Duration: {dt.durationString}
                      </span>
                    </div>
                  </div>
                ),
              }))}
            />
          </div>
        </div>
      )}
    </div>
  );
};

export default HostDetail;
