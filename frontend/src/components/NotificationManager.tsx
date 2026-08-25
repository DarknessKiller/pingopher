import React, { useEffect, useState, useCallback, useRef } from "react";
import {
  getNotifications,
  createNotification,
  updateNotification,
  deleteNotification,
  type Host,
  type Notification,
  type CreateNotificationRequest,
} from "../api";
import { useToast } from "./Toast";
import { PlusIcon, EditIcon, DeleteIcon } from "./icons";
import ResponsiveButton from "./ResponsiveButton";
import Modal from "./Modal";
import ConfirmDialog from "./ConfirmDialog";

interface NotificationManagerProps {
  host: Host;
}

const NotificationManager: React.FC<NotificationManagerProps> = ({ host }) => {
  const toast = useToast();
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalLoading, setModalLoading] = useState(false);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingNotification, setEditingNotification] = useState<Notification | undefined>(undefined);
  const [confirmTarget, setConfirmTarget] = useState<{ id: string; label: string } | null>(null);

  const [form, setForm] = useState<CreateNotificationRequest>({
    name: "",
    type: "discord",
    active: true,
  });

  const controllerRef = useRef<AbortController | null>(null);

  const fetchNotifications = useCallback(async ({ showLoading = true } = {}) => {
    if (controllerRef.current) controllerRef.current.abort();
    const controller = new AbortController();
    controllerRef.current = controller;
    if (showLoading) setLoading(true);
    try {
      const response = await getNotifications(host.id, { signal: controller.signal });
      if (!controller.signal.aborted) setNotifications(response.data || []);
    } catch (error) {
      if (!controller.signal.aborted) toast.error((error as Error).message);
    } finally {
      if (!controller.signal.aborted && showLoading) setLoading(false);
    }
  }, [host.id, toast]);

  useEffect(() => {
    let cancelled = false;
    const controller = new AbortController();
    controllerRef.current = controller;

    const load = async () => {
      if (controllerRef.current !== controller) return;
      setLoading(true);
      try {
        const response = await getNotifications(host.id, { signal: controller.signal });
        if (!cancelled && !controller.signal.aborted) setNotifications(response.data || []);
      } catch (error) {
        if (!cancelled && !controller.signal.aborted) toast.error((error as Error).message);
      } finally {
        if (!cancelled && !controller.signal.aborted) setLoading(false);
      }
    };

    load();
    return () => {
      cancelled = true;
      controller.abort();
      if (controllerRef.current === controller) controllerRef.current = null;
    };
  }, [host.id, toast]);

  const setField = <K extends keyof CreateNotificationRequest>(key: K, val: CreateNotificationRequest[K]) =>
    setForm((prev) => ({ ...prev, [key]: val }));

  const handleCreate = () => {
    setEditingNotification(undefined);
    setForm({ name: "", type: "discord", active: true });
    setIsModalVisible(true);
  };

  const handleEdit = (record: Notification) => {
    setEditingNotification(record);
    setForm({
      name: record.name,
      type: record.type,
      active: record.active,
      discordWebhookUrl: record.discordWebhookUrl,
      discordUsername: record.discordUsername,
      discordPrefixMessage: record.discordPrefixMessage,
    });
    setIsModalVisible(true);
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteNotification(host.id, id);
      toast.success("Notification deleted");
      fetchNotifications();
    } catch (error) {
      toast.error((error as Error).message);
    }
  };

  const onFinish = async (e: React.FormEvent) => {
    e.preventDefault();
    setModalLoading(true);
    try {
      if (editingNotification) {
        await updateNotification(host.id, editingNotification.id, form);
        toast.success("Notification updated");
      } else {
        await createNotification(host.id, form);
        toast.success("Notification created");
      }
      setIsModalVisible(false);
      fetchNotifications();
    } catch (error) {
      toast.error((error as Error).message);
    } finally {
      setModalLoading(false);
    }
  };

  return (
    <div>
      <div className="flex" style={{ justifyContent: "flex-end", marginBottom: "var(--space-md)" }}>
        <ResponsiveButton text="Add Notification" icon={<PlusIcon size={16} />} onClick={handleCreate} />
      </div>

      {loading ? (
        <div style={{ display: "flex", justifyContent: "center", padding: "var(--space-xl)" }}>
          <div className="spinner" />
        </div>
      ) : (
        <div className="table-container">
          <table className="table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Type</th>
                <th>Active</th>
                <th style={{ textAlign: "right" }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {notifications.map((n) => (
                <tr key={n.id}>
                  <td style={{ fontWeight: 500 }}>{n.name}</td>
                  <td><span className="tag tag-info">{n.type.toUpperCase()}</span></td>
                  <td>
                    <span className={`tag ${n.active ? "tag-success" : "tag-danger"}`}>
                      {n.active ? "Active" : "Inactive"}
                    </span>
                  </td>
                  <td>
                    <div className="table-actions">
                      <button className="btn btn-ghost btn-icon btn-sm" onClick={() => handleEdit(n)}>
                        <EditIcon size={14} />
                      </button>
                      <button
                        className="btn btn-ghost btn-icon btn-sm"
                        style={{ color: "var(--danger)" }}
                        onClick={() => setConfirmTarget({ id: n.id, label: n.name })}
                      >
                        <DeleteIcon size={14} />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
              {notifications.length === 0 && (
                <tr><td colSpan={4} style={{ textAlign: "center", padding: "var(--space-xl)", color: "var(--text-tertiary)" }}>No notifications</td></tr>
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* Add/Edit Modal */}
      <Modal
        open={isModalVisible}
        onClose={() => setIsModalVisible(false)}
        title={editingNotification ? "Edit Notification" : "Add Notification"}
        zIndex={1100}
      >
        <form onSubmit={onFinish}>
          <div className="form-group">
            <label className="form-label">Name</label>
            <input
              className="form-input"
              placeholder="My Discord Alert"
              required
              value={form.name}
              onChange={(e) => setField("name", e.target.value)}
            />
          </div>

          <div className="form-group">
            <label className="form-label">Type</label>
            <select className="form-select" value={form.type} onChange={(e) => setField("type", e.target.value)}>
              <option value="discord">Discord</option>
            </select>
          </div>

          <div className="form-group">
            <label className="form-label">Active</label>
            <button
              type="button"
              className={`toggle-switch ${form.active ? "active" : ""}`}
              onClick={() => setField("active", !form.active)}
            />
          </div>

          {form.type === "discord" && (
            <>
              <div className="form-group">
                <label className="form-label">Webhook URL</label>
                <input
                  className="form-input"
                  type="url"
                  required
                  value={form.discordWebhookUrl ?? ""}
                  onChange={(e) => setField("discordWebhookUrl", e.target.value)}
                />
              </div>
              <div className="form-group">
                <label className="form-label">Username</label>
                <input
                  className="form-input"
                  placeholder="Pingopher"
                  value={form.discordUsername ?? ""}
                  onChange={(e) => setField("discordUsername", e.target.value)}
                />
              </div>
              <div className="form-group">
                <label className="form-label">Prefix Message</label>
                <textarea
                  className="form-input"
                  placeholder="@here"
                  rows={2}
                  style={{ resize: "vertical" }}
                  value={form.discordPrefixMessage ?? ""}
                  onChange={(e) => setField("discordPrefixMessage", e.target.value)}
                />
              </div>
            </>
          )}

          <div className="flex gap-sm" style={{ justifyContent: "flex-end", marginTop: "var(--space-lg)" }}>
            <button type="button" className="btn" onClick={() => setIsModalVisible(false)}>Cancel</button>
            <button type="submit" className="btn btn-primary" disabled={modalLoading}>
              {modalLoading ? "Saving..." : editingNotification ? "Update" : "Create"}
            </button>
          </div>
        </form>
      </Modal>

      <ConfirmDialog
        open={!!confirmTarget}
        title="Are you sure?"
        message={`Delete notification "${confirmTarget?.label}"? This action cannot be undone.`}
        onConfirm={() => {
          if (confirmTarget) handleDelete(confirmTarget.id);
          setConfirmTarget(null);
        }}
        onCancel={() => setConfirmTarget(null)}
      />
    </div>
  );
};

export default NotificationManager;
