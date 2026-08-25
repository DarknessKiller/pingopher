import React, { useEffect, useRef, useState, useMemo, Suspense, lazy } from "react";
import { getHosts, deleteHost, type Host } from "../api";
import { useToast } from "./Toast";
import { PlusIcon, BellIcon, HistoryIcon, EditIcon, DeleteIcon } from "./icons";
import HostForm from "./HostForm";
import ResponsiveButton from "./ResponsiveButton";
import Modal from "./Modal";
import ConfirmDialog from "./ConfirmDialog";

const HostDetail = lazy(() => import("./HostDetail"));
const NotificationManager = lazy(() => import("./NotificationManager"));

const POLL_INTERVAL = 60000;

const getHostUrl = (host: Host) => {
  const portPart = host.port && host.port !== 0 ? `:${host.port}` : "";
  switch (host.protocol) {
    case "tcp":
    case "udp":
      return `${host.hostUrl}${portPart}`;
    case "ping":
      return host.hostUrl;
    default:
      return `${host.protocol}://${host.hostUrl}${portPart}`;
  }
};

const comparators: Record<string, (a: Host, b: Host) => number> = {
  name: (a, b) => a.name.localeCompare(b.name),
  status: (a, b) => a.status.localeCompare(b.status),
  pingInterval: (a, b) => a.pingInterval - b.pingInterval,
  hostUrl: (a, b) => getHostUrl(a).localeCompare(getHostUrl(b)),
};

type SortField = keyof Host | "url";

