import React, { useState } from "react";
import { createHost, updateHost, type Host, type CreateHostRequest } from "../api";
import { useToast } from "./Toast";
import { PlusIcon, CloseIcon } from "./icons";

interface HostFormProps {
  initialValues?: Host;
  onSuccess: () => void;
}

const DEFAULT_VALUES: CreateHostRequest = {
  name: "",
  protocol: "https",
  hostUrl: "",
  port: undefined,
  pingInterval: 60,
  failThreshold: 3,
  acceptedStatusCodes: ["200-299"],
  tls: { no_verify: false },
  dns: [],
};

const PRESET_STATUS_CODES = [
  { label: "200 OK", value: "200" },
  { label: "200–299", value: "200-299" },
  { label: "2xx", value: "2xx" },
  { label: "301 Redirect", value: "301" },
  { label: "4xx", value: "4xx" },
  { label: "5xx", value: "5xx" },
];

interface DnsEntry {
  name: string;
  ip: string;
  port: number;
  protocol: string;
}

const emptyDns: DnsEntry = { name: "", ip: "", port: 53, protocol: "udp" };

function initForm(h?: Host): CreateHostRequest {
  if (!h) return { ...DEFAULT_VALUES, dns: [] };
  return {
    name: h.name,
    protocol: h.protocol,
    hostUrl: h.hostUrl,
    port: h.port,
    pingInterval: h.pingInterval,
    failThreshold: h.failThreshold,
    acceptedStatusCodes: h.acceptedStatusCodes ? [...h.acceptedStatusCodes] : undefined,
    tls: { ...h.tls },
    dns: (h.dns ?? []).map((d) => ({ ...d })),
  };
}

