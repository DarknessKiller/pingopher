  import React, { useEffect, useRef, useState, Suspense, lazy } from "react";
  import {
    Table,
    Button,
    Tag,
    Space,
    Modal,
    message,
    Row,
    Col,
    Card,
    Select,
    App,
    FormInstance,
  } from "antd";
  import {
    EditOutlined,
    DeleteOutlined,
    PlusOutlined,
    HistoryOutlined,
    BellOutlined,
  } from "@ant-design/icons";
  import { CreateHostRequest, getHosts, deleteHost, type Host } from "../api";
  import HostForm from "./HostForm";
  const HostDetail = lazy(() => import("./HostDetail"));
  import NotificationManager from "./NotificationManager";
  import type { ColumnsType } from "antd/es/table";
  import ResponsiveButton from "./ResponsiveButton";

  const POLL_INTERVAL = 10000;

  const HostList: React.FC = () => {
    const { modal } = App.useApp()
    const [hosts, setHosts] = useState<Host[]>([]);
    const [sortedHosts, setSortedHosts] = useState<Host[]>([]);
    const [loading, setLoading] = useState(false);

    const [isModalVisible, setIsModalVisible] = useState(false);
    const [editingHost, setEditingHost] = useState<Host | undefined>(undefined);
    const formRef = useRef<FormInstance<CreateHostRequest>>(null);

    const [detailVisible, setDetailVisible] = useState(false);
    const [notificationVisible, setNotificationVisible] = useState(false);

    const [selectedHost, setSelectedHost] = useState<Host | undefined>();

    const pollTimerRef = useRef<NodeJS.Timeout | null>(null);
    const fetchControllerRef = useRef<AbortController | null>(null);

    // --------------------------
    // Helpers
    // --------------------------
    const getHostUrl = (host: Host) => {
      const portPart = host.port && host.port !== 0 ? `:${host.port}` : "";
      return `${host.protocol}://${host.hostUrl}${portPart}`;
    };

    const renderStatusTag = (status: Host["status"]) => {
      const colorMap: Record<Host["status"], string> = {
        up: "green",
        down: "red",
        unknown: "gold",
      };
      return <Tag color={colorMap[status]}>{status.toUpperCase()}</Tag>;
    };

    const handleSort = (field: keyof Host, order: "ascend" | "descend") => {
      const sorted = [...hosts].sort((a, b) => {
        let compare = 0;
        switch (field) {
          case "name":
            compare = a.name.localeCompare(b.name);
            break;
          case "status":
            compare = a.status.localeCompare(b.status);
            break;
          case "pingInterval":
            compare = a.pingInterval - b.pingInterval;
            break;
          case "hostUrl":
            compare = getHostUrl(a).localeCompare(getHostUrl(b));
            break;
        }
        return order === "ascend" ? compare : -compare;
      });
      setSortedHosts(sorted);
    };

    // --------------------------
    // Fetch hosts
    // --------------------------
    const fetchHosts = async (showLoading = true) => {
      if (fetchControllerRef.current) fetchControllerRef.current.abort();
      const controller = new AbortController();
      fetchControllerRef.current = controller;

      try {
        if (showLoading) setLoading(true);
        const response = await getHosts({ signal: controller.signal });
        if (!controller.signal.aborted) {
          setHosts(response.data.hosts);
          setSortedHosts(response.data.hosts);
        }
      } catch (err) {
        if (!controller.signal.aborted) message.error((err as Error).message);
      } finally {
        if (!controller.signal.aborted && showLoading) setLoading(false);
      }
    };

    // --------------------------
    // Polling effect
    // --------------------------
    useEffect(() => {
      const controller = new AbortController();

      const runPolling = async () => {
        await fetchHosts(false);
        if (!controller.signal.aborted) {
          pollTimerRef.current = setTimeout(runPolling, POLL_INTERVAL);
        }
      };

      fetchHosts(true).then(() => {
        if (!controller.signal.aborted) {
          pollTimerRef.current = setTimeout(runPolling, POLL_INTERVAL);
        }
      });

      return () => {
        controller.abort();
        if (pollTimerRef.current) clearTimeout(pollTimerRef.current);
      };
    }, []);

    // --------------------------
    // Handlers
    // --------------------------
    const handleDelete = async (id: string) => {
      try {
        await deleteHost(id);
        message.success("Host deleted");
        await fetchHosts(false);
      } catch (err) {
        message.error((err as Error).message);
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

    const showDeleteConfirm = (host: Host) => {
      modal.confirm({
        title: "Are you sure?",
        content: "This action cannot be undone.",
        centered: true,
        onOk: () => handleDelete(host.id),
      });
    };

    // --------------------------
    // Table columns
    // --------------------------
    const columns: ColumnsType<Host> = [
      {
        title: "Name",
        dataIndex: "name",
        key: "name",
        sorter: (a, b) => a.name.localeCompare(b.name),
      },
      {
        title: "URL",
        key: "url",
        render: (_, record) => getHostUrl(record),
        sorter: (a, b) => getHostUrl(a).localeCompare(getHostUrl(b)),
      },
      {
        title: "Status",
        dataIndex: "status",
        render: renderStatusTag,
        sorter: (a, b) => a.status.localeCompare(b.status),
      },
      {
        title: "Interval",
        dataIndex: "pingInterval",
        render: (v: number) => `${v}s`,
        sorter: (a, b) => a.pingInterval - b.pingInterval,
      },
      {
        title: "Actions",
        key: "actions",
        align: "right",
        render: (_, record) => (
          <Space size="middle">
            <Button
              icon={<BellOutlined />}
              onClick={() => {
                setSelectedHost(record);
                setNotificationVisible(true);
              }}
            />
            <Button
              icon={<HistoryOutlined />}
              onClick={() => handleShowDetail(record)}
            />
            <Button icon={<EditOutlined />} onClick={() => handleEdit(record)} />
            <Button
              danger
              icon={<DeleteOutlined />}
              onClick={() => showDeleteConfirm(record)}
            />
          </Space>
        ),
      },
    ];

    // --------------------------
    // Render
    // --------------------------
    return (
      <div>
        {/* Header */}
        <div
          style={{
            marginBottom: 16,
            display: "flex",
            justifyContent: "space-between",
            flexWrap: "wrap",
            alignItems: "center",
          }}
        >
          <h2 style={{ margin: 0 }}>Monitoring Hosts</h2>
          <ResponsiveButton
            icon={<PlusOutlined />}
            text="Add Host"
            onClick={handleCreate}
          />
        </div>

        {/* Desktop Table */}
        <div className="desktop-table" style={{ display: "none" }}>
          <Table
            columns={columns}
            dataSource={hosts}
            rowKey="id"
            loading={loading}
            scroll={{ x: "max-content" }}
            onChange={(_pagination, _filters, sorter) => {
              if (!Array.isArray(sorter) && sorter.order && sorter.field) {
                handleSort(sorter.field as keyof Host, sorter.order);
              } else {
                setSortedHosts(hosts);
              }
            }}
          />
        </div>

        {/* Mobile Cards */}
        <div className="mobile-cards" style={{ display: "none" }}>
          {/* Mobile sorting dropdown */}
          <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
            <Col span={24}>
              <Select
                placeholder="Sort by"
                onChange={(value) => {
                  const [field, order] = (value as string).split("-");
                  handleSort(field as keyof Host, order as "ascend" | "descend");
                }}
                style={{ width: 200 }}
                allowClear
              >
                <Select.Option value="name-ascend">Name ↑</Select.Option>
                <Select.Option value="name-descend">Name ↓</Select.Option>
                <Select.Option value="status-ascend">Status ↑</Select.Option>
                <Select.Option value="status-descend">Status ↓</Select.Option>
                <Select.Option value="pingInterval-ascend">
                  Interval ↑
                </Select.Option>
                <Select.Option value="pingInterval-descend">
                  Interval ↓
                </Select.Option>
                <Select.Option value="url-ascend">URL ↑</Select.Option>
                <Select.Option value="url-descend">URL ↓</Select.Option>
              </Select>
            </Col>
          </Row>

          <Row gutter={[16, 16]}>
            {(sortedHosts.length > 0 ? sortedHosts : hosts).map((host) => (
              <Col xs={24} sm={12} key={host.id}>
                <Card
                  title={host.name}
                  extra={
                    <Space size="small">
                      <Button
                        icon={<BellOutlined />}
                        size="small"
                        onClick={() => {
                          setSelectedHost(host);
                          setNotificationVisible(true);
                        }}
                      />
                      <Button
                        icon={<HistoryOutlined />}
                        size="small"
                        onClick={() => handleShowDetail(host)}
                      />
                      <Button
                        icon={<EditOutlined />}
                        size="small"
                        onClick={() => handleEdit(host)}
                      />
                      <Button
                        danger
                        icon={<DeleteOutlined />}
                        size="small"
                        onClick={() => showDeleteConfirm(host)}
                      />
                    </Space>
                  }
                >
                  <p>
                    <strong>URL:</strong> {getHostUrl(host)}
                  </p>
                  <p>
                    <strong>Status:</strong> {renderStatusTag(host.status)}
                  </p>
                  <p>
                    <strong>Interval:</strong> {host.pingInterval}s
                  </p>
                </Card>
              </Col>
            ))}
          </Row>
        </div>

        {/* Modals */}
      <Modal
        title={editingHost ? "Edit Host" : "Add Host"}
        open={isModalVisible}
        onCancel={() => setIsModalVisible(false)}
        footer={null}
        centered
        afterOpenChange={(open) => {
          if (open && !editingHost && formRef.current) {
            formRef.current.resetFields();
          }
        }}
      >
        <HostForm
          ref={formRef}
          initialValues={editingHost}
          onSuccess={handleSuccess}
        />
      </Modal>

        <Modal
          title={`Details: ${selectedHost?.name ?? ""}`}
          open={detailVisible}
          onCancel={() => setDetailVisible(false)}
          footer={null}
          centered
          width="90%"
        >
          {selectedHost ? (
            <Suspense fallback={<div style={{ textAlign: "center", padding: "50px" }}>Loading Chart...</div>}>
              <HostDetail host={selectedHost} />
            </Suspense>
          ) : null}
        </Modal>

        <Modal
          title={`Notifications: ${selectedHost?.name ?? ""}`}
          open={notificationVisible}
          onCancel={() => setNotificationVisible(false)}
          footer={null}
          centered
          width="90%"
        >
          {selectedHost ? <NotificationManager host={selectedHost} /> : null}
        </Modal>

        {/* Responsive CSS */}
        <style>
          {`
            @media (min-width: 768px) {
              .desktop-table { display: block !important; }
              .mobile-cards { display: none !important; }
            }
            @media (max-width: 767px) {
              .desktop-table { display: none !important; }
              .mobile-cards { display: block !important; }
            }
          `}
        </style>
      </div>
    );
  };

  export default HostList;