const HostList: React.FC = () => {
  const toast = useToast();
  const [hosts, setHosts] = useState<Host[]>([]);
  const [loading, setLoading] = useState(true);
  const [sortField, setSortField] = useState<SortField | null>(null);
  const [sortOrder, setSortOrder] = useState<"ascend" | "descend" | null>(null);

  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingHost, setEditingHost] = useState<Host | undefined>(undefined);
  const [detailVisible, setDetailVisible] = useState(false);
  const [notificationVisible, setNotificationVisible] = useState(false);
  const [selectedHost, setSelectedHost] = useState<Host | undefined>();
  const [confirmTarget, setConfirmTarget] = useState<{ id: string; label: string } | null>(null);

  const pollTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const fetchControllerRef = useRef<AbortController | null>(null);

  const sortedHosts = useMemo(() => {
    if (!sortField || !sortOrder) return hosts;
    const field = sortField === "url" ? "hostUrl" : sortField;
    const cmp = comparators[field];
    if (!cmp) return hosts;
    const dir = sortOrder === "ascend" ? 1 : -1;
    return [...hosts].sort((a, b) => dir * cmp(a, b));
  }, [hosts, sortField, sortOrder]);

  const fetchHosts = async (showLoading = true) => {
    if (fetchControllerRef.current) fetchControllerRef.current.abort();
    const controller = new AbortController();
    fetchControllerRef.current = controller;
    try {
      if (showLoading) setLoading(true);
      const response = await getHosts({ signal: controller.signal });
      if (!controller.signal.aborted) setHosts(response.data.hosts);
    } catch (err) {
      if (!controller.signal.aborted) toast.error((err as Error).message);
    } finally {
      if (!controller.signal.aborted && showLoading) setLoading(false);
    }
  };

  useEffect(() => {
    const controller = new AbortController();
    const runPolling = async () => {
      await fetchHosts(false);
      if (!controller.signal.aborted) pollTimerRef.current = setTimeout(runPolling, POLL_INTERVAL);
    };
    (async () => {
      await fetchHosts(true);
      if (!controller.signal.aborted) pollTimerRef.current = setTimeout(runPolling, POLL_INTERVAL);
    })();
    return () => {
      controller.abort();
      if (pollTimerRef.current) clearTimeout(pollTimerRef.current);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleDelete = async (id: string) => {
    try {
      await deleteHost(id);
      toast.success("Host deleted");
      await fetchHosts(false);
    } catch (err) {
      toast.error((err as Error).message);
    }
  };

  const handleEdit = (host: Host) => {
    setEditingHost(host);
    setIsModalVisible(true);
  };

  const handleCreate = () => {
    setEditingHost(undefined);
    setIsModalVisible(true);
  };

  const handleSuccess = async () => {
    setIsModalVisible(false);
    await fetchHosts(false);
  };

  const handleShowDetail = (host: Host) => {
    setSelectedHost(host);
    setDetailVisible(true);
  };

  const openNotifications = (host: Host) => {
    setSelectedHost(host);
    setNotificationVisible(true);
  };

  const toggleSort = (field: SortField) => {
    if (sortField === field) {
      if (sortOrder === "ascend") setSortOrder("descend");
      else if (sortOrder === "descend") {
        setSortField(null);
        setSortOrder(null);
      }
    } else {
      setSortField(field);
      setSortOrder("ascend");
    }
  };

  const sortArrow = (field: SortField) => {
    if (sortField !== field) return null;
    return <span className="sort-arrow">{sortOrder === "ascend" ? "▲" : "▼"}</span>;
  };

  const renderStatus = (status: Host["status"]) => (
    <span className="flex-center" style={{ gap: 0 }}>
      <span className={`status-dot status-${status}`} />
      <span className="status-text">{status}</span>
    </span>
  );

  const HostActions: React.FC<{ host: Host; size?: "small" | "normal" }> = ({ host, size = "normal" }) => (
    <div className="table-actions">
      <button className={`btn btn-ghost btn-icon ${size === "small" ? "btn-sm" : ""}`} onClick={() => openNotifications(host)}>
        <BellIcon size={size === "small" ? 14 : 16} />
      </button>
      <button className={`btn btn-ghost btn-icon ${size === "small" ? "btn-sm" : ""}`} onClick={() => handleShowDetail(host)}>
        <HistoryIcon size={size === "small" ? 14 : 16} />
      </button>
      <button className={`btn btn-ghost btn-icon ${size === "small" ? "btn-sm" : ""}`} onClick={() => handleEdit(host)}>
        <EditIcon size={size === "small" ? 14 : 16} />
      </button>
      <button
        className={`btn btn-ghost btn-icon ${size === "small" ? "btn-sm" : ""}`}
        style={{ color: "var(--danger)" }}
        onClick={() => setConfirmTarget({ id: host.id, label: host.name })}
      >
        <DeleteIcon size={size === "small" ? 14 : 16} />
      </button>
    </div>
  );

  return (
    <div>
      {/* Header */}
      <div className="section-header">
        <div>
          <span className="mobile-brand">Pingopher</span>
          <h2>Monitoring Hosts</h2>
        </div>
        <ResponsiveButton icon={<PlusIcon size={16} />} text="Add Host" onClick={handleCreate} />
      </div>

      {/* Desktop Table */}
      <div className="host-table-wrap">
        {loading ? (
          <div className="card" style={{ display: "flex", justifyContent: "center", padding: "var(--space-xl)" }}>
            <div className="spinner" />
          </div>
        ) : (
          <div className="table-container">
            <table className="table">
              <thead>
                <tr>
                  <th className={sortField === "name" ? "sort-active" : ""} onClick={() => toggleSort("name")}>
                    Name {sortArrow("name")}
                  </th>
                  <th className={sortField === "url" ? "sort-active" : ""} onClick={() => toggleSort("url")}>
                    URL {sortArrow("url")}
                  </th>
                  <th className={sortField === "status" ? "sort-active" : ""} onClick={() => toggleSort("status")}>
                    Status {sortArrow("status")}
                  </th>
                  <th className={sortField === "pingInterval" ? "sort-active" : ""} onClick={() => toggleSort("pingInterval")}>
                    Interval {sortArrow("pingInterval")}
                  </th>
                  <th style={{ textAlign: "right" }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {sortedHosts.map((host) => (
                  <tr key={host.id}>
                    <td style={{ fontWeight: 500 }}>{host.name}</td>
                    <td style={{ color: "var(--text-secondary)", fontSize: "0.875rem" }}>{getHostUrl(host)}</td>
                    <td>{renderStatus(host.status as Host["status"])}</td>
                    <td>{host.pingInterval}s</td>
                    <td><HostActions host={host} /></td>
                  </tr>
                ))}
                {sortedHosts.length === 0 && (
                  <tr><td colSpan={5} style={{ textAlign: "center", padding: "var(--space-xl)", color: "var(--text-tertiary)" }}>No hosts yet</td></tr>
                )}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Mobile Cards */}
      <div className="host-cards">
        <div className="sort-bar">
          <select
            className="form-select"
            value={sortField && sortOrder ? `${sortField}-${sortOrder}` : ""}
            onChange={(e) => {
              if (!e.target.value) {
                setSortField(null);
                setSortOrder(null);
              } else {
                const [field, order] = e.target.value.split("-");
                setSortField(field as SortField);
                setSortOrder(order as "ascend" | "descend");
              }
            }}
          >
            <option value="">Sort by...</option>
            <option value="name-ascend">Name ↑</option>
            <option value="name-descend">Name ↓</option>
            <option value="status-ascend">Status ↑</option>
            <option value="status-descend">Status ↓</option>
            <option value="pingInterval-ascend">Interval ↑</option>
            <option value="pingInterval-descend">Interval ↓</option>
            <option value="url-ascend">URL ↑</option>
            <option value="url-descend">URL ↓</option>
          </select>
        </div>
        {sortedHosts.map((host) => (
          <div key={host.id} className="card host-card mb-md">
            <div className="host-card-header">
              <span className="host-card-name">{host.name}</span>
              <HostActions host={host} size="small" />
            </div>
            <div className="host-card-url">{getHostUrl(host)}</div>
            <div className="host-card-meta">
              {renderStatus(host.status as Host["status"])}
              <span>Interval: {host.pingInterval}s</span>
            </div>
          </div>
        ))}
      </div>

      {/* Edit/Create Modal */}
      <Modal
        open={isModalVisible}
        onClose={() => setIsModalVisible(false)}
        title={editingHost ? "Edit Host" : "Add Host"}
      >
        <HostForm
          key={editingHost?.id ?? "new"}
          initialValues={editingHost}
          onSuccess={handleSuccess}
        />
      </Modal>

      {/* Detail Modal */}
      <Modal open={detailVisible} onClose={() => setDetailVisible(false)} title={`Details: ${selectedHost?.name ?? ""}`} width={900}>
        {selectedHost && (
          <Suspense fallback={<div style={{ textAlign: "center", padding: "var(--space-xl)" }}><div className="spinner spinner-lg" /></div>}>
            <HostDetail key={selectedHost.id} host={selectedHost} />
          </Suspense>
        )}
      </Modal>

      {/* Notification Modal */}
      <Modal open={notificationVisible} onClose={() => setNotificationVisible(false)} title={`Notifications: ${selectedHost?.name ?? ""}`} width={900}>
        {selectedHost && (
          <Suspense fallback={<div style={{ textAlign: "center", padding: "var(--space-xl)" }}><div className="spinner spinner-lg" /></div>}>
            <NotificationManager key={selectedHost.id} host={selectedHost} />
          </Suspense>
        )}
      </Modal>

      {/* Delete Confirmation */}
      <ConfirmDialog
        open={!!confirmTarget}
        title="Are you sure?"
        message={`Delete "${confirmTarget?.label}"? This action cannot be undone.`}
        onConfirm={() => {
          if (confirmTarget) handleDelete(confirmTarget.id);
          setConfirmTarget(null);
        }}
        onCancel={() => setConfirmTarget(null)}
      />
    </div>
  );
};

export default HostList;
