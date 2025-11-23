import React, { useEffect, useState } from "react";
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
} from "antd";
import {
  EditOutlined,
  DeleteOutlined,
  PlusOutlined,
  HistoryOutlined,
  BellOutlined,
} from "@ant-design/icons";
import { getHosts, deleteHost, type Host } from "../api";
import HostForm from "./HostForm";
import HostDetail from "./HostDetail";
import NotificationManager from "./NotificationManager";

const HostList: React.FC = () => {
  const [hosts, setHosts] = useState<Host[]>([]);
  const [loading, setLoading] = useState(false);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingHost, setEditingHost] = useState<Host | undefined>(undefined);
  const [detailVisible, setDetailVisible] = useState(false);
  const [notificationVisible, setNotificationVisible] = useState(false);
  const [selectedHost, setSelectedHost] = useState<Host | undefined>(undefined);

  const fetchHosts = async (showLoading = true) => {
    if (showLoading) setLoading(true);
    try {
      const response = await getHosts();
      setHosts(response.data.hosts);
    } catch (error) {
      message.error((error as Error).message);
    } finally {
      if (showLoading) setLoading(false);
    }
  };

  useEffect(() => {
    let isMounted = true;
    let timeoutId: ReturnType<typeof setTimeout>;

    const loop = async () => {
      await fetchHosts(false);
      if (isMounted) {
        timeoutId = setTimeout(loop, 10000);
      }
    };

    fetchHosts(true).then(() => {
      if (isMounted) {
        timeoutId = setTimeout(loop, 10000);
      }
    });

    return () => {
      isMounted = false;
      clearTimeout(timeoutId);
    };
  }, []);

  const handleDelete = async (id: string) => {
    try {
      await deleteHost(id);
      message.success("Host deleted");
      fetchHosts();
    } catch (error) {
      message.error((error as Error).message);
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

  const handleSuccess = () => {
    setIsModalVisible(false);
    fetchHosts();
  };

  const handleShowDetail = (host: Host) => {
    setSelectedHost(host);
    setDetailVisible(true);
  };

  const renderStatusTag = (status: string) => {
    const colorMap: Record<string, string> = {
      up: "green",
      down: "red",
      unknown: "yellow",
    };
    return (
      <Tag color={colorMap[status] || "default"}>{status.toUpperCase()}</Tag>
    );
  };

  // Responsive Table for Desktop
  const columns = [
    {
      title: "Name",
      dataIndex: "name",
      key: "name",
    },
    {
      title: "URL",
      key: "url",
      render: (_value: unknown, record: Host) => {
        const portStr =
          record.port && record.port !== 0 ? `:${record.port}` : "";
        return `${record.protocol}://${record.hostUrl}${portStr}`;
      },
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      render: renderStatusTag,
    },
    {
      title: "Interval",
      dataIndex: "pingInterval",
      key: "pingInterval",
      render: (text: number) => `${text}s`,
    },
    {
      title: "Actions",
      key: "actions",
      align: "right" as const,
      render: (_value: unknown, record: Host) => (
        <Space wrap size="middle">
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
            icon={<DeleteOutlined />}
            danger
            onClick={() =>
              Modal.confirm({
                title: "Are you sure?",
                content: "This action cannot be undone.",
                onOk: () => handleDelete(record.id),
              })
            }
          />
        </Space>
      ),
    },
  ];

  return (
    <div>
      {/* Header */}
      <div
        style={{
          marginBottom: 16,
          display: "flex",
          flexWrap: "wrap",
          justifyContent: "space-between",
          alignItems: "center",
          // gap: 8,
        }}
      >
        <h2 style={{ margin: 0 }}>Monitoring Hosts</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
          Add Host
        </Button>
      </div>

      {/* Desktop Table */}
      <div className="desktop-table" style={{ display: "none" }}>
        <Table
          columns={columns}
          dataSource={hosts}
          rowKey="id"
          loading={loading}
          scroll={{ x: "max-content" }}
        />
      </div>

      {/* Mobile Card Layout */}
      <div className="mobile-cards" style={{ display: "none" }}>
        <Row gutter={[16, 16]}>
          {hosts.map((host) => (
            <Col xs={24} sm={12} key={host.id}>
              <Card
                title={host.name}
                extra={
                  <Space wrap size="small">
                    <Button
                      icon={<BellOutlined />}
                      onClick={() => {
                        setSelectedHost(host);
                        setNotificationVisible(true);
                      }}
                      size="small"
                    />
                    <Button
                      icon={<HistoryOutlined />}
                      onClick={() => handleShowDetail(host)}
                      size="small"
                    />
                    <Button
                      icon={<EditOutlined />}
                      onClick={() => handleEdit(host)}
                      size="small"
                    />
                    <Button
                      icon={<DeleteOutlined />}
                      danger
                      size="small"
                      onClick={() =>
                        Modal.confirm({
                          title: "Are you sure?",
                          content: "This action cannot be undone.",
                          onOk: () => handleDelete(host.id),
                        })
                      }
                    />
                  </Space>
                }
              >
                <p>
                  <strong>URL:</strong>{" "}
                  {`${host.protocol}://${host.hostUrl}${host.port && host.port !== 0 ? `:${host.port}` : ""
                    }`}
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
      >
        <HostForm initialValues={editingHost} onSuccess={handleSuccess} />
      </Modal>

      <Modal
        title={`Details: ${selectedHost?.name}`}
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={null}
        centered
        width={"90%"}
      >
        {selectedHost && <HostDetail host={selectedHost} />}
      </Modal>

      <Modal
        title={`Notifications: ${selectedHost?.name}`}
        open={notificationVisible}
        onCancel={() => setNotificationVisible(false)}
        footer={null}
        centered
        width={"90%"}
      >
        {selectedHost && <NotificationManager host={selectedHost} />}
      </Modal>

      {/* CSS to switch layouts based on screen size */}
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