const HostForm: React.FC<HostFormProps> = ({ initialValues, onSuccess }) => {
  const toast = useToast();
  const [form, setForm] = useState<CreateHostRequest>(() => initForm(initialValues));
  const [loading, setLoading] = useState(false);
  const [codeInput, setCodeInput] = useState("");

  // Reinitialize the form when switching between create and edit targets.
  const [lastInitial, setLastInitial] = useState(initialValues);
  if (initialValues !== lastInitial) {
    setLastInitial(initialValues);
    setForm(initForm(initialValues));
  }

  const set = <K extends keyof CreateHostRequest>(key: K, val: CreateHostRequest[K]) =>
    setForm((prev) => ({ ...prev, [key]: val }));

  const isHTTP = form.protocol === "http" || form.protocol === "https";
  const isHTTPS = form.protocol === "https";
  const isPing = form.protocol === "ping";
  const needsPort = form.protocol === "tcp" || form.protocol === "udp";
  const showPort = !isPing;

  const changeProtocol = (protocol: CreateHostRequest["protocol"]) => {
    set("protocol", protocol);
    // Accepted status codes only mean something for http(s); reset on switch.
    const http = protocol === "http" || protocol === "https";
    set("acceptedStatusCodes", http ? ["200-299"] : undefined);
    if (http) set("tls", { no_verify: false });
    // ICMP ping doesn't use a port; drop any stale value.
    if (protocol === "ping") set("port", undefined);
  };

  const addCode = (code: string) => {
    const trimmed = code.trim();
    const codes = form.acceptedStatusCodes ?? [];
    if (trimmed && !codes.includes(trimmed)) {
      set("acceptedStatusCodes", [...codes, trimmed]);
    }
  };

  const removeCode = (code: string) => {
    set("acceptedStatusCodes", (form.acceptedStatusCodes ?? []).filter((c) => c !== code));
  };

  const handleCodeKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter" || e.key === ",") {
      e.preventDefault();
      addCode(codeInput);
      setCodeInput("");
    }
    if (e.key === "Backspace" && !codeInput && (form.acceptedStatusCodes ?? []).length > 0) {
      removeCode(form.acceptedStatusCodes![form.acceptedStatusCodes!.length - 1]);
    }
  };

  const addDns = () => set("dns", [...(form.dns ?? []), { ...emptyDns }]);

  const removeDns = (idx: number) =>
    set("dns", (form.dns ?? []).filter((_, i) => i !== idx));

  const updateDns = (idx: number, field: keyof DnsEntry, value: string | number) => {
    const dns = [...(form.dns ?? [])];
    dns[idx] = { ...dns[idx], [field]: value };
    set("dns", dns);
  };

  const onFinish = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.name || !form.hostUrl) {
      toast.error("Name and Host URL are required");
      return;
    }
    if (needsPort && !form.port) {
      toast.error("Port is required for TCP/UDP");
      return;
    }
    if (isHTTP && !form.acceptedStatusCodes?.length) {
      toast.error("At least one accepted status code is required");
      return;
    }
    setLoading(true);
    try {
      if (initialValues) {
        await updateHost(initialValues.id, form);
        toast.success("Host updated");
      } else {
        await createHost(form);
        toast.success("Host created");
      }
      onSuccess();
    } catch (err) {
      toast.error((err as Error).message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={onFinish}>
      <div className="form-grid form-grid-name-protocol mb-md">
        <div className="form-group">
          <label className="form-label">Name</label>
          <input
            className="form-input"
            placeholder="My Server"
            required
            value={form.name}
            onChange={(e) => set("name", e.target.value)}
          />
        </div>
        <div className="form-group">
          <label className="form-label">Protocol</label>
          <select
            className="form-select"
            value={form.protocol}
            onChange={(e) => changeProtocol(e.target.value as CreateHostRequest["protocol"])}
          >
            <option value="http">HTTP</option>
            <option value="https">HTTPS</option>
            <option value="tcp">TCP</option>
            <option value="udp">UDP</option>
            <option value="ping">Ping (ICMP)</option>
          </select>
        </div>
      </div>

      <div className={`form-grid ${showPort ? "form-grid-2" : ""} mb-md`}>
        <div className="form-group">
          <label className="form-label">Host URL / IP</label>
          <input
            className="form-input"
            placeholder="example.com"
            required
            value={form.hostUrl}
            onChange={(e) => set("hostUrl", e.target.value)}
          />
        </div>
        {showPort && (
          <div className="form-group">
            <label className="form-label">Port {needsPort && <span className="form-required">*</span>}</label>
            <input
              className="form-input"
              type="number"
              min={1}
              max={65535}
              required={needsPort}
              placeholder={isHTTP ? "Optional (e.g. 443)" : "Required"}
              value={form.port ?? ""}
              onChange={(e) => set("port", parseInt(e.target.value) || undefined)}
            />
          </div>
        )}
      </div>

      <div className="form-grid form-grid-3 mb-lg">
        <div className="form-group">
          <label className="form-label">Ping Interval (s)</label>
          <input
            className="form-input"
            type="number"
            min={1}
            required
            value={form.pingInterval}
            onChange={(e) => set("pingInterval", parseInt(e.target.value) || 60)}
          />
        </div>
        <div className="form-group">
          <label className="form-label">Fail Threshold</label>
          <input
            className="form-input"
            type="number"
            min={1}
            required
            value={form.failThreshold}
            onChange={(e) => set("failThreshold", parseInt(e.target.value) || 3)}
          />
        </div>
      </div>

      {/* DNS List */}
      <div className="mb-lg">
        <label className="form-label">DNS Servers</label>
        {(form.dns ?? []).map((entry, idx) => (
          <div key={idx} className="dns-entry">
            <div className="dns-entry-fields">
              <input
                className="form-input"
                placeholder="DNS Name"
                required
                value={entry.name}
                onChange={(e) => updateDns(idx, "name", e.target.value)}
              />
              <input
                className="form-input"
                placeholder="IP Address"
                required
                value={entry.ip}
                onChange={(e) => updateDns(idx, "ip", e.target.value)}
              />
              <select
                className="form-select"
                value={entry.protocol}
                onChange={(e) => updateDns(idx, "protocol", e.target.value)}
              >
                <option value="udp">UDP</option>
                <option value="tcp">TCP</option>
              </select>
              <input
                className="form-input"
                type="number"
                min={1}
                placeholder="Port"
                required
                value={entry.port}
                onChange={(e) => updateDns(idx, "port", parseInt(e.target.value) || 53)}
              />
            </div>
            <button type="button" className="btn btn-ghost btn-icon btn-sm btn-remove" onClick={() => removeDns(idx)}>
              <CloseIcon size={16} />
            </button>
          </div>
        ))}
        <button type="button" className="btn btn-ghost w-full" onClick={addDns} style={{ borderStyle: "dashed" }}>
          <PlusIcon size={16} /> Add DNS Server
        </button>
      </div>

      {/* Advanced: TLS + accepted status codes only apply to HTTP(S) */}
      {isHTTP && (
        <details className="mb-lg">
          <summary>Advanced</summary>
          <div className="details-body">
            {isHTTPS && (
              <div className="form-group">
                <label style={{ display: "flex", alignItems: "center", gap: "var(--space-sm)", cursor: "pointer" }}>
                  <input
                    type="checkbox"
                    checked={form.tls.no_verify}
                    onChange={(e) => set("tls", { no_verify: e.target.checked })}
                  />
                  Ignore TLS/SSL errors for HTTPS websites
                </label>
              </div>
            )}

            <div className="form-group">
              <label className="form-label">Accepted Status Codes</label>
              <p className="form-hint mb-sm">Formats: 200, 200-299, 2xx, 20x, 2*. Press Enter to add.</p>
              <div className="chip-input-wrap" onClick={(e) => (e.currentTarget.querySelector("input") as HTMLInputElement)?.focus()}>
                {(form.acceptedStatusCodes ?? []).map((code) => (
                  <span key={code} className="chip">
                    {code}
                    <button type="button" className="chip-remove" onClick={() => removeCode(code)}>×</button>
                  </span>
                ))}
                <input
                  className="chip-input"
                  value={codeInput}
                  onChange={(e) => setCodeInput(e.target.value)}
                  onKeyDown={handleCodeKeyDown}
                  placeholder={(form.acceptedStatusCodes ?? []).length === 0 ? "Type a code..." : ""}
                />
              </div>
              <div style={{ display: "flex", flexWrap: "wrap", gap: "var(--space-xs)", marginTop: "var(--space-sm)" }}>
                {PRESET_STATUS_CODES.filter((p) => !(form.acceptedStatusCodes ?? []).includes(p.value)).map((p) => (
                  <button
                    key={p.value}
                    type="button"
                    className="btn btn-ghost btn-sm"
                    onClick={() => addCode(p.value)}
                    style={{ fontSize: "0.75rem" }}
                  >
                    {p.label}
                  </button>
                ))}
              </div>
            </div>
          </div>
        </details>
      )}

      <button type="submit" className="btn btn-primary w-full" disabled={loading}>
        {loading ? "Saving..." : initialValues ? "Update Host" : "Create Host"}
      </button>
    </form>
  );
};

export default HostForm;
